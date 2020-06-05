package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type someStruct struct {
}

func TestGetStructName(t *testing.T) {
	n := NameFromEntity(someStruct{})

	assert.Equal(t, "d3/orm/entity/someStruct", string(n))
}

func TestGetStructPtrName(t *testing.T) {
	n := NameFromEntity(&someStruct{})

	assert.Equal(t, "d3/orm/entity/someStruct", string(n))
}

func TestShortName(t *testing.T) {
	n := NameFromEntity(&someStruct{})

	assert.Equal(t, "someStruct", n.Short())
}

func TestNameEqual(t *testing.T) {
	n := NameFromEntity(&someStruct{})
	n2 := NameFromEntity(&someStruct{})

	assert.True(t, n.Equal(n2))

	h := NameFromEntity(struct{}{})

	assert.False(t, n.Equal(h))
}
