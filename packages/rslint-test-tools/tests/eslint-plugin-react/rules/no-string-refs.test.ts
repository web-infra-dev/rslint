import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// NOTE: `this.refs` detection depends on `settings.react.version`, which the
// shared JS rule-tester does not thread through to `lint()`. Those cases are
// covered by the Go unit tests (`no_string_refs_test.go`). The JS suite here
// focuses on the JSX `ref={...}` detection, which is version-independent.

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
  ],
  invalid: [
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
