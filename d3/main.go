package main

import (
	"flag"
	"fmt"
	d3parser "github.com/godzie44/d3/d3/parser"
	"github.com/godzie44/d3/orm/gen/bootstrap"
	"golang.org/x/sync/errgroup"
	"os"
	"path/filepath"
	"strings"
)

var debug = flag.Bool("debug", false, "dont delete temporary files")

func main() {
	flag.Parse()

	g := &errgroup.Group{}
	for _, fileName := range flag.Args() {
		if err := walkAndCreateTask(fileName, g); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err := g.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func walkAndCreateTask(where string, taskPool *errgroup.Group) error {
	return filepath.Walk(where,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && !strings.HasSuffix(path, "_test.go") {
				taskPool.Go(func() error {
					return generate(path)
				})
				return nil
			}

			return nil
		})
}

func generate(fileName string) error {
	p := d3parser.Parser{}
	if err := p.Parse(fileName); err != nil {
		return fmt.Errorf("file parse error %s: %w", fileName, err)
	}

	if len(p.Metas) == 0 {
		return nil
	}

	var outName = strings.TrimSuffix(fileName, ".go") + "_d3.go"

	boot := bootstrap.Generator{
		PkgPath:   p.PkgPath,
		PkgName:   p.PkgName,
		Metas:     p.Metas,
		OutName:   outName,
		BuildTags: "",
		Debug:     *debug,
	}

	println(outName)
	return boot.Run()
}
