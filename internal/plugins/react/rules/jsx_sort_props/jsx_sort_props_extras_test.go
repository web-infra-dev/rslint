// TestJsxSortPropsExtras locks in tsgo shapes and upstream branches not fully
// represented by the upstream migration in jsx_sort_props_upstream_test.go.
package jsx_sort_props

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxSortPropsExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxSortPropsRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: JSX namespaced names are read as their complete name. ----
		{Code: `<App aria:a aria:b />`, Tsx: true},
		// ---- Dimension 4: a non-DOM component does not reserve dangerouslySetInnerHTML. ----
		{Code: `<App a dangerouslySetInnerHTML={{ __html: "x" }} />`, Tsx: true, Options: map[string]any{"reservedFirst": true}},
		// ---- Locks in upstream spread arm: a trailing spread does not report or panic. ----
		{Code: `<App b {...props} a />`, Tsx: true},
		// ---- Real-user: issue #3612 comment-bearing attribute blocks preserve comments. ----
		{Code: `<App a /* explanation */ b />`, Tsx: true},
		// ---- Real-user: issue #1632 gives reserved props precedence over callbacks. ----
		{Code: `<App key={1} a onClick={fn} />`, Tsx: true, Options: map[string]any{"reservedFirst": true, "callbacksLast": true}},
	}, []rule_tester.InvalidTestCase{
		// A moved line comment carries its exact terminator when the destination
		// slot would otherwise let it consume the closing tag.
		{Code: "<App b={1} // c\n  a={1} />", Tsx: true, Output: []string{"<App a={1}\n  b={1} // c\n />"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		// The source slot keeps its separator even with no indentation, so the
		// replacement cannot fuse two shorthand prop names.
		{Code: "<App b // c\na />", Tsx: true, Output: []string{"<App a\nb // c\n />"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		// The same guard protects ordinary attributes and spreads later on the
		// destination line, not just an inline closing tag.
		{Code: "<App\n b // b\n c a\n/>", Tsx: true, Output: []string{"<App\n a\n b // b\n c\n/>"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		{Code: "<App\n b // b\n a {...p}\n/>", Tsx: true, Output: []string{"<App\n a\n b // b\n {...p}\n/>"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		{Code: "<App b // b\n a>child</App>", Tsx: true, Output: []string{"<App a\n b // b\n>child</App>"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		// CRLF and the non-ASCII ECMAScript line separators are preserved rather
		// than normalized while moving a trailing line comment.
		{Code: "<App b // b\r\n a />", Tsx: true, Output: []string{"<App a\r\n b // b\r\n />"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		{Code: "<App b // b\r a />", Tsx: true, Output: []string{"<App a\r b // b\r />"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		{Code: "<App b // b\u2028 a />", Tsx: true, Output: []string{"<App a\u2028 b // b\u2028 />"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		{Code: "<App b // b\u2029 a />", Tsx: true, Output: []string{"<App a\u2029 b // b\u2029 />"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		// Each moved line-comment source is checked independently; a source that
		// remains in its own slot does not receive another terminator.
		{Code: "<App c // c\n b // b\n a d />", Tsx: true, Output: []string{"<App a\n b // b\n c // c\n d />"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}, {MessageId: "sortPropsByAlpha"}}},
		// Sorting comment blocks must not reverse duplicate props: JSX uses the
		// last duplicate value, so this group is reported but deliberately not fixed.
		{Code: `<App z="z" /* c */ x="first" a x="last" />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}, {MessageId: "sortPropsByAlpha"}, {MessageId: "sortPropsByAlpha"}}},
		// Attributes excluded by a complex comment sequence are stationary but
		// still participate in the projected duplicate order.
		{Code: "<App z x=\"first\" /* one */ // two\n x=\"last\" />", Tsx: true, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}, {MessageId: "sortPropsByAlpha"}}},
		// Stable duplicate order remains fixable when the projected sequence
		// keeps the same last-value-wins attribute.
		{Code: `<App b a="first" a="last" />`, Tsx: true, Output: []string{`<App a="first" a="last" b />`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}, {MessageId: "sortPropsByAlpha"}}},
		// Explicit locale ordering follows localeCompare for the Nordic cases
		// where x/text's old bundled tables disagree with Node 24 / ICU 78.
		{Code: `<App aa b />`, Tsx: true, Options: map[string]any{"locale": "nb"}, Output: []string{`<App b aa />`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		{Code: `<App AA aA />`, Tsx: true, Options: map[string]any{"locale": "da"}, Output: []string{`<App aA AA />`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		{Code: `<App a A />`, Tsx: true, Options: map[string]any{"locale": "mt"}, Output: []string{`<App A a />`}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}}},
		// A comment after a spread stays outside the sortable group that follows it.
		{Code: `<App {...p} /* gap */ d c />`, Tsx: true, Output: []string{`<App {...p} /* gap */ c d />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 25)}},
		{Code: "<App {...p}\n /* gap */\n d c />", Tsx: true, Output: []string{"<App {...p}\n /* gap */\n c d />"}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 3, 4)}},
		// Comments within the next attribute's initializer are not inter-attribute comments.
		{Code: `<App c a={/* inside */ 1} b />`, Tsx: true, Output: []string{`<App a={/* inside */ 1} b c />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 8), jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 27)}},
		// A comment sequence that cannot be grouped excludes only its own attribute from fixing.
		{Code: "<App c a={1} /* one */ // two\n b />", Tsx: true, Output: []string{"<App b a={1} /* one */ // two\n c />"}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 8), jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 2, 2)}},
		// Reports from one JSX element follow source order, as ESLint's diagnostics do.
		{Code: `<App onClick onBlur a />`, Tsx: true, Options: map[string]any{"callbacksLast": true}, Output: []string{`<App a onBlur onClick />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listCallbacksLast", "Callbacks must be listed after all other props", 1, 6), jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 14)}},
		// Locks in upstream alphabetic arm: errors are evaluated independently for each inversion.
		{Code: `<App c a b />`, Tsx: true, Output: []string{`<App a b c />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 8), jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 10)}},
		// Locks in upstream custom-reserved-list branch and the user-visible data substitution.
		{Code: `<App ref={ref} key={key} />`, Tsx: true, Options: map[string]any{"reservedFirst": []any{"key"}}, Output: []string{`<App key={key} ref={ref} />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("listReservedPropsFirst", "Reserved props must be listed before all other props", 1, 16)}},
		// An empty locale follows upstream's falsy fallback to "auto".
		{Code: `<App Z a />`, Tsx: true, Options: map[string]any{"locale": ""}},
		// ICU's Danish collation puts the uppercase spelling first for equal letters.
		{Code: `<App cH ch />`, Tsx: true, Options: map[string]any{"locale": "da"}},
		// ---- Dimension 4: comments move with the preceding prop in an autofix. ----
		{Code: `<App b /* b comment */ a />`, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 24)}},
		// ---- Dimension 4: member and intrinsic tag forms use the same attribute order. ----
		{Code: `<UI.Button b a />`, Tsx: true, Output: []string{`<UI.Button a b />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 14)}},
		{Code: `<svg:path b a />`, Tsx: true, Output: []string{`<svg:path a b />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 13)}},
		// ---- Real-user: issue #3936 trailing comments are included in the moved attribute. ----
		{Code: "<App\n  b // b comment\n  a\n/>", Tsx: true, Output: []string{"<App\n  a\n  b // b comment\n/>"}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 3, 3)}},
		// A comment before a spread must not let the fixer move either side across the spread boundary.
		{Code: `<App b /* b */ {...p} d c />`, Tsx: true, Output: []string{`<App b /* b */ {...p} c d />`}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 1, 25)}},
		// A trailing comment on the attribute absorbed by a preceding comment stays with that attribute.
		{Code: "<App\n b /* b */\n // leading a\n a // a\n c\n/>", Tsx: true, Output: []string{"<App\n c\n b /* b */\n // leading a\n a // a\n/>"}, Errors: []rule_tester.InvalidTestCaseError{jsxSortError("sortPropsByAlpha", "Props should be sorted alphabetically", 4, 2)}},
		// An absorbed attribute's trailing // also receives a terminator when its
		// block moves into an inline closing-tag slot.
		{Code: "<App\n z /* z */\n // leading b\n b // b\n a />", Tsx: true, Output: []string{"<App\n a\n z /* z */\n // leading b\n b // b\n />"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "sortPropsByAlpha"}, {MessageId: "sortPropsByAlpha"}}},
		// Invalid custom lists report every ordinary prop without applying the sorter.
		{Code: `<App b a />`, Tsx: true, Options: map[string]any{"reservedFirst": []any{}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "listIsEmpty"}, {MessageId: "listIsEmpty"}}},
		// A leading spread does not hide the following ordinary prop from validation.
		{Code: `<App {...p} key={1} />`, Tsx: true, Options: map[string]any{"reservedFirst": []any{}}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "listIsEmpty"}}},
	})
}

func TestDestinationHasLineTerminator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		gap  string
		want bool
	}{
		{name: "line feed", gap: "  \n next", want: true},
		{name: "line separator", gap: "\u2028next", want: true},
		{name: "same line token", gap: "  next"},
		// A comment is a token boundary for this scan. Its internal newline
		// cannot protect the comment opener from a moved // comment.
		{name: "block comment containing newline", gap: " /* gap\n */ next"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := destinationHasLineTerminator(test.gap, 0, len(test.gap)); got != test.want {
				t.Fatalf("destinationHasLineTerminator(%q) = %v, want %v", test.gap, got, test.want)
			}
		})
	}
}

// TestJsxSortPropsEditDemand locks the deferred-fix contract: asking only for
// diagnostics or suggestions must not materialize the shared element fix plan.
func TestJsxSortPropsEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(`<App c b a />`, "edit-demand.tsx", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     JsxSortPropsRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return JsxSortPropsRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index := range allEdits {
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got, want := withoutEdits(diagnostics[index]), withoutEdits(allEdits[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d changed diagnostic %d:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
		if diagnosticsOnly[index].FixesPtr != nil || suggestionOnly[index].FixesPtr != nil {
			t.Fatalf("diagnostic %d: non-autofix demand materialized fixes", index)
		}
		if autofixOnly[index].FixesPtr == nil ||
			!reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Fatalf("diagnostic %d: autofix and all-edits demands produced different fixes", index)
		}
		if len(*allEdits[index].FixesPtr) == 0 {
			t.Fatalf("diagnostic %d: all-edits demand produced no fixes", index)
		}
		if allEdits[index].Suggestions != nil {
			t.Fatalf("diagnostic %d: autofix-only rule materialized suggestions", index)
		}
	}
}
