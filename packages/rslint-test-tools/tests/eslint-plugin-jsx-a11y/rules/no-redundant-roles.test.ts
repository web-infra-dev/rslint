import { RuleTester } from '../rule-tester';

const errorMessage = (element: string, implicitRole: string): string =>
  `The element ${element} has an implicit role of ${implicitRole}. Defining this explicitly is redundant and should be avoided.`;

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

new RuleTester().run('no-redundant-roles', null as never, {
  valid: [
    // ============================================================
    // Upstream alwaysValid + default-config <nav role="navigation" />
    // ============================================================
    { code: '<div />' },
    { code: '<button role="main" />' },
    { code: '<MyComponent role="button" />' },
    { code: '<button role={`${foo}button`} />' },
    { code: '<Button role={`${foo}button`} />', settings: componentsSettings },
    {
      code: '<select role="menu"><option>1</option><option>2</option></select>',
    },
    {
      code: '<select role="menu" size={2}><option>1</option><option>2</option></select>',
    },
    {
      code: '<select role="menu" multiple><option>1</option><option>2</option></select>',
    },
    { code: '<nav role="navigation" />' },

    // ============================================================
    // Options override: `{ ul: ['list'], ol: ['list'] }` allows
    // role="list" on ul/ol; nav allowance is preserved via default.
    // ============================================================
    {
      code: '<ul role="list" />',
      options: [{ ul: ['list'], ol: ['list'] }],
    },
    {
      code: '<ol role="list" />',
      options: [{ ul: ['list'], ol: ['list'] }],
    },
    // dl has no implicit role → no comparison happens.
    {
      code: '<dl role="list" />',
      options: [{ ul: ['list'], ol: ['list'] }],
    },
    // img + literal .svg src → SVG arm returns '' → no implicit role.
    {
      code: '<img src="example.svg" role="img" />',
      options: [{ ul: ['list'], ol: ['list'] }],
    },
    // svg not in implicitRoles table → no comparison.
    {
      code: '<svg role="img" />',
      options: [{ ul: ['list'], ol: ['list'] }],
    },

    // ============================================================
    // Implicit-role table arms — every non-matching role short-
    // circuits the report.
    // ============================================================
    { code: '<input type="text" role="button" />' },
    { code: '<input role="button" />' }, // implicit textbox
    { code: '<input type={someType} role="button" />' },
    { code: '<input type="range" role="button" />' },
    { code: '<select multiple role="combobox" />' },
    { code: '<select size={5} role="combobox" />' },
    { code: '<select size={1} role="listbox" />' },
    { code: '<select size={undefined} role="listbox" />' },
    { code: '<select size="abc" role="listbox" />' },
    { code: '<a role="link" />' }, // no href → no implicit link
    { code: '<a href="/x" role="button" />' },
    { code: '<menu role="menu" />' }, // no type → no implicit role
    { code: '<menuitem role="menuitem" />' }, // no type → no implicit role
    { code: '<menu type="toolbar" role="menu" />' }, // implicit toolbar
    { code: '<menuitem type="command" role="checkbox" />' },

    // ============================================================
    // Non-implicit elements skip the rule.
    // ============================================================
    { code: '<div role="document" />' },
    { code: '<span role="button" />' },
    { code: '<p role="paragraph" />' },

    // ============================================================
    // Empty / non-literal explicit role short-circuits.
    // ============================================================
    { code: '<button role={someRole} />' },
    { code: '<button role={getRole()} />' },
    { code: '<button role={cond ? "button" : "main"} />' },
    { code: '<button role="" />' },
    { code: '<button role="not-a-real-role" />' },
    { code: '<button role={`role-${suffix}`} />' },

    // ============================================================
    // implicitRoleForImg branches
    // ============================================================
    { code: '<img alt="" role="img" />' }, // empty alt suppresses img
    { code: '<img src="logo.svg" role="img" />' },

    // ============================================================
    // Element-type forms — capitalization, member access.
    // ============================================================
    { code: '<BODY role="document" />' },
    { code: '<Foo.Button role="button" />' },
    { code: '<svg:rect role="img" />' },

    // ============================================================
    // Case-insensitive attribute name (jsx-ast-utils default).
    // ============================================================
    { code: '<button ROLE="main" />' },

    // ============================================================
    // Allow-list option permutations
    // ============================================================
    {
      code: '<button role="button" />',
      options: [{ button: ['button'] }],
    },
    {
      code: '<nav role="navigation" />',
      options: [{ nav: ['navigation'] }],
    },

    // ============================================================
    // Polymorphic / components — no role match
    // ============================================================
    { code: '<Button role="main" />', settings: componentsSettings },
    { code: '<Foo as="div" role="button" />', settings: polymorphicSettings },
  ],
  invalid: [
    // ============================================================
    // Upstream neverValid
    // ============================================================
    {
      code: '<body role="DOCUMENT" />',
      errors: [{ message: errorMessage('body', 'document') }],
    },
    {
      code: '<button role="button" />',
      errors: [{ message: errorMessage('button', 'button') }],
    },
    {
      code: '<Button role="button" />',
      settings: componentsSettings,
      errors: [{ message: errorMessage('button', 'button') }],
    },
    {
      code: '<select role="combobox"><option>1</option><option>2</option></select>',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="combobox" size="" />',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="combobox" size={1} />',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="combobox" size="1" />',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="combobox" size={null}></select>',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="combobox" size={undefined}></select>',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="combobox" multiple={undefined}></select>',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="combobox" multiple={false}></select>',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="combobox" multiple=""></select>',
      errors: [{ message: errorMessage('select', 'combobox') }],
    },
    {
      code: '<select role="listbox" size="3" />',
      errors: [{ message: errorMessage('select', 'listbox') }],
    },
    {
      code: '<select role="listbox" size={2} />',
      errors: [{ message: errorMessage('select', 'listbox') }],
    },
    {
      code: '<select role="listbox" multiple><option>1</option><option>2</option></select>',
      errors: [{ message: errorMessage('select', 'listbox') }],
    },
    {
      code: '<select role="listbox" multiple={true}></select>',
      errors: [{ message: errorMessage('select', 'listbox') }],
    },

    // ============================================================
    // nav allowance disabled
    // ============================================================
    {
      code: '<nav role="navigation" />',
      options: [{ nav: [] }],
      errors: [{ message: errorMessage('nav', 'navigation') }],
    },

    // ============================================================
    // img/ul/ol without override
    // ============================================================
    {
      code: '<ul role="list" />',
      errors: [{ message: errorMessage('ul', 'list') }],
    },
    {
      code: '<ol role="list" />',
      errors: [{ message: errorMessage('ol', 'list') }],
    },
    {
      code: '<img role="img" />',
      errors: [{ message: errorMessage('img', 'img') }],
    },
    {
      code: '<img src={someVariable} role="img" />',
      errors: [{ message: errorMessage('img', 'img') }],
    },

    // ============================================================
    // implicitRoleForImg — alt boolean / undefined / non-literal
    // ============================================================
    {
      code: '<img alt role="img" />',
      errors: [{ message: errorMessage('img', 'img') }],
    },
    {
      code: '<img alt={undefined} role="img" />',
      errors: [{ message: errorMessage('img', 'img') }],
    },

    // ============================================================
    // input type → ARIA role mapping
    // ============================================================
    {
      code: '<input type="button" role="button" />',
      errors: [{ message: errorMessage('input', 'button') }],
    },
    {
      code: '<input type="submit" role="button" />',
      errors: [{ message: errorMessage('input', 'button') }],
    },
    {
      code: '<input type="reset" role="button" />',
      errors: [{ message: errorMessage('input', 'button') }],
    },
    {
      code: '<input type="image" role="button" />',
      errors: [{ message: errorMessage('input', 'button') }],
    },
    {
      code: '<input type="checkbox" role="checkbox" />',
      errors: [{ message: errorMessage('input', 'checkbox') }],
    },
    {
      code: '<input type="radio" role="radio" />',
      errors: [{ message: errorMessage('input', 'radio') }],
    },
    {
      code: '<input type="range" role="slider" />',
      errors: [{ message: errorMessage('input', 'slider') }],
    },
    {
      code: '<input role="textbox" />',
      errors: [{ message: errorMessage('input', 'textbox') }],
    },
    {
      code: '<input type="text" role="textbox" />',
      errors: [{ message: errorMessage('input', 'textbox') }],
    },
    {
      code: '<input type="hidden" role="textbox" />',
      errors: [{ message: errorMessage('input', 'textbox') }],
    },

    // ============================================================
    // Case-insensitive role value
    // ============================================================
    {
      code: '<button role="BUTTON" />',
      errors: [{ message: errorMessage('button', 'button') }],
    },

    // ============================================================
    // Implicit-role table coverage
    // ============================================================
    {
      code: '<article role="article" />',
      errors: [{ message: errorMessage('article', 'article') }],
    },
    {
      code: '<aside role="complementary" />',
      errors: [{ message: errorMessage('aside', 'complementary') }],
    },
    {
      code: '<datalist role="listbox" />',
      errors: [{ message: errorMessage('datalist', 'listbox') }],
    },
    {
      code: '<details role="group" />',
      errors: [{ message: errorMessage('details', 'group') }],
    },
    {
      code: '<dialog role="dialog" />',
      errors: [{ message: errorMessage('dialog', 'dialog') }],
    },
    {
      code: '<form role="form" />',
      errors: [{ message: errorMessage('form', 'form') }],
    },
    {
      code: '<h1 role="heading" />',
      errors: [{ message: errorMessage('h1', 'heading') }],
    },
    {
      code: '<h6 role="heading" />',
      errors: [{ message: errorMessage('h6', 'heading') }],
    },
    {
      code: '<hr role="separator" />',
      errors: [{ message: errorMessage('hr', 'separator') }],
    },
    {
      code: '<li role="listitem" />',
      errors: [{ message: errorMessage('li', 'listitem') }],
    },
    {
      code: '<meter role="progressbar" />',
      errors: [{ message: errorMessage('meter', 'progressbar') }],
    },
    {
      code: '<progress role="progressbar" />',
      errors: [{ message: errorMessage('progress', 'progressbar') }],
    },
    {
      code: '<option role="option" />',
      errors: [{ message: errorMessage('option', 'option') }],
    },
    {
      code: '<output role="status" />',
      errors: [{ message: errorMessage('output', 'status') }],
    },
    {
      code: '<section role="region" />',
      errors: [{ message: errorMessage('section', 'region') }],
    },
    {
      code: '<tbody role="rowgroup" />',
      errors: [{ message: errorMessage('tbody', 'rowgroup') }],
    },
    {
      code: '<textarea role="textbox" />',
      errors: [{ message: errorMessage('textarea', 'textbox') }],
    },

    // ============================================================
    // a/area/link href → link role
    // ============================================================
    {
      code: '<a href="/x" role="link" />',
      errors: [{ message: errorMessage('a', 'link') }],
    },
    {
      code: '<area href="#x" role="link" />',
      errors: [{ message: errorMessage('area', 'link') }],
    },
    {
      code: '<link href="/x" role="link" />',
      errors: [{ message: errorMessage('link', 'link') }],
    },

    // ============================================================
    // menu / menuitem type branches
    // ============================================================
    {
      code: '<menu type="toolbar" role="toolbar" />',
      errors: [{ message: errorMessage('menu', 'toolbar') }],
    },
    {
      code: '<menuitem type="command" role="menuitem" />',
      errors: [{ message: errorMessage('menuitem', 'menuitem') }],
    },
    {
      code: '<menuitem type="checkbox" role="menuitemcheckbox" />',
      errors: [{ message: errorMessage('menuitem', 'menuitemcheckbox') }],
    },
    {
      code: '<menuitem type="radio" role="menuitemradio" />',
      errors: [{ message: errorMessage('menuitem', 'menuitemradio') }],
    },

    // ============================================================
    // polymorphicPropName remap to redundant target
    // ============================================================
    {
      code: '<Foo as="button" role="button" />',
      settings: polymorphicSettings,
      errors: [{ message: errorMessage('button', 'button') }],
    },

    // ============================================================
    // Paired form — fires on opening tag
    // ============================================================
    {
      code: '<button role="button">Click</button>',
      errors: [{ message: errorMessage('button', 'button') }],
    },

    // ============================================================
    // Literal spread role
    // ============================================================
    {
      code: '<button {...{role: "button"}} />',
      errors: [{ message: errorMessage('button', 'button') }],
    },
  ],
});
