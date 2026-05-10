import { RuleTester } from '../rule-tester';

const errorMessage =
  'The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users.';
const expectedError = { message: errorMessage };

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

new RuleTester().run('no-autofocus', null as never, {
  valid: [
    // ---- Upstream valid ----
    { code: '<div />;' },
    // Lowercase `autofocus` is the HTML DOM attribute, not React's prop.
    { code: '<div autofocus />;' },
    { code: '<input autofocus="true" />;' },
    { code: '<Foo bar />' },
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
    { code: '<div autoFocus={"false"!} />' },
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
    { code: '<div autoFocus />', errors: [expectedError] },
    { code: '<div autoFocus={true} />', errors: [expectedError] },
    { code: '<div autoFocus={undefined} />', errors: [expectedError] },
    { code: '<div autoFocus="true" />', errors: [expectedError] },
    { code: '<input autoFocus />', errors: [expectedError] },
    { code: '<Foo autoFocus />', errors: [expectedError] },
    {
      code: '<Button autoFocus />',
      errors: [expectedError],
      settings: componentsSettings,
    },
    {
      code: '<Button autoFocus />',
      errors: [expectedError],
      options: ignoreNonDOMSchema,
      settings: componentsSettings,
    },

    // ---- rslint extras: paired element + nested listener boundary ----
    { code: '<div autoFocus>child</div>', errors: [expectedError] },
    {
      code: '<a autoFocus><span autoFocus /></a>',
      errors: [expectedError, expectedError],
    },

    // ---- Element kind survey ----
    { code: '<a autoFocus />', errors: [expectedError] },
    { code: '<Component autoFocus />', errors: [expectedError] },
    { code: '<UX.Layout autoFocus />', errors: [expectedError] },
    { code: '<svg:circle autoFocus />', errors: [expectedError] },

    // ---- Truthy literal coercions ----
    { code: '<div autoFocus="True" />', errors: [expectedError] },
    { code: '<div autoFocus="TRUE" />', errors: [expectedError] },
    { code: '<div autoFocus={"true"} />', errors: [expectedError] },

    // ---- Non-coerced falsy values still report ----
    { code: '<div autoFocus={null} />', errors: [expectedError] },
    { code: '<div autoFocus={0} />', errors: [expectedError] },
    { code: '<div autoFocus={""} />', errors: [expectedError] },
    { code: '<div autoFocus="" />', errors: [expectedError] },
    { code: '<div autoFocus={void 0} />', errors: [expectedError] },

    // ---- Numeric / BigInt / Identifier expressions ----
    { code: '<div autoFocus={1} />', errors: [expectedError] },
    { code: '<div autoFocus={1n} />', errors: [expectedError] },
    { code: '<div autoFocus={someVar} />', errors: [expectedError] },

    // ---- Conditional / logical resolving to truthy ----
    { code: '<div autoFocus={true && true} />', errors: [expectedError] },
    { code: '<div autoFocus={false || true} />', errors: [expectedError] },
    { code: '<div autoFocus={"x" && true} />', errors: [expectedError] },

    // ---- Call / Member / Template ----
    { code: '<div autoFocus={fn()} />', errors: [expectedError] },
    { code: '<div autoFocus={obj.x} />', errors: [expectedError] },
    { code: '<div autoFocus={obj?.x} />', errors: [expectedError] },
    { code: '<div autoFocus={tag`x`} />', errors: [expectedError] },
    { code: '<div autoFocus={`${x}`} />', errors: [expectedError] },
    { code: '<div autoFocus={`true`} />', errors: [expectedError] },

    // ---- TS wrappers around truthy ----
    { code: '<div autoFocus={true as boolean} />', errors: [expectedError] },
    { code: '<div autoFocus={(true)} />', errors: [expectedError] },
    { code: '<div autoFocus={(true)!} />', errors: [expectedError] },

    // ---- `satisfies` is opaque — even `false satisfies boolean` reports ----
    {
      code: '<div autoFocus={false satisfies boolean} />',
      errors: [expectedError],
    },
    {
      code: '<div autoFocus={"false" satisfies string} />',
      errors: [expectedError],
    },

    // ---- TemplateExpression with `${false}` substitution → truthy string ----
    { code: '<div autoFocus={`${false}`} />', errors: [expectedError] },
    {
      code: '<div autoFocus={`prefix${false}suffix`} />',
      errors: [expectedError],
    },

    // ---- Deep nesting × ignoreNonDOM ----
    {
      code: '<Outer><Mid><input autoFocus /></Mid></Outer>',
      options: ignoreNonDOMSchema,
      errors: [expectedError],
    },
    {
      code: '<Custom autoFocus><input autoFocus /></Custom>',
      options: ignoreNonDOMSchema,
      errors: [expectedError],
    },
    {
      code: '<div autoFocus><span autoFocus /></div>',
      options: ignoreNonDOMSchema,
      errors: [expectedError, expectedError],
    },

    // ---- ignoreNonDOM: true × DOM element ----
    {
      code: '<input autoFocus />',
      options: ignoreNonDOMSchema,
      errors: [expectedError],
    },
    {
      code: '<button autoFocus />',
      options: ignoreNonDOMSchema,
      errors: [expectedError],
    },

    // ---- ignoreNonDOM: true × polymorphicPropName resolving to DOM ----
    {
      code: '<Box as="input" autoFocus />',
      options: ignoreNonDOMSchema,
      settings: polymorphicSettings,
      errors: [expectedError],
    },
    {
      code: '<Box as="div" autoFocus />',
      options: ignoreNonDOMSchema,
      settings: polymorphicSettings,
      errors: [expectedError],
    },

    // ---- Multiple autoFocus attributes on one element ----
    {
      code: '<div autoFocus autoFocus />',
      errors: [expectedError, expectedError],
    },

    // ---- Real-world component patterns ----
    {
      code: 'function LoginField() { return <input type="text" autoFocus />; }',
      errors: [expectedError],
    },
    {
      code: 'const x = items.map(item => <li autoFocus key={item.id} />)',
      errors: [expectedError],
    },
    {
      code: 'const x = <>{cond && <input autoFocus />}</>',
      errors: [expectedError],
    },

    // ---- Form patterns: search / login / signup ----
    {
      code: 'function SearchBar() { return <input type="search" autoFocus placeholder="Search…" />; }',
      errors: [expectedError],
    },
    {
      code: 'function LoginForm() { return <form><input autoFocus name="user" /><input type="password" name="pass" /></form>; }',
      errors: [expectedError],
    },
    {
      code: 'function Signup() { return (<form><label>Email<input type="email" autoFocus /></label><label>Pass<input type="password" autoFocus /></label></form>); }',
      errors: [expectedError, expectedError],
    },

    // ---- Modal / Dialog with offending autoFocus ----
    {
      code: 'function Modal() { return <dialog open><input autoFocus name="title" /><button>OK</button></dialog>; }',
      errors: [expectedError],
    },

    // ---- HOC / forwardRef / memo wrappers carrying autoFocus ----
    {
      code: 'const Enhanced = withTracking(({ value }) => <input value={value} autoFocus />);',
      errors: [expectedError],
    },
    {
      code: 'const FocusInput = React.forwardRef((props, ref) => <input ref={ref} autoFocus {...props} />);',
      errors: [expectedError],
    },
    {
      code: 'const Item = React.memo(({ id }) => <li id={id} autoFocus>{id}</li>);',
      errors: [expectedError],
    },

    // ---- TypeScript generic JSX ----
    { code: '<List<string> autoFocus />', errors: [expectedError] },
    { code: '<Cell<{a: number}> autoFocus={true} />', errors: [expectedError] },

    // ---- Long-chain member-expression tags (without ignoreNonDOM) ----
    { code: '<Foo.Bar.Baz autoFocus />', errors: [expectedError] },
    { code: '<this.Foo autoFocus />', errors: [expectedError] },

    // ---- Hyphenated DOM tags without ignoreNonDOM ----
    { code: '<my-element autoFocus />', errors: [expectedError] },

    // ---- Comments around / inside the prop don't suppress ----
    { code: '<div /* a */ autoFocus /* b */ />', errors: [expectedError] },
    {
      code: '<div autoFocus={/* truthy */ true} />',
      errors: [expectedError],
    },

    // ---- Complex value expressions ----
    { code: '<div autoFocus={[1, 2, 3]} />', errors: [expectedError] },
    { code: '<div autoFocus={{nested: true}} />', errors: [expectedError] },
    {
      code: '<div autoFocus={new Boolean(false)} />',
      errors: [expectedError],
    },

    // ---- Optional chain / nullish ----
    { code: '<div autoFocus={cfg?.autoFocus} />', errors: [expectedError] },
    { code: '<div autoFocus={cfg ?? true} />', errors: [expectedError] },

    // ---- ignoreNonDOM × polymorphicAllowList ----
    {
      code: '<Box as="input" autoFocus />',
      options: ignoreNonDOMSchema,
      settings: polymorphicAllowListSettings,
      errors: [expectedError],
    },

    // ---- Conditional rendering forms ----
    {
      code: 'function Foo({cond}) { return cond && <input autoFocus />; }',
      errors: [expectedError],
    },
    {
      code: 'function Foo({cond}) { return cond ? <input autoFocus /> : <div />; }',
      errors: [expectedError],
    },
    {
      code: 'function Foo({a, b}) { return a ? <input autoFocus /> : <textarea autoFocus />; }',
      errors: [expectedError, expectedError],
    },

    // ---- Switch / list / multi-component / generator / async / IIFE ----
    {
      code: "function Foo({type}) { switch(type) { case 'input': return <input autoFocus />; default: return null; } }",
      errors: [expectedError],
    },
    {
      code: 'const items = arr.map((x, i) => <li key={i} autoFocus>{x}</li>);',
      errors: [expectedError],
    },
    {
      code: 'function A() { return <input autoFocus />; }\nfunction B() { return <input autoFocus />; }',
      errors: [expectedError, expectedError],
    },
    {
      code: 'class Form extends React.Component { state = {ready: true}; render() { return this.state.ready ? <input autoFocus /> : <div />; } }',
      errors: [expectedError],
    },
    {
      code: 'function* render() { yield <input autoFocus />; yield <textarea autoFocus />; }',
      errors: [expectedError, expectedError],
    },
    {
      code: 'async function render() { return <input autoFocus />; }',
      errors: [expectedError],
    },
    {
      code: 'const x = (() => <input autoFocus />)();',
      errors: [expectedError],
    },
    {
      code: 'function Form({ initial: { autoFocus = true } }) { return <input autoFocus={autoFocus} />; }',
      errors: [expectedError],
    },
  ],
});
