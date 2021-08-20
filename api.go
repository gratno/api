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
	"os"
	"path/filepath"
	"strings"
)

type Struct interface {
	Import() string
	String() string
	AbsPath() string
	References() []Struct
}

type structPath struct {
	absPath    string
	importPath string
}

type _struct struct {
	imports map[string]structPath
	Source  string
	Name    string
	path    string
	structs []Struct
}

func (s _struct) Import() string {
	return s.Source
}

func (s _struct) String() string {
	return s.Name
}
func (s _struct) AbsPath() string {
	return s.path
}

func (s _struct) References() []Struct {
	return s.structs
}

type Api struct {
	gomod   *gomod
	root    string
	Structs []Struct
}

func NewApi(root string) *Api {
	p := &Api{root: root}
	p.gomod = p.parseGoModule()
	return p
}

func (p *Api) parseGoModule() *gomod {
	b, err := ioutil.ReadFile(filepath.Join(p.root, "go.mod"))
	if err != nil {
		panic(err)
	}
	return &gomod{
		Module: modfile.ModulePath(b),
		Root:   p.root,
	}
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

func (p *Api) joinImport(filename string) string {
	mstrs := strings.Split(p.gomod.Module, "/")
	filename = strings.TrimPrefix(filename, p.root)
	filename = strings.TrimPrefix(filename, string(os.PathSeparator))
	mstrs = append(mstrs, strings.Split(filename, string(os.PathSeparator))...)
	mstrs = mstrs[:len(mstrs)-1]
	return strings.Join(mstrs, "/")
}

type visitorFunc func(node ast.Node) (w ast.Visitor)

func (f visitorFunc) Visit(node ast.Node) (w ast.Visitor) {
	return f(node)
}

func (p *Api) visit(packageAlias, path string) visitorFunc {
	imports := make(map[string]structPath)
	basePath := filepath.Base(path)
	return func(node ast.Node) (w ast.Visitor) {
		switch expr := node.(type) {
		case *ast.GenDecl:
		case *ast.ImportSpec:
			alias, absPath, importPath := p.gomod.getImports(expr)
			imports[alias] = structPath{absPath: absPath, importPath: importPath}
		case *ast.TypeSpec:
			xexpr, ok := expr.Type.(*ast.StructType)
			if !ok {
				return
			}
			firstRune := expr.Name.Name[0]
			if firstRune >= 'A' && firstRune <= 'Z' {
				s := _struct{
					Source: p.joinImport(path),
					Name:   expr.Name.Name,
					path:   basePath,
				}
				p.Structs = append(p.Structs)
				for _, field := range xexpr.Fields.List {
					if filedType := p.getExprType(field.Type); strings.ContainsRune(filedType, '.') {
						strs := strings.Split(filedType, ".")
						s.structs = append(s.structs, _struct{
							imports: s.imports,
							Source:  s.imports[strs[0]].importPath,
							Name:    strs[1],
							path:    s.imports[strs[0]].absPath,
						})
					}
				}
			}
		}
		return p.visit(packageAlias, path)
	}

}

func (p *Api) getStructsFromFile(path string) ([]Struct, error) {
	results := make([]Struct, 0)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.DeclarationErrors|parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if f.Name.Name == "main" {
		return nil, nil
	}
	ast.Walk(p.visit(f.Name.Name, path), f)
	return results, nil
}

func (p *Api) getExprType(field ast.Expr) string {
	var fieldType string
	switch expr := field.(type) {
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType, *ast.ChanType:
	case *ast.ArrayType:
		return "[]" + p.getExprType(expr.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", p.getExprType(expr.Key), p.getExprType(expr.Value))
	case *ast.SelectorExpr:
		return expr.X.(*ast.Ident).Name + "." + expr.Sel.Name
	case *ast.Ident:
		fieldType = expr.Name
	case *ast.StarExpr:
		return p.getExprType(expr.X)
	default:
		fieldType = fmt.Sprintf("%v", field)
	}
	return fieldType
}

func main() {
	p := NewApi("D:\\GolandProjects\\awesomeProject")
	files, err := p.getFiles()
	if err != nil {
		log.Println("getfiles", err)
		return
	}
	var results []Struct
	for _, file := range files {
		structs, err := p.getStructsFromFile(file)
		if err != nil {
			log.Println("get struct", err)
			return
		}
		results = append(results, structs...)
	}
	g := global{
		structs:     results,
		filename:    "global_generate.go",
		PackageName: "main",
	}
	fmt.Println(g.GenerateFile())
}
