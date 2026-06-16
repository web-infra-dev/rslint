/**
 * @fileoverview Validate props indentation in JSX
 * @author Vitor Balocco
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-indent-props/jsx-indent-props.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-indent-props', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaVersion: 2018`, `sourceType: 'module'`,
 *    `ecmaFeatures.jsx: true`) dropped — rslint resolves via tsconfig and the
 *    RuleTester routes JSX fixtures to `.tsx`.
 *
 * The upstream file wraps its cases in the `valids()` / `invalids()` helpers from
 * `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case across
 * several PARSERS (ESLint-default, @babel/eslint-parser, @typescript-eslint/parser)
 * and append a `// features: [...], parser: ...` comment to `code`/`output`. With
 * the resolved toolchain (ESLint 10.x) the babel variant is always skipped
 * (`skipNewBabel ⊇ skipBabel = gte(ESLint.version, '10.0.0')` === true). No case in
 * this file carries a `features` array, so every case runs on default +
 * @typescript-eslint — identical under ts-go for a `.tsx` fixture. Each case is
 * therefore ported ONCE as its literal source (the appended parser-comment is
 * dropped); the multi-line plain-backtick templates — which are INDENTATION
 * SENSITIVE for this rule — are preserved byte-for-byte, including the leading
 * newline, the literal `\t` tab-indent fixtures, and the trailing whitespace.
 *
 * The expected message resolves from the plugin's own `meta.messages`:
 *   wrongIndent: "Expected indentation of {{needed}} {{type}} {{characters}} but found {{gotten}}."
 *
 * The upstream file contains NO `$` unindent template tags, NO `readFileSync`
 * external-fixture cases, NO spread/custom error helpers, NO `suggestions`, and
 * only the single `run()` block above (no skipBabel block). The `._css_` /
 * `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-indent-props', null as never, {
  valid: [
    {
      code: `
        <App foo
        />
      `,
    },
    {
      code: `
        <App
          foo
        />
      `,
      options: [2],
    },
    {
      code: `
        const Test = () => ([
          (x
            ? <div key="1" />
            : <div key="2" />),
          <div
            key="3"
            align="left"
          />,
          <div
            key="4"
            align="left"
          />,
        ]);
      `,
      options: [2],
    },
    {
      code: `
        <App
        foo
        />
      `,
      options: [0],
    },
    {
      code: `
          <App
        foo
          />
      `,
      options: [-2],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: ['tab'],
    },
    {
      code: `
        <App/>
      `,
      options: ['first'],
    },
    {
      code: `
        <App aaa
             b
             cc
        />
      `,
      options: ['first'],
    },
    {
      code: `
        <App   aaa
               b
               cc
        />
      `,
      options: ['first'],
    },
    {
      code: `
        const test = <App aaa
                          b
                          cc
                     />
      `,
      options: ['first'],
    },
    {
      code: `
        <App aaa x
             b y
             cc
        />
      `,
      options: ['first'],
    },
    {
      code: `
        const test = <App aaa x
                          b y
                          cc
                     />
      `,
      options: ['first'],
    },
    {
      code: `
        <App aaa
             b
        >
            <Child c
                   d/>
        </App>
      `,
      options: ['first'],
    },
    {
      code: `
        <Fragment>
          <App aaa
               b
               cc
          />
          <OtherApp a
                    bbb
                    c
          />
        </Fragment>
      `,
      options: ['first'],
    },
    {
      code: `
        <App
          a
          b
        />
      `,
      options: ['first'],
    },
    {
      code: `
        {this.props.ignoreTernaryOperatorFalse
          ? <span
              className="value"
              some={{aaa}}
            />
          : null}
      `,
      options: [
        {
          indentMode: 2,
          ignoreTernaryOperator: false,
        },
      ],
    },
    {
      code: `
        const F = () => {
          const foo = true
            ? <div id="id">test</div>
            : false;

          return <div
            id="id"
          >
            test
          </div>
        }
      `,
      options: [
        {
          indentMode: 2,
          ignoreTernaryOperator: false,
        },
      ],
    },
    {
      code: `
        const F = () => {
          const foo = true
            ? <div id="id">test</div>
            : false;

          return <div
            id="id"
          >
            test
          </div>
        }
      `,
      options: [
        {
          indentMode: 2,
          ignoreTernaryOperator: true,
        },
      ],
    },
    {
      code: `
\t\t\t\tconst F = () => {
\t\t\t\t\tconst foo = true
\t\t\t\t\t\t? <div id="id">test</div>
\t\t\t\t\t\t: false;

\t\t\t\t\treturn <div
\t\t\t\t\t\tid="id"
\t\t\t\t\t>
\t\t\t\t\t\ttest
\t\t\t\t\t</div>
\t\t\t\t}
`,
      options: [
        {
          indentMode: 'tab',
          ignoreTernaryOperator: false,
        },
      ],
    },
    {
      code: `
\t\t\t\tconst F = () => {
\t\t\t\t\tconst foo = true
\t\t\t\t\t\t? <div id="id">test</div>
\t\t\t\t\t\t: false;

\t\t\t\t\treturn <div
\t\t\t\t\t\tid="id"
\t\t\t\t\t>
\t\t\t\t\t\ttest
\t\t\t\t\t</div>
\t\t\t\t}
`,
      options: [
        {
          indentMode: 'tab',
          ignoreTernaryOperator: true,
        },
      ],
    },
    {
      code: `
        {this.props.ignoreTernaryOperatorTrue
          ? <span
            className="value"
            some={{aaa}}
            />
          : null}
      `,
      options: [
        {
          indentMode: 2,
          ignoreTernaryOperator: true,
        },
      ],
    },
    {
      code: `
        <a
          role={'button'}
          className={\`navbar-burger \${open ? 'is-active' : ''}\`}
          href={'#'}
          aria-label={'menu'}
          aria-expanded={false}
          onClick={openMenu}>
          <span aria-hidden={'true'}/>
          <span aria-hidden={'true'}/>
          <span aria-hidden={'true'}/>
        </a>
      `,
      options: [{ indentMode: 2 }],
    },
    {
      code: `
        <a role={'button'}
           className={\`navbar-burger \${open ? 'is-active' : ''}\`}
           href={'#'}
           aria-label={'menu'}
           aria-expanded={false}
           onClick={openMenu}>
          <span aria-hidden={'true'}/>
          <span aria-hidden={'true'}/>
          <span aria-hidden={'true'}/>
        </a>
      `,
      options: ['first'],
    },
  ],

  invalid: [
    {
      code: `
        <App
          foo
        />
      `,
      output: `
        <App
            foo
        />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 10,
          },
        },
      ],
    },
    {
      code: `
        <App
            foo
        />
      `,
      output: `
        <App
          foo
        />
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 10,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
      code: `
        const test = true
          ? <span
            attr="value"
            />
          : <span
            attr="otherValue"
            />
      `,
      output: `
        const test = true
          ? <span
              attr="value"
            />
          : <span
              attr="otherValue"
            />
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
      code: `
        const test = true
          ? <span attr="value" />
          : (
            <span
                attr="otherValue"
            />
          )
      `,
      output: `
        const test = true
          ? <span attr="value" />
          : (
            <span
              attr="otherValue"
            />
          )
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 16,
          },
        },
      ],
    },
    {
      code: `
        {test.isLoading
          ? <Value/>
          : <OtherValue
            some={aaa}/>
        }
      `,
      output: `
        {test.isLoading
          ? <Value/>
          : <OtherValue
              some={aaa}/>
        }
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
      code: `
        {test.isLoading
          ? <Value/>
          : <OtherValue
            some={aaa}
            other={bbb}/>
        }
      `,
      output: `
        {test.isLoading
          ? <Value/>
          : <OtherValue
              some={aaa}
              other={bbb}/>
        }
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
      code: `
        {this.props.test
          ? <span
            className="value"
            some={{aaa}}
            />
          : null}
      `,
      output: `
        {this.props.test
          ? <span
              className="value"
              some={{aaa}}
            />
          : null}
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
      code: `
        <App1
            foo
        />
      `,
      output: `
        <App1
\tfoo
        />
      `,
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 1,
            type: 'tab',
            characters: 'character',
            gotten: 0,
          },
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 5,
            type: 'tab',
            characters: 'characters',
            gotten: 7,
          },
        },
      ],
    },
    {
      code: `
        <App a
          b
        />
      `,
      output: `
        <App a
             b
        />
      `,
      options: ['first'],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 13,
            type: 'space',
            characters: 'characters',
            gotten: 10,
          },
        },
      ],
    },
    {
      code: `
        <App  a
           b
        />
      `,
      output: `
        <App  a
              b
        />
      `,
      options: ['first'],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 11,
          },
        },
      ],
    },
    {
      code: `
        <App
              a
           b
        />
      `,
      output: `
        <App
              a
              b
        />
      `,
      options: ['first'],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 14,
            type: 'space',
            characters: 'characters',
            gotten: 11,
          },
        },
      ],
    },
    {
      code: `
        <App
          a
         b
           c
        />
      `,
      output: `
        <App
          a
          b
          c
        />
      `,
      options: ['first'],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 10,
            type: 'space',
            characters: 'characters',
            gotten: 9,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 10,
            type: 'space',
            characters: 'characters',
            gotten: 11,
          },
        },
      ],
    },
    {
      code: `
        const F = () => {
          const foo = true
            ? <div id="id">test</div>
            : false;

          return <div
              id="id"
          >
            test
          </div>
        }
      `,
      output: `
        const F = () => {
          const foo = true
            ? <div id="id">test</div>
            : false;

          return <div
            id="id"
          >
            test
          </div>
        }
      `,
      options: [
        {
          indentMode: 2,
          ignoreTernaryOperator: false,
        },
      ],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 14,
          },
        },
      ],
    },
    {
      code: `
        const F = () => {
          const foo = true
            ? <div id="id">test</div>
            : false;

          return <div
              id="id"
          >
            test
          </div>
        }
      `,
      output: `
        const F = () => {
          const foo = true
            ? <div id="id">test</div>
            : false;

          return <div
            id="id"
          >
            test
          </div>
        }
      `,
      options: [
        {
          indentMode: 2,
          ignoreTernaryOperator: true,
        },
      ],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 14,
          },
        },
      ],
    },
    {
      code: `
\t\t\t\tconst F = () => {
\t\t\t\t\tconst foo = true
\t\t\t\t\t\t? <div id="id">test</div>
\t\t\t\t\t\t: false;

\t\t\t\t\treturn <div
\t\t\t\t\t\t\tid="id"
\t\t\t\t\t>
\t\t\t\t\t\ttest
\t\t\t\t\t</div>
\t\t\t\t}
`,
      output: `
\t\t\t\tconst F = () => {
\t\t\t\t\tconst foo = true
\t\t\t\t\t\t? <div id="id">test</div>
\t\t\t\t\t\t: false;

\t\t\t\t\treturn <div
\t\t\t\t\t\tid="id"
\t\t\t\t\t>
\t\t\t\t\t\ttest
\t\t\t\t\t</div>
\t\t\t\t}
`,
      options: [
        {
          indentMode: 'tab',
          ignoreTernaryOperator: false,
        },
      ],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 6,
            type: 'tab',
            characters: 'characters',
            gotten: 7,
          },
        },
      ],
    },
    {
      code: `
\t\t\t\tconst F = () => {
\t\t\t\t\tconst foo = true
\t\t\t\t\t\t? <div id="id">test</div>
\t\t\t\t\t\t: false;

\t\t\t\t\treturn <div
\t\t\t\t\t\t\tid="id"
\t\t\t\t\t>
\t\t\t\t\t\ttest
\t\t\t\t\t</div>
\t\t\t\t}
`,
      output: `
\t\t\t\tconst F = () => {
\t\t\t\t\tconst foo = true
\t\t\t\t\t\t? <div id="id">test</div>
\t\t\t\t\t\t: false;

\t\t\t\t\treturn <div
\t\t\t\t\t\tid="id"
\t\t\t\t\t>
\t\t\t\t\t\ttest
\t\t\t\t\t</div>
\t\t\t\t}
`,
      options: [
        {
          indentMode: 'tab',
          ignoreTernaryOperator: true,
        },
      ],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 6,
            type: 'tab',
            characters: 'characters',
            gotten: 7,
          },
        },
      ],
    },
  ],
});
