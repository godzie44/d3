package entity

import (
	d3reflect "d3/reflect"
	"errors"
	"reflect"
	"regexp"
	"strings"
)

type MetaInfo struct {
	Tpl         interface{}
	Fields      map[string]*FieldInfo
	TableName   string
	RelatedMeta map[Name]*MetaInfo
	EntityName  Name
}

type FieldInfo struct {
	Name           string
	Relation       Relation
	associatedType interface{}
	DbAlias        string
	Pk             bool
}

func (f *FieldInfo) IsRelation() bool {
	return f.Relation != nil
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

	fields := make(map[string]*FieldInfo, eType.NumField())
	for i := 0; i < eType.NumField(); i++ {
		fieldReflection := eType.Field(i)

		if fieldReflection.Name == "entity" {
			continue
		}

		tag := parseTag(fieldReflection.Tag)
		relation := extractRelation(tag)

		var dbAlias string
		if relation != nil {
			switch rel := relation.(type) {
			case *OneToOne:
				dbAlias = rel.JoinColumn
			case *OneToMany:
			}
		} else {
			dbAlias = extractDbFieldAlias(tag, fieldReflection.Name)
		}

		fields[fieldReflection.Name] = &FieldInfo{
			Relation:       relation,
			associatedType: fieldReflection.Type,
			Name:           fieldReflection.Name,
			DbAlias:        dbAlias,
			Pk:             tag.hasProperty("pk"),
		}
	}

	return &MetaInfo{
		Tpl:         e,
		Fields:      fields,
		TableName:   tableName,
		RelatedMeta: make(map[Name]*MetaInfo),
		EntityName:  Name(d3reflect.FullName(eType)),
	}, nil
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

func extractRelation(tag *parsedTag) Relation {
	if tag == nil {
		return nil
	}

	var relTypeAlias string
	relType, exists := tag.getProperty("type")
	if !exists {
		relTypeAlias = "lazy"
	} else {
		relTypeAlias = relType.val
	}

	if prop, exists := tag.getProperty("one_to_one"); exists {
		return &OneToOne{
			baseRelation:    baseRelation{RelType: relTypeAlias, TargetEntity: Name(prop.getSubPropVal("target_entity"))},
			JoinColumn:      prop.getSubPropVal("join_on"),
			ReferenceColumn: prop.getSubPropVal("reference_on"),
		}
	}

	if prop, exists := tag.getProperty("one_to_many"); exists {
		return &OneToMany{
			baseRelation:    baseRelation{RelType: relTypeAlias, TargetEntity: Name(prop.getSubPropVal("target_entity"))},
			JoinColumn:      prop.getSubPropVal("join_on"),
			ReferenceColumn: prop.getSubPropVal("reference_on"),
		}
	}

	if prop, exists := tag.getProperty("many_to_many"); exists {
		return &ManyToMany{
			baseRelation:    baseRelation{RelType: relTypeAlias, TargetEntity: Name(prop.getSubPropVal("target_entity"))},
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

func (m *MetaInfo) DependencyEntities() []Name {
	var dependencies []Name

	for _, field := range m.Fields {
		if field.Relation != nil {
			dependencies = append(dependencies, field.Relation.RelatedWith())
		}
	}

	return dependencies
}

func (m *MetaInfo) dependenciesIsSet() bool {
	return len(m.RelatedMeta) == len(m.DependencyEntities())
}

func (m *MetaInfo) PkField() *FieldInfo {
	for i := range m.Fields {
		if m.Fields[i].Pk != false {
			return m.Fields[i]
		}
	}

	panic("pk field must be set")
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

func (m *MetaInfo) FullFieldAlias(field *FieldInfo) string {
	return m.FullColumnAlias(field.DbAlias)
}

func (m *MetaInfo) FullColumnAlias(colName string) string {
	return m.TableName + "." + colName
}
