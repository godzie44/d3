package orm

import (
	d3entity "d3/orm/entity"
	"d3/orm/query"
	"fmt"
)

type Extractor func() *d3entity.Collection

func (s *session) makeOneToOneExtractor(id interface{}, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() *d3entity.Collection {
		entities, err := s.execute(
			query.NewQuery(relatedMeta).AndWhere(relatedMeta.Pk.FullDbAlias()+"=?", id),
		)
		if err != nil {
			return nil
		}

		return entities
	}
}

func (s *session) makeOneToManyExtractor(joinId interface{}, relation *d3entity.OneToMany, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() *d3entity.Collection {
		entities, err := s.execute(
			query.NewQuery(relatedMeta).AndWhere(relatedMeta.FullColumnAlias(relation.JoinColumn)+"=?", joinId),
		)
		if err != nil {
			return nil
		}

		return entities
	}
}

func (s *session) makeManyToManyExtractor(id interface{}, rel *d3entity.ManyToMany, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() *d3entity.Collection {
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
