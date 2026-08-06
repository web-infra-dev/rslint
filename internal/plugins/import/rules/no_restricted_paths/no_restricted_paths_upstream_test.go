// TestNoRestrictedPathsUpstream migrates the full valid/invalid suite from
// upstream tests/src/rules/no-restricted-paths.js 1:1, including the
// `context('Typescript')` block. Upstream's fixture root `tests/files/` maps
// onto this plugin's fixture root, so upstream's
// `./tests/files/restricted-paths/...` zone paths become
// `./restricted-paths/...` here, and its `.js` fixtures become `.ts` ones
// except for the CommonJS cases, which stay on `.js` so `require` specifiers
// take part in module resolution.
// Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in no_restricted_paths_extras_test.go.
package no_restricted_paths_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/no_restricted_paths"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func zones(entries ...map[string]interface{}) any {
	return []interface{}{map[string]interface{}{"zones": toAnySlice(entries)}}
}

func zonesWithBasePath(basePath string, entries ...map[string]interface{}) any {
	return []interface{}{map[string]interface{}{
		"zones":    toAnySlice(entries),
		"basePath": basePath,
	}}
}

func toAnySlice(entries []map[string]interface{}) []interface{} {
	values := make([]interface{}, len(entries))
	for i := range entries {
		values[i] = entries[i]
	}
	return values
}

// list mirrors the `[]interface{}` shape a JSON/JS config produces for the
// array-valued `target`, `from` and `except` options.
func list(values ...string) []interface{} {
	items := make([]interface{}, len(values))
	for i, value := range values {
		items[i] = value
	}
	return items
}

func errorAt(messageID string, message string, line int, column int, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: messageID,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}
}

func unexpectedPath(specifier string, line int, column int) rule_tester.InvalidTestCaseError {
	return errorAt(
		"unexpectedPath",
		fmt.Sprintf("Unexpected path \"%s\" imported in restricted zone.", specifier),
		line, column, column+len(specifier)+2,
	)
}

func unexpectedPathWithMessage(specifier string, custom string, line int, column int) rule_tester.InvalidTestCaseError {
	return errorAt(
		"unexpectedPath",
		fmt.Sprintf("Unexpected path \"%s\" imported in restricted zone. %s", specifier, custom),
		line, column, column+len(specifier)+2,
	)
}

func invalidExceptionPath(specifier string, line int, column int) rule_tester.InvalidTestCaseError {
	return errorAt(
		"invalidExceptionPath",
		"Restricted path exceptions must be descendants of the configured `from` path for that zone.",
		line, column, column+len(specifier)+2,
	)
}

func invalidExceptionGlob(specifier string, line int, column int) rule_tester.InvalidTestCaseError {
	return errorAt(
		"invalidExceptionGlob",
		"Restricted path exceptions must be glob patterns when `from` contains glob patterns",
		line, column, column+len(specifier)+2,
	)
}

func mixedGlobAndNonGlob(specifier string, line int, column int) rule_tester.InvalidTestCaseError {
	return errorAt(
		"mixedGlobAndNonGlob",
		"Restricted path `from` must contain either only glob patterns or none",
		line, column, column+len(specifier)+2,
	)
}

func TestNoRestrictedPathsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_restricted_paths.NoRestrictedPathsRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: zone does not cover the imported path ----
			{
				Code:     `import a from "../client/a"`,
				FileName: "restricted-paths/server/b.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server",
					"from":   "./restricted-paths/other",
				}),
			},
			{
				Code:     `import a from "../client/a"`,
				FileName: "restricted-paths/server/b.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/other",
				}),
			},
			{
				Code:     `import a from "../client/a"`,
				FileName: "restricted-paths/client/b.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/!(client)/**/*",
					"from":   "./restricted-paths/client/**/*",
				}),
			},
			{
				// CommonJS specifiers only take part in module resolution inside
				// JavaScript files, so the require cases run on a `.js` fixture.
				Code:     `const a = require("../client/a")`,
				FileName: "restricted-paths/server/consumer.js",
				TSConfig: "tsconfig.allow-js.json",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server",
					"from":   "./restricted-paths/other",
				}),
			},
			{
				Code:     `import b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/other",
				}),
			},

			// ---- upstream valid: except lifts the restriction ----
			{
				Code:     `import a from "./a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server",
					"except": list("./one"),
				}),
			},
			{
				Code:     `import a from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server",
					"except": list("./two"),
				}),
			},
			{
				Code:     `import a from "../one/a"`,
				FileName: "restricted-paths/server/two-new/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/two",
					"from":   "./restricted-paths/server",
					"except": list(),
				}),
			},
			{
				Code:     `import A from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
					"except": list("**/a.ts"),
				}),
			},

			// ---- upstream valid: support of arrays for from and target ----
			{
				Code:     `import a from "../client/a"`,
				FileName: "restricted-paths/server/b.ts",
				Options: zones(map[string]interface{}{
					"target": list("./restricted-paths/server"),
					"from":   "./restricted-paths/other",
				}),
			},
			{
				Code:     `import a from "../client/a"`,
				FileName: "restricted-paths/server/b.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server",
					"from":   list("./restricted-paths/other"),
				}),
			},
			{
				Code:     `import a from "../one/a"`,
				FileName: "restricted-paths/server/two-new/a.ts",
				Options: zones(map[string]interface{}{
					"target": list("./restricted-paths/server/two", "./restricted-paths/server/three"),
					"from":   "./restricted-paths/server",
				}),
			},
			{
				Code:     `import a from "../one/a"`,
				FileName: "restricted-paths/server/two-new/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server",
					"from":   list("./restricted-paths/server/two", "./restricted-paths/server/three"),
					"except": list(),
				}),
			},
			{
				Code:     `import a from "../client/a"`,
				FileName: "restricted-paths/client/b.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/!(client)/**/*",
					"from":   list("./restricted-paths/client/*", "./restricted-paths/client/one/*"),
				}),
			},
			{
				Code:     `import a from "../client/a"`,
				FileName: "restricted-paths/client/b.ts",
				Options: zones(map[string]interface{}{
					"target": list("./restricted-paths/!(client)/**/*", "./restricted-paths/client/a/"),
					"from":   "./restricted-paths/client/**/*",
				}),
			},

			// ---- upstream valid: irrelevant function calls ----
			{Code: `notRequire("../server/b")`},
			{
				Code:     `notRequire("../server/b")`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
			},

			// ---- upstream valid: no config ----
			{Code: `require("../server/b")`},
			{Code: `import b from "../server/b"`},

			// ---- upstream valid: builtin (ignore) ----
			{Code: `require("os")`},

			// ---- upstream valid (Typescript): type-only imports ----
			{
				Code:     `import type a from "../client/a"`,
				FileName: "restricted-paths/server/b.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server",
					"from":   "./restricted-paths/other",
				}),
			},
			{
				Code:     `import type a from "../client/a"`,
				FileName: "restricted-paths/server/b.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/other",
				}),
			},
			{
				Code:     `import type a from "../client/a"`,
				FileName: "restricted-paths/client/b.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/!(client)/**/*",
					"from":   "./restricted-paths/client/**/*",
				}),
			},
			{
				Code:     `import type b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/other",
				}),
			},
			{
				Code:     `import type a from "./a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server",
					"except": list("./one"),
				}),
			},
			{
				Code:     `import type a from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server",
					"except": list("./two"),
				}),
			},
			{
				Code:     `import type a from "../one/a"`,
				FileName: "restricted-paths/server/two-new/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/two",
					"from":   "./restricted-paths/server",
					"except": list(),
				}),
			},
			{
				Code:     `import type A from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
					"except": list("**/a.ts"),
				}),
			},
			{Code: `import type b from "../server/b"`},
			{Code: `import type * as b from "../server/b"`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid: directory and glob targets ----
			{
				Code:     `import b from "../server/b"; // 1`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},
			{
				Code:     `import b from "../server/b"; // 2`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client/**/*",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},
			{
				Code:     `import b from "../server/b";`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client/*.ts",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},
			{
				Code:     `import b from "../server/b"; // 2 ter`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client/**",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// ---- upstream invalid: several zones, one report each ----
			{
				Code:     "import a from \"../client/a\"\nimport c from \"./c\"",
				FileName: "restricted-paths/server/b.ts",
				Options: zones(
					map[string]interface{}{
						"target": "./restricted-paths/server",
						"from":   "./restricted-paths/client",
					},
					map[string]interface{}{
						"target": "./restricted-paths/server",
						"from":   "./restricted-paths/server/c.ts",
					},
				),
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedPath("../client/a", 1, 15),
					unexpectedPath("./c", 2, 15),
				},
			},

			// ---- upstream invalid: basePath ----
			{
				Code:     `import b from "../server/b"; // 3`,
				FileName: "restricted-paths/client/a.ts",
				Options: zonesWithBasePath("/restricted-paths", map[string]interface{}{
					"target": "./client",
					"from":   "./server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// ---- upstream invalid: commonjs require ----
			{
				Code:     `const b = require("../server/b")`,
				FileName: "restricted-paths/client/consumer.js",
				TSConfig: "tsconfig.allow-js.json",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 19)},
			},

			// ---- upstream invalid: except and message ----
			{
				Code:     `import b from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server",
					"except": list("./one"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../two/a", 1, 15)},
			},
			{
				Code:     `import b from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target":  "./restricted-paths/server/one",
					"from":    "./restricted-paths/server",
					"except":  list("./one"),
					"message": "Custom message",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedPathWithMessage("../two/a", "Custom message", 1, 15),
				},
			},
			{
				Code:     `import b from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server",
					"except": list("../client/a"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{invalidExceptionPath("../two/a", 1, 15)},
			},

			// ---- upstream invalid: glob from ----
			{
				Code:     `import A from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../two/a", 1, 15)},
			},
			{
				Code:     `import A from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
					"except": list("a.ts"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{invalidExceptionGlob("../two/a", 1, 15)},
			},

			// ---- upstream invalid: support of arrays for from and target ----
			{
				Code:     `import b from "../server/b"; // 4`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": list("./restricted-paths/client"),
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},
			{
				Code:     `import b from "../server/b"; // 5`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   list("./restricted-paths/server"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},
			{
				Code:     `import b from "../server/b"; // 6`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": list("./restricted-paths/client/one", "./restricted-paths/client"),
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},
			{
				Code:     "import b from \"../server/one/b\"\nimport a from \"../server/two/a\"",
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   list("./restricted-paths/server/one", "./restricted-paths/server/two"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedPath("../server/one/b", 1, 15),
					unexpectedPath("../server/two/a", 2, 15),
				},
			},
			{
				Code:     "import b from \"../server/one/b\"\nimport a from \"../server/two/a\"",
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   list("./restricted-paths/server/one/*", "./restricted-paths/server/two/*"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedPath("../server/one/b", 1, 15),
					unexpectedPath("../server/two/a", 2, 15),
				},
			},
			{
				Code:     `import b from "../server/b"; // 7`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": list("./restricted-paths/client/one", "./restricted-paths/client/**/*"),
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 15)},
			},

			// ---- upstream invalid: configuration format ----
			{
				Code:     `import A from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   list("./restricted-paths/server/**/*"),
					"except": list("a.ts"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{invalidExceptionGlob("../two/a", 1, 15)},
			},
			{
				Code:     `import b from "../server/one/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   list("./restricted-paths/server/one", "./restricted-paths/server/two/*"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{mixedGlobAndNonGlob("../server/one/b", 1, 15)},
			},

			// ---- upstream invalid (Typescript): type-only imports ----
			{
				Code:     `import type b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 20)},
			},
			{
				Code:     `import type b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/client/**/*",
					"from":   "./restricted-paths/server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 20)},
			},
			{
				Code:     "import type a from \"../client/a\"\nimport type c from \"./c\"",
				FileName: "restricted-paths/server/b.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server",
					"from":   list("./restricted-paths/client", "./restricted-paths/server/c.ts"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedPath("../client/a", 1, 20),
					unexpectedPath("./c", 2, 20),
				},
			},
			{
				Code:     `import type b from "../server/b"`,
				FileName: "restricted-paths/client/a.ts",
				Options: zonesWithBasePath("/restricted-paths", map[string]interface{}{
					"target": "./client",
					"from":   "./server",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../server/b", 1, 20)},
			},
			{
				Code:     `import type b from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server",
					"except": list("./one"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../two/a", 1, 20)},
			},
			{
				Code:     `import type b from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target":  "./restricted-paths/server/one",
					"from":    "./restricted-paths/server",
					"except":  list("./one"),
					"message": "Custom message",
				}),
				Errors: []rule_tester.InvalidTestCaseError{
					unexpectedPathWithMessage("../two/a", "Custom message", 1, 20),
				},
			},
			{
				Code:     `import type b from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "./restricted-paths/server/one",
					"from":   "./restricted-paths/server",
					"except": list("../client/a"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{invalidExceptionPath("../two/a", 1, 20)},
			},
			{
				Code:     `import type A from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
				}),
				Errors: []rule_tester.InvalidTestCaseError{unexpectedPath("../two/a", 1, 20)},
			},
			{
				Code:     `import type A from "../two/a"`,
				FileName: "restricted-paths/server/one/a.ts",
				Options: zones(map[string]interface{}{
					"target": "**/*",
					"from":   "./restricted-paths/server/**/*",
					"except": list("a.ts"),
				}),
				Errors: []rule_tester.InvalidTestCaseError{invalidExceptionGlob("../two/a", 1, 20)},
			},
		},
	)
}
