package relation

import "d3/orm/entity"

//d3:entity
//d3_table:shop
type ShopLR struct {
	Id    int32              `d3:"pk:auto"`
	Books *entity.Collection `d3:"one_to_many:<target_entity:d3/tests/integration/relation/BookLR,join_on:t1_id>,type:lazy"`
	Name  string
}

//d3:entity
//d3_table:book
type BookLR struct {
	Id   int32 `d3:"pk:auto"`
	Name string
}

//d3:entity
//d3_table:shop
type ShopER struct {
	Id    int32              `d3:"pk:auto"`
	Books *entity.Collection `d3:"one_to_many:<target_entity:d3/tests/integration/relation/BookER,join_on:t1_id,reference_on:id>,type:eager"`
	Name  string
}

//d3:entity
//d3_table:book
type BookER struct {
	Id        int32              `d3:"pk:auto"`
	Discounts *entity.Collection `d3:"one_to_many:<target_entity:d3/tests/integration/relation/DiscountER,join_on:t2_id,reference_on:id>,type:eager"`
	Name      string
}

//d3:entity
//d3_table:discount
type DiscountER struct {
	Id    int32 `d3:"pk:auto"`
	Value int32
}
