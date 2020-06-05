package orm

import (
	d3entity "d3/orm/entity"
	"d3/orm/query"
	"fmt"
	"reflect"
)

type RawDataMapper func(data interface{}, into reflect.Kind) interface{}

type Hydrator struct {
	session            *Session
	meta               *d3entity.MetaInfo
	afterHydrateEntity func(b *d3entity.Box)
	rawMapper          RawDataMapper
}

func (h *Hydrator) Hydrate(fetchedData []map[string]interface{}, plan *query.FetchPlan) (*d3entity.Collection, error) {
	groupByEntityData := make(map[interface{}][]map[string]interface{})

	for _, rowData := range fetchedData {
		pkVal := rowData[h.meta.Pk.FullDbAlias()]
		groupByEntityData[pkVal] = append(groupByEntityData[pkVal], rowData)
	}

	collection := d3entity.NewCollection()

	for _, entityData := range groupByEntityData {
		newEntity := h.meta.Tools.NewInstance()
		err := h.hydrateOne(newEntity, entityData, plan)
		if err != nil {
			return nil, err
		}

		h.afterHydrateEntity(d3entity.NewBox(newEntity, h.meta))

		collection.Add(newEntity)
	}

	return collection, nil
}

func (h *Hydrator) hydrateOne(entity interface{}, entityData []map[string]interface{}, plan *query.FetchPlan) error {
	for _, field := range h.meta.Fields {
		fieldValue, exists := entityData[0][field.FullDbAlias]
		if !exists {
			continue
		}

		if err := h.meta.Tools.SetFieldVal(entity, field.Name, h.rawMapper(fieldValue, field.AssociatedType.Kind())); err != nil {
			return err
		}
	}

	for _, rel := range h.meta.Relations {
		var fieldValue interface{}
		var err error
		if plan.CanFetchRelation(rel) {
			if entityData[0][h.meta.Pk.FullDbAlias()] == nil {
				fieldValue = nil
			} else {
				fieldValue, err = h.fetchRelation(rel, entityData, plan)
			}
		} else {
			fieldValue, err = h.createRelation(entity, rel, entityData[0])
		}
		if err != nil {
			return err
		}

		if err = h.meta.Tools.SetFieldVal(entity, rel.Field().Name, fieldValue); err != nil {
			return err
		}
	}

	return nil
}

func (h *Hydrator) fetchRelation(relation d3entity.Relation, entityData []map[string]interface{}, plan *query.FetchPlan) (interface{}, error) {
	relationMeta := h.meta.RelatedMeta[relation.RelatedWith()]

	relationHydrator := &Hydrator{
		session:            h.session,
		meta:               relationMeta,
		afterHydrateEntity: h.afterHydrateEntity,
		rawMapper:          h.rawMapper,
	}

	switch relation.(type) {
	case *d3entity.OneToOne:
		relationPkVal := entityData[0][relationMeta.Pk.FullDbAlias()]

		var entity interface{}
		if relationPkVal == nil {
			return d3entity.NewCell(nil), nil
		}

		entity = relationMeta.Tools.NewInstance()
		err := relationHydrator.hydrateOne(entity, entityData, plan.GetChildPlan(relation))
		if err != nil {
			return nil, fmt.Errorf("hydration: %w", err)
		}

		h.afterHydrateEntity(d3entity.NewBox(entity, relationMeta))

		return d3entity.NewCell(entity), nil
	case *d3entity.OneToMany, *d3entity.ManyToMany:
		var entities = d3entity.NewCollection()

		groupByEntity := make(map[interface{}][]map[string]interface{})

		for _, entityData := range entityData {
			pkVal := entityData[relationMeta.Pk.FullDbAlias()]
			if pkVal == nil {
				continue
			}

			groupByEntity[pkVal] = append(groupByEntity[pkVal], entityData)
		}

		for _, data := range groupByEntity {
			entity := relationMeta.Tools.NewInstance()
			err := relationHydrator.hydrateOne(entity, data, plan.GetChildPlan(relation))
			if err != nil {
				return nil, fmt.Errorf("hydration: %w", err)
			}
			entities.Add(entity)

			h.afterHydrateEntity(d3entity.NewBox(entity, relationMeta))
		}

		return entities, nil
	}

	return nil, nil
}

func (h *Hydrator) createRelation(entity interface{}, relation d3entity.Relation, entityData map[string]interface{}) (interface{}, error) {
	switch rel := relation.(type) {
	case *d3entity.OneToOne:
		relatedId, exists := entityData[h.meta.FullColumnAlias(rel.JoinColumn)]
		if !exists {
			return nil, fmt.Errorf("hydration: realated relation not exists")
		}

		if relatedId == nil {
			return d3entity.NewCell(nil), nil
		}

		extractor := h.session.makeOneToOneExtractor(relatedId, h.meta.RelatedMeta[rel.RelatedWith()])

		switch rel.Type() {
		case d3entity.Lazy:
			lazy := d3entity.NewLazyWrappedEntity(extractor, func(cell *d3entity.Cell) {
				h.session.uow.updateFieldOfOriginal(d3entity.NewBox(entity, h.meta), relation.Field().Name, cell)
			})

			return d3entity.NewCellFromWrapper(lazy), nil
		case d3entity.Eager:
			collection := extractor()
			if collection.Empty() {
				return d3entity.NewCell(nil), nil
			}
			return d3entity.NewCell(collection.Get(0)), nil
		}
	case *d3entity.OneToMany, *d3entity.ManyToMany:
		relatedId, exists := entityData[h.meta.Pk.FullDbAlias()]
		if !exists {
			return nil, fmt.Errorf("hydration: owner pk not exists")
		}

		var extractor Extractor
		switch rel := rel.(type) {
		case *d3entity.OneToMany:
			extractor = h.session.makeOneToManyExtractor(relatedId, rel, h.meta.RelatedMeta[rel.RelatedWith()])
		case *d3entity.ManyToMany:
			extractor = h.session.makeManyToManyExtractor(relatedId, rel, h.meta.RelatedMeta[rel.RelatedWith()])
		}

		switch rel.Type() {
		case d3entity.Lazy:
			lazyCol := d3entity.NewLazyCollection(extractor, func(c *d3entity.Collection) {
				h.session.uow.updateFieldOfOriginal(d3entity.NewBox(entity, h.meta), relation.Field().Name, c)
			})

			return d3entity.NewCollectionFromCollectionner(lazyCol), nil
		case d3entity.Eager:
			return extractor(), nil
		}
	}

	return nil, fmt.Errorf("hydration: unsupported relation type")
}
