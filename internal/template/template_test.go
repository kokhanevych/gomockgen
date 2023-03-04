package template

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kokhanevych/gomockgen/internal"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		assertion      assert.ValueAssertionFunc
		errorAssertion assert.ErrorAssertionFunc
	}{
		{
			name:           "nominal",
			fileName:       "mock.tmpl",
			assertion:      assert.NotNil,
			errorAssertion: assert.NoError,
		}, {
			name:           "error",
			fileName:       "not_found.tmpl",
			assertion:      assert.Nil,
			errorAssertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.fileName)

			tt.assertion(t, got)
			tt.errorAssertion(t, err)
		})
	}
}

func TestDefault(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		assertion      assert.ValueAssertionFunc
		errorAssertion assert.ErrorAssertionFunc
	}{
		{
			name:           "nominal",
			template:       defaultTemplate,
			assertion:      assert.NotNil,
			errorAssertion: assert.NoError,
		}, {
			name:           "error",
			template:       "{{}}",
			assertion:      assert.Nil,
			errorAssertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func(t string) { defaultTemplate = t }(defaultTemplate)
			defaultTemplate = tt.template

			got, err := Default()

			tt.assertion(t, got)
			tt.errorAssertion(t, err)
		})
	}
}

func TestTemplate_Render(t *testing.T) {
	pkg := internal.Package{
		Name:    "b",
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
						Name:       "Print",
						Parameters: []internal.Variable{{Name: "p0", Type: "io.Writer"}, {Name: "p1", Type: "[]byte"}},
						Results:    []internal.Variable{{Name: "n", Type: "int"}, {Name: "err", Type: "error"}},
					},
				},
			}, {
				Name: "I2",
			},
		},
	}

	want := `package b

import (	
	"github.com/stretchr/testify/mock"
	"golang.org/fake/b"
	"io"
)

// I1 is a mock.
type I1 struct { mock.Mock }

// F is a mocked method on I1.
func (i *I1) F(b B, args ...string) (error) {
	args := i.Called(b, args)
	return args.Error(0)
}

// Print is a mocked method on I1.
func (i *I1) Print(p0 io.Writer, p1 []byte) (int, error) {
	args := i.Called(p0, p1)
	return args.Int(0), args.Error(1)
}

// I2 is an interface mock.
type I2 struct { mock.Mock }

`

	tmpl, err := Default()
	require.NoError(t, err)

	type args struct {
		pkg           internal.Package
		substitutions map[string]string
	}
	tests := []struct {
		name      string
		tmpl      *Template
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "nominal",
			tmpl:      tmpl,
			args:      args{pkg, map[string]string{"I1Receiver": "i", "I2Comment": "I2 is an interface mock."}},
			want:      want,
			assertion: assert.NoError,
		},
		{
			name:      "error",
			tmpl:      &Template{&template.Template{}},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bytes.Buffer
			tt.assertion(t, tt.tmpl.Render(&got, tt.args.pkg, tt.args.substitutions))
			assert.Equal(t, tt.want, got.String())
		})
	}
}
