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
	entity.Wrap(2)
	assert.Equal(t, 2, entity.base.inner)
}

func TestEagerEntityIsNil(t *testing.T) {
	entity := NewWrapEntity(nil)
	assert.True(t, entity.IsNil())
}

func TestLazyEntityGet(t *testing.T) {
	entity := NewLazyWrappedEntity(func() interface{} {
		return 1
	})
	assert.Equal(t, 1, entity.Unwrap())
}

func TestLazyEntitySet(t *testing.T) {
	entity := NewLazyWrappedEntity(func() interface{} {
		return 1
	})
	entity.Wrap(2)
	assert.Equal(t, 2, entity.entity.inner)
}

func TestLazyEntityIsNil(t *testing.T) {
	entity := NewLazyWrappedEntity(func() interface{} {
		return nil
	})
	assert.True(t, entity.IsNil())
}
