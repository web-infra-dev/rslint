import { RuleTester } from '../rule-tester';

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

const bareMessage = (name: string) =>
  `${name}: This attribute is an invalid ARIA attribute.`;
const withSuggestions = (name: string, ...suggestions: string[]) =>
  `${bareMessage(name)} Did you mean to use ${suggestions.join(',')}?`;

// Each recognized aria attribute round-trips clean on a `<div>` element.
const basicValidityTests = ariaAttributes.map((prop) => ({
  code: `<div ${prop}="foobar" />`,
}));

new RuleTester().run('aria-props', null as never, {
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
      errors: [{ message: bareMessage('aria-') }],
    },
    {
      code: '<div aria-labeledby="foobar" />',
      errors: [
        { message: withSuggestions('aria-labeledby', 'aria-labelledby') },
      ],
    },
    {
      code: '<div aria-skldjfaria-klajsd="foobar" />',
      errors: [{ message: bareMessage('aria-skldjfaria-klajsd') }],
    },
    // Multiple invalid aria attributes on one element.
    {
      code: '<div aria-labeledby="x" aria-describeby="y" />',
      errors: [
        { message: withSuggestions('aria-labeledby', 'aria-labelledby') },
        { message: withSuggestions('aria-describeby', 'aria-describedby') },
      ],
    },
    // Mix of valid + invalid — only invalid reports.
    {
      code: '<div aria-hidden aria-foo="x" aria-label="y" />',
      errors: [{ message: bareMessage('aria-foo') }],
    },
    // Multi-line.
    {
      code: '<div\n  aria-typo="x"\n/>',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Boolean form of invalid aria-*.
    {
      code: '<div aria-foobar />',
      errors: [{ message: bareMessage('aria-foobar') }],
    },
    // Component (capitalized) with invalid aria-*.
    {
      code: '<MyComponent aria-typo="x" />',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Child JSX element carries the invalid attribute.
    {
      code: '<div><span aria-bogus="x" /></div>',
      errors: [{ message: bareMessage('aria-bogus') }],
    },
    // Inside Array.map callback.
    {
      code: 'function L({xs}) { return xs.map(x => <div aria-typo="y" key={x} />); }',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Self-closing without trailing space.
    {
      code: '<div aria-foo="x"/>',
      errors: [{ message: bareMessage('aria-foo') }],
    },
    // JSX fragment.
    {
      code: '<>{cond && <div aria-typo="y">x</div>}</>',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // TS type-assertion on value of invalid attribute — name extraction
    // is unaffected by value-side wrappers.
    {
      code: '<div aria-typo={"x" as const} />',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Suggestion ranking with a close typo to a long attribute.
    {
      code: '<div aria-describeby="x" />',
      errors: [
        { message: withSuggestions('aria-describeby', 'aria-describedby') },
      ],
    },
    // Lowercase prefix + uppercase suffix — `aria.has` is case-sensitive
    // so the lookup misses; `getSuggestion` upper-cases both sides, so
    // the canonical lowercase form surfaces as a suggestion.
    {
      code: '<div aria-LABEL="x" />',
      errors: [
        { message: withSuggestions('aria-LABEL', 'aria-label', 'aria-level') },
      ],
    },
    {
      code: '<div aria-Hidden />',
      errors: [{ message: withSuggestions('aria-Hidden', 'aria-hidden') }],
    },
    // Deeply nested JSX.
    {
      code: '<section><article><header><h1 aria-typo="x" /></header></article></section>',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Invalid attribute on outer AND inner — no bleed across boundaries.
    // Far-from-canonical names so suggestion ranking doesn't interfere.
    {
      code: '<div aria-zzzz="o"><span aria-qxqq="i" /></div>',
      errors: [
        { message: bareMessage('aria-zzzz') },
        { message: bareMessage('aria-qxqq') },
      ],
    },
    // Conditional rendering — JSX in a ternary arm.
    {
      code: 'function F({c}: {c: boolean}) { return c ? <div aria-typo /> : null; }',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // JSX as a call argument.
    {
      code: 'render(<div aria-typo="x" />);',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Distance-2 typo — still within THRESHOLD.
    {
      code: '<div aria-rowind="x" />',
      errors: [{ message: withSuggestions('aria-rowind', 'aria-rowindex') }],
    },
    // Distance-3 typo — past THRESHOLD, no suggestion suffix.
    {
      code: '<div aria-rowin="x" />',
      errors: [{ message: bareMessage('aria-rowin') }],
    },
    // Adjacent transposition typo — OSA distance 1 (swap i↔d from
    // `aria-hidden`). Locks the transposition arm of OSA.
    {
      code: '<div aria-hdiden />',
      errors: [{ message: withSuggestions('aria-hdiden', 'aria-hidden') }],
    },
    // Real-user typos with suggestions.
    {
      code: '<div aria-readyonly="true" />',
      errors: [{ message: withSuggestions('aria-readyonly', 'aria-readonly') }],
    },
    {
      code: '<div aria-busys="true" />',
      errors: [{ message: withSuggestions('aria-busys', 'aria-busy') }],
    },
    {
      code: '<div aria-checke="true" />',
      errors: [{ message: withSuggestions('aria-checke', 'aria-checked') }],
    },
    {
      code: '<div aria-haspup="menu" />',
      errors: [{ message: withSuggestions('aria-haspup', 'aria-haspopup') }],
    },
    // User confuses native HTML / React props with ARIA.
    {
      code: '<div aria-tabindex="0" />',
      errors: [{ message: bareMessage('aria-tabindex') }],
    },
    {
      code: '<div aria-onclick={handler} />',
      errors: [{ message: bareMessage('aria-onclick') }],
    },
    {
      code: '<div aria-class="x" />',
      errors: [{ message: bareMessage('aria-class') }],
    },
    {
      code: '<div aria-id="x" aria-style="color: red" />',
      errors: [
        { message: bareMessage('aria-id') },
        { message: bareMessage('aria-style') },
      ],
    },
    // AST shape edges: JSX in expression child / prop value / array /
    // object / arrow body / library wrappers.
    {
      code: '<div>{<span aria-typo="x" />}</div>',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    {
      code: '<Outer icon={<span aria-typo="x" />} />',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    {
      code: 'const xs = [<div key="a" aria-typo="x" />];',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    {
      code: 'const m = { a: <div aria-typo="x" /> };',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    {
      code: 'const f = () => <div aria-typo="x" />;',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    {
      code: 'const X = React.forwardRef((p, r) => <div ref={r} aria-typo="x" />);',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    {
      code: 'function F() { return React.useMemo(() => <div aria-typo="x" />, []); }',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    {
      code: '<Suspense fallback={<div aria-typo="x" />}><Page /></Suspense>',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Spread interleaved with invalid named attr — spread not inspected.
    {
      code: '<div {...a} aria-typo="x" {...b} />',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    {
      code: "<div {...{'aria-hidden': true}} aria-typo />",
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // settings cannot make `aria-typo` valid — rule doesn't read it.
    {
      code: '<Foo aria-typo="x" />',
      settings: { 'jsx-a11y': { components: { Foo: 'div' } } },
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Stress: many invalid + valid attrs interleaved.
    {
      code: '<div aria-foo="1" aria-hidden aria-bar="2" aria-label="x" aria-baz="3" aria-qux="4" aria-quux="5" />',
      errors: [
        { message: bareMessage('aria-foo') },
        { message: bareMessage('aria-bar') },
        { message: bareMessage('aria-baz') },
        { message: bareMessage('aria-qux') },
        { message: bareMessage('aria-quux') },
      ],
    },
    // Algorithm extremes.
    {
      code: '<div aria-thisisaveryverylongbogusattributename="x" />',
      errors: [
        { message: bareMessage('aria-thisisaveryverylongbogusattributename') },
      ],
    },
    {
      code: '<div aria-a="x" />',
      errors: [{ message: bareMessage('aria-a') }],
    },
    {
      code: '<div aria-foo123="x" />',
      errors: [{ message: bareMessage('aria-foo123') }],
    },
    {
      code: '<div aria-label- />',
      errors: [{ message: withSuggestions('aria-label-', 'aria-label') }],
    },
    {
      code: '<div aria--label />',
      errors: [{ message: withSuggestions('aria--label', 'aria-label') }],
    },
    // Deeply nested JSX with invalid attrs at multiple depths.
    {
      code: '<main><section><article><h2 aria-t1="a"><span aria-t2="b" /></h2></article></section></main>',
      errors: [
        { message: bareMessage('aria-t1') },
        { message: bareMessage('aria-t2') },
      ],
    },
    // Component with mixed custom + invalid aria attrs.
    {
      code: '<MyButton onClick={fn} disabled aria-typo title="x">label</MyButton>',
      errors: [{ message: bareMessage('aria-typo') }],
    },
    // Outer + inner invalid aria — both report.
    {
      code: '<div aria-outer="x" content={<span aria-inner="y" />}>k</div>',
      errors: [
        { message: bareMessage('aria-outer') },
        { message: bareMessage('aria-inner') },
      ],
    },
  ],
});
