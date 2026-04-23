package service

import (
	"encoding/json"
	"go/format"
	"strconv"
	"strings"
	"text/template"

	gentemplates "github.com/lvjiaben/goweb-scaffold/internal/gen/templates"
)

func renderTemplate(templatePath string, data any) ([]byte, error) {
	content, err := gentemplates.Render(templatePath, data, template.FuncMap{
		"quote":            Quote,
		"json":             MarshalIndent,
		"js":               toJSLiteral,
		"optionFuncName":   optionFuncName,
		"fieldOptions":     fieldOptions,
		"formComponent":    formComponent,
		"searchComponent":  searchComponent,
		"columnAlign":      columnAlign,
		"columnWidth":      columnWidth,
		"columnMinWidth":   columnMinWidth,
		"labelPlaceholder": labelPlaceholder,
		"taggableField":    taggableField,
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

func toJSLiteral(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "null"
	}
	return string(raw)
}

func optionFuncName(modulePascal string, field templateField) string {
	return "get" + modulePascal + GoFieldName(field.ColumnName) + "Options"
}

func fieldOptions(field templateField) []FieldOption {
	if len(field.Options) > 0 {
		return append([]FieldOption{}, field.Options...)
	}
	if field.IsBoolean {
		return []FieldOption{
			{Label: "启用", Value: 1},
			{Label: "禁用", Value: 0},
		}
	}
	return nil
}

func formComponent(field templateField) string {
	switch field.Component {
	case "select":
		if len(fieldOptions(field)) > 0 && isStatusLike(field.ColumnName) {
			return "RadioGroup"
		}
		return "Select"
	case "radio":
		return "RadioGroup"
	case "switch":
		return "Switch"
	case "number-input":
		return "InputNumber"
	case "textarea":
		return "Textarea"
	default:
		return "Input"
	}
}

func searchComponent(field templateField) string {
	switch field.SearchComponent {
	case "select":
		return "Select"
	case "datetime-range":
		return "RangePicker"
	default:
		if field.IsTimestamp {
			return "RangePicker"
		}
		return "Input"
	}
}

func columnAlign(field templateField) string {
	if field.IsPrimaryKey || field.IsInteger || field.IsBigInteger || field.IsTimestamp || taggableField(field) {
		return "center"
	}
	return "left"
}

func columnWidth(field templateField) int {
	if width := parseWidth(field.Width); width > 0 {
		return width
	}
	switch {
	case field.IsPrimaryKey:
		return 80
	case field.IsTimestamp:
		return 180
	case taggableField(field):
		return 120
	case field.IsInteger || field.IsBigInteger:
		return 100
	default:
		return 160
	}
}

func columnMinWidth(field templateField) int {
	if width := parseWidth(field.Width); width > 0 {
		return width
	}
	switch {
	case field.IsPrimaryKey:
		return 80
	case field.IsTimestamp:
		return 180
	case taggableField(field):
		return 120
	default:
		return 140
	}
}

func labelPlaceholder(field templateField) string {
	if strings.TrimSpace(field.Placeholder) != "" {
		return strings.TrimSpace(field.Placeholder)
	}
	switch formComponent(field) {
	case "RadioGroup", "Select", "Switch":
		return "请选择" + field.Label
	default:
		return "请输入" + field.Label
	}
}

func taggableField(field templateField) bool {
	return len(fieldOptions(field)) > 0
}

func parseWidth(raw string) int {
	value := strings.TrimSpace(strings.TrimSuffix(raw, "px"))
	if value == "" {
		return 0
	}
	width, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return width
}

func isStatusLike(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	return lower == "status" || strings.HasSuffix(lower, "_status") || lower == "state" || strings.HasSuffix(lower, "_state")
}
