import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectBracketLocationError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bracketLocation',
    message: 'The closing bracket must be placed after the opening tag',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBracketLocationError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bracketLocation',
    message: 'The closing bracket must be placed after the last prop',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBracketLocationError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bracketLocation',
    message:
      'The closing bracket must be aligned with the opening tag (expected column 9 on the next line)',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBracketLocationError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bracketLocation',
    message:
      'The closing bracket must be aligned with the line containing the opening tag (expected column 9)',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBracketLocationError5(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bracketLocation',
    message:
      'The closing bracket must be aligned with the opening tag (expected column 11)',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBracketLocationError6(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bracketLocation',
    message:
      'The closing bracket must be aligned with the opening tag (expected column 5 on the next line)',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBracketLocationError7(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'bracketLocation',
    message:
      'The closing bracket must be aligned with the line containing the opening tag (expected column 5)',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-closing-bracket-location', {
  valid: [
    { code: `<App />` },
    { code: `<App foo />` },
    { code: `<App\n  foo\n/>` },
    {
      code: `<App\n  foo />`,
      options: ['after-props'],
    },
    {
      code: `<App\n  foo\n  />`,
      options: ['props-aligned'],
    },
    {
      code: `<App foo></App>`,
    },
    {
      code: `<App>\n  <Foo\n    bar\n  >\n  </Foo>\n  <Foo\n    bar />\n</App>`,
      options: [{ nonEmpty: false, selfClosing: 'after-props' }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        <App />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        <App foo />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        <App
          foo
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        <App foo />
      `,
      options: [{ location: 'after-props' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        <App foo />
      `,
      options: [{ location: 'tag-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        <App foo />
      `,
      options: [{ location: 'line-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        <App
          foo />
      `,
      options: ['after-props'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        <App
          foo
          />
      `,
      options: ['props-aligned'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        <App
          foo />
      `,
      options: [{ location: 'after-props' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        <App
          foo
        />
      `,
      options: [{ location: 'tag-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        <App
          foo
        />
      `,
      options: [{ location: 'line-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        <App
          foo
          />
      `,
      options: [{ location: 'props-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        <App foo></App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        <App
          foo
        ></App>
      `,
      options: [{ location: 'tag-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `
        <App
          foo
        ></App>
      `,
      options: [{ location: 'line-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `
        <App
          foo
          ></App>
      `,
      options: [{ location: 'props-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        <App
          foo={function() {
            console.log('bar');
          }} />
      `,
      options: [{ location: 'after-props' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `
        <App
          foo={function() {
            console.log('bar');
          }}
          />
      `,
      options: [{ location: 'props-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `
        <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
      options: [{ location: 'tag-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
      options: [{ location: 'line-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        <App foo={function() {
          console.log('bar');
        }}/>
      `,
      options: [{ location: 'after-props' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
         <App foo={function() {
                console.log('bar');
              }}
              />
      `,
      options: [{ location: 'props-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        <App foo={function() {
          console.log('bar');
        }}
        />
      `,
      options: [{ location: 'tag-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        <App foo={function() {
          console.log('bar');
        }}
        />
      `,
      options: [{ location: 'line-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `
        <Provider store>
          <App
            foo />
        </Provider>
      `,
      options: [{ selfClosing: 'after-props' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        <Provider
          store
        >
          <App
            foo />
        </Provider>
      `,
      options: [{ selfClosing: 'after-props' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        <Provider
          store>
          <App
            foo
          />
        </Provider>
      `,
      options: [{ nonEmpty: 'after-props' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        <Provider store>
          <App
            foo
            />
        </Provider>
      `,
      options: [{ selfClosing: 'props-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        var x = function() {
          return <App
            foo
                 />
        }
      `,
      options: [{ location: 'tag-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        var x = <App
          foo
                />
      `,
      options: [{ location: 'tag-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        var x = <App
          foo={function() {
            console.log('bar');
          }}
        />
      `,
      options: [{ location: 'line-aligned' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
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
      options: [{ nonEmpty: false, selfClosing: 'after-props' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
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
      options: [{ nonEmpty: 'after-props', selfClosing: false }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: `<App\n/>`,
      options: ['tag-aligned'],
      errors: [expectBracketLocationError1([2, 1, 2, 2])],
      output: `<App />`,
    },
    {
      code: `<App foo\n/>`,
      options: ['tag-aligned'],
      errors: [expectBracketLocationError2([2, 1, 2, 2])],
      output: `<App foo/>`,
    },
    {
      code: `<App\n  foo />`,
      options: ['tag-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 1 on the next line)',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 8,
        },
      ],
      output: `<App
  foo
/>`,
    },
    {
      code: `<App\n  foo />`,
      options: ['props-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 3 on the next line)',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 8,
        },
      ],
      output: `<App
  foo
  />`,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <App
        />
      `,
      errors: [expectBracketLocationError1([3, 9, 3, 10])],
      output: `
        <App />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        <App foo
        />
      `,
      errors: [expectBracketLocationError2([3, 9, 3, 10])],
      output: `
        <App foo/>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        <App foo
        ></App>
      `,
      errors: [expectBracketLocationError2([3, 9, 3, 10])],
      output: `
        <App foo></App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        <App
          foo />
      `,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 11 on the next line)',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 16,
        },
      ],
      output: `
        <App
          foo
          />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        <App
          foo />
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [expectBracketLocationError3([3, 15, 3, 16])],
      output: `
        <App
          foo
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        <App
          foo />
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 9 on the next line)',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 16,
        },
      ],
      output: `
        <App
          foo
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        <App
          foo
        />
      `,
      options: [{ location: 'after-props' }],
      errors: [expectBracketLocationError2([4, 9, 4, 10])],
      output: `
        <App
          foo/>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        <App
          foo
        />
      `,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 11)',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 10,
        },
      ],
      output: `
        <App
          foo
          />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        <App
          foo
          />
      `,
      options: [{ location: 'after-props' }],
      errors: [expectBracketLocationError2([4, 11, 4, 12])],
      output: `
        <App
          foo/>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        <App
          foo
          />
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 9)',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 12,
        },
      ],
      output: `
        <App
          foo
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        <App
          foo
          />
      `,
      options: [{ location: 'line-aligned' }],
      errors: [expectBracketLocationError4([4, 11, 4, 12])],
      output: `
        <App
          foo
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        <App
          foo
        ></App>
      `,
      options: [{ location: 'after-props' }],
      errors: [expectBracketLocationError2([4, 9, 4, 10])],
      output: `
        <App
          foo></App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
        <App
          foo
        ></App>
      `,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 11)',
          line: 4,
          column: 9,
          endLine: 4,
          endColumn: 10,
        },
      ],
      output: `
        <App
          foo
          ></App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
        <App
          foo
          ></App>
      `,
      options: [{ location: 'after-props' }],
      errors: [expectBracketLocationError2([4, 11, 4, 12])],
      output: `
        <App
          foo></App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        <App
          foo
          ></App>
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 9)',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 12,
        },
      ],
      output: `
        <App
          foo
        ></App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        <App
          foo
          ></App>
      `,
      options: [{ location: 'line-aligned' }],
      errors: [expectBracketLocationError4([4, 11, 4, 12])],
      output: `
        <App
          foo
        ></App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        <Provider
          store>
          <App
            foo
            />
        </Provider>
      `,
      options: [{ selfClosing: 'props-aligned' }],
      errors: [expectBracketLocationError3([3, 16, 3, 17])],
      output: `
        <Provider
          store
        >
          <App
            foo
            />
        </Provider>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
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
      options: ['props-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 15)',
          line: 7,
          column: 45,
          endLine: 7,
          endColumn: 46,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
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
      options: ['tag-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 13)',
          line: 7,
          column: 45,
          endLine: 7,
          endColumn: 46,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
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
      options: ['line-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 13)',
          line: 7,
          column: 45,
          endLine: 7,
          endColumn: 46,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
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
      errors: [expectBracketLocationError5([7, 13, 7, 14])],
      output: `
        <Provider
          store
          >
          <App
            foo
          />
        </Provider>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        <Provider
          store>
          <App
            foo />
        </Provider>
      `,
      options: [{ selfClosing: 'after-props' }],
      errors: [expectBracketLocationError3([3, 16, 3, 17])],
      output: `
        <Provider
          store
        >
          <App
            foo />
        </Provider>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        <Provider
          store>
          <App
            foo
            />
        </Provider>
      `,
      options: [{ nonEmpty: 'after-props' }],
      errors: [expectBracketLocationError5([6, 13, 6, 14])],
      output: `
        <Provider
          store>
          <App
            foo
          />
        </Provider>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
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
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 11)',
          line: 5,
          column: 17,
          endLine: 5,
          endColumn: 18,
        },
      ],
      output: `
        var x = function() {
          return <App
            foo
          />
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        var x = <App
          foo
                />
      `,
      options: [{ location: 'line-aligned' }],
      errors: [expectBracketLocationError4([4, 17, 4, 18])],
      output: `
        var x = <App
          foo
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        var x = (
          <div
            className="MyComponent"
            {...props} />
        )
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 11 on the next line)',
          line: 5,
          column: 24,
          endLine: 5,
          endColumn: 25,
        },
      ],
      output: `
        var x = (
          <div
            className="MyComponent"
            {...props}
          />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        var x = (
          <Something
            content={<Foo />} />
        )
      `,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 11 on the next line)',
          line: 4,
          column: 31,
          endLine: 4,
          endColumn: 32,
        },
      ],
      output: `
        var x = (
          <Something
            content={<Foo />}
          />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        var x = (
          <Something
            />
        )
      `,
      options: [{ location: 'line-aligned' }],
      errors: [expectBracketLocationError1([4, 13, 4, 14])],
      output: `
        var x = (
          <Something />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        <div className={[
          "some",
          "stuff",
          2 ]}>
          Some text
        </div>
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [expectBracketLocationError3([5, 15, 5, 16])],
      output: `
        <div className={[
          "some",
          "stuff",
          2 ]}
        >
          Some text
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo />
\t\t\t`,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 6 on the next line)',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo />
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [expectBracketLocationError6([3, 10, 3, 11])],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo />
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 5 on the next line)',
          line: 3,
          column: 10,
          endLine: 3,
          endColumn: 11,
        },
      ],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'after-props' }],
      errors: [expectBracketLocationError2([4, 5, 4, 6])],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 6)',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 6,
        },
      ],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'after-props' }],
      errors: [expectBracketLocationError2([4, 6, 4, 7])],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 5)',
          line: 4,
          column: 6,
          endLine: 4,
          endColumn: 7,
        },
      ],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [expectBracketLocationError7([4, 6, 4, 7])],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t></App>
\t\t\t`,
      options: [{ location: 'after-props' }],
      errors: [expectBracketLocationError2([4, 5, 4, 6])],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo></App>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t></App>
\t\t\t`,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 6)',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 6,
        },
      ],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t></App>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t></App>
