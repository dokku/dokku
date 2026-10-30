package datastore

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/dokku/dokku-datastore/sdk"
)

// makefileList returns the names a Makefile variable links, without the
// directory each is listed under
func makefileList(t *testing.T, variable string) []string {
	t.Helper()

	contents, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatalf("unable to read the Makefile: %s", err)
	}

	for _, line := range strings.Split(string(contents), "\n") {
		value, found := strings.CutPrefix(line, variable+" = ")
		if !found {
			continue
		}

		names := []string{}
		for _, entry := range strings.Fields(value) {
			names = append(names, entry[strings.Index(entry, "/")+1:])
		}

		slices.Sort(names)
		return names
	}

	t.Fatalf("the Makefile does not set %s", variable)
	return nil
}

// Every command a datastore may implement needs a subcommand linked, or dokku
// has nothing to run when it is asked for
func TestEverySubcommandIsLinked(t *testing.T) {
	if linked, expected := makefileList(t, "SUBCOMMANDS"), sdk.SubcommandNames(); !slices.Equal(linked, expected) {
		t.Errorf("expected the Makefile to link the subcommands %v, got %v", expected, linked)
	}
}

// Every trigger a datastore implements needs a trigger linked, or dokku never
// runs it
func TestEveryTriggerIsLinked(t *testing.T) {
	if linked, expected := makefileList(t, "TRIGGERS"), sdk.TriggerNames(); !slices.Equal(linked, expected) {
		t.Errorf("expected the Makefile to link the triggers %v, got %v", expected, linked)
	}
}
