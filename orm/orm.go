package orm

import (
	"context"
	"fmt"
	d3Entity "github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/schema"
)

type ctxKey string

const sessionKey ctxKey = "d3_session"

// Orm - d3 orm instance.
type Orm struct {
	storage      Driver
	metaRegistry *d3Entity.MetaRegistry
}

// New - create an instance of d3 orm.
//
// driver - d3 wrapper on database driver. Find it in adapter package.
func New(driver Driver) *Orm {
	return &Orm{
		storage:      driver,
		metaRegistry: d3Entity.NewMetaRegistry(),
	}
}

// Register - register entities in d3 orm.
// Note that entity must be structure and must implement D3Entity interface (it's implement it after use code generation tool).
// Besides if you register entity with dependencies (for example: one to one relation) you must registered depended entities too, in the same call.
func (o *Orm) Register(entities ...interface{}) error {
	return o.metaRegistry.Add(entities...)
}

// CtxWithSession append new session instance to context.
func (o *Orm) CtxWithSession(ctx context.Context) context.Context {
	return context.WithValue(ctx, sessionKey, o.MakeSession())
}

// Session extract session from context, return nil if session not found.
func Session(ctx context.Context) *session {
	if sess, ok := ctx.Value(sessionKey).(*session); ok {
		return sess
	}
	return nil
}

// MakeSession - create new instance of session.
func (o *Orm) MakeSession() *session {
	return newSession(o.storage, newUOW(o.storage))
}

// MakeRepository - create new repository for entity.
//
// entity - entity to store in repository.
//
//deprecated
//func (o *Orm) MakeRepository(entity interface{}) (*Repository, error) {
//	entityMeta, err := o.metaRegistry.GetMeta(entity)
//	if err != nil {
//		return nil, fmt.Errorf("repository: %w", err)
//	}
//
//	return &Repository{
//		entityMeta: entityMeta,
//	}, nil
//}

// MakeRepository - create new repository for entity T.
func MakeRepository[T d3Entity.D3Entity](o *Orm) (*Repository[T], error) {
	var entity T
	entityMeta, err := o.metaRegistry.GetMeta(entity)
	if err != nil {
		return nil, fmt.Errorf("repository: %w", err)
	}

	return &Repository[T]{
		entityMeta: entityMeta,
	}, nil
}

// GenerateSchema - create sql DDL for persist all registered entities in database.
// May return error if driver nonsupport schema generation.
func (o *Orm) GenerateSchema() (string, error) {
	generator, adapterCanGenerateSchema := o.storage.(schema.StorageSchemaGenerator)
	if !adapterCanGenerateSchema {
		return "", fmt.Errorf("adapter unsupport schema generation")
	}
	return schema.NewBuilder(generator).Build(o.metaRegistry)
}
