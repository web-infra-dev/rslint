import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

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

new RuleTester(config).run('no-redundant-roles', {
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
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element body has an implicit role of document. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<button role="button" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element button has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<Button role="button" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element button has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<select role="combobox"><option>1</option><option>2</option></select>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<select role="combobox" size="" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<select role="combobox" size={1} />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<select role="combobox" size="1" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<select role="combobox" size={null}></select>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<select role="combobox" size={undefined}></select>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<select role="combobox" multiple={undefined}></select>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: '<select role="combobox" multiple={false}></select>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<select role="combobox" multiple=""></select>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of combobox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<select role="listbox" size="3" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of listbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<select role="listbox" size={2} />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of listbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<select role="listbox" multiple><option>1</option><option>2</option></select>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of listbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: '<select role="listbox" multiple={true}></select>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element select has an implicit role of listbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },

    // ============================================================
    // nav allowance disabled
    // ============================================================
    {
      code: '<nav role="navigation" />',
      options: [{ nav: [] }],
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element nav has an implicit role of navigation. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },

    // ============================================================
    // img/ul/ol without override
    // ============================================================
    {
      code: '<ul role="list" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element ul has an implicit role of list. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<ol role="list" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element ol has an implicit role of list. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<img role="img" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element img has an implicit role of img. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<img src={someVariable} role="img" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element img has an implicit role of img. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },

    // ============================================================
    // implicitRoleForImg — alt boolean / undefined / non-literal
    // ============================================================
    {
      code: '<img alt role="img" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element img has an implicit role of img. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<img alt={undefined} role="img" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element img has an implicit role of img. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },

    // ============================================================
    // input type → ARIA role mapping
    // ============================================================
    {
      code: '<input type="button" role="button" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<input type="submit" role="button" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: '<input type="reset" role="button" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<input type="image" role="button" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<input type="checkbox" role="checkbox" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of checkbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<input type="radio" role="radio" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of radio. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<input type="range" role="slider" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of slider. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<input role="textbox" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of textbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<input type="text" role="textbox" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of textbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<input type="hidden" role="textbox" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element input has an implicit role of textbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },

    // ============================================================
    // Case-insensitive role value
    // ============================================================
    {
      code: '<button role="BUTTON" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element button has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },

    // ============================================================
    // Implicit-role table coverage
    // ============================================================
    {
      code: '<article role="article" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element article has an implicit role of article. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<aside role="complementary" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element aside has an implicit role of complementary. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<datalist role="listbox" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element datalist has an implicit role of listbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<details role="group" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element details has an implicit role of group. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<dialog role="dialog" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element dialog has an implicit role of dialog. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<form role="form" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element form has an implicit role of form. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<h1 role="heading" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element h1 has an implicit role of heading. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<h6 role="heading" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element h6 has an implicit role of heading. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<hr role="separator" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element hr has an implicit role of separator. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<li role="listitem" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element li has an implicit role of listitem. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<meter role="progressbar" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element meter has an implicit role of progressbar. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<progress role="progressbar" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element progress has an implicit role of progressbar. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<option role="option" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element option has an implicit role of option. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<output role="status" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element output has an implicit role of status. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<section role="region" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element section has an implicit role of region. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<tbody role="rowgroup" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element tbody has an implicit role of rowgroup. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<textarea role="textbox" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element textarea has an implicit role of textbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },

    // ============================================================
    // a/area/link href → link role
    // ============================================================
    {
      code: '<a href="/x" role="link" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element a has an implicit role of link. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<area href="#x" role="link" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element area has an implicit role of link. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<link href="/x" role="link" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element link has an implicit role of link. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },

    // ============================================================
    // menu / menuitem type branches
    // ============================================================
    {
      code: '<menu type="toolbar" role="toolbar" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element menu has an implicit role of toolbar. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<menuitem type="command" role="menuitem" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element menuitem has an implicit role of menuitem. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: '<menuitem type="checkbox" role="menuitemcheckbox" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element menuitem has an implicit role of menuitemcheckbox. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<menuitem type="radio" role="menuitemradio" />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element menuitem has an implicit role of menuitemradio. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },

    // ============================================================
    // polymorphicPropName remap to redundant target
    // ============================================================
    {
      code: '<Foo as="button" role="button" />',
      settings: polymorphicSettings,
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element button has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },

    // ============================================================
    // Paired form — fires on opening tag
    // ============================================================
    {
      code: '<button role="button">Click</button>',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element button has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },

    // ============================================================
    // Literal spread role
    // ============================================================
    {
      code: '<button {...{role: "button"}} />',
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element button has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
  ],
});

// Additional upstream v6.10.2 inputs. Matrix rows freeze the expected
// classification and configuration; ranges follow the authored source templates.
// No rule implementation or lint output is consulted when generating cases.
const upstreamOptions0 = [{ nav: [] }];
const upstreamSettings0 = { 'jsx-a11y': { components: { Button: 'button' } } };

new RuleTester(config).run('no-redundant-roles', {
  valid: [
    {
      name: 'upstream v6.10.2 jsx-a11y/no-redundant-roles:recommended valid 0',
      code: '<div />;',
      options: upstreamOptions0,
    },
    ...(
      [
        ['main', '', 1, 'button'],
        ['button', '', 2, 'MyComponent'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-redundant-roles:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-redundant-roles:recommended valid 3',
      code: '<button role={`${foo}button`} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-redundant-roles:recommended valid 4',
      code: '<Button role={`${foo}button`} />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
    },
  ],
  invalid: [
    {
      name: 'upstream v6.10.2 jsx-a11y/no-redundant-roles:recommended invalid 0',
      code: '<button role="button" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element button has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-redundant-roles:recommended invalid 1',
      code: '<body role="DOCUMENT" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element body has an implicit role of document. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-redundant-roles:recommended invalid 2',
      code: '<Button role="button" />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
      errors: [
        {
          messageId: 'noRedundantRoles',
          message:
            'The element button has an implicit role of button. Defining this explicitly is redundant and should be avoided.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
  ],
});
