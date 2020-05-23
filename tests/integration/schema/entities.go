package schema

import (
	"d3/orm/entity"
	"database/sql"
	"time"
)

//d3:entity
type shop struct {
	Id      sql.NullInt32        `d3:"pk:auto"`
	Books   entity.Collection    `d3:"one_to_many:<target_entity:d3/tests/integration/schema/book,join_on:shop_id,delete:nullable>,type:lazy"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/schema/profile,join_on:profile_id,delete:cascade>,type:lazy"`
	Name    string
}

//d3:entity
type profile struct {
	Id          sql.NullInt32 `d3:"pk:auto"`
	Description string
}

//d3:entity
type book struct {
	Id      sql.NullInt32     `d3:"pk:auto"`
	Authors entity.Collection `d3:"many_to_many:<target_entity:d3/tests/integration/schema/author,join_on:book_id,reference_on:author_id,join_table:book_author_m>,type:lazy"`
	Name    string
}

//d3:entity
type author struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string
}

//d3:entity
type allTypeStruct struct {
	ID               sql.NullInt32 `d3:"pk:auto"`
	BoolField        bool
	IntField         int
	Int32Field       int32
	Int64Field       int64
	Float32Field     float32
	Float64Field     float64
	StringField      string
	TimeField        time.Time
	NullBoolField    sql.NullBool
	NullI32Field     sql.NullInt32
	NullI64Field     sql.NullInt64
	NullFloat64Field sql.NullFloat64
	NullStringField  sql.NullString
	NullTimeField    sql.NullTime
}
