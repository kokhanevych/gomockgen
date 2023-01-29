package importer

import (
	"go/build"
	"go/types"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages/packagestest"

	"github.com/kokhanevych/gomockgen/internal"
)

func TestImporter_Parse(t *testing.T) { packagestest.TestAll(t, testImporter_Parse) }
func testImporter_Parse(t *testing.T, exporter packagestest.Exporter) {
	e := packagestest.Export(t, exporter, []packagestest.Module{{
		Name: "golang.org/fake",
		Files: map[string]interface{}{
			"a/a.go": `package a; import "io"; import "golang.org/fake/b"; type I1 interface { io.Writer; F(b b.B, args ...string) error }; type I2 interface {}`,
			"b/b.go": `package b; type B string`,
			"c/c.go": `package c; import "io"; import "golang.org/fake/b"; type I interface { F(b b.B, w io.Writer) }`,
		}}})
	defer e.Cleanup()

	var dir string
	if exporter == packagestest.GOPATH {
		t.Setenv("GO111MODULE", "off")
		build.Default.GOPATH = filepath.Dir(e.Config.Dir)
		dir = filepath.Join(e.Config.Dir, "golang.org", "fake", "b")
	} else {
		build.Default.Dir = e.Config.Dir
		dir = filepath.Join(e.Config.Dir, "b")
	}

	qf, err := NewDirectoryQualifier(dir)
	require.NoError(t, err)

	pkgA := internal.Package{
		Name:    "a",
		Imports: []internal.Import{{Name: "b", Path: "golang.org/fake/b"}, {Name: "io", Path: "io"}},
		Interfaces: []internal.Interface{
			{
				Name: "I1",
				Methods: []internal.Method{
					{
						Name:       "F",
						Parameters: []internal.Variable{{Name: "b", Type: "golang.org/fake/b.B"}, {Name: "args", Type: "[]string"}},
						Results:    []internal.Variable{{Type: "error"}},
						Variadic:   true,
					},
					{
						Name:       "Write",
						Parameters: []internal.Variable{{Name: "p", Type: "[]byte"}},
						Results:    []internal.Variable{{Name: "n", Type: "int"}, {Name: "err", Type: "error"}},
					},
				},
			}, {
				Name: "I2",
			},
		},
	}

	pkgC := internal.Package{
		Name:    "c",
		Imports: []internal.Import{{Name: "b", Path: "golang.org/fake/b"}, {Name: "io", Path: "io"}},
		Interfaces: []internal.Interface{
			{
				Name: "I",
				Methods: []internal.Method{
					{
						Name:       "F",
						Parameters: []internal.Variable{{Name: "b", Type: "B"}, {Name: "w", Type: "io.Writer"}},
					},
				},
			},
		},
	}

	type args struct {
		qualifier  types.Qualifier
		importPath string
		interfaces []string
	}
	tests := []struct {
		name      string
		args      args
		want      internal.Package
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "nominal",
			args:      args{nil, "golang.org/fake/a", []string{"I1", "I2"}},
			want:      pkgA,
			assertion: assert.NoError,
		}, {
			name:      "package import path qualifier",
			args:      args{NewImportPathQualifier("golang.org/fake/b"), "golang.org/fake/c", nil},
			want:      pkgC,
			assertion: assert.NoError,
		}, {
			name:      "package name qualifier",
			args:      args{NewPackageNameQualifier("b"), "golang.org/fake/c", nil},
			want:      pkgC,
			assertion: assert.NoError,
		}, {
			name:      "package directory qualifier",
			args:      args{qf, "golang.org/fake/c", nil},
			want:      pkgC,
			assertion: assert.NoError,
		}, {
			name:      "no interface filtering",
			args:      args{nil, "golang.org/fake/a", nil},
			want:      pkgA,
			assertion: assert.NoError,
		}, {
			name:      "package not found",
			args:      args{nil, "golang.org/fake", nil},
			assertion: assert.Error,
		}, {
			name:      "interface not found",
			args:      args{nil, "golang.org/fake/a", []string{"I3"}},
			assertion: assert.Error,
		}, {
			name:      "multiple packages found",
			args:      args{nil, "golang.org/fake...", nil},
			assertion: assert.Error,
		}, {
			name:      "not interface",
			args:      args{nil, "golang.org/fake/b", []string{"B"}},
			assertion: assert.Error,
		}, {
			name:      "no interfaces",
			args:      args{nil, "golang.org/fake/b", nil},
			want:      internal.Package{Name: "b"},
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			im := New(tt.args.qualifier)
			im.config.Dir = e.Config.Dir
			im.config.Env = e.Config.Env

			got, err := im.Parse(tt.args.importPath, tt.args.interfaces...)

			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
