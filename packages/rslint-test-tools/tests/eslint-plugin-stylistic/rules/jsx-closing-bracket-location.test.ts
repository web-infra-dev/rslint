import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const MESSAGE_AFTER_PROPS = 'placed after the last prop';
const MESSAGE_AFTER_TAG = 'placed after the opening tag';
const MESSAGE_PROPS_ALIGNED = 'aligned with the last prop';
const MESSAGE_TAG_ALIGNED = 'aligned with the opening tag';
const MESSAGE_LINE_ALIGNED = 'aligned with the line containing the opening tag';

function details(expectedColumn: number, expectedNextLine: boolean) {
  return ` (expected column ${expectedColumn}${expectedNextLine ? ' on the next line)' : ')'}`;
}

ruleTester.run('jsx-closing-bracket-location', null as never, {
  valid: [
    // default tag-aligned
    { code: `<App />` },
    { code: `<App foo />` },
    { code: `<App\n  foo\n/>` },
    { code: `<App\n  foo\n  // comment\n/>` },
    { code: `<App\n  {...foo}\n/>` },
    // string shortcut
    { code: `<App\n  foo />`, options: ['after-props'] },
    { code: `<App\n  foo\n  />`, options: ['props-aligned'] },
    // object location form
    { code: `<App\n  foo />`, options: [{ location: 'after-props' }] },
    { code: `<App\n  foo\n/>`, options: [{ location: 'tag-aligned' }] },
    { code: `<App\n  foo\n/>`, options: [{ location: 'line-aligned' }] },
    { code: `<App\n  foo\n  />`, options: [{ location: 'props-aligned' }] },
    // non-self-closing
    { code: `<App foo></App>` },
    { code: `<App\n  foo\n></App>`, options: [{ location: 'tag-aligned' }] },
    // selfClosing / nonEmpty per-form
    {
      code: `<Provider store>\n  <App\n    foo />\n</Provider>`,
      options: [{ selfClosing: 'after-props' }],
    },
    {
      code: `<Provider\n  store>\n  <App\n    foo\n  />\n</Provider>`,
      options: [{ nonEmpty: 'after-props' }],
    },
    // nonEmpty:false / selfClosing:false disable per-form
    {
      code: `<App>\n  <Foo\n    bar\n  >\n  </Foo>\n  <Foo\n    bar />\n</App>`,
      options: [{ nonEmpty: false, selfClosing: 'after-props' }],
    },
  ],

  invalid: [
    // after-tag default, fix collapses to single-line
    {
      code: `<App\n/>`,
      output: `<App />`,
      errors: [
        {
          messageId: 'bracketLocation',
          message: `The closing bracket must be ${MESSAGE_AFTER_TAG}`,
        },
      ],
    },
    // after-props default, prop on opening line
    {
      code: `<App foo\n/>`,
      output: `<App foo/>`,
      errors: [
        {
          messageId: 'bracketLocation',
          message: `The closing bracket must be ${MESSAGE_AFTER_PROPS}`,
        },
      ],
    },
    // tag-aligned message + details
    {
      code: `<App\n  foo />`,
      output: `<App\n  foo\n/>`,
      options: [{ location: 'tag-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message: `The closing bracket must be ${MESSAGE_TAG_ALIGNED}${details(1, true)}`,
        },
      ],
    },
    // props-aligned with very-far closing bracket
    {
      code: `<App\n  foo\n/>`,
      output: `<App\n  foo\n  />`,
      options: [{ location: 'props-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message: `The closing bracket must be ${MESSAGE_PROPS_ALIGNED}${details(3, false)}`,
        },
      ],
    },
    // line-aligned, paren-wrapped JSX
    {
      code: `var x = (\n  <div\n    className="X"\n    {...props} />\n)`,
      output: `var x = (\n  <div\n    className="X"\n    {...props}\n  />\n)`,
      options: [{ location: 'line-aligned' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message: `The closing bracket must be ${MESSAGE_LINE_ALIGNED}${details(3, true)}`,
        },
      ],
    },
    // trailing-comment upgrade after-props → line-aligned (stylistic-specific)
    {
      code: `<input\n  // comment\n  type="text"\n  // comment\n  />`,
      output: `<input\n  // comment\n  type="text"\n  // comment\n/>`,
      options: [{ location: 'after-props' }],
      errors: [
        {
          messageId: 'bracketLocation',
          message: `The closing bracket must be ${MESSAGE_LINE_ALIGNED}${details(1, false)}`,
        },
      ],
    },
  ],
});
