package gen

import (
	"reflect"
	"strings"
)

func extractTypeAndPackageName(t reflect.Type, currPkgName string) (string, string) {
	var isPtr bool
	if isPtr = t.Kind() == reflect.Ptr; isPtr {
		t = t.Elem()
	}

	var name = t.Name()

	var pkgPath = t.PkgPath()
	if pkgPath != "" && pkgPath != currPkgName {
		pkgNameAt := strings.LastIndex(pkgPath, "/")
		name = pkgPath[pkgNameAt+1:] + "." + name
	}

	if isPtr {
		name = "*" + name
	}

	return name, pkgPath
}
