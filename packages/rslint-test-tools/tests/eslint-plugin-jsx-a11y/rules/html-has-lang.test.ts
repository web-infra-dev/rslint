import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const htmlTopSettings = {
  'jsx-a11y': {
    components: {
      HTMLTop: 'html',
    },
  },
};

new RuleTester(config).run('html-has-lang', {
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
      errors: [
        {
          messageId: 'htmlHasLang',
          message: '<html> elements must have the lang prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<html {...props} />',
      errors: [
        {
          messageId: 'htmlHasLang',
          message: '<html> elements must have the lang prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<html lang={undefined} />',
      errors: [
        {
          messageId: 'htmlHasLang',
          message: '<html> elements must have the lang prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<HTMLTop />',
      settings: htmlTopSettings,
      errors: [
        {
          messageId: 'htmlHasLang',
          message: '<html> elements must have the lang prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
  ],
});
