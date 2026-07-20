// TestPreferDestructuringExtrasPrecedence verifies the complete range of
// receiver expression families that can be exposed by an object-access fix.
package prefer_destructuring

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferDestructuringExtrasPrecedence(t *testing.T) {
	// fixedReceiver is the exact receiver spelling ESLint emits after removing
	// only source parentheses that are unnecessary on an initializer RHS.
	// Sequence expressions are the notable case that must regain parentheses.
	receivers := []struct {
		receiver      string
		fixedReceiver string
	}{
		{receiver: "object", fixedReceiver: "object"},
		{receiver: "((object))", fixedReceiver: "object"},
		{receiver: "(a, b)", fixedReceiver: "(a, b)"},
		{receiver: "(a = b)", fixedReceiver: "a = b"},
		{receiver: "(a ||= b)", fixedReceiver: "a ||= b"},
		{receiver: "(condition ? a : b)", fixedReceiver: "condition ? a : b"},
		{receiver: "(a ?? b)", fixedReceiver: "a ?? b"},
		{receiver: "(a || b)", fixedReceiver: "a || b"},
		{receiver: "(a && b)", fixedReceiver: "a && b"},
		{receiver: "(a | b)", fixedReceiver: "a | b"},
		{receiver: "(a ^ b)", fixedReceiver: "a ^ b"},
		{receiver: "(a & b)", fixedReceiver: "a & b"},
		{receiver: "(a === b)", fixedReceiver: "a === b"},
		{receiver: "(a in b)", fixedReceiver: "a in b"},
		{receiver: "(a << b)", fixedReceiver: "a << b"},
		{receiver: "(a + b)", fixedReceiver: "a + b"},
		{receiver: "(a * b)", fixedReceiver: "a * b"},
		{receiver: "(a ** b)", fixedReceiver: "a ** b"},
		{receiver: "(!a)", fixedReceiver: "!a"},
		{receiver: "(typeof a)", fixedReceiver: "typeof a"},
		{receiver: "(delete object.bar)", fixedReceiver: "delete object.bar"},
		{receiver: "(++a)", fixedReceiver: "++a"},
		{receiver: "(a++)", fixedReceiver: "a++"},
		{receiver: "getObject()", fixedReceiver: "getObject()"},
		{receiver: "new Box()", fixedReceiver: "new Box()"},
		{receiver: "tag`value`", fixedReceiver: "tag`value`"},
		{receiver: "import(\"package\")", fixedReceiver: "import(\"package\")"},
		{receiver: "import.meta", fixedReceiver: "import.meta"},
		{receiver: "[1, 2, 3]", fixedReceiver: "[1, 2, 3]"},
		{receiver: "({ foo: 1 })", fixedReceiver: "{ foo: 1 }"},
		{receiver: "(function () {})", fixedReceiver: "function () {}"},
		{receiver: "(function* () {})", fixedReceiver: "function* () {}"},
		{receiver: "(async function () {})", fixedReceiver: "async function () {}"},
		{receiver: "(class {})", fixedReceiver: "class {}"},
		{receiver: "(() => object)", fixedReceiver: "() => object"},
		{receiver: "(async () => object)", fixedReceiver: "async () => object"},
		{receiver: "\"text\"", fixedReceiver: "\"text\""},
		{receiver: "`text`", fixedReceiver: "`text`"},
		{receiver: "/value/u", fixedReceiver: "/value/u"},
		{receiver: "(123)", fixedReceiver: "123"},
		{receiver: "true", fixedReceiver: "true"},
		{receiver: "null", fixedReceiver: "null"},
		// ESTree does not assign a core precedence to TypeScript-only
		// expressions, so ESLint conservatively keeps parentheses around them.
		{
			receiver:      "(object as { foo: unknown })",
			fixedReceiver: "(object as { foo: unknown })",
		},
		{
			receiver:      "(<{ foo: unknown }>object)",
			fixedReceiver: "(<{ foo: unknown }>object)",
		},
		{
			receiver:      "(object satisfies { foo: unknown })",
			fixedReceiver: "(object satisfies { foo: unknown })",
		},
		{receiver: "object!", fixedReceiver: "(object!)"},
		{
			receiver:      "getObject<string>()",
			fixedReceiver: "getObject<string>()",
		},
	}

	invalid := make([]rule_tester.InvalidTestCase, 0, len(receivers)+2)
	for _, testCase := range receivers {
		code := "const foo = " + testCase.receiver + ".foo;"
		invalid = append(invalid, oneLineMatrixError(
			code,
			7,
			"object",
			"const {foo} = "+testCase.fixedReceiver+";",
			nil,
			false,
		))
	}

	// Await and yield are only legal in their respective enclosing contexts,
	// but otherwise follow the same parenthesis-removal rules.
	invalid = append(invalid,
		rule_tester.InvalidTestCase{
			Code:   "async function f() { const foo = (await getObject()).foo; }",
			Output: []string{"async function f() { const {foo} = await getObject(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 28, 1, 57),
			},
		},
		rule_tester.InvalidTestCase{
			Code:   "function* f() { const foo = (yield getObject()).foo; }",
			Output: []string{"function* f() { const {foo} = yield getObject(); }"},
			Errors: []rule_tester.InvalidTestCaseError{
				preferError("object", 1, 23, 1, 52),
			},
		},
	)

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferDestructuringRule,
		nil,
		invalid,
	)
}
