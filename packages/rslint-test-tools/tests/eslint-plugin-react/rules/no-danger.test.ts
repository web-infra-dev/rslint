import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-danger', {} as never, {
  valid: [
    // ---- Upstream valid cases ----
    { code: `<App />;` },
    { code: `<App dangerouslySetInnerHTML={{ __html: "" }} />;` },
    { code: `<div className="bar"></div>;` },
    {
      code: `<div className="bar"></div>;`,
      options: [{ customComponentNames: ['*'] }],
    },
    {
      code: `
        function App() {
          return <Title dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['Home'] }],
    },
    {
      code: `
        function App() {
          return <TextMUI dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['MUI*'] }],
    },

    // ---- Additional edge cases ----
    // Case-sensitive: lowercase 'h' in 'Html' is not the dangerous attribute.
    { code: `<div dangerouslySetInnerHtml={{ __html: "" }} />;` },
    // Similar-looking but distinct attribute name.
    { code: `<div innerHTML="<span />" />;` },
    // Multiple patterns, none matching the user component.
    {
      code: `<Widget dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Foo*', 'Bar*'] }],
    },
    // Empty customComponentNames disables checks on user components.
    {
      code: `<MyComponent dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: [] }],
    },
    // No options at all is the same as empty customComponentNames.
    { code: `<MyComponent dangerouslySetInnerHTML={{ __html: "" }} />;` },
    // Fragment has no attributes — nothing to flag.
    { code: `<React.Fragment></React.Fragment>;` },
    // Spread without the dangerous prop is fine.
    { code: `const props: any = {}; <div {...props} />;` },
    // `?` matches exactly one character — 4-char suffix doesn't match `???`.
    {
      code: `<TextMUIx dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Text???'] }],
    },
    // Literal `.` — non-matching dotted name.
    {
      code: `<Foo.Bar dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Foo.Baz'] }],
    },
    // UPPERCASE-base member tag is a user component — no pattern, no report.
    {
      code: `
        const Foo: any = {};
        Foo.Bar = () => null;
        function App() {
          return <Foo.Bar dangerouslySetInnerHTML={{ __html: "" }} />;
        }
      `,
    },
    // Defensive: non-array customComponentNames is ignored silently.
    {
      code: `<MyComponent dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: 'MyComponent' }],
    },
    // Defensive: non-string entries in the array are skipped.
    {
      code: `<MyComponent dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: [42, null, false] }],
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `<div dangerouslySetInnerHTML={{ __html: "" }}></div>;`,
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    {
      code: `<App dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;`,
      options: [{ customComponentNames: ['*'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    {
      code: `
        function App() {
          return <Title dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['Title'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    {
      code: `
        function App() {
          return <TextFoo dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['*Foo'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    {
      code: `
        function App() {
          return <FooText dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['Foo*'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    {
      code: `
        function App() {
          return <TextMUI dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['*MUI'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    {
      code: `
        const Comp = "div";
        const Component = () => null;
        function App() {
          return (
            <>
              <div dangerouslySetInnerHTML={{ __html: "<div>aaa</div>" }} />
              <Comp dangerouslySetInnerHTML={{ __html: "<div>aaa</div>" }} />

              <Component.NestedComponent
                dangerouslySetInnerHTML={{ __html: '<div>aaa</div>' }}
              />
            </>
          );
        }
      `,
      options: [{ customComponentNames: ['*'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },

    // ---- Additional edge cases ----

    // Boolean-shorthand dangerous prop still reports.
    {
      code: `<div dangerouslySetInnerHTML />;`,
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // Attribute on its own line.
    {
      code: `<div\n\tdangerouslySetInnerHTML={{ __html: '' }}\n/>;`,
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // Deeply-nested member tag matched via '*'.
    {
      code: `
        const A: any = {};
        A.B = { C: () => null };
        function App() {
          return <A.B.C dangerouslySetInnerHTML={{ __html: "" }} />;
        }
      `,
      options: [{ customComponentNames: ['*'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // One matching pattern is enough in a multi-pattern list.
    {
      code: `<Widget dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Foo*', 'Widget', 'Bar*'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // Spread attribute co-located with the dangerous prop — still detects.
    {
      code: `const props: any = {}; const x = <div {...props} dangerouslySetInnerHTML={{ __html: "" }} />;`,
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // `?` glob wildcard.
    {
      code: `<TextMUI dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Text???'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // Literal `.` — exact-dot match on a member tag.
    {
      code: `<Foo.Bar dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Foo.Bar'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // Generic component `<Foo<T> ...>`.
    {
      code: `
        function Foo<T>(_: { x: T; dangerouslySetInnerHTML?: unknown }) { return null; }
        function App() {
          return <Foo<string> dangerouslySetInnerHTML={{ __html: "" }} />;
        }
      `,
      options: [{ customComponentNames: ['Foo'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // JSX namespaced element `<svg:path>` — intrinsic DOM.
    {
      code: `<svg:path dangerouslySetInnerHTML={{ __html: "" }} />;`,
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // LOWERCASE-base member tag `<foo.bar>` — `isDOMComponent` matches on
    // the first character of elementType, so it's DOM-classified and fires
    // without customComponentNames.
    {
      code: `
        const foo: any = {};
        foo.bar = () => null;
        function App() {
          return <foo.bar dangerouslySetInnerHTML={{ __html: "" }} />;
        }
      `,
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // `<this.Foo>` — elementType returns "this.Foo", first char lowercase →
    // DOM-classified.
    {
      code: `
        class C {
          render() {
            return <this._Widget dangerouslySetInnerHTML={{ __html: "" }} />;
          }
        }
      `,
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // Locks "Differences from ESLint" #1: deep-member tag matched by its
    // full dotted name. eslint-plugin-react would compute "undefined.C" and
    // NOT match here; rslint uses source text so both patterns below fire.
    {
      code: `
        const A: any = {};
        A.B = { C: () => null };
        function App() {
          return <A.B.C dangerouslySetInnerHTML={{ __html: "" }} />;
        }
      `,
      options: [{ customComponentNames: ['A.B.C'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    {
      code: `
        const A: any = {};
        A.B = { C: () => null };
        function App() {
          return <A.B.C dangerouslySetInnerHTML={{ __html: "" }} />;
        }
      `,
      options: [{ customComponentNames: ['A.*'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
    // Locks "Differences from ESLint" #2: namespaced tag matched by a
    // literal `"ns:name"` pattern via the customComponentNames branch.
    {
      code: `<svg:path dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['svg:path'] }],
      errors: [
        { message: "Dangerous property 'dangerouslySetInnerHTML' found" },
      ],
    },
  ],
});
