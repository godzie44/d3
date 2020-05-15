package schema

import (
	"d3/orm/entity"
	"database/sql"
	"fmt"
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
	Id      sql.NullInt32        `d3:"pk:auto"`
	Books   entity.Collection    `d3:"one_to_many:<target_entity:d3/orm/schema/book,join_on:shop_id,delete:nullable>,type:lazy"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/schema/profile,join_on:profile_uuid,delete:cascade>,type:lazy"`
	Name    string
}

type profile struct {
	UUID        string `d3:"pk:manual"`
	Description string
}

type book struct {
	Id      sql.NullInt32     `d3:"pk:auto"`
	Authors entity.Collection `d3:"many_to_many:<target_entity:d3/orm/schema/author,join_on:book_id,reference_on:author_id,join_table:book_author>,type:lazy"`
	Name    string
}

type author struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string
}

func TestCreateTables(t *testing.T) {
	registry := entity.NewMetaRegistry()
	assert.NoError(t, registry.Add(
		entity.UserMapping{
			Entity:    &shop{},
			TableName: "shop",
		},
		entity.UserMapping{
			Entity:    &profile{},
			TableName: "profile",
		},
		entity.UserMapping{
			Entity:    &book{},
			TableName: "book",
		},
		entity.UserMapping{
			Entity:    &author{},
			TableName: "author",
		},
	))

	builder := &Builder{}
	commands, err := builder.createNewTableCommands(registry)
	assert.NoError(t, err)

	assert.Equal(t, &newTableCmd{
		tableName:  "shop",
		columns:    map[string]ColumnType{"id": NullInt32, "profile_uuid": String, "name": String},
		pkColumns:  []string{"id"},
		pkStrategy: entity.Auto,
	}, commands["d3/orm/schema/shop"])
	assert.Equal(t, &newTableCmd{
		tableName:  "profile",
		columns:    map[string]ColumnType{"uuid": String, "description": String},
		pkColumns:  []string{"uuid"},
		pkStrategy: entity.Manual,
	}, commands["d3/orm/schema/profile"])
	assert.Equal(t, &newTableCmd{
		tableName:  "book",
		columns:    map[string]ColumnType{"id": NullInt32, "shop_id": NullInt32, "name": String},
		pkColumns:  []string{"id"},
		pkStrategy: entity.Auto,
	}, commands["d3/orm/schema/book"])
	assert.Equal(t, &newTableCmd{
		tableName:  "author",
		columns:    map[string]ColumnType{"id": NullInt32, "name": String},
		pkColumns:  []string{"id"},
		pkStrategy: entity.Auto,
	}, commands["d3/orm/schema/author"])
	assert.Equal(t, &newTableCmd{
		tableName: "book_author",
		columns:   map[string]ColumnType{"book_id": Int32, "author_id": Int32},
		pkColumns: nil,
	}, commands["book_author"])
	fmt.Println(commands)
}
