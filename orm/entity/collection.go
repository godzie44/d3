package entity

type (
	Copiable interface {
		DeepCopy() interface{}
	}

	// Collection - d3 container for multiple entities of same type.
	Collection struct {
		base collectionner
	}

	collectionner interface {
		Copiable
		ToSlice() []interface{}
		Add(el interface{})
		Get(index int) interface{}
		Count() int
		Empty() bool
		Remove(index int)
	}
)

// NewCollection - create new collection of entities.
func NewCollection(entities ...interface{}) *Collection {
	return &Collection{base: &eagerCollection{holder: &dataHolder{data: entities}}}
}

func NewCollectionFromCollectionner(c collectionner) *Collection {
	return &Collection{base: c}
}

func (c *Collection) DeepCopy() interface{} {
	return &Collection{base: c.base.DeepCopy().(collectionner)}
}

// toSlice - return slice of entities.
func (c *Collection) ToSlice() []interface{} {
	return c.base.ToSlice()
}

// add - add element to collection.
func (c *Collection) Add(el interface{}) {
	c.base.Add(el)
}

// get - get element from collection by index.
func (c *Collection) Get(index int) interface{} {
	return c.base.Get(index)
}

// count - return count of elements in collection.
func (c *Collection) Count() int {
	return c.base.Count()
}

// empty - return true if collection has 0 entities, false otherwise.
func (c *Collection) Empty() bool {
	return c.base.Empty()
}

// remove - delete element from collection by index.
func (c *Collection) Remove(index int) {
	c.base.Remove(index)
}

type dataHolder struct {
	data []interface{}
}

func (e *dataHolder) toSlice() []interface{} {
	return e.data
}

func (e *dataHolder) add(el interface{}) {
	e.data = append(e.data, el)
}

func (e *dataHolder) get(index int) interface{} {
	return e.data[index]
}

func (e *dataHolder) count() int {
	return len(e.data)
}

func (e *dataHolder) empty() bool {
	return len(e.data) == 0
}

func (e *dataHolder) remove(index int) {
	copy(e.data[index:], e.data[index+1:])
	e.data[len(e.data)-1] = nil
	e.data = e.data[:len(e.data)-1]
}

type eagerCollection struct {
	holder *dataHolder
}

func (e *eagerCollection) DeepCopy() interface{} {
	dstData := make([]interface{}, len(e.holder.data))
	copy(dstData, e.holder.data)
	return &eagerCollection{holder: &dataHolder{data: dstData}}
}

func (e *eagerCollection) ToSlice() []interface{} {
	return e.holder.toSlice()
}

func (e *eagerCollection) Add(el interface{}) {
	e.holder.add(el)
}

func (e *eagerCollection) Get(index int) interface{} {
	return e.holder.get(index)
}

func (e *eagerCollection) Count() int {
	return e.holder.count()
}

func (e *eagerCollection) Empty() bool {
	return e.holder.empty()
}

func (e *eagerCollection) Remove(index int) {
	e.holder.remove(index)
}

type lazyCollection struct {
	holder    *dataHolder
	extractor func() *Collection
	afterInit func(collection *Collection)
}

func NewLazyCollection(extractor func() *Collection, afterInit func(collection *Collection)) *lazyCollection {
	return &lazyCollection{extractor: extractor, afterInit: afterInit}
}

func (l *lazyCollection) DeepCopy() interface{} {
	if l.holder == nil {
		return &lazyCollection{holder: nil}
	}

	dstData := make([]interface{}, len(l.holder.data))
	copy(dstData, l.holder.data)
	return &lazyCollection{holder: &dataHolder{data: dstData}}
}

func (l *lazyCollection) ToSlice() []interface{} {
	l.initIfNeeded()
	return l.holder.toSlice()
}

func (l *lazyCollection) Add(el interface{}) {
	l.initIfNeeded()
	l.holder.add(el)
}

func (l *lazyCollection) Get(index int) interface{} {
	l.initIfNeeded()
	return l.holder.get(index)
}

func (l *lazyCollection) Count() int {
	l.initIfNeeded()
	return l.holder.count()
}

func (l *lazyCollection) Empty() bool {
	l.initIfNeeded()
	return l.holder.empty()
}

func (l *lazyCollection) Remove(index int) {
	l.initIfNeeded()
	l.holder.remove(index)
}

func (l *lazyCollection) initIfNeeded() {
	if !l.IsInitialized() {
		l.holder = &dataHolder{data: l.extractor().ToSlice()}
		l.afterInit(&Collection{base: l})
	}
}

func (l *lazyCollection) IsInitialized() bool {
	return l.holder != nil
}

type iterator struct {
	currPos int
	c       *Collection
}

// Rewind - set iterator to start of collection
// Note that after rewind an iterator is in invalid state, use Next() for move on first collection element.
func (i *iterator) Rewind() {
	i.currPos = -1
}

// Next - move iterator to next collection element.
func (i *iterator) Next() bool {
	if i.currPos+1 >= i.c.base.Count() {
		return false
	}

	i.currPos++
	return true
}

// Value - get entity under iterator.
func (i *iterator) Value() interface{} {
	return i.c.base.Get(i.currPos)
}

// MakeIter - creates structure for iterate over collection.
func (c *Collection) MakeIter() *iterator {
	return &iterator{c: c, currPos: -1}
}
