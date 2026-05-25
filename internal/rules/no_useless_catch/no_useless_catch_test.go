package no_useless_catch

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessCatchRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessCatchRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `try { foo(); } catch (err) { console.error(err); }`},
			{Code: `try { foo(); } catch (err) { console.error(err); } finally { bar(); }`},
			{Code: `try { foo(); } catch (err) { doSomethingBeforeRethrow(); throw err; }`},
			{Code: `try { foo(); } catch (err) { throw err.msg; }`},
			{Code: `try { foo(); } catch (err) { throw new Error("whoops!"); }`},
			{Code: `try { foo(); } catch (err) { throw bar; }`},
			{Code: `try { foo(); } catch (err) { }`},
			{Code: `try { foo(); } catch ({ err }) { throw err; }`},
			{Code: `try { foo(); } catch ([ err ]) { throw err; }`},
			// Optional catch binding (ES2019): no param means rule can't match.
			{Code: `try { throw new Error('foo'); } catch { throw new Error('foo'); }`},
			// Non-null assertion wraps the identifier, so it's not a plain rethrow.
			{Code: `try { foo(); } catch (err) { throw err!; }`},
			// `as` expression wraps the identifier.
			{Code: `try { foo(); } catch (err) { throw err as Error; }`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `try { foo(); } catch (err) { throw err; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCatch", Line: 1, Column: 1},
				},
			},
			{
				Code: `try { foo(); } catch (err) { throw err; } finally { foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCatchClause", Line: 1, Column: 16},
				},
			},
			{
				Code: `try { foo(); } catch (err) { /* comment */ throw err; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCatch", Line: 1, Column: 1},
				},
			},
			// Type annotation on the catch param: still an Identifier name.
			{
				Code: `try { foo(); } catch (err: unknown) { throw err; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCatch", Line: 1, Column: 1},
				},
			},
			// Parentheses around the thrown identifier are flattened in ESTree.
			{
				Code: `try { foo(); } catch (err) { throw (err); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCatch", Line: 1, Column: 1},
				},
			},
			// ESLint fires when the first statement is the rethrow, even if
			// later statements exist (they are unreachable).
			{
				Code: `try { foo(); } catch (err) { throw err; unreachable(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryCatch", Line: 1, Column: 1},
				},
			},
		},
	)
}
