import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Mirrors aria-query's `aria.keys()` — the canonical list of recognized
// ARIA states and properties. Used to drive procedural valid-test generation
// (`<div ${prop}="foobar" />` for every entry), matching upstream's
// `__tests__/src/rules/aria-props-test.js`.
const ariaAttributes = [
  'aria-activedescendant',
  'aria-atomic',
  'aria-autocomplete',
  'aria-braillelabel',
  'aria-brailleroledescription',
  'aria-busy',
  'aria-checked',
  'aria-colcount',
  'aria-colindex',
  'aria-colspan',
  'aria-controls',
  'aria-current',
  'aria-describedby',
  'aria-description',
  'aria-details',
  'aria-disabled',
  'aria-dropeffect',
  'aria-errormessage',
  'aria-expanded',
  'aria-flowto',
  'aria-grabbed',
  'aria-haspopup',
  'aria-hidden',
  'aria-invalid',
  'aria-keyshortcuts',
  'aria-label',
  'aria-labelledby',
  'aria-level',
  'aria-live',
  'aria-modal',
  'aria-multiline',
  'aria-multiselectable',
  'aria-orientation',
  'aria-owns',
  'aria-placeholder',
  'aria-posinset',
  'aria-pressed',
  'aria-readonly',
  'aria-relevant',
  'aria-required',
  'aria-roledescription',
  'aria-rowcount',
  'aria-rowindex',
  'aria-rowspan',
  'aria-selected',
  'aria-setsize',
  'aria-sort',
  'aria-valuemax',
  'aria-valuemin',
  'aria-valuenow',
  'aria-valuetext',
];

// Each recognized aria attribute round-trips clean on a `<div>` element.
const basicValidityTests = ariaAttributes.map((prop) => ({
  code: `<div ${prop}="foobar" />`,
}));

