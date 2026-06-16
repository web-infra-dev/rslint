/**
 * @fileoverview Tests for jsx-one-expression-per-line rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-one-expression-per-line/jsx-one-expression-per-line.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid, invalid })`
 *    -> `ruleTester.run('jsx-one-expression-per-line', null as never, { valid, invalid })`.
 *  - Upstream wraps every case in `valids(...)` / `invalids(...)` (from
 *    `#test/parsers-jsx`), which fan each case out across the default / babel /
 *    typescript-eslint parsers and append a `// features: [...], parser: ...`
 *    line comment to `code`/`output`. That is upstream test-harness machinery,
 *    not part of the case data — rslint runs the one ts-go parser, so the
 *    underlying case (code / output / options / errors) is ported and the
 *    fan-out + comment-append are dropped.
 *  - `features: ['fragment']` / `['fragment', 'no-ts-old']` dropped — they only
 *    select upstream parser variants. (`no-ts-old` is upstream's marker that the
 *    OLD typescript-eslint parser mis-handled the case; it is irrelevant to
 *    ts-go.) The fragment/JSX code is valid TSX and ts-go parses it natively.
 *  - `parserOptions.ecmaFeatures.jsx` dropped — JSX routes to a `.tsx` fixture.
 *  - The `$` (unindent) tag, used on 4 cases, is evaluated to its real string
 *    (strip the common leading indent, drop the leading/trailing blank lines).
 *  - All other `code`/`output` are plain backtick templates; their leading
 *    newline + shared indentation is load-bearing for this whitespace rule and
 *    is preserved byte-for-byte, including the literal `\t\t\t\t` tab escapes in
 *    the *WithTabs* valid cases and the `{' '}` JSX-text lines.
 *  - Upstream `output` strings carry many interpolations of the form
 *    "dollar-brace, a single-space string, an inline block comment reading
 *    'intentional trailing space', close-brace". They are kept verbatim — each
 *    evaluates to one literal trailing space, exactly the byte the fixer emits,
 *    and stays visible (an editor/formatter can't silently strip it).
 *  - The sole message `moveToNewLine: '\`{{descriptor}}\` must be placed on a new
 *    line'` is interpolated from each error's `data: { descriptor }`, so the
 *    rendered text is asserted exactly.
 *
 * No suggestions, no `skipBabel`-gated block, no external-fixture cases upstream.
 * The `._css_` / `._json_` / `._markdown_` files don't exist for this rule.
 *
 * KNOWN GAPS (real rslint<->upstream differences) are moved out of the live
 * `valid`/`invalid` arrays into the commented block at the bottom, each annotated
 * with upstream-expected vs. rslint-actual. The dominant gap class here is
 * SINGLE-PASS vs MULTI-PASS autofix: several upstream invalid cases are flagged
 * `// Would be nice to handle in one pass, but multipass works fine` and pin a
 * single-pass `output` that still contains a violation (upstream runs with
 * `verifyAfterFix: false`); rslint's `--fix` runs to a stable fixed point and
 * therefore yields a different (fully-fixed) `output`. The diagnostics still
 * match; only the `output` differs.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-one-expression-per-line', null as never, {
  valid: [
    {
      code: '<App />',
    },
    {
      code: `
\t\t\t\t<AllTabs>
\t\t\t\t\tFail
\t\t\t\t</AllTabs>
      `,
    },
    {
      code: `
\t\t\t\t<TagsWithTabs>
          Fail
\t\t\t\t</TagsWithTabs>
      `,
    },
    {
      code: `
        <ClosedTagWithTabs>
          Fail
\t\t\t\t</ClosedTagWithTabs>
      `,
    },
    {
      code: `
\t\t\t\t<OpenTagWithTabs>
          OK
        </OpenTagWithTabs>
      `,
    },
    {
      code: `
        <TextWithTabs>
\t\t\t\t\t\tOK
        </TextWithTabs>
      `,
    },
    {
      code: `
        <AllSpaces>
          OK
        </AllSpaces>
      `,
    },
    {
      code: '<App></App>',
    },
    {
      code: '<App foo="bar" />',
    },
    {
      code: `
        <App>
          <Foo />
        </App>
      `,
    },
    {
      code: `
        <App>
          <Foo />
          <Bar />
        </App>
      `,
    },
    {
      code: `
        <App>
          <Foo></Foo>
        </App>
      `,
    },
    {
      code: `
        <App>
          foo bar baz  whatever
        </App>
      `,
    },
    {
      code: `
        <App>
          <Foo>
          </Foo>
        </App>
      `,
    },
    {
      code: `
        <App
          foo="bar"
        >
        <Foo />
        </App>
      `,
    },
    {
      code: `
        <
        App
        >
          <
            Foo
          />
        </
        App
        >
      `,
    },
    {
      code: '<App>foo</App>',
      options: [{ allow: 'literal' }],
    },
    {
      code: '<App>123</App>',
      options: [{ allow: 'literal' }],
    },
    {
      code: '<App>foo</App>',
      options: [{ allow: 'single-child' }],
    },
    {
      code: '<App>{"foo"}</App>',
      options: [{ allow: 'single-child' }],
    },
    {
      code: '<App>123</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>foo</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>{"foo"}</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>{<Bar />}</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>{foo && <Bar />}</App>',
      options: [{ allow: 'single-child' }],
    },
    {
      code: '<App><Foo /></App>',
      options: [{ allow: 'single-child' }],
    },
    {
      code: '<></>',
    },
    {
      code: `
        <>
          <Foo />
        </>
      `,
    },
    {
      code: `
        <>
          <Foo />
          <Bar />
        </>
      `,
    },
    {
      code: '<App>Hello {name}</App>',
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <App>
          Hello {name} there!
        </App>`,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <App>
          Hello {<Bar />} there!
        </App>`,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <App>
          Hello {(<Bar />)} there!
        </App>`,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <App>
          Hello {(() => <Bar />)()} there!
        </App>`,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `<>
               123
               <Foo/>
             </>`,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <>
          <Foo/>
            Bar
            <Baz>
          </Baz>
        </>
      `,
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: '<App>{"foo"}</App>',
      options: [{ allow: 'single-line' }],
    },
    {
      code: '<App>{foo && <Bar />}</App>',
      options: [{ allow: 'single-line' }],
    },
    {
      code: '<App><Foo /></App>',
      options: [{ allow: 'single-line' }],
    },
    {
      code: `<>123<Foo/></>`,
      options: [{ allow: 'single-line' }],
    },
    {
      code: `
        <>
          123<Foo/><Bar/>
        </>
      `,
      options: [{ allow: 'single-line' }],
    },
    {
      code: `
        <>
          <Foo/><Bar/>{'baz'}
        </>
      `,
      options: [{ allow: 'single-line' }],
    },
    {
      code: `
        <>
          <Foo/><Bar/>baz
        </>
      `,
      options: [{ allow: 'single-line' }],
    },
    {
      code: `
        <>
          <Foo/>Bar<Baz></Baz>
        </>
      `,
      options: [{ allow: 'single-line' }],
    },
    {
      code: `
        <App>
          <Hello /> <ESLint />
        </App>
      `,
      options: [{ allow: 'single-line' }],
    },
    {
      code: `<App>{"Hello"} {"ESLint"}</App>`,
      options: [{ allow: 'single-line' }],
    },
    {
      code: `<App>Hello <span>ESLint</span></App>`,
      options: [{ allow: 'single-line' }],
    },
  ],
  invalid: [
    {
      code: `
        <App>{"foo"}</App>
      `,
      output: `
        <App>
{"foo"}
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{"foo"}' },
        },
      ],
    },
    {
      code: `
        <App>foo</App>
      `,
      output: `
        <App>
foo
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'foo' },
        },
      ],
    },
    {
      code: `<App
  foo
>bar
</App>`,
      output: `<App
  foo
>
bar
</App>`,
      errors: [
        { messageId: 'moveToNewLine' },
      ],
    },
    {
      code: `
        <div>
          foo {"bar"}
        </div>
      `,
      output: `
        <div>
          foo${' '/* intentional trailing space */}
{' '}
{"bar"}
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{"bar"}' },
        },
      ],
    },
    {
      code: `
        <div>
          {"foo"} bar
        </div>
      `,
      output: `
        <div>
          {"foo"}
{' '}
bar
</div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: ' bar        ' },
        },
      ],
    },
    {
      code: `
        <App>
          <Foo /><Bar />
        </App>
      `,
      output: `
        <App>
          <Foo />
<Bar />
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bar' },
        },
      ],
    },
    {
      code: `
        <div>
          <span />foo
        </div>
      `,
      output: `
        <div>
          <span />
foo
</div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'foo        ' },
        },
      ],
    },
    {
      code: `
        <div>
          <span />{"foo"}
        </div>
      `,
      output: `
        <div>
          <span />
{"foo"}
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{"foo"}' },
        },
      ],
    },
    {
      code: `
        <div>
          {"foo"} { I18n.t('baz') }
        </div>
      `,
      output: `
        <div>
          {"foo"}${' '/* intentional trailing space */}
{' '}
{ I18n.t('baz') }
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{ I18n.t(\'baz\') }' },
        },
      ],
    },
    {
      code: `
        <Text style={styles.foo}>{ bar } <Text/> { I18n.t('baz') }</Text>
      `,
      output: `
        <Text style={styles.foo}>
{ bar }${' '/* intentional trailing space */}
{' '}
<Text/>${' '/* intentional trailing space */}
{' '}
{ I18n.t('baz') }
</Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{ bar }' },
        },
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Text' },
        },
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{ I18n.t(\'baz\') }' },
        },
      ],

    },
    {
      code: `
        <Text style={styles.foo}> <Bar/> <Baz/></Text>
      `,
      output: `
        <Text style={styles.foo}>${' '/* intentional trailing space */}
{' '}
<Bar/>${' '/* intentional trailing space */}
{' '}
<Baz/>
</Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bar' },
        },
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Baz' },
        },
      ],
    },
    {
      code: `
        <Text style={styles.foo}> <Bar/> <Baz/> <Bunk/> <Bruno/> </Text>
      `,
      output: `
        <Text style={styles.foo}>${' '/* intentional trailing space */}
{' '}
<Bar/>${' '/* intentional trailing space */}
{' '}
<Baz/>${' '/* intentional trailing space */}
{' '}
<Bunk/>${' '/* intentional trailing space */}
{' '}
<Bruno/>
{' '}
 </Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bar' },
        },
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Baz' },
        },
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bunk' },
        },
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bruno' },
        },
      ],
    },
    {
      code: `
        <Text style={styles.foo}> <Bar /></Text>
      `,
      output: `
        <Text style={styles.foo}>${' '/* intentional trailing space */}
{' '}
<Bar />
</Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bar' },
        },
      ],
    },
    {
      code: `
        <Text style={styles.foo}> <Bar />
        </Text>
      `,
      output: `
        <Text style={styles.foo}>${' '/* intentional trailing space */}
{' '}
<Bar />
        </Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bar' },
        },
      ],
    },
    {
      code: `
        <Text style={styles.foo}>
          <Bar /> <Baz />
        </Text>
      `,
      output: `
        <Text style={styles.foo}>
          <Bar />${' '/* intentional trailing space */}
{' '}
<Baz />
        </Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Baz' },
        },
      ],
    },
    {
      code: `
        <Text style={styles.foo}>
          <Bar /> <Baz />
        </Text>
      `,
      output: `
        <Text style={styles.foo}>
          <Bar />${' '/* intentional trailing space */}
{' '}
<Baz />
        </Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Baz' },
        },
      ],
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
        <Text style={styles.foo}>
          { bar } { I18n.t('baz') }
        </Text>
      `,
      output: `
        <Text style={styles.foo}>
          { bar }${' '/* intentional trailing space */}
{' '}
{ I18n.t('baz') }
        </Text>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{ I18n.t(\'baz\') }' },
        },
      ],
    },
    {
      code: `
        <div>
          foo<input />
        </div>
      `,
      output: `
        <div>
          foo
<input />
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'input' },
        },
      ],
    },
    {
      code: `
        <div>
          {"foo"}<span />
        </div>
      `,
      output: `
        <div>
          {"foo"}
<span />
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'span' },
        },
      ],
    },
    {
      code: `
        <div>
          foo <input />
        </div>
      `,
      output: `
        <div>
          foo${' '/* intentional trailing space */}
{' '}
<input />
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'input' },
        },
      ],
    },
    {
      code: `
        <div>
          <input /> foo
        </div>
      `,
      output: `
        <div>
          <input />
{' '}
foo
</div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: ' foo        ' },
        },
      ],
    },
    {
      code: `
        <div>
          <span /> <input />
        </div>
      `,
      output: `
        <div>
          <span />${' '/* intentional trailing space */}
{' '}
<input />
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'input' },
        },
      ],
    },
    {
      code: `
        <div>
          <span />
        {' '}<input />
        </div>
      `,
      output: `
        <div>
          <span />
        {' '}
<input />
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'input' },
        },
      ],
    },
    {
      code: `
        <div>
          {"foo"} <input />
        </div>
      `,
      output: `
        <div>
          {"foo"}${' '/* intentional trailing space */}
{' '}
<input />
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'input' },
        },
      ],
    },
    {
      code: `
        <div>
          <input /> {"foo"}
        </div>
      `,
      output: `
        <div>
          <input />${' '/* intentional trailing space */}
{' '}
{"foo"}
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{"foo"}' },
        },
      ],
    },
    {
      code: `
        <App>
          <Foo></Foo><Bar></Bar>
        </App>
      `,
      output: `
        <App>
          <Foo></Foo>
<Bar></Bar>
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bar' },
        },
      ],
    },
    {
      code: `
        <App>
        <Foo></Foo></App>
      `,
      output: `
        <App>
        <Foo></Foo>
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App><Foo />
        </App>
      `,
      output: `
        <App>
<Foo />
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App>
        <Foo/></App>
      `,
      output: `
        <App>
        <Foo/>
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App><Foo
        />
        </App>
      `,
      output: `
        <App>
<Foo
        />
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App
        >
        <Foo /></App>
      `,
      output: `
        <App
        >
        <Foo />
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App
        >
        <Foo
        /></App>
      `,
      output: `
        <App
        >
        <Foo
        />
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App
        ><Foo />
        </App>
      `,
      output: `
        <App
        >
<Foo />
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App>
          <Foo></Foo
        ></App>
      `,
      output: `
        <App>
          <Foo></Foo
        >
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App>
          <Foo></
        Foo></App>
      `,
      output: `
        <App>
          <Foo></
        Foo>
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App>
          <Foo></
        Foo><Bar />
        </App>
      `,
      output: `
        <App>
          <Foo></
        Foo>
<Bar />
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bar' },
        },
      ],
    },
    {
      code: `
        <App>
          <Foo>
            <Bar /></Foo>
        </App>
      `,
      output: `
        <App>
          <Foo>
            <Bar />
</Foo>
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Bar' },
        },
      ],
    },
    {
      code: `
        <App>
          <Foo>
            <Bar> baz </Bar>
          </Foo>
        </App>
      `,
      output: `
        <App>
          <Foo>
            <Bar>
{' '}
baz
{' '}
</Bar>
          </Foo>
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: ' baz ' },
        },
      ],
    },
    // (the `foo {"bar"} baz` single-pass case that pinned a still-unfixed
    //  `output` is moved to KNOWN GAPS — see bottom)
    {
    // Would be nice to handle in one pass, but multipass works fine.
      code: `
        <App>
          foo {"bar"}
        </App>
      `,
      output: `
        <App>
          foo${' '/* intentional trailing space */}
{' '}
{"bar"}
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{"bar"}' },
        },
      ],
    },
    {
    // Would be nice to handle in one pass, but multipass works fine.
      code: `
        <App>
          foo
        {' '}
        {"bar"} baz
        </App>
      `,
      output: `
        <App>
          foo
        {' '}
        {"bar"}
{' '}
baz
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: ' baz        ' },
        },
      ],
    },
    // (the blank-line-padded `foo {"bar"} baz` single-pass case is moved to
    //  KNOWN GAPS — see bottom)
    {
    // Would be nice to handle in one pass, but multipass works fine.
      code: `
        <App>

          foo
        {' '}
        {"bar"} baz

        </App>
      `,
      output: `
        <App>

          foo
        {' '}
        {"bar"}
{' '}
baz

</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: ' baz        ' },
        },
      ],
    },
    {
      code: `
        <App>{
          foo
        }</App>
      `,
      output: `
        <App>
{
          foo
        }
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{          foo        }' },
        },
      ],
    },
    {
      code: `
        <App> {
          foo
        } </App>
      `,
      output: `
        <App>${' '/* intentional trailing space */}
{' '}
{
          foo
        }
{' '}
 </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{          foo        }' },
        },
      ],
    },
    {
      code: `
        <App>
        {' '}
        {
          foo
        } </App>
      `,
      output: `
        <App>
        {' '}
        {
          foo
        }
{' '}
 </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{          foo        }' },
        },
      ],
    },
    {
      code: `
        <App><Foo /></App>
      `,
      output: `
        <App>
<Foo />
</App>
      `,
      options: [{ allow: 'none' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App>foo</App>
      `,
      output: `
        <App>
foo
</App>
      `,
      options: [{ allow: 'none' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'foo' },
        },
      ],
    },
    {
      code: `
        <App>{"foo"}</App>
      `,
      output: `
        <App>
{"foo"}
</App>
      `,
      options: [{ allow: 'none' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{"foo"}' },
        },
      ],
    },
    {
      code: `
        <App>foo
        </App>
      `,
      output: `
        <App>
foo
</App>
      `,
      options: [{ allow: 'literal' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'foo        ' },
        },
      ],
    },
    {
      code: `
        <App><Foo /></App>
      `,
      output: `
        <App>
<Foo />
</App>
      `,
      options: [{ allow: 'literal' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App><Foo /></App>
      `,
      output: `
        <App>
<Foo />
</App>
      `,
      options: [{ allow: 'non-jsx' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <App
          foo="1"
          bar="2"
        >baz</App>
      `,
      options: [{ allow: 'literal' }],
      output: `
        <App
          foo="1"
          bar="2"
        >
baz
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'baz' },
        },
      ],
    },
    {
      code: `
        <App>foo
        bar
        </App>
      `,
      options: [{ allow: 'literal' }],
      output: `
        <App>
foo
        bar
</App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'foo        bar        ' },
        },
      ],
    },
    {
      code: `
        <>{"foo"}</>
      `,
      output: `
        <>
{"foo"}
</>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{"foo"}' },
        },
      ],
    },
    {
      code: `
        <App>
          <Foo /><></>
        </App>
      `,
      output: `
        <App>
          <Foo />
<></>
        </App>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '<></>' },
        },
      ],
    },
    {
      code: `
        <
        ><Foo />
        </>
      `,
      output: `
        <
        >
<Foo />
        </>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Foo' },
        },
      ],
    },
    {
      code: `
        <div>
        <MyComponent>a</MyComponent>
        <MyOther>{a}</MyOther>
        </div>
      `,
      output: `
        <div>
        <MyComponent>
a
</MyComponent>
        <MyOther>
{a}
</MyOther>
        </div>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'a' },
        },
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{a}' },
        },
      ],
    },
    // (the first-pass `<h1>{"Hi people"}<button/></h1>` case is moved to KNOWN
    //  GAPS — see bottom; its second pass below stays live)
    {
      code: `
        const IndexPage = () => (
          <h1>
{"Hi people"}<button/></h1>
        );
      `,
      output: `
        const IndexPage = () => (
          <h1>
{"Hi people"}
<button/>
</h1>
        );
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'button' },
        },
      ],
    },
    // (the first-pass multi-element `<Layout>` case is moved to KNOWN GAPS —
    //  see bottom; its second pass below stays live)
    {
      code: `
        <Layout>
        <p>
Welcome to your new Gatsby site.
</p>
        <p>
Now go build something great.
</p>
        <h1>
Hi people<button/></h1>
        </Layout>
      `,
      output: `
        <Layout>
        <p>
Welcome to your new Gatsby site.
</p>
        <p>
Now go build something great.
</p>
        <h1>
Hi people
<button/>
</h1>
        </Layout>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'button' },
        },
      ],
    },
    // (the first-pass `</div><Link ...>` case is moved to KNOWN GAPS — see
    //  bottom; its second pass below stays live)
    {
      code: `
        <Layout>
          <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}>
            <Image />
          </div>
