package orm

import (
	"context"
	"errors"
	d3entity "github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/query"
)

var (
	ErrEntityNotFound = errors.New("entity not found")
	ErrSessionNotSet  = errors.New("session not found in context")
)

type Repository struct {
	entityMeta d3entity.MetaInfo
}

// FindOne - return one entity fetched by query. If entity not found ErrEntityNotFound will returned.
func (r *Repository) FindOne(ctx context.Context, q *query.Query) (interface{}, error) {
	session, err := sessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	coll, err := session.execute(q)
	if err != nil {
		return nil, err
	}

	if coll.Count() == 0 {
		return nil, ErrEntityNotFound
	}

	return coll.Get(0), nil
}

// FindOne - return collection of entities fetched by query.
func (r *Repository) FindAll(ctx context.Context, q *query.Query) (*d3entity.Collection, error) {
	session, err := sessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	return session.execute(q)
}

// Persists - add entities to repository.
func (r *Repository) Persists(ctx context.Context, entities ...interface{}) error {
	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	for _, entity := range entities {
		if err := session.uow.registerNew(d3entity.NewBox(entity, &r.entityMeta)); err != nil {
			return err
		}
	}
	return nil
}

// Delete - delete entities from repository.
func (r *Repository) Delete(ctx context.Context, entities ...interface{}) error {
	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	for _, e := range entities {
		if err := session.uow.registerRemove(d3entity.NewBox(e, &r.entityMeta)); err != nil {
			return err
		}
	}
	return nil
}

func sessionFromCtx(ctx context.Context) (*session, error) {
	s := Session(ctx)
	if s == nil {
		return nil, ErrSessionNotSet
	}

	return s, nil
}

// MakeQuery create query for fetch entity with the same type as the repository.
func (r *Repository) MakeQuery() *query.Query {
	return query.NewQuery(&r.entityMeta)
}
