package entity

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var testCases = map[string]*parsedTag{
	"d3:\"one_to_one:<target_entity:Address, join_column:address_id, reference:id>\"": {properties: map[string]property{
		"one_to_one": {
			name: "one_to_one",
			val:  "target_entity:Address, join_column:address_id, reference:id",
			subProperty: map[string]property{
				"target_entity": {
					name: "target_entity",
					val:  "Address",
				},
				"join_column": {
					name: "join_column",
					val:  "address_id",
				},
				"reference": {
					name: "reference",
					val:  "id",
				},
			},
		},
	}},
	"d3:\"one_to_one: <target_entity:Address , join_column: address_id , reference:  id> \",unknown_tag:\"key1:val1,key2:val2\"": {properties: map[string]property{
		"one_to_one": {
			name: "one_to_one",
			val:  "target_entity:Address , join_column: address_id , reference:  id",
			subProperty: map[string]property{
				"target_entity": {
					name: "target_entity",
					val:  "Address",
				},
				"join_column": {
					name: "join_column",
					val:  "address_id",
				},
				"reference": {
					name: "reference",
					val:  "id",
				},
			},
		},
	}},
	"d3:\"one_to_one:<target_entity:Address,join_column:address_id,reference:id>,many_to_one:<target_entity:User>,type:lazy\"": {properties: map[string]property{
		"one_to_one": {
			name: "one_to_one",
			val:  "target_entity:Address,join_column:address_id,reference:id",
			subProperty: map[string]property{
				"target_entity": {
					name: "target_entity",
					val:  "Address",
				},
				"join_column": {
					name: "join_column",
					val:  "address_id",
				},
				"reference": {
					name: "reference",
					val:  "id",
				},
			},
		},
		"many_to_one": {
			name: "many_to_one",
			val:  "target_entity:User",
			subProperty: map[string]property{
				"target_entity": {
					name: "target_entity",
					val:  "User",
				},
			},
		},
		"type": {
			name:        "type",
			val:         "lazy",
			subProperty: map[string]property{},
		},
	}},
}

func TestTagParsing(t *testing.T) {
	for tag, expectedResult := range testCases {
		result := parseTag(reflect.StructTag(tag))
		assert.Equal(t, expectedResult, result)
	}
}
