import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectDangerousPropError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'dangerousProp',
    "Dangerous property 'dangerouslySetInnerHTML' found",
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-danger', {
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
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `<div dangerouslySetInnerHTML={{ __html: "" }}></div>;`,
      errors: [expectDangerousPropError([1, 6, 1, 46])],
    },
    {
      code: `<App dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;`,
      options: [{ customComponentNames: ['*'] }],
      errors: [expectDangerousPropError([1, 6, 1, 64])],
    },
    {
      code: `
        function App() {
          return <Title dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['Title'] }],
      errors: [expectDangerousPropError([3, 25, 3, 83])],
    },
    {
      code: `
        function App() {
          return <TextFoo dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['*Foo'] }],
      errors: [expectDangerousPropError([3, 27, 3, 85])],
    },
    {
      code: `
        function App() {
          return <FooText dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['Foo*'] }],
      errors: [expectDangerousPropError([3, 27, 3, 85])],
    },
    {
      code: `
        function App() {
          return <TextMUI dangerouslySetInnerHTML={{ __html: "<span>hello</span>" }} />;
        }
      `,
      options: [{ customComponentNames: ['*MUI'] }],
      errors: [expectDangerousPropError([3, 27, 3, 85])],
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
        expectDangerousPropError([7, 20, 7, 74]),
        expectDangerousPropError([8, 21, 8, 75]),
        expectDangerousPropError([11, 17, 11, 71]),
      ],
    },

    // ---- Additional edge cases ----

    // Boolean-shorthand dangerous prop still reports.
    {
      code: `<div dangerouslySetInnerHTML />;`,
      errors: [expectDangerousPropError([1, 6, 1, 29])],
    },
    // Attribute on its own line.
    {
      code: `<div\n\tdangerouslySetInnerHTML={{ __html: '' }}\n/>;`,
      errors: [expectDangerousPropError([2, 2, 2, 42])],
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
      errors: [expectDangerousPropError([5, 25, 5, 65])],
    },
    // One matching pattern is enough in a multi-pattern list.
    {
      code: `<Widget dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Foo*', 'Widget', 'Bar*'] }],
      errors: [expectDangerousPropError([1, 9, 1, 49])],
    },
    // Spread attribute co-located with the dangerous prop — still detects.
    {
      code: `const props: any = {}; const x = <div {...props} dangerouslySetInnerHTML={{ __html: "" }} />;`,
      errors: [expectDangerousPropError([1, 50, 1, 90])],
    },
    // `?` glob wildcard.
    {
      code: `<TextMUI dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Text???'] }],
      errors: [expectDangerousPropError([1, 10, 1, 50])],
    },
    // Literal `.` — exact-dot match on a member tag.
    {
      code: `<Foo.Bar dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['Foo.Bar'] }],
      errors: [expectDangerousPropError([1, 10, 1, 50])],
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
      errors: [expectDangerousPropError([4, 31, 4, 71])],
    },
    // JSX namespaced element `<svg:path>` — intrinsic DOM.
    {
      code: `<svg:path dangerouslySetInnerHTML={{ __html: "" }} />;`,
      errors: [expectDangerousPropError([1, 11, 1, 51])],
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
      errors: [expectDangerousPropError([5, 27, 5, 67])],
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
      errors: [expectDangerousPropError([4, 34, 4, 74])],
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
      errors: [expectDangerousPropError([5, 25, 5, 65])],
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
      errors: [expectDangerousPropError([5, 25, 5, 65])],
    },
    // Locks "Differences from ESLint" #2: namespaced tag matched by a
    // literal `"ns:name"` pattern via the customComponentNames branch.
    {
      code: `<svg:path dangerouslySetInnerHTML={{ __html: "" }} />;`,
      options: [{ customComponentNames: ['svg:path'] }],
      errors: [expectDangerousPropError([1, 11, 1, 51])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        import type { ComponentProps } from "react";

        const Comp = "div";
        const Component = () => <></>;

        const NestedComponent = (_props: ComponentProps<"div">) => <></>;

        Component.NestedComponent = NestedComponent;

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
        expectDangerousPropError([14, 20, 14, 74]),
        expectDangerousPropError([15, 21, 15, 75]),
        expectDangerousPropError([18, 17, 18, 71]),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
