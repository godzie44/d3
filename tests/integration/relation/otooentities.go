package relation

import (
	"database/sql"
	"github.com/godzie44/d3/orm/entity"
)

//d3:entity
//d3_table:shop
type ShopLL struct {
	ID      sql.NullInt32 `d3:"pk:auto"`
	Profile *entity.Cell  `d3:"one_to_one:<target_entity:github.com/godzie44/d3/tests/integration/relation/ProfileLL,join_on:t2_id>,type:lazy"`
	Data    string
}

//d3:entity
//d3_table:profile
type ProfileLL struct {
	ID    int32        `d3:"pk:auto"`
	Photo *entity.Cell `d3:"one_to_one:<target_entity:github.com/godzie44/d3/tests/integration/relation/PhotoLL,join_on:t3_id,reference_on:id>,type:eager"`
	Data  string
}

//d3:entity
//d3_table:photo
type PhotoLL struct {
	ID   int32 `d3:"pk:auto"`
	Data string
}

//d3:entity
//d3_table:shop
type ShopEL struct {
	Id      int32        `d3:"pk:auto"`
	Profile *entity.Cell `d3:"one_to_one:<target_entity:github.com/godzie44/d3/tests/integration/relation/ProfileLL,join_on:t2_id,reference_on:id>,type:eager"`
	Data    string
}
