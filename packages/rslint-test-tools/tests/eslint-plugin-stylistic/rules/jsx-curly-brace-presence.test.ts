/**
 * @fileoverview Disallow unnecessary JSX expressions when literals alone are
 * sufficient or enforce JSX expressions on literals in JSX children or attributes.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-curly-brace-presence/jsx-curly-brace-presence.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid, invalid })` ->
 *    `ruleTester.run('jsx-curly-brace-presence', null as never, { valid, invalid })`.
 *  - The upstream `valids(...)` / `invalids(...)` wrappers are NOT identity
 *    helpers — `applyAllParsers` fans each case out across the ESLint parser
 *    matrix (default / `@babel/eslint-parser` / `@typescript-eslint/parser`) and
 *    appends a `// features: […], parser: …, parserOptions: {…}` bookkeeping
 *    comment to `code`/`output`. With ESLint 10.5.0 (`skipBabel = gte(version,
 *    '10.0.0')` is true) the babel slot is always dropped, so a feature-less case
 *    reduces to the SAME logical fixture under the default + ts parsers, differing
 *    only by that cosmetic trailing comment. rslint has a single parser (ts-go),
 *    so each logical case is ported once with the harness scaffolding stripped —
 *    identical to how the reference `quotes` port drops `parserOptions`/`lang`.
 *  - `features` dropped after accounting for what each flag means under ts-go:
 *      • `['fragment']`  — JSX fragments; ts-go supports `<>…</>`. GREEN.
 *      • `['no-ts-old']` — excludes only the OLD typescript-eslint parser (the
 *        comment says it "hangs forever" / mis-parses); ts-go is unaffected and
 *        the cases were verified GREEN.
 *      • `['no-ts']`     — excludes the ts-parser slot because `horror=<div />`
 *        (a bare JSX-element attribute value) is non-standard syntax. ts-go
 *        nonetheless PARSES it and reports correctly for `propElementValues`
 *        'always'/'ignore'/default — those cases are GREEN. The single `'never'`
 *        invalid case is a real fix gap (see KNOWN GAPS).
 *  - Error helpers are already plain `{ messageId }` objects; no inlining needed.
 *  - No `suggestions`, `._css_`/`._json_`/`._markdown_`, or `readFileSync`
 *    fixtures exist for this rule.
 *
 * JSX code (`</Tag>`, `/>`, `<>`) is auto-routed by the RuleTester to a `.tsx`
 * fixture, which ts-go parses correctly.
 *
 * Three items are isolated into the `KNOWN GAPS` block at the bottom (kept as
 * commented fixtures, never deleted, each annotated with upstream-vs-rslint):
 *  1. The `propElementValues: 'never'` unwrap of `<App horror={<div />} />`:
 *     ts-go reports the diagnostic (count matches) but cannot autofix to the
 *     bare-element form `horror=<div />` (not valid TSX), so the fix output
 *     diverges.
 *  2. The `features: ['no-default', 'no-ts-new', 'no-babel-new']` multi-pass
 *     case: ESLint 10 leaves ZERO parser variants for it (`no-default` drops the
 *     default parser too), so upstream never runs it. Its expected output only
 *     fixes 2 of 4 curlies (broken old-parser behavior, per its own
 *     `TODO: FIXME … and fix`); ts-go correctly fixes all 4.
 *  3. The `if (!skipBabel)` `_babel` run() block: skipped entirely under ESLint
 *     >= 10 and keyed to `parser: BABEL_ESLINT`. Its expected 4-curly fix is what
 *     ts-go actually produces, but it is a Babel-parser-only case.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-curly-brace-presence', null as never, {
  valid: [
    {
      code: '<App {...props}>foo</App>',
    },
    {
      code: '<>foo</>',
    },
    {
      code: '<App {...props}>foo</App>',
      options: [{ props: 'never' }],
    },
    /**
     * There is no way to inject the space into JSX without an expression container
     * so this format should always be allowed regardless of the `children` option.
     */
    {
      code: '<App>{\' \'}</App>',
    },
    {
      code: '<App>{\' \'}\n</App>',
    },
    {
      code: '<App>{\'     \'}</App>',
    },
    {
      code: '<App>{\'     \'}\n</App>',
    },
    {
      code: '<App>{\' \'}</App>',
      options: [{ children: 'never' }],
    },
    {
      code: '<App>{\'    \'}</App>',
      options: [{ children: 'never' }],
    },
    {
      code: '<App>{\' \'}</App>',
      options: [{ children: 'always' }],
    },
    {
      code: '<App>{\'        \'}</App>',
      options: [{ children: 'always' }],
    },
    {
      code: '<App {...props}>foo</App>',
      options: [{ props: 'always' }],
    },
    {
      code: '<App>{`Hello ${word} World`}</App>',
      options: [{ children: 'never' }],
    },
    {
      code: `
        <React.Fragment>
          foo{' '}
          <span>bar</span>
        </React.Fragment>
      `,
      options: [{ children: 'never' }],
    },
    {
      code: `
        <>
          foo{' '}
          <span>bar</span>
        </>
      `,
      options: [{ children: 'never' }],
    },
    {
      code: '<App>{`Hello \\n World`}</App>',
      options: [{ children: 'never' }],
    },
    {
      code: '<App>{`Hello ${word} World`}{`foo`}</App>',
      options: [{ children: 'never' }],
    },
    {
      code: '<App prop={`foo ${word} bar`}>foo</App>',
      options: [{ props: 'never' }],
    },
    {
      code: '<App prop={`foo ${word} bar`} />',
      options: [{ props: 'never' }],
    },
    {
      code: '<App>{<myApp></myApp>}</App>',
      options: [{ children: 'always' }],
    },
    {
      code: '<App>{[]}</App>',
    },
    {
      code: '<App>foo</App>',
    },
    {
      code: '<App>{"foo"}{<Component>bar</Component>}</App>',
    },
    {
      code: `<App prop='bar'>foo</App>`,
    },
    {
      code: '<App prop={true}>foo</App>',
    },
    {
      code: '<App prop>foo</App>',
    },
    {
      code: `<App prop='bar'>{'foo \\n bar'}</App>`,
    },
    {
      code: `<App prop={ ' ' }/>`,
    },
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: [{ props: 'never' }],
    },
    {
      code: `<MyComponent prop="bar">foo</MyComponent>`,
      options: [{ props: 'never' }],
    },
    {
      code: '<MyComponent>foo</MyComponent>',
      options: [{ children: 'never' }],
    },
    {
      code: '<MyComponent>{<App/>}{"123"}</MyComponent>',
      options: [{ children: 'never' }],
    },
    {
      code: `<App>{"foo 'bar' \\"foo\\" bar"}</App>`,
      options: [{ children: 'never' }],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ props: 'always' }],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      options: [{ children: 'always' }],
    },
    {
      code: `<MyComponent prop={"bar"}>foo</MyComponent>`,
      options: [{ props: 'always' }],
    },
    {
      code: `<MyComponent>{"foo"}</MyComponent>`,
      options: [{ children: 'always' }],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      options: [{ children: 'ignore' }],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ props: 'ignore' }],
    },
    {
      code: '<MyComponent>foo</MyComponent>',
      options: [{ children: 'ignore' }],
    },
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: [{ props: 'ignore' }],
    },
    {
      code: `<MyComponent prop="bar">foo</MyComponent>`,
      options: [{ props: 'ignore' }],
    },
    {
      code: `<MyComponent prop='bar'>{'foo'}</MyComponent>`,
      options: [{ children: 'always', props: 'never' }],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ children: 'never', props: 'always' }],
    },
    {
      code: `<MyComponent prop={'bar'}>{'foo'}</MyComponent>`,
      options: ['always'],
    },
    {
      code: `<MyComponent prop={"bar"}>{"foo"}</MyComponent>`,
      options: ['always'],
    },
    {
      code: `<MyComponent prop={"bar"} attr={'foo'} />`,
      options: ['always'],
    },
    {
      code: `<MyComponent prop="bar" attr='foo' />`,
      options: ['never'],
    },
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: ['never'],
    },
    {
      code: '<MyComponent prop={`bar ${word} foo`}>{`foo ${word}`}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{"div { margin-top: 0; }"}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{"<Foo />"}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent prop={"Hello \\u1026 world"}>bar</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{"Hello \\u1026 world"}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent prop={"Hello &middot; world"}>bar</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{"Hello &middot; world"}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{"Hello \\n world"}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{"space after "}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{" space before"}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{`space after `}</MyComponent>',
      options: ['never'],
    },
    {
      code: '<MyComponent>{` space before`}</MyComponent>',
      options: ['never'],
    },
    {
      code: ['<a a={"start\\', '\\', 'end"}/>'].join('/n'),
      options: ['never'],
    },
    {
      code: `
        <App prop={\`
          a
          b
        \`} />
      `,
      options: ['never'],
    },
    {
      code: `
        <App prop={\`
          a
          b
        \`} />
      `,
      options: ['always'],
    },
    {
      code: `
        <App>
          {\`
            a
            b
          \`}
        </App>
      `,
      options: ['never'],
    },
    {
      code: `
        <App>{\`
          a
          b
        \`}</App>
      `,
      options: ['always'],
    },
    {
      code: `
        <MyComponent>
          %
        </MyComponent>
      `,
      options: [{ children: 'never' }],
    },
    {
      code: `
        <MyComponent>
          { 'space after ' }
          <b>foo</b>
          { ' space before' }
        </MyComponent>
      `,
      options: [{ children: 'never' }],
    },
    {
      code: `
        <MyComponent>
          { \`space after \` }
          <b>foo</b>
          { \` space before\` }
        </MyComponent>
      `,
      options: [{ children: 'never' }],
    },
    {
      code: `
        <MyComponent>
          foo
          <div>bar</div>
        </MyComponent>
      `,
      options: [{ children: 'never' }],
    },
    {
      code: `
        <MyComponent p={<Foo>Bar</Foo>}>
        </MyComponent>
      `,
    },
    {
      code: `
        <MyComponent>
          <div>
            <p>
              <span>
                {"foo"}
              </span>
            </p>
          </div>
        </MyComponent>
      `,
      options: [{ children: 'always' }],
    },
    {
      code: `
        <App>
          <Component />&nbsp;
          &nbsp;
        </App>
      `,
      options: [{ children: 'always' }],
    },
    {
      code: `
        const Component2 = () => {
          return <span>/*</span>;
        };
      `,
    },
    {
      code: `
        const Component2 = () => {
          return <span>/*</span>;
        };
      `,
      options: [{ props: 'never', children: 'never' }],
    },
    {
      code: `
        import React from "react";

        const Component = () => {
          return <span>{"/*"}</span>;
        };
      `,
      options: [{ props: 'never', children: 'never' }],
    },
    {
      code: `<App>{/* comment */}</App>`,
    },
    {
      code: `<App>{/* comment */ <Foo />}</App>`,
    },
    {
      code: `<App>{/* comment */ 'foo'}</App>`,
    },
    {
      code: `<App prop={/* comment */ 'foo'} />`,
    },
    {
      code: `
        <App>
          {
            // comment
            <Foo />
          }
        </App>
      `,
    },
    {
      code: `<App horror=<div /> />`,
    },
    {
      code: `<App horror={<div />} />`,
    },
    {
      code: `<App horror=<div /> />`,
      options: [{ propElementValues: 'ignore' }],
    },
    {
      code: `<App horror={<div />} />`,
      options: [{ propElementValues: 'ignore' }],
    },
    {
      code: `
        <script>{\`window.foo = "bar"\`}</script>
      `,
    },
    {
      code: `
        <CollapsibleTitle
          extra={<span className="activity-type">{activity.type}</span>}
        />
      `,
      options: ['never'],
    },
    // legit as this single template literal might be used for stringifying
    {
      code: '<App label={`${label}`} />',
      options: ['never'],
    },
    {
      code: '<App>{`${label}`}</App>',
      options: ['never'],
    },
  ],

  invalid: [
    {
      code: '<App prop={`foo`} />',
      output: '<App prop="foo" />',
      options: [{ props: 'never' }],
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: '<App>{<myApp></myApp>}</App>',
      output: '<App><myApp></myApp></App>',
      options: [{ children: 'never' }],
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: '<App>{<myApp></myApp>}</App>',
      output: '<App><myApp></myApp></App>',
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: '<App prop={`foo`}>foo</App>',
      output: '<App prop="foo">foo</App>',
      options: [{ props: 'never' }],
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: '<App>{`foo`}</App>',
      output: '<App>foo</App>',
      options: [{ children: 'never' }],
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: '<>{`foo`}</>',
      output: '<>foo</>',
      options: [{ children: 'never' }],
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      output: '<MyComponent>foo</MyComponent>',
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      output: `<MyComponent prop="bar">foo</MyComponent>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      output: '<MyComponent>foo</MyComponent>',
      options: [{ children: 'never' }],
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      output: '<MyComponent prop="bar">foo</MyComponent>',
      options: [{ props: 'never' }],
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `
        <MyComponent>
          {'%'}
        </MyComponent>
      `,
      output: `
        <MyComponent>
          %
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      errors: [{ messageId: 'unnecessaryCurly' }],
    },
    {
      code: `
        <MyComponent>
          {'foo'}
          <div>
            {'bar'}
          </div>
          {'baz'}
        </MyComponent>
      `,
      output: `
        <MyComponent>
          foo
          <div>
            bar
          </div>
          baz
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      errors: [
        { messageId: 'unnecessaryCurly' },
        { messageId: 'unnecessaryCurly' },
        { messageId: 'unnecessaryCurly' },
      ],
    },
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      output: '<MyComponent prop={"bar"}>foo</MyComponent>',
      options: [{ props: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent prop="foo 'bar'">foo</MyComponent>`,
      output: `<MyComponent prop={"foo 'bar'"}>foo</MyComponent>`,
      options: [{ props: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent prop='foo "bar"'>foo</MyComponent>`,
      output: `<MyComponent prop={"foo \\"bar\\""}>foo</MyComponent>`,
      options: [{ props: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: '<MyComponent>foo bar </MyComponent>',
      output: `<MyComponent>{"foo bar "}</MyComponent>`,
      options: [{ children: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent prop="foo 'bar' \\n ">foo</MyComponent>`,
      output: `<MyComponent prop={"foo 'bar' \\\\n "}>foo</MyComponent>`,
      options: [{ props: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: '<MyComponent>foo bar \\r </MyComponent>',
      output: '<MyComponent>{"foo bar \\\\r "}</MyComponent>',
      options: [{ children: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent>foo bar 'foo'</MyComponent>`,
      output: `<MyComponent>{"foo bar 'foo'"}</MyComponent>`,
      options: [{ children: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: '<MyComponent>foo bar "foo"</MyComponent>',
      output: '<MyComponent>{"foo bar \\"foo\\""}</MyComponent>',
      options: [{ children: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: '<MyComponent>foo bar <App/></MyComponent>',
      output: '<MyComponent>{"foo bar "}<App/></MyComponent>',
      options: [{ children: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: '<MyComponent>foo \\n bar</MyComponent>',
      output: '<MyComponent>{"foo \\\\n bar"}</MyComponent>',
      options: [{ children: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: '<MyComponent>foo \\u1234 bar</MyComponent>',
      output: '<MyComponent>{"foo \\\\u1234 bar"}</MyComponent>',
      options: [{ children: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent prop='foo \\u1234 bar' />`,
      output: '<MyComponent prop={"foo \\\\u1234 bar"} />',
      options: [{ props: 'always' }],
      errors: [{ messageId: 'missingCurly' }],
    },
    {
      code: `<MyComponent prop={'bar'}>{'foo'}</MyComponent>`,
      output: '<MyComponent prop="bar">foo</MyComponent>',
      options: ['never'],
      errors: [
        { messageId: 'unnecessaryCurly' },
        { messageId: 'unnecessaryCurly' },
      ],
    },
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      output: '<MyComponent prop={"bar"}>{"foo"}</MyComponent>',
      options: ['always'],
      errors: [
        { messageId: 'missingCurly' },
        { messageId: 'missingCurly' },
      ],
    },
    {
      code: `<App prop={'foo'} attr={" foo "} />`,
      output: '<App prop="foo" attr=" foo " />',
      errors: [
        { messageId: 'unnecessaryCurly' },
        { messageId: 'unnecessaryCurly' },
      ],
      options: [{ props: 'never' }],
    },
    {
      code: `<App prop='foo' attr="bar" />`,
      output: '<App prop={"foo"} attr={"bar"} />',
      errors: [
        { messageId: 'missingCurly' },
        { messageId: 'missingCurly' },
      ],
      options: [{ props: 'always' }],
    },
    {
      code: `<App prop='foo' attr={"bar"} />`,
      output: `<App prop={"foo"} attr={"bar"} />`,
      errors: [{ messageId: 'missingCurly' }],
      options: [{ props: 'always' }],
    },
    {
      code: `<App prop={'foo'} attr='bar' />`,
      output: `<App prop={'foo'} attr={"bar"} />`,
      errors: [{ messageId: 'missingCurly' }],
      options: [{ props: 'always' }],
    },
    {
      code: `<App prop='foo &middot; bar' />`,
      errors: [{ messageId: 'missingCurly' }],
      options: [{ props: 'always' }],
      output: `<App prop={"foo &middot; bar"} />`,
    },
    {
      code: '<App>foo &middot; bar</App>',
      errors: [{ messageId: 'missingCurly' }],
      options: [{ children: 'always' }],
      output: '<App>{"foo &middot; bar"}</App>',
    },
    {
      code: `<App>{'foo "bar"'}</App>`,
      output: `<App>foo "bar"</App>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
      options: [{ children: 'never' }],
    },
    {
      code: `<App>{"foo 'bar'"}</App>`,
      output: `<App>foo 'bar'</App>`,
      errors: [{ messageId: 'unnecessaryCurly' }],
      options: [{ children: 'never' }],
    },
    {
      code: `
        <App prop="${'    '}
           a${'     '}
             b      c
                d
        ">
          a
              b     c${'   '}
                 d${'      '}
        </App>
      `,
      errors: [
        { messageId: 'missingCurly' },
        { messageId: 'missingCurly' },
      ],
      options: ['always'],
      output: `
        <App prop="${'    '}
           a${'     '}
             b      c
                d
        ">
          {"a"}
              {"b     c   "}
                 {"d      "}
        </App>
      `,
    },
    {
      code: `
        <App prop='${'    '}
           a${'     '}
             b      c
                d
        '>
          a
              b     c${'   '}
                 d${'      '}
        </App>
      `,
      errors: [
        { messageId: 'missingCurly' },
        { messageId: 'missingCurly' },
      ],
      options: ['always'],
      output: `
        <App prop='${'    '}
           a${'     '}
             b      c
                d
        '>
          {"a"}
              {"b     c   "}
                 {"d      "}
        </App>
      `,
    },
    {
      code: `
        <App>
          foo bar
          <div>foo bar foo</div>
          <span>
            foo bar <i>foo bar</i>
            <strong>
              foo bar
            </strong>
          </span>
        </App>
      `,
      output: `
        <App>
          {"foo bar"}
          <div>{"foo bar foo"}</div>
          <span>
            {"foo bar "}<i>{"foo bar"}</i>
            <strong>
              {"foo bar"}
            </strong>
          </span>
        </App>
      `,
      errors: [
        { messageId: 'missingCurly' },
        { messageId: 'missingCurly' },
        { messageId: 'missingCurly' },
        { messageId: 'missingCurly' },
        { messageId: 'missingCurly' },
      ],
      options: [{ children: 'always' }],
    },
    {
      code: `
        <App>
          &lt;Component&gt;
          &nbsp;<Component />&nbsp;
          &nbsp;
        </App>
      `,
      output: `
        <App>
          &lt;{"Component"}&gt;
          &nbsp;<Component />&nbsp;
          &nbsp;
        </App>
      `,
      errors: [{ messageId: 'missingCurly' }],
      options: [{ children: 'always' }],
    },
    {
      code: `
        <Box mb={'1rem'} />
      `,
      output: `
        <Box mb="1rem" />
      `,
      errors: [
        { messageId: 'unnecessaryCurly' },
      ],
      options: [{ props: 'never' }],
    },
    {
      code: `
        <Box mb={'1rem {}'} />
      `,
      output: `
        <Box mb="1rem {}" />
      `,
      errors: [{ messageId: 'unnecessaryCurly' }],
      options: ['never'],
    },
    {
      code: '<MyComponent prop={"{ style: true }"}>bar</MyComponent>',
      output: '<MyComponent prop="{ style: true }">bar</MyComponent>',
      errors: [{ messageId: 'unnecessaryCurly' }],
      options: ['never'],
    },
    {
      code: '<MyComponent prop={"< style: true >"}>foo</MyComponent>',
      output: '<MyComponent prop="< style: true >">foo</MyComponent>',
      errors: [{ messageId: 'unnecessaryCurly' }],
      options: ['never'],
    },
    {
      code: `<App horror=<div /> />`,
      output: `<App horror={<div />} />`,
      errors: [{ messageId: 'missingCurly' }],
      options: [{ props: 'always', children: 'always', propElementValues: 'always' }],
    },
  ],
});

// ──────────────────────────────────────────────────────────────────────────
// KNOWN GAPS — upstream cases that DON'T align with rslint/ts-go.
// Kept verbatim (commented) so the divergence is documented, never deleted.
// ──────────────────────────────────────────────────────────────────────────

// GAP 1 — `propElementValues: 'never'` unwrap of a JSX-element attribute value.
// Upstream (`features: ['no-ts']`) expects the curly braces stripped, producing
// the bare-element form `<App horror=<div /> />`:
//
//   {
//     code: `<App horror={<div />} />`,
//     output: `<App horror=<div /> />`,
//     errors: [{ messageId: 'unnecessaryCurly' }],
//     options: [{ props: 'never', children: 'never', propElementValues: 'never' }],
//     features: ['no-ts'],
//   }
//
// rslint: ts-go PARSES the fixture and REPORTS the diagnostic (count = 1, the
// `unnecessaryCurly` message matches), but its autofix cannot emit the
// bare-element form `horror=<div />` — that is not standard TSX syntax — so the
// `{<div />}` is left unchanged and the fix `output` diverges from upstream.
// Only the fix output differs; the diagnostic itself aligns.

// GAP 2 — multi-pass fix where a "complicated" expression is also unwrapped.
// Upstream marks this `features: ['no-default', 'no-ts-new', 'no-babel-new']`
// with `TODO: FIXME: remove no-default and no-ts-new and fix`. Under ESLint
// 10.5.0 every parser slot is dropped (`no-default` removes the default parser,
// `no-ts-new` the ts parser, `no-babel-new` the babel parser), so upstream never
// actually runs this case:
//
//   {
//     code: `
//       <MyComponent>
//         {'foo'}
//         <div>
//           {'bar'}
//         </div>
//         {'baz'}
//         {'some-complicated-exp'}
//       </MyComponent>
//     `,
//     output: `
//       <MyComponent>
//         foo
//         <div>
//           bar
//         </div>
//         {'baz'}
//         {'some-complicated-exp'}
//       </MyComponent>
//     `,
//     features: ['no-default', 'no-ts-new', 'no-babel-new'],
//     options: [{ children: 'never' }],
//     errors: [
//       { messageId: 'unnecessaryCurly', line: 3 },
//       { messageId: 'unnecessaryCurly', line: 5 },
//     ],
//   }
//
// rslint: ts-go reports 4 `unnecessaryCurly` diagnostics (lines 3, 5, 7, 8) and
// its multi-pass `--fix` unwraps ALL four curlies, including `{'baz'}` and
// `{'some-complicated-exp'}` — i.e. ts-go matches the CORRECT behavior the
// upstream `_babel` block documents (see GAP 3), not the broken 2-of-4 output
// pinned here. Both the error count and the fix output differ from this case.

// GAP 3 — the `if (!skipBabel) { run(...) }` `_babel` run() block.
// `skipBabel = gte(ESLint.version, '10.0.0')` is true (ESLint 10.5.0), so this
// block never executes upstream. It is keyed to `parser: BABEL_ESLINT`:
//
//   {
//     code: `
//       <MyComponent>
//         {'foo'}
//         <div>
//           {'bar'}
//         </div>
//         {'baz'}
//         {'some-complicated-exp'}
//       </MyComponent>
//     `,
//     output: `
//       <MyComponent>
//         foo
//         <div>
//           bar
//         </div>
//         baz
//         some-complicated-exp
//       </MyComponent>
//     `,
//     parser: BABEL_ESLINT,
//     parserOptions: babelParserOptions({}, new Set()),
//     options: [{ children: 'never' }],
//     errors: [
//       { messageId: 'unnecessaryCurly', line: 3 },
//       { messageId: 'unnecessaryCurly', line: 5 },
//       { messageId: 'unnecessaryCurly', line: 7 },
//       { messageId: 'unnecessaryCurly', line: 8 },
//     ],
//   }
//
// rslint: ts-go reproduces exactly this (4 diagnostics on lines 3/5/7/8, full
// unwrap of all four curlies) — so functionally rslint is aligned with the
// Babel behavior. It is isolated here only because it is a Babel-parser-keyed
// block that does not run under the project's ESLint version.
