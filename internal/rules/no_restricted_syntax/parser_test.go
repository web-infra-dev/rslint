package no_restricted_syntax

import (
	"strings"
	"testing"
)

func TestParseSelector_Wellformed(t *testing.T) {
	cases := []string{
		"Identifier",
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
		"Literal[regex.flags=/./]",
		"Literal[regex.flags=/i/]",
		`ImportDeclaration[source.value=/^some\/path$/]`,
		"ArrowFunctionExpression > BlockStatement",
		"Property > Literal.key",
		"FunctionDeclaration FunctionExpression",
		"Literal + Literal",
		"* ~ *",
		":is(Identifier, Literal)",
		":matches(Identifier, Literal)",
		":not(VariableDeclaration)",
		":has(Literal)",
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
		":nth-child",                         // missing arg
		":nth-child(",                        // unterminated paren
		":nth-child(abc)",                    // non-numeric arg
		":is(",                               // unterminated paren
		":is(,)",                             // empty selector in args
		"Identifier,",                        // trailing comma
		"Identifier >",                       // dangling combinator
		"Identifier >> Foo",                  // double combinator
		":unknownPseudo",                     // unsupported pseudo
		"FunctionDeclaration[params.length>", // missing operand
		"BinaryExpression[name=&&]",          // illegal operator value
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
	t.Run("single string (CLI / single-option config shape)", func(t *testing.T) {
		got := parseRuleOptions("Identifier")
		if len(got) != 1 || got[0].selector != "Identifier" {
			t.Fatalf("unexpected entries: %#v", got)
		}
	})
	t.Run("single map (CLI / single-option config shape)", func(t *testing.T) {
		got := parseRuleOptions(map[string]interface{}{
			"selector": "Identifier",
			"message":  "no identifiers",
		})
		if len(got) != 1 || got[0].selector != "Identifier" || got[0].message != "no identifiers" {
			t.Fatalf("unexpected entries: %#v", got)
		}
	})
	t.Run("multi-element array of mixed forms (rule_tester shape)", func(t *testing.T) {
		got := parseRuleOptions([]interface{}{
			"WithStatement",
			map[string]interface{}{"selector": "VariableDeclaration", "message": "x"},
		})
		if len(got) != 2 {
			t.Fatalf("expected 2, got %d", len(got))
		}
		if got[0].selector != "WithStatement" || got[0].message != "" {
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
	got := parseRuleOptions("Identifier")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if msg := got[0].formatMessage(); msg != "Using 'Identifier' is not allowed." {
		t.Fatalf("default message mismatch: %q", msg)
	}
}

func TestParseRuleOptions_CustomMessage(t *testing.T) {
	got := parseRuleOptions(map[string]interface{}{
		"selector": "Identifier",
		"message":  "no",
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if msg := got[0].formatMessage(); msg != "no" {
		t.Fatalf("custom message mismatch: %q", msg)
	}
}
