package no_useless_backreference

import (
	"fmt"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoUselessBackreferenceRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUselessBackreferenceRule,
		[]rule_tester.ValidTestCase{
			// ---- not a regular expression ----
			{Code: `'\\1(a)'`},
			{Code: `regExp('\\1(a)')`},
			{Code: `new Regexp('\\1(a)', 'u')`},
			{Code: `RegExp.foo('\\1(a)', 'u')`},
			{Code: `new foo.RegExp('\\1(a)')`},

			// ---- unknown pattern ----
			{Code: `RegExp(p)`},
			{Code: `new RegExp(p, 'u')`},
			{Code: `RegExp('\\1(a)' + suffix)`},
			{Code: "new RegExp(`${prefix}\\\\1(a)`)"},
			{Code: "{ const String = { raw: () => '\\\\1(a)' }; new RegExp(String.raw`\\1(a)`); }"},
			{Code: `RegExp(1); RegExp(1n); RegExp(true); RegExp(false); RegExp(null);`},
			{Code: `
declare function unknownPattern(): unknown;
RegExp({});
RegExp([]);
RegExp(() => '\\1(a)');
RegExp(function () { return '\\1(a)'; });
RegExp(class {});
RegExp(unknownPattern());
RegExp(new String('\\1(a)'));`},

			// ---- not the global RegExp ----
			{Code: `let RegExp; new RegExp('\\1(a)');`},
			{Code: `function foo() { var RegExp; RegExp('\\1(a)', 'u'); }`},
			{Code: `function foo(RegExp) { new RegExp('\\1(a)'); }`},
			{Code: `if (foo) { const RegExp = bar; RegExp('\\1(a)'); }`},
			{Code: `namespace RegExp {} RegExp('\\1(a)');`},
			// SKIP: rslint does not support ESLint's /*globals*/ directive comments
			// `/* globals RegExp:off */ new RegExp('\\1(a)');`
			// `RegExp('\\1(a)');` with languageOptions.globals { RegExp: "off" }

			// ---- no capturing groups ----
			{Code: `/(?:)/`},
			{Code: `/(?:a)/`},
			{Code: `new RegExp('')`},
			{Code: `RegExp('(?:a)|(?:b)*')`},
			{Code: `/^ab|[cd].\n$/`},

			// ---- no backreferences ----
			{Code: `/(a)/`},
			{Code: `RegExp('(a)|(b)')`},
			{Code: `new RegExp('\\n\\d(a)')`},
			{Code: `/\0(a)/`},
			{Code: `/\0(a)/u`},
			{Code: `/(?<=(a))(b)(?=(c))/`},
			{Code: `/(?<!(a))(b)(?!(c))/`},
			{Code: `/(?<foo>a)/`},

			// ---- not really a backreference ----
			// SKIP: rslint fixtures use strict TS, which rejects '\1' octal escapes in string literals
			// `RegExp('\1(a)')` // string octal escape
			{Code: `RegExp('\\\\1(a)')`}, // escaped backslash → JS string "\\1(a)" → pattern "\\1(a)" (literal \, then 1(a))
			{Code: `/\\1(a)/`},           // escaped backslash in literal
			{Code: `/\1/`},               // group 1 doesn't exist, regex octal escape
			{Code: `/^\1$/`},
			{Code: `/\2(a)/`},
			{Code: `/\1(?:a)/`},
			{Code: `/\1(?=a)/`},
			{Code: `/\1(?!a)/`},
			{Code: `/^[\1](a)$/`}, // \N in a character class is octal
			{Code: `new RegExp('[\\1](a)')`},
			{Code: `/\11(a)/`},     // octal escape \11
			{Code: `/\k<foo>(a)/`}, // no named groups → literal "k<foo>a"
			{Code: `/^(a)\1\2$/`},  // \1 backref, \2 octal

			// ---- valid backreferences: correct position, after the group ----
			{Code: `/(a)\1/`},
			{Code: `/(a).\1/`},
			{Code: `RegExp('(a)\\1(b)')`},
			{Code: `/(a)(b)\2(c)/`},
			{Code: `/(?<foo>a)\k<foo>/`},
			{Code: `new RegExp('(.)\\1')`},
			{Code: `RegExp('(a)\\1(?:b)')`},
			{Code: `/(a)b\1/`},
			{Code: `/((a)\2)/`},
			{Code: `/((a)b\2c)/`},
			{Code: `/^(?:(a)\1)$/`},
			{Code: `/^((a)\2)$/`},
			{Code: `/^(((a)\3))|b$/`},
			{Code: `/a(?<foo>(.)b\2)/`},
			{Code: `/(a)?(b)*(\1)(c)/`},
			{Code: `/(a)?(b)*(\2)(c)/`},
			{Code: `/(?<=(a))b\1/`},
			{Code: `/(?<=(?=(a)\1))b/`},

			// ---- valid backreferences: correct position before the group when both in same lookbehind ----
			{Code: `/(?<!\1(a))b/`},
			{Code: `/(?<=\1(a))b/`},
			{Code: `/(?<!\1.(a))b/`},
			{Code: `/(?<=\1.(a))b/`},
			{Code: `/(?<=(?:\1.(a)))b/`},
			{Code: `/(?<!(?:\1)((a)))b/`},
			{Code: `/(?<!(?:\2)((a)))b/`},
			{Code: `/(?=(?<=\1(a)))b/`},
			{Code: `/(?=(?<!\1(a)))b/`},
			{Code: `/(.)(?<=\2(a))b/`},

			// ---- valid: not a reference into another alternative ----
			{Code: `/^(a)\1|b/`},
			{Code: `/^a|(b)\1/`},
			{Code: `/^a|(b|c)\1/`},
			{Code: `/^(a)|(b)\2/`},
			{Code: `/^(?:(a)|(b)\2)$/`},
			{Code: `/^a|(?:.|(b)\1)/`},
			{Code: `/^a|(?:.|(b).(\1))/`},
			{Code: `/^a|(?:.|(?:(b)).(\1))/`},
			{Code: `/^a|(?:.|(?:(b)|c).(\1))/`},
			{Code: `/^a|(?:.|(?:(b)).(\1|c))/`},
			{Code: `/^a|(?:.|(?:(b)|c).(\1|d))/`},

			// ---- valid: not a reference into a negative lookaround (reference from within is allowed) ----
			{Code: `/.(?=(b))\1/`},
			{Code: `/.(?<=(b))\1/`},
			{Code: `/a(?!(b)\1)./`},
			{Code: `/a(?<!\1(b))./`},
			{Code: `/a(?!(b)(\1))./`},
			{Code: `/a(?!(?:(b)\1))./`},
			{Code: `/a(?!(?:(b))\1)./`},
			{Code: `/a(?<!(?:\1)(b))./`},
			{Code: `/a(?<!(?:(?:\1)(b)))./`},
			{Code: `/(?<!(a))(b)(?!(c))\2/`},
			{Code: `/a(?!(b|c)\1)./`},

			// ---- ignore regular expressions with syntax errors ----
			{Code: `RegExp('\\1(a)[')`},                              // unterminated [
			{Code: `new RegExp('\\1(a){', 'u')`},                     // unterminated { in u mode
			{Code: `new RegExp('\\1(a)\\2', 'ug')`},                  // \2 syntax error in u mode
			{Code: `const flags = 'gus'; RegExp('\\1(a){', flags);`}, // u flag known via const
			{Code: `RegExp('\\1(a)\\k<foo>', 'u')`},                  // \k<foo> with no named group is syntax error in u
			{Code: `new RegExp('\\k<foo>(?<foo>a)\\k<bar>')`},        // \k<bar> for unknown group is syntax error

			// ---- ES2024 ----
			{Code: `new RegExp('([[A--B]])\\1', 'v')`},
			{Code: `new RegExp('[[]\\1](a)', 'v')`}, // SyntaxError

			// ---- ES2025 ----
			{Code: `/((?<foo>bar)\k<foo>|(?<foo>baz))/`},

			// ---- Extra edge cases (rslint-side hardening) ----
			// Deeply nested capturing groups with valid backref
			{Code: `/(((((((((a)))))))))\1/`},
			// Multi-digit valid backref (group 10 referenced by \10)
			{Code: `/(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)\10/`},
			// Quantified group followed by valid backref
			{Code: `/(a)+\1/`},
			{Code: `/(a){2,5}\1/`},
			// Backref inside character class is treated as octal escape
			{Code: `/(a)[\1]/`},
			// Unicode-named group with \k<...>
			{Code: `/(?<日本>a)\k<日本>/u`},
			// Empty alternative + backref (edge: `(|a)\1` valid because \1 is after group)
			{Code: `/(|a)\1/`},

			// Config `off` un-declares `RegExp` — the constructor path goes dead
			{Code: `new RegExp("\\1(a)");`, Globals: map[string]bool{"RegExp": false}},
			{Code: `RegExp("\\1(a)");`, Globals: map[string]bool{"RegExp": false}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- ambient augmentations extend the global RegExp, they do not shadow it ----
			{
				Code: "export {};\ndeclare global {\n\tnamespace RegExp {}\n}\nnew RegExp('\\\\1(a)');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 5, Column: 1},
				},
			},
			{
				Code: "export {};\ndeclare global {\n\tvar RegExp: any;\n}\nnew RegExp('\\\\1(a)');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 5, Column: 1},
				},
			},

			// ---- full message tests ----
			{
				Code: `/(b)(\2a)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nested", Message: `Backreference '\2' will be ignored. It references group '(\2a)' from within that group.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/\k<foo>(?<foo>bar)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `RegExp('(a|bc)|\\1')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "disjunctive", Message: `Backreference '\1' will be ignored. It references group '(a|bc)' which is in another alternative.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('(?!(?<foo>\\n))\\1')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "intoNegativeLookaround", Message: `Backreference '\1' will be ignored. It references group '(?<foo>\n)' which is in a negative lookaround.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/(?<!(a)\1)b/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "backward", Message: `Backreference '\1' will be ignored. It references group '(a)' which appears before in the same lookbehind.`, Line: 1, Column: 1},
				},
			},

			// ---- nested ----
			{Code: `new RegExp('(\\1)')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/^(a\1)$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/^((a)\1)$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `new RegExp('^(a\\1b)$')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `RegExp('^((\\1))$')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/((\2))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/a(?<foo>(.)b\1)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/a(?<foo>\k<foo>)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/^(\1)*$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/^(?:a)(?:((?:\1)))*$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(?!(\1))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/a|(b\1c)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(a|(\1))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(a|(\2))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(?:a|(\1))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(a)?(b)*(\3)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},
			{Code: `/(?<=(a\1))b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "nested", Line: 1, Column: 1}}},

			// ---- forward ----
			{Code: `/\1(a)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\1.(a)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:\1)(?:(a))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:\1)(?:((a)))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:\2)(?:((a)))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:\1)(?:((?:a)))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(\2)(a)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `RegExp('(a)\\2(b)')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:a)(b)\2(c)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\k<foo>(?<foo>a)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?:a(b)\2)(c)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `new RegExp('(a)(b)\\3(c)')`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\1(?<=(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\1(?<!(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?<=\1)(?<=(a))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?<!\1)(?<!(a))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?=\1(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/(?!\1(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},

			// ---- backward in the same lookbehind ----
			{Code: `/(?<=(a)\1)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<!.(a).\1.)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(.)(?<!(b|c)\2)d/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<=(?:(a)\1))b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<=(?:(a))\1)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<=(a)(?:\1))b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<!(?:(a))(?:\1))b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/(?<!(?:(a))(?:\1)|.)b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/.(?!(?<!(a)\1))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/.(?=(?<!(a)\1))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/.(?!(?<=(a)\1))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},
			{Code: `/.(?=(?<=(a)\1))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "backward", Line: 1, Column: 1}}},

			// ---- into another alternative ----
			{Code: `/(a)|\1b/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a)|\1b)$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a)|b(?:c|\1))$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:a|b(?:(c)|\1))$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a(?!b))|\1b)+$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(?:(a)(?!b))|\1b)+$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a(?=a))|\1b)+$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/^(?:(a)(?=a)|\1b)+$/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/.(?:a|(b)).|(?:(\1)|c)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/.(?!(a)|\1)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},
			{Code: `/.(?<=\1|(a))./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "disjunctive", Line: 1, Column: 1}}},

			// ---- into a negative lookaround ----
			{Code: `/a(?!(b)).\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/(?<!(a))b\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/(?<!(a))(?:\1)/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?<!a|(b)).\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?!(a)).(?!\1)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?<!(a)).(?<!\1)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?=(?!(a))\1)./`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},
			{Code: `/.(?<!\1(?!(a)))/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "intoNegativeLookaround", Line: 1, Column: 1}}},

			// ---- valid and invalid ----
			{Code: `/\1(a)(b)\2/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},
			{Code: `/\1(a)\1/`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "forward", Line: 1, Column: 1}}},

			// ---- multiple invalid ----
			{
				Code: `/\1(a)\2(b)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			{
				Code: `/\1.(?<=(a)\1)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
					{MessageId: "backward", Line: 1, Column: 1},
				},
			},
			{
				Code: `/(?!\1(a)).\1/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
					{MessageId: "intoNegativeLookaround", Line: 1, Column: 1},
				},
			},
			{
				Code: `/(a)\2(b)/; RegExp('(\\1)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
					{MessageId: "nested", Line: 1, Column: 13},
				},
			},

			// ---- non-evaluable flags assumed to lack 'u' ----
			{
				Code: `RegExp('\\1(a){', flags);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},

			// ---- statically known expressions ----
			{
				Code: `const r = RegExp, p = '\\1', s = '(a)'; new r(p + s);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 41},
				},
			},
			{
				Code: `const pattern = '\\1(a)'; RegExp((((pattern as string) satisfies unknown)!));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
				},
			},
			{
				Code: `const slash = "\u005c"; const pattern = slash + "1(a)"; RegExp(pattern);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
				},
			},
			{
				Code: "new RegExp(String.raw`\\1${\"(a)\"}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
				},
			},
			{
				Code: `const r = RegExp; r('\\1(a)'); { const r = (value: string) => value; r('\\1(a)'); } r('\\1(a)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
					{MessageId: "forward"},
				},
			},
			{
				Code: `declare const regexpFactory: RegExpConstructor; regexpFactory('\\1(a)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
				},
			},
			{
				Code: `
declare const condition: boolean;
declare function isRegExpConstructor(
	value: RegExpConstructor | ((pattern: string) => string),
): value is RegExpConstructor;
const r: RegExpConstructor | ((pattern: string) => string) =
	condition ? RegExp : value => value;
if (isRegExpConstructor(r)) {
	r('\\1(a)');
} else {
	r('\\1(a)');
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
				},
			},
			{
				Code: `
let r: RegExpConstructor | ((pattern: string) => string) = value => value;
r('\\1(a)');
r = RegExp;
r('\\1(a)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
				},
			},
			{
				Code: `const r = globalThis["RegExp"]; r('\\1(a)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
				},
			},
			{
				Code: `R\u0065gExp('\\1(a)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward"},
				},
			},
			{
				Code: `new RegExp(true ? '\\1(a)' : '(a)');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp(String('\\1(a)'));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			{
				Code: "const RawString = String; new RegExp(RawString.raw`\\1(a)`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 27},
				},
			},
			{
				Code: `RegExp(/\1(a)/)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			{
				Code: `RegExp(/\1(a)/, "")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			{
				Code: `const flags = getFlags(); RegExp(/\1(a)/, flags);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 27},
				},
			},
			// Config `off` un-declares `RegExp`: the constructor no longer owns
			// its regex-literal argument, so the literal is checked standalone
			// with its own flags and reported at its own position.
			{
				Code:    `new RegExp(/\1(a)/, "u");`,
				Globals: map[string]bool{"RegExp": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 12},
				},
			},

			// ---- ES2024 ----
			{
				Code: `new RegExp('\\1([[A--B]])', 'v')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\1' will be ignored. It references group '([[A--B]])' which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},

			// ---- ES2025 ----
			{
				Code: `/\k<foo>((?<foo>bar)|(?<foo>baz))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and another group which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|\k<foo>(?<foo>baz))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>baz)' which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/\k<foo>((?<foo>bar)|(?<foo>baz)|(?<foo>qux))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and other 2 groups which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|\k<foo>(?<foo>baz)|(?<foo>qux))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>baz)' which appears later in the pattern.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|\k<foo>|(?<foo>baz))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "disjunctive", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and another group which is in another alternative.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|\k<foo>|(?<foo>baz)|(?<foo>qux))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "disjunctive", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and other 2 groups which is in another alternative.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?<foo>bar)|(?<foo>baz\k<foo>)|(?<foo>qux\k<foo>))/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nested", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>baz\k<foo>)' from within that group.`, Line: 1, Column: 1},
					{MessageId: "nested", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>qux\k<foo>)' from within that group.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/(?<=((?<foo>bar)|(?<foo>baz))\k<foo>)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "backward", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and another group which appears before in the same lookbehind.`, Line: 1, Column: 1},
				},
			},
			{
				Code: `/((?!(?<foo>bar))|(?!(?<foo>baz)))\k<foo>/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "intoNegativeLookaround", Message: `Backreference '\k<foo>' will be ignored. It references group '(?<foo>bar)' and another group which is in a negative lookaround.`, Line: 1, Column: 1},
				},
			},

			// ---- Extra edge cases (rslint-side hardening) ----
			// Multi-digit forward reference (group 10 referenced before declared)
			{
				Code: `/\10(a)(b)(c)(d)(e)(f)(g)(h)(i)(j)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			// Deeply nested forward reference
			{
				Code: `/((((((\1))))))(a)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nested", Line: 1, Column: 1},
				},
			},
			// Quantified backref before its group
			{
				Code: `/\1+(a)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			// Unicode-named group forward reference
			{
				Code: `/\k<日本>(?<日本>a)/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "forward", Line: 1, Column: 1},
				},
			},
			// Nested lookbehind — inner lookbehind still triggers backward
			{
				Code: `/(?<=(?<=(a)\1))b/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "backward", Line: 1, Column: 1},
				},
			},
		},
	)
}

