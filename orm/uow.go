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
	dirtyEntities   map[entity.Name]map[interface{}]dirtyEl
	deletedEntities map[entity.Name]map[interface{}]interface{}

	storage     Storage
	identityMap *identityMap
}

func NewUOW(storage Storage) *UnitOfWork {
	return &UnitOfWork{
		newEntities:     make(map[entity.Name][]*entity.Box),
		dirtyEntities:   make(map[entity.Name]map[interface{}]dirtyEl),
		deletedEntities: make(map[entity.Name]map[interface{}]interface{}),
		storage:         storage,
		identityMap:     newIdentityMap(),
	}
}

func (uow *UnitOfWork) registerNew(box *entity.Box) error {
	pkVal, err := box.ExtractPk()
	if err != nil {
		return fmt.Errorf("while adding entity to new: %w", err)
	}

	if _, exists := uow.newEntities[box.GetEName()]; !exists {
		uow.newEntities[box.Meta.EntityName] = make([]*entity.Box, 0)
	}

	if _, exists := uow.dirtyEntities[box.GetEName()][pkVal]; exists {
		return nil
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
		uow.dirtyEntities[box.GetEName()] = make(map[interface{}]dirtyEl, 0)
	}

	uow.dirtyEntities[box.GetEName()][pkVal] = dirtyEl{
		box:      box,
		original: d3reflect.Copy(box.Entity),
	}

	return nil
}

//
//func (uow *UnitOfWork) registerRemove(entity DomainEntity) {
//	uow.registerClean(entity)
//	uow.deletedEntities[entity.GetId()] = entity
//}
//
//func (uow *UnitOfWork) registerClean(entity DomainEntity) {
//	delete(uow.newEntities, entity.GetId())
//	delete(uow.dirtyEntities, entity.GetId())
//}

func (uow *UnitOfWork) Commit() error {
	graph := persistence.NewPersistGraph(uow.checkInDirty, uow.getOriginal)

	err := uow.processGraphByNew(graph)
	if err != nil {
		return err
	}

	err = uow.processGraphByDirty(graph)
	if err != nil {
		return err
	}

	defer func() {
		uow.newEntities = make(map[entity.Name][]*entity.Box)
	}()

	return persistence.NewExecutor(uow.storage, uow.moveInsertedBoxToDirty).Exec(graph)
}

func (uow *UnitOfWork) moveInsertedBoxToDirty(act persistence.CompositeAction) {
	if ia, ok := act.(*persistence.InsertAction); ok {
		if box := ia.Box(); box != nil {
			_ = uow.registerDirty(box)
		}
	}
}

func (uow *UnitOfWork) processGraphByNew(graph *persistence.PersistGraph) error {
	for _, newEntities := range uow.newEntities {
		for _, b := range newEntities {
			if err := graph.ProcessEntity(b); err != nil {
				return err
			}
		}
	}

	return nil
}

func (uow *UnitOfWork) processGraphByDirty(graph *persistence.PersistGraph) error {
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
