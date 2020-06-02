package gen

import (
	"d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
	"time"
)

type copierTestStruct2 struct {
}

type copierTestStruct struct {
	int     int
	intPtr  *int
	string  string
	setter2 *copierTestStruct2
	t       time.Time
	wrap    entity.WrappedEntity
	coll    entity.Collection
}

var expectedCopierCode = `func (c *copierTestStruct) __d3_makeCopier() entity.Copier {
	return func(src interface{}) interface{} {
		srcTyped, ok := src.(*copierTestStruct)
		if !ok {
			return fmt.Errorf("invalid entity type")
		}
		
		copy := &copierTestStruct{}
		
		copy.int = srcTyped.int 
		copy.intPtr = srcTyped.intPtr 
		copy.string = srcTyped.string 
		copy.setter2 = srcTyped.setter2 
		copy.t = srcTyped.t 
		
		if srcTyped.wrap != nil {
			copy.wrap = srcTyped.wrap.(entity.Copiable).DeepCopy().(entity.WrappedEntity)
		} 
		if srcTyped.coll != nil {
			copy.coll = srcTyped.coll.(entity.Copiable).DeepCopy().(entity.Collection)
		} 

		return copy
	}
}`

func TestCopierGeneration(t *testing.T) {
	buff := &strings.Builder{}
	gen := &copier{out: buff, imports: map[string]struct{}{}}

	gen.handle(reflect.TypeOf(copierTestStruct{}))

	assert.Equal(t, expectedCopierCode, strings.Trim(buff.String(), "\n"))
	assert.Equal(t, []string{"d3/orm/entity"}, gen.preamble())
}
