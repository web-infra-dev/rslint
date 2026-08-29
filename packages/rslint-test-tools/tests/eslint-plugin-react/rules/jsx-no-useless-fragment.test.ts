import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-no-useless-fragment', {} as never, {
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
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
    {
      code: '<>{}</>',
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
    {
      code: '<p>moo<>foo</></p>',
      errors: [
        { messageId: 'NeedsMoreChildren' },
        { messageId: 'ChildOfHtmlElement' },
      ],
    },
    {
      code: '<>{meow}</>',
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
    {
      code: '<p><>{meow}</></p>',
      errors: [
        { messageId: 'NeedsMoreChildren' },
        { messageId: 'ChildOfHtmlElement' },
      ],
    },
    {
      code: '<><div/></>',
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
    {
      code: `
        <>
          <div/>
        </>
      `,
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
    {
      code: '<Fragment />',
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
    {
      code: `
        <React.Fragment>
          <Foo />
        </React.Fragment>
      `,
      errors: [{ messageId: 'NeedsMoreChildren' }],
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
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
    {
      // Not safe to fix this case because `Eeee` might require child be ReactElement
      code: '<Eeee><>foo</></Eeee>',
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
    {
      code: '<div><>foo</></div>',
      errors: [
        { messageId: 'NeedsMoreChildren' },
        { messageId: 'ChildOfHtmlElement' },
      ],
    },
    {
      code: '<div><>{"a"}{"b"}</></div>',
      errors: [{ messageId: 'ChildOfHtmlElement' }],
    },
    {
      code: `
        <section>
          <Eeee />
          <Eeee />
          <>{"a"}{"b"}</>
        </section>`,
      errors: [{ messageId: 'ChildOfHtmlElement' }],
    },
    {
      code: '<div><Fragment>{"a"}{"b"}</Fragment></div>',
      errors: [{ messageId: 'ChildOfHtmlElement' }],
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
        { messageId: 'ChildOfHtmlElement', line: 3 },
        { messageId: 'ChildOfHtmlElement', line: 7 },
      ],
    },
    {
      code: '<div>a <>{""}{""}</> a</div>',
      errors: [{ messageId: 'ChildOfHtmlElement' }],
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
        { messageId: 'NeedsMoreChildren', line: 4 },
        { messageId: 'ChildOfHtmlElement', line: 4 },
      ],
    },
    {
      // Ensure allowExpressions still catches expected violations
      code: '<><Foo>{moo}</Foo></>',
      options: [{ allowExpressions: true }],
      errors: [{ messageId: 'NeedsMoreChildren' }],
    },
  ],
});
