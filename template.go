package main

import (
	"strings"
	"text/template"
)

const ignoreTmpl = `# gitignore generated by mkignore
# using templates from: https://github.com/github/gitignore

{{range $gi := .}}
# Template: {{ $gi.Name }}
{{ $gi.Data }}

{{end}}`

func BuildIgnoreTemplate() *template.Template {
	return template.Must(template.New("ignore").Parse(ignoreTmpl))
}

func ExecIgnoreTmpl(gis []*IgnoreFile) (string, error) {
	tmpl := BuildIgnoreTemplate()
	var buf strings.Builder
	err := tmpl.Execute(&buf, gis)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
