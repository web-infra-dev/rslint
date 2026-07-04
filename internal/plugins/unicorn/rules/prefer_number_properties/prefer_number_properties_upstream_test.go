package prefer_number_properties_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_number_properties"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	errorID      = "error"
	suggestionID = "suggestion"
)

// TestPreferNumberPropertiesUpstream migrates the full valid/invalid suite from
// upstream test/prefer-number-properties.js in eslint-plugin-unicorn v64.0.0
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in the prefer_number_properties_extras_test.go file.
func TestPreferNumberPropertiesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_number_properties.PreferNumberPropertiesRule,
		[]rule_tester.ValidTestCase{
			// ---- Methods ----
			{Code: `Number.parseInt("10", 2);`},
			{Code: `Number.parseFloat("10.5");`},
			{Code: `Number.isNaN(10);`},
			{Code: `Number.isFinite(10);`},

			// Shadowed
			{Code: "const parseInt = function() {};\nparseInt(\"10\", 2);"},
			{Code: "const parseFloat = function() {};\nparseFloat(\"10.5\");"},
			{Code: "const isNaN = function() {};\nisNaN(10);"},
			{Code: "const isFinite = function() {};\nisFinite(10);"},
			{Code: "const {parseInt} = Number;\nparseInt(\"10\", 2);"},
			{Code: "const {parseFloat} = Number;\nparseFloat(\"10.5\");"},
			{Code: "const {isNaN} = Number;\nisNaN(10);"},
			{Code: "const {isFinite} = Number;\nisFinite(10);"},
			{Code: "const parseInt = function() {};\nfunction inner() {\n\treturn parseInt(\"10\", 2);\n}"},
			{Code: "const parseFloat = function() {};\nfunction inner() {\n\treturn parseFloat(\"10.5\");\n}"},
			{Code: "const isNaN = function() {};\nfunction inner() {\n\treturn isNaN(10);\n}"},
			{Code: "const isFinite = function() {};\nfunction inner() {\n\treturn isFinite(10);\n}"},

			// Not read
			{Code: `global.isFinite = Number.isFinite;`},
			{Code: `global.isFinite ??= 1;`},
			{Code: `isFinite ||= 1;`},
			{Code: `[global.isFinite] = [];`},
			{Code: `[global.isFinite = 1] = [];`},
			{Code: `[[global.isFinite = 1]] = [];`},
			{Code: `[isFinite] = [];`},
			{Code: `[isFinite = 1] = [];`},
			{Code: `[[isFinite = 1]] = [];`},
			{Code: `({foo: global.isFinite} = {});`},
			{Code: `({foo: global.isFinite = 1} = {});`},
			{Code: `({foo: {bar: global.isFinite = 1}} = {});`},
			{Code: `({foo: isFinite} = {});`},
			{Code: `({foo: isFinite = 1} = {});`},
			{Code: `({foo: {bar: isFinite = 1}} = {});`},
			{Code: `delete global.isFinite;`},

			// ---- `NaN` and `Infinity` ----
			{Code: `const foo = Number.NaN;`},
			{Code: `const foo = window.Number.NaN;`},
			{Code: `const foo = bar.NaN;`},
			{Code: `const foo = nan;`},
			{Code: `const foo = "NaN";`},
			{Code: "function foo () {\n\tconst NaN = 2;\n\treturn NaN;\n}"},
			{Code: `const {NaN} = {};`},
			{Code: `const {a: NaN} = {};`},
			{Code: `const {[a]: NaN} = {};`},
			{Code: `const [NaN] = [];`},
			{Code: `function NaN() {}`},
			{Code: `const foo = function NaN() {}`},
			{Code: `function foo(NaN) {}`},
			{Code: `foo = function (NaN) {}`},
			{Code: `foo = (NaN) => {}`},
			{Code: `function foo({NaN}) {}`},
			{Code: `function foo({a: NaN}) {}`},
			{Code: `function foo({[a]: NaN}) {}`},
			{Code: `function foo([NaN]) {}`},
			{Code: `class NaN {}`},
			{Code: `const Foo = class NaN {}`},
			{Code: `class Foo {NaN(){}}`},
			{Code: `class Foo {#NaN(){}}`},
			{Code: `class Foo3 {NaN = 1}`},
			{Code: `class Foo {#NaN = 1}`},
			{Code: "NaN: for (const foo of bar) {\n\tif (a) {\n\t\tcontinue NaN;\n\t} else {\n\t\tbreak NaN;\n\t}\n}"},
			{Code: `import {NaN} from "foo"`},
			{Code: `import {NaN as NaN} from "foo"`},
			{Code: `import NaN from "foo"`},
			{Code: `import * as NaN from "foo"`},
			{Code: `export {NaN} from "foo"`},
			{Code: `export {NaN as NaN} from "foo"`},
			{Code: `export * as NaN from "foo"`},
			{Code: `const foo = Number.POSITIVE_INFINITY;`},
			{Code: `const foo = window.Number.POSITIVE_INFINITY;`},
			{Code: `const foo = bar.POSITIVE_INFINITY;`},
			{Code: `const foo = Number.Infinity;`},
			{Code: `const foo = window.Number.Infinity;`},
			{Code: `const foo = bar.Infinity;`},
			{Code: `const foo = infinity;`},
			{Code: `const foo = "Infinity";`},
			{Code: `const foo = "-Infinity";`},
			{Code: "function foo () {\n\tconst Infinity = 2;\n\treturn Infinity;\n}"},
			{Code: `const {Infinity} = {};`},
			{Code: `function Infinity() {}`},
			{Code: `class Infinity {}`},
			{Code: `class Foo { Infinity(){}}`},
			{Code: `const foo = Infinity;`},
			{Code: `const foo = -Infinity;`},
			{Code: `const foo = NaN`, Options: map[string]interface{}{"checkNaN": false}},

			// ---- TypeScript ----
			{Code: "export enum NumberSymbol {\n\tDecimal,\n\tNaN,\n}"},
			{Code: `declare var NaN: number;`},
			{Code: "interface NumberConstructor {\n\treadonly NaN: number;\n}"},
			{Code: `declare function NaN(s: string, radix?: number): number;`},
			{Code: `class Foo {NaN = 1}`},

			// ---- Upstream snapshot: valid ----
			{Code: `const foo = ++Infinity;`},
			{Code: `const foo = --Infinity;`},
			{Code: `const foo = -(--Infinity);`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Methods ----
			fixed(`parseInt("10", 2);`, `Number.parseInt("10", 2);`, "parseInt", "parseInt", 1, 1, 1, 9),
			fixed(`parseFloat("10.5");`, `Number.parseFloat("10.5");`, "parseFloat", "parseFloat", 1, 1, 1, 11),
			suggested(`isNaN(10);`, "isNaN", "isNaN", `Number.isNaN(10);`, 1, 1, 1, 6),
			suggested(`isFinite(10);`, "isFinite", "isFinite", `Number.isFinite(10);`, 1, 1, 1, 9),
			{
				Code:   "const a = parseInt(\"10\", 2);\nconst b = parseFloat(\"10.5\");\nconst c = isNaN(10);\nconst d = isFinite(10);",
				Output: []string{"const a = Number.parseInt(\"10\", 2);\nconst b = Number.parseFloat(\"10.5\");\nconst c = isNaN(10);\nconst d = isFinite(10);"},
				Errors: []rule_tester.InvalidTestCaseError{
					expected("parseInt", "parseInt", 1, 11, 1, 19),
					expected("parseFloat", "parseFloat", 2, 11, 2, 21),
					expectedSuggestion("isNaN", "isNaN", `const a = parseInt("10", 2);
const b = parseFloat("10.5");
const c = Number.isNaN(10);
const d = isFinite(10);`, 3, 11, 3, 16),
					expectedSuggestion("isFinite", "isFinite", `const a = parseInt("10", 2);
const b = parseFloat("10.5");
const c = isNaN(10);
const d = Number.isFinite(10);`, 4, 11, 4, 19),
				},
			},

			// ---- `NaN` and `Infinity` ----
			fixed(`const foo = NaN;`, `const foo = Number.NaN;`, "NaN", "NaN", 1, 13, 1, 16),
			fixed(`if (Number.isNaN(NaN)) {}`, `if (Number.isNaN(Number.NaN)) {}`, "NaN", "NaN", 1, 18, 1, 21),
			fixed(`if (Object.is(foo, NaN)) {}`, `if (Object.is(foo, Number.NaN)) {}`, "NaN", "NaN", 1, 20, 1, 23),
			fixed(`const foo = bar[NaN];`, `const foo = bar[Number.NaN];`, "NaN", "NaN", 1, 17, 1, 20),
			fixed(`const foo = {NaN};`, `const foo = {NaN: Number.NaN};`, "NaN", "NaN", 1, 14, 1, 17),
			fixed(`const foo = {NaN: NaN};`, `const foo = {NaN: Number.NaN};`, "NaN", "NaN", 1, 19, 1, 22),
			fixed(`const {foo = NaN} = {};`, `const {foo = Number.NaN} = {};`, "NaN", "NaN", 1, 14, 1, 17),
			fixed(`const foo = NaN.toString();`, `const foo = Number.NaN.toString();`, "NaN", "NaN", 1, 13, 1, 16),
			fixed(`class Foo3 {[NaN] = 1}`, `class Foo3 {[Number.NaN] = 1}`, "NaN", "NaN", 1, 14, 1, 17),
			fixedWithOptions(`const foo = Infinity;`, `const foo = Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 13, 1, 21),
			fixedWithOptions(`const foo = -Infinity;`, `const foo = Number.NEGATIVE_INFINITY;`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 13, 1, 22),

			// ---- TypeScript ----
			fixed(`class Foo {[NaN] = 1}`, `class Foo {[Number.NaN] = 1}`, "NaN", "NaN", 1, 13, 1, 16),

			// ---- Upstream snapshot: invalid ----
			fixed(`const foo = {[NaN]: 1}`, `const foo = {[Number.NaN]: 1}`, "NaN", "NaN", 1, 15, 1, 18),
			fixed(`const foo = {[NaN]() {}}`, `const foo = {[Number.NaN]() {}}`, "NaN", "NaN", 1, 15, 1, 18),
			fixed(`foo[NaN] = 1;`, `foo[Number.NaN] = 1;`, "NaN", "NaN", 1, 5, 1, 8),
			fixed(`class A {[NaN](){}}`, `class A {[Number.NaN](){}}`, "NaN", "NaN", 1, 11, 1, 14),
			fixed(`foo = {[NaN]: 1}`, `foo = {[Number.NaN]: 1}`, "NaN", "NaN", 1, 9, 1, 12),
			fixedWithOptions(`if (Number.isNaN(Infinity)) {}`, `if (Number.isNaN(Number.POSITIVE_INFINITY)) {}`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 18, 1, 26),
			fixedWithOptions(`if (Object.is(foo, Infinity)) {}`, `if (Object.is(foo, Number.POSITIVE_INFINITY)) {}`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 20, 1, 28),
			fixedWithOptions(`const foo = bar[Infinity];`, `const foo = bar[Number.POSITIVE_INFINITY];`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 17, 1, 25),
			fixedWithOptions(`const foo = {Infinity};`, `const foo = {Infinity: Number.POSITIVE_INFINITY};`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 14, 1, 22),
			fixedWithOptions(`const foo = {Infinity: Infinity};`, `const foo = {Infinity: Number.POSITIVE_INFINITY};`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 24, 1, 32),
			{
				Code:    `const foo = {[Infinity]: -Infinity};`,
				Options: map[string]interface{}{"checkInfinity": true},
				Output:  []string{`const foo = {[Number.POSITIVE_INFINITY]: Number.NEGATIVE_INFINITY};`},
				Errors: []rule_tester.InvalidTestCaseError{
					expected("Infinity", "POSITIVE_INFINITY", 1, 15, 1, 23),
					expected("-Infinity", "NEGATIVE_INFINITY", 1, 26, 1, 35),
				},
			},
			{
				Code:    `const foo = {[-Infinity]: Infinity};`,
				Options: map[string]interface{}{"checkInfinity": true},
				Output:  []string{`const foo = {[Number.NEGATIVE_INFINITY]: Number.POSITIVE_INFINITY};`},
				Errors: []rule_tester.InvalidTestCaseError{
					expected("-Infinity", "NEGATIVE_INFINITY", 1, 15, 1, 24),
					expected("Infinity", "POSITIVE_INFINITY", 1, 27, 1, 35),
				},
			},
			fixedWithOptions(`const foo = {Infinity: -Infinity};`, `const foo = {Infinity: Number.NEGATIVE_INFINITY};`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 24, 1, 33),
			fixedWithOptions(`const {foo = Infinity} = {};`, `const {foo = Number.POSITIVE_INFINITY} = {};`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 14, 1, 22),
			fixedWithOptions(`const {foo = -Infinity} = {};`, `const {foo = Number.NEGATIVE_INFINITY} = {};`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 14, 1, 23),
			fixedWithOptions(`const foo = Infinity.toString();`, `const foo = Number.POSITIVE_INFINITY.toString();`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 13, 1, 21),
			fixedWithOptions(`const foo = -Infinity.toString();`, `const foo = -Number.POSITIVE_INFINITY.toString();`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 14, 1, 22),
			fixedWithOptions(`const foo = (-Infinity).toString();`, `const foo = (Number.NEGATIVE_INFINITY).toString();`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 14, 1, 23),
			fixedWithOptions(`const foo = +Infinity;`, `const foo = +Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 14, 1, 22),
			fixedWithOptions(`const foo = +-Infinity;`, `const foo = +Number.NEGATIVE_INFINITY;`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 14, 1, 23),
			fixedWithOptions(`const foo = -(-Infinity);`, `const foo = -(Number.NEGATIVE_INFINITY);`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 15, 1, 24),
			fixedWithOptions(`const foo = 1 - Infinity;`, `const foo = 1 - Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 17, 1, 25),
			fixedWithOptions(`const foo = 1 - -Infinity;`, `const foo = 1 - Number.NEGATIVE_INFINITY;`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 17, 1, 26),
			fixedWithOptions(`const isPositiveZero = value => value === 0 && 1 / value === Infinity;`, `const isPositiveZero = value => value === 0 && 1 / value === Number.POSITIVE_INFINITY;`, "Infinity", "POSITIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 62, 1, 70),
			fixedWithOptions(`const isNegativeZero = value => value === 0 && 1 / value === -Infinity;`, `const isNegativeZero = value => value === 0 && 1 / value === Number.NEGATIVE_INFINITY;`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 62, 1, 71),
			fixed(`const {a = NaN} = {};`, `const {a = Number.NaN} = {};`, "NaN", "NaN", 1, 12, 1, 15),
			{
				Code:   `const {[NaN]: a = NaN} = {};`,
				Output: []string{`const {[Number.NaN]: a = Number.NaN} = {};`},
				Errors: []rule_tester.InvalidTestCaseError{
					expected("NaN", "NaN", 1, 9, 1, 12),
					expected("NaN", "NaN", 1, 19, 1, 22),
				},
			},
			fixed(`const [a = NaN] = [];`, `const [a = Number.NaN] = [];`, "NaN", "NaN", 1, 12, 1, 15),
			fixed(`function foo({a = NaN}) {}`, `function foo({a = Number.NaN}) {}`, "NaN", "NaN", 1, 19, 1, 22),
			{
				Code:   `function foo({[NaN]: a = NaN}) {}`,
				Output: []string{`function foo({[Number.NaN]: a = Number.NaN}) {}`},
				Errors: []rule_tester.InvalidTestCaseError{
					expected("NaN", "NaN", 1, 16, 1, 19),
					expected("NaN", "NaN", 1, 26, 1, 29),
				},
			},
			fixed(`function foo([a = NaN]) {}`, `function foo([a = Number.NaN]) {}`, "NaN", "NaN", 1, 19, 1, 22),
			fixedWithOptions(`function foo() {return-Infinity}`, `function foo() {return Number.NEGATIVE_INFINITY}`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 23, 1, 32),
			suggested(`globalThis.isNaN(foo);`, "isNaN", "isNaN", `Number.isNaN(foo);`, 1, 1, 1, 17),
			suggested(`global.isNaN(foo);`, "isNaN", "isNaN", `Number.isNaN(foo);`, 1, 1, 1, 13),
			suggested(`window.isNaN(foo);`, "isNaN", "isNaN", `Number.isNaN(foo);`, 1, 1, 1, 13),
			suggested(`self.isNaN(foo);`, "isNaN", "isNaN", `Number.isNaN(foo);`, 1, 1, 1, 11),
			fixed(`globalThis.parseFloat(foo);`, `Number.parseFloat(foo);`, "parseFloat", "parseFloat", 1, 1, 1, 22),
			fixed(`global.parseFloat(foo);`, `Number.parseFloat(foo);`, "parseFloat", "parseFloat", 1, 1, 1, 18),
			fixed(`window.parseFloat(foo);`, `Number.parseFloat(foo);`, "parseFloat", "parseFloat", 1, 1, 1, 18),
			fixed(`self.parseFloat(foo);`, `Number.parseFloat(foo);`, "parseFloat", "parseFloat", 1, 1, 1, 16),
			fixed(`globalThis.NaN`, `Number.NaN`, "NaN", "NaN", 1, 1, 1, 15),
			fixedWithOptions(`-globalThis.Infinity`, `Number.NEGATIVE_INFINITY`, "-Infinity", "NEGATIVE_INFINITY", map[string]interface{}{"checkInfinity": true}, 1, 1, 1, 21),
			{
				Code:   "const options = {\n\tnormalize: parseFloat,\n\tparseInt,\n};\n\nrun(foo, options);",
				Output: []string{"const options = {\n\tnormalize: Number.parseFloat,\n\tparseInt: Number.parseInt,\n};\n\nrun(foo, options);"},
				Errors: []rule_tester.InvalidTestCaseError{
					expected("parseFloat", "parseFloat", 2, 13, 2, 23),
					expected("parseInt", "parseInt", 3, 2, 3, 10),
				},
			},
		},
	)
}

func fixed(code, output, description, property string, line, column, endLine, endColumn int) rule_tester.InvalidTestCase {
	return fixedWithOptions(code, output, description, property, nil, line, column, endLine, endColumn)
}

func fixedWithOptions(code, output, description, property string, options any, line, column, endLine, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:    code,
		Options: options,
		Output:  []string{output},
		Errors:  []rule_tester.InvalidTestCaseError{expected(description, property, line, column, endLine, endColumn)},
	}
}

func suggested(code, description, property, suggestionOutput string, line, column, endLine, endColumn int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:   code,
		Errors: []rule_tester.InvalidTestCaseError{expectedSuggestion(description, property, suggestionOutput, line, column, endLine, endColumn)},
	}
}

func expected(description, property string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: errorID,
		Message:   fmt.Sprintf("Prefer `Number.%s` over `%s`.", property, description),
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

func expectedSuggestion(description, property, output string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	err := expected(description, property, line, column, endLine, endColumn)
	err.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
		MessageId: suggestionID,
		Output:    output,
	}}
	return err
}
