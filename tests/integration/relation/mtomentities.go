package relation

import "github.com/godzie44/d3/orm/entity"

//d3:entity
//d3_table:book
type BookLL struct {
	ID      int32              `d3:"pk:auto"`
	Authors *entity.Collection `d3:"many_to_many:<target_entity:github.com/godzie44/d3/tests/integration/relation/AuthorLL,join_on:book_id,reference_on:author_id,join_table:book_author>,type:lazy"`
	Name    string
}

//d3:entity
//d3_table:author
type AuthorLL struct {
	ID   int32 `d3:"pk:auto"`
	Name string
}

//d3:entity
//d3_table:book
type BookEL struct {
	Id   int32              `d3:"pk:auto"`
	Rel  *entity.Collection `d3:"many_to_many:<target_entity:AuthorEL,join_on:book_id,reference_on:author_id,join_table:book_author>,type:eager"`
	Name string
}

//d3:entity
//d3_table:author
type AuthorEL struct {
	Id   int32              `d3:"pk:auto"`
	Rel  *entity.Collection `d3:"many_to_many:<target_entity:Redactor,join_on:author_id,reference_on:redactor_id,join_table:author_redactor>,type:eager"`
	Name string
}

//d3:entity
//d3_table:redactor
type Redactor struct {
	Id   int32 `d3:"pk:auto"`
	Name string
}
