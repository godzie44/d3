package query

import (
	"database/sql"
	"github.com/godzie44/d3/orm/entity"
)

//d3:entity
//d3_table:q_user
type User struct {
	id     sql.NullInt64      `d3:"pk:auto"`
	photos *entity.Collection `d3:"one_to_many:<target_entity:Photo,join_on:user_id>,type:lazy"`
	name   string
	age    int
}

//d3:entity
//d3_table:q_photo
type Photo struct {
	id  sql.NullInt64 `d3:"pk:auto"`
	src string
}
