package entity

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

type testEntity struct {
	ID   int `d3:"pk:auto"`
	data string
}

func TestRegistryAdd(t *testing.T) {
	registry := NewMetaRegistry()
	_ = registry.Add((*testEntity)(nil), (*testEntity)(nil))

	assert.Len(t, registry.metaMap, 1)
}

func TestRegistryGet(t *testing.T) {
	registry := NewMetaRegistry()

	_ = registry.Add((*testEntity)(nil))

	meta, _ := registry.GetMeta((*testEntity)(nil))
	assert.NotEmpty(t, meta)

	meta, _ = registry.GetMeta(&testEntity{})
	assert.NotEmpty(t, meta)
}

type testEntity2 struct {
	Id   int `d3:"pk:auto"`
	data string
}

func TestRegistryGetMetaParallel(t *testing.T) {
	registry := NewMetaRegistry()

	_ = registry.Add((*testEntity)(nil), (*testEntity2)(nil))

	var meta1, meta2 MetaInfo
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		meta1, _ = registry.GetMeta((*testEntity)(nil))
	}()

	go func() {
		defer wg.Done()
		meta2, _ = registry.GetMeta((*testEntity2)(nil))
	}()

	wg.Wait()

	assert.NotEmpty(t, meta1)
	assert.NotEmpty(t, meta2)
}
