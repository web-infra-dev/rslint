// TestJsxClosingBracketLocationExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise. Each case carries an inline
// comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk
// it covers, so future refactors can't silently regress them without breaking
// a named lock-in.
//
// stylistic-specific delta vs. react/jsx-closing-bracket-location: the
// trailing-comment-upgrades-after-tag/after-props-to-line-aligned branch is
// unique to this rule. Several lock-ins below name that branch directly so
// regressing it surfaces immediately.
package jsx_closing_bracket_location

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxClosingBracketLocationExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxClosingBracketLocationRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: receiver / tag-name forms ----
		// JsxMemberExpression tag (`<Foo.Bar />`) — after-tag, no props.
		{Code: `<Foo.Bar />`, Tsx: true},
		// JsxMemberExpression tag, multi-line — default tag-aligned.
		{Code: "<Foo.Bar\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		// JsxNamespacedName tag (`<svg:circle />`) — after-tag.
		{Code: `<svg:circle />`, Tsx: true},
		// TS type arguments (`<App<T> foo />`) — single line, after-props.
		{Code: `<App<T> foo />`, Tsx: true},
		// TS type arguments, multi-line tag-aligned.
		{Code: "<App<T>\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- Dimension 4: '>' inside attribute expression must not be confused ----
		// `x > y` inside `{}` is a valid expression; closing-detection must
		// pick the JSX-element '>' / '/' at the end.
		{Code: `<App foo={x > y} />`, Tsx: true},
		// Nested JSX as attribute value — both elements are valid.
		{Code: `<App content={<Foo />} />`, Tsx: true},
		{Code: "<App\n  content={<Foo />}\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- Dimension 4: TS-only expression wrappers in attribute values ----
		// `as`, `satisfies`, `!`, `?.` — all may contain `>` and must not
		// disturb closing detection.
		{Code: `<App foo={x as T} />`, Tsx: true},
		{Code: `<App foo={x!} />`, Tsx: true},
		{Code: `<App foo={x?.y} />`, Tsx: true},
		{Code: `<App foo={x satisfies T} />`, Tsx: true},

		// ---- Dimension 4: spread attribute as last prop ----
		// `{...rest}` is JsxSpreadAttribute (not JsxAttribute); the
		// lastProp-driven column should still be the spread's column.
		{Code: `<App {...props} />`, Tsx: true},
		{Code: "<App\n  foo\n  {...rest}\n/>", Tsx: true},
		// Spread is last prop, props-aligned.
		{Code: "<App\n  foo\n  {...rest}\n  />", Tsx: true, Options: map[string]interface{}{"location": "props-aligned"}},

		// ---- Dimension 4: nesting boundaries ----
		// Multi-level nested elements — every OpeningElement processed
		// independently, no boundary leak.
		{Code: "<App>\n  <B>\n    <C />\n  </B>\n</App>", Tsx: true},
		// Fragment containing inner element — fragments are NOT JsxOpening/
		// SelfClosing kinds, so they don't fire this rule; inner element does.
		{Code: "<>\n  <App foo />\n</>", Tsx: true},

		// ---- Dimension 4: graceful degradation ----
		// Empty attribute list. `<App></App>` non-self-closing zero attrs.
		{Code: `<App></App>`, Tsx: true},
		// `<App />` single-line — after-tag passes (tag.line === closing.line).
		{Code: `<App />`, Tsx: true},

		// ---- Options JSON path robustness (CLI / rule_tester / map / array shapes) ----
		// Array-wrapped string.
		{Code: `<App foo />`, Tsx: true, Options: []interface{}{"after-props"}},
		// Array-wrapped map.
		{Code: `<App foo />`, Tsx: true, Options: []interface{}{map[string]interface{}{"location": "after-props"}}},
		// Empty array of options falls back to defaults.
		{Code: `<App />`, Tsx: true, Options: []interface{}{}},
		// nonEmpty:false alone — closing of `<App\n>...</App>` unchecked.
		{Code: "<App\n  foo\n  ></App>", Tsx: true, Options: map[string]interface{}{"nonEmpty": false}},
		// selfClosing:false alone — closing of `<App\n  foo\n  />` unchecked.
		{Code: "<App\n  foo\n  />", Tsx: true, Options: map[string]interface{}{"selfClosing": false}},

		// ---- Real-user: CRLF line terminators (Windows) ----
		{Code: "<App\r\n  foo\r\n/>", Tsx: true},
		// CRLF with nested provider.
		{Code: "<Provider\r\n  store\r\n>\r\n  <App\r\n    foo\r\n  />\r\n</Provider>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- Real-user: JSX expression-internal comments ----
		// `/* */` inside attribute value expression — comment trivia inside
		// `{}` is at JS comment level, NOT a JSX trailing comment, so it
		// must NOT trigger the after-tag/after-props upgrade.
		{Code: `<App foo={/* leading */ 1} />`, Tsx: true},
		// `//` comment inside expression on its own line.
		{Code: "<App\n  foo={\n    // line comment\n    1\n  }\n/>", Tsx: true},

		// ---- Locks in `findTrailingComment` discriminator ----
		// Comment INSIDE attribute value `{}` should NOT count as a trailing
		// comment relative to the last attribute — it's inside the
		// attribute's source range, not in the trivia after it.
		{Code: "<App\n  foo={/* x */ 1}\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- JSX as arrow-fn body / ternary branch / logical-AND ----
		{Code: "var f = () => <App\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		{Code: "var x = cond && <App\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		{Code: "var x = cond ? <App\n  foo\n/> : null", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},

		// ---- options nil / empty / unknown shapes degrade to defaults ----
		// nil options → tag-aligned default applied.
		{Code: `<App />`, Tsx: true, Options: nil},
		// Empty map options → defaults.
		{Code: `<App />`, Tsx: true, Options: map[string]interface{}{}},
		// Unknown enum string → falls back to that value being used as a
		// location key; hasCorrectLocation returns true on unknown, so no
		// report on the simple `<App />` form.
		{Code: `<App />`, Tsx: true, Options: "unknown-location-value"},
		// `location` form sets both nonEmpty and selfClosing.
		{Code: "<App\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		// `location` and `nonEmpty/selfClosing` are mutually exclusive
		// upstream-side via `'location' in config!`. When `location` is
		// present, the per-form keys are IGNORED — even though the JSON
		// schema rejects this mix, rslint's runtime preserves the same
		// precedence. Here both keys say tag-aligned (selfClosing override
		// would say props-aligned, but is ignored).
		{Code: "<App\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned", "selfClosing": "props-aligned"}},
		{Code: "<App\n  foo\n></App>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned", "nonEmpty": "props-aligned"}},

		// ---- Dimension 4: attribute value robustness ----
		// Regex literal containing '>' in attribute expression — closing
		// detection must not be confused.
		{Code: `<App pattern={/>/} />`, Tsx: true},
		// Template literal with embedded expression containing '>'.
		{Code: "<App t={`${a > b}`} />", Tsx: true},
		// String literal containing JSX-like text.
		{Code: `<App title="<X>" />`, Tsx: true},
		// Multi-line array literal as attribute value, single-line outer
		// element — opening line != lastProp.lastLine, so 'after-props' does
		// NOT collapse; goes to default tag-aligned.
		{Code: "<App foo={[\n  1,\n  2,\n]}\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- Dimension 4: deeply nested elements (5 levels) ----
		// Lock-in: each opening element checked independently; no boundary
		// leak between levels. All levels self-closing, single-line — no
		// reports.
		{Code: `<L1><L2><L3><L4><L5 /></L4></L3></L2></L1>`, Tsx: true},
		// Same, multi-line at innermost; tag-aligned should match per-level.
		{Code: "<L1>\n  <L2>\n    <L3>\n      <L4>\n        <L5\n          foo\n        />\n      </L4>\n    </L3>\n  </L2>\n</L1>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- Real-user: long lastProp spanning many lines (object literal) ----
		{
			Code: "<App\n  config={{\n    a: 1,\n    b: 2,\n    c: {\n      nested: true,\n    },\n  }}\n/>",
			Tsx:  true,
		},
		// ---- Real-user: JSX attribute name with hyphen (data-*, aria-*) ----
		{Code: `<div data-testid="x" aria-label="y" />`, Tsx: true},
		{Code: "<div\n  data-testid=\"x\"\n  aria-label=\"y\"\n/>", Tsx: true},

		// ---- Dimension 4: TypeScript nested generic with `>>` token ----
		// `Map<K, V>>` produces nested `>` chars. tsgo parses type-args
		// separately; elemEnd is still after JSX closing `>`. The reverse
		// gtPos scan must land on the JSX `>`, not a type-arg `>`.
		{Code: `<App<Map<string, number>> foo />`, Tsx: true},
		{Code: "<App<Map<string, number>>\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		// Triple-nested generic.
		{Code: "<App<Promise<Map<string, Array<number>>>>\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- Dimension 4: complex generic with object type literal ----
		{Code: "<App<{ a: number; b: string }>\n  foo\n/>", Tsx: true},

		// ---- Dimension 4: JSX as expression statement in various contexts ----
		// throw expression — `/>` tag-aligned to `<Err` column (8).
		{Code: "function f() {\n  throw <Err\n          code={1}\n        />;\n}", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		// yield in generator — `/>` tag-aligned to `<App` column (8).
		{Code: "function* g() {\n  yield <App\n          foo\n        />;\n}", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		// await in async function — `/>` line-aligned to the line indent.
		{Code: "async function f() {\n  return await Promise.resolve(<App\n    foo\n  />);\n}", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		// default parameter — `/>` line-aligned to the function-decl line indent.
		{Code: "function f(x = <App\n  foo\n/>) { return x; }", Tsx: true, Options: map[string]interface{}{"location": "line-aligned"}},
		// Inside template literal expression — single line element, no report.
		{Code: "var x = `${(<App foo />)}`;", Tsx: true},

		// ---- Dimension 4: multiple consecutive spreads ----
		{Code: `<App {...a} {...b} {...c} />`, Tsx: true},
		{Code: "<App\n  {...a}\n  {...b}\n  {...c}\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- Dimension 4: HTML entities in attribute (decoded by tsgo) ----
		{Code: `<App title="&amp;" />`, Tsx: true},
		{Code: `<App title="&#x3E;&#x3C;" />`, Tsx: true},

		// ---- Real-user: styled-components / emotion css prop ----
		{
			Code: "<Box\n  css={{\n    display: 'flex',\n    flexDirection: 'column',\n    padding: '1rem',\n  }}\n/>",
			Tsx:  true,
		},
		// Styled component with template literal.
		{
			Code: "<Box\n  css={`\n    display: flex;\n    color: red;\n  `}\n/>",
			Tsx:  true,
		},

		// ---- Real-user: redux Provider with multi-line store ----
		{
			Code: "<Provider\n  store={configureStore({\n    reducer,\n    middleware,\n  })}\n>\n  <App />\n</Provider>",
			Tsx:  true,
		},

		// ---- Real-user: React Router shape ----
		{
			Code: "<Route\n  path=\"/users/:id\"\n  exact\n  render={({ match }) => (\n    <UserPage id={match.params.id} />\n  )}\n/>",
			Tsx:  true,
		},

		// ---- Real-user: React.memo / forwardRef-wrapped components ----
		{
			Code: "<ForwardedRefComp\n  ref={ref}\n  {...props}\n/>",
			Tsx:  true,
		},

		// ---- Real-user: conditional rendering with ternary in JSX child ----
		{
			Code: "<Wrapper>\n  {flag ? (\n    <On\n      foo\n    />\n  ) : (\n    <Off\n      bar\n    />\n  )}\n</Wrapper>",
			Tsx:  true,
		},

		// ---- Comment edge cases ----
		// Block comment containing `>` chars — must not confuse closing
		// detection (closing scan looks at last `>` in element range).
		{Code: "<App\n  /* >>> comment >>> */\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},
		// Block comment spanning multiple lines, between two attrs.
		{Code: "<App\n  foo\n  /*\n   * long\n   * comment\n   */\n  bar\n/>", Tsx: true},
		// Multiple consecutive block comments before closing.
		{Code: "<App\n  foo\n  /* a *//* b */\n/>", Tsx: true},
		// Line comment with no content.
		{Code: "<App\n  //\n  foo\n/>", Tsx: true},

		// ---- Dimension 4: very long indent ----
		// Locks in that column accounting handles 80-col indent correctly.
		{
			Code: "                                                                                <App\n                                                                                  foo\n                                                                                />",
			Tsx:  true,
		},

		// ---- Dimension 4: mixed tab+space indent ----
		// `\t ` (tab then space) is the line-indent. UTF16Length counts each
		// as 1 unit, matching ESLint's loc.column. tag-aligned requires
		// closing at the `<` column (which is col 0 here).
		{Code: "<App\n\t foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned"}},

		// ---- Dimension 4: JSX containing a JsxFragment child ----
		{Code: "<App>\n  <>\n    <Child />\n  </>\n</App>", Tsx: true},

		// ---- Dimension 4: deeply nested mixed self-closing / paired ----
		{Code: "<A><B><C><D><E foo /></D></C></B></A>", Tsx: true},

		// ---- Dimension 4: attribute identifier in unicode ----
		{Code: `<App データ="x" />`, Tsx: true},
		{Code: "<App\n  データ=\"x\"\n/>", Tsx: true},

		// ---- Dimension 4: TS satisfies in attribute value ----
		{Code: `<App foo={x satisfies T} />`, Tsx: true},

		// ---- Dimension 4: TS readonly assertion / const assertion ----
		{Code: `<App foo={[1, 2] as const} />`, Tsx: true},

		// ---- Real-user: large props object as last attribute ----
		{
			Code: "<Component\n  config={{\n    a: 1,\n    b: 2,\n    c: 3,\n    d: { nested: true },\n    e: [1, 2, 3],\n  }}\n/>",
			Tsx:  true,
		},

		// ---- Real-user: JSX with key prop (React list pattern) ----
		{Code: "{items.map(item =>\n  <Item\n    key={item.id}\n    {...item}\n  />\n)}", Tsx: true},

		// ---- Real-user: render-prop component ----
		{
			Code: "<Consumer>\n  {value => (\n    <Display\n      value={value}\n    />\n  )}\n</Consumer>",
			Tsx:  true,
		},

		// ---- Options matrix: location is "after-props" (string form) ----
		{Code: `<App foo />`, Tsx: true, Options: "after-props"},
		// Options matrix: nested array `[[map]]` should not crash (rslint
		// loader collapses single-element arrays once; double-nested is
		// non-standard input that should degrade to defaults).
		{Code: `<App />`, Tsx: true, Options: []interface{}{[]interface{}{map[string]interface{}{"location": "props-aligned"}}}},
		// Options matrix: array with non-string non-map first element.
		{Code: `<App />`, Tsx: true, Options: []interface{}{42}},
		// Options matrix: array with nil first element.
		{Code: `<App />`, Tsx: true, Options: []interface{}{nil}},
		// Options matrix: map with extra unknown keys alongside valid ones.
		{Code: "<App\n  foo\n/>", Tsx: true, Options: map[string]interface{}{"location": "tag-aligned", "extraKey": "ignored"}},
	}, []rule_tester.InvalidTestCase{
		// ---- Locks in upstream's findTrailingComment + upgrade branch (after-tag) ----
		// Zero attributes, comment between tag-name and `/>`. Expected:
		// upgrade after-tag → line-aligned, fix from comment.end. Without
		// the upgrade, the fix would collapse `<App\n  //c\n  />` to
		// `<App //c />`, breaking the comment.
		// Note: with location=tag-aligned (default for `nonEmpty`/`selfClosing`)
		// the upgrade does NOT fire (because expectedLocation is `after-tag`,
		// not "tag-aligned"). The upgrade kicks in for `after-tag` only when
		// hasCorrectLocation is false. Here `after-tag` is wrong (closing on
		// line 4, tag on line 2), so it upgrades.
		{
			Code:   "<App\n  // comment\n  />",
			Tsx:    true,
			Output: []string{"<App\n  // comment\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 1)",
				Line:      3, Column: 3,
			}},
		},
		// Locks in upstream getExpectedLocation() arm 2: `openingLine ===
		// lastPropLastLine` → 'after-props'. Code: `<App foo\n  />`.
		// Default tag-aligned, but expectedLocation collapses to after-props
		// because the prop is on the opening line.
		{
			Code:   "<App foo\n  />",
			Tsx:    true,
			Output: []string{"<App foo/>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		// Locks in upstream getExpectedLocation() arm 3 self-closing branch:
		// `tokens.selfClosing ? selfClosing : nonEmpty`. With selfClosing
		// explicitly set, this routes to selfClosing's value.
		{
			Code:    "<App\n  foo\n  />",
			Tsx:     true,
			Output:  []string{"<App\n  foo\n/>"},
			Options: map[string]interface{}{"selfClosing": "tag-aligned", "nonEmpty": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
				Line:      3, Column: 3,
			}},
		},
		// Locks in upstream getExpectedLocation() arm 3 non-self-closing branch:
		// `tokens.selfClosing ? selfClosing : nonEmpty`. With nonEmpty
		// explicitly set, this routes to nonEmpty's value.
		{
			Code:    "<App\n  foo\n  ></App>",
			Tsx:     true,
			Output:  []string{"<App\n  foo\n></App>"},
			Options: map[string]interface{}{"selfClosing": "line-aligned", "nonEmpty": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
				Line:      3, Column: 3,
			}},
		},
		// ---- Dimension 4: JsxMemberExpression tag — column accounting uses `<` ----
		{
			Code:    "<Foo.Bar\n  foo />",
			Tsx:     true,
			Output:  []string{"<Foo.Bar\n  foo\n/>"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1 on the next line)",
				Line:      2, Column: 7,
			}},
		},
		// ---- Dimension 4: JsxNamespacedName tag, multi-line ----
		{
			Code:    "<svg:circle\n  cx={1} />",
			Tsx:     true,
			Output:  []string{"<svg:circle\n  cx={1}\n/>"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1 on the next line)",
				Line:      2, Column: 10,
			}},
		},
		// ---- Dimension 4: '>' inside attribute expression on a later line ----
		// Closing detection must pick the JSX-element '>' / '/' at the end,
		// not the comparison's `>`. tsgo's BinaryExpression `>` is part of
		// the attribute's range and is excluded by our trimmed scan.
		{
			Code:   "<App foo={x > y}\n/>",
			Tsx:    true,
			Output: []string{`<App foo={x > y}/>`},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the last prop",
			}},
		},
		// ---- Dimension 4: spread attribute as last prop ----
		// JsxSpreadAttribute is a different node kind from JsxAttribute; the
		// lastProp accounting must extract `{...rest}` correctly.
		{
			Code:    "<App\n  foo\n  {...rest}\n  />",
			Tsx:     true,
			Output:  []string{"<App\n  foo\n  {...rest}\n/>"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
				Line:      4, Column: 3,
			}},
		},
		// ---- Locks in trailing-comment upgrade vs after-tag (zero-prop case) ----
		// Two comments between tag-name and `/>`. expectedLocation=after-tag
		// (no props) but trailing comment → upgrade to line-aligned. Fix
		// rangeStart = lastComment.end (NOT tagNameEnd), so the comments stay.
		{
			Code:   "<input\n  // a\n  // b\n  />",
			Tsx:    true,
			Output: []string{"<input\n  // a\n  // b\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 1)",
				Line:      4, Column: 3,
			}},
		},
		// ---- Real-user: multi-byte UTF-16 column accounting ----
		// Opening line contains multi-byte identifier characters; the
		// "expected column N" in the message MUST be UTF-16 (matches IDE),
		// NOT byte offset.
		{
			Code:    "var x = (() => {\n  return <中文Comp\n    foo />\n})",
			Tsx:     true,
			Output:  []string{"var x = (() => {\n  return <中文Comp\n    foo\n  />\n})"},
			Options: map[string]interface{}{"location": "line-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				// `  return <中文Comp` is line 2; line-aligned indent = 2 spaces.
				// Expected column is 3 (1-based UTF-16).
				Message: "The closing bracket must be aligned with the line containing the opening tag (expected column 3 on the next line)",
				Line:    3, Column: 9,
			}},
		},
		// Tag-aligned with multi-byte chars before opening — expected column
		// is the UTF-16 position of '<', NOT byte offset.
		{
			Code:    "var 中文Var = <App\n  foo />",
			Tsx:     true,
			Output:  []string{"var 中文Var = <App\n  foo\n            />"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				// `var 中文Var = <App` — '<' at UTF-16 col 13 (0-based 12).
				Message: "The closing bracket must be aligned with the opening tag (expected column 13 on the next line)",
				Line:    2, Column: 7,
			}},
		},
		// ---- Real-user: after-tag with whitespace-only blank line between tag and `/>` ----
		// Blank line is whitespace only (no comment), so the trailing-comment
		// upgrade does NOT fire; expectedLocation stays `after-tag` and the
		// fix collapses to single-line.
		{
			Code:   "<App\n\n/>",
			Tsx:    true,
			Output: []string{`<App />`},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be placed after the opening tag",
			}},
		},
		// ---- Real-user: nested element column lock-in (3 levels) ----
		// Three-level nesting: outer `<App>`, middle `<B foo>`, innermost
		// `<C\n  bar\n   />`. Innermost props-aligned — column derives from
		// `bar`, not outer elements.
		{
			Code:    "<App>\n  <B foo>\n    <C\n      bar\n        />\n  </B>\n</App>",
			Tsx:     true,
			Output:  []string{"<App>\n  <B foo>\n    <C\n      bar\n      />\n  </B>\n</App>"},
			Options: map[string]interface{}{"location": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				// `bar` at UTF-16 col 7 (1-based) → "expected column 7".
				Message: "The closing bracket must be aligned with the last prop (expected column 7)",
				Line:    5, Column: 9,
			}},
		},
		// ---- Trailing block-comment specifically ----
		// Verifies `findTrailingComment` handles `/* ... */` and not just
		// `// ...`. After-props upgrade fires.
		{
			Code:    "<input\n  foo\n  /* trailing */\n  />",
			Tsx:     true,
			Output:  []string{"<input\n  foo\n  /* trailing */\n/>"},
			Options: map[string]interface{}{"location": "after-props"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the line containing the opening tag (expected column 1)",
				Line:      4, Column: 3,
			}},
		},
		// ---- Options 42 (numeric) falls back to default tag-aligned ----
		{
			Code:    "<App\n  foo\n  />",
			Tsx:     true,
			Options: 42,
			Output:  []string{"<App\n  foo\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
				Line:      3, Column: 3,
			}},
		},
		// Map with unknown key — defaults apply.
		{
			Code:    "<App\n  foo\n  />",
			Tsx:     true,
			Options: map[string]interface{}{"unknownKey": "props-aligned"},
			Output:  []string{"<App\n  foo\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
				Line:      3, Column: 3,
			}},
		},
		// ---- Dimension 4: regex literal in attribute value, multi-line outer ----
		// The `>` inside `/regex/` must not be picked up as closing.
		{
			Code:    "<App\n  pattern={/>/}\n  />",
			Tsx:     true,
			Output:  []string{"<App\n  pattern={/>/}\n/>"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
				Line:      3, Column: 3,
			}},
		},
		// ---- Dimension 4: template literal containing `>` ----
		{
			Code:    "<App\n  t={`<X>`}\n  />",
			Tsx:     true,
			Output:  []string{"<App\n  t={`<X>`}\n/>"},
			Options: map[string]interface{}{"location": "tag-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
				Line:      3, Column: 3,
			}},
		},
		// ---- Locks in mutual exclusion: location with nonEmpty present ----
		// `'location' in config!` short-circuits: nonEmpty is IGNORED when
		// location exists. Without that short-circuit, props-aligned would
		// apply and `</App>` at col 1 would be wrong. With the short-circuit,
		// tag-aligned applies; `</App>` at col 1 matches `<` col 1 → would
		// be valid. To exercise the override path being suppressed, we test
		// an actually-misaligned closing under location=props-aligned (and
		// nonEmpty="tag-aligned" which should NOT take effect).
		{
			Code:    "<App\n  foo\n  ></App>",
			Tsx:     true,
			Output:  []string{"<App\n  foo\n></App>"},
			Options: map[string]interface{}{"location": "tag-aligned", "nonEmpty": "props-aligned"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "bracketLocation",
				Message:   "The closing bracket must be aligned with the opening tag (expected column 1)",
				Line:      3, Column: 3,
			}},
		},
	})
}
