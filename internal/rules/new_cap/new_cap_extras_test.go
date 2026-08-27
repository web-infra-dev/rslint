// TestNewCapExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in. See
// new_cap_upstream_test.go for the migrated upstream suite.
package new_cap

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNewCapExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NewCapRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: a TS assertion around the entire callee remains a
		// distinct expression. TSESTree does not treat it as an Identifier or
		// MemberExpression, so neither does this rule. ----
		{Code: "new (factory as any)();"},
		{Code: "new (factory!)();"},
		{Code: "new (factory satisfies unknown)();"},

		// typescript-eslint rejects empty type argument lists before rules run.
		// ts-go recovers call/new nodes, which must not produce extra diagnostics.
		{Code: "Factory<>(1, 2);"},
		{Code: "new factory<>(1, 2);"},
		{Code: "service.Factory<>(1);"},
		{Code: "new service.factory<>(1);"},

		// ---- Dimension 4: dynamic computed keys are not static names. ----
		{Code: "new namespace[Factory]();"},
		{Code: "new namespace[`factory${suffix}`]();"},
		{Code: "namespace[Symbol.iterator]();"},

		// ---- Dimension 4: numeric and private keys begin with non-alphabetic
		// characters and are accepted. ----
		{Code: "new namespace[0]();"},
		{Code: "class C { #factory; make() { return new this.#factory(); } }"},

		// ---- Dimension 4: an astral letter starts with a UTF-16 high
		// surrogate under JavaScript's charAt(0), so upstream classifies both
		// astral-uppercase and astral-lowercase names as non-alphabetic. ----
		{Code: `namespace["𝐀"]();`},
		{Code: `new namespace["𝐚"]();`},

		// Locks in isCapAllowed() Date.UTC arm: static element access is a
		// MemberExpression too, and its object is the direct Date identifier.
		{Code: `Date["UTC"]();`},
		{Code: `Date.UTC();`, Options: map[string]any{"properties": false}},
		{Code: `Date["UTC"]();`, Options: map[string]any{"properties": false}},

		// Locks in isCapAllowed() allowedMap arm 1: property-name exception.
		{Code: "service.Factory();", Options: map[string]any{"capIsNewExceptions": []any{"Factory"}}},

		// Locks in isCapAllowed() allowedMap arm 2: full callee-source exception.
		// Parentheses do not become part of sourceCode.getText(callee) upstream.
		{Code: "((service.Factory))();", Options: map[string]any{"capIsNewExceptions": []any{"service.Factory"}}},

		// Locks in isCapAllowed() pattern arm: the regexp sees the full callee,
		// not only the final property name (ESLint issue #12874).
		{Code: "api.deep.Factory();", Options: map[string]any{"capIsNewExceptionPattern": `^api\.deep\.`}},

		// Locks in isCapAllowed() skipProperties arm for each listener.
		{Code: "service.Factory();", Options: map[string]any{"properties": false}},
		{Code: "new service.factory();", Options: map[string]any{"properties": false}},

		// Locks in listener construction: disabling each half removes only that
		// listener.
		{Code: "Factory();", Options: map[string]any{"capIsNew": false}},
		{Code: "new factory();", Options: map[string]any{"newIsCap": false}},

		// ---- N/A: declaration/container forms are not rule inputs. new-cap
		// inspects CallExpression and NewExpression callees independently of
		// the surrounding function/class declaration form. ----

		// ---- N/A: object-literal spreads, binding rests, body-less members,
		// and empty bodies are not inspected. Spread arguments and empty
		// argument lists are covered on the invalid side below. ----
	}, []rule_tester.InvalidTestCase{
		// A non-empty type argument list is valid TypeScript and does not hide
		// the direct callee. This guards the boundary of the recovery check above.
		{Code: "Factory<string>();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},
		{Code: "new factory<string>();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 5, EndLine: 1, EndColumn: 12}}},

		// ---- Dimension 4: multi-level parenthesized callee. ESTree discards
		// both wrappers, so the inner identifier remains the report target. ----
		{Code: "new ((factory))();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 7, EndLine: 1, EndColumn: 14}}},

		// ---- Dimension 4: TS non-null/assertion/satisfies wrappers on the
		// receiver do not hide the static property name. ----
		{Code: "new namespace!.factory();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 16, EndLine: 1, EndColumn: 23}}},
		{Code: "new (namespace as any).factory();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 24, EndLine: 1, EndColumn: 31}}},
		{Code: "new (namespace satisfies object).factory();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 34, EndLine: 1, EndColumn: 41}}},

		// ---- Dimension 4: optional-call root. The optional-chain flag is on
		// the CallExpression; its Identifier callee is still checked. ----
		{Code: "Factory?.();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},

		// ---- Dimension 4: static string key is checked and the diagnostic
		// covers the literal; dynamic/numeric/private counterparts are valid
		// cases above. ----
		{Code: `new namespace["factory"]();`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 15, EndLine: 1, EndColumn: 24}}},

		// ---- Dimension 4: same-kind nesting. The outer NewExpression and
		// inner CallExpression are judged independently without listener state
		// bleeding across the boundary. ----
		{
			Code: "new factory(Factory());",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "lower", Line: 1, Column: 5, EndLine: 1, EndColumn: 12},
				{MessageId: "upper", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
			},
		},

		// ---- Dimension 4: spread arguments and an otherwise empty argument
		// list do not affect callee classification or cause a crash. ----
		{Code: "Factory(...args);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},
		{Code: "Factory();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},

		// Locks in extractNameFromExpression() Identifier arm and getCap()
		// lowercase arm under an explicit true option.
		{Code: "new factory();", Options: map[string]any{"newIsCap": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 5, EndLine: 1, EndColumn: 12}}},

		// rslint intentionally uses a true string set for exceptions instead of
		// inheriting accidental exception names from Object.prototype upstream.
		{Code: "new constructor();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "lower", Line: 1, Column: 5, EndLine: 1, EndColumn: 16}}},

		// Locks in getCap() uppercase arm under an explicit true option.
		{Code: "Factory();", Options: map[string]any{"capIsNew": true}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},

		// Locks in the default-vs-empty-object contract: both retain all three
		// true defaults and produce the same diagnostic.
		{Code: "service.Factory();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},
		{Code: "service.Factory();", Options: map[string]any{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 9, EndLine: 1, EndColumn: 16}}},

		// Locks in properties:false direct-callee branch: it exempts only
		// MemberExpressions, never a direct Identifier.
		{Code: "Factory();", Options: map[string]any{"properties": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 1, EndLine: 1, EndColumn: 8}}},

		// UTC is special only on Date. Other receivers remain reportable even
		// when ordinary property calls are disabled.
		{Code: "service.UTC();", Options: map[string]any{"properties": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}}},
		{Code: `service["UTC"]();`, Options: map[string]any{"properties": false}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 1, Column: 9, EndLine: 1, EndColumn: 14}}},

		// ---- Real-user: ESLint #14970. Intl.DateTimeFormat is callable
		// without new at runtime, but the rule intentionally still reports it
		// as a constructor-style name. ----
		{
			Code: "Intl.DateTimeFormat().resolvedOptions().timeZone;",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "upper", Message: "A function with a name starting with an uppercase letter should only be used as a constructor.",
				Line: 1, Column: 6, EndLine: 1, EndColumn: 20,
			}},
		},

		// ---- Real-user: ESLint #15724. Imported names are not exempted;
		// Deserialize() remains reportable unless the user configures it as an
		// exception. ----
		{
			Code:   "import { Deserialize } from 'cerialize';\nconst value = Deserialize(json, Type);",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "upper", Line: 2, Column: 15, EndLine: 2, EndColumn: 26}},
		},
	})
}

var getCapBenchmarkResult capitalization

func benchmarkGetCapLongName(b *testing.B) {
	name := strings.Repeat("a", 1<<20)
	b.ResetTimer()
	for range b.N {
		getCapBenchmarkResult = getCap(name)
	}
}

func TestGetCapLongNameAllocation(t *testing.T) {
	result := testing.Benchmark(benchmarkGetCapLongName)
	if bytes := result.AllocedBytesPerOp(); bytes > 1024 {
		t.Fatalf("getCap allocated %d bytes per long name; want constant-space classification", bytes)
	}
}

func TestNewCapPatternSchema(t *testing.T) {
	for _, key := range []string{"newIsCapExceptionPattern", "capIsNewExceptionPattern"} {
		invalid := []any{map[string]any{key: "("}}
		if err := NewCapRule.Schema.Validate(invalid); err == nil {
			t.Errorf("expected invalid %s regexp to fail schema validation", key)
		}
		valid := []any{map[string]any{key: "(?<=namespace\\.)Factory$"}}
		if err := NewCapRule.Schema.Validate(valid); err != nil {
			t.Errorf("expected ECMAScript lookbehind in %s to validate: %v", key, err)
		}
	}
}
