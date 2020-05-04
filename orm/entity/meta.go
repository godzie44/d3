package entity

import (
	d3reflect "d3/reflect"
	"errors"
	"reflect"
	"regexp"
	"strings"
)

type PkStrategy int

const (
	_ PkStrategy = iota
	Auto
	Manual
)

type MetaInfo struct {
	Tpl        interface{}
	EntityName Name
	TableName  string

	Relations map[string]Relation
	Fields    map[string]*FieldInfo
	Pk        *pk

	RelatedMeta map[Name]*MetaInfo
}

type FieldInfo struct {
	Name           string
	associatedType reflect.Type
	DbAlias        string
	FullDbAlias    string
}

func CreateMeta(e interface{}) (*MetaInfo, error) {
	eType, err := d3reflect.IntoStructType(reflect.TypeOf(e))
	if err != nil {
		return nil, err
	}

	tableName, err := parseEntityTableName(eType)
	if err != nil {
		tableName = strings.ToLower(eType.Name())
	}

	meta := &MetaInfo{
		Tpl:         e,
		TableName:   tableName,
		Fields:      make(map[string]*FieldInfo),
		Relations:   make(map[string]Relation),
		RelatedMeta: make(map[Name]*MetaInfo),
		EntityName:  Name(d3reflect.FullName(eType)),
	}

	for i := 0; i < eType.NumField(); i++ {
		fieldReflection := eType.Field(i)

		if fieldReflection.Name == "entity" {
			continue
		}

		tag := parseTag(fieldReflection.Tag)

		field := &FieldInfo{
			Name:           fieldReflection.Name,
			associatedType: fieldReflection.Type,
		}

		if tag.hasRelation() {
			relation := extractRelation(tag, field)
			meta.Relations[fieldReflection.Name] = relation
		} else {
			field.DbAlias = extractDbFieldAlias(tag, fieldReflection.Name)
			meta.Fields[fieldReflection.Name] = field
		}

		field.FullDbAlias = meta.FullColumnAlias(field.DbAlias)

		if tag.hasProperty("pk") {
			meta.Pk = &pk{field, extractPkStrategy(tag)}
		}
	}

	if meta.Pk == nil {
		panic("pk not found in entity: " + meta.EntityName)
	}

	return meta, nil
}

func parseEntityTableName(eType reflect.Type) (string, error) {
	if metaField, ok := eType.FieldByName("entity"); ok {
		parsedTag := parseTag(metaField.Tag)
		if parsedTag == nil {
			return "", errors.New("tags not found")
		}

		tableNameProp, found := parsedTag.getProperty("table_name")
		if !found {
			return "", errors.New("tag table_name not found")
		}

		return tableNameProp.val, nil
	}

	return "", errors.New("field entity not found")
}

func extractRelation(tag *parsedTag, field *FieldInfo) Relation {
	var relTypeAlias string
	relType, exists := tag.getProperty("type")
	if !exists {
		relTypeAlias = "lazy"
	} else {
		relTypeAlias = relType.val
	}

	if prop, exists := tag.getProperty("one_to_one"); exists {
		return &OneToOne{
			baseRelation:    baseRelation{relType: relTypeAlias, targetEntity: Name(prop.getSubPropVal("target_entity")), field: field},
			JoinColumn:      prop.getSubPropVal("join_on"),
			ReferenceColumn: prop.getSubPropVal("reference_on"),
		}
	}

	if prop, exists := tag.getProperty("one_to_many"); exists {
		return &OneToMany{
			baseRelation:    baseRelation{relType: relTypeAlias, targetEntity: Name(prop.getSubPropVal("target_entity")), field: field},
			JoinColumn:      prop.getSubPropVal("join_on"),
			ReferenceColumn: prop.getSubPropVal("reference_on"),
		}
	}

	if prop, exists := tag.getProperty("many_to_many"); exists {
		return &ManyToMany{
			baseRelation:    baseRelation{relType: relTypeAlias, targetEntity: Name(prop.getSubPropVal("target_entity")), field: field},
			JoinColumn:      prop.getSubPropVal("join_on"),
			ReferenceColumn: prop.getSubPropVal("reference_on"),
			JoinTable:       prop.getSubPropVal("join_table"),
		}
	}

	return nil
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

//https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6
func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func extractDbFieldAlias(tag *parsedTag, fieldName string) string {
	if tag == nil {
		return toSnakeCase(fieldName)
	}

	prop, exists := tag.getProperty("alias")
	if !exists {
		return toSnakeCase(fieldName)
	}

	return prop.val
}

func (m *MetaInfo) DependencyEntities() map[Name]struct{} {
	dependencies := make(map[Name]struct{})

	for _, relation := range m.Relations {
		dependencies[relation.RelatedWith()] = struct{}{}
	}

	return dependencies
}

func (m *MetaInfo) dependenciesIsSet() bool {
	return len(m.RelatedMeta) == len(m.DependencyEntities())
}

func (m *MetaInfo) FindRelativeMetaRecursive(entityName Name) (*MetaInfo, bool) {
	meta, exists := m.RelatedMeta[entityName]
	if exists {
		return meta, true
	}

	for _, meta := range m.RelatedMeta {
		if m, exists := meta.FindRelativeMetaRecursive(entityName); exists {
			return m, true
		}
	}

	return nil, false
}

func (m *MetaInfo) FullColumnAlias(colName string) string {
	return m.TableName + "." + colName
}

func (m *MetaInfo) OneToOneRelations() []*OneToOne {
	var result []*OneToOne
	for _, relation := range m.Relations {
		if rel, ok := relation.(*OneToOne); ok {
			result = append(result, rel)
		}
	}
	return result
}

func (m *MetaInfo) OneToManyRelations() []*OneToMany {
	var result []*OneToMany
	for _, relation := range m.Relations {
		if r, ok := relation.(*OneToMany); ok {
			result = append(result, r)
		}
	}
	return result
}

func (m *MetaInfo) ManyToManyRelations() []*ManyToMany {
	var result []*ManyToMany
	for _, relation := range m.Relations {
		if rel, ok := relation.(*ManyToMany); ok {
			result = append(result, rel)
		}
	}
	return result
}

func extractPkStrategy(tag *parsedTag) PkStrategy {
	strategy, _ := tag.getProperty("pk")
	switch strategy.val {
	case "auto":
		return Auto
	case "manual":
		return Manual
	}

	return 0
}
