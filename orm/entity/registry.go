package entity

import (
	d3reflect "d3/reflect"
	"fmt"
	"reflect"
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

func (r *MetaRegistry) Add(entities ...interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, e := range entities {
		meta, err := CreateMeta(e)
		if err != nil {
			return err
		}
		r.metaMap[meta.EntityName] = meta
	}

	return nil
}

func (r *MetaRegistry) GetMeta(entity interface{}) (MetaInfo, error) {
	key := Name(d3reflect.FullName(reflect.TypeOf(entity)))

	return r.GetMetaByName(key)
}

func (r *MetaRegistry) GetMetaByName(entityName Name) (MetaInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if meta, exists := r.metaMap[entityName]; exists {
		if err := r.enrichWithRelations(meta); err != nil {
			return MetaInfo{}, err
		}
		return *meta, nil
	}

	return MetaInfo{}, fmt.Errorf("unregister entity: %s", entityName)
}

func (r *MetaRegistry) enrichWithRelations(meta *MetaInfo) error {
	if !meta.dependenciesIsSet() {
		for entityName := range meta.DependencyEntities() {
			if _, exists := r.metaMap[entityName]; !exists {
				return fmt.Errorf("unregister entity: %s", entityName)
			}
			meta.RelatedMeta[entityName] = r.metaMap[entityName]
		}
		for key := range meta.RelatedMeta {
			if err := r.enrichWithRelations(meta.RelatedMeta[key]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *MetaRegistry) ForEach(f func(meta *MetaInfo)) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, meta := range r.metaMap {
		f(meta)
	}
}
