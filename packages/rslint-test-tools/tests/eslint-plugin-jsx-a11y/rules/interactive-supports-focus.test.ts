import { RuleTester } from '../rule-tester';

const tabbableMessage = (role: string) =>
  `Elements with the '${role}' interactive role must be tabbable.`;
const focusableMessage = (role: string) =>
  `Elements with the '${role}' interactive role must be focusable.`;

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Div: 'div',
    },
  },
};

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

const recommendedOptions = [
  {
    tabbable: [
      'button',
      'checkbox',
      'link',
      'searchbox',
      'spinbutton',
      'switch',
      'textbox',
    ],
  },
];

const strictOptions = [
  {
    tabbable: [
      'button',
      'checkbox',
      'link',
      'progressbar',
      'searchbox',
      'slider',
      'spinbutton',
      'switch',
      'textbox',
    ],
  },
];

new RuleTester().run('interactive-supports-focus', null as never, {
  valid: [
    // ============================================================
    // Upstream alwaysValid sample (mirrors the file's `alwaysValid` array)
    // ============================================================
    { code: '<div />' },
    { code: '<div aria-hidden onClick={() => void 0} />' },
    { code: '<div onClick={() => void 0} />;' },
    { code: '<div onClick={() => void 0} tabIndex={undefined} />;' },
    { code: '<div onClick={() => void 0} tabIndex="bad" />;' },
    { code: '<div onClick={() => void 0} role={undefined} />;' },
    { code: '<div role="section" onClick={() => void 0} />' },
    { code: '<div onClick={() => void 0} {...props} />;' },
    { code: '<input type="text" onClick={() => void 0} />' },
    { code: '<button onClick={() => void 0} className="foo" />' },
    { code: '<select onClick={() => void 0} className="foo" />' },
    { code: '<textarea onClick={() => void 0} className="foo" />' },
    { code: '<a onClick={() => void 0} />' },
    { code: '<a tabIndex="0" onClick={() => void 0} />' },
    { code: '<a tabIndex={0} onClick={() => void 0} />' },
    { code: '<a role="button" href="#" onClick={() => void 0} />' },
    { code: '<a onClick={() => void 0} href="http://x.y.z" />' },
    { code: '<a onClick={() => void 0} href="http://x.y.z" role="button" />' },
    { code: '<TestComponent onClick={doFoo} />' },
    { code: '<input onClick={() => void 0} type="hidden" />;' },
    { code: '<section onClick={() => void 0} />;' },
    { code: '<main onClick={() => void 0} />;' },
    { code: '<header onClick={() => void 0} />;' },
    {
      code: '<div role="textbox" aria-disabled="true" onClick={() => void 0} />',
    },
    { code: '<div role="button" tabIndex="0" onClick={() => void 0} />' },
    { code: '<div role="menuitem" tabIndex="0" onClick={() => void 0} />' },
    { code: '<div role="link" tabIndex="0" onClick={() => void 0} />' },

    // ---- componentsSetting maps <Div> → div; <Div role="button" tabIndex="0">
    //      is already focusable.
    {
      code: '<Div onClick={() => void 0} role="button" tabIndex="0" />',
      settings: componentsSettings,
    },

    // ============================================================
    // Recommended config (default tabbable list)
    // ============================================================
    // ---- gridcell is interactive but NOT in recommended.tabbable; adding
    //      tabIndex="0" satisfies the rule under recommended.
    {
      code: '<div role="gridcell" tabIndex="0" onClick={() => void 0} />',
      options: recommendedOptions,
    },
    // ---- onFocus is not a triggering handler — rule short-circuits.
    {
      code: '<div role="button" onFocus={() => void 0} />',
      options: recommendedOptions,
    },

    // ============================================================
    // Strict config (broader tabbable list)
    // ============================================================
    // ---- progressbar is in strict.tabbable; tabIndex="0" works.
    {
      code: '<div role="progressbar" tabIndex="0" onClick={() => void 0} />',
      options: strictOptions,
    },

    // ============================================================
    // Polymorphic — `as="button"` resolves the JSX component to <button>
    // which is inherently focusable.
    // ============================================================
    {
      code: '<Foo as="button" onClick={() => void 0} />',
      settings: polymorphicSettings,
    },
  ],
  invalid: [
    // ============================================================
    // Default (no options) — tabbable list is empty so everything that
    // qualifies receives the "focusable" message with two suggestions.
    // ============================================================
    {
      code: '<div role="button" onClick={() => void 0} />',
      errors: [{ message: focusableMessage('button') }],
    },
    {
      code: '<div role="checkbox" onMouseDown={() => void 0} />',
      errors: [{ message: focusableMessage('checkbox') }],
    },
    {
      code: '<div role="slider" onKeyDown={() => void 0} />',
      errors: [{ message: focusableMessage('slider') }],
    },
    // ---- componentsSetting → role="button" + onClick.
    {
      code: '<Div onClick={() => void 0} role="button" />',
      settings: componentsSettings,
      errors: [{ message: focusableMessage('button') }],
    },

    // ============================================================
    // Recommended config — button is in tabbable → "must be tabbable"
    // ============================================================
    {
      code: '<div role="button" onClick={() => void 0} />',
      options: recommendedOptions,
      errors: [{ message: tabbableMessage('button') }],
    },
    {
      code: '<div role="checkbox" onClick={() => void 0} />',
      options: recommendedOptions,
      errors: [{ message: tabbableMessage('checkbox') }],
    },
    // ---- gridcell is NOT in recommended.tabbable → "must be focusable"
    {
      code: '<div role="gridcell" onClick={() => void 0} />',
      options: recommendedOptions,
      errors: [{ message: focusableMessage('gridcell') }],
    },
    {
      code: '<div role="menuitem" onClick={() => void 0} />',
      options: recommendedOptions,
      errors: [{ message: focusableMessage('menuitem') }],
    },

    // ============================================================
    // Strict config — slider / progressbar are in tabbable
    // ============================================================
    {
      code: '<div role="slider" onClick={() => void 0} />',
      options: strictOptions,
      errors: [{ message: tabbableMessage('slider') }],
    },
    {
      code: '<div role="progressbar" onClick={() => void 0} />',
      options: strictOptions,
      errors: [{ message: tabbableMessage('progressbar') }],
    },
    // ---- gridcell still gets "focusable" under strict.
    {
      code: '<div role="gridcell" onClick={() => void 0} />',
      options: strictOptions,
      errors: [{ message: focusableMessage('gridcell') }],
    },

    // ============================================================
    // role first valid is interactive — multi-role string variants.
    // ============================================================
    {
      code: '<div role="button heading" onClick={() => void 0} />',
      errors: [{ message: focusableMessage('button heading') }],
    },

    // ============================================================
    // Polymorphic — `as="div"` resolves to a plain div, which still trips.
    // ============================================================
    {
      code: '<Foo as="div" role="button" onClick={() => void 0} />',
      settings: polymorphicSettings,
      errors: [{ message: focusableMessage('button') }],
    },
  ],
});
