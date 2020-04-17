package orm

import (
	d3Entity "d3/orm/entity"
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

func (o *Orm) Register(entities ...interface{}) error {
	err := o.metaRegistry.Add(entities...)
	if err != nil {
		return err
	}

	allDependencies := make(map[d3Entity.Name]struct{})
	o.metaRegistry.ForEach(func(meta *d3Entity.MetaInfo) {
		for entityName := range meta.DependencyEntities() {
			allDependencies[entityName] = struct{}{}
		}
	})

	for dep := range allDependencies {
		_, err := o.metaRegistry.GetMetaByName(dep)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Orm) CreateSession() *Session {
	return NewSession(o.PgDb, NewUOW(o.PgDb), o.metaRegistry)
}

func (o *Orm) CreateRepository(session *Session, entity interface{}) (*Repository, error) {
	entityMeta, err := o.metaRegistry.GetMeta(entity)
	if err != nil {
		return nil, fmt.Errorf("repository: %w", err)
	}

	return &Repository{
		session:    session,
		entityMeta: entityMeta,
	}, nil
}
