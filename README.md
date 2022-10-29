# gomockgen
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