func TestImportedRegExpConstructorAlias(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "file.ts")
	helperPath := tspath.ResolvePath(rootDir, "foo.ts")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		filePath:   `import { regexpFactory } from "./foo"; regexpFactory("\\1(a)");`,
		helperPath: `export const regexpFactory = RegExp;`,
	})
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	sourceFile := program.GetSourceFile("file.ts")
	if sourceFile == nil {
		t.Fatal("file.ts was not loaded")
		return
	}

	var diagnostics []rule.RuleDiagnostic
	linter.RunLinterInProgram(
		program,
		[]string{sourceFile.FileName()},
		nil,
		nil,
		func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     NoUselessBackreferenceRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUselessBackreferenceRule.Run(ctx, nil)
				},
			}}
		},
		false,
		func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
		nil,
		nil,
	)
	if len(diagnostics) != 1 || diagnostics[0].Message.Id != "forward" {
		t.Fatalf("diagnostics = %#v; want one forward report", diagnostics)
	}
}

func TestRegExpResolutionWithoutTypeInfo(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "file.ts")
	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		filePath: `
RegExp("\\1(a)");
{
	const RegExp = (pattern: string) => pattern;
	RegExp("\\1(a)");
}`,
	})
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatal("file.ts was not loaded")
	}

	var diagnostics []rule.RuleDiagnostic
	linter.RunLinterInProgram(
		program,
		[]string{sourceFile.FileName()},
		nil,
		nil,
		func(*ast.SourceFile) []linter.ConfiguredRule {
			return []linter.ConfiguredRule{{
				Name:     NoUselessBackreferenceRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return NoUselessBackreferenceRule.Run(ctx, nil)
				},
			}}
		},
		false,
		func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
		map[string]struct{}{}, // Keep the Program/RefStore but withhold the checker.
		nil,
	)
	if len(diagnostics) != 1 || diagnostics[0].Message.Id != "forward" {
		t.Fatalf("diagnostics = %#v; want one forward report", diagnostics)
	}
}

