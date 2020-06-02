package reflect

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type testEntity struct {
	ID    string
	Count int
}

func TestEntityToSlice(t *testing.T) {
	entity := (*testEntity)(nil)
	slice := CreateSliceOfStructPtrs(reflect.TypeOf(entity), 1)
	assert.IsType(t, []*testEntity{}, slice)
}

func TestGetFirstElementFromSlice(t *testing.T) {
	someSlice := []int{1, 2, 3}
	el, err := GetFirstElementFromSlice(someSlice)
	assert.NoError(t, err)
	assert.Equal(t, el.(int), 1)
}
