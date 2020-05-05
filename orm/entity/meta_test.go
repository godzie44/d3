package entity

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestCreateSimpleMeta(t *testing.T) {
	meta, _ := CreateMeta((*testEntity)(nil))

	assert.Equal(t, meta.EntityName, Name("d3/orm/entity/testEntity"))
}

func TestCreateMetaFromVariousReflectionsOfOneEntity(t *testing.T) {
	meta1, _ := CreateMeta((*testEntity)(nil))
	meta2, _ := CreateMeta(&testEntity{})

	//cause (*testEntity)(nil) != &testEntity{} do this
	meta2.Tpl = meta1.Tpl
	assert.Equal(t, meta1, meta2)

	meta3, _ := CreateMeta(testEntity{})
	meta3.Tpl = meta1.Tpl

	assert.Equal(t, meta1, meta3)
}

type shop struct {
	entity  struct{}      `d3:"table_name:shop"`
	ID      sql.NullInt32 `d3:"pk:auto"`
	Books   Collection    `d3:"one_to_many:<target_entity:book,join_on:shop_id>,type:lazy"`
	Profile WrappedEntity `d3:"one_to_one:<target_entity:shopProfile,join_on:profile_id,delete:cascade>"`
	Name    string
}

func TestCreateMeta(t *testing.T) {
	meta, _ := CreateMeta((*shop)(nil))

	assert.Equal(t, Name("d3/orm/entity/shop"), meta.EntityName)
	assert.Equal(t, "shop", meta.TableName)
	assert.Equal(t, FieldInfo{
		Name:           "ID",
		associatedType: reflect.TypeOf(sql.NullInt32{}),
		DbAlias:        "id",
		FullDbAlias:    "shop.id",
	}, *meta.Pk.Field)
	assert.Equal(t, OneToOne{
		baseRelation: baseRelation{
			relType:        Lazy,
			deleteStrategy: Cascade,
			targetEntity:   Name("shopProfile"),
			field:          meta.Relations["Profile"].(*OneToOne).field,
		},
		JoinColumn:      "profile_id",
		ReferenceColumn: "",
	}, *meta.Relations["Profile"].(*OneToOne))
	assert.Equal(t, OneToMany{
		baseRelation: baseRelation{
			relType:        Lazy,
			deleteStrategy: None,
			targetEntity:   Name("book"),
			field:          meta.Relations["Books"].(*OneToMany).field,
		},
		JoinColumn:      "shop_id",
		ReferenceColumn: "",
	}, *meta.Relations["Books"].(*OneToMany))

	assert.Equal(t, FieldInfo{
		Name:           "Name",
		associatedType: reflect.TypeOf(""),
		DbAlias:        "name",
		FullDbAlias:    "shop.name",
	}, *meta.Fields["Name"])
}
