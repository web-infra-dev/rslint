import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';
import { singleLineError } from '../matrix';

const componentsSettings = {
  'jsx-a11y': {
    components: {
      Button: 'button',
      Link: 'a',
    },
  },
};

// Mirrors `configs.recommended.rules['jsx-a11y/no-interactive-element-to-noninteractive-role'][1]`.
const recommendedOptions = [
  {
    tr: ['none', 'presentation'],
    canvas: ['img'],
  },
];

new RuleTester(config).run('no-interactive-element-to-noninteractive-role', {
  valid: [
    // ============================================================
    // Upstream alwaysValid (run under both :recommended and :strict)
    // ============================================================
    // Custom JSX components — element type isn't in aria-query's dom map.
    { code: '<TestComponent onClick={doFoo} />' },
    { code: '<Button onClick={doFoo} />' },

    // Interactive elements with an interactive role.
    { code: '<a href="http://x.y.z" role="button" />' },
    { code: '<a href="http://x.y.z" tabIndex="0" role="button" />' },
    { code: '<button className="foo" role="button" />' },

    // All flavors of input with role="button".
    { code: '<input role="button" />' },
    { code: '<input type="button" role="button" />' },
    { code: '<input type="checkbox" role="button" />' },
    { code: '<input type="color" role="button" />' },
    { code: '<input type="date" role="button" />' },
    { code: '<input type="datetime" role="button" />' },
    { code: '<input type="datetime-local" role="button" />' },
    { code: '<input type="email" role="button" />' },
    { code: '<input type="file" role="button" />' },
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
    { code: '<input type="hidden" role="button" />' },

    // Other inherently interactive controls.
    { code: '<menuitem role="button" />;' },
    { code: '<option className="foo" role="button" />' },
    { code: '<select className="foo" role="button" />' },
    { code: '<textarea className="foo" role="button" />' },
    { code: '<tr role="button" />;' },

    // HTML elements with neither interactive nor non-interactive valence.
    { code: '<a role="button" />' },
    { code: '<a role="img" />;' },
    { code: '<a tabIndex="0" role="button" />' },
    { code: '<a tabIndex="0" role="img" />' },
    { code: '<div role="button" />;' },
    { code: '<div role={undefined} role="button" />;' },
    { code: '<div role="img" />;' },
    { code: '<canvas role="button" />;' },

    // Non-DOM custom components are exempt.
    { code: '<MyButton role="img" />' },

    // Namespaced roles — propName isn't literally `role`.
    { code: '<div mynamespace:role="term" />' },
    { code: '<input mynamespace:role="img" />' },

    // Components with explicit aria mapping.
    { code: '<Link href="http://x.y.z" role="img" />' },
    { code: '<Link href="http://x.y.z" />', settings: componentsSettings },
    { code: '<Button onClick={doFoo} />', settings: componentsSettings },

    // ============================================================
    // Recommended-only valid (tr / canvas allow-list, non-DOM component)
    // ============================================================
    {
      code: '<tr role="presentation" />;',
      options: recommendedOptions,
    },
    {
      code: '<canvas role="img" />;',
      options: recommendedOptions,
    },
    {
      code: '<Component role="presentation" />;',
      options: recommendedOptions,
    },
  ],
  invalid: [
    // ============================================================
    // Upstream neverValid — invalid in both recommended and strict
    // ============================================================
    // Interactive elements + non-interactive role.
    {
      code: '<a href="http://x.y.z" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<a href="http://x.y.z" tabIndex="0" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },

    // All flavors of input + role="img".
    {
      code: '<input role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<input type="img" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<input type="checkbox" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<input type="color" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<input type="date" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<input type="datetime" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<input type="datetime-local" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: '<input type="email" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<input type="file" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<input type="hidden" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<input type="image" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<input type="month" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<input type="number" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<input type="password" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<input type="radio" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<input type="range" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<input type="reset" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<input type="search" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<input type="submit" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<input type="tel" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<input type="text" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<input type="time" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<input type="url" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<input type="week" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<menuitem role="img" />;',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<option className="foo" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<select className="foo" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: '<textarea className="foo" role="img" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<tr role="img" />;',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },

    // Interactive elements + role="listitem".
    {
      code: '<a href="http://x.y.z" role="listitem" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<a href="http://x.y.z" tabIndex="0" role="listitem" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<input role="listitem" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<input type="listitem" role="listitem" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<input type="checkbox" role="listitem" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<menuitem role="listitem" />;',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<tr role="listitem" />;',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },

    // Custom element resolved to <a> via settings.
    {
      code: '<Link href="http://x.y.z" role="img" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },

    // ============================================================
    // Strict-only invalid (no allowed-roles override for tr / canvas)
    // ============================================================
    {
      code: '<tr role="presentation" />;',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<canvas role="img" />;',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
  ],
});

// Additional upstream v6.10.2 inputs. Matrix rows freeze the expected
// classification and configuration; ranges follow the authored source templates.
// No rule implementation or lint output is consulted when generating cases.
const upstreamOptions0 = [{ tr: ['none', 'presentation'], canvas: ['img'] }];
const upstreamSettings0 = {
  'jsx-a11y': { components: { Button: 'button', Link: 'a' } },
};

new RuleTester(config).run('no-interactive-element-to-noninteractive-role', {
  valid: [
    ...(
      [
        [
          'button',
          ';',
          38,
          'acronym address applet aside audio b base bdi bdo big blink blockquote body br',
        ],
        [
          'button',
          ';',
          53,
          'caption center cite code col colgroup content data datalist del details dir',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict valid 66',
      code: '<div className="foo" role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict valid 67',
      code: '<div className="foo" {...props} role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict valid 68',
      code: '<div aria-hidden role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict valid 69',
      code: '<div aria-hidden={true} role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict valid 72',
      code: '<div {...props} role="button" />;',
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict valid 73',
      code: '<div onKeyUp={() => void 0} aria-hidden={false} role="button" />;',
    },
    ...(
      [
        [
          'button',
          ';',
          74,
          'dl em embed figcaption font footer frameset head header hgroup html i iframe ins kbd keygen label legend link map mark marquee menu meta meter noembed noscript object optgroup output p param picture pre progress q rp rt rtc ruby s samp script section small source spacer span strike strong style sub summary sup th time title track tt u var video wbr xmp',
        ],
        ['checkbox', ';', 139, 'div'],
        ['columnheader', ';', 140, 'div'],
        ['combobox', ';', 141, 'div'],
        ['grid', ';', 142, 'div'],
        ['gridcell', ';', 143, 'div'],
        ['link', ';', 144, 'div'],
        ['listbox', ';', 145, 'div'],
        ['menu', ';', 146, 'div'],
        ['menubar', ';', 147, 'div'],
        ['menuitem', ';', 148, 'div'],
        ['menuitemcheckbox', ';', 149, 'div'],
        ['menuitemradio', ';', 150, 'div'],
        ['option', ';', 151, 'div'],
        ['progressbar', ';', 152, 'div'],
        ['radio', ';', 153, 'div'],
        ['radiogroup', ';', 154, 'div'],
        ['row', ';', 155, 'div'],
        ['rowheader', ';', 156, 'div'],
        ['searchbox', ';', 157, 'div'],
        ['slider', ';', 158, 'div'],
        ['spinbutton', ';', 159, 'div'],
        ['switch', ';', 160, 'div'],
        ['tab', ';', 161, 'div'],
        ['textbox', ';', 162, 'div'],
        ['treeitem', ';', 163, 'div'],
        ['presentation', ';', 164, 'div'],
        ['command', ';', 165, 'div'],
        ['composite', ';', 166, 'div'],
        ['input', ';', 167, 'div'],
        ['landmark', ';', 168, 'div'],
        ['range', ';', 169, 'div'],
        ['roletype', ';', 170, 'div'],
        ['section', ';', 171, 'div'],
        ['sectionhead', ';', 172, 'div'],
        ['select', ';', 173, 'div'],
        ['structure', ';', 174, 'div'],
        ['tablist', ';', 175, 'div'],
        ['toolbar', ';', 176, 'div'],
        ['tree', ';', 177, 'div'],
        ['treegrid', ';', 178, 'div'],
        ['widget', ';', 179, 'div'],
        ['window', ';', 180, 'div'],
        ['button', ';', 181, 'main area article'],
        [
          'button',
          ';',
          185,
          'dd dfn dt fieldset figure form frame h1 h2 h3 h4 h5 h6 hr img li',
        ],
        ['presentation', ';', 201, 'li'],
        ['button', ';', 202, 'nav ol table tbody td tfoot thead ul'],
        ['alert', ';', 210, 'div'],
        ['alertdialog', ';', 211, 'div'],
        ['application', ';', 212, 'div'],
        ['article', ';', 213, 'div'],
        ['banner', ';', 214, 'div'],
        ['cell', ';', 215, 'div'],
        ['complementary', ';', 216, 'div'],
        ['contentinfo', ';', 217, 'div'],
        ['definition', ';', 218, 'div'],
        ['dialog', ';', 219, 'div'],
        ['directory', ';', 220, 'div'],
        ['document', ';', 221, 'div'],
        ['feed', ';', 222, 'div'],
        ['figure', ';', 223, 'div'],
        ['form', ';', 224, 'div'],
        ['group', ';', 225, 'div'],
        ['heading', ';', 226, 'div'],
        ['list', ';', 228, 'div'],
        ['listitem', ';', 229, 'div'],
        ['log', ';', 230, 'div'],
        ['main', ';', 231, 'div'],
        ['marquee', ';', 232, 'div'],
        ['math', ';', 233, 'div'],
        ['navigation', ';', 234, 'div'],
        ['note', ';', 235, 'div'],
        ['region', ';', 236, 'div'],
        ['rowgroup', ';', 237, 'div'],
        ['search', ';', 238, 'div'],
        ['separator', ';', 239, 'div'],
        ['scrollbar', ';', 240, 'div'],
        ['status', ';', 241, 'div'],
        ['table', ';', 242, 'div'],
        ['tabpanel', ';', 243, 'div'],
        ['term', ';', 244, 'div'],
        ['timer', ';', 245, 'div'],
        ['tooltip', ';', 246, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict valid ${firstCase + index}`,
          code,
        };
      }),
    ),
    ...(
      [
        [
          'button',
          ';',
          38,
          'acronym address applet aside audio b base bdi bdo big blink blockquote body br',
        ],
        [
          'button',
          ';',
          53,
          'caption center cite code col colgroup content data datalist del details dir',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 66',
      code: '<div className="foo" role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 67',
      code: '<div className="foo" {...props} role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 68',
      code: '<div aria-hidden role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 69',
      code: '<div aria-hidden={true} role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 72',
      code: '<div {...props} role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 73',
      code: '<div onKeyUp={() => void 0} aria-hidden={false} role="button" />;',
      options: upstreamOptions0,
    },
    ...(
      [
        [
          'button',
          ';',
          74,
          'dl em embed figcaption font footer frameset head header hgroup html i iframe ins kbd keygen label legend link map mark marquee menu meta meter noembed noscript object optgroup output p param picture pre progress q rp rt rtc ruby s samp script section small source spacer span strike strong style sub summary sup th time title track tt u var video wbr xmp',
        ],
        ['checkbox', ';', 139, 'div'],
        ['columnheader', ';', 140, 'div'],
        ['combobox', ';', 141, 'div'],
        ['grid', ';', 142, 'div'],
        ['gridcell', ';', 143, 'div'],
        ['link', ';', 144, 'div'],
        ['listbox', ';', 145, 'div'],
        ['menu', ';', 146, 'div'],
        ['menubar', ';', 147, 'div'],
        ['menuitem', ';', 148, 'div'],
        ['menuitemcheckbox', ';', 149, 'div'],
        ['menuitemradio', ';', 150, 'div'],
        ['option', ';', 151, 'div'],
        ['progressbar', ';', 152, 'div'],
        ['radio', ';', 153, 'div'],
        ['radiogroup', ';', 154, 'div'],
        ['row', ';', 155, 'div'],
        ['rowheader', ';', 156, 'div'],
        ['searchbox', ';', 157, 'div'],
        ['slider', ';', 158, 'div'],
        ['spinbutton', ';', 159, 'div'],
        ['switch', ';', 160, 'div'],
        ['tab', ';', 161, 'div'],
        ['textbox', ';', 162, 'div'],
        ['treeitem', ';', 163, 'div'],
        ['presentation', ';', 164, 'div'],
        ['command', ';', 165, 'div'],
        ['composite', ';', 166, 'div'],
        ['input', ';', 167, 'div'],
        ['landmark', ';', 168, 'div'],
        ['range', ';', 169, 'div'],
        ['roletype', ';', 170, 'div'],
        ['section', ';', 171, 'div'],
        ['sectionhead', ';', 172, 'div'],
        ['select', ';', 173, 'div'],
        ['structure', ';', 174, 'div'],
        ['tablist', ';', 175, 'div'],
        ['toolbar', ';', 176, 'div'],
        ['tree', ';', 177, 'div'],
        ['treegrid', ';', 178, 'div'],
        ['widget', ';', 179, 'div'],
        ['window', ';', 180, 'div'],
        ['button', ';', 181, 'main area article'],
        [
          'button',
          ';',
          185,
          'dd dfn dt fieldset figure form frame h1 h2 h3 h4 h5 h6 hr img li',
        ],
        ['presentation', ';', 201, 'li'],
        ['button', ';', 202, 'nav ol table tbody td tfoot thead ul'],
        ['alert', ';', 210, 'div'],
        ['alertdialog', ';', 211, 'div'],
        ['application', ';', 212, 'div'],
        ['article', ';', 213, 'div'],
        ['banner', ';', 214, 'div'],
        ['cell', ';', 215, 'div'],
        ['complementary', ';', 216, 'div'],
        ['contentinfo', ';', 217, 'div'],
        ['definition', ';', 218, 'div'],
        ['dialog', ';', 219, 'div'],
        ['directory', ';', 220, 'div'],
        ['document', ';', 221, 'div'],
        ['feed', ';', 222, 'div'],
        ['figure', ';', 223, 'div'],
        ['form', ';', 224, 'div'],
        ['group', ';', 225, 'div'],
        ['heading', ';', 226, 'div'],
        ['list', ';', 228, 'div'],
        ['listitem', ';', 229, 'div'],
        ['log', ';', 230, 'div'],
        ['main', ';', 231, 'div'],
        ['marquee', ';', 232, 'div'],
        ['math', ';', 233, 'div'],
        ['navigation', ';', 234, 'div'],
        ['note', ';', 235, 'div'],
        ['region', ';', 236, 'div'],
        ['rowgroup', ';', 237, 'div'],
        ['search', ';', 238, 'div'],
        ['separator', ';', 239, 'div'],
        ['scrollbar', ';', 240, 'div'],
        ['status', ';', 241, 'div'],
        ['table', ';', 242, 'div'],
        ['tabpanel', ';', 243, 'div'],
        ['term', ';', 244, 'div'],
        ['timer', ';', 245, 'div'],
        ['tooltip', ';', 246, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 0',
      code: '<TestComponent onClick={doFoo} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 1',
      code: '<Button onClick={doFoo} />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 2',
      code: '<a href="http://x.y.z" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 3',
      code: '<a href="http://x.y.z" tabIndex="0" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 4',
      code: '<button className="foo" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 5',
      code: '<input role="button" />',
      options: upstreamOptions0,
    },
    ...(
      [
        [
          'button',
          '',
          6,
          'button checkbox color date datetime datetime-local email file image month number password radio range reset search submit tel text time url week hidden',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 29',
      code: '<menuitem role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 30',
      code: '<option className="foo" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 31',
      code: '<select className="foo" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 32',
      code: '<textarea className="foo" role="button" />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['button', ';', 33, 'tr'],
        ['button', '', 34, 'a'],
        ['img', ';', 35, 'a'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 36',
      code: '<a tabIndex="0" role="button" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 37',
      code: '<a tabIndex="0" role="img" />',
      options: upstreamOptions0,
    },
    ...(
      [
        ['button', ';', 52, 'canvas'],
        ['button', ';', 65, 'div'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 71',
      code: '<div role={undefined} role="button" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 227',
      code: '<div role="img" />;',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 247',
      code: '<div mynamespace:role="term" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 248',
      code: '<input mynamespace:role="img" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 249',
      code: '<Link href="http://x.y.z" role="img" />',
      options: upstreamOptions0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 250',
      code: '<Link href="http://x.y.z" />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended valid 251',
      code: '<Button onClick={doFoo} />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
    },
  ],
  invalid: [
    ...(
      [
        [
          'listitem',
          '',
          36,
          'color date datetime datetime-local email file image month number password radio range reset search submit tel text time url week',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict invalid ${firstCase + index}`,
          code,
          errors: [
            singleLineError(
              'noInteractiveElementToNoninteractiveRole',
              'Interactive elements should not be assigned non-interactive roles.',
              `role="${role}"`,
              `<input type="${inputType}" `.length + 1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict invalid 57',
      code: '<option className="foo" role="listitem" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict invalid 58',
      code: '<select className="foo" role="listitem" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:strict invalid 59',
      code: '<textarea className="foo" role="listitem" />',
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    ...(
      [
        [
          'listitem',
          '',
          36,
          'color date datetime datetime-local email file image month number password radio range reset search submit tel text time url week',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noInteractiveElementToNoninteractiveRole',
              'Interactive elements should not be assigned non-interactive roles.',
              `role="${role}"`,
              `<input type="${inputType}" `.length + 1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 57',
      code: '<option className="foo" role="listitem" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 58',
      code: '<select className="foo" role="listitem" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 59',
      code: '<textarea className="foo" role="listitem" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 0',
      code: '<a href="http://x.y.z" role="img" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 1',
      code: '<a href="http://x.y.z" tabIndex="0" role="img" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 2',
      code: '<input role="img" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    ...(
      [
        [
          'img',
          '',
          3,
          'img checkbox color date datetime datetime-local email file hidden image month number password radio range reset search submit tel text time url week',
        ],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((inputType, index) => {
        const openingElement = `<input type="${inputType}" role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noInteractiveElementToNoninteractiveRole',
              'Interactive elements should not be assigned non-interactive roles.',
              `role="${role}"`,
              `<input type="${inputType}" `.length + 1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 26',
      code: '<menuitem role="img" />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 27',
      code: '<option className="foo" role="img" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 28',
      code: '<select className="foo" role="img" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 29',
      code: '<textarea className="foo" role="img" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 30',
      code: '<tr role="img" />;',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 31',
      code: '<a href="http://x.y.z" role="listitem" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 24,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 32',
      code: '<a href="http://x.y.z" tabIndex="0" role="listitem" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 33',
      code: '<input role="listitem" />',
      options: upstreamOptions0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    ...([['listitem', '', 34, 'listitem checkbox']] as const).flatMap(
      ([role, suffix, firstCase, values]) =>
        values.split(' ').map((inputType, index) => {
          const openingElement = `<input type="${inputType}" role="${role}" />`;
          const code = openingElement + suffix;
          return {
            name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid ${firstCase + index}`,
            code,
            options: upstreamOptions0,
            errors: [
              singleLineError(
                'noInteractiveElementToNoninteractiveRole',
                'Interactive elements should not be assigned non-interactive roles.',
                `role="${role}"`,
                `<input type="${inputType}" `.length + 1,
              ),
            ],
          };
        }),
    ),
    ...(
      [
        ['listitem', ';', 56, 'menuitem'],
        ['listitem', ';', 60, 'tr'],
      ] as const
    ).flatMap(([role, suffix, firstCase, values]) =>
      values.split(' ').map((tag, index) => {
        const openingElement = `<${tag} role="${role}" />`;
        const code = openingElement + suffix;
        return {
          name: `upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid ${firstCase + index}`,
          code,
          options: upstreamOptions0,
          errors: [
            singleLineError(
              'noInteractiveElementToNoninteractiveRole',
              'Interactive elements should not be assigned non-interactive roles.',
              `role="${role}"`,
              `<${tag} `.length + 1,
            ),
          ],
        };
      }),
    ),
    {
      name: 'upstream v6.10.2 jsx-a11y/no-interactive-element-to-noninteractive-role:recommended invalid 61',
      code: '<Link href="http://x.y.z" role="img" />',
      options: upstreamOptions0,
      settings: upstreamSettings0,
      errors: [
        {
          messageId: 'noInteractiveElementToNoninteractiveRole',
          message:
            'Interactive elements should not be assigned non-interactive roles.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
  ],
});
