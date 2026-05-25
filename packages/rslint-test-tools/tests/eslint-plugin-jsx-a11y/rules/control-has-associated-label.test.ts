import { RuleTester } from '../rule-tester';

const errorMessage = 'A control must be associated with a text label.';
const expectedError = { message: errorMessage };

// Mirrors `configs.recommended` / `configs.strict` from
// eslint-plugin-jsx-a11y. Both presets ship identical options for this
// rule; depth is absent so per-test overrides win.
const recommendedOptions = {
  ignoreElements: [
    'audio',
    'canvas',
    'embed',
    'input',
    'textarea',
    'tr',
    'video',
  ],
  ignoreRoles: [
    'grid',
    'listbox',
    'menu',
    'menubar',
    'radiogroup',
    'row',
    'tablist',
    'toolbar',
    'tree',
    'treegrid',
  ],
};

const customControlSettings = {
  'jsx-a11y': {
    components: {
      CustomControl: 'button',
    },
  },
};

const polymorphicSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
  },
};

new RuleTester().run('control-has-associated-label', null as never, {
  valid: [
    // ============================================================
    // Upstream :recommended / :strict valid suite (mirrors
    // __tests__/src/rules/control-has-associated-label-test.js, with
    // recommendedOptions merged in via ruleOptionsMapperFactory).
    // ============================================================
    {
      code: '<CustomControl><span><span>Save</span></span></CustomControl>',
      options: [
        {
          depth: 3,
          controlComponents: ['CustomControl'],
          ...recommendedOptions,
        },
      ],
    },
    {
      code: '<CustomControl>Save</CustomControl>',
      options: [recommendedOptions],
      settings: customControlSettings,
    },
    { code: '<button>Save</button>', options: [recommendedOptions] },
    {
      code: '<button><span>Save</span></button>',
      options: [recommendedOptions],
    },
    {
      code: '<button aria-label="Save" />',
      options: [recommendedOptions],
    },
    {
      code: '<button aria-labelledby="js_1" />',
      options: [recommendedOptions],
    },
    {
      code: '<button>{sureWhyNot}</button>',
      options: [recommendedOptions],
    },
    { code: '<a href="#">Save</a>', options: [recommendedOptions] },
    { code: '<link>Save</link>', options: [recommendedOptions] },
    { code: '<th>Save</th>', options: [recommendedOptions] },
    {
      code: '<div role="button">Save</div>',
      options: [recommendedOptions],
    },
    {
      code: '<div role="button" aria-label="Save" />',
      options: [recommendedOptions],
    },
    // Non-interactive elements / roles
    { code: '<div role="alert" />', options: [recommendedOptions] },
    { code: '<article />', options: [recommendedOptions] },
    { code: '<p />', options: [recommendedOptions] },
    // Inputs / marginal interactive elements — skipped via ignoreElements.
    { code: '<input />', options: [recommendedOptions] },
    { code: '<input type="text" />', options: [recommendedOptions] },
    { code: '<textarea />', options: [recommendedOptions] },
    { code: '<tr />', options: [recommendedOptions] },
    // Interactive roles in ignoreRoles.
    { code: '<div role="grid" />', options: [recommendedOptions] },
    { code: '<div role="tablist" />', options: [recommendedOptions] },

    // ============================================================
    // :no-config — hidden elements skip regardless of options.
    // ============================================================
    { code: '<input type="hidden" />' },
    { code: '<input type="text" aria-hidden="true" />' },

    // ============================================================
    // Listener gate edge cases.
    // ============================================================
    // Hard-coded `link` ignore — not removable via empty ignoreElements.
    { code: '<link />', options: [{ ignoreElements: [] }] },
    // aria-hidden on a control short-circuits.
    { code: '<button aria-hidden />' },
    { code: '<button aria-hidden={true} />' },
    { code: '<button aria-hidden="true" />' },
    // Spread attribute opacity — counts as labelling.
    { code: '<button {...props} />' },
    // JsxExpression child — assumed label.
    { code: '<button>{maybeLabel}</button>' },
    // React-component fallback inside recursion.
    { code: '<button><MyLabel /></button>' },
    // settings.components: <MyButton aria-label="Save" /> demotes to button.
    {
      code: '<MyButton aria-label="Save" />',
      settings: {
        'jsx-a11y': { components: { MyButton: 'button' } },
      },
    },
    // settings.polymorphicPropName: <Foo as="button">Save</Foo>.
    {
      code: '<Foo as="button">Save</Foo>',
      settings: polymorphicSettings,
    },

    // ============================================================
    // Real-world JSX patterns
    // ============================================================
    // Multi-line button with label.
    {
      code: ['<button', '  className="primary"', '>Save</button>'].join('\n'),
    },
    // Newline-indented text child.
    { code: '<button>\n  Save\n</button>' },
    // Non-ASCII text.
    { code: '<button>保存</button>' },
    // Render-prop pattern with labelled inner control.
    { code: '<DataLoader>{(data) => <button>Save</button>}</DataLoader>' },
    // Suspense + labelled child.
    {
      code: '<Suspense fallback={<Spinner />}><button>Save</button></Suspense>',
    },
    // Array.map with labelled controls.
    { code: 'items.map((item) => <button key={item.id}>{item.name}</button>)' },
    // Logical && rendering.
    { code: '{loading && <button aria-label="Loading" />}' },
    // Conditional with both branches labelled.
    { code: '{cond ? <button>A</button> : <button>B</button>}' },
    // TS generic on a component (treated as custom — no trigger).
    { code: '<Select<string> options={opts} />' },
    // Member-expression JSX tag (custom — no trigger).
    { code: '<Foo.Bar />' },
    // Namespaced lowercase JSX tag (not in DOM map — no trigger).
    { code: '<svg:path />' },

    // ============================================================
    // labelAttributes — hyphenated prop name.
    // ============================================================
    {
      code: '<button data-label="Save" />',
      options: [{ labelAttributes: ['data-label'] }],
    },
    // labelAttributes — duplicate entry, defensive.
    {
      code: '<button title="Save" />',
      options: [{ labelAttributes: ['title', 'title'] }],
    },

    // ============================================================
    // ignoreRoles vs case-sensitivity / multi-role
    // ============================================================
    // role="GRID" not in ignoreRoles=['grid'] (case-sensitive) but
    // isInteractiveRole lowercases — combined: role-trigger fires →
    // label required → "Save" satisfies.
    {
      code: '<div role="GRID">Save</div>',
      options: [{ ignoreRoles: ['grid'] }],
    },
    // Multi-role: first valid role wins for isInteractiveRole.
    { code: '<div role="button switch">Save</div>' },

    // ============================================================
    // depth extremes
    // ============================================================
    // depth=0 with label directly on root.
    { code: '<button aria-label="Save" />', options: [{ depth: 0 }] },
    // depth=999 clamped to 25 — still resolves a shallow label.
    { code: '<button><span>Save</span></button>', options: [{ depth: 999 }] },
  ],
  invalid: [
    // ============================================================
    // Upstream :recommended / :strict invalid suite.
    // ============================================================
    {
      code: '<button />',
      options: [recommendedOptions],
      errors: [expectedError],
    },
    {
      code: '<button><span /></button>',
      options: [recommendedOptions],
      errors: [expectedError],
    },
    {
      code: '<button><img /></button>',
      options: [recommendedOptions],
      errors: [expectedError],
    },
    {
      code: '<CustomControl><span><span></span></span></CustomControl>',
      options: [
        {
          depth: 3,
          controlComponents: ['CustomControl'],
          ...recommendedOptions,
        },
      ],
      errors: [expectedError],
    },
    {
      code: '<CustomControl></CustomControl>',
      options: [recommendedOptions],
      settings: customControlSettings,
      errors: [expectedError],
    },
    {
      code: '<a href="#" />',
      options: [recommendedOptions],
      errors: [expectedError],
    },
    {
      code: '<area href="#" />',
      options: [recommendedOptions],
      errors: [expectedError],
    },
    {
      code: '<menuitem />',
      options: [recommendedOptions],
      errors: [expectedError],
    },
    {
      code: '<option />',
      options: [recommendedOptions],
      errors: [expectedError],
    },
    { code: '<th />', options: [recommendedOptions], errors: [expectedError] },
    { code: '<td />', options: [recommendedOptions], errors: [expectedError] },
    {
      code: '<div role="button" />',
      options: [recommendedOptions],
      errors: [expectedError],
    },
    {
      code: '<div role="link" />',
      options: [recommendedOptions],
      errors: [expectedError],
    },

    // ============================================================
    // :no-config — <input type="text" /> reports because input is
    // interactive when ignoreElements is empty.
    // ============================================================
    { code: '<input type="text" />', errors: [expectedError] },

    // ============================================================
    // Listener gate: aria-label trim semantics.
    // ============================================================
    { code: '<button aria-label="" />', errors: [expectedError] },
    { code: '<button aria-label="   " />', errors: [expectedError] },
    { code: '<button aria-label={null} />', errors: [expectedError] },
    { code: '<button aria-label={false} />', errors: [expectedError] },

    // ============================================================
    // Real-world failure modes
    // ============================================================
    // Icon-only button (SVG).
    { code: '<button><svg /></button>', errors: [expectedError] },
    // Icon class on void child.
    {
      code: '<button><i className="icon-save" /></button>',
      errors: [expectedError],
    },
    // onClick without label — common mistake.
    { code: '<button onClick={fn} />', errors: [expectedError] },
    { code: '<a href="#" onClick={fn} />', errors: [expectedError] },
    // depth=0 cannot see labelled descendant.
    {
      code: '<button><span>Save</span></button>',
      options: [{ depth: 0 }],
      errors: [expectedError],
    },
    // Element interactive regardless of role.
    { code: '<button role="presentation" />', errors: [expectedError] },
    { code: '<button role="img" />', errors: [expectedError] },
    // Multi-role: first valid role "button" → trigger.
    { code: '<div role="button presentation" />', errors: [expectedError] },
    // Settings.components remap exposes interactivity.
    {
      code: '<Submit />',
      settings: { 'jsx-a11y': { components: { Submit: 'button' } } },
      errors: [expectedError],
    },
    // polymorphicAllowList lets `<Foo as="th" />` become th.
    {
      code: '<Foo as="th" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'as',
          polymorphicAllowList: ['Foo'],
        },
      },
      errors: [expectedError],
    },
    // Multi-line position assertion.
    {
      code: ['(', '  <button', '    onClick={fn}', '  />', ')'].join('\n'),
      errors: [expectedError],
    },
    // controlComponents glob — minimatch in recursion.
    {
      code: '<button><CustomIcon /></button>',
      options: [{ controlComponents: ['Custom*'] }],
      errors: [expectedError],
    },
  ],
});
