/**
 * @fileoverview Tests for jsx-closing-bracket-location
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-closing-bracket-location/jsx-closing-bracket-location.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-closing-bracket-location', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaFeatures.jsx`) dropped — rslint resolves via tsconfig
 *    and the RuleTester routes every JSX fixture (these carry `</Tag` / `/>`) to a
 *    `.tsx` file, where ts-go parses JSX correctly.
 *  - The single messageId `bracketLocation` = "The closing bracket must be
 *    {{location}}{{details}}"; the RuleTester renders it from the plugin's own
 *    `meta.messages` with each case's `data`. The upstream test computes that
 *    `data` from the local string consts (`MESSAGE_AFTER_PROPS` etc.) and the
 *    local `details()` helper — both are plain string helpers (NOT parser
 *    machinery), so they are KEPT verbatim and referenced from the cases exactly
 *    as upstream does.
 *
 * The upstream file wraps every case in the `valids()` / `invalids()` helpers from
 * `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case across
 * several PARSERS (ESLint-default, @babel/eslint-parser, @typescript-eslint/parser)
 * and append a `// features: [...], parser: ...` comment to `code`/`output`. With
 * the resolved toolchain (ESLint 10.x) the babel variant is skipped
 * (`skipBabel = gte(ESLint.version, '10.0.0')` === true), leaving the default +
 * @typescript-eslint variants — identical under rslint's single ts-go parser. No
 * case sets `features`. That parser-multiplexing is a pure upstream-harness
 * artifact with no rslint analog, so each case is ported ONCE as its literal
 * source (the appended parser-comment is dropped); the code itself is verbatim,
 * including the leading newline + indentation of every plain backtick template
 * and the tab-indented (\t) fixtures (load-bearing: the `line`/`column`/`output`
 * pins are computed against that exact indented source). The four
 * inline-comment-marker template substitutions (which evaluate to an empty
 * string) are kept verbatim.
 *
 * Every invalid case pins BOTH `errors` and `output` — there are NO output-only
 * invalid cases. There are NO `$` unindent template tags, NO `readFileSync`
 * external-fixture cases, and NO `suggestions`. The `._css_` / `._json_` /
 * `._markdown_` test files don't exist for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const MESSAGE_AFTER_PROPS = 'placed after the last prop';
const MESSAGE_AFTER_TAG = 'placed after the opening tag';

const MESSAGE_PROPS_ALIGNED = 'aligned with the last prop';
const MESSAGE_TAG_ALIGNED = 'aligned with the opening tag';
const MESSAGE_LINE_ALIGNED = 'aligned with the line containing the opening tag';

function details(expectedColumn: number, expectedNextLine: boolean) {
  return ` (expected column ${expectedColumn}${expectedNextLine ? ' on the next line)' : ')'}`;
}

ruleTester.run('jsx-closing-bracket-location', null as never, {
  valid: [
    {
      code: `
        <App />
      `,
    },
    {
      code: `
        <App
          // comment
        />
      `,
    },
    {
      code: `
        <App /** comment */ />
      `,
    },
    {
      code: `
        <App foo />
      `,
    },
    {
      code: `
        <App
          foo
        />
      `,
    },
    {
      code: `
        <App
          foo
          // comment
        />
      `,
    },
    {
      code: `
        <App
          {...foo}
        />
      `,
    },
    {
      code: `
        <App
          {...foo}
          // comment
        />
      `,
    },
    {
      code: `
        <App foo />
      `,
      options: [{ location: 'after-props' }],
    },
    {
      code: `
        <App foo />
      `,
      options: [{ location: 'tag-aligned' }],
    },
    {
      code: `
        <App foo />
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        <App
          foo />
      `,
      options: ['after-props'],
    },
    {
      code: `
        <App
          foo
          />
      `,
      options: ['props-aligned'],
    },
    {
      code: `
        <App
          foo />
      `,
      options: [{ location: 'after-props' }],
    },
    {
      code: `
        <App
          foo
        />
      `,
      options: [{ location: 'tag-aligned' }],
    },
    {
      code: `
        <App
          foo
        />
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        <App
          foo
          />
      `,
      options: [{ location: 'props-aligned' }],
    },
    {
      code: `
        <App foo></App>
      `,
    },
    {
      code: `
        <App
          foo
        ></App>
      `,
      options: [{ location: 'tag-aligned' }],
    },
    {
      code: `
        <App
          foo
        ></App>
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        <App
          foo
          ></App>
      `,
      options: [{ location: 'props-aligned' }],
    },
    {
      code: `
        <App
          foo={function() {
            console.log('bar');
          }} />
      `,
      options: [{ location: 'after-props' }],
    },
    {
      code: `
        <App
          foo={function() {
            console.log('bar');
          }}
          />
      `,
      options: [{ location: 'props-aligned' }],
    },
    {
      code: `
        <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
      options: [{ location: 'tag-aligned' }],
    },
    {
      code: `
        <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        <App foo={function() {
          console.log('bar');
        }}/>
      `,
      options: [{ location: 'after-props' }],
    },
    {
      code: `
         <App foo={function() {
                console.log('bar');
              }}
              />
      `,
      options: [{ location: 'props-aligned' }],
    },
    {
      code: `
        <App foo={function() {
          console.log('bar');
        }}
        />
      `,
      options: [{ location: 'tag-aligned' }],
    },
    {
      code: `
        <App foo={function() {
          console.log('bar');
        }}
        />
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        <Provider store>
          <App
            foo />
        </Provider>
      `,
      options: [{ selfClosing: 'after-props' }],
    },
    {
      code: `
        <Provider
          store
        >
          <App
            foo />
        </Provider>
      `,
      options: [{ selfClosing: 'after-props' }],
    },
    {
      code: `
        <Provider
          store>
          <App
            foo
          />
        </Provider>
      `,
      options: [{ nonEmpty: 'after-props' }],
    },
    {
      code: `
        <Provider store>
          <App
            foo
            />
        </Provider>
      `,
      options: [{ selfClosing: 'props-aligned' }],
    },
    {
      code: `
        <Provider
          store
          >
          <App
            foo
          />
        </Provider>
      `,
      options: [{ nonEmpty: 'props-aligned' }],
    },
    {
      code: `
        var x = function() {
          return <App
            foo
                 >
              bar
                 </App>
        }
      `,
      options: [{ location: 'tag-aligned' }],
    },
    {
      code: `
        var x = function() {
          return <App
            foo
                 />
        }
      `,
      options: [{ location: 'tag-aligned' }],
    },
    {
      code: `
        var x = <App
          foo
                />
      `,
      options: [{ location: 'tag-aligned' }],
    },
    {
      code: `
        var x = function() {
          return <App
            foo={function() {
              console.log('bar');
            }}
          />
        }
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        var x = <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        <Provider
          store
        >
          <App
            foo={function() {
              console.log('bar');
            }}
          />
        </Provider>
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        <Provider
          store
        >
          {baz && <App
            foo={function() {
              console.log('bar');
            }}
          />}
        </Provider>
      `,
      options: [{ location: 'line-aligned' }],
    },
    {
      code: `
        <App>
          <Foo
            bar
          >
          </Foo>
          <Foo
            bar />
        </App>
      `,
      options: [
        {
          nonEmpty: false,
          selfClosing: 'after-props',
        },
      ],
    },
    {
      code: `
        <App>
          <Foo
            bar>
          </Foo>
          <Foo
            bar
          />
        </App>
      `,
      options: [
        {
          nonEmpty: 'after-props',
          selfClosing: false,
        },
      ],
    },
    {
      code: `
        <div className={[
          "some",
          "stuff",
          2 ]}
        >
          Some text
        </div>
      `,
      options: [{ location: 'tag-aligned' }],
    },
  ],

  invalid: [
    {
      code: `
        <App
        />
      `,
      output: `
        <App />
      `,
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_TAG, details: '' },
        },
      ],
    },
    {
      code: `
        <App foo
        />
      `,
      output: `
        <App foo/>
      `,
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
        },
      ],
    },
    {
      code: `
        <App foo
        ></App>
      `,
      output: `
        <App foo></App>
      `,
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
        },
      ],
    },
    {
      code: `
        <App
          foo />
      `,
      output: `
        <App
          foo
          />
      `,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_PROPS_ALIGNED,
            details: details(11, true),
          },
          line: 3,
          column: 15,
        },
      ],
    },
    {
      code: `
        <App
          foo />
      `,
      output: `
        <App
          foo
        />
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(9, true),
          },
          line: 3,
          column: 15,
        },
      ],
    },
    {
      code: `
        <App
          foo />
      `,
      output: `
        <App
          foo
        />
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(9, true),
          },
          line: 3,
          column: 15,
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
          foo/>
      `,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
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
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_PROPS_ALIGNED,
            details: details(11, false),
          },
          line: 4,
          column: 9,
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
          foo/>
      `,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
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
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(9, false),
          },
          line: 4,
          column: 11,
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
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(9, false),
          },
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `
        <App
          foo
        ></App>
      `,
      output: `
        <App
          foo></App>
      `,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
        },
      ],
    },
    {
      code: `
        <App
          foo
        ></App>
      `,
      output: `
        <App
          foo
          ></App>
      `,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_PROPS_ALIGNED,
            details: details(11, false),
          },
          line: 4,
          column: 9,
        },
      ],
    },
    {
      code: `
        <App
          foo
          ></App>
      `,
      output: `
        <App
          foo></App>
      `,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
        },
      ],
    },
    {
      code: `
        <App
          foo
          ></App>
      `,
      output: `
        <App
          foo
        ></App>
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(9, false),
          },
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `
        <App
          foo
          ></App>
      `,
      output: `
        <App
          foo
        ></App>
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(9, false),
          },
          line: 4,
          column: 11,
        },
      ],
    },
    {
      code: `
        <Provider
          store>${/* <-- */ ''}
          <App
            foo
            />
        </Provider>
      `,
      output: `
        <Provider
          store
        >
          <App
            foo
            />
        </Provider>
      `,
      options: [{ selfClosing: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(9, true),
          },
          line: 3,
          column: 16,
        },
      ],
    },
    {
      code: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
                                            >
              Button Text
            </Button>
          );
        };
      `,
      output: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
              >
              Button Text
            </Button>
          );
        };
      `,
      options: ['props-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_PROPS_ALIGNED,
            details: details(15, false),
          },
          line: 7,
          column: 45,
        },
      ],
    },
    {
      code: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
                                            >
              Button Text
            </Button>
          );
        };
      `,
      output: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
            >
              Button Text
            </Button>
          );
        };
      `,
      options: ['tag-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(13, false),
          },
          line: 7,
          column: 45,
        },
      ],
    },
    {
      code: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
                                            >
              Button Text
            </Button>
          );
        };
      `,
      output: `
        const Button = function(props) {
          return (
            <Button
              size={size}
              onClick={onClick}
            >
              Button Text
            </Button>
          );
        };
      `,
      options: ['line-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(13, false),
          },
          line: 7,
          column: 45,
        },
      ],
    },
    {
      code: `
        <Provider
          store
          >
          <App
            foo
            />${'' /* <-- */}
        </Provider>
      `,
      output: `
        <Provider
          store
          >
          <App
            foo
          />
        </Provider>
      `,
      options: [{ nonEmpty: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(11, false),
          },
          line: 7,
          column: 13,
        },
      ],
    },
    {
      code: `
        <Provider
          store>${'' /* <-- */}
          <App
            foo />
        </Provider>
      `,
      output: `
        <Provider
          store
        >
          <App
            foo />
        </Provider>
      `,
      options: [{ selfClosing: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(9, true),
          },
          line: 3,
          column: 16,
        },
      ],
    },
    {
      code: `
        <Provider
          store>
          <App
            foo
            />${'' /* <-- */}
        </Provider>
      `,
      output: `
        <Provider
          store>
          <App
            foo
          />
        </Provider>
      `,
      options: [{ nonEmpty: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(11, false),
          },
          line: 6,
          column: 13,
        },
      ],
    },
    {
      code: `
        var x = function() {
          return <App
            foo
                />
        }
      `,
      output: `
        var x = function() {
          return <App
            foo
          />
        }
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(11, false),
          },
          line: 5,
          column: 17,
        },
      ],
    },
    {
      code: `
        var x = <App
          foo
                />
      `,
      output: `
        var x = <App
          foo
        />
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(9, false),
          },
          line: 4,
          column: 17,
        },
      ],
    },
    {
      code: `
        var x = (
          <div
            className="MyComponent"
            {...props} />
        )
      `,
      output: `
        var x = (
          <div
            className="MyComponent"
            {...props}
          />
        )
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(11, true),
          },
          line: 5,
          column: 24,
        },
      ],
    },
    {
      code: `
        var x = (
          <Something
            content={<Foo />} />
        )
      `,
      output: `
        var x = (
          <Something
            content={<Foo />}
          />
        )
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(11, true),
          },
          line: 4,
          column: 31,
        },
      ],
    },
    {
      code: `
        var x = (
          <Something
            />
        )
      `,
      output: `
        var x = (
          <Something />
        )
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_TAG, details: '' },
        },
      ],
    },
    {
      code: `
        <div className={[
          "some",
          "stuff",
          2 ]}>
          Some text
        </div>
      `,
      output: `
        <div className={[
          "some",
          "stuff",
          2 ]}
        >
          Some text
        </div>
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(9, true),
          },
          line: 5,
          column: 15,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo />
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_PROPS_ALIGNED,
            details: details(6, true),
          },
          line: 3,
          column: 10,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo />
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(5, true),
          },
          line: 3,
          column: 10,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo />
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(5, true),
          },
          line: 3,
          column: 10,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo/>
\t\t\t`,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_PROPS_ALIGNED,
            details: details(6, false),
          },
          line: 4,
          column: 5,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo/>
