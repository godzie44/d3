package parser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/godzie44/d3/orm/entity"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	entityAnnotation = "d3:entity"
	tableAnnotation  = "d3_table:"
	indexAnnotation  = "d3_index:"
	uniqueAnnotation = "d3_index_unique:"
)

type Parser struct {
	PkgPath string
	PkgName string
	Metas   []EntityMeta
}

type EntityMeta struct {
	Name      string
	TableName string
	Indexes   []entity.Index
}

func (p *Parser) needProcess(comments string) bool {
	for _, v := range strings.Split(comments, "\n") {
		if strings.HasPrefix(v, entityAnnotation) {
			return true
		}
	}
	return false
}

func (p *Parser) extractTableName(comments string) string {
	for _, v := range strings.Split(comments, "\n") {
		if strings.HasPrefix(v, tableAnnotation) {
			return strings.TrimSpace(strings.TrimPrefix(v, tableAnnotation))
		}
	}
	return ""
}

func (p *Parser) extractIndexes(comments string, annotation string) []entity.Index {
	var result []entity.Index
	for _, v := range strings.Split(comments, "\n") {
		if strings.HasPrefix(v, annotation) {
			indexDec := strings.TrimSpace(strings.TrimPrefix(v, annotation))

			obIndex := strings.Index(indexDec, "(")
			cbIndex := strings.Index(indexDec, ")")

			if obIndex == -1 || cbIndex == -1 {
				continue
			}

			cols := strings.Split(indexDec[obIndex+1:cbIndex], ",")
			for i := range cols {
				cols[i] = strings.TrimSpace(cols[i])
			}

			result = append(result, entity.Index{
				Name:    indexDec[:obIndex],
				Columns: cols,
				Unique:  annotation == uniqueAnnotation,
			})
		}
	}
	return result
}

func (p *Parser) Visit(n ast.Node) (w ast.Visitor) {
	switch n := n.(type) {
	case *ast.Package:
		return p
	case *ast.File:
		p.PkgName = n.Name.String()
		return p

	case *ast.GenDecl:
		if p.needProcess(n.Doc.Text()) {
			for _, nc := range n.Specs {
				switch nct := nc.(type) {
				case *ast.TypeSpec:
					nct.Doc = n.Doc
				}
			}
		}

		return p
	case *ast.TypeSpec:
		if !p.needProcess(n.Doc.Text()) {
			return nil
		}

		p.Metas = append(p.Metas, EntityMeta{
			Name:      n.Name.String(),
			TableName: p.extractTableName(n.Doc.Text()),
			Indexes:   append(p.extractIndexes(n.Doc.Text(), indexAnnotation), p.extractIndexes(n.Doc.Text(), uniqueAnnotation)...),
		})
		return nil
	}
	return nil
}

func (p *Parser) Parse(fileName string) error {
	var err error
	if p.PkgPath, err = getPkgPath(fileName); err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Walk(p, f)
	return nil
}

func getPkgPath(fname string) (string, error) {
	if !filepath.IsAbs(fname) {
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		fname = filepath.Join(pwd, fname)
	}

	goModPath, _ := goModPath(fname)
	if strings.Contains(goModPath, "go.mod") {
		pkgPath, err := getPkgPathFromGoMod(fname, goModPath)
		if err != nil {
			return "", err
		}

		return pkgPath, nil
	}

	return "", errors.New("only mod")
}

// empty if no go.mod, GO111MODULE=off or go without go modules support
func goModPath(fname string) (string, error) {
	root := filepath.Dir(fname)

	cmd := exec.Command("go", "env", "GOMOD")
	cmd.Dir = root

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(stdout)), nil
}

func getPkgPathFromGoMod(fname string, goModPath string) (string, error) {
	modulePath := getModulePath(goModPath)
	if modulePath == "" {
		return "", fmt.Errorf("cannot determine module path from %s", goModPath)
	}

	rel := path.Join(modulePath, filePathToPackagePath(strings.TrimPrefix(fname, filepath.Dir(goModPath))))

	return path.Dir(rel), nil
}

func getModulePath(goModPath string) string {
	data, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return ""
	}

	return ModulePath(data)
}

func filePathToPackagePath(path string) string {
	return filepath.ToSlash(path)
}
