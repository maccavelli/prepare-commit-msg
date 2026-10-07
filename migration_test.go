package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/maccavelli/go-selfupdate-lib/buildinfo"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/cli"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/selfupdatetest"
)

// here is the running platform, the only one the fixture releases publish.
var here = selfupdate.Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}

// releaseID is the identity the go-selfupdate-lib migration fixtures assume.
var releaseID = buildinfo.Info{Version: "v1.0.0", Kind: buildinfo.KindRelease}

// failSource fails discovery with the fixture's fixed error.
type failSource struct{ *selfupdatetest.FakeSource }

func (failSource) Latest(context.Context) (selfupdate.Release, error) {
	return selfupdate.Release{}, errors.New("fixture: source unavailable")
}

func fixtureRelease(tag string) selfupdatetest.ReleaseSpec {
	return selfupdatetest.NewRelease(AppTitle, tag, []selfupdate.Platform{here},
		func(selfupdate.Platform) []byte { return []byte(AppTitle + " " + tag + "\n") })
}

// fixtureUpdater builds an updater over src whose installer may only touch a
// temporary executable, never the test binary.
func fixtureUpdater(t *testing.T, src selfupdate.ReleaseSource) func() (*selfupdate.Updater, error) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, AppTitle)
	if runtime.GOOS == "windows" {
		target += ".exe"
	}
	if err := os.WriteFile(target, []byte("old binary\n"), 0o700); err != nil { //nolint:gosec // an executable fixture
		t.Fatal(err)
	}
	return func() (*selfupdate.Updater, error) {
		assets, err := selfupdate.NewExactAssetSelector([]selfupdate.Platform{here})
		if err != nil {
			return nil, err
		}
		inst, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{
			TargetPolicy: selfupdate.TargetPolicy{ExecutablePath: target, AllowedRoots: []string{dir}},
		})
		if err != nil {
			return nil, err
		}
		return selfupdate.New(selfupdate.Config{
			Source:    src,
			Versions:  selfupdate.NewStrictVersionPolicy(),
			Assets:    assets,
			Installer: inst,
			Reporter:  selfupdate.DiscardReporter(),
			Confirmer: selfupdate.NonInteractiveConfirmer(),
			Limits:    selfupdate.DefaultLimits(),
		})
	}
}

// runMain runs main with args and the update seams replaced. It returns
// everything written to stdout, by the update command's stream and by the
// process itself, what the update command wrote to stderr, and the exit code.
func runMain(t *testing.T, args []string, id buildinfo.Info, newUpdater func() (*selfupdate.Updater, error)) (stdout, stderr string, code int) {
	t.Helper()
	var out, errOut bytes.Buffer
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	prevStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = prevStdout
		_ = w.Close()
	})
	procOut := make(chan []byte)
	go func() {
		b, _ := io.ReadAll(r)
		procOut <- b
	}()
	prevArgs, prevExit := os.Args, osExit
	prevID, prevUpdater, prevOptions := buildIdentity, newUpdateUpdater, updateOptions
	t.Cleanup(func() {
		os.Args, osExit = prevArgs, prevExit
		buildIdentity, newUpdateUpdater, updateOptions = prevID, prevUpdater, prevOptions
	})
	os.Args = append([]string{AppTitle}, args...)
	code = -1
	osExit = func(c int) { code = c }
	buildIdentity = func() buildinfo.Info { return id }
	newUpdateUpdater = newUpdater
	updateOptions = func() cli.Options {
		return cli.Options{Stdout: &out, Stderr: &errOut, Stdin: strings.NewReader(""), Signals: []os.Signal{}}
	}
	main()
	os.Stdout = prevStdout
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return string(<-procOut) + out.String(), errOut.String(), code
}

