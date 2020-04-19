package orm

import (
	"d3/orm/entity"
	"d3/orm/persistence"
	"d3/orm/query"
)

type Storage interface {
	persistence.Storage

	ExecuteQuery(query *query.Query) ([]map[string]interface{}, error)
	BeforeQuery(fn func(query string, args ...interface{}))
	AfterQuery(fn func(query string, args ...interface{}))
}

type Session struct {
	storage      Storage
	uow          *UnitOfWork
	MetaRegistry *entity.MetaRegistry
}

func NewSession(storage Storage, uow *UnitOfWork, metaRegistry *entity.MetaRegistry) *Session {
	return &Session{storage: storage, uow: uow, MetaRegistry: metaRegistry}
}

func (s *Session) execute(q *query.Query) (interface{}, error) {
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

	hydrator := &Hydrator{session: s, meta: q.OwnerMeta()}

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
