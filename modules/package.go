package modules

import (
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
)

var (
	moduleFilePatterns = []string{
		"*.tf",
		"*.tf.tmpl",
		"README.md",
	}
	excludes = map[string]struct{}{
		"__backend__.tf": struct{}{},
	}
)

func Package(manifest *Manifest, version string) (string, error) {
	excludeFn := func(entry artifacts.GlobEntry) bool {
		_, ok := excludes[entry.Path]
		return ok
	}

	tarballFilename := fmt.Sprintf("%s.tar.gz", manifest.Name)
	if version != "" {
		tarballFilename = fmt.Sprintf("%s-%s.tar.gz", manifest.Name, version)
	}
	return tarballFilename, artifacts.PackageModule(".", tarballFilename, moduleFilePatterns, excludeFn)
}
