import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

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

new RuleTester(config).run('no-aria-hidden-on-focusable', {
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
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<input aria-hidden="true" />;',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<a href="/" aria-hidden="true" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<button aria-hidden="true" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<textarea aria-hidden="true" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<p tabindex="0" aria-hidden="true">text</p>;',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- aria-hidden value-extraction → boolean true ----
    {
      code: '<button aria-hidden />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<button aria-hidden={true} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<button aria-hidden="True" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<button aria-hidden="TRUE" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<button aria-hidden={"true"} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<button aria-hidden={"True"} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // ---- Interactive element survey, no tabIndex ----
    {
      code: '<a href="#" aria-hidden="true" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<area href="#" aria-hidden="true" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<select aria-hidden="true" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },

    // ---- Non-interactive + tabIndex >= 0 ----
    {
      code: '<div aria-hidden="true" tabIndex={0} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<div aria-hidden="true" tabIndex={1} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<div aria-hidden="true" tabIndex={42} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<span aria-hidden="true" tabIndex="0">x</span>',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<p aria-hidden="true" tabIndex="2">x</p>',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- Boolean attribute form + interactive ----
    {
      code: '<input aria-hidden />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<a href="/" aria-hidden />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<select aria-hidden />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ---- TS wrappers around boolean true on interactive ----
    {
      code: '<button aria-hidden={(true)} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<button aria-hidden={true as boolean} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<input aria-hidden={(true)} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ---- Logical / conditional resolving to true ----
    {
      code: '<button aria-hidden={true && true} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<button aria-hidden={false || true} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<button aria-hidden={true ? true : false} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },

    // ---- components map remapping to interactive ----
    {
      code: '<Btn aria-hidden="true" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },

    // ---- polymorphicPropName resolving to interactive ----
    {
      code: '<Box as="input" aria-hidden="true" />',
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<Box as="button" aria-hidden="true" />',
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },

    // ---- Custom component with explicit positive tabIndex ----
    {
      code: '<MyButton aria-hidden="true" tabIndex={0} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '<MyButton aria-hidden="true" tabIndex="1" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '<UX.Layout aria-hidden="true" tabIndex={3} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },

    // ---- tabIndex via expressions that staticEval resolves ----
    {
      code: '<div aria-hidden="true" tabIndex={true ? 0 : -1} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<div aria-hidden="true" tabIndex={"5"} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },

    // ---- Lowercase tabindex on interactive (case-insensitive prop) ----
    {
      code: '<button aria-hidden="true" tabindex="0" />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },

    // ---- Nested elements report independently ----
    {
      code: '<button aria-hidden="true"><a href="/" aria-hidden="true">x</a></button>',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },

    // ---- Multiple offending elements ----
    {
      code: '<form><input aria-hidden="true" /><textarea aria-hidden="true" /></form>',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 35,
        },
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 66,
        },
      ],
    },

    // ---- Real-world component patterns ----
    {
      code: 'function Icon() { return <button aria-hidden="true">x</button>; }',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: 'function Foo({cond}) { return cond ? <input aria-hidden="true" /> : <div />; }',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 66,
        },
      ],
    },
    {
      code: 'const xs = arr.map((x) => <button aria-hidden="true" key={x}>{x}</button>);',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 62,
        },
      ],
    },
    {
      code: 'const Wrapped = withTracking(({v}) => <input aria-hidden="true" value={v} />);',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 77,
        },
      ],
    },
    {
      code: 'const x = <>{cond && <button aria-hidden="true" />}</>',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'function A() { return <input aria-hidden="true" />; }\nfunction B() { return <textarea aria-hidden="true" />; }',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 51,
        },
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 2,
          column: 23,
          endLine: 2,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'function* render() { yield <input aria-hidden="true" />; yield <textarea aria-hidden="true" />; }',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 56,
        },
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 64,
          endLine: 1,
          endColumn: 95,
        },
      ],
    },
    {
      code: 'async function render() { return <button aria-hidden="true">x</button>; }',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: 'const x = (() => <input aria-hidden="true" />)();',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: 'function Modal() { return <dialog open><button aria-hidden="true">OK</button></dialog>; }',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 67,
        },
      ],
    },
    {
      code: 'class Form extends React.Component { render() { return <input aria-hidden="true" />; } }',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 56,
          endLine: 1,
          endColumn: 84,
        },
      ],
    },

    // ---- TS generic JSX ----
    {
      code: '<Cell<{a: number}> aria-hidden="true" tabIndex={0} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },

    // ---- Comments don't suppress ----
    {
      code: '<input /* a */ aria-hidden="true" /* b */ />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 45,
        },
      ],
    },
    {
      code: '<button aria-hidden={/* truthy */ true} />',
      errors: [
        {
          messageId: 'noAriaHiddenOnFocusable',
          message: 'aria-hidden="true" must not be set on focusable elements.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
  ],
});
