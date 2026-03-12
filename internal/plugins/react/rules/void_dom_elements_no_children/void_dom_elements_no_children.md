# react/void-dom-elements-no-children

## Rule Details

Prevent void DOM elements (e.g. `<img />`, `<br />`, `<hr />`) from receiving children. Void elements are HTML elements that cannot have any content. Passing children or `dangerouslySetInnerHTML` to these elements is a mistake.

Examples of **incorrect** code for this rule:

```jsx
<br>Children</br>
<img alt="content">Children</img>
<img dangerouslySetInnerHTML={{ __html: 'content' }} />
React.createElement('br', {}, 'children')
React.createElement('br', null, 'children')
```

Examples of **correct** code for this rule:

```jsx
<br />
<img alt="content" />
<div>Children</div>
```

## Limitations

- Only detects `React.createElement(...)` calls. Destructured `createElement` (e.g. `import { createElement } from 'react'`) and custom pragma (e.g. `Preact.h`) are not supported.

## Original Documentation

- [react/void-dom-elements-no-children](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/void-dom-elements-no-children.md)
