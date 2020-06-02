package entity

import (
	"fmt"
	"sync"
)

type MetaRegistry struct {
	metaMap map[Name]*MetaInfo

	mutex sync.RWMutex
}

func NewMetaRegistry() *MetaRegistry {
	return &MetaRegistry{
		metaMap: make(map[Name]*MetaInfo),
	}
}

type UserMapping struct {
	Entity    interface{}
	TableName string
}

type promise func() error

func (r *MetaRegistry) makeDepInstaller(meta *MetaInfo, depName Name) promise {
	return func() error {
		if _, exists := r.metaMap[depName]; !exists {
			return fmt.Errorf("found unregister entity: %s", depName)
		}
		meta.RelatedMeta[depName] = r.metaMap[depName]
		return nil
	}
}

func (r *MetaRegistry) Add(mappings ...UserMapping) error {
	var dependencyInstallers []promise

	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, mapping := range mappings {
		meta, err := CreateMeta(mapping)
		if err != nil {
			return err
		}
		r.metaMap[meta.EntityName] = meta

		for entityName := range meta.DependencyEntities() {
			dependencyInstallers = append(dependencyInstallers, r.makeDepInstaller(meta, entityName))
		}
	}

	for _, installer := range dependencyInstallers {
		if err := installer(); err != nil {
			return err
		}
	}

	return nil
}

func (r *MetaRegistry) GetMeta(entity interface{}) (MetaInfo, error) {
	return r.GetMetaByName(NameFromEntity(entity))
}

func (r *MetaRegistry) GetMetaByName(entityName Name) (MetaInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if meta, exists := r.metaMap[entityName]; exists {
		return *meta, nil
	}

	return MetaInfo{}, fmt.Errorf("unregister entity: %s", entityName)
}

func (r *MetaRegistry) ForEach(f func(meta *MetaInfo) error) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, meta := range r.metaMap {
		err := f(meta)
		if err != nil {
			return err
		}
	}

	return nil
}
