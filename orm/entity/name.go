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

func (n Name) Short() string {
	path := strings.Split(string(n), "/")

	return path[len(path)-1]
}

func (n Name) Equal(name Name) bool {
	//return n == name || n.Short() == name.Short()
	return n == name
}
