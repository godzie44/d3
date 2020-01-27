package d3

import (
	"github.com/stretchr/testify/mock"
	"testing"
)

type testEntity struct {
	Id     string
	Field1 int
	Field2 string
}

func (t *testEntity) GetId() interface{} {
	return t.Id
}

type dbAdapterMock struct {
	mock.Mock
}

func newDbAdapterMock() *dbAdapterMock {
	adapter := &dbAdapterMock{}
	adapter.On("Insert", mock.Anything).Return(nil)
	adapter.On("Update", mock.Anything).Return(nil)
	adapter.On("Delete", mock.Anything).Return(nil)
	adapter.On("DoInTransaction", mock.Anything).Return(nil)
	return adapter
}

func (d *dbAdapterMock) Insert(interface{}) error {
	d.Called()
	return nil
}

func (d *dbAdapterMock) Update(interface{}) error {
	d.Called()
	return nil
}

func (d *dbAdapterMock) Remove(interface{}) error {
	d.Called()
	return nil
}

func (d *dbAdapterMock) DoInTransaction(f func()) error {
	f()
	d.Called()
	return nil
}

func TestNewUOW(t *testing.T) {
	dbAdapter := newDbAdapterMock()
	uow := NewUOW(dbAdapter)

	uow.RegisterNew(&testEntity{Id: "te1"})
	_ = uow.Commit()

	dbAdapter.AssertNumberOfCalls(t, "DoInTransaction", 1)
	dbAdapter.AssertNumberOfCalls(t, "Insert", 1)
}