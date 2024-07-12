package templates

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
)

//go:embed *.html
var templateFiles embed.FS

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseFS(templateFiles, "*.html"))
}

func Execute(w http.ResponseWriter, name string, data interface{}) error {
	return tmpl.ExecuteTemplate(w, name, data)
}

func Render(name string, data interface{}) (template.HTML, error) {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, name, data)
	return template.HTML(buf.String()), err
}
