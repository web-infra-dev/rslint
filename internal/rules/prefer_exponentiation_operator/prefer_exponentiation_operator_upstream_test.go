package prefer_exponentiation_operator

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestPreferExponentiationOperatorUpstream migrates the full valid/invalid suite from upstream tests/lib/rules/prefer-exponentiation-operator.js 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific lock-in cases live in the prefer_exponentiation_operator_extras_test.go file.
func TestPreferExponentiationOperatorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferExponentiationOperatorRule,
		[]rule_tester.ValidTestCase{
			// ---- not Math.pow() ----
			{Code: "Object.pow(a, b)"},
			{Code: "Math.max(a, b)"},
			{Code: "Math"},
			{Code: "Math(a, b)"},
			{Code: "pow"},
			{Code: "pow(a, b)"},
			{Code: "Math.pow"},
			{Code: "Math.Pow(a, b)"},
			{Code: "math.pow(a, b)"},
			{Code: "foo.Math.pow(a, b)"},
			{Code: "new Math.pow(a, b)"},
			{Code: "Math[pow](a, b)"},
			{Code: "globalThis.Object.pow(a, b)"},
			{Code: "globalThis.Math.max(a, b)"},

			// ---- not the global Math ----
			{Code: "/* globals Math:off*/ Math.pow(a, b)", Skip: true}, // SKIP: rslint does not support ESLint's /* globals */ directive comments.
			{Code: "let Math; Math.pow(a, b);"},
			{Code: "if (foo) { const Math = 1; Math.pow(a, b); }"},
			{Code: "var x = function Math() { Math.pow(a, b); }"},
			{Code: "function foo(Math) { Math.pow(a, b); }"},
			{Code: "function foo() { Math.pow(a, b); var Math; }"},

			{Code: "globalThis.Math.pow(a, b)", Skip: true}, // SKIP: mirrors upstream ecmaVersion 2019 case; rslint does not expose ecmaVersion-specific globalThis availability.
			{Code: "globalThis.Math.pow(a, b)", Skip: true}, // SKIP: mirrors upstream ecmaVersion 6 case; rslint does not expose ecmaVersion-specific globalThis availability.
			{Code: "globalThis.Math.pow(a, b)", Skip: true}, // SKIP: mirrors upstream ecmaVersion 2017 case; rslint does not expose ecmaVersion-specific globalThis availability.
			{Code: `
                var globalThis = bar;
                globalThis.Math.pow(a, b)
            `},

			{Code: "class C { #pow; foo() { Math.#pow(a, b); } }"},
		},
		[]rule_tester.InvalidTestCase{
			invalidFixed("Math.pow(a, b)", "a**b"),
			invalidFixed("(Math).pow(a, b)", "a**b"),
			invalidFixed("Math['pow'](a, b)", "a**b"),
			invalidFixed("(Math)['pow'](a, b)", "a**b"),
			invalidFixed("var x=Math\n.  pow( a, \n b )", "var x=a**b"),
			invalidFixed("globalThis.Math.pow(a, b)", "a**b"),
			invalidFixed("globalThis.Math['pow'](a, b)", "a**b"),

			// ---- able to catch some workarounds ----
			invalidFixed("Math[`pow`](a, b)", "a**b"),
			invalidFixed("Math[`${'pow'}`](a, b)", "a**b"),
			invalidFixed("Math['p' + 'o' + 'w'](a, b)", "a**b"),

			// ---- non-expression parents that don't require parens ----
			invalidFixed("var x = Math.pow(a, b);", "var x = a**b;"),
			invalidFixed("if(Math.pow(a, b)){}", "if(a**b){}"),
			invalidFixed("for(;Math.pow(a, b);){}", "for(;a**b;){}"),
			invalidFixed("switch(foo){ case Math.pow(a, b): break; }", "switch(foo){ case a**b: break; }"),
			invalidFixed("{ foo: Math.pow(a, b) }", "{ foo: a**b }"),
			invalidFixed("function foo(bar, baz = Math.pow(a, b), quux){}", "function foo(bar, baz = a**b, quux){}"),
			invalidFixed("`${Math.pow(a, b)}`", "`${a**b}`"),

			// ---- non-expression parents that do require parens ----
			invalidFixed("class C extends Math.pow(a, b) {}", "class C extends (a**b) {}"),

			// ---- parents with a higher precedence ----
			invalidFixed("+ Math.pow(a, b)", "+ (a**b)"),
			invalidFixed("- Math.pow(a, b)", "- (a**b)"),
			invalidFixed("! Math.pow(a, b)", "! (a**b)"),
			invalidFixed("typeof Math.pow(a, b)", "typeof (a**b)"),
			invalidFixed("void Math.pow(a, b)", "void (a**b)"),
			invalidFixed("Math.pow(a, b) .toString()", "(a**b) .toString()"),
			invalidFixed("Math.pow(a, b) ()", "(a**b) ()"),
			invalidFixed("Math.pow(a, b) ``", "(a**b) ``"),
			invalidFixed("(class extends Math.pow(a, b) {})", "(class extends (a**b) {})"),

			// ---- already parenthesised, shouldn't insert extra parens ----
			invalidFixed("+(Math.pow(a, b))", "+(a**b)"),
			invalidFixed("(Math.pow(a, b)).toString()", "(a**b).toString()"),
			invalidFixed("(class extends (Math.pow(a, b)) {})", "(class extends (a**b) {})"),
			invalidFixed("class C extends (Math.pow(a, b)) {}", "class C extends (a**b) {}"),

			// ---- parents with a higher precedence, but the expression's role doesn't require parens ----
			invalidFixed("f(Math.pow(a, b))", "f(a**b)"),
			invalidFixed("f(foo, Math.pow(a, b))", "f(foo, a**b)"),
			invalidFixed("f(Math.pow(a, b), foo)", "f(a**b, foo)"),
			invalidFixed("f(foo, Math.pow(a, b), bar)", "f(foo, a**b, bar)"),
			invalidFixed("new F(Math.pow(a, b))", "new F(a**b)"),
			invalidFixed("new F(foo, Math.pow(a, b))", "new F(foo, a**b)"),
			invalidFixed("new F(Math.pow(a, b), foo)", "new F(a**b, foo)"),
			invalidFixed("new F(foo, Math.pow(a, b), bar)", "new F(foo, a**b, bar)"),
			invalidFixed("obj[Math.pow(a, b)]", "obj[a**b]"),
			invalidFixed("[foo, Math.pow(a, b), bar]", "[foo, a**b, bar]"),

			// ---- parents with a lower precedence ----
			invalidFixed("a * Math.pow(b, c)", "a * b**c"),
			invalidFixed("Math.pow(a, b) * c", "a**b * c"),
			invalidFixed("a + Math.pow(b, c)", "a + b**c"),
			invalidFixed("Math.pow(a, b)/c", "a**b/c"),
			invalidFixed("a < Math.pow(b, c)", "a < b**c"),
			invalidFixed("Math.pow(a, b) > c", "a**b > c"),
			invalidFixed("a === Math.pow(b, c)", "a === b**c"),
			invalidFixed("a ? Math.pow(b, c) : d", "a ? b**c : d"),
			invalidFixed("a = Math.pow(b, c)", "a = b**c"),
			invalidFixed("a += Math.pow(b, c)", "a += b**c"),
			invalidFixed("function *f() { yield Math.pow(a, b) }", "function *f() { yield a**b }"),
			invalidFixed("a, Math.pow(b, c), d", "a, b**c, d"),

			// ---- '**' is right-associative, that applies to both parent and child nodes ----
			invalidFixed("a ** Math.pow(b, c)", "a ** b**c"),
			invalidFixed("Math.pow(a, b) ** c", "(a**b) ** c"),
			invalidFixed("Math.pow(a, b ** c)", "a**b ** c"),
			invalidFixed("Math.pow(a ** b, c)", "(a ** b)**c"),
			invalidFixed("a ** Math.pow(b ** c, d ** e) ** f", "a ** ((b ** c)**d ** e) ** f"),

			// ---- doesn't remove already existing unnecessary parens around the whole expression ----
			invalidFixed("(Math.pow(a, b))", "(a**b)"),
			invalidFixed("foo + (Math.pow(a, b))", "foo + (a**b)"),
			invalidFixed("(Math.pow(a, b)) + foo", "(a**b) + foo"),
			invalidFixed("`${(Math.pow(a, b))}`", "`${(a**b)}`"),

			// ---- base and exponent with a higher precedence ----
			invalidFixed("Math.pow(2, 3)", "2**3"),
			invalidFixed("Math.pow(a.foo, b)", "a.foo**b"),
			invalidFixed("Math.pow(a, b.foo)", "a**b.foo"),
			invalidFixed("Math.pow(a(), b)", "a()**b"),
			invalidFixed("Math.pow(a, b())", "a**b()"),
			invalidFixed("Math.pow(++a, ++b)", "++a**++b"),
			invalidFixed("Math.pow(a++, ++b)", "a++**++b"),
			invalidFixed("Math.pow(a--, b--)", "a--**b--"),
			invalidFixed("Math.pow(--a, b--)", "--a**b--"),

			// ---- doesn't preserve unnecessary parens around base and exponent ----
			invalidFixed("Math.pow((a), (b))", "a**b"),
			invalidFixed("Math.pow(((a)), ((b)))", "a**b"),
			invalidFixed("Math.pow((a.foo), b)", "a.foo**b"),
			invalidFixed("Math.pow(a, (b.foo))", "a**b.foo"),
			invalidFixed("Math.pow((a()), b)", "a()**b"),
			invalidFixed("Math.pow(a, (b()))", "a**b()"),

			// ---- unary expressions are exception by the language ----
			invalidFixed("Math.pow(+a, b)", "(+a)**b"),
			invalidFixed("Math.pow(a, +b)", "a**+b"),
			invalidFixed("Math.pow(-a, b)", "(-a)**b"),
			invalidFixed("Math.pow(a, -b)", "a**-b"),
			invalidFixed("Math.pow(-2, 3)", "(-2)**3"),
			invalidFixed("Math.pow(2, -3)", "2**-3"),
			invalidFixed("async () => Math.pow(await a, b)", "async () => (await a)**b"),
			invalidFixed("async () => Math.pow(a, await b)", "async () => a**await b"),

			// ---- base and exponent with a lower precedence ----
			invalidFixed("Math.pow(a * b, c)", "(a * b)**c"),
			invalidFixed("Math.pow(a, b * c)", "a**(b * c)"),
			invalidFixed("Math.pow(a / b, c)", "(a / b)**c"),
			invalidFixed("Math.pow(a, b / c)", "a**(b / c)"),
			invalidFixed("Math.pow(a + b, 3)", "(a + b)**3"),
			invalidFixed("Math.pow(2, a - b)", "2**(a - b)"),
			invalidFixed("Math.pow(a + b, c + d)", "(a + b)**(c + d)"),
			invalidFixed("Math.pow(a = b, c = d)", "(a = b)**(c = d)"),
			invalidFixed("Math.pow(a += b, c -= d)", "(a += b)**(c -= d)"),
			invalidFixed("Math.pow((a, b), (c, d))", "(a, b)**(c, d)"),
			invalidFixed("function *f() { Math.pow(yield, yield) }", "function *f() { (yield)**(yield) }"),

			// ---- doesn't put extra parens ----
			invalidFixed("Math.pow((a + b), (c + d))", "(a + b)**(c + d)"),

			// ---- tokens that can be adjacent ----
			invalidFixed("a+Math.pow(b, c)+d", "a+b**c+d"),

			// ---- tokens that cannot be adjacent ----
			invalidFixed("a+Math.pow(++b, c)", "a+ ++b**c"),
			invalidFixed("(a)+(Math).pow((++b), c)", "(a)+ ++b**c"),
			invalidFixed("Math.pow(a, b)in c", "a**b in c"),
			invalidFixed("Math.pow(a, (b))in (c)", "a**b in (c)"),
			invalidFixed("a+Math.pow(++b, c)in d", "a+ ++b**c in d"),
			invalidFixed("a+Math.pow( ++b, c )in d", "a+ ++b**c in d"),

			// ---- tokens that cannot be adjacent, but there is already space or something else between ----
			invalidFixed("a+ Math.pow(++b, c) in d", "a+ ++b**c in d"),
			invalidFixed("a+/**/Math.pow(++b, c)/**/in d", "a+/**/++b**c/**/in d"),
			invalidFixed("a+(Math.pow(++b, c))in d", "a+(++b**c)in d"),

			// ---- tokens that cannot be adjacent, but parens required for precedence make extra space unnecessary ----
			invalidFixed("+Math.pow(++a, b)", "+(++a**b)"),
			invalidFixed("Math.pow(a, b + c)in d", "a**(b + c)in d"),

			invalidFixed("Math.pow(a, b) + Math.pow(c,\n d)", "a**b + c**d"),
			invalidWithOutputs("Math.pow(Math.pow(a, b), Math.pow(c, d))", "Math.pow(a, b)**Math.pow(c, d)", "(a**b)**c**d"),
			invalidWithOutputs("Math.pow(a, b)**Math.pow(c, d)", "(a**b)**c**d"),

			// ---- shouldn't autofix if the call doesn't have exactly two arguments ----
			invalidNoFix("Math.pow()"),
			invalidNoFix("Math.pow(a)"),
			invalidNoFix("Math.pow(a, b, c)"),
			invalidNoFix("Math.pow(a, b, c, d)"),

			// ---- shouldn't autofix if any of the arguments is spread ----
			invalidNoFix("Math.pow(...a)"),
			invalidNoFix("Math.pow(...a, b)"),
			invalidNoFix("Math.pow(a, ...b)"),
			invalidNoFix("Math.pow(a, b, ...c)"),

			// ---- shouldn't autofix if that would remove comments ----
			invalidFixed("/* comment */Math.pow(a, b)", "/* comment */a**b"),
			invalidNoFix("Math/**/.pow(a, b)"),
			invalidNoFix("Math//\n.pow(a, b)"),
			invalidNoFix("Math[//\n'pow'](a, b)"),
			invalidNoFix("Math['pow'/**/](a, b)"),
			invalidNoFix("Math./**/pow(a, b)"),
			invalidNoFix("Math.pow/**/(a, b)"),
			invalidNoFix("Math.pow//\n(a, b)"),
			invalidNoFix("Math.pow(/**/a, b)"),
			invalidNoFix("Math.pow(a,//\n b)"),
			invalidNoFix("Math.pow(a, b/**/)"),
			invalidNoFix("Math.pow(a, b//\n)"),
			invalidFixed("Math.pow(a, b)/* comment */;", "a**b/* comment */;"),
			invalidFixed("Math.pow(a, b)// comment\n;", "a**b// comment\n;"),

			// ---- Optional chaining ----
			invalidFixed("Math.pow?.(a, b)", "a**b"),
			invalidFixed("Math?.pow(a, b)", "a**b"),
			invalidFixed("Math?.pow?.(a, b)", "a**b"),
			invalidFixed("(Math?.pow)(a, b)", "a**b"),
			invalidFixed("(Math?.pow)?.(a, b)", "a**b"),

			// ---- https://github.com/eslint/eslint/issues/17173 ----
			invalidFixed("Math.pow(a, b as any)", "a**(b as any)"),
			invalidFixed("Math.pow(a as any, b)", "(a as any)**b"),
			invalidFixed("Math.pow(a, b) as any", "(a**b) as any"),

			// ---- https://github.com/eslint/eslint/issues/20987 ----
			invalidFixed("Math.pow({a:1}.a, 2);", "({a:1}.a**2);"),
			invalidFixed("Math.pow({a:1}.a, 2) + 100;", "({a:1}.a**2) + 100;"),
			invalidFixed("(Math.pow({a:1}.a, 2));", "({a:1}.a**2);"),
			invalidFixed("100 + Math.pow({a:1}.a, 2);", "100 + {a:1}.a**2;"),
			invalidFixed("Math.pow({a:1}.a + 100, 2);", "({a:1}.a + 100)**2;"),
			invalidFixed("Math.pow(function(){return 2}(), 3);", "(function(){return 2}()**3);"),
			invalidFixed("Math.pow(function(){return 2}(), 3) + 100;", "(function(){return 2}()**3) + 100;"),
			invalidFixed("(Math.pow(function(){return 2}(), 3));", "(function(){return 2}()**3);"),
			invalidFixed("100 + Math.pow(function(){return 2}(), 3);", "100 + function(){return 2}()**3;"),
			invalidFixed("Math.pow(function(){return 2}() + 100, 3);", "(function(){return 2}() + 100)**3;"),
			invalidFixed("Math.pow(class{static x=2}.x, 4);", "(class{static x=2}.x**4);"),
			invalidFixed("Math.pow(class{static x=2}.x, 4) + 100;", "(class{static x=2}.x**4) + 100;"),
			invalidFixed("(Math.pow(class{static x=2}.x, 4));", "(class{static x=2}.x**4);"),
			invalidFixed("100 + Math.pow(class{static x=2}.x, 4);", "100 + class{static x=2}.x**4;"),
			invalidFixed("Math.pow(class{static x=2}.x + 100, 4);", "(class{static x=2}.x + 100)**4;"),

			// ---- preceding semicolon needed ----
			invalidFixed("foo\nMath.pow(a + b, c)", "foo\n;(a + b)**c"),
			invalidFixed("foo\nMath.pow(+a, b)", "foo\n;(+a)**b"),
			invalidFixed("foo\nMath.pow(-a, b)", "foo\n;(-a)**b"),
			invalidFixed("foo\nMath.pow({a:1}.a, 2)", "foo\n;({a:1}.a**2)"),
			invalidFixed("foo\nMath.pow((a).b, c)", "foo\n;(a).b**c"),
			invalidFixed("foo\nMath.pow([a, b].find(fn), c)", "foo\n;[a, b].find(fn)**c"),
			invalidFixed("foo\nMath.pow(/regex/, 2)", "foo\n;/regex/**2"),
			invalidFixed("foo\nMath.pow(`template literal`, 2)", "foo\n;`template literal`**2"),

			// ---- preceding semicolon not needed ----
			invalidFixed("foo\n100 + Math.pow((a).b, c)", "foo\n100 + (a).b**c"),
			invalidFixed("foo\nMath.pow(a.b, c)", "foo\na.b**c"),
			invalidFixed("Math.pow((a).b, c)", "(a).b**c"),
			invalidFixed("foo;\nMath.pow((a).b, c)", "foo;\n(a).b**c"),
			invalidFixed("if (foo) {}\nMath.pow((a).b, c)", "if (foo) {}\n(a).b**c"),
		},
	)
}

