package importer

import (
	"go/build"
	"go/types"
	"path/filepath"
)

// NewImportPathQualifier returns a new qualifier using the package import path.
func NewImportPathQualifier(importPath string) types.Qualifier {
	return func(pkg *types.Package) string {
		if pkg.Path() == importPath {
			return ""
		}

		return pkg.Name()
	}
}

// NewImportPathQualifier returns a new qualifier using the package directory.
func NewDirectoryQualifier(dir string) (types.Qualifier, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	dirs := make(map[string]string)

	return func(pkg *types.Package) string {
		if dirs[pkg.Path()] == "" {
			p, err := build.Import(pkg.Path(), "", build.FindOnly)
			if err != nil {
				panic(err)
			}

			dirs[pkg.Path()] = p.Dir
		}

		if dirs[pkg.Path()] == dir {
			return ""
		}

		return pkg.Name()
	}, nil
}
