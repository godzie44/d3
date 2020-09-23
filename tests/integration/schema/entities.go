package schema

import (
	"database/sql"
	"github.com/godzie44/d3/orm/entity"
	"github.com/gofrs/uuid"
	"time"
)

//d3:entity
//d3_table:shop_m
//d3_index:shop_name_idx(name)
type shop struct {
	Id      sql.NullInt32      `d3:"pk:auto"`
	Books   *entity.Collection `d3:"one_to_many:<target_entity:book,join_on:shop_id,delete:nullable>,type:lazy"`
	Profile *entity.Cell       `d3:"one_to_one:<target_entity:profile,join_on:profile_id,delete:cascade>,type:lazy"`
	Name    string
}

//d3:entity
//d3_table:profile_m
type profile struct {
	Id          sql.NullInt32 `d3:"pk:auto"`
	Description string
}

//d3:entity
//d3_table:book_m
//d3_index:book_name_idx(name)
//d3_index_unique:book_isbn_idx(isbn)
type book struct {
	Id      sql.NullInt32      `d3:"pk:auto"`
	Authors *entity.Collection `d3:"many_to_many:<target_entity:author,join_on:book_id,reference_on:author_id,join_table:book_author_m>,type:lazy"`
	Name    string
	ISBN    string
}

//d3:entity
//d3_table:author_m
//d3_index:author_name_idx(name, surname)
type author struct {
	Id      sql.NullInt32 `d3:"pk:auto"`
	Name    string
	Surname string
}

//d3:entity
//d3_table:all_types
type allTypeStruct struct {
	ID               sql.NullInt32 `d3:"pk:auto"`
	Uuid             uuid.UUID
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

type Email string
type myEmail Email

//d3:entity
//d3_table:test_aliases
type entityWithAliases struct {
	ID          sql.NullInt32 `d3:"pk:auto"`
	email       Email
	secretEmail myEmail
}
