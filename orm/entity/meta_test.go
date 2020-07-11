package entity

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewSimpleMeta(t *testing.T) {
	meta, _ := NewMeta(
		(*testEntity)(nil),
	)

	assert.Equal(t, meta.EntityName, Name("github.com/godzie44/d3/orm/entity/testEntity"))
}

func TestNewMetaFromVariousReflectionsOfOneEntity(t *testing.T) {
	meta1, _ := NewMeta((*testEntity)(nil))
	meta2, _ := NewMeta(&testEntity{})

	//cause (*testEntity)(nil) != &testEntity{} do this
	meta2.Tpl = meta1.Tpl
	assert.Equal(t, meta1, meta2)
}

type shop struct {
	ID      sql.NullInt32 `d3:"pk:auto"`
	Books   *Collection   `d3:"one_to_many:<target_entity:book,join_on:shop_id>,type:lazy"`
	Profile *Cell         `d3:"one_to_one:<target_entity:shopProfile,join_on:profile_id,delete:cascade>"`
	Name    string
}

func (s *shop) D3Token() MetaToken {
	return MetaToken{}
}

func TestNewMetaWithRelations(t *testing.T) {
	meta, err := NewMeta((*shop)(nil))
	assert.NoError(t, err)

	assert.Equal(t, Name("github.com/godzie44/d3/orm/entity/shop"), meta.EntityName)
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
			targetEntity:   "github.com/godzie44/d3/orm/entity/shopProfile",
			field:          meta.Relations["Profile"].(*OneToOne).field,
		},
		JoinColumn:      "profile_id",
		ReferenceColumn: "",
	}, *meta.Relations["Profile"].(*OneToOne))
	assert.Equal(t, OneToMany{
		baseRelation: baseRelation{
			relType:        Lazy,
			deleteStrategy: None,
			targetEntity:   "github.com/godzie44/d3/orm/entity/book",
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

func TestNewMetaWithFieldAlias(t *testing.T) {
	meta, _ := NewMeta((*author)(nil))

	assert.Equal(t, FieldInfo{
		Name:           "Name",
		AssociatedType: reflect.TypeOf(""),
		DbAlias:        "author_name",
		FullDbAlias:    "author.author_name",
	}, *meta.Fields["Name"])
}

type shop2 struct {
	ID      sql.NullInt32 `d3:"pk:auto"`
	Books   *Collection   `d3:"one_to_many:<target_entity:book,join_on:shop_id>,type:lazy"`
	Profile *Cell         `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/entity/shopProfile,join_on:profile_id,delete:cascade>"`
}

func (s *shop2) D3Token() MetaToken {
	return MetaToken{}
}

func TestShortNamesInRelationSwitchTooFullNames(t *testing.T) {
	meta, err := NewMeta((*shop2)(nil))
	assert.NoError(t, err)

	assert.Equal(t, "github.com/godzie44/d3/orm/entity/book", string(meta.Relations["Books"].RelatedWith()))
	assert.Equal(t, "github.com/godzie44/d3/orm/entity/shopProfile", string(meta.Relations["Profile"].RelatedWith()))
}
