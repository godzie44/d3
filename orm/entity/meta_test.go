package entity

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestCreateSimpleMeta(t *testing.T) {
	meta, _ := CreateMeta(
		UserMapping{
			Entity: (*testEntity)(nil),
		})

	assert.Equal(t, meta.EntityName, Name("d3/orm/entity/testEntity"))
}

func TestCreateMetaFromVariousReflectionsOfOneEntity(t *testing.T) {
	meta1, _ := CreateMeta(UserMapping{
		Entity: (*testEntity)(nil),
	})
	meta2, _ := CreateMeta(UserMapping{
		Entity: &testEntity{},
	})

	//cause (*testEntity)(nil) != &testEntity{} do this
	meta2.Tpl = meta1.Tpl
	assert.Equal(t, meta1, meta2)
}

type shop struct {
	ID      sql.NullInt32 `d3:"pk:auto"`
	Books   Collection    `d3:"one_to_many:<target_entity:book,join_on:shop_id>,type:lazy"`
	Profile WrappedEntity `d3:"one_to_one:<target_entity:shopProfile,join_on:profile_id,delete:cascade>"`
	Name    string
}

func (s *shop) D3Token() MetaToken {
	return MetaToken{}
}

func TestCreateMetaWithRelations(t *testing.T) {
	meta, err := CreateMeta(UserMapping{
		TableName: "shop",
		Entity:    (*shop)(nil),
	})
	assert.NoError(t, err)

	assert.Equal(t, Name("d3/orm/entity/shop"), meta.EntityName)
	assert.Equal(t, "shop", meta.TableName)
	assert.Equal(t, FieldInfo{
		Name:           "ID",
		AssociatedType: reflect.TypeOf(sql.NullInt32{}),
		DbAlias:        "id",
		FullDbAlias:    "shop.id",
	}, *meta.Pk.Field)
	assert.Equal(t, OneToOne{
		baseRelation: baseRelation{
			relType:        Lazy,
			deleteStrategy: Cascade,
			targetEntity:   "shopProfile",
			field:          meta.Relations["Profile"].(*OneToOne).field,
		},
		JoinColumn:      "profile_id",
		ReferenceColumn: "",
	}, *meta.Relations["Profile"].(*OneToOne))
	assert.Equal(t, OneToMany{
		baseRelation: baseRelation{
			relType:        Lazy,
			deleteStrategy: None,
			targetEntity:   "book",
			field:          meta.Relations["Books"].(*OneToMany).field,
		},
		JoinColumn:      "shop_id",
		ReferenceColumn: "",
	}, *meta.Relations["Books"].(*OneToMany))

	assert.Equal(t, FieldInfo{
		Name:           "Name",
		AssociatedType: reflect.TypeOf(""),
		DbAlias:        "name",
		FullDbAlias:    "shop.name",
	}, *meta.Fields["Name"])
}

type author struct {
	ID   sql.NullInt32 `d3:"pk:manual"`
	Name string        `d3:"column:author_name"`
}

func (a *author) D3Token() MetaToken {
	return MetaToken{}
}

func TestCreateMetaWithFieldAlias(t *testing.T) {
	meta, _ := CreateMeta(UserMapping{
		TableName: "author",
		Entity:    (*author)(nil),
	})

	assert.Equal(t, FieldInfo{
		Name:           "Name",
		AssociatedType: reflect.TypeOf(""),
		DbAlias:        "author_name",
		FullDbAlias:    "author.author_name",
	}, *meta.Fields["Name"])
}
