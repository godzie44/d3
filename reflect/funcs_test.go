package reflect

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type testEntity struct {
	Id string
}

func TestEntityToSlice(t *testing.T) {
	entity := (*testEntity)(nil)
	slice := CreateSliceOfEntities(reflect.TypeOf(entity), 1)
	assert.IsType(t, []*testEntity{}, slice)
}

func TestEntityToEntity(t *testing.T) {
	entity := &testEntity{}
	result := CreateEmptyEntity(entity)
	e := result.(*testEntity)
	e.Id = "1"

	assert.IsType(t, &testEntity{}, result)
}

func TestGetFirstElementFromSlice(t *testing.T) {
	someSlice := []int{1, 2, 3}
	el, err := GetFirstElementFromSlice(someSlice)
	assert.NoError(t, err)
	assert.Equal(t, el.(int), 1)
}
