import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Intentional difference from eslint-plugin-jsx-a11y v6.10.2: Rslint allows
// values statically resolved to false or "false"; upstream reports any autoFocus
// prop. The message uses enabled rather than used. See no_autofocus.md.

const ignoreNonDOMSchema = [{ ignoreNonDOM: true }];

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Button: 'button',
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

const componentsToCustomSettings = {
  'jsx-a11y': {
    components: { Foo: 'Bar' },
  },
};

const emptyJsxA11ySettings = {
  'jsx-a11y': {},
};

new RuleTester(config).run('no-autofocus', {
  valid: [
    // ---- Upstream valid ----
    { code: '<div />;' },
    // Lowercase `autofocus` is the HTML DOM attribute, not React's prop.
    { code: '<div autofocus />;' },
    { code: '<input autofocus="true" />;' },
    { code: '<Foo bar />' },
    // These two upstream invalid inputs are intentionally valid in Rslint.
    { code: '<div autoFocus={false} />' },
    { code: '<div autoFocus="false" />' },
    // ignoreNonDOM: true skips custom components.
    { code: '<Foo autoFocus />', options: ignoreNonDOMSchema },
    { code: '<div><div autofocus /></div>', options: ignoreNonDOMSchema },
    { code: '<Button />', settings: componentsSettings },
    {
      code: '<Button />',
      options: ignoreNonDOMSchema,
      settings: componentsSettings,
    },

    // ---- rslint extras: case-sensitivity ----
    { code: '<div AUTOFOCUS />' },
    { code: '<div AutoFocus />' },
    { code: '<div autoFocuS />' },
    { code: '<div xml:autoFocus />' },

    // ---- "false" literal forms — all coerce to boolean false ----
    { code: '<div autoFocus="False" />' },
    { code: '<div autoFocus="FALSE" />' },
    { code: '<div autoFocus={"false"} />' },
    { code: '<div autoFocus={"False"} />' },
    { code: '<div autoFocus={("false")} />' },
    { code: '<div autoFocus={"false" as string} />' },
    { code: '<div autoFocus={false as boolean} />' },
    { code: '<div autoFocus={`false`} />' },

    // ---- Logical / conditional resolving to false ----
    { code: '<div autoFocus={true && false} />' },
    { code: '<div autoFocus={true ? false : true} />' },
    { code: '<div autoFocus={false || false} />' },

    // ---- ignoreNonDOM: true skips non-DOM tags ----
    { code: '<UX.Layout autoFocus />', options: ignoreNonDOMSchema },
    { code: '<svg:circle autoFocus />', options: ignoreNonDOMSchema },

    // ---- ignoreNonDOM × components map: false short-circuits ----
    {
      code: '<Button autoFocus={false} />',
      options: ignoreNonDOMSchema,
      settings: componentsSettings,
    },

    // ---- ignoreNonDOM × polymorphicPropName ----
    {
      code: '<Box as="div" autoFocus={false} />',
      options: ignoreNonDOMSchema,
      settings: polymorphicSettings,
    },
    {
      code: '<Box as="ComponentName" autoFocus />',
      options: ignoreNonDOMSchema,
      settings: polymorphicSettings,
    },

    // ---- Default options: bare {} and array-wrapped [{}] ----
    { code: '<Foo autoFocus={false} />', options: [{}] },

    // ---- Spread attributes are NOT JsxAttribute → listener never fires ----
    { code: '<div {...{autoFocus: true}} />' },
    { code: '<div {...props} />' },
    { code: '<div {...{autoFocus: true}} {...props} />' },

    // ---- Real-world valid: explicit false suppresses across patterns ----
    {
      code: 'function Modal({ open }) { return <dialog open={open}><input autoFocus={false} placeholder="search" /></dialog>; }',
    },
    {
      code: 'function Search() { const ref = useRef(null); useEffect(() => ref.current?.focus(), []); return <input ref={ref} />; }',
    },
    {
      code: 'const Enhanced = withTracking(({ value }) => <input value={value} />);',
    },
    {
      code: 'const FocusInput = React.forwardRef((props, ref) => <input ref={ref} autoFocus={false} {...props} />);',
    },
    {
      code: 'const Item = React.memo(({ id }) => <li id={id} autoFocus={false}>{id}</li>);',
    },

    // ---- TypeScript generic JSX components ----
    { code: '<List<string> autoFocus={false} />' },
    { code: '<List<{a: number}> autoFocus />', options: ignoreNonDOMSchema },

    // ---- Long-chain member-expression tags ----
    { code: '<Foo.Bar.Baz autoFocus={false} />' },
    { code: '<Foo.Bar.Baz autoFocus />', options: ignoreNonDOMSchema },

    // ---- Hyphenated DOM tags (web components) ----
    { code: '<my-element autoFocus />', options: ignoreNonDOMSchema },

    // ---- components map: custom → custom (still custom) ----
    {
      code: '<Foo autoFocus />',
      options: ignoreNonDOMSchema,
      settings: componentsToCustomSettings,
    },

    // ---- empty `jsx-a11y` settings ----
    {
      code: '<Foo autoFocus />',
      options: ignoreNonDOMSchema,
      settings: emptyJsxA11ySettings,
    },

    // ---- polymorphicAllowList restricts the `as` swap ----
    {
      code: '<Box as="input" autoFocus={false} />',
      options: ignoreNonDOMSchema,
      settings: polymorphicAllowListSettings,
    },
    {
      code: '<Other as="input" autoFocus />',
      options: ignoreNonDOMSchema,
      settings: polymorphicAllowListSettings,
    },

    // ---- Comments around the prop don't break extraction ----
    { code: '<div /* before */ autoFocus={false} /* after */ />' },
    { code: '<div autoFocus={/* false */ false} />' },
  ],
  invalid: [
    // ---- Upstream invalid ----
    {
      code: '<div autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<div autoFocus={true} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div autoFocus={undefined} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div autoFocus="true" />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<input autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<Foo autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<Button autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      settings: componentsSettings,
    },
    {
      code: '<Button autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
      options: ignoreNonDOMSchema,
      settings: componentsSettings,
    },

    // ---- rslint extras: paired element + nested listener boundary ----
    {
      code: '<div autoFocus>child</div>',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<a autoFocus><span autoFocus /></a>',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 13,
        },
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },

    // ---- Element kind survey ----
    {
      code: '<a autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 4,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<Component autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<UX.Layout autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<svg:circle autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ---- Truthy literal coercions ----
    {
      code: '<div autoFocus="True" />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div autoFocus="TRUE" />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div autoFocus={"true"} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // ---- Non-coerced falsy values still report ----
    {
      code: '<div autoFocus={null} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div autoFocus={0} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div autoFocus={""} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div autoFocus="" />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div autoFocus={void 0} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // ---- Numeric / BigInt / Identifier expressions ----
    {
      code: '<div autoFocus={1} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div autoFocus={1n} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div autoFocus={someVar} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ---- Conditional / logical resolving to truthy ----
    {
      code: '<div autoFocus={true && true} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<div autoFocus={false || true} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div autoFocus={"x" && true} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },

    // ---- Call / Member / Template ----
    {
      code: '<div autoFocus={fn()} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div autoFocus={obj.x} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div autoFocus={obj?.x} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div autoFocus={tag`x`} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div autoFocus={`${x}`} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div autoFocus={`true`} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // ---- TS wrappers around truthy ----
    {
      code: '<div autoFocus={true as boolean} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div autoFocus={(true)} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div autoFocus={(true)!} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    // TSNonNullExpression on string literal: upstream stringifies to "false!"
    // (NOT coerced via case-insensitive "true"/"false" Literal rule).
    // "false!" !== false → autoFocus enabled → reports.
    {
      code: '<div autoFocus={"false"!} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },

    // ---- `satisfies` is opaque — even `false satisfies boolean` reports ----
    {
      code: '<div autoFocus={false satisfies boolean} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: '<div autoFocus={"false" satisfies string} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },

    // ---- TemplateExpression with `${false}` substitution → truthy string ----
    {
      code: '<div autoFocus={`${false}`} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<div autoFocus={`prefix${false}suffix`} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },

    // ---- Deep nesting × ignoreNonDOM ----
    {
      code: '<Outer><Mid><input autoFocus /></Mid></Outer>',
      options: ignoreNonDOMSchema,
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<Custom autoFocus><input autoFocus /></Custom>',
      options: ignoreNonDOMSchema,
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<div autoFocus><span autoFocus /></div>',
      options: ignoreNonDOMSchema,
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ---- ignoreNonDOM: true × DOM element ----
    {
      code: '<input autoFocus />',
      options: ignoreNonDOMSchema,
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<button autoFocus />',
      options: ignoreNonDOMSchema,
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ---- ignoreNonDOM: true × polymorphicPropName resolving to DOM ----
    {
      code: '<Box as="input" autoFocus />',
      options: ignoreNonDOMSchema,
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<Box as="div" autoFocus />',
      options: ignoreNonDOMSchema,
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },

    // ---- Multiple autoFocus attributes on one element ----
    {
      code: '<div autoFocus autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ---- Real-world component patterns ----
    {
      code: 'function LoginField() { return <input type="text" autoFocus />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 60,
        },
      ],
    },
    {
      code: 'const x = items.map(item => <li autoFocus key={item.id} />)',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: 'const x = <>{cond && <input autoFocus />}</>',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },

    // ---- Form patterns: search / login / signup ----
    {
      code: 'function SearchBar() { return <input type="search" autoFocus placeholder="Search…" />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 52,
          endLine: 1,
          endColumn: 61,
        },
      ],
    },
    {
      code: 'function LoginForm() { return <form><input autoFocus name="user" /><input type="password" name="pass" /></form>; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: 'function Signup() { return (<form><label>Email<input type="email" autoFocus /></label><label>Pass<input type="password" autoFocus /></label></form>); }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 67,
          endLine: 1,
          endColumn: 76,
        },
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 121,
          endLine: 1,
          endColumn: 130,
        },
      ],
    },

    // ---- Modal / Dialog with offending autoFocus ----
    {
      code: 'function Modal() { return <dialog open><input autoFocus name="title" /><button>OK</button></dialog>; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 47,
          endLine: 1,
          endColumn: 56,
        },
      ],
    },

    // ---- HOC / forwardRef / memo wrappers carrying autoFocus ----
    {
      code: 'const Enhanced = withTracking(({ value }) => <input value={value} autoFocus />);',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 67,
          endLine: 1,
          endColumn: 76,
        },
      ],
    },
    {
      code: 'const FocusInput = React.forwardRef((props, ref) => <input ref={ref} autoFocus {...props} />);',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 70,
          endLine: 1,
          endColumn: 79,
        },
      ],
    },
    {
      code: 'const Item = React.memo(({ id }) => <li id={id} autoFocus>{id}</li>);',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 49,
          endLine: 1,
          endColumn: 58,
        },
      ],
    },

    // ---- TypeScript generic JSX ----
    {
      code: '<List<string> autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<Cell<{a: number}> autoFocus={true} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- Long-chain member-expression tags (without ignoreNonDOM) ----
    {
      code: '<Foo.Bar.Baz autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<this.Foo autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ---- Hyphenated DOM tags without ignoreNonDOM ----
    {
      code: '<my-element autoFocus />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ---- Comments around / inside the prop don't suppress ----
    {
      code: '<div /* a */ autoFocus /* b */ />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div autoFocus={/* truthy */ true} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },

    // ---- Complex value expressions ----
    {
      code: '<div autoFocus={[1, 2, 3]} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<div autoFocus={{nested: true}} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div autoFocus={new Boolean(false)} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- Optional chain / nullish ----
    {
      code: '<div autoFocus={cfg?.autoFocus} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div autoFocus={cfg ?? true} />',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },

    // ---- ignoreNonDOM × polymorphicAllowList ----
    {
      code: '<Box as="input" autoFocus />',
      options: ignoreNonDOMSchema,
      settings: polymorphicAllowListSettings,
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },

    // ---- Conditional rendering forms ----
    {
      code: 'function Foo({cond}) { return cond && <input autoFocus />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 55,
        },
      ],
    },
    {
      code: 'function Foo({cond}) { return cond ? <input autoFocus /> : <div />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'function Foo({a, b}) { return a ? <input autoFocus /> : <textarea autoFocus />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 42,
          endLine: 1,
          endColumn: 51,
        },
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 67,
          endLine: 1,
          endColumn: 76,
        },
      ],
    },

    // ---- Switch / list / multi-component / generator / async / IIFE ----
    {
      code: "function Foo({type}) { switch(type) { case 'input': return <input autoFocus />; default: return null; } }",
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 67,
          endLine: 1,
          endColumn: 76,
        },
      ],
    },
    {
      code: 'const items = arr.map((x, i) => <li key={i} autoFocus>{x}</li>);',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'function A() { return <input autoFocus />; }\nfunction B() { return <input autoFocus />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 39,
        },
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 2,
          column: 30,
          endLine: 2,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'class Form extends React.Component { state = {ready: true}; render() { return this.state.ready ? <input autoFocus /> : <div />; } }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 105,
          endLine: 1,
          endColumn: 114,
        },
      ],
    },
    {
      code: 'function* render() { yield <input autoFocus />; yield <textarea autoFocus />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 44,
        },
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 65,
          endLine: 1,
          endColumn: 74,
        },
      ],
    },
    {
      code: 'async function render() { return <input autoFocus />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'const x = (() => <input autoFocus />)();',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: 'function Form({ initial: { autoFocus = true } }) { return <input autoFocus={autoFocus} />; }',
      errors: [
        {
          messageId: 'noAutoFocus',
          message:
            'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.',
          line: 1,
          column: 66,
          endLine: 1,
          endColumn: 87,
        },
      ],
    },
  ],
});
