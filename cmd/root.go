package cmd

import (
	"go/types"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/kokhanevych/gomockgen/internal/generator"
	"github.com/kokhanevych/gomockgen/internal/importer"
	"github.com/kokhanevych/gomockgen/internal/template"
)

var (
	options          generator.Options
	templateFileName string
)

var cmd = &cobra.Command{
	Use:   "gomockgen <import-path> [<interface>...]",
	Short: "Mock generator for Go interfaces based on text/template",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var qf types.Qualifier
		importPath := args[0]
		out := os.Stdout

		switch {
		case options.FileName != "":
			dir := filepath.Dir(options.FileName)

			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}

			out, err = os.Create(options.FileName)
			if err != nil {
				return err
			}
			defer out.Close()

			qf, err = importer.NewDirectoryQualifier(dir)
			if err != nil {
				return err
			}
		case options.MockPackage != "":
			qf = importer.NewPackageNameQualifier(options.MockPackage)
		default:
			qf = importer.NewImportPathQualifier(importPath)
		}

		i := importer.New(qf)

		t, err := newTemplate(templateFileName)
		if err != nil {
			return err
		}

		g := generator.New(i, t)

		if err := g.Generate(importPath, out, options, args[1:]...); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	cmd.Flags().StringToStringVarP(&options.MockNames, "names", "n", nil, "comma-separated interfaceName=mockName pairs of explicit mock names to use. Default mock names are interface names")
	cmd.Flags().StringVarP(&options.FileName, "out", "o", "", "output file instead of stdout")
	cmd.Flags().StringVarP(&options.MockPackage, "package", "p", "", "package of the generated code (default is the package of the interfaces)")
	cmd.Flags().StringVarP(&templateFileName, "template", "t", "", "template file used to generate the mock (default is the testify template)")
}

// Execute executes the root command.
func Execute() error {
	return cmd.Execute()
}

func newTemplate(fileName string) (*template.Template, error) {
	if fileName == "" {
		return template.Default()
	}

	return template.New(templateFileName)
}
