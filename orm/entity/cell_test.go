package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEagerEntityGet(t *testing.T) {
	entity := NewCell(1)
	assert.Equal(t, 1, entity.Unwrap())
}

func TestEagerEntityIsNil(t *testing.T) {
	entity := NewCell(nil)
	assert.True(t, entity.IsNil())
}

func TestLazyEntityGet(t *testing.T) {
	entity := NewLazyWrappedEntity(func() *Collection {
		return NewCollection(1)
	}, func(_ *Cell) {})
	assert.Equal(t, 1, entity.Unwrap())
}

func TestLazyEntitySet(t *testing.T) {
	entity := NewLazyWrappedEntity(func() *Collection {
		return NewCollection(1)
	}, func(_ *Cell) {})
	entity.wrap(2)
	assert.Equal(t, 2, entity.entity.inner)
}

func TestLazyEntityIsNil(t *testing.T) {
	entity := NewLazyWrappedEntity(func() *Collection {
		return NewCollection()
	}, func(_ *Cell) {})
	assert.True(t, entity.IsNil())
}
