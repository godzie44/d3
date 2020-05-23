package gen

import (
	"io"
	"reflect"
	"strings"
	"text/template"
)

type extractor struct {
	out io.Writer
}

func (e *extractor) run(t reflect.Type) {
	name := t.Name()

	receiver := strings.ToLower(strings.Split(name, "")[0])

	tpl, err := template.New("registrar").Parse(`

func ({{.receiver}} *{{.entity}}) __d3_createFieldExtractor() entity.FieldExtractor {
	return func(s interface{}, name string) (interface{}, error) {
		sTyped, ok := s.(*{{.entity}})
		if !ok {
			return nil, fmt.Errorf("invalid entity type")
		}
		
		switch name {
		{{range .fields}}
		case "{{.}}":
			return sTyped.{{.}}, nil
		{{end}}
		default:
			return nil, fmt.Errorf("field %s not found", name)
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

	if err := tpl.Execute(e.out, map[string]interface{}{"receiver": receiver, "entity": t.Name(), "fields": fields}); err != nil {
		return
	}
}
