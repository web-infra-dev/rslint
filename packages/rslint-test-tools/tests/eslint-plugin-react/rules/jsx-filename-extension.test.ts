import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectNoJSXWithExtensionError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'noJSXWithExtension',
    "JSX not allowed in files with extension '.js'",
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-filename-extension', {
  valid: [
    {
      code: `const App = () => <div>Hello</div>`,
      filename: 'src/virtual.tsx',
      options: [{ extensions: ['.tsx'] }],
    },
    {
      code: `const App = () => <div />`,
      filename: 'src/virtual.tsx',
      options: [{ extensions: ['.tsx', '.jsx'] }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      skip: 'Upstream uses the <text> sentinel filename; native file selection requires a supported source extension.',
      code: `module.exports = function MyComponent() { return <div>
<div />
</div>; }`,
      filename: '<text>',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `module.exports = function MyComponent() { return <div>
<div />
</div>; }`,
      filename: 'MyComponent.jsx',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `module.exports = {}`,
      options: [{ allow: 'as-needed' }],
      filename: 'MyComponent.js',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `module.exports = function MyComponent() { return <div>
<div />
</div>; }`,
      options: [{ allow: 'as-needed' }],
      filename: 'MyComponent.jsx',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `module.exports = function MyComponent() { return <div>
<div />
</div>; }`,
      options: [{ extensions: ['.js', '.jsx'] }],
      filename: 'MyComponent.js',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `module.exports = {}`,
      filename: 'notAComponent.js',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      skip: 'Upstream uses the <text> sentinel filename; native file selection requires a supported source extension.',
      code: `module.exports = function MyComponent() { return <>
</>; }`,
      filename: '<text>',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `module.exports = function MyComponent() { return <>
</>; }`,
      filename: 'MyComponent.jsx',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `module.exports = function MyComponent() { return <>
</>; }`,
      options: [{ extensions: ['.js', '.jsx'] }],
      filename: 'MyComponent.js',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `// some initial comment

/* multiline
 * comment
 */`,
      options: [{ allow: 'as-needed' }],
      filename: 'MyComponent.js',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `// some initial comment

/* multiline
 * comment
 */`,
      options: [{ allow: 'as-needed', ignoreFilesWithoutCode: true }],
      filename: 'MyComponent.jsx',
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: ``,
      options: [{ allow: 'as-needed', ignoreFilesWithoutCode: true }],
      filename: 'MyComponent.jsx',
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      // .tsx file has no JSX, but extension is .tsx and allow is "as-needed"
      code: `const x = 1;`,
      filename: 'src/virtual.tsx',
      options: [{ extensions: ['.tsx'], allow: 'as-needed' }],
      errors: [
        diagnostic(
          'extensionOnlyForJSX',
          "Only files containing JSX may use the extension '.tsx'",
          [1, 1, 1, 1],
        ),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `module.exports = function MyComponent() { return <div>
<div />
</div>; }`,
      filename: 'MyComponent.js',
      errors: [expectNoJSXWithExtensionError([1, 50, 3, 7])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `module.exports = {}`,
      options: [{ allow: 'as-needed' }],
      filename: 'MyComponent.jsx',
      errors: [
        diagnostic(
          'extensionOnlyForJSX',
          "Only files containing JSX may use the extension '.jsx'",
          [1, 1, 1, 1],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `module.exports = function MyComponent() { return <div>
<div />
</div>; }`,
      options: [{ allow: 'as-needed' }],
      filename: 'notAComponent.js',
      errors: [expectNoJSXWithExtensionError([1, 50, 3, 7])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `module.exports = function MyComponent() { return <div>
<div />
</div>; }`,
      options: [{ extensions: ['.js'] }],
      filename: 'MyComponent.jsx',
      errors: [
        diagnostic(
          'noJSXWithExtension',
          "JSX not allowed in files with extension '.jsx'",
          [1, 50, 3, 7],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `module.exports = function MyComponent() { return <>
</>; }`,
      filename: 'MyComponent.js',
      errors: [expectNoJSXWithExtensionError([1, 50, 2, 4])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `module.exports = function MyComponent() { return <>
</>; }`,
      options: [{ extensions: ['.js'] }],
      filename: 'MyComponent.jsx',
      errors: [
        diagnostic(
          'noJSXWithExtension',
          "JSX not allowed in files with extension '.jsx'",
          [1, 50, 2, 4],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
