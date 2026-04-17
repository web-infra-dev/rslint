package no_unescaped_entities

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// entitySuggestions returns the expected suggestions for a default-entity
// diagnostic, given the source before/after the single char that will be
// replaced by each alternative.
func entitySuggestions(alts []string, before, after string) []rule_tester.InvalidTestCaseSuggestion {
	out := make([]rule_tester.InvalidTestCaseSuggestion, len(alts))
	for i, alt := range alts {
		out[i] = rule_tester.InvalidTestCaseSuggestion{
			MessageId: "replaceWithAlt",
			Output:    before + alt + after,
		}
	}
	return out
}

func apostropheSuggestions(before, after string) []rule_tester.InvalidTestCaseSuggestion {
	return entitySuggestions([]string{"&apos;", "&lsquo;", "&#39;", "&rsquo;"}, before, after)
}

func quoteSuggestions(before, after string) []rule_tester.InvalidTestCaseSuggestion {
	return entitySuggestions([]string{"&quot;", "&ldquo;", "&#34;", "&rdquo;"}, before, after)
}

func TestNoUnescapedEntitiesRule(t *testing.T) {
	// NOTE: TypeScript's JSX parser rejects unescaped `>` and `}` in JSX text
	// (unlike Babel/Acorn), so those cases cannot be exercised as test inputs
	// even though the rule logic supports them.
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnescapedEntitiesRule, []rule_tester.ValidTestCase{
		// --- structural: no JsxText is present ---
		{Code: `var Hello = <div/>`, Tsx: true},
		{Code: `var Hello = <div></div>`, Tsx: true},
		{Code: `var Hello = <></>`, Tsx: true},
		{Code: `var Hello = <Component name="don't" />`, Tsx: true},
		{Code: `var Hello = <div title="it's fine"></div>`, Tsx: true},

		// --- plain text / already-escaped ---
		{Code: `var Hello = <div>Here is some text!</div>`, Tsx: true},
		{Code: `var Hello = <div>I&rsquo;ve escaped some entities: &gt; &lt; &amp;</div>`, Tsx: true},
		{Code: "var Hello = <div>first line is ok\n            so is second\n            and here are some escaped entities: &gt; &lt; &amp;</div>", Tsx: true},

		// --- whitespace-only JsxText ---
		{Code: "var Hello = <div>   </div>", Tsx: true},
		{Code: "var Hello = <div>\n\n\t</div>", Tsx: true},

		// --- string literals inside expression containers are not JsxText ---
		{Code: `var Hello = <div>{">" + "<" + "&" + '"'}</div>`, Tsx: true},
		{Code: `var Hello = <>{">" + "<" + "&" + '"'}</>`, Tsx: true},
		{Code: `var Hello = <div>{"it's fine"}</div>`, Tsx: true},

		// --- fragments with safe text ---
		{Code: `var Hello = <>Here is some text!</>`, Tsx: true},
		{Code: `var Hello = <>I&rsquo;ve escaped some entities: &gt; &lt; &amp;</>`, Tsx: true},

		// --- nested JSX where inner text is escaped ---
		{Code: `var Hello = <Outer><Inner>safe text</Inner></Outer>`, Tsx: true},
		{Code: `var Hello = <ul><li>one</li><li>two</li></ul>`, Tsx: true},
		{Code: `var Hello = <div>{<span>nested</span>}</div>`, Tsx: true},

		// --- JSX in various expression positions ---
		{Code: `var Hello = [<div key="1">safe</div>]`, Tsx: true},
		{Code: `function f() { return <div>safe</div>; }`, Tsx: true},
		{Code: `var Hello = true ? <div>safe</div> : null`, Tsx: true},

		// --- JSX as prop value (inner JsxText is checked — escaped version is safe) ---
		{Code: `var Hello = <Outer child={<Inner>safe</Inner>} />`, Tsx: true},

		// --- custom forbid: disables defaults ---
		{Code: `var Hello = <div>don't forget</div>`, Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{"&"}}},

		// --- forbid: [] explicitly disables all entities ---
		{Code: `var Hello = <div>don't do that</div>`, Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{}}},

		// --- forbid entry with empty char string never matches ---
		{Code: `var Hello = <div>don't do that</div>`, Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{""}}},
	}, []rule_tester.InvalidTestCase{
		// --- default: bare apostrophe ---
		{
			Code: `var Hello = <div>'</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 18,
					Suggestions: apostropheSuggestions(`var Hello = <div>`, `</div>`),
				},
			},
		},
		// --- apostrophe in word ---
		{
			Code: `var Hello = <div>Don't do that</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 21,
					Suggestions: apostropheSuggestions(`var Hello = <div>Don`, `t do that</div>`),
				},
			},
		},
		// --- fragment with apostrophe ---
		{
			Code: `var Hello = <>it's a trap</>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 17,
					Suggestions: apostropheSuggestions(`var Hello = <>it`, `s a trap</>`),
				},
			},
		},
		// --- multi-line: multiple apostrophes on different lines ---
		{
			Code: "var Hello = <div>line one's\n            line two's</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 26,
					Suggestions: apostropheSuggestions("var Hello = <div>line one", "s\n            line two's</div>"),
				},
				{
					MessageId:   "unescapedEntityAlts",
					Line:        2, Column: 21,
					Suggestions: apostropheSuggestions("var Hello = <div>line one's\n            line two", "s</div>"),
				},
			},
		},
		// --- adjacent apostrophes ---
		{
			Code: `var Hello = <div>a''b</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 19,
					Suggestions: apostropheSuggestions(`var Hello = <div>a`, `'b</div>`),
				},
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 20,
					Suggestions: apostropheSuggestions(`var Hello = <div>a'`, `b</div>`),
				},
			},
		},
		// --- both `"` and `'` in same JsxText, reported in source order ---
		{
			Code: `var Hello = <div>"don't"</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 18,
					Suggestions: quoteSuggestions(`var Hello = <div>`, `don't"</div>`),
				},
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 22,
					Suggestions: apostropheSuggestions(`var Hello = <div>"don`, `t"</div>`),
				},
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 24,
					Suggestions: quoteSuggestions(`var Hello = <div>"don't`, `</div>`),
				},
			},
		},
		// --- script tag: two unescaped double quotes with precise positions ---
		{
			Code: `var Hello = <script>window.foo = "bar"</script>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 34,
					Suggestions: quoteSuggestions(`var Hello = <script>window.foo = `, `bar"</script>`),
				},
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 38,
					Suggestions: quoteSuggestions(`var Hello = <script>window.foo = "bar`, `</script>`),
				},
			},
		},
		// --- nested JSX: flag in inner element only ---
		{
			Code: `var Hello = <Outer>outer text<Inner>inner's</Inner></Outer>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unescapedEntityAlts",
					Line:      1, Column: 42,
					Suggestions: apostropheSuggestions(
						`var Hello = <Outer>outer text<Inner>inner`,
						`s</Inner></Outer>`,
					),
				},
			},
		},
		// --- nested JSX: flag at multiple nesting levels ---
		{
			Code: `var Hello = <Outer>outer's<Inner>inner's</Inner></Outer>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unescapedEntityAlts",
					Line:      1, Column: 25,
					Suggestions: apostropheSuggestions(
						`var Hello = <Outer>outer`,
						`s<Inner>inner's</Inner></Outer>`,
					),
				},
				{
					MessageId: "unescapedEntityAlts",
					Line:      1, Column: 39,
					Suggestions: apostropheSuggestions(
						`var Hello = <Outer>outer's<Inner>inner`,
						`s</Inner></Outer>`,
					),
				},
			},
		},
		// --- JSX inside expression container inside JSX ---
		{
			Code: `var Hello = <Outer>{<Inner>inner's</Inner>}</Outer>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unescapedEntityAlts",
					Line:      1, Column: 33,
					Suggestions: apostropheSuggestions(
						`var Hello = <Outer>{<Inner>inner`,
						`s</Inner>}</Outer>`,
					),
				},
			},
		},
		// --- JSX as prop value: inner JsxText is scanned ---
		{
			Code: `var Hello = <Outer child={<Inner>inner's</Inner>} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unescapedEntityAlts",
					Line:      1, Column: 39,
					Suggestions: apostropheSuggestions(
						`var Hello = <Outer child={<Inner>inner`,
						`s</Inner>} />`,
					),
				},
			},
		},
		// --- forbid: simple string form, no suggestions ---
		{
			Code:    `var Hello = <span>foo & bar</span>`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"&"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unescapedEntity", Line: 1, Column: 23},
			},
		},
		// --- forbid: object form with suggestions ---
		{
			Code: `var Hello = <span>foo & bar</span>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"forbid": []interface{}{
					map[string]interface{}{"char": "&", "alternatives": []interface{}{"&amp;"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unescapedEntityAlts",
					Line:      1, Column: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceWithAlt", Output: `var Hello = <span>foo &amp; bar</span>`},
					},
				},
			},
		},
		// --- forbid: mix of string + object entries ---
		{
			Code: `var Hello = <div>a & b $ c</div>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"forbid": []interface{}{
					"&",
					map[string]interface{}{"char": "$", "alternatives": []interface{}{"&#36;"}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unescapedEntity", Line: 1, Column: 20},
				{
					MessageId: "unescapedEntityAlts",
					Line:      1, Column: 24,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceWithAlt", Output: `var Hello = <div>a & b &#36; c</div>`},
					},
				},
			},
		},
		// --- forbid: object form with empty alternatives reports without suggestions ---
		{
			Code: `var Hello = <div>foo & bar</div>`,
			Tsx:  true,
			Options: map[string]interface{}{
				"forbid": []interface{}{
					map[string]interface{}{"char": "&", "alternatives": []interface{}{}},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unescapedEntity", Line: 1, Column: 22},
			},
		},
		// --- NUL (U+0000) as a custom forbid char: must not be confused with
		// the "unset rune" sentinel in the implementation. ---
		{
			Code:    "var Hello = <div>a\u0000b</div>",
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"\u0000"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unescapedEntity", Line: 1, Column: 19},
			},
		},
		// --- multi-byte Unicode custom forbid char (fullwidth apostrophe U+FF07) ---
		{
			Code:    "var Hello = <div>fullwidth\uff07quote</div>",
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"\uff07"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unescapedEntity", Line: 1, Column: 27},
			},
		},
		// --- UTF-16 column accuracy when a multi-byte char precedes the match ---
		// `é` (U+00E9) is 2 bytes in UTF-8 but 1 UTF-16 code unit, so the `'`
		// must still report as column 22 (1-based UTF-16), not a byte offset.
		{
			Code: "var Hello = <div>café's</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 22,
					Suggestions: apostropheSuggestions("var Hello = <div>café", "s</div>"),
				},
			},
		},
		// --- UTF-16 column accuracy for a surrogate pair (supra-BMP) char ---
		// `🚀` (U+1F680) is 4 bytes in UTF-8 and occupies 2 UTF-16 code units,
		// so the `'` must report as column 20 (17 leading chars + 2 for the
		// surrogate pair + 1 for the apostrophe itself).
		{
			Code: "var Hello = <div>🚀's</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 20,
					Suggestions: apostropheSuggestions("var Hello = <div>🚀", "s</div>"),
				},
			},
		},
		// --- JSX in a type assertion: inner JsxText still scanned ---
		{
			Code: `var Hello = (<div>a's</div>) as any`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 20,
					Suggestions: apostropheSuggestions(`var Hello = (<div>a`, `s</div>) as any`),
				},
			},
		},
		// --- JSX in logical short-circuit expression ---
		{
			Code: `var Hello = cond && <div>a's</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 27,
					Suggestions: apostropheSuggestions(`var Hello = cond && <div>a`, `s</div>`),
				},
			},
		},
		// --- ESLint parity: whitespace-only JsxText with a whitespace forbid char ---
		// An obscure configuration but aligned with ESLint's char-by-char scan.
		{
			Code:    `var Hello = <div>  </div>`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{" "}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unescapedEntity", Line: 1, Column: 18},
				{MessageId: "unescapedEntity", Line: 1, Column: 19},
			},
		},
		// --- two JsxText nodes split by an expression container in the same parent ---
		{
			Code: `var Hello = <div>a's {x} b's</div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 19,
					Suggestions: apostropheSuggestions(`var Hello = <div>a`, `s {x} b's</div>`),
				},
				{
					MessageId:   "unescapedEntityAlts",
					Line:        1, Column: 27,
					Suggestions: apostropheSuggestions(`var Hello = <div>a's {x} b`, `s</div>`),
				},
			},
		},
	})
}
