package relation

import d3entity "d3/orm/entity"

//d3:entity
type fwTestEntity1 struct {
	Id   int32                  `d3:"pk:auto"`
	Rel  d3entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/tests/integration/relation/fwTestEntity2,join_on:e2_id,reference_on:id>,type:lazy"`
	Data string
}

//d3:entity
type fwTestEntity2 struct {
	Id   int32               `d3:"pk:auto"`
	Rel  d3entity.Collection `d3:"one_to_many:<target_entity:d3/tests/integration/relation/fwTestEntity3,join_on:e2_id>,type:lazy"`
	Data string
}

//d3:entity
type fwTestEntity3 struct {
	Id   int32               `d3:"pk:auto"`
	Rel  d3entity.Collection `d3:"many_to_many:<target_entity:d3/tests/integration/relation/fwTestEntity4,join_on:t3_id,reference_on:t4_id,join_table:t3_t4>,type:lazy"`
	Data string
}

//d3:entity
type fwTestEntity4 struct {
	Id   int32 `d3:"pk:auto"`
	Data string
}
