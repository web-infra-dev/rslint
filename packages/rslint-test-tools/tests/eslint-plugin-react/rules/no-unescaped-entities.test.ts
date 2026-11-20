import { RuleTester, type ExpectedSuggestion } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectUnescapedEntityAltsError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unescapedEntityAlts',
    message: "`'` can be escaped with `&apos;`, `&lsquo;`, `&#39;`, `&rsquo;`.",
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

function expectUnescapedEntityAltsError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
  suggestions?: ExpectedSuggestion[],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unescapedEntityAlts',
    message: '`"` can be escaped with `&quot;`, `&ldquo;`, `&#34;`, `&rdquo;`.',
    line,
    column,
    endLine,
    endColumn,
    ...(suggestions === undefined ? {} : { suggestions }),
  };
}

function expectUnescapedEntityError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unescapedEntity',
    message: 'HTML entity, `&` , must be escaped.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-unescaped-entities', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return (
              <div/>
            );
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>Here is some text!</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>I&rsquo;ve escaped some entities: &gt; &lt; &amp;</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>first line is ok
            so is second
            and here are some escaped entities: &gt; &lt; &amp;</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>{">" + "<" + "&" + '"'}</div>;
          },
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <>Here is some text!</>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <>I&rsquo;ve escaped some entities: &gt; &lt; &amp;</>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <>{">" + "<" + "&" + '"'}</>;
          },
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `var Hello = <div>'</div>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 18, 1, 19],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: 'var Hello = <div>&apos;</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: 'var Hello = <div>&lsquo;</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = <div>&#39;</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: 'var Hello = <div>&rsquo;</div>',
            },
          ],
        ),
      ],
    },
    {
      code: `var Hello = <div>Don't do that</div>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 21, 1, 22],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: 'var Hello = <div>Don&apos;t do that</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: 'var Hello = <div>Don&lsquo;t do that</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = <div>Don&#39;t do that</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: 'var Hello = <div>Don&rsquo;t do that</div>',
            },
          ],
        ),
      ],
    },
    {
      code: `var Hello = <>it's a trap</>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 17, 1, 18],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: 'var Hello = <>it&apos;s a trap</>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: 'var Hello = <>it&lsquo;s a trap</>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = <>it&#39;s a trap</>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: 'var Hello = <>it&rsquo;s a trap</>',
            },
          ],
        ),
      ],
    },
    // nested: flag inner only
    {
      code: `var Hello = <Outer>outer text<Inner>inner's</Inner></Outer>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 42, 1, 43],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output:
                'var Hello = <Outer>outer text<Inner>inner&apos;s</Inner></Outer>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output:
                'var Hello = <Outer>outer text<Inner>inner&lsquo;s</Inner></Outer>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output:
                'var Hello = <Outer>outer text<Inner>inner&#39;s</Inner></Outer>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output:
                'var Hello = <Outer>outer text<Inner>inner&rsquo;s</Inner></Outer>',
            },
          ],
        ),
      ],
    },
    // nested: flag at multiple levels
    {
      code: `var Hello = <Outer>outer's<Inner>inner's</Inner></Outer>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 25, 1, 26],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output:
                "var Hello = <Outer>outer&apos;s<Inner>inner's</Inner></Outer>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output:
                "var Hello = <Outer>outer&lsquo;s<Inner>inner's</Inner></Outer>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output:
                "var Hello = <Outer>outer&#39;s<Inner>inner's</Inner></Outer>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output:
                "var Hello = <Outer>outer&rsquo;s<Inner>inner's</Inner></Outer>",
            },
          ],
        ),
        expectUnescapedEntityAltsError1(
          [1, 39, 1, 40],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output:
                "var Hello = <Outer>outer's<Inner>inner&apos;s</Inner></Outer>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output:
                "var Hello = <Outer>outer's<Inner>inner&lsquo;s</Inner></Outer>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output:
                "var Hello = <Outer>outer's<Inner>inner&#39;s</Inner></Outer>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output:
                "var Hello = <Outer>outer's<Inner>inner&rsquo;s</Inner></Outer>",
            },
          ],
        ),
      ],
    },
    // JSX inside expression container inside JSX
    {
      code: `var Hello = <Outer>{<Inner>inner's</Inner>}</Outer>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 33, 1, 34],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output:
                'var Hello = <Outer>{<Inner>inner&apos;s</Inner>}</Outer>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output:
                'var Hello = <Outer>{<Inner>inner&lsquo;s</Inner>}</Outer>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = <Outer>{<Inner>inner&#39;s</Inner>}</Outer>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output:
                'var Hello = <Outer>{<Inner>inner&rsquo;s</Inner>}</Outer>',
            },
          ],
        ),
      ],
    },
    // JSX as prop value: inner JsxText is scanned
    {
      code: `var Hello = <Outer child={<Inner>inner's</Inner>} />`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 39, 1, 40],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output:
                'var Hello = <Outer child={<Inner>inner&apos;s</Inner>} />',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output:
                'var Hello = <Outer child={<Inner>inner&lsquo;s</Inner>} />',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output:
                'var Hello = <Outer child={<Inner>inner&#39;s</Inner>} />',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output:
                'var Hello = <Outer child={<Inner>inner&rsquo;s</Inner>} />',
            },
          ],
        ),
      ],
    },
    // adjacent duplicates
    {
      code: `var Hello = <div>a''b</div>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 19, 1, 20],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: "var Hello = <div>a&apos;'b</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: "var Hello = <div>a&lsquo;'b</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: "var Hello = <div>a&#39;'b</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: "var Hello = <div>a&rsquo;'b</div>",
            },
          ],
        ),
        expectUnescapedEntityAltsError1(
          [1, 20, 1, 21],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: "var Hello = <div>a'&apos;b</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: "var Hello = <div>a'&lsquo;b</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: "var Hello = <div>a'&#39;b</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: "var Hello = <div>a'&rsquo;b</div>",
            },
          ],
        ),
      ],
    },
    // mixed default chars in same JsxText (source order)
    {
      code: `var Hello = <div>"don't"</div>`,
      errors: [
        expectUnescapedEntityAltsError2(
          [1, 18, 1, 19],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&quot;`.',
              output: 'var Hello = <div>&quot;don\'t"</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&ldquo;`.',
              output: 'var Hello = <div>&ldquo;don\'t"</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#34;`.',
              output: 'var Hello = <div>&#34;don\'t"</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rdquo;`.',
              output: 'var Hello = <div>&rdquo;don\'t"</div>',
            },
          ],
        ),
        expectUnescapedEntityAltsError1(
          [1, 22, 1, 23],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: 'var Hello = <div>"don&apos;t"</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: 'var Hello = <div>"don&lsquo;t"</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = <div>"don&#39;t"</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: 'var Hello = <div>"don&rsquo;t"</div>',
            },
          ],
        ),
        expectUnescapedEntityAltsError2(
          [1, 24, 1, 25],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&quot;`.',
              output: 'var Hello = <div>"don\'t&quot;</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&ldquo;`.',
              output: 'var Hello = <div>"don\'t&ldquo;</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#34;`.',
              output: 'var Hello = <div>"don\'t&#34;</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rdquo;`.',
              output: 'var Hello = <div>"don\'t&rdquo;</div>',
            },
          ],
        ),
      ],
    },
    // script tag: unescaped quotes
    {
      code: `var Hello = <script>window.foo = "bar"</script>`,
      errors: [
        expectUnescapedEntityAltsError2(
          [1, 34, 1, 35],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&quot;`.',
              output: 'var Hello = <script>window.foo = &quot;bar"</script>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&ldquo;`.',
              output: 'var Hello = <script>window.foo = &ldquo;bar"</script>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#34;`.',
              output: 'var Hello = <script>window.foo = &#34;bar"</script>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rdquo;`.',
              output: 'var Hello = <script>window.foo = &rdquo;bar"</script>',
            },
          ],
        ),
        expectUnescapedEntityAltsError2(
          [1, 38, 1, 39],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&quot;`.',
              output: 'var Hello = <script>window.foo = "bar&quot;</script>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&ldquo;`.',
              output: 'var Hello = <script>window.foo = "bar&ldquo;</script>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#34;`.',
              output: 'var Hello = <script>window.foo = "bar&#34;</script>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rdquo;`.',
              output: 'var Hello = <script>window.foo = "bar&rdquo;</script>',
            },
          ],
        ),
      ],
    },
    // custom forbid: simple string form
    {
      code: `var Hello = <span>foo & bar</span>`,
      options: [{ forbid: ['&'] }],
      errors: [expectUnescapedEntityError([1, 23, 1, 24])],
    },
    // custom forbid: object form
    {
      code: `var Hello = <span>foo & bar</span>`,
      options: [{ forbid: [{ char: '&', alternatives: ['&amp;'] }] }],
      errors: [
        {
          messageId: 'unescapedEntityAlts',
          message: '`&` can be escaped with `&amp;`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 24,
          suggestions: [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&amp;`.',
              output: 'var Hello = <span>foo &amp; bar</span>',
            },
          ],
        },
      ],
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
        expectUnescapedEntityError([1, 20, 1, 21]),
        {
          messageId: 'unescapedEntityAlts',
          message: '`$` can be escaped with `&#36;`.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 25,
          suggestions: [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#36;`.',
              output: 'var Hello = <div>a & b &#36; c</div>',
            },
          ],
        },
      ],
    },
    // multi-byte Unicode custom forbid char (fullwidth apostrophe U+FF07)
    {
      code: `var Hello = <div>fullwidth\uff07quote</div>`,
      options: [{ forbid: ['\uff07'] }],
      errors: [
        {
          messageId: 'unescapedEntity',
          message: 'HTML entity, `＇` , must be escaped.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    // UTF-16 column stays correct when a multi-byte char precedes the match
    {
      code: `var Hello = <div>café's</div>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 22, 1, 23],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: 'var Hello = <div>café&apos;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: 'var Hello = <div>café&lsquo;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = <div>café&#39;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: 'var Hello = <div>café&rsquo;s</div>',
            },
          ],
        ),
      ],
    },
    // UTF-16 column stays correct across a surrogate pair (supra-BMP char)
    {
      code: `var Hello = <div>🚀's</div>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 20, 1, 21],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: 'var Hello = <div>🚀&apos;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: 'var Hello = <div>🚀&lsquo;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = <div>🚀&#39;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: 'var Hello = <div>🚀&rsquo;s</div>',
            },
          ],
        ),
      ],
    },
    // JSX in type assertion
    {
      code: `var Hello = (<div>a's</div>) as any`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 20, 1, 21],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: 'var Hello = (<div>a&apos;s</div>) as any',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: 'var Hello = (<div>a&lsquo;s</div>) as any',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = (<div>a&#39;s</div>) as any',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: 'var Hello = (<div>a&rsquo;s</div>) as any',
            },
          ],
        ),
      ],
    },
    // JSX in logical short-circuit
    {
      code: `var Hello = cond && <div>a's</div>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 27, 1, 28],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: 'var Hello = cond && <div>a&apos;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: 'var Hello = cond && <div>a&lsquo;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: 'var Hello = cond && <div>a&#39;s</div>',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: 'var Hello = cond && <div>a&rsquo;s</div>',
            },
          ],
        ),
      ],
    },
    // Two JsxText nodes split by an expression container
    {
      code: `var Hello = <div>a's {x} b's</div>`,
      errors: [
        expectUnescapedEntityAltsError1(
          [1, 19, 1, 20],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: "var Hello = <div>a&apos;s {x} b's</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: "var Hello = <div>a&lsquo;s {x} b's</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: "var Hello = <div>a&#39;s {x} b's</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: "var Hello = <div>a&rsquo;s {x} b's</div>",
            },
          ],
        ),
        expectUnescapedEntityAltsError1(
          [1, 27, 1, 28],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output: "var Hello = <div>a's {x} b&apos;s</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output: "var Hello = <div>a's {x} b&lsquo;s</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output: "var Hello = <div>a's {x} b&#39;s</div>",
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output: "var Hello = <div>a's {x} b&rsquo;s</div>",
            },
          ],
        ),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      skip: "The native TypeScript parser rejects this JSX form (Unexpected token. Did you mean `{'>'}` or `&gt;`?); upstream uses a different parser.",
      // Preserve the historical fixture without inventing native diagnostics.
      create: () => {
        throw new Error(
          JSON.stringify({
            code: '\n        var Hello = createReactClass({\n          render: function() {\n            return <>> babel-eslint</>;\n          }\n        });\n      ',
            settings: {},
            parserOptions: {
              ecmaVersion: 2018,
              sourceType: 'module',
              ecmaFeatures: {
                jsx: true,
              },
            },
            features: ['fragment', 'no-ts', 'no-default'],
            upstreamErrors: [
              {
                messageId: 'unescapedEntityAlts',
                data: {
                  entity: '>',
                  alts: '`&gt;`',
                },
                suggestions: [
                  {
                    messageId: 'replaceWithAlt',
                    data: {
                      alt: '&gt;',
                    },
                    output:
                      '\n        var Hello = createReactClass({\n          render: function() {\n            return <>&gt; babel-eslint</>;\n          }\n        });\n      ',
                  },
                ],
              },
            ],
          }),
        );
      },
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      skip: "The native TypeScript parser rejects this JSX form (Unexpected token. Did you mean `{'>'}` or `&gt;`?); upstream uses a different parser.",
      // Preserve the historical fixture without inventing native diagnostics.
      create: () => {
        throw new Error(
          JSON.stringify({
            code: '\n        var Hello = createReactClass({\n          render: function() {\n            return <>first line is ok\n            so is second\n            and here are some bad entities: ></>\n          }\n        });\n      ',
            settings: {},
            parserOptions: {
              ecmaVersion: 2018,
              sourceType: 'module',
              ecmaFeatures: {
                jsx: true,
              },
            },
            features: ['fragment', 'no-ts', 'no-default'],
            upstreamErrors: [
              {
                messageId: 'unescapedEntityAlts',
                data: {
                  entity: '>',
                  alts: '`&gt;`',
                },
                suggestions: [
                  {
                    messageId: 'replaceWithAlt',
                    data: {
                      alt: '&gt;',
                    },
                    output:
                      '\n        var Hello = createReactClass({\n          render: function() {\n            return <>first line is ok\n            so is second\n            and here are some bad entities: &gt;</>\n          }\n        });\n      ',
                  },
                ],
              },
            ],
          }),
        );
      },
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div>'</div>;
          }
        });
      `,
      errors: [
        expectUnescapedEntityAltsError1(
          [4, 25, 4, 26],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&apos;`.',
              output:
                '\n        var Hello = createReactClass({\n          render: function() {\n            return <div>&apos;</div>;\n          }\n        });\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&lsquo;`.',
              output:
                '\n        var Hello = createReactClass({\n          render: function() {\n            return <div>&lsquo;</div>;\n          }\n        });\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#39;`.',
              output:
                '\n        var Hello = createReactClass({\n          render: function() {\n            return <div>&#39;</div>;\n          }\n        });\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rsquo;`.',
              output:
                '\n        var Hello = createReactClass({\n          render: function() {\n            return <div>&rsquo;</div>;\n          }\n        });\n      ',
            },
          ],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      skip: "The native TypeScript parser rejects this JSX form (Unexpected token. Did you mean `{'}'}` or `&rbrace;`?); upstream uses a different parser.",
      // Preserve the historical fixture without inventing native diagnostics.
      create: () => {
        throw new Error(
          JSON.stringify({
            code: '\n        var Hello = createReactClass({\n          render: function() {\n            return <>{"Unbalanced braces - babel-eslint"}}</>;\n          }\n        });\n      ',
            settings: {},
            parserOptions: {
              ecmaVersion: 2018,
              sourceType: 'module',
              ecmaFeatures: {
                jsx: true,
              },
            },
            features: ['fragment', 'no-ts', 'no-default'],
            upstreamErrors: [
              {
                messageId: 'unescapedEntityAlts',
                data: {
                  entity: '}',
                  alts: '`&#125;`',
                },
                suggestions: [
                  {
                    messageId: 'replaceWithAlt',
                    data: {
                      alt: '&#125;',
                    },
                    output:
                      '\n        var Hello = createReactClass({\n          render: function() {\n            return <>{"Unbalanced braces - babel-eslint"}&#125;</>;\n          }\n        });\n      ',
                  },
                ],
              },
            ],
          }),
        );
      },
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <>foo & bar</>;
          }
        });
      `,
      options: [{ forbid: ['&'] }],
      errors: [expectUnescapedEntityError([4, 26, 4, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <span>foo & bar</span>;
          }
        });
      `,
      options: [{ forbid: ['&'] }],
      errors: [expectUnescapedEntityError([4, 30, 4, 31])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <span>foo & bar</span>;
          }
        });
      `,
      options: [{ forbid: [{ char: '&', alternatives: ['&amp;'] }] }],
      errors: [
        {
          messageId: 'unescapedEntityAlts',
          message: '`&` can be escaped with `&amp;`.',
          line: 4,
          column: 30,
          endLine: 4,
          endColumn: 31,
          suggestions: [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&amp;`.',
              output:
                '\n        var Hello = createReactClass({\n          render: function() {\n            return <span>foo &amp; bar</span>;\n          }\n        });\n      ',
            },
          ],
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        <script>window.foo = "bar"</script>
      `,
      errors: [
        expectUnescapedEntityAltsError2(
          [2, 30, 2, 31],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&quot;`.',
              output:
                '\n        <script>window.foo = &quot;bar"</script>\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&ldquo;`.',
              output:
                '\n        <script>window.foo = &ldquo;bar"</script>\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#34;`.',
              output:
                '\n        <script>window.foo = &#34;bar"</script>\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rdquo;`.',
              output:
                '\n        <script>window.foo = &rdquo;bar"</script>\n      ',
            },
          ],
        ),
        expectUnescapedEntityAltsError2(
          [2, 34, 2, 35],
          [
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&quot;`.',
              output:
                '\n        <script>window.foo = "bar&quot;</script>\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&ldquo;`.',
              output:
                '\n        <script>window.foo = "bar&ldquo;</script>\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&#34;`.',
              output:
                '\n        <script>window.foo = "bar&#34;</script>\n      ',
            },
            {
              messageId: 'replaceWithAlt',
              desc: 'Replace with `&rdquo;`.',
              output:
                '\n        <script>window.foo = "bar&rdquo;</script>\n      ',
            },
          ],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
