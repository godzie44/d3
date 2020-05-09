package orm

import (
	"d3/orm/entity"
	"d3/orm/persistence"
	d3reflect "d3/reflect"
	"fmt"
)

type dirtyEl struct {
	box      *entity.Box
	original interface{}
}

type UnitOfWork struct {
	newEntities     map[entity.Name][]*entity.Box
	dirtyEntities   map[entity.Name]map[interface{}]*dirtyEl
	deletedEntities map[entity.Name]map[interface{}]*entity.Box

	storage     Storage
	identityMap *identityMap

	currentTx Transaction
}

func NewUOW(storage Storage) *UnitOfWork {
	return &UnitOfWork{
		newEntities:     make(map[entity.Name][]*entity.Box),
		dirtyEntities:   make(map[entity.Name]map[interface{}]*dirtyEl),
		deletedEntities: make(map[entity.Name]map[interface{}]*entity.Box),
		storage:         storage,
		identityMap:     newIdentityMap(),
	}
}

func (uow *UnitOfWork) registerNew(box *entity.Box) error {
	pkVal, err := box.ExtractPk()
	if err != nil {
		return fmt.Errorf("while adding entity to new: %w", err)
	}

	if _, exists := uow.dirtyEntities[box.GetEName()][pkVal]; exists {
		return nil
	}

	if _, exists := uow.newEntities[box.GetEName()]; !exists {
		uow.newEntities[box.Meta.EntityName] = make([]*entity.Box, 0)
	}

	uow.newEntities[box.GetEName()] = append(uow.newEntities[box.GetEName()], box)

	return nil
}

func (uow *UnitOfWork) registerDirty(box *entity.Box) error {
	pkVal, err := box.ExtractPk()
	if err != nil {
		return fmt.Errorf("while adding entity to dirty: %w", err)
	}

	if _, exists := uow.dirtyEntities[box.GetEName()]; !exists {
		uow.dirtyEntities[box.GetEName()] = make(map[interface{}]*dirtyEl, 0)
	}

	uow.dirtyEntities[box.GetEName()][pkVal] = &dirtyEl{
		box:      box,
		original: d3reflect.Copy(box.Entity),
	}

	return nil
}

func (uow *UnitOfWork) updateFieldOfOriginal(box *entity.Box, fieldName string, newVal interface{}) {
	pkVal, err := box.ExtractPk()
	if err != nil {
		return
	}

	if _, exists := uow.dirtyEntities[box.GetEName()]; !exists {
		return
	}

	if _, exists := uow.dirtyEntities[box.GetEName()][pkVal]; !exists {
		return
	}

	cp := d3reflect.Copy(newVal)

	_ = d3reflect.SetFields(
		uow.dirtyEntities[box.GetEName()][pkVal].original,
		map[string]interface{}{fieldName: cp},
	)
}

func (uow *UnitOfWork) registerRemove(box *entity.Box) error {
	pkVal, err := box.ExtractPk()
	if err != nil {
		return err
	}

	uow.clean(box, pkVal)

	if _, exists := uow.deletedEntities[box.GetEName()]; !exists {
		uow.deletedEntities[box.GetEName()] = make(map[interface{}]*entity.Box)
	}
	uow.deletedEntities[box.GetEName()][pkVal] = box

	return nil
}

func (uow *UnitOfWork) clean(box *entity.Box, pk interface{}) {
	var i int
	for _, b := range uow.newEntities[box.GetEName()] {
		if b != box {
			uow.newEntities[box.GetEName()][i] = b
			i++
		}
	}

	for j := i; j < len(uow.newEntities); j++ {
		uow.newEntities[box.GetEName()][j] = nil
	}
	uow.newEntities[box.GetEName()] = uow.newEntities[box.GetEName()][:i]

	delete(uow.dirtyEntities[box.GetEName()], pk)
}

func (uow *UnitOfWork) Commit() error {
	graph := persistence.NewPersistGraph(uow.checkInDirty, uow.getOriginal)

	err := uow.processNew(graph)
	if err != nil {
		return err
	}

	err = uow.processDirty(graph)
	if err != nil {
		return err
	}

	err = uow.processDelete(graph)
	if err != nil {
		return err
	}

	defer func() {
		uow.newEntities = make(map[entity.Name][]*entity.Box)
		uow.deletedEntities = make(map[entity.Name]map[interface{}]*entity.Box)
	}()

	if uow.currentTx == nil {
		tx, err := uow.storage.BeginTx()
		if err != nil {
			return err
		}

		err = persistence.NewExecutor(uow.storage.MakePusher(tx), uow.moveInsertedBoxToDirty).Exec(graph)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		return tx.Commit()
	}

	return persistence.NewExecutor(uow.storage.MakePusher(uow.currentTx), uow.moveInsertedBoxToDirty).Exec(graph)
}

func (uow *UnitOfWork) moveInsertedBoxToDirty(act persistence.CompositeAction) {
	if ia, ok := act.(*persistence.InsertAction); ok {
		if box := ia.Box(); box != nil {
			_ = uow.registerDirty(box)
		}
	}
}

func (uow *UnitOfWork) processNew(graph *persistence.PersistGraph) error {
	for _, newEntities := range uow.newEntities {
		for _, b := range newEntities {
			if err := graph.ProcessEntity(b); err != nil {
				return err
			}
		}
	}

	return nil
}

func (uow *UnitOfWork) processDirty(graph *persistence.PersistGraph) error {
	for _, dirtyEntities := range uow.dirtyEntities {
		for _, dirtyEntity := range dirtyEntities {
			err := graph.ProcessEntity(dirtyEntity.box)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (uow *UnitOfWork) processDelete(graph *persistence.PersistGraph) error {
	for _, deletedEntities := range uow.deletedEntities {
		for _, b := range deletedEntities {
			if err := graph.ProcessDeletedEntity(b); err != nil {
				return err
			}
		}
	}

	return nil
}

func (uow *UnitOfWork) getOriginal(box *entity.Box) interface{} {
	if pk, err := box.ExtractPk(); err == nil {
		el, exists := uow.dirtyEntities[box.GetEName()][pk]
		if exists {
			return el.original
		}
	}

	return nil
}

func (uow *UnitOfWork) checkInDirty(box *entity.Box) (bool, error) {
	if pk, err := box.ExtractPk(); err == nil {
		_, exists := uow.dirtyEntities[box.GetEName()][pk]
		return exists, nil
	} else {
		return false, err
	}
}

func (uow *UnitOfWork) beginTx() error {
	tx, err := uow.storage.BeginTx()
	if err != nil {
		return err
	}

	uow.currentTx = tx
	return nil
}

func (uow *UnitOfWork) commitTx() error {
	if uow.currentTx == nil {
		return fmt.Errorf("begin transaction before commit")
	}
	defer func() {
		uow.currentTx = nil
	}()
	return uow.currentTx.Commit()
}

func (uow *UnitOfWork) rollbackTx() error {
	if uow.currentTx == nil {
		return fmt.Errorf("begin transaction before rollback")
	}
	defer func() {
		uow.currentTx = nil
	}()
	return uow.currentTx.Rollback()
}
