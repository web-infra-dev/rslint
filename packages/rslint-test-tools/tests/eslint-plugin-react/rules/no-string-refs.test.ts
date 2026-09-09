import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-string-refs', {} as never, {
  valid: [
    // Callback ref is fine.
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={c => this.hello = c}>Hello</div>;
          }
        });
      `,
    },
    // Template literal without noTemplateLiterals is fine.
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={\`hello\`}>Hello</div>;
          }
        });
      `,
    },
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={\`hello\${index}\`}>Hello</div>;
          }
        });
      `,
    },
    // Non-ref attributes with string literals are fine.
    { code: `<div title="hello" />;` },
    // `ref` without an initializer (boolean shorthand) is not flagged.
    { code: `<div ref />;` },
    // An identifier in a ref is fine (it's a variable, not a string).
    { code: `const myRef = () => {}; <div ref={myRef} />;` },
    // TypeScript `as` / `!` / `satisfies` wrappers on the expression mean
    // `expression.type !== 'Literal'` in upstream ESTree, so these are NOT
    // reported. Locks alignment with eslint-plugin-react.
    { code: `<div ref={'hello' as string} />;` },
    { code: `<div ref={'hello'!} />;` },
    { code: `<div ref={('hello' as string)} />;` },
    { code: `<div ref={'hello' satisfies string} />;` },
    // Intentional divergences from upstream: computed identifier and private
    // names do not necessarily refer to React's public `refs` property.
    {
      code: `class App extends React.Component { method() { const refs = 'not-react-refs'; return this[refs]; } }`,
      settings: { react: { version: '18.2.0' } },
    },
    {
      code: `class App extends React.Component { #refs; method() { return this.#refs; } }`,
      settings: { react: { version: '18.2.0' } },
    },
    {
      code: `/** @jsx Preact.h */ class App extends Other.Component { method() { return this.refs; } }`,
      settings: { react: { version: '18.2.0', pragma: 'Other' } },
    },
  ],
  invalid: [
    // Custom JSX components have the same string-ref semantics as intrinsic
    // elements, including member-expression component names.
    {
      code: `<Widget ref="instance" />;`,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    {
      code: `<UI.Widget ref={'instance'} />;`,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    {
      code: `/** @jsx Preact.h */ class App extends Preact.Component { method() { return this.refs; } }`,
      settings: { react: { version: '18.2.0', pragma: 'Other' } },
      errors: [{ message: 'Using this.refs is deprecated.' }],
    },
    {
      code: `/** @jsx Preact.h */ var App = Preact.createClass({ method() { return this.refs; } });`,
      settings: {
        react: {
          version: '18.2.0',
          pragma: 'Other',
          createClass: 'createClass',
        },
      },
      errors: [{ message: 'Using this.refs is deprecated.' }],
    },
    {
      code: `class App extends React.Component { method() { return this.refs; } }`,
      settings: { react: { defaultVersion: '18.2.0' } },
      errors: [{ message: 'Using this.refs is deprecated.' }],
    },
    {
      code: `class App extends React.Component { method() { return this.refs; } }`,
      settings: {
        react: { version: 'detect', defaultVersion: '18.2.0' },
      },
      errors: [{ message: 'Using this.refs is deprecated.' }],
    },
    {
      code: `class App extends React.Component { method() { return this.refs; } }`,
      settings: { react: { version: 17 } },
      errors: [{ message: 'Using this.refs is deprecated.' }],
    },
    {
      code: `var App = createReactClass({ method: (function() { return this.refs; }) });`,
      settings: { react: { version: '18.2.0' } },
      errors: [{ message: 'Using this.refs is deprecated.' }],
    },
    {
      code: `var App = new createReactClass({ method() { return this.refs; } });`,
      settings: { react: { version: '18.2.0' } },
      errors: [{ message: 'Using this.refs is deprecated.' }],
    },
    // String literal directly.
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref="hello">Hello</div>;
          }
        });
      `,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // String literal inside expression container.
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={'hello'}>Hello</div>;
          }
        });
      `,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // Double-quoted string literal.
    {
      code: `<div ref={"hello"} />;`,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // Template literal with noTemplateLiterals: true.
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={\`hello\`}>Hello</div>;
          }
        });
      `,
      options: [{ noTemplateLiterals: true }],
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // Template literal with interpolation + noTemplateLiterals: true.
    {
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={\`hello\${index}\`}>Hello</div>;
          }
        });
      `,
      options: [{ noTemplateLiterals: true }],
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // Multiple offending refs in the same tree.
    {
      code: `
        function App() {
          return (
            <div>
              <div ref="first" />
              <div ref={'second'} />
            </div>
          );
        }
      `,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // Parenthesized string literal inside the expression container.
    {
      code: `<div ref={('hello')} />;`,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // Double-parenthesized string literal.
    {
      code: `<div ref={(('hello'))} />;`,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // Paren-wrapped template literal with noTemplateLiterals: true.
    {
      code: `<div ref={(\`hello\`)} />;`,
      options: [{ noTemplateLiterals: true }],
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
    // Paren-wrapped object literal argument to createReactClass — ESTree
    // flattens parens, tsgo preserves them. Regression case for GH-PR comment.
    {
      code: `
        var Hello = createReactClass(({
          render: function() { return <div ref="x" />; }
        }));
      `,
      errors: [
        { message: 'Using string literals in ref attributes is deprecated.' },
      ],
    },
  ],
});
