package orm

import (
	"github.com/godzie44/d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

type imTestEntity1 struct {
	ID   int64 `d3:"pk:auto"`
	Data string
}

func (i *imTestEntity1) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*imTestEntity1).ID, nil
				case "Data":
					return s.(*imTestEntity1).Data, nil
				default:
					return nil, nil
				}
			},
		},
	}
}

type imTestEntity2 struct {
	ID   int64 `d3:"pk:auto"`
	Data string
}

func (i *imTestEntity2) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			ExtractField: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*imTestEntity2).ID, nil
				case "Data":
					return s.(*imTestEntity2).Data, nil
				default:
					return nil, nil
				}
			},
		},
	}
}

func TestPutEntities(t *testing.T) {
	im := newIdentityMap()

	meta, err := entity.NewMeta((*imTestEntity1)(nil))
	assert.NoError(t, err)

	collection := entity.NewCollection()
	collection.Add(&imTestEntity1{
		ID:   1,
		Data: "1",
	})
	collection.Add(&imTestEntity1{
		ID:   2,
		Data: "1",
	})

	im.putEntities(meta, collection)

	assert.Len(t, im.data, 1)
	assert.Len(t, im.data["github.com/godzie44/d3/orm/imTestEntity1"], 2)

	coll2 := entity.NewCollection(
		&imTestEntity2{
			ID:   1,
			Data: "1",
		},
	)

	meta2, err := entity.NewMeta((*imTestEntity2)(nil))
	assert.NoError(t, err)
	im.putEntities(meta2, coll2)

	assert.Len(t, im.data, 2)
	assert.Len(t, im.data["github.com/godzie44/d3/orm/imTestEntity2"], 1)
}
