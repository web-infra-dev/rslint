// TestNoAwaitExpressionMemberUpstream migrates every valid/invalid case from
// Unicorn v74.0.0 test/no-await-expression-member.js and its documentation.
// Every diagnostic has position and message assertions. Additional edge shapes
// and branch lock-ins live in no_await_expression_member_extras_test.go.
package no_await_expression_member_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_await_expression_member"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAwaitExpressionMemberUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&no_await_expression_member.NoAwaitExpressionMemberRule,
		[]rule_tester.ValidTestCase{
			{Code: `const foo = await promise`},
			{Code: `const {foo: bar} = await promise`},
			{Code: `const foo = !await promise`},
			{Code: `const foo = typeof await promise`},
			{Code: `const foo = await notPromise.method()`},
			{Code: `const foo = foo[await promise]`},
			// These await expressions need parentheses, but are rarely used.
			{Code: `new (await promiseReturnsAClass)`},
			{Code: `(await promiseReturnsAFunction)()`},
			// ---- TypeScript ----
			{Code: `function foo () {return (await promise) as string;}`},
			{Code: `(await promise)!.property`},
			// ---- Documentation ----
			{Code: `const {default: foo} = await import('./foo.js');`},
			{Code: `const [, secondElement] = await getArray();`},
			{Code: `const {property} = await getObject();`},
			{Code: `const response = await fetch('/foo'); const data = await response.json();`},
		},
		[]rule_tester.InvalidTestCase{
			memberInvalid(`(await promise)[0]`, "0"),
			memberInvalid(`(await promise).property`, "property"),
			memberInvalid(`const foo = (await promise).bar()`, "bar"),
			memberInvalid(`const foo = (await promise).bar?.()`, "bar"),
			memberInvalid(`const foo = (await promise)?.bar()`, "bar"),

			memberInvalid(`const firstElement = (await getArray())[0]`, "0", `const [firstElement] = await getArray()`),
			memberInvalid(`const secondElement = (await getArray())[1]`, "1", `const [, secondElement] = await getArray()`),
			memberInvalid(`const thirdElement = (await getArray())[2]`, "2"),
			memberInvalid(`const optionalFirstElement = (await getArray())?.[0]`, "0"),
			memberInvalid(`const {propertyOfFirstElement} = (await getArray())[0]`, "0"),
			memberInvalid(`const [firstElementOfFirstElement] = (await getArray())[0]`, "0"),
			memberInvalid(`let foo, firstElement = (await getArray())[0]`, "0", `let foo, [firstElement] = await getArray()`),
			memberInvalid(`var firstElement = (await getArray())[0], bar`, "0", `var [firstElement] = await getArray(), bar`),

			memberInvalid(`const property = (await getObject()).property`, "property", `const {property} = await getObject()`),
			memberInvalid(`let property = (await getObject()).property`, "property", `let {property} = await getObject()`),
			memberInvalid(`const renamed = (await getObject()).property`, "property"),
			// Only the member directly off await is flagged, not the chained .bar.
			memberInvalid(`(await promise).foo.bar`, "foo"),
			memberInvalid(`const property = (await getObject())[property]`, "property"),
			memberInvalid(`const property = (await getObject())?.property`, "property"),
			memberInvalid(`const {propertyOfProperty} = (await getObject()).property`, "property"),
			memberInvalid(`const {propertyOfProperty} = (await getObject()).propertyOfProperty`, "propertyOfProperty"),
			memberInvalid(`const [firstElementOfProperty] = (await getObject()).property`, "property"),
			memberInvalid(`const [firstElementOfProperty] = (await getObject()).firstElementOfProperty`, "firstElementOfProperty"),

			memberInvalid(`firstElement = (await getArray())[0]`, "0"),
			memberInvalid(`property = (await getArray()).property`, "property"),
			// ---- TypeScript ----
			memberInvalid(`const foo: Type = (await promise)[0]`, "0"),
			memberInvalid(`const foo: Type | A = (await promise).foo`, "foo"),
			// ---- Documentation (the other two examples are covered above) ----
			memberInvalid(`const foo = (await import('./foo.js')).default;`, "default"),
			memberInvalid(`const data = await (await fetch('/foo')).json();`, "json"),
		},
	)
}

func memberInvalid(code, property string, output ...string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:     code,
		FileName: "file.mts",
		Output:   output,
		Errors:   []rule_tester.InvalidTestCaseError{memberError(code, property)},
	}
}

func memberError(code, property string) rule_tester.InvalidTestCaseError {
	start := strings.LastIndex(code, property)
	if property == "" || start < 0 {
		panic("missing property in test input: " + property)
	}
	end := start + len(property)
	return rule_tester.InvalidTestCaseError{
		MessageId: "no-await-expression-member",
		Message:   "Do not access a member directly from an await expression.",
		Line:      strings.Count(code[:start], "\n") + 1,
		Column:    start - strings.LastIndex(code[:start], "\n"),
		EndLine:   strings.Count(code[:end], "\n") + 1,
		EndColumn: end - strings.LastIndex(code[:end], "\n"),
	}
}
