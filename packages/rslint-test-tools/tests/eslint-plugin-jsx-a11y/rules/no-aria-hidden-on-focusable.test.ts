import { RuleTester } from '../rule-tester';

const errorMessage =
  'aria-hidden="true" must not be set on focusable elements.';
const expectedError = { message: errorMessage };

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Btn: 'button',
    },
  },
};

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

new RuleTester().run('no-aria-hidden-on-focusable', null as never, {
  valid: [
    // ---- Upstream valid ----
    { code: '<div aria-hidden="true" />;' },
    { code: '<div onClick={() => void 0} aria-hidden="true" />;' },
    { code: '<img aria-hidden="true" />' },
    { code: '<a aria-hidden="false" href="#" />' },
    { code: '<button aria-hidden="true" tabIndex="-1" />' },
    { code: '<button />' },
    { code: '<a href="/" />' },

    // ---- aria-hidden does NOT resolve to true → safe regardless of focus ----
    { code: '<input />' },
    { code: '<button aria-hidden={false} />' },
    { code: '<input aria-hidden={false} />' },
    { code: '<button aria-hidden="false" />' },
    { code: '<button aria-hidden="FALSE" />' },
    { code: '<button aria-hidden={"false"} />' },
    { code: '<button aria-hidden={undefined} />' },
    { code: '<button aria-hidden={0} />' },
    { code: '<button aria-hidden="yes" />' },
    { code: '<button aria-hidden={someVar} />' },
    { code: '<button aria-hidden={obj.x} />' },
    { code: '<button aria-hidden={fn()} />' },
    // TSNonNullExpression stringifies the operand with `!` → "true!" !== true.
    { code: '<button aria-hidden={true!} />' },
    { code: '<input aria-hidden={true!} />' },

    // ---- Non-interactive elements without tabIndex ----
    { code: '<div aria-hidden />' },
    { code: '<div aria-hidden={true} />' },
    { code: '<div aria-hidden="True" />' },
    { code: '<span aria-hidden="true" />' },
    { code: '<p aria-hidden="true" />' },
    { code: '<section aria-hidden="true" />' },
    { code: '<br aria-hidden="true" />' },
    { code: '<hr aria-hidden="true" />' },

    // ---- Non-interactive with tabIndex < 0 ----
    { code: '<div aria-hidden="true" tabIndex="-1" />' },
    { code: '<div aria-hidden="true" tabIndex={-1} />' },

    // ---- Interactive with tabIndex < 0 ----
    { code: '<button aria-hidden="true" tabIndex={-1} />' },
    { code: '<button aria-hidden="true" tabIndex="-2" />' },
    { code: '<input aria-hidden="true" tabIndex={-1} />' },
    { code: '<textarea aria-hidden="true" tabIndex={-1} />' },
    { code: '<select aria-hidden="true" tabIndex={-1} />' },
    { code: '<a href="/" aria-hidden="true" tabIndex={-1} />' },

    // ---- Custom components without explicit positive tabIndex ----
    { code: '<MyButton aria-hidden="true" />' },
    { code: '<UX.Layout aria-hidden="true" />' },
    { code: '<svg:circle aria-hidden="true" />' },
    { code: '<Foo.Bar.Baz aria-hidden="true" />' },
    { code: '<MyButton aria-hidden="true" tabIndex={-1} />' },

    // ---- <a> without href is non-interactive ----
    { code: '<a aria-hidden="true" />' },
    { code: '<area aria-hidden="true" />' },

    // ---- tabIndex extracts to undefined → non-interactive arm rejects ----
    { code: '<div aria-hidden="true" tabIndex="" />' },
    { code: '<div aria-hidden="true" tabIndex={undefined} />' },
    { code: '<div aria-hidden="true" tabIndex={true} />' },
    { code: '<div aria-hidden="true" tabIndex={1.5} />' },

    // ---- components map: non-DOM → interactive resolution ----
    {
      code: '<Btn aria-hidden="true" tabIndex={-1} />',
      settings: componentsSettings,
    },

    // ---- polymorphicPropName ----
    {
      code: '<Box as="button" aria-hidden="true" tabIndex={-1} />',
      settings: polymorphicSettings,
    },

    // ---- TS wrappers around the aria-hidden value, but element non-focusable ----
    { code: '<div aria-hidden={(true)} />' },
    { code: '<div aria-hidden={true as boolean} />' },

    // ---- Spread literal cannot match aria-hidden (Identifier key only, hyphen) ----
    { code: '<div {...{"aria-hidden": true}} />' },
    { code: '<button {...{"aria-hidden": true}} />' },
    { code: '<input {...{"aria-hidden": true}} />' },
    { code: '<button {...props} />' },

    // ---- Empty / unresolvable expressions ----
    { code: '<button aria-hidden={} />' },

    // ---- Real-world: explicit tabIndex={-1} suppresses ----
    {
      code: 'function Decoration() { return <button aria-hidden="true" tabIndex={-1}>icon</button>; }',
    },

    // ---- Comments around / inside the prop ----
    { code: '<div /* before */ aria-hidden="true" /* after */ />' },
    { code: '<div aria-hidden={/* true */ true} />' },
  ],
  invalid: [
    // ---- Upstream invalid ----
    {
      code: '<div aria-hidden="true" tabIndex="0" />;',
      errors: [expectedError],
    },
    { code: '<input aria-hidden="true" />;', errors: [expectedError] },
    { code: '<a href="/" aria-hidden="true" />', errors: [expectedError] },
    { code: '<button aria-hidden="true" />', errors: [expectedError] },
    { code: '<textarea aria-hidden="true" />', errors: [expectedError] },
    {
      code: '<p tabindex="0" aria-hidden="true">text</p>;',
      errors: [expectedError],
    },

    // ---- aria-hidden value-extraction → boolean true ----
    { code: '<button aria-hidden />', errors: [expectedError] },
    { code: '<button aria-hidden={true} />', errors: [expectedError] },
    { code: '<button aria-hidden="True" />', errors: [expectedError] },
    { code: '<button aria-hidden="TRUE" />', errors: [expectedError] },
    { code: '<button aria-hidden={"true"} />', errors: [expectedError] },
    { code: '<button aria-hidden={"True"} />', errors: [expectedError] },

    // ---- Interactive element survey, no tabIndex ----
    { code: '<a href="#" aria-hidden="true" />', errors: [expectedError] },
    { code: '<area href="#" aria-hidden="true" />', errors: [expectedError] },
    { code: '<select aria-hidden="true" />', errors: [expectedError] },

    // ---- Non-interactive + tabIndex >= 0 ----
    {
      code: '<div aria-hidden="true" tabIndex={0} />',
      errors: [expectedError],
    },
    {
      code: '<div aria-hidden="true" tabIndex={1} />',
      errors: [expectedError],
    },
    {
      code: '<div aria-hidden="true" tabIndex={42} />',
      errors: [expectedError],
    },
    {
      code: '<span aria-hidden="true" tabIndex="0">x</span>',
      errors: [expectedError],
    },
    {
      code: '<p aria-hidden="true" tabIndex="2">x</p>',
      errors: [expectedError],
    },

    // ---- Boolean attribute form + interactive ----
    { code: '<input aria-hidden />', errors: [expectedError] },
    { code: '<a href="/" aria-hidden />', errors: [expectedError] },
    { code: '<select aria-hidden />', errors: [expectedError] },

    // ---- TS wrappers around boolean true on interactive ----
    { code: '<button aria-hidden={(true)} />', errors: [expectedError] },
    {
      code: '<button aria-hidden={true as boolean} />',
      errors: [expectedError],
    },
    { code: '<input aria-hidden={(true)} />', errors: [expectedError] },

    // ---- Logical / conditional resolving to true ----
    {
      code: '<button aria-hidden={true && true} />',
      errors: [expectedError],
    },
    {
      code: '<button aria-hidden={false || true} />',
      errors: [expectedError],
    },
    {
      code: '<button aria-hidden={true ? true : false} />',
      errors: [expectedError],
    },

    // ---- components map remapping to interactive ----
    {
      code: '<Btn aria-hidden="true" />',
      settings: componentsSettings,
      errors: [expectedError],
    },

    // ---- polymorphicPropName resolving to interactive ----
    {
      code: '<Box as="input" aria-hidden="true" />',
      settings: polymorphicSettings,
      errors: [expectedError],
    },
    {
      code: '<Box as="button" aria-hidden="true" />',
      settings: polymorphicSettings,
      errors: [expectedError],
    },

    // ---- Custom component with explicit positive tabIndex ----
    {
      code: '<MyButton aria-hidden="true" tabIndex={0} />',
      errors: [expectedError],
    },
    {
      code: '<MyButton aria-hidden="true" tabIndex="1" />',
      errors: [expectedError],
    },
    {
      code: '<UX.Layout aria-hidden="true" tabIndex={3} />',
      errors: [expectedError],
    },

    // ---- tabIndex via expressions that staticEval resolves ----
    {
      code: '<div aria-hidden="true" tabIndex={true ? 0 : -1} />',
      errors: [expectedError],
    },
    {
      code: '<div aria-hidden="true" tabIndex={"5"} />',
      errors: [expectedError],
    },

    // ---- Lowercase tabindex on interactive (case-insensitive prop) ----
    {
      code: '<button aria-hidden="true" tabindex="0" />',
      errors: [expectedError],
    },

    // ---- Nested elements report independently ----
    {
      code: '<button aria-hidden="true"><a href="/" aria-hidden="true">x</a></button>',
      errors: [expectedError, expectedError],
    },

    // ---- Multiple offending elements ----
    {
      code: '<form><input aria-hidden="true" /><textarea aria-hidden="true" /></form>',
      errors: [expectedError, expectedError],
    },

    // ---- Real-world component patterns ----
    {
      code: 'function Icon() { return <button aria-hidden="true">x</button>; }',
      errors: [expectedError],
    },
    {
      code: 'function Foo({cond}) { return cond ? <input aria-hidden="true" /> : <div />; }',
      errors: [expectedError],
    },
    {
      code: 'const xs = arr.map((x) => <button aria-hidden="true" key={x}>{x}</button>);',
      errors: [expectedError],
    },
    {
      code: 'const Wrapped = withTracking(({v}) => <input aria-hidden="true" value={v} />);',
      errors: [expectedError],
    },
    {
      code: 'const x = <>{cond && <button aria-hidden="true" />}</>',
      errors: [expectedError],
    },
    {
      code: 'function A() { return <input aria-hidden="true" />; }\nfunction B() { return <textarea aria-hidden="true" />; }',
      errors: [expectedError, expectedError],
    },
    {
      code: 'function* render() { yield <input aria-hidden="true" />; yield <textarea aria-hidden="true" />; }',
      errors: [expectedError, expectedError],
    },
    {
      code: 'async function render() { return <button aria-hidden="true">x</button>; }',
      errors: [expectedError],
    },
    {
      code: 'const x = (() => <input aria-hidden="true" />)();',
      errors: [expectedError],
    },
    {
      code: 'function Modal() { return <dialog open><button aria-hidden="true">OK</button></dialog>; }',
      errors: [expectedError],
    },
    {
      code: 'class Form extends React.Component { render() { return <input aria-hidden="true" />; } }',
      errors: [expectedError],
    },

    // ---- TS generic JSX ----
    {
      code: '<Cell<{a: number}> aria-hidden="true" tabIndex={0} />',
      errors: [expectedError],
    },

    // ---- Comments don't suppress ----
    {
      code: '<input /* a */ aria-hidden="true" /* b */ />',
      errors: [expectedError],
    },
    {
      code: '<button aria-hidden={/* truthy */ true} />',
      errors: [expectedError],
    },
  ],
});
