package entity

type Copiable interface {
	DeepCopy() interface{}
}

type (
	Collection struct {
		base Collectionner
	}
	Collectionner interface {
		Copiable
		ToSlice() []interface{}
		Add(el interface{})
		Get(index int) interface{}
		Count() int
		Empty() bool
		Remove(index int)
	}
)

func NewCollection(entities ...interface{}) *Collection {
	return &Collection{base: &eagerCollection{holder: &dataHolder{data: entities}}}
}

func NewCollectionFromCollectionner(c Collectionner) *Collection {
	return &Collection{base: c}
}

func (c *Collection) DeepCopy() interface{} {
	return &Collection{base: c.base.DeepCopy().(Collectionner)}
}

func (c *Collection) ToSlice() []interface{} {
	return c.base.ToSlice()
}

func (c *Collection) Add(el interface{}) {
	c.base.Add(el)
}

func (c *Collection) Get(index int) interface{} {
	return c.base.Get(index)
}

func (c *Collection) Count() int {
	return c.base.Count()
}

func (c *Collection) Empty() bool {
	return c.base.Empty()
}

func (c *Collection) Remove(index int) {
	c.base.Remove(index)
}

type dataHolder struct {
	data []interface{}
}

func (e *dataHolder) ToSlice() []interface{} {
	return e.data
}

func (e *dataHolder) Add(el interface{}) {
	e.data = append(e.data, el)
}

func (e *dataHolder) Get(index int) interface{} {
	return e.data[index]
}

func (e *dataHolder) Count() int {
	return len(e.data)
}

func (e *dataHolder) Empty() bool {
	return len(e.data) == 0
}

func (e *dataHolder) Remove(index int) {
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
	return e.holder.ToSlice()
}

func (e *eagerCollection) Add(el interface{}) {
	e.holder.Add(el)
}

func (e *eagerCollection) Get(index int) interface{} {
	return e.holder.Get(index)
}

func (e *eagerCollection) Count() int {
	return e.holder.Count()
}

func (e *eagerCollection) Empty() bool {
	return e.holder.Empty()
}

func (e *eagerCollection) Remove(index int) {
	e.holder.Remove(index)
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
	return l.holder.ToSlice()
}

func (l *lazyCollection) Add(el interface{}) {
	l.initIfNeeded()
	l.holder.Add(el)
}

func (l *lazyCollection) Get(index int) interface{} {
	l.initIfNeeded()
	return l.holder.Get(index)
}

func (l *lazyCollection) Count() int {
	l.initIfNeeded()
	return l.holder.Count()
}

func (l *lazyCollection) Empty() bool {
	l.initIfNeeded()
	return l.holder.Empty()
}

func (l *lazyCollection) Remove(index int) {
	l.initIfNeeded()
	l.holder.Remove(index)
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
