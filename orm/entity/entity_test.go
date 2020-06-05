package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEagerEntityGet(t *testing.T) {
	entity := NewWrapEntity(1)
	assert.Equal(t, 1, entity.Unwrap())
}

func TestEagerEntitySet(t *testing.T) {
	entity := NewWrapEntity(1)
	entity.wrap(2)
	assert.Equal(t, 2, entity.base.inner)
}

func TestEagerEntityIsNil(t *testing.T) {
	entity := NewWrapEntity(nil)
	assert.True(t, entity.IsNil())
}

func TestLazyEntityGet(t *testing.T) {
	entity := NewLazyWrappedEntity(func() *Collection {
		return NewCollection(1)
	}, func(_ WrappedEntity) {})
	assert.Equal(t, 1, entity.Unwrap())
}

func TestLazyEntitySet(t *testing.T) {
	entity := NewLazyWrappedEntity(func() *Collection {
		return NewCollection(1)
	}, func(_ WrappedEntity) {})
	entity.wrap(2)
	assert.Equal(t, 2, entity.entity.inner)
}

func TestLazyEntityIsNil(t *testing.T) {
	entity := NewLazyWrappedEntity(func() *Collection {
		return NewCollection()
	}, func(_ WrappedEntity) {})
	assert.True(t, entity.IsNil())
}
