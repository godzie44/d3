package orm

import (
	"d3"
	"d3/orm/entity"
	"d3/orm/query"
	"d3/reflect"
	"fmt"
)

type Repository struct {
	sourceEntity interface{}
	session      *Session
	uow          *d3.UnitOfWork
	entityMeta   entity.MetaInfo
}

func (r *Repository) FindOne(q *query.Query) (interface{}, error) {
	entities, err := r.session.Execute(q)
	if err != nil {
		return nil, err
	}

	el, err := reflect.GetFirstElementFromSlice(entities)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	return el, nil
}

func (r *Repository) FindAll(q *query.Query) (interface{}, error) {
	return r.session.Execute(q)
}

func (r *Repository) Persists(entity d3.DomainEntity) {
	r.uow.RegisterNew(entity)
}

func (r *Repository) CreateQuery() *query.Query {
	return query.NewQuery(&r.entityMeta)
}
