package internal

// Import represents an imported Go package.
type Import struct {
	Name  string
	Alias string
	Path  string
}

// Variable represents a Go variable.
type Variable struct {
	Name string
	Type string
}

// Method represents a Go interface method.
type Method struct {
	Name       string
	Parameters []Variable
	Variadic   bool
	Results    []Variable
}

// Interface represents a Go interface.
type Interface struct {
	Name    string
	Methods []Method
}

// Package represents a Go package.
type Package struct {
	Name       string
	Imports    []Import
	Interfaces []Interface
}
