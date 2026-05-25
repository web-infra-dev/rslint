// cspell:ignore asdjfl
package jsx_no_comment_textnodes

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxNoCommentTextnodesRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxNoCommentTextnodesRule, []rule_tester.ValidTestCase{
		// ---- Upstream: expression-container comment inside JsxElement ----
		{Code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
                {/* valid */}
              </div>
            );
          }
        }
      `, Tsx: true},

		// ---- Upstream: expression-container comment inside JsxFragment ----
		{Code: `
        class Comp1 extends Component {
          render() {
            return (
              <>
                {/* valid */}
              </>
            );
          }
        }
      `, Tsx: true},

		// ---- Upstream: inline expression-container comment ----
		{Code: `
        class Comp1 extends Component {
          render() {
            return (<div>{/* valid */}</div>);
          }
        }
      `, Tsx: true},

		// ---- Upstream: parenthesized JSX assigned to variable ----
		{Code: `
        class Comp1 extends Component {
          render() {
            const bar = (<div>{/* valid */}</div>);
            return bar;
          }
        }
      `, Tsx: true},

		// ---- Upstream: createReactClass property is JSX with embedded comment ----
		{Code: `
        var Hello = createReactClass({
          foo: (<div>{/* valid */}</div>),
          render() {
            return this.foo;
          },
        });
      `, Tsx: true},

		// ---- Upstream: multiple valid expression-container comments ----
		{Code: `
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
      `, Tsx: true},

		// ---- Upstream: empty element ----
		{Code: `
        class Comp1 extends Component {
          render() {
            return (
              <div>
              </div>
            );
          }
        }
      `, Tsx: true},

		// ---- Upstream: non-JSX source ----
		{Code: `
        var foo = require('foo');
      `, Tsx: true},

		// ---- Upstream: expression-container comment with attribute on parent ----
		{Code: `
        <Foo bar='test'>
          {/* valid */}
        </Foo>
      `, Tsx: true},

		// ---- Upstream: URL-like text with `//` embedded (not at line start) ----
		{Code: `
        <strong>
          &nbsp;https://www.example.com/attachment/download/1
        </strong>
      `, Tsx: true},

		// ---- Upstream: leading-trivia comment inside element declaration ----
		{Code: `
        <Foo /* valid */ placeholder={'foo'}/>
      `, Tsx: true},

		// ---- Upstream: leading-trivia comment between fragment tokens ----
		{Code: `
        </* valid */></>
      `, Tsx: true},

		// ---- Upstream: leading-trivia comment inside attribute expression ----
		{Code: `
        <Foo title={'foo' /* valid */}/>
      `, Tsx: true},

		// ---- Upstream: HTML-entity `//` is NOT decoded before the regex test ----
		{Code: `<pre>&#x2F;&#x2F; TODO: Write perfect code</pre>`, Tsx: true},

		// ---- Upstream: HTML-entity `/* ... */` is NOT decoded before the regex test ----
		{Code: `<pre>&#x2F;&#42; TODO: Write perfect code &#42;&#x2F;</pre>`, Tsx: true},

		// ---- Upstream: `//` entities inside a nested <span> with sibling text ----
		{Code: `
        <div>
          <span className="pl-c"><span className="pl-c">&#47;&#47;</span> ...</span><br />
        </div>
      `, Tsx: true},

		// ---- Edge: text contains `//` mid-line (not at line start) ----
		{Code: `<div>value // trailing</div>`, Tsx: true},

		// ---- Edge: purely whitespace text between siblings ----
		{Code: `
        <div>
          <span/>
          <span/>
        </div>
      `, Tsx: true},

		// ---- Edge: single slash is not comment-like ----
		{Code: `<div>/ not a comment</div>`, Tsx: true},

		// ---- Edge: `//` appears mid-line and end-of-text, no line-start occurrence ----
		{Code: `<div>hello //</div>`, Tsx: true},

		// ---- Edge: `/ *` (space between) is not a block-comment open ----
		{Code: `<div>/ * not a block comment</div>`, Tsx: true},

		// ---- Edge: JSX inside a conditional expression — text has no `//` ----
		{Code: `
        const render = (ok) => ok ? <div>ok</div> : <div>fail</div>;
      `, Tsx: true},

		// ---- Edge: JSX inside a .map callback — text has no `//` ----
		{Code: `
        const list = items.map((i) => <li key={i}>{i}</li>);
      `, Tsx: true},

		// ---- Edge: JsxText between sibling JsxElements carries `//` INSIDE an
		// expression container, not as bare text — still valid ----
		{Code: `
        <div>
          <span />
          {'// string literal is fine'}
          <span />
        </div>
      `, Tsx: true},

		// ---- Edge: deeply-nested JSX without any comment-like text ----
		{Code: `
        <div><section><article><p>hello world</p></article></section></div>
      `, Tsx: true},

		// ---- Edge: member-expression tag (Foo.Bar) with valid embedded comment ----
		{Code: `<Foo.Bar>{/* valid */}</Foo.Bar>`, Tsx: true},

		// ---- SKIP: ESLint accepts `<></* valid *//>` as a self-closing fragment
		// with a leading-trivia comment between `<>` and `</>`. tsgo's parser
		// rejects this as invalid JSX syntax, so there is no AST for the rule
		// to inspect. Tracked upstream as `features: ['no-ts']`.
		{Code: `<></* valid *//>`, Tsx: true, Skip: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: inline `// invalid` in JsxElement ----
		{
			Code: `
        class Comp1 extends Component {
          render() {
            return (<div>// invalid</div>);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 4, Column: 26},
			},
		},

		// ---- Upstream: inline `// invalid` in JsxFragment ----
		{
			Code: `
        class Comp1 extends Component {
          render() {
            return (<>// invalid</>);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 4, Column: 23},
			},
		},

		// ---- Upstream: inline `/* invalid */` in JsxElement ----
		{
			Code: `
        class Comp1 extends Component {
          render() {
            return (<div>/* invalid */</div>);
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 4, Column: 26},
			},
		},

		// ---- Upstream: multi-line `// invalid` ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 5, Column: 20},
			},
		},

		// ---- Upstream: `/* invalid */` surrounded by text lines ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 5, Column: 20},
			},
		},

		// ---- Upstream: `// invalid` between two expression containers splits
		// JsxText into three — only the middle text node reports ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 6, Column: 27},
			},
		},

		// ---- Upstream: arrow function returning JSX with `/*` text ----
		{
			Code: `
        const Component2 = () => {
          return <span>/*</span>;
        };
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces",
					Message: "Comments inside children section of tag should be placed inside braces",
					Line:    3, Column: 24},
			},
		},

		// ---- Edge: comment-like text in a JsxFragment across multiple lines ----
		{
			Code: `
        const C = () => (
          <>
            // invalid
          </>
        );
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 3, Column: 13},
			},
		},

		// ---- Edge: two sibling JsxElements, each with its own comment-like
		// text — both should report (two separate JsxText nodes) ----
		{
			Code: `
        const C = () => (
          <div>
            <span>// a</span>
            <span>// b</span>
          </div>
        );
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 4, Column: 19},
				{MessageId: "putCommentInBraces", Line: 5, Column: 19},
			},
		},

		// ---- Edge: nested JsxElement — the INNER <span> hosts the bad text,
		// so exactly one diagnostic fires even though outer JsxText also
		// surrounds it ----
		{
			Code: `
        const C = () => (
          <div>
            <span>/* bad */</span>
          </div>
        );
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 4, Column: 19},
			},
		},

		// ---- Edge: deeply-nested JsxElement (5 levels), bad text at innermost ----
		{
			Code: `
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
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 7, Column: 25},
			},
		},

		// ---- Edge: same JsxText contains BOTH `//` and `/*` — still a single
		// diagnostic (one diagnostic per JsxText node, regardless of how many
		// comment markers fall inside it) ----
		{
			Code: `
        const C = () => (
          <div>
            // first
            /* second */
          </div>
        );
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 3, Column: 16},
			},
		},

		// ---- Edge: tab-indented comment line (upstream regex uses `\s*`, so
		// tabs qualify as leading whitespace) ----
		{
			Code: "<div>\n\t\t// tab-indented\n</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 1, Column: 6},
			},
		},

		// ---- Edge: JSX in a conditional expression, bad text in one branch ----
		{
			Code: `
        const render = (ok) => ok ? <div>// bad</div> : <div>ok</div>;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 2, Column: 42},
			},
		},

		// ---- Edge: JSX inside `.map` callback with bad text ----
		{
			Code: `
        const list = items.map((i) => <li key={i}>// {i}</li>);
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 2, Column: 51},
			},
		},

		// ---- Edge: member-expression tag (Foo.Bar) with bad text ----
		{
			Code: `<Foo.Bar>// bad</Foo.Bar>`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 1, Column: 10},
			},
		},

		// ---- Edge: JsxText with bad text split across multiple non-leading
		// lines (content has text on first line, then `//` on a later line) ----
		{
			Code: `
        const C = () => (
          <div>
            line1
            line2
            // bad
          </div>
        );
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 3, Column: 16},
			},
		},

		// ---- Edge: NBSP (U+00A0) as leading whitespace before `//`. ESLint's
		// `\s` covers Unicode WhiteSpace — mirrored via `unicode.Is(Zs, r)`. ----
		{
			Code: "<div>\n\u00A0\u00A0// bad\n</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 1, Column: 6},
			},
		},

		// ---- Edge: LINE SEPARATOR (U+2028) acts as a line boundary under
		// ECMAScript `/m` — so text after it should be scanned for `//`. ----
		{
			Code: "<div>\u2028// bad</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 1, Column: 6},
			},
		},

		// ---- Edge: IDEOGRAPHIC SPACE (U+3000, Zs category) as indent ----
		{
			Code: "<div>\n\u3000// bad\n</div>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 1, Column: 6},
			},
		},

		// ---- Edge: multi-byte characters (CJK, 3 UTF-8 bytes / 1 UTF-16 unit)
		// preceding the JsxText. Locks in that node.Pos()/End() are UTF-8 byte
		// offsets into SourceFile.Text() — slicing must stay byte-aligned while
		// the reported column remains in UTF-16 units (matching ESLint). ----
		{
			Code: "const x = '你好';\nconst y = <div>// bad</div>;\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 2, Column: 16},
			},
		},

		// ---- Edge: CJK inside the JsxText, followed by `//` on the next line.
		// The slice must include the multi-byte prefix without corrupting it. ----
		{
			Code: "const C = () => <div>你好\n// bad\n</div>;\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 1, Column: 22},
			},
		},

		// ---- Edge: emoji (4 UTF-8 bytes / 2 UTF-16 surrogate units) before
		// the JsxText. Verifies the byte-slice path on supplementary-plane
		// characters too. ----
		{
			Code: "const e = '🚀';\nconst y = <div>/* bad */</div>;\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "putCommentInBraces", Line: 2, Column: 16},
			},
		},
	})
}
