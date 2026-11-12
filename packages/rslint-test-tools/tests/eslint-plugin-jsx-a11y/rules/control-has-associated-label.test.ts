import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';
import { singleLineError } from '../matrix';

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

new RuleTester(config).run('control-has-associated-label', {
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
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<button><span /></button>',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<button><img /></button>',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
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
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<CustomControl></CustomControl>',
      options: [recommendedOptions],
      settings: customControlSettings,
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<a href="#" />',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<area href="#" />',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<menuitem />',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<option />',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<th />',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: '<td />',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: '<div role="button" />',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div role="link" />',
      options: [recommendedOptions],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // ============================================================
    // :no-config — <input type="text" /> reports because input is
    // interactive when ignoreElements is empty.
    // ============================================================
    {
      code: '<input type="text" />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // ============================================================
    // Listener gate: aria-label trim semantics.
    // ============================================================
    {
      code: '<button aria-label="" />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<button aria-label="   " />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<button aria-label={null} />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<button aria-label={false} />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },

    // ============================================================
    // Real-world failure modes
    // ============================================================
    // Icon-only button (SVG).
    {
      code: '<button><svg /></button>',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    // Icon class on void child.
    {
      code: '<button><i className="icon-save" /></button>',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    // onClick without label — common mistake.
    {
      code: '<button onClick={fn} />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<a href="#" onClick={fn} />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    // depth=0 cannot see labelled descendant.
    {
      code: '<button><span>Save</span></button>',
      options: [{ depth: 0 }],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    // Element interactive regardless of role.
    {
      code: '<button role="presentation" />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<button role="img" />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    // Multi-role: first valid role "button" → trigger.
    {
      code: '<div role="button presentation" />',
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    // Settings.components remap exposes interactivity.
    {
      code: '<Submit />',
      settings: { 'jsx-a11y': { components: { Submit: 'button' } } },
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
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
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    // Multi-line position assertion.
    {
      code: ['(', '  <button', '    onClick={fn}', '  />', ')'].join('\n'),
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 2,
          column: 3,
          endLine: 4,
          endColumn: 5,
        },
      ],
    },
    // controlComponents glob — minimatch in recursion.
    {
      code: '<button><CustomIcon /></button>',
      options: [{ controlComponents: ['Custom*'] }],
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
  ],
});

// Upstream presets also contain includeRoles, which the upstream rule ignores.
// Omit that no-op field because the Rslint schema does not accept it.
// Additional upstream v6.10.2 inputs. Matrix rows freeze the expected
// classification and configuration; ranges follow the authored source templates.
// No rule implementation or lint output is consulted when generating cases.
const upstreamOptions0 = [
  {
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
  },
];
const upstreamOptions1 = [
  {
    depth: 3,
    controlComponents: ['CustomControl'],
    labelAttributes: ['label'],
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
  },
];
const upstreamOptions2 = [
  {
    depth: 3,
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
  },
];
const upstreamOptions3 = [
  {
    depth: 9,
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
  },
];
const upstreamOptions4 = [
  {
    depth: 3,
    labelAttributes: ['label'],
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
  },
];

new RuleTester(config).run('control-has-associated-label', {
  valid: [
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 7',
      code: '<button><img alt="Save" /></button>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 9',
      code: '<button><span aria-label="Save" /></button>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 11',
      code: '<button><span aria-labelledby="js_1" /></button>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 15',
      code: '<area href="#">Save</area>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 17',
      code: '<menuitem>Save</menuitem>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 18',
      code: '<option>Save</option>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 21',
      code: '<div role="checkbox">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 22',
      code: '<div role="columnheader">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 23',
      code: '<div role="combobox">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 24',
      code: '<div role="gridcell">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 25',
      code: '<div role="link">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 26',
      code: '<div role="menuitem">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 27',
      code: '<div role="menuitemcheckbox">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 28',
      code: '<div role="menuitemradio">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 29',
      code: '<div role="option">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 30',
      code: '<div role="progressbar">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 31',
      code: '<div role="radio">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 32',
      code: '<div role="rowheader">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 33',
      code: '<div role="searchbox">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 34',
      code: '<div role="slider">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 35',
      code: '<div role="spinbutton">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 36',
      code: '<div role="switch">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 37',
      code: '<div role="tab">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 38',
      code: '<div role="textbox">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 39',
      code: '<div role="treeitem">Save</div>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 41',
      code: '<div role="checkbox" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 42',
      code: '<div role="columnheader" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 43',
      code: '<div role="combobox" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 44',
      code: '<div role="gridcell" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 45',
      code: '<div role="link" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 46',
      code: '<div role="menuitem" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 47',
      code: '<div role="menuitemcheckbox" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 48',
      code: '<div role="menuitemradio" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 49',
      code: '<div role="option" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 50',
      code: '<div role="progressbar" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 51',
      code: '<div role="radio" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 52',
      code: '<div role="rowheader" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 53',
      code: '<div role="searchbox" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 54',
      code: '<div role="slider" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 55',
      code: '<div role="spinbutton" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 56',
      code: '<div role="switch" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 57',
      code: '<div role="tab" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 58',
      code: '<div role="textbox" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 59',
      code: '<div role="treeitem" aria-label="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 60',
      code: '<div role="button" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 61',
      code: '<div role="checkbox" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 62',
      code: '<div role="columnheader" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 63',
      code: '<div role="combobox" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 64',
      code: '<div role="gridcell" aria-labelledby="Save" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 65',
      code: '<div role="link" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 66',
      code: '<div role="menuitem" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 67',
      code: '<div role="menuitemcheckbox" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 68',
      code: '<div role="menuitemradio" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 69',
      code: '<div role="option" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 70',
      code: '<div role="progressbar" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 71',
      code: '<div role="radio" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 72',
      code: '<div role="rowheader" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 73',
      code: '<div role="searchbox" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 74',
      code: '<div role="slider" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 75',
      code: '<div role="spinbutton" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 76',
      code: '<div role="switch" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 77',
      code: '<div role="tab" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 78',
      code: '<div role="textbox" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 79',
      code: '<div role="treeitem" aria-labelledby="js_1" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 80',
      code: '<abbr />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 82',
      code: '<blockquote />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 83',
      code: '<br />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 84',
      code: '<caption />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 85',
      code: '<dd />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 86',
      code: '<details />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 87',
      code: '<dfn />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 88',
      code: '<dialog />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 89',
      code: '<dir />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 90',
      code: '<dl />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 91',
      code: '<dt />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 92',
      code: '<fieldset />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 93',
      code: '<figcaption />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 94',
      code: '<figure />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 95',
      code: '<footer />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 96',
      code: '<form />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 97',
      code: '<frame />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 98',
      code: '<h1 />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 99',
      code: '<h2 />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 100',
      code: '<h3 />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 101',
      code: '<h4 />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 102',
      code: '<h5 />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 103',
      code: '<h6 />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 104',
      code: '<hr />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 105',
      code: '<iframe />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 106',
      code: '<img />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 107',
      code: '<label />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 108',
      code: '<legend />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 109',
      code: '<li />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 111',
      code: '<main />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 112',
      code: '<mark />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 113',
      code: '<marquee />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 114',
      code: '<menu />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 115',
      code: '<meter />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 116',
      code: '<nav />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 117',
      code: '<ol />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 119',
      code: '<pre />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 120',
      code: '<progress />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 121',
      code: '<ruby />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 122',
      code: '<section />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 123',
      code: '<table />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 124',
      code: '<tbody />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 125',
      code: '<tfoot />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 126',
      code: '<thead />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 127',
      code: '<time />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 128',
      code: '<ul />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['alertdialog', '', 130, 'div'],
        ['application', '', 131, 'div'],
        ['article', '', 132, 'div'],
        ['banner', '', 133, 'div'],
        ['cell', '', 134, 'div'],
        ['complementary', '', 135, 'div'],
        ['contentinfo', '', 136, 'div'],
        ['definition', '', 137, 'div'],
        ['dialog', '', 138, 'div'],
        ['directory', '', 139, 'div'],
        ['document', '', 140, 'div'],
        ['feed', '', 141, 'div'],
        ['figure', '', 142, 'div'],
        ['form', '', 143, 'div'],
        ['group', '', 144, 'div'],
        ['heading', '', 145, 'div'],
        ['img', '', 146, 'div'],
        ['list', '', 147, 'div'],
        ['listitem', '', 148, 'div'],
        ['log', '', 149, 'div'],
        ['main', '', 150, 'div'],
        ['marquee', '', 151, 'div'],
        ['math', '', 152, 'div'],
        ['navigation', '', 153, 'div'],
        ['none', '', 154, 'div'],
        ['note', '', 155, 'div'],
        ['presentation', '', 156, 'div'],
        ['region', '', 157, 'div'],
        ['rowgroup', '', 158, 'div'],
        ['search', '', 159, 'div'],
        ['separator', '', 160, 'div'],
        ['status', '', 161, 'div'],
        ['table', '', 162, 'div'],
        ['tabpanel', '', 163, 'div'],
        ['term', '', 164, 'div'],
        ['timer', '', 165, 'div'],
        ['tooltip', '', 166, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 168',
      code: '<input type="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 169',
      code: '<input type="checkbox" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 170',
      code: '<input type="color" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 171',
      code: '<input type="date" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 172',
      code: '<input type="datetime" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 173',
      code: '<input type="email" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 174',
      code: '<input type="file" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 176',
      code: '<input type="hidden" name="bot-field"/>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 177',
      code: '<input type="hidden" name="form-name" value="Contact Form"/>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 178',
      code: '<input type="image" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 179',
      code: '<input type="month" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 180',
      code: '<input type="number" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 181',
      code: '<input type="password" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 182',
      code: '<input type="radio" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 183',
      code: '<input type="range" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 184',
      code: '<input type="reset" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 185',
      code: '<input type="search" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 186',
      code: '<input type="submit" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 187',
      code: '<input type="tel" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 189',
      code: '<label>Foo <input type="text" /></label>',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 190',
      code: '<input name={field.name} id="foo" type="text" value={field.value} disabled={isDisabled} onChange={changeText(field.onChange, field.name)} onBlur={field.onBlur} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 191',
      code: '<input type="time" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 192',
      code: '<input type="url" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 193',
      code: '<input type="week" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 194',
      code: '<audio />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 195',
      code: '<canvas />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 196',
      code: '<embed />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 199',
      code: '<video />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['listbox', '', 201, 'div'],
        ['menu', '', 202, 'div'],
        ['menubar', '', 203, 'div'],
        ['radiogroup', '', 204, 'div'],
        ['row', '', 205, 'div'],
        ['toolbar', '', 207, 'div'],
        ['tree', '', 208, 'div'],
        ['treegrid', '', 209, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 110',
      code: '<link />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 175',
      code: '<input type="hidden" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 1',
      code: '<CustomControl><span><span label="Save"></span></span></CustomControl>',
      options: upstreamOptions1,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 5',
      code: '<button><span><span>Save</span></span></button>',
      options: upstreamOptions2,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 6',
      code: '<button><span><span><span><span><span><span><span><span>Save</span></span></span></span></span></span></span></span></button>',
      options: upstreamOptions3,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended valid 13',
      code: '<button><span><span label="Save"></span></span></button>',
      options: upstreamOptions4,
    },
  ],
  invalid: [
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended invalid 3',
      code: '<button><span title="This is not a real label" /></button>',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    ...(
      [
        ['checkbox', '', 14, 'div'],
        ['columnheader', '', 15, 'div'],
        ['combobox', '', 16, 'div'],
        ['gridcell', '', 18, 'div'],
        ['menuitem', '', 19, 'div'],
        ['menuitemcheckbox', '', 20, 'div'],
        ['menuitemradio', '', 21, 'div'],
        ['option', '', 22, 'div'],
        ['progressbar', '', 23, 'div'],
        ['radio', '', 24, 'div'],
        ['rowheader', '', 25, 'div'],
        ['scrollbar', '', 26, 'div'],
        ['searchbox', '', 27, 'div'],
        ['slider', '', 28, 'div'],
        ['spinbutton', '', 29, 'div'],
        ['switch', '', 30, 'div'],
        ['tab', '', 31, 'div'],
        ['textbox', '', 32, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'controlHasAssociatedLabel',
              'A control must be associated with a text label.',
              openingElement,
              1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/control-has-associated-label:recommended invalid 4',
      code: '<button><span><span><span>Save</span></span></span></button>',
      options: upstreamOptions2,
      errors: [
        {
          messageId: 'controlHasAssociatedLabel',
          message: 'A control must be associated with a text label.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
  ],
});
