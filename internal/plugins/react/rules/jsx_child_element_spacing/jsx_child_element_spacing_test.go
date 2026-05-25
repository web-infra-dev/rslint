package jsx_child_element_spacing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxChildElementSpacingRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxChildElementSpacingRule, []rule_tester.ValidTestCase{
		// ---- Upstream: text-only child ----
		{Code: `
        <App>
          foo
        </App>
      `, Tsx: true},

		// ---- Upstream: text-only fragment ----
		{Code: `
        <>
          foo
        </>
      `, Tsx: true},

		// ---- Upstream: single inline element child (no surrounding text) ----
		{Code: `
        <App>
          <a>bar</a>
        </App>
      `, Tsx: true},

		// ---- Upstream: nested inline elements ----
		{Code: `
        <App>
          <a>
            <b>nested</b>
          </a>
        </App>
      `, Tsx: true},

		// ---- Upstream: multiple text-only lines ----
		{Code: `
        <App>
          foo
          bar
        </App>
      `, Tsx: true},

		// ---- Upstream: text and inline element on same line ----
		{Code: `
        <App>
          foo<a>bar</a>baz
        </App>
      `, Tsx: true},

		// ---- Upstream: explicit {' '} between text and inline element ----
		{Code: `
        <App>
          foo
          {' '}
          <a>bar</a>
          {' '}
          baz
        </App>
      `, Tsx: true},

		// ---- Upstream: {' '} on same line as inline element ----
		{Code: `
        <App>
          foo
          {' '}<a>bar</a>{' '}
          baz
        </App>
      `, Tsx: true},

		// ---- Upstream: {' '} attached to text ----
		{Code: `
        <App>
          foo{' '}
          <a>bar</a>
          {' '}baz
        </App>
      `, Tsx: true},

		// ---- Upstream: comment expression containers as spacing markers ----
		{Code: `
        <App>
          foo{/*
          */}<a>bar</a>{/*
          */}baz
        </App>
      `, Tsx: true},

		// ---- Upstream: sentence with inline link, all on one line ----
		{Code: `
        <App>
          Please take a look at <a href="https://js.org">this link</a>.
        </App>
      `, Tsx: true},

		// ---- Upstream: sentence with explicit {' '} before linked text ----
		{Code: `
        <App>
          Please take a look at
          {' '}
          <a href="https://js.org">this link</a>.
        </App>
      `, Tsx: true},

		// ---- Upstream: block-level <p> elements (not inline) ----
		{Code: `
        <App>
          <p>A</p>
          <p>B</p>
        </App>
      `, Tsx: true},

		// ---- Upstream: block-level elements adjacent ----
		{Code: `
        <App>
          <p>A</p><p>B</p>
        </App>
      `, Tsx: true},

		// ---- Upstream: inline elements separated by whitespace-only text
		// (no `\S` between newlines, so neither pattern fires) ----
		{Code: `
        <App>
          <a>foo</a>
          <a>bar</a>
        </App>
      `, Tsx: true},

		// ---- Upstream: nested inline elements with whitespace siblings ----
		{Code: `
        <App>
          <a>
            <b>nested1</b>
            <b>nested2</b>
          </a>
        </App>
      `, Tsx: true},

		// ---- Upstream: text-only single-letter lines ----
		{Code: `
        <App>
          A
          B
        </App>
      `, Tsx: true},

		// ---- Upstream: <br/> is not in inline set, so surrounding text is fine ----
		{Code: `
        <App>
          A
          <br/>
          B
        </App>
      `, Tsx: true},

		// ---- Upstream: <br/> after text on same line, then text on next line ----
		{Code: `
        <App>
          A<br/>
          B
        </App>
      `, Tsx: true},

		// ---- Upstream: <br/> tightly between text ----
		{Code: `
        <App>
          A<br/>B
        </App>
      `, Tsx: true},

		// ---- Upstream: single-line all-tight ----
		{Code: `
        <App>A<br/>B</App>
      `, Tsx: true},

		// ---- Edge: dotted tag name (Foo.Bar) is never treated as inline ----
		{Code: `
        <App>
          foo
          <Foo.Bar>bar</Foo.Bar>
        </App>
      `, Tsx: true},

		// ---- Edge: capitalized custom component is not in the inline set ----
		{Code: `
        <App>
          foo
          <Link>bar</Link>
        </App>
      `, Tsx: true},

		// ---- Edge: fragment children with all-text ----
		{Code: `
        <>
          A
          B
        </>
      `, Tsx: true},

		// ---- Edge: empty element ----
		{Code: `<App></App>`, Tsx: true},

		// ---- Edge: self-closing inline element child has no children to walk ----
		{Code: `<App><a /></App>`, Tsx: true},

		// ---- Edge: text contains a newline mid-string but no leading/trailing
		// newline against an inline element — no report ----
		{Code: `<App>foo
bar<a>x</a></App>`, Tsx: true},

		// ---- Edge: self-closing inline element (<img/>) on same line as text ----
		{Code: `<App>foo<img/>bar</App>`, Tsx: true},

		// ---- Edge: self-closing inline element with explicit {' '} ----
		{Code: `
        <App>
          foo
          {' '}
          <img/>
          {' '}
          bar
        </App>
      `, Tsx: true},

		// ---- Edge: text + self-closing inline element on same line ----
		{Code: `
        <App>
          foo<img/>
        </App>
      `, Tsx: true},

		// ---- Edge: deeply nested mix of block and inline with no ambiguity ----
		{Code: `
        <article>
          <section>
            <p>
              hello <span>world</span> today
            </p>
          </section>
        </article>
      `, Tsx: true},

		// ---- Edge: namespaced JSX tag (svg:circle) is not an Identifier, so
		// not in inline set — adjacent-line text should NOT be flagged ----
		{Code: `
        <App>
          foo
          <a:b>bar</a:b>
        </App>
      `, Tsx: true},

		// ---- Edge: <this/> tag (ThisKeyword base) — not an Identifier, not inline ----
		{Code: `
        <App>
          foo
          <this/>
        </App>
      `, Tsx: true},

		// ---- Edge: only text-only children (no inline neighbor) — guard against
		// the "(lastChild || nextChild)" being false on bare text in a fragment ----
		{Code: `<>foo</>`, Tsx: true},

		// ---- Edge: triple-element window where the middle text is whitespace
		// only between two inline elements — neither pattern matches \S ----
		{Code: `<App><a>x</a> <b>y</b></App>`, Tsx: true},

		// ---- Edge: nested JsxElement listener fires on inner element too —
		// inner uses a block-level child (<p>), so no inline-pairing matches. ----
		{Code: `
        <App>
          <div>
            <p>A</p>
            <p>B</p>
          </div>
        </App>
      `, Tsx: true},

		// ---- Edge: nested inline span tightly surrounded by text on the same
		// line — neither pattern matches because no `\n` boundary exists. ----
		{Code: `
        <App>
          <p>text <span>x</span> text</p>
        </App>
      `, Tsx: true},

		// ---- Inline-set coverage: every tag in the upstream set must be
		// recognized as inline; spacing them with explicit {' '} on the same
		// line keeps the input valid while exercising the lookup. ----
		{Code: `<App>x<a>y</a>z</App>`, Tsx: true},
		{Code: `<App>x<abbr>y</abbr>z</App>`, Tsx: true},
		{Code: `<App>x<acronym>y</acronym>z</App>`, Tsx: true},
		{Code: `<App>x<b>y</b>z</App>`, Tsx: true},
		{Code: `<App>x<bdo>y</bdo>z</App>`, Tsx: true},
		{Code: `<App>x<big>y</big>z</App>`, Tsx: true},
		{Code: `<App>x<button>y</button>z</App>`, Tsx: true},
		{Code: `<App>x<cite>y</cite>z</App>`, Tsx: true},
		{Code: `<App>x<code>y</code>z</App>`, Tsx: true},
		{Code: `<App>x<dfn>y</dfn>z</App>`, Tsx: true},
		{Code: `<App>x<em>y</em>z</App>`, Tsx: true},
		{Code: `<App>x<i>y</i>z</App>`, Tsx: true},
		{Code: `<App>x<img/>z</App>`, Tsx: true},
		{Code: `<App>x<input/>z</App>`, Tsx: true},
		{Code: `<App>x<kbd>y</kbd>z</App>`, Tsx: true},
		{Code: `<App>x<label>y</label>z</App>`, Tsx: true},
		{Code: `<App>x<map>y</map>z</App>`, Tsx: true},
		{Code: `<App>x<object>y</object>z</App>`, Tsx: true},
		{Code: `<App>x<q>y</q>z</App>`, Tsx: true},
		{Code: `<App>x<samp>y</samp>z</App>`, Tsx: true},
		{Code: `<App>x<script>y</script>z</App>`, Tsx: true},
		{Code: `<App>x<select>y</select>z</App>`, Tsx: true},
		{Code: `<App>x<small>y</small>z</App>`, Tsx: true},
		{Code: `<App>x<span>y</span>z</App>`, Tsx: true},
		{Code: `<App>x<strong>y</strong>z</App>`, Tsx: true},
		{Code: `<App>x<sub>y</sub>z</App>`, Tsx: true},
		{Code: `<App>x<sup>y</sup>z</App>`, Tsx: true},
		{Code: `<App>x<textarea>y</textarea>z</App>`, Tsx: true},
		{Code: `<App>x<tt>y</tt>z</App>`, Tsx: true},
		{Code: `<App>x<var>y</var>z</App>`, Tsx: true},

		// ---- Block-level negative coverage: common HTML block tags must NOT
		// trigger the rule even with text on adjacent lines. ----
		{Code: `
        <main>
          intro
          <section>body</section>
        </main>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <article>body</article>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <header>body</header>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <footer>body</footer>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <nav>body</nav>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <aside>body</aside>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <div>body</div>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <ul>body</ul>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <li>item</li>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          intro
          <h1>title</h1>
        </App>
      `, Tsx: true},

		// ---- Edge: JsxExpression child between text and inline element acts
		// as a separator — the 3-element window for the inline pair never
		// includes a JsxText, so no report fires. ----
		{Code: `
        <App>
          foo
          {variable}
          <a>bar</a>
        </App>
      `, Tsx: true},

		// ---- Edge: numeric / boolean / null JsxExpression values — same. ----
		{Code: `
        <App>
          foo
          {0}
          <a>bar</a>
        </App>
      `, Tsx: true},
		{Code: `
        <App>
          foo
          {null}
          <a>bar</a>
        </App>
      `, Tsx: true},

		// ---- Edge: JSX in conditional expression — neither branch ambiguous. ----
		{Code: `
        const r = (ok) => ok ? <App>foo<a>x</a>bar</App> : <App>baz</App>;
      `, Tsx: true},

		// ---- Edge: JSX in .map callback — text sits on same line as inline. ----
		{Code: `
        items.map((i) => <li key={i}>before <a>{i}</a> after</li>);
      `, Tsx: true},

		// ---- Edge: JSX in array literal expression — no spacing ambiguity. ----
		{Code: `
        const a = [<span key="1">one</span>, <span key="2">two</span>];
      `, Tsx: true},

		// ---- Edge: JSX inside attribute value (recursion through attributes
		// must not blow up) — outer App has no children-spacing issue. ----
		{Code: `
        <App tooltip={<span>hint</span>}>body</App>
      `, Tsx: true},

		// ---- Edge: spread attribute on inline element — attributes don't
		// affect children-walk. ----
		{Code: `<App>foo<a {...props}>bar</a>baz</App>`, Tsx: true},

		// ---- Edge: TypeScript generic component (no inline tag involvement). ----
		{Code: `
        <List<string>>
          <a>x</a>
        </List>
      `, Tsx: true},

		// ---- Edge: text with mid-line CJK characters — no \n boundary, so no
		// pattern match. Locks in that JsxText.Text reads tsgo's text correctly. ----
		{Code: `<App>你好<a>link</a>世界</App>`, Tsx: true},

		// ---- Edge: text that contains ONLY a single newline between two
		// inline elements (no \S on either side). ----
		{Code: "<App><a>x</a>\n<b>y</b></App>", Tsx: true},

		// ---- Edge: text that contains ONLY whitespace+newline+whitespace
		// (no \S anywhere) — neither pattern fires. ----
		{Code: "<App><a>x</a>   \n   <b>y</b></App>", Tsx: true},

		// ---- Edge: CRLF line endings between inline elements + whitespace
		// only — Go regex treats \r as part of \s, so still no match. ----
		{Code: "<App><a>x</a>\r\n<b>y</b></App>", Tsx: true},

		// ---- Edge: real-world prose pattern — sentence with embedded link
		// that fits on one line. ----
		{Code: `
        <p>
          Click <a href="/x">here</a> to continue.
        </p>
      `, Tsx: true},

		// ---- Edge: real-world prose with code reference. ----
		{Code: `
        <p>
          Use <code>useState</code> to track state.
        </p>
      `, Tsx: true},

		// ---- Edge: trailing punctuation glued to inline element on same line. ----
		{Code: `
        <p>
          See <a>docs</a>.
        </p>
      `, Tsx: true},

		// ---- Edge: leading punctuation before inline element on same line. ----
		{Code: `<p>(<a>note</a>) trailing</p>`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream: text before inline element on a new line ----
		{
			Code: `
        <App>
          foo
          <a>bar</a>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11},
			},
		},

		// ---- Upstream: same as above but inside a fragment ----
		{
			Code: `
        <>
          foo
          <a>bar</a>
        </>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11},
			},
		},

		// ---- Upstream: inline element followed by text on a new line ----
		{
			Code: `
        <App>
          <a>bar</a>
          baz
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 3, Column: 21},
			},
		},

		// ---- Upstream: explicit {' '} BEFORE inline element doesn't suppress
		// the AFTER report when the next text starts with a newline ----
		{
			Code: `
        <App>
          {' '}<a>bar</a>
          baz
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 3, Column: 26},
			},
		},

		// ---- Upstream: sentence split before inline link ----
		{
			Code: `
        <App>
          Please take a look at
          <a href="https://js.org">this link</a>.
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11},
			},
		},

		// ---- Upstream: text between two inline `<code>` elements matches the
		// PRECEDING pattern (ends with newline + whitespace), not FOLLOWING —
		// so the report anchors on the second `<code>` ----
		{
			Code: `
        <App>
          Some <code>loops</code> and some
          <code>if</code> statements.
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11,
					Message: "Ambiguous spacing before next element code"},
			},
		},

		// ---- Upstream: two separate spacingBeforeNext reports in a row ----
		{
			Code: `
        <App>
          Here is
          <a href="https://js.org">a link</a> and here is
          <a href="https://js.org">another</a>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11},
				{MessageId: "spacingBeforeNext", Line: 5, Column: 11},
			},
		},

		// ---- Edge: spacingAfterPrev message text exact match ----
		{
			Code: `
        <App>
          <a>bar</a>
          baz
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 3, Column: 21,
					Message: "Ambiguous spacing after previous element a"},
			},
		},

		// ---- Edge: spacingBeforeNext message text exact match ----
		{
			Code: `
        <App>
          foo
          <a>bar</a>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11,
					Message: "Ambiguous spacing before next element a"},
			},
		},

		// ---- Edge: inline element at the end (followed by text-only) inside a
		// fragment — exercise spacingAfterPrev within JsxFragment too ----
		{
			Code: `
        <>
          <a>bar</a>
          baz
        </>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 3, Column: 21},
			},
		},

		// ---- Edge: text between two inline elements where the text matches
		// the FOLLOWING pattern (starts with newline + content) — anchors on
		// the FIRST element via spacingAfterPrev. ----
		{
			Code: `
        <App>
          <a>x</a>
          mid <b>y</b>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 3, Column: 19,
					Message: "Ambiguous spacing after previous element a"},
			},
		},

		// ---- Edge: self-closing inline element (<img/>) before text on a new
		// line — locks in tsgo's KindJsxSelfClosingElement equivalence with
		// ESTree's JSXElement for the inline check. ----
		{
			Code: `
        <App>
          <img/>
          baz
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 3, Column: 17,
					Message: "Ambiguous spacing after previous element img"},
			},
		},

		// ---- Edge: text before self-closing inline element on a new line ----
		{
			Code: `
        <App>
          foo
          <input/>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11,
					Message: "Ambiguous spacing before next element input"},
			},
		},

		// ---- Edge: paired AND self-closing inline siblings — text matches
		// PRECEDING, anchors on the self-closing element. ----
		{
			Code: `
        <App>
          Some <code>x</code> and
          <img/>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11,
					Message: "Ambiguous spacing before next element img"},
			},
		},

		// ---- Edge: nested JsxElement reports independently per scope — the
		// inner <p> hosts the bad text, so exactly one diagnostic fires from
		// the inner element's listener call. ----
		{
			Code: `
        <App>
          <p>
            foo
            <a>bar</a>
          </p>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 5, Column: 13},
			},
		},

		// ---- Edge: deeply nested (4 levels) with bad text at innermost ----
		{
			Code: `
        <App>
          <section>
            <article>
              <p>
                foo
                <a>bar</a>
              </p>
            </article>
          </section>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 7, Column: 17},
			},
		},

		// ---- Edge: fragment containing nested fragment with bad text — both
		// JsxFragment listeners fire, but only the inner one matches. ----
		{
			Code: `
        <>
          <>
            foo
            <a>bar</a>
          </>
        </>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 5, Column: 13},
			},
		},

		// ---- Edge: outer + inner element BOTH host the same kind of bug —
		// each fires its own diagnostic from its own listener call. ----
		{
			Code: `
        <App>
          <p>
            foo
            <a>bar</a>
          </p>
          <span>x</span>
          trailing
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 7, Column: 25},
				{MessageId: "spacingBeforeNext", Line: 5, Column: 13},
			},
		},

		// ---- Edge: three-inline-element chain separated by line-broken text
		// blocks — each text matches PRECEDING, two reports anchor on the
		// 2nd and 3rd inline elements. ----
		{
			Code: `
        <App>
          one
          <a>X</a> two
          <b>Y</b> three
          <em>Z</em>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11,
					Message: "Ambiguous spacing before next element a"},
				{MessageId: "spacingBeforeNext", Line: 5, Column: 11,
					Message: "Ambiguous spacing before next element b"},
				{MessageId: "spacingBeforeNext", Line: 6, Column: 11,
					Message: "Ambiguous spacing before next element em"},
			},
		},

		// ---- Edge: column reporting with multi-byte CJK before the inline
		// element on the same line — verifies that lastChild.End() / Pos()
		// produce the correct UTF-16 column the test framework expects. ----
		{
			Code: `
        <App>
          你好
          <a>x</a>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11,
					Message: "Ambiguous spacing before next element a"},
			},
		},

		// ---- Edge: trailing-text case where the previous-element loc.end is
		// reported correctly when followed by trailing punctuation on next line. ----
		{
			Code: `
        <App>
          <strong>important</strong>
          .
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 3, Column: 37,
					Message: "Ambiguous spacing after previous element strong"},
			},
		},

		// ---- Edge: button (uncommon inline tag) — sanity-check the inline set. ----
		{
			Code: `
        <App>
          <button>OK</button>
          baz
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 3, Column: 30,
					Message: "Ambiguous spacing after previous element button"},
			},
		},

		// ---- Edge: textarea (uncommon inline tag) — both directions. ----
		{
			Code: `
        <App>
          foo
          <textarea/>
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11,
					Message: "Ambiguous spacing before next element textarea"},
			},
		},

		// ---- Edge: text contains CRLF line ending — \r is whitespace under
		// Go's regex, so the pattern still fires on `\S\r\n\s*$`. ----
		{
			Code: "<App>foo\r\n<a>bar</a></App>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext"},
			},
		},

		// ---- Edge: a JsxExpression in the window position would normally
		// suppress the report, but here the JsxExpression sits AFTER the
		// inline element — text+inline+expr+text reduces to just text+inline
		// for the first window, which still fires for the leading text. ----
		{
			Code: `
        <App>
          foo
          <a>bar</a>
          {expr}
        </App>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 4, Column: 11,
					Message: "Ambiguous spacing before next element a"},
			},
		},

		// ---- Edge: text immediately preceding inline ends with non-newline
		// space — the regex requires a \n in the trailing whitespace, so
		// trailing tabs/spaces alone don't match. We assert the boundary
		// case where the LAST char is `\n` followed by spaces. ----
		{
			Code: "<App>x  \n<a>y</a></App>",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext"},
			},
		},

		// ---- Edge: JSX in arrow function body returning paragraph with bad
		// link spacing — real-world code path. ----
		{
			Code: `
        const Doc = () => (
          <p>
            Read more
            <a href="/docs">here</a>
          </p>
        );
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingBeforeNext", Line: 5, Column: 13,
					Message: "Ambiguous spacing before next element a"},
			},
		},

		// ---- Edge: JSX returned from a component method — class component
		// path, lock against listener boundary issues with class bodies. ----
		{
			Code: `
        class C extends Component {
          render() {
            return (
              <p>
                <code>x</code>
                trailing
              </p>
            );
          }
        }
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "spacingAfterPrev", Line: 6, Column: 31,
					Message: "Ambiguous spacing after previous element code"},
			},
		},
	})
}
