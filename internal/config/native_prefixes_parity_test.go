package config

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// Single source of truth for "what counts as a native plugin namespace"
// lives in two places that must stay in sync:
//
//   - Go side: `internal/config/config.go::RegisterAllRules` registers
//     each ported plugin's rules with a slashed name like
//     `@typescript-eslint/no-explicit-any`. The set of distinct prefixes
//     across all registered slashed rule names IS the canonical native
//     namespace set.
//
//   - JS side: `packages/rslint/src/define-config.ts` exports
//     `NATIVE_PLUGIN_PREFIXES` (and the matching `KnownPlugin` union)
//     for the host-side collision validator. If the user puts one of
//     these prefixes in `eslintPlugins`, normalizeConfig fails fast.
//
// This test reads the JS source file, extracts the array contents,
// and asserts they equal the prefixes the Go registry produces. A
// drift in either direction makes CI red — the JS validator either
// misses a newly-ported plugin (false-negative — user collision
// silently shadowed) or rejects a legitimate prefix (false-positive
// — user can't use a plugin whose rule rslint never ported).
//
// Why parse the .ts file from Go: the alternative (codegen a JSON
// file from Go) is fragile (forget to regenerate → silent drift).
// Parsing the source ensures the truth is read on every test run.
func TestNativePluginPrefixes_JSAndGoInSync(t *testing.T) {
	// 1. Compute Go-side prefix set from the global registry.
	RegisterAllRules()
	allRules := GlobalRuleRegistry.GetAllRules()

	goPrefixes := map[string]struct{}{}
	for name := range allRules {
		idx := strings.IndexByte(name, '/')
		if idx <= 0 {
			continue // unprefixed core ESLint rule
		}
		prefix := name[:idx]
		// `eslint_plugin_test.go` registers test-only fake plugins
		// under `spike1__` / `dedupe_test__` / `shadow_test__` prefixes
		// in the SAME global registry the parity test reads from.
		// Filter those out by their `__` suffix convention so the
		// parity test reflects production prefixes only.
		if strings.HasSuffix(prefix, "__") {
			continue
		}
		goPrefixes[prefix] = struct{}{}
	}

	goSorted := make([]string, 0, len(goPrefixes))
	for p := range goPrefixes {
		goSorted = append(goSorted, p)
	}
	sort.Strings(goSorted)

	// 2. Parse JS side: load define-config.ts and extract the
	// NATIVE_PLUGIN_PREFIXES array literal contents.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// thisFile = internal/config/native_prefixes_parity_test.go
	// Walk up to repo root: ../../  (internal/config → repo root)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	defineConfigPath := filepath.Join(repoRoot, "packages", "rslint", "src", "define-config.ts")
	src, err := os.ReadFile(defineConfigPath)
	if err != nil {
		t.Fatalf("read define-config.ts: %v", err)
	}

	// Grab the array body between `NATIVE_PLUGIN_PREFIXES = [` and the
	// closing `]`. Tolerate whitespace + a `as const satisfies ...`
	// type assertion tail.
	re := regexp.MustCompile(`(?s)NATIVE_PLUGIN_PREFIXES\s*=\s*\[([^\]]*)\]`)
	m := re.FindSubmatch(src)
	if m == nil {
		t.Fatalf("could not find NATIVE_PLUGIN_PREFIXES array in %s", defineConfigPath)
	}
	body := string(m[1])

	// Pull every single- or double-quoted string literal out of the
	// array body.
	strRe := regexp.MustCompile(`['"]([^'"]+)['"]`)
	matches := strRe.FindAllStringSubmatch(body, -1)
	jsPrefixes := map[string]struct{}{}
	for _, mm := range matches {
		jsPrefixes[mm[1]] = struct{}{}
	}
	jsSorted := make([]string, 0, len(jsPrefixes))
	for p := range jsPrefixes {
		jsSorted = append(jsSorted, p)
	}
	sort.Strings(jsSorted)

	// 3. Sets must be equal. Surface BOTH directions of drift so the
	// fixer sees exactly which side is out of date.
	missingFromJS := []string{}
	for _, p := range goSorted {
		if _, ok := jsPrefixes[p]; !ok {
			missingFromJS = append(missingFromJS, p)
		}
	}
	missingFromGo := []string{}
	for _, p := range jsSorted {
		if _, ok := goPrefixes[p]; !ok {
			missingFromGo = append(missingFromGo, p)
		}
	}

	if len(missingFromJS) > 0 || len(missingFromGo) > 0 {
		t.Fatalf(
			"Native plugin prefix set drift between Go and JS sides:\n"+
				"  Go side (from GlobalRuleRegistry):    %v\n"+
				"  JS side (NATIVE_PLUGIN_PREFIXES):     %v\n"+
				"  missing from JS NATIVE_PLUGIN_PREFIXES (add these to define-config.ts): %v\n"+
				"  missing from Go registry (remove from JS, or add the Go-side rules): %v",
			goSorted, jsSorted, missingFromJS, missingFromGo,
		)
	}
}
