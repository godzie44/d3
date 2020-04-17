package orm

import (
	d3entity "d3/orm/entity"
	"d3/orm/query"
	d3reflect "d3/reflect"
	"fmt"
)

type Repository struct {
	session    *Session
	entityMeta d3entity.MetaInfo
}

func (r *Repository) FindOne(q *query.Query) (interface{}, error) {
	entities, err := r.session.Execute(q)
	if err != nil {
		return nil, err
	}

	el, err := d3reflect.GetFirstElementFromSlice(entities)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	return el, nil
}

func (r *Repository) FindAll(q *query.Query) (interface{}, error) {
	return r.session.Execute(q)
}

func (r *Repository) Persists(entity interface{}) error {
	return r.session.uow.registerNew(d3entity.NewBox(entity, &r.entityMeta))
}

func (r *Repository) CreateQuery() *query.Query {
	return query.NewQuery(&r.entityMeta)
}
