package rule

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestECMAScriptGlobalsMatchESLint10_8(t *testing.T) {
	t.Parallel()
	if LatestECMAScriptVersion != 2026 {
		t.Fatalf("ESLint 10.8 snapshot expects latest=2026, got %d; update the snapshot with the target ESLint version", LatestECMAScriptVersion)
	}

	// Snapshot of ESLint 10.8 conf/globals.js, grouped by first edition.
	additions := []struct {
		version int
		names   string
	}{
		{
			version: 3,
			names: "Array Boolean constructor Date decodeURI decodeURIComponent " +
				"encodeURI encodeURIComponent Error escape eval EvalError Function " +
				"hasOwnProperty Infinity isFinite isNaN isPrototypeOf Math NaN " +
				"Number Object parseFloat parseInt propertyIsEnumerable RangeError " +
				"ReferenceError RegExp String SyntaxError toLocaleString toString " +
				"TypeError undefined unescape URIError valueOf",
		},
		{version: 5, names: "JSON"},
		{
			version: 2015,
			names: "ArrayBuffer DataView Float32Array Float64Array Int16Array " +
				"Int32Array Int8Array Intl Map Promise Proxy Reflect Set Symbol " +
				"Uint16Array Uint32Array Uint8Array Uint8ClampedArray WeakMap WeakSet",
		},
		{version: 2017, names: "Atomics SharedArrayBuffer"},
		{version: 2020, names: "BigInt BigInt64Array BigUint64Array globalThis"},
		{version: 2021, names: "AggregateError FinalizationRegistry WeakRef"},
		{version: 2025, names: "Float16Array Iterator"},
		{
			version: 2026,
			names:   "AsyncDisposableStack DisposableStack SuppressedError Temporal",
		},
	}

	expected := map[string]bool{}
	groupIndex := 0
	for _, version := range []int{
		3, 5, 2015, 2016, 2017, 2018, 2019, 2020, 2021, 2022, 2023, 2024, 2025, 2026,
	} {
		for groupIndex < len(additions) && additions[groupIndex].version <= version {
			for _, name := range strings.Fields(additions[groupIndex].names) {
				expected[name] = true
			}
			groupIndex++
		}

		globals := NewGlobals(LanguageOptions{ECMAVersion: version}, GlobalsInit{}, nil, nil, nil)
		actualCount := 0
		for name := range ecmaScriptGlobalIntroducedIn {
			if globals.LanguageAccess(name).IsDeclared() {
				actualCount++
			}
		}
		if actualCount != len(expected) {
			t.Errorf("ES%d global count = %d, want %d", version, actualCount, len(expected))
		}
		for name := range expected {
			if !globals.LanguageAccess(name).IsDeclared() {
				t.Errorf("ES%d is missing ESLint global %q", version, name)
			}
		}
	}

	latest := Globals{}
	for name := range ecmaScriptGlobalIntroducedIn {
		if !expected[name] {
			t.Errorf("latest contains non-ESLint global %q", name)
		}
		if !latest.LanguageAccess(name).IsDeclared() {
			t.Errorf("latest unexpectedly excludes %q", name)
		}
	}
	if latest.Access("AsyncIterator").IsDeclared() {
		t.Fatal("AsyncIterator is TypeScript-only and must not be a runtime language global")
	}
}

func TestGlobalsPrecedenceMatrix(t *testing.T) {
	t.Parallel()

	levels := []utils.GlobalAccess{
		utils.GlobalAccessUnset,
		utils.GlobalAccessOff,
		utils.GlobalAccessReadonly,
		utils.GlobalAccessWritable,
	}
	for _, name := range []string{"Promise", "customGlobal"} {
		language := utils.GlobalAccessUnset
		if name == "Promise" {
			language = utils.GlobalAccessReadonly
		}
		for _, configAccess := range levels {
			for _, inlineAccess := range levels {
				config := accessMap(name, configAccess)
				inline := accessMap(name, inlineAccess)
				globals := NewGlobals(LanguageOptions{ECMAVersion: 2015}, GlobalsInit{}, config, inline, nil)

				wantConfigured := language
				if configAccess != utils.GlobalAccessUnset {
					wantConfigured = configAccess
				}
				wantFinal := wantConfigured
				if inlineAccess != utils.GlobalAccessUnset {
					wantFinal = inlineAccess
				}
				wantOverride := configAccess
				if inlineAccess != utils.GlobalAccessUnset {
					wantOverride = inlineAccess
				}

				if got := globals.LanguageAccess(name); got != language {
					t.Errorf("%s config=%s inline=%s: LanguageAccess=%s, want %s", name, configAccess, inlineAccess, got, language)
				}
				if got := globals.ConfigOverride(name); got != configAccess {
					t.Errorf("%s config=%s inline=%s: ConfigOverride=%s, want %s", name, configAccess, inlineAccess, got, configAccess)
				}
				if got := globals.ConfiguredAccess(name); got != wantConfigured {
					t.Errorf("%s config=%s inline=%s: ConfiguredAccess=%s, want %s", name, configAccess, inlineAccess, got, wantConfigured)
				}
				if got := globals.Access(name); got != wantFinal {
					t.Errorf("%s config=%s inline=%s: Access=%s, want %s", name, configAccess, inlineAccess, got, wantFinal)
				}
				if got := globals.Override(name); got != wantOverride {
					t.Errorf("%s config=%s inline=%s: Override=%s, want %s", name, configAccess, inlineAccess, got, wantOverride)
				}
			}
		}
	}
}

