package entity

import (
	"reflect"
	"strings"
)

type property struct {
	name        string
	val         string
	subProperty map[string]property
}

func (p *property) getSubPropVal(name string) string {
	property, exists := p.subProperty[name]
	if !exists {
		return ""
	}

	return property.val
}

type parsedTag struct {
	properties map[string]property
}

func (p *parsedTag) hasProperty(name string) bool {
	_, exists := p.properties[name]
	return exists
}

func (p *parsedTag) getProperty(name string) (property, bool) {
	property, exists := p.properties[name]
	return property, exists
}

// one_to_one:"target_entity:Address, join_column:address_id, reference:id"
func parseTag(tag reflect.StructTag) *parsedTag {
	result := &parsedTag{
		properties: make(map[string]property),
	}

	tagVal, ok := tag.Lookup("d3")
	if !ok {
		return result
	}

	for name, val := range extractKVFromTag(tagVal) {
		d3Property := property{
			name:        name,
			val:         val,
			subProperty: make(map[string]property),
		}

		for subName, subVal := range extractKVFromVal(val) {
			d3Property.subProperty[subName] = property{
				name: subName,
				val:  subVal,
			}
		}

		result.properties[name] = d3Property
	}

	return result
}

func extractKVFromTag(tag string) map[string]string {
	result := make(map[string]string)

	var i int
	for tag != "" {
		for i < len(tag) && tag[i] != ':' {
			i++
		}
		if i >= len(tag) {
			break
		}

		name := strings.Trim(tag[:i], " :,")
		tag = tag[i+1:]
		i = 0

		for i < len(tag) && tag[i] != '<' {
			i++
		}
		if i >= len(tag) {
			result[name] = strings.Trim(tag[:i], "\" ,")
			break
		}

		tag = tag[i+1:]
		i = 0
		for i < len(tag) && tag[i] != '>' {
			i++
		}

		result[name] = strings.Trim(tag[:i], "\" ,")
		if i >= len(tag) {
			break
		}

		tag = tag[i+1:]
		i = 0
	}

	return result
}

func extractKVFromVal(tag string) map[string]string {
	result := make(map[string]string)

	var i int
	for tag != "" {
		for i < len(tag) && tag[i] != ':' {
			i++
		}
		if i >= len(tag) {
			break
		}

		name := strings.Trim(tag[:i], " :")
		tag = tag[i+1:]
		i = 0

		for i < len(tag) && tag[i] != ',' {
			i++
		}
		result[name] = strings.Trim(tag[:i], " ")

		if i >= len(tag) {
			break
		}
		tag = tag[i+1:]
		i = 0
	}

	return result
}