func TestMayEvaluateToRegexPatternAdversarial(t *testing.T) {
	rootDir := fixtures.GetRootDir()
	filePath := tspath.ResolvePath(rootDir, "file.ts")

	var code strings.Builder
	code.WriteString(`
declare function makePattern(): unknown;
const stablePattern = "\\1(a)";
let counter = 0;
enum Pattern {
	Value = "\\1(a)",
}
`)

	baseExpressions := []string{
		`"\\1(a)"`,
		"`\\\\1(a)`",
		"`\\\\1${\"(a)\"}`",
		"String.raw`\\1(a)`",
		`stablePattern`,
		`Pattern.Value`,
		`Pattern["Value"]`,
		`"\\1" + "(a)"`,
		`true ? "\\1(a)" : "(a)"`,
		`String("\\1(a)")`,
		`String()`,
		`"" || "\\1(a)"`,
		`null ?? "\\1(a)"`,
		`makePattern()`,
		`new String("\\1(a)")`,
		`typeof stablePattern`,
	}
	wrappers := []string{
		`%s`,
		`(%s)`,
		`(%s as string)`,
		`(%s satisfies unknown)`,
		`(%s)!`,
		`true ? %s : "(a)"`,
		`String(%s)`,
	}
	candidateCount := 0
	frontier := baseExpressions
	for depth := range 2 {
		next := make([]string, 0, len(frontier)*len(wrappers))
		for expressionIndex, expression := range frontier {
			for wrapperIndex, wrapper := range wrappers {
				wrapped := fmt.Sprintf(wrapper, expression)
				fmt.Fprintf(
					&code,
					"const candidate_%d_%d_%d = %s;\n",
					depth,
					expressionIndex,
					wrapperIndex,
					wrapped,
				)
				next = append(next, wrapped)
				candidateCount++
			}
		}
		frontier = next
	}

	rejectedExpressions := []string{
		`1`,
		`1n`,
		`true`,
		`false`,
		`null`,
		`{ value: "\\1(a)" }`,
		`["\\1(a)"]`,
		`() => "\\1(a)"`,
		`function () { return "\\1(a)"; }`,
		`class {}`,
		`-1`,
		`counter++`,
		`delete ({ value: 1 }).value`,
		`void stablePattern`,
	}
	for index, expression := range rejectedExpressions {
		fmt.Fprintf(&code, "const rejected_%d = %s;\n", index, expression)
	}

	fs := utils.NewOverlayVFS(bundled.WrapFS(osvfs.FS()), map[string]string{
		filePath: code.String(),
	})
	host := utils.CreateCompilerHost(rootDir, fs)
	program, err := utils.CreateProgram(true, fs, rootDir, "tsconfig.json", host)
	if err != nil {
		t.Fatal(err)
	}
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		t.Fatalf("%s was not loaded", filePath)
		return
	}
	typeChecker, done := program.GetTypeChecker(t.Context())
	defer done()
	evaluator := utils.NewStaticStringEvaluatorWithSourceFile(typeChecker, sourceFile)

	seenCandidates := 0
	seenRejected := 0
	evaluatedStrings := 0
	evaluatedCandidateStrings := 0
	var nonEvaluatedCandidates []string
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			declaration := node.AsVariableDeclaration()
			if declaration != nil &&
				declaration.Name() != nil &&
				declaration.Name().Kind == ast.KindIdentifier &&
				declaration.Initializer != nil {
				name := declaration.Name().AsIdentifier().Text
				mayEvaluate := mayEvaluateToRegexPattern(declaration.Initializer)
				if strings.HasPrefix(name, "candidate_") {
					seenCandidates++
					if !mayEvaluate {
						t.Errorf("%s: candidate expression was rejected", name)
					}
				}
				if strings.HasPrefix(name, "rejected_") {
					seenRejected++
					if mayEvaluate {
						t.Errorf("%s: known non-string expression passed the filter", name)
					}
				}
				if _, ok := evaluator.Eval(declaration.Initializer); ok {
					evaluatedStrings++
					if strings.HasPrefix(name, "candidate_") {
						evaluatedCandidateStrings++
					}
					if !mayEvaluate {
						t.Errorf("%s: StaticStringEvaluator produced a string rejected by the filter", name)
					}
				} else if strings.HasPrefix(name, "candidate_") {
					nonEvaluatedCandidates = append(nonEvaluatedCandidates, name)
				}
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)

	if seenCandidates != candidateCount {
		t.Fatalf("visited %d candidate expressions, want %d", seenCandidates, candidateCount)
	}
	if seenRejected != len(rejectedExpressions) {
		t.Fatalf("visited %d rejected expressions, want %d", seenRejected, len(rejectedExpressions))
	}
	const evaluableBaseExpressionCount = 11
	minimumEvaluatedCandidates := evaluableBaseExpressionCount *
		(len(wrappers) + len(wrappers)*len(wrappers))
	if evaluatedCandidateStrings < minimumEvaluatedCandidates {
		t.Fatalf(
			"only evaluated %d of %d adversarial candidates to strings; all strings=%d, misses: %v",
			evaluatedCandidateStrings,
			candidateCount,
			evaluatedStrings,
			nonEvaluatedCandidates,
		)
	}
}

