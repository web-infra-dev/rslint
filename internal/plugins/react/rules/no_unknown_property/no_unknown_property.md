# react/no-unknown-property

Disallow usage of unknown DOM property.

## Rule Details

In JSX most DOM properties and attributes should be camelCased to be consistent
with standard JavaScript style. This can be a possible source of error if you
are used to writing plain HTML. Only `data-*` and `aria-*` attributes use
hyphens and lowercase letters in JSX.

The rule flags four classes of issue on lowercase HTML/DOM tags:

1. **`unknownProp`** — an attribute that does not match any known React DOM
   property / aria / data-\* attribute.
2. **`unknownPropWithStandardName`** — an attribute whose lowercased form
   matches a standard React property name, but whose casing differs (for
   example `class` → `className`, `onmousedown` → `onMouseDown`). Automatically
   fixable.
3. **`invalidPropOnTag`** — an attribute that is recognized, but only valid on
   a specific set of tags (for example `crossOrigin` on `script` / `img` but
   not `div`).
4. **`dataLowercaseRequired`** — a `data-*` attribute that contains uppercase
   characters, reported only when the `requireDataLowercase` option is set.

Examples of **incorrect** code for this rule:

```jsx
var React = require('react');

var Hello = <div class="hello">Hello World</div>;
var Alphabet = <div abc="something">Alphabet</div>;

// Invalid aria-* attribute
var IconButton = <div aria-foo="bar" />;
```

Examples of **correct** code for this rule:

```jsx
var React = require('react');

var Hello = <div className="hello">Hello World</div>;
var Button = <button disabled>Cannot click me</button>;
var Img = <img src={catImage} alt="A cat sleeping on a keyboard" />;

// aria-* attributes
var IconButton = <button aria-label="Close" onClick={this.close}>{closeIcon}</button>;

// data-* attributes
var Data = <div data-index={12}>Some data</div>;

// React components are ignored
var MyComponent = <App class="foo-bar"/>;
var AnotherComponent = <Foo.bar for="bar" />;

// Custom web components are ignored
var MyElem = <div class="foo" is="my-elem"></div>;
var AtomPanel = <atom-panel class="foo"></atom-panel>;
```

## Rule Options

- `ignore`: optional array of property and attribute names to skip during
  validation.
- `requireDataLowercase`: optional boolean (default `false`). When `true`, any
  `data-*` attribute that contains uppercase characters is reported. React
  itself warns about such attributes at runtime; enabling this option catches
  them statically.

Examples of **correct** code with `{ "ignore": ["css"] }`:

```json
{ "react/no-unknown-property": ["error", { "ignore": ["css"] }] }
```

```jsx
var StyledDiv = <div css={{ color: 'pink' }}></div>;
```

Examples of **incorrect** code with `{ "requireDataLowercase": true }`:

```json
{ "react/no-unknown-property": ["error", { "requireDataLowercase": true }] }
```

```jsx
var Row = <div data-testID="row-1" />;
```

## React version detection

Some recognized attributes depend on the configured React version
(`settings.react.version`):

- React `< 16.1` still accepts `allowTransparency`.
- React `>= 16.4` adds the pointer-event handlers (`onPointerDown`,
  `onPointerUp`, …).
- React `>= 19` adds `precedence` for stylesheet support.

When the version is not set, the rule assumes the latest React.

## Original Documentation

- [eslint-plugin-react / no-unknown-property](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-unknown-property.md)
