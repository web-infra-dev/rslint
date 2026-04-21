// cspell:disable — tests deliberately contain misspelled / lowercased DOM
// attribute names (e.g. `crossorigin`, `nomodule`, `onmousedown`, `webkitdirectory`)
// that this rule is designed to flag.
package no_unknown_property

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnknownPropertyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnknownPropertyRule, []rule_tester.ValidTestCase{
		// ---- React components and their props/attributes should be fine ----
		{Code: `<App class="bar" />;`, Tsx: true},
		{Code: `<App for="bar" />;`, Tsx: true},
		{Code: `<App someProp="bar" />;`, Tsx: true},
		{Code: `<Foo.bar for="bar" />;`, Tsx: true},
		{Code: `<App accept-charset="bar" />;`, Tsx: true},
		{Code: `<App http-equiv="bar" />;`, Tsx: true},
		{Code: `<App xlink:href="bar" />;`, Tsx: true},
		{Code: `<App clip-path="bar" />;`, Tsx: true},
		{
			Code:    `<App dataNotAnDataAttribute="yes" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"requireDataLowercase": true},
		},

		// ---- Some HTML/DOM elements with common attributes ----
		{Code: `<div className="bar"></div>;`, Tsx: true},
		{Code: `<div onMouseDown={this._onMouseDown}></div>;`, Tsx: true},
		{Code: `<div onScrollEnd={this._onScrollEnd}></div>;`, Tsx: true},
		{Code: `<div onScrollEndCapture={this._onScrollEndCapture}></div>;`, Tsx: true},
		{Code: `<a href="someLink" download="foo">Read more</a>`, Tsx: true},
		{Code: `<area download="foo" />`, Tsx: true},
		{Code: `<img src="cat_keyboard.jpeg" alt="A cat sleeping on a keyboard" align="top" fetchPriority="high" />`, Tsx: true},
		{Code: `<input type="password" required />`, Tsx: true},
		{Code: `<input ref={this.input} type="radio" />`, Tsx: true},
		{Code: `<input type="file" webkitdirectory="" />`, Tsx: true},
		{Code: `<input type="file" webkitDirectory="" />`, Tsx: true},
		{Code: `<div inert children="anything" />`, Tsx: true},
		{Code: `<iframe scrolling="?" onLoad={a} onError={b} align="top" />`, Tsx: true},
		{Code: `<input key="bar" type="radio" />`, Tsx: true},
		{Code: `<button disabled>You cannot click me</button>;`, Tsx: true},
		{Code: `<svg key="lock" viewBox="box" fill={10} d="d" stroke={1} strokeWidth={2} strokeLinecap={3} strokeLinejoin={4} transform="something" clipRule="else" x1={5} x2="6" y1="7" y2="8"></svg>`, Tsx: true},
		{Code: `<g fill="#7B82A0" fillRule="evenodd"></g>`, Tsx: true},
		{Code: `<mask fill="#7B82A0"></mask>`, Tsx: true},
		{Code: `<symbol fill="#7B82A0"></symbol>`, Tsx: true},
		{Code: `<meta property="og:type" content="website" />`, Tsx: true},
		{Code: `<input type="checkbox" checked={checked} disabled={disabled} id={id} onChange={onChange} />`, Tsx: true},
		{Code: `<video playsInline />`, Tsx: true},
		{Code: `<img onError={foo} onLoad={bar} />`, Tsx: true},
		{Code: `<picture inert={false} onError={foo} onLoad={bar} />`, Tsx: true},
		{Code: `<iframe onError={foo} onLoad={bar} />`, Tsx: true},
		{Code: `<script onLoad={bar} onError={foo} />`, Tsx: true},
		{Code: `<source onLoad={bar} onError={foo} />`, Tsx: true},
		{Code: `<link onLoad={bar} onError={foo} />`, Tsx: true},
		{Code: `<link rel="preload" as="image" href="someHref" imageSrcSet="someImageSrcSet" imageSizes="someImageSizes" />`, Tsx: true},
		{Code: `<object onLoad={bar} />`, Tsx: true},
		{Code: `<body onLoad={bar} />`, Tsx: true},
		{Code: `<video allowFullScreen webkitAllowFullScreen mozAllowFullScreen />`, Tsx: true},
		{Code: `<iframe allowFullScreen webkitAllowFullScreen mozAllowFullScreen />`, Tsx: true},
		{Code: `<table border="1" />`, Tsx: true},
		{Code: `<th abbr="abbr" />`, Tsx: true},
		{Code: `<td abbr="abbr" />`, Tsx: true},
		{Code: `<template shadowrootmode="open" shadowrootclonable shadowrootdelegatesfocus shadowrootserializable />`, Tsx: true},
		{
			Code:     `<div allowTransparency="true" />`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.0.99"}},
		},

		// ---- React-related attributes ----
		{Code: `<div onPointerDown={this.onDown} onPointerUp={this.onUp} />`, Tsx: true},
		{Code: `<input type="checkbox" defaultChecked={this.state.checkbox} />`, Tsx: true},
		{Code: `<div onTouchStart={this.startAnimation} onTouchEnd={this.stopAnimation} onTouchCancel={this.cancel} onTouchMove={this.move} onMouseMoveCapture={this.capture} onTouchCancelCapture={this.log} />`, Tsx: true},
		{
			Code:     `<link precedence="medium" href="https://foo.bar" rel="canonical" />`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "19.0.0"}},
		},

		// ---- Case-ignored attributes ----
		{Code: `<meta charset="utf-8" />;`, Tsx: true},
		{Code: `<meta charSet="utf-8" />;`, Tsx: true},

		// ---- Custom web components allowed to use `class` ----
		{Code: `<div class="foo" is="my-elem"></div>;`, Tsx: true},
		{Code: `<div {...this.props} class="foo" is="my-elem"></div>;`, Tsx: true},
		{Code: `<atom-panel class="foo"></atom-panel>;`, Tsx: true},

		// ---- data-* attributes ----
		{Code: `<div data-foo="bar"></div>;`, Tsx: true},
		{Code: `<div data-foo-bar="baz"></div>;`, Tsx: true},
		{Code: `<div data-parent="parent"></div>;`, Tsx: true},
		{Code: `<div data-index-number="1234"></div>;`, Tsx: true},
		{Code: `<div data-e2e-id="5678"></div>;`, Tsx: true},
		{Code: `<div data-testID="bar" data-under_sCoRe="bar" />;`, Tsx: true},
		{
			Code:    `<div data-testID="bar" data-under_sCoRe="bar" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"requireDataLowercase": false},
		},

		// ---- Ignoring via options ----
		{
			Code:    `<div class="bar"></div>;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"class"}},
		},
		{
			Code:    `<div someProp="bar"></div>;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"someProp"}},
		},
		{
			Code:    `<div css={{flex: 1}}></div>;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"css"}},
		},

		// ---- aria-* attributes ----
		{Code: `<button aria-haspopup="true">Click me to open pop up</button>;`, Tsx: true},
		{Code: `<button aria-label="Close" onClick={someThing.close} />;`, Tsx: true},

		// ---- Attributes on allowed elements ----
		{Code: `<script crossOrigin noModule />`, Tsx: true},
		{Code: `<audio crossOrigin />`, Tsx: true},
		{Code: `<svg focusable><image crossOrigin /></svg>`, Tsx: true},
		{Code: `<details onToggle={this.onToggle}>Some details</details>`, Tsx: true},
		{Code: `<path fill="pink" d="M 10,30 A 20,20 0,0,1 50,30 A 20,20 0,0,1 90,30 Q 90,60 50,90 Q 10,60 10,30 z"></path>`, Tsx: true},
		{Code: `<line fill="pink" x1="0" y1="80" x2="100" y2="20"></line>`, Tsx: true},
		{Code: `<link as="audio">Audio content</link>`, Tsx: true},
		{Code: `<video controlsList="nodownload" controls={this.controls} loop={true} muted={false} src={this.videoSrc} playsInline={true} onResize={this.onResize}></video>`, Tsx: true},
		{Code: `<audio controlsList="nodownload" controls={this.controls} crossOrigin="anonymous" disableRemotePlayback loop muted preload="none" src="something" onAbort={this.abort} onDurationChange={this.durationChange} onEmptied={this.emptied} onEnded={this.end} onError={this.error} onResize={this.onResize}></audio>`, Tsx: true},
		{Code: `<marker id={markerId} viewBox="0 0 2 2" refX="1" refY="1" markerWidth="1" markerHeight="1" orient="auto" />`, Tsx: true},
		{Code: `<pattern id="pattern" viewBox="0,0,10,10" width="10%" height="10%" />`, Tsx: true},
		{Code: `<symbol id="myDot" width="10" height="10" viewBox="0 0 2 2" />`, Tsx: true},
		{Code: `<view id="one" viewBox="0 0 100 100" />`, Tsx: true},
		{Code: `<hr align="top" />`, Tsx: true},
		{Code: `<applet align="top" />`, Tsx: true},
		{Code: `<marker fill="#000" />`, Tsx: true},
		{Code: `<dialog closedby="something" onClose={handler} open id="dialog" returnValue="something" onCancel={handler2} />`, Tsx: true},
		{Code: `
        <table align="top">
          <caption align="top">Table Caption</caption>
          <colgroup valign="top" align="top">
            <col valign="top" align="top"/>
          </colgroup>
          <thead valign="top" align="top">
            <tr valign="top" align="top">
              <th valign="top" align="top">Header</th>
              <td valign="top" align="top">Cell</td>
            </tr>
          </thead>
          <tbody valign="top" align="top" />
          <tfoot valign="top" align="top" />
        </table>
      `, Tsx: true},

		// ---- fbt / fbs are bypassed ----
		{Code: `<fbt desc="foo" doNotExtract />;`, Tsx: true},
		{Code: `<fbs desc="foo" doNotExtract />;`, Tsx: true},
		{Code: `<math displaystyle="true" />;`, Tsx: true},

		// ---- Extra-long data-* attribute name ----
		{Code: `
        <div className="App" data-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash="customValue">
          Hello, world!
        </div>
      `, Tsx: true},

		// ---- Popover attributes ----
		{Code: `
        <div>
          <button popovertarget="my-popover" popovertargetaction="toggle">Open Popover</button>

          <div popover id="my-popover">Greetings, one and all!</div>
        </div>
      `, Tsx: true},
		{Code: `
        <div>
          <button popoverTarget="my-popover" popoverTargetAction="toggle">Open Popover</button>

          <div id="my-popover" onBeforeToggle={this.onBeforeToggle} popover>Greetings, one and all!</div>
        </div>
      `, Tsx: true},

		// ---- Extra edge cases ----
		// JSX fragment wrapping children — each child's attrs are independent.
		{Code: `<><div className="a" /><div className="b" /></>;`, Tsx: true},
		// Custom elements (hyphenated tag) — rule does not apply.
		{Code: `<my-custom someUnknownProp="x" />;`, Tsx: true},
		// Nested dotted components still get skipped at any depth.
		{Code: `<Foo.Bar.Baz class="x" for="y" someProp="z" />;`, Tsx: true},
		// Spread preceding a valid prop on a DOM tag.
		{Code: `<div {...rest} className="x" />;`, Tsx: true},
		// Spread between attributes.
		{Code: `<div className="a" {...rest} id="b" />;`, Tsx: true},
		// Unknown prop on a component (upper-case tag) is always allowed.
		{Code: `<Foo unknownWeirdProp="x" />;`, Tsx: true},
		// Multiple nested JSX elements with mixed DOM / component tags.
		{Code: `<App><div className="ok"><span data-foo="bar" /></div></App>;`, Tsx: true},
		// Boolean `is` attribute (no value) still marks the tag as custom-elements.
		{Code: `<div is class="foo" />;`, Tsx: true},
		// `is={expr}` form (non-string value) — upstream only checks attr name.
		{Code: `<div is={someVar} class="foo" />;`, Tsx: true},
		// aria-* exact-case lowercase is valid.
		{Code: `<div aria-hidden="true" />;`, Tsx: true},
		// data- with exact 5-char prefix only (no suffix) — regex allows zero-length `[^:]*`.
		{Code: `<div data- />;`, Tsx: true},
		// `fbt` gets a bye even for obviously unknown props.
		{Code: `<fbt someNonsense="x" />;`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- allowTransparency with newer React (removed from DOM prop list) ----
		{
			Code:     `<div allowTransparency="true" />`,
			Tsx:      true,
			Settings: map[string]interface{}{"react": map[string]interface{}{"version": "16.1.0"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 6},
			},
		},

		// ---- Unknown names ----
		{
			Code: `<div hasOwnProperty="should not be allowed property"></div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div abc="should not be allowed property"></div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div aria-fake="should not be allowed property"></div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div someProp="bar"></div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 6},
			},
		},

		// ---- unknownPropWithStandardName + autofix ----
		{
			Code:   `<div class="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div className="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<div for="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div htmlFor="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<div accept-charset="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div acceptCharset="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<div http-equiv="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div httpEquiv="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<div accesskey="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div accessKey="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<div onclick="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div onClick="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<div onmousedown="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div onMouseDown="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<div onMousedown="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div onMouseDown="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<use xlink:href="bar" />;`,
			Tsx:    true,
			Output: []string{`<use xlinkHref="bar" />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},
		{
			Code:   `<rect clip-path="bar" transform-origin="center" />;`,
			Tsx:    true,
			Output: []string{`<rect clipPath="bar" transform-origin="center" />;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 7},
			},
		},
		{
			Code:   `<script crossorigin nomodule />`,
			Tsx:    true,
			Output: []string{`<script crossOrigin noModule />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 9},
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 21},
			},
		},
		{
			Code:   `<div crossorigin />`,
			Tsx:    true,
			Output: []string{`<div crossOrigin />`},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownPropWithStandardName", Line: 1, Column: 6},
			},
		},

		// ---- invalidPropOnTag (attribute allowed only on specific tags) ----
		{
			Code: `<div crossOrigin />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div as="audio" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div onAbort={this.abort} onDurationChange={this.durationChange} onEmptied={this.emptied} onEnded={this.end} onResize={this.resize} onError={this.error} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 27},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 66},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 91},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 110},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 133},
			},
		},
		{
			Code: `<div onLoad={this.load} />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div fill="pink" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div controls={this.controls} loop={true} muted={false} src={this.videoSrc} playsInline={true} allowFullScreen></div>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 31},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 43},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 77},
				{MessageId: "invalidPropOnTag", Line: 1, Column: 96},
			},
		},
		{
			Code: `<div download="foo" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div imageSrcSet="someImageSrcSet" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div imageSizes="someImageSizes" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},

		// ---- data-xml-* reserved → unknownProp ----
		{
			Code: `<div data-xml-anything="invalid" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 6},
			},
		},

		// ---- requireDataLowercase (only checks data-* with uppercase) ----
		{
			Code:    `<div data-testID="bar" data-under_sCoRe="bar" dataNotAnDataAttribute="yes" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"requireDataLowercase": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dataLowercaseRequired", Line: 1, Column: 6},
				{MessageId: "dataLowercaseRequired", Line: 1, Column: 24},
				{MessageId: "unknownProp", Line: 1, Column: 47},
			},
		},
		{
			Code:    `<App data-testID="bar" data-under_sCoRe="bar" dataNotAnDataAttribute="yes" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"requireDataLowercase": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "dataLowercaseRequired", Line: 1, Column: 6},
				{MessageId: "dataLowercaseRequired", Line: 1, Column: 24},
			},
		},

		// ---- Element-specific attribute on wrong tag ----
		{
			Code: `<div abbr="abbr" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div webkitDirectory="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		{
			Code: `<div webkitdirectory="" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "invalidPropOnTag", Line: 1, Column: 6},
			},
		},
		// Upstream line 715 — JsxNamespacedName `data-<lots-of-crash>:c`, the
		// long hyphenated namespace fails upstream's `/^data-[^:]*$/` (contains
		// `:`), so it falls through to unknownProp. Upstream marks this case
		// `features: ['no-ts']` because TS ESTree parsers failed to parse it;
		// tsgo does parse it as JsxNamespacedName, so we can exercise it here.
		{
			Code: `
        <div className="App" data-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash:c="customValue">
          Hello, world!
        </div>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 2, Column: 30},
			},
		},

		// ---- Extra edge cases (multi-line / nested / contract assertions) ----
		// Multi-line JSX — Line / Column on an invalid prop.
		{
			Code: `<div
  class="x"
/>`,
			Tsx:    true,
			Output: []string{"<div\n  className=\"x\"\n/>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unknownPropWithStandardName",
					Line:      2, Column: 3,
				},
			},
		},
		// Nested JSX: inner DOM element has an unknown prop; outer component is fine.
		{
			Code: `<App><div abc="x" /></App>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 11},
			},
		},
		// Two nested invalid usages at different nesting depths.
		{
			Code: `<div><section abc="x"><span def="y" /></section></div>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 15},
				{MessageId: "unknownProp", Line: 1, Column: 29},
			},
		},
		// Exact message text (contract assertion for unknownPropWithStandardName).
		{
			Code:   `<div class="bar"></div>;`,
			Tsx:    true,
			Output: []string{`<div className="bar"></div>;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unknownPropWithStandardName",
					Message:   "Unknown property 'class' found, use 'className' instead",
					Line:      1, Column: 6,
				},
			},
		},
		// Exact message text for invalidPropOnTag with preserved allowedTags order.
		{
			Code: `<div crossOrigin />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "invalidPropOnTag",
					Message:   "Invalid property 'crossOrigin' found on tag 'div', but it is only allowed on: script, img, video, audio, link, image",
					Line:      1, Column: 6,
				},
			},
		},
		// Exact message text for unknownProp.
		{
			Code: `<div abc="1" />`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unknownProp",
					Message:   "Unknown property 'abc' found",
					Line:      1, Column: 6,
				},
			},
		},
		// Exact message text for dataLowercaseRequired.
		{
			Code:    `<div data-testID="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"requireDataLowercase": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "dataLowercaseRequired",
					Message:   "React does not recognize data-* props with uppercase characters on a DOM element. Found 'data-testID', use 'data-testid' instead",
					Line:      1, Column: 6,
				},
			},
		},
		// Namespaced `xlink:*` on a DOM tag is reported with the full namespaced text.
		{
			Code: `<use xlink:unknown="x" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 6},
			},
		},
		// Spread + unknown prop: spread is ignored, unknown prop reports.
		{
			Code: `<div {...rest} abc="x" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 16},
			},
		},
		// Fragment wrapping two DOM tags, both unknown props.
		{
			Code: `<><div abc="1" /><span def="2" /></>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 8},
				{MessageId: "unknownProp", Line: 1, Column: 24},
			},
		},
		// ignore matches `actualName` exactly (pre-normalization) — `class`
		// is ignored but `Class` (different case) is NOT; upstream's `indexOf`
		// is case-sensitive and `Class` has no close-case standard property
		// either (neither `class` nor `className` collide with `Class` under
		// the lowercased-comparison lookup), so it falls through to unknownProp.
		{
			Code:    `<div Class="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"ignore": []interface{}{"class"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unknownProp", Line: 1, Column: 6},
			},
		},
	})
}
