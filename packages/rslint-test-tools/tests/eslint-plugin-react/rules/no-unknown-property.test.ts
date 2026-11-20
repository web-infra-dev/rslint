import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

const ruleTester = new RuleTester(fixtureConfig);

// NOTE: Cases that depend on `settings.react.version` (e.g. `allowTransparency`
// on React < 16.1, `precedence` on React >= 19) are version-gated and the
// shared JS rule-tester does not thread `settings` through to `lint()`. Those
// cases are covered by the Go unit tests (`no_unknown_property_test.go`); the
// JS suite here assumes the default (latest) React version.

ruleTester.run('no-unknown-property', {
  valid: [
    // React components and their props/attributes should be fine
    { code: '<App class="bar" />;' },
    { code: '<App for="bar" />;' },
    { code: '<App someProp="bar" />;' },
    { code: '<Foo.bar for="bar" />;' },
    { code: '<App accept-charset="bar" />;' },
    { code: '<App http-equiv="bar" />;' },
    { code: '<App xlink:href="bar" />;' },
    { code: '<App clip-path="bar" />;' },
    {
      code: '<App dataNotAnDataAttribute="yes" />;',
      options: [{ requireDataLowercase: true }],
    },
    // Some HTML/DOM elements with common attributes should work
    { code: '<div className="bar"></div>;' },
    { code: '<div onMouseDown={this._onMouseDown}></div>;' },
    { code: '<div onScrollEnd={this._onScrollEnd}></div>;' },
    { code: '<div onScrollEndCapture={this._onScrollEndCapture}></div>;' },
    { code: '<a href="someLink" download="foo">Read more</a>' },
    { code: '<area download="foo" />' },
    {
      code: '<img src="cat_keyboard.jpeg" alt="A cat sleeping on a keyboard" align="top" fetchPriority="high" />',
    },
    { code: '<input type="password" required />' },
    { code: '<input ref={this.input} type="radio" />' },
    { code: '<input type="file" webkitdirectory="" />' },
    { code: '<input type="file" webkitDirectory="" />' },
    { code: '<div inert children="anything" />' },
    { code: '<iframe scrolling="?" onLoad={a} onError={b} align="top" />' },
    { code: '<input key="bar" type="radio" />' },
    { code: '<button disabled>You cannot click me</button>;' },
    {
      code: '<svg key="lock" viewBox="box" fill={10} d="d" stroke={1} strokeWidth={2} strokeLinecap={3} strokeLinejoin={4} transform="something" clipRule="else" x1={5} x2="6" y1="7" y2="8"></svg>',
    },
    { code: '<g fill="#7B82A0" fillRule="evenodd"></g>' },
    { code: '<mask fill="#7B82A0"></mask>' },
    { code: '<symbol fill="#7B82A0"></symbol>' },
    { code: '<meta property="og:type" content="website" />' },
    {
      code: '<input type="checkbox" checked={checked} disabled={disabled} id={id} onChange={onChange} />',
    },
    { code: '<video playsInline />' },
    { code: '<img onError={foo} onLoad={bar} />' },
    { code: '<picture inert={false} onError={foo} onLoad={bar} />' },
    { code: '<iframe onError={foo} onLoad={bar} />' },
    { code: '<script onLoad={bar} onError={foo} />' },
    { code: '<source onLoad={bar} onError={foo} />' },
    { code: '<link onLoad={bar} onError={foo} />' },
    {
      code: '<link rel="preload" as="image" href="someHref" imageSrcSet="someImageSrcSet" imageSizes="someImageSizes" />',
    },
    { code: '<object onLoad={bar} />' },
    { code: '<body onLoad={bar} />' },
    {
      code: '<video allowFullScreen webkitAllowFullScreen mozAllowFullScreen />',
    },
    {
      code: '<iframe allowFullScreen webkitAllowFullScreen mozAllowFullScreen />',
    },
    { code: '<table border="1" />' },
    { code: '<th abbr="abbr" />' },
    { code: '<td abbr="abbr" />' },
    {
      code: '<template shadowrootmode="open" shadowrootclonable shadowrootdelegatesfocus shadowrootserializable />',
    },
    // React related attributes
    { code: '<div onPointerDown={this.onDown} onPointerUp={this.onUp} />' },
    { code: '<input type="checkbox" defaultChecked={this.state.checkbox} />' },
    {
      code: '<div onTouchStart={this.startAnimation} onTouchEnd={this.stopAnimation} onTouchCancel={this.cancel} onTouchMove={this.move} onMouseMoveCapture={this.capture} onTouchCancelCapture={this.log} />',
    },
    // Case ignored attributes
    { code: '<meta charset="utf-8" />;' },
    { code: '<meta charSet="utf-8" />;' },
    // Some custom web components that are allowed to use `class` instead of `className`
    { code: '<div class="foo" is="my-elem"></div>;' },
    { code: '<div {...this.props} class="foo" is="my-elem"></div>;' },
    { code: '<atom-panel class="foo"></atom-panel>;' },
    // data-* attributes should work
    { code: '<div data-foo="bar"></div>;' },
    { code: '<div data-foo-bar="baz"></div>;' },
    { code: '<div data-parent="parent"></div>;' },
    { code: '<div data-index-number="1234"></div>;' },
    { code: '<div data-e2e-id="5678"></div>;' },
    { code: '<div data-testID="bar" data-under_sCoRe="bar" />;' },
    {
      code: '<div data-testID="bar" data-under_sCoRe="bar" />;',
      options: [{ requireDataLowercase: false }],
    },
    // Ignoring should work
    {
      code: '<div class="bar"></div>;',
      options: [{ ignore: ['class'] }],
    },
    {
      code: '<div someProp="bar"></div>;',
      options: [{ ignore: ['someProp'] }],
    },
    {
      code: '<div css={{flex: 1}}></div>;',
      options: [{ ignore: ['css'] }],
    },
    // aria-* attributes should work
    { code: '<button aria-haspopup="true">Click me to open pop up</button>;' },
    { code: '<button aria-label="Close" onClick={someThing.close} />;' },
    // Attributes on allowed elements should work
    { code: '<script crossOrigin noModule />' },
    { code: '<audio crossOrigin />' },
    { code: '<svg focusable><image crossOrigin /></svg>' },
    { code: '<details onToggle={this.onToggle}>Some details</details>' },
    {
      code: '<path fill="pink" d="M 10,30 A 20,20 0,0,1 50,30 A 20,20 0,0,1 90,30 Q 90,60 50,90 Q 10,60 10,30 z"></path>',
    },
    { code: '<line fill="pink" x1="0" y1="80" x2="100" y2="20"></line>' },
    { code: '<link as="audio">Audio content</link>' },
    {
      code: '<video controlsList="nodownload" controls={this.controls} loop={true} muted={false} src={this.videoSrc} playsInline={true} onResize={this.onResize}></video>',
    },
    {
      code: '<audio controlsList="nodownload" controls={this.controls} crossOrigin="anonymous" disableRemotePlayback loop muted preload="none" src="something" onAbort={this.abort} onDurationChange={this.durationChange} onEmptied={this.emptied} onEnded={this.end} onError={this.error} onResize={this.onResize}></audio>',
    },
    {
      code: '<marker id={markerId} viewBox="0 0 2 2" refX="1" refY="1" markerWidth="1" markerHeight="1" orient="auto" />',
    },
    {
      code: '<pattern id="pattern" viewBox="0,0,10,10" width="10%" height="10%" />',
    },
    { code: '<symbol id="myDot" width="10" height="10" viewBox="0 0 2 2" />' },
    { code: '<view id="one" viewBox="0 0 100 100" />' },
    { code: '<hr align="top" />' },
    { code: '<applet align="top" />' },
    { code: '<marker fill="#000" />' },
    {
      code: '<dialog closedby="something" onClose={handler} open id="dialog" returnValue="something" onCancel={handler2} />',
    },
    {
      code: `
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
      `,
    },
    // fbt / fbs
    { code: '<fbt desc="foo" doNotExtract />;' },
    { code: '<fbs desc="foo" doNotExtract />;' },
    { code: '<math displaystyle="true" />;' },
    {
      code: `
        <div className="App" data-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash="customValue">
          Hello, world!
        </div>
      `,
    },
    {
      code: `
        <div>
          <button popovertarget="my-popover" popovertargetaction="toggle">Open Popover</button>

          <div popover id="my-popover">Greetings, one and all!</div>
        </div>
      `,
    },
    {
      code: `
        <div>
          <button popoverTarget="my-popover" popoverTargetAction="toggle">Open Popover</button>

          <div id="my-popover" onBeforeToggle={this.onBeforeToggle} popover>Greetings, one and all!</div>
        </div>
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `<div allowTransparency="true" />`,
      settings: { react: { version: '16.0.99' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `<link precedence="medium" href="https://foo.bar" rel="canonical" />`,
      settings: { react: { version: '19.0.0' } },
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 81 (7.37.5)',
      code: `<dialog onClose={handler} open id="dialog" returnValue="something" onCancel={handler2} />`,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: '<div hasOwnProperty="should not be allowed property"></div>;',
      errors: [
        {
          messageId: 'unknownProp',
          message: "Unknown property 'hasOwnProperty' found",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<div abc="should not be allowed property"></div>;',
      errors: [
        {
          messageId: 'unknownProp',
          message: "Unknown property 'abc' found",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<div aria-fake="should not be allowed property"></div>;',
      errors: [
        {
          messageId: 'unknownProp',
          message: "Unknown property 'aria-fake' found",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<div someProp="bar"></div>;',
      errors: [
        {
          messageId: 'unknownProp',
          message: "Unknown property 'someProp' found",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div class="bar"></div>;',
      output: '<div className="bar"></div>;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message: "Unknown property 'class' found, use 'className' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<div for="bar"></div>;',
      output: '<div htmlFor="bar"></div>;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message: "Unknown property 'for' found, use 'htmlFor' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<div accept-charset="bar"></div>;',
      output: '<div acceptCharset="bar"></div>;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message:
            "Unknown property 'accept-charset' found, use 'acceptCharset' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div http-equiv="bar"></div>;',
      output: '<div httpEquiv="bar"></div>;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message:
            "Unknown property 'http-equiv' found, use 'httpEquiv' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div accesskey="bar"></div>;',
      output: '<div accessKey="bar"></div>;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message:
            "Unknown property 'accesskey' found, use 'accessKey' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div onclick="bar"></div>;',
      output: '<div onClick="bar"></div>;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message: "Unknown property 'onclick' found, use 'onClick' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div onmousedown="bar"></div>;',
      output: '<div onMouseDown="bar"></div>;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message:
            "Unknown property 'onmousedown' found, use 'onMouseDown' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div onMousedown="bar"></div>;',
      output: '<div onMouseDown="bar"></div>;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message:
            "Unknown property 'onMousedown' found, use 'onMouseDown' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<use xlink:href="bar" />;',
      output: '<use xlinkHref="bar" />;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message:
            "Unknown property 'xlink:href' found, use 'xlinkHref' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<rect clip-path="bar" transform-origin="center" />;',
      output: '<rect clipPath="bar" transform-origin="center" />;',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message: "Unknown property 'clip-path' found, use 'clipPath' instead",
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<script crossorigin nomodule />',
      output: '<script crossOrigin noModule />',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message:
            "Unknown property 'crossorigin' found, use 'crossOrigin' instead",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'unknownPropWithStandardName',
          message: "Unknown property 'nomodule' found, use 'noModule' instead",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div crossorigin />',
      output: '<div crossOrigin />',
      errors: [
        {
          messageId: 'unknownPropWithStandardName',
          message:
            "Unknown property 'crossorigin' found, use 'crossOrigin' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<div crossOrigin />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'crossOrigin' found on tag 'div', but it is only allowed on: script, img, video, audio, link, image",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<div as="audio" />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'as' found on tag 'div', but it is only allowed on: link",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<div onAbort={this.abort} onDurationChange={this.durationChange} onEmptied={this.emptied} onEnded={this.end} onResize={this.resize} onError={this.error} />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'onAbort' found on tag 'div', but it is only allowed on: audio, video",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'onDurationChange' found on tag 'div', but it is only allowed on: audio, video",
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 65,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'onEmptied' found on tag 'div', but it is only allowed on: audio, video",
          line: 1,
          column: 66,
          endLine: 1,
          endColumn: 90,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'onEnded' found on tag 'div', but it is only allowed on: audio, video",
          line: 1,
          column: 91,
          endLine: 1,
          endColumn: 109,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'onResize' found on tag 'div', but it is only allowed on: audio, video",
          line: 1,
          column: 110,
          endLine: 1,
          endColumn: 132,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'onError' found on tag 'div', but it is only allowed on: audio, video, img, link, source, script, picture, iframe",
          line: 1,
          column: 133,
          endLine: 1,
          endColumn: 153,
        },
      ],
    },
    {
      code: '<div onLoad={this.load} />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'onLoad' found on tag 'div', but it is only allowed on: script, img, link, picture, iframe, object, source, body",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div fill="pink" />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'fill' found on tag 'div', but it is only allowed on: altGlyph, circle, ellipse, g, line, marker, mask, path, polygon, polyline, rect, svg, symbol, text, textPath, tref, tspan, use, animate, animateColor, animateMotion, animateTransform, set",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<div controls={this.controls} loop={true} muted={false} src={this.videoSrc} playsInline={true} allowFullScreen></div>',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'controls' found on tag 'div', but it is only allowed on: audio, video",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'loop' found on tag 'div', but it is only allowed on: audio, video",
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 42,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'muted' found on tag 'div', but it is only allowed on: audio, video",
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 56,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'playsInline' found on tag 'div', but it is only allowed on: video",
          line: 1,
          column: 77,
          endLine: 1,
          endColumn: 95,
        },
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'allowFullScreen' found on tag 'div', but it is only allowed on: iframe, video",
          line: 1,
          column: 96,
          endLine: 1,
          endColumn: 111,
        },
      ],
    },
    {
      code: '<div download="foo" />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'download' found on tag 'div', but it is only allowed on: a, area",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div imageSrcSet="someImageSrcSet" />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'imageSrcSet' found on tag 'div', but it is only allowed on: link",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<div imageSizes="someImageSizes" />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'imageSizes' found on tag 'div', but it is only allowed on: link",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div data-xml-anything="invalid" />',
      errors: [
        {
          messageId: 'unknownProp',
          message: "Unknown property 'data-xml-anything' found",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div data-testID="bar" data-under_sCoRe="bar" dataNotAnDataAttribute="yes" />;',
      errors: [
        {
          messageId: 'dataLowercaseRequired',
          message:
            "React does not recognize data-* props with uppercase characters on a DOM element. Found 'data-testID', use 'data-testid' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
        {
          messageId: 'dataLowercaseRequired',
          message:
            "React does not recognize data-* props with uppercase characters on a DOM element. Found 'data-under_sCoRe', use 'data-under_score' instead",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 46,
        },
        {
          messageId: 'unknownProp',
          message: "Unknown property 'dataNotAnDataAttribute' found",
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 75,
        },
      ],
      options: [{ requireDataLowercase: true }],
    },
    {
      code: '<App data-testID="bar" data-under_sCoRe="bar" dataNotAnDataAttribute="yes" />;',
      errors: [
        {
          messageId: 'dataLowercaseRequired',
          message:
            "React does not recognize data-* props with uppercase characters on a DOM element. Found 'data-testID', use 'data-testid' instead",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
        {
          messageId: 'dataLowercaseRequired',
          message:
            "React does not recognize data-* props with uppercase characters on a DOM element. Found 'data-under_sCoRe', use 'data-under_score' instead",
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 46,
        },
      ],
      options: [{ requireDataLowercase: true }],
    },
    {
      code: '<div abbr="abbr" />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'abbr' found on tag 'div', but it is only allowed on: th, td",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<div webkitDirectory="" />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'webkitDirectory' found on tag 'div', but it is only allowed on: input",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div webkitdirectory="" />',
      errors: [
        {
          messageId: 'invalidPropOnTag',
          message:
            "Invalid property 'webkitdirectory' found on tag 'div', but it is only allowed on: input",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<div allowTransparency="true" />`,
      settings: { react: { version: '16.1.0' } },
      errors: [
        {
          messageId: 'unknownProp',
          message: "Unknown property 'allowTransparency' found",
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        <div className="App" data-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash:c="customValue">
          Hello, world!
        </div>
      `,
      errors: [
        {
          messageId: 'unknownProp',
          message:
            "Unknown property 'data-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash-crash:c' found",
          line: 2,
          column: 30,
          endLine: 2,
          endColumn: 326,
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
