package importer

import (
	"fmt"
	"go/build"
	"go/types"
	"path/filepath"

	"github.com/kokhanevych/gomockgen/internal"
)

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

// PackagePathQualifier represents a qualifier using the package path.
type PackagePathQualifier struct {
	qualifier
	packagePath string
}

// NewPackagePathQualifier returns a new PackagePathQualifier.
func NewPackagePathQualifier(path string) *PackagePathQualifier {
	return &PackagePathQualifier{
		qualifier:   newQualifier(),
		packagePath: path,
	}
}

// Qualify controls how named package-level objects are printed.
func (q *PackagePathQualifier) Qualify(pkg *types.Package) string {
	if pkg.Path() == q.packagePath {
		return ""
	}

	return q.qualify(pkg)
}

// PackageNameQualifier represents a qualifier using the package name.
type PackageNameQualifier struct {
	qualifier
	packageName string
}

// NewPackageNameQualifier returns a new PackageNameQualifier.
func NewPackageNameQualifier(name string) *PackageNameQualifier {
	return &PackageNameQualifier{
		qualifier:   newQualifier(),
		packageName: name,
	}
}

// Qualify controls how named package-level objects are printed.
func (q *PackageNameQualifier) Qualify(pkg *types.Package) string {
	if pkg.Name() == q.packageName {
		return ""
	}

	return q.qualify(pkg)
}

// PackageDirectoryQualifier represents a qualifier using the package directory.
type PackageDirectoryQualifier struct {
	qualifier
	dirs             map[string]string
	packageDirectory string
}

// NewPackageDirectoryQualifier returns a new PackageDirectoryQualifier.
func NewPackageDirectoryQualifier(dir string) (*PackageDirectoryQualifier, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return &PackageDirectoryQualifier{
		qualifier:        newQualifier(),
		dirs:             make(map[string]string),
		packageDirectory: dir,
	}, nil
}

// Qualify controls how named package-level objects are printed.
func (q *PackageDirectoryQualifier) Qualify(pkg *types.Package) string {
	if q.dirs[pkg.Path()] == "" {
		p, err := build.Import(pkg.Path(), "", build.FindOnly)
		if err != nil {
			panic(err)
		}

		q.dirs[pkg.Path()] = p.Dir
	}

	if q.dirs[pkg.Path()] == q.packageDirectory {
		return ""
	}

	return q.qualify(pkg)
}
