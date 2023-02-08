package modules

import (
	"context"
	"fmt"
	"gopkg.in/nullstone-io/nullstone.v0/artifacts"
	"os"
	"os/exec"
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

func Package(ctx context.Context, manifest *Manifest, version string, addlFiles []string) (string, error) {
	fmt.Fprintf(os.Stderr, "updating .terraform.lock.hcl with all platforms...\n")
	if err := lockProviders(ctx); err != nil {
		// If an error happens, we won't prevent packaging, but we will report it
		fmt.Fprintf(os.Stderr, "could not update .terraform.lock.hcl: %s\n", err)
	}

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

func lockProviders(ctx context.Context) error {
	process := "terraform"
	args := []string{
		"providers",
		"lock",
		"-platform=linux_amd64",
		"-platform=darwin_amd64",
		"-platform=windows_amd64",
		"-platform=darwin_arm64",
		"-platform=linux_arm64",
	}
	cmd := exec.CommandContext(ctx, process, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
