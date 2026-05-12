package modules

import (
	"fmt"
	"log"
	"os"

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

// Package writes the module tarball into the current working directory and returns
// the filename plus a content-digest checksum (see HashArchiveContents) suitable
// for comparison against a previously published version.
func Package(logger *log.Logger, manifest *types.ModuleManifest, version string, addlFiles []string) (string, string, error) {
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
		return tarballFilename, "", err
	}
	colorstring.Fprintln(logger.Writer(), "[green]Packaged module")

	data, err := os.ReadFile(tarballFilename)
	if err != nil {
		return tarballFilename, "", fmt.Errorf("reading packaged tarball for checksum: %w", err)
	}
	checksum, err := HashArchiveContents(data, ".tar.gz")
	if err != nil {
		return tarballFilename, "", fmt.Errorf("computing tarball checksum: %w", err)
	}
	return tarballFilename, checksum, nil
}
