package importer

import (
	"fmt"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/packages"

	"github.com/kokhanevych/gomockgen/internal"
)

// QualifierBuilder builds a qualifier.
type QualifierBuilder struct {
	env         []string
	packageDir  string
	packageName string
	packagePath string
}

// WithPackageDir sets the package directory to use for the qualifier.
func (b *QualifierBuilder) WithPackageDir(dir string) *QualifierBuilder {
	b.packageDir = dir
	return b
}

// WithPackageName sets the package name to use for the qualifier.
func (b *QualifierBuilder) WithPackageName(name string) *QualifierBuilder {
	b.packageName = name
	return b
}

// WithPackagePath sets the package path to use for the qualifier.
func (b *QualifierBuilder) WithPackagePath(path string) *QualifierBuilder {
	b.packagePath = path
	return b
}

// Build returns a new qualifier.
func (b *QualifierBuilder) Build() (Qualifier, error) {
	p := b.packagePath

	switch {
	case b.packageDir != "":
		var err error
		if p, err = b.path(b.packageDir); err != nil {
			return nil, err
		}
	case b.packageName != "":
		return &packageNameQualifier{
			qualifier:   newQualifier(),
			packageName: b.packageName,
		}, nil
	}

	return &packagePathQualifier{
		qualifier:   newQualifier(),
		packagePath: p,
	}, nil
}

func (b *QualifierBuilder) path(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	cfg := &packages.Config{Mode: packages.NeedName, Dir: dir, Env: b.env}
	pkgs, err := packages.Load(cfg, dir)
	if err == nil && len(pkgs) == 1 {
		return pkgs[0].PkgPath, nil
	}

	return "", nil
}

type qualifier struct {
	nameCount     map[string]int
	importsByPath map[string]internal.Import
	imports       []internal.Import
}

func newQualifier() qualifier {
	return qualifier{
		nameCount:     make(map[string]int),
		importsByPath: make(map[string]internal.Import),
	}
}

// Imports returns imported packages.
func (q *qualifier) Imports() []internal.Import {
	return q.imports
}

func (q *qualifier) qualify(pkg *types.Package) string {
	i, ok := q.importsByPath[pkg.Path()]
	if !ok {
		i = internal.Import{
			Name:  pkg.Name(),
			Alias: q.alias(pkg.Name()),
			Path:  pkg.Path(),
		}
		q.importsByPath[pkg.Path()] = i
		q.imports = append(q.imports, i)
	}

	if i.Alias == "" {
		return i.Name
	}

	return i.Alias
}

func (q *qualifier) alias(name string) string {
	q.nameCount[name]++
	c := q.nameCount[name]

	if c == 1 {
		return ""
	}

	return fmt.Sprintf("%s%d", name, c)
}

// packagePathQualifier represents a qualifier using the package path.
type packagePathQualifier struct {
	qualifier
	packagePath string
}

// Qualify controls how named package-level objects are printed.
func (q *packagePathQualifier) Qualify(pkg *types.Package) string {
	if pkg.Path() == q.packagePath {
		return ""
	}

	return q.qualify(pkg)
}

// packageNameQualifier represents a qualifier using the package name.
type packageNameQualifier struct {
	qualifier
	packageName string
}

// Qualify controls how named package-level objects are printed.
func (q *packageNameQualifier) Qualify(pkg *types.Package) string {
	if pkg.Name() == q.packageName {
		return ""
	}

	return q.qualify(pkg)
}
