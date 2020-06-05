package entity

type Cell struct {
	w WrappedEntity
}

type WrappedEntity interface {
	Copiable
	IsNil() bool
	Unwrap() interface{}
}

func NewCell(entity interface{}) *Cell {
	return &Cell{w: &eagerEntity{base: &baseEntity{inner: entity}}}
}

func NewCellFromWrapper(w WrappedEntity) *Cell {
	return &Cell{w: w}
}

func (c *Cell) DeepCopy() interface{} {
	return &Cell{w: c.w.DeepCopy().(WrappedEntity)}
}

func (c *Cell) IsNil() bool {
	return c.w.IsNil()
}

func (c *Cell) Unwrap() interface{} {
	return c.w.Unwrap()
}

type LazyContainer interface {
	IsInitialized() bool
}

type baseEntity struct {
	inner interface{}
}

func (b *baseEntity) IsNil() bool {
	return b.inner == nil
}

func (b *baseEntity) Unwrap() interface{} {
	return b.inner
}

func (b *baseEntity) wrap(entity interface{}) {
	b.inner = entity
}

type eagerEntity struct {
	base *baseEntity
}

func (e *eagerEntity) IsNil() bool {
	return e.base.IsNil()
}

func (e *eagerEntity) Unwrap() interface{} {
	return e.base.Unwrap()
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

func (l *lazyEntity) IsNil() bool {
	l.initIfNeeded()

	return l.entity.IsNil()
}

func (l *lazyEntity) Unwrap() interface{} {
	l.initIfNeeded()
	return l.entity.Unwrap()
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
