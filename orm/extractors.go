package orm

import (
	d3entity "d3/orm/entity"
	"d3/orm/query"
	d3reflect "d3/reflect"
	"fmt"
)

type Extractor func() interface{}

func createOneToOneExtractor(session *Session, id interface{}, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() interface{} {
		entities, err := session.Execute(
			query.NewQuery(relatedMeta).AndWhere(relatedMeta.FullColumnAlias(relatedMeta.PkField().DbAlias)+"=?", id),
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

func createOneToManyExtractor(session *Session, joinId interface{}, relation *d3entity.OneToMany, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() interface{} {
		entities, err := session.Execute(
			query.NewQuery(relatedMeta).AndWhere(relatedMeta.FullColumnAlias(relation.JoinColumn)+"=?", joinId),
		)
		if err != nil {
			return nil
		}

		return entities
	}
}

func createManyToManyExtractor(session *Session, id interface{}, rel *d3entity.ManyToMany, relatedMeta *d3entity.MetaInfo) Extractor {
	return func() interface{} {
		entities, err := session.Execute(
			query.
				NewQuery(relatedMeta).
				Join(query.JoinInner, rel.JoinTable, fmt.Sprintf("%s.%s=%s", rel.JoinTable, rel.ReferenceColumn, relatedMeta.FullFieldAlias(relatedMeta.PkField()))).
				AndWhere(fmt.Sprintf("%s.%s=?", rel.JoinTable, rel.JoinColumn), id),
		)
		if err != nil {
			return nil
		}

		return entities
	}
}
