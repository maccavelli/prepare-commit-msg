package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/maccavelli/mcplib/selfupdate"
)

type scriptSource struct {
	rel    selfupdate.Release
	bodies map[int64][]byte
	err    error
}

func (s *scriptSource) Latest(ctx context.Context) (selfupdate.Release, error) {
	if err := ctx.Err(); err != nil {
		return selfupdate.Release{}, err
	}
	return s.rel, s.err
}
func (s *scriptSource) ByTag(ctx context.Context, tag string) (selfupdate.Release, error) {
	if err := ctx.Err(); err != nil {
		return selfupdate.Release{}, err
	}
	rel := s.rel
	rel.Tag = tag
	return rel, s.err
}
func (s *scriptSource) OpenAsset(_ context.Context, _ selfupdate.Release, a selfupdate.Asset) (io.ReadCloser, error) {
	body, ok := s.bodies[a.ID]
	if !ok {
		return nil, fmt.Errorf("missing asset %d", a.ID)
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}

type scriptConfirmer struct {
	ok    bool
	err   error
	calls int
}

func (c *scriptConfirmer) Confirm(context.Context, selfupdate.Prompt) (bool, error) {
	c.calls++
	return c.ok, c.err
}

func productPlatforms() []selfupdate.Platform {
	return []selfupdate.Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
		{OS: "windows", Arch: "amd64"},
		{OS: "windows", Arch: "arm64"},
	}
}

func fixtureRelease(t *testing.T, tag string) (selfupdate.Release, map[int64][]byte) {
	t.Helper()
	plat := selfupdate.Platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
	bin := []byte("hello-bin")
	sum := sha256.Sum256(bin)
	hexsum := hex.EncodeToString(sum[:])
	name := AppTitle + "-" + plat.OS + "-" + plat.Arch
	if plat.OS == "windows" {
		name += ".exe"
	}
	manifest := []byte(hexsum + "  " + name + "\n")
	rel := selfupdate.Release{
		ID: 1, Tag: tag, URL: "https://example.invalid/" + tag, Immutable: true,
		Assets: []selfupdate.Asset{
			{ID: 2, Name: name, State: "uploaded", Size: int64(len(bin)), Digest: "sha256:" + hexsum},
			{ID: 3, Name: "SHA256SUMS", State: "uploaded", Size: int64(len(manifest))},
		},
	}
	return rel, map[int64][]byte{2: bin, 3: manifest}
}

func withTempTarget(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	exe := filepath.Join(home, AppTitle)
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	if err := os.WriteFile(exe, []byte("old-bytes"), 0o755); err != nil {
		t.Fatal(err)
	}
	return exe
}

