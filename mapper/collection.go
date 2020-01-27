package mapper

import "d3/reflect"

type Collection interface {
	ToSlice() []interface{}
	Add(el interface{})
	Get(index int) interface{}
	Count() int
	Empty() bool
}

type baseCollection struct {
	data []interface{}
}

func (e *baseCollection) ToSlice() []interface{} {
	return e.data
}

func (e *baseCollection) Add(el interface{}) {
	e.data = append(e.data, el)
}

func (e *baseCollection) Get(index int) interface{} {
	return e.data[index]
}

func (e *baseCollection) Count() int {
	return len(e.data)
}

func (e *baseCollection) Empty() bool {
	return len(e.data) == 0
}

type EagerCollection struct {
	baseCollection
}

func NewEagerCollection(data []interface{}) *EagerCollection {
	return &EagerCollection{baseCollection: baseCollection{data: data}}
}

type lazyCollection struct {
	collection *baseCollection
	extractor  func() interface{}
}

func NewLazyCollection(extractor func() interface{}) *lazyCollection {
	return &lazyCollection{extractor: extractor}
}

func (l *lazyCollection) ToSlice() []interface{} {
	l.initIfNeeded()
	return l.collection.ToSlice()
}

func (l *lazyCollection) Add(el interface{}) {
	l.initIfNeeded()
	l.collection.Add(el)
}

func (l *lazyCollection) Get(index int) interface{} {
	l.initIfNeeded()
	return l.collection.Get(index)
}

func (l *lazyCollection) Count() int {
	l.initIfNeeded()
	return l.collection.Count()
}

func (l *lazyCollection) Empty() bool {
	l.initIfNeeded()
	return l.collection.Empty()
}

func (l *lazyCollection) initIfNeeded() {
	if l.collection == nil {
		l.collection = &baseCollection{data: reflect.BreakUpSlice(l.extractor())}
	}
}
