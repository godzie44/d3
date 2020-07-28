package entity

import (
	"reflect"
	"strings"
)

type Name string

func NameFromEntity(e interface{}) Name {
	t := reflect.TypeOf(e)
	switch t.Kind() {
	case reflect.Ptr:
		return Name(t.Elem().PkgPath() + "/" + t.Elem().Name())
	default:
		return Name(t.PkgPath() + "/" + t.Name())
	}
}

func nameFromTag(tag string, parentName Name) Name {
	defined := Name(tag)
	if defined.IsShort() {
		return parentName.Combine(defined)
	}

	return defined
}

func (n Name) Short() string {
	path := strings.Split(string(n), "/")

	return path[len(path)-1]
}

func (n Name) IsShort() bool {
	return !strings.Contains(string(n), "/")
}

func (n Name) Equal(name Name) bool {
	return n == name
}

func (n Name) Combine(entity Name) Name {
	path := strings.Split(string(n), "/")

	return Name(strings.Join(append(path[:len(path)-1], entity.Short()), "/"))
}
