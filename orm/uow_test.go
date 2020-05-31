package orm

import (
	"d3/orm/entity"
	"d3/orm/persistence"
	"d3/orm/query"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
	//"testing"
)

type uowTestEntity struct {
	ID     int `d3:"pk:manual"`
	Field1 int
	Field2 string
}

func (u *uowTestEntity) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tools: entity.InternalTools{
			FieldExtractor: func(s interface{}, name string) (interface{}, error) {
				switch name {
				case "ID":
					return s.(*uowTestEntity).ID, nil
				case "Field1":
					return s.(*uowTestEntity).Field1, nil
				case "Field2":
					return s.(*uowTestEntity).Field2, nil
				default:
					return nil, nil
				}
			},
			Copier: func(src interface{}) interface{} {
				return nil
			},
			CompareFields: func(e1, e2 interface{}, fName string) bool {
				if e1 == nil || e2 == nil {
					return e1 == e2
				}
				e1T := e1.(*uowTestEntity)
				e2T := e2.(*uowTestEntity)
				switch fName {
				case "ID":
					return e1T.ID == e2T.ID
				case "Field1":
					return e1T.Field1 == e2T.Field1
				case "Field2":
					return e1T.Field2 == e2T.Field2
				default:
					return false
				}
			},
		},
	}
}

var testEntityMeta, _ = entity.CreateMeta(entity.UserMapping{
	TableName: "t",
	Entity:    (*uowTestEntity)(nil),
})

func TestRegisterNewEntity(t *testing.T) {
	uow := NewUOW(nil)

	te1 := &uowTestEntity{ID: 1}
	te2 := &uowTestEntity{ID: 2}
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

	te1 := &uowTestEntity{ID: 1}

	box := entity.NewBox(te1, testEntityMeta)
	err := uow.registerDirty(box)
	assert.NoError(t, err)

	err = uow.registerNew(box)
	assert.NoError(t, err)

	assert.Empty(t, uow.newEntities[box.GetEName()])
}

func TestTransactionRollbackIfError(t *testing.T) {
	storageMock := &storageMock{}
	storageMock.On("MakePusher").Return(&alwaysErrPusher{})
	txMock := &transactionMock{}
	txMock.On("Rollback")

	storageMock.On("BeginTx").Return(txMock)

	uow := NewUOW(storageMock)

	assert.NoError(t, uow.registerNew(entity.NewBox(&uowTestEntity{}, testEntityMeta)))
	_ = uow.Commit()

	txMock.AssertNumberOfCalls(t, "Rollback", 1)
	txMock.AssertNumberOfCalls(t, "Commit", 0)
}

func TestTransactionCommitIfNoError(t *testing.T) {
	storageMock := &storageMock{}
	storageMock.On("MakePusher").Return(&alwaysOkPusher{})
	txMock := &transactionMock{}
	txMock.On("Commit")

	storageMock.On("BeginTx").Return(txMock)

	uow := NewUOW(storageMock)

	assert.NoError(t, uow.registerNew(entity.NewBox(&uowTestEntity{}, testEntityMeta)))
	_ = uow.Commit()

	txMock.AssertNumberOfCalls(t, "Rollback", 0)
	txMock.AssertNumberOfCalls(t, "Commit", 1)
}

type storageMock struct {
	mock.Mock
}

func (s *storageMock) MakeRawDataMapper() RawDataMapper {
	return func(data interface{}, into reflect.Kind) interface{} {
		return data
	}
}

func (s *storageMock) MakePusher(_ Transaction) persistence.Pusher {
	args := s.Called()
	return args.Get(0).(persistence.Pusher)
}

func (s *storageMock) ExecuteQuery(_ *query.Query) ([]map[string]interface{}, error) {
	return nil, nil
}

func (s *storageMock) BeforeQuery(_ func(query string, args ...interface{})) {
}

func (s *storageMock) AfterQuery(_ func(query string, args ...interface{})) {
}

func (s *storageMock) BeginTx() (Transaction, error) {
	args := s.Called()
	return args.Get(0).(Transaction), nil
}

type alwaysErrPusher struct {
}

var pusherErr = errors.New("pusher err")

func (a *alwaysErrPusher) Insert(_ string, _ []string, _ []interface{}) error {
	return pusherErr
}

func (a alwaysErrPusher) InsertWithReturn(_ string, _ []string, _ []interface{}, _ []string, _ func(scanner persistence.Scanner) error) error {
	return pusherErr
}

func (a alwaysErrPusher) Update(_ string, _ []string, _ []interface{}, _ map[string]interface{}) error {
	return pusherErr
}

func (a alwaysErrPusher) Remove(_ string, _ map[string]interface{}) error {
	return pusherErr
}

type alwaysOkPusher struct {
}

func (a alwaysOkPusher) Insert(_ string, _ []string, _ []interface{}) error {
	return nil
}

func (a alwaysOkPusher) InsertWithReturn(_ string, _ []string, _ []interface{}, _ []string, _ func(scanner persistence.Scanner) error) error {
	return nil
}

func (a alwaysOkPusher) Update(_ string, _ []string, _ []interface{}, _ map[string]interface{}) error {
	return nil
}

func (a alwaysOkPusher) Remove(_ string, _ map[string]interface{}) error {
	return nil
}

type transactionMock struct {
	mock.Mock
}

func (t *transactionMock) Commit() error {
	t.Called()
	return nil
}

func (t *transactionMock) Rollback() error {
	t.Called()
	return nil
}
