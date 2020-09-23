package gen

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type registrarTestStruct struct {
	int int //nolint
}

var expectedRegistrarCode = `func (r *registrarTestStruct) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tpl: (*registrarTestStruct)(nil),
		TableName: "test_tab",
		Tools: entity.InternalTools{
			ExtractField: r.__d3_makeFieldExtractor(),
			SetFieldVal: r.__d3_makeFieldSetter(),
			CompareFields: r.__d3_makeComparator(),
			NewInstance: r.__d3_makeInstantiator(),
			Copy: r.__d3_makeCopier(),
		},
		Indexes: []entity.Index{
			
		},
	}
}


func (r *registrarTestStruct) __d3_makeFieldExtractor() entity.FieldExtractor {
	return func(s interface{}, name string) (interface{}, error) {
		sTyped, ok := s.(*registrarTestStruct)
		if !ok {
			return nil, fmt.Errorf("invalid entity type")
		}
		
		switch name {
		
		case "int":
			return sTyped.int, nil
		
		default:
			return nil, fmt.Errorf("field %s not found", name)
		}
	}
}


func (r *registrarTestStruct) __d3_makeInstantiator() entity.Instantiator {
	return func() interface{} {
		return &registrarTestStruct{}
	}
}


func (r *registrarTestStruct) __d3_makeFieldSetter() entity.FieldSetter {
	return func(s interface{}, name string, val interface{}) error {
		eTyped, ok := s.(*registrarTestStruct)
		if !ok {
			return fmt.Errorf("invalid entity type")
		}
		
		switch name { 
		case "int":
			eTyped.int = val.(int)
			return nil 
		
		
		default:
			return fmt.Errorf("field %s not found", name)
		}
	}
}


func (r *registrarTestStruct) __d3_makeCopier() entity.Copier {
	return func(src interface{}) interface{} {
		srcTyped, ok := src.(*registrarTestStruct)
		if !ok {
			return fmt.Errorf("invalid entity type")
		}
		
		copy := &registrarTestStruct{}
		
		copy.int = srcTyped.int 
		
		

		return copy
	}
}


func (r *registrarTestStruct) __d3_makeComparator() entity.FieldComparator {
	return func(e1, e2 interface{}, fName string) bool {
		if e1 == nil || e2 == nil {
			return e1 == e2
		}

		e1Typed, ok := e1.(*registrarTestStruct)
		if !ok {
			return false
		}
		e2Typed, ok := e2.(*registrarTestStruct)
		if !ok {
			return false
		}
		
		switch fName {
		
		case "int":
			return e1Typed.int == e2Typed.int
		default:
			return false
		}
	}
}`

func TestCodeGenerator(t *testing.T) {
	buff := &strings.Builder{}
	gen := NewGenerator(buff, "")

	gen.Prepare(reflect.TypeOf(registrarTestStruct{}), "test_tab")

	assert.Equal(t, expectedRegistrarCode, strings.Trim(gen.tempBuffer.String(), "\n"))

	gen.Write()

	assert.Contains(t, buff.String(), "import \"fmt\"")
	assert.Contains(t, buff.String(), "import \"github.com/godzie44/d3/orm/entity\"")
}
