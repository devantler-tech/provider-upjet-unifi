// Package main implements the Upjet code generator for the provider; it runs
// the Upjet pipeline to emit the provider's API types, controllers and CRDs,
// and records the generated Terraform resources in config/generated.lst.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
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
	pc := config.GetProvider()
	pipeline.Run(pc, config.GetProviderNamespaced(), absRootDir)
	if err := writeGeneratedList(filepath.Join(absRootDir, "config"), pc); err != nil {
		panic(fmt.Sprintf("cannot write the generated resource list: %v", err))
	}
}

// writeGeneratedList writes generated.lst in configDir: the Terraform names of
// the resources the pipeline generated, as a sorted JSON array. `make
// schema-version-diff` reads it to know which resources' native state schema
// versions to compare. The write is confined to configDir.
func writeGeneratedList(configDir string, pc *ujconfig.Provider) error {
	names := make([]string, 0, len(pc.Resources))
	for name := range pc.Resources {
		names = append(names, name)
	}
	if len(names) == 0 {
		return fmt.Errorf("the provider configures no resources")
	}
	sort.Strings(names)
	data, err := json.Marshal(names)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(configDir)
	if err != nil {
		return err
	}
	if err := root.WriteFile("generated.lst", append(data, '\n'), 0o600); err != nil {
		_ = root.Close()
		return err
	}
	return root.Close()
}
