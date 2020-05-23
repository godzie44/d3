package main

import (
	d3parser "d3/cmd/d3/parser"
	"d3/orm/gen/bootstrap"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var debug = flag.Bool("debug", false, "dont delete temporary files")

func main() {
	flag.Parse()

	for _, fileName := range flag.Args() {
		if err := walkAndGenerate(fileName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func walkAndGenerate(where string) error {
	return filepath.Walk(where,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && !strings.HasSuffix(path, "_test.go") {
				return generate(path)
			}

			return nil
		})
}

func generate(fileName string) error {
	p := d3parser.Parser{}
	if err := p.Parse(fileName); err != nil {
		return fmt.Errorf("file parse error %s: %w", fileName, err)
	}

	if len(p.EntityNames) == 0 {
		return nil
	}

	var outName = strings.TrimSuffix(fileName, ".go") + "_d3.go"

	boot := bootstrap.Generator{
		PkgPath:   p.PkgPath,
		PkgName:   p.PkgName,
		Entities:  p.EntityNames,
		OutName:   outName,
		BuildTags: "",
		Debug:     *debug,
	}

	println(outName)
	return boot.Run()
}
