package orm

import (
	"d3/orm/entity"
	"d3/orm/persistence"
	"d3/orm/query"
	"fmt"
)

type Storage interface {
	MakePusher(tx Transaction) persistence.Pusher

	ExecuteQuery(query *query.Query) ([]map[string]interface{}, error)
	BeforeQuery(fn func(query string, args ...interface{}))
	AfterQuery(fn func(query string, args ...interface{}))

	BeginTx() (Transaction, error)

	MakeRawDataMapper() RawDataMapper
}

type Session struct {
	storage      Storage
	uow          *UnitOfWork
	metaRegistry *entity.MetaRegistry
}

func NewSession(storage Storage, uow *UnitOfWork, metaRegistry *entity.MetaRegistry) *Session {
	return &Session{storage: storage, uow: uow, metaRegistry: metaRegistry}
}

func (s *Session) execute(q *query.Query) (*entity.Collection, error) {
	fetchPlan := query.Preprocessor.CreateFetchPlan(q)

	if s.uow.identityMap.canApply(fetchPlan) {
		entities, err := s.uow.identityMap.executePlan(fetchPlan)
		if err == nil {
			return entities, nil
		}
	}

	data, err := s.storage.ExecuteQuery(q)
	if err != nil {
		return nil, err
	}

	hydrator := &Hydrator{session: s, meta: q.OwnerMeta(), rawMapper: s.storage.MakeRawDataMapper(),
		afterHydrateEntity: func(b *entity.Box) {
			_ = s.uow.registerDirty(b)
		}}

	result, err := hydrator.Hydrate(data, fetchPlan)
	if err != nil {
		return nil, err
	}

	s.uow.identityMap.putEntities(q.OwnerMeta(), result)

	return result, nil
}

//Flush save all created, update changed and delete deleted entities within the session.
func (s *Session) Flush() error {
	return s.uow.Commit()
}

func (s *Session) MakeRepository(entity interface{}) (*Repository, error) {
	entityMeta, err := s.metaRegistry.GetMeta(entity)
	if err != nil {
		return nil, fmt.Errorf("repository: %w", err)
	}

	return &Repository{
		session:    s,
		entityMeta: entityMeta,
	}, nil
}

type Transaction interface {
	Commit() error
	Rollback() error
}

func (s *Session) BeginTx() error {
	return s.uow.beginTx()
}

func (s *Session) CommitTx() error {
	return s.uow.commitTx()
}

func (s *Session) RollbackTx() error {
	return s.uow.rollbackTx()
}
