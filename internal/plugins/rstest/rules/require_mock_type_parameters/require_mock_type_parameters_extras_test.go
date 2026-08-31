// TestRequireMockTypeParametersExtras covers the shapes the migrated suite does
// not: the tsgo/ESTree divergences on every child node the rule reads, the two
// Rstest module loaders with no counterpart upstream, and one locked-in case
// per reachable branch.
package require_mock_type_parameters

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func missing(member string, line, column, endColumn int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "missingTypeParameter",
		Message:   "'" + member + "' is called without a type parameter.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: endColumn,
	}}
}

func TestRequireMockTypeParametersExtras(t *testing.T) {
	checkImports := []any{map[string]any{"checkImportFunctions": true}}
	emptyOptions := []any{map[string]any{}}
	dontCheckImports := []any{map[string]any{"checkImportFunctions": false}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireMockTypeParametersRule,
		[]rule_tester.ValidTestCase{
			// ---- Locks in the option's default ----
			// An empty options object and an explicit `false` both leave the
			// module loaders alone, exactly as passing no options does.
			{Code: `rs.importActual('./example.js')`, Options: emptyOptions},
			{Code: `rs.requireMock('./example.js')`, Options: dontCheckImports},
			{Code: `rstest.importMock('./example.js')`},

			// ---- Dimension 4: access / key forms ----
			// `fn` is reached by resolving the receiver and reading a member
			// name, so a computed key names no member to report.
			{Code: `rs['fn']()`},
			{Code: "rs[`fn`]()"},
			{Code: `rs[0]()`},
			// The loader rewrite does read a computed key, but only a plain
			// string one: a substitution and a non-string key are unknown until
			// the code runs, so the call stays the throwing stub.
			{Code: "rs[`import${'Actual'}`]('./example.js')", Options: checkImports},
			{Code: `rs[loaderName]('./example.js')`, Options: checkImports},
			{Code: `rs?.['importActual']('./example.js')`, Options: checkImports},

			// ---- Dimension 4: nesting / traversal boundaries ----
			// `fn` has to be the member the call goes through. An uncalled
			// `rs.fn` reached as a value carries no call to write a type
			// parameter on.
			{Code: `rs.fn.mockReturnValue(1)`},
			{Code: `const factory = rs.fn;`},
			{Code: `new rs.fn()`},

			// Other members of the utilities object are untouched.
			{Code: `rs.spyOn(console, 'log')`},
			{Code: `rs.mocked(load)`},
			{Code: `rs.clearAllMocks()`},
			// `resetModules` and `hoisted` are plugin-managed too, but neither
			// hands back a module to type.
			{Code: `rs.resetModules()`, Options: checkImports},
			{Code: `rs.hoisted(() => ({}))`, Options: checkImports},

			// ---- Locks in the receiver resolution for `fn` ----
			// A local declaration of the name really is a different object,
			// because `fn` is an ordinary function reached through an ordinary
			// binding.
			{Code: `const rs = { fn: () => {} }; rs.fn();`},
			{Code: `function register(rstest) { rstest.fn(); }`},
			{Code: `import { rs } from './helpers'; rs.fn();`},
			// A member of something that is not the utilities object.
			{Code: `import * as core from '@rstest/core'; core.fn();`},
			{Code: `helpers.rs.fn();`},
			// `import.meta.rstest` exposes the registration APIs only; there is
			// no `fn` on it to type.
			{Code: `import.meta.rstest.fn();`},

			// ---- Accepted false negatives ----
			// A false negative is the side this rule errs on, so each of these
			// stays unreported until the shape it needs is understood
			// everywhere. Any change that starts reporting one belongs in the
			// invalid list, deliberately.
			//
			// A tagged template is a distinct node, not a `CallExpression`, so
			// the listener never sees it. Nothing is lost in practice: `fn`
			// takes the implementation to infer from, and a tag is handed a
			// `TemplateStringsArray` that no `FunctionLike` accepts.
			{Code: "import { rs } from '@rstest/core'; rs.fn`service`;"},
			// `import = require` binds the module namespace, but the shared
			// resolver in internal/utils/test_framework recognizes only a
			// namespace import and a `require` initializer as one.
			{Code: `import core = require('@rstest/core'); core.rs.fn();`},
			// The same resolver has no branch for a dynamic import, so a
			// destructure of one — renamed or not — is not resolved either.
			{Code: `async function setup() { const { rs } = await import('@rstest/core'); rs.fn(); }`},
			{Code: `async function setup() { const { rs: mocker } = await import('@rstest/core'); mocker.fn(); }`},

			// ---- Locks in the syntax match for the module loaders ----
			// Rstest rewrites these four on the name written at the call site,
			// so a renamed binding is never rewritten and throws where it
			// stands; a type parameter would not change that.
			{Code: `import { rs as mocker } from '@rstest/core'; mocker.importActual('./example.js');`, Options: checkImports},
			{Code: `import { rstest as vi } from '@rstest/core'; vi.importMock('./example.js');`, Options: checkImports},
			{Code: `import * as core from '@rstest/core'; core.rs.importActual('./example.js');`, Options: checkImports},
			// Unlike the hoisted module-mock APIs, a local binding of the name
			// keeps the rewrite from happening at all.
			{Code: `const rs = { importActual: async (p) => ({}) }; rs.importActual('./example.js');`, Options: checkImports},
			{Code: `function load(rs) { return rs.requireActual('./example.js'); }`, Options: checkImports},
			{Code: `const rs = { importActual: async (p) => ({}) }; rs['importActual']('./example.js');`, Options: checkImports},
			// An optional chain on the receiver is left as the runtime stub.
			{Code: `rs?.importActual('./example.js')`, Options: checkImports},

			// The build resolves the path, so anything it cannot read at build
			// time is either a build failure or the runtime stub.
			{Code: `const path = './example.js'; rs.importActual(path);`, Options: checkImports},
			{Code: "rs.importActual(`./example.js`)", Options: checkImports},
			{Code: `rs.importActual(...paths)`, Options: checkImports},
			{Code: `rs.importActual()`, Options: checkImports},
			{Code: `rs.importActual('./example.js', {})`, Options: checkImports},
			{Code: `rs.requireMock('./example.js', {})`, Options: checkImports},
			{Code: `rs['importActual'](path)`, Options: checkImports},

			// ---- Dimension 4: receiver / expression wrappers ----
			// The type parameter is what the rule asks for, through any wrapper.
			{Code: `(rs).fn<() => void>()`},
			{Code: `(rs as any).fn<() => void>()`},
			{Code: `rs!.fn<() => void>()`},
			{Code: `(rs.fn)<() => void>()`},
			{Code: `(rs as any).importActual<MyModule>('./example.js')`, Options: checkImports},
			{Code: `rs.requireActual<MyModule>('./example.js')`, Options: checkImports},
			{Code: `rs.requireMock<MyModule>('./example.js')`, Options: checkImports},

			// ---- Graceful degradation ----
			{Code: `rs.fn<() => void>(...factories)`},
			{Code: `rs.fn<() => void>();;`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: receiver / expression wrappers ----
			{Code: `(rs).fn()`, Errors: missing("fn", 1, 6, 8)},
			{Code: `((rs)).fn()`, Errors: missing("fn", 1, 8, 10)},
			{Code: `rs!.fn()`, Errors: missing("fn", 1, 5, 7)},
			{Code: `(rs as any).fn()`, Errors: missing("fn", 1, 13, 15)},
			{Code: `(rs satisfies object).fn()`, Errors: missing("fn", 1, 23, 25)},
			{Code: `(rs.fn)()`, Errors: missing("fn", 1, 5, 7)},
			{Code: `(rs.fn as any)()`, Errors: missing("fn", 1, 5, 7)},
			// An optional chain still reaches the same function, so the mock it
			// creates is still untyped.
			{Code: `rs?.fn()`, Errors: missing("fn", 1, 5, 7)},
			{Code: `rs.fn?.()`, Errors: missing("fn", 1, 4, 6)},
			{Code: `(rs as any).importActual('./example.js')`, Options: checkImports, Errors: missing("importActual", 1, 13, 25)},
			{Code: `rs!.requireActual('./example.js')`, Options: checkImports, Errors: missing("requireActual", 1, 5, 18)},
			{Code: `(rs.importMock)('./example.js')`, Options: checkImports, Errors: missing("importMock", 1, 5, 15)},
			{Code: `rs.importActual(('./example.js'))`, Options: checkImports, Errors: missing("importActual", 1, 4, 16)},
			{Code: `rs.importActual('./example.js' as string)`, Options: checkImports, Errors: missing("importActual", 1, 4, 16)},

			// ---- Dimension 4: the loader rewrite's wider call surface ----
			// A computed string key and an optional call are both rewritten,
			// measured on rstest 0.11.8, so both leave an untyped module.
			{Code: `rs['importActual']('./example.js')`, Options: checkImports, Errors: missing("importActual", 1, 4, 18)},
			{Code: "rs[`requireActual`]('./example.js')", Options: checkImports, Errors: missing("requireActual", 1, 4, 19)},
			{Code: `rs.importActual?.('./example.js')`, Options: checkImports, Errors: missing("importActual", 1, 4, 16)},
			{Code: `rs['importMock']?.('./example.js')`, Options: checkImports, Errors: missing("importMock", 1, 4, 16)},
			{Code: `(rs as any)['importActual']('./example.js')`, Options: checkImports, Errors: missing("importActual", 1, 13, 27)},
			// The shadowing rule is the same whichever way the member is
			// written, and so is the argument shape.
			{Code: `rs['requireMock']('./example.js')`, Options: checkImports, Errors: missing("requireMock", 1, 4, 17)},

			// ---- The two module loaders Rstest adds ----
			{
				Code:    `rs.requireActual('./example.js')`,
				Options: checkImports,
				Errors:  missing("requireActual", 1, 4, 17),
			},
			{
				Code:    `rs.requireMock('./example.js')`,
				Options: checkImports,
				Errors:  missing("requireMock", 1, 4, 15),
			},

			// ---- Locks in both namespace spellings ----
			{Code: `rstest.fn()`, Errors: missing("fn", 1, 8, 10)},
			{Code: `rstest.importActual('./example.js')`, Options: checkImports, Errors: missing("importActual", 1, 8, 20)},

			// ---- Locks in the receiver resolution for `fn` ----
			{
				Code:   `import { rs } from '@rstest/core'; rs.fn();`,
				Errors: missing("fn", 1, 39, 41),
			},
			{
				// A renamed binding is the same function, so it is still a
				// mock with no call signature.
				Code:   `import { rs as vi } from '@rstest/core'; vi.fn();`,
				Errors: missing("fn", 1, 45, 47),
			},
			{
				Code:   `import { rstest as mocker } from '@rstest/core'; mocker.fn();`,
				Errors: missing("fn", 1, 57, 59),
			},
			{
				Code:   `const { rs } = require('@rstest/core'); rs.fn();`,
				Errors: missing("fn", 1, 44, 46),
			},
			{
				// A namespace import reaches the same object through one more
				// member.
				Code:   `import * as core from '@rstest/core'; core.rs.fn();`,
				Errors: missing("fn", 1, 47, 49),
			},
			{
				Code:   `const core = require('@rstest/core'); core.rstest.fn();`,
				Errors: missing("fn", 1, 51, 53),
			},

			// ---- Real-user: a mock created inline as an argument ----
			// The mock is just as untyped when its value is consumed on the
			// spot instead of being bound to a name.
			{
				Code:   `render(<Button onClick={rs.fn()} />);`,
				Tsx:    true,
				Errors: missing("fn", 1, 28, 30),
			},
			{
				Code:   `const handlers = { onSave: rs.fn(), onCancel: rs.fn() };`,
				Errors: append(missing("fn", 1, 31, 33), missing("fn", 1, 50, 52)...),
			},
			// ---- Real-user: the mock is configured on the same expression ----
			{
				Code:   `rs.fn().mockResolvedValue({ id: 1 });`,
				Errors: missing("fn", 1, 4, 6),
			},
			{
				Code:   `rs.spyOn(api, 'load').mockImplementation(rs.fn());`,
				Errors: missing("fn", 1, 45, 47),
			},

			// `fn` is checked whatever the option says, since the option only
			// widens the rule to the module loaders.
			{Code: `rs.fn()`, Options: emptyOptions, Errors: missing("fn", 1, 4, 6)},
			{Code: `rs.fn()`, Options: dontCheckImports, Errors: missing("fn", 1, 4, 6)},
			{Code: `rs.fn()`, Options: checkImports, Errors: missing("fn", 1, 4, 6)},

			// ---- Locks in that each call is judged on its own ----
			{
				Code: `import { rs } from '@rstest/core';
const load = rs.fn<() => void>();
const save = rs.fn();`,
				Errors: missing("fn", 3, 17, 19),
			},
		},
	)
}
