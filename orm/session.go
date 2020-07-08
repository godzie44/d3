package orm

import (
	"github.com/godzie44/d3/orm/entity"
	"github.com/godzie44/d3/orm/persistence"
	"github.com/godzie44/d3/orm/query"
)

type Driver interface {
	MakePusher(tx Transaction) persistence.Pusher

	ExecuteQuery(query *query.Query) ([]map[string]interface{}, error)
	BeforeQuery(fn func(query string, args ...interface{}))
	AfterQuery(fn func(query string, args ...interface{}))

	BeginTx() (Transaction, error)

	MakeScalarDataMapper() ScalarDataMapper
}

type session struct {
	storage Driver
	uow     *unitOfWork
}

func newSession(storage Driver, uow *unitOfWork) *session {
	return &session{storage: storage, uow: uow}
}

func (s *session) execute(q *query.Query, entityMeta *entity.MetaInfo) (*entity.Collection, error) {
	fetchPlan := query.Preprocessor.MakeFetchPlan(q)

	if s.uow.identityMap.canApply(fetchPlan) {
		entities, err := s.uow.identityMap.executePlan(fetchPlan)
		if err == nil {
			return entities, nil
		}
	}

	data, err := s.Execute(q)
	if err != nil {
		return nil, err
	}

	hydrator := &hydrator{session: s, meta: entityMeta, scalarMapper: s.storage.MakeScalarDataMapper(),
		afterHydrateEntity: func(b *entity.Box) {
			_ = s.uow.registerDirty(b)
		}}

	result, err := hydrator.hydrate(data, fetchPlan)
	if err != nil {
		return nil, err
	}

	s.uow.identityMap.putEntities(entityMeta, result)

	return result, nil
}

// Execute - execute query and return slice of result rows.
func (s *session) Execute(q *query.Query) ([]map[string]interface{}, error) {
	return s.storage.ExecuteQuery(q)
}

// Flush save all created, update changed and delete deleted entities within the session.
func (s *session) Flush() error {
	return s.uow.commit()
}

// Transaction for control transaction driver must provide instance of this interface.
type Transaction interface {
	Commit() error
	Rollback() error
}

// BeginTx - start transaction manually.
func (s *session) BeginTx() error {
	return s.uow.beginTx()
}

// CommitTx - commit transaction manually.
func (s *session) CommitTx() error {
	return s.uow.commitTx()
}

// RollbackTx - rollback transaction manually.
func (s *session) RollbackTx() error {
	return s.uow.rollbackTx()
}
