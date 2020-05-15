package shema

import (
	"d3/orm/entity"
	"database/sql"
)

type shop struct {
	Id      sql.NullInt32        `d3:"pk:auto"`
	Books   entity.Collection    `d3:"one_to_many:<target_entity:d3/tests/integration/shema/book,join_on:shop_id,delete:nullable>,type:lazy"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/shema/profile,join_on:profile_id,delete:cascade>,type:lazy"`
	Name    string
}

type profile struct {
	Id          sql.NullInt32 `d3:"pk:auto"`
	Description string
}

type book struct {
	Id      sql.NullInt32     `d3:"pk:auto"`
	Authors entity.Collection `d3:"many_to_many:<target_entity:d3/tests/integration/shema/author,join_on:book_id,reference_on:author_id,join_table:book_author_m>,type:lazy"`
	Name    string
}

type author struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string
}
