package iac

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover_StackScopedTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".nullstone", "dev.yml"), "version: \"0.1\"\n# flat\n")
	mustWrite(t, filepath.Join(dir, ".nullstone", "stacks", "primary", "dev.yml"), "version: \"0.1\"\n# stack-scoped\n")

	pmr, err := Discover(dir, "primary", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	override := pmr.Overrides["dev"]
	if override == nil {
		t.Fatalf("expected dev override, got none")
	}
	got := override.IacContext.Filename
	want := filepath.Join(dir, ".nullstone", "stacks", "primary", "dev.yml")
	if got != want {
		t.Errorf("expected stack-scoped file %q, got %q", want, got)
	}
}

func TestDiscover_FallsBackToFlatWhenStackDirEmpty(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".nullstone", "dev.yml"), "version: \"0.1\"\n# flat\n")
	if err := os.MkdirAll(filepath.Join(dir, ".nullstone", "stacks", "primary"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	pmr, err := Discover(dir, "primary", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	override := pmr.Overrides["dev"]
	if override == nil {
		t.Fatalf("expected dev override from flat dir, got none")
	}
	got := override.IacContext.Filename
	want := filepath.Join(dir, ".nullstone", "dev.yml")
	if got != want {
		t.Errorf("expected flat fallback file %q, got %q", want, got)
	}
}

func TestDiscover_NoStackNameUsesFlat(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".nullstone", "dev.yml"), "version: \"0.1\"\n# flat\n")
	mustWrite(t, filepath.Join(dir, ".nullstone", "stacks", "primary", "dev.yml"), "version: \"0.1\"\n# stack-scoped\n")

	pmr, err := Discover(dir, "", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	override := pmr.Overrides["dev"]
	if override == nil {
		t.Fatalf("expected dev override, got none")
	}
	got := override.IacContext.Filename
	want := filepath.Join(dir, ".nullstone", "dev.yml")
	if got != want {
		t.Errorf("expected flat file %q (stack name unset), got %q", want, got)
	}
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
