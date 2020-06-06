package cache

import d3Entity "github.com/godzie44/d3/orm/entity"

//d3:entity
//d3_table:im_test_entity_1
type entity1 struct {
	Id   int32                `d3:"pk:auto"`
	Rel  *d3Entity.Collection `d3:"one_to_many:<target_entity:github.com/godzie44/d3/tests/integration/cache/entity2,join_on:t1_id>,type:eager"`
	Data string
}

//d3:entity
//d3_table:im_test_entity_2
type entity2 struct {
	Id   int32 `d3:"pk:auto"`
	Data string
}
