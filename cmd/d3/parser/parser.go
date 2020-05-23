package parser

import (
	"bytes"
	"errors"
	"fmt"
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
)

type Parser struct {
	PkgPath     string
	PkgName     string
	EntityNames []string
}

func (p *Parser) needProcess(comments string) bool {
	for _, v := range strings.Split(comments, "\n") {
		if strings.HasPrefix(v, entityAnnotation) {
			return true
		}
	}
	return false
}

func (p *Parser) Visit(n ast.Node) (w ast.Visitor) {
	switch n := n.(type) {
	case *ast.Package:
		fmt.Println(n)
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

		p.EntityNames = append(p.EntityNames, n.Name.String())

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
