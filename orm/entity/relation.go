package entity

import (
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
	DeleteStrategy() DeleteStrategy
	RelatedWith() Name

	Field() *FieldInfo

	setField(f *FieldInfo)
	fillFromTag(tag *parsedTag, parent *MetaInfo)
}

type baseRelation struct {
	relType        RelationType
	deleteStrategy DeleteStrategy
	targetEntity   Name
	field          *FieldInfo
}

func (b *baseRelation) Type() RelationType {
	return b.relType
}

func (b *baseRelation) DeleteStrategy() DeleteStrategy {
	return b.deleteStrategy
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

func (o *OneToMany) fillFromTag(tag *parsedTag, parent *MetaInfo) {
	prop, _ := tag.getProperty("one_to_many")

	relType, _ := tag.getProperty("type")
	o.baseRelation = baseRelation{
		relType:        relationTypeFromAlias(relType.val),
		targetEntity:   nameFromTag(prop.getSubPropVal("target_entity"), parent.EntityName),
		deleteStrategy: deleteStrategyFromAlias(prop.getSubPropVal("delete")),
	}
	o.JoinColumn = prop.getSubPropVal("join_on")
	o.ReferenceColumn = prop.getSubPropVal("reference_on")
}

func (o *OneToMany) ExtractCollection(ownerBox *Box) (*Collection, error) {
	val, err := ownerBox.Meta.Tools.ExtractField(ownerBox.Entity, o.Field().Name)
	if err != nil {
		return nil, err
	}

	if val == nil || val == nilCollection {
		return NewCollection(), nil
	}

	collection, ok := val.(*Collection)
	if !ok {
		return nil, errors.New("field type must be Collection")
	}

	if lc, ok := collection.base.(LazyContainer); ok && !lc.IsInitialized() {
		return NewCollection(), nil
	}

	return collection, nil
}

type OneToOne struct {
	baseRelation
	JoinColumn      string
	ReferenceColumn string
}

func (o *OneToOne) fillFromTag(tag *parsedTag, parent *MetaInfo) {
	prop, _ := tag.getProperty("one_to_one")
	relType, _ := tag.getProperty("type")

	o.baseRelation = baseRelation{
		relType:        relationTypeFromAlias(relType.val),
		targetEntity:   nameFromTag(prop.getSubPropVal("target_entity"), parent.EntityName),
		deleteStrategy: deleteStrategyFromAlias(prop.getSubPropVal("delete")),
	}
	o.JoinColumn = prop.getSubPropVal("join_on")
	o.ReferenceColumn = prop.getSubPropVal("reference_on")
}

var nilCell *Cell

func (o *OneToOne) Extract(ownerBox *Box) (*Cell, error) {
	val, err := ownerBox.Meta.Tools.ExtractField(ownerBox.Entity, o.Field().Name)
	if err != nil {
		return nil, err
	}

	if val == nil || val == nilCell {
		return NewCell(nil), nil
	}

	cell, ok := val.(*Cell)
	if !ok {
		return nil, errors.New("field type must be wrapper")
	}

	return cell, nil
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

func (m *ManyToMany) fillFromTag(tag *parsedTag, parent *MetaInfo) {
	prop, _ := tag.getProperty("many_to_many")
	relType, _ := tag.getProperty("type")

	m.baseRelation = baseRelation{
		relType:        relationTypeFromAlias(relType.val),
		targetEntity:   nameFromTag(prop.getSubPropVal("target_entity"), parent.EntityName),
		deleteStrategy: deleteStrategyFromAlias(prop.getSubPropVal("delete")),
	}
	m.JoinColumn = prop.getSubPropVal("join_on")
	m.ReferenceColumn = prop.getSubPropVal("reference_on")
	m.JoinTable = prop.getSubPropVal("join_table")
}

var nilCollection *Collection

func (m *ManyToMany) ExtractCollection(ownerBox *Box) (*Collection, error) {
	val, err := ownerBox.Meta.Tools.ExtractField(ownerBox.Entity, m.Field().Name)
	if err != nil {
		return nil, err
	}

	if val == nil || val == nilCollection {
		return NewCollection(), nil
	}

	collection, ok := val.(*Collection)
	if !ok {
		return nil, errors.New("field type must be Collection")
	}

	if lc, ok := collection.base.(LazyContainer); ok && !lc.IsInitialized() {
		return NewCollection(), nil
	}

	return collection, nil
}
