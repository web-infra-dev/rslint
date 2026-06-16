/**
 * @fileoverview Tests for jsx-indent rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-indent/jsx-indent.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid, invalid })`
 *    -> `ruleTester.run('jsx-indent', null as never, { valid, invalid })`
 *    (the `name`, `rule`, `parserOptions` keys are dropped).
 *  - Upstream wraps its arrays in `valids(...)` / `invalids(...)` (from
 *    `#test/parsers-jsx`). Those helpers MULTIPLY every case across three parsers
 *    (default ESLint, `@babel/eslint-parser`, `@typescript-eslint/parser`) and
 *    append a `// features: [...], parser: ...` comment to `code`/`output`. That
 *    is upstream test-harness machinery, NOT rule semantics — rslint runs the
 *    single ts-go path (equivalent to the `@typescript-eslint/parser` variant), so
 *    each upstream entry is ported as ONE case with NO appended parser comment.
 *  - The per-case `features: [...]` property is dropped: `['fragment']` /
 *    `['fragment', 'no-ts-old']` mark `<>` fragments (valid TSX — they ARE run by
 *    upstream's new `@typescript-eslint/parser` variant; `no-ts-old` only skips the
 *    OLD ts parser); `['class fields']` (class field syntax) is likewise valid
 *    modern TS. These stay in the main blocks. `['do expressions']` and `['flow']`
 *    are NOT valid TypeScript — they are isolated into KNOWN GAPS (see bottom).
 *  - `parserOptions.ecmaFeatures.jsx` dropped — rslint resolves JSX via tsconfig;
 *    the RuleTester routes JSX code to a `.tsx` fixture.
 *  - The single messageId `wrongIndent` carries `data: { needed, type, characters,
 *    gotten }`; its message
 *    'Expected indentation of {{needed}} {{type}} {{characters}} but found {{gotten}}.'
 *    is fully resolved from that data and asserted. Cases that pin only
 *    `{ messageId, line }` (no data) keep their count+line assertion; the message
 *    template stays un-interpolated so it is not asserted (RuleTester semantics).
 *  - All `code`/`output` are plain backtick template literals whose indentation and
 *    `\n` / `\t` escapes are part of the literal string value — preserved
 *    byte-for-byte (jsx-indent is indentation-sensitive).
 *
 * No `$`-unindent, suggestion, or external-fixture (`readFileSync`) cases exist in
 * the upstream jsx-indent test. The `._css_` / `._json_` / `._markdown_` test files
 * don't exist for this rule.
 *
 * Counts after porting: 95 valid + 57 invalid live cases. 14 do-expression + 1 flow
 * cases isolated as unparseable, and 4 invalid cases isolated for a multi-pass fix
 * divergence — all in KNOWN GAPS below.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-indent', null as never, {
  valid: [
    {
      code: `
        <App></App>
      `,
    },
    {
      code: `
        <></>
      `,
    },
    {
      code: `
        <App>
        </App>
      `,
    },
    {
      code: `
        <>
        </>
      `,
    },
    {
      code: `
        <App>
          <Foo />
        </App>
      `,
      options: [2],
    },
    {
      code: `
        <App>
          <></>
        </App>
      `,
      options: [2],
    },
    {
      code: `
        <>
          <Foo />
        </>
      `,
      options: [2],
    },
    {
      code: `
        <App>
        <Foo />
        </App>
      `,
      options: [0],
    },
    {
      code: `
          <App>
        <Foo />
          </App>
      `,
      options: [-2],
    },
    {
      code: `
\t\t\t\t<App>
\t\t\t\t\t<Foo />
\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
    },
    {
      code: `
        function App() {
          return <App>
            <Foo />
          </App>;
        }
      `,
      options: [2],
    },
    {
      code: `
        function App() {
          return <App>
            <></>
          </App>;
        }
      `,
      options: [2],
    },
    {
      code: `
        function App() {
          return (<App>
            <Foo />
          </App>);
        }
      `,
      options: [2],
    },
    {
      code: `
        function App() {
          return (<App>
            <></>
          </App>);
        }
      `,
      options: [2],
    },
    {
      code: `
        function App() {
          return (
            <App>
              <Foo />
            </App>
          );
        }
      `,
      options: [2],
    },
    {
      code: `
        function App() {
          return (
            <App>
              <></>
            </App>
          );
        }
      `,
      options: [2],
    },
    {
      code: `
        it(
          (
            <div>
              <span />
            </div>
          )
        )
      `,
      options: [2],
    },
    {
      code: `
        it(
          (
            <div>
              <></>
            </div>
          )
        )
      `,
      options: [2],
    },
    {
      code: `
        it(
          (<div>
            <span />
            <span />
            <span />
          </div>)
        )
      `,
      options: [2],
    },
    {
      code: `
        (
          <div>
            <span />
          </div>
        )
      `,
      options: [2],
    },
    {
      code: `
        {
          head.title &&
          <h1>
            {head.title}
          </h1>
        }
      `,
      options: [2],
    },
    {
      code: `
        {
          head.title &&
          <>
            {head.title}
          </>
        }
      `,
      options: [2],
    },
    {
      code: `
        {
          head.title &&
            <h1>
              {head.title}
            </h1>
        }
      `,
      options: [2],
    },
    {
      code: `
        {
          head.title && (
          <h1>
            {head.title}
          </h1>)
        }
      `,
      options: [2],
    },
    {
      code: `
        {
          head.title && (
            <h1>
              {head.title}
            </h1>
          )
        }
      `,
      options: [2],
    },
    {
      code: `
        [
          <div />,
          <div />
        ]
      `,
      options: [2],
    },
    {
      code: `
        [
          <></>,
          <></>
        ]
      `,
      options: [2],
    },
    {
      code: `
        <div>
            {
                [
                    <Foo />,
                    <Bar />
                ]
            }
        </div>
      `,
    },
    {
      code: `
        <div>
            {foo &&
                [
                    <Foo />,
                    <Bar />
                ]
            }
        </div>
      `,
    },
    {
      code: `
        <div>
            {foo &&
                [
                    <></>,
                    <></>
                ]
            }
        </div>
      `,
    },
    {
      code: `
        <div>
            bar <div>
                bar
                bar {foo}
                bar </div>
        </div>
      `,
    },
    {
      code: `
        <>
            bar <>
                bar
                bar {foo}
                bar </>
        </>
      `,
    },
    {
    // Multiline ternary
    // (colon at the end of the first expression)
      code: `
        foo ?
            <Foo /> :
            <Bar />
      `,
    },
    {
      code: `
        foo ?
            <></> :
            <></>
      `,
    },
    {
    // Multiline ternary
    // (colon at the start of the second expression)
      code: `
        foo ?
            <Foo />
            : <Bar />
      `,
    },
    {
      code: `
        foo ?
            <></>
            : <></>
      `,
    },
    {
    // Multiline ternary
    // (colon on its own line)
      code: `
        foo ?
            <Foo />
        :
            <Bar />
      `,
    },
    {
      code: `
        foo ?
            <></>
        :
            <></>
      `,
    },
    {
    // Multiline ternary
    // (multiline JSX, colon on its own line)
      code: `
        {!foo ?
            <Foo
                onClick={this.onClick}
            />
        :
            <Bar
                onClick={this.onClick}
            />
        }
      `,
    },
    {
    // Multiline ternary
    // (first expression on test line, colon at the end of the first expression)
      code: `
        foo ? <Foo /> :
        <Bar />
      `,
    },
    {
      code: `
        foo ? <></> :
        <></>
      `,
    },
    {
    // Multiline ternary
    // (first expression on test line, colon at the start of the second expression)
      code: `
        foo ? <Foo />
        : <Bar />
      `,
    },
    {
      code: `
        foo ? <></>
        : <></>
      `,
    },
    {
    // Multiline ternary
    // (first expression on test line, colon on its own line)
      code: `
        foo ? <Foo />
        :
        <Bar />
      `,
    },
    {
      code: `
        foo ? <></>
        :
        <></>
      `,
    },
    {
    // Multiline ternary
    // (colon at the end of the first expression, parenthesized first expression)
      code: `
        foo ? (
            <Foo />
        ) :
            <Bar />
      `,
    },
    {
      code: `
        foo ? (
            <></>
        ) :
            <></>
      `,
    },
    {
    // Multiline ternary
    // (colon at the start of the second expression, parenthesized first expression)
      code: `
        foo ? (
            <Foo />
        )
            : <Bar />
      `,
    },
    {
      code: `
        foo ? (
            <></>
        )
            : <></>
      `,
    },
    {
    // Multiline ternary
    // (colon on its own line, parenthesized first expression)
      code: `
        foo ? (
            <Foo />
        )
        :
            <Bar />
      `,
    },
    {
      code: `
        foo ? (
            <></>
        )
        :
            <></>
      `,
    },
    {
    // Multiline ternary
    // (colon at the end of the first expression, parenthesized second expression)
      code: `
        foo ?
            <Foo /> : (
                <Bar />
            )
      `,
    },
    {
      code: `
        foo ?
            <></> : (
                <></>
            )
      `,
    },
    {
    // Multiline ternary
    // (colon on its own line, parenthesized second expression)
      code: `
        foo ?
            <Foo />
        : (
            <Bar />
        )
      `,
    },
    {
      code: `
        foo ?
            <></>
        : (
            <></>
        )
      `,
    },
    {
    // Multiline ternary
    // (colon indented on its own line, parenthesized second expression)
      code: `
        foo ?
            <Foo />
            : (
                <Bar />
            )
      `,
    },
    {
      code: `
        foo ?
            <></>
            : (
                <></>
            )
      `,
    },
    {
    // Multiline ternary
    // (colon at the end of the first expression, both expression parenthesized)
      code: `
        foo ? (
            <Foo />
        ) : (
            <Bar />
        )
      `,
    },
    {
      code: `
        foo ? (
            <></>
        ) : (
            <></>
        )
      `,
    },
    {
    // Multiline ternary
    // (colon on its own line, both expression parenthesized)
      code: `
        foo ? (
            <Foo />
        )
        : (
            <Bar />
        )
      `,
    },
    {
      code: `
        foo ? (
            <></>
        )
        : (
            <></>
        )
      `,
    },
    {
    // Multiline ternary
    // (colon on its own line, both expression parenthesized)
      code: `
        foo ? (
            <Foo />
        )
        :
        (
            <Bar />
        )
      `,
    },
    {
      code: `
        foo ? (
            <></>
        )
        :
        (
            <></>
        )
      `,
    },
    {
    // Multiline ternary
    // (first expression on test line, colon at the end of the first expression, parenthesized second expression)
      code: `
        foo ? <Foo /> : (
            <Bar />
        )
      `,
    },
    {
      code: `
        foo ? <></> : (
            <></>
        )
      `,
    },
    {
    // Multiline ternary
    // (first expression on test line, colon at the start of the second expression, parenthesized second expression)
      code: `
        foo ? <Foo />
        : (<Bar />)
      `,
    },
    {
      code: `
        foo ? <></>
        : (<></>)
      `,
    },
    {
    // Multiline ternary
    // (first expression on test line, colon on its own line, parenthesized second expression)
      code: `
        foo ? <Foo />
        : (
            <Bar />
        )
      `,
    },
    {
      code: `
        foo ? <></>
        : (
            <></>
        )
      `,
    },
    {
      code: `
        <span>
          {condition ?
            <Thing
              foo={\`bar\`}
            /> :
            <Thing/>
          }
        </span>
      `,
      options: [2],
    },
    {
      code: `
        <span>
          {condition ?
            <Thing
              foo={"bar"}
            /> :
            <Thing/>
          }
        </span>
      `,
      options: [2],
    },
    {
      code: `
        function foo() {
          <span>
            {condition ?
              <Thing
                foo={superFoo}
              /> :
              <Thing/>
            }
          </span>
        }
      `,
      options: [2],
    },
    {
      code: `
        function foo() {
          <span>
            {condition ?
              <Thing
                foo={superFoo}
              /> :
              <></>
            }
          </span>
        }
      `,
      options: [2],
    },
    {
      code: `
        class Test extends React.Component {
          render() {
            return (
              <div>
                <div />
                <div />
              </div>
            );
          }
        }
      `,
      options: [2],
    },
    {
      code: `
        class Test extends React.Component {
          render() {
            return (
              <>
                <></>
                <></>
              </>
            );
          }
        }
      `,
      options: [2],
    },
    {
      code: `
        const Component = () => (
          <View
            ListFooterComponent={(
              <View
                rowSpan={3}
                placeholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"
              />
        )}
          />
        );
      `,
      options: [2],
    },
    {
      code: `
const Component = () => (
\t<View
\t\tListFooterComponent={(
\t\t\t<View
\t\t\t\trowSpan={3}
\t\t\t\tplaceholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"
\t\t\t/>
)}
\t/>
);
    `,
      options: ['tab'],
    },
    {
      code: `
        const Component = () => (
          <View
            ListFooterComponent={(
              <View
                rowSpan={3}
                placeholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"
              />
        )}
          />
        );
      `,
      options: [2, { checkAttributes: false }],
    },
    {
      code: `
const Component = () => (
\t<View
\t\tListFooterComponent={(
\t\t\t<View
\t\t\t\trowSpan={3}
\t\t\t\tplaceholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"
\t\t\t/>
)}
\t/>
);
    `,
      options: ['tab', { checkAttributes: false }],
    },
    {
      code: `
        function Foo() {
          return (
            <input
              type="radio"
              defaultChecked
            />
          );
        }
      `,
      options: [2, { checkAttributes: true }],
    },
    {
      code: `
        function Foo() {
          return (
            <div>
              {condition && (
                <p>Bar</p>
              )}
            </div>
          );
        }
      `,
      options: [2, { indentLogicalExpressions: true }],
    },
    {
      code: `
        <App>
            text
        </App>
      `,
    },
    {
      code: `
        <App>
            text
            text
            text
        </App>
      `,
    },
    {
      code: `
\t\t\t\t<App>
\t\t\t\t\ttext
\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
    },
    {
      code: `
\t\t\t\t<App>
\t\t\t\t\t{undefined}
\t\t\t\t\t{null}
\t\t\t\t\t{true}
\t\t\t\t\t{false}
\t\t\t\t\t{42}
\t\t\t\t\t{NaN}
\t\t\t\t\t{"foo"}
\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
    },
    {
    // don't check literals not within JSX. See #2563
      code: `
        function foo() {
          const a = \`aa\`;
          const b = \`b\nb\`;
        }
      `,
    },
    {
      code: `
        function App() {
          return (
            <App />
          );
        }
      `,
      options: [2],
    },
    {
      code: `
        function App() {
          return <App>
            <Foo />
          </App>;
        }
      `,
      options: [2],
    },
    {
      code: `
        const myFunction = () => (
          [
            <Tag
              {...properties}
            />,
            <Tag
              {...properties}
            />,
            <Tag
              {...properties}
            />,
          ]
        )
      `,
      options: [2],
    },
    {
      code: `
        const Item = ({ id, name, onSelect }) => <div onClick={onSelect}>
          {id}: {name}
        </div>;
      `,
      options: [2],
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
      options: [2],
    },
    {
      code: `
        export default class App extends React.Component {
          state = {
            name: '',
          }

          componentDidMount() {
            this.fetchName()
              .then(name => {
                this.setState({name})
              });
          }

          fetchName = () => {
            const url = 'https://api.github.com/users/job13er'
            return fetch(url)
              .then(resp => resp.json())
              .then(json => json.name)
          }

          render() {
            const {name} = this.state
            return (
              <h1>Hello, {name}</h1>
            )
          }
        }
      `,
      options: [2],
    },
    {
      code: `
        function test (foo) {
          return foo != null
            ? Math.max(0, Math.min(1, 10))
            : 0
        }
      `,
      options: [99],
    },
    {
      code: `
        function test (foo) {
          return foo != null
            ? <div>foo</div>
            : <div>bar</div>
        }
      `,
      options: [2],
    },
    {
      options: [2, { checkAttributes: true, indentLogicalExpressions: true }],
      code: `
      <>
        <div
          foo={
            condition
              ? [
                'bar'
              ]
              : [
                'baz',
                'qux'
              ]
          }
        />
        <div
          style={
            true
              ? {
                  color: 'red',
                }
              : {
                  height: 1,
                }
          }
        />
      </>
      `,
    },
  ],
  invalid: [
    {
      code: `
        <App>
          <Foo />
        </App>
      `,
      output: `
        <App>
            <Foo />
        </App>
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
        <App>
          <></>
        </App>
      `,
      output: `
        <App>
            <></>
        </App>
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
        <>
          <Foo />
        </>
      `,
      output: `
        <>
            <Foo />
        </>
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
        <App>
            <Foo />
        </App>
      `,
      output: `
        <App>
          <Foo />
        </App>
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
        <App>
            <Foo />
        </App>
      `,
      output: `
        <App>
\t<Foo />
        </App>
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
        function App() {
          return <App>
            <Foo />
                 </App>;
        }
      `,
      output: `
        function App() {
          return <App>
            <Foo />
          </App>;
        }
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          line: 3,
          data: {
            needed: 10,
            type: 'space',
            characters: 'characters',
            gotten: 17,
          },
        },
        {
          messageId: 'wrongIndent',
          line: 5,
          data: {
            needed: 10,
            type: 'space',
            characters: 'characters',
            gotten: 17,
          },
        },
      ],
    },
    {
      code: `
        function App() {
          return (<App>
            <Foo />
            </App>);
        }
      `,
      output: `
        function App() {
          return (<App>
            <Foo />
          </App>);
        }
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          line: 3,
          data: {
            needed: 10,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
        {
          messageId: 'wrongIndent',
          line: 5,
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
        <App>
           {test}
        </App>
      `,
      output: `
        <App>
            {test}
        </App>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 11,
          },
        },
      ],
    },
    {
      code: `
        <App>
            {options.map((option, index) => (
                <option key={index} value={option.key}>
                   {option.name}
                </option>
            ))}
        </App>
      `,
      output: `
        <App>
            {options.map((option, index) => (
                <option key={index} value={option.key}>
                    {option.name}
                </option>
            ))}
        </App>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 20,
            type: 'space',
            characters: 'characters',
            gotten: 19,
          },
        },
      ],
    },
    {
      code: `
        <App>
        {test}
        </App>
      `,
      output: `
        <App>
\t{test}
        </App>
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
\t\t\t\t<App>
\t\t\t\t\t{options.map((option, index) => (
\t\t\t\t\t\t<option key={index} value={option.key}>
\t\t\t\t\t\t{option.name}
\t\t\t\t\t\t</option>
\t\t\t\t\t))}
\t\t\t\t</App>
\t\t\t`,
      output: `
\t\t\t\t<App>
\t\t\t\t\t{options.map((option, index) => (
\t\t\t\t\t\t<option key={index} value={option.key}>
\t\t\t\t\t\t\t{option.name}
\t\t\t\t\t\t</option>
\t\t\t\t\t))}
\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 7,
            type: 'tab',
            characters: 'characters',
            gotten: 6,
          },
        },
      ],
    },
    {
      code: `
\t\t\t\t<App>\n
\t\t\t\t<Foo />\n
\t\t\t\t</App>
\t\t\t`,
      output: `
\t\t\t\t<App>\n
\t\t\t\t\t<Foo />\n
\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 5,
            type: 'tab',
            characters: 'characters',
            gotten: 4,
          },
        },
      ],
    },
    {
      code: `
        [
          <div />,
            <div />
        ]
      `,
      output: `
        [
          <div />,
          <div />
        ]
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
        [
          <div />,
            <></>
        ]
      `,
      output: `
        [
          <div />,
          <></>
        ]
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
        <App>

         <Foo />

        </App>
      `,
      output: `
        <App>

\t<Foo />

        </App>
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
        <App>

        \t<Foo />

        </App>
      `,
      output: `
        <App>

          <Foo />

        </App>
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 10,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        <div>
            {
                [
                    <Foo />,
                <Bar />
                ]
            }
        </div>
      `,
      output: `
        <div>
            {
                [
                    <Foo />,
                    <Bar />
                ]
            }
        </div>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 20,
            type: 'space',
            characters: 'characters',
            gotten: 16,
          },
        },
      ],
    },
    {
      code: `
        <div>
            {foo &&
                [
                    <Foo />,
                <Bar />
                ]
            }
        </div>
      `,
      output: `
        <div>
            {foo &&
                [
                    <Foo />,
                    <Bar />
                ]
            }
        </div>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 20,
            type: 'space',
            characters: 'characters',
            gotten: 16,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon at the end of the first expression)
      code: `
        foo ?
            <Foo /> :
        <Bar />
      `,
      output: `
        foo ?
            <Foo /> :
            <Bar />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        foo ?
            <Foo /> :
        <></>
      `,
      output: `
        foo ?
            <Foo /> :
            <></>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon on its own line)
      code: `
        foo ?
            <Foo />
        :
        <Bar />
      `,
      output: `
        foo ?
            <Foo />
        :
            <Bar />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (first expression on test line, colon at the end of the first expression)
      code: `
        foo ? <Foo /> :
            <Bar />
      `,
      output: `
        foo ? <Foo /> :
        <Bar />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 8,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
      code: `
        foo ?
            <Foo />
        :
        <></>
      `,
      output: `
        foo ?
            <Foo />
        :
            <></>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (first expression on test line, colon on its own line)
      code: `
        foo ? <Foo />
        :
              <Bar />
      `,
      output: `
        foo ? <Foo />
        :
        <Bar />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 8,
            type: 'space',
            characters: 'characters',
            gotten: 14,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon at the end of the first expression, parenthesized first expression)
      code: `
        foo ? (
            <Foo />
        ) :
        <Bar />
      `,
      output: `
        foo ? (
            <Foo />
        ) :
            <Bar />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        foo ? (
            <Foo />
        ) :
        <></>
      `,
      output: `
        foo ? (
            <Foo />
        ) :
            <></>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon on its own line, parenthesized first expression)
      code: `
        foo ? (
            <Foo />
        )
        :
        <Bar />
      `,
      output: `
        foo ? (
            <Foo />
        )
        :
            <Bar />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon at the end of the first expression, parenthesized second expression)
      code: `
        foo ?
            <Foo /> : (
            <Bar />
            )
      `,
      output: `
        foo ?
            <Foo /> : (
                <Bar />
            )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 16,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
      code: `
        foo ?
            <Foo /> : (
            <></>
            )
      `,
      output: `
        foo ?
            <Foo /> : (
                <></>
            )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 16,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon on its own line, parenthesized second expression)
      code: `
        foo ?
            <Foo />
        : (
        <Bar />
        )
      `,
      output: `
        foo ?
            <Foo />
        : (
            <Bar />
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon indented on its own line, parenthesized second expression)
      code: `
        foo ?
            <Foo />
            : (
            <Bar />
            )
      `,
      output: `
        foo ?
            <Foo />
            : (
                <Bar />
            )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 16,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
      code: `
        foo ?
            <Foo />
            : (
            <></>
            )
      `,
      output: `
        foo ?
            <Foo />
            : (
                <></>
            )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 16,
            type: 'space',
            characters: 'characters',
            gotten: 12,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon at the end of the first expression, both expression parenthesized)
      code: `
        foo ? (
        <Foo />
        ) : (
        <Bar />
        )
      `,
      output: `
        foo ? (
            <Foo />
        ) : (
            <Bar />
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        foo ? (
        <></>
        ) : (
        <></>
        )
      `,
      output: `
        foo ? (
            <></>
        ) : (
            <></>
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon on its own line, both expression parenthesized)
      code: `
        foo ? (
        <Foo />
        )
        : (
        <Bar />
        )
      `,
      output: `
        foo ? (
            <Foo />
        )
        : (
            <Bar />
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (colon on its own line, both expression parenthesized)
      code: `
        foo ? (
        <Foo />
        )
        :
        (
        <Bar />
        )
      `,
      output: `
        foo ? (
            <Foo />
        )
        :
        (
            <Bar />
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        foo ? (
        <></>
        )
        :
        (
        <></>
        )
      `,
      output: `
        foo ? (
            <></>
        )
        :
        (
            <></>
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
    // Multiline ternary
    // (first expression on test line, colon at the end of the first expression, parenthesized second expression)
      code: `
        foo ? <Foo /> : (
        <Bar />
        )
      `,
      output: `
        foo ? <Foo /> : (
            <Bar />
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        foo ? <Foo /> : (
        <></>
        )
      `,
      output: `
        foo ? <Foo /> : (
            <></>
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      // Multiline ternary
      // (first expression on test line, colon on its own line, parenthesized second expression)
      code: `
        foo ? <Foo />
        : (
        <Bar />
        )
      `,
      output: `
        foo ? <Foo />
        : (
            <Bar />
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        foo ? <Foo />
        : (
        <></>
        )
      `,
      output: `
        foo ? <Foo />
        : (
            <></>
        )
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        <p>
            <div>
                <SelfClosingTag />Text
          </div>
        </p>
      `,
      output: `
        <p>
            <div>
                <SelfClosingTag />Text
            </div>
        </p>
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
        const Component = () => (
          <View
            ListFooterComponent={(
              <View
                rowSpan={3}
                placeholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"
              />
        )}
          />
        );
      `,
      output: `
        const Component = () => (
          <View
            ListFooterComponent={(
              <View
                rowSpan={3}
                placeholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"
              />
            )}
          />
        );
      `,
      options: [2, { checkAttributes: true }],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
const Component = () => (
\t<View
\t\tListFooterComponent={(
\t\t\t<View
\t\t\t\trowSpan={3}
\t\t\t\tplaceholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"
\t\t\t/>
)}
\t/>
);
    `,
      output: `
const Component = () => (
\t<View
\t\tListFooterComponent={(
\t\t\t<View
\t\t\t\trowSpan={3}
\t\t\t\tplaceholder="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do"
\t\t\t/>
\t\t)}
\t/>
);
    `,
      options: ['tab', { checkAttributes: true }],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 2,
            type: 'tab',
            characters: 'characters',
            gotten: 0,
          },
        },
      ],
    },
    {
      code: `
        function Foo() {
          return (
            <div>
              {condition && (
              <p>Bar</p>
              )}
            </div>
          );
        }
      `,
      output: `
        function Foo() {
          return (
            <div>
              {condition && (
                <p>Bar</p>
              )}
            </div>
          );
        }
      `,
      options: [2, { indentLogicalExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 16,
            type: 'space',
            characters: 'characters',
            gotten: 14,
          },
        },
      ],
    },
    {
      code: `
        <div>
        text
        </div>
      `,
      output: `
        <div>
            text
        </div>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        <div>
          text
        text
        </div>
      `,
      output: `
        <div>
            text
            text
        </div>
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
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        <div>
        \t  text
          \t  text
        </div>
      `,
      output: `
        <div>
            text
            text
        </div>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
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
        <div>
        \t  text
          \t  text
        </div>
      `,
      output: `
        <div>
            text
            text
        </div>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
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
        <div>
        \t\ttext
        </div>
      `,
      options: ['tab'],
      output: `
        <div>
\ttext
        </div>
      `,
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
        <>
        aaa
        </>
      `,
      output: `
        <>
            aaa
        </>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        const StatelessComponent = () => {
          if (new Date() % 2) {
              return (
        <div>Hello</div>
              );
          }
          return null;
        };
      `,
      output: `
        const StatelessComponent = () => {
          if (new Date() % 2) {
              return (
                  <div>Hello</div>
              );
          }
          return null;
        };
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 18,
            gotten: 8,
            type: 'space',
            characters: 'characters',
          },
        },
      ],
    },
    {
      code: `
        function App() {
          return (
            <App />
            );
        }
      `,
      output: `
        function App() {
          return (
            <App />
          );
        }
      `,
      options: [2],
      errors: [{ message: 'Expected indentation of 10 space characters but found 12.' }],
    },
    {
      code: `
        function App() {
          return (
            <App />
        );
        }
      `,
      output: `
        function App() {
          return (
            <App />
          );
        }
      `,
      options: [2],
      errors: [{ message: 'Expected indentation of 10 space characters but found 8.' }],
    },
    {
      code: `
        {condition && [
            <Tag key="a" onClick={() => {
              // some code
            }} />,
            <Tag key="b" onClick={() => {
              // some code
            }} />,
          ]
        }
      `,
      output: `
        {condition && [
          <Tag key="a" onClick={() => {
              // some code
            }} />,
          <Tag key="b" onClick={() => {
              // some code
            }} />,
          ]
        }
      `,
      options: [2],
      errors: [
        {
          message: 'Expected indentation of 10 space characters but found 12.',
          line: 3,
        },
        {
          message: 'Expected indentation of 10 space characters but found 12.',
          line: 6,
        },
      ],
    },
    {
      code: `
        const IndexPage = () => (
          <h1>
        {"Hi people"}
        <button/>
        </h1>
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
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
        {
          messageId: 'wrongIndent',
          data: {
            needed: 10,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
    {
      code: `
        const IndexPage = () => (
          <h1>
            Hi people
        <button/>
          </h1>
        );
      `,

      output: `
        const IndexPage = () => (
          <h1>
            Hi people
            <button/>
          </h1>
        );
      `,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          data: {
            needed: 12,
            type: 'space',
            characters: 'characters',
            gotten: 8,
          },
        },
      ],
    },
  ],
});

/*
 * ============================ jsx-indent — KNOWN GAPS ============================
 *
 * The cases below are NOT ported into the live blocks above. They are preserved
 * verbatim (commented out, with the upstream `features` flag / comment where
 * present), never deleted and never altered to pass. Each was verified against
 * rslint directly.
 *
 *
 * ---------------------------------------------------------------------------------
 * (1) `do { ... }` do-expressions (features: ['do expressions']) — UNPARSEABLE
 * ---------------------------------------------------------------------------------
 * Babel-only Stage-1 syntax. ts-go emits TS1109 "Expression expected" at the `do`
 * keyword (and TS1381 around the closing `}}`). Upstream runs these only on the
 * default + `@babel/eslint-parser` variants, never on `@typescript-eslint/parser`.
 * 10 valid + 4 invalid cases:
 *
 * {
 *       code: `
 *         <span>
 *             {do {
 *                 const num = rollDice();
 *                 <Thing num={num} />;
 *             }}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {(do {
 *                 const num = rollDice();
 *                 <Thing num={num} />;
 *             })}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {do {
 *                 const purposeOfLife = getPurposeOfLife();
 *                 if (purposeOfLife == 42) {
 *                     <Thing />;
 *                 } else {
 *                     <AnotherThing />;
 *                 }
 *             }}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {(do {
 *                 const purposeOfLife = getPurposeOfLife();
 *                 if (purposeOfLife == 42) {
 *                     <Thing />;
 *                 } else {
 *                     <AnotherThing />;
 *                 }
 *             })}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {do {
 *                 <Thing num={rollDice()} />;
 *             }}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {(do {
 *                 <Thing num={rollDice()} />;
 *             })}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {do {
 *                 <Thing num={rollDice()} />;
 *                 <Thing num={rollDice()} />;
 *             }}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {(do {
 *                 <Thing num={rollDice()} />;
 *                 <Thing num={rollDice()} />;
 *             })}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {do {
 *                 const purposeOfLife = 42;
 *                 <Thing num={purposeOfLife} />;
 *                 <Thing num={purposeOfLife} />;
 *             }}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {(do {
 *                 const purposeOfLife = 42;
 *                 <Thing num={purposeOfLife} />;
 *                 <Thing num={purposeOfLife} />;
 *             })}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {do {
 *                 const num = rollDice();
 *                     <Thing num={num} />;
 *             }}
 *         </span>
 *       `,
 *       output: `
 *         <span>
 *             {do {
 *                 const num = rollDice();
 *                 <Thing num={num} />;
 *             }}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *       errors: [
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 16,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 20,
 *           },
 *         },
 *       ],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {(do {
 *                 const num = rollDice();
 *                     <Thing num={num} />;
 *             })}
 *         </span>
 *       `,
 *       output: `
 *         <span>
 *             {(do {
 *                 const num = rollDice();
 *                 <Thing num={num} />;
 *             })}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *       errors: [
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 16,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 20,
 *           },
 *         },
 *       ],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {do {
 *             <Thing num={getPurposeOfLife()} />;
 *             }}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *       output: `
 *         <span>
 *             {do {
 *                 <Thing num={getPurposeOfLife()} />;
 *             }}
 *         </span>
 *       `,
 *       errors: [
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 16,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 12,
 *           },
 *         },
 *       ],
 *     }
 *
 * {
 *       code: `
 *         <span>
 *             {(do {
 *             <Thing num={getPurposeOfLife()} />;
 *             })}
 *         </span>
 *       `,
 *       output: `
 *         <span>
 *             {(do {
 *                 <Thing num={getPurposeOfLife()} />;
 *             })}
 *         </span>
 *       `,
 *       features: ['do expressions'],
 *       errors: [
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 16,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 12,
 *           },
 *         },
 *       ],
 *     }
 *
 *
 * ---------------------------------------------------------------------------------
 * (2) Flow type-annotated arrow with comment-only body (features: ['flow']) — UNPARSEABLE
 * ---------------------------------------------------------------------------------
 * Upstream runs this only on `@babel/eslint-parser` (skipBase + skipTS exclude the
 * TS parser for `flow`). The arrow body `( // JSX )` is an empty parenthesized
 * expression (only a comment, no expression) — ts-go emits TS1109 "Expression
 * expected". 1 valid case:
 *
 * {
 *       code: `
 *         type Props = {
 *           email: string,
 *           password: string,
 *           error: string,
 *         }
 *
 *         const SomeFormComponent = ({
 *           email,
 *           password,
 *           error,
 *         }: Props) => (
 *           // JSX
 *         );
 *       `,
 *       features: ['flow'],
 *     }
 *
 *
 * ---------------------------------------------------------------------------------
 * (3) Multi-pass fix divergence — 4 invalid cases
 * ---------------------------------------------------------------------------------
 * For these four invalid cases the DIAGNOSTIC set matches upstream exactly (same
 * count, same messages), but the autofix OUTPUT differs. ESLint's RuleTester applies
 * a SINGLE fix pass (`verifyAfterFix: false` upstream), so its pinned `output` is a
 * deliberately PARTIAL re-indent — every upstream `output` here carries a comment
 * admitting this (`See #608`, `multipass works fine`, `TODO: remove two spaces`).
 * rslint applies fixes to a stable fixed point (multi-pass), so it re-indents the
 * still-wrong descendant lines too, producing a MORE complete result. This is the
 * documented single-pass-vs-multi-pass gap (see rule-tester.ts) — a fixer-iteration
 * difference, NOT a detection difference. Preserved verbatim with each upstream
 * single-pass `output` and the rslint multi-pass `output` noted:
 *
 *
 * --- multi-pass case (was invalid[0] in upstream order) ---
 * {
 *       code: `
 *         <div>
 *         bar <div>
 *            bar
 *            bar {foo}
 *            bar </div>
 *         </div>
 *       `,
 *       output: `
 *         <div>
 *             bar <div>
 *             bar
 *             bar {foo}
 *             bar </div>
 *         </div>
 *       `,
 *       errors: [
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 12,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 8,
 *           },
 *         },
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 12,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 11,
 *           },
 *         },
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 12,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 11,
 *           },
 *         },
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 12,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 11,
 *           },
 *         },
 *       ],
 *     }
 *
 * rslint multi-pass output (differs from the single-pass `output` above):
 * `
 *
 *         <div>
 *             bar <div>
 *                 bar
 *                 bar {foo}
 *                 bar </div>
 *         </div>
 *       
 * `
 *
 *
 * --- multi-pass case (was invalid[8] in upstream order) ---
 * {
 *       code: `
 *         function App() {
 *           return (
 *         <App>
 *           <Foo />
 *         </App>
 *           );
 *         }
 *       `,
 *       // The detection logic only thinks <App> is indented wrong, not the other
 *       // two lines following. I *think* because it incorrectly uses <App>'s indention
 *       // as the baseline for the next two, instead of the realizing the entire three
 *       // lines are wrong together. See #608
 *       output: `
 *         function App() {
 *           return (
 *             <App>
 *           <Foo />
 *         </App>
 *           );
 *         }
 *       `,
 *       options: [2],
 *       errors: [
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 12,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 8,
 *           },
 *         },
 *       ],
 *     }
 *
 * rslint multi-pass output (differs from the single-pass `output` above):
 * `
 *
 *         function App() {
 *           return (
 *             <App>
 *               <Foo />
 *             </App>
 *           );
 *         }
 *       
 * `
 *
 *
 * --- multi-pass case (was invalid[58] in upstream order) ---
 * {
 *       code: `
 *         const IndexPage = () => (
 *           <h1>
 *         Hi people
 *         <button/>
 *         </h1>
 *         );
 *       `,
 *
 *       output: `
 *         const IndexPage = () => (
 *           <h1>
 *             Hi people
 *         <button/>
 *           </h1>
 *         );
 *       `,
 *       options: [2],
 *       errors: [
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 12,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 8,
 *           },
 *         },
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 12,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 8,
 *           },
 *         },
 *         {
 *           messageId: 'wrongIndent',
 *           data: {
 *             needed: 10,
 *             type: 'space',
 *             characters: 'characters',
 *             gotten: 8,
 *           },
 *         },
 *       ],
 *     }
 *
 * rslint multi-pass output (differs from the single-pass `output` above):
 * `
 *
 *         const IndexPage = () => (
 *           <h1>
 *             Hi people
 *             <button/>
 *           </h1>
 *         );
 *       
 * `
 *
 *
 * --- multi-pass case (was invalid[60] in upstream order) ---
 * {
 *       code: `
 *         import React from 'react';
 *
 *         export default function () {
 *             return (
 *                 <div>
 *                             Test1
 *
 *                       <p>Test2</p>
 *                 </div>
 *             );
 *         }
 *       `,
 *       // TODO: remove two spaces from the Test2 output line
 *       output: `
 *         import React from 'react';
 *
 *         export default function () {
 *             return (
 *                 <div>
 *                     Test1
 *
 *                       <p>Test2</p>
 *                 </div>
 *             );
 *         }
 *       `,
 *       options: [4],
 *       errors: [
 *         { messageId: 'wrongIndent', line: 6 },
 *         { messageId: 'wrongIndent', line: 9 },
 *       ],
 *     }
 *
 * rslint multi-pass output (differs from the single-pass `output` above):
 * `
 *
 *         import React from 'react';
 *
 *         export default function () {
 *             return (
 *                 <div>
 *                     Test1
 *
 *                     <p>Test2</p>
 *                 </div>
 *             );
 *         }
 *       
 * `
 *
 */
