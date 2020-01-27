package mapper

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEagerCollectionCount(t *testing.T) {
	collection := NewEagerCollection([]interface{}{1, 2})

	assert.Equal(t, 2, collection.Count())
}

func TestEagerCollectionEmpty(t *testing.T) {
	collection := NewEagerCollection([]interface{}{})

	assert.True(t, collection.Empty())
}

func TestEagerCollectionAdd(t *testing.T) {
	collection := NewEagerCollection([]interface{}{})

	collection.Add(1)
	collection.Add(2)

	assert.Equal(t, 2, collection.Count())
}

func TestEagerCollectionGet(t *testing.T) {
	collection := NewEagerCollection([]interface{}{1, 2})

	assert.Equal(t, 1, collection.Get(0))
}

func TestEagerCollectionToSlice(t *testing.T) {
	collection := NewEagerCollection([]interface{}{1, 2})
	collection.Add(3)

	assert.Equal(t, []interface{}{1, 2, 3}, collection.ToSlice())
}

func TestLazyCollectionCount(t *testing.T) {
	collection := NewLazyCollection(func() interface{} {
		return []int{1, 2}
	})

	assert.Equal(t, 2, collection.Count())
}

func TestLazyCollectionEmpty(t *testing.T) {
	collection := NewLazyCollection(func() interface{} {
		return []int{}
	})

	assert.True(t, collection.Empty())
}

func TestLazyCollectionAdd(t *testing.T) {
	collection := NewLazyCollection(func() interface{} {
		return []int{}
	})

	collection.Add(1)
	collection.Add(2)

	assert.Equal(t, 2, collection.Count())
}

func TestLazyCollectionGet(t *testing.T) {
	collection := NewLazyCollection(func() interface{} {
		return []int{1, 2}
	})

	assert.Equal(t, 1, collection.Get(0))
}

func TestLazyCollectionToSlice(t *testing.T) {
	collection := NewLazyCollection(func() interface{} {
		return []int{1, 2}
	})
	collection.Add(3)

	assert.Equal(t, []interface{}{1, 2, 3}, collection.ToSlice())
}
