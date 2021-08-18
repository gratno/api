package main

import (
	"bytes"
	"io/ioutil"
)

type global struct {
	structs  []Struct
	filename string
}

func (g *global) GenerateFile() error {
	buf := bytes.Buffer{}
	buf.WriteString(g.Package())
	buf.WriteString("\n")
	buf.WriteString("import (\n")
	for _, v := range g.Imports() {
		buf.WriteString("\""+v+"\"\n")
	}
	buf.WriteString("\n")
	buf.WriteString(g.Body())
	return ioutil.WriteFile(g.filename, buf.Bytes(), 0644)
}

func (g *global) Package() string {
	return "package main"
}

func (g *global) Imports() []string {
	results := make([]string, len(g.structs))
	for i, v := range g.structs {
		results[i] = v.Import()
	}
	return results
}

func (g *global) Body() string {
	return `var global = []interface{}{
       {{range .Globals}}new({{.}}),
       {{end}}
}`
}

