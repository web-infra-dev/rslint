# import/no-webpack-loader-syntax

## Rule Details

Disallows the use of webpack loader syntax (`!`) in `import` statements and `require()` calls. Webpack allows specifying loaders inline using `!` in the module path (e.g., `css-loader!./styles.css`), but this couples the code to webpack and makes it non-portable to other bundlers or environments. Loader configuration should be specified in the webpack configuration file instead.

Examples of **incorrect** code for this rule:

```javascript
import styles from 'css-loader!./styles.css';

import content from 'html-loader!./template.html';

const styles = require('style-loader!css-loader!./styles.css');
```

Examples of **correct** code for this rule:

```javascript
import styles from './styles.css';

import content from './template.html';

const styles = require('./styles.css');
```

## Original Documentation

- [import/no-webpack-loader-syntax](https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/no-webpack-loader-syntax.md)
