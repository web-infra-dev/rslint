package no_restricted_syntax

import (
	"strings"
	"testing"
)

func TestParseSelector_Wellformed(t *testing.T) {
	cases := []string{
		"Identifier",
		"123",
		"#identifier",
		"*",
		"FunctionExpression",
		"BinaryExpression",
		`Identifier[name="bar"]`,
		"BreakStatement[label]",
		`VariableDeclaration[kind='using']`,
		`VariableDeclaration[kind='await using']`,
		"FunctionDeclaration[params.length>2]",
		"FunctionDeclaration[params.length>=2]",
		"FunctionDeclaration[params.length<=2]",
		"FunctionDeclaration[params.length<2]",
		"FunctionDeclaration[params.0.name=x]",
		"ImportDeclaration[source.value=some/path]",
		"Literal[value=-1]",
		"BinaryExpression[name=&&]",
		"Identifier[name=type(string)]",
		"Literal[value=.5]",
		"Literal[regex.flags=/./]",
		"Literal[regex.flags=/i/]",
		`ImportDeclaration[source.value=/^some\/path$/]`,
		"ArrowFunctionExpression > BlockStatement",
		"Property > Literal.key",
		".body.declarations.init",
		"FunctionDeclaration FunctionExpression",
		"Literal + Literal",
		"* ~ *",
		"!IfStatement > BlockStatement",
		"ExpressionStatement + !ExpressionStatement",
		":is(Identifier, Literal)",
		":matches(Identifier, Literal)",
		":not(VariableDeclaration)",
		":has(Literal)",
		":has(> Identifier)",
		":has(+ Identifier)",
		":has(~ Identifier)",
		":Expression",
		":nth-child(1)",
		":nth-last-child(2)",
		":first-child",
		":last-child",
		"ChainExpression",
		"[optional=true]",
		"MemberExpression[computed=true]",
		"MethodDefinition:not([static=true])",
		":is(Identifier[name='foo'], Identifier[name='bar'])",
		"FunctionDeclaration[generator=true]",
		"FunctionDeclaration[async=true]",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, err := parseSelector(c); err != nil {
				t.Fatalf("parseSelector(%q) returned %v", c, err)
			}
		})
	}
}

func TestParseSelector_Malformed(t *testing.T) {
	cases := []string{
		"",                                   // empty
		"   ",                                // whitespace only
		"[",                                  // unterminated bracket
		"BinaryExpression[",                  // unterminated bracket after head
		"BinaryExpression[name",              // unterminated attribute
		"BinaryExpression[name=",             // missing value
		"BinaryExpression[name=']",           // unterminated string
		"BinaryExpression[name=/]",           // unterminated regex (no closing /)
		"Identifier[name=//]",                // empty regex
		"Identifier[name=/foo/g]",            // unsupported regex flag
		"Identifier[name=/foo/ii]",           // duplicate regex flag
		"Identifier[name=type()]",            // empty typeof operand
		"Literal[value=5.]",                  // esquery requires a fractional digit
		":nth-child",                         // missing arg
		":nth-child(",                        // unterminated paren
		":nth-child(abc)",                    // non-numeric arg
		":is(",                               // unterminated paren
		":is(,)",                             // empty selector in args
		"Identifier,",                        // trailing comma
		"Identifier >",                       // dangling combinator
		"Identifier >> Foo",                  // double combinator
		":unknownPseudo",                     // unsupported pseudo
		":HAS(Identifier)",                   // named pseudos are case-sensitive
		"FunctionDeclaration[params.length>", // missing operand
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, err := parseSelector(c); err == nil {
				t.Fatalf("parseSelector(%q) should have failed", c)
			}
		})
	}
}

