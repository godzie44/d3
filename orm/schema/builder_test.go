package schema

import (
	"database/sql"
	"github.com/godzie44/d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestTypeMapping(t *testing.T) {
	tp, err := reflectTypeToDbType(reflect.TypeOf(int(1)))
	assert.NoError(t, err)
	assert.Equal(t, Int, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(int32(1)))
	assert.NoError(t, err)
	assert.Equal(t, Int32, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(int64(1)))
	assert.NoError(t, err)
	assert.Equal(t, Int64, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(""))
	assert.NoError(t, err)
	assert.Equal(t, String, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(true))
	assert.NoError(t, err)
	assert.Equal(t, Bool, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(float32(1)))
	assert.NoError(t, err)
	assert.Equal(t, Float32, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(float64(1)))
	assert.NoError(t, err)
	assert.Equal(t, Float64, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(time.Now()))
	assert.NoError(t, err)
	assert.Equal(t, Time, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(sql.NullBool{}))
	assert.NoError(t, err)
	assert.Equal(t, NullBool, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(sql.NullInt32{}))
	assert.NoError(t, err)
	assert.Equal(t, NullInt32, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(sql.NullInt64{}))
	assert.NoError(t, err)
	assert.Equal(t, NullInt64, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(sql.NullFloat64{}))
	assert.NoError(t, err)
	assert.Equal(t, NullFloat64, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(sql.NullString{}))
	assert.NoError(t, err)
	assert.Equal(t, NullString, tp)

	tp, err = reflectTypeToDbType(reflect.TypeOf(sql.NullTime{}))
	assert.NoError(t, err)
	assert.Equal(t, NullTime, tp)

	var i = 1
	_, err = reflectTypeToDbType(reflect.TypeOf(&i))
	assert.Error(t, err)
}

type shop struct {
	Id      sql.NullInt32      `d3:"pk:auto"`
	Books   *entity.Collection `d3:"one_to_many:<target_entity:github.com/godzie44/d3/orm/schema/book,join_on:shop_id,delete:nullable>,type:lazy"`
	Profile *entity.Cell       `d3:"one_to_one:<target_entity:github.com/godzie44/d3/orm/schema/profile,join_on:profile_uuid,delete:cascade>,type:lazy"`
	Name    string
}

func (s *shop) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Indexes: []entity.Index{
			{Name: "name_idx", Unique: false, Columns: []string{"name"}},
		},
	}
}

type profile struct {
	UUID        string `d3:"pk:manual"`
	Description string
}

func (p *profile) D3Token() entity.MetaToken {
	return entity.MetaToken{}
}

type book struct {
	Id      sql.NullInt32      `d3:"pk:auto"`
	Authors *entity.Collection `d3:"many_to_many:<target_entity:github.com/godzie44/d3/orm/schema/author,join_on:book_id,reference_on:author_id,join_table:book_author>,type:lazy"`
	Name    string
}

func (b *book) D3Token() entity.MetaToken {
	return entity.MetaToken{}
}

type author struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string
}

func (a *author) D3Token() entity.MetaToken {
	return entity.MetaToken{}
}

func TestCreateTables(t *testing.T) {
	registry := entity.NewMetaRegistry()
	assert.NoError(t, registry.Add(
		&shop{},
		&profile{},
		&book{},
		&author{},
	))

	builder := &Builder{}
	commands, err := builder.createNewTableCommands(registry)
	assert.NoError(t, err)

	assert.Equal(t, &newTableCmd{
		tableName:  "shop",
		columns:    map[string]ColumnType{"id": NullInt32, "profile_uuid": String, "name": String},
		pkColumns:  []string{"id"},
		pkStrategy: entity.Auto,
		indexes: []entity.Index{
			{Name: "name_idx", Unique: false, Columns: []string{"name"}},
		},
	}, commands["github.com/godzie44/d3/orm/schema/shop"])
	assert.Equal(t, &newTableCmd{
		tableName:  "profile",
		columns:    map[string]ColumnType{"uuid": String, "description": String},
		pkColumns:  []string{"uuid"},
		pkStrategy: entity.Manual,
	}, commands["github.com/godzie44/d3/orm/schema/profile"])
	assert.Equal(t, &newTableCmd{
		tableName:  "book",
		columns:    map[string]ColumnType{"id": NullInt32, "shop_id": NullInt32, "name": String},
		pkColumns:  []string{"id"},
		pkStrategy: entity.Auto,
	}, commands["github.com/godzie44/d3/orm/schema/book"])
	assert.Equal(t, &newTableCmd{
		tableName:  "author",
		columns:    map[string]ColumnType{"id": NullInt32, "name": String},
		pkColumns:  []string{"id"},
		pkStrategy: entity.Auto,
	}, commands["github.com/godzie44/d3/orm/schema/author"])
	assert.Equal(t, &newTableCmd{
		tableName: "book_author",
		columns:   map[string]ColumnType{"book_id": Int32, "author_id": Int32},
		pkColumns: []string{"book_id", "author_id"},
	}, commands["book_author"])
}
