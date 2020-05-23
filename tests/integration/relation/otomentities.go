package relation

import entity2 "d3/orm/entity"

//d3:entity
type ShopLR struct {
	Id    int32              `d3:"pk:auto"`
	Books entity2.Collection `d3:"one_to_many:<target_entity:d3/tests/integration/relation/BookLR,join_on:t1_id>,type:lazy"`
	Name  string
}

//d3:entity
type BookLR struct {
	Id int32 `d3:"pk:auto"`
	//Profile    entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/relation/PhotoLL,join_on:t3_id,reference_on:id>,type:eager"`
	Name string
}

//d3:entity
type ShopER struct {
	Id    int32              `d3:"pk:auto"`
	Books entity2.Collection `d3:"one_to_many:<target_entity:d3/tests/integration/relation/BookER,join_on:t1_id,reference_on:id>,type:eager"`
	Name  string
}

//d3:entity
type BookER struct {
	Id        int32              `d3:"pk:auto"`
	Discounts entity2.Collection `d3:"one_to_many:<target_entity:d3/tests/integration/relation/DiscountER,join_on:t2_id,reference_on:id>,type:eager"`
	Name      string
}

//d3:entity
type DiscountER struct {
	Id    int32 `d3:"pk:auto"`
	Value int32
}