import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const componentsSettings = {
  'jsx-a11y': {
    components: {
      FooComponent: 'iframe',
    },
  },
};

new RuleTester(config).run('iframe-has-title', {
  valid: [
    { code: '<div />;' },
    { code: '<iframe title="Unique title" />' },
    { code: '<iframe title={foo} />' },
    { code: '<FooComponent />' },
    {
      code: '<FooComponent title="Unique title" />',
      settings: componentsSettings,
    },
  ],
  invalid: [
    {
      code: '<iframe />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<iframe {...props} />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<iframe title={undefined} />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<iframe title="" />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<iframe title={false} />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<iframe title={true} />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: "<iframe title={''} />",
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<iframe title={``} />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<iframe title={""} />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<iframe title={42} />',
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<FooComponent />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'iframeHasTitle',
          message: '<iframe> elements must have a unique title property.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
  ],
});
