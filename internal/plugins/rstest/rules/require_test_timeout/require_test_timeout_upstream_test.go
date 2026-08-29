// TestRequireTestTimeoutUpstream migrates the upstream Vitest suite one case at
// a time. Cases upstream reports but this port deliberately accepts are kept
// here, in the valid list, each labelled with what upstream does and why this
// rule stays silent.
package require_test_timeout_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_test_timeout"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const missingTimeoutMessage = "Test is missing a timeout. Add an explicit timeout."

// invalidCase builds the expected diagnostics by locating the source text of
// each reported registration, so a case stays readable while still asserting
// the whole range the rule reports.
func invalidCase(code string, calls ...string) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(calls))
	searchFrom := 0
	for _, call := range calls {
		offset := strings.Index(code[searchFrom:], call)
		if offset < 0 {
			panic("registration not found in test code: " + call)
		}
		start := searchFrom + offset
		searchFrom = start + len(call)
		line, column := positionOf(code, start)
		endLine, endColumn := positionOf(code, start+len(call))
		errors = append(errors, rule_tester.InvalidTestCaseError{
			MessageId: "missingTimeout",
			Message:   missingTimeoutMessage,
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
		})
	}
	return rule_tester.InvalidTestCase{Code: code, Errors: errors}
}

func positionOf(code string, offset int) (int, int) {
	line, column := 1, 1
	for _, character := range code[:offset] {
		if character == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}

func TestRequireTestTimeoutUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_test_timeout.RequireTestTimeoutRule,
		[]rule_tester.ValidTestCase{
			{Code: `test.todo("a")`},
			// Upstream skips names beginning with `x`. Rstest exports no `xit`,
			// so this call registers nothing the rule knows about.
			{Code: `xit("a", () => {})`},
			{Code: `test("a", () => {}, 0)`},
			{Code: `it("a", () => {}, 500)`},
			{Code: `it.skip("a", () => {})`},
			{Code: `test.skip("a", () => {})`},
			{Code: `test("a", () => {}, 1000)`},
			{Code: `it.only("a", () => {}, 1234)`},
			{Code: `test.only("a", () => {}, 1234)`},
			{Code: `it.concurrent("a", () => {}, 400)`},
			{Code: `test("a", () => {}, { timeout: 0 })`},
			{Code: `test.concurrent("a", () => {}, 400)`},
			{Code: `test("a", () => {}, { timeout: 500 })`},
			{Code: `test("a", { timeout: 500 }, () => {})`},
			{Code: `const t = 500; test("a", { timeout: t }, () => {})`},
			{Code: `const t = 500; test("a", () => {}, t)`},
			{Code: `const opts = { timeout: 500 }; test("a", opts, () => {})`},
			{Code: `const T = 1000; rs.setConfig({ testTimeout: T }); test("a", () => {})`},
			{Code: `rs.setConfig({ testTimeout: 1000 }); test("a", () => {})`},

			// Mirrored `it` variants.
			{Code: `const t = 500; it("a", { timeout: t }, () => {})`},
			{Code: `const t = 500; it("a", () => {}, t)`},
			{Code: `const opts = { timeout: 500 }; it("a", opts, () => {})`},
			{Code: `const T = 1000; rs.setConfig({ testTimeout: T }); it("a", () => {})`},
			{Code: `rs.setConfig({ testTimeout: 1000 }); it("a", () => {})`},

			// More arguments than either Rstest overload takes.
			{Code: `test("a", { foo: 1 }, { timeout: 500 }, () => {})`},
			{Code: `test("a", { timeout: 500 }, 1000, () => {})`},
			{Code: `test("a", () => {}, 1000, { extra: true })`},

			// In-source unit tests.
			{Code: `if (import.meta.rstest) { const opts = { timeout: 500 }; describe("outer", () => { it("repro: same-file opts object", opts, () => {}); }); }`},
			{Code: `if (import.meta.rstest) { const T = 500; describe("outer", () => { describe("inner", () => { it("repro: same-file const timeout", () => {}, T); }); }); }`},

			// An imported binding named in the timeout position.
			{Code: `import { TIMEOUT } from "./test-constants"; test("a", () => {}, TIMEOUT)`},
			{Code: `import { TIMEOUT } from "./test-constants"; it("a", () => {}, TIMEOUT)`},
			{Code: `import { TIMEOUT } from "./test-constants"; test("a", () => {}, { timeout: TIMEOUT })`},
			{Code: `import { TIMEOUT } from "./test-constants"; test("a", { timeout: TIMEOUT }, () => {})`},
			{Code: `import { OPTS } from "./test-constants"; test("a", OPTS, () => {})`},
			{Code: `import T from "./test-constants"; test("a", () => {}, T)`},

			// ---- Upstream reports these; this port stays silent ----
			// Upstream resolves the timeout identifier and reports every
			// binding it fails to prove is a numeric `const`. A binding this
			// rule cannot read still means the author wrote a timeout down, so
			// reading it wrong is the only way to produce a false report.
			{Code: `let t = 500; test("a", () => {}, t)`},
			{Code: `let t = 500; it("a", () => {}, t)`},
			{Code: `const t = getTimeout(); test("a", () => {}, t)`},
			{Code: `const t = getTimeout(); it("a", () => {}, t)`},
			// Upstream's `Number.NaN` branch reports a `timeout` property whose
			// value is not a number. TypeScript already rejects these, so the
			// only thing a lint report adds is noise.
			{Code: `test("a", () => {}, { timeout: null })`},
			{Code: `test("a", () => {}, { timeout: undefined })`},
			{Code: `rs.setConfig({ testTimeout: null }); test("a", () => {})`},
			{Code: `rs.setConfig({ testTimeout: undefined }); test("a", () => {})`},
			// A spread can carry `timeout`, and upstream reports it anyway.
			{Code: `const opts = { timeout: 1000 }; test("a", { ...opts }, () => {})`},
			{Code: `const opts = { timeout: 1000 }; test("a", { ...opts }, { foo: 1 }, () => {})`},
			// Upstream scans every argument for a negative timeout. None of
			// these calls matches an Rstest overload at all.
			{Code: `test("a", () => {}, { timeout: -1 }, { timeout: 500 })`},
			{Code: `test("a", () => {}, { timeout: 500 }, { timeout: -1 })`},
			{Code: `test("a", () => {}, { timeout: -1 }, 1000)`},
			{Code: `test("a", () => {}, 1000, { timeout: -1 })`},
		},
		[]rule_tester.InvalidTestCase{
			invalidCase(`test("a", () => {})`, `test("a", () => {})`),
			invalidCase(`it("a", () => {})`, `it("a", () => {})`),
			invalidCase(`test.only("a", () => {})`, `test.only("a", () => {})`),
			invalidCase(`test.concurrent("a", () => {})`, `test.concurrent("a", () => {})`),
			invalidCase(`it.concurrent("a", () => {})`, `it.concurrent("a", () => {})`),
			invalidCase(`rs.setConfig({}); test("a", () => {})`, `test("a", () => {})`),
			invalidCase(`test("a", () => {}, -100)`, `test("a", () => {}, -100)`),
			invalidCase(`test("a", () => {}, { timeout: -1 })`, `test("a", () => {}, { timeout: -1 })`),
			// A `setConfig` after the test does not reach back to it.
			invalidCase(`test("a", () => {}); rs.setConfig({ testTimeout: 1000 })`, `test("a", () => {})`),
			// An import elsewhere in the file is not a timeout.
			invalidCase(
				`import { TIMEOUT } from "./test-constants"; test("a", () => {})`,
				`test("a", () => {})`,
			),
		},
	)
}
