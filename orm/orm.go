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

func (o *Orm) MakeSession() *Session {
	return NewSession(o.PgDb, NewUOW(o.PgDb), o.metaRegistry)
}
