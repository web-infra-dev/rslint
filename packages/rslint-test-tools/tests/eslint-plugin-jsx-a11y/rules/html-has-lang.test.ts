import { RuleTester } from '../rule-tester';

const htmlTopSettings = {
  'jsx-a11y': {
    components: {
      HTMLTop: 'html',
    },
  },
};

const errorMessage = '<html> elements must have the lang prop.';

new RuleTester().run('html-has-lang', null as never, {
  valid: [
    { code: '<div />;' },
    { code: '<html lang="en" />' },
    { code: '<html lang="en-US" />' },
    { code: '<html lang={foo} />' },
    { code: '<html lang />' },
    { code: '<HTML />' },
    { code: '<HTMLTop lang="en" />', settings: htmlTopSettings },
  ],
  invalid: [
    {
      code: '<html />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<html {...props} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<html lang={undefined} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<HTMLTop />',
      settings: htmlTopSettings,
      errors: [{ message: errorMessage }],
    },
  ],
});
