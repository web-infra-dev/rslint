import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectWrongIndentError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'wrongIndent',
    message: 'Expected indentation of 1 tab character but found 0.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectWrongIndentError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'wrongIndent',
    message: 'Expected indentation of 4 space characters but found 0.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectWrongIndentError3(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'wrongIndent',
    message: 'Expected indentation of 12 space characters but found 8.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectWrongIndentError4(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'wrongIndent',
    message: 'Expected indentation of 12 space characters but found 11.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectWrongIndentError5(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'wrongIndent',
    message: 'Expected indentation of 12 space characters but found 10.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectWrongIndentError6(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'wrongIndent',
    message: 'Expected indentation of 10 space characters but found 12.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectWrongIndentError7(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'wrongIndent',
    message: 'Expected indentation of 10 space characters but found 8.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectWrongIndentError8(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'wrongIndent',
    message: 'Expected indentation of 16 space characters but found 12.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-indent', {
  valid: [
    // ---- Single-line / multi-line element basics ----
    { code: `<App></App>;` },
    { code: `<></>;` },
    { code: `<App>\n    <Foo />\n</App>;` },
    {
      code: `<App>\n  <Foo />\n</App>;`,
      options: [2],
    },
    {
      code: `<App>\n\t<Foo />\n</App>;`,
      options: ['tab'],
    },
    {
      code: `<App>\n<Foo />\n</App>;`,
      options: [0],
    },
    // ---- Function/Arrow returns ----
    {
      code: `function App() {\n  return (\n    <App>\n      <Foo />\n    </App>\n  );\n}`,
      options: [2],
    },
    {
      code: `var x = () => (\n  <App>\n    <Foo />\n  </App>\n);`,
      options: [2],
    },
    // ---- Logical / conditional renders ----
    {
      code: `var x = (\n  <div>\n    {condition && (\n    <p>Bar</p>\n    )}\n  </div>\n);`,
      options: [2],
    },
    {
      code: `var x = (\n  <div>\n    {condition && (\n      <p>Bar</p>\n    )}\n  </div>\n);`,
      options: [2, { indentLogicalExpressions: true }],
    },
    {
      code: `var x = foo ?\n    <Foo /> :\n    <Bar />;`,
    },
    // ---- TS wrappers (as / satisfies / non-null) on a return — must
    // not enter the ReturnStatement check. ----
    {
      code: `function App() {\n  return (\n    <App>\n      <Foo />\n    </App>\n    ) as React.ReactElement;\n}`,
      options: [2],
    },
    // ---- React.memo / forwardRef wrap ----
    {
      code: `var W = React.memo(() => (\n  <App>\n    <Foo />\n  </App>\n));`,
      options: [2],
    },
    // ---- map render ----
    {
      code: `var els = items.map((it) => (\n  <Item key={it.id}>\n    {it.name}\n  </Item>\n));`,
      options: [2],
    },
    // ---- Member / namespaced JSX tag names ----
    {
      code: `var x = (\n  <Foo.Bar.Baz>\n    <Quux />\n  </Foo.Bar.Baz>\n);`,
      options: [2],
    },
    {
      code: `var x = (\n  <svg:svg>\n    <svg:rect />\n  </svg:svg>\n);`,
      options: [2],
    },
    // ---- Generic JSX (`<Foo<T>>`) ----
    {
      code: `var x = (\n  <Foo<string>>\n    <Bar />\n  </Foo>\n);`,
      options: [2],
    },
    // ---- 6-level nesting ----
    {
      code: `var x = (\n  <A>\n    <B>\n      <C>\n        <D>\n          <E>\n            <F />\n          </E>\n        </D>\n      </C>\n    </B>\n  </A>\n);`,
      options: [2],
    },
    // ---- Both boolean options on ----
    {
      code: `var x = (\n  <App>\n    {a && (\n      <View\n        Foo={(\n          <Inner />\n        )}\n      />\n    )}\n  </App>\n);`,
      options: [2, { checkAttributes: true, indentLogicalExpressions: true }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        <App></App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        <App>
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        <>
        </>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        <App>
          <Foo />
        </App>
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        <App>
          <></>
        </App>
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        <>
          <Foo />
        </>
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        <App>
        <Foo />
        </App>
      `,
      options: [0],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
          <App>
        <Foo />
          </App>
      `,
      options: [-2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
\t\t\t\t<App>
\t\t\t\t\t<Foo />
\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `
        function App() {
          return <App>
            <Foo />
          </App>;
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        function App() {
          return <App>
            <></>
          </App>;
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `
        function App() {
          return (<App>
            <Foo />
          </App>);
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `
        function App() {
          return (<App>
            <></>
          </App>);
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `
        (
          <div>
            <span />
          </div>
        )
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `
        {
          head.title &&
          <h1>
            {head.title}
          </h1>
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `
        {
          head.title &&
          <>
            {head.title}
          </>
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `
        {
          head.title &&
            <h1>
              {head.title}
            </h1>
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `
        {
          head.title && (
          <h1>
            {head.title}
          </h1>)
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `
        [
          <div />,
          <div />
        ]
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        [
          <></>,
          <></>
        ]
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        <div>
            bar <div>
                bar
                bar {foo}
                bar </div>
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        <>
            bar <>
                bar
                bar {foo}
                bar </>
        </>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `
        foo ?
            <Foo /> :
            <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `
        foo ?
            <></> :
            <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `
        foo ?
            <Foo />
            : <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `
        foo ?
            <></>
            : <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `
        foo ?
            <Foo />
        :
            <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `
        foo ?
            <></>
        :
            <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `
        foo ? <Foo /> :
        <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `
        foo ? <></> :
        <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
        foo ? <Foo />
        : <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `
        foo ? <></>
        : <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 43 (7.37.5)',
      code: `
        foo ? <Foo />
        :
        <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 44 (7.37.5)',
      code: `
        foo ? <></>
        :
        <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 45 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        ) :
            <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 46 (7.37.5)',
      code: `
        foo ? (
            <></>
        ) :
            <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 47 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        )
            : <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 48 (7.37.5)',
      code: `
        foo ? (
            <></>
        )
            : <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 49 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        )
        :
            <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 50 (7.37.5)',
      code: `
        foo ? (
            <></>
        )
        :
            <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 51 (7.37.5)',
      code: `
        foo ?
            <Foo /> : (
                <Bar />
            )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 52 (7.37.5)',
      code: `
        foo ?
            <></> : (
                <></>
            )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 53 (7.37.5)',
      code: `
        foo ?
            <Foo />
        : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 54 (7.37.5)',
      code: `
        foo ?
            <></>
        : (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 55 (7.37.5)',
      code: `
        foo ?
            <Foo />
            : (
                <Bar />
            )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 56 (7.37.5)',
      code: `
        foo ?
            <></>
            : (
                <></>
            )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 57 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        ) : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 58 (7.37.5)',
      code: `
        foo ? (
            <></>
        ) : (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 59 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        )
        : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 60 (7.37.5)',
      code: `
        foo ? (
            <></>
        )
        : (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 61 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        )
        :
        (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 62 (7.37.5)',
      code: `
        foo ? (
            <></>
        )
        :
        (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 63 (7.37.5)',
      code: `
        foo ? <Foo /> : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 64 (7.37.5)',
      code: `
        foo ? <></> : (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 65 (7.37.5)',
      code: `
        foo ? <Foo />
        : (<Bar />)
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 66 (7.37.5)',
      code: `
        foo ? <></>
        : (<></>)
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 67 (7.37.5)',
      code: `
        foo ? <Foo />
        : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 68 (7.37.5)',
      code: `
        foo ? <></>
        : (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 69 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 70 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 71 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 72 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 73 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {do {
                const num = rollDice();
                <Thing num={num} />;
            }}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 74 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {(do {
                const num = rollDice();
                <Thing num={num} />;
            })}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 75 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {do {
                const purposeOfLife = getPurposeOfLife();
                if (purposeOfLife == 42) {
                    <Thing />;
                } else {
                    <AnotherThing />;
                }
            }}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 76 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {(do {
                const purposeOfLife = getPurposeOfLife();
                if (purposeOfLife == 42) {
                    <Thing />;
                } else {
                    <AnotherThing />;
                }
            })}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 77 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {do {
                <Thing num={rollDice()} />;
            }}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 78 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {(do {
                <Thing num={rollDice()} />;
            })}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 79 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {do {
                <Thing num={rollDice()} />;
                <Thing num={rollDice()} />;
            }}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 80 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {(do {
                <Thing num={rollDice()} />;
                <Thing num={rollDice()} />;
            })}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 81 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {do {
                const purposeOfLife = 42;
                <Thing num={purposeOfLife} />;
                <Thing num={purposeOfLife} />;
            }}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 82 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {(do {
                const purposeOfLife = 42;
                <Thing num={purposeOfLife} />;
                <Thing num={purposeOfLife} />;
            })}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 83 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 84 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 85 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 86 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 87 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 88 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 89 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 90 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 91 (7.37.5)',
      code: `
        <App>
            text
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 92 (7.37.5)',
      code: `
        <App>
            text
            text
            text
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 93 (7.37.5)',
      code: `
\t\t\t\t<App>
\t\t\t\t\ttext
\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 94 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 95 (7.37.5)',
      code: `
        function foo() {
          const a = \`aa\`;
          const b = \`b
b\`;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 96 (7.37.5)',
      code: `
        function App() {
          return (
            <App />
          );
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 98 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 99 (7.37.5)',
      code: `
        const Item = ({ id, name, onSelect }) => <div onClick={onSelect}>
          {id}: {name}
        </div>;
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 100 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream flow syntax (Expression expected.).',
      code: `
        type Props = {
          email: string,
          password: string,
          error: string,
        }

        const SomeFormComponent = ({
          email,
          password,
          error,
        }: Props) => (
          // JSX
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 101 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 102 (7.37.5)',
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 103 (7.37.5)',
      code: `
        function test (foo) {
          return foo != null
            ? Math.max(0, Math.min(1, 10))
            : 0
        }
      `,
      options: [99],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 104 (7.37.5)',
      code: `
        function test (foo) {
          return foo != null
            ? <div>foo</div>
            : <div>bar</div>
        }
      `,
      options: [2],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Default 4-space, child under-indented ----
    {
      code: `<App>\n  <Foo />\n</App>;`,
      output: `<App>\n    <Foo />\n</App>;`,
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 4 space characters but found 2.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    // ---- 2-space option, child over-indented ----
    {
      code: `<App>\n    <Foo />\n</App>;`,
      output: `<App>\n  <Foo />\n</App>;`,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 2 space characters but found 4.',
          line: 2,
          column: 5,
          endLine: 2,
          endColumn: 12,
        },
      ],
    },
    // ---- Tab option ----
    {
      code: `<App>\n    <Foo />\n</App>;`,
      output: `<App>\n\t<Foo />\n</App>;`,
      options: ['tab'],
      errors: [expectWrongIndentError1([2, 5, 2, 12])],
    },
    // ---- Closing tag mis-indented ----
    {
      code: `<App>\n    <Foo />\n      </App>;`,
      output: `<App>\n    <Foo />\n</App>;`,
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 0 space characters but found 6.',
          line: 3,
          column: 7,
          endLine: 3,
          endColumn: 13,
        },
      ],
    },
    // ---- JsxExpression mis-indented ----
    {
      code: `<App>\n   {test}\n</App>;`,
      output: `<App>\n    {test}\n</App>;`,
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 4 space characters but found 3.',
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 10,
        },
      ],
    },
    // ---- JsxText mis-indented ----
    {
      code: `<div>\ntext\n</div>;`,
      output: `<div>\n    text\n</div>;`,
      errors: [expectWrongIndentError2([1, 6, 3, 1])],
    },
    // ---- indentLogicalExpressions:true forces deeper indent ----
    {
      code: `var x = (\n  <div>\n    {condition && (\n    <p>Bar</p>\n    )}\n  </div>\n);`,
      output: `var x = (\n  <div>\n    {condition && (\n      <p>Bar</p>\n    )}\n  </div>\n);`,
      options: [2, { indentLogicalExpressions: true }],
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 6 space characters but found 4.',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 8,
        },
      ],
    },
    // ---- Multiline ternary alternate at wrong indent ----
    {
      code: `var x = foo ?\n    <Foo /> :\n<Bar />;`,
      output: `var x = foo ?\n    <Foo /> :\n    <Bar />;`,
      errors: [expectWrongIndentError2([3, 1, 3, 8])],
    },
    // ---- React.memo wrap with mis-indented child ----
    {
      code: `var W = React.memo(() => (\n  <App>\n        <Foo />\n  </App>\n));`,
      output: `var W = React.memo(() => (\n  <App>\n    <Foo />\n  </App>\n));`,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 4 space characters but found 8.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 16,
        },
      ],
    },
    // ---- Member expression tag with mis-indented child ----
    {
      code: `var x = (\n  <Foo.Bar>\n        <Foo.Baz />\n  </Foo.Bar>\n);`,
      output: `var x = (\n  <Foo.Bar>\n    <Foo.Baz />\n  </Foo.Bar>\n);`,
      options: [2],
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 4 space characters but found 8.',
          line: 3,
          column: 9,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    // ---- checkAttributes:true — `)}` realigned to attribute name ----
    {
      code: `const x = (\n  <View\n    Foo={(\n      <Inner />\n)}\n  />\n);`,
      output: `const x = (\n  <View\n    Foo={(\n      <Inner />\n    )}\n  />\n);`,
      options: [2, { checkAttributes: true }],
      errors: [expectWrongIndentError2([5, 1, 5, 2])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <div>
        bar <div>
           bar
           bar {foo}
           bar </div>
        </div>
      `,
      errors: [
        expectWrongIndentError3([2, 14, 3, 13]),
        expectWrongIndentError4([3, 18, 5, 16]),
        expectWrongIndentError4([3, 18, 5, 16]),
        expectWrongIndentError4([5, 21, 6, 16]),
      ],
      output: `
        <div>
            bar <div>
                bar
                bar {foo}
                bar </div>
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        <App>
          <Foo />
        </App>
      `,
      errors: [expectWrongIndentError5([3, 11, 3, 18])],
      output: `
        <App>
            <Foo />
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        <App>
          <></>
        </App>
      `,
      errors: [expectWrongIndentError5([3, 11, 3, 13])],
      output: `
        <App>
            <></>
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        <>
          <Foo />
        </>
      `,
      errors: [expectWrongIndentError5([3, 11, 3, 18])],
      output: `
        <>
            <Foo />
        </>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        <App>
            <Foo />
        </App>
      `,
      options: [2],
      errors: [expectWrongIndentError6([3, 13, 3, 20])],
      output: `
        <App>
          <Foo />
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        <App>
            <Foo />
        </App>
      `,
      options: ['tab'],
      errors: [expectWrongIndentError1([3, 13, 3, 20])],
      output: `
        <App>
\t<Foo />
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
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
          message: 'Expected indentation of 10 space characters but found 17.',
          line: 3,
          column: 11,
          endLine: 5,
          endColumn: 25,
        },
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 10 space characters but found 17.',
          line: 5,
          column: 18,
          endLine: 5,
          endColumn: 24,
        },
      ],
      output: `
        function App() {
          return <App>
            <Foo />
          </App>;
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        function App() {
          return (<App>
            <Foo />
            </App>);
        }
      `,
      options: [2],
      errors: [
        expectWrongIndentError6([3, 11, 5, 21]),
        expectWrongIndentError6([5, 13, 5, 19]),
      ],
      output: `
        function App() {
          return (<App>
            <Foo />
          </App>);
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
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
      errors: [expectWrongIndentError3([4, 9, 4, 14])],
      output: `
        function App() {
          return (
            <App>
              <Foo />
            </App>
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        <App>
           {test}
        </App>
      `,
      errors: [expectWrongIndentError4([3, 12, 3, 18])],
      output: `
        <App>
            {test}
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
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
          message: 'Expected indentation of 20 space characters but found 19.',
          line: 5,
          column: 20,
          endLine: 5,
          endColumn: 33,
        },
      ],
      output: `
        <App>
            {options.map((option, index) => (
                <option key={index} value={option.key}>
                    {option.name}
                </option>
            ))}
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        <App>
        {test}
        </App>
      `,
      options: ['tab'],
      errors: [expectWrongIndentError1([3, 9, 3, 15])],
      output: `
        <App>
\t{test}
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `
\t\t\t\t<App>
\t\t\t\t\t{options.map((option, index) => (
\t\t\t\t\t\t<option key={index} value={option.key}>
\t\t\t\t\t\t{option.name}
\t\t\t\t\t\t</option>
\t\t\t\t\t))}
\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 7 tab characters but found 6.',
          line: 5,
          column: 7,
          endLine: 5,
          endColumn: 20,
        },
      ],
      output: `
\t\t\t\t<App>
\t\t\t\t\t{options.map((option, index) => (
\t\t\t\t\t\t<option key={index} value={option.key}>
\t\t\t\t\t\t\t{option.name}
\t\t\t\t\t\t</option>
\t\t\t\t\t))}
\t\t\t\t</App>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `
\t\t\t\t<App>

\t\t\t\t<Foo />

\t\t\t\t</App>
\t\t\t`,
      options: ['tab'],
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 5 tab characters but found 4.',
          line: 4,
          column: 5,
          endLine: 4,
          endColumn: 12,
        },
      ],
      output: `
\t\t\t\t<App>

\t\t\t\t\t<Foo />

\t\t\t\t</App>
\t\t\t`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `
        [
          <div />,
            <div />
        ]
      `,
      options: [2],
      errors: [expectWrongIndentError6([4, 13, 4, 20])],
      output: `
        [
          <div />,
          <div />
        ]
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        [
          <div />,
            <></>
        ]
      `,
      options: [2],
      errors: [expectWrongIndentError6([4, 13, 4, 15])],
      output: `
        [
          <div />,
          <></>
        ]
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `
        <App>

         <Foo />

        </App>
      `,
      options: ['tab'],
      errors: [expectWrongIndentError1([4, 10, 4, 17])],
      output: `
        <App>

\t<Foo />

        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        <App>

        \t<Foo />

        </App>
      `,
      options: [2],
      errors: [expectWrongIndentError7([4, 10, 4, 17])],
      output: `
        <App>

          <Foo />

        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
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
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 20 space characters but found 16.',
          line: 6,
          column: 17,
          endLine: 6,
          endColumn: 24,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
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
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 20 space characters but found 16.',
          line: 6,
          column: 17,
          endLine: 6,
          endColumn: 24,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `
        foo ?
            <Foo /> :
        <Bar />
      `,
      errors: [expectWrongIndentError3([4, 9, 4, 16])],
      output: `
        foo ?
            <Foo /> :
            <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `
        foo ?
            <Foo /> :
        <></>
      `,
      errors: [expectWrongIndentError3([4, 9, 4, 11])],
      output: `
        foo ?
            <Foo /> :
            <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `
        foo ?
            <Foo />
        :
        <Bar />
      `,
      errors: [expectWrongIndentError3([5, 9, 5, 16])],
      output: `
        foo ?
            <Foo />
        :
            <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `
        foo ? <Foo /> :
            <Bar />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 8 space characters but found 12.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 20,
        },
      ],
      output: `
        foo ? <Foo /> :
        <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `
        foo ?
            <Foo />
        :
        <></>
      `,
      errors: [expectWrongIndentError3([5, 9, 5, 11])],
      output: `
        foo ?
            <Foo />
        :
            <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `
        foo ? <Foo />
        :
              <Bar />
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 8 space characters but found 14.',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 22,
        },
      ],
      output: `
        foo ? <Foo />
        :
        <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        ) :
        <Bar />
      `,
      errors: [expectWrongIndentError3([5, 9, 5, 16])],
      output: `
        foo ? (
            <Foo />
        ) :
            <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        ) :
        <></>
      `,
      errors: [expectWrongIndentError3([5, 9, 5, 11])],
      output: `
        foo ? (
            <Foo />
        ) :
            <></>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 28 (7.37.5)',
      code: `
        foo ? (
            <Foo />
        )
        :
        <Bar />
      `,
      errors: [expectWrongIndentError3([6, 9, 6, 16])],
      output: `
        foo ? (
            <Foo />
        )
        :
            <Bar />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `
        foo ?
            <Foo /> : (
            <Bar />
            )
      `,
      errors: [expectWrongIndentError8([4, 13, 4, 20])],
      output: `
        foo ?
            <Foo /> : (
                <Bar />
            )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `
        foo ?
            <Foo /> : (
            <></>
            )
      `,
      errors: [expectWrongIndentError8([4, 13, 4, 15])],
      output: `
        foo ?
            <Foo /> : (
                <></>
            )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `
        foo ?
            <Foo />
        : (
        <Bar />
        )
      `,
      errors: [expectWrongIndentError3([5, 9, 5, 16])],
      output: `
        foo ?
            <Foo />
        : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `
        foo ?
            <Foo />
            : (
            <Bar />
            )
      `,
      errors: [expectWrongIndentError8([5, 13, 5, 20])],
      output: `
        foo ?
            <Foo />
            : (
                <Bar />
            )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `
        foo ?
            <Foo />
            : (
            <></>
            )
      `,
      errors: [expectWrongIndentError8([5, 13, 5, 15])],
      output: `
        foo ?
            <Foo />
            : (
                <></>
            )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `
        foo ? (
        <Foo />
        ) : (
        <Bar />
        )
      `,
      errors: [
        expectWrongIndentError3([3, 9, 3, 16]),
        expectWrongIndentError3([5, 9, 5, 16]),
      ],
      output: `
        foo ? (
            <Foo />
        ) : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `
        foo ? (
        <></>
        ) : (
        <></>
        )
      `,
      errors: [
        expectWrongIndentError3([3, 9, 3, 11]),
        expectWrongIndentError3([5, 9, 5, 11]),
      ],
      output: `
        foo ? (
            <></>
        ) : (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 36 (7.37.5)',
      code: `
        foo ? (
        <Foo />
        )
        : (
        <Bar />
        )
      `,
      errors: [
        expectWrongIndentError3([3, 9, 3, 16]),
        expectWrongIndentError3([6, 9, 6, 16]),
      ],
      output: `
        foo ? (
            <Foo />
        )
        : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `
        foo ? (
        <Foo />
        )
        :
        (
        <Bar />
        )
      `,
      errors: [
        expectWrongIndentError3([3, 9, 3, 16]),
        expectWrongIndentError3([7, 9, 7, 16]),
      ],
      output: `
        foo ? (
            <Foo />
        )
        :
        (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `
        foo ? (
        <></>
        )
        :
        (
        <></>
        )
      `,
      errors: [
        expectWrongIndentError3([3, 9, 3, 11]),
        expectWrongIndentError3([7, 9, 7, 11]),
      ],
      output: `
        foo ? (
            <></>
        )
        :
        (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        foo ? <Foo /> : (
        <Bar />
        )
      `,
      errors: [expectWrongIndentError3([3, 9, 3, 16])],
      output: `
        foo ? <Foo /> : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        foo ? <Foo /> : (
        <></>
        )
      `,
      errors: [expectWrongIndentError3([3, 9, 3, 11])],
      output: `
        foo ? <Foo /> : (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
        foo ? <Foo />
        : (
        <Bar />
        )
      `,
      errors: [expectWrongIndentError3([4, 9, 4, 16])],
      output: `
        foo ? <Foo />
        : (
            <Bar />
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
        foo ? <Foo />
        : (
        <></>
        )
      `,
      errors: [expectWrongIndentError3([4, 9, 4, 11])],
      output: `
        foo ? <Foo />
        : (
            <></>
        )
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
        <p>
            <div>
                <SelfClosingTag />Text
          </div>
        </p>
      `,
      errors: [expectWrongIndentError5([5, 11, 5, 17])],
      output: `
        <p>
            <div>
                <SelfClosingTag />Text
            </div>
        </p>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
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
      options: [2, { checkAttributes: true }],
      errors: [expectWrongIndentError3([9, 9, 9, 10])],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
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
      options: ['tab', { checkAttributes: true }],
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 2 tab characters but found 0.',
          line: 9,
          column: 1,
          endLine: 9,
          endColumn: 2,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
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
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 16 space characters but found 14.',
          line: 6,
          column: 15,
          endLine: 6,
          endColumn: 18,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {do {
                const num = rollDice();
                    <Thing num={num} />;
            }}
        </span>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 16 space characters but found 20.',
          line: 5,
          column: 21,
          endLine: 5,
          endColumn: 40,
        },
      ],
      output: `
        <span>
            {do {
                const num = rollDice();
                <Thing num={num} />;
            }}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {(do {
                const num = rollDice();
                    <Thing num={num} />;
            })}
        </span>
      `,
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 16 space characters but found 20.',
          line: 5,
          column: 21,
          endLine: 5,
          endColumn: 40,
        },
      ],
      output: `
        <span>
            {(do {
                const num = rollDice();
                <Thing num={num} />;
            })}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {do {
            <Thing num={getPurposeOfLife()} />;
            }}
        </span>
      `,
      errors: [expectWrongIndentError8([4, 13, 4, 47])],
      output: `
        <span>
            {do {
                <Thing num={getPurposeOfLife()} />;
            }}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      skip: 'The native TypeScript parser cannot parse this upstream do expressions syntax (Expression expected.).',
      code: `
        <span>
            {(do {
            <Thing num={getPurposeOfLife()} />;
            })}
        </span>
      `,
      errors: [expectWrongIndentError8([4, 13, 4, 47])],
      output: `
        <span>
            {(do {
                <Thing num={getPurposeOfLife()} />;
            })}
        </span>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 51 (7.37.5)',
      code: `
        <div>
        text
        </div>
      `,
      errors: [expectWrongIndentError3([2, 14, 4, 9])],
      output: `
        <div>
            text
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 52 (7.37.5)',
      code: `
        <div>
          text
        text
        </div>
      `,
      errors: [
        expectWrongIndentError5([2, 14, 5, 9]),
        expectWrongIndentError3([2, 14, 5, 9]),
      ],
      output: `
        <div>
            text
            text
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 53 (7.37.5)',
      code: `
        <div>
        \t  text
          \t  text
        </div>
      `,
      errors: [
        expectWrongIndentError3([2, 14, 5, 9]),
        expectWrongIndentError5([2, 14, 5, 9]),
      ],
      output: `
        <div>
            text
            text
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 54 (7.37.5)',
      code: `
        <div>
        \t\ttext
        </div>
      `,
      options: ['tab'],
      errors: [expectWrongIndentError1([2, 14, 4, 9])],
      output: `
        <div>
