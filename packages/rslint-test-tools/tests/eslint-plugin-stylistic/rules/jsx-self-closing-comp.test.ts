/**
 * @fileoverview Prevent extra closing tags for components without children.
 * @author Yannick Croissant
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-self-closing-comp/jsx-self-closing-comp.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-self-closing-comp', null as never, { valid, invalid })`
 *    (upstream's `name: 'self-closing-comp'` is an alias; the rule is registered
 *    in the plugin as `jsx-self-closing-comp`, which is the id rslint must use.)
 *  - `parserOptions` (`ecmaFeatures.jsx: true`) dropped — rslint resolves via
 *    tsconfig and the RuleTester routes JSX fixtures to `.tsx`.
 *
 * The upstream file wraps its cases in the `valids()` / `invalids()` helpers from
 * `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case across
 * several PARSERS (ESLint-default, @babel/eslint-parser, @typescript-eslint/parser)
 * and append a `// features: [...], parser: ...` comment to `code`/`output`. With
 * the resolved toolchain (ESLint 10.5.0) the babel variant is skipped
 * (`skipBabel = gte(ESLint.version, '10.0.0')` === true), leaving only the
 * default + @typescript-eslint variants — which are identical under rslint's
 * single ts-go parser. That parser-multiplexing is a pure upstream-harness
 * artifact with no rslint analog, so each case is ported ONCE as its literal
 * source (the appended parser-comment is dropped); the JSX code is verbatim and
 * runs as a `.tsx` fixture. No case carries a `features` array, so no case was
 * dropped by the `skipBase` filter either.
 *
 * The expected message resolves from the plugin's own `meta.messages`:
 *   notSelfClosing: "Empty components are self-closing"
 *
 * The upstream file contains NO `$` unindent template tags, NO `readFileSync`
 * external-fixture cases, NO `suggestions`, and only the single `run()` block
 * above. (Two valid cases use plain multi-line template literals; their leading
 * whitespace is preserved verbatim.) The `._css_` / `._json_` / `._markdown_`
 * test files don't exist for this rule.
 *
 * Every case is plain JSX that ts-go parses identically to the upstream parser,
 * so no case surfaces a rslint<->upstream gap and nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-self-closing-comp', null as never, {
  valid: [
    {
      code: 'var HelloJohn = <Hello name="John" />;',
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John" />;',
    },
    {
      code: 'var Profile = <Hello name="John"><img src="picture.png" /></Hello>;',
    },
    {
      code: 'var Profile = <Hello.Compound name="John"><img src="picture.png" /></Hello.Compound>;',
    },
    {
      code: `
        <Hello>
          <Hello name="John" />
        </Hello>
      `,
    },
    {
      code: `
        <Hello.Compound>
          <Hello.Compound name="John" />
        </Hello.Compound>
      `,
    },
    {
      code: 'var HelloJohn = <Hello name="John"> </Hello>;',
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John"> </Hello.Compound>;',
    },
    {
      code: 'var HelloJohn = <Hello name="John">        </Hello>;',
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John">        </Hello.Compound>;',
    },
    {
      code: 'var HelloJohn = <div>&nbsp;</div>;',
    },
    {
      code: 'var HelloJohn = <div>{\' \'}</div>;',
    },
    {
      code: 'var HelloJohn = <Hello name="John">&nbsp;</Hello>;',
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John">&nbsp;</Hello.Compound>;',
    },
    {
      code: 'var HelloJohn = <Hello name="John" />;',
      options: [],
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John" />;',
      options: [],
    },
    {
      code: 'var Profile = <Hello name="John"><img src="picture.png" /></Hello>;',
      options: [],
    },
    {
      code: 'var Profile = <Hello.Compound name="John"><img src="picture.png" /></Hello.Compound>;',
      options: [],
    },
    {
      code: `
        <Hello>
          <Hello name="John" />
        </Hello>
      `,
      options: [],
    },
    {
      code: `
        <Hello.Compound>
          <Hello.Compound name="John" />
        </Hello.Compound>
      `,
      options: [],
    },
    {
      code: 'var HelloJohn = <div> </div>;',
      options: [],
    },
    {
      code: 'var HelloJohn = <div>        </div>;',
      options: [],
    },
    {
      code: 'var HelloJohn = <div>&nbsp;</div>;',
      options: [],
    },
    {
      code: 'var HelloJohn = <div>{\' \'}</div>;',
      options: [],
    },
    {
      code: 'var HelloJohn = <Hello name="John">&nbsp;</Hello>;',
      options: [],
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John">&nbsp;</Hello.Compound>;',
      options: [],
    },
    {
      code: 'var HelloJohn = <Hello name="John"></Hello>;',
      options: [{ component: false }],
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John"></Hello.Compound>;',
      options: [{ component: false }],
    },
    {
      code: 'var HelloJohn = <Hello name="John">\n</Hello>;',
      options: [{ component: false }],
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John">\n</Hello.Compound>;',
      options: [{ component: false }],
    },
    {
      code: 'var HelloJohn = <Hello name="John"> </Hello>;',
      options: [{ component: false }],
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John"> </Hello.Compound>;',
      options: [{ component: false }],
    },
    {
      code: 'var contentContainer = <div className="content" />;',
      options: [{ html: true }],
    },
    {
      code: 'var contentContainer = <div className="content"><img src="picture.png" /></div>;',
      options: [{ html: true }],
    },
    {
      code: `
        <div>
          <div className="content" />
        </div>
      `,
      options: [{ html: true }],
    },
  ],

  invalid: [
    {
      code: 'var contentContainer = <div className="content"></div>;',
      output: 'var contentContainer = <div className="content" />;',
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var contentContainer = <div className="content"></div>;',
      output: 'var contentContainer = <div className="content" />;',
      options: [],
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var HelloJohn = <Hello name="John"></Hello>;',
      output: 'var HelloJohn = <Hello name="John" />;',
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var CompoundHelloJohn = <Hello.Compound name="John"></Hello.Compound>;',
      output: 'var CompoundHelloJohn = <Hello.Compound name="John" />;',
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var HelloJohn = <Hello name="John">\n</Hello>;',
      output: 'var HelloJohn = <Hello name="John" />;',
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John">\n</Hello.Compound>;',
      output: 'var HelloJohn = <Hello.Compound name="John" />;',
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var HelloJohn = <Hello name="John"></Hello>;',
      output: 'var HelloJohn = <Hello name="John" />;',
      options: [],
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John"></Hello.Compound>;',
      output: 'var HelloJohn = <Hello.Compound name="John" />;',
      options: [],
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var HelloJohn = <Hello name="John">\n</Hello>;',
      output: 'var HelloJohn = <Hello name="John" />;',
      options: [],
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var HelloJohn = <Hello.Compound name="John">\n</Hello.Compound>;',
      output: 'var HelloJohn = <Hello.Compound name="John" />;',
      options: [],
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var contentContainer = <div className="content"></div>;',
      output: 'var contentContainer = <div className="content" />;',
      options: [{ html: true }],
      errors: [{ messageId: 'notSelfClosing' }],
    },
    {
      code: 'var contentContainer = <div className="content">\n</div>;',
      output: 'var contentContainer = <div className="content" />;',
      options: [{ html: true }],
      errors: [{ messageId: 'notSelfClosing' }],
    },
  ],
});
