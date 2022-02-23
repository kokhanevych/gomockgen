package importer

import (
	"fmt"
	"go/importer"
	"go/token"
	"go/types"

	"github.com/kokhanevych/gomockgen/internal"
)

// Importer resolves import paths to packages.
type Importer struct {
	types.Importer
	qualifier types.Qualifier
}

// New returns an Importer for importing directly from the source.
func New(qf types.Qualifier) *Importer {
	return &Importer{importer.ForCompiler(token.NewFileSet(), "source", nil), qf}
}

// Parse returns the package for the given import path with filtered interfaces.
func (im *Importer) Parse(importPath string, interfaces ...string) (internal.Package, error) {
	pkg, err := im.Importer.Import(importPath)
	if err != nil {
		return internal.Package{}, err
	}

	return im.toPackage(pkg, interfaces)
}

func (im *Importer) toPackage(pkg *types.Package, interfaceNames []string) (r internal.Package, err error) {
	r = internal.Package{
		Name: pkg.Name(),
	}

	r.Interfaces, err = im.lookup(pkg, interfaceNames)
	if err != nil {
		return internal.Package{}, err
	}

	for _, i := range pkg.Imports() {
		r.Imports = append(r.Imports, im.toImport(i))
	}

	return r, nil
}

func (im *Importer) lookup(pkg *types.Package, interfaceNames []string) ([]internal.Interface, error) {
	names := interfaceNames
	if len(names) == 0 {
		names = pkg.Scope().Names()
	}

	var ifaces []internal.Interface
	for _, n := range names {
		obj := pkg.Scope().Lookup(n)

		if obj == nil {
			return nil, fmt.Errorf("interface %s missing", n)
		}

		if types.IsInterface(obj.Type()) {
			iface := obj.Type().Underlying().(*types.Interface).Complete()
			ifaces = append(ifaces, im.toInterface(n, iface))
		} else if len(interfaceNames) > 0 {
			return nil, fmt.Errorf("%s should be an interface, was %s", n, obj.Type())
		}
	}

	return ifaces, nil
}

func (im *Importer) toInterface(name string, iface *types.Interface) internal.Interface {
	n := iface.NumMethods()

	r := internal.Interface{
		Name: name,
	}

	for i := 0; i < n; i++ {
		r.Methods = append(r.Methods, im.toMethod(iface.Method(i)))
	}

	return r
}

func (im *Importer) toMethod(f *types.Func) internal.Method {
	s := f.Type().(*types.Signature)

	r := internal.Method{
		Name:     f.Name(),
		Variadic: s.Variadic(),
	}

	for i := 0; i < len(r.Parameters); i++ {
		r.Parameters = append(r.Parameters, im.toVariable(s.Params().At(i)))
	}

	for i := 0; i < len(r.Results); i++ {
		r.Results = append(r.Results, im.toVariable(s.Results().At(i)))
	}

	return r
}

func (im *Importer) toVariable(v *types.Var) internal.Variable {
	return internal.Variable{
		Name: v.Name(),
		Type: types.TypeString(v.Type(), im.qualifier),
	}
}

func (im *Importer) toImport(p *types.Package) internal.Import {
	return internal.Import{
		Name: p.Name(),
		Path: p.Path(),
	}
}
