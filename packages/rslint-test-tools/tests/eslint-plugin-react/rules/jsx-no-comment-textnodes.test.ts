import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectPutCommentInBracesError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'putCommentInBraces',
    message:
      'Comments inside children section of tag should be placed inside braces',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-comment-textnodes', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            const bar = (<div>{/* valid */}</div>);
            return bar;
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                {/* valid */}
                {/* valid 2 */}
                {/* valid 3 */}
              </div>
            );
          }
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        var foo = require('foo');
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        <Foo bar='test'>
          {/* valid */}
        </Foo>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        </* valid */></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      skip: 'The native TypeScript parser rejects this JSX form (Expected corresponding closing tag for JSX fragment.); upstream uses a different parser.',
      code: `
        <></* valid *//>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        <Foo title={'foo' /* valid */}/>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        <div>
          <span className="pl-c"><span className="pl-c">&#47;&#47;</span> ...</span><br />
        </div>
      `,
      languageOptions: upstreamModuleOptions,
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
      errors: [expectPutCommentInBracesError([4, 26, 4, 36])],
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
      errors: [expectPutCommentInBracesError([4, 23, 4, 33])],
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
      errors: [expectPutCommentInBracesError([4, 26, 4, 39])],
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
      errors: [expectPutCommentInBracesError([5, 20, 7, 15])],
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
      errors: [expectPutCommentInBracesError([5, 20, 9, 15])],
    },
    // ---- Upstream: arrow function returning JSX with `/*` text ----
    {
      code: `
        const Component2 = () => {
          return <span>/*</span>;
        };
      `,
      errors: [expectPutCommentInBracesError([3, 24, 3, 26])],
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
        expectPutCommentInBracesError([4, 19, 4, 23]),
        expectPutCommentInBracesError([5, 19, 5, 23]),
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
      errors: [expectPutCommentInBracesError([7, 25, 7, 32])],
    },
    // ---- Edge: JSX in conditional expression — bad text in one branch ----
    {
      code: `const render = (ok) => ok ? <div>// bad</div> : <div>ok</div>;`,
      errors: [expectPutCommentInBracesError([1, 34, 1, 40])],
    },
    // ---- Edge: JSX in .map callback with bad text ----
    {
      code: `const list = items.map((i) => <li key={i}>// {i}</li>);`,
      errors: [expectPutCommentInBracesError([1, 43, 1, 46])],
    },
    // ---- Edge: member-expression tag with bad text ----
    {
      code: `<Foo.Bar>// bad</Foo.Bar>`,
      errors: [expectPutCommentInBracesError([1, 10, 1, 16])],
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
      errors: [expectPutCommentInBracesError([3, 16, 6, 11])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                {'asdjfl'}
                // invalid
                {'foo'}
              </div>
            );
          }
        }
      `,
      errors: [expectPutCommentInBracesError([6, 27, 8, 17])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
