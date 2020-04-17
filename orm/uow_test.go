package orm

import (
	"d3/mapper"
	"d3/orm/entity"
	"github.com/stretchr/testify/assert"
	"testing"
	//"testing"
)

type uowTestEntity struct {
	entity struct{} `d3:"table_name:t"`
	Id     int      `d3:"pk:auto"`
	Field1 int
	Field2 string
}

var testEntityMeta, _ = entity.CreateMeta((*uowTestEntity)(nil))

func TestRegisterNewEntity(t *testing.T) {
	uow := NewUOW(nil)

	te1 := &uowTestEntity{Id: 1}
	te2 := &uowTestEntity{Id: 2}
	err := uow.registerNew(entity.NewBox(te1, testEntityMeta))
	assert.NoError(t, err)
	err = uow.registerNew(entity.NewBox(te2, testEntityMeta))
	assert.NoError(t, err)

	assert.Equal(t, map[entity.Name][]*entity.Box{
		testEntityMeta.EntityName: {
			entity.NewBox(te1, testEntityMeta),
			entity.NewBox(te2, testEntityMeta),
		},
	}, uow.newEntities)
}

func TestRegisterNewEntityIfEntityInDirty(t *testing.T) {
	uow := NewUOW(nil)

	te1 := &uowTestEntity{Id: 1}
	te2 := &uowTestEntity{Id: 2}
	err := uow.registerNew(entity.NewBox(te1, testEntityMeta))
	assert.NoError(t, err)
	err = uow.registerNew(entity.NewBox(te2, testEntityMeta))
	assert.NoError(t, err)

	assert.Equal(t, map[entity.Name][]*entity.Box{
		testEntityMeta.EntityName: {
			entity.NewBox(te1, testEntityMeta),
			entity.NewBox(te2, testEntityMeta),
		},
	}, uow.newEntities)
}

type uowTestEntity2 struct {
	entity       struct{}             `d3:"table_name:t2"`
	Id           int                  `d3:"pk:auto"`
	OneToOneRel  entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/uowTestEntity3,join_on:t3_id>,type:lazy"`
	OneToManyRel mapper.Collection    `d3:"one_to_many:<target_entity:d3/orm/uowTestEntity3,join_on:t2_id>,type:lazy"`
}

type uowTestEntity3 struct {
	entity      struct{}             `d3:"table_name:t3"`
	Id          int                  `d3:"pk:auto"`
	OneToOneRel entity.WrappedEntity `d3:"one_to_one:<target_entity:d3/orm/uowTestEntity4,join_on:t4_id>,type:lazy"`
}

type uowTestEntity4 struct {
	entity struct{} `d3:"table_name:t4"`
	Id     int      `d3:"pk:auto"`
}

//func TestInsertGenerationOneToOneEntity(t *testing.T)  {
//	metaRegistry := entity.NewMetaRegistry()
//	metaRegistry.Add((*uowTestEntity2)(nil), (*uowTestEntity3)(nil), (*uowTestEntity4)(nil))
//
//	entityJoinedOneToOne := &uowTestEntity3{ID: 3}
//	te1 := &uowTestEntity2{ID: 1, OneToOneRel: entity.NewWrapEntity(entityJoinedOneToOne)}
//
//	meta2, _ := metaRegistry.GetMeta((*uowTestEntity2)(nil))
//
//	act, err := generateInsert(te1, &meta2, nil)
//	assert.NoError(t, err)
//
//	fmt.Println(act)
//}
//
//type dbAdapterMock struct {
//	mock.Mock
//}
//
//func newDbAdapterMock() *dbAdapterMock {
//	adapter := &dbAdapterMock{}
//	adapter.On("Insert", mock.Anything).Return(nil)
//	adapter.On("Update", mock.Anything).Return(nil)
//	adapter.On("Delete", mock.Anything).Return(nil)
//	adapter.On("DoInTransaction", mock.Anything).Return(nil)
//	return adapter
//}
//
//func (d *dbAdapterMock) Insert(_ ...*persistence.InsertAction) {
//	d.Called()
//}
//
//func (d *dbAdapterMock) Update(_ []interface{}, _ *entity.MetaInfo) {
//	d.Called()
//}
//
//func (d *dbAdapterMock) Remove(_ []interface{}, _ *entity.MetaInfo) {
//	d.Called()
//}
//
//func (d *dbAdapterMock) DoInTransaction(f func() error) error {
//	_ = f()
//	d.Called()
//	return nil
//}
//
//func TestNewUOW(t *testing.T) {
//	dbAdapter := newDbAdapterMock()
//	uow := NewUOW(dbAdapter)
//
//	uow.registerNew(&uowTestEntity{ID: 1}, testEntityMeta)
//	_ = uow.Commit()
//
//	dbAdapter.AssertNumberOfCalls(t, "DoInTransaction", 1)
//	dbAdapter.AssertNumberOfCalls(t, "Insert", 1)
//}
