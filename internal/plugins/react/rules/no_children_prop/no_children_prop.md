# no-children-prop

## Rule Details

Children should always be actual children, not passed in as a prop.

When using JSX, the `children` should be nested between the opening and closing tags. When not using JSX, the `children` should be passed as additional arguments to `React.createElement`.

Examples of **incorrect** code for this rule:

```jsx
<div children='Children' />

<MyComponent children={<AnotherComponent />} />
<MyComponent children={['Child 1', 'Child 2']} />

React.createElement("div", { children: 'Children' })
```

Examples of **correct** code for this rule:

```jsx
<div>Children</div>

<MyComponent>Children</MyComponent>

<MyComponent>
  <span>Child 1</span>
  <span>Child 2</span>
</MyComponent>

React.createElement("div", {}, 'Children')
React.createElement("div", 'Child 1', 'Child 2')
```

## Rule Options

```json
{ "react/no-children-prop": ["error", { "allowFunctions": false }] }
```

### allowFunctions

Default: `false`.

When enabled, function children are required to be passed as the `children` prop rather than nested between tags or passed as an additional `React.createElement` argument. This can be useful for libraries that rely on render-callbacks.

Examples of **correct** code for this rule with `{ "allowFunctions": true }`:

```json
{ "react/no-children-prop": ["error", { "allowFunctions": true }] }
```

```jsx
<MyComponent children={() => <div />} />
React.createElement(MyComponent, { children: () => <div /> })
```

Examples of **incorrect** code for this rule with `{ "allowFunctions": true }`:

```json
{ "react/no-children-prop": ["error", { "allowFunctions": true }] }
```

```jsx
<MyComponent>{() => <div />}</MyComponent>
React.createElement(MyComponent, {}, () => <div />)
```

## Original Documentation

- [eslint-plugin-react/no-children-prop](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-children-prop.md)
