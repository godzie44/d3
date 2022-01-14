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

type Repository[T d3entity.D3Entity] struct {
	entityMeta d3entity.MetaInfo
}

// FindOne - return one entity fetched by query. If entity not found ErrEntityNotFound will returned.
func (r *Repository[T]) FindOne(ctx context.Context, q *query.Query) (entity T, err error) {
	session, err := sessionFromCtx(ctx)
	if err != nil {
		return entity, err
	}

	coll, err := session.execute(q, &r.entityMeta)
	if err != nil {
		return entity, err
	}

	if coll.Count() == 0 {
		return entity, ErrEntityNotFound
	}

	return coll.Get(0).(T), nil
}

// FindOne - return collection of entities fetched by query.
func (r *Repository[T]) FindAll(ctx context.Context, q *query.Query) (*d3entity.Collection, error) {
	session, err := sessionFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	return session.execute(q, &r.entityMeta)
}

// Persists - add entities to repository.
func (r *Repository[T]) Persists(ctx context.Context, entities ...interface{}) error {
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
func (r *Repository[T]) Delete(ctx context.Context, entities ...interface{}) error {
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

// Select - create query for fetch entity with the same type as the repository.
func (r *Repository[T]) Select() *query.Query {
	return query.New().ForEntity(&r.entityMeta)
}
