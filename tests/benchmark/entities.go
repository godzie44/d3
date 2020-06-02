package benchmark

import (
	"d3/orm/entity"
	"database/sql"
)

//d3:entity
type shop struct {
	id      sql.NullInt32        `d3:"pk:auto"`
	books   entity.Collection    `d3:"one_to_many:<target_entity:d3/tests/benchmark/book,join_on:shop_id,delete:nullable>,type:eager"`
	profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/benchmark/profile,join_on:profile_id,delete:cascade>,type:eager"`
	name    string               `d3:"column:name"`
}

//d3:entity
type profile struct {
	Id          sql.NullInt32 `d3:"pk:auto"`
	Description string
}

//d3:entity
type book struct {
	Id      sql.NullInt32     `d3:"pk:auto"`
	Authors entity.Collection `d3:"many_to_many:<target_entity:d3/tests/benchmark/author,join_on:book_id,reference_on:author_id,join_table:book_author_m>,type:lazy"`
	Name    string
}

//d3:entity
type author struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string
}