// TestMigrationByteForByte drives the real main with `update --check` and
// compares it with go-selfupdate-lib's selfupdate/cli/testdata/migration
// fixtures, substituting only the asset name (rule 7).
func TestMigrationByteForByte(t *testing.T) {
	for _, sc := range []struct {
		name   string
		latest string
		fail   bool
	}{
		{"up-to-date", "v1.0.0", false},
		{"available", "v1.1.0", false},
		{"failed", "v1.1.0", true},
	} {
		t.Run(sc.name, func(t *testing.T) {
			fake := selfupdatetest.NewFakeSource(sc.latest, fixtureRelease("v1.0.0"), fixtureRelease("v1.1.0"))
			var src selfupdate.ReleaseSource = fake
			if sc.fail {
				src = failSource{fake}
			}
			stdout, stderr, code := runMain(t, []string{"update", "--check"}, releaseID, fixtureUpdater(t, src))
			asset := selfupdate.ExactAssetName(AppTitle, here)
			got := map[string]string{
				".stdout": strings.ReplaceAll(stdout, asset, "{{asset}}"),
				".stderr": strings.ReplaceAll(stderr, asset, "{{asset}}"),
				".code":   strconv.Itoa(code) + "\n",
			}
			for _, ext := range []string{".stdout", ".stderr", ".code"} {
				want, err := os.ReadFile(filepath.Join("testdata", "migration", sc.name+ext))
				if err != nil {
					t.Fatal(err)
				}
				if got[ext] != string(want) {
					t.Errorf("%s%s differs\n--- got ---\n%s--- want ---\n%s", sc.name, ext, got[ext], want)
				}
			}
		})
	}
}

// wantResultSchema is the update result's schema version from
// go-selfupdate-lib v1.6.0 on, which README.md's "Self-Update" states
// (0010-MADR D3, D4).
const wantResultSchema = 2

// TestUpdateCheckJSONSchema: `update --check --json` ends with one result
// object, whose schema version is the one the README documents. A library
// release that changes it fails here first (0010-MADR D3).
func TestUpdateCheckJSONSchema(t *testing.T) {
	fake := selfupdatetest.NewFakeSource("v1.0.0", fixtureRelease("v1.0.0"), fixtureRelease("v1.1.0"))
	stdout, stderr, code := runMain(t, []string{"update", "--check", "--json"}, releaseID, fixtureUpdater(t, fake))
	if code != 0 {
		t.Fatalf("update --check --json: exit %d, stderr %q", code, stderr)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	var last struct {
		Kind   string `json:"kind"`
		Result struct {
			SchemaVersion int `json:"schema_version"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last stdout line %q: %v", lines[len(lines)-1], err)
	}
	if last.Kind != "result" || last.Result.SchemaVersion != wantResultSchema {
		t.Fatalf("last stdout line is kind %q, schema_version %d; want result, %d\n%s",
			last.Kind, last.Result.SchemaVersion, wantResultSchema, stdout)
	}
}

// TestDefaultNewUpdateUpdater: the production updater builds, for every
// release platform, without touching the network.
func TestDefaultNewUpdateUpdater(t *testing.T) {
	u, err := defaultNewUpdateUpdater()
	if err != nil {
		t.Fatalf("defaultNewUpdateUpdater: %v", err)
	}
	if u == nil {
		t.Fatal("defaultNewUpdateUpdater returned a nil updater")
	}
}

// TestUpdateRefusedBeforeBuild: a usage error exits 1 and never builds the
// updater.
func TestUpdateRefusedBeforeBuild(t *testing.T) {
	never := func() (*selfupdate.Updater, error) {
		t.Fatal("the updater was built")
		return nil, nil
	}
	for _, args := range [][]string{{"update", "--bogus"}, {"update", "now"}, {"update", "--check", "--yes"}} {
		stdout, stderr, code := runMain(t, args, releaseID, never)
		if code != 1 || stdout != "" || !strings.Contains(stderr, "update failed: ") {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", args, code, stdout, stderr)
		}
	}
}

// TestVersionPrintsIdentity: version prints buildinfo's identity.
func TestVersionPrintsIdentity(t *testing.T) {
	out, _, code := runMain(t, []string{"version"}, buildinfo.Info{Version: "v1.2.3", Kind: buildinfo.KindRelease, Revision: "0123456789abcdef"}, nil)
	if want := AppTitle + " version v1.2.3 (release) 0123456789ab\n"; out != want || code != -1 {
		t.Fatalf("version: %q exit %d; want %q and no exit", out, code, want)
	}
}

// TestMakefileStamps: every release build stamps buildinfo's two variables.
func TestMakefileStamps(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make is not installed")
	}
	out, err := exec.Command("make", "-n", "build-all", "VERSION=v1.0.0").CombinedOutput()
	if err != nil {
		t.Fatalf("make -n build-all: %v\n%s", err, out)
	}
	for _, name := range []string{buildinfo.VersionVar + "=v1.0.0", buildinfo.KindVar + "=release"} {
		if n := strings.Count(string(out), "-X "+name); n != 6 {
			t.Errorf("%s stamped %d times, want 6", name, n)
		}
	}
}
