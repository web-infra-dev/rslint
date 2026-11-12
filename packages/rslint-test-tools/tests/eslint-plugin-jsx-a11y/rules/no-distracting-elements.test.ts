import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const blinkComponentSettings = {
  'jsx-a11y': {
    components: {
      Blink: 'blink',
    },
  },
};

new RuleTester(config).run('no-distracting-elements', {
  valid: [
    // ---- Upstream valid ----
    { code: '<div />;' },
    { code: '<Marquee />' },
    { code: '<div marquee />' },
    { code: '<Blink />' },
    { code: '<div blink />' },

    // ---- Case sensitivity: only lowercase intrinsic tags match ----
    { code: '<MARQUEE />' },
    { code: '<MarQuee />' },
    { code: '<BLINK />' },

    // ---- PropertyAccess / namespaced tags do NOT match ----
    { code: '<UI.Marquee />' },
    { code: '<svg:marquee />' },

    // ---- Element kind survey: rule is a no-op for every non-listed tag ----
    { code: '<a />' },
    { code: '<input />' },
    { code: '<Component />' },

    // ---- Empty `elements` array disables every check ----
    { code: '<marquee />', options: [{ elements: [] }] },
    { code: '<blink />', options: [{ elements: [] }] },

    // ---- Custom `elements` list — types not in the list pass ----
    { code: '<marquee />', options: [{ elements: ['blink'] }] },
    { code: '<blink />', options: [{ elements: ['marquee'] }] },

    // ---- Components map without a matching entry ----
    {
      code: '<Blink />',
      settings: { 'jsx-a11y': { components: { OtherTag: 'marquee' } } },
    },

    // ---- Polymorphic with allow-list NOT containing the tag ----
    {
      code: '<Foo as="marquee" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Bar'],
        },
      },
    },

    // ---- Multi-segment PropertyAccess + this.Foo + lowercase final ----
    { code: '<A.B.C />' },
    { code: '<this.Foo />' },
    { code: '<UI.marquee />' }, // type "UI.marquee" ≠ "marquee"

    // ---- Components map: map-to-non-distracting / non-string value ----
    {
      code: '<Foo />',
      settings: { 'jsx-a11y': { components: { Foo: 'div' } } },
    },
    // Reverse aliasing: <marquee/> → "div" → no report.
    {
      code: '<marquee />',
      settings: { 'jsx-a11y': { components: { marquee: 'div' } } },
    },
    {
      code: '<Foo />',
      settings: { 'jsx-a11y': { components: { Foo: 123 } } },
    },

    // ---- Polymorphic prop edge cases ----
    {
      code: '<Foo />', // no `as` prop
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<Foo as={x} />', // dynamic Identifier
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<Foo as="" />', // empty string falsy
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<Foo as={cond ? "marquee" : "blink"} />', // Conditional opaque
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<Foo as="marquee" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: [], // empty list → never replace
        },
      },
    },

    // ---- Polymorphic reverse exemption ----
    {
      code: '<marquee as="div" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    {
      code: '<blink as="span" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },
    // Polymorphic alone replaces with non-distracting; no components chain.
    {
      code: '<Foo as="Bar" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
    },

    // ---- Defensive type handling for settings ----
    // `settings['jsx-a11y']` is not a map.
    {
      code: '<Foo />',
      settings: { 'jsx-a11y': 'invalid' as unknown as object },
    },
    // `settings['jsx-a11y']` is null.
    { code: '<Foo />', settings: { 'jsx-a11y': null as unknown as object } },
    // `components` is not a map.
    {
      code: '<Foo />',
      settings: { 'jsx-a11y': { components: 'invalid' as unknown as object } },
    },
    // `polymorphicPropName` is a number — upstream throws TypeError on
    // this; rslint silently skips the polymorphic block (Go-natural
    // robustness divergence). Locking in our behavior.
    {
      code: '<Foo as="marquee" />',
      settings: {
        'jsx-a11y': { polymorphicPropName: 123 as unknown as string },
      },
    },
    // allowList contains only non-strings.
    {
      code: '<Foo as="marquee" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: [123 as unknown as string],
        },
      },
    },
  ],
  invalid: [
    // ---- Upstream invalid ----
    {
      code: '<marquee />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '<marquee {...props} />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<marquee lang={undefined} />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<blink />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<blink {...props} />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<blink foo={undefined} />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<Blink />',
      settings: blinkComponentSettings,
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },

    // ---- Paired form: report on the OPENING element only ----
    {
      code: '<marquee>scrolling</marquee>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<marquee></marquee>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },

    // ---- Same-kind nesting: outer + inner each report ----
    {
      code: '<marquee><marquee /></marquee>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },

    // ---- Boolean / spread / various attribute forms ----
    {
      code: '<marquee draggable />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<marquee className="x" id={n} />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },

    // ---- Custom `elements` list ----
    {
      code: '<blink />',
      options: [{ elements: ['blink'] }],
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },

    // ---- Components map with a matching entry ----
    {
      code: '<CustomMarquee />',
      settings: {
        'jsx-a11y': { components: { CustomMarquee: 'marquee' } },
      },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // ---- Polymorphic prop without an allow-list — every truthy `as` value
    //      replaces rawType ----
    {
      code: '<Foo as="marquee" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    // Polymorphic with allow-list that DOES include the rawType.
    {
      code: '<Foo as="blink" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Foo'],
        },
      },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },

    // ---- Listener boundary: each violation reports independently ----
    {
      code: '<div><marquee /></div>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<marquee><blink /></marquee>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },

    // ---- Real-world component patterns ----
    {
      code: 'function Banner() { return <marquee>News</marquee>; }',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'const x = items.map(item => <blink key={item.id}>{item.text}</blink>)',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'const x = cond ? <marquee /> : <div />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'class C { render() { return <div><marquee /><blink /></div>; } }',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 45,
        },
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },

    // ---- Deep nesting (3 levels) ----
    {
      code: '<marquee><marquee><marquee /></marquee></marquee>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 19,
        },
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },

    // ---- JsxFragment wrapper ----
    {
      code: '<><marquee /></>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },

    // ---- JSX inside expression containers (children, attribute, &&) ----
    {
      code: '<div>{<marquee />}</div>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div>{cond && <marquee />}</div>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<div content={<marquee />} />',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<Provider value={data}>{<marquee />}</Provider>',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },

    // ---- Components map: map-to-distracting / map-to-self ----
    {
      code: '<Foo />',
      settings: { 'jsx-a11y': { components: { Foo: 'marquee' } } },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<marquee />',
      settings: { 'jsx-a11y': { components: { marquee: 'marquee' } } },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },

    // ---- Polymorphic + components combination ----
    // Polymorphic replaces FIRST → rawType becomes "marquee"; components
    // has no "marquee" key, so it's a no-op. Reports marquee.
    {
      code: '<Foo as="marquee" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          components: { Foo: 'div' },
        },
      },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },

    // Polymorphic + components CHAIN: Foo → Bar (polymorphic) → marquee
    // (components). Locks in that components looks up the post-polymorphic
    // name.
    {
      code: '<Foo as="Bar" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          components: { Bar: 'marquee' },
        },
      },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },

    // ---- Polymorphic prop value forms ----
    {
      code: '<Foo as={"marquee"} />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<Foo as={`marquee`} />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ---- Real-world patterns (extended) ----
    {
      code: 'const Banner = withTheme(() => <marquee />);',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 43,
        },
      ],
    },
    {
      code: 'const x = items.filter(x => x).map(x => <marquee key={x.id} />)',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 63,
        },
      ],
    },
    {
      code: 'const items = [<marquee key="1" />, <blink key="2" />];',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 35,
        },
        {
          messageId: 'distractingElement',
          message:
            'Do not use <blink> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'wrap(<marquee />);',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: 'const x = { content: <marquee /> };',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'const Marquee = React.forwardRef((props, ref) => <marquee ref={ref} {...props} />);',
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 50,
          endLine: 1,
          endColumn: 82,
        },
      ],
    },

    // ---- Defensive type handling (invalid side) ----
    // Mixed-type allowList — non-string entries dropped, "Foo" honored.
    {
      code: '<Foo as="marquee" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: [123 as unknown as string, 'Foo'],
        },
      },
      errors: [
        {
          messageId: 'distractingElement',
          message:
            'Do not use <marquee> elements as they can create visual accessibility issues and are deprecated.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
  ],
});
