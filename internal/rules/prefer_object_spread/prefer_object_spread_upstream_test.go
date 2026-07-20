package prefer_object_spread

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferObjectSpreadUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/prefer-object-spread.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases
// live in the prefer_object_spread_extras_test.go file.
func TestPreferObjectSpreadUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferObjectSpreadRule,
		[]rule_tester.ValidTestCase{
			{Code: `Object.assign()`},
			{Code: `let a = Object.assign(a, b)`},
			{Code: `Object.assign(a, b)`},
			{Code: `let a = Object.assign(b, { c: 1 })`},
			{Code: `const bar = { ...foo }`},
			{Code: `Object.assign(...foo)`},
			{Code: `Object.assign(foo, { bar: baz })`},
			{Code: `Object.assign({}, ...objects)`},
			{Code: `foo({ foo: 'bar' })`},
			{Code: "\n        const Object = {};\n        Object.assign({}, foo);\n        "},
			// SKIP: relies on ESLint's ReferenceTracker refusing to track a
			// global variable that is ever the target of a bare (declaration-less)
			// assignment anywhere in the file. rslint's shadow check only detects
			// shadowing via a new declaration — see "Differences from ESLint".
			{Code: "\n        Object = {};\n        Object.assign({}, foo);\n        ", Skip: true},
			{Code: "\n        const Object = {};\n        Object.assign({ foo: 'bar' });\n        "},
			// SKIP: see the bare-reassignment note above.
			{Code: "\n        Object = {};\n        Object.assign({ foo: 'bar' });\n        ", Skip: true},
			{Code: "\n        const Object = require('foo');\n        Object.assign({ foo: 'bar' });\n        "},
			{Code: "\n        import Object from 'foo';\n        Object.assign({ foo: 'bar' });\n        "},
			{Code: "\n        import { Something as Object } from 'foo';\n        Object.assign({ foo: 'bar' });\n        "},
			{Code: "\n        import { Object, Array } from 'globals';\n        Object.assign({ foo: 'bar' });\n        "},
			// SKIP: relies on ESLint's ecmaVersion 2018 default not declaring
			// `globalThis` as a known global, so ReferenceTracker never chains
			// through it. rslint always recognizes `globalThis` regardless of
			// any ecmaVersion-like setting — see "Differences from ESLint".
			{Code: `globalThis.Object.assign({}, foo)`, Skip: true},
			{Code: `globalThis.Object.assign({}, { foo: 'bar' })`, Skip: true},
			{Code: `globalThis.Object.assign({}, baz, { foo: 'bar' })`, Skip: true},
			{Code: "\n                var globalThis = foo;\n                globalThis.Object.assign({}, foo)\n                "},
			{Code: `class C { #assign; foo() { Object.#assign({}, foo); } }`},

			// ---- ignore Object.assign() with > 1 arguments if any of the arguments is an object expression with a getter/setter ----
			{Code: `Object.assign({ get a() {} }, {})`},
			{Code: `Object.assign({ set a(val) {} }, {})`},
			{Code: `Object.assign({ get a() {} }, foo)`},
			{Code: `Object.assign({ set a(val) {} }, foo)`},
			{Code: `Object.assign({ foo: 'bar', get a() {}, baz: 'quux' }, quuux)`},
			{Code: `Object.assign({ foo: 'bar', set a(val) {} }, { baz: 'quux' })`},
			{Code: `Object.assign({}, { get a() {} })`},
			{Code: `Object.assign({}, { set a(val) {} })`},
			{Code: `Object.assign({}, { foo: 'bar', get a() {} }, {})`},
			{Code: `Object.assign({ foo }, bar, {}, { baz: 'quux', set a(val) {}, quuux }, {})`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `Object.assign({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign  ({}, foo)`,
				Output: []string{`({ ...foo})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign({}, { foo: 'bar' })`,
				Output: []string{`({ foo: 'bar'})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign({}, baz, { foo: 'bar' })`,
				Output: []string{`({ ...baz, foo: 'bar'})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign({}, { foo: 'bar', baz: 'foo' })`,
				Output: []string{`({ foo: 'bar', baz: 'foo'})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign({ foo: 'bar' }, baz)`,
				Output: []string{`({foo: 'bar', ...baz})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Many args ----
			{
				Code:   `Object.assign({ foo: 'bar' }, cats, dogs, trees, birds)`,
				Output: []string{`({foo: 'bar', ...cats, ...dogs, ...trees, ...birds})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				// Upstream's own RuleTester asserts only the first (single) fix
				// pass here — the outer call's fix range subsumes the inner
				// call, so ESLint's fixer applies just the outer fix in one
				// pass and leaves the inner Object.assign call for the next
				// lint pass. rslint's tester fixes to convergence, so the
				// second (fully-converged) pass is asserted too.
				Code:   `Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, baz))`,
				Output: []string{`({foo: 'bar', ...Object.assign({ bar: 'foo' }, baz)})`, `({foo: 'bar', ...({bar: 'foo', ...baz})})`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useSpreadMessage", Line: 1, Column: 1},
					{MessageId: "useSpreadMessage", Line: 1, Column: 31},
				},
			},
			{
				// Same convergence note as the previous case, one nesting level deeper.
				Code: `Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, Object.assign({}, { superNested: 'butwhy' })))`,
				Output: []string{
					`({foo: 'bar', ...Object.assign({ bar: 'foo' }, Object.assign({}, { superNested: 'butwhy' }))})`,
					`({foo: 'bar', ...({bar: 'foo', ...Object.assign({}, { superNested: 'butwhy' })})})`,
					`({foo: 'bar', ...({bar: 'foo', ...({ superNested: 'butwhy'})})})`,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "useSpreadMessage", Line: 1, Column: 1},
					{MessageId: "useSpreadMessage", Line: 1, Column: 31},
					{MessageId: "useSpreadMessage", Line: 1, Column: 61},
				},
			},

			// ---- Mix spread in argument ----
			{
				Code:   `Object.assign({foo: 'bar', ...bar}, baz)`,
				Output: []string{`({foo: 'bar', ...bar, ...baz})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Object shorthand ----
			{
				Code:   `Object.assign({}, { foo, bar, baz })`,
				Output: []string{`({ foo, bar, baz})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Objects with computed properties ----
			{
				Code:   `Object.assign({}, { [bar]: 'foo' })`,
				Output: []string{`({ [bar]: 'foo'})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Objects with spread properties ----
			{
				Code:   `Object.assign({ ...bar }, { ...baz })`,
				Output: []string{`({...bar, ...baz})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- Multiline objects ----
			{
				Code: "Object.assign({ ...bar }, {\n                // this is a bar\n                foo: 'bar',\n                baz: \"cats\"\n            })",
				Output: []string{
					"({...bar, // this is a bar\n                foo: 'bar',\n                baz: \"cats\"})",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code: "Object.assign({\n                boo: \"lol\",\n                // I'm a comment\n                dog: \"cat\"\n             }, {\n                // this is a bar\n                foo: 'bar',\n                baz: \"cats\"\n            })",
				Output: []string{
					"({boo: \"lol\",\n                // I'm a comment\n                dog: \"cat\", // this is a bar\n                foo: 'bar',\n                baz: \"cats\"})",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},

			// ---- HTML comment ----
			// SKIP: requires Annex B `sourceType: "script"` HTML-comment
			// syntax (`<!-- -->` as a line comment), which rslint's
			// TypeScript-based parser does not support.
			{
				Code:   "const test = Object.assign({ ...bar }, {\n                <!-- html comment\n                foo: 'bar',\n                baz: \"cats\"\n                --> weird\n            })",
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 14}},
			},
			{
				Code: "const test = Object.assign({ ...bar }, {\n                foo: 'bar', // inline comment\n                baz: \"cats\"\n            })",
				Output: []string{
					"const test = {...bar, foo: 'bar', // inline comment\n                baz: \"cats\"}",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 14}},
			},
			{
				Code: "const test = Object.assign({ ...bar }, {\n                /**\n                 * foo\n                 */\n                foo: 'bar',\n                baz: \"cats\"\n            })",
				Output: []string{
					"const test = {...bar, /**\n                 * foo\n                 */\n                foo: 'bar',\n                baz: \"cats\"}",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 14}},
			},

			// ---- ASI safety ----
			{
				Code:   "const result = doSomething()\nObject.assign({}, myData)",
				Output: []string{"const result = doSomething()\n;({ ...myData})"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 1}},
			},
			{
				Code:   `let a = foo + Object.assign({}, bar)`,
				Output: []string{`let a = foo + ({ ...bar})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 15}},
			},
			{
				Code:   "let foo = function() {};\nfoo\nObject.assign({}, bar)",
				Output: []string{"let foo = function() {};\nfoo\n;({ ...bar})"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 3, Column: 1}},
			},
			{
				Code:   "foo\nObject.assign({ foo: bar })",
				Output: []string{"foo\n;({foo: bar})"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 2, Column: 1}},
			},
			{
				Code:   "foo();\nObject.assign({}, bar)",
				Output: []string{"foo();\n({ ...bar})"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 1}},
			},
			{
				Code:   "const x = [1]\nObject.assign({}, bar)",
				Output: []string{"const x = [1]\n;({ ...bar})"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 1}},
			},
			{
				Code:   "foo\nObject.assign({}, bar).doSomething()",
				Output: []string{"foo\n;({ ...bar}).doSomething()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 1}},
			},
			{
				Code:   "foo\nObject.assign({}, bar), 2",
				Output: []string{"foo\n;({ ...bar}), 2"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 1}},
			},

			// ---- single-argument object literal ----
			{
				Code:   `Object.assign({})`,
				Output: []string{`({})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign({ foo: bar })`,
				Output: []string{`({foo: bar})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 1, Column: 1}},
			},
			{
				Code:   "\n                const foo = 'bar';\n                Object.assign({ foo: bar })\n            ",
				Output: []string{"\n                const foo = 'bar';\n                ({foo: bar})\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 3, Column: 17}},
			},
			{
				Code:   "\n                foo = 'bar';\n                Object.assign({ foo: bar })\n            ",
				Output: []string{"\n                foo = 'bar';\n                ({foo: bar})\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 3, Column: 17}},
			},
			{
				Code:   `let a = Object.assign({})`,
				Output: []string{`let a = {}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 1, Column: 9}},
			},
			{
				Code:   `let a = Object.assign({}, a)`,
				Output: []string{`let a = { ...a}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 9}},
			},
			{
				Code:   `let a = Object.assign   ({}, a)`,
				Output: []string{`let a = { ...a}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 9}},
			},
			{
				Code:   `let a = Object.assign({ a: 1 }, b)`,
				Output: []string{`let a = {a: 1, ...b}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 9}},
			},
			{
				Code:   `Object.assign(  {},  a,      b,   )`,
				Output: []string{`({    ...a,      ...b,   })`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `Object.assign({}, a ? b : {}, b => c, a = 2)`,
				Output: []string{`({ ...(a ? b : {}), ...(b => c), ...(a = 2)})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
			{
				Code:   "\n                const someVar = 'foo';\n                Object.assign({}, a ? b : {}, b => c, a = 2)\n            ",
				Output: []string{"\n                const someVar = 'foo';\n                ({ ...(a ? b : {}), ...(b => c), ...(a = 2)})\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 3, Column: 17}},
			},
			{
				Code:   "\n                someVar = 'foo';\n                Object.assign({}, a ? b : {}, b => c, a = 2)\n            ",
				Output: []string{"\n                someVar = 'foo';\n                ({ ...(a ? b : {}), ...(b => c), ...(a = 2)})\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 3, Column: 17}},
			},

			// ---- Cases where you don't need parens around an object literal ----
			{
				Code:   `[1, 2, Object.assign({}, a)]`,
				Output: []string{`[1, 2, { ...a}]`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 8}},
			},
			{
				Code:   `const foo = Object.assign({}, a)`,
				Output: []string{`const foo = { ...a}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 13}},
			},
			{
				Code:   `function foo() { return Object.assign({}, a) }`,
				Output: []string{`function foo() { return { ...a} }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 25}},
			},
			{
				Code:   `foo(Object.assign({}, a));`,
				Output: []string{`foo({ ...a});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 5}},
			},
			{
				Code:   `const x = { foo: 'bar', baz: Object.assign({}, a) }`,
				Output: []string{`const x = { foo: 'bar', baz: { ...a} }`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 30}},
			},
			{
				Code:   "\n                import Foo from 'foo';\n                Object.assign({ foo: Foo });\n            ",
				Output: []string{"\n                import Foo from 'foo';\n                ({foo: Foo});\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 3, Column: 17}},
			},
			{
				Code:   "\n                import Foo from 'foo';\n                Object.assign({}, Foo);\n            ",
				Output: []string{"\n                import Foo from 'foo';\n                ({ ...Foo});\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 3, Column: 17}},
			},
			{
				Code:   "\n                const Foo = require('foo');\n                Object.assign({ foo: Foo });\n            ",
				Output: []string{"\n                const Foo = require('foo');\n                ({foo: Foo});\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 3, Column: 17}},
			},
			{
				Code:   "\n                import { Something as somethingelse } from 'foo';\n                Object.assign({}, somethingelse);\n            ",
				Output: []string{"\n                import { Something as somethingelse } from 'foo';\n                ({ ...somethingelse});\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 3, Column: 17}},
			},
			{
				Code:   "\n                import { foo } from 'foo';\n                Object.assign({ foo: Foo });\n            ",
				Output: []string{"\n                import { foo } from 'foo';\n                ({foo: Foo});\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 3, Column: 17}},
			},
			{
				Code:   "\n                const Foo = require('foo');\n                Object.assign({}, Foo);\n            ",
				Output: []string{"\n                const Foo = require('foo');\n                ({ ...Foo});\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 3, Column: 17}},
			},
			{
				Code: "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput,\n                    },\n                    this.props.actions\n                );\n            ",
				Output: []string{
					"\n                const actions = {\n                    onChangeInput: this.handleChangeInput,\n                    ...this.props.actions\n                };\n            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 33}},
			},
			{
				Code: "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput, //\n                    },\n                    this.props.actions\n                );\n            ",
				Output: []string{
					"\n                const actions = {\n                    onChangeInput: this.handleChangeInput, //\n                    \n                    ...this.props.actions\n                };\n            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 33}},
			},
			{
				Code: "\n                const actions = Object.assign(\n                    {\n                        onChangeInput: this.handleChangeInput //\n                    },\n                    this.props.actions\n                );\n            ",
				Output: []string{
					"\n                const actions = {\n                    onChangeInput: this.handleChangeInput //\n                    ,\n                    ...this.props.actions\n                };\n            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 33}},
			},
			{
				Code: "\n                const actions = Object.assign(\n                    (\n                        {\n                            onChangeInput: this.handleChangeInput\n                        }\n                    ),\n                    (\n                        this.props.actions\n                    )\n                );\n            ",
				Output: []string{
					"\n                const actions = {\n                    \n                            onChangeInput: this.handleChangeInput\n                        ,\n                    ...(\n                        this.props.actions\n                    )\n                };\n            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 33}},
			},
			{
				Code:   "\n                eventData = Object.assign({}, eventData, { outsideLocality: `${originLocality} - ${destinationLocality}` })\n            ",
				Output: []string{"\n                eventData = { ...eventData, outsideLocality: `${originLocality} - ${destinationLocality}`}\n            "},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 2, Column: 29}},
			},

			// ---- https://github.com/eslint/eslint/issues/10646 ----
			{
				Code:   `Object.assign({ });`,
				Output: []string{`({});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 1, Column: 1}},
			},
			{
				Code:   "Object.assign({\n});",
				Output: []string{`({});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 1, Column: 1}},
			},
			{
				Code:   `globalThis.Object.assign({ });`,
				Output: []string{`({});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 1, Column: 1}},
			},
			{
				Code:   "globalThis.Object.assign({\n});",
				Output: []string{`({});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 1, Column: 1}},
			},
			{
				Code: "\n                function foo () { var globalThis = bar; }\n                globalThis.Object.assign({ });\n            ",
				Output: []string{
					"\n                function foo () { var globalThis = bar; }\n                ({});\n            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 3, Column: 17}},
			},
			{
				Code: "\n                const Foo = require('foo');\n                globalThis.Object.assign({ foo: Foo });\n            ",
				Output: []string{
					"\n                const Foo = require('foo');\n                ({foo: Foo});\n            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 3, Column: 17}},
			},

			// ---- report Object.assign() with getters/setters if the function call has only 1 argument ----
			{
				Code:   `Object.assign({ get a() {}, set b(val) {} })`,
				Output: []string{`({get a() {}, set b(val) {}})`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useLiteralMessage", Line: 1, Column: 1}},
			},

			// ---- https://github.com/eslint/eslint/issues/13058 ----
			// Upstream uses a custom TS-generic parser fixture here because
			// espree can't parse call type arguments; rslint parses TypeScript
			// natively, so these are migrated as regular TS syntax instead.
			{
				Code:   `const obj = Object.assign<{}, Record<string, string[]>>({}, getObject());`,
				Output: []string{`const obj = { ...getObject()};`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 13}},
			},
			{
				Code:   `Object.assign<{}, A>({}, foo);`,
				Output: []string{`({ ...foo});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useSpreadMessage", Line: 1, Column: 1}},
			},
		},
	)
}
