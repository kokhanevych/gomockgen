package generator

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kokhanevych/gomockgen/internal"
)

// Parser returns the package for the given import path with filtered interfaces.
type Parser interface {
	Parse(importPath string, interfaces ...string) (internal.Package, error)
}

// Renderer allows mock implementation rendering.
type Renderer interface {
	Render(io.Writer, internal.Package) error
}

// Generator generates mock implementations of Go interfaces.
type Generator struct {
	parser   Parser
	renderer Renderer
}

// New returns a new generator.
func New(p Parser, r Renderer) *Generator {
	return &Generator{p, r}
}

// Generate generates mock implementations of the specified Go interfaces for the given import path.
func (g *Generator) Generate(importPath string, interfaces ...string) error {
	pkg, err := g.parser.Parse(importPath, interfaces...)
	if err != nil {
		return err
	}

	for _, i := range pkg.Interfaces {
		for _, m := range i.Methods {
			if m.Variadic {
				l := len(m.Parameters)
				m.Parameters[l-1].Type = strings.Replace(m.Parameters[l-1].Type, "[]", "...", 1)
			}

			if len(m.Parameters) > 0 && m.Parameters[0].Name == "" {
				for i := range m.Parameters {
					m.Parameters[i].Name = fmt.Sprintf("p%d", i)
				}
			}
		}
	}

	return g.renderer.Render(os.Stdout, pkg)
}
