package entity

type WrappedEntity interface {
	IsNil() bool
	Unwrap() interface{}
	Wrap(interface{})
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
	baseEntity
}

func NewEagerEntity(source interface{}) *eagerEntity {
	return &eagerEntity{baseEntity{inner: source}}
}

type lazyEntity struct {
	entity    *baseEntity
	extractor func() interface{}
}

func NewLazyEntity(extractor func() interface{}) *lazyEntity {
	return &lazyEntity{extractor: extractor}
}

func (l *lazyEntity) initIfNeeded() {
	if l.entity == nil {
		l.entity = &baseEntity{inner: l.extractor()}
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

type Name string

func (e Name) Short() string {
	return string(e)
}

func (e Name) Equal(name Name) bool {
	return e == name || e.Short() == name.Short()
}
