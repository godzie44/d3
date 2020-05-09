package orm

import (
	"d3/orm/entity"
	"d3/orm/persistence"
	"d3/orm/query"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (s *storageMock) MakePusher(_ Transaction) persistence.Pusher {
	args := s.Called()
	return args.Get(0).(persistence.Pusher)
}

func (s *storageMock) ExecuteQuery(_ *query.Query) ([]map[string]interface{}, error) {
	return nil, nil
}

func (s *storageMock) BeforeQuery(_ func(query string, args ...interface{})) {
	return
}

func (s *storageMock) AfterQuery(_ func(query string, args ...interface{})) {
	return
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
