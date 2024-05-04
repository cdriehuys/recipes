package templates

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"path"
	"path/filepath"
)

type DiskTemplateEngine struct {
	IncludePath string
	LayoutPath  string

	Logger *slog.Logger
}

func (e *DiskTemplateEngine) Write(w io.Writer, name string, data any) error {
	includes, err := filepath.Glob(path.Join(e.IncludePath, "*.html.tmpl"))
	if err != nil {
		return fmt.Errorf("could not find includes: %w", err)
	}

	templatePath := path.Join(e.LayoutPath, name+".html.tmpl")
	templateFiles := append(includes, templatePath)
	e.Logger.Debug("Collected template files.", "templateFiles", templateFiles)

	tpl, err := template.ParseFiles(templateFiles...)
	if err != nil {
		return fmt.Errorf("failed to parse template files: %w", err)
	}

	if err := tpl.ExecuteTemplate(w, "base", data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