\t\t\t`,
      options: [{ location: 'after-props' }],
      errors: [expectBracketLocationError2([4, 6, 4, 7])],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo></App>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t></App>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 5)',
          line: 4,
          column: 6,
          endLine: 4,
          endColumn: 7,
        },
      ],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t></App>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t\t></App>
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [expectBracketLocationError7([4, 6, 4, 7])],
      output: `
\t\t\t\t<App
\t\t\t\t\tfoo
\t\t\t\t></App>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
\t\t\t\t<Provider
\t\t\t\t\tstore>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      options: [{ selfClosing: 'props-aligned' }],
      errors: [expectBracketLocationError6([3, 11, 3, 12])],
      output: `
\t\t\t\t<Provider
\t\t\t\t\tstore
\t\t\t\t>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
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
      options: ['props-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the last prop (expected column 8)',
          line: 7,
          column: 23,
          endLine: 7,
          endColumn: 24,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
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
      options: ['tag-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 7)',
          line: 7,
          column: 23,
          endLine: 7,
          endColumn: 24,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
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
      options: ['line-aligned'],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 7)',
          line: 7,
          column: 23,
          endLine: 7,
          endColumn: 24,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
\t\t\t\t<Provider
\t\t\t\t\tstore
\t\t\t\t\t>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      options: [{ nonEmpty: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 6)',
          line: 7,
          column: 7,
          endLine: 7,
          endColumn: 8,
        },
      ],
      output: `
