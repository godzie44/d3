package orm

import (
	"fmt"
	d3entity "github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/query"
)

type Extractor func() *d3entity.Collection

func (s *session) makeOneToOneExtractor(id interface{}, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() *d3entity.Collection {
		entities, err := s.execute(
			query.New().ForEntity(relatedMeta).Where(relatedMeta.Pk.FullDbAlias(), "=", id), relatedMeta,
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
			query.New().ForEntity(relatedMeta).Where(relatedMeta.FullColumnAlias(relation.JoinColumn), "=", joinId), relatedMeta,
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
			query.New().
				ForEntity(relatedMeta).
				Join(query.JoinInner, rel.JoinTable, fmt.Sprintf("%s.%s=%s", rel.JoinTable, rel.ReferenceColumn, relatedMeta.Pk.FullDbAlias())).
				Where(fmt.Sprintf("%s.%s", rel.JoinTable, rel.JoinColumn), "=", id),
			relatedMeta,
		)
		if err != nil {
			return nil
		}

		return entities
	}
}
