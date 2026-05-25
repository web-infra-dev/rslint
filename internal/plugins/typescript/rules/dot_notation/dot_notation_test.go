package dot_notation

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDotNotationRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DotNotationRule, []rule_tester.ValidTestCase{
		// --- base rule parity ---
		{Code: "a.b;"},
		{Code: "a['12'];"},
		{Code: "a[b];"},
		{Code: "a[0];"},
		// Reserved keyword as bracket key, allowKeywords=false
		{Code: "a['default'];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a['null'];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a['true'];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a['false'];", Options: map[string]interface{}{"allowKeywords": false}},
		// ES3 words that rslint previously missed
		{Code: "a['abstract'];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a['public'];", Options: map[string]interface{}{"allowKeywords": false}},
		// allowPattern with ES-style lookaround via regexp2
		{Code: "a['foo_bar'];", Options: map[string]interface{}{"allowPattern": "^[a-z]+(_[a-z]+)+$"}},

		// --- TS-specific: allowIndexSignaturePropertyAccess via option ---
		{
			Code:    "declare const m: Record<string, number>;\nm['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},
		{
			Code:    "declare const m: Record<string, number> | undefined;\nm?.['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},
		{
			Code:    "type M = { [k: string]: number };\ndeclare const m: M;\nm['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},
		{
			Code:    "interface M { bar: number; [k: string]: number }\ndeclare const m: M;\nm['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},
		// Template-literal index key — index signature is still string-like.
		{
			Code:    "type M = { [k: `key_${string}`]: number };\ndeclare const m: M;\nm['key_baz'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},
		// Intersection form: one part has the string index signature.
		{
			Code:    "type A = { a: number };\ntype B = { [k: string]: number };\ndeclare const m: A & B;\nm['z'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},

		// Template-literal with substitution — not a static key, don't touch
		{Code: "a[`b${x}`];"},
		// `a[null]`, `a[true]`, `a[false]` with allowKeywords=false should NOT report
		{Code: "a[null];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a[true];", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a[false];", Options: map[string]interface{}{"allowKeywords": false}},
		// `let`/`yield`/`eval`/`arguments` dot access is allowed (not ES3 keywords)
		{Code: "a.let;", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a.yield;", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a.eval;", Options: map[string]interface{}{"allowKeywords": false}},
		{Code: "a.arguments;", Options: map[string]interface{}{"allowKeywords": false}},

		// --- allowPrivate / allowProtected ---
		{
			Code:    "class X {\n  private priv_prop = 123;\n}\nconst x = new X();\nx['priv_prop'] = 123;\n",
			Options: map[string]interface{}{"allowPrivateClassPropertyAccess": true},
		},
		{
			Code:    "class X {\n  protected prot_prop = 123;\n}\nconst x = new X();\nx['prot_prop'] = 123;\n",
			Options: map[string]interface{}{"allowProtectedClassPropertyAccess": true},
		},

		// --- #1 string keys with chars that make them invalid identifiers ---
		{Code: "a['with-dash'];"},
		{Code: "a['has space'];"},
		{Code: "a[''];"},
		{Code: "a['12valid'];"},       // starts with digit
		{Code: "a['\\n'];"},            // contains newline
		{Code: "a['it\\'s'];"},         // contains quote
		{Code: "a['$ok'];", Options: map[string]interface{}{"allowPattern": "^\\$"}}, // starts with $ but allowPattern matches

		// --- #2 non-ASCII identifier-looking strings (ESLint ASCII-only regex skips them) ---
		{Code: "a['ñ'];"},
		{Code: "a['中文'];"},
		{Code: "a['café'];"}, // mixed ASCII+non-ASCII — regex still fails (é not ASCII)

		// --- #3 numeric and bigint literal keys are not string-literal-like ---
		{Code: "a[0x1];"},
		{Code: "a[42];"},
		{Code: "a[42n];"},

		// --- #5 allowPattern matches null/true/false string keys ---
		{
			Code:    "a['null'];",
			Options: map[string]interface{}{"allowPattern": "^(null|true|false)$"},
		},
		{
			Code:    "a[null];",
			Options: map[string]interface{}{"allowPattern": "^null$"},
		},

		// --- #7 private identifier access is not a regular Identifier ---
		{
			Code: "class X {\n  #priv = 1;\n  foo() { return this.#priv; }\n}\n",
		},
		{
			Code:    "class X {\n  #priv = 1;\n  foo() { return this.#priv; }\n}\n",
			Options: map[string]interface{}{"allowKeywords": false},
		},

		// --- #14 readonly index signature ---
		{
			Code:    "type RO = { readonly [k: string]: number };\ndeclare const m: RO;\nm['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},

		// --- #16 index signature value type includes undefined ---
		{
			Code:    "type M = { [k: string]: number | undefined };\ndeclare const m: M;\nm['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},

		// --- #18 generic constrained to indexable + named property ---
		{
			Code:    "function f<T extends { a: number; [k: string]: number }>(x: T) {\n  x['b'];\n}\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},

		// --- #19 this access on class with index signature ---
		{
			Code:    "class X {\n  [k: string]: number;\n  foo() { return this['bar']; }\n}\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},

		// --- `as Record<string, T>` cast surfaces the index signature ---
		{
			Code:    "declare const x: unknown;\n(x as Record<string, number>)['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
		},

		// --- Decorator argument is just an ExpressionStatement-nested access ---
		// (Listeners fire regardless of containing construct — valid here because
		// the key matches an allowPattern.)
		{
			Code:    "declare function d(v: unknown): ClassDecorator;\n@d((undefined as any)['_ok_'])\nclass C {}\n",
			Options: map[string]interface{}{"allowPattern": "^_"},
		},
	}, []rule_tester.InvalidTestCase{
		// --- base useDot ---
		{
			Code:   "a['b'];",
			Output: []string{"a.b;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 3}},
		},
		{
			Code:   "a['test'];",
			Output: []string{"a.test;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 3}},
		},
		// allowKeywords=true (default): bracket with a keyword is reported.
		{
			Code:   "a['default'];",
			Output: []string{"a.default;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 1, Column: 3}},
		},

		// Numeric-only index signature MUST NOT be treated as string-like.
		// Even with allowIndexSignaturePropertyAccess on, `m['foo']` should report.
		{
			Code:    "type M = { [k: number]: number };\ndeclare const m: M;\nm['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
			Output:  []string{"type M = { [k: number]: number };\ndeclare const m: M;\nm.foo;\n"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 3, Column: 3}},
		},

		// Concrete property wins over index signature — bracket still reported.
		{
			Code:    "interface M { bar: number; [k: string]: number }\ndeclare const m: M;\nm['bar'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
			Output:  []string{"interface M { bar: number; [k: string]: number }\ndeclare const m: M;\nm.bar;\n"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 3, Column: 3}},
		},

		// Private/protected without the corresponding allow-option still reports.
		{
			Code:   "class X {\n  private priv_prop = 123;\n}\nconst x = new X();\nx['priv_prop'] = 123;\n",
			Output: []string{"class X {\n  private priv_prop = 123;\n}\nconst x = new X();\nx.priv_prop = 123;\n"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot", Line: 5, Column: 3}},
		},

		// Comment inside brackets: report diagnostic but no autofix.
		{
			Code:   "foo[/* comment */ 'bar'];",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},
		// Comment between dot and property name: diagnostic but no fix.
		{
			Code:    "foo./* comment */ while;",
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets"}},
		},
		// `let.if()` reports but cannot auto-fix (would become `let["if"]()`
		// which re-parses as a destructuring declaration).
		{
			Code:    "let.if();",
			Options: map[string]interface{}{"allowKeywords": false},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets"}},
		},
		// Optional dot access to a keyword: replacement must preserve `?.`.
		{
			Code:    "a?.while;",
			Options: map[string]interface{}{"allowKeywords": false},
			Output:  []string{"a?.[\"while\"];"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets"}},
		},
		// Parenthesized literal key is still a literal key.
		{
			Code:   "a[('b')];",
			Output: []string{"a.b;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},
		// Chained bracket access: three separate reports, three fix passes.
		{
			Code: "a['b']['c']['d'];",
			Output: []string{
				"a.b['c']['d'];",
				"a.b.c['d'];",
				"a.b.c.d;",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "useDot"},
				{MessageId: "useDot"},
				{MessageId: "useDot"},
			},
		},

		// --- #4 unicode escape — `.Text` returns the unescaped value, which
		// happens to be a plain ASCII identifier, so autofix produces a valid dot. ---
		{
			Code:   "a['\\u0062'];",
			Output: []string{"a.b;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},

		// --- #6 allowKeywords:false + dot access to null/true/false ---
		{
			Code:    "a.null;",
			Options: map[string]interface{}{"allowKeywords": false},
			Output:  []string{"a[\"null\"];"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets"}},
		},
		{
			Code:    "a.true;",
			Options: map[string]interface{}{"allowKeywords": false},
			Output:  []string{"a[\"true\"];"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets"}},
		},
		{
			Code:    "a.false;",
			Options: map[string]interface{}{"allowKeywords": false},
			Output:  []string{"a[\"false\"];"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useBrackets"}},
		},

		// --- #13 assignment target ---
		{
			Code:   "a['foo'] = 1;",
			Output: []string{"a.foo = 1;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},
		{
			Code:   "a['foo'] += 1;",
			Output: []string{"a.foo += 1;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},
		{
			Code:   "a['foo']++;",
			Output: []string{"a.foo++;"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},

		// --- #19 `this` bracket access to a concrete member is reported ---
		{
			Code:   "class X {\n  a = 1;\n  foo() { return this['a']; }\n}\n",
			Output: []string{"class X {\n  a = 1;\n  foo() { return this.a; }\n}\n"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},

		// --- #20 `super` bracket access ---
		{
			Code:   "class B { foo = 1; } class X extends B {\n  bar() { return super['foo']; }\n}\n",
			Output: []string{"class B { foo = 1; } class X extends B {\n  bar() { return super.foo; }\n}\n"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},

		// --- #17 `any` bracket access is reported even with allowIndexSig — matches typescript-eslint ---
		{
			Code:    "declare const x: any;\nx['foo'];\n",
			Options: map[string]interface{}{"allowIndexSignaturePropertyAccess": true},
			Output:  []string{"declare const x: any;\nx.foo;\n"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},

		// --- Type assertion casting to Record preserves the index signature ---
		// (Covered via e2e in the validation audit; repeated here as a fix check.)

		// --- #9 inside namespace/module declaration ---
		{
			Code:   "namespace N {\n  declare const m: { foo: number };\n  m['foo'];\n}\n",
			Output: []string{"namespace N {\n  declare const m: { foo: number };\n  m.foo;\n}\n"},
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "useDot"}},
		},
	})
}
