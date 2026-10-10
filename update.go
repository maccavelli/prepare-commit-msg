package main

import (
	_ "embed"
	"net/http"
	"time"

	"github.com/maccavelli/go-selfupdate-lib/buildinfo"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/cli"
	"github.com/maccavelli/go-selfupdate-lib/selfupdate/releasespec"
)

// releaseSpec is the release spec the build workflow reads: the products,
// the platforms and the packaging of every release
// (docs/decisions/0012-MADR-adopt-go-1-27-2-go-selfupdate-lib-v1-13-0-and-its-release-pipeline.md B1).
//
//go:embed selfupdate-release.json
var releaseSpec []byte

// releaseAssets is the asset selector the embedded spec decides. It fails
// when the spec does not name this program, so a spec and a program that
// disagree fail in the tests, before any release.
func releaseAssets() (selfupdate.AssetSelector, error) {
	spec, err := releasespec.Parse(releaseSpec)
	if err != nil {
		return nil, err
	}
	if _, err := spec.Product(AppTitle); err != nil {
		return nil, err
	}
	return spec.AssetSelector()
}

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
	selector, err := releaseAssets()
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
