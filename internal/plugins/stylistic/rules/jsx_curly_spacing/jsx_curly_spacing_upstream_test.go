// TestJsxCurlySpacingUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/jsx-curly-spacing/
// jsx-curly-spacing.test.ts 1:1. Upstream asserts only messageId + token on
// its invalid cases; the Line/Column here are computed from the exact source
// each case carries (every report anchors at the single-character `{` or `}`
// brace token). rslint-specific lock-in cases live in
// jsx_curly_spacing_extras_test.go.
//
// The implementation is shared with react/jsx-curly-spacing via BuildRule; the
// two upstream rules are case-identical, so this suite mirrors the react port's
// upstream cases verbatim (same inputs, same positions).
package jsx_curly_spacing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	mlAttrBar             = "\n        <App foo={\n        bar\n        } />;\n      "
	mlAttrBarFixedNoSpace = "\n        <App foo={bar} />;\n      "
	mlAttrBarFixedAlways  = "\n        <App foo={ bar } />;\n      "

	mlAttrObj             = "\n        <App foo={\n        { bar: true, baz: true }\n        } />;\n      "
	mlAttrObjFixedNoSpace = "\n        <App foo={{ bar: true, baz: true }} />;\n      "
	mlAttrObjFixedAlways  = "\n        <App foo={ { bar: true, baz: true } } />;\n      "

	mlChildBar             = "\n        <App>{\n        bar\n        }</App>;\n      "
	mlChildBarFixedNoSpace = "\n        <App>{bar}</App>;\n      "
	mlChildBarFixedAlways  = "\n        <App>{ bar }</App>;\n      "

	mlChildObj = "\n        <App>{\n        { bar: true, baz: true }\n        }</App>;\n      "

	mlAttrSpread             = "\n        <App {\n        ...bar\n        } />;\n      "
	mlAttrSpreadFixedNoSpace = "\n        <App {...bar} />;\n      "
	mlAttrSpreadFixedAlways  = "\n        <App { ...bar } />;\n      "

	mlAttrBarAndSpread = "\n        <App foo={\n        bar\n        } {\n        ...bar\n        } />;\n      "

	mlAttrBarAndSpreadComma             = "\n        <App foo={\n        bar\n        } {\n        ...baz\n        } />;\n      "
	mlAttrBarAndSpreadCommaFixedNoSpace = "\n        <App foo={bar} {...baz} />;\n      "
	mlAttrBarAndSpreadCommaFixedAlways  = "\n        <App foo={ bar } { ...baz } />;\n      "

	mlChildBarAndBar             = "\n        <App>{\n        bar\n        } {\n        bar\n        }</App>;\n      "
	mlChildBarAndBaz             = "\n        <App>{\n        bar\n        } {\n        baz\n        }</App>;\n      "
	mlChildBarAndBazFixedNoSpace = "\n        <App>{bar} {baz}</App>;\n      "
	mlChildBarAndBazFixedAlways  = "\n        <App>{ bar } { baz }</App>;\n      "

	mlChildTemplate = "\n        <App>{`\n        text\n        `}</App>\n      "

	mlChildAppFooBaz  = "\n        <App foo={ 42 } { ...bar } baz={{ 4: 2 }}>\n        {foo} {{ bar: baz }}\n        </App>\n      "
	mlChildAppFooBaz2 = "\n        <App foo={42} {...bar} baz={ { 4: 2 } }>\n        {foo} { { bar: baz } }\n        </App>\n      "

	mlChildEmptyComment            = "\n        <App>\n        { /* comment 27 */ }\n        </App>;\n      "
	mlChildEmptyCommentFixedNever  = "\n        <App>\n        {/* comment 27 */}\n        </App>;\n      "
	mlChildEmptyCommentAlways      = "\n        <App>\n        {/* comment 28 */}\n        </App>;\n      "
	mlChildEmptyCommentAlwaysFixed = "\n        <App>\n        { /* comment 28 */ }\n        </App>;\n      "

	mlChildCommentNewline      = "\n        <App>\n        {/*comment29*/\n        }\n        </App>\n      "
	mlChildCommentNewlineFixed = "\n        <App>\n        {/*comment29*/}\n        </App>\n      "
	mlChildNewlineComment      = "\n        <App>\n        {\n        /*comment30*/}\n        </App>\n      "
	mlChildNewlineCommentFixed = "\n        <App>\n        {/*comment30*/}\n        </App>\n      "

	mlChildBarBazComments           = "\n        <App>{ /* comment 31 */\n        bar\n        } {\n        baz\n        /* comment 32 */ }</App>;\n      "
	mlChildBarBazCommentsFixedNever = "\n        <App>{/* comment 31 */\n        bar\n        } {\n        baz\n        /* comment 32 */}</App>;\n      "
	mlChildBarBazComments2          = "\n        <App>{/* comment 33 */\n        bar\n        } {\n        baz\n        /* comment 33 */}</App>;\n      "
	mlChildBarBazComments2Fixed     = "\n        <App>{ /* comment 33 */\n        bar\n        } {\n        baz\n        /* comment 33 */ }</App>;\n      "

	mlTernary      = "\n        <div className={ this.state.renderInfo ? \"infoPanel col-xs-12\" : \"unToggled col-xs-12\" } />\n      "
	mlTernaryFixed = "\n        <div className={this.state.renderInfo ? \"infoPanel col-xs-12\" : \"unToggled col-xs-12\"} />\n      "
)

