import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// `aria-query`'s `roles.keys()`, partitioned by `abstract === false` / `=== true`.
// Order matches the live aria-query output (ARIA 1.x core alphabetic,
// followed by DPUB-ARIA `doc-*`, then Graphics-ARIA `graphics-*`). Mirrors
// the Go-side `jsxa11yutil.AriaRoleNonAbstract` / `AriaRoleAbstract` data.
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
  // DPUB-ARIA roles.
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
  // Graphics-ARIA roles.
  'graphics-document',
  'graphics-object',
  'graphics-symbol',
];

const allowedInvalidRolesOption = [
  { allowedInvalidRoles: ['invalid-role', 'other-invalid-role'] },
];
const ignoreNonDOMOption = [{ ignoreNonDOM: true }];

const customDivSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'asChild',
    components: { Div: 'div' },
  },
};

new RuleTester(config).run('aria-role', {
  valid: [
    // Upstream test file — order preserved.
    { code: '<div />' },
    { code: '<div></div>' },
    { code: '<div role={role} />' },
    { code: '<div role={role || "button"} />' },
    { code: '<div role={role || "foobar"} />' },
    { code: '<div role="tabpanel row" />' },
    { code: '<div role="switch" />' },
    { code: '<div role="doc-abstract" />' },
    { code: '<div role="doc-appendix doc-bibliography" />' },
    { code: '<Bar baz />' },
    {
      code: '<img role="invalid-role" />',
      options: allowedInvalidRolesOption,
    },
    {
      code: '<img role="invalid-role tabpanel" />',
      options: allowedInvalidRolesOption,
    },
    {
      code: '<img role="invalid-role other-invalid-role" />',
      options: allowedInvalidRolesOption,
    },
    { code: '<Foo role="bar" />', options: ignoreNonDOMOption },
    { code: '<fakeDOM role="bar" />', options: ignoreNonDOMOption },
    { code: '<img role="presentation" />', options: ignoreNonDOMOption },
    { code: '<Div role="button" />', settings: customDivSettings },
    {
      code: '<Box asChild="div" role="button" />',
      settings: customDivSettings,
    },
    { code: '<svg role="graphics-document document" />' },
    { code: '<svg role="img" />' },
    // Upstream's createTests(validRoles) — one per non-abstract role.
    ...validRoles.map((role) => ({ code: `<div role="${role}" />` })),
  ],
  invalid: [
    {
      code: '<div role="foobar" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div role="datepicker"></div>',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<div role="range"></div>',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div role="Button"></div>',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div role=""></div>',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<div role="tabpanel row foobar"></div>',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<div role="tabpanel row range"></div>',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div role="doc-endnotes range"></div>',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<div role />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<div role="unknown-invalid-role" />',
      options: [
        {
          allowedInvalidRoles: ['invalid-role', 'other-invalid-role'],
        },
      ],
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div role={null}></div>',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<Foo role="datepicker" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<Foo role="Button" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<Div role="Button" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'asChild',
          components: {
            Div: 'div',
          },
        },
      },
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<Div role="Button" />',
      options: [
        {
          ignoreNonDOM: true,
        },
      ],
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'asChild',
          components: {
            Div: 'div',
          },
        },
      },
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<Box asChild="div" role="Button" />',
      settings: {
        'jsx-a11y': {
          polymorphicPropName: 'asChild',
          components: {
            Div: 'div',
          },
        },
      },
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<div role="command" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div role="composite" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div role="input" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div role="landmark" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div role="range" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<div role="roletype" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<div role="section" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<div role="sectionhead" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<div role="select" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div role="structure" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<div role="widget" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<div role="window" />',
      errors: [
        {
          messageId: 'invalidAriaRole',
          message:
            'Elements with ARIA roles must use a valid, non-abstract ARIA role.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
  ],
});