\t\t\t\t<Provider
\t\t\t\t\tstore
\t\t\t\t\t>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
\t\t\t\t<Provider
\t\t\t\t\tstore>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo />
\t\t\t\t</Provider>
\t\t\t`,
      options: [{ selfClosing: 'after-props' }],
      errors: [expectBracketLocationError6([3, 11, 3, 12])],
      output: `
\t\t\t\t<Provider
\t\t\t\t\tstore
\t\t\t\t>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo />
\t\t\t\t</Provider>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
\t\t\t\t<Provider
\t\t\t\t\tstore>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      options: [{ nonEmpty: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 6)',
          line: 6,
          column: 7,
          endLine: 6,
          endColumn: 8,
        },
      ],
      output: `
\t\t\t\t<Provider
\t\t\t\t\tstore>
\t\t\t\t\t<App
\t\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t\t</Provider>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
\t\t\t\tvar x = function() {
\t\t\t\t\treturn <App
\t\t\t\t\t\tfoo
\t\t\t\t\t\t\t\t/>
\t\t\t\t}
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 6)',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 10,
        },
      ],
      output: `
\t\t\t\tvar x = function() {
\t\t\t\t\treturn <App
\t\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t\t}
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `
\t\t\t\tvar x = <App
\t\t\t\t\tfoo
\t\t\t\t\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [expectBracketLocationError7([4, 9, 4, 10])],
      output: `
