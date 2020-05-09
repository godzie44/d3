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
	entities, err := r.session.execute(q)
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
	return r.session.execute(q)
}

func (r *Repository) Persists(entities ...interface{}) error {
	for _, entity := range entities {
		if err := r.session.uow.registerNew(d3entity.NewBox(entity, &r.entityMeta)); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) CreateQuery() *query.Query {
	return query.NewQuery(&r.entityMeta)
}

func (r *Repository) Delete(entities ...interface{}) error {
	for _, e := range entities {
		if err := r.session.uow.registerRemove(d3entity.NewBox(e, &r.entityMeta)); err != nil {
			return err
		}
	}
	return nil
}
