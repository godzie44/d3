package gen

import (
	"github.com/godzie44/d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
	"time"
)

type extractorTestStruct struct {
	int    int                //nolint
	intPtr *int               //nolint
	string string             //nolint
	t      time.Time          //nolint
	wrap   *entity.Cell       //nolint
	coll   *entity.Collection //nolint
}

var expectedExtractorCode = `func (e *extractorTestStruct) __d3_makeFieldExtractor() entity.FieldExtractor {
	return func(s interface{}, name string) (interface{}, error) {
		sTyped, ok := s.(*extractorTestStruct)
		if !ok {
			return nil, fmt.Errorf("invalid entity type")
		}
		
		switch name {
		
		case "int":
			return sTyped.int, nil
		
		case "intPtr":
			return sTyped.intPtr, nil
		
		case "string":
			return sTyped.string, nil
		
		case "t":
			return sTyped.t, nil
		
		case "wrap":
			return sTyped.wrap, nil
		
		case "coll":
			return sTyped.coll, nil
		
		default:
			return nil, fmt.Errorf("field %s not found", name)
		}
	}
}`

func TestExtractorGeneration(t *testing.T) {
	buff := &strings.Builder{}
	gen := &extractor{out: buff}

	gen.handle(reflect.TypeOf(extractorTestStruct{}))

	assert.Equal(t, expectedExtractorCode, strings.Trim(buff.String(), "\n"))
}
