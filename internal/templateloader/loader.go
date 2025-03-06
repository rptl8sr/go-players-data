package templateloader

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// templatesDirDefault defines the default directory name where template files are stored if no other directory is specified.
const (
	templatesDirDefault = "templates"
)

// Loader is a struct that manages the loading of templates from a specified directory.
type Loader struct {
	templatesDir string
}

// New initializes a Loader instance with the provided template directories
// or a default directory if none are specified.
// Returns an error if the specified directory does not exist.
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

// Load loads a template by name from the loader's templates directory and applies the given template functions.
// Returns the parsed template or an error if the file is not found or cannot be parsed.
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
