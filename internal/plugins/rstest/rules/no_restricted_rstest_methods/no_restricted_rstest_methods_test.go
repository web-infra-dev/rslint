package no_restricted_rstest_methods

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func disallowed(member string, line, column, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "restrictedRstestMethod",
		Message:   "Use of `" + member + "` is disallowed",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}}
}

func withMessage(message string, line, column, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "restrictedRstestMethodWithMessage",
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}}
}

func TestNoRestrictedRstestMethods(t *testing.T) {
	noFn := []any{map[string]any{"fn": nil}}
	fnWithMessage := []any{map[string]any{"fn": "Use the shared factory instead."}}
	noMock := []any{map[string]any{"mock": nil}}
	noImportActual := []any{map[string]any{"importActual": nil}}
	noImportMock := []any{map[string]any{"importMock": nil}}
	noUnknownMember := []any{map[string]any{"notAnRstestMember": nil}}
	empty := []any{map[string]any{}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t, &NoRestrictedRstestMethodsRule,
		[]rule_tester.ValidTestCase{
			// Nothing is disallowed until the option object says so.
			{Code: `rs.fn();`},
			{Code: `rs.fn();`, Options: empty},
			{Code: `rs.spyOn(target, 'method');`, Options: noFn},
			// A member of something else is not a member of the utilities
			// object, whichever kind of member it is named after.
			{Code: `helpers.fn();`, Options: noFn},
			{Code: `helpers.mock('./m');`, Options: noMock},
			{Code: `fn();`, Options: noFn},
			{Code: `helpers.notAnRstestMember();`, Options: noUnknownMember},
			// A whole-module require binds the module namespace, not the
			// utilities object, which is the `rs` on it.
			{Code: `const rs = require('@rstest/core'); rs.fn();`, Options: noFn},
			// A type-only import is erased before runtime, so a namespace it
			// binds reaches nothing: `core` stands for no value at all.
			{
				Code:    `import type core = require('@rstest/core'); core.rs.fn();`,
				Options: noFn,
			},
			{
				Code:    `import type * as core from '@rstest/core'; core.rs.fn();`,
				Options: noFn,
			},
			// Erased under a further name, the call reaches neither the import
			// nor a global: nothing named `mocker` is Rstest's.
			{
				Code:    `import type { rs as mocker } from '@rstest/core'; mocker.fn();`,
				Options: noFn,
			},

			// ---- The ordinary members follow the binding ----
			// A receiver the file declares itself is a different object.
			{Code: `const rs = { fn: () => {} }; rs.fn();`, Options: noFn},
			{Code: `function make(rs) { return rs.fn(); }`, Options: noFn},
			{Code: `import { rs } from './helpers'; rs.fn();`, Options: noFn},

			// ---- The plugin-managed members are read as written ----
			// A renamed binding is never rewritten: the call throws where it
			// stands rather than mocking anything.
			{Code: `import { rs as mocker } from '@rstest/core'; mocker.mock('./m');`, Options: noMock},
			{Code: `import * as core from '@rstest/core'; core.rs.mock('./m');`, Options: noMock},
			// Shapes the build does not rewrite reach the stub that throws.
			{Code: `rs['mock']('./m');`, Options: noMock},
			{Code: `rs.mock?.('./m');`, Options: noMock},
			{Code: `rs?.mock('./m');`, Options: noMock},
			{Code: "rs[`importActual`]('./m');", Options: noImportActual},
			{Code: `rs?.importActual('./m');`, Options: noImportActual},
			{Code: `rs['importMock']('./m');`, Options: noImportMock},
			// These two are handed back to a local declaration of the
			// receiver, so the call runs that object's method.
			{Code: `const rs = { importActual: async (p) => ({}) }; rs.importActual('./m');`, Options: noImportActual},
			{Code: `function load(rs) { return rs.requireActual('./m'); }`, Options: []any{map[string]any{"requireActual": nil}}},
			// `import.meta.rstest` carries the module namespace, and the
			// plugin-managed members are not rewritten through it.
			{Code: `import.meta.rstest.rs.mock('./m');`, Options: noMock},
		},
		[]rule_tester.InvalidTestCase{
			{Code: `rs.fn();`, Options: noFn, Errors: disallowed("fn", 1, 4, 6)},
			{Code: `rstest.fn();`, Options: noFn, Errors: disallowed("fn", 1, 8, 10)},
			{
				Code:    `rs.fn();`,
				Options: fnWithMessage,
				Errors:  withMessage("Use the shared factory instead.", 1, 4, 6),
			},

			// ---- The ordinary members are real functions in every shape ----
			{Code: `rs['fn']();`, Options: noFn, Errors: disallowed("fn", 1, 4, 8)},
			{Code: `rs.fn?.();`, Options: noFn, Errors: disallowed("fn", 1, 4, 6)},
			{Code: `rs?.fn();`, Options: noFn, Errors: disallowed("fn", 1, 5, 7)},
			// However the binding was made, it names the same function.
			{
				Code:    `import { rs as mocker } from '@rstest/core'; mocker.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 53, 55),
			},
			{
				Code:    `import { rstest as testing } from 'rstack/test'; testing.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 58, 60),
			},
			{
				Code:    `import * as core from '@rstest/core'; core.rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 47, 49),
			},
			{
				Code:    `import.meta.rstest.rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 23, 25),
			},
			// A require reaches the same bindings as an import: the utilities
			// object destructured off the module, under either name, and the
			// module namespace it is a member of.
			{
				Code:    `const { rs } = require('@rstest/core'); rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 44, 46),
			},
			{
				Code:    `const { rs: mocker } = require('@rstest/core'); mocker.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 56, 58),
			},
			{
				Code:    `const core = require('@rstest/core'); core.rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 47, 49),
			},
			// An assertion on the require call is erased before runtime, so
			// the binding is still the module namespace.
			{
				Code:    `const core = require('@rstest/core') as any; core.rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 54, 56),
			},
			// `import x = require(...)` binds the module namespace the same
			// way a namespace import does.
			{
				Code:    `import core = require('@rstest/core'); core.rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 48, 50),
			},
			// A type-only import of the utilities object binds no value, so
			// the call is the global `rs` under `globals: true` rather than
			// the erased binding — the same reading the registrations get.
			{
				Code:    `import type { rs } from '@rstest/core'; rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 44, 46),
			},
			{
				Code:    `import { type rs } from '@rstest/core'; rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 44, 46),
			},
			{
				Code:    `import type rs = require('@rstest/core'); rs.importActual('./m');`,
				Options: noImportActual,
				Errors:  disallowed("importActual", 1, 46, 58),
			},
			// An assertion on the specifier is erased too.
			{
				Code:    `const core = require('@rstest/core' as any); core.rs.fn();`,
				Options: noFn,
				Errors:  disallowed("fn", 1, 54, 56),
			},
			// The option object is taken as written: a name that is not a
			// member of the utilities object matches nothing real, but it is
			// still matched where it is written on the object.
			{
				Code:    `rs.notAnRstestMember();`,
				Options: noUnknownMember,
				Errors:  disallowed("notAnRstestMember", 1, 4, 21),
			},

			// ---- The plugin-managed members are read as written ----
			{Code: `rs.mock('./m');`, Options: noMock, Errors: disallowed("mock", 1, 4, 8)},
			{Code: `rstest.mock('./m');`, Options: noMock, Errors: disallowed("mock", 1, 8, 12)},
			// A local declaration of the receiver is bypassed rather than
			// honored, so the call is still Rstest's.
			{
				Code:    `const rs = { mock() {} }; rs.mock('./m');`,
				Options: noMock,
				Errors:  disallowed("mock", 1, 30, 34),
			},
			{
				Code:    `const rs = { importMock: async (p) => ({}) }; rs.importMock('./m');`,
				Options: noImportMock,
				Errors:  disallowed("importMock", 1, 50, 60),
			},
			// The two members whose rewrite reads a wider set of shapes.
			{Code: `rs['importActual']('./m');`, Options: noImportActual, Errors: disallowed("importActual", 1, 4, 18)},
			{Code: `rs.importActual?.('./m');`, Options: noImportActual, Errors: disallowed("importActual", 1, 4, 16)},
			// Parentheses and TypeScript's type-only syntax are transparent on
			// both sides of the callee.
			{Code: `(rs as any).mock('./m');`, Options: noMock, Errors: disallowed("mock", 1, 13, 17)},
			{Code: `(rs.fn)();`, Options: noFn, Errors: disallowed("fn", 1, 5, 7)},
		},
	)
}
