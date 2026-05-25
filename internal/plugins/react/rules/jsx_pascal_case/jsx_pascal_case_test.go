package jsx_pascal_case

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxPascalCaseRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxPascalCaseRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		// Lowercase-leading tags are DOM intrinsics and not checked.
		{Code: `<testcomponent />;`, Tsx: true},
		{Code: `<testComponent />;`, Tsx: true},
		{Code: `<test_component />;`, Tsx: true},
		// Canonical PascalCase names.
		{Code: `<TestComponent />;`, Tsx: true},
		{Code: `<CSSTransitionGroup />;`, Tsx: true},
		{Code: `<BetterThanCSS />;`, Tsx: true},
		{Code: `<TestComponent><div /></TestComponent>;`, Tsx: true},
		{Code: `<Test1Component />;`, Tsx: true},
		{Code: `<TestComponent1 />;`, Tsx: true},
		{Code: `<T3StComp0Nent />;`, Tsx: true},
		// Unicode letters in PascalCase names.
		{Code: `<Éurströmming />;`, Tsx: true},
		{Code: `<Año />;`, Tsx: true},
		{Code: `<Søknad />;`, Tsx: true},
		// Single-char names always allowed.
		{Code: `<T />;`, Tsx: true},
		// allowAllCaps: SCREAMING_SNAKE_CASE names accepted.
		{Code: `<YMCA />;`, Tsx: true, Options: map[string]interface{}{"allowAllCaps": true}},
		{Code: `<TEST_COMPONENT />;`, Tsx: true, Options: map[string]interface{}{"allowAllCaps": true}},
		// Member-access tag: each dotted part is checked independently.
		{Code: `<Modal.Header />;`, Tsx: true},
		{Code: `<qualification.T3StComp0Nent />;`, Tsx: true},
		// ignore list exact / glob / extglob entries.
		{Code: `<IGNORED />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"IGNORED"}}},
		{Code: `<Foo_DEPRECATED />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"*_D*D"}}},
		{Code: `<Foo_DEPRECATED />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"*_+(DEPRECATED|IGNORED)"}}},
		// Single-char `$` / `_` — length-1 short-circuit exempts them.
		{Code: `<$ />;`, Tsx: true},
		{Code: `<_ />;`, Tsx: true},
		// `H1` is PascalCase (upper, then digit counts as a follow-up char).
		{Code: `<H1>Hello!</H1>;`, Tsx: true},
		// Without allowNamespace, each dotted part is still checked — "P" is
		// length-1 so it's allowed.
		{Code: `<Typography.P />;`, Tsx: true},
		// allowNamespace exits after the first valid part; "h1" is not
		// checked because "Styled" already passed.
		{Code: `<Styled.h1 />;`, Tsx: true, Options: map[string]interface{}{"allowNamespace": true}},
		// allowLeadingUnderscore strips `_` before pascal/all-caps tests.
		{Code: `<_TEST_COMPONENT />;`, Tsx: true, Options: map[string]interface{}{"allowAllCaps": true, "allowLeadingUnderscore": true}},
		{Code: `<_TestComponent />;`, Tsx: true, Options: map[string]interface{}{"allowLeadingUnderscore": true}},

		// ---- Additional edge cases ----
		// JSX within JSX — inner DOM child isn't checked.
		{Code: `<Wrapper><span /></Wrapper>;`, Tsx: true},
		// Self-closing on a member-access base with single-char leaf.
		{Code: `<Foo.T />;`, Tsx: true},
		// `<this.Foo>` — IsDOMComponent classifies as DOM (leading "this"
		// lowercase), so no report is emitted.
		{Code: `<this.Foo />;`, Tsx: true},
		// Chained member access with every level PascalCase.
		{Code: `<Foo.Bar.Baz />;`, Tsx: true},
		{Code: `<Foo.Bar.Baz.Qux />;`, Tsx: true},
		// Member access where a middle / leaf part is single-char — the
		// length-1 short-circuit exempts the whole element.
		{Code: `<Foo.T.Bar />;`, Tsx: true},
		{Code: `<Foo.Bar.T />;`, Tsx: true},
		// allowAllCaps combined with member access — ALL-CAPS leaf allowed.
		{Code: `<Foo.BAR />;`, Tsx: true, Options: map[string]interface{}{"allowAllCaps": true}},
		// allowNamespace combined with allowAllCaps — first ALL-CAPS passes,
		// `h1` never reached.
		{
			Code:    `<FOO.h1 />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true, "allowNamespace": true},
		},
		// All four options together on a "happy path" name.
		{
			Code:    `<_SAFE />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true, "allowLeadingUnderscore": true, "allowNamespace": true, "ignore": []interface{}{"Never"}},
		},
		// Multiple JSX elements in a single source — DOM intrinsics are
		// skipped, user components validated independently.
		{Code: `<div><Wrapper /><span /><OtherOK /></div>;`, Tsx: true},
		// JSX nested inside a JSX expression container — both outer and
		// inner elements are checked.
		{Code: `<Outer>{true && <Inner />}</Outer>;`, Tsx: true},
		// JSX passed as a JSX attribute value (prop children-style) — the
		// nested <Inner /> is its own element and must also pass.
		{Code: `<Outer prop={<Inner />} />;`, Tsx: true},
		// JSX fragment wrapping user components — the fragment itself has
		// no tag name, children are still checked.
		{Code: `<><Foo /><Bar /></>;`, Tsx: true},
		// Conditional / ternary with two user components.
		{Code: `const cond = true; <Outer>{cond ? <Yes /> : <No />}</Outer>;`, Tsx: true},
		// Ignore patterns: character class `[A-Z]` — exact length 2.
		{Code: `<AB />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"[A-Z][A-Z]"}}},
		// Ignore patterns: @() extglob — "match exactly one of".
		{Code: `<Foo_BAR />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"Foo_@(BAR|BAZ)"}}},
		// Ignore patterns: ? single-char wildcard.
		{Code: `<A_B />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"A_?"}}},
		// Upstream valid case — JSX namespaced tag. tsgo parses this
		// natively (KindJsxNamespacedName), so unlike what the upstream
		// test's `features: ['jsx namespace']` flag suggests, no special
		// parser opt-in is needed here.
		{Code: `<Modal:Header />;`, Tsx: true},

		// ---- Robustness: Unicode in ignore patterns ----
		// Non-ASCII literal in pattern — must match rune-for-rune, not
		// byte-for-byte (byte iteration would produce broken regex).
		{Code: `<Año_Old />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"Año_*"}}},
		// Unicode + `?` single-rune wildcard — one `?` must match exactly
		// one rune, not one byte.
		{Code: `<Año_KEEP />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"Año_KE?P"}}},

		// ---- Robustness: malformed ignore patterns don't crash ----
		// Unclosed `[` — the `[` is treated as a literal (regex class
		// never opened); ignore entry effectively exact-matches "[invalid"
		// which doesn't match "Good_Name" → rule runs normally and sees
		// PascalCase → valid.
		{Code: `<Good />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"[invalid"}}},
		// `[^]` — only one rune inside the class (plus the negation marker)
		// would produce the invalid regex `[^]` under a naive translation;
		// our guard emits `\[` instead so compile succeeds with no match.
		{Code: `<Good />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"[^]"}}},
		// Empty `[]` class — same guard, no crash.
		{Code: `<Good />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"[]"}}},

		// ---- TypeScript generics on JSX tag ----
		// Type arguments don't change TagName; rule inspects the tag only.
		{Code: `<Foo<string> />;`, Tsx: true},
		// Real-world namespaced intrinsics: SVG with `svg:` prefix. The
		// elementType ("svg:path") starts lowercase → DOM → skipped.
		{Code: `<svg:path />;`, Tsx: true},
		// Namespaced tag with SCREAMING_SNAKE_CASE leaf + allowAllCaps.
		{Code: `<Modal:NAMED />;`, Tsx: true, Options: map[string]interface{}{"allowAllCaps": true}},
		// Namespace + leading underscore leaf, stripped and validated.
		{Code: `<Modal:_Header />;`, Tsx: true, Options: map[string]interface{}{"allowLeadingUnderscore": true}},
		// Namespace + allowNamespace: first part passes → loop exits
		// before checking the lowercase leaf "h1".
		{Code: `<Styled:h1 />;`, Tsx: true, Options: map[string]interface{}{"allowNamespace": true}},
		// Namespace + ignore: second part matched via glob.
		{Code: `<Modal:Foo_DEPRECATED />;`, Tsx: true, Options: map[string]interface{}{"ignore": []interface{}{"*_DEPRECATED"}}},

		// tsgo rejects these shapes at parse time (they're invalid JSX per
		// spec — namespaces are single-level and don't mix with `.`):
		//   <A:B:C />                 → TS1003 Identifier expected
		//   <Modal:Header.Sub />      → TS1003 Identifier expected
		//   <Foo.Modal:Header />      → TS1003 Identifier expected
		// ESLint/Babel's parser behaves the same.
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `<Test_component />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component Test_component must be in PascalCase", Line: 1, Column: 1},
			},
		},
		{
			Code: `<TEST_COMPONENT />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component TEST_COMPONENT must be in PascalCase", Line: 1, Column: 1},
			},
		},
		{
			Code: `<YMCA />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component YMCA must be in PascalCase", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<_TEST_COMPONENT />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalOrSnakeCase", Message: "Imported JSX component _TEST_COMPONENT must be in PascalCase or SCREAMING_SNAKE_CASE", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<TEST_COMPONENT_ />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalOrSnakeCase", Message: "Imported JSX component TEST_COMPONENT_ must be in PascalCase or SCREAMING_SNAKE_CASE", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<TEST-COMPONENT />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalOrSnakeCase", Message: "Imported JSX component TEST-COMPONENT must be in PascalCase or SCREAMING_SNAKE_CASE", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<__ />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalOrSnakeCase", Message: "Imported JSX component __ must be in PascalCase or SCREAMING_SNAKE_CASE", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<_div />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowLeadingUnderscore": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component _div must be in PascalCase", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<__ />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true, "allowLeadingUnderscore": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalOrSnakeCase", Message: "Imported JSX component __ must be in PascalCase or SCREAMING_SNAKE_CASE", Line: 1, Column: 1},
			},
		},
		{
			Code: `<$a />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component $a must be in PascalCase", Line: 1, Column: 1},
			},
		},
		{
			Code:    `<Foo_DEPRECATED />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"*_FOO"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component Foo_DEPRECATED must be in PascalCase", Line: 1, Column: 1},
			},
		},
		{
			// Member-access: second part "h1" starts lowercase → usePascalCase
			// with name "h1".
			Code: `<Styled.h1 />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component h1 must be in PascalCase", Line: 1, Column: 1},
			},
		},
		{
			Code: `<$Typography.P />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component $Typography must be in PascalCase", Line: 1, Column: 1},
			},
		},
		{
			// With allowNamespace, the outer loop still checks the FIRST
			// part; "STYLED" fails both PascalCase and ALL_CAPS (allowAllCaps
			// is off), so we report before reaching "h1".
			Code:    `<STYLED.h1 />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowNamespace": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component STYLED must be in PascalCase", Line: 1, Column: 1},
			},
		},

		// ---- Additional edge cases ----
		// Multi-line JSX: position assertions anchor at the element opening.
		{
			Code: "<TEST_COMPONENT\n  foo=\"bar\"\n/>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Line: 1, Column: 1, EndLine: 3, EndColumn: 3},
			},
		},
		// Opening form (not self-closing) still reports — listener covers both.
		{
			Code: `<TEST_COMPONENT>hi</TEST_COMPONENT>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Line: 1, Column: 1},
			},
		},
		// allowLeadingUnderscore doesn't rescue a name that's invalid even
		// after stripping the underscore.
		{
			Code:    `<_test />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowLeadingUnderscore": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Line: 1, Column: 1},
			},
		},
		// ignore doesn't match a different casing of the name.
		{
			Code:    `<BAD_NAME />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"GOOD_NAME"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Line: 1, Column: 1},
			},
		},
		// Member access where the first part is valid pascal but second is
		// lowercase — reports on the second part by its name.
		{
			Code: `<Foo.bar />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component bar must be in PascalCase", Line: 1, Column: 1},
			},
		},
		// Chained member access: middle part is lowercase — reports that
		// middle name, not the leaf or the root.
		{
			Code: `<Foo.bad_name.Baz />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component bad_name must be in PascalCase", Line: 1, Column: 1},
			},
		},
		// Leaf is invalid (trailing lowercase) — reports leaf.
		{
			Code: `<Foo.Bar.baz />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component baz must be in PascalCase", Line: 1, Column: 1},
			},
		},
		// Multiple invalid components in one element — rule fires once per
		// offending element, not once per file. `ALL_CAPS` names fail
		// PascalCase (the underscore trips the non-alphanumeric check).
		{
			Code: `<div><FOO_A /><FOO_B /></div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component FOO_A must be in PascalCase", Line: 1, Column: 6},
				{MessageId: "usePascalCase", Message: "Imported JSX component FOO_B must be in PascalCase", Line: 1, Column: 15},
			},
		},
		// Invalid component nested inside a JSX expression container.
		{
			Code: `<Wrapper>{true && <Bad_Inner />}</Wrapper>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component Bad_Inner must be in PascalCase", Line: 1, Column: 19},
			},
		},
		// Invalid component passed as a JSX attribute value.
		{
			Code: `<Wrapper prop={<Bad_Attr />} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component Bad_Attr must be in PascalCase", Line: 1, Column: 16},
			},
		},
		// allowAllCaps switches the messageId to usePascalOrSnakeCase even
		// when the failing reason isn't a SCREAMING_SNAKE_CASE candidate.
		{
			Code:    `<$Weird />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalOrSnakeCase", Message: "Imported JSX component $Weird must be in PascalCase or SCREAMING_SNAKE_CASE", Line: 1, Column: 1},
			},
		},
		// Member access multi-line — position assertion across lines.
		{
			Code: "<Foo.\n  bar />;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component bar must be in PascalCase", Line: 1, Column: 1, EndLine: 2, EndColumn: 9},
			},
		},
		// Namespace with lowercase-Name-with-underscore leaf — report on leaf.
		{
			Code: `<Modal:bad_header />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component bad_header must be in PascalCase", Line: 1, Column: 1},
			},
		},
		// Namespace with uppercase namespace that fails PascalCase — report
		// on namespace part (first split segment).
		{
			Code: `<BAD_NS:Header />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component BAD_NS must be in PascalCase", Line: 1, Column: 1},
			},
		},
		// Namespace + allowAllCaps: leaf is not SCREAMING_SNAKE_CASE and
		// not PascalCase → usePascalOrSnakeCase.
		{
			Code:    `<Modal:bad_Name />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowAllCaps": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalOrSnakeCase", Message: "Imported JSX component bad_Name must be in PascalCase or SCREAMING_SNAKE_CASE", Line: 1, Column: 1},
			},
		},
		// Namespace + allowNamespace: first part fails → still reported
		// (allowNamespace only skips subsequent parts after first success).
		{
			Code:    `<BAD_NS:Header />;`,
			Tsx:     true,
			Options: map[string]interface{}{"allowNamespace": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component BAD_NS must be in PascalCase", Line: 1, Column: 1},
			},
		},
		// Member-access leaf with leading special char — reports the
		// specific failing segment, not the whole chain.
		{
			Code: `<Foo.$Bar />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component $Bar must be in PascalCase", Line: 1, Column: 1},
			},
		},
		// TS generic on an otherwise-invalid tag name still reports.
		{
			Code: `<Bad_Name<string> />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component Bad_Name must be in PascalCase", Line: 1, Column: 1},
			},
		},
		// Unicode ignore pattern — if the ignore regex is built byte-wise
		// instead of rune-wise, it fails to match `<Año_BAD />` via
		// `Año_B*`, and the rule falsely reports usePascalCase. The
		// inverse assertion: a non-matching Unicode pattern lets the
		// rule fire on a genuinely bad name.
		{
			Code:    `<Año_BAD />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"Año_DIFFERENT"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "usePascalCase", Message: "Imported JSX component Año_BAD must be in PascalCase", Line: 1, Column: 1},
			},
		},
	})
}
