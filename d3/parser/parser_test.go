package parser

import (
	"github.com/godzie44/d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

//d3:entity
//d3_index:parsed_struct_idx(col1,col2)
//d3_index:parsed_struct_idx2(col3,col4)
//d3_index_unique:parsed_struct_uidx(ucol1,ucol2)
type parsedStruct struct { //nolint
}

type notParsedStruct struct { //nolint
}

func TestParser(t *testing.T) {
	p := Parser{}
	err := p.Parse("./parser_test.go")
	assert.NoError(t, err)

	assert.Equal(t, "github.com/godzie44/d3/d3/parser", p.PkgPath)
	assert.Equal(t, "parser", p.PkgName)
	assert.Equal(t, []EntityMeta{{
		Name: "parsedStruct",
		Indexes: []entity.Index{
			{Name: "parsed_struct_idx", Columns: []string{"col1", "col2"}},
			{Name: "parsed_struct_idx2", Columns: []string{"col3", "col4"}},
			{Name: "parsed_struct_uidx", Columns: []string{"ucol1", "ucol2"}, Unique: true},
		},
	}}, p.Metas)
}
