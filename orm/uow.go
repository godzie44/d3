package orm

type DomainEntity interface {
	GetId() interface{}
}

type DbAdapter interface {
	Insert(interface{}) error
	Update(interface{}) error
	Remove(interface{}) error

	DoInTransaction(func()) error
}

type UnitOfWork struct {
	newEntities     map[interface{}]DomainEntity
	dirtyEntities   map[interface{}]DomainEntity
	deletedEntities map[interface{}]DomainEntity

	storage     DbAdapter
	identityMap *identityMap
}

func NewUOW(storage DbAdapter) *UnitOfWork {
	return &UnitOfWork{
		newEntities:     make(map[interface{}]DomainEntity),
		dirtyEntities:   make(map[interface{}]DomainEntity),
		deletedEntities: make(map[interface{}]DomainEntity),
		storage:         storage,
		identityMap:     newIdentityMap(),
	}
}

func (uow *UnitOfWork) RegisterNew(entity DomainEntity) {
	uow.newEntities[entity.GetId()] = entity
}

func (uow *UnitOfWork) registerDirty(entity DomainEntity) {
	_, isNewEntity := uow.newEntities[entity.GetId()]
	_, isRemovedEntity := uow.deletedEntities[entity.GetId()]
	if !isNewEntity && !isRemovedEntity {
		uow.dirtyEntities[entity.GetId()] = entity
	}
}

func (uow *UnitOfWork) registerRemove(entity DomainEntity) {
	uow.registerClean(entity)
	uow.deletedEntities[entity.GetId()] = entity
}

func (uow *UnitOfWork) registerClean(entity DomainEntity) {
	delete(uow.newEntities, entity.GetId())
	delete(uow.dirtyEntities, entity.GetId())
}

func (uow *UnitOfWork) Commit() error {
	return uow.storage.DoInTransaction(func() {
		for _, newEntity := range uow.newEntities {
			_ = uow.storage.Insert(newEntity)
		}

		for _, entity := range uow.dirtyEntities {
			_ = uow.storage.Update(entity)
		}

		for _, deletedEntity := range uow.deletedEntities {
			_ = uow.storage.Remove(deletedEntity)
		}
	})
}
