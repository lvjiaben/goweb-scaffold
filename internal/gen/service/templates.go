package service

import (
	"go/format"
	"strings"
	"text/template"

	gentemplates "github.com/lvjiaben/goweb-scaffold/internal/gen/templates"
)

func renderTemplate(templatePath string, data any) ([]byte, error) {
	content, err := gentemplates.Render(templatePath, data, template.FuncMap{
		"quote": Quote,
	})
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(templatePath, "backend/") {
		formatted, err := format.Source(content)
		if err == nil {
			return formatted, nil
		}
	}
	return content, nil
}
