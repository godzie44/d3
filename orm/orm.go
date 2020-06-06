package orm

import (
	"context"
	d3Entity "d3/orm/entity"
	"d3/orm/schema"
	"fmt"
)

const sessionKey = "d3_session"

type Orm struct {
	PgDb         Storage
	metaRegistry *d3Entity.MetaRegistry
}

func NewOrm(adapter Storage) *Orm {
	return &Orm{
		PgDb:         adapter,
		metaRegistry: d3Entity.NewMetaRegistry(),
	}
}

func (o *Orm) Register(entities ...interface{}) error {
	err := o.metaRegistry.Add(entities...)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orm) CtxWithSession(ctx context.Context) context.Context {
	return context.WithValue(ctx, sessionKey, o.MakeSession())
}

func Session(ctx context.Context) *session {
	if sess, ok := ctx.Value(sessionKey).(*session); ok {
		return sess
	}
	return nil
}

func (o *Orm) MakeSession() *session {
	return newSession(o.PgDb, NewUOW(o.PgDb))
}

func (o *Orm) MakeRepository(entity interface{}) (*Repository, error) {
	entityMeta, err := o.metaRegistry.GetMeta(entity)
	if err != nil {
		return nil, fmt.Errorf("repository: %w", err)
	}

	return &Repository{
		entityMeta: entityMeta,
	}, nil
}

func (o *Orm) GenerateSchema() (string, error) {
	generator, adapterCanGenerateSchema := o.PgDb.(schema.StorageSchemaGenerator)
	if !adapterCanGenerateSchema {
		return "", fmt.Errorf("adapter unsupport schema generation")
	}
	return schema.NewBuilder(generator).Build(o.metaRegistry)
}
