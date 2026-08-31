package no_empty

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoEmptyUpstream migrates the complete valid/invalid suite from ESLint
// v10.9.1 tests/lib/rules/no-empty.js. Every invalid case additionally locks
// in the full diagnostic range and suggestion output.
func TestNoEmptyUpstream(t *testing.T) {
	allowEmptyCatch := []any{map[string]any{"allowEmptyCatch": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyRule,
		[]rule_tester.ValidTestCase{
			{Code: `if (foo) { bar() }`},
			{Code: `while (foo) { bar() }`},
			{Code: `for (;foo;) { bar() }`},
			{Code: `try { foo() } catch (ex) { foo() }`},
			{Code: `switch(foo) {case 'foo': break;}`},
			{Code: `(function() { }())`},
			{Code: `var foo = () => {};`},
			{Code: `function foo() { }`},
			{Code: `if (foo) {/* empty */}`},
			{Code: `while (foo) {/* empty */}`},
			{Code: `switch (foo) {/* empty */}`},
			{Code: `for (;foo;) {/* empty */}`},
			{Code: `try { foo() } catch (ex) {/* empty */}`},
			{Code: "try { foo() } catch (ex) {// empty\n}"},
			{Code: "try { foo() } finally {// empty\n}"},
			{Code: "try { foo() } finally {// test\n}"},
			{Code: "try { foo() } finally {\n \n // hi i am off no use\n}"},
			{Code: `try { foo() } catch (ex) {/* test111 */}`},
			{Code: "if (foo) { bar() } else { // nothing in me \n}"},
			{Code: "if (foo) { bar() } else { /**/ \n}"},
			{Code: "if (foo) { bar() } else { // \n}"},
			{Code: `try { foo(); } catch (ex) {}`, Options: allowEmptyCatch},
			{Code: `try { foo(); } catch (ex) {} finally { bar(); }`, Options: allowEmptyCatch},
		},
		[]rule_tester.InvalidTestCase{
			invalidNoEmptyCase(
				`try {} catch (ex) {throw ex}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `try { /* empty */ } catch (ex) {throw ex}`},
			),
			invalidNoEmptyCase(
				`try { foo() } catch (ex) {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `try { foo() } catch (ex) { /* empty */ }`},
			),
			invalidNoEmptyCase(
				`try { foo() } catch (ex) {throw ex} finally {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `try { foo() } catch (ex) {throw ex} finally { /* empty */ }`},
			),
			invalidNoEmptyCase(
				`if (foo) {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `if (foo) { /* empty */ }`},
			),
			invalidNoEmptyCase(
				`while (foo) {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `while (foo) { /* empty */ }`},
			),
			invalidNoEmptyCase(
				`for (;foo;) {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `for (;foo;) { /* empty */ }`},
			),
			invalidNoEmptyCase(
				`switch(foo) {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "switch", output: `switch(foo) { /* empty */ }`},
			),
			invalidNoEmptyCase(
				`switch /* empty */ (/* empty */ foo /* empty */) /* empty */ {} /* empty */`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "switch", output: `switch /* empty */ (/* empty */ foo /* empty */) /* empty */ { /* empty */ } /* empty */`},
			),
			invalidNoEmptyCase(
				`try {} catch (ex) {}`,
				allowEmptyCatch,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `try { /* empty */ } catch (ex) {}`},
			),
			invalidNoEmptyCase(
				`try { foo(); } catch (ex) {} finally {}`,
				allowEmptyCatch,
				noEmptyErrorSpec{occurrence: 1, statementType: "block", output: `try { foo(); } catch (ex) {} finally { /* empty */ }`},
			),
			invalidNoEmptyCase(
				`try {} catch (ex) {} finally {}`,
				allowEmptyCatch,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `try { /* empty */ } catch (ex) {} finally {}`},
				noEmptyErrorSpec{occurrence: 2, statementType: "block", output: `try {} catch (ex) {} finally { /* empty */ }`},
			),
			invalidNoEmptyCase(
				`try { foo(); } catch (ex) {} finally {}`,
				nil,
				noEmptyErrorSpec{occurrence: 0, statementType: "block", output: `try { foo(); } catch (ex) { /* empty */ } finally {}`},
				noEmptyErrorSpec{occurrence: 1, statementType: "block", output: `try { foo(); } catch (ex) {} finally { /* empty */ }`},
			),
		},
	)
}

type noEmptyErrorSpec struct {
	occurrence    int
	statementType string
	output        string
}

func invalidNoEmptyCase(code string, options any, specs ...noEmptyErrorSpec) rule_tester.InvalidTestCase {
	errors := make([]rule_tester.InvalidTestCaseError, 0, len(specs))
	for _, spec := range specs {
		start := nthEmptyBraces(code, spec.occurrence)
		if start < 0 {
			panic("no-empty upstream case does not contain enough empty blocks: " + code)
		}
		line, column := noEmptyLineColumn(code, start)
		endLine, endColumn := noEmptyLineColumn(code, start+2)
		errors = append(errors, rule_tester.InvalidTestCaseError{
			MessageId: "unexpected",
			Message:   "Empty " + spec.statementType + " statement.",
			Line:      line,
			Column:    column,
			EndLine:   endLine,
			EndColumn: endColumn,
			Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
				MessageId: "suggestComment",
				Output:    spec.output,
			}},
		})
	}
	return rule_tester.InvalidTestCase{Code: code, Options: options, Errors: errors}
}

func nthEmptyBraces(code string, occurrence int) int {
	start := 0
	for current := 0; current <= occurrence; current++ {
		offset := strings.Index(code[start:], "{}")
		if offset < 0 {
			return -1
		}
		if current == occurrence {
			return start + offset
		}
		start += offset + 2
	}
	return -1
}

func noEmptyLineColumn(code string, offset int) (line, column int) {
	line, column = 1, 1
	for index, char := range code {
		if index >= offset {
			break
		}
		if char == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}
