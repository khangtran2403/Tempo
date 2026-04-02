package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"path/filepath"
)

func TransformData(
	templateStr string,
	data map[string]interface{},
) (string, error) {
	tmpl, err := template.New("transform").
		Funcs(getTemplateFunctions()).
		Parse(templateStr)

	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"title":   cases.Title(language.Und).String,
		"trim":    strings.TrimSpace,
		"replace": strings.ReplaceAll,
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"now": func() time.Time {
			return time.Now()
		},
		"formatDate": func(format string, t time.Time) string {
			return t.Format(format)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
		"base": func(path string) string {
			return filepath.Base(path)
		},
	}
}
