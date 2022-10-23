package template

import (
	_ "embed"
	"io"
	"text/template"

	"github.com/kokhanevych/gomockgen/internal"
)

//go:embed mock.tmpl
var defaultTemplate string

type data struct {
	Package       internal.Package
	Substitutions map[string]string
}

// Template is the representation of a parsed template.
type Template struct {
	*template.Template
}

// New returns a new template.
func New(fileName string) (*Template, error) {
	tmpl, err := template.ParseFiles(fileName)
	if err != nil {
		return nil, err
	}

	return &Template{tmpl}, nil
}

// Default returns the default template.
func Default() (*Template, error) {
	tmpl, err := template.New("mock").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}

	return &Template{tmpl}, nil
}

// Render writes the generated code in the io.Writer.
func (t *Template) Render(wr io.Writer, pkg internal.Package, substitutions map[string]string) error {
	return t.Execute(wr, data{pkg, substitutions})
}
