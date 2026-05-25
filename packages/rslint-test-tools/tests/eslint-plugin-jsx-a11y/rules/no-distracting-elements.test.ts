import { RuleTester } from '../rule-tester';

const expectedError = (element: string) => ({
  message: `Do not use <${element}> elements as they can create visual accessibility issues and are deprecated.`,
});

const blinkComponentSettings = {
  'jsx-a11y': {
    components: {
      Blink: 'blink',
    },
  },
};

new RuleTester().run('no-distracting-elements', null as never, {
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

    // ---- options: null elements falls back to default ----
    { code: '<div />', options: [{ elements: null }] },
    // rslint extension: array-of-non-strings silences the rule.
    { code: '<marquee />', options: [{ elements: [123, true] }] },

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
    { code: '<marquee />', errors: [expectedError('marquee')] },
    { code: '<marquee {...props} />', errors: [expectedError('marquee')] },
    {
      code: '<marquee lang={undefined} />',
      errors: [expectedError('marquee')],
    },
    { code: '<blink />', errors: [expectedError('blink')] },
    { code: '<blink {...props} />', errors: [expectedError('blink')] },
    { code: '<blink foo={undefined} />', errors: [expectedError('blink')] },
    {
      code: '<Blink />',
      settings: blinkComponentSettings,
      errors: [expectedError('blink')],
    },

    // ---- Paired form: report on the OPENING element only ----
    {
      code: '<marquee>scrolling</marquee>',
      errors: [expectedError('marquee')],
    },
    { code: '<marquee></marquee>', errors: [expectedError('marquee')] },

    // ---- Same-kind nesting: outer + inner each report ----
    {
      code: '<marquee><marquee /></marquee>',
      errors: [expectedError('marquee'), expectedError('marquee')],
    },

    // ---- Boolean / spread / various attribute forms ----
    { code: '<marquee draggable />', errors: [expectedError('marquee')] },
    {
      code: '<marquee className="x" id={n} />',
      errors: [expectedError('marquee')],
    },

    // ---- Custom `elements` list ----
    {
      code: '<custom />',
      options: [{ elements: ['custom'] }],
      errors: [expectedError('custom')],
    },
    {
      code: '<blink />',
      options: [{ elements: ['blink'] }],
      errors: [expectedError('blink')],
    },

    // ---- Components map with a matching entry ----
    {
      code: '<CustomMarquee />',
      settings: {
        'jsx-a11y': { components: { CustomMarquee: 'marquee' } },
      },
      errors: [expectedError('marquee')],
    },

    // ---- Polymorphic prop without an allow-list — every truthy `as` value
    //      replaces rawType ----
    {
      code: '<Foo as="marquee" />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [expectedError('marquee')],
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
      errors: [expectedError('blink')],
    },

    // ---- Listener boundary: each violation reports independently ----
    { code: '<div><marquee /></div>', errors: [expectedError('marquee')] },
    {
      code: '<marquee><blink /></marquee>',
      errors: [expectedError('marquee'), expectedError('blink')],
    },

    // ---- Real-world component patterns ----
    {
      code: 'function Banner() { return <marquee>News</marquee>; }',
      errors: [expectedError('marquee')],
    },
    {
      code: 'const x = items.map(item => <blink key={item.id}>{item.text}</blink>)',
      errors: [expectedError('blink')],
    },
    {
      code: 'const x = cond ? <marquee /> : <div />',
      errors: [expectedError('marquee')],
    },
    {
      code: 'class C { render() { return <div><marquee /><blink /></div>; } }',
      errors: [expectedError('marquee'), expectedError('blink')],
    },

    // ---- Deep nesting (3 levels) ----
    {
      code: '<marquee><marquee><marquee /></marquee></marquee>',
      errors: [
        expectedError('marquee'),
        expectedError('marquee'),
        expectedError('marquee'),
      ],
    },

    // ---- JsxFragment wrapper ----
    { code: '<><marquee /></>', errors: [expectedError('marquee')] },

    // ---- JSX inside expression containers (children, attribute, &&) ----
    { code: '<div>{<marquee />}</div>', errors: [expectedError('marquee')] },
    {
      code: '<div>{cond && <marquee />}</div>',
      errors: [expectedError('marquee')],
    },
    {
      code: '<div content={<marquee />} />',
      errors: [expectedError('marquee')],
    },
    {
      code: '<Provider value={data}>{<marquee />}</Provider>',
      errors: [expectedError('marquee')],
    },

    // ---- Components map: map-to-distracting / map-to-self ----
    {
      code: '<Foo />',
      settings: { 'jsx-a11y': { components: { Foo: 'marquee' } } },
      errors: [expectedError('marquee')],
    },
    {
      code: '<marquee />',
      settings: { 'jsx-a11y': { components: { marquee: 'marquee' } } },
      errors: [expectedError('marquee')],
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
      errors: [expectedError('marquee')],
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
      errors: [expectedError('marquee')],
    },

    // ---- Polymorphic prop value forms ----
    {
      code: '<Foo as={"marquee"} />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [expectedError('marquee')],
    },
    {
      code: '<Foo as={`marquee`} />',
      settings: { 'jsx-a11y': { polymorphicPropName: 'as' } },
      errors: [expectedError('marquee')],
    },

    // ---- Real-world patterns (extended) ----
    {
      code: 'const Banner = withTheme(() => <marquee />);',
      errors: [expectedError('marquee')],
    },
    {
      code: 'const x = items.filter(x => x).map(x => <marquee key={x.id} />)',
      errors: [expectedError('marquee')],
    },
    {
      code: 'const items = [<marquee key="1" />, <blink key="2" />];',
      errors: [expectedError('marquee'), expectedError('blink')],
    },
    {
      code: 'wrap(<marquee />);',
      errors: [expectedError('marquee')],
    },
    {
      code: 'const x = { content: <marquee /> };',
      errors: [expectedError('marquee')],
    },
    {
      code: 'const Marquee = React.forwardRef((props, ref) => <marquee ref={ref} {...props} />);',
      errors: [expectedError('marquee')],
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
      errors: [expectedError('marquee')],
    },

    // ---- rslint-specific options (no schema validation) ----
    {
      code: '<marquee />',
      options: [{ elements: ['marquee', 123] }],
      errors: [expectedError('marquee')],
    },
    {
      code: '<marquee />',
      options: [{ elements: ['', 'marquee'] }],
      errors: [expectedError('marquee')],
    },
    {
      code: '<marquee />',
      options: [{ elements: ['marquee', 'marquee'] }],
      errors: [expectedError('marquee')],
    },
  ],
});
