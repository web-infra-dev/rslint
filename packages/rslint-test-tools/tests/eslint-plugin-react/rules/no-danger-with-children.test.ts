import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectDangerWithChildrenError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'dangerWithChildren',
    'Only set one of `children` or `props.dangerouslySetInnerHTML`',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-danger-with-children', {
  valid: [
    // ---- Upstream valid cases ----
    { code: `<div>Children</div>;` },
    { code: `const props: any = {}; <div {...props} />;` },
    { code: `<div dangerouslySetInnerHTML={{ __html: "HTML" }} />;` },
    { code: `<div children="Children" />;` },
    {
      code: `
        const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
        const x = <div {...props} />;
      `,
    },
    {
      code: `
        const moreProps = { className: "eslint" };
        const props = { children: "Children", ...moreProps };
        const x = <div {...props} />;
      `,
    },
    {
      code: `
        const otherProps = { children: "Children" };
        const { a, b, ...props } = otherProps as any;
        const x = <div {...props} />;
      `,
    },
    { code: `<Hello>Children</Hello>;` },
    { code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }} />;` },
    {
      code: `
        <Hello dangerouslySetInnerHTML={{ __html: "HTML" }}>
        </Hello>;
      `,
    },
    {
      code: `React.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } });`,
    },
    { code: `React.createElement("div", {}, "Children");` },
    {
      code: `React.createElement("Hello", { dangerouslySetInnerHTML: { __html: "HTML" } });`,
    },
    { code: `React.createElement("Hello", {}, "Children");` },
    { code: `<Hello {...undefined}>Children</Hello>;` },
    { code: `React.createElement("Hello", undefined, "Children");` },
    {
      code: `
        declare const shallow: any;
        declare const TaskEditableTitle: any;
        const props = { ...props, scratch: { mode: 'edit' } } as any;
        const component = shallow(<TaskEditableTitle {...props} />);
      `,
    },

    // ---- Additional edge cases ----
    // Empty element, no props, no children.
    { code: `<div />;` },
    // dangerouslySetInnerHTML is case-sensitive.
    {
      code: `<div dangerouslySetInnerHtml={{ __html: "HTML" }}>Children</div>;`,
    },
    // Bare createElement — upstream only matches `x.createElement(...)`.
    {
      code: `createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`,
    },
    // Computed createElement access — upstream skips via `'name' in property`.
    {
      code: `React["createElement"]("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`,
    },
    // Single argument createElement — nothing to check.
    { code: `React.createElement("div");` },
    // Whitespace-only multi-line children — treated as empty.
    {
      code: `
        <div dangerouslySetInnerHTML={{ __html: "HTML" }}>
        </div>;
      `,
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<div>Children</div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<div {...props} />`,
      languageOptions: {
        ecmaVersion: 2018,
        sourceType: 'module',
        parserOptions: { ecmaFeatures: { jsx: true } },
        globals: { props: true },
      },
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<div dangerouslySetInnerHTML={{ __html: "HTML" }} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<div children="Children" />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
        <div {...props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `
        const moreProps = { className: "eslint" };
        const props = { children: "Children", ...moreProps };
        <div {...props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `
        const otherProps = { children: "Children" };
        const { a, b, ...props } = otherProps;
        <div {...props} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<Hello>Children</Hello>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `
        <Hello dangerouslySetInnerHTML={{ __html: "HTML" }}>
        </Hello>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<Hello {...undefined}>Children</Hello>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `React.createElement("Hello", undefined, "Children")`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `
        const props = {...props, scratch: {mode: 'edit'}};
        const component = shallow(<TaskEditableTitle {...props} />);
      `,
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `
        <div dangerouslySetInnerHTML={{ __html: "HTML" }}>
          Children
        </div>;
      `,
      errors: [expectDangerWithChildrenError([2, 9, 4, 15])],
    },
    {
      code: `<div dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />;`,
      errors: [expectDangerWithChildrenError([1, 1, 1, 73])],
    },
    {
      code: `
        const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
        const x = <div {...props}>Children</div>;
      `,
      errors: [expectDangerWithChildrenError([3, 19, 3, 49])],
    },
    {
      code: `
        const props = { children: "Children", dangerouslySetInnerHTML: { __html: "HTML" } };
        const x = <div {...props} />;
      `,
      errors: [expectDangerWithChildrenError([3, 19, 3, 37])],
    },
    {
      code: `
        <Hello dangerouslySetInnerHTML={{ __html: "HTML" }}>
          Children
        </Hello>;
      `,
      errors: [expectDangerWithChildrenError([2, 9, 4, 17])],
    },
    {
      code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />;`,
      errors: [expectDangerWithChildrenError([1, 1, 1, 75])],
    },
    {
      code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }}> </Hello>;`,
      errors: [expectDangerWithChildrenError([1, 1, 1, 62])],
    },
    {
      code: `
        React.createElement(
          "div",
          { dangerouslySetInnerHTML: { __html: "HTML" } },
          "Children"
        );
      `,
      errors: [expectDangerWithChildrenError([2, 9, 6, 10])],
    },
    {
      code: `
        React.createElement(
          "div",
          {
            dangerouslySetInnerHTML: { __html: "HTML" },
            children: "Children",
          }
        );
      `,
      errors: [expectDangerWithChildrenError([2, 9, 8, 10])],
    },
    {
      code: `
        React.createElement(
          "Hello",
          { dangerouslySetInnerHTML: { __html: "HTML" } },
          "Children"
        );
      `,
      errors: [expectDangerWithChildrenError([2, 9, 6, 10])],
    },
    {
      code: `
        React.createElement(
          "Hello",
          {
            dangerouslySetInnerHTML: { __html: "HTML" },
            children: "Children",
          }
        );
      `,
      errors: [expectDangerWithChildrenError([2, 9, 8, 10])],
    },
    {
      code: `
        const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
        React.createElement("div", props, "Children");
      `,
      errors: [expectDangerWithChildrenError([3, 9, 3, 54])],
    },
    {
      code: `
        const props = { children: "Children", dangerouslySetInnerHTML: { __html: "HTML" } };
        React.createElement("div", props);
      `,
      errors: [expectDangerWithChildrenError([3, 9, 3, 42])],
    },
    {
      code: `
        const moreProps = { children: "Children" };
        const otherProps = { ...moreProps };
        const props = { ...otherProps, dangerouslySetInnerHTML: { __html: "HTML" } };
        React.createElement("div", props);
      `,
      errors: [expectDangerWithChildrenError([5, 9, 5, 42])],
    },

    // ---- Additional edge cases ----
    // JSX expression container child — counts as meaningful children.
    {
      code: `const children: any = "x"; const x = <div dangerouslySetInnerHTML={{ __html: "HTML" }}>{children}</div>;`,
      errors: [expectDangerWithChildrenError([1, 38, 1, 104])],
    },
    // Any `x.createElement(...)` member call, pragma-agnostic.
    {
      code: `declare const createFoo: any; createFoo.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`,
      errors: [expectDangerWithChildrenError([1, 31, 1, 122])],
    },
    // Nested member callee (`lib.h.createElement(...)`).
    {
      code: `declare const lib: { h: { createElement: (...a: any[]) => any } }; lib.h.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`,
      errors: [expectDangerWithChildrenError([1, 68, 1, 155])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <div dangerouslySetInnerHTML={{ __html: "HTML" }}>
          Children
        </div>
      `,
      errors: [expectDangerWithChildrenError([2, 9, 4, 15])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<div dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />`,
      errors: [expectDangerWithChildrenError([1, 1, 1, 73])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
        <div {...props}>Children</div>
      `,
      errors: [expectDangerWithChildrenError([3, 9, 3, 39])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        const props = { children: "Children", dangerouslySetInnerHTML: { __html: "HTML" } };
        <div {...props} />
      `,
      errors: [expectDangerWithChildrenError([3, 9, 3, 27])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        <Hello dangerouslySetInnerHTML={{ __html: "HTML" }}>
          Children
        </Hello>
      `,
      errors: [expectDangerWithChildrenError([2, 9, 4, 17])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />`,
      errors: [expectDangerWithChildrenError([1, 1, 1, 75])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }}> </Hello>`,
      errors: [expectDangerWithChildrenError([1, 1, 1, 62])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
