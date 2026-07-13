package no_thenable_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_thenable"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoThenableExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / upstream issue it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
func TestNoThenableExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_thenable.NoThenableRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: unresolved computed identifiers are dynamic keys ----
			tsValid(`const foo = {[(then)]: 1}`),
			tsValid(`class Foo {[(then)]() {}}`),
			tsValid(`foo[(then)] = 1`),

			// ---- Dimension 4: optional Object/Reflect method calls do not match upstream optionalCall/member false ----
			tsValid(`Object?.defineProperty(foo, "then", 1)`),
			tsValid(`Object.defineProperty?.(foo, "then", 1)`),
			tsValid(`(Object?.defineProperty)(foo, "then", 1)`),
			tsValid(`Reflect?.defineProperty(foo, "then", 1)`),
			tsValid(`Object?.fromEntries([["then", 1]])`),
			tsValid(`Object.fromEntries?.([["then", 1]])`),
			tsValid(`(Object?.fromEntries)([["then", 1]])`),

			// ---- Dimension 4: upstream isMethodCall uses computed:false for Object helpers ----
			tsValid(`Object["defineProperty"](foo, "then", 1)`),
			tsValid(`Reflect["defineProperty"](foo, "then", 1)`),
			tsValid(`Object["fromEntries"]([["then", 1]])`),

			// ---- Dimension 4: dynamic and non-string key forms do not match ----
			tsValid(`const foo = {0: 1, [Symbol.iterator]: 2}`),
			tsValid(`class Foo {#then; #thenMethod() {}}`),
			tsValid(`foo[Symbol.iterator] = 1`),
			tsValid(`Object.fromEntries([[Symbol.iterator, 1]])`),
			tsValid(`const prefix = "th"; const foo = {[prefix + suffix]: 1}`),
			tsValid(`const key = flag ? "then" : "no"; const foo = {[key]: 1}`),
			tsValid(`{ const String = value => "then"; Object.defineProperty(foo, String("x"), {value: 1}) }`),
			tsValid("{ const String = { raw: () => \"then\" }; const foo = {[String.raw`then`]: 1} }"),
			tsValid("let RawString = String; RawString = {raw: () => \"then\"} as any; const foo = {[RawString.raw`then`]: 1}"),
			tsValid(`let key = "then"; key = "other"; const foo = {[key]: 1}`),

			// ---- Dimension 4: non-assignment member access is allowed ----
			tsValid(`const value = foo.then`),
			tsValid(`foo.then++`),
			tsValid(`delete foo.then`),

			// ---- Dimension 4: graceful degradation for empty and spread-only containers ----
			tsValid(`const foo = {...bar}`),
			tsValid(`class Foo {}`),
			tsValid(`function foo({}) {}`),

			// ---- Upstream final implementation only checks array-entry pairs in fromEntries ----
			tsValid(`Object.fromEntries(["then", 1])`),

			// N/A: autofix boundaries do not apply; this rule has no fix.
			// N/A: call/new declaration container variants do not apply beyond object, class, assignment, call, and export nodes.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: direct string-literal object key ----
			tsObjectInvalid(`const foo = {"then": 1}`, `"then"`),

			// ---- Dimension 4: shorthand object property is still an object `then` key ----
			tsObjectInvalid(`const then = 1; const foo = {then}`, `then`, 2),

			// ---- Dimension 4: parenthesized and TS-wrapped computed keys resolve through const literals ----
			tsObjectInvalid(`const key = "then"; const foo = {[(key)]: 1}`, `key`, 2),
			tsObjectInvalid(`const key = "then"; const foo = {[key as string]: 1}`, `key`, 2),
			tsClassInvalid(`const key = "then"; class Foo {[(key)]() {}}`, `key`, 2),
			tsClassInvalid(`const key = "then"; class Foo {[key satisfies string] = 1}`, `key`, 2),
			tsObjectInvalid(`const key = "then"; foo[(key)] = 1`, `key`, 2),
			tsObjectInvalid(`const key = "then"; foo[key as string] = 1`, `key`, 2),

			// ---- Dimension 4: computed keys use static string evaluation through nested expressions ----
			tsObjectInvalid(`const prefix = "th"; const foo = {[prefix + "en"]: 1}`, `prefix + "en"`),
			tsClassInvalid("const prefix = \"th\"; const key = `${prefix}en`; class Foo {[key]() {}}", `key`, 2),
			tsObjectInvalid(`const key = "then"; foo.bar[key] = 1`, `key`, 2),
			tsObjectInvalid(`const key = ("then" satisfies string); foo[key] = 1`, `key`, 2),
			tsObjectInvalid(`const prefix = "th"; Object.defineProperty(foo, prefix + "en", {value: 1})`, `prefix + "en"`),
			tsObjectInvalid(`const prefix = "th"; Object.fromEntries([[prefix + "en", 1]])`, `prefix + "en"`),
			tsObjectInvalid(`const prefix = "th"; Object.fromEntries([([prefix + "en", 1])])`, `prefix + "en"`),
			tsObjectInvalid(`const key = true ? "then" : "no"; const foo = {[key]: 1}`, `key`, 2),
			tsClassInvalid(`const key = "" || "then"; class Foo {[key]() {}}`, `key`, 2),
			tsObjectInvalid(`const key = "then" && "then"; foo[key] = 1`, `key`, 2),
			tsObjectInvalid("const key = String.raw`then`; Object.defineProperty(foo, key, {value: 1})", `key`, 2),
			tsObjectInvalid("const RawString = String; const foo = {[RawString.raw`then`]: 1}", "RawString.raw`then`"),
			tsObjectInvalid(`const key = String("then"); Object.fromEntries([[key, 1]])`, `key`, 2),
			tsObjectInvalid(`let key = "then"; const foo = {[key]: 1}`, `key`, 2),
			tsObjectInvalid(`var key = "then"; Object.defineProperty(foo, key, {value: 1})`, `key`, 2),

			// ---- Dimension 4: spread properties do not mask sibling checks ----
			tsObjectInvalid(`const foo = {...bar, then: 1}`, `then`),

			// ---- Dimension 4: TS body-absent class member forms are still thenable class members ----
			tsClassInvalid(`declare class Foo { then(): void }`, `then`),
			tsClassInvalid(`declare class Foo { then: string }`, `then`),
			tsClassInvalid(`class Foo { accessor then = 1 }`, `then`),

			// ---- Dimension 4: nested containers report independently without traversal bleed ----
			{
				Code:     `const foo = {then: 1, nested: {then: 2}}`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError(`const foo = {then: 1, nested: {then: 2}}`, `then`, messageIDObject, messageObject),
					expectedError(`const foo = {then: 1, nested: {then: 2}}`, `then`, messageIDObject, messageObject, 2),
				},
			},
			{
				Code:     `const prefix = "th"; const foo = {[prefix + "en"]: {nested: {[prefix + "en"]: 1}}}`,
				FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					expectedError(`const prefix = "th"; const foo = {[prefix + "en"]: {nested: {[prefix + "en"]: 1}}}`, `prefix + "en"`, messageIDObject, messageObject),
					expectedError(`const prefix = "th"; const foo = {[prefix + "en"]: {nested: {[prefix + "en"]: 1}}}`, `prefix + "en"`, messageIDObject, messageObject, 2),
				},
			},

			// ---- Real-user: #1710 yup conditional validation object reports the `then` branch ----
			tsObjectInvalid(`
const validationSchema = Yup.object().shape(
	{
		addEntry: Yup.string()
			.nullable()
			.notRequired()
			.when('addEntry', {
				is: () => entryType === 'EMAIL',
				then: rule => rule.email('Please enter a valid email.'),
				otherwise: rule => rule.matches(anyDomainRegex, 'Please enter a correct domain format.'),
			}),
	},
	[['addEntry', 'addEntry']],
);`, `then`),

			// ---- Real-user: #1841 yup object schema conditional validation reports the `then` branch ----
			tsObjectInvalid(`
const mySchema = yup.object({
	requireUsername: yup.bool(),
	username: yup.string().when('requireUsername', {
		is: true,
		then: yup.string().required('Username is required'),
	}),
});`, `then`),

			// Locks in upstream ObjectExpression arm: method/getter/setter/property all use isThenKey().
			tsObjectInvalid(`const foo = {get then() { return 1 }}`, `then`),

			// Locks in upstream class-member arm: class fields and methods share the class message.
			tsClassInvalid(`class Foo {static then = 1}`, `then`),

			// Locks in upstream MemberExpression arm: only assignment left-hand sides report.
			tsObjectInvalid(`foo.then = foo.then`, `then`),

			// Locks in member-chain assignments from the original rule proposal.
			tsObjectInvalid(`Foo.prototype.then = () => {}`, `then`),

			// Locks in upstream defineProperty arm: Object and Reflect calls both report their second argument.
			tsObjectInvalid(`Reflect.defineProperty(foo, "then", {value: 1})`, `"then"`),

			// Locks in upstream fromEntries arm: non-pair elements are skipped, later pairs still report.
			tsObjectInvalid(`Object.fromEntries([foo, , ["then", 1]])`, `"then"`),

			// Locks in upstream export-specifier arm: exported name, not local name, is checked.
			tsExportInvalid(`const local = 1; export {local as then}`, `then`),

			// Locks in upstream exported declaration arm: named export reports, default export does not.
			tsExportInvalid(`export async function * then() {}`, `then`),

			// Locks in upstream exported variable arm: nested binding identifiers are collected.
			tsExportInvalid(`export const {foo: [{bar: then}]} = value`, `then`),
		},
	)
}

func tsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.ts"}
}

func tsObjectInvalid(code string, target string, occurrence ...int) rule_tester.InvalidTestCase {
	return tsInvalid(code, target, messageIDObject, messageObject, occurrence...)
}

func tsClassInvalid(code string, target string, occurrence ...int) rule_tester.InvalidTestCase {
	return tsInvalid(code, target, messageIDClass, messageClass, occurrence...)
}

func tsExportInvalid(code string, target string, occurrence ...int) rule_tester.InvalidTestCase {
	return tsInvalid(code, target, messageIDExport, messageExport, occurrence...)
}

func tsInvalid(code string, target string, messageID string, message string, occurrence ...int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.ts",
		Errors: []rule_tester.InvalidTestCaseError{
			expectedError(code, target, messageID, message, occurrence...),
		},
	}
}
