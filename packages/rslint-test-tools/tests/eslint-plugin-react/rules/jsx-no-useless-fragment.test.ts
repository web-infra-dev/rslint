import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNeedsMoreChildrenError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'NeedsMoreChildren',
    message:
      'Fragments should contain more than one child - otherwise, there’s no need for a Fragment at all.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectChildOfHtmlElementError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'ChildOfHtmlElement',
    message: 'Passing a fragment to an HTML element is useless.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-no-useless-fragment', {
  valid: [
    { code: '<><Foo /><Bar /></>' },
    { code: '<>foo<div /></>' },
    { code: '<> <div /></>' },
    { code: '<>{"moo"} </>' },
    { code: '<NotFragment />' },
    { code: '<React.NotFragment />' },
    { code: '<NotReact.Fragment />' },
    { code: '<Foo><><div /><div /></></Foo>' },
    { code: '<div p={<>{"a"}{"b"}</>} />' },
    { code: '<Fragment key={item.id}>{item.value}</Fragment>' },
    { code: '<Fooo content={<>eeee ee eeeeeee eeeeeeee</>} />' },
    { code: '<>{foos.map(foo => foo)}</>' },
    {
      code: '<>{moo}</>',
      options: [{ allowExpressions: true }],
    },
    {
      code: `
        <>
          {moo}
        </>
      `,
      options: [{ allowExpressions: true }],
    },
  ],
  invalid: [
    {
      code: '<></>',
      errors: [expectNeedsMoreChildrenError([1, 1, 1, 6])],
    },
    {
      code: '<>{}</>',
      errors: [expectNeedsMoreChildrenError([1, 1, 1, 8])],
    },
    {
      code: '<p>moo<>foo</></p>',
      errors: [
        expectNeedsMoreChildrenError([1, 7, 1, 15]),
        expectChildOfHtmlElementError([1, 7, 1, 15]),
      ],
      output: `<p>moofoo</p>`,
    },
    {
      code: '<>{meow}</>',
      errors: [expectNeedsMoreChildrenError([1, 1, 1, 12])],
    },
    {
      code: '<p><>{meow}</></p>',
      errors: [
        expectNeedsMoreChildrenError([1, 4, 1, 15]),
        expectChildOfHtmlElementError([1, 4, 1, 15]),
      ],
      output: `<p>{meow}</p>`,
    },
    {
      code: '<><div/></>',
      errors: [expectNeedsMoreChildrenError([1, 1, 1, 12])],
      output: `<div/>`,
    },
    {
      code: `
        <>
          <div/>
        </>
      `,
      errors: [expectNeedsMoreChildrenError([2, 9, 4, 12])],
      output: `
        <div/>
      `,
    },
    {
      code: '<Fragment />',
      errors: [expectNeedsMoreChildrenError([1, 1, 1, 13])],
    },
    {
      code: `
        <React.Fragment>
          <Foo />
        </React.Fragment>
      `,
      errors: [expectNeedsMoreChildrenError([2, 9, 4, 26])],
      output: `
        <Foo />
      `,
    },
    {
      code: `
        <SomeReact.SomeFragment>
          {foo}
        </SomeReact.SomeFragment>
      `,
      settings: {
        react: {
          pragma: 'SomeReact',
          fragment: 'SomeFragment',
        },
      },
      errors: [expectNeedsMoreChildrenError([2, 9, 4, 34])],
    },
    {
      // Not safe to fix this case because `Eeee` might require child be ReactElement
      code: '<Eeee><>foo</></Eeee>',
      errors: [expectNeedsMoreChildrenError([1, 7, 1, 15])],
    },
    {
      code: '<div><>foo</></div>',
      errors: [
        expectNeedsMoreChildrenError([1, 6, 1, 14]),
        expectChildOfHtmlElementError([1, 6, 1, 14]),
      ],
      output: `<div>foo</div>`,
    },
    {
      code: '<div><>{"a"}{"b"}</></div>',
      errors: [expectChildOfHtmlElementError([1, 6, 1, 21])],
      output: `<div>{"a"}{"b"}</div>`,
    },
    {
      code: `
        <section>
          <Eeee />
          <Eeee />
          <>{"a"}{"b"}</>
        </section>`,
      errors: [expectChildOfHtmlElementError([5, 11, 5, 26])],
      output: `
        <section>
          <Eeee />
          <Eeee />
          {"a"}{"b"}
        </section>`,
    },
    {
      code: '<div><Fragment>{"a"}{"b"}</Fragment></div>',
      errors: [expectChildOfHtmlElementError([1, 6, 1, 37])],
      output: `<div>{"a"}{"b"}</div>`,
    },
    {
      // whitespace tricky case
      code: `
        <section>
          git<>
            <b>hub</b>.
          </>

          git<> <b>hub</b></>
        </section>`,
      errors: [
        expectChildOfHtmlElementError([3, 14, 5, 14]),
        expectChildOfHtmlElementError([7, 14, 7, 30]),
      ],
      output: `
        <section>
          git<b>hub</b>.

          git <b>hub</b>
        </section>`,
    },
    {
      code: '<div>a <>{""}{""}</> a</div>',
      errors: [expectChildOfHtmlElementError([1, 8, 1, 21])],
      output: `<div>a {""}{""} a</div>`,
    },
    {
      code: `
        const Comp = () => (
          <html>
            <React.Fragment />
          </html>
        );
      `,
      errors: [
        expectNeedsMoreChildrenError([4, 13, 4, 31]),
        expectChildOfHtmlElementError([4, 13, 4, 31]),
      ],
      output: `
        const Comp = () => (
          <html>
\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20\x20
          </html>
        );
      `,
    },
    {
      // Ensure allowExpressions still catches expected violations
      code: '<><Foo>{moo}</Foo></>',
      options: [{ allowExpressions: true }],
      errors: [expectNeedsMoreChildrenError([1, 1, 1, 22])],
      output: `<Foo>{moo}</Foo>`,
    },
  ],
});
