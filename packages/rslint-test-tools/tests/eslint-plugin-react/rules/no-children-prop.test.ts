import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-children-prop', {} as never, {
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
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `<div children />;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `<div children="Children" />;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `<div children={<div />} />;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `<div children={[<div />, <div />]} />;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `<div children="Children">Children</div>;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `React.createElement("div", {children: "Children"});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    {
      code: `React.createElement("div", {children: "Children"}, "Children");`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    {
      code: `React.createElement("div", {children: React.createElement("div")});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    {
      code: `React.createElement("div", {children: [React.createElement("div"), React.createElement("div")]});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    {
      code: `<MyComponent children="Children" />`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `React.createElement(MyComponent, {children: "Children"});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    {
      code: `<MyComponent className="class-name" children="Children" />;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `React.createElement(MyComponent, {children: "Children", className: "class-name"});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    {
      code: `const props: any = {}; const x = <MyComponent {...props} children="Children" />;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `const props: any = {}; React.createElement(MyComponent, {...props, children: "Children"})`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    {
      code: `<MyComponent>{() => {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not nest a function between the opening and closing tags. Instead, pass it as a prop.',
        },
      ],
    },
    {
      code: `<MyComponent>{function() {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not nest a function between the opening and closing tags. Instead, pass it as a prop.',
        },
      ],
    },
    {
      code: `<MyComponent>{async function() {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not nest a function between the opening and closing tags. Instead, pass it as a prop.',
        },
      ],
    },
    {
      code: `<MyComponent>{function* () {}}</MyComponent>;`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not nest a function between the opening and closing tags. Instead, pass it as a prop.',
        },
      ],
    },
    {
      code: `React.createElement(MyComponent, {}, () => {});`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not pass a function as an additional argument to React.createElement. Instead, pass it as a prop.',
        },
      ],
    },
    {
      code: `React.createElement(MyComponent, {}, function() {});`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not pass a function as an additional argument to React.createElement. Instead, pass it as a prop.',
        },
      ],
    },
    {
      code: `React.createElement(MyComponent, {}, async function() {});`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not pass a function as an additional argument to React.createElement. Instead, pass it as a prop.',
        },
      ],
    },
    {
      code: `React.createElement(MyComponent, {}, function* () {});`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not pass a function as an additional argument to React.createElement. Instead, pass it as a prop.',
        },
      ],
    },

    // ---- Additional edge cases ----
    // Shorthand `{children}` is still reported.
    {
      code: `const children = "x"; React.createElement("div", {children});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    // Without allowFunctions, a function children prop still reports.
    {
      code: `<MyComponent children={() => {}} />;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    {
      code: `React.createElement(MyComponent, {children: () => {}});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    // With allowFunctions, a non-function value (variable reference) still reports.
    {
      code: `const fn = () => {}; const x = <MyComponent children={fn} />;`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
    // Parenthesized second arg — still unwrapped and inspected.
    {
      code: `React.createElement("div", ({children: "x"}));`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    // Parenthesized createElement callee — IsCreateElementCall unwraps it.
    {
      code: `(React.createElement)("div", {children: "x"});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    // Nested createElement: only the outer call has a `children` prop.
    {
      code: `React.createElement("div", {children: React.createElement("span", {})});`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, pass them as additional arguments to React.createElement.',
        },
      ],
    },
    // Deeply nested JSX: only the innermost JsxElement with a function child
    // triggers `nestFunction`.
    {
      code: `<Outer><Mid><Inner>{() => {}}</Inner></Mid></Outer>;`,
      options: [{ allowFunctions: true }],
      errors: [
        {
          message:
            'Do not nest a function between the opening and closing tags. Instead, pass it as a prop.',
        },
      ],
    },
    // Member-based user component `<Foo.Bar>` — `children` prop still reports.
    {
      code: `const Foo: any = {}; Foo.Bar = () => null; const x = <Foo.Bar children="x" />;`,
      errors: [
        {
          message:
            'Do not pass children as props. Instead, nest children between the opening and closing tags.',
        },
      ],
    },
  ],
});
