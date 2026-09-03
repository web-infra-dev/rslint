package rule

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestDefaultTypeScriptTypeGlobals(t *testing.T) {
	t.Parallel()

	if got, want := len(defaultTypeScriptTypeGlobals), 194; got != want {
		t.Fatalf("len(defaultTypeScriptTypeGlobals) = %d, want %d", got, want)
	}
	if !slices.IsSorted(defaultTypeScriptTypeGlobals[:]) {
		t.Fatal("defaultTypeScriptTypeGlobals must be sorted for binary search")
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(defaultTypeScriptTypeGlobals[:], "\n")))), "ed4f69d96156ce34f2d1e402a13f234744bb9e26fdb3eec2eda444bfa2272745"; got != want {
		t.Fatalf("defaultTypeScriptTypeGlobals SHA-256 = %s, want @typescript-eslint/scope-manager 8.69 esnext set %s", got, want)
	}

	for _, name := range []string{
		"Record",
		"Intl",
		"Reflect",
		"Temporal",
		"IteratorObjectConstructor",
		"const",
	} {
		if !IsDefaultTypeScriptTypeGlobal(name) {
			t.Errorf("defaultTypeScriptTypeGlobals is missing %q", name)
		}
	}

	for _, name := range []string{
		"globalThis",
		"IteratorConstructor",
		"HTMLElement",
		"NodeJS",
		"Buffer",
	} {
		if IsDefaultTypeScriptTypeGlobal(name) {
			t.Errorf("defaultTypeScriptTypeGlobals unexpectedly contains %q", name)
		}
	}
}
