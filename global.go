package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
)

type global struct {
	structs     []Struct
	filename    string
	PackageName string
}

func (g *global) GenerateFile() error {
	buf := bytes.Buffer{}
	buf.WriteString(g.Package())
	buf.WriteString("\n")
	buf.WriteString("import (\n")
	imports, vars := g.ImportsAndVars()
	for _, v := range imports {
		buf.WriteString(v)
		buf.WriteString("\n")
	}
	buf.WriteString(")\n")
	template.Must(template.New("").Parse(g.Body())).Execute(&buf, map[string]interface{}{
		"Globals": vars,
	})
	return ioutil.WriteFile(g.filename, buf.Bytes(), 0644)
}

func (g *global) Package() string {
	return "package " + g.PackageName
}

func (g *global) ImportsAndVars() (imports []string, vars []string) {
	imports = make([]string, len(g.structs))
	vars = make([]string, len(g.structs))
	for i, v := range g.structs {
		imp := v.Import()
		strs := strings.Split(imp, "/")
		strs = strs[len(strs)-2:]
		alias := strings.Join(strs, "")
		imports[i] = fmt.Sprintf("%s \"%s\"", alias, imp)
		vars[i] = alias + "." + v.String()
	}
	return imports, vars
}

func (g *global) Vars() []string {
	results := make([]string, len(g.structs))
	for i, v := range g.structs {
		results[i] = v.String()
	}
	return results
}

func (g *global) Body() string {
	return `var globals = []interface{}{
       {{range .Globals}}new({{.}}),
       {{end}}
}`
}
