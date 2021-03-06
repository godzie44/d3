// Code generated by d3. DO NOT EDIT.

package query

import "github.com/godzie44/d3/orm/entity"
import "database/sql/driver"
import "fmt"

func (u *User) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tpl:       (*User)(nil),
		TableName: "q_user",
		Tools: entity.InternalTools{
			ExtractField:  u.__d3_makeFieldExtractor(),
			SetFieldVal:   u.__d3_makeFieldSetter(),
			CompareFields: u.__d3_makeComparator(),
			NewInstance:   u.__d3_makeInstantiator(),
			Copy:          u.__d3_makeCopier(),
		},
		Indexes: []entity.Index{},
	}
}

func (u *User) __d3_makeFieldExtractor() entity.FieldExtractor {
	return func(s interface{}, name string) (interface{}, error) {
		sTyped, ok := s.(*User)
		if !ok {
			return nil, fmt.Errorf("invalid entity type")
		}

		switch name {

		case "id":
			return sTyped.id, nil

		case "photos":
			return sTyped.photos, nil

		case "name":
			return sTyped.name, nil

		case "age":
			return sTyped.age, nil

		default:
			return nil, fmt.Errorf("field %s not found", name)
		}
	}
}

func (u *User) __d3_makeInstantiator() entity.Instantiator {
	return func() interface{} {
		return &User{}
	}
}

func (u *User) __d3_makeFieldSetter() entity.FieldSetter {
	return func(s interface{}, name string, val interface{}) error {
		eTyped, ok := s.(*User)
		if !ok {
			return fmt.Errorf("invalid entity type")
		}

		switch name {
		case "photos":
			eTyped.photos = val.(*entity.Collection)
			return nil
		case "name":
			eTyped.name = val.(string)
			return nil
		case "age":
			eTyped.age = val.(int)
			return nil

		case "id":
			if valuer, isValuer := val.(driver.Valuer); isValuer {
				v, err := valuer.Value()
				if err != nil {
					return eTyped.id.Scan(nil)
				}
				return eTyped.id.Scan(v)
			}
			return eTyped.id.Scan(val)
		default:
			return fmt.Errorf("field %s not found", name)
		}
	}
}

func (u *User) __d3_makeCopier() entity.Copier {
	return func(src interface{}) interface{} {
		srcTyped, ok := src.(*User)
		if !ok {
			return fmt.Errorf("invalid entity type")
		}

		copy := &User{}

		copy.id = srcTyped.id
		copy.name = srcTyped.name
		copy.age = srcTyped.age

		if srcTyped.photos != nil {
			copy.photos = srcTyped.photos.DeepCopy().(*entity.Collection)
		}

		return copy
	}
}

func (u *User) __d3_makeComparator() entity.FieldComparator {
	return func(e1, e2 interface{}, fName string) bool {
		if e1 == nil || e2 == nil {
			return e1 == e2
		}

		e1Typed, ok := e1.(*User)
		if !ok {
			return false
		}
		e2Typed, ok := e2.(*User)
		if !ok {
			return false
		}

		switch fName {

		case "id":
			return e1Typed.id == e2Typed.id
		case "photos":
			return e1Typed.photos == e2Typed.photos
		case "name":
			return e1Typed.name == e2Typed.name
		case "age":
			return e1Typed.age == e2Typed.age
		default:
			return false
		}
	}
}

func (p *Photo) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tpl:       (*Photo)(nil),
		TableName: "q_photo",
		Tools: entity.InternalTools{
			ExtractField:  p.__d3_makeFieldExtractor(),
			SetFieldVal:   p.__d3_makeFieldSetter(),
			CompareFields: p.__d3_makeComparator(),
			NewInstance:   p.__d3_makeInstantiator(),
			Copy:          p.__d3_makeCopier(),
		},
		Indexes: []entity.Index{},
	}
}

func (p *Photo) __d3_makeFieldExtractor() entity.FieldExtractor {
	return func(s interface{}, name string) (interface{}, error) {
		sTyped, ok := s.(*Photo)
		if !ok {
			return nil, fmt.Errorf("invalid entity type")
		}

		switch name {

		case "id":
			return sTyped.id, nil

		case "src":
			return sTyped.src, nil

		default:
			return nil, fmt.Errorf("field %s not found", name)
		}
	}
}

func (p *Photo) __d3_makeInstantiator() entity.Instantiator {
	return func() interface{} {
		return &Photo{}
	}
}

func (p *Photo) __d3_makeFieldSetter() entity.FieldSetter {
	return func(s interface{}, name string, val interface{}) error {
		eTyped, ok := s.(*Photo)
		if !ok {
			return fmt.Errorf("invalid entity type")
		}

		switch name {
		case "src":
			eTyped.src = val.(string)
			return nil

		case "id":
			if valuer, isValuer := val.(driver.Valuer); isValuer {
				v, err := valuer.Value()
				if err != nil {
					return eTyped.id.Scan(nil)
				}
				return eTyped.id.Scan(v)
			}
			return eTyped.id.Scan(val)
		default:
			return fmt.Errorf("field %s not found", name)
		}
	}
}

func (p *Photo) __d3_makeCopier() entity.Copier {
	return func(src interface{}) interface{} {
		srcTyped, ok := src.(*Photo)
		if !ok {
			return fmt.Errorf("invalid entity type")
		}

		copy := &Photo{}

		copy.id = srcTyped.id
		copy.src = srcTyped.src

		return copy
	}
}

func (p *Photo) __d3_makeComparator() entity.FieldComparator {
	return func(e1, e2 interface{}, fName string) bool {
		if e1 == nil || e2 == nil {
			return e1 == e2
		}

		e1Typed, ok := e1.(*Photo)
		if !ok {
			return false
		}
		e2Typed, ok := e2.(*Photo)
		if !ok {
			return false
		}

		switch fName {

		case "id":
			return e1Typed.id == e2Typed.id
		case "src":
			return e1Typed.src == e2Typed.src
		default:
			return false
		}
	}
}
