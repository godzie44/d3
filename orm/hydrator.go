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
		pkVal := rowData[h.meta.FullFieldAlias(h.meta.PkField())]
		groupByEntityData[pkVal] = append(groupByEntityData[pkVal], rowData)
	}

	entityType := reflect.TypeOf(h.meta.Tpl)

	modelSlice := d3reflect.CreateSliceOfEntities(entityType, len(groupByEntityData))
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
		if !f.IsValid() || !f.CanSet() {
			continue
		}

		fieldMeta, exists := h.meta.Fields[modelType.Field(i).Name]
		if !exists {
			continue
		}

		var fieldValue interface{}
		var err error

		if fieldMeta.IsRelation() {
			if plan.CanFetchRelation(fieldMeta.Relation) {
				if entityData[0][h.meta.FullFieldAlias(h.meta.PkField())] == nil {
					fieldValue = nil
				} else {
					fieldValue, err = h.fetchRelation(fieldMeta.Relation, entityData, plan)
				}
			} else {
				fieldValue, err = h.createLazyRelation(fieldMeta.Relation, entityData[0])
			}

			if err != nil {
				return err
			}
		} else {
			fieldValue, exists = entityData[0][h.meta.FullFieldAlias(fieldMeta)]
			if !exists {
				continue
			}
		}

		f.Set(reflect.ValueOf(fieldValue))
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
		relationPkVal := entityData[0][relationMeta.FullFieldAlias(relationMeta.PkField())]

		var entity interface{}
		if relationPkVal == nil {
			return d3entity.NewEagerEntity(nil), nil
		}

		entity = d3reflect.CreateEmptyEntity(relationMeta.Tpl)
		err := relationHydrator.hydrateOne(entity, entityData, plan.GetChildPlan(relation))
		if err != nil {
			return nil, fmt.Errorf("hydration: %w", err)
		}

		return d3entity.NewEagerEntity(entity), nil
	case *d3entity.OneToMany, *d3entity.ManyToMany:
		var entities []interface{}

		groupByEntity := make(map[interface{}][]map[string]interface{})

		for _, entityData := range entityData {
			pkVal := entityData[relationMeta.FullFieldAlias(relationMeta.PkField())]
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

		return mapper.NewEagerCollection(entities), nil
		//case *d3entity.ManyToMany:

	}

	return nil, nil
}

func (h *Hydrator) createLazyRelation(relation d3entity.Relation, entityData map[string]interface{}) (interface{}, error) {
	switch rel := relation.(type) {
	case *d3entity.OneToOne:
		relatedId, exists := entityData[h.meta.FullColumnAlias(rel.JoinColumn)]
		if !exists {
			return nil, fmt.Errorf("hydration: realated relation not exists")
		}

		if relatedId == nil {
			return d3entity.NewEagerEntity(nil), nil
		}

		extractor := createOneToOneExtractor(h.session, relatedId, h.meta.RelatedMeta[rel.TargetEntity])

		if rel.IsLazy() {
			return d3entity.NewLazyEntity(extractor), nil
		}

		if rel.IsEager() {
			return d3entity.NewEagerEntity(extractor()), nil
		}
	case *d3entity.OneToMany, *d3entity.ManyToMany:
		relatedId, exists := entityData[h.meta.FullColumnAlias(h.meta.PkField().DbAlias)]
		if !exists {
			return nil, fmt.Errorf("hydration: owner pk not exists")
		}

		var extractor func() interface{}
		switch rel := rel.(type) {
		case *d3entity.OneToMany:
			extractor = createOneToManyExtractor(h.session, relatedId, rel, h.meta.RelatedMeta[rel.TargetEntity])
		case *d3entity.ManyToMany:
			extractor = createManyToManyExtractor(h.session, relatedId, rel, h.meta.RelatedMeta[rel.TargetEntity])
		default:
			panic("unreachable statement")
		}

		if rel.IsLazy() {
			return mapper.NewLazyCollection(extractor), nil
		}

		if rel.IsEager() {
			return mapper.NewEagerCollection(d3reflect.BreakUpSlice(extractor())), nil
		}
	}

	return nil, fmt.Errorf("hydration: unsupported relation type")
}
