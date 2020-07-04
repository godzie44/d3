package entity

// Cell - d3 container for single entity.
type Cell struct {
	w wrapper
}

type wrapper interface {
	Copiable
	isNil() bool
	unwrap() interface{}
}

// NewCell - create new Cell.
func NewCell(entity interface{}) *Cell {
	return &Cell{w: &eagerEntity{base: &baseEntity{inner: entity}}}
}

func NewCellFromWrapper(w wrapper) *Cell {
	return &Cell{w: w}
}

func (c *Cell) DeepCopy() interface{} {
	return &Cell{w: c.w.DeepCopy().(wrapper)}
}

// isNil - return true if nil in Cell.
func (c *Cell) IsNil() bool {
	return c.w.isNil()
}

// unwrap - return entity under Cell.
func (c *Cell) Unwrap() interface{} {
	return c.w.unwrap()
}

type LazyContainer interface {
	IsInitialized() bool
}

type baseEntity struct {
	inner interface{}
}

func (b *baseEntity) isNil() bool {
	return b.inner == nil
}

func (b *baseEntity) unwrap() interface{} {
	return b.inner
}

func (b *baseEntity) wrap(entity interface{}) {
	b.inner = entity
}

type eagerEntity struct {
	base *baseEntity
}

func (e *eagerEntity) isNil() bool {
	return e.base.isNil()
}

func (e *eagerEntity) unwrap() interface{} {
	return e.base.unwrap()
}

func (e *eagerEntity) DeepCopy() interface{} {
	return &eagerEntity{base: &baseEntity{inner: e.base.inner}}
}

type lazyEntity struct {
	entity    *baseEntity
	extractor func() *Collection
	afterInit func(entity *Cell)
}

func NewLazyWrappedEntity(extractor func() *Collection, afterInit func(entity *Cell)) *lazyEntity {
	return &lazyEntity{extractor: extractor, afterInit: afterInit}
}

func (l *lazyEntity) DeepCopy() interface{} {
	if l.entity == nil {
		return &lazyEntity{entity: nil}
	}
	return &lazyEntity{entity: &baseEntity{inner: l.entity.inner}}
}

func (l *lazyEntity) initIfNeeded() {
	if !l.IsInitialized() {
		collection := l.extractor()
		if collection.Empty() {
			l.entity = &baseEntity{inner: nil}
		} else {
			l.entity = &baseEntity{inner: collection.Get(0)}
		}

		l.afterInit(&Cell{w: l})
	}
}

func (l *lazyEntity) isNil() bool {
	l.initIfNeeded()

	return l.entity.isNil()
}

func (l *lazyEntity) unwrap() interface{} {
	l.initIfNeeded()
	return l.entity.unwrap()
}

func (l *lazyEntity) wrap(entity interface{}) {
	l.initIfNeeded()
	l.entity.wrap(entity)
}

func (l *lazyEntity) IsInitialized() bool {
	return l.entity != nil
}

func CellIsLazy(cell *Cell) bool {
	_, ok := cell.w.(LazyContainer)
	return ok
}
