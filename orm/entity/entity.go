package entity

type WrappedEntity interface {
	IsNil() bool
	Unwrap() interface{}
	Wrap(interface{})
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

func (b *baseEntity) Wrap(entity interface{}) {
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

func (e *eagerEntity) Wrap(i interface{}) {
	e.base.Wrap(i)
}

func (e *eagerEntity) DeepCopy() interface{} {
	return &eagerEntity{base: &baseEntity{inner: e.base.inner}}
}

func NewWrapEntity(source interface{}) *eagerEntity {
	return &eagerEntity{base: &baseEntity{inner: source}}
}

type lazyEntity struct {
	entity    *baseEntity
	extractor func() interface{}
	afterInit func(entity WrappedEntity)
}

func NewLazyWrappedEntity(extractor func() interface{}, afterInit func(entity WrappedEntity)) *lazyEntity {
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
		l.entity = &baseEntity{inner: l.extractor()}
		l.afterInit(l)
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

func (l *lazyEntity) Wrap(entity interface{}) {
	l.initIfNeeded()
	l.entity.Wrap(entity)
}

func (l *lazyEntity) IsInitialized() bool {
	return l.entity != nil
}

type Name string

func (e Name) Short() string {
	return string(e)
}

func (e Name) Equal(name Name) bool {
	return e == name || e.Short() == name.Short()
}
