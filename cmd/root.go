package cmd

import (
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
		importPath := args[0]

		i, err := newImporter(importPath, options.FileName, options.MockPackage)
		if err != nil {
			return err
		}

		t, err := newTemplate(templateFileName)
		if err != nil {
			return err
		}

		g := generator.New(i, t)

		b, err := g.Generate(importPath, options, args[1:]...)
		if err != nil {
			return err
		}

		return write(options.FileName, b)
	},
}

func init() {
	cmd.Flags().StringToStringVarP(&options.MockNames, "names", "n", nil, "comma-separated interfaceName=mockName pairs of explicit mock names to use. Default mock names are interface names")
	cmd.Flags().StringVarP(&options.FileName, "out", "o", "", "output file instead of stdout")
	cmd.Flags().StringVarP(&options.MockPackage, "package", "p", "", "package of the generated code (default is the package of the interfaces)")
	cmd.Flags().StringVarP(&templateFileName, "template", "t", "", "template file used to generate the mock (default is the testify template)")
	cmd.Flags().StringToStringVarP(&options.Substitutions, "substitutions", "s", nil, "comma-separated key=value pairs of substitutions to make when expanding the template")
}

// Execute executes the root command.
func Execute() error {
	return cmd.Execute()
}

func newImporter(importPath, fileName, mockPackage string) (i *importer.Importer, err error) {
	b := &importer.QualifierBuilder{}

	if fileName != "" {
		b = b.WithPackageDir(filepath.Dir(fileName))
	}

	qf, err := b.WithPackageName(mockPackage).
		WithPackagePath(importPath).
		Build()
	if err != nil {
		return nil, err
	}

	return importer.New(qf), nil
}

func newTemplate(fileName string) (*template.Template, error) {
	if fileName == "" {
		return template.Default()
	}

	return template.New(fileName)
}

func write(fileName string, data []byte) error {
	if fileName != "" {
		if err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm); err != nil {
			return err
		}

		return os.WriteFile(fileName, data, 0666)
	}

	_, err := os.Stdout.Write(data)

	return err
}
