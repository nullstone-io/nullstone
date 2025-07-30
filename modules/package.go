package modules

import (
	"fmt"
	"gopkg.in/nullstone-io/go-api-client.v0/types"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
)

var (
	moduleFilePatterns = []string{
		"*.tf",
		"*.tf.tmpl",
		".terraform.lock.hcl",
		"README.md",
		"CHANGELOG.md",
	}
	excludes = map[string]struct{}{
		"__backend__.tf": {},
	}
)

func Package(manifest *types.ModuleManifest, version string, addlFiles []string) (string, error) {
	excludeFn := func(entry artifacts.GlobEntry) bool {
		_, ok := excludes[entry.Path]
		return ok
	}

	tarballFilename := fmt.Sprintf("%s.tar.gz", manifest.Name)
	if version != "" {
		tarballFilename = fmt.Sprintf("%s-%s.tar.gz", manifest.Name, version)
	}
	allPatterns := append(moduleFilePatterns, addlFiles...)
	return tarballFilename, artifacts.PackageModule(".", tarballFilename, allPatterns, excludeFn)
}
