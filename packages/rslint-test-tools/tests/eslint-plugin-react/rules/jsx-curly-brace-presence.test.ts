import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamES2015Options } from '../config';
import { diagnostic } from '../expectations';

function expectUnnecessaryCurlyError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'unnecessaryCurly',
    'Curly braces are unnecessary here.',
    range,
  );
}

function expectMissingCurlyError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'missingCurly',
    'Need to wrap this literal in a JSX expression.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-curly-brace-presence', {
  valid: [
    // ---- Defaults ----
    { code: `<App {...props}>foo</App>` },
    { code: `<>foo</>` },
    { code: `<App {...props}>foo</App>`, options: [{ props: 'never' }] },

    // ---- Whitespace expressions are always allowed ----
    { code: `<App>{' '}</App>` },
    { code: `<App>{'     '}</App>` },
    { code: `<App>{' '}</App>`, options: [{ children: 'never' }] },
    { code: `<App>{' '}</App>`, options: [{ children: 'always' }] },

    // ---- Templates with substitutions stay wrapped ----
    {
      code: '<App>{`Hello ${word} World`}</App>',
      options: [{ children: 'never' }],
    },
    {
      code: '<App prop={`foo ${word} bar`} />',
      options: [{ props: 'never' }],
    },
    { code: '<App label={`${label}`} />', options: ['never'] },

    // ---- always-children allows braces around JSX child elements ----
    {
      code: `<App>{<myApp></myApp>}</App>`,
      options: [{ children: 'always' }],
    },

    // ---- never-children: identifiers, arrays, booleans pass through ----
    { code: `<App>{[]}</App>` },
    { code: `<App>foo</App>` },
    { code: `<App prop={true}>foo</App>` },
    { code: `<App prop>foo</App>` },

    // ---- Per-option matrix ----
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: [{ props: 'never' }],
    },
    {
      code: `<MyComponent>foo</MyComponent>`,
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
      code: `<MyComponent>{'foo'}</MyComponent>`,
      options: [{ children: 'ignore' }],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ props: 'ignore' }],
    },
    {
      code: `<MyComponent prop={'bar'}>{'foo'}</MyComponent>`,
      options: ['always'],
    },
    {
      code: `<MyComponent prop="bar" attr='foo' />`,
      options: ['never'],
    },

    // ---- never-children: literals containing JSX-disallowed chars stay wrapped ----
    {
      code: `<MyComponent>{"div { margin-top: 0; }"}</MyComponent>`,
      options: ['never'],
    },
    {
      code: `<MyComponent>{"<Foo />"}</MyComponent>`,
      options: ['never'],
    },

    // ---- HTML entities preserved ----
    {
      code: `<MyComponent prop={"Hello &middot; world"}>bar</MyComponent>`,
      options: ['never'],
    },
    {
      code: `<MyComponent>{"Hello &middot; world"}</MyComponent>`,
      options: ['never'],
    },

    // ---- Trailing whitespace / leading whitespace strings ----
    {
      code: `<MyComponent>{"space after "}</MyComponent>`,
      options: ['never'],
    },
    {
      code: `<MyComponent>{" space before"}</MyComponent>`,
      options: ['never'],
    },

    // ---- propElementValues default ('ignore'): JSX prop values pass ----
    { code: `<MyComponent p={<Foo>Bar</Foo>} />` },
    { code: `<App horror={<div />} />` },
    {
      code: `<App horror={<div />} />`,
      options: [{ propElementValues: 'ignore' }],
    },

    // ---- Comments inside JSXExpression: braces never removed ----
    { code: `<App>{/* comment */}</App>` },
    { code: `<App>{/* comment */ <Foo />}</App>` },
    { code: `<App>{/* comment */ 'foo'}</App>` },
    { code: `<App prop={/* comment */ 'foo'} />` },

    // ---- script-like child: never-children leaves the template alone ----
    {
      code: '<script>{`window.foo = "bar"`}</script>',
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<App>{' '}
</App>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<App>{'     '}
</App>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<App>{'    '}</App>`,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<App>{'        '}</App>`,
      options: [{ children: 'always' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<App {...props}>foo</App>`,
      options: [{ props: 'always' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        <React.Fragment>
          foo{' '}
          <span>bar</span>
        </React.Fragment>
      `,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        <>
          foo{' '}
          <span>bar</span>
        </>
      `,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `<App>{\`Hello \\n World\`}</App>`,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `<App>{\`Hello \${word} World\`}{\`foo\`}</App>`,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `<App prop={\`foo \${word} bar\`}>foo</App>`,
      options: [{ props: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `<App>{"foo"}{<Component>bar</Component>}</App>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `<App prop='bar'>foo</App>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `<App prop='bar'>{'foo \\n bar'}</App>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `<App prop={ ' ' }/>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `<MyComponent prop="bar">foo</MyComponent>`,
      options: [{ props: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `<MyComponent>{<App/>}{"123"}</MyComponent>`,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `<App>{"foo 'bar' \\"foo\\" bar"}</App>`,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `<MyComponent prop={"bar"}>foo</MyComponent>`,
      options: [{ props: 'always' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `<MyComponent>{"foo"}</MyComponent>`,
      options: [{ children: 'always' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `<MyComponent>foo</MyComponent>`,
      options: [{ children: 'ignore' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: [{ props: 'ignore' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `<MyComponent prop="bar">foo</MyComponent>`,
      options: [{ props: 'ignore' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `<MyComponent prop='bar'>{'foo'}</MyComponent>`,
      options: [{ children: 'always', props: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ children: 'never', props: 'always' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
      code: `<MyComponent prop={"bar"}>{"foo"}</MyComponent>`,
      options: ['always'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 46 (7.37.5)',
      code: `<MyComponent prop={"bar"} attr={'foo'} />`,
      options: ['always'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `<MyComponent prop={\`bar \${word} foo\`}>{\`foo \${word}\`}</MyComponent>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `<MyComponent prop={"Hello \\u1026 world"}>bar</MyComponent>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `<MyComponent>{"Hello \\u1026 world"}</MyComponent>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `<MyComponent>{"Hello \\n world"}</MyComponent>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 59 (7.37.5)',
      code: `<MyComponent>{\`space after \`}</MyComponent>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 60 (7.37.5)',
      code: `<MyComponent>{\` space before\`}</MyComponent>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 61 (7.37.5)',
      code: `<a a={"start\\/n\\/nend"}/>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 62 (7.37.5)',
      code: `
        <App prop={\`
          a
          b
        \`} />
      `,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 63 (7.37.5)',
      code: `
        <App prop={\`
          a
          b
        \`} />
      `,
      options: ['always'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 64 (7.37.5)',
      code: `
        <App>
          {\`
            a
            b
          \`}
        </App>
      `,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 65 (7.37.5)',
      code: `
        <App>{\`
          a
          b
        \`}</App>
      `,
      options: ['always'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 66 (7.37.5)',
      code: `
        <MyComponent>
          %
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 67 (7.37.5)',
      code: `
        <MyComponent>
          { 'space after ' }
          <b>foo</b>
          { ' space before' }
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 68 (7.37.5)',
      code: `
        <MyComponent>
          { \`space after \` }
          <b>foo</b>
          { \` space before\` }
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 69 (7.37.5)',
      code: `
        <MyComponent>
          foo
          <div>bar</div>
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 70 (7.37.5)',
      code: `
        <MyComponent p={<Foo>Bar</Foo>}>
        </MyComponent>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 71 (7.37.5)',
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
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 72 (7.37.5)',
      code: `
        <App>
          <Component />&nbsp;
          &nbsp;
        </App>
      `,
      options: [{ children: 'always' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 73 (7.37.5)',
      code: `
        const Component2 = () => {
          return <span>/*</span>;
        };
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 74 (7.37.5)',
      code: `
        const Component2 = () => {
          return <span>/*</span>;
        };
      `,
      options: [{ props: 'never', children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 75 (7.37.5)',
      code: `
        import React from "react";

        const Component = () => {
          return <span>{"/*"}</span>;
        };
      `,
      options: [{ props: 'never', children: 'never' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 80 (7.37.5)',
      code: `
          <App>
            {
              // comment
              <Foo />
            }
          </App>
        `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 81 (7.37.5)',
      code: `<App horror=<div /> />`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 83 (7.37.5)',
      code: `<App horror=<div /> />`,
      options: [{ propElementValues: 'ignore' }],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 85 (7.37.5)',
      code: `
        <script>{\`window.foo = "bar"\`}</script>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 86 (7.37.5)',
      code: `
        <CollapsibleTitle
          extra={<span className="activity-type">{activity.type}</span>}
        />
      `,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream valid 88 (7.37.5)',
      code: `<App>{\`\${label}\`}</App>`,
      options: ['never'],
      languageOptions: upstreamES2015Options,
    },
  ],
  invalid: [
    // ---- Unnecessary curly: template literal in props ----
    {
      code: '<App prop={`foo`} />',
      options: [{ props: 'never' }],
      output: '<App prop="foo" />',
      errors: [expectUnnecessaryCurlyError([1, 11, 1, 18])],
    },
    // ---- Unnecessary curly: JSX element child wrapped in `{}` ----
    {
      code: `<App>{<myApp></myApp>}</App>`,
      options: [{ children: 'never' }],
      output: `<App><myApp></myApp></App>`,
      errors: [expectUnnecessaryCurlyError([1, 6, 1, 23])],
    },
    {
      code: `<App>{<myApp></myApp>}</App>`,
      output: `<App><myApp></myApp></App>`,
      errors: [expectUnnecessaryCurlyError([1, 6, 1, 23])],
    },
    // ---- Children template literal collapsed ----
    {
      code: '<App>{`foo`}</App>',
      options: [{ children: 'never' }],
      output: '<App>foo</App>',
      errors: [expectUnnecessaryCurlyError([1, 6, 1, 13])],
    },
    {
      code: '<>{`foo`}</>',
      options: [{ children: 'never' }],
      output: '<>foo</>',
      errors: [expectUnnecessaryCurlyError([1, 3, 1, 10])],
    },
    {
      code: `<MyComponent>{'foo'}</MyComponent>`,
      output: `<MyComponent>foo</MyComponent>`,
      errors: [expectUnnecessaryCurlyError([1, 14, 1, 21])],
    },
    {
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      output: `<MyComponent prop="bar">foo</MyComponent>`,
      errors: [expectUnnecessaryCurlyError([1, 19, 1, 26])],
    },

    // ---- Missing curly: props=always ----
    {
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: [{ props: 'always' }],
      output: `<MyComponent prop={"bar"}>foo</MyComponent>`,
      errors: [expectMissingCurlyError([1, 19, 1, 24])],
    },
    {
      code: `<MyComponent prop="foo 'bar'">foo</MyComponent>`,
      options: [{ props: 'always' }],
      output: `<MyComponent prop={"foo 'bar'"}>foo</MyComponent>`,
      errors: [expectMissingCurlyError([1, 19, 1, 30])],
    },
    {
      code: `<MyComponent prop='foo "bar"'>foo</MyComponent>`,
      options: [{ props: 'always' }],
      output: `<MyComponent prop={"foo \\"bar\\""}>foo</MyComponent>`,
      errors: [expectMissingCurlyError([1, 19, 1, 30])],
    },

    // ---- Missing curly: children=always ----
    {
      code: `<MyComponent>foo bar </MyComponent>`,
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar "}</MyComponent>`,
      errors: [expectMissingCurlyError([1, 14, 1, 22])],
    },
    {
      code: `<MyComponent>foo bar 'foo'</MyComponent>`,
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar 'foo'"}</MyComponent>`,
      errors: [expectMissingCurlyError([1, 14, 1, 27])],
    },
    {
      code: '<MyComponent>foo bar "foo"</MyComponent>',
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar \\"foo\\""}</MyComponent>`,
      errors: [expectMissingCurlyError([1, 14, 1, 27])],
    },
    {
      code: '<MyComponent>foo bar <App/></MyComponent>',
      options: [{ children: 'always' }],
      output: `<MyComponent>{"foo bar "}<App/></MyComponent>`,
      errors: [expectMissingCurlyError([1, 14, 1, 22])],
    },

    // ---- Two-prop attribute fixes ----
    {
      code: `<App prop={'foo'} attr={" foo "} />`,
      options: [{ props: 'never' }],
      output: `<App prop="foo" attr=" foo " />`,
      errors: [
        expectUnnecessaryCurlyError([1, 11, 1, 18]),
        expectUnnecessaryCurlyError([1, 24, 1, 33]),
      ],
    },
    {
      code: `<App prop='foo' attr="bar" />`,
      options: [{ props: 'always' }],
      output: `<App prop={"foo"} attr={"bar"} />`,
      errors: [
        expectMissingCurlyError([1, 11, 1, 16]),
        expectMissingCurlyError([1, 22, 1, 27]),
      ],
    },

    // ---- HTML entities + always-children: whole text wraps ----
    {
      code: `<App>foo &middot; bar</App>`,
      options: [{ children: 'always' }],
      output: `<App>{"foo &middot; bar"}</App>`,
      errors: [expectMissingCurlyError([1, 6, 1, 22])],
    },
    {
      code: `<App prop='foo &middot; bar' />`,
      options: [{ props: 'always' }],
      output: `<App prop={"foo &middot; bar"} />`,
      errors: [expectMissingCurlyError([1, 11, 1, 29])],
    },

    // ---- Quote payload, never-children unwraps ----
    {
      code: `<App>{'foo "bar"'}</App>`,
      options: [{ children: 'never' }],
      output: `<App>foo "bar"</App>`,
      errors: [expectUnnecessaryCurlyError([1, 6, 1, 19])],
    },
    {
      code: `<App>{"foo 'bar'"}</App>`,
      options: [{ children: 'never' }],
      output: `<App>foo 'bar'</App>`,
      errors: [expectUnnecessaryCurlyError([1, 6, 1, 19])],
    },

    // ---- propElementValues=never collapses braces around JSX prop value ----
    // Known upstream/native autofix cycle: removing these braces is followed by
    // reinserting them on the next round. Preserve the direct removal edit here.
    {
      code: `<App horror={<div />} />`,
      options: [
        { props: 'never', children: 'never', propElementValues: 'never' },
      ],
      // Assert the complete output after all autofix rounds.

      errors: [
        {
          ...expectUnnecessaryCurlyError([1, 13, 1, 22]),
          fix: { range: [12, 21], text: '<div />' },
        },
      ],
    },

    // ---- Quote-only payload ----
    {
      code: `<Foo bar={"'"} />`,
      options: [
        { props: 'never', children: 'never', propElementValues: 'never' },
      ],
      output: `<Foo bar="'" />`,
      errors: [expectUnnecessaryCurlyError([1, 10, 1, 15])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<App prop={\`foo\`}>foo</App>`,
      options: [{ props: 'never' }],
      errors: [expectUnnecessaryCurlyError([1, 11, 1, 18])],
      output: `<App prop="foo">foo</App>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `<MyComponent>{'foo'}</MyComponent>`,
      options: [{ children: 'never' }],
      errors: [expectUnnecessaryCurlyError([1, 14, 1, 21])],
      output: `<MyComponent>foo</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `<MyComponent prop={'bar'}>foo</MyComponent>`,
      options: [{ props: 'never' }],
      errors: [expectUnnecessaryCurlyError([1, 19, 1, 26])],
      output: `<MyComponent prop="bar">foo</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        <MyComponent>
          {'%'}
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      errors: [expectUnnecessaryCurlyError([3, 11, 3, 16])],
      output: `
        <MyComponent>
          %
        </MyComponent>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        <MyComponent>
          {'foo'}
          <div>
            {'bar'}
          </div>
          {'baz'}
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      errors: [
        expectUnnecessaryCurlyError([3, 11, 3, 18]),
        expectUnnecessaryCurlyError([5, 13, 5, 20]),
        expectUnnecessaryCurlyError([7, 11, 7, 18]),
      ],
      output: `
        <MyComponent>
          foo
          <div>
            bar
          </div>
          baz
        </MyComponent>
      `,
      languageOptions: upstreamES2015Options,
    },
    // The historical upstream assertion expects 2 errors; both pinned upstream and native report 4.
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        <MyComponent>
          {'foo'}
          <div>
            {'bar'}
          </div>
          {'baz'}
          {'some-complicated-exp'}
        </MyComponent>
      `,
      options: [{ children: 'never' }],
      errors: [
        expectUnnecessaryCurlyError([3, 11, 3, 18]),
        expectUnnecessaryCurlyError([5, 13, 5, 20]),
        expectUnnecessaryCurlyError([7, 11, 7, 18]),
        expectUnnecessaryCurlyError([8, 11, 8, 35]),
      ],
      output: `
        <MyComponent>
          foo
          <div>
            bar
          </div>
          baz
          some-complicated-exp
        </MyComponent>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `<MyComponent prop="foo 'bar' \\n ">foo</MyComponent>`,
      options: [{ props: 'always' }],
      errors: [expectMissingCurlyError([1, 19, 1, 34])],
      output: `<MyComponent prop={"foo 'bar' \\\\n "}>foo</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `<MyComponent>foo bar \\r </MyComponent>`,
      options: [{ children: 'always' }],
      errors: [expectMissingCurlyError([1, 14, 1, 25])],
      output: `<MyComponent>{"foo bar \\\\r "}</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `<MyComponent>foo \\n bar</MyComponent>`,
      options: [{ children: 'always' }],
      errors: [expectMissingCurlyError([1, 14, 1, 24])],
      output: `<MyComponent>{"foo \\\\n bar"}</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `<MyComponent>foo \\u1234 bar</MyComponent>`,
      options: [{ children: 'always' }],
      errors: [expectMissingCurlyError([1, 14, 1, 28])],
      output: `<MyComponent>{"foo \\\\u1234 bar"}</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `<MyComponent prop='foo \\u1234 bar' />`,
      options: [{ props: 'always' }],
      errors: [expectMissingCurlyError([1, 19, 1, 35])],
      output: `<MyComponent prop={"foo \\\\u1234 bar"} />`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `<MyComponent prop={'bar'}>{'foo'}</MyComponent>`,
      options: ['never'],
      errors: [
        expectUnnecessaryCurlyError([1, 19, 1, 26]),
        expectUnnecessaryCurlyError([1, 27, 1, 34]),
      ],
      output: `<MyComponent prop="bar">foo</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `<MyComponent prop='bar'>foo</MyComponent>`,
      options: ['always'],
      errors: [
        expectMissingCurlyError([1, 19, 1, 24]),
        expectMissingCurlyError([1, 25, 1, 28]),
      ],
      output: `<MyComponent prop={"bar"}>{"foo"}</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `<App prop='foo' attr={"bar"} />`,
      options: [{ props: 'always' }],
      errors: [expectMissingCurlyError([1, 11, 1, 16])],
      output: `<App prop={"foo"} attr={"bar"} />`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `<App prop={'foo'} attr='bar' />`,
      options: [{ props: 'always' }],
      errors: [expectMissingCurlyError([1, 24, 1, 29])],
      output: `<App prop={'foo'} attr={"bar"} />`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        <App prop="\x20\x20\x20\x20
           a\x20\x20\x20\x20\x20
             b      c
                d
        ">
          a
              b     c\x20\x20\x20
                 d\x20\x20\x20\x20\x20\x20
        </App>
      `,
      options: ['always'],
      errors: [
        expectMissingCurlyError([2, 19, 6, 10]),
        expectMissingCurlyError([6, 11, 10, 9]),
      ],
      output: `
        <App prop="\x20\x20\x20\x20
           a\x20\x20\x20\x20\x20
             b      c
                d
        ">
          {"a"}
              {"b     c   "}
                 {"d      "}
        </App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        <App prop='\x20\x20\x20\x20
           a\x20\x20\x20\x20\x20
             b      c
                d
        '>
          a
              b     c\x20\x20\x20
                 d\x20\x20\x20\x20\x20\x20
        </App>
      `,
      options: ['always'],
      errors: [
        expectMissingCurlyError([2, 19, 6, 10]),
        expectMissingCurlyError([6, 11, 10, 9]),
      ],
      output: `
        <App prop='\x20\x20\x20\x20
           a\x20\x20\x20\x20\x20
             b      c
                d
        '>
          {"a"}
              {"b     c   "}
                 {"d      "}
        </App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
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
      options: [{ children: 'always' }],
      errors: [
        expectMissingCurlyError([2, 14, 4, 11]),
        expectMissingCurlyError([4, 16, 4, 27]),
        expectMissingCurlyError([5, 17, 6, 21]),
        expectMissingCurlyError([6, 24, 6, 31]),
        expectMissingCurlyError([7, 21, 9, 13]),
      ],
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
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        <App>
          &lt;Component&gt;
          &nbsp;<Component />&nbsp;
          &nbsp;
        </App>
      `,
      options: [{ children: 'always' }],
      errors: [expectMissingCurlyError([2, 14, 4, 17])],
      output: `
        <App>
          &lt;{"Component"}&gt;
          &nbsp;<Component />&nbsp;
          &nbsp;
        </App>
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        <Box mb={'1rem'} />
      `,
      options: [{ props: 'never' }],
      errors: [expectUnnecessaryCurlyError([2, 17, 2, 25])],
      output: `
        <Box mb="1rem" />
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        <Box mb={'1rem {}'} />
      `,
      options: ['never'],
      errors: [expectUnnecessaryCurlyError([2, 17, 2, 28])],
      output: `
        <Box mb="1rem {}" />
      `,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `<MyComponent prop={"{ style: true }"}>bar</MyComponent>`,
      options: ['never'],
      errors: [expectUnnecessaryCurlyError([1, 19, 1, 38])],
      output: `<MyComponent prop="{ style: true }">bar</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `<MyComponent prop={"< style: true >"}>foo</MyComponent>`,
      options: ['never'],
      errors: [expectUnnecessaryCurlyError([1, 19, 1, 38])],
      output: `<MyComponent prop="< style: true >">foo</MyComponent>`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `<App horror=<div /> />`,
      options: [
        { props: 'always', children: 'always', propElementValues: 'always' },
      ],
      errors: [expectMissingCurlyError([1, 13, 1, 20])],
      output: `<App horror={<div />} />`,
      languageOptions: upstreamES2015Options,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
        <Foo help={'The maximum time range for searches. (i.e. "P30D" for 30 days, "PT24H" for 24 hours)'} />
      `,
      options: ['never'],
      errors: [expectUnnecessaryCurlyError([2, 19, 2, 107])],
      output: `
        <Foo help='The maximum time range for searches. (i.e. "P30D" for 30 days, "PT24H" for 24 hours)' />
      `,
      languageOptions: upstreamES2015Options,
    },
  ],
});
