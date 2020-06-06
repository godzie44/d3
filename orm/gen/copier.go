package gen

import "C"
import (
	"github.com/godzie44/d3/orm/entity"
	"io"
	"reflect"
	"strings"
	"text/template"
)

type copier struct {
	out     io.Writer
	pkgPath string
	imports map[string]struct{}
}

var copiableType = reflect.TypeOf((*entity.Copiable)(nil)).Elem()

func (c *copier) handle(t reflect.Type) {
	name := t.Name()

	receiver := strings.ToLower(strings.Split(name, "")[0])

	tpl, err := template.New("registrar").Parse(`

func ({{.receiver}} *{{.entity}}) __d3_makeCopier() entity.Copier {
	return func(src interface{}) interface{} {
		srcTyped, ok := src.(*{{.entity}})
		if !ok {
			return fmt.Errorf("invalid entity type")
		}
		
		copy := &{{.entity}}{}
		{{range .simple_fields}}
		copy.{{.}} = srcTyped.{{.}} {{end}}
		{{range .copy_fields_interf}}
		if srcTyped.{{.FieldName}} != nil {
			copy.{{.FieldName}} = srcTyped.{{.FieldName}}.(entity.Copiable).DeepCopy().({{.TypeName}})
		} {{end}}
		{{range .copy_fields_struct}}
		if srcTyped.{{.FieldName}} != nil {
			copy.{{.FieldName}} = srcTyped.{{.FieldName}}.DeepCopy().({{.TypeName}})
		} {{end}}

		return copy
	}
}
`)
	if err != nil {
		return
	}

	var fields []string
	var copiableInterfaceFields, copiableStructFields []struct{ FieldName, TypeName string }
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Implements(copiableType) {
			typeName, pkgName := extractTypeAndPackageName(t.Field(i).Type, c.pkgPath)
			if pkgName != "" && pkgName != c.pkgPath {
				c.imports[pkgName] = struct{}{}
			}

			if t.Field(i).Type.Kind() == reflect.Interface {
				copiableInterfaceFields = append(copiableInterfaceFields, struct{ FieldName, TypeName string }{FieldName: t.Field(i).Name, TypeName: typeName})
			} else {
				copiableStructFields = append(copiableStructFields, struct{ FieldName, TypeName string }{FieldName: t.Field(i).Name, TypeName: typeName})
			}
		} else {
			fields = append(fields, t.Field(i).Name)
		}
	}

	if err := tpl.Execute(c.out, map[string]interface{}{
		"receiver":           receiver,
		"entity":             name,
		"simple_fields":      fields,
		"copy_fields_interf": copiableInterfaceFields,
		"copy_fields_struct": copiableStructFields,
	}); err != nil {
		return
	}
}

func (c *copier) preamble() []string {
	var result []string
	for imp := range c.imports {
		result = append(result, imp)
	}
	return result
}
