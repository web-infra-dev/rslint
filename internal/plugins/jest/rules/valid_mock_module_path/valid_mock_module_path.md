# valid-mock-module-path

## Rule Details

Disallow mocking of non-existing module paths.

This rule reports `jest.mock()` and `jest.doMock()` calls whose first argument names a module that cannot be found. Only a string literal is checked; a variable, template literal or other expression is ignored.

A path starting with `.` is resolved against the directory of the linted file. It exists when a file or directory is found at that path, or at that path followed by one of the `moduleFileExtensions`. Any other specifier is resolved the way Node's `require.resolve` resolves it: Node built-in modules are always accepted, packages are looked up in `node_modules` directories, and a package's `exports` field is honored with the `require` condition, so a path the package does not export is reported even when the file exists.

A call passing a third argument that is an object literal with a `virtual` property set to a truthy literal, such as `{ virtual: true }`, is not checked, because Jest does not require a virtual module to exist.

Examples of **incorrect** code for this rule:

```javascript
// Module(s) that cannot be found
jest.mock('@org/some-module-not-in-package-json');
jest.mock('some-module-not-in-package-json');

// Local module (directory) that cannot be found
jest.mock('../../this/module/does/not/exist');

// Local file that cannot be found
jest.mock('../../this/path/does/not/exist.js');

// Local file that cannot be found and is NOT virtual
jest.mock('../../this/path/does/not/exist.js', undefined, { virtual: false });
```

Examples of **correct** code for this rule:

```javascript
// Module(s) that can be found
jest.mock('@org/some-module-in-package-json');
jest.mock('some-module-in-package-json');

// Local module that can be found
jest.mock('../../this/module/really/does/exist');

// Local file that can be found
jest.mock('../../this/path/really/does/exist.js');

// Module(s) that cannot be found but are configured as virtual
jest.mock('@org/some-module-not-in-package-json', undefined, { virtual: true });
jest.mock('some-module-not-in-package-json', undefined, { virtual: true });

// Local file that cannot be found but is configured as virtual
jest.mock('../../this/path/does/not/exist.js', undefined, { virtual: true });
```

## Options

```json
{
  "jest/valid-mock-module-path": [
    "error",
    {
      "moduleFileExtensions": [".js", ".ts", ".jsx", ".tsx", ".json"]
    }
  ]
}
```

### `moduleFileExtensions`

The file extensions tried after a local path that does not exist as written. The default extensions are:

- `".js"`
- `".ts"`
- `".jsx"`
- `".tsx"`
- `".json"`

Each extension is appended to the path as written, so a custom extension **must** include its leading dot. This option does not affect package specifiers.

## Differences from ESLint

- Package specifiers are looked up in `node_modules` directories starting from the linted file's directory, the way Jest looks them up when running the test. eslint-plugin-jest looks them up from its own install location instead. The results differ only when a package is installed where one location can see it and the other cannot, for example in a nested package of a monorepo.
- For the same reason, a `#` subpath import such as `jest.mock('#utils')` is resolved through the `imports` field of the `package.json` that owns the linted file, and is accepted when that field maps it to an existing file. eslint-plugin-jest reports every `#` import.

## When Not To Use It

Don't use this rule on non-jest test files.

## Original Documentation

- [eslint-plugin-jest: valid-mock-module-path](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/docs/rules/valid-mock-module-path.md)
- [Source code](https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/valid-mock-module-path.ts)
