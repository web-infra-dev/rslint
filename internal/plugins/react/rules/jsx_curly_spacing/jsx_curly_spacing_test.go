package jsx_curly_spacing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Multi-line code blocks reused across both valid and invalid suites.
// Indentation matches upstream's template-literal formatting verbatim so the
// reported (line, column) numbers stay aligned with the upstream source.
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

func TestJsxCurlySpacingRule(t *testing.T) {
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

		// ===== Dimension 4 lock-in: tsgo-specific shapes & real-world =====
		// These tests are not in the upstream suite. They exercise AST shapes
		// and content patterns that are easy to break in a port (string
		// content, regex literals, TS-only operators, parenthesized wrappers,
		// nested containers). The contract is: rules of brace-spacing depend
		// ONLY on the source text immediately surrounding `{` and `}` — the
		// container's inner expression is opaque.

		// ---- TS expression wrappers — should not trigger "object literal" ----
		{Code: `<App foo={(bar)} />`, Tsx: true},             // ParenthesizedExpression
		{Code: `<App foo={((bar))} />`, Tsx: true},           // multi-level paren
		{Code: `<App foo={bar!} />`, Tsx: true},              // non-null assertion
		{Code: `<App foo={bar as any} />`, Tsx: true},        // type assertion
		{Code: `<App foo={bar satisfies Foo} />`, Tsx: true}, // satisfies
		{Code: `<App foo={bar?.baz} />`, Tsx: true},          // optional chain
		{Code: `<App foo={bar?.()} />`, Tsx: true},           // optional call
		{Code: `<App foo={(bar)} />`, Tsx: true, Options: []interface{}{opts{"when": "never"}}},
		{Code: `<App foo={ (bar) } />`, Tsx: true, Options: []interface{}{opts{"when": "always"}}},

		// ---- Non-`{` second tokens — must use `when`, NOT objectLiteralSpaces ----
		{Code: `<App foo={[1, 2]} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={<Bar />} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={() => {}} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={typeof bar} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={!bar} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={-1} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},
		{Code: `<App foo={1 + 2} />`, Tsx: true, Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}}},

		// ---- Strings / templates / regex containing trivia-like text ----
		// These exercise the scanner-based body scan: the inner content
		// includes characters that a naive byte-scanner could mistake for
		// comments or unbalanced braces.
		{Code: `<App foo={"// not a comment"} />`, Tsx: true},
		{Code: `<App foo={"/* not a comment */"} />`, Tsx: true},
		{Code: `<App foo={"{ not an object }"} />`, Tsx: true},
		{Code: `<App foo={'single \'quote\''} />`, Tsx: true},
		{Code: "<App foo={`tpl ${a} ${b}`} />", Tsx: true},
		{Code: "<App foo={`with } brace inside`} />", Tsx: true},
		{Code: `<App foo={/regex/g} />`, Tsx: true},
		{Code: `<App foo={ /regex/g } />`, Tsx: true, Options: []interface{}{opts{"when": "always"}}},

		// ---- Nested containers ----
		// `children` defaults to false: only inner attribute container is checked.
		{Code: `<App><Bar foo={baz} /></App>`, Tsx: true},
		// fragment + element + attribute container — all clean.
		{Code: `<><Bar foo={baz} /></>`, Tsx: true},
		// nested fragment.
		{Code: `<><></></>`, Tsx: true},
		// attribute container with JSX expression inside.
		{Code: `<App foo={<Bar baz={qux} />} />`, Tsx: true},
		// children container with nested element with attribute container.
		{Code: `<App>{<Bar foo={baz} />}</App>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},

		// ---- Spread variants ----
		// Children spread (`<App>{...arr}</App>`) is JSXSpreadChild in ESTree
		// and NOT covered by upstream's listener — verify our skip keeps the
		// rule silent across all configs, including ones that would otherwise
		// fire on the surrounding spaces.
		{Code: `<Foo>{...arr}</Foo>`, Tsx: true},
		{Code: `<Foo>{...arr}</Foo>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<Foo>{ ...arr }</Foo>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		{Code: `<Foo>{ ...arr }</Foo>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: `<Foo>{...arr}</Foo>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "always"}}}},
		{Code: `<App {...obj1} {...obj2} />`, Tsx: true},
		// Spread of a parenthesized object — second token is `.`, NOT `{`.
		{Code: `<App {...({a: 1, ...rest})} />`, Tsx: true},

		// ---- Real-world idioms ----
		{Code: `<div className={isActive ? 'active' : 'inactive'} />`, Tsx: true},
		{Code: `<button onClick={() => handleClick()}>Click</button>`, Tsx: true},
		{Code: `<App data-x={JSON.stringify({a: 1})} />`, Tsx: true},
		{Code: `<List items={items.map((item) => <Item key={item.id} />)} />`, Tsx: true},
		// Generic call inside attribute container.
		{Code: `<App foo={callMe<T>()} />`, Tsx: true},
		// Object literal as attribute (default never spacing for objectLiterals via lastPass).
		{Code: `<div style={{color: 'red'}} />`, Tsx: true},
		// Line comment immediately before `}` on the next line — allowed
		// under the default allowMultiline=true. Verifies the scanner-based
		// body scan correctly recognizes the comment as the previous
		// "thing" (so the multiline detection / reporting path agrees with
		// upstream's getTokenBefore({includeComments:true})).
		{Code: "<App>{foo // c\n}</App>", Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		// Same shape with a BLOCK comment immediately before `}` on the
		// next line — also allowed by default.
		{Code: "<App>{foo /* c */\n}</App>", Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		// Empty container `{}` and whitespace-only `{ }` — `never` accepts
		// both `{}` (no inner space) and same with `always` accepts `{ }`
		// (a single space). Locks in the empty-body branch where
		// `secondPos == innerHigh` and `isObjectLiteral` is correctly false.
		{Code: `<App>{}</App>`, Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},
		// Triple-nested attribute containers — listener triggers
		// independently for each JsxExpression, regardless of depth.
		{Code: `<Outer foo={<Mid bar={<Inner baz={qux} />} />} />`, Tsx: true},
		// Mixed normal + spread + normal attributes (ESLint passes them
		// through three separate listeners; we use one callback that runs
		// per node in source order).
		{Code: `<App foo={bar} {...spread} qux={quux} />`, Tsx: true},
		// CRLF line endings — newline detection must accept `\r\n`.
		{Code: "<App foo={\r\n  bar\r\n} />", Tsx: true},

		// ===== Robustness: complex inner contents under multi-line `always` =====
		// All of the following are multi-line attribute containers under
		// `always` mode with the default `allowMultiline: true`. The body
		// contains characters that a token-level scan would mis-classify
		// (template `}`, regex `}`, string `}`, nested template, nested
		// object literal `}`, JSX in body). The rule must remain SILENT —
		// brace spacing only depends on trivia immediately around the
		// outer `{` / `}`.

		// regex literal containing `}` in a character class.
		{Code: "<X foo={\n  /[}]+/g\n} />", Tsx: true, Options: []interface{}{"always"}},
		// regex literal containing escaped slash and `}`.
		{Code: "<X foo={\n  /a\\/b}c/g\n} />", Tsx: true, Options: []interface{}{"always"}},
		// String literal containing `}`.
		{Code: "<X foo={\n  \"hello}world\"\n} />", Tsx: true, Options: []interface{}{"always"}},
		// String literal containing `{` and `}`.
		{Code: "<X foo={\n  \"{ not an object }\"\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Nested template literal: outer template's substitution itself contains a template.
		{Code: "<X foo={\n  `outer ${`inner ${x}`}`\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Multiple `${ }` substitutions in one template (the regression shape).
		{Code: "<X foo={\n  `${a}${b}${c}`\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Arrow function body with block statement (raw `{` `}` inside).
		{Code: "<X foo={\n  () => { return 1; }\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Multi-line object literal as body (not "always" promotion of objectLiterals — body has surrounding newlines).
		{Code: "<X foo={\n  {\n    a: 1,\n    b: 2\n  }\n} />", Tsx: true, Options: []interface{}{"always", opts{"spacing": spc{"objectLiterals": "never"}}}},
		// Conditional with two object literals.
		{Code: "<X foo={\n  cond ? { a: 1 } : { b: 2 }\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Fragment in body.
		{Code: "<X foo={\n  <>{inner}</>\n} />", Tsx: true, Options: []interface{}{"always"}},
		// arr.map returning JSX with nested attribute container — inner
		// `{ x.id }` carries spaces so the whole tree is `always`-clean.
		{Code: "<X foo={\n  arr.map((x) => <Item key={ x.id } />)\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Tagged template with substitutions.
		{Code: "<X foo={\n  tag`hello ${name}`\n} />", Tsx: true, Options: []interface{}{"always"}},
		// Class expression body.
		{Code: "<X foo={\n  class { method() { return 1; } }\n} />", Tsx: true, Options: []interface{}{"always"}},

		// ===== Same-range edge cases (empty body with whitespace) =====
		// `<App>{ }</App>` body is whitespace-only — both noSpaceAfter and
		// noSpaceBefore fixes target the same range; rule_tester must apply
		// them without conflict.
		// Skipped: rule_tester rejects overlapping autofix ranges, even
		// when both replace the same span with the same content. The rule
		// itself does the right thing semantically (matches upstream's
		// double-report on `{ }`); we just can't assert the autofix
		// output through this harness when both fixes point at the same
		// span. Documented and asserted via the no-fix invalid case below.

		// A single tab between `{` and `}` should report just like a single space.
		{Code: "<X>{}</X>", Tsx: true, Options: []interface{}{opts{"children": opts{"when": "never"}}}},

		// ===== BOM + Unicode lock-in =====
		// rslint reports columns as 1-based UTF-16 character offsets (via
		// scanner.GetECMALineAndUTF16CharacterOfPosition), matching ESLint.
		// These tests lock that behaviour against three classes of source:
		//   1. BOM byte at file start — must NOT be counted as a column,
		//   2. BMP non-ASCII identifiers/strings/comments (e.g. `中`) —
		//      must each count as exactly 1 UTF-16 character,
		//   3. SMP characters represented by UTF-16 surrogate pairs (e.g.
		//      `🚀`) — must each count as 2 UTF-16 characters.

		// BOM at file start: column for `{` and `}` should match the
		// non-BOM equivalent.
		{Code: "\uFEFF<App foo={bar} />", Tsx: true},
		// BMP non-ASCII identifier inside braces — no surrounding spaces, valid under default never.
		{Code: `<App foo={中文} />`, Tsx: true},
		// String literal containing non-ASCII characters.
		{Code: `<App foo={"中文 with spaces"} />`, Tsx: true},
		// Block comment with non-ASCII text — both sides flush, valid.
		{Code: `<App foo={/* 中文注释 */ bar} />`, Tsx: true},
		// Single emoji (SMP, UTF-16 surrogate pair) inside braces.
		{Code: `<App foo={"🚀"} />`, Tsx: true},
		// `always` mode with surrounding spaces and non-ASCII content — valid.
		{Code: `<App foo={ 中文 } />`, Tsx: true, Options: []interface{}{"always"}},
		{Code: `<App foo={ "🚀" } />`, Tsx: true, Options: []interface{}{"always"}},

		// ===== Unicode WhiteSpace + LineTerminator (ECMAScript §12.2/§12.3) =====
		// NBSP (U+00A0) counts as ECMAScript WhiteSpace → satisfies `always`,
		// triggers `never` extra. Parity with ESLint verified via local probe.
		{Code: "<App foo={\u00A0bar\u00A0} />", Tsx: true, Options: []interface{}{"always"}},
		// LS (U+2028) / PS (U+2029) count as LineTerminator → cross-line
		// short-circuit on that side. Both modes valid when the OTHER side
		// hugs the brace.
		{Code: "<App foo={bar\u2028} />", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<App foo={\u2028bar} />", Tsx: true, Options: []interface{}{"never"}},
		{Code: "<App foo={bar\u2029} />", Tsx: true, Options: []interface{}{"never"}},
		// Template literal with `${ … }` substitutions — earlier impl using
		// the bare tsgo Scanner without parser context mis-tokenized the
		// closing `}` of a substitution as a real `}` token, corrupting
		// `penultimateEnd` for any enclosing multi-line attribute container.
		// Locks in: a multi-line `always` attribute whose body contains a
		// template with substitutions must not falsely report
		// `spaceNeededBefore` on the closing brace.
		{
			Code:    "<App\n  foo={\n    a + `/${b}${c}`\n  }\n/>",
			Tsx:     true,
			Options: []interface{}{"always"},
		},
		{
			Code:    "<App\n  foo={\n    `${x}-${y}-${z}`\n  }\n/>",
			Tsx:     true,
			Options: []interface{}{"always"},
		},
		// Nested attribute container whose body is a JSX subtree that
		// itself contains template-literal attributes — the same shape
		// that triggered the regression in rsbuild website/theme/index.tsx.
		// Inner `href={ ... }` carries surrounding spaces so it complies
		// with `always` on its own; the test asserts the OUTER multi-line
		// `beforeNav={ ... }` doesn't falsely report due to template `}`s.
		{
			Code:    "<Outer\n  beforeNav={\n    cond ? (\n      <Inner href={ `/${pre}${post}` } />\n    ) : null\n  }\n/>",
			Tsx:     true,
			Options: []interface{}{"always"},
		},

		// ---- JS-truthy semantics on schema-invalid `attributes` / `children` ----
		// ESLint's JSON-schema validator rejects these before the rule runs;
		// rslint does not validate schemas, so the rule must reproduce JS's
		// `value ? cfg : null` semantics (0/null/"" → disabled).
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": nil}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": 0}}},
		{Code: `<App foo={ bar } />;`, Tsx: true, Options: []interface{}{opts{"attributes": ""}}},
		{Code: `<App>{ bar }</App>;`, Tsx: true, Options: []interface{}{opts{"children": nil}}},
		{Code: `<App>{ bar }</App>;`, Tsx: true, Options: []interface{}{opts{"children": 0}}},
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

		// ===== Dimension 4 invalid: tsgo-specific shapes & real-world =====

		// ---- TS expression wrappers — default `never` reports the outer brace ----
		{
			Code:   `<App foo={ (bar) } />`,
			Tsx:    true,
			Output: []string{`<App foo={(bar)} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 18},
			},
		},
		{
			Code:   `<App foo={ bar! } />`,
			Tsx:    true,
			Output: []string{`<App foo={bar!} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 17},
			},
		},
		{
			Code:   `<App foo={ bar as any } />`,
			Tsx:    true,
			Output: []string{`<App foo={bar as any} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 23},
			},
		},
		{
			Code:   `<App foo={ bar satisfies Foo } />`,
			Tsx:    true,
			Output: []string{`<App foo={bar satisfies Foo} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 30},
			},
		},
		{
			Code:   `<App foo={ bar?.baz } />`,
			Tsx:    true,
			Output: []string{`<App foo={bar?.baz} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},

		// ---- Non-`{` second tokens — `objectLiterals: 'always'` must NOT promote them ----
		// Confirms isObjectLiteral correctly checks the literal `{` character,
		// not "is the inner expression a JS object". `[1,2]` is an array, not
		// an object, so `objectLiterals` config does not apply.
		{
			Code:    `<App foo={ [1, 2] } />`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={[1, 2]} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 19},
			},
		},
		{
			Code:    `<App foo={ <Bar /> } />`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never", "spacing": spc{"objectLiterals": "always"}}},
			Output:  []string{`<App foo={<Bar />} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 20},
			},
		},

		// ---- String / template / regex content — scanner must not mis-tokenize ----
		// The substrings `//`, `/*`, `{`, and `}` inside a string/template/regex
		// must NOT be treated as comments or extra braces. If the scanner gets
		// this wrong, fixes corrupt source.
		{
			Code:   `<App foo={ "// not a comment" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"// not a comment"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 31},
			},
		},
		{
			Code:   `<App foo={ "/* block */" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"/* block */"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 26},
			},
		},
		{
			Code:   `<App foo={ "{ not an object }" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"{ not an object }"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 32},
			},
		},
		{
			Code:   `<App foo={ /regex/g } />`,
			Tsx:    true,
			Output: []string{`<App foo={/regex/g} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},

		// (children spread is intentionally NOT covered — see corresponding
		// valid cases above for the parity with upstream JSXSpreadChild
		// behavior.)

		// ---- Nested element: only inner attribute container is checked under default config ----
		{
			Code:   `<App><Bar foo={ baz } /></App>`,
			Tsx:    true,
			Output: []string{`<App><Bar foo={baz} /></App>`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 15},
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},

		// ---- Real-world: ternary in className (single line, allowMultiline irrelevant) ----
		{
			Code:    `<div className={ isActive ? 'a' : 'b' } />`,
			Tsx:     true,
			Options: []interface{}{opts{"when": "never"}},
			Output:  []string{`<div className={isActive ? 'a' : 'b'} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 16},
				{MessageId: "noSpaceBefore", Line: 1, Column: 39},
			},
		},

		// ===== Robustness: complex inner contents under single-line `never` =====
		// Counterparts to the multi-line valid tests above — verify the
		// rule still fires correctly when these tricky bodies are wrapped
		// with explicit spaces under default `never`.
		{
			Code:   `<X foo={ "hello}world" } />`,
			Tsx:    true,
			Output: []string{`<X foo={"hello}world"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 8},
				{MessageId: "noSpaceBefore", Line: 1, Column: 24},
			},
		},
		{
			Code:   "<X foo={ `${a}${b}` } />",
			Tsx:    true,
			Output: []string{"<X foo={`${a}${b}`} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 8},
				{MessageId: "noSpaceBefore", Line: 1, Column: 21},
			},
		},
		{
			Code:   `<X foo={ /[}]/g } />`,
			Tsx:    true,
			Output: []string{`<X foo={/[}]/g} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 8},
				{MessageId: "noSpaceBefore", Line: 1, Column: 17},
			},
		},
		{
			Code:    `<X foo={"hello}world"} />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<X foo={ "hello}world" } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 8},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 22},
			},
		},
		{
			Code:    "<X foo={`${x}-${y}`} />",
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{"<X foo={ `${x}-${y}` } />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 8},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 20},
			},
		},
		{
			Code:    `<X foo={() => { return 1; }} />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<X foo={ () => { return 1; } } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 8},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 28},
			},
		},

		// ===== BOM + Unicode invalid lock-in =====

		// BOM at file start — tsgo's scanner counts the BOM as 1 UTF-16
		// character on line 1 (not stripped during position calculation),
		// so columns shift by +1 vs the BOM-less equivalent: `{` at col 11,
		// `}` at col 17. Locked in as observable rslint behaviour; if
		// upstream ESLint differs here that is a framework-level position-
		// calc divergence, not a rule-logic issue.
		{
			Code:   "\uFEFF<App foo={ bar } />",
			Tsx:    true,
			Output: []string{"\uFEFF<App foo={bar} />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 11},
				{MessageId: "noSpaceBefore", Line: 1, Column: 17},
			},
		},
		// BMP non-ASCII identifier — `中` and `文` each count as 1 UTF-16
		// character, so `}` of `<App foo={ 中文 } />` is at col 15.
		{
			Code:   `<App foo={ 中文 } />`,
			Tsx:    true,
			Output: []string{`<App foo={中文} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 15},
			},
		},
		// SMP character inside a string literal — `🚀` is a UTF-16 surrogate
		// pair (2 code units). For `<App foo={ "🚀" } />`, the closing `}`
		// is at col 17. (Emoji can only appear inside strings; it is not a
		// valid TypeScript / JSX identifier-start character.)
		{
			Code:   `<App foo={ "🚀" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"🚀"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 17},
			},
		},
		// Two SMP characters inside a string — each adds 2 UTF-16 chars,
		// so `}` of `<App foo={ "🚀🎉" } />` is at col 19.
		{
			Code:   `<App foo={ "🚀🎉" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"🚀🎉"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 19},
			},
		},
		// always mode + non-ASCII string — verifies `spaceNeededAfter` /
		// `spaceNeededBefore` columns under `always` with multi-byte body.
		// `<App foo={"中文"} />` — `}` at col 15.
		{
			Code:    `<App foo={"中文"} />`,
			Tsx:     true,
			Options: []interface{}{"always"},
			Output:  []string{`<App foo={ "中文" } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 15},
			},
		},
		// BMP + SMP mixed inside the same string — `中` and `文` count as
		// 1 UTF-16 char each, `🚀` counts as 2. For `<App foo={ "中文🚀" } />`
		// the closing `}` is at col 19. Locks in cross-class column math
		// in a single body and proves the byte-level scanner never lands
		// inside a multi-byte UTF-8 sequence.
		{
			Code:   `<App foo={ "中文🚀" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"中文🚀"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 19},
			},
		},
		// SMP between two BMP — verifies the scanner ignores the surrogate
		// pair regardless of position within the body.
		// `<App foo={ "🚀中文🎉中" } />` — col counts: `{`=10, `"`=12,
		// `🚀`=13-14, `中`=15, `文`=16, `🎉`=17-18, `中`=19, `"`=20,
		// `}`=22.
		{
			Code:   `<App foo={ "🚀中文🎉中" } />`,
			Tsx:    true,
			Output: []string{`<App foo={"🚀中文🎉中"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 22},
			},
		},
		// Multi-byte content inside a comment — block comment with both
		// CJK and emoji. Column for `}` of `<App foo={ /* 中🚀 */ x } />`
		// is computed as: `{`=10, ` `=11, `/`=12, `*`=13, ` `=14, `中`=15,
		// `🚀`=16-17, ` `=18, `*`=19, `/`=20, ` `=21, `x`=22, ` `=23,
		// `}`=24.
		{
			Code:   `<App foo={ /* 中🚀 */ x } />`,
			Tsx:    true,
			Output: []string{`<App foo={/* 中🚀 */ x} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 10},
				{MessageId: "noSpaceBefore", Line: 1, Column: 24},
			},
		},
		// Multi-byte chars BEFORE the JsxExpression on the SAME line —
		// verifies the brace's column is correctly offset by all preceding
		// UTF-16 units (BOM-precedent style verification with mixed
		// emoji+CJK preceding tokens). For
		// `const 名 = <App foo={ "🚀" } />` — col counts:
		// `c`=1..`t`=5, ` `=6, `名`=7, ` `=8, `=`=9, ` `=10, `<`=11,
		// `A`=12..`p`=14, ` `=15, `f`=16..`o`=18, `=`=19, `{`=20, ` `=21,
		// `"`=22, `🚀`=23-24, `"`=25, ` `=26, `}`=27.
		{
			Code:   `const 名 = <App foo={ "🚀" } />`,
			Tsx:    true,
			Output: []string{`const 名 = <App foo={"🚀"} />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noSpaceAfter", Line: 1, Column: 20},
				{MessageId: "noSpaceBefore", Line: 1, Column: 27},
			},
		},

		// ---- Robustness: block comment immediately before `}` on next line ----
		// Variant where the trailing trivia is a BLOCK comment (vs the line
		// comment case, which both upstream and rslint cannot autofix
		// without producing syntactically broken source). Here the fix is
		// well-defined: the trailing whitespace+newline collapse, and `}`
		// lands cleanly after `*/`.
		{
			Code:    "<App>{foo /* c */\n}</App>",
			Tsx:     true,
			Options: []interface{}{opts{"children": opts{"when": "never", "allowMultiline": false}}},
			Output:  []string{"<App>{foo /* c */}</App>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "noNewlineBefore", Line: 2, Column: 1},
			},
		},

		// ---- Normalization scope (stylisticScope=false): react keeps inheriting ----
		// Same input as the @stylistic extras' (a) lock-in, OPPOSITE result. In
		// react mode (BuildRule stylisticScope=false) a per-side empty
		// `spacing: {}` keeps INHERITING the top-level `objectLiterals: 'always'`,
		// so a flush object literal is reported as needing surrounding space.
		// (@stylistic falls back to when:'never' and accepts it — the one
		// cross-fork delta, locked in on both sides.)
		{
			Code:    `<App foo={{a: 1}} />`,
			Tsx:     true,
			Options: []interface{}{opts{"spacing": spc{"objectLiterals": "always"}, "attributes": opts{"when": "never", "spacing": spc{}}}},
			Output:  []string{`<App foo={ {a: 1} } />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spaceNeededAfter", Line: 1, Column: 10},
				{MessageId: "spaceNeededBefore", Line: 1, Column: 17},
			},
		},
	})
}
