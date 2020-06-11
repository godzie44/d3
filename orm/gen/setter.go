package gen

import (
	"database/sql"
	"io"
	"reflect"
	"strings"
	"text/template"
)

type setter struct {
	out     io.Writer
	imports map[string]struct{}
	pkgPath string
}

var scannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

func (s *setter) preamble() []string {
	var result []string
	for imp := range s.imports {
		result = append(result, imp)
	}
	return result
}

func (s *setter) handle(t reflect.Type) {
	name := t.Name()

	receiver := strings.ToLower(strings.Split(name, "")[0])

	tpl, err := template.New("setter").Parse(`

func ({{.receiver}} *{{.entity}}) __d3_makeFieldSetter() entity.FieldSetter {
	return func(s interface{}, name string, val interface{}) error {
		eTyped, ok := s.(*{{.entity}})
		if !ok {
			return fmt.Errorf("invalid entity type")
		}
		
		switch name { {{range .fields}}
		case "{{.FieldName}}":
			eTyped.{{.FieldName}} = val.({{.TypeName}})
			return nil {{end}}
		{{range .custom_type_fields}}
		case "{{.FieldName}}":
			eTyped.{{.FieldName}} = {{.CustomTypeName}}(val.({{.TypeName}}))
			return nil {{end}}
		{{range .scanner_fields}}
		case "{{.FieldName}}":
			if valuer, isValuer := val.(driver.Valuer); isValuer {
				v, err := valuer.Value()
				if err != nil {
					return eTyped.{{.FieldName}}.Scan(nil)
				} 
				return eTyped.{{.FieldName}}.Scan(v)
			}
			return eTyped.{{.FieldName}}.Scan(val) {{end}}
		default:
			return fmt.Errorf("field %s not found", name)
		}
	}
}
`)
	if err != nil {
		return
	}

	var fields, scannerFields []struct {
		FieldName, TypeName string
	}
	var customTypeFields []struct {
		FieldName, TypeName, CustomTypeName string
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		typeName, pkgName := extractTypeAndPackageName(field.Type, s.pkgPath)
		kind := field.Type.Kind()
		if reflect.PtrTo(field.Type).Implements(scannerType) {
			scannerFields = append(scannerFields, struct{ FieldName, TypeName string }{FieldName: field.Name, TypeName: typeName})
		} else {
			if pkgName != "" && pkgName != s.pkgPath {
				s.imports[pkgName] = struct{}{}
			}

			if kind != reflect.Ptr && kind != reflect.Struct && kind != reflect.Interface && kind.String() != typeName {
				customTypeFields = append(customTypeFields, struct{ FieldName, TypeName, CustomTypeName string }{FieldName: field.Name, TypeName: kind.String(), CustomTypeName: typeName})
			} else {
				fields = append(fields, struct{ FieldName, TypeName string }{FieldName: field.Name, TypeName: typeName})
			}
		}
	}

	if len(scannerFields) > 0 {
		s.imports["database/sql/driver"] = struct{}{}
	}

	if err := tpl.Execute(s.out, map[string]interface{}{
		"receiver": receiver, "entity": name, "fields": fields, "scanner_fields": scannerFields, "custom_type_fields": customTypeFields,
	}); err != nil {
		return
	}
}