\ttext
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 55 (7.37.5)',
      code: `
        <>
        aaa
        </>
      `,
      errors: [expectWrongIndentError3([2, 11, 4, 9])],
      output: `
        <>
            aaa
        </>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 56 (7.37.5)',
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
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 18 space characters but found 8.',
          line: 5,
          column: 9,
          endLine: 5,
          endColumn: 14,
        },
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 57 (7.37.5)',
      code: `
        function App() {
          return (
            <App />
            );
        }
      `,
      options: [2],
      errors: [expectWrongIndentError6([3, 11, 5, 15])],
      output: `
        function App() {
          return (
            <App />
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 58 (7.37.5)',
      code: `
        function App() {
          return (
            <App />
        );
        }
      `,
      options: [2],
      errors: [expectWrongIndentError7([3, 11, 5, 11])],
      output: `
        function App() {
          return (
            <App />
          );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 59 (7.37.5)',
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
      options: [2],
      errors: [
        expectWrongIndentError6([3, 13, 5, 18]),
        expectWrongIndentError6([6, 13, 8, 18]),
      ],
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
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 60 (7.37.5)',
      code: `
        const IndexPage = () => (
          <h1>
        {"Hi people"}
        <button/>
        </h1>
        );
      `,
      options: [2],
      errors: [
        expectWrongIndentError3([4, 9, 4, 22]),
        expectWrongIndentError3([5, 9, 5, 18]),
        expectWrongIndentError7([6, 9, 6, 14]),
      ],
      output: `
        const IndexPage = () => (
          <h1>
            {"Hi people"}
            <button/>
          </h1>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 61 (7.37.5)',
      code: `
        const IndexPage = () => (
          <h1>
        Hi people
        <button/>
        </h1>
        );
      `,
      options: [2],
      errors: [
        expectWrongIndentError3([3, 15, 5, 9]),
        expectWrongIndentError3([5, 9, 5, 18]),
        expectWrongIndentError7([6, 9, 6, 14]),
      ],
      output: `
        const IndexPage = () => (
          <h1>
            Hi people
            <button/>
          </h1>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 62 (7.37.5)',
      code: `
        const IndexPage = () => (
          <h1>
            Hi people
        <button/>
          </h1>
        );
      `,
      options: [2],
      errors: [expectWrongIndentError3([5, 9, 5, 18])],
      output: `
        const IndexPage = () => (
          <h1>
            Hi people
            <button/>
          </h1>
        );
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 63 (7.37.5)',
      code: `
        import React from 'react';

        export default function () {
            return (
                <div>
                            Test1

                      <p>Test2</p>
                </div>
            );
        }
      `,
      options: [4],
      errors: [
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 20 space characters but found 28.',
          line: 6,
          column: 22,
          endLine: 9,
          endColumn: 23,
        },
        {
          messageId: 'wrongIndent',
          message: 'Expected indentation of 20 space characters but found 22.',
          line: 9,
          column: 23,
          endLine: 9,
          endColumn: 26,
        },
      ],
      output: `
        import React from 'react';

        export default function () {
            return (
                <div>
                    Test1

                    <p>Test2</p>
                </div>
            );
        }
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
});
