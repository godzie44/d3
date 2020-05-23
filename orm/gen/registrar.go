package gen

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"
)

type CodeGenerator struct {
	out        io.Writer
	extractGen *extractor
}

func NewGenerator(out io.Writer) *CodeGenerator {
	return &CodeGenerator{out: out, extractGen: &extractor{out: out}}
}

func (r *CodeGenerator) WritePreamble() {
	fmt.Fprintf(r.out, "import \"fmt\"\n")
	fmt.Fprintf(r.out, "import \"d3/orm/entity\"\n")
}

func (r *CodeGenerator) Run(t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	name := t.Name()

	receiverName := strings.ToLower(strings.Split(name, "")[0])

	tpl, err := template.New("registrar").Parse(`

func ({{.receiver}} *{{.entity}}) D3Token() entity.MetaToken {
	return entity.MetaToken{
		Tpl: (*{{.entity}})(nil),
		TableName: "",
		Tools: entity.InternalTools{
			FieldExtractor: {{.receiver}}.__d3_createFieldExtractor(),
		},
	}
}
`)
	if err != nil {
		return
	}

	if err := tpl.Execute(r.out, map[string]interface{}{"receiver": receiverName, "entity": t.Name()}); err != nil {
		return
	}

	r.extractGen.run(t)
}
