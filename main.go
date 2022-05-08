package main

import (
	"os"

	"github.com/kokhanevych/gomockgen/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
