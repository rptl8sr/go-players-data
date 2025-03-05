package templateloader

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

const (
	templatesDirDefault = "templates"
)

type Loader struct {
	templatesDir string
}

func New(templatesDir ...string) (*Loader, error) {
	if len(templatesDir) == 0 {
		templatesDir = []string{templatesDirDefault}
	}

	if _, err := os.Stat(templatesDir[0]); os.IsNotExist(err) {
		return nil, fmt.Errorf("loader.Must: directory %s", templatesDir)
	}

	return &Loader{
		templatesDir: templatesDir[0],
	}, nil
}

func (t *Loader) Load(name string, funcs template.FuncMap) (*template.Template, error) {
	tmplPath := filepath.Join(t.templatesDir, fmt.Sprintf("%s.tmpl", name))

	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("loader.Must: template file not found: %s", tmplPath)
	}

	tmpl, err := template.New(filepath.Base(tmplPath)).
		Funcs(funcs).
		ParseFiles(tmplPath)

	if err != nil {
		return nil, fmt.Errorf("loader.Must: failed to parse template: %w", err)
	}

	return tmpl, nil
}
