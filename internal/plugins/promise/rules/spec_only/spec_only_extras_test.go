package spec_only_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/promise/rules/spec_only"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Additional cases checked against eslint-plugin-promise v7.3.0 with
// @typescript-eslint/parser 8.65.0, or ESLint's default parser for JavaScript.
// Go assertions include complete diagnostic ranges.
func TestSpecOnlyExtras(t *testing.T) {
	root := fixtures.GetRootDir()
	// Use tsgo paths for the in-memory filesystem on every platform.
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "tsconfig.allowJs.json"): `{"extends":"./tsconfig.json","compilerOptions":{"allowJs":true}}`,
	})
	rule_tester.RunRuleTester(root, "tsconfig.json", t, &spec_only.SpecOnlyRule,
		[]rule_tester.ValidTestCase{
			// Every standard method, including statics missing from upstream tests.
			{Code: `Promise.allSettled([]); Promise.any([]); Promise["withResolvers"](); Promise.prototype.then; Promise.prototype.finally; Promise.prototype["catch"];`},
			// Identifiers remain dynamic even when a variable has a known string value.
			{Code: `const method = "done"; Promise[(method)]; Promise.prototype[(method)]; Promise[prototype]; Promise[prototype].then;`},
			// Only direct Promise members are checked; promise instances and aliases are not tracked.
			{Code: `const P = Promise; P.done(); globalThis.Promise.done(); Promise.resolve().done(); new Promise(fn).done(); Promise().done(); data.a.b.c.d.e.f.g.Promise.done;`},
			// Static and instance allow lists share names and accept string keys.
			{Code: `Promise.done(); Promise["done"]; Promise.prototype.done; Promise.prototype["done"]; Promise["prototype"].done;`,
				Options: []any{map[string]any{"allowedMethods": []any{"done", "prototype", "done"}}}},
			// String values are matched exactly, including empty and Unicode names.
			{Code: `Promise[""]; Promise["完成"];`,
				Options: []any{map[string]any{"allowedMethods": []any{"", "完成"}}}},
			// Standard methods remain valid with an explicitly empty options object.
			{Code: `Promise.resolve(); Promise.prototype.then;`,
				Options: []any{map[string]any{}}},
			// Parentheses terminating optional chains introduce ChainExpression parents.
			{Code: `(Promise?.prototype).done; (Promise?.prototype)?.done; target[Promise?.prototype]; Promise?.prototype;`},
			// TS assertions and non-null wrappers remain visible to ESTree.
			{Code: `(Promise as any).done; Promise!.done; (Promise satisfies any).done; (Promise.prototype as any).done; Promise.prototype!.done; (Promise.prototype satisfies any).done;`},
			// Dotted JSX tag names and ordinary type names are not member expressions.
			{Code: `const node = <Promise.done />; const nested = <Other.Promise.done />; const deep = <Promise.foo.bar />; type T = Promise.done; type U = typeof Promise.done;`,
				Tsx: true},
			// A private prototype name takes the upstream prototype branch, but instance methods must still be permitted.
			{Code: `class C { #prototype; m() { Promise.#prototype; Promise.#prototype.then; } }`},
			// Allowed names are strings, including Object.prototype names and empty keys.
			{Code: `Promise.__proto__; Promise.constructor; Promise.toString; Promise[""];`,
				Options: []any{map[string]any{"allowedMethods": []any{"__proto__", "constructor", "toString", ""}}},
			},
			// A JSDoc cast does not remove the ChainExpression terminating optional access.
			{Code: `(/** @type {any} */ (Promise?.prototype)).done;`,
				FileName: "chain.js", TSConfig: "tsconfig.allowJs.json",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Both explicit defaults reject a non-standard method.
			{Code: `Promise.done;`,
				Options: []any{map[string]any{}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				}},
			{Code: `Promise.done;`,
				Options: []any{map[string]any{"allowedMethods": []any{}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
				}},
			// Static and prototype standard sets are distinct; prototype reports cover the receiver.
			{Code: `Promise.then; Promise.catch; Promise.finally; Promise.prototype.resolve; Promise.prototype.withResolvers;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.then'", Line: 1, Column: 1, EndLine: 1, EndColumn: 13},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.catch'", Line: 1, Column: 15, EndLine: 1, EndColumn: 28},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.finally'", Line: 1, Column: 30, EndLine: 1, EndColumn: 45},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 47, EndLine: 1, EndColumn: 64},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 74, EndLine: 1, EndColumn: 91},
				}},
			// The special prototype branch uses property.name, including computed identifiers.
			{Code: `Promise["prototype"]; Promise[prototype].done; Promise.prototype["done"]; target[Promise.prototype];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 23, EndLine: 1, EndColumn: 41},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 48, EndLine: 1, EndColumn: 65},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 82, EndLine: 1, EndColumn: 99},
				}},
			// Allowing prototype does not allow arbitrary prototype members.
			{Code: `Promise.prototype.done;`,
				Options: []any{map[string]any{"allowedMethods": []any{"prototype"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 1, EndLine: 1, EndColumn: 18},
				}},
			// Parentheses are transparent around receivers, keys and prototype member expressions.
			{Code: `((Promise)).done; (Promise.prototype).done; Promise[(("done"))]; Promise.prototype[("done")];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 1, EndLine: 1, EndColumn: 17},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 20, EndLine: 1, EndColumn: 37},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 45, EndLine: 1, EndColumn: 64},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 66, EndLine: 1, EndColumn: 83},
				}},
			// Optional access reports members within a chain, including consecutive optional links.
			{Code: `Promise?.done?.(); Promise?.["done"]; Promise?.prototype.done; Promise.prototype?.done; Promise?.prototype?.done;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 20, EndLine: 1, EndColumn: 37},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 39, EndLine: 1, EndColumn: 57},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 64, EndLine: 1, EndColumn: 81},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 89, EndLine: 1, EndColumn: 107},
				}},
			// Numeric, boolean, null, bigint and regexp keys are never coerced for allowedMethods.
			{Code: `Promise[0x10]; Promise[1e21]; Promise[1_000]; Promise[true]; Promise[false]; Promise[null]; Promise[0x10n]; Promise[/done/mi];`,
				Options: []any{map[string]any{"allowedMethods": []any{"16", "1e+21", "1000", "true", "false", "null", "/done/im"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.16'", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.1e+21'", Line: 1, Column: 16, EndLine: 1, EndColumn: 29},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.1000'", Line: 1, Column: 31, EndLine: 1, EndColumn: 45},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.true'", Line: 1, Column: 47, EndLine: 1, EndColumn: 60},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.false'", Line: 1, Column: 62, EndLine: 1, EndColumn: 76},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.null'", Line: 1, Column: 78, EndLine: 1, EndColumn: 91},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.16'", Line: 1, Column: 93, EndLine: 1, EndColumn: 107},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise./done/im'", Line: 1, Column: 109, EndLine: 1, EndColumn: 126},
				}},
			// Expressions and templates have no ESTree name or literal value.
			{Code: "Promise[`resolve`]; Promise[`done${suffix}`]; Promise[\"re\" + \"solve\"]; Promise[getMethod()]; Promise[-1]; Promise[{}]; Promise[[]];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 1, EndLine: 1, EndColumn: 19},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 21, EndLine: 1, EndColumn: 45},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 47, EndLine: 1, EndColumn: 70},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 72, EndLine: 1, EndColumn: 92},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 94, EndLine: 1, EndColumn: 105},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 107, EndLine: 1, EndColumn: 118},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 120, EndLine: 1, EndColumn: 131},
				}},
			// Private properties report their name without # and cannot be allowed by string.
			{Code: `class C { #done; #prototype; m() { Promise.#done; Promise.#prototype.done; Promise.prototype.#done; } }`,
				Options: []any{map[string]any{"allowedMethods": []any{"done"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 36, EndLine: 1, EndColumn: 49},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 76, EndLine: 1, EndColumn: 93},
				}},
			// The rule intentionally checks shadowed identifiers named Promise.
			{Code: `function run(Promise) { return Promise.done(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 32, EndLine: 1, EndColumn: 44},
				}},
			// TS wrappers around a property value have neither a name nor literal value.
			{Code: `Promise["resolve" as string]; Promise[method!]; Promise.done<string>();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 1, EndLine: 1, EndColumn: 29},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.undefined'", Line: 1, Column: 31, EndLine: 1, EndColumn: 47},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 49, EndLine: 1, EndColumn: 61},
				}},
			// TS heritage names are ESTree members; heritage type arguments remain types.
			{Code: `class C extends Promise.done {} interface I extends Promise.done {} class D implements Promise.prototype.done {} interface J extends Base<Promise.done> {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 17, EndLine: 1, EndColumn: 29},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 53, EndLine: 1, EndColumn: 65},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 88, EndLine: 1, EndColumn: 105},
				}},
			// Runtime JSX expressions are checked while dotted tags are ignored.
			{Code: `const node = <Promise.done value={Promise.done}>{Promise.prototype.done}</Promise.done>;`,
				Tsx: true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 35, EndLine: 1, EndColumn: 47},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 50, EndLine: 1, EndColumn: 67},
				}},
			// JSDoc casts in JavaScript are transparent to ESTree.
			{Code: `/** @type {any} */ (Promise).done; (/** @type {any} */ (Promise.prototype)).done;`,
				FileName: "cast.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 20, EndLine: 1, EndColumn: 34},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 1, Column: 57, EndLine: 1, EndColumn: 74},
				}},
			// Unicode text and line breaks require complete UTF-16 diagnostic ranges.
			{Code: `"😀"; Promise["完成"]
Promise
  .prototype
  .done;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.完成'", Line: 1, Column: 7, EndLine: 1, EndColumn: 20},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 2, Column: 1, EndLine: 3, EndColumn: 13},
				}},
			// Escaped names are decoded before matching and reporting.
			// cspell:disable-next-line
			{Code: `Pr\u006fmise.d\u006fne; Promise["d\u006fne"];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 1, EndLine: 1, EndColumn: 23},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 25, EndLine: 1, EndColumn: 45},
				}},
			// A Bluebird-style migration has nested static references and prototype extensions.
			{Code: `import Promise from "bluebird";
export const load = files => Promise.map(files, readFile).tap(log);
Promise.prototype.timeout = timeout;
const join = Promise.join;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.map'", Line: 2, Column: 30, EndLine: 2, EndColumn: 41},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 3, Column: 1, EndLine: 3, EndColumn: 18},
					{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.join'", Line: 4, Column: 14, EndLine: 4, EndColumn: 26},
				}},
			// Numeric rounding and regexp stringification use the shared ECMAScript helpers.
			// Regexp modifiers were checked against ESLint on Node 24.19.0.
			{Code: `Promise[0x20000000000001]; Promise[1e309]; Promise[0x100000000000000000000n]; Promise[/a/yg]; Promise[/(?i:a)/];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.9007199254740992'", Line: 1, Column: 1, EndLine: 1, EndColumn: 26}, {MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.Infinity'", Line: 1, Column: 28, EndLine: 1, EndColumn: 42}, {MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.1208925819614629174706176'", Line: 1, Column: 44, EndLine: 1, EndColumn: 77}, {MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise./a/gy'", Line: 1, Column: 79, EndLine: 1, EndColumn: 93}, {MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise./(?i:a)/'", Line: 1, Column: 95, EndLine: 1, EndColumn: 112}},
			},
			// BOM, CRLF and Unicode line separators preserve the full member range.
			{Code: "\ufeff// before\r\nPromise\u2028 .prototype\u2029 .done;",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.prototype'", Line: 2, Column: 1, EndLine: 3, EndColumn: 12}},
			},
			// Line directives suppress diagnostics before and beside a statement.
			{Code: `// eslint-disable-next-line
Promise.done;
Promise.done; // eslint-disable-line
Promise.something;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.something'", Line: 4, Column: 1, EndLine: 4, EndColumn: 18}},
			},
			// Block directives can re-enable reporting later on the same line.
			{Code: `/* eslint-disable */ Promise.done; /* eslint-enable */ Promise.done;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 56, EndLine: 1, EndColumn: 68}},
			},
			// ESLint accepts optional private access, but tsgo reports TS18030 before rule execution.
			{Code: `class C { #done; m() { Promise?.#done; } }`,
				FileName: "private.js", TSConfig: "tsconfig.allowJs.json",
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "avoidNonStandard", Message: "Avoid using non-standard 'Promise.done'", Line: 1, Column: 24, EndLine: 1, EndColumn: 38}},
			},
		},
	)
}