func TestMayContainBackreference(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		{name: "plain", pattern: `abc`, want: false},
		{name: "non backreference escapes", pattern: `\d+\s+\w+`, want: false},
		{name: "numeric", pattern: `\1(a)`, want: true},
		{name: "named", pattern: `\k<name>(?<name>a)`, want: true},
		{name: "escaped numeric", pattern: `\\1(a)`, want: false},
		{name: "escaped named", pattern: `\\k<name>`, want: false},
		{name: "backreference after escaped slash", pattern: `\\\1(a)`, want: true},
		{name: "zero escape", pattern: `\0(a)`, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mayContainBackreference(test.pattern); got != test.want {
				t.Fatalf("mayContainBackreference(%q) = %v, want %v", test.pattern, got, test.want)
			}
		})
	}
}

func TestMayContainBackreferenceDoesNotMissParsedBackrefs(t *testing.T) {
	parts := []string{
		`a`,
		`(a)`,
		`(?:a)`,
		`(?<name>a)`,
		`\1`,
		`\2`,
		`\k<name>`,
		`\\`,
		`\\\1`,
		`[\1]`,
		`[\\1]`,
		`|`,
	}
	parsedBackrefs := 0
	for _, first := range parts {
		for _, second := range parts {
			for _, third := range parts {
				pattern := first + second + third
				_, backrefs, ok := parsePattern(pattern, utils.RegexFlags{})
				if !ok || len(backrefs) == 0 {
					continue
				}
				parsedBackrefs += len(backrefs)
				if !mayContainBackreference(pattern) {
					t.Fatalf("mayContainBackreference(%q) missed %d parsed backreferences", pattern, len(backrefs))
				}
			}
		}
	}
	if parsedBackrefs < 100 {
		t.Fatalf("only exercised %d parsed backreferences; want an adversarial corpus", parsedBackrefs)
	}
}

