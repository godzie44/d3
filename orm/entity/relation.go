package entity

import (
	"d3/mapper"
	"d3/reflect"
	"errors"
)

type Relation interface {
	IsLazy() bool
	IsEager() bool
	IsSmartLazy() bool

	RelatedWith() Name

	Field() *FieldInfo
}

//type QueryExecutor interface {
//	Execute(query *query.Query) interface{}
//}

type baseRelation struct {
	relType      string
	targetEntity Name
	field        *FieldInfo
}

func (b *baseRelation) IsLazy() bool {
	return b.relType == "lazy"
}

func (b *baseRelation) IsEager() bool {
	return b.relType == "eager"
}

func (b *baseRelation) IsSmartLazy() bool {
	return b.relType == "smart_lazy"
}

func (b *baseRelation) RelatedWith() Name {
	return b.targetEntity
}

func (b *baseRelation) Field() *FieldInfo {
	return b.field
}

type ManyToOne struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
}

type ManyToOneInverse struct {
	baseRelation
	MappedBy string
}

type OneToMany struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
}

func (o *OneToMany) ExtractCollection(owner interface{}) (mapper.Collection, error) {
	val, err := reflect.ExtractStructField(owner, o.Field().Name)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return mapper.NewCollection(nil), nil
	}

	collection, ok := val.(mapper.Collection)
	if !ok {
		return nil, errors.New("field type must be Collection")
	}

	return collection, nil
}

type OneToOne struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
}

func (o *OneToOne) Extract(owner interface{}) (WrappedEntity, error) {
	val, err := reflect.ExtractStructField(owner, o.Field().Name)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return NewWrapEntity(nil), nil
	}

	wrappedEntity, ok := val.(WrappedEntity)
	if !ok {
		return nil, errors.New("field type must be WrappedEntity")
	}

	return wrappedEntity, nil
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

func (o *ManyToMany) ExtractCollection(owner interface{}) (mapper.Collection, error) {
	val, err := reflect.ExtractStructField(owner, o.Field().Name)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return mapper.NewCollection(nil), nil
	}

	collection, ok := val.(mapper.Collection)
	if !ok {
		return nil, errors.New("field type must be Collection")
	}

	return collection, nil
}
