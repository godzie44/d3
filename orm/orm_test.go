package orm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testEntity1 struct {
	Id  int64       `d3:",pk"`
	Rel interface{} `d3:"one_to_one:<target_entity:d3/orm/testEntity2>"`
}

type testEntity2 struct {
	Id int `d3:",pk"`
}

func TestRegisterEntities(t *testing.T) {
	orm1 := NewOrm(nil)

	assert.Panics(t, func() {
		_ = orm1.Register(&testEntity1{})
	})

	orm2 := NewOrm(nil)

	assert.NotPanics(t, func() {
		_ = orm2.Register((*testEntity1)(nil), &testEntity2{})
	})
}
