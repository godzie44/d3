package persistence

import (
	d3entity "d3/orm/entity"
	"d3/reflect"
	"fmt"
	"math"
)

type state int

const (
	create state = iota
	update
	createInProcess
	updateInProcess
	createProcessed
	updateProcessed
)

func (s state) isUpdate() bool {
	switch s {
	case update, updateInProcess, updateProcessed:
		return true
	}
	return false
}

func (s state) isCreate() bool {
	switch s {
	case create, createInProcess, createProcessed:
		return true
	}
	return false
}

func (s state) isInProcess() bool {
	switch s {
	case createInProcess, updateInProcess:
		return true
	}
	return false
}

func (s *state) toInProcess() {
	switch *s {
	case create, createProcessed:
		*s = createInProcess
	case update, updateProcessed:
		*s = updateInProcess
	}
}

func (s state) isProcessed() bool {
	switch s {
	case createProcessed, updateProcessed:
		return true
	}
	return false
}

func (s *state) toProcessed() {
	switch *s {
	case create, createInProcess:
		*s = createProcessed
	case update, updateInProcess:
		*s = updateProcessed
	}
}

type persistBox struct {
	*d3entity.Box
	entityPk  interface{}
	action    CompositeAction
	currState state
	original  interface{}
}

func newPersistBox(b *d3entity.Box, original interface{}) (*persistBox, error) {
	pk, err := b.ExtractPk()
	if err != nil {
		return nil, err
	}

	return &persistBox{Box: b, currState: create, entityPk: pk, original: original}, nil
}

func (p *persistBox) makeAction() CompositeAction {
	switch {
	case p.currState.isUpdate():
		return makeUpdateAction(p)
	case p.currState.isCreate():
		return makeInsertAction(p)
	default:
		panic("unknown state type")
	}
}

func makeUpdateAction(box *persistBox) *UpdateAction {
	a := NewUpdateAction(map[string]interface{}{box.Meta.Pk.Field.Name: createIDPromise(box)})
	a.setTableName(box.Meta.TableName)
	return a
}

func makeInsertAction(box *persistBox) *InsertAction {
	a := NewInsertAction(func(pk []interface{}) error {
		//currently only 1 pk
		return reflect.SetFields(box.Entity, map[string]interface{}{
			box.Meta.Pk.Field.Name: pk[0],
		})
	}, box)
	a.setTableName(box.Meta.TableName)
	return a
}

type pBoxContainer struct {
	boxes            map[interface{}]*persistBox
	stateInitializer func(box *d3entity.Box) (bool, error)
	originalFactory  func(box *d3entity.Box) interface{}
}

func (p *pBoxContainer) get(parent *d3entity.Box) (*persistBox, error) {
	if _, exists := p.boxes[parent.Entity]; !exists {
		var err error
		p.boxes[parent.Entity], err = newPersistBox(parent, p.originalFactory(parent))
		if err != nil {
			return nil, fmt.Errorf("box cannot be generated: %w", err)
		}

		isDirty, err := p.stateInitializer(p.boxes[parent.Entity].Box)
		if err != nil {
			return nil, fmt.Errorf("box cannot be generated: %w", err)
		}

		if isDirty {
			p.boxes[parent.Entity].currState = update
		}
	}

	return p.boxes[parent.Entity], nil
}

func (p *pBoxContainer) getRaw(entity interface{}, meta *d3entity.MetaInfo) (*persistBox, error) {
	return p.get(d3entity.NewBox(entity, meta))
}

func (p *pBoxContainer) flattActions() []CompositeAction {
	var result []CompositeAction
	for _, pb := range p.boxes {
		result = append(result, pb.action)
	}
	return result
}

//PersistGraph graph of database actions derived from entities.
type PersistGraph struct {
	knownBoxes *pBoxContainer
}

//NewPersistGraph create new graph.
func NewPersistGraph(
	checkInDirty func(box *d3entity.Box) (bool, error),
	originalFactory func(box *d3entity.Box) interface{},
) *PersistGraph {
	return &PersistGraph{
		knownBoxes: &pBoxContainer{boxes: map[interface{}]*persistBox{}, stateInitializer: checkInDirty, originalFactory: originalFactory},
	}
}

