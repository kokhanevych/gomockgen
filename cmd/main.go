package main

import (
	"os"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "gomockgen",
	Short: "Mock generator for Go interfaces based on text/template",
}

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
