package orm

import (
	d3Entity "d3/orm/entity"
	"d3/orm/schema"
	"fmt"
)

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

type Mapping struct {
	Table  string
	Entity interface{}
}

func NewMapping(tableName string, entity interface{}) Mapping {
	return Mapping{
		Table:  tableName,
		Entity: entity,
	}
}

func (o *Orm) Register(mappings ...Mapping) error {
	ms := make([]d3Entity.UserMapping, len(mappings))
	for i := range mappings {
		ms[i] = d3Entity.UserMapping{
			Entity:    mappings[i].Entity,
			TableName: mappings[i].Table,
		}
	}

	err := o.metaRegistry.Add(ms...)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orm) MakeSession() *Session {
	return NewSession(o.PgDb, NewUOW(o.PgDb), o.metaRegistry)
}

func (o *Orm) GenerateSchema() (string, error) {
	generator, adapterCanGenerateSchema := o.PgDb.(schema.StorageSchemaGenerator)
	if !adapterCanGenerateSchema {
		return "", fmt.Errorf("adapter unsupport schema generation")
	}
	return schema.NewBuilder(generator).Build(o.metaRegistry)
}