func TestGlobalsApplyToGatesSeededECMAScriptNames(t *testing.T) {
	t.Parallel()

	globals := NewGlobals(
		LanguageOptions{ECMAVersion: 5},
		GlobalsInit{},
		map[string]utils.GlobalAccess{"customGlobal": utils.GlobalAccessReadonly},
		map[string]utils.GlobalAccess{"JSON": utils.GlobalAccessOff},
		nil,
	)
	dst := map[string]bool{
		"Promise":  true,
		"Temporal": true,
		"window":   true,
	}
	globals.ApplyTo(dst)

	for _, name := range []string{"Promise", "Temporal", "JSON"} {
		if dst[name] {
			t.Errorf("ApplyTo retained unavailable or disabled ECMAScript global %q", name)
		}
	}
	for _, name := range []string{"Object", "customGlobal", "window"} {
		if !dst[name] {
			t.Errorf("ApplyTo removed declared or unrelated seeded global %q", name)
		}
	}
}

func TestGlobalsApplyToCanRestoreOlderEditionNameFromConfig(t *testing.T) {
	t.Parallel()

	globals := NewGlobals(
		LanguageOptions{ECMAVersion: 5},
		GlobalsInit{},
		map[string]utils.GlobalAccess{"Promise": utils.GlobalAccessReadonly},
		nil,
		nil,
	)
	dst := map[string]bool{"Promise": true, "Temporal": true}
	globals.ApplyTo(dst)

	if !dst["Promise"] {
		t.Error("ApplyTo removed Promise even though config explicitly restored it")
	}
	if dst["Temporal"] {
		t.Error("ApplyTo retained an unconfigured future-edition global")
	}
}

func TestLanguageDefaultsPrecedenceAndApplyTo(t *testing.T) {
	t.Parallel()

	globalsInit, _ := ResolveLanguageDefaults("/repo/file.cjs")
	globals := NewGlobals(
		LanguageOptions{},
		globalsInit,
		map[string]utils.GlobalAccess{
			"exports": utils.GlobalAccessReadonly,
			"require": utils.GlobalAccessOff,
		},
		map[string]utils.GlobalAccess{
			"module":  utils.GlobalAccessWritable,
			"require": utils.GlobalAccessWritable,
		},
		nil,
	)

	wantLanguage := map[string]utils.GlobalAccess{
		"exports": utils.GlobalAccessWritable,
		"global":  utils.GlobalAccessReadonly,
		"module":  utils.GlobalAccessReadonly,
		"require": utils.GlobalAccessReadonly,
	}
	wantFinal := map[string]utils.GlobalAccess{
		"exports": utils.GlobalAccessReadonly,
		"global":  utils.GlobalAccessReadonly,
		"module":  utils.GlobalAccessWritable,
		"require": utils.GlobalAccessWritable,
	}
	for name, want := range wantLanguage {
		if got := globals.LanguageAccess(name); got != want {
			t.Errorf("LanguageAccess(%q) = %s, want %s", name, got, want)
		}
		if got := globals.Access(name); got != wantFinal[name] {
			t.Errorf("Access(%q) = %s, want %s", name, got, wantFinal[name])
		}
		if globals.IsECMAScriptGlobalName(name) {
			t.Errorf("CommonJS name %q was classified as an ECMAScript global", name)
		}
	}

	dst := map[string]bool{
		"__dirname": true,
		"process":   true,
	}
	globals.ApplyTo(dst)
	for name := range wantFinal {
		if !dst[name] {
			t.Errorf("ApplyTo did not declare %q", name)
		}
	}
	for _, name := range []string{"__filename", "Buffer", "console"} {
		if dst[name] {
			t.Errorf("ApplyTo unexpectedly declared %q", name)
		}
	}
	// ApplyTo preserves unrelated seed data; the resolved defaults themselves must
	// not introduce any of these host names.
	if !dst["__dirname"] || !dst["process"] {
		t.Fatal("ApplyTo removed unrelated seed data")
	}
}

func TestLanguageDefaultsCanBeDisabled(t *testing.T) {
	t.Parallel()

	globalsInit, _ := ResolveLanguageDefaults("file.cjs")
	globals := NewGlobals(
		LanguageOptions{},
		globalsInit,
		map[string]utils.GlobalAccess{"require": utils.GlobalAccessReadonly},
		map[string]utils.GlobalAccess{"require": utils.GlobalAccessOff},
		nil,
	)
	dst := map[string]bool{"require": true}
	globals.ApplyTo(dst)
	if globals.Access("require") != utils.GlobalAccessOff || dst["require"] {
		t.Fatal("final inline off did not remove the CommonJS require global")
	}
}

func TestLanguageDefaultsDoNotLeakThroughApplyToSeed(t *testing.T) {
	t.Parallel()

	globalsInit, _ := ResolveLanguageDefaults("file.js")
	globals := NewGlobals(LanguageOptions{}, globalsInit, nil, nil, nil)
	dst := map[string]bool{
		"exports": true,
		"global":  true,
		"module":  true,
		"require": true,
		"custom":  true,
	}

	globals.ApplyTo(dst)
	for _, entry := range languageGlobalCatalog {
		if dst[entry.name] {
			t.Errorf("ApplyTo retained inactive language-default global %q", entry.name)
		}
	}
	if !dst["custom"] {
		t.Fatal("ApplyTo removed an unrelated seed entry")
	}
}

func accessMap(name string, access utils.GlobalAccess) map[string]utils.GlobalAccess {
	if access == utils.GlobalAccessUnset {
		return nil
	}
	return map[string]utils.GlobalAccess{name: access}
}
