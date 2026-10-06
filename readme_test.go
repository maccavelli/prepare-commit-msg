package main

import (
	"os"
	"regexp"
	"slices"
	"testing"

	"github.com/maccavelli/go-llmprovider-sdk/llmprovider"
	"github.com/maccavelli/go-llmprovider-sdk/llmprovider/catalog"
)

// readmeProviderRow matches a row of README.md's provider table, capturing
// the provider id and the "Default / Curated Models" cell.
var readmeProviderRow = regexp.MustCompile("(?m)^\\| \\*\\*`([a-z-]+)`\\*\\* \\|[^|]*\\|[^|]*\\| ([^|]*) \\|")

// readmeModelID matches one backquoted model id in a table cell.
var readmeModelID = regexp.MustCompile("`([^`]+)`")

// TestReadmeCuratedModels: README.md's provider table lists, for every
// provider with a curated catalog, exactly the SDK's catalog in its order, so
// the "(recommended)" model it names first is the default configure picks.
// An SDK release that changes a curated list fails here until the README
// follows (0009-MADR D4).
func TestReadmeCuratedModels(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	rows := map[llmprovider.ProviderID][]string{}
	for _, m := range readmeProviderRow.FindAllStringSubmatch(string(readme), -1) {
		var ids []string
		for _, id := range readmeModelID.FindAllStringSubmatch(m[2], -1) {
			ids = append(ids, id[1])
		}
		rows[llmprovider.ProviderID(m[1])] = ids
	}
	checked := 0
	for id := range llmprovider.ProviderEnvVars() {
		want := catalog.Static(id)
		if len(want) == 0 {
			continue
		}
		checked++
		got, ok := rows[id]
		if !ok {
			t.Errorf("README.md has no provider row for %s", id)
			continue
		}
		if !slices.Equal(got, want) {
			t.Errorf("README.md lists %s's curated models as %q; the SDK's catalog is %q", id, got, want)
		}
	}
	if checked == 0 {
		t.Fatal("no provider with a curated catalog was checked")
	}
}
