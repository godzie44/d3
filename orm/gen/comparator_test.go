package gen

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type testStruct struct {
	num64   int64
	str     string
	nullInt sql.NullInt64
}

var expectedComparatorCode = `func (t *testStruct) __d3_makeComparator() entity.FieldComparator {
	return func(e1, e2 interface{}, fName string) bool {
		if e1 == nil || e2 == nil {
			return e1 == e2
		}

		e1Typed, ok := e1.(*testStruct)
		if !ok {
			return false
		}
		e2Typed, ok := e2.(*testStruct)
		if !ok {
			return false
		}
		
		switch fName {
		
		case "num64":
			return e1Typed.num64 == e2Typed.num64
		case "str":
			return e1Typed.str == e2Typed.str
		case "nullInt":
			return e1Typed.nullInt == e2Typed.nullInt
		default:
			return false
		}
	}
}`

func TestComparatorsGenGenerate(t *testing.T) {
	buff := &strings.Builder{}
	gen := &comparator{buff}

	gen.handle(reflect.TypeOf(testStruct{}))

	assert.Equal(t, expectedComparatorCode, strings.Trim(buff.String(), "\n"))
}
