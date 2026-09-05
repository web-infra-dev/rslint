// TestNoArrayFromFillExtras covers tsgo edge shapes, every rule branch, and
// real-user examples absent from the upstream suite. The exact upstream
// migration lives in no_array_from_fill_upstream_test.go.
package no_array_from_fill_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_array_from_fill"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoArrayFromFillExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_array_from_fill.NoArrayFromFillRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream isArrayFromFillCall() arm 2: only a direct
			// Array.from(...) receiver reaches isArrayFromLengthCall().
			extraValid(`Array.from({length: 3}).slice().fill(0)`, "file.js"),

			// ---- Dimension 4: authored TypeScript wrappers remain visible ----
			extraValid(`(Array.from({length: 3}) as unknown[]).fill(0)`, "file.ts"),
			extraValid(`Array!.from({length: 3}).fill(0)`, "file.ts"),
			extraValid(`(Array satisfies typeof Array).from({length: 3}).fill(0)`, "file.ts"),

			// ---- Dimension 4: optional member access does not match ----
			extraValid(`Array?.from({length: 3}).fill(0)`, "file.js"),

			// Locks in upstream isMethodCall() computed:false and
			// allowSpreadElement:false branches for both calls.
			extraValid(`Array["from"]({length: 3}).fill(0)`, "file.js"),
			extraValid("Array[`from`]({length: 3}).fill(0)", "file.js"),
			extraValid(`Array[0]({length: 3}).fill(0)`, "file.js"),
			extraValid(`Array[Symbol.from]({length: 3}).fill(0)`, "file.js"),
			extraValid(`Array.from({length: 3})["fill"](0)`, "file.js"),
			extraValid(`Array.from(...[{length: 3}]).fill(0)`, "file.js"),
			extraValid(`Array.from({length: 3}).fill(...[])`, "file.js"),

			// ---- Dimension 4: computed and numeric length-key forms are rejected ----
			extraValid(`Array.from({["length"]: 3}).fill(0)`, "file.js"),
			extraValid(`Array.from({0: 3}).fill(0)`, "file.js"),

			// Locks in upstream isLengthProperty() method/kind branches: methods
			// and accessors are not ordinary init data properties.
			extraValid(`Array.from({length() { return 3; }}).fill(0)`, "file.js"),
			extraValid(`Array.from({get length() { return 3; }}).fill(0)`, "file.js"),

			// Locks in upstream isArrayFromLengthCall() global-reference guard:
			// only the direct environment Array binding is accepted.
			extraValid(`globalThis.Array.from({length: 3}).fill(0)`, "file.js"),
			{
				Code:     `Array.from({length: 3}).fill(0)`,
				FileName: "file.js",
				Globals:  map[string]any{"Array": "off"},
			},

			// ---- Dimension 4: SpreadAssignment degrades without masking shape checks ----
			extraValid(`Array.from({...source}).fill(0)`, "file.js"),

			// N/A: PrivateIdentifier keys cannot appear in object literals.
			// N/A: declaration/container forms; the rule targets call expressions.
			// N/A: rest binding patterns; the rule does not inspect bindings.
			// N/A: empty/abstract/declare bodies; the rule does not inspect bodies.
			// N/A: rule-specific ancestor walks; the rule performs none.
			// N/A: autofix boundaries; the rule has no fix or suggestion.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: single/multi-level receiver and argument parentheses ----
			extraInvalid(`(Array).from(({length: 3})).fill(0)`, "file.js"),
			extraInvalid(`((Array.from({length: 3}))).fill()`, "file.js"),

			// ---- Dimension 4: TypeScript type arguments do not wrap the callee ----
			extraInvalid(`Array.from<number>({length: 3}).fill(0)`, "file.ts"),

			// Locks in upstream isLengthPropertyKey() arm 1: identifier key.
			extraInvalid(`Array.from({length}).fill(undefined)`, "file.js"),
			// Locks in upstream isLengthPropertyKey() arm 2: string-literal key.
			extraInvalid(`Array.from({'length': 3}).fill(0)`, "file.js"),

			// ---- Dimension 4: same-kind call nesting reports only the inner fill ----
			extraInvalid(`consume(Array.from({length: 3}).fill(0))`, "file.js"),

			// ---- Real-user: issue #2943 multiline fill().map() chain ----
			extraInvalid(
				"Array.from({ length: 5 })\n\t.fill(null)\n\t.map((_, index) => `Item number: ${index}`);",
				"file.js",
			),
			// ---- Real-user: issue #2943 shared object fill value ----
			extraInvalid(`Array.from({ length: 5 }).fill({ something: "special" })`, "file.js"),
		},
	)
}

func extraValid(code, fileName string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: fileName}
}

func extraInvalid(code, fileName string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: fileName,
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, "fill", 0),
		},
	}
}
