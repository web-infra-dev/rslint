import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const linkComponentSettings = {
  'jsx-a11y': {
    components: {
      Link: 'a',
    },
  },
};

const imageComponentSettings = {
  'jsx-a11y': {
    components: {
      Image: 'img',
    },
  },
};

new RuleTester(config).run('anchor-ambiguous-text', {
  valid: [
    { code: '<a>documentation</a>;' },
    { code: '<a>${here}</a>;' },
    {
      code: '<a aria-label="tutorial on using eslint-plugin-jsx-a11y">click here</a>;',
    },
    {
      code: '<a><span aria-label="tutorial on using eslint-plugin-jsx-a11y">click here</span></a>;',
    },
    { code: '<a><img alt="documentation" /></a>;' },
    {
      code: '<a>click here</a>',
      options: [{ words: ['disabling the defaults'] }],
    },
    { code: '<Link>documentation</Link>;', settings: linkComponentSettings },
    {
      code: '<a><Image alt="documentation" /></a>;',
      settings: imageComponentSettings,
    },
    { code: '<Link>${here}</Link>;', settings: linkComponentSettings },
    {
      code: '<Link aria-label="tutorial on using eslint-plugin-jsx-a11y">click here</Link>;',
      settings: linkComponentSettings,
    },
    {
      code: '<Link>click here</Link>',
      options: [{ words: ['disabling the defaults with components'] }],
      settings: linkComponentSettings,
    },
  ],
  invalid: [
    {
      code: '<a>here</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>HERE</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>click here</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>learn more</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>learn      more</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>learn more.</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>learn more?</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>learn more,</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>learn more!</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>learn more;</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>learn more:</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>link</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>a link</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a aria-label="click here">something</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<a> a link </a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>a<i></i> link</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a><i></i>a link</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a><span>click</span> here</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a><span> click </span> here</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a><span aria-hidden>more text</span>learn more</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a><span aria-hidden="true">more text</span>learn more</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a><img alt="click here"/></a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a alt="tutorial on using eslint-plugin-jsx-a11y">click here</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '<a><span alt="tutorial on using eslint-plugin-jsx-a11y">click here</span></a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a><CustomElement>click</CustomElement> here</a>;',
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<Link>here</Link>',
      settings: linkComponentSettings,
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 7,
        },
      ],
    },
    {
      code: '<a><Image alt="click here" /></a>',
      settings: imageComponentSettings,
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "click here", "here", "link", "a link", "learn more" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '<a>a disallowed word</a>',
      options: [{ words: ['a disallowed word'] }],
      errors: [
        {
          messageId: 'anchorAmbiguousText',
          message:
            'Ambiguous text within anchor. Screen reader users rely on link text for context; the words "a disallowed word" are ambiguous and do not provide enough context.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
  ],
});
