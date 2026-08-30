# jsx-no-useless-fragment

## Rule Details

Disallow unnecessary fragments. A fragment is redundant if it contains only one
child, or if it is the child of an HTML element, and is not a
[keyed fragment](https://react.dev/reference/react/Fragment#rendering-a-list-of-fragments).

Both fragment spellings are checked: the `<>…</>` shorthand and the long form
named by the `react` / `fragment` shared settings (`<Fragment>` and
`<React.Fragment>` by default). A `@jsx` annotation in the file — such as
`/** @jsx Preact.h */` — renames the object half of the long form for that file
and takes precedence over `settings.react.pragma`.

Examples of **incorrect** code for this rule:

```jsx
<>{foo}</>

<><Foo /></>

<p><>foo</></p>

<></>

<Fragment>foo</Fragment>

<React.Fragment>foo</React.Fragment>

<section>
  <>
    <div />
    <div />
  </>
</section>

{showFullName ? <>{fullName}</> : <>{firstName}</>}
```

Examples of **correct** code for this rule:

```jsx
{foo}

<Foo />

<>
  <Foo />
  <Bar />
</>

<>foo {bar}</>

<> {foo}</>

const cat = <>meow</>

<SomeComponent>
  <>
    <div />
    <div />
  </>
</SomeComponent>

<Fragment key={item.id}>{item.value}</Fragment>

{showFullName ? fullName : firstName}
```

## Options

- `allowExpressions` (default: `false`): when `true`, a fragment wrapping a
  single expression is allowed. This is useful in TypeScript, where `string`
  does not satisfy the expected `JSX.Element` return type and wrapping the
  value in a fragment is a common workaround.

Examples of **correct** code when `allowExpressions` is `true`:

```jsx
<>{foo}</>

<>
  {foo}
</>
```

Note that `allowExpressions` only relaxes the "needs more children" check — a
fragment passed to an HTML element is still reported.

## Original Documentation

- [eslint-plugin-react: jsx-no-useless-fragment](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/docs/rules/jsx-no-useless-fragment.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/lib/rules/jsx-no-useless-fragment.js)
