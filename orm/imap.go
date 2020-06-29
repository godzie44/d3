package orm

import (
	"errors"
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/query"
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

func (im *identityMap) executePlan(plan *query.FetchPlan) (*entity.Collection, error) {
	im.RLock()
	defer im.RUnlock()

	collection := entity.NewCollection()
	for _, id := range plan.PKs() {
		if e, exists := im.get(plan.EntityName(), id); exists {
			collection.Add(e)
		} else {
			return nil, ErrCantExecutePlan
		}
	}

	return collection, nil
}

func (im *identityMap) putEntities(meta *entity.MetaInfo, collection *entity.Collection) {
	iter := collection.MakeIter()

	for iter.Next() {
		pkVal, err := meta.ExtractPkValue(iter.Value())
		if err != nil {
			continue
		}

		im.Lock()
		im.add(meta.EntityName, pkVal, iter.Value())
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
