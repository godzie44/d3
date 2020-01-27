package entity

type Relation interface {
	IsLazy() bool
	IsEager() bool
	IsSmartLazy() bool

	RelatedWith() Name

	CreateExtractor() func() interface{}
}

//type QueryExecutor interface {
//	Execute(query *query.Query) interface{}
//}

type baseRelation struct {
	RelType      string
	TargetEntity Name
}

func (b *baseRelation) IsLazy() bool {
	return b.RelType == "lazy"
}

func (b *baseRelation) IsEager() bool {
	return b.RelType == "eager"
}

func (b *baseRelation) IsSmartLazy() bool {
	return b.RelType == "smart_lazy"
}

func (b *baseRelation) RelatedWith() Name {
	return b.TargetEntity
}

type ManyToOne struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
}

func (o *ManyToOne) CreateExtractor() func() interface{} {
	return nil
}

type ManyToOneInverse struct {
	baseRelation
	MappedBy string
}

func (o *ManyToOneInverse) CreateExtractor() func() interface{} {
	return nil
}

// done
type OneToMany struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
}

func (o *OneToMany) CreateExtractor() func() interface{} {
	return nil
}

//done
type OneToOne struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
}

func (o *OneToOne) CreateExtractor() func() interface{} {
	return nil
}

//type OneToOneInverse struct {
//	baseRelation
//	MappedBy string
//}

type ManyToMany struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
	JoinTable       string
}

func (o *ManyToMany) CreateExtractor() func() interface{} {
	return nil
}