\t\t\t\tvar x = <App
\t\t\t\t\tfoo
\t\t\t\t/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      code: `
\t\t\t\tvar x = (
\t\t\t\t\t<div
\t\t\t\t\t\tclassName="MyComponent"
\t\t\t\t\t\t{...props} />
\t\t\t\t)
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 6 on the next line)',
          line: 5,
          column: 18,
          endLine: 5,
          endColumn: 19,
        },
      ],
      output: `
\t\t\t\tvar x = (
\t\t\t\t\t<div
\t\t\t\t\t\tclassName="MyComponent"
\t\t\t\t\t\t{...props}
\t\t\t\t\t/>
\t\t\t\t)
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 52 (7.37.5)',
      code: `
\t\t\t\tvar x = (
\t\t\t\t\t<Something
\t\t\t\t\t\tcontent={<Foo />} />
\t\t\t\t)
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the line containing the opening tag (expected column 6 on the next line)',
          line: 4,
          column: 25,
          endLine: 4,
          endColumn: 26,
        },
      ],
      output: `
\t\t\t\tvar x = (
\t\t\t\t\t<Something
\t\t\t\t\t\tcontent={<Foo />}
\t\t\t\t\t/>
\t\t\t\t)
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 53 (7.37.5)',
      code: `
\t\t\t\tvar x = (
\t\t\t\t\t<Something
\t\t\t\t\t\t/>
\t\t\t\t)
\t\t\t`,
      options: [{ location: 'line-aligned' }],
      errors: [expectBracketLocationError1([4, 7, 4, 8])],
      output: `
\t\t\t\tvar x = (
\t\t\t\t\t<Something />
\t\t\t\t)
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 54 (7.37.5)',
      code: `
\t\t\t\t<div className={[
\t\t\t\t\t"some",
\t\t\t\t\t"stuff",
\t\t\t\t\t2 ]}>
\t\t\t\t\tSome text
\t\t\t\t</div>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [expectBracketLocationError6([5, 10, 5, 11])],
      output: `
\t\t\t\t<div className={[
\t\t\t\t\t"some",
\t\t\t\t\t"stuff",
\t\t\t\t\t2 ]}
\t\t\t\t>
\t\t\t\t\tSome text
\t\t\t\t</div>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 55 (7.37.5)',
      code: `
\t\t\t\t\t\t\t<div
\t\t\t\t\t\t\t\tclassName={styles}
\t\t\t\t\t >
\t\t\t\t\t\t\t\t{props}
\t\t\t\t\t\t\t</div>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 8)',
          line: 4,
          column: 7,
          endLine: 4,
          endColumn: 8,
        },
      ],
      output: `
\t\t\t\t\t\t\t<div
\t\t\t\t\t\t\t\tclassName={styles}
\t\t\t\t\t\t\t>
\t\t\t\t\t\t\t\t{props}
\t\t\t\t\t\t\t</div>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 56 (7.37.5)',
      code: `
          <div
            className={styles}
            >
            {props}
          </div>
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [expectBracketLocationError5([4, 13, 4, 14])],
      output: `
          <div
            className={styles}
          >
            {props}
          </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 57 (7.37.5)',
      code: `
          <App
            foo
            />
      `,
      options: [{ location: 'tag-aligned' }],
      errors: [expectBracketLocationError5([4, 13, 4, 14])],
      output: `
          <App
            foo
          />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 58 (7.37.5)',
      code: `
\t\t\t\t\t\t<App
\t\t\t\t\t\t\tfoo
\t\t\t\t\t/>
\t\t\t`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message:
            'The closing bracket must be aligned with the opening tag (expected column 7)',
          line: 4,
          column: 6,
          endLine: 4,
          endColumn: 7,
        },
      ],
      output: `
\t\t\t\t\t\t<App
\t\t\t\t\t\t\tfoo
\t\t\t\t\t\t/>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
