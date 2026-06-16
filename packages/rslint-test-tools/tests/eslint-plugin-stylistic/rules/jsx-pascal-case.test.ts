/**
 * @fileoverview Tests for jsx-pascal-case rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-pascal-case/jsx-pascal-case.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid, invalid })`
 *      -> `ruleTester.run('jsx-pascal-case', null as never, { valid, invalid })`
 *  - The upstream `valids(...)` / `invalids(...)` wrappers (`#test/parsers-jsx`)
 *    are a multi-parser fan-out harness: they replay each case through the
 *    default / `@babel/eslint-parser` / `@typescript-eslint/parser` parsers and
 *    append a `// features: [...], parser: ...` trailing comment to the code.
 *    That augmentation is purely a test-harness artifact (it does not change the
 *    rule's diagnostics for these JSX-name cases), so each case is ported to its
 *    bare `code` / `options` / `errors`, evaluated to its final form.
 *  - `parserOptions` (ecmaFeatures.jsx) dropped — rslint resolves via tsconfig and
 *    the RuleTester routes JSX fixtures to `.tsx`.
 *  - Per-case `features` dropped (see below for the two that carried one).
 *
 * `@stylistic/jsx-pascal-case` is NOT fixable (meta.fixable === undefined), so no
 * upstream case pins `output`, and none is added here.
 *
 * Feature-tagged cases — both verified to parse cleanly under ts-go and to match
 * the upstream expectation, so both stay in the green set:
 *  - `<Modal:Header />` carried `features: ['jsx namespace']` (upstream skips it on
 *    the TS parser). ts-go parses the JSX namespaced name without a syntax error
 *    and the rule emits 0 diagnostics, matching upstream's `valid` expectation.
 *  - (jsx-pascal-case has no `['ts', 'no-babel']`-tagged case; that one lives in
 *    the jsx-props-no-multi-spaces test.)
 *
 * No `._css_` / `._json_` / `._markdown_` test files exist for this rule, and the
 * single `.test.ts` has exactly one `run()` block (no skipBabel block). No
 * suggestions are pinned anywhere in the upstream file. No KNOWN GAPS surfaced:
 * every fixture parses under ts-go and aligns with upstream.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-pascal-case', null as never, {
  valid: [
    {
      // The rule must not warn on components that start with a lowercase
      // because they are interpreted as HTML elements by React
      code: '<testcomponent />',
    },
    {
      code: '<testComponent />',
    },
    {
      code: '<test_component />',
    },
    {
      code: '<TestComponent />',
    },
    {
      code: '<CSSTransitionGroup />',
    },
    {
      code: '<BetterThanCSS />',
    },
    {
      code: '<TestComponent><div /></TestComponent>',
    },
    {
      code: '<Test1Component />',
    },
    {
      code: '<TestComponent1 />',
    },
    {
      code: '<T3StComp0Nent />',
    },
    {
      code: '<Éurströmming />',
    },
    {
      code: '<Año />',
    },
    {
      code: '<Søknad />',
    },
    {
      code: '<T />',
    },
    {
      code: '<YMCA />',
      options: [{ allowAllCaps: true }],
    },
    {
      code: '<TEST_COMPONENT />',
      options: [{ allowAllCaps: true }],
    },
    {
      code: '<Modal.Header />',
    },
    {
      code: '<qualification.T3StComp0Nent />',
    },
    {
      code: '<Modal:Header />',
    },
    {
      code: '<IGNORED />',
      options: [{ ignore: ['IGNORED'] }],
    },
    {
      code: '<Foo_DEPRECATED />',
      options: [{ ignore: ['*_D*D'] }],
    },
    {
      code: '<Foo_DEPRECATED />',
      options: [{ ignore: ['*_+(DEPRECATED|IGNORED)'] }],
    },
    {
      code: '<$ />',
    },
    {
      code: '<_ />',
    },
    {
      code: '<H1>Hello!</H1>',
    },
    {
      code: '<Typography.P />',
    },
    {
      code: '<Styled.h1 />',
      options: [{ allowNamespace: true }],
    },
    {
      code: '<_TEST_COMPONENT />',
      options: [{ allowAllCaps: true, allowLeadingUnderscore: true }],
    },
    {
      code: '<_TestComponent />',
      options: [{ allowLeadingUnderscore: true }],
    },
  ],

  invalid: [
    {
      code: '<Test_component />',
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: 'Test_component' },
        },
      ],
    },
    {
      code: '<TEST_COMPONENT />',
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: 'TEST_COMPONENT' },
        },
      ],
    },
    {
      code: '<YMCA />',
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: 'YMCA' },
        },
      ],
    },
    {
      code: '<_TEST_COMPONENT />',
      options: [{ allowAllCaps: true }],
      errors: [
        {
          messageId: 'usePascalOrSnakeCase',
          data: { name: '_TEST_COMPONENT' },
        },
      ],
    },
    {
      code: '<TEST_COMPONENT_ />',
      options: [{ allowAllCaps: true }],
      errors: [
        {
          messageId: 'usePascalOrSnakeCase',
          data: { name: 'TEST_COMPONENT_' },
        },
      ],
    },
    {
      code: '<TEST-COMPONENT />',
      options: [{ allowAllCaps: true }],
      errors: [
        {
          messageId: 'usePascalOrSnakeCase',
          data: { name: 'TEST-COMPONENT' },
        },
      ],
    },
    {
      code: '<__ />',
      options: [{ allowAllCaps: true }],
      errors: [
        {
          messageId: 'usePascalOrSnakeCase',
          data: { name: '__' },
        },
      ],
    },
    {
      code: '<_div />',
      options: [{ allowLeadingUnderscore: true }],
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: '_div' },
        },
      ],
    },
    {
      code: '<__ />',
      options: [{ allowAllCaps: true, allowLeadingUnderscore: true }],
      errors: [
        {
          messageId: 'usePascalOrSnakeCase',
          data: { name: '__' },
        },
      ],
    },
    {
      code: '<$a />',
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: '$a' },
        },
      ],
    },
    {
      code: '<Foo_DEPRECATED />',
      options: [{ ignore: ['*_FOO'] }],
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: 'Foo_DEPRECATED' },
        },
      ],
    },
    {
      code: '<Styled.h1 />',
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: 'h1' },
        },
      ],
    },
    {
      code: '<$Typography.P />',
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: '$Typography' },
        },
      ],
    },
    {
      code: '<STYLED.h1 />',
      options: [{ allowNamespace: true }],
      errors: [
        {
          messageId: 'usePascalCase',
          data: { name: 'STYLED' },
        },
      ],
    },
  ],
});
