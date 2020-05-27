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

type testCollection struct {
	Data []interface{}
}

func newTestCollection(d []interface{}) *testCollection {
	return &testCollection{Data: d}
}

type testEntity3 struct {
	ID    string
	Count int
	Rel   interface{}
}

func TestCopy(t *testing.T) {
	te := &testEntity3{ID: "1", Count: 1, Rel: newTestCollection([]interface{}{1, 2, 3})}

	copyTe := Copy(te).(*testEntity3)
	assert.Equal(t, te.ID, copyTe.ID)
	assert.Equal(t, te.Count, copyTe.Count)
	assert.Equal(t, te.Rel.(*testCollection).Data, copyTe.Rel.(*testCollection).Data)

	te.Rel.(*testCollection).Data = te.Rel.(*testCollection).Data[1:]

	assert.Equal(t, te.ID, copyTe.ID)
	assert.Equal(t, te.Count, copyTe.Count)
	assert.NotEqual(t, te.Rel.(*testCollection).Data, copyTe.Rel.(*testCollection).Data)
}

func TestIsFieldEquals(t *testing.T) {
	te1 := &testEntity{ID: "1", Count: 1}
	te2 := &testEntity{ID: "2", Count: 1}

	assert.True(t, IsFieldEquals(te1, te2, "Count"))
	assert.False(t, IsFieldEquals(te1, te2, "ID"))
}
