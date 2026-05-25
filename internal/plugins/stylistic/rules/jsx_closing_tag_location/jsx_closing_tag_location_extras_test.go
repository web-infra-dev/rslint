// TestJsxClosingTagLocationExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
//
// stylistic-specific deltas vs. react/jsx-closing-tag-location, locked in here:
//   - the autofix indent is always SPACES, even when the reference line is
//     tab-indented (react preserves the literal indentation under line-aligned);
//   - the "move to its own line" fix is a pure insertion that leaves any
//     whitespace already before the tag in place (react strips it);
//   - the option is string-only — a `{ location }` object is NOT honored
//     (react accepts it).
package jsx_closing_tag_location

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxClosingTagLocationExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxClosingTagLocationRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4 rows that do NOT apply to this rule ----
		// N/A: receiver / expression wrappers (paren, non-null `!`, `as`,
		//   `satisfies`, optional chain). The rule inspects only the opening /
		//   closing TAG positions, never a member-access receiver, so no child
		//   node that could be wrapped is ever read.
		// N/A: access / key forms (string / numeric / computed / private keys).
		//   The rule reads no property keys.
		// N/A: async / generator / declaration-vs-expression container variants.
		//   These are function/class shapes; the rule only sees JSX elements and
		//   fragments (their element-vs-fragment split IS covered below).

		// ---- Dimension 4: graceful degradation ----
		// Self-closing element has no closing tag — never visited, never reported.
		{Code: "<App\n  foo\n/>", Tsx: true},
		// Single-line element with an expression child — opening/closing share a
		// line. Locks in upstream `opening.line === node.line` early return.
		{Code: "<App>{cond ? a : b}</App>", Tsx: true},
		// Single-line fragment.
		{Code: "<>{x}</>", Tsx: true},
		// Empty multiline body, closing aligned with opening.
		{Code: "<App>\n</App>", Tsx: true},

		// ---- Dimension 4: container forms ----
		// JsxMemberExpression tag name — only the `<` position matters.
		{Code: "<Foo.Bar>\n  x\n</Foo.Bar>", Tsx: true},

		// ---- Real-user: multiline opening tag with attributes (docs example) ----
		{Code: "<Say\n  firstName=\"John\"\n  lastName=\"Smith\">\n  Hello\n</Say>", Tsx: true},
		// ---- Real-user: conditional render, paren-wrapped, properly aligned ----
		{Code: "<App>\n  {cond && (\n    <Foo>\n      bar\n    </Foo>\n  )}\n</App>", Tsx: true},

		// Locks in upstream column-match arm: when the closing tag shares its
		// line with content BUT lands on the target column, it is valid — the
		// column check returns before the first-in-line / messageId logic.
		// (upstream only tests this for line-aligned; this is the tag-aligned twin.)
		{Code: "  <App>\nab</App>", Tsx: true},

		// Locks in getIndentation() option branch: same source is VALID under
		// line-aligned (closing column 0 == opening line indent 0) but INVALID
		// under tag-aligned (see the matching invalid case below). Exercises the
		// bare-string option parse path.
		{Code: "var x = <App>\n  foo\n</App>", Tsx: true, Options: "line-aligned"},

		// ---- Real-user: list rendering via a .map() callback, all aligned ----
		{Code: "<ul>\n  {items.map(i => (\n    <li>\n      {i}\n    </li>\n  ))}\n</ul>", Tsx: true},
		// ---- Real-user: ternary rendering, both branches aligned ----
		{Code: "<div>\n  {cond ? (\n    <A>\n      x\n    </A>\n  ) : (\n    <B>\n      y\n    </B>\n  )}\n</div>", Tsx: true},
		// ---- Real-user: multi-line opening tag with several props; tag-aligned
		// closing aligns to the opening `<` column, not the `>` line. ----
		{Code: "const el = (\n  <Button\n    type=\"submit\"\n  >\n    Submit\n  </Button>\n)", Tsx: true},
		// ---- tsgo/UTF-16: multi-byte chars before the opening tag do NOT shift
		// line-aligned (which keys off the opening LINE indent, here 0). ----
		{Code: "const 名字 = <App>\n  foo\n</App>", Tsx: true, Options: "line-aligned"},
		// ---- Doc example (line-aligned, correct): closing aligns with the
		// opening LINE indent (0) even though the opening `<` is at column 12. ----
		{Code: "const App = <Bar>\n  Foo\n</Bar>", Tsx: true, Options: "line-aligned"},
	}, []rule_tester.InvalidTestCase{
		// ---- Layer 3: getIndentation() option branch (tag-aligned twin of the
		// line-aligned valid case above) — opening tag not at line start, so
		// tag-aligned (opening column 8) and line-aligned (line indent 0) diverge.
		{
			Code:   "var x = <App>\n  foo\n</App>",
			Tsx:    true,
			Output: []string{"var x = <App>\n  foo\n        </App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      3,
				Column:    1,
				EndLine:   3,
				EndColumn: 7,
			}},
		},

		// ---- Layer 3 (stylistic-specific): pure-insert fix preserves the
		// whitespace already before the tag. react strips the two spaces after
		// "foo"; stylistic's insertTextBefore leaves them. ----
		{
			Code:   "<App>\n  foo  </App>",
			Tsx:    true,
			Output: []string{"<App>\n  foo  \n</App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "onOwnLine",
				Line:      2,
				Column:    8,
				EndLine:   2,
				EndColumn: 14,
			}},
		},

		// ---- Layer 3 (stylistic-specific): tab-indented reference line, the
		// fix emits SPACES (one per leading whitespace char), not tabs. Under
		// line-aligned react would re-use the literal "\t". ----
		{
			Code:    "\t<App>\n\t\tfoo\n\t\t\t</App>",
			Tsx:     true,
			Options: []interface{}{"line-aligned"},
			Output:  []string{"\t<App>\n\t\tfoo\n </App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "alignWithOpening",
				Line:      3,
				Column:    4,
				EndLine:   3,
				EndColumn: 10,
			}},
		},

		// ---- Dimension 4: nesting boundary — only the misaligned INNER element
		// reports; the correctly-aligned outer element must not bleed in. ----
		{
			Code:   "<App>\n  <Inner>\n    x\n      </Inner>\n</App>",
			Tsx:    true,
			Output: []string{"<App>\n  <Inner>\n    x\n  </Inner>\n</App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      4,
				Column:    7,
				EndLine:   4,
				EndColumn: 15,
			}},
		},

		// ---- Dimension 4: element nested in a fragment — fragment closing is
		// aligned (valid), only the inner element reports. ----
		{
			Code:   "<>\n  <App>\n    x\n      </App>\n</>",
			Tsx:    true,
			Output: []string{"<>\n  <App>\n    x\n  </App>\n</>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      4,
				Column:    7,
				EndLine:   4,
				EndColumn: 13,
			}},
		},

		// ---- Dimension 4: container form — JsxMemberExpression tag name. ----
		{
			Code:   "<Foo.Bar>\n  x\n    </Foo.Bar>",
			Tsx:    true,
			Output: []string{"<Foo.Bar>\n  x\n</Foo.Bar>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      3,
				Column:    5,
				EndLine:   3,
				EndColumn: 15,
			}},
		},

		// ---- Real-user: multiline opening tag with attributes (docs example,
		// tag-aligned). ----
		{
			Code:   "<Say\n  firstName=\"John\"\n  lastName=\"Smith\">\n  Hello\n    </Say>",
			Tsx:    true,
			Output: []string{"<Say\n  firstName=\"John\"\n  lastName=\"Smith\">\n  Hello\n</Say>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      5,
				Column:    5,
				EndLine:   5,
				EndColumn: 11,
			}},
		},

		// ---- Real-user: paren-wrapped assignment, line-aligned (opening tag
		// indented 2, closing over-indented). ----
		{
			Code:    "var x = (\n  <App>\n    foo\n      </App>\n)",
			Tsx:     true,
			Options: []interface{}{"line-aligned"},
			Output:  []string{"var x = (\n  <App>\n    foo\n  </App>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "alignWithOpening",
				Line:      4,
				Column:    7,
				EndLine:   4,
				EndColumn: 13,
			}},
		},

		// ---- Dimension 4: graceful — unknown option string falls back to the
		// default (tag-aligned). Upstream's JSON schema rejects this value
		// before the rule runs; rslint has no schema validation, so the rule
		// must degrade rather than emit an empty messageId. ----
		{
			Code:    "<App>\n  foo\n    </App>",
			Tsx:     true,
			Options: "tag-align", // typo → not in enum → default
			Output:  []string{"<App>\n  foo\n</App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      3,
				Column:    5,
				EndLine:   3,
				EndColumn: 11,
			}},
		},

		// ---- Dimension 4: graceful — the `{ location }` object form is NOT in
		// stylistic's string-only schema, so it is ignored and the rule uses the
		// default tag-aligned (here flagging what line-aligned would have
		// accepted). This is the key option-shape divergence from the react rule,
		// which DOES honor `{ location }`. ----
		{
			Code:    "var x = <App>\n  foo\n</App>",
			Tsx:     true,
			Options: map[string]interface{}{"location": "line-aligned"},
			Output:  []string{"var x = <App>\n  foo\n        </App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      3,
				Column:    1,
				EndLine:   3,
				EndColumn: 7,
			}},
		},

		// ---- Layer 3 (faithfulness lock-in): a leading comma is CONTENT, not
		// whitespace, so the close is NOT first-in-line → onOwnLine, and the
		// pure-insert fix must leave the comma intact. reactutil.IsNodeFirstInLine
		// skips commas and would (a) mis-pick matchIndent and (b) DELETE the comma
		// via replaceTextRange — which is exactly why this rule computes
		// first-in-line itself instead of reusing that helper. ----
		{
			Code:   "<App>\n  ,</App>",
			Tsx:    true,
			Output: []string{"<App>\n  ,\n</App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "onOwnLine",
				Line:      2,
				Column:    4,
				EndLine:   2,
				EndColumn: 10,
			}},
		},

		// ---- Dimension 4: fragment nested in a fragment — only the misaligned
		// inner fragment reports; the aligned outer fragment must not bleed in. ----
		{
			Code:   "<>\n  <>\n    x\n      </>\n</>",
			Tsx:    true,
			Output: []string{"<>\n  <>\n    x\n  </>\n</>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      4,
				Column:    7,
				EndLine:   4,
				EndColumn: 10,
			}},
		},

		// ---- Dimension 4: 3-level deep nesting — only the innermost element is
		// misaligned; both correctly-aligned ancestors must stay silent. ----
		{
			Code:   "<A>\n  <B>\n    <C>\n      x\n        </C>\n  </B>\n</A>",
			Tsx:    true,
			Output: []string{"<A>\n  <B>\n    <C>\n      x\n    </C>\n  </B>\n</A>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      5,
				Column:    9,
				EndLine:   5,
				EndColumn: 13,
			}},
		},

		// ---- Dimension 4: container form — JsxNamespacedName tag (`<a:b>`).
		// The rule reads only the `<` position, so the tag-name shape is
		// irrelevant; this locks that in. ----
		{
			Code:   "<a:b>\n  x\n    </a:b>",
			Tsx:    true,
			Output: []string{"<a:b>\n  x\n</a:b>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      3,
				Column:    5,
				EndLine:   3,
				EndColumn: 11,
			}},
		},

		// ---- tsgo/UTF-16 lock-in: multi-byte chars on the opening line. The
		// tag-aligned target column AND the fix indent width are UTF-16 columns
		// (10 = "const 名 = "), not byte offsets (12). A byte-based impl would
		// emit 12 spaces and diverge from ESLint. ----
		{
			Code:   "const 名 = <App>\n  foo\n</App>",
			Tsx:    true,
			Output: []string{"const 名 = <App>\n  foo\n          </App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      3,
				Column:    1,
				EndLine:   3,
				EndColumn: 7,
			}},
		},

		// ---- tsgo/UTF-16 lock-in: multi-byte content on the closing line before
		// the tag — the reported column is UTF-16 (7), not the byte offset (9). ----
		{
			Code:   "<App>\n  名foo</App>",
			Tsx:    true,
			Output: []string{"<App>\n  名foo\n</App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "onOwnLine",
				Line:      2,
				Column:    7,
				EndLine:   2,
				EndColumn: 13,
			}},
		},

		// ---- Dimension 4 (line-ending lock-in): CRLF terminators must produce
		// the same line/column/fix as LF. ----
		{
			Code:   "<App>\r\n  foo\r\n      </App>",
			Tsx:    true,
			Output: []string{"<App>\r\n  foo\r\n</App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      3,
				Column:    7,
				EndLine:   3,
				EndColumn: 13,
			}},
		},

		// ---- Dimension 4: container form — TS generic component tag `<App<T>>`.
		// The opening `<` is the element start, not the `<` of the type args. ----
		{
			Code:   "<App<T>>\n  x\n    </App>",
			Tsx:    true,
			Output: []string{"<App<T>>\n  x\n</App>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      3,
				Column:    5,
				EndLine:   3,
				EndColumn: 11,
			}},
		},

		// ---- Real-user: JSX returned from a .map() callback, closing misaligned;
		// the outer <ul> stays aligned and silent. ----
		{
			Code:   "<ul>\n  {items.map(i =>\n    <li>\n      {i}\n      </li>\n  )}\n</ul>",
			Tsx:    true,
			Output: []string{"<ul>\n  {items.map(i =>\n    <li>\n      {i}\n    </li>\n  )}\n</ul>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      5,
				Column:    7,
				EndLine:   5,
				EndColumn: 12,
			}},
		},

		// ---- Real-user: JSX as an array element — the closing tag is followed by
		// a comma, which (being after the tag) must not affect first-in-line. ----
		{
			Code:   "[\n  <App>\n    x\n      </App>,\n]",
			Tsx:    true,
			Output: []string{"[\n  <App>\n    x\n  </App>,\n]"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "matchIndent",
				Line:      4,
				Column:    7,
				EndLine:   4,
				EndColumn: 13,
			}},
		},

		// ---- Real-user: multi-line opening tag, line-aligned. The closing keys
		// off the opening LINE indent (2), proving the `<` is taken from the
		// opening tag's line, not the `>` line. ----
		{
			Code:    "const el = (\n  <Button\n    type=\"submit\"\n    onClick={fn}\n  >\n    Submit\n      </Button>\n)",
			Tsx:     true,
			Options: []interface{}{"line-aligned"},
			Output:  []string{"const el = (\n  <Button\n    type=\"submit\"\n    onClick={fn}\n  >\n    Submit\n  </Button>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "alignWithOpening",
				Line:      7,
				Column:    7,
				EndLine:   7,
				EndColumn: 16,
			}},
		},

		// ---- Doc example (line-aligned, incorrect): mirrors the `.md` example;
		// the over-indented closing tag is re-aligned to the opening line indent. ----
		{
			Code:    "const App = <Bar>\n  Foo\n                </Bar>",
			Tsx:     true,
			Options: []interface{}{"line-aligned"},
			Output:  []string{"const App = <Bar>\n  Foo\n</Bar>"},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "alignWithOpening",
				Line:      3,
				Column:    17,
				EndLine:   3,
				EndColumn: 23,
			}},
		},
	})
}
