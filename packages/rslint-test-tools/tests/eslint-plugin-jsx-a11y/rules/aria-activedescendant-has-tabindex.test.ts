import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const componentsCustomDiv = {
  'jsx-a11y': {
    components: {
      CustomComponent: 'div',
    },
  },
};

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

new RuleTester(config).run('aria-activedescendant-has-tabindex', {
  valid: [
    // ============================================================
    // Upstream-suite valid cases
    // ============================================================
    { code: '<CustomComponent />;' },
    { code: '<CustomComponent aria-activedescendant={someID} />;' },
    {
      code: '<CustomComponent aria-activedescendant={someID} tabIndex={0} />;',
    },
    {
      code: '<CustomComponent aria-activedescendant={someID} tabIndex={-1} />;',
    },
    {
      code: '<CustomComponent aria-activedescendant={someID} tabIndex={0} />;',
      settings: componentsCustomDiv,
    },
    { code: '<div />;' },
    { code: '<input />;' },
    { code: '<div tabIndex={0} />;' },
    { code: '<div aria-activedescendant={someID} tabIndex={0} />;' },
    { code: '<div aria-activedescendant={someID} tabIndex="0" />;' },
    { code: '<div aria-activedescendant={someID} tabIndex={1} />;' },
    { code: '<input aria-activedescendant={someID} />;' },
    { code: '<input aria-activedescendant={someID} tabIndex={1} />;' },
    { code: '<input aria-activedescendant={someID} tabIndex={0} />;' },
    { code: '<input aria-activedescendant={someID} tabIndex={-1} />;' },
    { code: '<div aria-activedescendant={someID} tabIndex={-1} />;' },
    { code: '<div aria-activedescendant={someID} tabIndex="-1" />;' },
    { code: '<input aria-activedescendant={someID} tabIndex={-1} />;' },

    // ============================================================
    // Case-insensitive prop match — jsx-ast-utils default ignoreCase: true
    // ============================================================
    { code: '<div Aria-ActiveDescendant={x} tabIndex={0} />;' },
    { code: '<div ARIA-ACTIVEDESCENDANT={x} tabIndex={0} />;' },

    // ============================================================
    // Other inherently interactive elements — gate-3 skip
    // ============================================================
    { code: '<a href="#" aria-activedescendant={x} />;' },
    { code: '<button aria-activedescendant={x} />;' },
    { code: '<select size={1} aria-activedescendant={x} />;' },
    { code: '<textarea aria-activedescendant={x} />;' },
    { code: '<option aria-activedescendant={x} />;' },

    // ============================================================
    // Non-DOM tag forms — gate-2 skip
    // ============================================================
    { code: '<svg:path aria-activedescendant={x} />;' },
    { code: '<Foo.Bar aria-activedescendant={x} />;' },
    { code: '<Foo.Bar.Baz aria-activedescendant={x} />;' },

    // ============================================================
    // Polymorphic settings
    // ============================================================
    {
      code: '<Box as="span" aria-activedescendant={x} tabIndex={0} />;',
      settings: polymorphicSettings,
    },

    // ============================================================
    // tabIndex string variants — coerce via getTabIndex's StringToNumber
    // ============================================================
    { code: '<div aria-activedescendant={x} tabIndex="1" />;' },
    { code: '<div aria-activedescendant={x} tabIndex="100" />;' },
    { code: '<div aria-activedescendant={x} tabIndex="0x10" />;' },
    { code: '<div aria-activedescendant={x} tabIndex="0o7" />;' },

    // ============================================================
    // Spread without literal aria-activedescendant — opaque, gate-1 skip
    // ============================================================
    { code: '<div {...props} />;' },

    // ============================================================
    // Nested JSX where neither element trips the rule
    // ============================================================
    { code: '<span><div aria-activedescendant={x} tabIndex={0} /></span>;' },

    // ============================================================
    // Literal-spread tabIndex tracking (jsx-ast-utils' getProp walks
    // ObjectExpression spreads when the key is an Identifier)
    // ============================================================
    { code: '<div aria-activedescendant={x} {...{tabIndex: 0}} />;' },
    { code: '<div aria-activedescendant={x} {...{tabIndex: -1}} />;' },

    // ============================================================
    // TS-wrapper unwrapping in tabIndex value (paren / as).
    // Note: TSNonNullExpression `0!` is INVALID per upstream — it
    // stringifies to "0!" → NaN → undefined → REPORT (see invalid section).
    // ============================================================
    { code: '<div aria-activedescendant={x} tabIndex={(0)} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={(0 as number)} />;' },

    // ============================================================
    // Opaque expression types resolve to upstream `null`, which the
    // downstream `null >= -1` (ToNumber → 0) treats as a valid tabIndex.
    // Aligned with ESLint via GetTabIndexEx's nullLike classification.
    // ============================================================
    {
      code: '<div aria-activedescendant={x} tabIndex={0 satisfies number} />;',
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={(-2) satisfies number} />;',
    },
    {
      code: 'async function f() { return <div aria-activedescendant={x} tabIndex={await p} />; }',
    },
    {
      code: 'function* g() { yield <div aria-activedescendant={x} tabIndex={yield 0} />; }',
    },

    // ============================================================
    // NoSubstitutionTemplateLiteral tabIndex
    // ============================================================
    { code: '<div aria-activedescendant={x} tabIndex={`-1`} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={`0`} />;' },

    // ============================================================
    // Comments around the attribute don't suppress detection
    // ============================================================
    {
      code: '<div /* before */ aria-activedescendant={x} /* mid */ tabIndex={0} /* after */ />;',
    },

    // ============================================================
    // Expression-form coverage for tabIndex value (valid arms)
    // ============================================================
    { code: '<div aria-activedescendant={x} tabIndex={cond ? 0 : -1} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={true ? -1 : -5} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={false ? -5 : 0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={null ?? 0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={1 || 5} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={true && 0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={1 - 1} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={1 - 2} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={2 * 0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={10 / 10} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={[0]} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={[-1]} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={[null]} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={[]} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={0x0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={0o0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={0b0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={1e0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={1_000} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={Infinity} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={-0} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={-(-2)} />;' },
    { code: '<div aria-activedescendant={x} tabIndex={(((-1)))} />;' },

    // ============================================================
    // aria-activedescendant value-form coverage (gate-1 only checks
    // presence; tabIndex={0} keeps these valid)
    // ============================================================
    { code: '<div aria-activedescendant="staticID" tabIndex={0} />;' },
    { code: '<div aria-activedescendant={`staticID`} tabIndex={0} />;' },
    { code: '<div aria-activedescendant={state.focusedId} tabIndex={0} />;' },
    { code: '<div aria-activedescendant={cond ? a : b} tabIndex={0} />;' },
    { code: '<div aria-activedescendant={getId()} tabIndex={0} />;' },

    // ============================================================
    // Real-world a11y patterns
    // ============================================================
    { code: '<input role="combobox" aria-activedescendant={x} />;' },
    {
      code: '<ul role="listbox" tabIndex={0} aria-activedescendant={focusedId}><li>x</li></ul>;',
    },
    { code: '<a href={url} aria-activedescendant={x} />;' },
    {
      code: 'function L() { return items.map(i => <li tabIndex={-1} aria-activedescendant={i.id} key={i.id}>x</li>); }',
    },

    // ============================================================
    // Tag-name case sensitivity — uppercase is a component reference
    // ============================================================
    { code: '<DIV aria-activedescendant={x} />;' },
  ],
  invalid: [
    // ============================================================
    // TSNonNullExpression on tabIndex — jsx-ast-utils' TSNonNullExpression
    // extractor stringifies (`0!` → "0!"). Number("0!") = NaN → step-1
    // undefined → aria gate-4 `undefined >= -1` false → REPORT.
    // ============================================================
    {
      code: '<div aria-activedescendant={x} tabIndex={0!} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={(0)!} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },

    // ============================================================
    // Upstream-suite invalid cases
    // ============================================================
    {
      code: '<div aria-activedescendant={someID} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<CustomComponent aria-activedescendant={someID} />;',
      settings: componentsCustomDiv,
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },

    // ============================================================
    // gate-1: boolean attribute form / explicit-undefined value still
    // count as "defined" via jsx-ast-utils' getProp
    // ============================================================
    {
      code: '<div aria-activedescendant />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={undefined} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },

    // ============================================================
    // Case-insensitive prop match — INVALID arm
    // ============================================================
    {
      code: '<div Aria-ActiveDescendant={x} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<div ARIA-ACTIVEDESCENDANT={x} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },

    // ============================================================
    // Paired element form — listener fires once on JsxOpeningElement
    // ============================================================
    {
      code: '<div aria-activedescendant={x}></div>;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x}>text</div>;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // ============================================================
    // gate-3 boundary: interactive + tabIndex defined → don't skip
    // ============================================================
    {
      code: '<input aria-activedescendant={x} tabIndex={-2} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<button aria-activedescendant={x} tabIndex={-5} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },

    // ============================================================
    // a without href is NOT interactive
    // ============================================================
    {
      code: '<a aria-activedescendant={x} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },

    // ============================================================
    // gate-4 boundary: tabIndex < -1
    // ============================================================
    {
      code: '<div aria-activedescendant={x} tabIndex={-2} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={-100} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex="-2" />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },

    // ============================================================
    // gate-4: NaN-coercing tabIndex string → undefined → fall through
    // ============================================================
    {
      code: '<div aria-activedescendant={x} tabIndex="abc" />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={NaN} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={undefined} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },

    // ============================================================
    // polymorphicPropName resolution
    // ============================================================
    {
      code: '<Box as="div" aria-activedescendant={x} />;',
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },

    // ============================================================
    // Nested JSX — listener fires per element independently
    // ============================================================
    {
      code: '<span><div aria-activedescendant={x} /></span>;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x}><span aria-activedescendant={x} /></div>;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 66,
        },
      ],
    },

    // ============================================================
    // TS-wrapper tabIndex value resolving to invalid range
    // ============================================================
    {
      code: '<div aria-activedescendant={x} tabIndex={(-2 as number)} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={(-2)!} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },

    // ============================================================
    // Multiple sibling invalid elements
    // ============================================================
    {
      code: '<><div aria-activedescendant={x} /><span aria-activedescendant={y} /><p aria-activedescendant={z} /></>;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 36,
        },
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 36,
          endLine: 1,
          endColumn: 70,
        },
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 70,
          endLine: 1,
          endColumn: 101,
        },
      ],
    },

    // ============================================================
    // Real-world component patterns
    // ============================================================
    {
      code: 'const Item = () => <li aria-activedescendant={x} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: 'function R() { return <div aria-activedescendant={x} />; }',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },
    {
      code: 'class C extends React.Component { render() { return <div aria-activedescendant={x} />; } }',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 86,
        },
      ],
    },

    // ============================================================
    // Expression-form coverage for tabIndex value (invalid arms)
    // ============================================================
    {
      code: '<div aria-activedescendant={x} tabIndex={cond ? -2 : -3} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={false ? 0 : -2} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={x || 0} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={x ?? 0} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={1 - 5} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={[-2]} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={[5,6]} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={true} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={false} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={null} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={-Infinity} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={`${0}`} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<div aria-activedescendant={x} tabIndex={`${cond}`} />;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },

    // ============================================================
    // Real-world a11y misuse patterns
    // ============================================================
    {
      code: '<ul aria-activedescendant={x}><li>x</li></ul>;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<section aria-activedescendant={x} tabIndex={-2}>content</section>;',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'function L() { return items.map(i => <li aria-activedescendant={i.id}>x</li>); }',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 71,
        },
      ],
    },

    // ============================================================
    // Mixed nested tree — outer + inner each invalid
    // ============================================================
    {
      code: 'function App() { return (<ul aria-activedescendant={a}><li aria-activedescendant={b}>x</li></ul>); }',
      errors: [
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 56,
        },
        {
          messageId: 'tabIndexRequired',
          message:
            'An element that manages focus with `aria-activedescendant` must have a tabindex',
          line: 1,
          column: 56,
          endLine: 1,
          endColumn: 86,
        },
      ],
    },
  ],
});
