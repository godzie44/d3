package gen

import (
	"io"
	"reflect"
	"strings"
	"text/template"
)

type comparator struct {
	out io.Writer
}

func (c *comparator) handle(t reflect.Type) {
	name := t.Name()

	receiver := strings.ToLower(strings.Split(name, "")[0])

	tpl, err := template.New("registrar").Parse(`

func ({{.receiver}} *{{.entity}}) __d3_makeComparator() entity.FieldComparator {
	return func(e1, e2 interface{}, fName string) bool {
		if e1 == nil || e2 == nil {
			return e1 == e2
		}

		e1Typed, ok := e1.(*{{.entity}})
		if !ok {
			return false
		}
		e2Typed, ok := e2.(*{{.entity}})
		if !ok {
			return false
		}
		
		switch fName {
		{{range .fields}}
		case "{{.}}":
			return e1Typed.{{.}} == e2Typed.{{.}}{{end}}
		default:
			return false
		}
	}
}
`)
	if err != nil {
		return
	}

	var fields []string
	for i := 0; i < t.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}

	if err := tpl.Execute(c.out, map[string]interface{}{"receiver": receiver, "entity": name, "fields": fields}); err != nil {
		return
	}
}
