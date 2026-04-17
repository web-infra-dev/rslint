import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-unescaped-entities', {} as never, {
  valid: [
    // structural: no JsxText
    { code: `var Hello = <div/>` },
    { code: `var Hello = <div></div>` },
    { code: `var Hello = <></>` },
    { code: `var Hello = <Component name="don't" />` },
    { code: `var Hello = <div title="it's fine"></div>` },

    // plain / already-escaped text
    { code: `var Hello = <div>Here is some text!</div>` },
    {
      code: `var Hello = <div>I&rsquo;ve escaped some entities: &gt; &lt; &amp;</div>`,
    },

    // whitespace-only
    { code: `var Hello = <div>   </div>` },

    // string literals inside expression containers are not JsxText
    { code: `var Hello = <div>{">" + "<" + "&" + '"'}</div>` },
    { code: `var Hello = <div>{"it's fine"}</div>` },

    // nested safe content
    { code: `var Hello = <Outer><Inner>safe</Inner></Outer>` },
    { code: `var Hello = <ul><li>one</li><li>two</li></ul>` },
    { code: `var Hello = <Outer child={<Inner>safe</Inner>} />` },

    // custom forbid replaces defaults
    {
      code: `var Hello = <div>don't forget</div>`,
      options: [{ forbid: ['&'] }],
    },

    // forbid: [] explicitly disables all
    {
      code: `var Hello = <div>don't do that</div>`,
      options: [{ forbid: [] }],
    },
  ],
  invalid: [
    {
      code: `var Hello = <div>'</div>`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    {
      code: `var Hello = <div>Don't do that</div>`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    {
      code: `var Hello = <>it's a trap</>`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    // nested: flag inner only
    {
      code: `var Hello = <Outer>outer text<Inner>inner's</Inner></Outer>`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    // nested: flag at multiple levels
    {
      code: `var Hello = <Outer>outer's<Inner>inner's</Inner></Outer>`,
      errors: [
        { messageId: 'unescapedEntityAlts', data: { entity: "'" } },
        { messageId: 'unescapedEntityAlts', data: { entity: "'" } },
      ],
    },
    // JSX inside expression container inside JSX
    {
      code: `var Hello = <Outer>{<Inner>inner's</Inner>}</Outer>`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    // JSX as prop value: inner JsxText is scanned
    {
      code: `var Hello = <Outer child={<Inner>inner's</Inner>} />`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    // adjacent duplicates
    {
      code: `var Hello = <div>a''b</div>`,
      errors: [
        { messageId: 'unescapedEntityAlts', data: { entity: "'" } },
        { messageId: 'unescapedEntityAlts', data: { entity: "'" } },
      ],
    },
    // mixed default chars in same JsxText (source order)
    {
      code: `var Hello = <div>"don't"</div>`,
      errors: [
        { messageId: 'unescapedEntityAlts', data: { entity: '"' } },
        { messageId: 'unescapedEntityAlts', data: { entity: "'" } },
        { messageId: 'unescapedEntityAlts', data: { entity: '"' } },
      ],
    },
    // script tag: unescaped quotes
    {
      code: `var Hello = <script>window.foo = "bar"</script>`,
      errors: [
        { messageId: 'unescapedEntityAlts', data: { entity: '"' } },
        { messageId: 'unescapedEntityAlts', data: { entity: '"' } },
      ],
    },
    // custom forbid: simple string form
    {
      code: `var Hello = <span>foo & bar</span>`,
      options: [{ forbid: ['&'] }],
      errors: [{ messageId: 'unescapedEntity', data: { entity: '&' } }],
    },
    // custom forbid: object form
    {
      code: `var Hello = <span>foo & bar</span>`,
      options: [{ forbid: [{ char: '&', alternatives: ['&amp;'] }] }],
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: '&' } }],
    },
    // forbid: mix of string + object entries
    {
      code: `var Hello = <div>a & b $ c</div>`,
      options: [
        {
          forbid: ['&', { char: '$', alternatives: ['&#36;'] }],
        },
      ],
      errors: [
        { messageId: 'unescapedEntity', data: { entity: '&' } },
        { messageId: 'unescapedEntityAlts', data: { entity: '$' } },
      ],
    },
    // multi-byte Unicode custom forbid char (fullwidth apostrophe U+FF07)
    {
      code: `var Hello = <div>fullwidth\uff07quote</div>`,
      options: [{ forbid: ['\uff07'] }],
      errors: [{ messageId: 'unescapedEntity', data: { entity: '\uff07' } }],
    },
    // UTF-16 column stays correct when a multi-byte char precedes the match
    {
      code: `var Hello = <div>café's</div>`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    // UTF-16 column stays correct across a surrogate pair (supra-BMP char)
    {
      code: `var Hello = <div>🚀's</div>`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    // JSX in type assertion
    {
      code: `var Hello = (<div>a's</div>) as any`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    // JSX in logical short-circuit
    {
      code: `var Hello = cond && <div>a's</div>`,
      errors: [{ messageId: 'unescapedEntityAlts', data: { entity: "'" } }],
    },
    // Two JsxText nodes split by an expression container
    {
      code: `var Hello = <div>a's {x} b's</div>`,
      errors: [
        { messageId: 'unescapedEntityAlts', data: { entity: "'" } },
        { messageId: 'unescapedEntityAlts', data: { entity: "'" } },
      ],
    },
  ],
});
