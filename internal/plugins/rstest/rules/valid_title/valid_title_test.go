package valid_title_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/valid_title"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	patMustNotHashBad = `(?:#(?!unit|e2e))\w+`
	patMustMatchHash  = `^[^#]+$|(?:#(?:unit|e2e))`
	patHashTag        = `#(?:unit|integration|e2e)`
)

func TestValidTitleRule(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Registration shapes: all three overloads keep the title at
		// arguments[0], so none of them changes what this rule reads.
		{Code: `test("case", () => {});`},
		{Code: `it("case", () => {});`},
		{Code: `describe("suite", () => {});`},
		{Code: `test("case", { timeout: 100 }, () => {});`},
		{Code: `test("case", () => {}, 500);`},
		{Code: `describe("suite");`},
		{Code: `describe();`},
		{Code: `test();`},
		{Code: "test(`case`, () => {});"},
		{Code: "test(`case ${x}`, () => {});"},
		{Code: `test("a" + b, () => {});`},
		{Code: `test("is" + " a " + " string", () => {});`},
		{Code: `test("is a string" + suffix, () => {});`},
		{Code: `test(foo - "bar", () => {});`},
		{Code: `test.concurrent("case", () => {});`},
		{Code: `describe.skip("suite", () => {});`},
		{Code: `test.skipIf(cond)("case", () => {});`},

		// API sources.
		{Code: `import { test } from "@rstest/core";
test("case", () => {});`},
		{Code: `import { test as check } from "@rstest/core";
check("case", () => {});`},
		{Code: `const { describe } = require("@rstest/core");
describe("suite", () => {});`},
		{Code: `import * as rstest from "@rstest/core";
rstest.test("case", () => {});`},
		{Code: `import.meta.rstest.test("case", () => {});`},
		{Code: `const api = import.meta.rstest;
api.test("case", () => {});`},
		{Code: `import { test } from "@rstest/playwright";
test("case", () => {});`},
		{Code: `import { test } from "@rstest/playwright";
test.describe("test authentication", () => {});`},
		{
			Code: `import { test } from "@rstest/playwright";
test.describe("suite", () => {});`,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": map[string]interface{}{"test": "^suite$"},
				"mustMatch": map[string]interface{}{
					"describe": "^suite$",
					"test":     "^case$",
				},
			}},
		},

		// A factory call is not a registration, so its first argument — here a
		// fixtures object — must not be taken for a title.
		{Code: `const myTest = test.extend({ fixture: 1 });`},
		{Code: `const myTest = test.extend({ fixture: 1 });
myTest("case", () => {});`},
		{Code: `const cases = test.each([1, 2]);`},

		// Not Rstest.
		{Code: `import { test } from "vitest";
test("test foo", () => {});`},
		{Code: `import { test } from "@jest/globals";
test("test foo", () => {});`},
		{Code: `import { test } from "@playwright/test";
test("test foo", () => {});`},
		{Code: `import { test } from "node:test";
test("test foo", () => {});`},
		{Code: `const test = createRunner();
test("test foo", () => {});`},
		{Code: `someFn("test foo", function () {})`},
		{Code: `someFn("", function () {})`},

		// Titles that merely contain the keyword.
		{Code: `test("foo test", () => {});`},
		{Code: `it("foos it correctly", () => {});`},
		{Code: `describe("foo", () => {
  it("describes things correctly", () => {});
});`},

		{
			Code:    "test(`GIVEN... \n  `, () => {});",
			Options: []interface{}{map[string]interface{}{"ignoreSpaces": true}},
		},
		{
			Code:    `describe(myFunction, () => {});`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfDescribeName": true}},
		},
		{
			Code:    `test(String(/.+/), () => {});`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfTestName": true}},
		},
		{
			Code:    `const foo = "my-title"; test(foo, () => {});`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfTestName": true}},
		},

		// mustMatch / disallowedWords non-triggering shapes.
		{
			Code:    `it("correctly sets the value", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": map[string]interface{}{}}},
		},
		{
			Code:    `it("correctly sets the value", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": ` `}},
		},
		{
			Code:    `it("correctly sets the value #unit", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": patHashTag}},
		},
		{
			Code:    `it("correctly sets the value", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": map[string]interface{}{"test": patHashTag}}},
		},
		{
			Code: `describe("things to test", () => {
  describe("unit tests #unit", () => {
    it("is true", () => {});
  });

  describe("e2e tests #e2e", () => {
    it("is another test #rstest4life", () => {});
  });
});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": map[string]interface{}{"test": patMustMatchHash}}},
		},

		// Percent signs outside of a parameterized registration are plain text.
		{Code: `test("returns 100%", () => {});`},
		{Code: `test("returns a %percent", () => {});`},

		// D1, positive side: Rstest's formatRegExp.
		{Code: `test.each([])("%s %d %i %f %j %o", () => {});`},
		{Code: `test.each([])("%O", () => {});`},
		{Code: `test.each([])("%c", () => {});`},
		{Code: `test.each([])("%# %$", () => {});`},
		{Code: `test.each([])("%%", () => {});`},
		{Code: `test.each([])("x%%y", () => {});`},
		{Code: `test.each([])("%int", () => {});`},
		{Code: `test.only.each([])("%%%%", () => {});`},
		{
			Code: `test.each([
  [1, 1, 2],
  [1, 2, 3],
])(".add(%i, %i) = %i", (a, b, expected) => {});`,
		},
		{
			Code: `describe.skip.each([
  [1, 1, 2],
])(".add(%i, %i)", (a, b, expected) => {});`,
		},
		{Code: `test.each([{ a: 1, b: 1 }])(".add($a, $b)", ({ a, b }) => {});`},

		// D2, positive side: .for goes through the same formatName.
		{Code: `test.for([])("%s", () => {});`},
		{Code: `describe.for([])("%O", () => {});`},

		// D3: a tagged-template table interpolates with $var, so its title is
		// not printf-formatted and bare percent signs are literal.
		{Code: "test.each`\n  a    | b\n  ${1} | ${2}\n`(\"returns $a and $b\", ({ a, b }) => {});"},
		{Code: "test.each`\n  a\n  ${1}\n`(\"100%coverage\", ({ a }) => {});"},

		// Known gap, shared with the jest rule: a factory stored in a variable
		// loses the syntactic evidence that the title is printf-formatted.
		{Code: `const f = test.each([1]);
f("%z", () => {});`},

		// A non-parameterized factory must not be mistaken for one just because
		// its callee is also a call expression.
		{Code: `test.runIf(cond)("case", () => {});`},
		{Code: `test.skipIf(cond)("%z", () => {});`},
	}
	for _, param := range []string{"s", "d", "i", "f", "j", "o", "O", "c", "#", "$"} {
		valid = append(valid,
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`test.each([])("%%%s", () => {})`, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`test.each([])("%%%%%s", () => {})`, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`test.each([])("%%%s!", () => {})`, param)},
			rule_tester.ValidTestCase{Code: fmt.Sprintf(`test.for([])("%%%s", () => {})`, param)},
		)
	}

	invalid := []rule_tester.InvalidTestCase{
		// invalidPattern short-circuits the whole rule.
		{
			Code:    `it("it foo", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": "(unclosed"}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidPattern",
				Message:   "Invalid regular expression in `mustMatch` option: `(unclosed`: error parsing regexp: missing closing ) in `(unclosed`",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code:    `it("it foo", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustNotMatch": map[string]interface{}{"describe": "(unterminated"}}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidPattern",
				Message:   "Invalid regular expression in `mustNotMatch.describe` option: `(unterminated`: error parsing regexp: missing closing ) in `(unterminated`",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code:    `it("it foo", () => {});`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"ok", "(unterminated"}}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidPattern",
				Message:   "Invalid regular expression in `disallowedWords` option: `\\b(ok|(unterminated)\\b`: error parsing regexp: missing closing ) in `\\b(ok|(unterminated)\\b`",
				Line:      1,
				Column:    1,
			}},
		},

		// disallowedWord reports and returns, so accidentalSpace on the same
		// title stays silent.
		{
			Code:    `test("the correct way to properly handle all things", () => {});`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"correct", "properly", "all"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code:    `it("has ALL the things", () => {});`,
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"all"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},
		{
			Code:    "test(`that the value is set properly `, () => {});",
			Options: []interface{}{map[string]interface{}{"disallowedWords": []interface{}{"properly"}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "disallowedWord"}},
		},

		// titleMustBeString.
		{
			Code:   `test(123, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `test(myTitle, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `test(String(/.+/), () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `test(1 + 2 + 3, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		// A logical operator is not concatenation, so the string literal in it
		// does not make the title static.
		{
			Code:   `test("a" || "b", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `test(cond && "b", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `test(x = "foo", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `describe(myFunction, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `describe.skip(123, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `test.each([])(1, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:   `test.for([])(1, () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		// The two ignoreTypeOf options are keyed off the block kind, so the one
		// for describe does not relax a test title.
		{
			Code:    `test.skip(123, () => {});`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfDescribeName": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},
		{
			Code:    `describe(myFunction, () => {});`,
			Options: []interface{}{map[string]interface{}{"ignoreTypeOfTestName": true}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "titleMustBeString"}},
		},

		// emptyTitle reports on the whole call, not on the argument.
		{
			Code: `describe("", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "emptyTitle",
				Message:   "describe should not have an empty title",
				Line:      1,
				Column:    1,
			}},
		},
		{
			Code: `test("", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "emptyTitle",
				Message:   "test should not have an empty title",
			}},
		},
		{
			Code: `it("", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "emptyTitle",
				Message:   "test should not have an empty title",
			}},
		},
		{
			Code:   "test(``, () => {});",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},
		{
			Code:   `import.meta.rstest.describe("", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "emptyTitle"}},
		},

		// accidentalSpace.
		{
			Code:   `describe(" foo", () => {});`,
			Output: []string{`describe("foo", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `test("foo foe fum ", () => {});`,
			Output: []string{`test("foo foe fum", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   "test(` foo bar bang  `, () => {});",
			Output: []string{"test(`foo bar bang`, () => {});"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code:   `describe.each()(" foo", () => {});`,
			Output: []string{`describe.each()("foo", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},
		{
			Code: `import { test as testThat } from "@rstest/core";
testThat("foo works ", () => {});`,
			Output: []string{`import { test as testThat } from "@rstest/core";
testThat("foo works", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},

		// duplicatePrefix.
		{
			Code:   `describe("describe foo", () => {});`,
			Output: []string{`describe("foo", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `test("test foo", () => {});`,
			Output: []string{`test("foo", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code: `import { test } from "@rstest/playwright";
test.describe("describe authentication", () => {});`,
			Output: []string{`import { test } from "@rstest/playwright";
test.describe("authentication", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code: `import { test } from "@rstest/playwright";
const suite = test.describe.only;
suite("describe authentication", () => {});`,
			Output: []string{`import { test } from "@rstest/playwright";
const suite = test.describe.only;
suite("authentication", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		// Detection uses the decoded title, but a fix is only safe when the raw
		// source spells both the prefix and its separating space literally.
		{
			Code:   `test("test\u0020auth flow", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `test("test\u0020auth", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `test("te\u0073t auth flow", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `test("test auth\u0020flow", () => {});`,
			Output: []string{`test("auth\u0020flow", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   "test(`TEST authentication`, () => {});",
			Output: []string{"test(`authentication`, () => {});"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `it("it foo", () => {});`,
			Output: []string{`it("foo", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   `it("it foos it correctly", () => {});`,
			Output: []string{`it("foos it correctly", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		{
			Code:   "describe('describe \tfoo', () => {})",
			Output: []string{"describe('\tfoo', () => {})"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		// D4/D5: the prefix is compared against the API being registered, not
		// against the local identifier. Renaming `it` to `xit` on import does
		// not turn `it works` into a valid title, and Rstest has no `xit` of
		// its own for the jest rule's f/x prefix trimming to apply to.
		{
			Code: `import { it as xit } from "@rstest/core";
xit("it works", () => {});`,
			Output: []string{`import { it as xit } from "@rstest/core";
xit("works", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		// D5: the Rstest parser follows same-file const aliases, so this is a
		// test registration with a duplicated prefix. The jest parser does not
		// follow the alias and reports nothing here.
		{
			Code: `const t = test.skip;
t("test foo", () => {});`,
			Output: []string{`const t = test.skip;
t("foo", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "duplicatePrefix"}},
		},
		// accidentalSpace does not return, so a title that both has a stray
		// space and repeats the keyword reports twice. A *leading* space would
		// instead hide duplicatePrefix, because the first word of " describe
		// foo" is the empty string.
		{
			Code:   `describe("describe foo ", () => {});`,
			Output: []string{`describe("describe foo", () => {});`, `describe("foo", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "accidentalSpace"},
				{MessageId: "duplicatePrefix"},
			},
		},
		{
			Code:   `describe(" describe foo", () => {});`,
			Output: []string{`describe("describe foo", () => {});`, `describe("foo", () => {});`},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "accidentalSpace"}},
		},

		// invalidEachSpecifier.
		{
			Code: `test.each([])("%z", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidEachSpecifier",
				Message:   `"%z" is not a valid format specifier`,
			}},
		},
		{
			Code:   `test.each([])("%s+%z", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `test.each([])("%%%z", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		// D1: %p is a jest pretty-format placeholder. Rstest's formatRegExp has
		// no `p`, so the specifier survives into the reported title verbatim.
		{
			Code: `test.each([])("%p", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "invalidEachSpecifier",
				Message:   `"%p" is not a valid format specifier`,
			}},
		},
		{
			Code:   `test.each([[1, 2]])(".add(%i, %p)", (a, b) => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		// D2: .for formats its title the same way .each does.
		{
			Code:   `test.for([])("%z", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `describe.for([])("%z", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `test.for([])("%p", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `test.skip.each([])("%z", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code:   `describe.each()("%z", () => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},
		{
			Code: `const entries = [[1, 1, 2]];
test.each(entries)(".add(%i, %z)", (a, b, expected) => {});`,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		},

		// mustNotMatch / mustMatch, including the custom-message variants.
		{
			Code: `describe("things to test", () => {
  describe("unit tests #unit", () => {
    it("is true", () => {});
  });

  describe("e2e tests #e4e", () => {
    it("is another test #e2e #rstest4life", () => {});
  });
});`,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": patMustNotHashBad,
				"mustMatch":    patMustMatchHash,
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustNotMatch"},
				{MessageId: "mustNotMatch"},
			},
		},
		{
			Code: `describe("things to test", () => {
  describe("e2e tests #e4e", () => {
    it("is another test #e2e #rstest4life", () => {});
  });
});`,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": []interface{}{patMustNotHashBad, `Please include "#unit" or "#e2e" in titles`},
				"mustMatch":    []interface{}{patMustMatchHash, `Please include "#unit" or "#e2e" in titles`},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "mustNotMatchCustom", Message: `Please include "#unit" or "#e2e" in titles`},
				{MessageId: "mustNotMatchCustom", Message: `Please include "#unit" or "#e2e" in titles`},
			},
		},
		{
			Code: `describe("things to test", () => {
  describe("e2e tests #e4e", () => {
    it("is another test #e2e #rstest4life", () => {});
  });
});`,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": map[string]interface{}{"describe": patMustNotHashBad},
				"mustMatch":    map[string]interface{}{"it": patMustMatchHash},
			}},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "mustNotMatch"}},
		},
		{
			Code:    `test("the correct way to properly handle all things", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": patHashTag}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mustMatch",
				Message:   `test should match /#(?:unit|integration|e2e)/u`,
			}},
		},
		{
			Code:    `describe("the test", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": map[string]interface{}{"describe": patHashTag}}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mustMatch",
				Message:   `describe should match /#(?:unit|integration|e2e)/u`,
			}},
		},
		{
			Code:    `describe.skip("the test", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": map[string]interface{}{"describe": patHashTag}}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "mustMatch"}},
		},
		{
			Code: `import { test } from "@rstest/playwright";
test.describe("suite", () => {});`,
			Options: []interface{}{map[string]interface{}{
				"mustNotMatch": map[string]interface{}{"describe": "^suite$"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mustNotMatch",
				Message:   "describe should not match /^suite$/u",
			}},
		},
		{
			Code: `import { test } from "@rstest/playwright";
test.describe("suite", () => {});`,
			Options: []interface{}{map[string]interface{}{
				"mustMatch": map[string]interface{}{"describe": "^ok$"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mustMatch",
				Message:   "describe should match /^ok$/u",
			}},
		},
		// D5: the groups are keyed by the API being registered, so a renamed
		// import is still matched against the `test` group.
		{
			Code: `import { test as check } from "@rstest/core";
check("wrong", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": map[string]interface{}{"test": "^ok"}}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mustMatch",
				Message:   `test should match /^ok/u`,
			}},
		},
		// `it` and `test` are separate groups, as they are upstream.
		{
			Code:    `it("wrong", () => {});`,
			Options: []interface{}{map[string]interface{}{"mustMatch": map[string]interface{}{"it": "^ok"}}},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "mustMatch",
				Message:   `it should match /^ok/u`,
			}},
		},
	}
	for _, letter := range []string{"P", "S", "D", "I", "F", "J", "Z"} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:   fmt.Sprintf(`test.each([])("%%%s", () => {})`, letter),
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "invalidEachSpecifier"}},
		})
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_title.ValidTitleRule,
		valid,
		invalid,
	)
}

// TestValidTitleMatcherSchema locks in the shape of the `mustMatch` /
// `mustNotMatch` options, which is copied byte for byte from the jest rule.
func TestValidTitleMatcherSchema(t *testing.T) {
	valid := []any{
		"^\\d+", // one pattern for every block kind
		[]any{"^\\d+", "titles must start with a number"},     // pattern with a custom message
		map[string]any{"describe": "^\\d+", "it": []any{"x"}}, // per-block-kind patterns
	}
	for _, matcher := range valid {
		for _, key := range []string{"mustMatch", "mustNotMatch"} {
			options := []any{map[string]any{key: matcher}}
			if err := valid_title.ValidTitleRule.Schema.Validate(options); err != nil {
				t.Errorf("expected %s: %v to pass schema validation, got: %v", key, matcher, err)
			}
		}
	}

	invalid := []any{
		42,                             // neither a pattern nor a per-kind map
		[]any{},                        // a pattern is required
		[]any{"a", "b", "c"},           // at most a pattern and a message
		map[string]any{"describe": 42}, // a per-kind value is still a pattern
	}
	for _, matcher := range invalid {
		options := []any{map[string]any{"mustMatch": matcher}}
		if err := valid_title.ValidTitleRule.Schema.Validate(options); err == nil {
			t.Errorf("expected mustMatch: %v to fail schema validation", matcher)
		}
	}

	options := []any{map[string]any{"mustAlsoMatch": "^\\d+"}}
	if err := valid_title.ValidTitleRule.Schema.Validate(options); err == nil {
		t.Error("expected an unknown option name to fail schema validation")
	}
}