new RuleTester(config).run('aria-props', {
  valid: [
    // Upstream literal valid cases.
    { code: '<div />' },
    { code: '<div></div>' },
    { code: '<div aria="wee"></div>' },
    { code: '<div abcARIAdef="true"></div>' },
    { code: '<div fooaria-foobar="true"></div>' },
    { code: '<div fooaria-hidden="true"></div>' },
    { code: '<Bar baz />' },
    { code: '<input type="text" aria-errormessage="foobar" />' },
    // All recognized aria-* attributes on a <div>.
    ...basicValidityTests,
    // Extra rslint lockdowns:
    // Case-sensitive prefix — uppercase / mixed-case ARIA prefixes do not
    // enter the validation branch.
    { code: '<div ARIA-HIDDEN="true" />' },
    { code: '<div Aria-Hidden="true" />' },
    // Boolean form of a recognized aria attribute.
    { code: '<div aria-hidden />' },
    // JSXSpreadAttribute is not visited — invalid keys inside are skipped.
    { code: "<div {...{'aria-labeledby': 'x'}} />" },
    // JSXNamespacedName attribute — `aria:hidden` is not `aria-` prefixed.
    { code: '<svg aria:hidden="true" />' },
    // TS type-assertion wrappers on the value side — rule never reads value.
    { code: '<div aria-hidden={true as boolean} />' },
    { code: '<div aria-label={"x" as const} />' },
    // data-* is not aria-*.
    { code: '<div data-aria-hidden="true" />' },
    // Multi-line attribute layout.
    { code: '<div\n  aria-label="hi"\n  aria-hidden\n/>' },
    // Self-closing without trailing space before `/>`.
    { code: '<input aria-label="x"/>' },
    // Tag-name independence — rule fires on attributes regardless of tag.
    { code: '<Custom aria-hidden />' },
    { code: '<Foo.Bar aria-hidden />' },
    { code: '<my-element aria-hidden />' },
    // React patterns.
    {
      code: 'function L({xs}) { return xs.map(x => <div aria-hidden key={x} />); }',
    },
    { code: 'class C { render() { return <div aria-label="x" />; } }' },
    {
      code: 'function f<T>(x: T) { return <div aria-hidden />; }',
    },
    { code: '<>{cond && <div aria-hidden>x</div>}</>' },
    // Library wrappers — forwardRef / memo / useMemo / Suspense.
    {
      code: 'const Btn = React.forwardRef<HTMLButtonElement, {}>((props, ref) => <button ref={ref} aria-pressed="true" />);',
    },
    {
      code: 'const M = React.memo(function M() { return <div aria-label="x" />; });',
    },
    {
      code: 'function F() { const v = React.useMemo(() => <div aria-hidden>x</div>, []); return v; }',
    },
    {
      code: '<Suspense fallback={<div aria-busy="true">loading</div>}><Page aria-label="x" /></Suspense>',
    },
    // JSX in collection literals / object values / default params.
    { code: 'const renderMap = { ok: <div aria-live="polite" /> };' },
    {
      code: 'const arr = [<div key="1" aria-hidden />, <div key="2" aria-label="y" />];',
    },
    {
      code: 'function F({ child = <div aria-hidden /> }: { child?: any }) { return child; }',
    },
    { code: 'async function F() { return <div aria-busy="true" />; }' },
    { code: 'const f = () => <div aria-hidden />;' },
    { code: 'export default <div aria-label="x" />;' },
    {
      code: 'function F() { try { return <div aria-hidden />; } catch { return null; } }',
    },
    // Repeated identical aria attrs (syntactically legal JSX).
    { code: '<div aria-hidden="false" aria-hidden="true" />' },
    // Comments interleaved with attributes.
    {
      code: '<div /* lead */ aria-label="x" /* mid */ aria-hidden /* trail */ />',
    },
    // Multiple spreads interleaved with valid named attrs.
    { code: '<div {...a} aria-hidden {...b} aria-label="y" {...c} />' },
    // Render-prop pattern — JSX-in-JSX.
    {
      code: '<Outer renderIcon={<span aria-hidden />} aria-label="x" />',
    },
    // Settings supplied but rule does not consult them — must not panic
    // and must not change the verdict for an all-valid element.
    {
      code: '<Foo aria-hidden />',
      settings: {
        'jsx-a11y': {
          components: { Foo: 'div' },
          polymorphicPropName: 'as',
        },
      },
    },
    // Realistic compound elements with the recommended ARIA matrix.
    {
      code: '<button type="button" aria-pressed="true" aria-label="Save" aria-disabled="false" />',
    },
    {
      code: '<input type="checkbox" aria-checked="mixed" aria-required="true" />',
    },
    {
      code: '<ul role="listbox" aria-multiselectable="true" aria-activedescendant="opt-1" />',
    },
    // Computed / expression values — value side is never inspected.
    { code: '<div aria-level={depth + 1} aria-rowindex={index * 2} />' },
    { code: '<div aria-label={user?.name} />' },
    { code: '<div aria-label={`hello ${name}`} />' },
  ],
  invalid: [
    // Upstream literal invalid cases.
    {
      code: '<div aria-="foobar" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div aria-labeledby="foobar" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-labeledby: This attribute is an invalid ARIA attribute. Did you mean to use aria-labelledby?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div aria-skldjfaria-klajsd="foobar" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-skldjfaria-klajsd: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    // Multiple invalid aria attributes on one element.
    {
      code: '<div aria-labeledby="x" aria-describeby="y" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-labeledby: This attribute is an invalid ARIA attribute. Did you mean to use aria-labelledby?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-describeby: This attribute is an invalid ARIA attribute. Did you mean to use aria-describedby?',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    // Mix of valid + invalid — only invalid reports.
    {
      code: '<div aria-hidden aria-foo="x" aria-label="y" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-foo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    // Multi-line.
    {
      code: '<div\n  aria-typo="x"\n/>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 2,
          column: 3,
          endLine: 2,
          endColumn: 16,
        },
      ],
    },
    // Boolean form of invalid aria-*.
    {
      code: '<div aria-foobar />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-foobar: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    // Component (capitalized) with invalid aria-*.
    {
      code: '<MyComponent aria-typo="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    // Child JSX element carries the invalid attribute.
    {
      code: '<div><span aria-bogus="x" /></div>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-bogus: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    // Inside Array.map callback.
    {
      code: 'function L({xs}) { return xs.map(x => <div aria-typo="y" key={x} />); }',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 57,
        },
      ],
    },
    // Self-closing without trailing space.
    {
      code: '<div aria-foo="x"/>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-foo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    // JSX fragment.
    {
      code: '<>{cond && <div aria-typo="y">x</div>}</>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    // TS type-assertion on value of invalid attribute — name extraction
    // is unaffected by value-side wrappers.
    {
      code: '<div aria-typo={"x" as const} />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    // Suggestion ranking with a close typo to a long attribute.
    {
      code: '<div aria-describeby="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-describeby: This attribute is an invalid ARIA attribute. Did you mean to use aria-describedby?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    // Lowercase prefix + uppercase suffix — `aria.has` is case-sensitive
    // so the lookup misses; `getSuggestion` upper-cases both sides, so
    // the canonical lowercase form surfaces as a suggestion.
    {
      code: '<div aria-LABEL="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-LABEL: This attribute is an invalid ARIA attribute. Did you mean to use aria-label,aria-level?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div aria-Hidden />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-Hidden: This attribute is an invalid ARIA attribute. Did you mean to use aria-hidden?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    // Deeply nested JSX.
    {
      code: '<section><article><header><h1 aria-typo="x" /></header></article></section>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    // Invalid attribute on outer AND inner — no bleed across boundaries.
    // Far-from-canonical names so suggestion ranking doesn't interfere.
    {
      code: '<div aria-zzzz="o"><span aria-qxqq="i" /></div>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-zzzz: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
        {
          messageId: 'invalidAriaProp',
          message: 'aria-qxqq: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    // Conditional rendering — JSX in a ternary arm.
    {
      code: 'function F({c}: {c: boolean}) { return c ? <div aria-typo /> : null; }',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    // JSX as a call argument.
    {
      code: 'render(<div aria-typo="x" />);',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    // Distance-2 typo — still within THRESHOLD.
    {
      code: '<div aria-rowind="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-rowind: This attribute is an invalid ARIA attribute. Did you mean to use aria-rowindex?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    // Distance-3 typo — past THRESHOLD, no suggestion suffix.
    {
      code: '<div aria-rowin="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-rowin: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    // Adjacent transposition typo — OSA distance 1 (swap i↔d from
    // `aria-hidden`). Locks the transposition arm of OSA.
    {
      code: '<div aria-hdiden />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-hdiden: This attribute is an invalid ARIA attribute. Did you mean to use aria-hidden?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    // Real-user typos with suggestions.
    {
      code: '<div aria-readyonly="true" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-readyonly: This attribute is an invalid ARIA attribute. Did you mean to use aria-readonly?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div aria-busys="true" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-busys: This attribute is an invalid ARIA attribute. Did you mean to use aria-busy?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div aria-checke="true" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-checke: This attribute is an invalid ARIA attribute. Did you mean to use aria-checked?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div aria-haspup="menu" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-haspup: This attribute is an invalid ARIA attribute. Did you mean to use aria-haspopup?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    // User confuses native HTML / React props with ARIA.
    {
      code: '<div aria-tabindex="0" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-tabindex: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div aria-onclick={handler} />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-onclick: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div aria-class="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-class: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div aria-id="x" aria-style="color: red" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-id: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
        {
          messageId: 'invalidAriaProp',
          message: 'aria-style: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    // AST shape edges: JSX in expression child / prop value / array /
    // object / arrow body / library wrappers.
    {
      code: '<div>{<span aria-typo="x" />}</div>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<Outer icon={<span aria-typo="x" />} />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'const xs = [<div key="a" aria-typo="x" />];',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'const m = { a: <div aria-typo="x" /> };',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'const f = () => <div aria-typo="x" />;',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'const X = React.forwardRef((p, r) => <div ref={r} aria-typo="x" />);',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 64,
        },
      ],
    },
    {
      code: 'function F() { return React.useMemo(() => <div aria-typo="x" />, []); }',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: '<Suspense fallback={<div aria-typo="x" />}><Page /></Suspense>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    // Spread interleaved with invalid named attr — spread not inspected.
    {
      code: '<div {...a} aria-typo="x" {...b} />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: "<div {...{'aria-hidden': true}} aria-typo />",
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    // settings cannot make `aria-typo` valid — rule doesn't read it.
    {
      code: '<Foo aria-typo="x" />',
      settings: { 'jsx-a11y': { components: { Foo: 'div' } } },
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    // Stress: many invalid + valid attrs interleaved.
    {
      code: '<div aria-foo="1" aria-hidden aria-bar="2" aria-label="x" aria-baz="3" aria-qux="4" aria-quux="5" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-foo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
        {
          messageId: 'invalidAriaProp',
          message: 'aria-bar: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 43,
        },
        {
          messageId: 'invalidAriaProp',
          message: 'aria-baz: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 59,
          endLine: 1,
          endColumn: 71,
        },
        {
          messageId: 'invalidAriaProp',
          message: 'aria-qux: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 72,
          endLine: 1,
          endColumn: 84,
        },
        {
          messageId: 'invalidAriaProp',
          message: 'aria-quux: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 85,
          endLine: 1,
          endColumn: 98,
        },
      ],
    },
    // Algorithm extremes.
    {
      code: '<div aria-thisisaveryverylongbogusattributename="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-thisisaveryverylongbogusattributename: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<div aria-a="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-a: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<div aria-foo123="x" />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-foo123: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div aria-label- />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria-label-: This attribute is an invalid ARIA attribute. Did you mean to use aria-label?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<div aria--label />',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message:
            'aria--label: This attribute is an invalid ARIA attribute. Did you mean to use aria-label?',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    // Deeply nested JSX with invalid attrs at multiple depths.
    {
      code: '<main><section><article><h2 aria-t1="a"><span aria-t2="b" /></h2></article></section></main>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-t1: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 40,
        },
        {
          messageId: 'invalidAriaProp',
          message: 'aria-t2: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },
    // Component with mixed custom + invalid aria attrs.
    {
      code: '<MyButton onClick={fn} disabled aria-typo title="x">label</MyButton>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-typo: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    // Outer + inner invalid aria — both report.
    {
      code: '<div aria-outer="x" content={<span aria-inner="y" />}>k</div>',
      errors: [
        {
          messageId: 'invalidAriaProp',
          message: 'aria-outer: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'invalidAriaProp',
          message: 'aria-inner: This attribute is an invalid ARIA attribute.',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
  ],
});