// Helper aliases for option-shape clarity inside the suite.
type opts = map[string]interface{}
type spc = map[string]interface{}

func TestJsxCurlySpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxCurlySpacingRule, []rule_tester.ValidTestCase{
		// ---- Default config (when:never, attributes:true, children:false) ----
		{Code: `<App foo={bar} />;`, Tsx: true},
		{Code: `<App foo={bar}>{bar}</App>;`, Tsx: true},
		{Code: `<App foo={bar}>{ bar }</App>;`, Tsx: true},
		{Code: "\n        <App foo={\n        bar\n        }>\n        {bar}\n        </App>;\n      ", Tsx: true},
		{Code: `<App foo={{ bar: true, baz: true }}>{{ bar: true, baz: true }}</App>;`, Tsx: true},
		{Code: `<App foo={{ bar: true, baz: true }}>{ { bar: true, baz: true } }</App>;`, Tsx: true},
		{Code: "\n        <App foo={\n        { bar: true, baz: true }\n        } />;\n      ", Tsx: true},
		{Code: "\n        <App foo={\n        { bar: true, baz: true }\n        }>\n        {{ bar: true, baz: true }}\n        </App>;\n      ", Tsx: true},
		{Code: `<App>{ foo /* comment 1 */ }</App>`, Tsx: true},
		{Code: `<App>{ /* comment 1 */ foo }</App>`, Tsx: true},

		// ---- attributes: true ----
		{Code: `<App foo={bar} />;`, Tsx: true, Options: []interface{}{opts{"attributes": true}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"attributes": true}}},
		{Code: `<App foo={{ bar: true, baz: true }} />;`, Tsx: true, Options: []interface{}{opts{"attributes": true}}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{opts{"attributes": true}}},

		// ---- attributes: false (disables attribute brace check) ----
		{Code: `<App foo={bar} />;`, Tsx: true, Options: []interface{}{opts{"attributes": false}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"attributes": false}}},
		{Code: `<App foo={{ bar: true, baz: true }} />;`, Tsx: true, Options: []interface{}{opts{"attributes": false}}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{opts{"attributes": false}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": false}}},
		{Code: `<App foo={ { bar: true, baz: true } } />;`, Tsx: true, Options: []interface{}{opts{"attributes": false}}},

		// ---- children: true ----
		{Code: `<App>{bar}</App>;`, Tsx: true, Options: []interface{}{opts{"children": true}}},
		{Code: mlChildBar, Tsx: true, Options: []interface{}{opts{"children": true}}},
		{Code: `<App>{{ bar: true, baz: true }}</App>;`, Tsx: true, Options: []interface{}{opts{"children": true}}},
		{Code: mlChildObj, Tsx: true, Options: []interface{}{opts{"children": true}}},

		// ---- children: false (disables child brace check) ----
		{Code: `<App>{bar}</App>;`, Tsx: true, Options: []interface{}{opts{"children": false}}},
		{Code: mlChildBar, Tsx: true, Options: []interface{}{opts{"children": false}}},
		{Code: `<App>{{ bar: true, baz: true }}</App>;`, Tsx: true, Options: []interface{}{opts{"children": false}}},
		{Code: mlChildObj, Tsx: true, Options: []interface{}{opts{"children": false}}},
		{Code: `<App>{ bar }</App>;`, Tsx: true, Options: []interface{}{opts{"children": false}}},
		{Code: `<App>{ { bar: true, baz: true } }</App>;`, Tsx: true, Options: []interface{}{opts{"children": false}}},

		// ---- when:never (object form) ----
		{Code: `<App foo={bar} />;`, Tsx: true, Options: []interface{}{opts{"when": "never"}}},
		{Code: `<App foo={bar} />;`, Tsx: true, Options: []interface{}{opts{"when": "never", "allowMultiline": false}}},
		{Code: `<App foo={bar} />;`, Tsx: true, Options: []interface{}{opts{"when": "never", "allowMultiline": true}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"when": "never", "allowMultiline": true}}},
		{Code: `<App foo={{ bar: true, baz: true }} />;`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "never"}}}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "never"}}}},

		// ---- when:always (object form) ----
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"when": "always"}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"when": "always", "allowMultiline": false}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"when": "always", "allowMultiline": true}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"when": "always", "allowMultiline": true}}},
		{Code: `<App foo={{ bar: true, baz: true }} />;`, Tsx: true, Options: []interface{}{opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}}},

		// ---- attributes object: when overrides ----
		{Code: `<App foo={bar} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}}},
		{Code: `<App foo={{ bar:baz }} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: `<App foo={ {bar:baz} } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": true}}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never", "allowMultiline": true}}}},
		{Code: `<App foo={bar/* comment 2 */} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "spacing": spc{}}}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "spacing": spc{}}}}},
		{Code: `<App foo={{ bar: true, baz: true }} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}}}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}}}},

		// ---- children object: when overrides ----
		{Code: `<App>{bar}</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<App>{ bar }</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: `<App>{ bar }</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always", "allowMultiline": false}}}},
		{Code: `<App>{{ bar:baz }}</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: mlChildObj, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<App>{ {bar:baz} }</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: mlChildObj, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: mlChildBar, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: mlChildBar, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always", "allowMultiline": true}}}},
		{Code: mlChildBar, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: mlChildBar, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": true}}}},
		{Code: "\n        <App>{/* comment 3 */}</App>;\n      ", Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<App>{bar/* comment 4 */}</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<App>{ bar }</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always", "spacing": spc{}}}}},
		{Code: mlChildBar, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always", "spacing": spc{}}}}},
		{Code: `<App>{{ bar: true, baz: true }}</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}}}},
		{Code: mlChildObj, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}}}},

		// ---- Spread attribute ----
		{Code: `<App {...bar} />;`, Tsx: true},
		{Code: `<App {...bar} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: `<App { ...bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: `<App { ...bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}}},
		{Code: mlAttrSpread, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: mlAttrSpread, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": true}}}},
		{Code: mlAttrSpread, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: mlAttrSpread, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never", "allowMultiline": true}}}},
		{Code: `<App {...bar/* comment 5 */} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},

		// ---- Multiple attribute braces ----
		{Code: `<App foo={bar} {...baz} />;`, Tsx: true},
		{Code: `<App foo={bar} {...baz} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: `<App foo={ bar } { ...baz } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: `<App foo={ bar } { ...baz } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}}},
		{Code: `<App foo={{ bar:baz }} {...baz} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: `<App foo={ {bar:baz} } { ...baz } />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: mlAttrBarAndSpread, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: `<App foo={bar/* comment 6 */} {...baz/* comment 7 */} />;`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never"}}}},
		{Code: `<App foo={3} bar={ {a: 2} } />`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}}},

		// ---- Children with comments ----
		{Code: `<App>{bar/* comment 8 */}</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<App>{bar} {baz}</App>;`, Tsx: true},
		{Code: `<App>{bar} {baz}</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<App>{ bar } { baz }</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: `<App>{ bar } { baz }</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always", "allowMultiline": false}}}},
		{Code: `<App>{{ bar:baz }} {baz}</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<App>{ {bar:baz} } { baz }</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: mlChildBarAndBar, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: `<App>{bar/* comment 9 */} {baz/* comment 10 */}</App>;`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<App>{3} { {a: 2} }</App>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}}},

		// ---- attributes:always with default children=false (only attribute braces enforced) ----
		{Code: `<App foo={ bar }>{bar}</App>`, Tsx: true, Options: []interface{}{opts{"attributes": opts{"when": "always"}}}},
		{Code: mlChildAppFooBaz, Tsx: true, Options: []interface{}{opts{"when": "never", "attributes": opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}, "children": true}}},
		{Code: mlChildAppFooBaz2, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}, "attributes": true, "children": opts{"when": "never"}}}},
		{Code: mlChildAppFooBaz2, Tsx: true, Options: []interface{}{opts{"spacing": spc{"objectLiterals": "always"}, "attributes": opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}, "children": opts{"when": "never"}}}},

		// ---- String shorthand ----
		{Code: `<App foo={bar} />;`, Tsx: true, Options: []interface{}{"never"}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{"never", opts{"spacing": spc{"objectLiterals": "never"}}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{"always"}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{"always", opts{"allowMultiline": false}}},
		{Code: `<App foo={{ bar:baz }} />;`, Tsx: true, Options: []interface{}{"never"}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<App foo={ {bar:baz} } />;`, Tsx: true, Options: []interface{}{"always"}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{"always"}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{"always"}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{"never"}},
		{Code: "\n        <App>{/* comment 11 */}</App>;\n      ", Tsx: true, Options: []interface{}{"never"}},
		{Code: `<App foo={bar/* comment 12 */} />;`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{"always", opts{"spacing": spc{}}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{"always", opts{"spacing": spc{}}}},
		{Code: `<App foo={{ bar: true, baz: true }} />;`, Tsx: true, Options: []interface{}{"always", opts{"spacing": spc{"objectLiterals": "never"}}}},
		{Code: mlAttrBar, Tsx: true, Options: []interface{}{"always", opts{"allowMultiline": true}}},
		{Code: mlAttrObj, Tsx: true, Options: []interface{}{"always", opts{"spacing": spc{"objectLiterals": "never"}}}},

		// ---- Spread + string shorthand ----
		{Code: `<App {...bar} />;`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<App { ...bar } />;`, Tsx: true, Options: []interface{}{"always"}},
		{Code: `<App { ...bar } />;`, Tsx: true, Options: []interface{}{"always", opts{"allowMultiline": false}}},
		{Code: mlAttrSpread, Tsx: true, Options: []interface{}{"always"}},
		{Code: mlAttrSpread, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<App {...bar/* comment 13 */} />;`, Tsx: true, Options: []interface{}{"never"}},

		// ---- Multiple attribute braces + string shorthand ----
		{Code: `<App foo={bar} {...baz} />;`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<App foo={ bar } { ...baz } />;`, Tsx: true, Options: []interface{}{"always"}},
		{Code: `<App foo={ bar } { ...baz } />;`, Tsx: true, Options: []interface{}{"always", opts{"allowMultiline": false}}},
		{Code: `<App foo={{ bar:baz }} {...baz} />;`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<App foo={ {bar:baz} } { ...baz } />;`, Tsx: true, Options: []interface{}{"always"}},
		{Code: "\n        <App foo={\n        bar\n        } {\n        ...bar\n        }/>;\n      ", Tsx: true, Options: []interface{}{"always"}},
		{Code: `<App foo={bar/* comment 14 */} {...baz/* comment 15 */} />;`, Tsx: true, Options: []interface{}{"never"}},
		{Code: `<App foo={3} bar={ {a: 2} } />`, Tsx: true, Options: []interface{}{"never", opts{"spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={ bar }>{bar}</App>`, Tsx: true, Options: []interface{}{"always"}},

		// ---- children template literal that spans multiple lines ----
		{Code: mlChildTemplate, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}}},

		// ---- Fragment as parent of JsxExpression child ----
		{Code: `<>{bar} {baz}</>;`, Tsx: true},

		// ---- Arrow callback body with comment-only block ----
		{Code: `<div onLayout={() => { /* dummy callback to fix android bug with component measuring */ }} />`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Default config: attribute brace spacing reported, child untouched ----
		{
			Code:   `<App foo={ bar }>{bar}</App>;`,
			Tsx:    true,
			Output: []string{`<App foo={bar}>{bar}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:   `<App foo={ bar }>{ bar }</App>;`,
			Tsx:    true,
			Output: []string{`<App foo={bar}>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:   `<App foo={ { bar: true, baz: true } }>{{ bar: true, baz: true }}</App>;`,
			Tsx:    true,
			Output: []string{`<App foo={{ bar: true, baz: true }}>{{ bar: true, baz: true }}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 37},
			},
		},
		{
			Code:   `<App foo={ { bar: true, baz: true } }>{ { bar: true, baz: true } }</App>;`,
			Tsx:    true,
			Output: []string{`<App foo={{ bar: true, baz: true }}>{ { bar: true, baz: true } }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 37},
			},
		},

		// ---- attributes: true ----
		{
			Code:    `<App foo={ bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    `<App foo={ { bar: true, baz: true } } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true}},
			Output:  []string{`<App foo={{ bar: true, baz: true }} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 37},
			},
		},

		// ---- children: true ----
		{
			Code:    `<App>{ bar }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true}},
			Output:  []string{`<App>{bar}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 12},
			},
		},
		{
			Code:    `<>{ bar }</>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true}},
			Output:  []string{`<>{bar}</>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 3},
				{MessageId: "noSpaceBefore", Line: 1, Column: 9},
			},
		},
		{
			Code:    `<App>{ { bar: true, baz: true } }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true}},
			Output:  []string{`<App>{{ bar: true, baz: true }}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 33},
			},
		},

		// ---- when:never (object form) ----
		{
			Code:    `<App foo={ bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never"}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "allowMultiline": false}},
			Output:  []string{mlAttrBarFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={ { bar: true, baz: true } } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{`<App foo={{ bar: true, baz: true }} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 37},
			},
		},
		{
			Code:    mlAttrObj,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "allowMultiline": false, "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{mlAttrObjFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={{ bar: true, baz: true }} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={ { bar: true, baz: true } } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 35},
			},
		},
		{
			Code:    mlAttrObj,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "allowMultiline": false, "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{mlAttrObjFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},

		// ---- when:always (object form) ----
		{
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always"}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always", "allowMultiline": false}},
			Output:  []string{mlAttrBarFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={ { bar: true, baz: true } } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{`<App foo={{ bar: true, baz: true }} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 37},
			},
		},
		{
			Code:    mlAttrObj,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always", "allowMultiline": false, "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{mlAttrObjFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={{ bar: true, baz: true }} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={ { bar: true, baz: true } } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 35},
			},
		},
		{
			Code:    mlAttrObj,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always", "allowMultiline": false, "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{mlAttrObjFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},

		// ---- attributes:true + when ----
		{
			Code:    `<App foo={ bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true, "when": "never"}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true, "when": "never", "allowMultiline": false}},
			Output:  []string{mlAttrBarFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={ { bar: true, baz: true } } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true, "when": "never", "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{`<App foo={{ bar: true, baz: true }} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 37},
			},
		},
		{
			Code:    `<App foo={{ bar: true, baz: true }} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true, "when": "never", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={ { bar: true, baz: true } } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 35},
			},
		},
		{
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true, "when": "always"}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true, "when": "always", "allowMultiline": false}},
			Output:  []string{mlAttrBarFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={ { bar: true, baz: true } } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true, "when": "always", "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{`<App foo={{ bar: true, baz: true }} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 37},
			},
		},
		{
			Code:    `<App foo={{ bar: true, baz: true }} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": true, "when": "always", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={ { bar: true, baz: true } } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 35},
			},
		},

		// ---- attributes object: when ----
		{
			Code:    `<App foo={ bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    `<App foo={ bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    `<App foo={ bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    `<App foo={bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App foo={ bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App foo={bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{mlAttrBarFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{mlAttrBarFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "spacing": spc{}}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    `<App foo={ bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "spacing": spc{}}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    `<App foo={bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "spacing": spc{}}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App foo={ {bar: true, baz: true} } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}}},
			Output:  []string{`<App foo={{bar: true, baz: true}} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 35},
			},
		},

		// ---- children:true + when ----
		{
			Code:    `<App>{ bar }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true, "when": "never"}},
			Output:  []string{`<App>{bar}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 12},
			},
		},
		{
			Code:    mlChildBar,
			Tsx:     true,
			Options: []interface{}{opts{"children": true, "when": "never", "allowMultiline": false}},
			Output:  []string{mlChildBarFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App>{ { bar: true, baz: true } }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true, "when": "never", "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{`<App>{{ bar: true, baz: true }}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 33},
			},
		},
		{
			Code:    `<App>{{ bar: true, baz: true }}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true, "when": "never", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App>{ { bar: true, baz: true } }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 31},
			},
		},
		{
			Code:    `<App>{bar}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true, "when": "always"}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 10},
			},
		},
		{
			Code:    mlChildBar,
			Tsx:     true,
			Options: []interface{}{opts{"children": true, "when": "always", "allowMultiline": false}},
			Output:  []string{mlChildBarFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App>{ { bar: true, baz: true } }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true, "when": "always", "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{`<App>{{ bar: true, baz: true }}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 33},
			},
		},
		{
			Code:    `<App>{{ bar: true, baz: true }}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": true, "when": "always", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App>{ { bar: true, baz: true } }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 31},
			},
		},

		// ---- children object: when ----
		{
			Code:    `<App>{ bar }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never"}}},
			Output:  []string{`<App>{bar}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 12},
			},
		},
		{
			Code:    `<App>{ bar }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{`<App>{bar}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 12},
			},
		},
		{
			Code:    `<App>{bar}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App>{bar}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App>{ bar}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 11},
			},
		},
		{
			Code:    `<App>{bar }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<App>{ bar}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never"}}},
			Output:  []string{`<App>{bar}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<App>{bar }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never"}}},
			Output:  []string{`<App>{bar}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 11},
			},
		},
		{
			Code:    mlChildBar,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{mlChildBarFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlChildBar,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{mlChildBarFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App>{bar}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always", "spacing": spc{}}}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App>{ bar}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always", "spacing": spc{}}}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 11},
			},
		},
		{
			Code:    `<App>{bar }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always", "spacing": spc{}}}},
			Output:  []string{`<App>{ bar }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<App>{ {bar: true, baz: true} }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always", "spacing": spc{"objectLiterals": "never"}}}},
			Output:  []string{`<App>{{bar: true, baz: true}}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 31},
			},
		},

		// ---- Spread attribute ----
		{
			Code:    `<App { ...bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    `<App { ...bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    `<App {...bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 13},
			},
		},
		{
			Code:    `<App {...bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 13},
			},
		},
		{
			Code:    `<App { ...bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    `<App {...bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<App { ...bar} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<App {...bar } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    mlAttrSpread,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{mlAttrSpreadFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlAttrSpread,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{mlAttrSpreadFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},

		// ---- Multiple attribute braces ----
		{
			Code:    `<App foo={ bar } { ...baz } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App foo={bar} {...baz} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
				{MessageId: "noSpaceAfter", Line: 1, Column: 18},
				{MessageId: "noSpaceBefore", Line: 1, Column: 27},
			},
		},
		{
			Code:    `<App foo={ bar } { ...baz } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{`<App foo={bar} {...baz} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
				{MessageId: "noSpaceAfter", Line: 1, Column: 18},
				{MessageId: "noSpaceBefore", Line: 1, Column: 27},
			},
		},
		{
			Code:    `<App foo={bar} {...baz} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App foo={ bar } { ...baz } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 16},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 23},
			},
		},
		{
			Code:    `<App foo={bar} {...baz} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{`<App foo={ bar } { ...baz } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 16},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 23},
			},
		},
		{
			Code:    `<App foo={ bar} { ...baz} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App foo={ bar } { ...baz } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 15},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 25},
			},
		},
		{
			Code:    `<App foo={bar } {...baz } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App foo={ bar } { ...baz } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 17},
			},
		},
		{
			Code:    `<App foo={ bar} { ...baz} />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App foo={bar} {...baz} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceAfter", Line: 1, Column: 17},
			},
		},
		{
			Code:    `<App foo={bar } {...baz } />;`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never"}}},
			Output:  []string{`<App foo={bar} {...baz} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
				{MessageId: "noSpaceBefore", Line: 1, Column: 25},
			},
		},
		{
			Code:    mlAttrBarAndSpreadComma,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{mlAttrBarAndSpreadCommaFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
				{MessageId: "noNewlineAfter", Line: 4, Column: 11},
				{MessageId: "noNewlineBefore", Line: 6, Column: 9},
			},
		},
		{
			Code:    mlAttrBarAndSpreadComma,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{mlAttrBarAndSpreadCommaFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
				{MessageId: "noNewlineAfter", Line: 4, Column: 11},
				{MessageId: "noNewlineBefore", Line: 6, Column: 9},
			},
		},
		{
			Code:    `<App foo={ 3 } bar={{a: 2}} />`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
			Output:  []string{`<App foo={3} bar={ {a: 2} } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 14},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 20},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 27},
			},
		},

		// ---- Comments inside attribute braces ----
		{
			Code:   `<App foo={ foo /* comment 16 */ } />`,
			Tsx:    true,
			Output: []string{`<App foo={foo /* comment 16 */} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 33},
			},
		},
		{
			Code:    `<App foo={foo /* comment 17 */} />`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App foo={ foo /* comment 17 */ } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 31},
			},
		},
		{
			Code:   `<App foo={ /* comment 18 */ foo } />`,
			Tsx:    true,
			Output: []string{`<App foo={/* comment 18 */ foo} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 33},
			},
		},
		{
			Code:    `<App foo={/* comment 19 */ foo} />`,
			Tsx:     true,
			Options: []interface{}{opts{"attributes": opts{"when": "always"}}},
			Output:  []string{`<App foo={ /* comment 19 */ foo } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 31},
			},
		},

		// ---- Multiple child braces ----
		{
			Code:    `<App>{ bar } { baz }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never"}}},
			Output:  []string{`<App>{bar} {baz}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 12},
				{MessageId: "noSpaceAfter", Line: 1, Column: 14},
				{MessageId: "noSpaceBefore", Line: 1, Column: 20},
			},
		},
		{
			Code:    `<App>{ bar } { baz }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{`<App>{bar} {baz}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 12},
				{MessageId: "noSpaceAfter", Line: 1, Column: 14},
				{MessageId: "noSpaceBefore", Line: 1, Column: 20},
			},
		},
		{
			Code:    `<App>{bar} {baz}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ bar } { baz }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 10},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 12},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    `<App>{bar} {baz}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{`<App>{ bar } { baz }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 10},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 12},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    `<App>{ bar} { baz}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ bar } { baz }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 11},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 18},
			},
		},
		{
			Code:    `<App>{bar } {baz }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ bar } { baz }</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 13},
			},
		},
		{
			Code:    `<App>{ bar} { baz}</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never"}}},
			Output:  []string{`<App>{bar} {baz}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceAfter", Line: 1, Column: 13},
			},
		},
		{
			Code:    `<App>{bar } {baz }</App>;`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never"}}},
			Output:  []string{`<App>{bar} {baz}</App>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 11},
				{MessageId: "noSpaceBefore", Line: 1, Column: 18},
			},
		},
		{
			Code:    mlChildBarAndBaz,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{mlChildBarAndBazFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
				{MessageId: "noNewlineAfter", Line: 4, Column: 11},
				{MessageId: "noNewlineBefore", Line: 6, Column: 9},
			},
		},
		{
			Code:    mlChildBarAndBaz,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always", "allowMultiline": false}}},
			Output:  []string{mlChildBarAndBazFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
				{MessageId: "noNewlineAfter", Line: 4, Column: 11},
				{MessageId: "noNewlineBefore", Line: 6, Column: 9},
			},
		},
		{
			Code:    `<App>{ 3 } bar={{a: 2}}</App>`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
			Output:  []string{`<App>{3} bar={ {a: 2} }</App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 10},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 16},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 23},
			},
		},

		// ---- Children with comments ----
		{
			Code:    `<App>{foo /* comment 20 */}</App>`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ foo /* comment 20 */ }</App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 27},
			},
		},
		{
			Code:    `<App>{/* comment 21 */ foo}</App>`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ /* comment 21 */ foo }</App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 27},
			},
		},

		// ---- String shorthand ----
		{
			Code:    `<App foo={ bar } />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    `<App foo={ bar } />;`,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": false}},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
			},
		},
		{
			Code:    mlAttrObj,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": false, "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{mlAttrObjFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlAttrObj,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": false, "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{mlAttrObjFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlAttrObj,
			Tsx:     true,
			Options: []interface{}{"always", opts{"allowMultiline": false, "spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{mlAttrObjFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlAttrObj,
			Tsx:     true,
			Options: []interface{}{"always", opts{"allowMultiline": false, "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{mlAttrObjFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{"always", opts{"allowMultiline": false}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    `<App foo={ bar} />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    `<App foo={bar } />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App foo={ bar} />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App foo={bar } />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App foo={bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": false}},
			Output:  []string{mlAttrBarFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlAttrBar,
			Tsx:     true,
			Options: []interface{}{"always", opts{"allowMultiline": false}},
			Output:  []string{mlAttrBarFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    `<App foo={bar} />;`,
			Tsx:     true,
			Options: []interface{}{"always", opts{"spacing": spc{}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    `<App foo={ bar} />;`,
			Tsx:     true,
			Options: []interface{}{"always", opts{"spacing": spc{}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    `<App foo={bar } />;`,
			Tsx:     true,
			Options: []interface{}{"always", opts{"spacing": spc{}}},
			Output:  []string{`<App foo={ bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
			},
		},
		{
			Code:    `<App foo={ {bar: true, baz: true} } />;`,
			Tsx:     true,
			Options: []interface{}{"always", opts{"spacing": spc{"objectLiterals": "never"}}},
			Output:  []string{`<App foo={{bar: true, baz: true}} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 35},
			},
		},

		// ---- Spread + string shorthand ----
		{
			Code:    `<App { ...bar } />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    `<App { ...bar } />;`,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": false}},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
			},
		},
		{
			Code:    `<App {...bar} />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 13},
			},
		},
		{
			Code:    `<App {...bar} />;`,
			Tsx:     true,
			Options: []interface{}{"always", opts{"allowMultiline": false}},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 13},
			},
		},
		{
			Code:    `<App { ...bar} />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    `<App {...bar } />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App { ...bar } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<App { ...bar} />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<App {...bar } />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App {...bar} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 14},
			},
		},
		{
			Code:    mlAttrSpread,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": false}},
			Output:  []string{mlAttrSpreadFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlAttrSpread,
			Tsx:     true,
			Options: []interface{}{"always", opts{"allowMultiline": false}},
			Output:  []string{mlAttrSpreadFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 14},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},

		// ---- Multiple attribute braces + string shorthand ----
		{
			Code:    `<App foo={ bar } { ...baz } />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App foo={bar} {...baz} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
				{MessageId: "noSpaceAfter", Line: 1, Column: 18},
				{MessageId: "noSpaceBefore", Line: 1, Column: 27},
			},
		},
		{
			Code:    `<App foo={ bar } { ...baz } />;`,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": false}},
			Output:  []string{`<App foo={bar} {...baz} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 16},
				{MessageId: "noSpaceAfter", Line: 1, Column: 18},
				{MessageId: "noSpaceBefore", Line: 1, Column: 27},
			},
		},
		{
			Code:    `<App foo={bar} {...baz} />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ bar } { ...baz } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 16},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 23},
			},
		},
		{
			Code:    `<App foo={bar} {...baz} />;`,
			Tsx:     true,
			Options: []interface{}{"always", opts{"allowMultiline": false}},
			Output:  []string{`<App foo={ bar } { ...baz } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 14},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 16},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 23},
			},
		},
		{
			Code:    `<App foo={ bar} { ...baz} />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ bar } { ...baz } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededBefore", Line: 1, Column: 15},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 25},
			},
		},
		{
			Code:    `<App foo={bar } {...baz } />;`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ bar } { ...baz } />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 17},
			},
		},
		{
			Code:    `<App foo={ bar} { ...baz} />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App foo={bar} {...baz} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceAfter", Line: 1, Column: 17},
			},
		},
		{
			Code:    `<App foo={bar } {...baz } />;`,
			Tsx:     true,
			Options: []interface{}{"never"},
			Output:  []string{`<App foo={bar} {...baz} />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
				{MessageId: "noSpaceBefore", Line: 1, Column: 25},
			},
		},
		{
			Code:    mlAttrBarAndSpreadComma,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": false}},
			Output:  []string{mlAttrBarAndSpreadCommaFixedNoSpace},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
				{MessageId: "noNewlineAfter", Line: 4, Column: 11},
				{MessageId: "noNewlineBefore", Line: 6, Column: 9},
			},
		},
		{
			Code:    mlAttrBarAndSpreadComma,
			Tsx:     true,
			Options: []interface{}{"always", opts{"allowMultiline": false}},
			Output:  []string{mlAttrBarAndSpreadCommaFixedAlways},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 2, Column: 18},
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
				{MessageId: "noNewlineAfter", Line: 4, Column: 11},
				{MessageId: "noNewlineBefore", Line: 6, Column: 9},
			},
		},
		{
			Code:    `<App foo={ 3 } bar={{a: 2}} />`,
			Tsx:     true,
			Options: []interface{}{"never", opts{"spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={3} bar={ {a: 2} } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 14},
				{MessageId: "spaceNeededAfter", Line: 1, Column: 20},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 27},
			},
		},
		{
			Code:    `<App foo={foo /* comment 22 */} />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ foo /* comment 22 */ } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 31},
			},
		},
		{
			Code:    `<App foo={/* comment 23 */ foo} />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ /* comment 23 */ foo } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 31},
			},
		},

		// ---- Empty / comment-only braces ----
		{
			Code:    `<App>{/*comment24*/ }</App>`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never"}}},
			Output:  []string{`<App>{/*comment24*/}</App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},
		{
			Code:    `<App>{ /*comment25*/}</App>`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never"}}},
			Output:  []string{`<App>{/*comment25*/}</App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 6},
			},
		},
		{
			Code:    `<App>{/*comment26*/}</App>`,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "always"}}},
			Output:  []string{`<App>{ /*comment26*/ }</App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 6},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 20},
			},
		},
		{
			Code:    mlChildEmptyComment,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "children": true}},
			Output:  []string{mlChildEmptyCommentFixedNever},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 3, Column: 9},
				{MessageId: "noSpaceBefore", Line: 3, Column: 28},
			},
		},
		{
			Code:    mlChildEmptyCommentAlways,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always", "children": true}},
			Output:  []string{mlChildEmptyCommentAlwaysFixed},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 3, Column: 9},
				{MessageId: "spaceNeededBefore", Line: 3, Column: 26},
			},
		},
		{
			Code:    mlChildCommentNewline,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{mlChildCommentNewlineFixed},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineBefore", Line: 4, Column: 9},
			},
		},
		{
			Code:    mlChildNewlineComment,
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{mlChildNewlineCommentFixed},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineAfter", Line: 3, Column: 9},
			},
		},
		{
			Code:    mlChildBarBazComments,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "children": true}},
			Output:  []string{mlChildBarBazCommentsFixedNever},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 2, Column: 14},
				{MessageId: "noSpaceBefore", Line: 6, Column: 26},
			},
		},
		{
			Code:    mlChildBarBazComments2,
			Tsx:     true,
			Options: []interface{}{opts{"when": "always", "children": true}},
			Output:  []string{mlChildBarBazComments2Fixed},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 2, Column: 14},
				{MessageId: "spaceNeededBefore", Line: 6, Column: 25},
			},
		},

		// ---- Single-line ternary attribute (allowMultiline does not apply) ----
		{
			Code:    mlTernary,
			Tsx:     true,
			Options: []interface{}{"never", opts{"allowMultiline": true}},
			Output:  []string{mlTernaryFixed},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 2, Column: 24},
				{MessageId: "noSpaceBefore", Line: 2, Column: 96},
			},
		},
	})
}
