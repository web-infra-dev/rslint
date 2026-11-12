import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// (role, requiredProps) — mirrors aria-query's per-role `requiredProps`
// keys for every non-abstract role with at least one required prop.
// The order of the values matches `Object.keys(requiredProps)` (= aria-query
// insertion order), so the comma-joined error message matches upstream
// byte-for-byte. Mirrors the Go-side `jsxa11yutil.AriaRoleRequiredProps`.
const roleRequiredProps: Record<string, string[]> = {
  checkbox: ['aria-checked'],
  combobox: ['aria-controls', 'aria-expanded'],
  heading: ['aria-level'],
  menuitemcheckbox: ['aria-checked'],
  menuitemradio: ['aria-checked'],
  meter: ['aria-valuenow'],
  option: ['aria-selected'],
  radio: ['aria-checked'],
  scrollbar: ['aria-controls', 'aria-valuenow'],
  slider: ['aria-valuenow'],
  switch: ['aria-checked'],
  treeitem: ['aria-selected'],
};

const componentsSettings = {
  'jsx-a11y': {
    components: { MyComponent: 'div' },
  },
};

// aria-query's `roles.keys()` non-abstract subset — same list as
// `jsxa11yutil.AriaRoleNonAbstract`. Used to synthesize a "valid for every
// role with its required props" suite, mirroring upstream's
// `createTests(validRoleTests)`.
const validRoles = [
  'alert',
  'alertdialog',
  'application',
  'article',
  'banner',
  'blockquote',
  'button',
  'caption',
  'cell',
  'checkbox',
  'code',
  'columnheader',
  'combobox',
  'complementary',
  'contentinfo',
  'definition',
  'deletion',
  'dialog',
  'directory',
  'document',
  'emphasis',
  'feed',
  'figure',
  'form',
  'generic',
  'grid',
  'gridcell',
  'group',
  'heading',
  'img',
  'insertion',
  'link',
  'list',
  'listbox',
  'listitem',
  'log',
  'main',
  'mark',
  'marquee',
  'math',
  'menu',
  'menubar',
  'menuitem',
  'menuitemcheckbox',
  'menuitemradio',
  'meter',
  'navigation',
  'none',
  'note',
  'option',
  'paragraph',
  'presentation',
  'progressbar',
  'radio',
  'radiogroup',
  'region',
  'row',
  'rowgroup',
  'rowheader',
  'scrollbar',
  'search',
  'searchbox',
  'separator',
  'slider',
  'spinbutton',
  'status',
  'strong',
  'subscript',
  'superscript',
  'switch',
  'tab',
  'table',
  'tablist',
  'tabpanel',
  'term',
  'textbox',
  'time',
  'timer',
  'toolbar',
  'tooltip',
  'tree',
  'treegrid',
  'treeitem',
  // DPUB-ARIA.
  'doc-abstract',
  'doc-acknowledgments',
  'doc-afterword',
  'doc-appendix',
  'doc-backlink',
  'doc-biblioentry',
  'doc-bibliography',
  'doc-biblioref',
  'doc-chapter',
  'doc-colophon',
  'doc-conclusion',
  'doc-cover',
  'doc-credit',
  'doc-credits',
  'doc-dedication',
  'doc-endnote',
  'doc-endnotes',
  'doc-epigraph',
  'doc-epilogue',
  'doc-errata',
  'doc-example',
  'doc-footnote',
  'doc-foreword',
  'doc-glossary',
  'doc-glossref',
  'doc-index',
  'doc-introduction',
  'doc-noteref',
  'doc-notice',
  'doc-pagebreak',
  'doc-pagefooter',
  'doc-pageheader',
  'doc-pagelist',
  'doc-part',
  'doc-preface',
  'doc-prologue',
  'doc-pullquote',
  'doc-qna',
  'doc-subtitle',
  'doc-tip',
  'doc-toc',
  // Graphics-ARIA.
  'graphics-document',
  'graphics-object',
  'graphics-symbol',
];

