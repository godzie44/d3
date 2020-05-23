package relation

import (
	"d3/orm/entity"
	"database/sql"
)

//d3:entity
type ShopLL struct {
	ID      sql.NullInt32        `d3:"pk:auto"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/relation/ProfileLL,join_on:t2_id>,type:lazy"`
	Data    string
}

//d3:entity
type ProfileLL struct {
	ID    int32                `d3:"pk:auto"`
	Photo entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/relation/PhotoLL,join_on:t3_id,reference_on:id>,type:eager"`
	Data  string
}

//d3:entity
type PhotoLL struct {
	ID   int32 `d3:"pk:auto"`
	Data string
}

//d3:entity
type ShopEL struct {
	Id      int32                `d3:"pk:auto"`
	Profile entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/relation/ProfileLL,join_on:t2_id,reference_on:id>,type:eager"`
	Data    string
}
