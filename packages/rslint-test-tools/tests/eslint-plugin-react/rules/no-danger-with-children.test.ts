import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-danger-with-children', {} as never, {
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
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: `
        <div dangerouslySetInnerHTML={{ __html: "HTML" }}>
          Children
        </div>;
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `<div dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />;`,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `
        const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
        const x = <div {...props}>Children</div>;
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `
        const props = { children: "Children", dangerouslySetInnerHTML: { __html: "HTML" } };
        const x = <div {...props} />;
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `
        <Hello dangerouslySetInnerHTML={{ __html: "HTML" }}>
          Children
        </Hello>;
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />;`,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `<Hello dangerouslySetInnerHTML={{ __html: "HTML" }}> </Hello>;`,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `
        React.createElement(
          "div",
          { dangerouslySetInnerHTML: { __html: "HTML" } },
          "Children"
        );
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
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
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `
        React.createElement(
          "Hello",
          { dangerouslySetInnerHTML: { __html: "HTML" } },
          "Children"
        );
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
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
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `
        const props = { dangerouslySetInnerHTML: { __html: "HTML" } };
        React.createElement("div", props, "Children");
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `
        const props = { children: "Children", dangerouslySetInnerHTML: { __html: "HTML" } };
        React.createElement("div", props);
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    {
      code: `
        const moreProps = { children: "Children" };
        const otherProps = { ...moreProps };
        const props = { ...otherProps, dangerouslySetInnerHTML: { __html: "HTML" } };
        React.createElement("div", props);
      `,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },

    // ---- Additional edge cases ----
    // JSX expression container child — counts as meaningful children.
    {
      code: `const children: any = "x"; const x = <div dangerouslySetInnerHTML={{ __html: "HTML" }}>{children}</div>;`,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    // Any `x.createElement(...)` member call, pragma-agnostic.
    {
      code: `declare const createFoo: any; createFoo.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
    // Nested member callee (`lib.h.createElement(...)`).
    {
      code: `declare const lib: { h: { createElement: (...a: any[]) => any } }; lib.h.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } }, "Children");`,
      errors: [
        {
          message:
            'Only set one of `children` or `props.dangerouslySetInnerHTML`',
        },
      ],
    },
  ],
});