//ProcessEntity process entity and all related entities into database actions.
func (p *PersistGraph) ProcessEntity(box *d3entity.Box) error {
	pb, err := p.knownBoxes.get(box)
	if err != nil {
		return err
	}

	return p.processBox(pb)
}

func (p *PersistGraph) processBox(box *persistBox) error {
	if box.currState.isProcessed() || box.currState.isInProcess() {
		return nil
	}

	box.currState.toInProcess()
	defer box.currState.toProcessed()

	box.action = box.makeAction()

	extractedFields, err := extractSimpleFields(box)
	if err != nil {
		return err
	}

	box.action.setFields(extractedFields...)

	for _, rel := range box.Meta.OneToOneRelations() {
		if err := p.processOneToOneRel(box, rel); err != nil {
			return err
		}
	}

	for _, rel := range box.Meta.OneToManyRelations() {
		if err := p.processOneToManyRel(box, rel); err != nil {
			return err
		}
	}

	for _, rel := range box.Meta.ManyToManyRelations() {
		if err := p.processManyToManyRel(box, rel); err != nil {
			return err
		}
	}

	return nil
}

func (p *PersistGraph) processOneToOneRel(box *persistBox, relation *d3entity.OneToOne) error {
	relatedEntity, err := relation.Extract(box.Entity)
	if err != nil {
		return err
	}

	var origRelatedEntity d3entity.WrappedEntity = d3entity.NewWrapEntity(nil)
	if box.original != nil {
		origRelatedEntity, err = relation.Extract(box.original)
		if err != nil {
			return err
		}
	}

	if relatedEntity.Unwrap() == origRelatedEntity.Unwrap() {
		return nil
	}

	if relatedEntity.IsNil() {
		box.action.mergeFields(ActionField(relation.JoinColumn, nil))
	} else {
		relatedBox, err := p.knownBoxes.getRaw(relatedEntity.Unwrap(), box.GetRelatedMeta(relation.RelatedWith()))
		if err != nil {
			return err
		}

		//split here, cycle detected
		if relatedBox.currState.isProcessed() || relatedBox.currState.isInProcess() {
			doSplit(box.action, relatedBox.action, box, relation.JoinColumn, createIDPromise(relatedBox))
		} else {
			if err := p.processBox(relatedBox); err != nil {
				return err
			}
			relatedBox.action.addChild(box.action)
			box.action.mergeFields(ActionField(relation.JoinColumn, createIDPromise(relatedBox)))
		}
	}

	return nil
}

func (p *PersistGraph) processOneToManyRel(box *persistBox, relation *d3entity.OneToMany) error {
	newCollection, err := relation.ExtractCollection(box.Entity)
	if err != nil {
		return err
	}
	relatedEntities := revertIntoMap(newCollection.ToSlice())

	origRelatedEntities := make(map[interface{}]struct{})
	if box.original != nil {
		origCollection, err := relation.ExtractCollection(box.original)
		if err != nil {
			return err
		}
		origRelatedEntities = revertIntoMap(origCollection.ToSlice())
	}

	relatedMeta := box.GetRelatedMeta(relation.RelatedWith())
	for _, relatedEntity := range mapKeyDiff(relatedEntities, origRelatedEntities) {
		relatedBox, err := p.knownBoxes.getRaw(relatedEntity, relatedMeta)
		if err != nil {
			return err
		}

		//split here, cycle detected
		if relatedBox.currState.isProcessed() || relatedBox.currState.isInProcess() {
			doSplit(relatedBox.action, box.action, relatedBox, relation.JoinColumn, createIDPromise(box))
		} else {
			if err := p.processBox(relatedBox); err != nil {
				return err
			}
			box.action.addChild(relatedBox.action)
			relatedBox.action.mergeFields(ActionField(relation.JoinColumn, createIDPromise(box)))
		}
	}

	for _, origRelatedEntity := range mapKeyDiff(origRelatedEntities, relatedEntities) {
		updPk, err := relatedMeta.ExtractPkValue(origRelatedEntity)
		if err != nil {
			return err
		}

		updAction := NewUpdateAction(map[string]interface{}{
			relatedMeta.Pk.Field.Name: updPk,
		})
		updAction.setFields(ActionField(relation.JoinColumn, nil))
		updAction.setTableName(relatedMeta.TableName)

		box.action.addChild(updAction)
	}

	return nil
}

