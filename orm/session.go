package orm

import (
	"d3"
	"d3/orm/entity"
	"d3/orm/query"
)

type StorageAdapter interface {
	ExecuteQuery(query *query.Query) ([]map[string]interface{}, error)

	BeforeQuery(fn func(query string, args ...interface{}))
	AfterQuery(fn func(query string, args ...interface{}))

	Insert(interface{}) error
	Update(interface{}) error
	Remove(interface{}) error

	DoInTransaction(func()) error
}

type Session struct {
	storage      StorageAdapter
	uow          *d3.UnitOfWork
	MetaRegistry *entity.MetaRegistry
	iMap         *identityMap
}

func NewSession(storage StorageAdapter, uow *d3.UnitOfWork, metaRegistry *entity.MetaRegistry) *Session {
	return &Session{storage: storage, uow: uow, MetaRegistry: metaRegistry, iMap: newIdentityMap()}
}

func (s *Session) Execute(q *query.Query) (interface{}, error) {
	fetchPlan := query.Preprocessor.CreateFetchPlan(q)

	if s.iMap.canApply(fetchPlan) {
		entities, err := s.iMap.executePlan(fetchPlan)
		if err == nil {
			return entities, nil
		}
	}

	data, err := s.storage.ExecuteQuery(q)
	if err != nil {
		return nil, err
	}

	hydrator := &Hydrator{session: s, meta: q.OwnerMeta()}

	result, err := hydrator.Hydrate(data, fetchPlan)
	if err != nil {
		return nil, err
	}

	s.iMap.putEntities(q.OwnerMeta(), result)

	return result, nil
}

func (s *Session) Flush() error {
	return s.uow.Commit()
}
