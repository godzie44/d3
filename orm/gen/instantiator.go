package gen

import (
	"io"
	"reflect"
	"strings"
	"text/template"
)

type instantiator struct {
	out io.Writer
}

func (i *instantiator) run(t reflect.Type) {
	name := t.Name()

	receiver := strings.ToLower(strings.Split(name, "")[0])

	tpl, err := template.New("instantiator").Parse(`

func ({{.receiver}} *{{.entity}}) __d3_makeInstantiator() entity.Instantiator {
	return func() interface{} {
		return &{{.entity}}{}
	}
}
`)
	if err != nil {
		return
	}

	if err := tpl.Execute(i.out, map[string]interface{}{"receiver": receiver, "entity": name}); err != nil {
		return
	}
}
