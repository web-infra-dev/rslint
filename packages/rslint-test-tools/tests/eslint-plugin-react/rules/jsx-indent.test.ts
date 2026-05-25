import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-indent', {} as never, {
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
  ],
  invalid: [
    // ---- Default 4-space, child under-indented ----
    {
      code: `<App>\n  <Foo />\n</App>;`,
      output: `<App>\n    <Foo />\n</App>;`,
      errors: [
        { message: 'Expected indentation of 4 space characters but found 2.' },
      ],
    },
    // ---- 2-space option, child over-indented ----
    {
      code: `<App>\n    <Foo />\n</App>;`,
      output: `<App>\n  <Foo />\n</App>;`,
      options: [2],
      errors: [
        { message: 'Expected indentation of 2 space characters but found 4.' },
      ],
    },
    // ---- Tab option ----
    {
      code: `<App>\n    <Foo />\n</App>;`,
      output: `<App>\n\t<Foo />\n</App>;`,
      options: ['tab'],
      errors: [
        { message: 'Expected indentation of 1 tab character but found 0.' },
      ],
    },
    // ---- Closing tag mis-indented ----
    {
      code: `<App>\n    <Foo />\n      </App>;`,
      output: `<App>\n    <Foo />\n</App>;`,
      errors: [
        { message: 'Expected indentation of 0 space characters but found 6.' },
      ],
    },
    // ---- JsxExpression mis-indented ----
    {
      code: `<App>\n   {test}\n</App>;`,
      output: `<App>\n    {test}\n</App>;`,
      errors: [
        { message: 'Expected indentation of 4 space characters but found 3.' },
      ],
    },
    // ---- JsxText mis-indented ----
    {
      code: `<div>\ntext\n</div>;`,
      output: `<div>\n    text\n</div>;`,
      errors: [
        { message: 'Expected indentation of 4 space characters but found 0.' },
      ],
    },
    // ---- indentLogicalExpressions:true forces deeper indent ----
    {
      code: `var x = (\n  <div>\n    {condition && (\n    <p>Bar</p>\n    )}\n  </div>\n);`,
      output: `var x = (\n  <div>\n    {condition && (\n      <p>Bar</p>\n    )}\n  </div>\n);`,
      options: [2, { indentLogicalExpressions: true }],
      errors: [
        { message: 'Expected indentation of 6 space characters but found 4.' },
      ],
    },
    // ---- Multiline ternary alternate at wrong indent ----
    {
      code: `var x = foo ?\n    <Foo /> :\n<Bar />;`,
      output: `var x = foo ?\n    <Foo /> :\n    <Bar />;`,
      errors: [
        { message: 'Expected indentation of 4 space characters but found 0.' },
      ],
    },
    // ---- React.memo wrap with mis-indented child ----
    {
      code: `var W = React.memo(() => (\n  <App>\n        <Foo />\n  </App>\n));`,
      output: `var W = React.memo(() => (\n  <App>\n    <Foo />\n  </App>\n));`,
      options: [2],
      errors: [
        { message: 'Expected indentation of 4 space characters but found 8.' },
      ],
    },
    // ---- Member expression tag with mis-indented child ----
    {
      code: `var x = (\n  <Foo.Bar>\n        <Foo.Baz />\n  </Foo.Bar>\n);`,
      output: `var x = (\n  <Foo.Bar>\n    <Foo.Baz />\n  </Foo.Bar>\n);`,
      options: [2],
      errors: [
        { message: 'Expected indentation of 4 space characters but found 8.' },
      ],
    },
    // ---- checkAttributes:true — `)}` realigned to attribute name ----
    {
      code: `const x = (\n  <View\n    Foo={(\n      <Inner />\n)}\n  />\n);`,
      output: `const x = (\n  <View\n    Foo={(\n      <Inner />\n    )}\n  />\n);`,
      options: [2, { checkAttributes: true }],
      errors: [
        { message: 'Expected indentation of 4 space characters but found 0.' },
      ],
    },
  ],
});
