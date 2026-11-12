import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';
import { singleLineError } from '../matrix';

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Article: 'article',
      Input: 'input',
    },
  },
};

// Mirrors `configs.recommended.rules['jsx-a11y/no-noninteractive-element-to-interactive-role'][1]`.
const recommendedOptions = [
  {
    ul: [
      'listbox',
      'menu',
      'menubar',
      'radiogroup',
      'tablist',
      'tree',
      'treegrid',
    ],
    ol: [
      'listbox',
      'menu',
      'menubar',
      'radiogroup',
      'tablist',
      'tree',
      'treegrid',
    ],
    li: [
      'menuitem',
      'menuitemradio',
      'menuitemcheckbox',
      'option',
      'row',
      'tab',
      'treeitem',
    ],
    table: ['grid'],
    td: ['gridcell'],
    fieldset: ['radiogroup', 'presentation'],
  },
];

new RuleTester(config).run('no-noninteractive-element-to-interactive-role', {
  valid: [
    // ============================================================
    // Upstream alwaysValid (run under both :recommended and :strict)
    // ============================================================
    // Custom JSX components — element type isn't in aria-query's dom map.
    { code: '<TestComponent onClick={doFoo} />' },
    { code: '<Button onClick={doFoo} />' },

    // Interactive elements with an interactive role.
    { code: '<a tabIndex="0" role="button" />' },
    { code: '<a href="http://x.y.z" role="button" />' },
    { code: '<a href="http://x.y.z" tabIndex="0" role="button" />' },
    { code: '<area role="button" />;' },
    { code: '<area role="menuitem" />;' },
    { code: '<button className="foo" role="button" />' },
    { code: '<body role="button" />;' },
    { code: '<frame role="button" />;' },
    { code: '<td role="button" />;' },
    { code: '<frame role="menuitem" />;' },
    { code: '<td role="menuitem" />;' },

    // All flavors of input + role="button" — input is inherently interactive.
    { code: '<input role="button" />' },
    { code: '<input type="button" role="button" />' },
    { code: '<input type="checkbox" role="button" />' },
    { code: '<input type="color" role="button" />' },
    { code: '<input type="date" role="button" />' },
    { code: '<input type="datetime" role="button" />' },
    { code: '<input type="datetime-local" role="button" />' },
    { code: '<input type="email" role="button" />' },
    { code: '<input type="file" role="button" />' },
    { code: '<input type="hidden" role="button" />' },
    { code: '<input type="image" role="button" />' },
    { code: '<input type="month" role="button" />' },
    { code: '<input type="number" role="button" />' },
    { code: '<input type="password" role="button" />' },
    { code: '<input type="radio" role="button" />' },
    { code: '<input type="range" role="button" />' },
    { code: '<input type="reset" role="button" />' },
    { code: '<input type="search" role="button" />' },
    { code: '<input type="submit" role="button" />' },
    { code: '<input type="tel" role="button" />' },
    { code: '<input type="text" role="button" />' },
    { code: '<input type="time" role="button" />' },
    { code: '<input type="url" role="button" />' },
    { code: '<input type="week" role="button" />' },
    { code: '<input type="hidden" role="img" />' },

    // Other inherently interactive controls.
    { code: '<menuitem role="button" />;' },
    { code: '<option className="foo" role="button" />' },
    { code: '<select className="foo" role="button" />' },
    { code: '<textarea className="foo" role="button" />' },
    { code: '<tr role="button" />;' },
    { code: '<tr role="presentation" />;' },

    // HTML elements that have neither interactive nor non-interactive
    // valence — div, span, etc. Element gate fails.
    { code: '<acronym role="button" />;' },
    { code: '<canvas role="button" />;' },
    { code: '<div role="button" />;' },
    { code: '<div role={undefined} role="button" />;' },
    { code: '<div className="foo" {...props} role="button" />;' },
    { code: '<span role="button" />;' },
    { code: '<small role="button" />;' },
    // <header> is hard-coded false in isNonInteractiveElement (banner
    // landmark depends on parent context).
    { code: '<header role="button" />;' },

    // Non-interactive element + non-interactive role — role gate fails.
    { code: '<main role="listitem" />;' },
    { code: '<article role="listitem" />;' },
    { code: '<li role="listitem" />;' },
    { code: '<li role="presentation" />;' },
    { code: '<ul role="listitem" />;' },
    { code: '<ul role="list" />;' },
    { code: '<table role="listitem" />;' },

    // Abstract role — IsInteractiveRole returns false for abstract.
    { code: '<div role="command" />;' },
    { code: '<div role="composite" />;' },
    { code: '<div role="widget" />;' },

    // Non-DOM custom components are exempt.
    { code: '<Article role="button" />' },
    { code: '<Input role="button" />', settings: componentsSettings },

    // ============================================================
    // Recommended-only valid (list/li/fieldset allow-list)
    // ============================================================
    { code: '<ul role="menu" />;', options: recommendedOptions },
    { code: '<ul role="menubar" />;', options: recommendedOptions },
    { code: '<ul role="radiogroup" />;', options: recommendedOptions },
    { code: '<ul role="tablist" />;', options: recommendedOptions },
    { code: '<ul role="tree" />;', options: recommendedOptions },
    { code: '<ul role="treegrid" />;', options: recommendedOptions },
    { code: '<ol role="menu" />;', options: recommendedOptions },
    { code: '<ol role="menubar" />;', options: recommendedOptions },
    { code: '<ol role="radiogroup" />;', options: recommendedOptions },
    { code: '<ol role="tablist" />;', options: recommendedOptions },
    { code: '<ol role="tree" />;', options: recommendedOptions },
    { code: '<ol role="treegrid" />;', options: recommendedOptions },
    { code: '<li role="tab" />;', options: recommendedOptions },
    { code: '<li role="menuitem" />;', options: recommendedOptions },
    { code: '<li role="menuitemcheckbox" />;', options: recommendedOptions },
    { code: '<li role="menuitemradio" />;', options: recommendedOptions },
    { code: '<li role="row" />;', options: recommendedOptions },
    { code: '<li role="treeitem" />;', options: recommendedOptions },
    { code: '<Component role="treeitem" />;', options: recommendedOptions },
    { code: '<fieldset role="radiogroup" />;', options: recommendedOptions },
    {
      code: '<fieldset role="presentation" />;',
      options: recommendedOptions,
    },
  ],
  invalid: [
    // ============================================================
    // Upstream neverValid — invalid in both recommended and strict
    // ============================================================
    {
      code: '<address role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<article role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<aside role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<blockquote role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<br role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<caption role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<code role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<dd role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<del role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<details role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<dfn role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<dir role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<dl role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<dt role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<em role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<fieldset role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<figcaption role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<figure role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<footer role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<form role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<h1 role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<h2 role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<h3 role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<h4 role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<h5 role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<h6 role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<hr role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<html role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<iframe role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<img role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<ins role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<label role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<legend role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<li role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<main role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<mark role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<marquee role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<menu role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<meter role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<nav role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<ol role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<optgroup role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<output role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<pre role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<progress role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<ruby role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<strong role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<sub role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<sup role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<table role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<tbody role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<tfoot role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<thead role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<time role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<ul role="button" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },

    // Non-interactive elements + role="menuitem".
    {
      code: '<main role="menuitem" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<article role="menuitem" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<h1 role="menuitem" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<img role="menuitem" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<table role="menuitem" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // Custom component resolved to <article> via settings.
    {
      code: '<Article role="button" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ============================================================
    // Strict-only invalid (no allowed-roles override for ul/ol/li/...)
    // ============================================================
    {
      code: '<ul role="menu" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<ul role="menubar" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<ul role="tablist" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<ol role="menu" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: '<li role="tab" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<li role="menuitem" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<li role="row" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<li role="treeitem" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
  ],
});

// Additional upstream v6.10.2 inputs. Matrix rows freeze the expected
// classification and configuration; ranges follow the authored source templates.
// No rule implementation or lint output is consulted when generating cases.
const upstreamOptions0 = [
  {
    ul: [
      'listbox',
      'menu',
      'menubar',
      'radiogroup',
      'tablist',
      'tree',
      'treegrid',
    ],
    ol: [
      'listbox',
      'menu',
      'menubar',
      'radiogroup',
      'tablist',
      'tree',
      'treegrid',
    ],
    li: [
      'menuitem',
      'menuitemradio',
      'menuitemcheckbox',
      'option',
      'row',
      'tab',
      'treeitem',
    ],
    table: ['grid'],
    td: ['gridcell'],
    fieldset: ['radiogroup', 'presentation'],
  },
];
const upstreamSettings0 = {
  'jsx-a11y': { components: { Article: 'article', Input: 'input' } },
};

new RuleTester(config).run('no-noninteractive-element-to-interactive-role', {
  valid: [
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 44',
      code: '<a tabIndex="0" role="img" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 45',
      code: '<a href="http://x.y.z" role="img" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 46',
      code: '<a href="http://x.y.z" tabIndex="0" role="img" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 47',
      code: '<input role="img" />',
    },
    ...(
      [
        [
          'img',
          '',
          48,
          'img checkbox color date datetime datetime-local email file',
        ],
        [
          'img',
          '',
          57,
          'image month number password radio range reset search submit tel text time url week',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 71',
      code: '<menuitem role="img" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 72',
      code: '<option className="foo" role="img" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 73',
      code: '<select className="foo" role="img" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 74',
      code: '<textarea className="foo" role="img" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 75',
      code: '<tr role="img" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 76',
      code: '<a tabIndex="0" role="listitem" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 77',
      code: '<a href="http://x.y.z" role="listitem" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 78',
      code: '<a href="http://x.y.z" tabIndex="0" role="listitem" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 79',
      code: '<input role="listitem" />',
    },
    ...(
      [
        [
          'listitem',
          '',
          80,
          'listitem checkbox color date datetime datetime-local email file image month number password radio range reset search submit tel text time url week',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 102',
      code: '<menuitem role="listitem" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 103',
      code: '<option className="foo" role="listitem" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 104',
      code: '<select className="foo" role="listitem" />',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 105',
      code: '<textarea className="foo" role="listitem" />',
    },
    ...(
      [
        ['listitem', ';', 106, 'tr'],
        ['button', ';', 108, 'applet audio b base bdi bdo big blink'],
        ['button', ';', 117, 'center cite col colgroup content data datalist'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 125',
      code: '<div className="foo" role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 127',
      code: '<div aria-hidden role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 128',
      code: '<div aria-hidden={true} role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 131',
      code: '<div {...props} role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid 132',
      code: '<div onKeyUp={() => void 0} aria-hidden={false} role="button" />;',
    },
    ...(
      [
        ['button', ';', 133, 'embed font frameset head'],
        [
          'button',
          ';',
          138,
          'hgroup i kbd keygen link map meta noembed noscript object param picture q rp rt rtc s samp script',
        ],
        ['button', ';', 158, 'source spacer'],
        [
          'button',
          ';',
          161,
          'strike style summary th title track tt u var video wbr xmp',
        ],
        ['checkbox', ';', 174, 'div'],
        ['columnheader', ';', 175, 'div'],
        ['combobox', ';', 176, 'div'],
        ['grid', ';', 177, 'div'],
        ['gridcell', ';', 178, 'div'],
        ['link', ';', 179, 'div'],
        ['listbox', ';', 180, 'div'],
        ['menu', ';', 181, 'div'],
        ['menubar', ';', 182, 'div'],
        ['menuitem', ';', 183, 'div'],
        ['menuitemcheckbox', ';', 184, 'div'],
        ['menuitemradio', ';', 185, 'div'],
        ['option', ';', 186, 'div'],
        ['progressbar', ';', 187, 'div'],
        ['radio', ';', 188, 'div'],
        ['radiogroup', ';', 189, 'div'],
        ['row', ';', 190, 'div'],
        ['rowheader', ';', 191, 'div'],
        ['searchbox', ';', 192, 'div'],
        ['slider', ';', 193, 'div'],
        ['spinbutton', ';', 194, 'div'],
        ['switch', ';', 195, 'div'],
        ['tab', ';', 196, 'div'],
        ['textbox', ';', 197, 'div'],
        ['treeitem', ';', 198, 'div'],
        ['presentation', ';', 199, 'div'],
        ['input', ';', 202, 'div'],
        ['landmark', ';', 203, 'div'],
        ['range', ';', 204, 'div'],
        ['roletype', ';', 205, 'div'],
        ['section', ';', 206, 'div'],
        ['sectionhead', ';', 207, 'div'],
        ['select', ';', 208, 'div'],
        ['structure', ';', 209, 'div'],
        ['tablist', ';', 210, 'div'],
        ['toolbar', ';', 211, 'div'],
        ['tree', ';', 212, 'div'],
        ['treegrid', ';', 213, 'div'],
        ['window', ';', 215, 'div'],
        ['listitem', '', 217, 'a'],
        ['button', '', 219, 'a'],
        ['menuitem', '', 221, 'a'],
        ['listitem', ';', 223, 'area'],
        [
          'listitem',
          ';',
          226,
          'dd dfn dt fieldset figure form frame h1 h2 h3 h4 h5 h6 hr img',
        ],
        ['listitem', ';', 243, 'nav ol'],
        ['listitem', ';', 246, 'tbody td tfoot thead'],
        ['alert', ';', 251, 'div'],
        ['alertdialog', ';', 252, 'div'],
        ['application', ';', 253, 'div'],
        ['article', ';', 254, 'div'],
        ['banner', ';', 255, 'div'],
        ['cell', ';', 256, 'div'],
        ['complementary', ';', 257, 'div'],
        ['contentinfo', ';', 258, 'div'],
        ['definition', ';', 259, 'div'],
        ['dialog', ';', 260, 'div'],
        ['directory', ';', 261, 'div'],
        ['document', ';', 262, 'div'],
        ['feed', ';', 263, 'div'],
        ['figure', ';', 264, 'div'],
        ['form', ';', 265, 'div'],
        ['group', ';', 266, 'div'],
        ['heading', ';', 267, 'div'],
        ['img', ';', 268, 'div'],
        ['list', ';', 269, 'div'],
        ['listitem', ';', 270, 'div'],
        ['log', ';', 271, 'div'],
        ['main', ';', 272, 'div'],
        ['marquee', ';', 273, 'div'],
        ['math', ';', 274, 'div'],
        ['navigation', ';', 275, 'div'],
        ['note', ';', 276, 'div'],
        ['region', ';', 277, 'div'],
        ['rowgroup', ';', 278, 'div'],
        ['search', ';', 279, 'div'],
        ['separator', ';', 280, 'div'],
        ['scrollbar', ';', 281, 'div'],
        ['status', ';', 282, 'div'],
        ['table', ';', 283, 'div'],
        ['tabpanel', ';', 284, 'div'],
        ['term', ';', 285, 'div'],
        ['timer', ';', 286, 'div'],
        ['tooltip', ';', 287, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 44',
      code: '<a tabIndex="0" role="img" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 45',
      code: '<a href="http://x.y.z" role="img" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 46',
      code: '<a href="http://x.y.z" tabIndex="0" role="img" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 47',
      code: '<input role="img" />',
      options: upstreamOptions0,
    },
    ...(
      [
        [
          'img',
          '',
          48,
          'img checkbox color date datetime datetime-local email file',
        ],
        [
          'img',
          '',
          57,
          'image month number password radio range reset search submit tel text time url week',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 71',
      code: '<menuitem role="img" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 72',
      code: '<option className="foo" role="img" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 73',
      code: '<select className="foo" role="img" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 74',
      code: '<textarea className="foo" role="img" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 75',
      code: '<tr role="img" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 76',
      code: '<a tabIndex="0" role="listitem" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 77',
      code: '<a href="http://x.y.z" role="listitem" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 78',
      code: '<a href="http://x.y.z" tabIndex="0" role="listitem" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 79',
      code: '<input role="listitem" />',
      options: upstreamOptions0,
    },
    ...(
      [
        [
          'listitem',
          '',
          80,
          'listitem checkbox color date datetime datetime-local email file image month number password radio range reset search submit tel text time url week',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 102',
      code: '<menuitem role="listitem" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 103',
      code: '<option className="foo" role="listitem" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 104',
      code: '<select className="foo" role="listitem" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 105',
      code: '<textarea className="foo" role="listitem" />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['listitem', ';', 106, 'tr'],
        ['button', ';', 108, 'applet audio b base bdi bdo big blink'],
        ['button', ';', 117, 'center cite col colgroup content data datalist'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 125',
      code: '<div className="foo" role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 127',
      code: '<div aria-hidden role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 128',
      code: '<div aria-hidden={true} role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 131',
      code: '<div {...props} role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 132',
      code: '<div onKeyUp={() => void 0} aria-hidden={false} role="button" />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['button', ';', 133, 'embed font frameset head'],
        [
          'button',
          ';',
          138,
          'hgroup i kbd keygen link map meta noembed noscript object param picture q rp rt rtc s samp script',
        ],
        ['button', ';', 158, 'source spacer'],
        [
          'button',
          ';',
          161,
          'strike style summary th title track tt u var video wbr xmp',
        ],
        ['checkbox', ';', 174, 'div'],
        ['columnheader', ';', 175, 'div'],
        ['combobox', ';', 176, 'div'],
        ['grid', ';', 177, 'div'],
        ['gridcell', ';', 178, 'div'],
        ['link', ';', 179, 'div'],
        ['listbox', ';', 180, 'div'],
        ['menu', ';', 181, 'div'],
        ['menubar', ';', 182, 'div'],
        ['menuitem', ';', 183, 'div'],
        ['menuitemcheckbox', ';', 184, 'div'],
        ['menuitemradio', ';', 185, 'div'],
        ['option', ';', 186, 'div'],
        ['progressbar', ';', 187, 'div'],
        ['radio', ';', 188, 'div'],
        ['radiogroup', ';', 189, 'div'],
        ['row', ';', 190, 'div'],
        ['rowheader', ';', 191, 'div'],
        ['searchbox', ';', 192, 'div'],
        ['slider', ';', 193, 'div'],
        ['spinbutton', ';', 194, 'div'],
        ['switch', ';', 195, 'div'],
        ['tab', ';', 196, 'div'],
        ['textbox', ';', 197, 'div'],
        ['treeitem', ';', 198, 'div'],
        ['presentation', ';', 199, 'div'],
        ['input', ';', 202, 'div'],
        ['landmark', ';', 203, 'div'],
        ['range', ';', 204, 'div'],
        ['roletype', ';', 205, 'div'],
        ['section', ';', 206, 'div'],
        ['sectionhead', ';', 207, 'div'],
        ['select', ';', 208, 'div'],
        ['structure', ';', 209, 'div'],
        ['tablist', ';', 210, 'div'],
        ['toolbar', ';', 211, 'div'],
        ['tree', ';', 212, 'div'],
        ['treegrid', ';', 213, 'div'],
        ['window', ';', 215, 'div'],
        ['listitem', '', 217, 'a'],
        ['button', '', 219, 'a'],
        ['menuitem', '', 221, 'a'],
        ['listitem', ';', 223, 'area'],
        [
          'listitem',
          ';',
          226,
          'dd dfn dt fieldset figure form frame h1 h2 h3 h4 h5 h6 hr img',
        ],
        ['listitem', ';', 243, 'nav ol'],
        ['listitem', ';', 246, 'tbody td tfoot thead'],
        ['alert', ';', 251, 'div'],
        ['alertdialog', ';', 252, 'div'],
        ['application', ';', 253, 'div'],
        ['article', ';', 254, 'div'],
        ['banner', ';', 255, 'div'],
        ['cell', ';', 256, 'div'],
        ['complementary', ';', 257, 'div'],
        ['contentinfo', ';', 258, 'div'],
        ['definition', ';', 259, 'div'],
        ['dialog', ';', 260, 'div'],
        ['directory', ';', 261, 'div'],
        ['document', ';', 262, 'div'],
        ['feed', ';', 263, 'div'],
        ['figure', ';', 264, 'div'],
        ['form', ';', 265, 'div'],
        ['group', ';', 266, 'div'],
        ['heading', ';', 267, 'div'],
        ['img', ';', 268, 'div'],
        ['list', ';', 269, 'div'],
        ['listitem', ';', 270, 'div'],
        ['log', ';', 271, 'div'],
        ['main', ';', 272, 'div'],
        ['marquee', ';', 273, 'div'],
        ['math', ';', 274, 'div'],
        ['navigation', ';', 275, 'div'],
        ['note', ';', 276, 'div'],
        ['region', ';', 277, 'div'],
        ['rowgroup', ';', 278, 'div'],
        ['search', ';', 279, 'div'],
        ['separator', ';', 280, 'div'],
        ['scrollbar', ';', 281, 'div'],
        ['status', ';', 282, 'div'],
        ['table', ';', 283, 'div'],
        ['tabpanel', ';', 284, 'div'],
        ['term', ';', 285, 'div'],
        ['timer', ';', 286, 'div'],
        ['tooltip', ';', 287, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 0',
      code: '<TestComponent onClick={doFoo} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 1',
      code: '<Button onClick={doFoo} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 2',
      code: '<a tabIndex="0" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 3',
      code: '<a href="http://x.y.z" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 4',
      code: '<a href="http://x.y.z" tabIndex="0" role="button" />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['button', ';', 5, 'area'],
        ['menuitem', ';', 6, 'area'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 7',
      code: '<button className="foo" role="button" />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['button', ';', 8, 'body frame td'],
        ['menuitem', ';', 11, 'frame td'],
        ['button', '', 13, 'input'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    ...(
      [
        [
          'button',
          '',
          14,
          'button checkbox color date datetime datetime-local email file hidden image month number password radio range reset search submit tel text time url week',
        ],
        ['img', '', 37, 'hidden'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 38',
      code: '<menuitem role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 39',
      code: '<option className="foo" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 40',
      code: '<select className="foo" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 41',
      code: '<textarea className="foo" role="button" />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['button', ';', 42, 'tr'],
        ['presentation', ';', 43, 'tr'],
        ['button', ';', 107, 'acronym'],
        ['button', ';', 116, 'canvas'],
        ['button', ';', 124, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 126',
      code: '<div className="foo" {...props} role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 130',
      code: '<div role={undefined} role="button" />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['button', ';', 137, 'header'],
        ['button', ';', 157, 'small'],
        ['button', ';', 160, 'span'],
        ['command', ';', 200, 'div'],
        ['composite', ';', 201, 'div'],
        ['widget', ';', 214, 'div'],
        ['listitem', ';', 216, 'main'],
        ['listitem', ';', 224, 'article'],
        ['listitem', ';', 241, 'li'],
        ['presentation', ';', 242, 'li'],
        ['listitem', ';', 245, 'table'],
        ['listitem', ';', 250, 'ul'],
        ['list', ';', 288, 'ul'],
        ['button', '', 289, 'Article'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended valid 290',
      code: '<Input role="button" />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
    },
  ],
  invalid: [
    ...(
      [
        ['menuitem', ';', 57, 'dd dfn dt fieldset figure form'],
        ['menuitem', ';', 64, 'h2 h3 h4 h5 h6 hr'],
        ['menuitem', ';', 71, 'nav ol'],
        ['button', ';', 73, 'p'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict invalid ${firstCase + index}`,
          code,
          errors: [
            singleLineError(
              'noNoninteractiveElementToInteractiveRole',
              'Non-interactive elements should not be assigned interactive roles.',
              `role="${role}"`,
              `<${tag} `.length + 1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict invalid 74',
      code: '<section role="button" aria-label="Aardvark" />;',
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    ...(
      [
        ['menuitem', ';', 76, 'tbody tfoot thead'],
        ['radiogroup', ';', 82, 'ul'],
        ['tree', ';', 84, 'ul'],
        ['treegrid', ';', 85, 'ul'],
        ['menubar', ';', 87, 'ol'],
        ['radiogroup', ';', 88, 'ol'],
        ['tablist', ';', 89, 'ol'],
        ['tree', ';', 90, 'ol'],
        ['treegrid', ';', 91, 'ol'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:strict invalid ${firstCase + index}`,
          code,
          errors: [
            singleLineError(
              'noNoninteractiveElementToInteractiveRole',
              'Non-interactive elements should not be assigned interactive roles.',
              `role="${role}"`,
              `<${tag} `.length + 1,
            ),
          ],
        };
      }),
    ),
    ...(
      [
        ['menuitem', ';', 57, 'dd dfn dt fieldset figure form'],
        ['menuitem', ';', 64, 'h2 h3 h4 h5 h6 hr'],
        ['menuitem', ';', 71, 'nav ol'],
        ['button', ';', 73, 'p'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noNoninteractiveElementToInteractiveRole',
              'Non-interactive elements should not be assigned interactive roles.',
              `role="${role}"`,
              `<${tag} `.length + 1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended invalid 74',
      code: '<section role="button" aria-label="Aardvark" />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    ...(
      [
        ['menuitem', ';', 76, 'tbody tfoot thead'],
        [
          'button',
          ';',
          0,
          'address article aside blockquote br caption code dd del details dfn dir dl dt em fieldset figcaption figure footer form h1 h2 h3 h4 h5 h6 hr html iframe img ins label legend li main mark marquee menu meter nav ol optgroup output pre progress ruby strong sub sup table tbody tfoot thead time ul',
        ],
        ['menuitem', ';', 55, 'main article'],
        ['menuitem', ';', 63, 'h1'],
        ['menuitem', ';', 70, 'img'],
        ['menuitem', ';', 75, 'table'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noNoninteractiveElementToInteractiveRole',
              'Non-interactive elements should not be assigned interactive roles.',
              `role="${role}"`,
              `<${tag} `.length + 1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-noninteractive-element-to-interactive-role:recommended invalid 79',
      code: '<Article role="button" />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
      errors: [
        {
          messageId: 'noNoninteractiveElementToInteractiveRole',
          message:
            'Non-interactive elements should not be assigned interactive roles.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
  ],
});
