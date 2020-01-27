package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEagerEntityGet(t *testing.T) {
	entity := NewEagerEntity(1)
	assert.Equal(t, 1, entity.Unwrap())
}

func TestEagerEntitySet(t *testing.T) {
	entity := NewEagerEntity(1)
	entity.Wrap(2)
	assert.Equal(t, 2, entity.baseEntity.inner)
}

func TestEagerEntityIsNil(t *testing.T) {
	entity := NewEagerEntity(nil)
	assert.True(t, entity.IsNil())
}

func TestLazyEntityGet(t *testing.T) {
	entity := NewLazyEntity(func() interface{} {
		return 1
	})
	assert.Equal(t, 1, entity.Unwrap())
}

func TestLazyEntitySet(t *testing.T) {
	entity := NewLazyEntity(func() interface{} {
		return 1
	})
	entity.Wrap(2)
	assert.Equal(t, 2, entity.entity.inner)
}

func TestLazyEntityIsNil(t *testing.T) {
	entity := NewLazyEntity(func() interface{} {
		return nil
	})
	assert.True(t, entity.IsNil())
}