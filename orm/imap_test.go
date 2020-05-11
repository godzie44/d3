package orm

import (
	"d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

type imTestEntity1 struct {
	ID   int64 `d3:"pk:auto"`
	Data string
}

type imTestEntity2 struct {
	ID   int64 `d3:"pk:auto"`
	Data string
}

func TestPutEntities(t *testing.T) {
	im := newIdentityMap()

	meta, err := entity.CreateMeta(entity.UserMapping{
		TableName: "test_entity_t1",
		Entity:    (*imTestEntity1)(nil),
	})
	assert.NoError(t, err)

	entities := []*imTestEntity1{
		{
			ID:   1,
			Data: "1",
		},
		{
			ID:   2,
			Data: "1",
		},
	}

	im.putEntities(meta, entities)

	assert.Len(t, im.data, 1)
	assert.Len(t, im.data["d3/orm/imTestEntity1"], 2)

	entities2 := []*imTestEntity2{
		{
			ID:   1,
			Data: "1",
		},
	}

	meta2, err := entity.CreateMeta(entity.UserMapping{
		TableName: "test_entity_t2",
		Entity:    (*imTestEntity2)(nil),
	})
	assert.NoError(t, err)
	im.putEntities(meta2, entities2)

	assert.Len(t, im.data, 2)
	assert.Len(t, im.data["d3/orm/imTestEntity2"], 1)
}
