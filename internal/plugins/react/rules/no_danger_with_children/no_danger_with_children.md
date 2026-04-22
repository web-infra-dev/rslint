# no-danger-with-children

Disallow when a DOM element is using both `children` and `dangerouslySetInnerHTML`.

## Rule Details

This rule helps prevent problems caused by using `children` and
`dangerouslySetInnerHTML` at the same time. React will throw a warning if this
rule is ignored.

Examples of **incorrect** code for this rule:

```javascript
<div dangerouslySetInnerHTML={{ __html: "HTML" }}>Children</div>;

<Hello dangerouslySetInnerHTML={{ __html: "HTML" }}>Children</Hello>;

<div dangerouslySetInnerHTML={{ __html: "HTML" }} children="Children" />;

React.createElement(
  "div",
  { dangerouslySetInnerHTML: { __html: "HTML" } },
  "Children"
);

React.createElement("div", {
  dangerouslySetInnerHTML: { __html: "HTML" },
  children: "Children",
});
```

Examples of **correct** code for this rule:

```javascript
<div>Children</div>;

<div dangerouslySetInnerHTML={{ __html: "HTML" }} />;

React.createElement("div", { dangerouslySetInnerHTML: { __html: "HTML" } });

React.createElement("div", {}, "Children");
```

## Original Documentation

- https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-danger-with-children.md
