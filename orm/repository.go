package orm

import (
	d3entity "d3/orm/entity"
	"d3/orm/query"
	"errors"
)

var (
	ErrEntityNotFound = errors.New("entity not found")
)

type Repository struct {
	session    *Session
	entityMeta d3entity.MetaInfo
}

func (r *Repository) FindOne(q *query.Query) (interface{}, error) {
	coll, err := r.session.execute(q)
	if err != nil {
		return nil, err
	}

	if coll.Count() == 0 {
		return nil, ErrEntityNotFound
	}

	return coll.Get(0), nil
}

func (r *Repository) FindAll(q *query.Query) (*d3entity.Collection, error) {
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
