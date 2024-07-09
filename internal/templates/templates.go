package templates

import (
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
