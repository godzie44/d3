package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEagerCollectionCount(t *testing.T) {
	collection := NewCollection([]interface{}{1, 2})

	assert.Equal(t, 2, collection.Count())
}

func TestEagerCollectionEmpty(t *testing.T) {
	collection := NewCollection([]interface{}{})

	assert.True(t, collection.Empty())
}

func TestEagerCollectionAdd(t *testing.T) {
	collection := NewCollection([]interface{}{})

	collection.Add(1)
	collection.Add(2)

	assert.Equal(t, 2, collection.Count())
}

func TestEagerCollectionGet(t *testing.T) {
	collection := NewCollection([]interface{}{1, 2})

	assert.Equal(t, 1, collection.Get(0))
}

func TestEagerCollectionToSlice(t *testing.T) {
	collection := NewCollection([]interface{}{1, 2})
	collection.Add(3)

	assert.Equal(t, []interface{}{1, 2, 3}, collection.ToSlice())
}

func TestLazyCollectionCount(t *testing.T) {
	collection := NewCollectionFromCollectionner(NewLazyCollection(func() *Collection {
		return NewCollection([]interface{}{1, 2})
	}, func(_ *Collection) {}))

	assert.Equal(t, 2, collection.Count())
}

func TestLazyCollectionEmpty(t *testing.T) {
	collection := NewCollectionFromCollectionner(NewLazyCollection(func() *Collection {
		return NewCollection([]interface{}{})
	}, func(_ *Collection) {}))

	assert.True(t, collection.Empty())
}

func TestLazyCollectionAdd(t *testing.T) {
	collection := NewCollectionFromCollectionner(NewLazyCollection(func() *Collection {
		return NewCollection([]interface{}{})
	}, func(_ *Collection) {}))

	collection.Add(1)
	collection.Add(2)

	assert.Equal(t, 2, collection.Count())
}

func TestLazyCollectionGet(t *testing.T) {
	collection := NewCollectionFromCollectionner(NewLazyCollection(func() *Collection {
		return NewCollection([]interface{}{1, 2})
	}, func(_ *Collection) {}))

	assert.Equal(t, 1, collection.Get(0))
}

func TestLazyCollectionToSlice(t *testing.T) {
	collection := NewCollectionFromCollectionner(NewLazyCollection(func() *Collection {
		return NewCollection([]interface{}{1, 2})
	}, func(_ *Collection) {}))
	collection.Add(3)

	assert.Equal(t, []interface{}{1, 2, 3}, collection.ToSlice())
}
