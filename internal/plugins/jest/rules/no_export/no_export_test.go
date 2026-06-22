package no_export_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExportRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_export.NoExportRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe("a test", () => { expect(1).toBe(1); })`},
			{Code: `window.location = "valid"`},
			{Code: `module.somethingElse = "foo";`},
			{Code: `export const myThing = "valid"`},
			{Code: `export default function () {}`},
			{Code: `module.exports = function(){}`},
			{Code: `module.exports.myThing = "valid";`},
			{Code: `module.export.foo = "valid"; test("a test", () => {});`},
			{Code: `const exports = "exports"; module[exports] = {}; test("a test", () => {});`},
			{Code: `const module = { exports: {} }; module.exports.foo = "valid"; test("a test", () => {});`},
			{Code: `const run = (module: { exports: object }) => { module.exports.foo = "valid" }; test("a test", () => {});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `export const myThing = "invalid"; test("a test", () => { expect(1).toBe(1);});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 34},
				},
			},
			{
				Code: `
export const myThing = 'invalid';

test.each()('my code', () => {
  expect(1).toBe(1);
});
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 2, Column: 1, EndColumn: 34},
				},
			},
			{
				Code: `
export const myThing = 'invalid';

test.each` + "`" + "`('my code', () => {\n  expect(1).toBe(1);\n});\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 2, Column: 1, EndColumn: 34},
				},
			},
			{
				Code: `
export const myThing = 'invalid';

test.only.each` + "`" + "`('my code', () => {\n  expect(1).toBe(1);\n});\n",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 2, Column: 1, EndColumn: 34},
				},
			},
			{
				Code: `export default function() {};  test("a test", () => { expect(1).toBe(1);});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 29},
				},
			},
			{
				Code: `export = function() {}; test("a test", () => { expect(1).toBe(1);});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 24},
				},
			},
			{
				Code: `module.exports["invalid"] = function() {};  test("a test", () => { expect(1).toBe(1);});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 26},
				},
			},
			{
				Code: `module.exports = function() {}; ;  test("a test", () => { expect(1).toBe(1);});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 15},
				},
			},
			{
				Code: `module["exports"] = function() {}; test("a test", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 18},
				},
			},
			{
				Code: "module[`exports`].foo = function() {}; test(\"a test\", () => {});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 22},
				},
			},
			{
				Code: `module.exports.foo.bar = function() {}; test("a test", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 23},
				},
			},
			{
				Code: `module.exports ||= {}; test("a test", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 15},
				},
			},
			{
				Code: `value = module.exports; test("a test", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 9, EndColumn: 23},
				},
			},
			{
				Code: `export import foo = require("./foo"); test("a test", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 38},
				},
			},
			{
				Code: `export const myThing = "invalid"; describe("a suite", () => {});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedExport", Line: 1, Column: 1, EndColumn: 34},
				},
			},
		},
	)
}