\t\t\t`,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(5, false),
          },
          line: 4,
          column: 6,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(5, false),
          },
          line: 4,
          column: 6,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t></App>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo></App>
\t\t\t`,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t></App>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t></App>
\t\t\t`,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_PROPS_ALIGNED,
            details: details(6, false),
          },
          line: 4,
          column: 5,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t></App>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo></App>
\t\t\t`,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_PROPS, details: '' },
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t></App>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t></App>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(5, false),
          },
          line: 4,
          column: 6,
        },
      ],
    },
    {
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t></App>
\t\t\t`,
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t></App>
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(5, false),
          },
          line: 4,
          column: 6,
        },
      ],
    },
    {
      code: `
\t\t\t\t<Provider
\t\t\t\t\tstore>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      output: `
\t\t\t\t<Provider
\t\t\t\t\tstore
\t\t\t\t>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      options: [{ selfClosing: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(5, true),
          },
          line: 3,
          column: 11,
        },
      ],
    },
    {
      code: `
\t\t\t\tconst Button = function(props) {
\t\t\t\t\treturn (
\t\t\t\t\t\t<Button
\t\t\t\t\t\t\tsize={size}
\t\t\t\t\t\t\tonClick={onClick}
\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>
\t\t\t\t\t\t\tButton Text
\t\t\t\t\t\t</Button>
\t\t\t\t\t);
\t\t\t\t};
\t\t\t`,
      output: `
\t\t\t\tconst Button = function(props) {
\t\t\t\t\treturn (
\t\t\t\t\t\t<Button
\t\t\t\t\t\t\tsize={size}
\t\t\t\t\t\t\tonClick={onClick}
\t\t\t\t\t\t\t>
\t\t\t\t\t\t\tButton Text
\t\t\t\t\t\t</Button>
\t\t\t\t\t);
\t\t\t\t};
\t\t\t`,
      options: ['props-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_PROPS_ALIGNED,
            details: details(8, false),
          },
          line: 7,
          column: 23,
        },
      ],
    },
    {
      code: `
\t\t\t\tconst Button = function(props) {
\t\t\t\t\treturn (
\t\t\t\t\t\t<Button
\t\t\t\t\t\t\tsize={size}
\t\t\t\t\t\t\tonClick={onClick}
\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>
\t\t\t\t\t\t\tButton Text
\t\t\t\t\t\t</Button>
\t\t\t\t\t);
\t\t\t\t};
\t\t\t`,
      output: `
\t\t\t\tconst Button = function(props) {
\t\t\t\t\treturn (
\t\t\t\t\t\t<Button
\t\t\t\t\t\t\tsize={size}
\t\t\t\t\t\t\tonClick={onClick}
\t\t\t\t\t\t>
\t\t\t\t\t\t\tButton Text
\t\t\t\t\t\t</Button>
\t\t\t\t\t);
\t\t\t\t};
\t\t\t`,
      options: ['tag-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(7, false),
          },
          line: 7,
          column: 23,
        },
      ],
    },
    {
      code: `
\t\t\t\tconst Button = function(props) {
\t\t\t\t\treturn (
\t\t\t\t\t\t<Button
\t\t\t\t\t\t\tsize={size}
\t\t\t\t\t\t\tonClick={onClick}
\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t>
\t\t\t\t\t\t\tButton Text
\t\t\t\t\t\t</Button>
\t\t\t\t\t);
\t\t\t\t};
\t\t\t`,
      output: `
\t\t\t\tconst Button = function(props) {
\t\t\t\t\treturn (
\t\t\t\t\t\t<Button
\t\t\t\t\t\t\tsize={size}
\t\t\t\t\t\t\tonClick={onClick}
\t\t\t\t\t\t>
\t\t\t\t\t\t\tButton Text
\t\t\t\t\t\t</Button>
\t\t\t\t\t);
\t\t\t\t};
\t\t\t`,
      options: ['line-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(7, false),
          },
          line: 7,
          column: 23,
        },
      ],
    },
    {
      code: `
\t\t\t\t<Provider
\t\t\t\t\tstore
\t\t\t\t\t>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      output: `
\t\t\t\t<Provider
\t\t\t\t\tstore
\t\t\t\t\t>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      options: [{ nonEmpty: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(6, false),
          },
          line: 7,
          column: 7,
        },
      ],
    },
    {
      code: `
\t\t\t\t<Provider
\t\t\t\t\tstore>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo />
\t\t\t\t</Provider>
\t\t\t`,
      output: `
\t\t\t\t<Provider
\t\t\t\t\tstore
\t\t\t\t>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo />
\t\t\t\t</Provider>
\t\t\t`,
      options: [{ selfClosing: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(5, true),
          },
          line: 3,
          column: 11,
        },
      ],
    },
    {
      code: `
\t\t\t\t<Provider
\t\t\t\t\tstore>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      output: `
\t\t\t\t<Provider
\t\t\t\t\tstore>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      options: [{ nonEmpty: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(6, false),
          },
          line: 6,
          column: 7,
        },
      ],
    },
    {
      code: `
\t\t\t\tvar x = function() {
\t\t\t\t\treturn <App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t\t\t/>
\t\t\t\t}
\t\t\t`,
      output: `
\t\t\t\tvar x = function() {
\t\t\t\t\treturn <App
\t\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t\t}
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(6, false),
          },
          line: 5,
          column: 9,
        },
      ],
    },
    {
      code: `
\t\t\t\tvar x = <App
\t\t\t\t\tfoo
\t\t\t\t\t\t\t\t/>
\t\t\t`,
      output: `
\t\t\t\tvar x = <App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(5, false),
          },
          line: 4,
          column: 9,
        },
      ],
    },
    {
      code: `
\t\t\t\tvar x = (
\t\t\t\t\t<div
\t\t\t\t\t\tclassName="MyComponent"
\t\t\t\t\t\t{...props} />
\t\t\t\t)
\t\t\t`,
      output: `
\t\t\t\tvar x = (
\t\t\t\t\t<div
\t\t\t\t\t\tclassName="MyComponent"
\t\t\t\t\t\t{...props}
\t\t\t\t\t/>
\t\t\t\t)
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(6, true),
          },
          line: 5,
          column: 18,
        },
      ],
    },
    {
      code: `
\t\t\t\tvar x = (
\t\t\t\t\t<Something
\t\t\t\t\t\tcontent={<Foo />} />
\t\t\t\t)
\t\t\t`,
      output: `
\t\t\t\tvar x = (
\t\t\t\t\t<Something
\t\t\t\t\t\tcontent={<Foo />}
\t\t\t\t\t/>
\t\t\t\t)
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(6, true),
          },
          line: 4,
          column: 25,
        },
      ],
    },
    {
      code: `
\t\t\t\tvar x = (
\t\t\t\t\t<Something
\t\t\t\t\t\t/>
\t\t\t\t)
\t\t\t`,
      output: `
\t\t\t\tvar x = (
\t\t\t\t\t<Something />
\t\t\t\t)
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: { location: MESSAGE_AFTER_TAG, details: '' },
        },
      ],
    },
    {
      code: `
\t\t\t\t<div className={[
\t\t\t\t\t"some",
\t\t\t\t\t"stuff",
\t\t\t\t\t2 ]}>
\t\t\t\t\tSome text
\t\t\t\t</div>
\t\t\t`,
      output: `
\t\t\t\t<div className={[
\t\t\t\t\t"some",
\t\t\t\t\t"stuff",
\t\t\t\t\t2 ]}
\t\t\t\t>
\t\t\t\t\tSome text
\t\t\t\t</div>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(5, true),
          },
          line: 5,
          column: 10,
        },
      ],
    },
    {
      code: `
\t\t\t\t\t\t\t<div
\t\t\t\t\t\t\t\tclassName={styles}
\t\t\t\t\t >
\t\t\t\t\t\t\t\t{props}
\t\t\t\t\t\t\t</div>
\t\t\t`,
      output: `
\t\t\t\t\t\t\t<div
\t\t\t\t\t\t\t\tclassName={styles}
\t\t\t\t\t\t\t>
\t\t\t\t\t\t\t\t{props}
\t\t\t\t\t\t\t</div>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(8, false),
          },
          line: 4,
          column: 7,
        },
      ],
    },
    {
      code: `
          <div
            className={styles}
            >
            {props}
          </div>
      `,
      output: `
          <div
            className={styles}
          >
            {props}
          </div>
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(11, false),
          },
          line: 4,
          column: 13,
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
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(11, false),
          },
          line: 4,
          column: 13,
        },
      ],
    },
    {
      code: `
\t\t\t\t\t\t<App
\t\t\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      output: `
\t\t\t\t\t\t<App
\t\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(7, false),
          },
          line: 4,
          column: 6,
        },
      ],
    },
    {
      code: `
        <input
          // comment
          type="text"
          // comment
          />
      `,
      output: `
        <input
          // comment
          type="text"
          // comment
        />
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(9, false),
          },
          line: 6,
          column: 11,
        },
      ],
    },
    {
      code: `
        <input
          // comment
          type="text"
          /**
           * 
           * comment
           * 
           */
          />
      `,
      output: `
        <input
          // comment
          type="text"
          /**
           * 
           * comment
           * 
           */
        />
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_TAG_ALIGNED,
            details: details(9, false),
          },
          line: 10,
          column: 11,
        },
      ],
    },
    {
      code: `
        <input
          // comment
          type="text"

        />
      `,
      output: `
        <input
          // comment
          type="text"/>
      `,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_AFTER_PROPS,
            details: '',
          },
          line: 6,
          column: 9,
        },
      ],
    },
    {
      code: `
        <input
          // comment
          type="text"
          // comment
          />
      `,
      output: `
        <input
          // comment
          type="text"
          // comment
        />
      `,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(9, false),
          },
          line: 6,
          column: 11,
        },
      ],
    },
    {
      code: `
        <input
          // comment
          // comment
          />
      `,
      output: `
        <input
          // comment
          // comment
        />
      `,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(9, false),
          },
          line: 5,
          column: 11,
        },
      ],
    },
    {
      code: `
\t\t\t\t<a
\t\t\t\t\thref="javascript:;"
\t\t\t\t\t// comment
\t\t\t\t\t// comment
\t\t\t\t\t>
\t\t\t\t\ttext
\t\t\t\t</a>
      `,
      output: `
\t\t\t\t<a
\t\t\t\t\thref="javascript:;"
\t\t\t\t\t// comment
\t\t\t\t\t// comment
\t\t\t\t>
\t\t\t\t\ttext
\t\t\t\t</a>
      `,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          data: {
            location: MESSAGE_LINE_ALIGNED,
            details: details(5, false),
          },
          line: 6,
          column: 6,
        },
      ],
    },
  ],
});
