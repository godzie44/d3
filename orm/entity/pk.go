package entity

import (
	d3reflect "d3/reflect"
	"database/sql/driver"
	"reflect"
)

type pk struct {
	Field    *FieldInfo
	Strategy PkStrategy
}

func (p *pk) FullDbAlias() string {
	return p.Field.FullDbAlias
}

type KeyTpl struct {
	values     []reflect.Value
	projection []interface{}
}

func (k *KeyTpl) Projection() []interface{} {
	return k.projection
}

func (k *KeyTpl) Key() []interface{} {
	res := make([]interface{}, len(k.values))
	for i, val := range k.values {
		res[i] = val.Interface()
	}

	return res
}

func (m *MetaInfo) ExtractPkValue(entity interface{}) (interface{}, error) {
	val, err := d3reflect.ExtractStructField(entity, m.Pk.Field.Name)
	if err != nil {
		return nil, err
	}

	if val, ok := val.(driver.Valuer); ok {
		if pk, _ := val.Value(); pk == nil {
			return nil, nil
		}
	}

	return val, nil
}

func (m *MetaInfo) CreateKeyTpl() *KeyTpl {
	tpl := &KeyTpl{
		values:     make([]reflect.Value, 0, 1),
		projection: make([]interface{}, 0, 1),
	}

	value := reflect.New(reflect.PtrTo(m.Pk.Field.associatedType).Elem())

	tpl.values = append(tpl.values, value.Elem())
	tpl.projection = append(tpl.projection, value.Elem().Addr().Interface())

	return tpl
}
