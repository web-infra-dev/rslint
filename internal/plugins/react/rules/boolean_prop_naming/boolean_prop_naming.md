# boolean-prop-naming

Enforces consistent naming for boolean React props.

## Rule Details

This rule checks declared React component prop types and reports any
boolean-typed prop whose name does not match a configurable regex (default
`^(is|has)[A-Z]([A-Za-z0-9]?)+`). It looks at:

- `static propTypes = {...}` declared on a class that extends
  `React.Component` / `React.PureComponent` / `Component` / `PureComponent`
- `Foo.propTypes = {...}` assignments where `Foo` is a known React component
- `createReactClass({propTypes: {...}})` and `React.createClass(...)` literal
  arguments
- TypeScript type annotations on the first parameter of a functional
  component (`(props: Props) => ...` or `({a, b}: {a: boolean}) => ...`)
- TypeScript type arguments on the variable's annotation
  (`const Hello: React.FC<Props> = (props) => ...`), including
  intersection / union compositions

The rule is a **no-op** when no `rule` regex is configured. Add a `rule`
option to turn it on.

Examples of **incorrect** code for this rule:

```json
{ "react/boolean-prop-naming": ["error", { "rule": "^is[A-Z]([A-Za-z0-9]?)+" }] }
```

```javascript
class Hello extends React.Component {
  static propTypes = { something: PropTypes.bool };
  render() { return <div />; }
}
```

```json
{ "react/boolean-prop-naming": ["error", { "rule": "^is[A-Z]([A-Za-z0-9]?)+" }] }
```

```javascript
type Props = { enabled: boolean };
const Hello = (props: Props) => <div />;
```

Examples of **correct** code for this rule:

```json
{ "react/boolean-prop-naming": ["error", { "rule": "^is[A-Z]([A-Za-z0-9]?)+" }] }
```

```javascript
class Hello extends React.Component {
  static propTypes = { isSomething: PropTypes.bool };
  render() { return <div />; }
}
```

```json
{ "react/boolean-prop-naming": ["error", { "rule": "^is[A-Z]([A-Za-z0-9]?)+" }] }
```

```javascript
type Props = { isEnabled: boolean };
const Hello: React.FC<Props> = (props) => <div />;
```

## Options

```json
{
  "react/boolean-prop-naming": ["error", {
    "rule": "^(is|has)[A-Z]([A-Za-z0-9]?)+",
    "propTypeNames": ["bool"],
    "message": "Boolean prop names must begin with `is` or `has`",
    "validateNested": false
  }]
}
```

- `rule` (string) — regex source the prop name must match. **Required to
  enable the rule** (omitted / empty → no-op). Default in upstream docs is
  `^(is|has)[A-Z]([A-Za-z0-9]?)+` but the rule is not turned on unless you
  pass it explicitly.
- `propTypeNames` (string[]) — names treated as boolean PropTypes. Default
  `["bool"]`. Add custom validators (`["bool", "mutuallyExclusiveTrueProps"]`)
  to recognize them as boolean.
- `message` (string) — custom error message; supports `{{ propName }}` and
  `{{ pattern }}` placeholders.
- `validateNested` (boolean, default `false`) — when `true`, recurse into the
  first argument of any call inside a prop value (e.g. `nested:
  PropTypes.shape({ failingItIs: PropTypes.bool })`).

## Settings

`settings.propWrapperFunctions` — array of identifiers (or
`{ object, property }` pairs) that wrap propTypes objects, so the rule can
inspect inside them. Examples:

```json
{
  "settings": {
    "propWrapperFunctions": [
      "forbidExtraProps",
      "merge",
      { "object": "Object", "property": "assign" },
      { "object": "_", "property": "assign" }
    ]
  }
}
```

## Differences from ESLint

- Flow type annotations (`type Props = { x: boolean }` with
  `BooleanTypeAnnotation`, `ObjectTypeProperty`, etc.) are not validated —
  rslint parses these as TypeScript, so only the equivalent TypeScript shapes
  (`TypeLiteralNode`, `TSPropertySignature`, `BooleanKeyword`) are checked.
  TypeScript-flavored equivalents of every Flow case in the upstream test
  suite report identically.

## Original Documentation

- [eslint-plugin-react / boolean-prop-naming](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/boolean-prop-naming.md)
- [Source code](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/lib/rules/boolean-prop-naming.js)
