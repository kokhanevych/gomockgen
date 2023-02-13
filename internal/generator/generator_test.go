package generator

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kokhanevych/gomockgen/internal"
)

type parser struct{ mock.Mock }

// Parse is a mocked method on parser.
func (m *parser) Parse(importPath string, interfaces ...string) (internal.Package, error) {
	args := m.Called(importPath, interfaces)
	return args.Get(0).(internal.Package), args.Error(1)
}

type renderer struct{ mock.Mock }

// Render is a mocked method on renderer.
func (m *renderer) Render(w io.Writer, p internal.Package, substitutions map[string]string) error {
	args := m.Called(w, p, substitutions)
	return args.Error(0)
}

func TestGenerator_Generate(t *testing.T) {
	const importPath = "golang.org/fake/a"

	o := Options{
		MockPackage:   "b",
		MockNames:     map[string]string{"I2": "I3"},
		FileName:      "out.go",
		Substitutions: map[string]string{"k": "v"},
	}
	newPkg := func() internal.Package {
		return internal.Package{
			Name:    "a",
			Imports: []internal.Import{{Name: "b", Path: "golang.org/fake/b"}, {Name: "io", Path: "io"}},
			Interfaces: []internal.Interface{
				{
					Name: "I1",
					Methods: []internal.Method{
						{
							Name:       "F",
							Parameters: []internal.Variable{{Name: "b", Type: "B"}, {Name: "args", Type: "[]string"}},
							Results:    []internal.Variable{{Type: "error"}},
							Variadic:   true,
						},
						{
							Name:       "Print",
							Parameters: []internal.Variable{{Type: "io.Writer"}, {Type: "[]byte"}},
							Results:    []internal.Variable{{Name: "n", Type: "int"}, {Name: "err", Type: "error"}},
						},
					},
				}, {
					Name: "I2",
				},
			},
		}
	}
	pkgB := internal.Package{
		Name:    "b",
		Imports: []internal.Import{{Name: "b", Path: "golang.org/fake/b"}, {Name: "io", Path: "io"}, {Path: importPath}},
		Interfaces: []internal.Interface{
			{
				Name: "I1",
				Methods: []internal.Method{
					{
						Name:       "F",
						Parameters: []internal.Variable{{Name: "b", Type: "B"}, {Name: "args", Type: "...string"}},
						Results:    []internal.Variable{{Type: "error"}},
						Variadic:   true,
					},
					{
						Name: "Print",

						Parameters: []internal.Variable{{Name: "p0", Type: "io.Writer"}, {Name: "p1", Type: "[]byte"}},
						Results:    []internal.Variable{{Name: "n", Type: "int"}, {Name: "err", Type: "error"}},
					},
				},
			}, {
				Name: "I3",
			},
		},
	}
	pkgA := internal.Package{
		Name:    "a",
		Imports: []internal.Import{{Name: "b", Path: "golang.org/fake/b"}, {Name: "io", Path: "io"}},
		Interfaces: []internal.Interface{
			{
				Name: "I1",
				Methods: []internal.Method{
					{
						Name:       "F",
						Parameters: []internal.Variable{{Name: "b", Type: "B"}, {Name: "args", Type: "...string"}},
						Results:    []internal.Variable{{Type: "error"}},
						Variadic:   true,
					},
					{
						Name: "Print",

						Parameters: []internal.Variable{{Name: "p0", Type: "io.Writer"}, {Name: "p1", Type: "[]byte"}},
						Results:    []internal.Variable{{Name: "n", Type: "int"}, {Name: "err", Type: "error"}},
					},
				},
			}, {
				Name: "I2",
			},
		},
	}

	type args struct {
		importPath string
		options    Options
		interfaces []string
	}

	tests := []struct {
		name      string
		args      args
		expect    func(p *parser, r *renderer)
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "nominal",
			args: args{importPath: importPath, options: o, interfaces: []string{"I1", "I2"}},
			expect: func(p *parser, r *renderer) {
				p.On("Parse", importPath, []string{"I1", "I2"}).Return(newPkg(), nil).Once()

				r.On("Render", mock.Anything, pkgB, o.Substitutions).Run(func(args mock.Arguments) {
					_, _ = args.Get(0).(*bytes.Buffer).WriteString("package b")
				}).Return(nil).Once()
			},
			want:      "package b\n",
			assertion: assert.NoError,
		}, {
			name: "empty args",
			expect: func(p *parser, r *renderer) {
				p.On("Parse", "", []string(nil)).Return(newPkg(), nil).Once()

				r.On("Render", mock.Anything, pkgA, map[string]string(nil)).Run(func(args mock.Arguments) {
					_, _ = args.Get(0).(*bytes.Buffer).WriteString("package a")
				}).Return(nil).Once()
			},
			want:      "package a\n",
			assertion: assert.NoError,
		}, {
			name: "parse error",
			expect: func(p *parser, r *renderer) {
				p.On("Parse", "", []string(nil)).Return(internal.Package{}, assert.AnError).Once()
			},
			assertion: assert.Error,
		}, {
			name: "render error",
			expect: func(p *parser, r *renderer) {
				p.On("Parse", "", []string(nil)).Return(newPkg(), nil).Once()

				r.On("Render", mock.Anything, pkgA, map[string]string(nil)).Return(assert.AnError).Once()
			},
			assertion: assert.Error,
		}, {
			name: "format error",
			expect: func(p *parser, r *renderer) {
				p.On("Parse", "", []string(nil)).Return(newPkg(), nil).Once()

				r.On("Render", mock.Anything, pkgA, map[string]string(nil)).Return(nil).Once()
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := new(parser)
			r := new(renderer)
			tt.expect(p, r)

			g := New(p, r)

			var out bytes.Buffer
			err := g.Generate(tt.args.importPath, &out, tt.args.options, tt.args.interfaces...)

			mock.AssertExpectationsForObjects(t, p, r)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, out.String())
		})
	}
}
