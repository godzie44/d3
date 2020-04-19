package orm

import (
	d3entity "d3/orm/entity"
	"d3/orm/query"
	d3reflect "d3/reflect"
	"fmt"
)

type Extractor func() interface{}

func (s *Session) createOneToOneExtractor(id interface{}, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() interface{} {
		entities, err := s.execute(
			query.NewQuery(relatedMeta).AndWhere(relatedMeta.Pk.FullDbAlias()+"=?", id),
		)

		if err != nil {
			return nil
		}

		entity, err := d3reflect.GetFirstElementFromSlice(entities)
		if err != nil {
			return nil
		}

		return entity
	}
}

func (s *Session) createOneToManyExtractor(joinId interface{}, relation *d3entity.OneToMany, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() interface{} {
		entities, err := s.execute(
			query.NewQuery(relatedMeta).AndWhere(relatedMeta.FullColumnAlias(relation.JoinColumn)+"=?", joinId),
		)
		if err != nil {
			return nil
		}

		return entities
	}
}

func (s *Session) createManyToManyExtractor(id interface{}, rel *d3entity.ManyToMany, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() interface{} {
		entities, err := s.execute(
			query.
				NewQuery(relatedMeta).
				Join(query.JoinInner, rel.JoinTable, fmt.Sprintf("%s.%s=%s", rel.JoinTable, rel.ReferenceColumn, relatedMeta.Pk.FullDbAlias())).
				AndWhere(fmt.Sprintf("%s.%s=?", rel.JoinTable, rel.JoinColumn), id),
		)
		if err != nil {
			return nil
		}

		return entities
	}
}
