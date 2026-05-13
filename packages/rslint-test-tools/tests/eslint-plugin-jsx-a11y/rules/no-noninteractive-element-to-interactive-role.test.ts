import { RuleTester } from '../rule-tester';

const errorMessage =
  'Non-interactive elements should not be assigned interactive roles.';
const expectedError = { message: errorMessage };

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

new RuleTester().run(
  'no-noninteractive-element-to-interactive-role',
  null as never,
  {
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
      { code: '<address role="button" />;', errors: [expectedError] },
      { code: '<article role="button" />;', errors: [expectedError] },
      { code: '<aside role="button" />;', errors: [expectedError] },
      { code: '<blockquote role="button" />;', errors: [expectedError] },
      { code: '<br role="button" />;', errors: [expectedError] },
      { code: '<caption role="button" />;', errors: [expectedError] },
      { code: '<code role="button" />;', errors: [expectedError] },
      { code: '<dd role="button" />;', errors: [expectedError] },
      { code: '<del role="button" />;', errors: [expectedError] },
      { code: '<details role="button" />;', errors: [expectedError] },
      { code: '<dfn role="button" />;', errors: [expectedError] },
      { code: '<dir role="button" />;', errors: [expectedError] },
      { code: '<dl role="button" />;', errors: [expectedError] },
      { code: '<dt role="button" />;', errors: [expectedError] },
      { code: '<em role="button" />;', errors: [expectedError] },
      { code: '<fieldset role="button" />;', errors: [expectedError] },
      { code: '<figcaption role="button" />;', errors: [expectedError] },
      { code: '<figure role="button" />;', errors: [expectedError] },
      { code: '<footer role="button" />;', errors: [expectedError] },
      { code: '<form role="button" />;', errors: [expectedError] },
      { code: '<h1 role="button" />;', errors: [expectedError] },
      { code: '<h2 role="button" />;', errors: [expectedError] },
      { code: '<h3 role="button" />;', errors: [expectedError] },
      { code: '<h4 role="button" />;', errors: [expectedError] },
      { code: '<h5 role="button" />;', errors: [expectedError] },
      { code: '<h6 role="button" />;', errors: [expectedError] },
      { code: '<hr role="button" />;', errors: [expectedError] },
      { code: '<html role="button" />;', errors: [expectedError] },
      { code: '<iframe role="button" />;', errors: [expectedError] },
      { code: '<img role="button" />;', errors: [expectedError] },
      { code: '<ins role="button" />;', errors: [expectedError] },
      { code: '<label role="button" />;', errors: [expectedError] },
      { code: '<legend role="button" />;', errors: [expectedError] },
      { code: '<li role="button" />;', errors: [expectedError] },
      { code: '<main role="button" />;', errors: [expectedError] },
      { code: '<mark role="button" />;', errors: [expectedError] },
      { code: '<marquee role="button" />;', errors: [expectedError] },
      { code: '<menu role="button" />;', errors: [expectedError] },
      { code: '<meter role="button" />;', errors: [expectedError] },
      { code: '<nav role="button" />;', errors: [expectedError] },
      { code: '<ol role="button" />;', errors: [expectedError] },
      { code: '<optgroup role="button" />;', errors: [expectedError] },
      { code: '<output role="button" />;', errors: [expectedError] },
      { code: '<pre role="button" />;', errors: [expectedError] },
      { code: '<progress role="button" />;', errors: [expectedError] },
      { code: '<ruby role="button" />;', errors: [expectedError] },
      { code: '<strong role="button" />;', errors: [expectedError] },
      { code: '<sub role="button" />;', errors: [expectedError] },
      { code: '<sup role="button" />;', errors: [expectedError] },
      { code: '<table role="button" />;', errors: [expectedError] },
      { code: '<tbody role="button" />;', errors: [expectedError] },
      { code: '<tfoot role="button" />;', errors: [expectedError] },
      { code: '<thead role="button" />;', errors: [expectedError] },
      { code: '<time role="button" />;', errors: [expectedError] },
      { code: '<ul role="button" />;', errors: [expectedError] },

      // Non-interactive elements + role="menuitem".
      { code: '<main role="menuitem" />;', errors: [expectedError] },
      { code: '<article role="menuitem" />;', errors: [expectedError] },
      { code: '<h1 role="menuitem" />;', errors: [expectedError] },
      { code: '<img role="menuitem" />;', errors: [expectedError] },
      { code: '<table role="menuitem" />;', errors: [expectedError] },

      // Custom component resolved to <article> via settings.
      {
        code: '<Article role="button" />',
        settings: componentsSettings,
        errors: [expectedError],
      },

      // ============================================================
      // Strict-only invalid (no allowed-roles override for ul/ol/li/...)
      // ============================================================
      { code: '<ul role="menu" />;', errors: [expectedError] },
      { code: '<ul role="menubar" />;', errors: [expectedError] },
      { code: '<ul role="tablist" />;', errors: [expectedError] },
      { code: '<ol role="menu" />;', errors: [expectedError] },
      { code: '<li role="tab" />;', errors: [expectedError] },
      { code: '<li role="menuitem" />;', errors: [expectedError] },
      { code: '<li role="row" />;', errors: [expectedError] },
      { code: '<li role="treeitem" />;', errors: [expectedError] },
    ],
  },
);
