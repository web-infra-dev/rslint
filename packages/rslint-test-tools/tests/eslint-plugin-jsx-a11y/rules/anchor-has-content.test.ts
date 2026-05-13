import { RuleTester } from '../rule-tester';

const linkSettings = {
  'jsx-a11y': {
    components: {
      Link: 'a',
    },
  },
};

const errorMessage =
  'Anchors must have content and the content must be accessible by a screen reader.';

new RuleTester().run('anchor-has-content', null as never, {
  valid: [
    { code: '<div />;' },
    { code: '<a>Foo</a>' },
    { code: '<a><Bar /></a>' },
    { code: '<a>{foo}</a>' },
    { code: '<a>{foo.bar}</a>' },
    { code: '<a dangerouslySetInnerHTML={{ __html: "foo" }} />' },
    { code: '<a children={children} />' },
    { code: '<Link />' },
    { code: '<Link>foo</Link>', settings: linkSettings },
    { code: '<a title={title} />' },
    { code: '<a aria-label={ariaLabel} />' },
    { code: '<a title={title} aria-label={ariaLabel} />' },
  ],
  invalid: [
    {
      code: '<a />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<a><Bar aria-hidden /></a>',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<a>{undefined}</a>',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<Link />',
      settings: linkSettings,
      errors: [{ message: errorMessage }],
    },
  ],
});
