package persist

import (
	"d3/orm/entity"
	"database/sql"
)

//d3:entity
type Shop struct {
	Id      sql.NullInt32        `d3:"pk:auto"`
	Books   entity.Collection    `d3:"one_to_many:<target_entity:d3/tests/integration/persist/Book,join_on:shop_id,delete:nullable>,type:lazy"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/persist/ShopProfile,join_on:profile_id,delete:cascade>,type:lazy"`
	Name    string
}

//d3:entity
type ShopProfile struct {
	Id          sql.NullInt32 `d3:"pk:auto"`
	Description string
}

//d3:entity
type Book struct {
	Id      sql.NullInt32     `d3:"pk:auto"`
	Authors entity.Collection `d3:"many_to_many:<target_entity:d3/tests/integration/persist/Author,join_on:book_id,reference_on:author_id,join_table:book_author_p>,type:lazy"`
	Name    string
}

//d3:entity
type Author struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string
}