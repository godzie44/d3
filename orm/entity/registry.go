package entity

import (
	d3reflect "d3/reflect"
	"errors"
	"reflect"
	"sync"
)

var ErrMetaNotFound = errors.New("meta info not found")

type MetaRegistry struct {
	metaMap map[Name]*MetaInfo

	sync.RWMutex
}

func NewMetaRegistry() *MetaRegistry {
	return &MetaRegistry{
		metaMap: make(map[Name]*MetaInfo),
	}
}

func (r *MetaRegistry) Add(entities ...interface{}) error {
	r.Lock()
	defer r.Unlock()
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
	r.RLock()
	defer r.RUnlock()
	if meta, exists := r.metaMap[entityName]; exists {
		r.enrichWithRelations(meta)
		return *meta, nil
	}

	return MetaInfo{}, ErrMetaNotFound
}

func (r *MetaRegistry) enrichWithRelations(meta *MetaInfo) {
	if !meta.dependenciesIsSet() {
		for _, entityName := range meta.DependencyEntities() {
			meta.RelatedMeta[entityName] = r.metaMap[entityName]
		}
		for key := range meta.RelatedMeta {
			r.enrichWithRelations(meta.RelatedMeta[key])
		}
	}
}

func (r *MetaRegistry) ForEach(f func(meta *MetaInfo)) {
	r.Lock()
	defer r.Unlock()

	for _, meta := range r.metaMap {
		f(meta)
	}
}