new RuleTester(config).run('role-has-required-aria-props', {
  valid: [
    // Upstream test file — order preserved.
    { code: '<Bar baz />' },
    { code: '<MyComponent role="combobox" />' },
    { code: '<div />' },
    { code: '<div></div>' },
    { code: '<div role={role} />' },
    { code: '<div role={role || "button"} />' },
    { code: '<div role={role || "foobar"} />' },
    { code: '<div role="row" />' },
    {
      code: '<span role="checkbox" aria-checked="false" aria-labelledby="foo" tabindex="0"></span>',
    },
    {
      code: '<input role="checkbox" aria-checked="false" aria-labelledby="foo" tabindex="0" {...props} type="checkbox" />',
    },
    { code: '<input type="checkbox" role="switch" />' },
    {
      code: '<MyComponent role="checkbox" aria-checked="false" aria-labelledby="foo" tabindex="0" />',
      settings: componentsSettings,
    },
    { code: '<div role="heading" aria-level={2} />' },
    { code: '<div role="heading" aria-level="3" />' },
    // Generated valid cases — one per non-abstract role with required props.
    ...validRoles.map((role) => {
      const required = roleRequiredProps[role] ?? [];
      const propChain = required.length === 0 ? '' : ` ${required.join(' ')}`;
      return { code: `<div role="${role}"${propChain} />` };
    }),
  ],
  invalid: [
    // SLIDER.
    {
      code: '<div role="slider" />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "slider" must have the following attributes defined: aria-valuenow',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div role="slider" aria-valuemax />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "slider" must have the following attributes defined: aria-valuenow',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div role="slider" aria-valuemax aria-valuemin />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "slider" must have the following attributes defined: aria-valuenow',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },

    // CHECKBOX.
    {
      code: '<div role="checkbox" />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "checkbox" must have the following attributes defined: aria-checked',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div role="checkbox" checked />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "checkbox" must have the following attributes defined: aria-checked',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div role="checkbox" aria-chcked />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "checkbox" must have the following attributes defined: aria-checked',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<span role="checkbox" aria-labelledby="foo" tabindex="0"></span>',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "checkbox" must have the following attributes defined: aria-checked',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // COMBOBOX.
    {
      code: '<div role="combobox" />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "combobox" must have the following attributes defined: aria-controls,aria-expanded',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div role="combobox" expanded />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "combobox" must have the following attributes defined: aria-controls,aria-expanded',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div role="combobox" aria-expandd />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "combobox" must have the following attributes defined: aria-controls,aria-expanded',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },

    // SCROLLBAR.
    {
      code: '<div role="scrollbar" />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "scrollbar" must have the following attributes defined: aria-controls,aria-valuenow',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div role="scrollbar" aria-valuemax />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "scrollbar" must have the following attributes defined: aria-controls,aria-valuenow',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div role="scrollbar" aria-valuemax aria-valuemin />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "scrollbar" must have the following attributes defined: aria-controls,aria-valuenow',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div role="scrollbar" aria-valuemax aria-valuenow />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "scrollbar" must have the following attributes defined: aria-controls,aria-valuenow',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div role="scrollbar" aria-valuemin aria-valuenow />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "scrollbar" must have the following attributes defined: aria-controls,aria-valuenow',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },

    // HEADING / OPTION.
    {
      code: '<div role="heading" />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "heading" must have the following attributes defined: aria-level',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div role="option" />',
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "option" must have the following attributes defined: aria-selected',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },

    // Custom element + components map.
    {
      code: '<MyComponent role="combobox" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'role-has-required-aria-props',
          message:
            'Elements with the ARIA role "combobox" must have the following attributes defined: aria-controls,aria-expanded',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
  ],
});

// Additional upstream v6.10.2 inputs. Matrix rows freeze the expected
// classification and configuration; ranges follow the authored source templates.
// No rule implementation or lint output is consulted when generating cases.

new RuleTester(config).run('role-has-required-aria-props', {
  valid: [
    ...(
      [
        [
          14,
          'command composite input landmark range roletype section sectionhead select structure widget window',
        ],
      ] as const
    ).flatMap(([firstCase, values]) =>
      values.split(' ').map((role, index) => {
        const openingElement = `<div role="${role}"  />`;
        const code = openingElement;
        return {
          name: `upstream v6.10.2 role-has-required-aria-props valid ${firstCase + index}`,
          code,
        };
      }),
    ),
  ],
  invalid: [],
});
