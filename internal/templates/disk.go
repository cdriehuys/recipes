package templates

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
)

type DiskTemplateEngine struct {
	IncludePath string
	LayoutPath  string

	FuncMap template.FuncMap

	Logger *slog.Logger
}

func (e *DiskTemplateEngine) Write(w io.Writer, _ *http.Request, name string, data map[string]any) error {
	includes, err := filepath.Glob(path.Join(e.IncludePath, "*.html.tmpl"))
	if err != nil {
		return fmt.Errorf("could not find includes: %w", err)
	}

	templatePath := path.Join(e.LayoutPath, name+".html.tmpl")
	templateFiles := append(includes, templatePath)
	e.Logger.Debug("Collected template files.", "templateFiles", templateFiles)

	tpl := template.New(name).Funcs(e.FuncMap)

	tpl, err = tpl.ParseFiles(templateFiles...)
	if err != nil {
		return fmt.Errorf("failed to parse template files: %w", err)
	}

	if err := tpl.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
