package templates

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"text/template"
)

//go:embed backend/*.tmpl admin_frontend/*.tmpl
var FS embed.FS

func Render(path string, data any, funcMap template.FuncMap) ([]byte, error) {
	name := filepath.Base(path)
	tpl, err := template.New(name).Funcs(funcMap).ParseFS(FS, path)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", path, err)
	}

	var buffer bytes.Buffer
	if err := tpl.ExecuteTemplate(&buffer, name, data); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", path, err)
	}
	return buffer.Bytes(), nil
}
