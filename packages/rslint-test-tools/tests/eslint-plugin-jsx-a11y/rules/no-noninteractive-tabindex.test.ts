import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Article: 'article',
      MyButton: 'button',
    },
  },
};

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

const polymorphicAllowListSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
    polymorphicAllowList: ['Box'],
  },
};

const recommendedOptions = [
  {
    tags: [],
    roles: ['tabpanel'],
    allowExpressionValues: true,
  },
];

const allowExpressionValuesTrueOptions = [{ allowExpressionValues: true }];
const allowExpressionValuesFalseOptions = [{ allowExpressionValues: false }];
const tagsExemptDivOptions = [{ tags: ['div'] }];
const rolesExemptTabpanelOptions = [{ roles: ['tabpanel'] }];

new RuleTester(config).run('no-noninteractive-tabindex', {
  valid: [
    // ============================================================
    // Upstream alwaysValid (run under both :strict and :recommended)
    // ============================================================
    { code: '<MyButton tabIndex={0} />' },
    { code: '<button />' },
    { code: '<button tabIndex="0" />' },
    { code: '<button tabIndex={0} />' },
    { code: '<div />' },
    { code: '<div tabIndex="-1" />' },
    { code: '<div role="button" tabIndex="0" />' },
    { code: '<div role="article" tabIndex="-1" />' },
    { code: '<article tabIndex="-1" />' },
    { code: '<Article tabIndex="-1" />', settings: componentsSettings },
    { code: '<MyButton tabIndex={0} />', settings: componentsSettings },

    // ============================================================
    // Upstream recommended-only valid
    // ============================================================
    {
      code: '<div role="tabpanel" tabIndex="0" />',
      options: recommendedOptions,
    },
    {
      code: '<div role={ROLE_BUTTON} onClick={() => {}} tabIndex="0" />',
      options: recommendedOptions,
    },
    {
      code: '<div role={BUTTON} onClick={() => {}} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={isButton ? "button" : "link"} onClick={() => {}} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={isButton ? "button" : LINK} onClick={() => {}} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={isButton ? BUTTON : LINK} onClick={() => {}} tabIndex="0"/>',
      options: allowExpressionValuesTrueOptions,
    },

    // ============================================================
    // Inherent-interactive elements (full survey)
    // ============================================================
    { code: '<input tabIndex={0} />' },
    { code: '<input type="text" tabIndex={0} />' },
    { code: '<input type="button" tabIndex={0} />' },
    { code: '<input type="submit" tabIndex={0} />' },
    { code: '<input type="reset" tabIndex={0} />' },
    { code: '<input type="image" tabIndex={0} />' },
    { code: '<input type="checkbox" tabIndex={0} />' },
    { code: '<input type="radio" tabIndex={0} />' },
    { code: '<input type="range" tabIndex={0} />' },
    { code: '<input type="number" tabIndex={0} />' },
    { code: '<input type="email" list="x" tabIndex={0} />' },
    { code: '<input type="search" list="x" tabIndex={0} />' },
    { code: '<input type="color" tabIndex={0} />' },
    { code: '<input type="date" tabIndex={0} />' },
    { code: '<textarea tabIndex={0} />' },
    { code: '<select tabIndex={0} />' },
    { code: '<select multiple size={2} tabIndex={0} />' },
    { code: '<a href="x" tabIndex={0} />' },
    { code: '<a href="" tabIndex={0} />' },
    { code: '<a href tabIndex={0} />' }, // boolean form satisfies "href present"
    { code: '<area href="x" tabIndex={0} />' },
    { code: '<th tabIndex={0} />' },
    { code: '<th scope="col" tabIndex={0} />' },
    { code: '<td tabIndex={0} />' },
    { code: '<tr tabIndex={0} />' },
    { code: '<datalist tabIndex={0} />' },
    { code: '<option tabIndex={0} />' },
    { code: '<audio tabIndex={0} />' },
    { code: '<video tabIndex={0} />' },
    { code: '<canvas tabIndex={0} />' },
    { code: '<embed tabIndex={0} />' },
    { code: '<menuitem tabIndex={0} />' },
    { code: '<summary tabIndex={0} />' },

    // ============================================================
    // Interactive role overrides — full widget-role survey
    // ============================================================
    { code: '<div role="button" tabIndex={0} />' },
    { code: '<div role="checkbox" tabIndex={0} />' },
    { code: '<div role="combobox" tabIndex={0} />' },
    { code: '<div role="grid" tabIndex={0} />' },
    { code: '<div role="link" tabIndex={0} />' },
    { code: '<div role="listbox" tabIndex={0} />' },
    { code: '<div role="menu" tabIndex={0} />' },
    { code: '<div role="menubar" tabIndex={0} />' },
    { code: '<div role="menuitem" tabIndex={0} />' },
    { code: '<div role="option" tabIndex={0} />' },
    { code: '<div role="radio" tabIndex={0} />' },
    { code: '<div role="searchbox" tabIndex={0} />' },
    { code: '<div role="slider" tabIndex={0} />' },
    { code: '<div role="spinbutton" tabIndex={0} />' },
    { code: '<div role="switch" tabIndex={0} />' },
    { code: '<div role="tab" tabIndex={0} />' },
    { code: '<div role="tablist" tabIndex={0} />' },
    { code: '<div role="textbox" tabIndex={0} />' },
    { code: '<div role="tree" tabIndex={0} />' },
    { code: '<div role="treegrid" tabIndex={0} />' },
    { code: '<div role="treeitem" tabIndex={0} />' },
    { code: '<div role="toolbar" tabIndex={0} />' }, // upstream forces interactive
    // First-token rule: `button menu` → button (interactive).
    { code: '<div role="button menu" tabIndex={0} />' },
    { code: '<div role="link button" tabIndex={0} />' },
    { code: '<div role="button " tabIndex={0} />' }, // trailing space
    // Case-insensitive on the role value.
    { code: '<div role="BUTTON" tabIndex={0} />' },
    { code: '<div role="MENUITEM" tabIndex={0} />' },

    // ============================================================
    // tabIndex < 0 — always allowed
    // ============================================================
    { code: '<div tabIndex={-1} />' },
    { code: '<div tabIndex={-100} />' },
    { code: '<div tabIndex={-2147483648} />' },
    { code: '<article tabIndex="-2" />' },
    { code: '<div tabIndex=" -1 " />' },
    // Conditional / Logical resolving negative — staticEval-fallback path.
    { code: '<div tabIndex={cond ? -1 : -2} />' },
    { code: '<div tabIndex={false || -1} />' },
    { code: '<div tabIndex={x ?? -1} />' },

    // ============================================================
    // tabIndex undefined-resolving shapes
    // ============================================================
    { code: '<div tabIndex />' },
    { code: '<div tabIndex={true} />' },
    { code: '<div tabIndex={false} />' },
    { code: '<div tabIndex="true" />' },
    { code: '<div tabIndex="false" />' },
    { code: '<div tabIndex="True" />' },
    { code: '<div tabIndex="FALSE" />' },
    { code: '<div tabIndex="" />' },
    { code: '<div tabIndex={""} />' },
    { code: '<div tabIndex={1.5} />' }, // non-integer
    { code: '<div tabIndex={-0.5} />' },
    { code: '<div tabIndex={1e-5} />' },
    { code: '<div tabIndex="1.5" />' },
    { code: '<div tabIndex="0.5" />' },
    // BigInt 0n/1n/positive are INVALID per upstream (Number(BigInt)
    // coerces — see invalid section). Only negative BigInts skip here.
    { code: '<div tabIndex={-1n} />' },
    // progressbar role — widget-class in aria-query → interactive → skip.
    { code: '<div role="progressbar" tabIndex={0} />' },
    // TSNonNullExpression → stringifies to "0!" → NaN → undefined → skip.
    { code: '<div tabIndex={0!} />' },
    { code: '<div tabIndex={(0)!} />' },
    { code: '<div tabIndex={(5)!} />' },
    { code: '<div tabIndex={someVar} />' },
    { code: '<div tabIndex={fn()} />' },
    { code: '<div tabIndex={obj.x} />' },
    { code: '<div tabIndex={obj?.x} />' },
    { code: '<div tabIndex="abc" />' },
    // `-Infinity` < 0 → not reported. `Infinity` is reported (covered in
    // invalid below — upstream `Infinity >= 0` is true).
    { code: '<div tabIndex={-Infinity} />' },
    // Signed hex / oct / bin → JS Number rejects → NaN → not reported.
    { code: '<div tabIndex="-0x10" />' },
    { code: '<div tabIndex="+0x10" />' },
    // Malformed hex / oct / bin → ParseUint fails → not reported.
    { code: '<div tabIndex="0x" />' },
    { code: '<div tabIndex="0xZZ" />' },
    { code: '<div tabIndex={NaN} />' },
    { code: '<div tabIndex={undefined} />' },
    // null literal → step 1 jvString "null" (LITERAL_TYPES special) → parseFloat fails → undefined.
    { code: '<div tabIndex={null} />' },

    // ============================================================
    // TS wrappers
    // ============================================================
    { code: '<div tabIndex={(-1)} />' },
    { code: '<div tabIndex={((-1))} />' },
    { code: '<div tabIndex={-1 as number} />' },
    { code: '<div tabIndex={(-1) as number} />' },
    { code: '<div tabIndex={(-1)!} />' },
    { code: '<div tabIndex={("-1")} />' },
    { code: '<div tabIndex={("-1") as string} />' },
    // `<div tabIndex={X satisfies T} />` is INVALID under upstream
    // (TSSatisfiesExpression → null → `null >= 0` true → REPORT). See
    // the invalid section below for the lock-in.

    // ============================================================
    // Options matrix
    // ============================================================
    { code: '<div tabIndex="0" />', options: tagsExemptDivOptions },
    {
      code: '<div role="article" tabIndex="0" />',
      options: tagsExemptDivOptions,
    },
    {
      code: '<div role="tabpanel" tabIndex="0" />',
      options: rolesExemptTabpanelOptions,
    },
    {
      code: '<article role="tabpanel" tabIndex="0" />',
      options: rolesExemptTabpanelOptions,
    },
    // allowExpressionValues=true with diverse non-literal role shapes.
    {
      code: '<div role={SOME_ROLE} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={getRole()} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={obj.role} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={obj?.role} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={"button"} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={r || "button"} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={r ?? "button"} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    {
      code: '<div role={cond ? "button" : OTHER} tabIndex="0" />',
      options: allowExpressionValuesTrueOptions,
    },
    // Empty option object / array.
    { code: '<div tabIndex={-1} />', options: [{}] },
    { code: '<div tabIndex={-1} />', options: [] },

    // ============================================================
    // Custom components (skip via dom.has=false)
    // ============================================================
    { code: '<UX.Layout tabIndex={0} />' },
    { code: '<svg:circle tabIndex={0} />' },
    { code: '<this.Foo tabIndex={0} />' },
    { code: '<Foo.Bar.Baz tabIndex={0} />' },
    { code: '<my-element tabIndex={0} />' },
    { code: '<x-y-z tabIndex={0} />' },

    // ============================================================
    // TypeScript generic JSX
    // ============================================================
    { code: '<List<string> tabIndex={0} />' },
    { code: '<Cell<{a: number}> tabIndex={0} />' },
    { code: '<Map<string, number> tabIndex={0} />' },

    // ============================================================
    // polymorphicPropName remap to interactive tag
    // ============================================================
    {
      code: '<Box as="button" tabIndex={0} />',
      settings: polymorphicSettings,
    },
    {
      code: '<Box as="a" href="x" tabIndex={0} />',
      settings: polymorphicSettings,
    },
    {
      code: '<Box as="textarea" tabIndex={0} />',
      settings: polymorphicSettings,
    },
    {
      code: '<Box as="select" tabIndex={0} />',
      settings: polymorphicSettings,
    },
    {
      code: '<Box as="input" type="checkbox" tabIndex={0} />',
      settings: polymorphicSettings,
    },
    // polymorphicAllowList: `Other` not in list → swap skipped → resolved = "Other" → not DOM → skip.
    {
      code: '<Other as="article" tabIndex={0} />',
      settings: polymorphicAllowListSettings,
    },
    {
      code: '<Box as="button" tabIndex={0} />',
      settings: polymorphicAllowListSettings,
    },

    // ============================================================
    // Comments around / inside the prop
    // ============================================================
    { code: '<div /* before */ tabIndex={-1} /* after */ />' },
    { code: '<div tabIndex={/* explicit */ -1} />' },
    { code: '<div role={/* literal-string */ "button"} tabIndex={0} />' },

    // ============================================================
    // Spread literal
    // ============================================================
    { code: '<div {...props} />' },
    { code: '<div {...a} {...b} tabIndex={-1} />' },
    { code: '<div {...{tabIndex: -1}} />' },
    { code: '<div {...{role: "button", tabIndex: 0}} />' },

    // ============================================================
    // Real-world a11y patterns (no-report)
    // ============================================================
    {
      code: 'function Tabs() { return <div role="tablist"><button role="tab" tabIndex={0}>One</button><button role="tab" tabIndex={-1}>Two</button></div>; }',
    },
    {
      code: 'function Combobox() { return <div role="combobox"><input role="searchbox" tabIndex={0} /><ul role="listbox"><li role="option" tabIndex={-1}>A</li></ul></div>; }',
    },
    {
      code: 'function Tree() { return <ul role="tree"><li role="treeitem" tabIndex={0}>Root</li></ul>; }',
    },
    {
      code: 'function Modal({ open }) { return open ? <dialog open><button autoFocus tabIndex={0}>Close</button></dialog> : null; }',
    },
    {
      code: 'function Search() { const ref = useRef(null); useEffect(() => ref.current?.focus(), []); return <input ref={ref} />; }',
    },
    {
      code: 'const Inp = React.forwardRef((props, ref) => <input ref={ref} tabIndex={0} {...props} />);',
    },
    {
      code: 'const Item = React.memo(({ id }) => <button id={id} tabIndex={0}>{id}</button>);',
    },
  ],
  invalid: [
    // ============================================================
    // Opaque expression types — upstream returns null which passes the
    // `typeof === 'undefined'` guard, then `null >= 0` ToNumber-coerces
    // to `0 >= 0` = true → REPORT. Aligned via GetTabIndexEx's nullLike
    // arm. Locks against the lossy pre-Ex behavior that silently skipped.
    // ============================================================
    {
      code: '<div tabIndex={-1 satisfies number} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<div tabIndex={5 satisfies number} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'async function f() { return <div tabIndex={await p} />; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'function* g() { yield <div tabIndex={yield 0} />; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },

    // ============================================================
    // BigInt — Number(BigInt) coerces. 0n → 0 → REPORT; 1n → 1 → REPORT.
    // ============================================================
    {
      code: '<div tabIndex={0n} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={1n} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={2n} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={5n} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={true ? 1n : 0n} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // Empty JsxExpression `tabIndex={}` — JSXEmptyExpression → null →
    // `null >= 0` true → REPORT.
    // ============================================================
    {
      code: '<div tabIndex={} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ============================================================
    // Upstream neverValid
    // ============================================================
    {
      code: '<div tabIndex="0" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div role="article" tabIndex="0" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<article tabIndex="0" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<article tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<Article tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
      settings: componentsSettings,
    },

    // ============================================================
    // Upstream strict-only invalid
    // ============================================================
    {
      code: '<div role="tabpanel" tabIndex="0" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<div role={ROLE_BUTTON} onClick={() => {}} tabIndex="0" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: '<div role={BUTTON} onClick={() => {}} tabIndex="0" />',
      options: allowExpressionValuesFalseOptions,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<div role={isButton ? "button" : "link"} onClick={() => {}} tabIndex="0" />',
      options: allowExpressionValuesFalseOptions,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 73,
        },
      ],
    },

    // ============================================================
    // Listener boundary — nested elements report independently
    // ============================================================
    {
      code: '<article tabIndex={0}><span tabIndex={0} /></article>',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<article tabIndex={0}><section tabIndex={0}><span tabIndex={0} /></section></article>',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 44,
        },
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      code: '<button tabIndex={0}><span tabIndex={0} /></button>',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 40,
        },
      ], // outer interactive, inner reports
    },
    {
      code: '<><div tabIndex={0} /><span tabIndex={0} /></>',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },

    // ============================================================
    // Non-interactive element survey
    // ============================================================
    {
      code: '<div tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<span tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<section tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<header tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<footer tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<main tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<aside tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<nav tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<p tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<h1 tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<h6 tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<ul tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<ol tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<li tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<table tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<dl tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<details tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<dialog tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<fieldset tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    // `<a />` (no href) is non-interactive.
    {
      code: '<a tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<area tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },

    // ============================================================
    // Non-interactive role values
    // ============================================================
    {
      code: '<div role="article" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div role="document" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<div role="img" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div role="region" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div role="tabpanel" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    // Note: `<div role="progressbar" tabIndex={0} />` is VALID per
    // upstream — progressbar is widget-class in aria-query → interactive →
    // rule skips. Moved to valid section.
    {
      code: '<div role="alert" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div role="banner" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div role="navigation" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<div role="presentation" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<div role="none" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    // First-token rule: `article button` — first wins → non-interactive.
    {
      code: '<div role="article button" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<div role="madeUpRole" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<div role="" tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div role="   " tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },

    // ============================================================
    // Non-literal role + allowExpressionValues=false
    // ============================================================
    {
      code: '<div role={SOME_ROLE} tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<div role={SOME_ROLE} tabIndex={0} />',
      options: allowExpressionValuesFalseOptions,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    // `role={undefined}` is special-cased — does NOT count as non-literal.
    {
      code: '<div role={undefined} tabIndex={0} />',
      options: allowExpressionValuesTrueOptions,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    // `role={null}` → JSXExpressionContainer with NullKeyword → not undefined,
    // not JSXText → IsNonLiteralProperty=true → exempt under
    // allowExpressionValues=true. Without the option, IsInteractiveRole
    // returns false → reports.
    {
      code: '<div role={null} tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    // Non-string literal role values.
    {
      code: '<div role={0} tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div role={true} tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div role={false} tabIndex={0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // tabIndex value variants ≥ 0 — full numeric/string survey
    // ============================================================
    {
      code: '<div tabIndex={1} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabIndex={5} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabIndex={2147483647} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    // Hex / Octal / Binary integer literals.
    {
      code: '<div tabIndex={0x10} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={0o10} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={0b10} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    // Hex / oct / bin STRING forms: upstream Number("0x10") = 16 → reports.
    {
      code: '<div tabIndex="0x10" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="0X10" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="0o10" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="0b10" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex="0xff" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    // Step-2 fallback exercises the same hex/oct/bin dispatch.
    {
      code: '<div tabIndex={true ? "0x10" : "-1"} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    // `Infinity` identifier — `Infinity >= 0` is true → reports.
    {
      code: '<div tabIndex={Infinity} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    // ArrayLiteralExpression — Array.prototype.join + ToNumber.
    {
      code: '<div tabIndex={[]} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    }, // "" → 0
    {
      code: '<div tabIndex={[5]} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    }, // "5" → 5
    {
      code: '<div tabIndex={[0]} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={[null]} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div tabIndex={[Infinity]} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    // Unary `+`/`-` on string operand → ToNumber.
    {
      code: '<div tabIndex={+"5"} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div tabIndex={-"-5"} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex={+"0x10"} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    // Numeric separator (ES2021).
    {
      code: '<div tabIndex={1_000} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    // Scientific notation that lands on integer.
    {
      code: '<div tabIndex={1e2} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    // Negative zero — `-0 >= 0` is true in JS.
    {
      code: '<div tabIndex={-0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    // String forms.
    {
      code: '<div tabIndex="5" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div tabIndex="100" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex=" 0 " />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={" 0 "} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div tabIndex="+0" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex="+1" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex="-0" />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    // Numeric expressions (step-2 fallback).
    {
      code: '<div tabIndex={1+1} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={2-1} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={2*0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={5%3} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    // String concatenation → ToNumber → 0 → reports.
    {
      code: '<div tabIndex={"0" + "0"} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    // Unary plus / minus on numeric literals.
    {
      code: '<div tabIndex={+0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={+5} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div tabIndex={-(-1)} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    // NoSubstitutionTemplateLiteral.
    {
      code: '<div tabIndex={`0`} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={`5`} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    // Parenthesized number literal.
    {
      code: '<div tabIndex={(0)} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div tabIndex={((0))} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    // TS wrappers around numeric / string literals.
    {
      code: '<div tabIndex={0 as number} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div tabIndex={(0) as number} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div tabIndex={0 as any} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    // Note: `<div tabIndex={(0)!} />` is VALID (stringified to "0!" → NaN
    // → undefined → skip). Lock-in in the valid section.
    // Conditional / Logical / Nullish resolving to non-negative.
    {
      code: '<div tabIndex={cond ? 0 : -1} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div tabIndex={true ? 0 : 1} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<div tabIndex={true ? "0" : "-1"} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<div tabIndex={true && 0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div tabIndex={false || 1} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div tabIndex={null ?? 0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    // Boolean-ternary (very rare; step-2 boolean coercion → 0/1, both ≥0).
    {
      code: '<div tabIndex={true ? true : false} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ============================================================
    // tags / roles option does NOT exempt
    // ============================================================
    {
      code: '<div tabIndex={0} />',
      options: [{ tags: ['span'] }],
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div role="article" tabIndex={0} />',
      options: rolesExemptTabpanelOptions,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div role="article" tabIndex={0} />',
      options: [{ roles: [] }],
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },

    // ============================================================
    // Multiple tabIndex props
    // ============================================================
    {
      code: '<div tabIndex={0} tabIndex={5} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ============================================================
    // polymorphicPropName remap to non-interactive
    // ============================================================
    {
      code: '<Box as="div" tabIndex={0} />',
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<Box as="article" tabIndex={0} />',
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<Box as="article" tabIndex={0} />',
      settings: polymorphicAllowListSettings,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // Components map → non-interactive
    // ============================================================
    {
      code: '<MyArticle tabIndex={0} />',
      settings: {
        'jsx-a11y': {
          components: { MyArticle: 'article' },
        },
      },
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<MyDiv tabIndex={0} />',
      settings: {
        'jsx-a11y': {
          components: { MyDiv: 'div' },
        },
      },
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ============================================================
    // Real-world component patterns
    // ============================================================
    {
      code: 'function Outer() { return <div tabIndex={0}>focusable but not interactive</div>; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'const items = arr.map(item => <li key={item.id} tabIndex={0}>{item.name}</li>);',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: 'function Foo({cond}) { return cond ? <article tabIndex={0} /> : null; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: 'const Pane = React.forwardRef((props, ref) => <div ref={ref} tabIndex={0} {...props} />);',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 62,
          endLine: 1,
          endColumn: 74,
        },
      ],
    },
    {
      code: 'function A() { return <div tabIndex={0} />; }\nfunction B() { return <article tabIndex={0} />; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 40,
        },
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 2,
          column: 32,
          endLine: 2,
          endColumn: 44,
        },
      ],
    },
    {
      code: 'const Enhanced = withTracking(({ value }) => <div data-value={value} tabIndex={0} />);',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 70,
          endLine: 1,
          endColumn: 82,
        },
      ],
    },
    {
      code: 'const Pane = React.memo(({ id }) => <section id={id} tabIndex={0} />);',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 54,
          endLine: 1,
          endColumn: 66,
        },
      ],
    },

    // ============================================================
    // Real-world a11y misuse patterns
    // ============================================================
    {
      code: 'function FakeButton({ onClick, children }) { return <div onClick={onClick} tabIndex={0}>{children}</div>; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 76,
          endLine: 1,
          endColumn: 88,
        },
      ],
    },
    {
      code: 'function Dropdown({ items }) { return <ul tabIndex={0}>{items.map(x => <li key={x}>{x}</li>)}</ul>; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'function ClickCard() { return <article tabIndex={0} onClick={handler}>Title</article>; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 40,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },

    // ============================================================
    // Generator / async / IIFE / class
    // ============================================================
    {
      code: 'function* render() { yield <div tabIndex={0} />; yield <article tabIndex={0} />; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 45,
        },
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 65,
          endLine: 1,
          endColumn: 77,
        },
      ],
    },
    {
      code: 'async function render() { return <div tabIndex={0} />; }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 39,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: 'const x = (() => <article tabIndex={0} />)();',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'class Form extends React.Component { render() { return <div tabIndex={0}>ready</div>; } }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 61,
          endLine: 1,
          endColumn: 73,
        },
      ],
    },

    // ============================================================
    // Fragment + conditional rendering
    // ============================================================
    {
      code: 'const x = <>{cond && <div tabIndex={0} />}</>',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'const x = <React.Fragment><div tabIndex={0} /></React.Fragment>',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },

    // ============================================================
    // Comments around / inside the prop don't suppress
    // ============================================================
    {
      code: '<div /* a */ tabIndex={0} /* b */ />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div tabIndex={/* truthy */ 0} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // Spread literal — role / tabIndex flowing through
    // ============================================================
    {
      code: '<div {...{role: "article", tabIndex: 0}} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<div {...{tabIndex: 0}} />',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ============================================================
    // Multi-element file with mixed interactive / non-interactive
    // ============================================================
    {
      code: 'function App() { return (<><div tabIndex={0}>A</div><button tabIndex={0}>B</button><article tabIndex={0}>C</article><input tabIndex={0} /><section tabIndex={0}>D</section></>); }',
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 45,
        },
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 93,
          endLine: 1,
          endColumn: 105,
        },
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 148,
          endLine: 1,
          endColumn: 160,
        },
      ],
    },
  ],
});

// Additional upstream v6.10.2 inputs. Matrix rows freeze the expected
// classification and configuration; ranges follow the authored source templates.
// No rule implementation or lint output is consulted when generating cases.
const upstreamOptions0 = [
  { tags: [], roles: ['tabpanel'], allowExpressionValues: true },
];
const upstreamOptions1 = [
  { allowExpressionValues: true, tags: [], roles: ['tabpanel'] },
];
const upstreamSettings0 = {
  'jsx-a11y': { components: { Article: 'article', MyButton: 'button' } },
};

new RuleTester(config).run('no-noninteractive-tabindex', {
  valid: [
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 0',
      code: '<MyButton tabIndex={0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 1',
      code: '<button />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 2',
      code: '<button tabIndex="0" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 3',
      code: '<button tabIndex={0} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 4',
      code: '<div />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 5',
      code: '<div tabIndex="-1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 6',
      code: '<div role="button" tabIndex="0" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 7',
      code: '<div role="article" tabIndex="-1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 8',
      code: '<article tabIndex="-1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 9',
      code: '<Article tabIndex="-1" />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 10',
      code: '<MyButton tabIndex={0} />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 13',
      code: '<div role={BUTTON} onClick={() => {}} tabIndex="0" />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 14',
      code: '<div role={isButton ? "button" : "link"} onClick={() => {}} tabIndex="0" />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 15',
      code: '<div role={isButton ? "button" : LINK} onClick={() => {}} tabIndex="0" />;',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended valid 16',
      code: '<div role={isButton ? BUTTON : LINK} onClick={() => {}} tabIndex="0"/>;',
      options: upstreamOptions1,
    },
  ],
  invalid: [
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended invalid 0',
      code: '<div tabIndex="0" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended invalid 1',
      code: '<div role="article" tabIndex="0" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended invalid 2',
      code: '<article tabIndex="0" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended invalid 3',
      code: '<article tabIndex={0} />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 no-noninteractive-tabindex:recommended invalid 4',
      code: '<Article tabIndex={0} />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
      errors: [
        {
          messageId: 'noNoninteractiveTabindex',
          message:
            '`tabIndex` should only be declared on interactive elements.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
  ],
});
