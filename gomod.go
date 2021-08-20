package main

import (
	"go/ast"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type gomod struct {
	Version string
	Module  string
	Root    string
}

func (d *gomod) getRoot() string {
	return d.Root
}

func (d *gomod) getImports(spec *ast.ImportSpec) (alias, absPath, importPath string) {
	root := d.getRoot()
	if spec.Name != nil {
		alias = spec.Name.Name
	}
	importPath = strings.Trim(spec.Path.Value, "\"")
	if strings.Contains(importPath, d.Module) {
		importPath = strings.TrimPrefix(importPath, d.Module)
		absPath = filepath.Join(root, importPath)
		if index := strings.LastIndex(importPath, "/"); index > 0 && alias == "" {
			alias = importPath[index+1:]
		}
	}
	return
}

func (d *gomod) ModulePath(name, version string) (string, error) {
	// first we need GOMODCACHE
	cache, ok := os.LookupEnv("GOMODCACHE")
	if !ok {
		cache = path.Join(os.Getenv("GOPATH"), "pkg", "mod")
	}

	// then we need to escape path
	escapedPath, err := module.EscapePath(name)
	if err != nil {
		return "", err
	}

	// version also
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return "", err
	}

	return path.Join(cache, escapedPath+"@"+escapedVersion), nil
}

func (d *gomod) getGoModule() (*gomod, error) {
	b, err := ioutil.ReadFile(filepath.Join(d.getRoot(), "go.mod"))
	if err != nil {
		return nil, err
	}
	return &gomod{
		Module: modfile.ModulePath(b),
	}, nil
}
