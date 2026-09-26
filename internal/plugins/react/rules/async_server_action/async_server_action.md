# async-server-action

Require functions whose first statement is `"use server"` to be async.

## Rule Details

React Server Actions must be async even when they do not await anything or return a promise. This rule checks function declarations, function expressions, arrow functions, and methods. It offers an editor suggestion to insert `async`; it does not apply an automatic fix.

Examples of **incorrect** code for this rule:

```jsx
function action() {
  'use server';
  save();
}
```

```jsx
<form
  action={() => {
    'use server';
    save();
  }}
>
  Save
</form>;
```

Examples of **correct** code for this rule:

```jsx
async function action() {
  'use server';
  await save();
}
```

```jsx
<form
  action={async () => {
    'use server';
    await save();
  }}
>
  Save
</form>;
```

Only the first statement is checked. A preceding statement, including another directive such as `"use strict"`, prevents a match. Generators and template literals are ignored. A file-level `"use server"` directive does not cause this rule to check all exported functions.

## Options

This rule has no options.

## Known Limitations

Editor suggestions for computed methods, constructors, and accessors can produce invalid syntax. For example, the suggestion changes `[action]() { 'use server'; }` to `[async action]() { 'use server'; }`. Instead, place `async` before the opening bracket: `async [action]() { 'use server'; }`.

Constructors, getters, and setters cannot be async. Move the Server Action into an ordinary async function or method before applying a suggestion.

## When Not To Use It

If you are not using React Server Components.

## Original Documentation

This rule follows the upstream commit linked below; it is not included in upstream release `v7.37.5`.

- [React: use server](https://react.dev/reference/react/use-server)
- [eslint-plugin-react: async-server-action](https://github.com/jsx-eslint/eslint-plugin-react/blob/5d4bf128fb506f604f1bb61bcf76fa8caeea6a1a/docs/rules/async-server-action.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/5d4bf128fb506f604f1bb61bcf76fa8caeea6a1a/lib/rules/async-server-action.js)
