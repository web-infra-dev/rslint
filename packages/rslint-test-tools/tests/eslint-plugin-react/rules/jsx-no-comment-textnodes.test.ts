import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-comment-textnodes', {} as never, {
  valid: [
    // ---- Upstream: expression-container comment inside JsxElement ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                {/* valid */}
              </div>
            );
          }
        }
      `,
    },
    // ---- Upstream: expression-container comment inside JsxFragment ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (
              <>
                {/* valid */}
              </>
            );
          }
        }
      `,
    },
    // ---- Upstream: inline expression-container comment ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (<div>{/* valid */}</div>);
          }
        }
      `,
    },
    // ---- Upstream: createReactClass property is JSX with embedded comment ----
    {
      code: `
        var Hello = createReactClass({
          foo: (<div>{/* valid */}</div>),
          render() {
            return this.foo;
          },
        });
      `,
    },
    // ---- Upstream: empty element ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
              </div>
            );
          }
        }
      `,
    },
    // ---- Upstream: URL-like text with `//` embedded (not at line start) ----
    {
      code: `
        <strong>
          &nbsp;https://www.example.com/attachment/download/1
        </strong>
      `,
    },
    // ---- Upstream: leading-trivia comment inside element declaration ----
    {
      code: `
        <Foo /* valid */ placeholder={'foo'}/>
      `,
    },
    // ---- Upstream: HTML-entity `//` is NOT decoded before the regex test ----
    {
      code: `<pre>&#x2F;&#x2F; TODO: Write perfect code</pre>`,
    },
    // ---- Upstream: HTML-entity `/* ... */` is NOT decoded ----
    {
      code: `<pre>&#x2F;&#42; TODO: Write perfect code &#42;&#x2F;</pre>`,
    },
    // ---- Edge: text contains `//` mid-line (not at line start) ----
    {
      code: `<div>value // trailing</div>`,
    },
    // ---- Edge: single slash is not a comment marker ----
    {
      code: `<div>/ not a comment</div>`,
    },
    // ---- Edge: `/ *` (space between) is not a block-comment open ----
    {
      code: `<div>/ * not a block comment</div>`,
    },
    // ---- Edge: member-expression tag with valid embedded comment ----
    {
      code: `<Foo.Bar>{/* valid */}</Foo.Bar>`,
    },
    // ---- Edge: JSX in .map callback, no bad text ----
    {
      code: `const list = items.map((i) => <li key={i}>{i}</li>);`,
    },
  ],
  invalid: [
    // ---- Upstream: inline `// invalid` in JsxElement ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (<div>// invalid</div>);
          }
        }
      `,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Upstream: inline `// invalid` in JsxFragment ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (<>// invalid</>);
          }
        }
      `,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Upstream: inline `/* invalid */` in JsxElement ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (<div>/* invalid */</div>);
          }
        }
      `,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Upstream: multi-line `// invalid` ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                // invalid
              </div>
            );
          }
        }
      `,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Upstream: `/* invalid */` surrounded by text lines ----
    {
      code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                asdjfl
                /* invalid */
                foo
              </div>
            );
          }
        }
      `,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Upstream: arrow function returning JSX with `/*` text ----
    {
      code: `
        const Component2 = () => {
          return <span>/*</span>;
        };
      `,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Edge: two sibling JsxElements, each with its own comment-like text ----
    {
      code: `
        const C = () => (
          <div>
            <span>// a</span>
            <span>// b</span>
          </div>
        );
      `,
      errors: [
        { messageId: 'putCommentInBraces' },
        { messageId: 'putCommentInBraces' },
      ],
    },
    // ---- Edge: deeply-nested JSX with bad text at innermost ----
    {
      code: `
        const C = () => (
          <section>
            <article>
              <header>
                <h1>
                  <span>// deep</span>
                </h1>
              </header>
            </article>
          </section>
        );
      `,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Edge: JSX in conditional expression — bad text in one branch ----
    {
      code: `const render = (ok) => ok ? <div>// bad</div> : <div>ok</div>;`,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Edge: JSX in .map callback with bad text ----
    {
      code: `const list = items.map((i) => <li key={i}>// {i}</li>);`,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Edge: member-expression tag with bad text ----
    {
      code: `<Foo.Bar>// bad</Foo.Bar>`,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
    // ---- Edge: same JsxText carries BOTH `//` and `/*` → single diagnostic ----
    {
      code: `
        const C = () => (
          <div>
            // first
            /* second */
          </div>
        );
      `,
      errors: [{ messageId: 'putCommentInBraces' }],
    },
  ],
});
