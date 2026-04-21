# react/no-deprecated

## Rule Details

Several methods are deprecated between React versions. This rule warns about usage of methods that have been deprecated in the React version configured via `settings.react.version`. When no version is configured, the rule treats the project as "latest" and flags every known deprecation.

The rule detects deprecated APIs through member access, named imports, destructuring from `require()` / module bindings, and lifecycle methods declared inside React components (ES5 `createReactClass` objects and ES6 classes extending `Component` / `PureComponent`).

Examples of **incorrect** code for this rule:

```jsx
React.render(<MyComponent />, root);

React.unmountComponentAtNode(root);

React.findDOMNode(this.refs.foo);

React.renderToString(<MyComponent />);

React.renderToStaticMarkup(<MyComponent />);

React.createClass({ /* Class object */ });

const propTypes = {
  foo: PropTypes.bar,
};

// Any factories under React.DOM
React.DOM.div();

import React, { PropTypes } from 'react';

// old lifecycles (since React 16.9)
class Foo extends React.Component {
  componentWillMount() {}
  componentWillReceiveProps() {}
  componentWillUpdate() {}
}

// React 18 deprecations
import { render } from 'react-dom';
ReactDOM.render(<div></div>, container);

import { hydrate } from 'react-dom';
ReactDOM.hydrate(<div></div>, container);

import { unmountComponentAtNode } from 'react-dom';
ReactDOM.unmountComponentAtNode(container);

import { renderToNodeStream } from 'react-dom/server';
ReactDOMServer.renderToNodeStream(element);
```

Examples of **correct** code for this rule:

```jsx
// when React < 18
ReactDOM.render(<MyComponent />, root);

// when React is < 0.14
ReactDOM.findDOMNode(this.refs.foo);

import { PropTypes } from 'prop-types';

class Foo extends React.Component {
  UNSAFE_componentWillMount() {}
  UNSAFE_componentWillReceiveProps() {}
  UNSAFE_componentWillUpdate() {}
}

ReactDOM.createPortal(child, container);

import { createRoot } from 'react-dom/client';
const root = createRoot(container);
root.unmount();

import { hydrateRoot } from 'react-dom/client';
const root = hydrateRoot(container, <App/>);
```

## Settings

This rule is influenced by the shared React settings:

- `settings.react.version` — determines which deprecations are active (default: latest).
- `settings.react.pragma` — the `React` object name used for `<pragma>.X` deprecations (default: `"React"`). An inline `@jsx` comment overrides this setting for the file.

## Differences from ESLint

rslint flags these forms; ESLint does not:

- `(React).createClass` and any other parenthesized wrap around the receiver.
- `React?.createClass()` and optional-chain forms like `React?.addons?.TestUtils`.

## Original Documentation

- https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-deprecated.md
