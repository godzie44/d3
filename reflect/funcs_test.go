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

func TestIsFieldEquals(t *testing.T) {
	te1 := &testEntity{ID: "1", Count: 1}
	te2 := &testEntity{ID: "2", Count: 1}

	assert.True(t, IsFieldEquals(te1, te2, "Count"))
	assert.False(t, IsFieldEquals(te1, te2, "ID"))
}
