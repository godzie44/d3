package orm

import (
	stEntity "d3/orm/entity"
	"fmt"
)

type Orm struct {
	PgDb         StorageAdapter
	metaRegistry *stEntity.MetaRegistry
}

func NewOrm(adapter StorageAdapter) *Orm {
	return &Orm{
		PgDb:         adapter,
		metaRegistry: stEntity.NewMetaRegistry(),
	}
}

func (o *Orm) Register(entities ...interface{}) error {
	err := o.metaRegistry.Add(entities...)
	if err != nil {
		return err
	}

	dependencies := make(map[stEntity.Name]struct{})
	o.metaRegistry.ForEach(func(meta *stEntity.MetaInfo) {
		for _, entityName := range meta.DependencyEntities() {
			dependencies[entityName] = struct{}{}
		}
	})

	for dep, _ := range dependencies {
		_, err := o.metaRegistry.GetMetaByName(dep)
		if err != nil {
			panic(fmt.Errorf("found unregister entity: %s", dep))
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
		sourceEntity: entity,
		session:      session,
		entityMeta:   entityMeta,
	}, nil
}
