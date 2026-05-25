import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-child-element-spacing', {} as never, {
  valid: [
    // ---- Upstream: text-only child ----
    {
      code: `
        <App>
          foo
        </App>
      `,
    },
    // ---- Upstream: text-only fragment ----
    {
      code: `
        <>
          foo
        </>
      `,
    },
    // ---- Upstream: single inline element child (no surrounding text) ----
    {
      code: `
        <App>
          <a>bar</a>
        </App>
      `,
    },
    // ---- Upstream: nested inline elements ----
    {
      code: `
        <App>
          <a>
            <b>nested</b>
          </a>
        </App>
      `,
    },
    // ---- Upstream: multiple text-only lines ----
    {
      code: `
        <App>
          foo
          bar
        </App>
      `,
    },
    // ---- Upstream: text and inline element on same line ----
    {
      code: `
        <App>
          foo<a>bar</a>baz
        </App>
      `,
    },
    // ---- Upstream: explicit {' '} between text and inline element ----
    {
      code: `
        <App>
          foo
          {' '}
          <a>bar</a>
          {' '}
          baz
        </App>
      `,
    },
    // ---- Upstream: {' '} on same line as inline element ----
    {
      code: `
        <App>
          foo
          {' '}<a>bar</a>{' '}
          baz
        </App>
      `,
    },
    // ---- Upstream: {' '} attached to text ----
    {
      code: `
        <App>
          foo{' '}
          <a>bar</a>
          {' '}baz
        </App>
      `,
    },
    // ---- Upstream: comment expression containers as spacing markers ----
    {
      code: `
        <App>
          foo{/*
          */}<a>bar</a>{/*
          */}baz
        </App>
      `,
    },
    // ---- Upstream: sentence with inline link, all on one line ----
    {
      code: `
        <App>
          Please take a look at <a href="https://js.org">this link</a>.
        </App>
      `,
    },
    // ---- Upstream: sentence with explicit {' '} before linked text ----
    {
      code: `
        <App>
          Please take a look at
          {' '}
          <a href="https://js.org">this link</a>.
        </App>
      `,
    },
    // ---- Upstream: block-level <p> elements (not inline) ----
    {
      code: `
        <App>
          <p>A</p>
          <p>B</p>
        </App>
      `,
    },
    // ---- Upstream: block-level elements adjacent ----
    {
      code: `
        <App>
          <p>A</p><p>B</p>
        </App>
      `,
    },
    // ---- Upstream: inline elements separated by whitespace-only text ----
    {
      code: `
        <App>
          <a>foo</a>
          <a>bar</a>
        </App>
      `,
    },
    // ---- Upstream: <br/> is not in inline set ----
    {
      code: `
        <App>
          A
          <br/>
          B
        </App>
      `,
    },
    // ---- Upstream: <br/> tightly between text ----
    {
      code: `
        <App>
          A<br/>B
        </App>
      `,
    },
    // ---- Edge: dotted tag name (Foo.Bar) is never treated as inline ----
    {
      code: `
        <App>
          foo
          <Foo.Bar>bar</Foo.Bar>
        </App>
      `,
    },
    // ---- Edge: capitalized custom component is not in the inline set ----
    {
      code: `
        <App>
          foo
          <Link>bar</Link>
        </App>
      `,
    },
    // ---- Edge: empty element ----
    {
      code: `<App></App>`,
    },
    // ---- Edge: self-closing inline element child has no children to walk ----
    {
      code: `<App><a /></App>`,
    },
    // ---- Edge: self-closing inline (<img/>) on same line as text ----
    {
      code: `<App>foo<img/>bar</App>`,
    },
    // ---- Edge: nested inline span tightly on one line ----
    {
      code: `<App><p>text <span>x</span> text</p></App>`,
    },
    // ---- Edge: block-level <section> not flagged on adjacent lines ----
    {
      code: `
        <App>
          intro
          <section>body</section>
        </App>
      `,
    },
    // ---- Edge: <br/> not in inline set, multi-line valid ----
    {
      code: `
        <App>
          A
          <br/>
          B
        </App>
      `,
    },
    // ---- Edge: real-world prose with code reference ----
    {
      code: `
        <p>
          Use <code>useState</code> to track state.
        </p>
      `,
    },
    // ---- Edge: JsxExpression as separator between text and inline ----
    {
      code: `
        <App>
          foo
          {variable}
          <a>bar</a>
        </App>
      `,
    },
    // ---- Edge: namespaced tag (a:b) is not an Identifier — not inline ----
    {
      code: `
        <App>
          foo
          <a:b>bar</a:b>
        </App>
      `,
    },
  ],
  invalid: [
    // ---- Upstream: text before inline element on a new line ----
    {
      code: `
        <App>
          foo
          <a>bar</a>
        </App>
      `,
      errors: [{ messageId: 'spacingBeforeNext' }],
    },
    // ---- Upstream: same as above but inside a fragment ----
    {
      code: `
        <>
          foo
          <a>bar</a>
        </>
      `,
      errors: [{ messageId: 'spacingBeforeNext' }],
    },
    // ---- Upstream: inline element followed by text on a new line ----
    {
      code: `
        <App>
          <a>bar</a>
          baz
        </App>
      `,
      errors: [{ messageId: 'spacingAfterPrev' }],
    },
    // ---- Upstream: explicit {' '} before inline element doesn't suppress
    // the spacingAfterPrev report when the next text starts with newline ----
    {
      code: `
        <App>
          {' '}<a>bar</a>
          baz
        </App>
      `,
      errors: [{ messageId: 'spacingAfterPrev' }],
    },
    // ---- Upstream: sentence split before inline link ----
    {
      code: `
        <App>
          Please take a look at
          <a href="https://js.org">this link</a>.
        </App>
      `,
      errors: [{ messageId: 'spacingBeforeNext' }],
    },
    // ---- Upstream: text between two inline <code> elements anchors on the
    // SECOND element via spacingBeforeNext ----
    {
      code: `
        <App>
          Some <code>loops</code> and some
          <code>if</code> statements.
        </App>
      `,
      errors: [{ messageId: 'spacingBeforeNext' }],
    },
    // ---- Upstream: two separate spacingBeforeNext reports in a row ----
    {
      code: `
        <App>
          Here is
          <a href="https://js.org">a link</a> and here is
          <a href="https://js.org">another</a>
        </App>
      `,
      errors: [
        { messageId: 'spacingBeforeNext' },
        { messageId: 'spacingBeforeNext' },
      ],
    },
    // ---- Edge: spacingAfterPrev inside a JsxFragment ----
    {
      code: `
        <>
          <a>bar</a>
          baz
        </>
      `,
      errors: [{ messageId: 'spacingAfterPrev' }],
    },
    // ---- Edge: text between two inline elements where the text matches the
    // FOLLOWING pattern (starts with newline + content) — anchors on the FIRST
    // element via spacingAfterPrev ----
    {
      code: `
        <App>
          <a>x</a>
          mid <b>y</b>
        </App>
      `,
      errors: [{ messageId: 'spacingAfterPrev' }],
    },
    // ---- Edge: self-closing inline element followed by text on a new line ----
    {
      code: `
        <App>
          <img/>
          baz
        </App>
      `,
      errors: [{ messageId: 'spacingAfterPrev' }],
    },
    // ---- Edge: text before self-closing inline element on a new line ----
    {
      code: `
        <App>
          foo
          <input/>
        </App>
      `,
      errors: [{ messageId: 'spacingBeforeNext' }],
    },
    // ---- Edge: nested element reports independently — only inner <p> matches ----
    {
      code: `
        <App>
          <p>
            foo
            <a>bar</a>
          </p>
        </App>
      `,
      errors: [{ messageId: 'spacingBeforeNext' }],
    },
    // ---- Edge: real-world arrow component with bad link spacing ----
    {
      code: `
        const Doc = () => (
          <p>
            Read more
            <a href="/docs">here</a>
          </p>
        );
      `,
      errors: [{ messageId: 'spacingBeforeNext' }],
    },
    // ---- Edge: outer + inner JsxElement BOTH host a bug — two diagnostics ----
    {
      code: `
        <App>
          <p>
            foo
            <a>bar</a>
          </p>
          <span>x</span>
          trailing
        </App>
      `,
      errors: [
        { messageId: 'spacingAfterPrev' },
        { messageId: 'spacingBeforeNext' },
      ],
    },
    // ---- Edge: textarea (uncommon inline tag) before text on a new line ----
    {
      code: `
        <App>
          foo
          <textarea/>
        </App>
      `,
      errors: [{ messageId: 'spacingBeforeNext' }],
    },
  ],
});
