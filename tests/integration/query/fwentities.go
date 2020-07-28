package query

import d3entity "github.com/godzie44/d3/orm/entity"

//d3:entity
//d3_table:test_entity_1
type fwTestEntity1 struct {
	Id   int32          `d3:"pk:auto"`
	Rel  *d3entity.Cell `d3:"one_to_one:<target_entity:fwTestEntity2,join_on:e2_id,reference_on:id>,type:lazy"`
	Data string
}

//d3:entity
//d3_table:test_entity_2
type fwTestEntity2 struct {
	Id   int32                `d3:"pk:auto"`
	Rel  *d3entity.Collection `d3:"one_to_many:<target_entity:fwTestEntity3,join_on:e2_id>,type:lazy"`
	Data string
}

//d3:entity
//d3_table:test_entity_3
type fwTestEntity3 struct {
	Id   int32                `d3:"pk:auto"`
	Rel  *d3entity.Collection `d3:"many_to_many:<target_entity:fwTestEntity4,join_on:t3_id,reference_on:t4_id,join_table:t3_t4>,type:lazy"`
	Data string
}

//d3:entity
//d3_table:test_entity_4
type fwTestEntity4 struct {
	Id   int32 `d3:"pk:auto"`
	Data string
}
