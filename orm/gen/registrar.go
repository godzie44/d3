package gen

import (
	"bytes"
	"fmt"
	"github.com/godzie44/d3/orm/entity"
	"io"
	"reflect"
	"strings"
	"text/template"
)

type CodeGenerator struct {
	out             io.Writer
	tempBuffer      *bytes.Buffer
	extractGen      *extractor
	instantiatorGen *instantiator
	setterGen       *setter
	copierGen       *copier
	comparatorGen   *comparator
	pkgPath         string
}

func NewGenerator(out io.Writer, packagePath string) *CodeGenerator {
	tmpBuff := &bytes.Buffer{}
	return &CodeGenerator{
		out:             out,
		tempBuffer:      tmpBuff,
		extractGen:      &extractor{tmpBuff},
		instantiatorGen: &instantiator{tmpBuff},
		setterGen:       &setter{out: tmpBuff, imports: map[string]struct{}{}, pkgPath: packagePath},
		copierGen:       &copier{out: tmpBuff, imports: map[string]struct{}{}, pkgPath: packagePath},
		comparatorGen:   &comparator{out: tmpBuff},
		pkgPath:         packagePath,
	}
}

func (r *CodeGenerator) commonPreamble() []string {
	return []string{
		"fmt",
		"github.com/godzie44/d3/orm/entity",
	}
}

func (r *CodeGenerator) Prepare(t reflect.Type, tableName string, indexes ...entity.Index) {
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
		TableName: "{{.table}}",
		Tools: entity.InternalTools{
			ExtractField: {{.receiver}}.__d3_makeFieldExtractor(),
			SetFieldVal: {{.receiver}}.__d3_makeFieldSetter(),
			CompareFields: {{.receiver}}.__d3_makeComparator(),
			NewInstance: {{.receiver}}.__d3_makeInstantiator(),
			Copy: {{.receiver}}.__d3_makeCopier(),
		},
		Indexes: []entity.Index{
			{{range .indexes}}
			{Name: "{{.Name}}", Columns: []string{ {{range .Columns}} "{{.}}", {{end}} }, Unique: {{.Unique}} },
			{{end}}
		},
	}
}
`)
	if err != nil {
		return
	}

	if err := tpl.Execute(r.tempBuffer, map[string]interface{}{
		"receiver": receiverName,
		"entity":   name,
		"table":    tableName,
		"indexes":  indexes,
	}); err != nil {
		return
	}

	r.extractGen.handle(t)
	r.instantiatorGen.handle(t)
	r.setterGen.handle(t)
	r.copierGen.handle(t)
	r.comparatorGen.handle(t)
}

func (r *CodeGenerator) Write() {
	var imports = map[string]struct{}{}
	for _, imp := range append(r.commonPreamble(), append(r.copierGen.preamble(), r.setterGen.preamble()...)...) {
		if imp == r.pkgPath {
			continue
		}
		imports[imp] = struct{}{}
	}

	for imp := range imports {
		fmt.Fprintf(r.out, "import \"%s\"\n", imp)
	}
	_, _ = r.out.Write(r.tempBuffer.Bytes())
}
