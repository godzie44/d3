package orm

import (
	"d3/mapper"
	d3entity "d3/orm/entity"
	"d3/orm/query"
	d3reflect "d3/reflect"
	"fmt"
	"reflect"
)

type Hydrator struct {
	session *Session
	meta    *d3entity.MetaInfo
}

func (h *Hydrator) Hydrate(fetchedData []map[string]interface{}, plan *query.FetchPlan) (interface{}, error) {
	groupByEntityData := make(map[interface{}][]map[string]interface{})

	for _, rowData := range fetchedData {
		pkVal := rowData[h.meta.Pk.FullDbAlias()]
		groupByEntityData[pkVal] = append(groupByEntityData[pkVal], rowData)
	}

	entityType := reflect.TypeOf(h.meta.Tpl)

	modelSlice := d3reflect.CreateSliceOfStructPtrs(entityType, len(groupByEntityData))
	sliceVal := reflect.ValueOf(modelSlice)

	var lastInsertedNum int
	for _, entityData := range groupByEntityData {
		newEntity := reflect.New(entityType.Elem())

		err := h.hydrateOne(newEntity.Interface(), entityData, plan)
		if err != nil {
			return nil, err
		}

		sliceVal.Index(lastInsertedNum).Set(newEntity)
		lastInsertedNum++
	}

	return modelSlice, nil
}

func (h *Hydrator) hydrateOne(model interface{}, entityData []map[string]interface{}, plan *query.FetchPlan) error {
	modelReflectVal := reflect.ValueOf(model).Elem()
	modelType := modelReflectVal.Type()
	for i := 0; i < modelReflectVal.NumField(); i++ {
		f := modelReflectVal.Field(i)
		if err := d3reflect.ValidateField(&f); err != nil {
			continue
		}

		var fieldValue interface{}

		if fieldInfo, exists := h.meta.Fields[modelType.Field(i).Name]; exists {
			fieldValue, exists = entityData[0][fieldInfo.FullDbAlias]
			if !exists {
				continue
			}
		} else if relation, exists := h.meta.Relations[modelType.Field(i).Name]; exists {
			var err error
			if plan.CanFetchRelation(relation) {
				if entityData[0][h.meta.Pk.FullDbAlias()] == nil {
					fieldValue = nil
				} else {
					fieldValue, err = h.fetchRelation(relation, entityData, plan)
				}
			} else {
				fieldValue, err = h.createRelation(relation, entityData[0])
			}

			if err != nil {
				return err
			}
		}

		d3reflect.SetField(&f, fieldValue)
	}

	return nil
}

func (h *Hydrator) fetchRelation(relation d3entity.Relation, entityData []map[string]interface{}, plan *query.FetchPlan) (interface{}, error) {
	relationMeta := h.meta.RelatedMeta[relation.RelatedWith()]

	relationHydrator := &Hydrator{
		session: h.session,
		meta:    relationMeta,
	}

	switch relation.(type) {
	case *d3entity.OneToOne:
		relationPkVal := entityData[0][relationMeta.Pk.FullDbAlias()]

		var entity interface{}
		if relationPkVal == nil {
			return d3entity.NewWrapEntity(nil), nil
		}

		entity = d3reflect.CreateEmptyEntity(relationMeta.Tpl)
		err := relationHydrator.hydrateOne(entity, entityData, plan.GetChildPlan(relation))
		if err != nil {
			return nil, fmt.Errorf("hydration: %w", err)
		}

		return d3entity.NewWrapEntity(entity), nil
	case *d3entity.OneToMany, *d3entity.ManyToMany:
		var entities []interface{}

		groupByEntity := make(map[interface{}][]map[string]interface{})

		for _, entityData := range entityData {
			pkVal := entityData[relationMeta.Pk.FullDbAlias()]
			if pkVal == nil {
				continue
			}

			groupByEntity[pkVal] = append(groupByEntity[pkVal], entityData)
		}

		for _, data := range groupByEntity {
			entity := d3reflect.CreateEmptyEntity(relationMeta.Tpl)
			err := relationHydrator.hydrateOne(entity, data, plan.GetChildPlan(relation))
			if err != nil {
				return nil, fmt.Errorf("hydration: %w", err)
			}
			entities = append(entities, entity)
		}

		return mapper.NewCollection(entities), nil
	}

	return nil, nil
}

func (h *Hydrator) createRelation(relation d3entity.Relation, entityData map[string]interface{}) (interface{}, error) {
	switch rel := relation.(type) {
	case *d3entity.OneToOne:
		relatedId, exists := entityData[h.meta.FullColumnAlias(rel.JoinColumn)]
		if !exists {
			return nil, fmt.Errorf("hydration: realated relation not exists")
		}

		if relatedId == nil {
			return d3entity.NewWrapEntity(nil), nil
		}

		extractor := h.session.createOneToOneExtractor(relatedId, h.meta.RelatedMeta[rel.RelatedWith()])

		if rel.IsLazy() {
			return d3entity.NewLazyWrappedEntity(extractor), nil
		}

		if rel.IsEager() {
			return d3entity.NewWrapEntity(extractor()), nil
		}
	case *d3entity.OneToMany, *d3entity.ManyToMany:
		relatedId, exists := entityData[h.meta.Pk.FullDbAlias()]
		if !exists {
			return nil, fmt.Errorf("hydration: owner pk not exists")
		}

		var extractor func() interface{}
		switch rel := rel.(type) {
		case *d3entity.OneToMany:
			extractor = h.session.createOneToManyExtractor(relatedId, rel, h.meta.RelatedMeta[rel.RelatedWith()])
		case *d3entity.ManyToMany:
			extractor = h.session.createManyToManyExtractor(relatedId, rel, h.meta.RelatedMeta[rel.RelatedWith()])
		default:
			panic("unreachable statement")
		}

		if rel.IsLazy() {
			return mapper.NewLazyCollection(extractor), nil
		}

		if rel.IsEager() {
			return mapper.NewCollection(d3reflect.BreakUpSlice(extractor())), nil
		}
	}

	return nil, fmt.Errorf("hydration: unsupported relation type")
}
