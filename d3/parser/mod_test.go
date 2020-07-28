package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModulePath(t *testing.T) {
	goModContent := `
module github.com/godzie44/d3

go 1.13

require (
	github.com/go-sql-driver/mysql v1.4.1 // indirect
	)
`

	path := ModulePath([]byte(goModContent))
	assert.Equal(t, "github.com/godzie44/d3", path)
}
