package entity

import "d3/reflect"

type Copiable interface {
	DeepCopy() interface{}
}

type Collection interface {
	Copiable
	ToSlice() []interface{}
	Add(el interface{})
	Get(index int) interface{}
	Count() int
	Empty() bool
	Remove(index int)
}

type baseCollection struct {
	Data []interface{}
}

func (e *baseCollection) ToSlice() []interface{} {
	return e.Data
}

func (e *baseCollection) Add(el interface{}) {
	e.Data = append(e.Data, el)
}

func (e *baseCollection) Get(index int) interface{} {
	return e.Data[index]
}

func (e *baseCollection) Count() int {
	return len(e.Data)
}

func (e *baseCollection) Empty() bool {
	return len(e.Data) == 0
}

func (e *baseCollection) Remove(index int) {
	copy(e.Data[index:], e.Data[index+1:])
	e.Data[len(e.Data)-1] = nil
	e.Data = e.Data[:len(e.Data)-1]
}

type EagerCollection struct {
	base *baseCollection
}

func NewCollection(entities []interface{}) *EagerCollection {
	return &EagerCollection{base: &baseCollection{Data: entities}}
}

func (e *EagerCollection) DeepCopy() interface{} {
	dstData := make([]interface{}, len(e.base.Data))
	copy(dstData, e.base.Data)
	return &EagerCollection{base: &baseCollection{Data: dstData}}
}

func (e *EagerCollection) ToSlice() []interface{} {
	return e.base.ToSlice()
}

func (e *EagerCollection) Add(el interface{}) {
	e.base.Add(el)
}

func (e *EagerCollection) Get(index int) interface{} {
	return e.base.Get(index)
}

func (e *EagerCollection) Count() int {
	return e.base.Count()
}

func (e *EagerCollection) Empty() bool {
	return e.base.Empty()
}

func (e *EagerCollection) Remove(index int) {
	e.base.Remove(index)
}

type lazyCollection struct {
	collection *baseCollection
	extractor  func() interface{}
	afterInit  func(collection Collection)
}

func NewLazyCollection(extractor func() interface{}, afterInit func(collection Collection)) *lazyCollection {
	return &lazyCollection{extractor: extractor, afterInit: afterInit}
}

func (l *lazyCollection) DeepCopy() interface{} {
	if l.collection == nil {
		return &lazyCollection{collection: nil}
	}

	dstData := make([]interface{}, len(l.collection.Data))
	copy(dstData, l.collection.Data)
	return &lazyCollection{collection: &baseCollection{Data: dstData}}
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

func (l *lazyCollection) Remove(index int) {
	l.initIfNeeded()
	l.collection.Remove(index)
}

func (l *lazyCollection) initIfNeeded() {
	if !l.IsInitialized() {
		l.collection = &baseCollection{Data: reflect.BreakUpSlice(l.extractor())}
		l.afterInit(l)
	}
}

func (l *lazyCollection) IsInitialized() bool {
	return l.collection != nil
}
