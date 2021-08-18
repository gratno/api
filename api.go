package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/mod/modfile"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type Struct interface {
	Import() string
	String() string
}

type _struct struct {
	Source string
	Name   string
}

func (s _struct) Import() string {
	return s.Source
}

func (s _struct) String() string {
	return s.Name
}

type Api struct {
	Module string
	root string
}

func NewApi(root string) *Api {
	p:= &Api{root: root}
	module, err := p.parseGoModule()
	if err != nil {
		panic(err)
	}
	p.Module = module.Module
	fmt.Println(p.Module)
	return p
}

func (p *Api) parseGoModule() (*gomod, error) {
	b, err := ioutil.ReadFile(filepath.Join(p.root, "go.mod"))
	if err != nil {
		return nil, err
	}
	return &gomod{
		Module: modfile.ModulePath(b),
	}, nil
}

func (p *Api) getFiles() ([]string, error) {
	files := make([]string, 0)
	return files, filepath.Walk(p.root, func(path string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
			files = append(files, path)
		}
		return nil
	})
}

func (p *Api) isStruct(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.StructType:
		return true
	}
	return false
}

func (p *Api) getStructsFromFile(filename string) ([]Struct, error) {
	results := make([]Struct, 0)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.DeclarationErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}
	for _, v := range f.Decls {
		switch decl := v.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				t, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if  !p.isStruct(t.Type) {
					continue
				}
				results = append(results, _struct{
					Source: p.Module+"/"+f.Name.Name,
					Name:   t.Name.Name,
				})
			}
		}
	}
	return results, nil
}

type gomod struct {
	Version string
	Module  string
}

func main() {
	p := NewApi("/Users/jianglin/GOPATH/src/github.com/ponzu-cms/ponzu")
	files, err := p.getFiles()
	if err != nil {
		log.Println("getfiles",err)
		return
	}
	for _, file := range files {
		structs, err := p.getStructsFromFile(file)
		if err != nil {
			log.Println("get struct",err)
			return
		}
		for _, v := range structs{
			fmt.Println(v.Import(), v.String())
		}
	}
}
