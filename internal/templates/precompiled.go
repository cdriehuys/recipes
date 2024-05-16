package templates

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

type FSTemplateEngine struct {
	templates map[string]*template.Template
}

func (e *FSTemplateEngine) Write(w io.Writer, _ *http.Request, name string, data map[string]any) error {
	tpl, ok := e.templates[name]
	if !ok {
		return fmt.Errorf("no template named '%s'", name)
	}

	return tpl.ExecuteTemplate(w, "base", data)
}

func NewFSTemplateEngine(templates fs.FS, funcs template.FuncMap) (FSTemplateEngine, error) {
	includes, err := fs.Glob(templates, path.Join("templates", "includes", "*.html.tmpl"))
	if err != nil {
		return FSTemplateEngine{}, fmt.Errorf("failed to find includes: %w", err)
	}

	layoutTemplates := make(map[string]*template.Template)
	err = fs.WalkDir(templates, "templates/layouts", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".html.tmpl") {
			return nil
		}

		relativeToLayoutDir, err := filepath.Rel("templates/layouts", p)
		if err != nil {
			return fmt.Errorf("could not determine relative path for %s: %w", p, err)
		}

		templateName := strings.TrimSuffix(relativeToLayoutDir, ".html.tmpl")
		templateFiles := append(includes, p)
		tpl := template.New(templateName).Funcs(funcs)
		tpl, err = tpl.ParseFS(templates, templateFiles...)
		if err != nil {
			return fmt.Errorf("failed to parse template for %s: %w", p, err)
		}

		layoutTemplates[templateName] = tpl

		return nil
	})
	if err != nil {
		return FSTemplateEngine{}, fmt.Errorf("failed to compile layout templates: %w", err)
	}

	return FSTemplateEngine{layoutTemplates}, nil
}
