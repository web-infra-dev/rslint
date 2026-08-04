// TestNoRestrictedPathsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific upstream branch, Dimension 4 row or real-user
// report it covers, so future refactors can't silently regress them without
// breaking a named lock-in. The upstream 1:1 migration lives in
// no_restricted_paths_upstream_test.go.
package no_restricted_paths_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_restricted_paths"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// zonesBare produces the single-object option shape a CLI config passes
// directly, without the surrounding options array.
func zonesBare(entries ...map[string]interface{}) any {
	return map[string]interface{}{"zones": toAnySlice(entries)}
}

func TestNoRestrictedPathsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_restricted_paths.NoRestrictedPathsRule,
		[]rule_tester.ValidTestCase{
			// ---- Options: no options, empty options array and empty option object all disable the rule ----
			{Code: `import b from "../server/b"`, FileName: "restricted-paths/client/a.ts"},
			{Code: `import b from "../server/b"`, FileName: "restricted-paths/client/a.ts", Options: []interface{}{}},
			{Code: `import b from "../server/b"`, FileName: "restricted-paths/client/a.ts", Options: []interface{}{map[string]interface{}{}}},

			// Locks in upstream makePathValidators() arm: a zone without `from` produces no validators.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options:  zones(map[string]interface{}{"target": "./restricted-paths/client"}),
			},

			// Locks in upstream matchingZones filter: a zone without `target` never matches a file.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options:  zones(map[string]interface{}{"from": "./restricted-paths/server"}),
			},

			// Locks in upstream checkForRestrictedImportPath() arm 1: an import that
			// resolves to nothing is skipped before any zone is consulted.
			{
				Code:     `import missing from "./does-not-exist"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zonesBare(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths",
				}),
			},

			// ---- Real-user: #2089/#2090 — a sibling directory sharing a name prefix is not inside `from` ----
			{
				Code:     `import a from "../two-new/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server/two",
				}),
			},

			// ---- Real-user: #2818 — with a glob `from`, `except` patterns are matched
			// verbatim, so a relative exception glob can still match nothing. Here the
			// exception is absolute and does match, so the import is allowed. ----
			{
				Code:     `import a from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
					"except": list("/restricted-paths/server/two/**"),
				}),
			},

			// Locks in upstream computeAbsolutePathValidator() arm: `except: ['.']` resolves
			// to `from` itself, which excepts everything the zone would otherwise restrict.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
					"except": list("."),
				}),
			},

			// Locks in upstream isGlob() arm: a bare `?` is not glob syntax for is-glob,
			// so `serve?` stays a literal directory name and matches no real directory.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/serve?",
				}),
			},

			// Locks in upstream isGlob() arm: a leading `!` makes the whole string a glob
			// pattern, so it is no longer read as a directory path.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "!./restricted-paths/server",
				}),
			},

			// ---- Differences from ESLint: extended glob syntax is matched literally, so a
			// `!(...)` target selects no file and leaves the zone inactive ----
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/!(server)/**/*",
					"from":   "./restricted-paths/server",
				}),
			},

			// ---- Dimension 4: parenthesized callee — the shared module visitor only
			// matches a bare `require` identifier ----
			{
				Code:     `const b = (require)("../server/b")`,
				FileName: "restricted-paths/client/consumer.js",
				TSConfig: "tsconfig.allow-js.json",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
			},

			// ---- Dimension 4: graceful degradation — call shapes with no usable specifier ----
			{
				Code:     `const b = require()`,
				FileName: "restricted-paths/client/consumer.js",
				TSConfig: "tsconfig.allow-js.json",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
			},
			{
				Code:     `const name = "../server/b"; const b = require(name)`,
				FileName: "restricted-paths/client/consumer.js",
				TSConfig: "tsconfig.allow-js.json",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
			},

			// N/A: receiver/expression wrappers (`(X).y`, `X!.y`, `X as T`, `X?.y`),
			// access/key forms and declaration/container forms — this rule only ever
			// inspects a module specifier string literal and the linted file's path.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Options: the bare single-object shape a CLI config passes ----
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zonesBare(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// Locks in upstream containsPath() arm: `from` naming the imported file
			// itself is contained by itself.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client/a.ts",
					"from":   "./restricted-paths/server/b.ts",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// Locks in upstream reportImportsInRestrictedZone(): one report per matching
			// validator, so two overlapping `from` entries report the same import twice.
			{
				Code:     `import a from "../server/two/a"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   list("./restricted-paths/server", "./restricted-paths/server/two"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedPath("../server/two/a", 1, 15),
					unexpectedPath("../server/two/a", 1, 15),
				},
			},

			// Locks in upstream matchingZones.forEach(): two zones matching the same file
			// each report the same import.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(
					map[string]interface{}{
						"target": "./restricted-paths/client",
						"from":   "./restricted-paths/server",
					},
					map[string]interface{}{
						"target":  "./restricted-paths/client/**/*",
						"from":    "./restricted-paths",
						"message": "Second zone.",
					},
				),
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedPath("../server/b", 1, 15),
					unexpectedPathWithMessage("../server/b", "Second zone.", 1, 15),
				},
			},

			// Locks in upstream reportInvalidExceptions() ordering: exception errors for a
			// zone are emitted before its restricted-zone errors, even on the same import.
			{
				Code:     `import a from "./a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   list("./restricted-paths/server/one", "./restricted-paths/server"),
					"except": list("../one/b.ts"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					invalidExceptionPath("./a", 1, 15),
					unexpectedPath("./a", 1, 15),
				},
			},

			// Locks in upstream computeMixedGlobAndAbsolutePathValidator(): three `from`
			// entries collapse into a single validator, so the mix is reported once.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from": list(
						"./restricted-paths/server",
						"./restricted-paths/server/*",
						"./restricted-paths/other/*",
					),
				}),
				Errors: []rule_tester.InvalidTestCaseError{mixedGlobAndNonGlob("../server/b", 1, 15)},
			},

			// Locks in upstream computeGlobPatternPathValidator() arm: an empty `except`
			// list is valid and excepts nothing.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client/*",
					"from":   "./restricted-paths/server/*",
					"except": list(),
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// Locks in upstream reportImportsInRestrictedZone(): an empty `message` adds
			// no trailing separator to the default text.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target":  "./restricted-paths/client",
					"from":    "./restricted-paths/server",
					"message": "",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// Locks in upstream isGlob() arm: `{a,b}` alternatives make `from` a glob pattern.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/{server,other}/*",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// Locks in upstream isGlob() arm: a `[...]` character class makes `from` a glob pattern.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/serve[r]/*",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// Locks in upstream `basePath` default: with no `basePath`, relative zone paths
			// resolve against the current working directory, which is the project root here.
			{
				Code:     `import b from "../server/one/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "restricted-paths/client",
					"from":   "restricted-paths/server/one",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/one/b", 1, 15)},
			},

			// Locks in upstream `path.resolve(basePath, ...)`: a relative `basePath` is
			// made absolute against the current working directory before zone paths are
			// resolved from it, so the zone still matches absolute file names.
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zonesWithBasePath("./restricted-paths", map[string]interface{}{
					"target": "./client",
					"from":   "./server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// ---- Real-user: #2690 — a `dir/**/*` glob in `from` also covers the direct
			// children of `dir`, not only its nested subtrees ----
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// ---- Real-user: #2818 — a relative `except` glob is matched verbatim against
			// the resolved absolute import path, so it never lifts the restriction ----
			{
				Code:     `import a from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
					"except": list("./restricted-paths/server/two/**"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../two/a", 1, 15)},
			},

			// ---- Real-user: #2730 — `from` may name a single file through a glob suffix ----
			{
				Code:     `import c from "./c"`,
				FileName: "restricted-paths/server/b.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server",
					"from":   "./restricted-paths/server/c.*",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("./c", 1, 15)},
			},

			// ---- Dimension 4: module reference shapes the visitor must reach ----
			{
				Code:     `export { default as b } from "../server/b";`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 30)},
			},
			{
				Code:     `export * from "../server/b";`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},
			{
				Code:     `import "../server/b";`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 8)},
			},
			{
				Code:     `const load = () => import("../server/b");`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 27)},
			},
			{
				Code:     `function load() { return require("../server/b"); }`,
				FileName: "restricted-paths/client/consumer.js",
				TSConfig: "tsconfig.allow-js.json",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 34)},
			},
		},
	)
}