func TestAnalyzeBackrefMatchesLegacyDecision(t *testing.T) {
	patterns := []string{
		`\1(a)`,
		`(a)\1`,
		`(\1)`,
		`(a)|\1`,
		`(?!(a)).\1`,
		`(?<!(a)).\1`,
		`(?<=(a)\1)`,
		`(?<=\1(a))`,
		`\k<name>(?<name>a)`,
		`(?<name>a)|\k<name>`,
		`\k<name>((?<name>a)|(?<name>b)|(?<name>c))`,
		`((?<name>a)|\k<name>(?<name>b)|(?<name>c))`,
	}
	wrappers := []func(string) string{
		func(pattern string) string { return `(?:` + pattern + `)` },
		func(pattern string) string { return `(` + pattern + `)` },
		func(pattern string) string { return `(?=` + pattern + `)` },
		func(pattern string) string { return `(?!` + pattern + `)` },
		func(pattern string) string { return `(?<=` + pattern + `)` },
		func(pattern string) string { return `(?<!` + pattern + `)` },
		func(pattern string) string { return pattern + `|x` },
		func(pattern string) string { return `x|` + pattern },
	}
	frontier := append([]string(nil), patterns...)
	for range 2 {
		next := make([]string, 0, len(frontier)*len(wrappers))
		for _, pattern := range frontier {
			for _, wrap := range wrappers {
				next = append(next, wrap(pattern))
			}
		}
		patterns = append(patterns, next...)
		frontier = next
	}

	compared := 0
	for _, pattern := range patterns {
		_, backrefs, ok := parsePattern(pattern, utils.RegexFlags{})
		if !ok {
			continue
		}
		for index, backref := range backrefs {
			gotFirst, gotCount, gotOK := analyzeBackref(backref)
			wantFirst, wantCount, wantOK := legacyAnalyzeBackref(backref)
			compared++
			if gotOK != wantOK ||
				gotCount != wantCount ||
				gotFirst.messageId != wantFirst.messageId ||
				gotFirst.group != wantFirst.group {
				t.Fatalf(
					"pattern %q backref %d: analyzeBackref() = (%q, %d, %v), legacy = (%q, %d, %v)",
					pattern,
					index,
					gotFirst.messageId,
					gotCount,
					gotOK,
					wantFirst.messageId,
					wantCount,
					wantOK,
				)
			}
		}
	}
	if compared < 500 {
		t.Fatalf("only compared %d backreferences; want an adversarial corpus", compared)
	}
}

