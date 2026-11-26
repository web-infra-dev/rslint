import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectNestChildrenError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'nestChildren',
    message:
      'Do not pass children as props. Instead, nest children between the opening and closing tags.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectPassChildrenAsArgsError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'passChildrenAsArgs',
    message:
      'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectNestFunctionError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'nestFunction',
    message:
      'Do not nest a function between the opening and closing tags. Instead, pass it as a prop.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectPassFunctionAsArgsError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'passFunctionAsArgs',
    message:
      'Do not pass a function as an additional argument to React.createElement. Instead, pass it as a prop.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-children-prop', {
  valid: [
    // ---- Upstream valid cases ----
    { code: `<div />;` },
    { code: `<div></div>;` },
    { code: `React.createElement("div", {});` },
    { code: `React.createElement("div", undefined);` },
    { code: `<div className="class-name"></div>;` },
    { code: `React.createElement("div", {className: "class-name"});` },
    { code: `<div>Children</div>;` },
    { code: `React.createElement("div", "Children");` },
    { code: `React.createElement("div", {}, "Children");` },
    { code: `React.createElement("div", undefined, "Children");` },
    { code: `<div className="class-name">Children</div>;` },
    {
      code: `React.createElement("div", {className: "class-name"}, "Children");`,
    },
    { code: `<div><div /></div>;` },
    { code: `React.createElement("div", React.createElement("div"));` },
    { code: `React.createElement("div", {}, React.createElement("div"));` },
    {
      code: `React.createElement("div", undefined, React.createElement("div"));`,
    },
    { code: `<div><div /><div /></div>;` },
    {
      code: `React.createElement("div", React.createElement("div"), React.createElement("div"));`,
    },
    {
      code: `React.createElement("div", {}, React.createElement("div"), React.createElement("div"));`,
    },
    {
      code: `React.createElement("div", undefined, React.createElement("div"), React.createElement("div"));`,
    },
    {
      code: `React.createElement("div", [React.createElement("div"), React.createElement("div")]);`,
    },
    {
      code: `React.createElement("div", {}, [React.createElement("div"), React.createElement("div")]);`,
    },
    {
      code: `React.createElement("div", undefined, [React.createElement("div"), React.createElement("div")]);`,
    },
    { code: `<MyComponent />` },
    { code: `React.createElement(MyComponent);` },
    { code: `React.createElement(MyComponent, {});` },
    { code: `React.createElement(MyComponent, undefined);` },
    { code: `<MyComponent>Children</MyComponent>;` },
    { code: `React.createElement(MyComponent, "Children");` },
    { code: `React.createElement(MyComponent, {}, "Children");` },
    { code: `React.createElement(MyComponent, undefined, "Children");` },
    { code: `<MyComponent className="class-name"></MyComponent>;` },
    { code: `React.createElement(MyComponent, {className: "class-name"});` },
    { code: `<MyComponent className="class-name">Children</MyComponent>;` },
    {
      code: `React.createElement(MyComponent, {className: "class-name"}, "Children");`,
    },
    {
      code: `const props: any = {}; <MyComponent className="class-name" {...props} />;`,
    },
    {
      code: `const props: any = {}; React.createElement(MyComponent, {className: "class-name", ...props});`,
    },

    // ---- allowFunctions ----
    {
      code: `<MyComponent children={() => {}} />;`,
      options: [{ allowFunctions: true }],
    },
    {
      code: `<MyComponent children={function() {}} />;`,
      options: [{ allowFunctions: true }],
    },
    {
      code: `<MyComponent children={async function() {}} />;`,
      options: [{ allowFunctions: true }],
    },
    {
      code: `<MyComponent children={function* () {}} />;`,
      options: [{ allowFunctions: true }],
    },
    {
      code: `React.createElement(MyComponent, {children: () => {}});`,
      options: [{ allowFunctions: true }],
    },
    {
      code: `React.createElement(MyComponent, {children: function() {}});`,
      options: [{ allowFunctions: true }],
    },
    {
      code: `React.createElement(MyComponent, {children: async function() {}});`,
      options: [{ allowFunctions: true }],
    },
    {
      code: `React.createElement(MyComponent, {children: function* () {}});`,
      options: [{ allowFunctions: true }],
    },

    // ---- Additional edge cases ----
    // Without allowFunctions, a function as JSX child is fine.
    { code: `<MyComponent>{() => {}}</MyComponent>;` },
    // Without allowFunctions, a 3rd-arg function is also fine.
    { code: `React.createElement(MyComponent, {}, () => {});` },
    // Non-React createElement is ignored.
    {
      code: `const someOther: any = {}; someOther.createElement("div", {children: "x"});`,
    },
    // allowFunctions + non-function JSX child expression — no nestFunction.
    {
      code: `const someValue: any = null; <MyComponent>{someValue}</MyComponent>;`,
      options: [{ allowFunctions: true }],
    },
    // String-keyed `"children"` is NOT matched (upstream `'name' in prop.key`).
    { code: `React.createElement("div", {"children": "x"});` },
    // JSX Fragment child is ignored (only JsxElement is listened).
    {
      code: `<>{() => {}}</>;`,
      options: [{ allowFunctions: true }],
    },
    // JsxElement with more than one child never fires `nestFunction`.
    {
      code: `<MyComponent>{() => {}}{() => {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
    },
    // Paren-wrapped function inside `children={...}` is allowed with allowFunctions.
    {
      code: `<MyComponent children={(() => {})} />;`,
      options: [{ allowFunctions: true }],
    },
    // Paren-wrapped function in createElement children prop is allowed.
    {
      code: `React.createElement(MyComponent, {children: (() => {})});`,
      options: [{ allowFunctions: true }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `<MyComponent className="class-name" {...props} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `React.createElement(MyComponent, {className: "class-name", ...props});`,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `<div children />;`,
      errors: [expectNestChildrenError([1, 6, 1, 14])],
    },
    {
      code: `<div children="Children" />;`,
      errors: [expectNestChildrenError([1, 6, 1, 25])],
    },
    {
      code: `<div children={<div />} />;`,
      errors: [expectNestChildrenError([1, 6, 1, 24])],
    },
    {
      code: `<div children={[<div />, <div />]} />;`,
      errors: [expectNestChildrenError([1, 6, 1, 35])],
    },
    {
      code: `<div children="Children">Children</div>;`,
      errors: [expectNestChildrenError([1, 6, 1, 25])],
    },
    {
      code: `React.createElement("div", {children: "Children"});`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 51])],
    },
    {
      code: `React.createElement("div", {children: "Children"}, "Children");`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 63])],
    },
    {
      code: `React.createElement("div", {children: React.createElement("div")});`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 67])],
    },
    {
      code: `React.createElement("div", {children: [React.createElement("div"), React.createElement("div")]});`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 97])],
    },
    {
      code: `<MyComponent children="Children" />`,
      errors: [expectNestChildrenError([1, 14, 1, 33])],
    },
    {
      code: `React.createElement(MyComponent, {children: "Children"});`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 57])],
    },
    {
      code: `<MyComponent className="class-name" children="Children" />;`,
      errors: [expectNestChildrenError([1, 37, 1, 56])],
    },
    {
      code: `React.createElement(MyComponent, {children: "Children", className: "class-name"});`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 82])],
    },
    {
      code: `const props: any = {}; const x = <MyComponent {...props} children="Children" />;`,
      errors: [expectNestChildrenError([1, 58, 1, 77])],
    },
    {
      code: `const props: any = {}; React.createElement(MyComponent, {...props, children: "Children"})`,
      errors: [expectPassChildrenAsArgsError([1, 24, 1, 90])],
    },
    {
      code: `<MyComponent>{() => {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
      errors: [expectNestFunctionError([1, 1, 1, 38])],
    },
    {
      code: `<MyComponent>{function() {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
      errors: [expectNestFunctionError([1, 1, 1, 43])],
    },
    {
      code: `<MyComponent>{async function() {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
      errors: [expectNestFunctionError([1, 1, 1, 49])],
    },
    {
      code: `<MyComponent>{function* () {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
      errors: [expectNestFunctionError([1, 1, 1, 45])],
    },
    {
      code: `React.createElement(MyComponent, {}, () => {});`,
      options: [{ allowFunctions: true }],
      errors: [expectPassFunctionAsArgsError([1, 1, 1, 47])],
    },
    {
      code: `React.createElement(MyComponent, {}, function() {});`,
      options: [{ allowFunctions: true }],
      errors: [expectPassFunctionAsArgsError([1, 1, 1, 52])],
    },
    {
      code: `React.createElement(MyComponent, {}, async function() {});`,
      options: [{ allowFunctions: true }],
      errors: [expectPassFunctionAsArgsError([1, 1, 1, 58])],
    },
    {
      code: `React.createElement(MyComponent, {}, function* () {});`,
      options: [{ allowFunctions: true }],
      errors: [expectPassFunctionAsArgsError([1, 1, 1, 54])],
    },

    // ---- Additional edge cases ----
    // Shorthand `{children}` is still reported.
    {
      code: `const children = "x"; React.createElement("div", {children});`,
      errors: [expectPassChildrenAsArgsError([1, 23, 1, 61])],
    },
    // Without allowFunctions, a function children prop still reports.
    {
      code: `<MyComponent children={() => {}} />;`,
      errors: [expectNestChildrenError([1, 14, 1, 33])],
    },
    {
      code: `React.createElement(MyComponent, {children: () => {}});`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 55])],
    },
    // With allowFunctions, a non-function value (variable reference) still reports.
    {
      code: `const fn = () => {}; const x = <MyComponent children={fn} />;`,
      options: [{ allowFunctions: true }],
      errors: [expectNestChildrenError([1, 45, 1, 58])],
    },
    // Parenthesized second arg — still unwrapped and inspected.
    {
      code: `React.createElement("div", ({children: "x"}));`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 46])],
    },
    // Parenthesized createElement callee — IsCreateElementCall unwraps it.
    {
      code: `(React.createElement)("div", {children: "x"});`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 46])],
    },
    // Nested createElement: only the outer call has a `children` prop.
    {
      code: `React.createElement("div", {children: React.createElement("span", {})});`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 72])],
    },
    // Deeply nested JSX: only the innermost JsxElement with a function child
    // triggers `nestFunction`.
    {
      code: `<Outer><Mid><Inner>{() => {}}</Inner></Mid></Outer>;`,
      options: [{ allowFunctions: true }],
      errors: [expectNestFunctionError([1, 13, 1, 38])],
    },
    // Member-based user component `<Foo.Bar>` — `children` prop still reports.
    {
      code: `const Foo: any = {}; Foo.Bar = () => null; const x = <Foo.Bar children="x" />;`,
      errors: [expectNestChildrenError([1, 63, 1, 75])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `<MyComponent {...props} children="Children" />;`,
      errors: [expectNestChildrenError([1, 25, 1, 44])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `React.createElement(MyComponent, {...props, children: "Children"})`,
      errors: [expectPassChildrenAsArgsError([1, 1, 1, 67])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
