# gomockgen

[![Go](https://github.com/kokhanevych/gomockgen/actions/workflows/go.yml/badge.svg)](https://github.com/kokhanevych/gomockgen/actions/workflows/go.yml)

Mock generator for Go interfaces based on text/template

## Installation

```sh
$ go install github.com/kokhanevych/gomockgen@latest
```

## Usage

From the commandline:

```sh
$ gomockgen <import-path> [<interface>...] [flags]
```

Available options:

```
Flags:
  -h, --help                           help for gomockgen
  -n, --names stringToString           comma-separated interfaceName=mockName pairs of explicit mock names to use. Default mock names are interface names (default [])
  -o, --out string                     output file instead of stdout
  -p, --package string                 package of the generated code (default is the package of the interfaces)
  -s, --substitutions stringToString   comma-separated key=value pairs of substitutions to make when expanding the template (default [])
  -t, --template string                template file used to generate the mock (default is the testify template)
```

## Examples

Run:

```sh
$ gomockgen io Reader ReadWriter
```

You’ll see the following output:

```
package io

import (
        "github.com/stretchr/testify/mock"
)

// Reader is a mock.
type Reader struct{ mock.Mock }

// Read is a mocked method on Reader.
func (m *Reader) Read(p []byte) (int, error) {
        args := m.Called(p)
        return args.Int(0), args.Error(1)
}

// ReadWriter is a mock.
type ReadWriter struct{ mock.Mock }

// Read is a mocked method on ReadWriter.
func (m *ReadWriter) Read(p []byte) (int, error) {
        args := m.Called(p)
        return args.Int(0), args.Error(1)
}

// Write is a mocked method on ReadWriter.
func (m *ReadWriter) Write(p []byte) (int, error) {
        args := m.Called(p)
        return args.Int(0), args.Error(1)
}
```

To make substitutions when expanding the template, use `--substitutions`:

```sh
$ gomockgen io Reader ReadWriter --substitutions ReaderComment='Reader is a reader mock.',ReadWriterReceiver=w
```

This will generate the following output.

```
package io

import (
        "github.com/stretchr/testify/mock"
)

// Reader is a reader mock.
type Reader struct{ mock.Mock }

// Read is a mocked method on Reader.
func (m *Reader) Read(p []byte) (int, error) {
        args := m.Called(p)
        return args.Int(0), args.Error(1)
}

// ReadWriter is a mock.
type ReadWriter struct{ mock.Mock }

// Read is a mocked method on ReadWriter.
func (w *ReadWriter) Read(p []byte) (int, error) {
        args := w.Called(p)
        return args.Int(0), args.Error(1)
}

// Write is a mocked method on ReadWriter.
func (w *ReadWriter) Write(p []byte) (int, error) {
        args := w.Called(p)
        return args.Int(0), args.Error(1)
}
```

## Default template

```
{{$s := .Substitutions -}}
package {{.Package.Name}}

import (	
	"github.com/stretchr/testify/mock"
{{- range .Package.Imports}}
	"{{.Path}}"{{end}}
)

{{range $interface := .Package.Interfaces}}
{{- $k := printf "%sReceiver" .Name}}
{{- $receiver := index $s $k}}
{{- $receiver := or $receiver "m"}}
{{- $k := printf "%sComment" .Name}}
{{- $comment := index $s $k -}}
// {{if $comment}}{{$comment}}{{else}}{{.Name}} is a mock.{{end}}
type {{.Name}} struct { mock.Mock }

{{range .Methods -}}
// {{.Name}} is a mocked method on {{$interface.Name}}.
func ({{$receiver}} *{{$interface.Name}}) {{.Name}}(
	{{- range $index, $p := .Parameters}}{{if $index}}, {{end}}{{$p.Name}} {{$p.Type}}{{end -}}
) (
	{{- range $index, $r := .Results}}{{if $index}}, {{end}}{{$r.Type}}{{end -}}
) {
	{{if .Results}}args := {{end}}{{$receiver}}.Called(
		{{- range $index, $p := .Parameters}}{{if $index}}, {{end}}{{$p.Name}}{{end -}}
	)
{{- if .Results}}
	return {{range $index, $r := .Results}}
		{{- if $index}}, {{end}}
		{{- if eq $r.Type "bool"}}args.Bool({{$index}})
		{{- else if eq $r.Type "error"}}args.Error({{$index}})
		{{- else if eq $r.Type "int"}}args.Int({{$index}})
		{{- else if eq $r.Type "string"}}args.Int({{$index}})
		{{- else}}args.Get({{$index}}).({{$r.Type}}){{end}}
	{{- end}}
{{- end}}
}

{{end}}
{{- end}}
```