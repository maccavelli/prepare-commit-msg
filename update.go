package main

import (
	"net/http"
	"time"

	"github.com/maccavelli/go-selfupdate-lib/buildinfo"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/cli"
)

const (
	archAMD64 = "amd64"
	archARM64 = "arm64"
)

// updateTimeout bounds each GitHub request. cli.Run bounds the whole run.
const updateTimeout = 15 * time.Minute

// buildIdentity is the running binary's identity; tests replace it.
var buildIdentity = buildinfo.Identity

// newUpdateUpdater builds the updater; tests replace it.
var newUpdateUpdater = defaultNewUpdateUpdater

// updateOptions are the update command's streams; tests replace them.
var updateOptions = cli.StdioOptions

func defaultNewUpdateUpdater() (*selfupdate.Updater, error) {
	src, err := selfupdate.NewGitHubSource(selfupdate.GitHubOptions{
		Repository: selfupdate.Repository{Owner: "maccavelli", Name: "prepare-commit-msg"},
		Client:     &http.Client{Timeout: updateTimeout},
		UserAgent:  selfupdate.UserAgent(AppTitle, buildIdentity().Current()),
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
	// cli.Run supplies the reporter and the confirmer, on the right streams.
	return selfupdate.New(selfupdate.Config{
		Source:    src,
		Versions:  selfupdate.NewStrictVersionPolicy(),
		Assets:    selector,
		Installer: installer,
		Reporter:  selfupdate.DiscardReporter(),
		Confirmer: selfupdate.NonInteractiveConfirmer(),
		Limits:    selfupdate.DefaultLimits(),
	})
}
