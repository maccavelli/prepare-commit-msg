package main

import (
	"slices"
	"testing"

	"github.com/maccavelli/go-selfupdate-lib/buildinfo"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/releasespec"
)

// TestReleaseSpec: the embedded release spec parses, names this program
// with the identity command, and lists the six platforms every release has
// shipped, so a v1.7.0 client still finds its asset
// (docs/decisions/0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md B1).
func TestReleaseSpec(t *testing.T) {
	spec, err := releasespec.Parse(releaseSpec)
	if err != nil {
		t.Fatalf("parse the embedded spec: %v", err)
	}
	prod, err := spec.Product(AppTitle)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(prod.IdentityArgs, []string{"identity"}) {
		t.Fatalf("identity_args %q, want [identity]", prod.IdentityArgs)
	}
	want := []selfupdate.Platform{
		{OS: "linux", Arch: "amd64"}, {OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"}, {OS: "darwin", Arch: "arm64"},
		{OS: "windows", Arch: "amd64"}, {OS: "windows", Arch: "arm64"},
	}
	if got := spec.Targets(); !slices.Equal(got, want) {
		t.Fatalf("targets %v, want %v", got, want)
	}
	if _, err := releaseAssets(); err != nil {
		t.Fatalf("releaseAssets: %v", err)
	}
	// Every release carries both installers, which install into the
	// directory the user names and leave core.hooksPath alone (0012-MADR I1).
	if got := spec.InstallerScripts(); !slices.Equal(got, []string{"install.sh", "install.ps1"}) {
		t.Fatalf("installer scripts %v, want [install.sh install.ps1]", got)
	}
	if spec.Installer == nil || len(spec.Installer.Hooks) != 0 {
		t.Fatalf("installer %+v, want present with no hooks", spec.Installer)
	}
	// README.md names the installers' variables by this prefix.
	if prefix, err := spec.InstallerEnvPrefix(AppTitle); err != nil || prefix != "PREPARE_COMMIT_MSG" {
		t.Fatalf("installer env prefix %q, %v; want PREPARE_COMMIT_MSG", prefix, err)
	}
}

// TestIdentityCommand: identity prints the build identity alone, which the
// build workflow and the installers require on its first line.
func TestIdentityCommand(t *testing.T) {
	id := buildinfo.Info{Version: "v1.8.0", Kind: buildinfo.KindRelease, Revision: "0123456789abcdef"}
	stdout, _, code := runMain(t, []string{"identity"}, id, nil)
	if code != -1 {
		t.Fatalf("identity exited %d", code)
	}
	if want := id.String() + "\n"; stdout != want {
		t.Fatalf("identity printed %q, want %q", stdout, want)
	}
}