<Link to="/page-2/">Go to page 2</Link>
        </Layout>
      `,
      output: `
        <Layout>
          <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}>
            <Image />
          </div>
<Link to="/page-2/">
Go to page 2
</Link>
        </Layout>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Go to page 2' },
        },
      ],
    },
    {
      code: `
<Layout>
  <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}><Image /></div>{'Bar'}
  <Link to="/page-2/">Go to page 2</Link>
</Layout>
      `,
      output: `
<Layout>
  <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}>
<Image />
</div>
{'Bar'}
  <Link to="/page-2/">Go to page 2</Link>
</Layout>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'Image' },
        },
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{\'Bar\'}' },
        },
      ],
      options: [{ allow: 'non-jsx' }],
    },
    {
      code: `
<Layout>
  <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}><Image /></div>{'Bar'}
  <Link to="/page-2/">Go to page 2</Link>
</Layout>
      `,
      output: `
<Layout>
  <div style={{ maxWidth: \`300px\`, marginBottom: \`1.45rem\` }}><Image /></div>
{'Bar'}
  <Link to="/page-2/">Go to page 2</Link>
</Layout>
      `,
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: '{\'Bar\'}' },
        },
      ],
      options: [{ allow: 'single-line' }],
    },
    {
      code: `<div><span>foo</span>
</div>`,
      output: `<div>
<span>foo</span>
</div>`,
      options: [{ allow: 'single-line' }],
      errors: [
        {
          messageId: 'moveToNewLine',
          data: { descriptor: 'span' },
        },
      ],
    },
    // https://github.com/eslint-stylistic/eslint-stylistic/issues/869
    {
      code: `<App
  foo
>Up to {percent}% Off
</App>`,
      output: `<App
  foo
>
Up to {percent}% Off
</App>`,
      options: [{ allow: 'single-line' }],
      errors: [
        { messageId: 'moveToNewLine' },
      ],
    },
    {
      code: `<App
  foo
>
  Up to {percent}% Off</App>`,
      output: `<App
  foo
>
  Up to {percent}% Off
</App>`,
      options: [{ allow: 'single-line' }],
      errors: [
        { messageId: 'moveToNewLine' },
      ],
    },
  ],
});

