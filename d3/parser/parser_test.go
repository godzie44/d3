package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//d3:entity
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
	assert.Equal(t, []EntityMeta{{Name: "parsedStruct"}}, p.Metas)
}
