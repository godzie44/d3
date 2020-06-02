package cache

import d3Entity "d3/orm/entity"

//d3:entity
type entity1 struct {
	Id   int32               `d3:"pk:auto"`
	Rel  d3Entity.Collection `d3:"one_to_many:<target_entity:d3/tests/integration/cache/entity2,join_on:t1_id>,type:eager"`
	Data string
}

//d3:entity
type entity2 struct {
	Id   int32 `d3:"pk:auto"`
	Data string
}
