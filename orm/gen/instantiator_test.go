package gen

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type instantiatorTestStruct struct {
}

var expectedInstantiator = `func (i *instantiatorTestStruct) __d3_makeInstantiator() entity.Instantiator {
	return func() interface{} {
		return &instantiatorTestStruct{}
	}
}`

func TestInstantiatorGeneration(t *testing.T) {
	buff := &strings.Builder{}
	gen := &instantiator{buff}

	gen.run(reflect.TypeOf(instantiatorTestStruct{}))

	assert.Equal(t, expectedInstantiator, strings.Trim(buff.String(), "\n"))
}
