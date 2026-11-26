import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectStringInRefDeprecatedError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'stringInRefDeprecated',
    message: 'Using string literals in ref attributes is deprecated.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectThisRefsDeprecatedError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'thisRefsDeprecated',
    message: 'Using this.refs is deprecated.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-string-refs', {
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

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            var component = this.hello;
          },
          render: function() {
            return <div ref={c => this.hello = c}>Hello {this.props.name}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={\`hello\`}>Hello {this.props.name}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={\`hello\${index}\`}>Hello {this.props.name}</div>;
          }
        });
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            var component = this.refs.hello;
          },
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      settings: { react: { version: '18.3.0' } },
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    // Custom JSX components have the same string-ref semantics as intrinsic
    // elements, including member-expression component names.
    {
      code: `<Widget ref="instance" />;`,
      errors: [expectStringInRefDeprecatedError([1, 9, 1, 23])],
    },
    {
      code: `<UI.Widget ref={'instance'} />;`,
      errors: [expectStringInRefDeprecatedError([1, 12, 1, 28])],
    },
    {
      code: `/** @jsx Preact.h */ class App extends Preact.Component { method() { return this.refs; } }`,
      settings: { react: { version: '18.2.0', pragma: 'Other' } },
      errors: [expectThisRefsDeprecatedError([1, 77, 1, 86])],
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
      errors: [expectThisRefsDeprecatedError([1, 71, 1, 80])],
    },
    {
      code: `class App extends React.Component { method() { return this.refs; } }`,
      settings: { react: { defaultVersion: '18.2.0' } },
      errors: [expectThisRefsDeprecatedError([1, 55, 1, 64])],
    },
    {
      code: `class App extends React.Component { method() { return this.refs; } }`,
      settings: {
        react: { version: 'detect', defaultVersion: '18.2.0' },
      },
      errors: [expectThisRefsDeprecatedError([1, 55, 1, 64])],
    },
    {
      code: `class App extends React.Component { method() { return this.refs; } }`,
      settings: { react: { version: 17 } },
      errors: [expectThisRefsDeprecatedError([1, 55, 1, 64])],
    },
    {
      code: `var App = createReactClass({ method: (function() { return this.refs; }) });`,
      settings: { react: { version: '18.2.0' } },
      errors: [expectThisRefsDeprecatedError([1, 59, 1, 68])],
    },
    {
      code: `var App = new createReactClass({ method() { return this.refs; } });`,
      settings: { react: { version: '18.2.0' } },
      errors: [expectThisRefsDeprecatedError([1, 52, 1, 61])],
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
      errors: [expectStringInRefDeprecatedError([4, 25, 4, 36])],
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
      errors: [expectStringInRefDeprecatedError([4, 25, 4, 38])],
    },
    // Double-quoted string literal.
    {
      code: `<div ref={"hello"} />;`,
      errors: [expectStringInRefDeprecatedError([1, 6, 1, 19])],
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
      errors: [expectStringInRefDeprecatedError([4, 25, 4, 38])],
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
      errors: [expectStringInRefDeprecatedError([4, 25, 4, 46])],
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
        expectStringInRefDeprecatedError([5, 20, 5, 31]),
        expectStringInRefDeprecatedError([6, 20, 6, 34]),
      ],
    },
    // Parenthesized string literal inside the expression container.
    {
      code: `<div ref={('hello')} />;`,
      errors: [expectStringInRefDeprecatedError([1, 6, 1, 21])],
    },
    // Double-parenthesized string literal.
    {
      code: `<div ref={(('hello'))} />;`,
      errors: [expectStringInRefDeprecatedError([1, 6, 1, 23])],
    },
    // Paren-wrapped template literal with noTemplateLiterals: true.
    {
      code: `<div ref={(\`hello\`)} />;`,
      options: [{ noTemplateLiterals: true }],
      errors: [expectStringInRefDeprecatedError([1, 6, 1, 21])],
    },
    // Paren-wrapped object literal argument to createReactClass — ESTree
    // flattens parens, tsgo preserves them. Regression case for GH-PR comment.
    {
      code: `
        var Hello = createReactClass(({
          render: function() { return <div ref="x" />; }
        }));
      `,
      errors: [expectStringInRefDeprecatedError([3, 44, 3, 51])],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            var component = this.refs.hello;
          },
          render: function() {
            return <div>Hello {this.props.name}</div>;
          }
        });
      `,
      settings: { react: { version: '18.2.0' } },
      errors: [expectThisRefsDeprecatedError([4, 29, 4, 38])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref="hello">Hello {this.props.name}</div>;
          }
        });
      `,
      settings: { react: { version: '18.2.0' } },
      errors: [expectStringInRefDeprecatedError([4, 25, 4, 36])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        var Hello = createReactClass({
          render: function() {
            return <div ref={'hello'}>Hello {this.props.name}</div>;
          }
        });
      `,
      settings: { react: { version: '18.2.0' } },
      errors: [expectStringInRefDeprecatedError([4, 25, 4, 38])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
            var component = this.refs.hello;
          },
          render: function() {
            return <div ref="hello">Hello {this.props.name}</div>;
          }
        });
      `,
      settings: { react: { version: '18.2.0' } },
      errors: [
        expectThisRefsDeprecatedError([4, 29, 4, 38]),
        expectStringInRefDeprecatedError([7, 25, 7, 36]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
          var component = this.refs.hello;
          },
          render: function() {
            return <div ref={\`hello\`}>Hello {this.props.name}</div>;
          }
        });
      `,
      options: [{ noTemplateLiterals: true }],
      settings: { react: { version: '18.2.0' } },
      errors: [
        expectThisRefsDeprecatedError([4, 27, 4, 36]),
        expectStringInRefDeprecatedError([7, 25, 7, 38]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
          var component = this.refs.hello;
          },
          render: function() {
            return <div ref={\`hello\${index}\`}>Hello {this.props.name}</div>;
          }
        });
      `,
      options: [{ noTemplateLiterals: true }],
      settings: { react: { version: '18.2.0' } },
      errors: [
        expectThisRefsDeprecatedError([4, 27, 4, 36]),
        expectStringInRefDeprecatedError([7, 25, 7, 46]),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        var Hello = createReactClass({
          componentDidMount: function() {
          var component = this.refs.hello;
          },
          render: function() {
            return <div ref={\`hello\${index}\`}>Hello {this.props.name}</div>;
          }
        });
      `,
      options: [{ noTemplateLiterals: true }],
      settings: { react: { version: '18.3.0' } },
      errors: [expectStringInRefDeprecatedError([7, 25, 7, 46])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