/*
 * ================= jsx-one-expression-per-line — KNOWN GAPS =================
 *
 * GAP CLASS: SINGLE-PASS vs MULTI-PASS AUTOFIX (diagnostics identical; only
 * the fixed `output` differs). Upstream runs eslint-vitest-rule-tester with
 * `verifyAfterFix: false`, so each pinned `output` is the result of ONE fix
 * pass and may still contain a violation — every case below is flagged in
 * upstream source as `Would be nice to handle in one pass` / `TODO: handle in
 * a single pass`. rslint's `--fix` runs to a STABLE fixed point, re-applying
 * the same fix until no violation remains, so it yields the fully-split
 * output. The diagnostic set (count + rendered message + line) matches
 * upstream exactly for all of these; they are preserved here verbatim (as the
 * exact strings rslint/ESLint see) rather than asserted against a single-pass
 * `output` rslint by design does not reproduce.
 *
 * Strings are JSON-encoded so whitespace/newlines are unambiguous.
 *
 * ---- upstream invalid case  [// Would be nice to handle in one pass, but multipass works fine.] ----
 *   single-pass: text + expression + text on one line (foo {"bar"} baz)
 *   options: (default)   errors: 2x moveToNewLine (descriptor '{"bar"}' then ' baz        ')
 *   code:            "\n        <App>\n          foo {\"bar\"} baz\n        </App>\n      "
 *   upstream output: "\n        <App>\n          foo \n{' '}\n{\"bar\"} baz\n        </App>\n      "
 *   rslint  output:  "\n        <App>\n          foo \n{' '}\n{\"bar\"}\n{' '}\nbaz\n</App>\n      "
 *
 * ---- upstream invalid case  [// Would be nice to handle in one pass, but multipass works fine.] ----
 *   single-pass: same foo {"bar"} baz, padded with blank lines
 *   options: (default)   errors: 2x moveToNewLine (descriptor '{"bar"}' then ' baz        ')
 *   code:            "\n        <App>\n\n          foo {\"bar\"} baz\n\n        </App>\n      "
 *   upstream output: "\n        <App>\n\n          foo \n{' '}\n{\"bar\"} baz\n\n        </App>\n      "
 *   rslint  output:  "\n        <App>\n\n          foo \n{' '}\n{\"bar\"}\n{' '}\nbaz\n\n</App>\n      "
 *
 * ---- upstream invalid case  [// TODO: handle in a single pass] ----
 *   single-pass: <h1>{"Hi people"}<button/></h1>
 *   options: (default)   errors: 2x moveToNewLine (descriptor '{"Hi people"}' then 'button')
 *   code:            "\n        const IndexPage = () => (\n          <h1>{\"Hi people\"}<button/></h1>\n        );\n      "
 *   upstream output: "\n        const IndexPage = () => (\n          <h1>\n{\"Hi people\"}<button/></h1>\n        );\n      "
 *   rslint  output:  "\n        const IndexPage = () => (\n          <h1>\n{\"Hi people\"}\n<button/>\n</h1>\n        );\n      "
 *
 * ---- upstream invalid case  [// TODO: handle in a single pass (see above)] ----
 *   single-pass: multi-element <Layout> (two <p> + <h1>Hi people<button/>)
 *   options: (default)   errors: 4x moveToNewLine (descriptors 'Welcome to your new Gatsby site.', 'Now go build something great.', 'Hi people', 'button')
 *   code:            "\n        <Layout>\n        <p>Welcome to your new Gatsby site.</p>\n        <p>Now go build something great.</p>\n        <h1>Hi people<button/></h1>\n        </Layout>\n      "
 *   upstream output: "\n        <Layout>\n        <p>\nWelcome to your new Gatsby site.\n</p>\n        <p>\nNow go build something great.\n</p>\n        <h1>\nHi people<button/></h1>\n        </Layout>\n      "
 *   rslint  output:  "\n        <Layout>\n        <p>\nWelcome to your new Gatsby site.\n</p>\n        <p>\nNow go build something great.\n</p>\n        <h1>\nHi people\n<button/>\n</h1>\n        </Layout>\n      "
 *
 * ---- upstream invalid case  [// TODO: handle in a single pass] ----
 *   single-pass: </div><Link to="/page-2/">Go to page 2</Link>
 *   options: (default)   errors: 2x moveToNewLine (descriptor 'Link' then 'Go to page 2')
 *   code:            "\n        <Layout>\n          <div style={{ maxWidth: `300px`, marginBottom: `1.45rem` }}>\n            <Image />\n          </div><Link to=\"/page-2/\">Go to page 2</Link>\n        </Layout>\n      "
 *   upstream output: "\n        <Layout>\n          <div style={{ maxWidth: `300px`, marginBottom: `1.45rem` }}>\n            <Image />\n          </div>\n<Link to=\"/page-2/\">Go to page 2</Link>\n        </Layout>\n      "
 *   rslint  output:  "\n        <Layout>\n          <div style={{ maxWidth: `300px`, marginBottom: `1.45rem` }}>\n            <Image />\n          </div>\n<Link to=\"/page-2/\">\nGo to page 2\n</Link>\n        </Layout>\n      "
 *
 */