func installTestUpdater(t *testing.T, src selfupdate.ReleaseSource, exe string, conf selfupdate.Confirmer) {
	t.Helper()
	selector, err := selfupdate.NewExactAssetSelector(productPlatforms())
	if err != nil {
		t.Fatal(err)
	}
	installer, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{
		TargetPolicy: selfupdate.TargetPolicy{ExecutablePath: exe},
	})
	if err != nil {
		t.Fatal(err)
	}
	if conf == nil {
		conf = &scriptConfirmer{ok: true}
	}
	u, err := selfupdate.New(selfupdate.Config{
		Source:    src,
		Versions:  selfupdate.NewStrictVersionPolicy(),
		Assets:    selector,
		Installer: installer,
		Reporter:  selfupdate.NewTextReporter(io.Discard),
		Confirmer: conf,
		Limits:    selfupdate.DefaultLimits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	prev := newUpdateUpdater
	newUpdateUpdater = func() (*selfupdate.Updater, error) { return u, nil }
	t.Cleanup(func() { newUpdateUpdater = prev })
}

func TestRunUpdateRejectsExtraArgsAndContradictions(t *testing.T) {
	called := false
	prev := newUpdateUpdater
	newUpdateUpdater = func() (*selfupdate.Updater, error) {
		called = true
		return nil, errors.New("factory should not run")
	}
	t.Cleanup(func() { newUpdateUpdater = prev })

	ctx := context.Background()
	if _, err := runUpdate(ctx, []string{"extra"}); err == nil {
		t.Fatal("accepted positional args")
	}
	if _, err := runUpdate(ctx, []string{"--check", "--yes"}); err == nil {
		t.Fatal("accepted --check --yes")
	}
	if _, err := runUpdate(ctx, []string{"--check", "--force"}); err == nil {
		t.Fatal("accepted --check --force")
	}
	if _, err := runUpdate(ctx, []string{"--invalid-flag-12345"}); err == nil {
		t.Fatal("accepted invalid flag")
	}
	if called {
		t.Fatal("constructed source before rejecting flags")
	}
}

func TestRunUpdateAliasesAndHelp(t *testing.T) {
	if _, err := runUpdate(context.Background(), []string{"--help"}); err != nil {
		t.Fatal(err)
	}
	exe := withTempTarget(t)
	rel, bodies := fixtureRelease(t, "v1.3.0")
	installTestUpdater(t, &scriptSource{rel: rel, bodies: bodies}, exe, &scriptConfirmer{ok: true})
	prevKind, prevVer := RawBuildKind, RawVersion
	RawBuildKind, RawVersion = "local", "dev"
	t.Cleanup(func() { RawBuildKind, RawVersion = prevKind, prevVer })

	if _, err := runUpdate(context.Background(), []string{"-y", "--force"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunUpdateCheckExitCodes(t *testing.T) {
	exe := withTempTarget(t)
	rel, bodies := fixtureRelease(t, "v1.3.0")
	installTestUpdater(t, &scriptSource{rel: rel, bodies: bodies}, exe, nil)
	prevKind, prevVer := RawBuildKind, RawVersion
	t.Cleanup(func() { RawBuildKind, RawVersion = prevKind, prevVer })

	RawBuildKind, RawVersion = "release", "v1.2.0"
	res, err := runUpdate(context.Background(), []string{"--check"})
	if !errors.Is(err, selfupdate.ErrUpdateAvailable) {
		t.Fatalf("newer check err = %v", err)
	}
	if selfupdate.ExitCode(res, err) != 10 {
		t.Fatalf("exit = %d", selfupdate.ExitCode(res, err))
	}

	RawVersion = "v1.3.0"
	res, err = runUpdate(context.Background(), []string{"--check"})
	if err != nil {
		t.Fatal(err)
	}
	if selfupdate.ExitCode(res, err) != 0 || res.Operation != selfupdate.OperationNone {
		t.Fatalf("%+v %v", res, err)
	}

	res, err = runUpdate(context.Background(), []string{"--check", "--version", "v1.2.0"})
	if !errors.Is(err, selfupdate.ErrUpdateAvailable) || res.Operation != selfupdate.OperationRollback {
		t.Fatalf("rollback check %+v %v", res, err)
	}
	if selfupdate.ExitCode(res, err) != 10 {
		t.Fatalf("rollback exit = %d", selfupdate.ExitCode(res, err))
	}
}

func TestRunUpdateLocalBuildRequiresForce(t *testing.T) {
	exe := withTempTarget(t)
	rel, bodies := fixtureRelease(t, "v1.3.0")
	installTestUpdater(t, &scriptSource{rel: rel, bodies: bodies}, exe, nil)
	prevKind, prevVer := RawBuildKind, RawVersion
	RawBuildKind, RawVersion = "local", "dev"
	t.Cleanup(func() { RawBuildKind, RawVersion = prevKind, prevVer })

	_, err := runUpdate(context.Background(), []string{"--yes"})
	if err == nil {
		t.Fatal("local apply without force succeeded")
	}
	if selfupdate.ExitCode(selfupdate.Result{}, err) != 1 {
		t.Fatalf("exit = %d", selfupdate.ExitCode(selfupdate.Result{}, err))
	}

	res, err := runUpdate(context.Background(), []string{"--check"})
	if !errors.Is(err, selfupdate.ErrUpdateAvailable) || res.Operation != selfupdate.OperationReplaceLocal {
		t.Fatalf("%+v %v", res, err)
	}

	if _, err := runUpdate(context.Background(), []string{"--yes", "--force"}); err != nil {
		t.Fatal(err)
	}
}

func TestRunUpdateYesNoAndNonTTY(t *testing.T) {
	exe := withTempTarget(t)
	rel, bodies := fixtureRelease(t, "v1.3.0")
	conf := &scriptConfirmer{ok: false}
	installTestUpdater(t, &scriptSource{rel: rel, bodies: bodies}, exe, conf)
	prevKind, prevVer := RawBuildKind, RawVersion
	RawBuildKind, RawVersion = "release", "v1.2.0"
	t.Cleanup(func() { RawBuildKind, RawVersion = prevKind, prevVer })

	res, err := runUpdate(context.Background(), []string{})
	if err != nil || !res.Declined {
		t.Fatalf("decline %+v %v", res, err)
	}
	if conf.calls != 1 {
		t.Fatalf("confirms = %d", conf.calls)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})
	selector, err := selfupdate.NewExactAssetSelector(productPlatforms())
	if err != nil {
		t.Fatal(err)
	}
	installer, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{
		TargetPolicy: selfupdate.TargetPolicy{ExecutablePath: exe},
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := selfupdate.New(selfupdate.Config{
		Source:    &scriptSource{rel: rel, bodies: bodies},
		Versions:  selfupdate.NewStrictVersionPolicy(),
		Assets:    selector,
		Installer: installer,
		Reporter:  selfupdate.NewTextReporter(io.Discard),
		Confirmer: selfupdate.NewTerminalConfirmer(r, io.Discard),
		Limits:    selfupdate.DefaultLimits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	prev := newUpdateUpdater
	newUpdateUpdater = func() (*selfupdate.Updater, error) { return u, nil }
	t.Cleanup(func() { newUpdateUpdater = prev })
	_, err = runUpdate(context.Background(), []string{})
	if !errors.Is(err, selfupdate.ErrConfirmationRequired) {
		t.Fatalf("non-tty err = %v", err)
	}
}

func TestRunUpdateSameVersionForceAndCancel(t *testing.T) {
	exe := withTempTarget(t)
	rel, bodies := fixtureRelease(t, "v1.3.0")
	installTestUpdater(t, &scriptSource{rel: rel, bodies: bodies}, exe, &scriptConfirmer{ok: true})
	prevKind, prevVer := RawBuildKind, RawVersion
	RawBuildKind, RawVersion = "release", "v1.3.0"
	t.Cleanup(func() { RawBuildKind, RawVersion = prevKind, prevVer })

	res, err := runUpdate(context.Background(), []string{"--yes", "--force"})
	if err != nil || res.Operation != selfupdate.OperationReinstall {
		t.Fatalf("%+v %v", res, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = runUpdate(ctx, []string{"--check"})
	if err == nil {
		t.Fatal("expected cancellation")
	}
}

func TestRunUpdateUnsupportedPlatform(t *testing.T) {
	exe := withTempTarget(t)
	rel, bodies := fixtureRelease(t, "v1.3.0")
	src := &scriptSource{rel: rel, bodies: bodies}
	selector, err := selfupdate.NewExactAssetSelector([]selfupdate.Platform{{OS: "plan9", Arch: "amd64"}})
	if err != nil {
		t.Fatal(err)
	}
	installer, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{
		TargetPolicy: selfupdate.TargetPolicy{ExecutablePath: exe},
	})
	if err != nil {
		t.Fatal(err)
	}
	u, err := selfupdate.New(selfupdate.Config{
		Source:    src,
		Versions:  selfupdate.NewStrictVersionPolicy(),
		Assets:    selector,
		Installer: installer,
		Reporter:  selfupdate.NewTextReporter(io.Discard),
		Confirmer: &scriptConfirmer{ok: true},
		Limits:    selfupdate.DefaultLimits(),
	})
	if err != nil {
		t.Fatal(err)
	}
	prev := newUpdateUpdater
	newUpdateUpdater = func() (*selfupdate.Updater, error) { return u, nil }
	t.Cleanup(func() { newUpdateUpdater = prev })
	prevKind, prevVer := RawBuildKind, RawVersion
	RawBuildKind, RawVersion = "release", "v1.2.0"
	t.Cleanup(func() { RawBuildKind, RawVersion = prevKind, prevVer })
	_, err = runUpdate(context.Background(), []string{"--check"})
	if !errors.Is(err, selfupdate.ErrUnsupportedPlatform) {
		t.Fatalf("err = %v", err)
	}
}

func TestCurrentBuildKind(t *testing.T) {
	prev := RawBuildKind
	t.Cleanup(func() { RawBuildKind = prev })
	RawBuildKind = "local"
	if currentBuildKind() != selfupdate.LocalBuild {
		t.Fatal("local")
	}
	RawBuildKind = "release"
	if currentBuildKind() != selfupdate.ReleaseBuild {
		t.Fatal("release")
	}
	RawBuildKind = "Release"
	if currentBuildKind() != selfupdate.LocalBuild {
		t.Fatal("only exact release maps")
	}
}

func TestMainUpdateExit10(t *testing.T) {
	exe := withTempTarget(t)
	rel, bodies := fixtureRelease(t, "v1.3.0")
	installTestUpdater(t, &scriptSource{rel: rel, bodies: bodies}, exe, nil)
	prevKind, prevVer := RawBuildKind, RawVersion
	RawBuildKind, RawVersion = "release", "v1.2.0"
	t.Cleanup(func() { RawBuildKind, RawVersion = prevKind, prevVer })

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{AppTitle, "update", "--check"}

	var exitCode = -1
	oldExit := osExit
	t.Cleanup(func() { osExit = oldExit })
	osExit = func(code int) {
		exitCode = code
		panic("osExit")
	}
	defer func() { _ = recover() }()
	main()
	if exitCode != 10 {
		t.Fatalf("exit = %d", exitCode)
	}
}

func TestRunUpdateNoLiveGitHubOnFlagError(t *testing.T) {
	prev := newUpdateUpdater
	newUpdateUpdater = func() (*selfupdate.Updater, error) {
		t.Fatal("must not construct GitHub source")
		return nil, nil
	}
	t.Cleanup(func() { newUpdateUpdater = prev })
	_, err := runUpdate(context.Background(), []string{"--check", "--yes"})
	if err == nil {
		t.Fatal("expected error")
	}
}
