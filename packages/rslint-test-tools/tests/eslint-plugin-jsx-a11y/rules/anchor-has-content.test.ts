import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const linkSettings = {
  'jsx-a11y': {
    components: {
      Link: 'a',
    },
  },
};

new RuleTester(config).run('anchor-has-content', {
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
      errors: [
        {
          messageId: 'anchorHasContent',
          message:
            'Anchors must have content and the content must be accessible by a screen reader.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 6,
        },
      ],
    },
    {
      code: '<a><Bar aria-hidden /></a>',
      errors: [
        {
          messageId: 'anchorHasContent',
          message:
            'Anchors must have content and the content must be accessible by a screen reader.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>{undefined}</a>',
      errors: [
        {
          messageId: 'anchorHasContent',
          message:
            'Anchors must have content and the content must be accessible by a screen reader.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<Link />',
      settings: linkSettings,
      errors: [
        {
          messageId: 'anchorHasContent',
          message:
            'Anchors must have content and the content must be accessible by a screen reader.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
  ],
});
