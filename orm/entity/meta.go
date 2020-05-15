package entity

import (
	d3reflect "d3/reflect"
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
	AssociatedType reflect.Type
	DbAlias        string
	FullDbAlias    string
}

func CreateMeta(mapping UserMapping) (*MetaInfo, error) {
	eType, err := d3reflect.IntoStructType(reflect.TypeOf(mapping.Entity))
	if err != nil {
		return nil, err
	}

	meta := &MetaInfo{
		Tpl:         mapping.Entity,
		TableName:   mapping.TableName,
		Fields:      make(map[string]*FieldInfo),
		Relations:   make(map[string]Relation),
		RelatedMeta: make(map[Name]*MetaInfo),
		EntityName:  Name(d3reflect.FullName(eType)),
	}

	for i := 0; i < eType.NumField(); i++ {
		fieldReflection := eType.Field(i)

		// skip unexported fields
		if fieldReflection.PkgPath != "" {
			continue
		}

		tag := parseTag(fieldReflection.Tag)

		field := &FieldInfo{
			Name:           fieldReflection.Name,
			AssociatedType: fieldReflection.Type,
			DbAlias:        extractDbFieldAlias(tag, fieldReflection.Name),
		}

		var relation Relation
		switch {
		case tag.hasProperty("one_to_one"):
			relation = &OneToOne{}
		case tag.hasProperty("one_to_many"):
			relation = &OneToMany{}
		case tag.hasProperty("many_to_many"):
			relation = &ManyToMany{}
		}

		if relation != nil {
			relation.fillFromTag(tag)
			relation.setField(field)
			meta.Relations[fieldReflection.Name] = relation
		} else {
			field.FullDbAlias = meta.FullColumnAlias(field.DbAlias)
			meta.Fields[fieldReflection.Name] = field
		}

		if tag.hasProperty("pk") {
			meta.Pk = &pk{field, extractPkStrategy(tag)}
		}
	}

	if meta.Pk == nil {
		panic("pk not found in entity: " + meta.EntityName)
	}

	return meta, nil
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

	prop, exists := tag.getProperty("column")
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