func invalidFixed(code string, output string) rule_tester.InvalidTestCase {
	return invalidWithOutputs(code, output)
}

func invalidNoFix(code string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:   code,
		Errors: expectedFirstCallError(code),
	}
}

func invalidWithOutputs(code string, outputs ...string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code:   code,
		Output: outputs,
		Errors: expectedErrors(code),
	}
}

func expectedErrors(code string) []rule_tester.InvalidTestCaseError {
	opts := ast.SourceFileParseOptions{
		FileName: "/tmp/prefer_exponentiation_operator.ts",
		Path:     tspath.Path("/tmp/prefer_exponentiation_operator.ts"),
	}
	sf := parser.ParseSourceFile(opts, code, core.ScriptKindTS)
	var errors []rule_tester.InvalidTestCaseError
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindCallExpression && isMathPowCall(node, nil) {
			rng := utils.TrimNodeTextRange(sf, node)
			line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, rng.Pos())
			endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, rng.End())
			errors = append(errors, rule_tester.InvalidTestCaseError{
				MessageId: "useExponentiation",
				Message:   messageUseExponentiation,
				Line:      line + 1,
				Column:    int(column) + 1,
				EndLine:   endLine + 1,
				EndColumn: int(endColumn) + 1,
			})
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sf.AsNode())
	return errors
}

func expectedFirstCallError(code string) []rule_tester.InvalidTestCaseError {
	opts := ast.SourceFileParseOptions{
		FileName: "/tmp/prefer_exponentiation_operator.ts",
		Path:     tspath.Path("/tmp/prefer_exponentiation_operator.ts"),
	}
	sf := parser.ParseSourceFile(opts, code, core.ScriptKindTS)

	var callNode *ast.Node
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || callNode != nil {
			return
		}
		if node.Kind == ast.KindCallExpression {
			callNode = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return callNode != nil
		})
	}
	walk(sf.AsNode())
	if callNode == nil {
		return nil
	}

	rng := utils.TrimNodeTextRange(sf, callNode)
	line, column := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, rng.Pos())
	endLine, endColumn := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, rng.End())
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "useExponentiation",
		Message:   messageUseExponentiation,
		Line:      line + 1,
		Column:    int(column) + 1,
		EndLine:   endLine + 1,
		EndColumn: int(endColumn) + 1,
	}}
}