func (p *PersistGraph) processManyToManyRel(box *persistBox, relation *d3entity.ManyToMany) error {
	newCollection, err := relation.ExtractCollection(box.Entity)
	if err != nil {
		return err
	}
	relatedEntities := revertIntoMap(newCollection.ToSlice())

	origRelatedEntities := make(map[interface{}]struct{})
	if box.original != nil {
		origCollection, err := relation.ExtractCollection(box.original)
		if err != nil {
			return err
		}
		origRelatedEntities = revertIntoMap(origCollection.ToSlice())
	}

	relatedMeta := box.GetRelatedMeta(relation.RelatedWith())
	for _, relatedEntity := range mapKeyDiff(relatedEntities, origRelatedEntities) {
		relatedBox, err := p.knownBoxes.getRaw(relatedEntity, relatedMeta)
		if err != nil {
			return err
		}

		if err := p.processBox(relatedBox); err != nil {
			return err
		}

		linkTableInsertAction := NewInsertAction(nil, nil)
		linkTableInsertAction.setTableName(relation.JoinTable)
		linkTableInsertAction.setFields(
			ActionField(relation.JoinColumn, createIDPromise(box)),
			ActionField(relation.ReferenceColumn, createIDPromise(relatedBox)),
		)

		if !relatedBox.action.hasChild(linkTableInsertAction) && !box.action.hasChild(linkTableInsertAction) {
			relatedBox.action.addChild(linkTableInsertAction)
			box.action.addChild(linkTableInsertAction)
		}
	}

	pk, err := box.ExtractPk()
	if err != nil {
		return err
	}

	for _, origRelatedEntity := range mapKeyDiff(origRelatedEntities, relatedEntities) {
		relPk, err := relatedMeta.ExtractPkValue(origRelatedEntity)
		if err != nil {
			return err
		}

		delAction := NewDeleteAction(map[string]interface{}{
			relation.ReferenceColumn: relPk,
			relation.JoinColumn:      pk,
		})
		delAction.setTableName(relation.JoinTable)

		box.action.addChild(delAction)
	}

	return nil
}

func extractSimpleFields(box *persistBox) ([]*actionField, error) {
	fields := make([]*actionField, 0, len(box.Meta.Fields))
	for _, field := range box.Meta.Fields {
		if reflect.IsFieldEquals(box.Entity, box.original, field.Name) {
			continue
		}

		val, err := reflect.ExtractStructField(box.Entity, field.Name)
		if err != nil {
			return nil, err
		}
		fields = append(fields, ActionField(field.DbAlias, val))
	}

	return fields, nil
}

func (p *PersistGraph) filterRoots() []CompositeAction {
	actions := p.knownBoxes.flattActions()

	var result []CompositeAction

	for _, agg := range actions {
		if agg.subscriptionsCount() == 0 {
			result = append(result, agg)
		}
	}

	// if all graph nodes in cycle return nodes with lowest sub count
	if len(result) == 0 && len(actions) != 0 {
		var minSubscriptionsCount = math.MaxInt64

		for _, agg := range actions {
			subscriptionCount := agg.subscriptionsCount()

			if minSubscriptionsCount > subscriptionCount {
				result = result[:0]
				minSubscriptionsCount = subscriptionCount
			}

			if subscriptionCount == minSubscriptionsCount {
				result = append(result, agg)
			}
		}
	}

	return result
}

//promise using for return some fields of entity that have not initialized yet, but will be in future.
type promise struct {
	executable func() (interface{}, error)
	box        *persistBox
	field      string
}

func (p *promise) unwrap() (interface{}, error) {
	return p.executable()
}

func (p *promise) equal(e equaler) bool {
	p2, isPromise := e.(*promise)
	if isPromise {
		return p2.box == p.box && p2.field == p.field
	}
	return false
}

func createIDPromise(box *persistBox) *promise {
	return &promise{
		executable: func() (interface{}, error) {
			return box.ExtractPk()
		},
		box:   box,
		field: box.Meta.Pk.Field.Name,
	}
}

func doSplit(from, to CompositeAction, source *persistBox, col string, val interface{}) {
	splitAction := makeUpdateAction(source)
	splitAction.setFields(ActionField(col, val))

	if from.hasChild(splitAction) || to.hasChild(splitAction) {
		return
	}

	from.addChild(splitAction)
	to.addChild(splitAction)
}
