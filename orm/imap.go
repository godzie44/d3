package orm

import (
	"d3/orm/entity"
	"d3/orm/query"
	d3reflect "d3/reflect"
	"errors"
	"sync"
)

var ErrCantExecutePlan = errors.New("cant execute plan")

type identityMap struct {
	data map[entity.Name]map[interface{}]interface{}

	sync.RWMutex
}

func newIdentityMap() *identityMap {
	return &identityMap{data: make(map[entity.Name]map[interface{}]interface{})}
}

//canApply check that only id in where clause.
// query with joins not allowed, cause we don't know does entities in identityMap has related entities in memory.
func (im *identityMap) canApply(plan *query.FetchPlan) bool {
	return !plan.HasJoins() && len(plan.PKs()) != 0
}

func (im *identityMap) executePlan(plan *query.FetchPlan) (interface{}, error) {
	im.RLock()
	defer im.RUnlock()

	var entities []interface{}
	for _, id := range plan.PKs() {
		if e, exists := im.get(plan.Query().OwnerMeta().EntityName, id); exists {
			entities = append(entities, e)
		} else {
			return nil, ErrCantExecutePlan
		}
	}

	return entities, nil
}

func (im *identityMap) putEntities(meta *entity.MetaInfo, entities interface{}) {
	for _, el := range d3reflect.BreakUpSlice(entities) {
		pkVal, err := meta.ExtractPkValue(el)
		if err != nil {
			continue
		}

		im.Lock()
		im.add(meta.EntityName, pkVal, el)
		im.Unlock()
	}
}

func (im *identityMap) add(name entity.Name, key interface{}, e interface{}) {
	if _, exists := im.data[name]; !exists {
		im.data[name] = make(map[interface{}]interface{})
	}

	im.data[name][normalizeKey(key)] = e
}

func (im *identityMap) get(name entity.Name, key interface{}) (interface{}, bool) {
	e, exists := im.data[name][normalizeKey(key)]

	return e, exists
}

func normalizeKey(key interface{}) interface{} {
	switch k := key.(type) {
	case int:
		return int64(k)
	case int32:
		return int64(k)
	case int64:
		return k
	default:
		return k
	}
}
