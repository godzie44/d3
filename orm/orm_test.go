package orm

import (
	"github.com/godzie44/d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testEntity1 struct {
	Id  int64       `d3:"pk:auto"`
	Rel interface{} `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/testEntity2>"`
}

func (t *testEntity1) D3Token() entity.MetaToken {
	return entity.MetaToken{}
}

type testEntity2 struct {
	Id int `d3:"pk:auto"`
}

func (t *testEntity2) D3Token() entity.MetaToken {
	return entity.MetaToken{}
}

func TestRegisterEntities(t *testing.T) {
	orm1 := New(nil)

	assert.Error(t, orm1.Register(
		&testEntity1{},
	))

	orm2 := New(nil)

	assert.NoError(t, orm2.Register((*testEntity1)(nil), &testEntity2{}))
}
