// Package main implements the Upjet code generator for the provider; it runs
// the Upjet pipeline to emit the provider's API types, controllers and CRDs.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crossplane/upjet/v2/pkg/pipeline"

	"github.com/devantler-tech/provider-upjet-unifi/config"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" {
		panic("root directory is required to be given as argument")
	}
	rootDir := os.Args[1]
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		panic(fmt.Sprintf("cannot calculate the absolute path with %s", rootDir))
	}
	pipeline.Run(config.GetProvider(), config.GetProviderNamespaced(), absRootDir)
}
