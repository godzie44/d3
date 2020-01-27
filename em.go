package d3

const (
	StateNew = iota
	StateManaged
	StateDetach
	StateRemoved
)

type IdentityData struct {
	Meta   EntityMeta
	Entity *interface{}
	State  int
}

type EntityManager struct {
	identityMap map[string]map[string]*IdentityData
	uow UnitOfWork
}

type EntityMeta struct {
	Id string
}

func (em *EntityManager) Persist(entity *interface{}) {
	meta := processEntity(entity)

	identityData := &IdentityData{
		Meta:   meta,
		Entity: entity,
		State:  StateNew,
	}

	em.identityMap["entity_name"][meta.Id] = identityData

	//em.uow.WatchFor(*identityData)
}

func processEntity(entity *interface{}) EntityMeta {
	return EntityMeta{}
}

func (em *EntityManager) Flush() {
	//em.uow.Flush(em.identityMap)
}
