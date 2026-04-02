package modules

import (
	"fmt"
	"log"

	"github.com/mitchellh/colorstring"
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

func Package(logger *log.Logger, manifest *types.ModuleManifest, version string, addlFiles []string) (string, error) {
	excludeFn := func(entry artifacts.GlobEntry) bool {
		_, ok := excludes[entry.Path]
		return ok
	}

	tarballFilename := fmt.Sprintf("%s.tar.gz", manifest.Name)
	if version != "" {
		tarballFilename = fmt.Sprintf("%s-%s.tar.gz", manifest.Name, version)
	}
	logger.Println(fmt.Sprintf("Packaging module into %q...", tarballFilename))
	allPatterns := append(moduleFilePatterns, addlFiles...)
	logger.SetPrefix("    ")
	err := artifacts.PackageModule(logger, ".", tarballFilename, allPatterns, excludeFn)
	logger.SetPrefix("")
	if err != nil {
		return tarballFilename, err
	}
	colorstring.Fprintln(logger.Writer(), "[green]Packaged module")
	return tarballFilename, nil
}
