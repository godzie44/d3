package persist

import (
	"database/sql"
	"github.com/godzie44/d3/orm/entity"
)

//d3:entity
//d3_table:shop_p
type Shop struct {
	Id      sql.NullInt32      `d3:"pk:auto"`
	Books   *entity.Collection `d3:"one_to_many:<target_entity:Book,join_on:shop_id,delete:nullable>,type:lazy"`
	Profile *entity.Cell       `d3:"one_to_one:<target_entity:ShopProfile,join_on:profile_id,delete:cascade>,type:lazy"`
	Name    string
}

//d3:entity
//d3_table:profile_p
type ShopProfile struct {
	Id          sql.NullInt32 `d3:"pk:auto"`
	Description string
}

//d3:entity
//d3_table:book_p
type Book struct {
	Id      sql.NullInt32      `d3:"pk:auto"`
	Authors *entity.Collection `d3:"many_to_many:<target_entity:Author,join_on:book_id,reference_on:author_id,join_table:book_author_p>,type:lazy"`
	Name    string
}

//d3:entity
//d3_table:author_p
type Author struct {
	Id   sql.NullInt32 `d3:"pk:auto"`
	Name string
}