func TestParseRuleOptions_Shapes(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		if got := parseRuleOptions(nil); len(got) != 0 {
			t.Fatalf("nil should yield 0 entries, got %d", len(got))
		}
	})
	t.Run("empty array", func(t *testing.T) {
		if got := parseRuleOptions([]interface{}{}); len(got) != 0 {
			t.Fatalf("empty array should yield 0 entries, got %d", len(got))
		}
	})
	t.Run("single string", func(t *testing.T) {
		got := parseRuleOptions([]interface{}{"Identifier"})
		if len(got) != 1 || got[0].selector != "Identifier" {
			t.Fatalf("unexpected entries: %#v", got)
		}
	})
	t.Run("exit selector", func(t *testing.T) {
		got := parseRuleOptions([]interface{}{"Identifier:exit"})
		if len(got) != 1 || got[0].selector != "Identifier:exit" || got[0].compiled == nil {
			t.Fatalf("unexpected entries: %#v", got)
		}
	})
	t.Run("single map", func(t *testing.T) {
		got := parseRuleOptions([]interface{}{map[string]interface{}{
			"selector": "Identifier",
			"message":  "no identifiers",
		}})
		if len(got) != 1 || got[0].selector != "Identifier" || got[0].message != "no identifiers" {
			t.Fatalf("unexpected entries: %#v", got)
		}
	})
	t.Run("multi-element array of mixed forms", func(t *testing.T) {
		got := parseRuleOptions([]interface{}{
			"WithStatement",
			map[string]interface{}{"selector": "VariableDeclaration", "message": "x"},
		})
		if len(got) != 2 {
			t.Fatalf("expected 2, got %d", len(got))
		}
		if got[0].selector != "WithStatement" || got[0].message != "Using 'WithStatement' is not allowed." {
			t.Fatalf("entry 0 mismatch: %#v", got[0])
		}
		if got[1].selector != "VariableDeclaration" || got[1].message != "x" {
			t.Fatalf("entry 1 mismatch: %#v", got[1])
		}
	})
	t.Run("malformed selectors are silently dropped, well-formed kept", func(t *testing.T) {
		got := parseRuleOptions([]interface{}{
			"Identifier",
			"[", // unterminated
			"FunctionExpression",
		})
		if len(got) != 2 {
			t.Fatalf("expected 2 entries, got %d (%#v)", len(got), got)
		}
		names := []string{got[0].selector, got[1].selector}
		if !strings.Contains(strings.Join(names, ","), "Identifier") || !strings.Contains(strings.Join(names, ","), "FunctionExpression") {
			t.Fatalf("missing well-formed entry; got %v", names)
		}
	})
	t.Run("object missing selector key is dropped", func(t *testing.T) {
		got := parseRuleOptions([]interface{}{
			map[string]interface{}{"message": "no selector"},
		})
		if len(got) != 0 {
			t.Fatalf("expected 0 entries, got %d", len(got))
		}
	})
	t.Run("non-string non-map array elements are ignored", func(t *testing.T) {
		got := parseRuleOptions([]interface{}{
			"Identifier",
			42,
			true,
			nil,
		})
		if len(got) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(got))
		}
	})
}

func TestParseRuleOptions_DefaultMessage(t *testing.T) {
	got := parseRuleOptions([]interface{}{"Identifier"})
	if len(got) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if msg := got[0].formatMessage(); msg != "Using 'Identifier' is not allowed." {
		t.Fatalf("default message mismatch: %q", msg)
	}
}

func TestParseRuleOptions_CustomMessage(t *testing.T) {
	got := parseRuleOptions([]interface{}{map[string]interface{}{
		"selector": "Identifier",
		"message":  "no",
	}})
	if len(got) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if msg := got[0].formatMessage(); msg != "no" {
		t.Fatalf("custom message mismatch: %q", msg)
	}
}

func TestSelectorSpecificityMatchesESLint(t *testing.T) {
	tests := []struct {
		selector    string
		attributes  int
		identifiers int
	}{
		{selector: "Identifier", identifiers: 1},
		{selector: `[name='foo']`, attributes: 1},
		{selector: `Identifier[name='foo']`, attributes: 1, identifiers: 1},
		{selector: `FunctionDeclaration > Identifier.id`, attributes: 1, identifiers: 2},
		{selector: `:nth-child(2)`, attributes: 1},
		// ESLint deliberately excludes selectors nested inside :has().
		{selector: `CallExpression:has(Identifier MemberExpression)`, identifiers: 1},
	}
	for _, test := range tests {
		t.Run(test.selector, func(t *testing.T) {
			parsed, err := parseSelector(test.selector)
			if err != nil {
				t.Fatal(err)
			}
			got := analyzeSelectorSpecificity(parsed)
			if got.attributes != test.attributes || got.identifiers != test.identifiers {
				t.Fatalf("specificity = (%d, %d), want (%d, %d)", got.attributes, got.identifiers, test.attributes, test.identifiers)
			}
		})
	}
}

func TestCandidateKindsUniverseAbsorbsTypedBranches(t *testing.T) {
	parsed, err := parseSelector(`:matches(*, Identifier)`)
	if err != nil {
		t.Fatal(err)
	}
	if kinds := candidateKinds(parsed); !kinds.universe {
		t.Fatalf("candidate kinds = %#v, want universe", kinds)
	}
}
