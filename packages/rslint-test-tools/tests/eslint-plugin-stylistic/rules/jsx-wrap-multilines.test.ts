/**
 * @fileoverview Prevent missing parentheses around multilines JSX
 * @author Yannick Croissant
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-wrap-multilines/jsx-wrap-multilines.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid: valids<…>(…), invalid: invalids<…>(…) })`
 *    -> `ruleTester.run('jsx-wrap-multilines', null as never, { valid, invalid })`.
 *  - `parserOptions` (`ecmaVersion`/`sourceType`/`ecmaFeatures.jsx`) dropped — rslint
 *    resolves via tsconfig (`run`'s default `lang: 'ts'`) and the RuleTester routes
 *    JSX fixtures to `.tsx`.
 *  - The `valids()` / `invalids()` helpers (`shared/test-utils/parsers-jsx.ts`)
 *    are NOT semantic to the rule: they MULTIPLEX each case across several PARSERS
 *    (ESLint-default, @babel/eslint-parser, @typescript-eslint/parser) and append a
 *    `// features: [...], parser: ...` disambiguation comment to `code`/`output`.
 *    With the resolved toolchain (ESLint >=10) the babel variant is always skipped
 *    (`skipBabel = gte(ESLint.version, '10.0.0')` === true); a `features:
 *    ['fragment']` case additionally skips the default-parser variant (`fragment` ∈
 *    skipBase + skipNewBabel) and runs ONLY on @typescript-eslint — exactly the
 *    parser rslint's ts-go emulates for a `.tsx` fixture. So each case is ported
 *    ONCE as its literal source (the appended parser-comment is an artifact of the
 *    multi-parser runner and is dropped), and the per-case `features: ['fragment']`
 *    markers are dropped.
 *  - The shared `const …` JSX snippets and the `addNewLineSymbols` string helper
 *    are kept verbatim (they are plain test DATA — backtick literals preserved
 *    byte-for-byte incl. their leading newline + indentation; `addNewLineSymbols`
 *    is a pure `String.replace` transform on the expected `output`, evaluated to a
 *    concrete fixed string by the RuleTester before comparison).
 *
 * Expected messages resolve from the plugin's own `meta.messages`:
 *   missingParens:    "Missing parentheses around multilines JSX"
 *   parensOnNewLines: "Parentheses around JSX should be on separate lines"
 *
 * Every invalid case pins `output`, so the RuleTester runs `--fix` and asserts the
 * fixed source for all of them; none are output-only (all also pin `errors`). The
 * upstream file has NO `$` unindent tags (all multi-line code uses plain backtick
 * literals), NO `readFileSync` external-fixture cases, NO `suggestions`, and only
 * the single `run()` block. The `._css_` / `._json_` / `._markdown_` test files
 * don't exist for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// ------------------------------------------------------------------------------
// Constants/Code Snippets (verbatim from upstream)
// ------------------------------------------------------------------------------

const OPTIONS_ALL_NEW_LINES = {
  declaration: 'parens-new-line',
  assignment: 'parens-new-line',
  return: 'parens-new-line',
  arrow: 'parens-new-line',
  condition: 'parens-new-line',
  logical: 'parens-new-line',
  prop: 'parens-new-line',
};

const RETURN_SINGLE_LINE = `
  var Hello = createReactClass({
    render: function() {
      return <p>Hello {this.props.name}</p>;
    }
  });
`;

const RETURN_PAREN = `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`;

const RETURN_PAREN_FRAGMENT = `
  var Hello = createReactClass({
    render: function() {
      return (<>
        <p>Hello {this.props.name}</p>
      </>);
    }
  });
`;

const RETURN_NO_PAREN = `
  var Hello = createReactClass({
    render: function() {
      return <div>
        <p>Hello {this.props.name}</p>
      </div>;
    }
  });
`;

const RETURN_NO_PAREN_FRAGMENT = `
  var Hello = createReactClass({
    render: function() {
      return <>
        <p>Hello {this.props.name}</p>
      </>;
    }
  });
`;

const RETURN_PAREN_NEW_LINE = `
  var Hello = createReactClass({
    render: function() {
      return (
        <div>
          <p>Hello {this.props.name}</p>
        </div>
      );
    }
  });
`;

const RETURN_PAREN_NEW_LINE_OPENING = `
  var Hello = createReactClass({
    render: function() {
      return (

      <div>
        <p>Hello {this.props.name}</p>
      </div>);
    }
  });
`;

const RETURN_PAREN_NEW_LINE_OPENING_FIXED = `
  var Hello = createReactClass({
    render: function() {
      return (

      <div>
        <p>Hello {this.props.name}</p>
      </div>
);
    }
  });
`;

const RETURN_PAREN_NEW_LINE_CLOSING = `
  var Hello = createReactClass({
    render: function() {
      return (<div>
        <p>Hello {this.props.name}</p>
      </div>

      );
    }
  });
`;

const RETURN_PAREN_NEW_LINE_CLOSING_FIXED = `
  var Hello = createReactClass({
    render: function() {
      return (
<div>
        <p>Hello {this.props.name}</p>
      </div>

      );
    }
  });
`;

const RETURN_PAREN_NEW_LINE_FRAGMENT = `
  var Hello = createReactClass({
    render: function() {
      return (
        <>
          <p>Hello {this.props.name}</p>
        </>
      );
    }
  });
`;

const RETURN_SINGLE_LINE_FRAGMENT = `
  var Hello = createReactClass({
    render: function() {
      return <>Hello {this.props.name}</>;
    }
  });
`;

const DECLARATION_TERNARY_SINGLE_LINE = 'var hello = foo ? <p>Hello</p> : <p>Hi</p>;';

const DECLARATION_TERNARY_SINGLE_LINE_FRAGMENT = 'var hello = foo ? <>Hello</> : <>Hi</>;';

const DECLARATION_TERNARY_PAREN = `
  var hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`;

const DECLARATION_TERNARY_PAREN_FRAGMENT = `
  var hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`;

const DECLARATION_TERNARY_NO_PAREN = `
  var hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`;

const DECLARATION_TERNARY_NO_PAREN_FRAGMENT = `
  var hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`;

const DECLARATION_TERNARY_PAREN_NEW_LINE = `
  var hello = foo ? (
    <div>
    <p>Hello</p>
    </div>
  ) : (
    <div>
      <p>Hi</p>
    </div>
  );
`;

const ASSIGNMENT_TERNARY_SINGLE_LINE = 'var hello; hello = foo ? <p>Hello</p> : <p>Hi</p>;';

const ASSIGNMENT_TERNARY_PAREN = `
  var hello;
  hello = foo ? (<div>
    <p>Hello</p>
  </div>) : (<div>
    <p>Hi</p>
  </div>);
`;

const ASSIGNMENT_TERNARY_PAREN_FRAGMENT = `
  var hello;
  hello = foo ? (<>
    <p>Hello</p>
  </>) : (<>
    <p>Hi</p>
  </>);
`;

const ASSIGNMENT_TERNARY_NO_PAREN = `
  var hello;
  hello = foo ? <div>
    <p>Hello</p>
  </div> : <div>
    <p>Hi</p>
  </div>;
`;

const ASSIGNMENT_TERNARY_NO_PAREN_FRAGMENT = `
  var hello;
  hello = foo ? <>
    <p>Hello</p>
  </> : <>
    <p>Hi</p>
  </>;
`;

const ASSIGNMENT_TERNARY_PAREN_NEW_LINE = `
  var hello;
  hello = foo ? (
    <div>
      <p>Hello</p>
    </div>
  ) : (
    <div>
      <p>Hi</p>
    </div>
  );
`;

const DECLARATION_SINGLE_LINE = 'var hello = <p>Hello</p>;';

const DECLARATION_PAREN = `
  var hello = (<div>
    <p>Hello</p>
  </div>);
`;

const DECLARATION_PAREN_FRAGMENT = `
  var hello = (<>
    <p>Hello</p>
  </>);
`;

const DECLARATION_NO_PAREN = `
  var hello = <div>
    <p>Hello</p>
  </div>;
`;

const DECLARATION_NO_PAREN_FRAGMENT = `
  var hello = <>
    <p>Hello</p>
  </>;
`;

const DECLARATION_PAREN_NEW_LINE = `
  var hello = (
    <div>
      <p>Hello</p>
    </div>
  );
`;

const ASSIGNMENT_SINGLE_LINE = 'var hello; hello = <p>Hello</p>;';

const ASSIGNMENT_PAREN = `
  var hello;
  hello = (<div>
    <p>Hello</p>
  </div>);
`;

const ASSIGNMENT_PAREN_FRAGMENT = `
  var hello;
  hello = (<>
    <p>Hello</p>
  </>);
`;

const ASSIGNMENT_NO_PAREN = `
  var hello;
  hello = <div>
    <p>Hello</p>
  </div>;
`;

const ASSIGNMENT_NO_PAREN_FRAGMENT = `
  var hello;
  hello = <>
    <p>Hello</p>
  </>;
`;

const ASSIGNMENT_PAREN_NEW_LINE = `
  var hello;
  hello = (
    <div>
      <p>Hello</p>
    </div>
  );
`;

const ARROW_SINGLE_LINE = 'var hello = () => <p>Hello</p>;';

const ARROW_PAREN = `
  var hello = () => (<div>
    <p>Hello</p>
  </div>);
`;

const ARROW_PAREN_FRAGMENT = `
  var hello = () => (<>
    <p>Hello</p>
  </>);
`;

const ARROW_NO_PAREN = `
  var hello = () => <div>
    <p>Hello</p>
  </div>;
`;

const ARROW_NO_PAREN_NEWLINE_TABS = `
var hello = () =>
\t<div>
\t\t<p>Hello</p>
\t</div>;
`;

const ARROW_NO_PAREN_FRAGMENT = `
  var hello = () => <>
    <p>Hello</p>
  </>;
`;

const ARROW_PAREN_NEW_LINE = `
  var hello = () => (
    <div>
      <p>Hello</p>
    </div>
  );
`;

const ARROW_PAREN_NEWLINE_TABS = `
var hello = () => (
\t<div>
\t\t<p>Hello</p>
\t</div>
);
`;

const CONDITION_SINGLE_LINE = 'foo ? <p>Hello</p> : null;';

const CONDITION_PAREN = `
  <div>
    {foo ? (<div>
        <p>Hello</p>
      </div>) : null}
  </div>
`;

const CONDITION_PAREN_FRAGMENT = `
  <div>
    {foo ? (<>
        <p>Hello</p>
      </>) : null}
  </div>
`;

const CONDITION_NO_PAREN = `
  <div>
    {foo ? <div>
        <p>Hello</p>
      </div> : null}
  </div>
`;

const CONDITION_NO_PAREN_FRAGMENT = `
  <div>
    {foo ? <>
        <p>Hello</p>
      </> : null}
  </div>
`;

const CONDITION_PAREN_NEW_LINE = `
  <div>
    {foo ? (
      <div>
        <p>Hello</p>
      </div>
    ) : null}
  </div>
`;

const LOGICAL_SINGLE_LINE = 'foo && <p>Hello</p>;';

const LOGICAL_PAREN = `
  <div>
    {foo &&
      (<div>
        <p>Hello World</p>
      </div>)
    }
  </div>
`;

const LOGICAL_PAREN_FRAGMENT = `
  <div>
    {foo &&
      (<>
        <p>Hello World</p>
      </>)
    }
  </div>
`;

const LOGICAL_NO_PAREN = `
  <div>
    {foo &&
      <div>
        <p>Hello World</p>
      </div>
    }
  </div>
`;

const LOGICAL_NO_PAREN_FRAGMENT = `
  <div>
    {foo &&
      <>
        <p>Hello World</p>
      </>
    }
  </div>
`;

const LOGICAL_PAREN_NEW_LINE_AUTOFIX = `
  <div>
    {foo && (
      <div>
        <p>Hello World</p>
      </div>
    )}
  </div>
`;

const LOGICAL_PAREN_NEW_LINE_AUTOFIX_FRAGMENT = `
  <div>
    {foo && (
      <>
        <p>Hello World</p>
      </>
    )}
  </div>
`;

const LOGICAL_PAREN_NEW_LINE = `
  <div>
    {foo && (
      <div>
        <p>Hello World</p>
      </div>
    )}
  </div>
`;

const ATTR_SINGLE_LINE = '<div prop={<p>Hello</p>}></div>';

const ATTR_PAREN = `
  <div prop={
    (<div>
      <p>Hello</p>
    </div>)
  }>
    <p>Hello</p>
  </div>
`;

const ATTR_PAREN_FRAGMENT = `
  <div prop={
    (<>
      <p>Hello</p>
    </>)
  }>
    <p>Hello</p>
  </div>
`;

const ATTR_NO_PAREN = `
  <div prop={
    <div>
      <p>Hello</p>
    </div>
  }>
    <p>Hello</p>
  </div>
`;

const ATTR_NO_PAREN_FRAGMENT = `
  <div prop={
    <>
      <p>Hello</p>
    </>
  }>
    <p>Hello</p>
  </div>
`;

const ATTR_PAREN_NEW_LINE = `
  <div prop={(
    <div>
      <p>Hello</p>
    </div>
  )}>
    <p>Hello</p>
  </div>
`;

const ATTR_PAREN_NEW_LINE_AUTOFIX = `
  <div prop={(
    <div>
      <p>Hello</p>
    </div>
  )}>
    <p>Hello</p>
  </div>
`;

const ATTR_PAREN_NEW_LINE_AUTOFIX_FRAGMENT = `
  <div prop={(
    <>
      <p>Hello</p>
    </>
  )}>
    <p>Hello</p>
  </div>
`;

const SFC_NO_PARENS_NO_NEWLINE = `
  export default () =>
    <div>
        with newline without parentheses eslint crashes
    </div>`;

const SFC_NO_PARENS_AUTOFIX = `
  export default () => (
    <div>
        with newline without parentheses eslint crashes
    </div>
  )`;

const ARROW_WITH_EXPORT = `
  const Component = () =>
    <div>
      <p>Some text</p>
    </div>

  export { Component as default }
`;

const ARROW_WITH_EXPORT_AUTOFIX = `
  const Component = () => (
    <div>
      <p>Some text</p>
    </div>
  )

  export { Component as default }
`;

const ARROW_WITH_LOGICAL = `
const Component = props => (
  <div>
    {true &&
      <div>
        <p>Some text</p>
      </div>
    }
  </div>
)
`;

const ARROW_WITH_LOGICAL_AUTOFIX = `
const Component = props => (
  <div>
    {true && (
      <div>
        <p>Some text</p>
      </div>
    )}
  </div>
)
`;

function addNewLineSymbols(code: string) {
  return code.replace(/\(</g, '(\n<').replace(/>\)/g, '>\n)');
}

ruleTester.run('jsx-wrap-multilines', null as never, {
  valid: [
    {
      code: RETURN_SINGLE_LINE,
    },
    {
      code: RETURN_SINGLE_LINE_FRAGMENT,
    },
    {
      code: RETURN_PAREN,
    },
    {
      code: RETURN_SINGLE_LINE,
      options: [{ return: true }],
    },
    {
      code: RETURN_SINGLE_LINE_FRAGMENT,
      options: [{ return: true }],
    },
    {
      code: RETURN_PAREN,
      options: [{ return: true }],
    },
    {
      code: RETURN_NO_PAREN,
      options: [{ return: 'ignore' }],
    },
    {
      code: RETURN_NO_PAREN,
      options: [{ return: false }],
    },
    {
      code: DECLARATION_TERNARY_SINGLE_LINE,
    },
    {
      code: DECLARATION_TERNARY_SINGLE_LINE_FRAGMENT,
    },
    {
      code: DECLARATION_TERNARY_PAREN,
    },
    {
      code: DECLARATION_TERNARY_SINGLE_LINE,
      options: [{ declaration: true }],
    },
    {
      code: DECLARATION_TERNARY_PAREN,
      options: [{ declaration: true }],
    },
    {
      code: DECLARATION_TERNARY_NO_PAREN,
      options: [{ declaration: 'ignore' }],
    },
    {
      code: DECLARATION_TERNARY_NO_PAREN,
      options: [{ declaration: false }],
    },
    {
      code: ASSIGNMENT_TERNARY_SINGLE_LINE,
    },
    {
      code: ASSIGNMENT_TERNARY_PAREN,
    },
    {
      code: ASSIGNMENT_TERNARY_SINGLE_LINE,
      options: [{ assignment: true }],
    },
    {
      code: ASSIGNMENT_TERNARY_PAREN,
      options: [{ assignment: true }],
    },
    {
      code: ASSIGNMENT_TERNARY_NO_PAREN,
      options: [{ assignment: 'ignore' }],
    },
    {
      code: ASSIGNMENT_TERNARY_NO_PAREN,
      options: [{ assignment: false }],
    },
    {
      code: DECLARATION_SINGLE_LINE,
    },
    {
      code: DECLARATION_PAREN,
    },
    {
      code: DECLARATION_PAREN_FRAGMENT,
    },
    {
      code: DECLARATION_SINGLE_LINE,
      options: [{ declaration: true }],
    },
    {
      code: DECLARATION_PAREN,
      options: [{ declaration: true }],
    },
    {
      code: DECLARATION_NO_PAREN,
      options: [{ declaration: 'ignore' }],
    },
    {
      code: DECLARATION_NO_PAREN_FRAGMENT,
      options: [{ declaration: 'ignore' }],
    },
    {
      code: DECLARATION_NO_PAREN,
      options: [{ declaration: false }],
    },
    {
      code: ASSIGNMENT_SINGLE_LINE,
      options: [{ declaration: 'ignore' }],
    },
    {
      code: ASSIGNMENT_SINGLE_LINE,
      options: [{ declaration: false }],
    },
    {
      code: ASSIGNMENT_PAREN,
    },
    {
      code: ASSIGNMENT_PAREN_FRAGMENT,
    },
    {
      code: ASSIGNMENT_PAREN,
      options: [{ assignment: true }],
    },
    {
      code: ASSIGNMENT_NO_PAREN,
      options: [{ assignment: 'ignore' }],
    },
    {
      code: ASSIGNMENT_NO_PAREN_FRAGMENT,
      options: [{ assignment: 'ignore' }],
    },
    {
      code: ASSIGNMENT_NO_PAREN,
      options: [{ assignment: false }],
    },
    {
      code: ARROW_PAREN,
    },
    {
      code: ARROW_PAREN_FRAGMENT,
    },
    {
      code: ARROW_SINGLE_LINE,
    },
    {
      code: ARROW_PAREN,
      options: [{ arrow: true }],
    },
    {
      code: ARROW_SINGLE_LINE,
      options: [{ arrow: true }],
    },
    {
      code: ARROW_NO_PAREN,
      options: [{ arrow: 'ignore' }],
    },
    {
      code: ARROW_NO_PAREN_NEWLINE_TABS,
      options: [{ arrow: 'ignore' }],
    },
    {
      code: ARROW_PAREN_NEWLINE_TABS,
      options: [{ arrow: 'parens-new-line' }],
    },
    {
      code: ARROW_NO_PAREN_FRAGMENT,
      options: [{ arrow: 'ignore' }],
    },
    {
      code: ARROW_NO_PAREN,
      options: [{ arrow: false }],
    },
    {
      code: CONDITION_SINGLE_LINE,
    },
    {
      code: CONDITION_SINGLE_LINE,
      options: [{ condition: true }],
    },
    {
      code: CONDITION_NO_PAREN,
    },
    {
      code: CONDITION_PAREN,
      options: [{ condition: true }],
    },
    {
      code: CONDITION_PAREN_FRAGMENT,
      options: [{ condition: true }],
    },
    {
      code: LOGICAL_SINGLE_LINE,
    },
    {
      code: LOGICAL_NO_PAREN,
    },
    {
      code: LOGICAL_PAREN,
      options: [{ logical: true }],
    },
    {
      code: LOGICAL_PAREN_FRAGMENT,
      options: [{ logical: true }],
    },
    {
      code: ATTR_SINGLE_LINE,
    },
    {
      code: ATTR_NO_PAREN,
    },
    {
      code: ATTR_PAREN,
      options: [{ prop: true }],
    },
    {
      code: ATTR_PAREN_FRAGMENT,
      options: [{ prop: true }],
    },
    {
      code: RETURN_PAREN_NEW_LINE,
      options: [{ return: 'parens-new-line' }],
    },
    {
      code: RETURN_PAREN_NEW_LINE_FRAGMENT,
      options: [{ return: 'parens-new-line' }],
    },
    {
      code: DECLARATION_TERNARY_PAREN_NEW_LINE,
      options: [{ declaration: 'parens-new-line' }],
    },
    {
      code: ASSIGNMENT_TERNARY_PAREN_NEW_LINE,
      options: [{ assignment: 'parens-new-line' }],
    },
    {
      code: DECLARATION_PAREN_NEW_LINE,
      options: [{ declaration: 'parens-new-line' }],
    },
    {
      code: ASSIGNMENT_PAREN_NEW_LINE,
      options: [{ assignment: 'parens-new-line' }],
    },
    {
      code: ARROW_PAREN_NEW_LINE,
      options: [{ arrow: 'parens-new-line' }],
    },
    {
      code: CONDITION_PAREN_NEW_LINE,
      options: [{ condition: 'parens-new-line' }],
    },
    {
      code: LOGICAL_PAREN_NEW_LINE,
      options: [{ logical: 'parens-new-line' }],
    },
    {
      code: ATTR_PAREN_NEW_LINE,
      options: [{ prop: 'parens-new-line' }],
    },
    {
      code: `
        obj = {
          foo: <div><p>Hello</p></div>
        }
      `,
      options: [{ propertyValue: 'parens' }],
    },
  ],

  invalid: [
    {
      code: RETURN_NO_PAREN,
      output: RETURN_PAREN,
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: RETURN_NO_PAREN_FRAGMENT,
      output: RETURN_PAREN_FRAGMENT,
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: RETURN_NO_PAREN,
      output: RETURN_PAREN,
      options: [{ return: true }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: RETURN_NO_PAREN_FRAGMENT,
      output: RETURN_PAREN_FRAGMENT,
      options: [{ return: true }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: DECLARATION_TERNARY_NO_PAREN,
      output: DECLARATION_TERNARY_PAREN,
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: DECLARATION_TERNARY_NO_PAREN_FRAGMENT,
      output: DECLARATION_TERNARY_PAREN_FRAGMENT,
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: DECLARATION_TERNARY_NO_PAREN,
      output: DECLARATION_TERNARY_PAREN,
      options: [{ declaration: true }],
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: DECLARATION_TERNARY_NO_PAREN_FRAGMENT,
      output: DECLARATION_TERNARY_PAREN_FRAGMENT,
      options: [{ declaration: true }],
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: ASSIGNMENT_TERNARY_NO_PAREN,
      output: ASSIGNMENT_TERNARY_PAREN,
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: ASSIGNMENT_TERNARY_NO_PAREN_FRAGMENT,
      output: ASSIGNMENT_TERNARY_PAREN_FRAGMENT,
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: ASSIGNMENT_TERNARY_NO_PAREN,
      output: ASSIGNMENT_TERNARY_PAREN,
      options: [{ assignment: true }],
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: ASSIGNMENT_TERNARY_NO_PAREN_FRAGMENT,
      output: ASSIGNMENT_TERNARY_PAREN_FRAGMENT,
      options: [{ assignment: true }],
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: DECLARATION_NO_PAREN,
      output: DECLARATION_PAREN,
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: DECLARATION_NO_PAREN_FRAGMENT,
      output: DECLARATION_PAREN_FRAGMENT,
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: DECLARATION_NO_PAREN,
      output: DECLARATION_PAREN,
      options: [{ declaration: true }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ASSIGNMENT_NO_PAREN,
      output: ASSIGNMENT_PAREN,
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ASSIGNMENT_NO_PAREN_FRAGMENT,
      output: ASSIGNMENT_PAREN_FRAGMENT,
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ASSIGNMENT_NO_PAREN,
      output: ASSIGNMENT_PAREN,
      options: [{ assignment: true }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ARROW_NO_PAREN,
      output: ARROW_PAREN,
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ARROW_NO_PAREN_FRAGMENT,
      output: ARROW_PAREN_FRAGMENT,
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ARROW_NO_PAREN,
      output: ARROW_PAREN,
      options: [{ arrow: true }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ARROW_NO_PAREN_NEWLINE_TABS,
      output: ARROW_PAREN_NEWLINE_TABS,
      options: [{ arrow: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: CONDITION_NO_PAREN,
      output: CONDITION_PAREN,
      options: [{ condition: 'parens' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: CONDITION_NO_PAREN_FRAGMENT,
      output: CONDITION_PAREN_FRAGMENT,
      options: [{ condition: 'parens' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: CONDITION_NO_PAREN,
      output: CONDITION_PAREN,
      options: [{ condition: true }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: LOGICAL_NO_PAREN,
      output: LOGICAL_PAREN,
      options: [{ logical: 'parens' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: LOGICAL_NO_PAREN_FRAGMENT,
      output: LOGICAL_PAREN_FRAGMENT,
      options: [{ logical: 'parens' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: LOGICAL_NO_PAREN,
      output: LOGICAL_PAREN,
      options: [{ logical: true }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ATTR_NO_PAREN,
      output: ATTR_PAREN,
      options: [{ prop: 'parens' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ATTR_NO_PAREN_FRAGMENT,
      output: ATTR_PAREN_FRAGMENT,
      options: [{ prop: 'parens' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ATTR_NO_PAREN,
      output: ATTR_PAREN,
      options: [{ prop: true }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: RETURN_NO_PAREN,
      output: addNewLineSymbols(RETURN_PAREN),
      options: [{ return: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: RETURN_NO_PAREN_FRAGMENT,
      output: addNewLineSymbols(RETURN_PAREN_FRAGMENT),
      options: [{ return: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: RETURN_PAREN,
      output: addNewLineSymbols(RETURN_PAREN),
      options: [{ return: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: RETURN_PAREN_NEW_LINE_OPENING,
      output: RETURN_PAREN_NEW_LINE_OPENING_FIXED,
      options: [{ return: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: RETURN_PAREN_NEW_LINE_CLOSING,
      output: RETURN_PAREN_NEW_LINE_CLOSING_FIXED,
      options: [{ return: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: RETURN_PAREN_FRAGMENT,
      output: addNewLineSymbols(RETURN_PAREN_FRAGMENT),
      options: [{ return: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: DECLARATION_TERNARY_NO_PAREN,
      output: addNewLineSymbols(DECLARATION_TERNARY_PAREN),
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: DECLARATION_TERNARY_NO_PAREN_FRAGMENT,
      output: addNewLineSymbols(DECLARATION_TERNARY_PAREN_FRAGMENT),
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: DECLARATION_TERNARY_PAREN_FRAGMENT,
      output: addNewLineSymbols(DECLARATION_TERNARY_PAREN_FRAGMENT),
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        { messageId: 'parensOnNewLines' },
        { messageId: 'parensOnNewLines' },
      ],
    },
    {
      code: DECLARATION_TERNARY_PAREN,
      output: addNewLineSymbols(DECLARATION_TERNARY_PAREN),
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        { messageId: 'parensOnNewLines' },
        { messageId: 'parensOnNewLines' },
      ],
    },
    {
      code: DECLARATION_TERNARY_PAREN_FRAGMENT,
      output: addNewLineSymbols(DECLARATION_TERNARY_PAREN_FRAGMENT),
      options: [{ declaration: 'parens-new-line' }],
      errors: [
        { messageId: 'parensOnNewLines' },
        { messageId: 'parensOnNewLines' },
      ],
    },
    {
      code: ASSIGNMENT_TERNARY_NO_PAREN,
      output: addNewLineSymbols(ASSIGNMENT_TERNARY_PAREN),
      options: [{ assignment: 'parens-new-line' }],
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: ASSIGNMENT_TERNARY_NO_PAREN_FRAGMENT,
      output: addNewLineSymbols(ASSIGNMENT_TERNARY_PAREN_FRAGMENT),
      options: [{ assignment: 'parens-new-line' }],
      errors: [
        { messageId: 'missingParens' },
        { messageId: 'missingParens' },
      ],
    },
    {
      code: ASSIGNMENT_TERNARY_PAREN,
      output: addNewLineSymbols(ASSIGNMENT_TERNARY_PAREN),
      options: [{ assignment: 'parens-new-line' }],
      errors: [
        { messageId: 'parensOnNewLines' },
        { messageId: 'parensOnNewLines' },
      ],
    },
    {
      code: ASSIGNMENT_TERNARY_PAREN_FRAGMENT,
      output: addNewLineSymbols(ASSIGNMENT_TERNARY_PAREN_FRAGMENT),
      options: [{ assignment: 'parens-new-line' }],
      errors: [
        { messageId: 'parensOnNewLines' },
        { messageId: 'parensOnNewLines' },
      ],
    },
    {
      code: DECLARATION_NO_PAREN,
      output: addNewLineSymbols(DECLARATION_PAREN),
      options: [{ declaration: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: DECLARATION_PAREN,
      output: addNewLineSymbols(DECLARATION_PAREN),
      options: [{ declaration: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: ASSIGNMENT_NO_PAREN,
      output: addNewLineSymbols(ASSIGNMENT_PAREN),
      options: [{ assignment: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ASSIGNMENT_PAREN,
      output: addNewLineSymbols(ASSIGNMENT_PAREN),
      options: [{ assignment: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: ARROW_PAREN,
      output: addNewLineSymbols(ARROW_PAREN),
      options: [{ arrow: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: ARROW_PAREN_FRAGMENT,
      output: addNewLineSymbols(ARROW_PAREN_FRAGMENT),
      options: [{ arrow: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: ARROW_NO_PAREN,
      output: addNewLineSymbols(ARROW_PAREN),
      options: [{ arrow: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ARROW_NO_PAREN_FRAGMENT,
      output: addNewLineSymbols(ARROW_PAREN_FRAGMENT),
      options: [{ arrow: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: CONDITION_PAREN,
      output: addNewLineSymbols(CONDITION_PAREN),
      options: [{ condition: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: CONDITION_PAREN_FRAGMENT,
      output: addNewLineSymbols(CONDITION_PAREN_FRAGMENT),
      options: [{ condition: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: CONDITION_NO_PAREN,
      output: addNewLineSymbols(CONDITION_PAREN),
      options: [{ condition: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: CONDITION_NO_PAREN_FRAGMENT,
      output: addNewLineSymbols(CONDITION_PAREN_FRAGMENT),
      options: [{ condition: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: LOGICAL_PAREN,
      output: addNewLineSymbols(LOGICAL_PAREN),
      options: [{ logical: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: LOGICAL_NO_PAREN,
      output: LOGICAL_PAREN_NEW_LINE_AUTOFIX,
      options: [{ logical: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: LOGICAL_NO_PAREN_FRAGMENT,
      output: LOGICAL_PAREN_NEW_LINE_AUTOFIX_FRAGMENT,
      options: [{ logical: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ATTR_PAREN,
      output: addNewLineSymbols(ATTR_PAREN),
      options: [{ prop: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: ATTR_PAREN_FRAGMENT,
      output: addNewLineSymbols(ATTR_PAREN_FRAGMENT),
      options: [{ prop: 'parens-new-line' }],
      errors: [{ messageId: 'parensOnNewLines' }],
    },
    {
      code: ATTR_NO_PAREN,
      output: ATTR_PAREN_NEW_LINE_AUTOFIX,
      options: [{ prop: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ATTR_NO_PAREN_FRAGMENT,
      output: ATTR_PAREN_NEW_LINE_AUTOFIX_FRAGMENT,
      options: [{ prop: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: SFC_NO_PARENS_NO_NEWLINE,
      output: SFC_NO_PARENS_AUTOFIX,
      options: [OPTIONS_ALL_NEW_LINES],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ARROW_WITH_EXPORT,
      output: ARROW_WITH_EXPORT_AUTOFIX,
      options: [
        {
          declaration: 'parens-new-line',
          assignment: 'parens-new-line',
          return: 'parens-new-line',
          arrow: 'parens-new-line',
          condition: 'parens-new-line',
          logical: 'ignore',
          prop: 'ignore',
        },
      ],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: ARROW_WITH_LOGICAL,
      output: ARROW_WITH_LOGICAL_AUTOFIX,
      options: [{ logical: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: `
        import React from 'react';

        const A =
        <div>
            B
        </div>;
      `,
      output: `
        import React from 'react';

        const A = (
        <div>
            B
        </div>
      );
      `,
      options: [{ declaration: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: `
      const obj = {
        foo: <div>
          bar
        </div>,
        baz: 2,
      }
      `,
      output: `
      const obj = {
        foo: (
<div>
          bar
        </div>
),
        baz: 2,
      }
      `,
      options: [{ propertyValue: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: `
const obj = {
  foo: <div>
         bar
       </div>,
  baz: <div prop={
         <div>
           <p>Hello</p>
         </div>
       }>
         <p>Hello</p>
       </div>,
}
      `,
      output: `
const obj = {
  foo: (
<div>
         bar
       </div>
),
  baz: (
<div prop={
         <div>
           <p>Hello</p>
         </div>
       }>
         <p>Hello</p>
       </div>
),
}
      `,
      options: [{ propertyValue: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }, { messageId: 'missingParens' }],
    },
    {
      code: `
const obj = {
  foo: <div>
         bar
       </div>,
}
      `,
      output: `
const obj = {
  foo: (<div>
         bar
       </div>),
}
      `,
      options: [{ propertyValue: 'parens' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: `
const obj = {
  foo: {
    bar: <div>
      baz
    </div>,
  }
}
      `,
      output: `
const obj = {
  foo: {
    bar: (<div>
      baz
    </div>),
  }
}
      `,
      options: [{ propertyValue: 'parens' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: `
    const MissingParenWIthCOmment = () =>
      // Something about this line
      <>
        <input />
      </>
          `,
      output: `
    const MissingParenWIthCOmment = () => (
      // Something about this line
      <>
        <input />
      </>
    )
          `,
      options: [{ arrow: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
    {
      code: `
    var hello = {
      foo:
        // Something about this line
        <div>
          <p>Hello</p>
        </div>
    };
          `,
      output: `
    var hello = {
      foo: (
        // Something about this line
        <div>
          <p>Hello</p>
        </div>
      )};
          `,
      options: [{ propertyValue: 'parens-new-line' }],
      errors: [{ messageId: 'missingParens' }],
    },
  ],
});
