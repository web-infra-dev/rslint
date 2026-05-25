import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// NOTE: Cases that depend on `settings.react.version` (e.g. `allowTransparency`
// on React < 16.1, `precedence` on React >= 19) are version-gated and the
// shared JS rule-tester does not thread `settings` through to `lint()`. Those
// cases are covered by the Go unit tests (`no_unknown_property_test.go`); the
// JS suite here assumes the default (latest) React version.

ruleTester.run('no-unknown-property', {} as never, {
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
  ],
  invalid: [
    {
      code: '<div hasOwnProperty="should not be allowed property"></div>;',
      errors: [{ messageId: 'unknownProp' }],
    },
    {
      code: '<div abc="should not be allowed property"></div>;',
      errors: [{ messageId: 'unknownProp' }],
    },
    {
      code: '<div aria-fake="should not be allowed property"></div>;',
      errors: [{ messageId: 'unknownProp' }],
    },
    {
      code: '<div someProp="bar"></div>;',
      errors: [{ messageId: 'unknownProp' }],
    },
    {
      code: '<div class="bar"></div>;',
      output: '<div className="bar"></div>;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<div for="bar"></div>;',
      output: '<div htmlFor="bar"></div>;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<div accept-charset="bar"></div>;',
      output: '<div acceptCharset="bar"></div>;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<div http-equiv="bar"></div>;',
      output: '<div httpEquiv="bar"></div>;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<div accesskey="bar"></div>;',
      output: '<div accessKey="bar"></div>;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<div onclick="bar"></div>;',
      output: '<div onClick="bar"></div>;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<div onmousedown="bar"></div>;',
      output: '<div onMouseDown="bar"></div>;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<div onMousedown="bar"></div>;',
      output: '<div onMouseDown="bar"></div>;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<use xlink:href="bar" />;',
      output: '<use xlinkHref="bar" />;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<rect clip-path="bar" transform-origin="center" />;',
      output: '<rect clipPath="bar" transform-origin="center" />;',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<script crossorigin nomodule />',
      output: '<script crossOrigin noModule />',
      errors: [
        { messageId: 'unknownPropWithStandardName' },
        { messageId: 'unknownPropWithStandardName' },
      ],
    },
    {
      code: '<div crossorigin />',
      output: '<div crossOrigin />',
      errors: [{ messageId: 'unknownPropWithStandardName' }],
    },
    {
      code: '<div crossOrigin />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div as="audio" />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div onAbort={this.abort} onDurationChange={this.durationChange} onEmptied={this.emptied} onEnded={this.end} onResize={this.resize} onError={this.error} />',
      errors: [
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
      ],
    },
    {
      code: '<div onLoad={this.load} />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div fill="pink" />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div controls={this.controls} loop={true} muted={false} src={this.videoSrc} playsInline={true} allowFullScreen></div>',
      errors: [
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
        { messageId: 'invalidPropOnTag' },
      ],
    },
    {
      code: '<div download="foo" />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div imageSrcSet="someImageSrcSet" />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div imageSizes="someImageSizes" />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div data-xml-anything="invalid" />',
      errors: [{ messageId: 'unknownProp' }],
    },
    {
      code: '<div data-testID="bar" data-under_sCoRe="bar" dataNotAnDataAttribute="yes" />;',
      errors: [
        { messageId: 'dataLowercaseRequired' },
        { messageId: 'dataLowercaseRequired' },
        { messageId: 'unknownProp' },
      ],
      options: [{ requireDataLowercase: true }],
    },
    {
      code: '<App data-testID="bar" data-under_sCoRe="bar" dataNotAnDataAttribute="yes" />;',
      errors: [
        { messageId: 'dataLowercaseRequired' },
        { messageId: 'dataLowercaseRequired' },
      ],
      options: [{ requireDataLowercase: true }],
    },
    {
      code: '<div abbr="abbr" />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div webkitDirectory="" />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
    {
      code: '<div webkitdirectory="" />',
      errors: [{ messageId: 'invalidPropOnTag' }],
    },
  ],
});
