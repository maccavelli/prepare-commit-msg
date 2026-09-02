package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/maccavelli/mcplib/selfupdate"
)

const (
	localVersionIdentity = "dev"
	releaseBuildKind     = "release"
	archAMD64            = "amd64"
	archARM64            = "arm64"
)

// RawVersion is the running identity stamped by release builds. Local source
// builds keep the default and are never ordered as a release.
var RawVersion = localVersionIdentity

// RawBuildKind is a linker-stamped string. Only the exact value "release"
// maps to selfupdate.ReleaseBuild. The Go linker is not used on a bool.
var RawBuildKind = "local"

const updateTimeout = 15 * time.Minute

func currentBuildKind() selfupdate.BuildKind {
	if RawBuildKind == releaseBuildKind {
		return selfupdate.ReleaseBuild
	}
	return selfupdate.LocalBuild
}

func currentVersion() string {
	if RawVersion != "" && RawVersion != localVersionIdentity {
		return RawVersion
	}
	if Version != "" && Version != localVersionIdentity {
		return Version
	}
	return RawVersion
}

var newUpdateUpdater = defaultNewUpdateUpdater

func defaultNewUpdateUpdater() (*selfupdate.Updater, error) {
	src, err := selfupdate.NewGitHubSource(selfupdate.GitHubOptions{
		Repository: selfupdate.Repository{Owner: "maccavelli", Name: "prepare-commit-msg"},
		Client:     &http.Client{Timeout: updateTimeout},
		UserAgent:  AppTitle + "/" + currentVersion(),
		Limits:     selfupdate.DefaultLimits(),
	})
	if err != nil {
		return nil, err
	}
	selector, err := selfupdate.NewExactAssetSelector([]selfupdate.Platform{
		{OS: "linux", Arch: archAMD64},
		{OS: "linux", Arch: archARM64},
		{OS: "darwin", Arch: archAMD64},
		{OS: "darwin", Arch: archARM64},
		{OS: "windows", Arch: archAMD64},
		{OS: "windows", Arch: archARM64},
	})
	if err != nil {
		return nil, err
	}
	installer, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{})
	if err != nil {
		return nil, err
	}
	return selfupdate.New(selfupdate.Config{
		Source:    src,
		Versions:  selfupdate.NewStrictVersionPolicy(),
		Assets:    selector,
		Installer: installer,
		Reporter:  selfupdate.NewTextReporter(os.Stdout),
		Confirmer: selfupdate.NewTerminalConfirmer(os.Stdin, os.Stdout),
		Limits:    selfupdate.DefaultLimits(),
	})
}

func runUpdate(ctx context.Context, args []string) (selfupdate.Result, error) {
	if err := ctx.Err(); err != nil {
		return selfupdate.Result{}, err
	}
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	check := fs.Bool("check", false, "check for updates without applying")
	force := fs.Bool("force", false, "reinstall current version or force overwrite")
	targetVersion := fs.String("version", "", "target specific version tag (e.g. v1.2.0)")
	yes := fs.Bool("yes", false, "non-interactive update")
	fs.BoolVar(yes, "y", false, "non-interactive update (shorthand)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return selfupdate.Result{}, nil
		}
		return selfupdate.Result{}, err
	}
	if fs.NArg() > 0 {
		return selfupdate.Result{}, fmt.Errorf("positional arguments are not accepted")
	}
	if *check && *yes {
		return selfupdate.Result{}, fmt.Errorf("--check and --yes are contradictory")
	}
	if *check && *force {
		return selfupdate.Result{}, fmt.Errorf("--check and --force are contradictory")
	}

	u, err := newUpdateUpdater()
	if err != nil {
		return selfupdate.Result{}, err
	}

	req := selfupdate.Request{
		Product:        AppTitle,
		CurrentVersion: currentVersion(),
		CurrentBuild:   currentBuildKind(),
		TargetVersion:  *targetVersion,
		CheckOnly:      *check,
		Force:          *force,
		Yes:            *yes,
	}
	return u.Run(ctx, req)
}
