package main

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

const (
	unixInstallCommand       = "curl -fsSL https://github.com/maccavelli/prepare-commit-msg/releases/latest/download/install.sh | sh -s -- --dir ~/.global-git-hooks"
	powershellInstallCommand = "& ([scriptblock]::Create((irm https://github.com/maccavelli/prepare-commit-msg/releases/latest/download/install.ps1))) -InstallDir \"$env:USERPROFILE\\.global-git-hooks\" -NoPathUpdate"
)

var reusableWorkflowRef = regexp.MustCompile(`(?m)^    uses: (maccavelli/go-selfupdate-lib/\.github/workflows/(?:build|publish)-selfupdate-release\.yml@[0-9a-f]{40} # v[0-9]+\.[0-9]+\.[0-9]+)$`)

func TestCurrentDocumentation(t *testing.T) {
	if err := validateCurrentDocumentation(os.DirFS(".")); err != nil {
		t.Fatal(err)
	}
}

func TestCurrentDocumentationDetectsBrokenFixtures(t *testing.T) {
	files := documentationFixture(t)

	files["README.md"] = &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(files["README.md"].Data), "--dir ~/.global-git-hooks", "--dir ~/.wrong-hooks"))}
	if err := validateCurrentDocumentation(files); err == nil || !strings.Contains(err.Error(), "Unix installer command") {
		t.Fatalf("broken Unix installer fixture error = %v, want Unix installer command", err)
	}

	files = documentationFixture(t)
	workflow := string(files[".github/workflows/ci.yml"].Data)
	refs := reusableWorkflowRef.FindAllStringSubmatch(workflow, -1)
	if len(refs) != 2 {
		t.Fatalf("workflow reusable references = %d, want 2", len(refs))
	}
	ref := refs[0][1]
	shaStart := strings.Index(ref, "@") + 1
	brokenRef := ref[:shaStart] + strings.Repeat("0", 40) + ref[shaStart+40:]
	files[".github/workflows/ci.yml"] = &fstest.MapFile{Data: []byte(strings.Replace(workflow, ref, brokenRef, 1))}
	if err := validateCurrentDocumentation(files); err == nil || !strings.Contains(err.Error(), "operations guide") {
		t.Fatalf("broken workflow fixture error = %v, want operations guide", err)
	}
}

func documentationFixture(t *testing.T) fstest.MapFS {
	t.Helper()
	paths := []string{"README.md", "docs/guides/cicd-operations.md", ".github/workflows/ci.yml"}
	files := make(fstest.MapFS, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		files[path] = &fstest.MapFile{Data: data}
	}
	return files
}

func validateCurrentDocumentation(files fs.FS) error {
	read := func(path string) (string, error) {
		data, err := fs.ReadFile(files, path)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", path, err)
		}
		return string(data), nil
	}

	readme, err := read("README.md")
	if err != nil {
		return err
	}
	if !strings.Contains(readme, unixInstallCommand) {
		return fmt.Errorf("README.md does not contain the documented Unix installer command")
	}
	if !strings.Contains(readme, powershellInstallCommand) {
		return fmt.Errorf("README.md does not contain the documented PowerShell installer command")
	}
	if !strings.Contains(readme, "Release `v1.10.1` moves to Go 1.27.3") {
		return fmt.Errorf("README.md does not identify v1.10.1 as the released Go 1.27.3 update")
	}

	workflow, err := read(".github/workflows/ci.yml")
	if err != nil {
		return err
	}
	refs := reusableWorkflowRef.FindAllStringSubmatch(workflow, -1)
	if len(refs) != 2 {
		return fmt.Errorf("ci workflow has %d reusable self-update workflow references, want 2", len(refs))
	}

	operations, err := read("docs/guides/cicd-operations.md")
	if err != nil {
		return err
	}
	for _, ref := range refs {
		if !strings.Contains(operations, ref[1]) {
			return fmt.Errorf("operations guide does not cite workflow reference %q", ref[1])
		}
	}
	return nil
}
