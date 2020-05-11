package orm

import (
	d3Entity "d3/orm/entity"
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
