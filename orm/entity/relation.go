package entity

import (
	"d3/reflect"
	"errors"
)

type DeleteStrategy int

const (
	_ DeleteStrategy = iota
	None
	Cascade
	Nullable
)

func deleteStrategyFromAlias(alias string) DeleteStrategy {
	switch alias {
	case "cascade":
		return Cascade
	case "nullable":
		return Nullable
	default:
		return None
	}
}

type RelationType int

const (
	_ RelationType = iota
	Lazy
	Eager
	SmartLazy
)

func relationTypeFromAlias(alias string) RelationType {
	switch alias {
	case "lazy":
		return Lazy
	case "eager":
		return Eager
	default:
		return Lazy
	}
}

type Relation interface {
	Type() RelationType
	RelatedWith() Name

	Field() *FieldInfo

	setField(f *FieldInfo)
	fillFromTag(tag *parsedTag)
}

type baseRelation struct {
	relType      RelationType
	onDelete     DeleteStrategy
	targetEntity Name
	field        *FieldInfo
}

func (b *baseRelation) Type() RelationType {
	return b.relType
}

func (b *baseRelation) RelatedWith() Name {
	return b.targetEntity
}

func (b *baseRelation) Field() *FieldInfo {
	return b.field
}

func (b *baseRelation) setField(f *FieldInfo) {
	b.field = f
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

func (o *OneToMany) fillFromTag(tag *parsedTag) {
	prop, _ := tag.getProperty("one_to_many")

	relType, _ := tag.getProperty("type")
	o.baseRelation = baseRelation{
		relType:      relationTypeFromAlias(relType.val),
		targetEntity: Name(prop.getSubPropVal("target_entity")),
		onDelete:     deleteStrategyFromAlias(prop.getSubPropVal("on_delete")),
	}
	o.JoinColumn = prop.getSubPropVal("join_on")
	o.ReferenceColumn = prop.getSubPropVal("reference_on")
}

func (o *OneToMany) ExtractCollection(owner interface{}) (Collection, error) {
	val, err := reflect.ExtractStructField(owner, o.Field().Name)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return NewCollection(nil), nil
	}

	collection, ok := val.(Collection)
	if !ok {
		return nil, errors.New("field type must be Collection")
	}

	if lc, ok := collection.(LazyContainer); ok && !lc.IsInitialized() {
		return NewCollection(nil), nil
	}

	return collection, nil
}

type OneToOne struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
}

func (o *OneToOne) fillFromTag(tag *parsedTag) {
	prop, _ := tag.getProperty("one_to_one")
	relType, _ := tag.getProperty("type")

	o.baseRelation = baseRelation{
		relType:      relationTypeFromAlias(relType.val),
		targetEntity: Name(prop.getSubPropVal("target_entity")),
		onDelete:     deleteStrategyFromAlias(prop.getSubPropVal("on_delete")),
	}
	o.JoinColumn = prop.getSubPropVal("join_on")
	o.ReferenceColumn = prop.getSubPropVal("reference_on")
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

func (m *ManyToMany) fillFromTag(tag *parsedTag) {
	prop, _ := tag.getProperty("many_to_many")
	relType, _ := tag.getProperty("type")

	m.baseRelation = baseRelation{
		relType:      relationTypeFromAlias(relType.val),
		targetEntity: Name(prop.getSubPropVal("target_entity")),
		onDelete:     deleteStrategyFromAlias(prop.getSubPropVal("on_delete")),
	}
	m.JoinColumn = prop.getSubPropVal("join_on")
	m.ReferenceColumn = prop.getSubPropVal("reference_on")
	m.JoinTable = prop.getSubPropVal("join_table")
}

func (m *ManyToMany) ExtractCollection(owner interface{}) (Collection, error) {
	val, err := reflect.ExtractStructField(owner, m.Field().Name)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return NewCollection(nil), nil
	}

	collection, ok := val.(Collection)
	if !ok {
		return nil, errors.New("field type must be Collection")
	}

	if lc, ok := collection.(LazyContainer); ok && !lc.IsInitialized() {
		return NewCollection(nil), nil
	}

	return collection, nil
}