func legacyAnalyzeBackref(bref *rxNode) (first problem, problemCount int, hasProblem bool) {
	groups := bref.resolved
	if len(groups) == 0 {
		return problem{}, 0, false
	}
	brefPath := legacyPathToRoot(bref)
	problems := make([]*problem, 0, len(groups))
	for _, group := range groups {
		problems = append(problems, legacyClassifyPair(brefPath, bref, group))
	}
	for _, current := range problems {
		if current == nil {
			return problem{}, 0, false
		}
	}

	sameDisjunction := make([]*problem, 0, len(problems))
	for _, current := range problems {
		if current.messageId != "disjunctive" {
			sameDisjunction = append(sameDisjunction, current)
		}
	}
	toReport := problems
	if len(sameDisjunction) > 0 {
		toReport = sameDisjunction
	}
	if len(toReport) == 0 {
		panic("legacy analyzer produced no result")
	}
	return *toReport[0], len(toReport), true
}

func legacyClassifyPair(brefPath []*rxNode, bref *rxNode, group *rxNode) *problem {
	if legacyNodeContains(brefPath, group) {
		return &problem{messageId: "nested", group: group}
	}

	groupPath := legacyPathToRoot(group)
	i := len(brefPath) - 1
	j := len(groupPath) - 1
	for i >= 0 && j >= 0 && brefPath[i] == groupPath[j] {
		i--
		j--
	}
	indexOfLCA := j + 1
	groupCut := groupPath[:indexOfLCA]
	commonPath := groupPath[indexOfLCA:]

	var lowestCommonLookaround *rxNode
	for _, current := range commonPath {
		if isLookaround(current) {
			lowestCommonLookaround = current
			break
		}
	}
	matchingBackward := lowestCommonLookaround != nil && isLookbehind(lowestCommonLookaround)

	if len(groupCut) > 0 && groupCut[len(groupCut)-1].kind == nkAlternative {
		return &problem{messageId: "disjunctive", group: group}
	}
	if !matchingBackward && bref.end <= group.start {
		return &problem{messageId: "forward", group: group}
	}
	if matchingBackward && group.end <= bref.start {
		return &problem{messageId: "backward", group: group}
	}
	for _, current := range groupCut {
		if isNegativeLookaround(current) {
			return &problem{messageId: "intoNegativeLookaround", group: group}
		}
	}
	return nil
}

func legacyPathToRoot(node *rxNode) []*rxNode {
	var path []*rxNode
	for current := node; current != nil; current = current.parent {
		path = append(path, current)
	}
	return path
}

func legacyNodeContains(path []*rxNode, target *rxNode) bool {
	for _, current := range path {
		if current == target {
			return true
		}
	}
	return false
}
