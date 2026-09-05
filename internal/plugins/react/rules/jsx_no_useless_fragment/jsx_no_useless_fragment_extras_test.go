package jsx_no_useless_fragment

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

// TestJsxNoUselessFragmentExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// naming the specific branch / Dimension 4 row / tsgo AST quirk it covers, so
// a future refactor can't silently regress it without breaking a named
// lock-in. The 1:1 upstream migration lives in
// jsx_no_useless_fragment_upstream_test.go.
//
// Dimension walk notes for jsx-no-useless-fragment:
//   - Dimension 2 (scoping & nesting): the rule performs no scope lookup; its
//     only ancestor step is `node.parent`, so the relevant nesting axis is
//     JSX containment, covered under "Nesting / traversal boundaries" below.
//   - Dimension 4 (declaration / container forms): N/A — the rule targets JSX
//     elements and fragments only; it never inspects a function or class.
//   - Dimension 4 (element access `X['y']`, private names, computed keys):
//     N/A — a JSX tag name and a JSX attribute name have no computed or
//     private spelling.
func TestJsxNoUselessFragmentExtras(t *testing.T) {
	allowExpressions := []interface{}{map[string]interface{}{"allowExpressions": true}}
	emptyOptions := []interface{}{map[string]interface{}{}}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoUselessFragmentRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: expression wrappers around the container's expression ----
		// ESTree has no ParenthesizedExpression node, so upstream's
		// `expression.type === 'CallExpression'` sees straight through
		// parentheses. tsgo keeps them, so containsCallExpression must skip
		// them to stay aligned.
		{Code: `<>{(foos.map(foo => foo))}</>`, Tsx: true},
		{Code: `<>{((foos.map(foo => foo)))}</>`, Tsx: true},
		// Locks in upstream containsCallExpression() arm 2: a chain broken by
		// parentheses is a plain CallExpression in ESTree too.
		{Code: `<>{(foo?.bar)()}</>`, Tsx: true},

		// ---- Dimension 4: fragment tag-name forms that must NOT match ----
		{Code: `<Fragment.Foo><Bar /></Fragment.Foo>`, Tsx: true},
		{Code: `<React.Fragment.Foo>{foo}</React.Fragment.Foo>`, Tsx: true},
		{Code: `<A.React.Fragment>{foo}</A.React.Fragment>`, Tsx: true},
		{Code: `<this.Fragment>{foo}</this.Fragment>`, Tsx: true},
		{Code: `<react:Fragment>{foo}</react:Fragment>`, Tsx: true},

		// ---- Locks in upstream isKeyedElement() ----
		// A keyed fragment element is meaningful, including the self-closing
		// and value-less attribute spellings that upstream never tests.
		{Code: `<Fragment key="k" />`, Tsx: true},
		{Code: `<React.Fragment key={item.id}>{item.value}</React.Fragment>`, Tsx: true},

		// ---- Locks in upstream isFragmentWithOnlyTextAndIsNotChild() ----
		// A fragment holding a single text child outside any JSX parent is
		// useful (`<Foo content={<>text</>} />`), and stays useful when that
		// text is whitespace only.
		{Code: `<> </>`, Tsx: true},
		{Code: `const a = <>meow</>;`, Tsx: true},
		// Upstream's own `<>cat {meow}</>` comment shape: two children, so
		// nothing is reported and canFix is never consulted.
		{Code: `<>cat {meow}</>`, Tsx: true},

		// ---- Locks in upstream hasLessThanTwoChildren() call-expression arm ----
		{Code: `<>{foo()}</>`, Tsx: true},
		{Code: `<>{foo.bar()}</>`, Tsx: true},

		// ---- Dimension 4: the `@jsx` annotation renames the pragma ----
		// Upstream's pragmaUtil reads the annotation before `settings.react`,
		// so under a Preact pragma `<React.Fragment>` is an ordinary component
		// and must not be reported.
		{Code: "/** @jsx Preact.h */\n<React.Fragment>{value}</React.Fragment>;", Tsx: true},
		// An annotation that is not a valid identifier falls back to `React`,
		// matching upstream's warn-and-default path, so the configured pragma
		// is ignored once any annotation is present.
		{Code: "/** @jsx Foo-Bar */\n<Preact.Fragment>{value}</Preact.Fragment>;", Tsx: true, Settings: fragmentSettings("Preact", "")},
		// `@jsxFrag` is not `@jsx`: upstream's `/@jsx\s+/` needs whitespace
		// straight after the marker, and the fragment pragma is settings-only.
		{Code: "/** @jsxFrag Preact.Fragment */\n<Preact.Fragment>{value}</Preact.Fragment>;", Tsx: true},
		// The annotation is read from comments, not from arbitrary source
		// text, so a string literal cannot rename the pragma.
		{Code: "const marker = '@jsx Preact.h';\n<Preact.Fragment>{value}</Preact.Fragment>;", Tsx: true},

		// ---- Nesting / traversal boundaries ----
		// A fragment whose parent is another fragment is never a child of an
		// HTML element, and two children keep it useful.
		{Code: `<><><div /><div /></><div /></>`, Tsx: true},

		// ---- Option: allowExpressions ----
		{Code: `<Fragment>{moo}</Fragment>`, Tsx: true, Options: allowExpressions},
		{Code: `<React.Fragment>{moo}</React.Fragment>`, Tsx: true, Options: allowExpressions},
		// `{}` is a JSXExpressionContainer to upstream as well, so
		// allowExpressions covers it.
		{Code: `<>{}</>`, Tsx: true, Options: allowExpressions},

		// ---- Settings: pragma / fragment resolved independently ----
		{Code: `<Fragment>{foo}</Fragment>`, Tsx: true, Settings: fragmentSettings("", "Frag")},
		{Code: `<React.Fragment>{foo}</React.Fragment>`, Tsx: true, Settings: fragmentSettings("Act", "")},

		// ---- Real-user: keyed fragments emitted from a list render ----
		{
			Code: `
        function List({ items }) {
          return items.map((item) => (
            <React.Fragment key={item.id}>
              <dt>{item.term}</dt>
              <dd>{item.description}</dd>
            </React.Fragment>
          ));
        }
      `,
			Tsx: true,
		},

		// ---- Real-user: a string-returning component wrapped for JSX.Element ----
		// The workaround the allowExpressions option exists for.
		{
			Code: `
        const Label = ({ text }: { text: string }) => {
          return <>{text}</>;
        };
      `,
			Tsx:     true,
			Options: allowExpressions,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Dimension 4: optional chains are NOT call expressions ----
		// ESTree wraps an optional call in a ChildExpression-free
		// ChainExpression, which upstream's `type === 'CallExpression'` test
		// rejects; tsgo instead flags the CallExpression itself, so the rule
		// must exclude optional chains to stay aligned.
		{
			Code: `<>{foo?.()}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code: `<>{foo?.bar()}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		// ---- Dimension 4: a dynamic import is not a call expression ----
		// ESTree models `import("x")` as an ImportExpression, so upstream's
		// call-expression exemption does not apply; tsgo parses it as a
		// KindCallExpression with an ImportKeyword callee.
		{
			Code: `<>{import("x")}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},

		// ---- Dimension 4: JSX spread children are not expression containers ----
		// ESTree gives `{...value}` its own JSXSpreadChild type, which none of
		// upstream's `type === 'JSXExpressionContainer'` tests match. So the
		// fragment is reported, the call-expression exemption does not apply
		// to `{...make()}`, and — because canFix's non-space-curly guard also
		// skips it — the fix is emitted.
		{
			Code:   `const a = <>{...value}</>;`,
			Tsx:    true,
			Output: []string{`const a = {...value};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 11},
			},
		},
		{
			Code:   `const a = <>{...make()}</>;`,
			Tsx:    true,
			Output: []string{`const a = {...make()};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 11},
			},
		},
		// `allowExpressions` covers `{expr}` only, so a spread child is still
		// reported with the option on.
		{
			Code:    `const a = <>{...value}</>;`,
			Tsx:     true,
			Options: allowExpressions,
			Output:  []string{`const a = {...value};`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 11},
			},
		},

		// ---- Dimension 4: the `@jsx` annotation renames the pragma ----
		// The mirror of the valid case above: the annotated pragma's fragment
		// is the one that gets reported.
		{
			Code: "/** @jsx Preact.h */\n<Preact.Fragment>{value}</Preact.Fragment>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 2, Column: 1},
			},
		},
		// A line comment carries the annotation just as well, and it wins over
		// `settings.react.pragma`.
		{
			Code:     "// @jsx Preact.h\n<Preact.Fragment>{value}</Preact.Fragment>;",
			Tsx:      true,
			Settings: fragmentSettings("Other", ""),
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 2, Column: 1},
			},
		},

		// ---- Dimension 4: TS-only expression wrappers ----
		// ESTree models these as TSAsExpression / TSNonNullExpression /
		// TSSatisfiesExpression — none of them is a CallExpression, so the
		// fragment still needs more children.
		{
			Code: `<>{foo() as JSX.Element}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code: `<>{foo()!}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		// Locks in upstream containsCallExpression() rejecting non-call
		// expression shapes that look call-like.
		{
			Code: `<>{new Foo()}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},

		// ---- Dimension 4: graceful degradation on empty / comment-only containers ----
		// `{/* comment */}` parses to a container with no expression at all;
		// it must not crash and must not be mistaken for a call.
		{
			Code: `<>{/* just a note */}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		// ---- Locks in upstream isKeyedElement() attribute scan ----
		// Only a `key` attribute spares the fragment; other attributes and a
		// spread do not.
		{
			Code: `<Fragment id="x">{foo}</Fragment>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code: `<Fragment {...props}>{foo}</Fragment>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},

		// ---- Locks in upstream isChildOfHtmlElement() `/^[a-z]+$/` ----
		// A digit or a hyphen in the tag name takes the parent out of the
		// HTML-element branch, so only NeedsMoreChildren fires — and, because
		// the parent then counts as a component element, the fragment is left
		// unfixed.
		{
			Code: `<h1><>foo</></h1>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 5},
			},
		},
		{
			Code: `<foo-bar><>foo</></foo-bar>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 10},
			},
		},
		// A non-Identifier parent tag name (member access) also takes the
		// HTML branch out, and additionally blocks the fix because the parent
		// is a component element.
		{
			Code: `<Foo.Bar><>foo</></Foo.Bar>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 10},
			},
		},

		// ---- Locks in upstream isChildOfComponentElement() fragment arm ----
		// A fragment nested directly in another fragment element is fixable:
		// the parent is a fragment, so it can't require a ReactElement child.
		{
			Code:   `<Fragment><>foo</></Fragment>`,
			Tsx:    true,
			Output: []string{`<>foo</>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 11},
			},
		},

		// ---- Nesting / traversal boundaries ----
		// Both the outer and the inner fragment are visited; only the outer
		// one is short of children.
		{
			Code:   `<><><div /><div /></></>`,
			Tsx:    true,
			Output: []string{`<><div /><div /></>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},

		// ---- Locks in upstream trimLikeReact() ----
		// A leading / trailing whitespace run without a "\n" is preserved,
		// because React itself preserves it.
		{
			Code:   `<div>  <>  <b />  </>  </div>`,
			Tsx:    true,
			Output: []string{`<div>    <b />    </div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 8},
			},
		},
		// An all-whitespace children run containing "\n" collapses to the
		// empty string — the JavaScript `slice(start, end)` arm where start
		// runs past end.
		{
			Code: `<div>
<>
</>
</div>`,
			Tsx: true,
			Output: []string{`<div>

</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 2, Column: 1},
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 2, Column: 1},
			},
		},

		// ---- Locks in upstream isFragmentWithOnlyTextAndIsNotChild() ----
		// The single child is an expression container, not text, so the
		// "useful text fragment" escape hatch does not apply even though the
		// fragment has no JSX parent.
		{
			Code: `<div p={<>{foo}</>} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 9},
			},
		},

		// ---- Option contract: allowExpressions defaults to false ----
		// Running with no options and with `[{}]` must produce the same
		// diagnostic.
		{
			Code: `<>{moo}</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		{
			Code:    `<>{moo}</>`,
			Tsx:     true,
			Options: emptyOptions,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 1},
			},
		},
		// allowExpressions only silences NeedsMoreChildren — a fragment passed
		// to an HTML element is still useless.
		{
			Code:    `<div><>{moo}</></div>`,
			Tsx:     true,
			Options: allowExpressions,
			Output:  []string{`<div>{moo}</div>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "ChildOfHtmlElement", Message: childOfHtmlElementDescription, Line: 1, Column: 6},
			},
		},

		// ---- Real-user: ternary branches wrapped in fragments ----
		{
			Code: `const name = showFullName ? <>{fullName}</> : <>{firstName}</>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 29},
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 1, Column: 47},
			},
		},
		// ---- Real-user: a pass-through component wrapping its children ----
		{
			Code: `
        function Wrapper({ children }) {
          return <>{children}</>;
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "NeedsMoreChildren", Message: needsMoreChildrenDescription, Line: 3, Column: 18},
			},
		},
	})
}

// fragmentSettings builds a `settings.react` map with only the entries the
// caller cares about, so a test can vary `pragma` and `fragment`
// independently and confirm each falls back to its own default.
func fragmentSettings(pragma, fragment string) map[string]interface{} {
	react := map[string]interface{}{}
	if pragma != "" {
		react["pragma"] = pragma
	}
	if fragment != "" {
		react["fragment"] = fragment
	}
	return map[string]interface{}{"react": react}
}

// TestJsxNoUselessFragmentEditDemand exercises Dimension 3 (autofix
// boundaries): diagnostic count, message, and range must stay identical
// across every edit demand, and the replacement text must materialize only
// when an autofix is actually requested.
func TestJsxNoUselessFragmentEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		"const a = <><div /></>;\nconst b = <><span /></>;\nconst c = <this><><A/><B/></></this>;",
		"edit-demand.tsx",
		"tsconfig.json",
	)
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
					Name:     JsxNoUselessFragmentRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return JsxNoUselessFragmentRule.Run(ctx, nil)
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
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: diagnostics = %d, want 3", demand, len(diagnostics))
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
		if fixes := *allEdits[index].FixesPtr; len(fixes) == 0 {
			t.Fatalf("diagnostic %d: all-edits demand produced no fixes", index)
		}
		if diagnostic := allEdits[index]; diagnostic.Suggestions != nil {
			t.Fatalf("diagnostic %d: autofix-only rule materialized suggestions", index)
		}
	}
}
