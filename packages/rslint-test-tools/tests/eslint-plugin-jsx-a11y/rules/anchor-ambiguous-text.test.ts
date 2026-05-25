import { RuleTester } from '../rule-tester';

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

const DEFAULT_AMBIGUOUS_WORDS = [
  'click here',
  'here',
  'link',
  'a link',
  'learn more',
];

const expectedMessage = (words: string[]) =>
  `Ambiguous text within anchor. Screen reader users rely on link text for context; the words "${words.join('", "')}" are ambiguous and do not provide enough context.`;

const defaultErrorMessage = expectedMessage(DEFAULT_AMBIGUOUS_WORDS);

new RuleTester().run('anchor-ambiguous-text', null as never, {
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
    { code: '<a>here</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>HERE</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>click here</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>learn more</a>;', errors: [{ message: defaultErrorMessage }] },
    {
      code: '<a>learn      more</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    { code: '<a>learn more.</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>learn more?</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>learn more,</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>learn more!</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>learn more;</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>learn more:</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>link</a>;', errors: [{ message: defaultErrorMessage }] },
    { code: '<a>a link</a>;', errors: [{ message: defaultErrorMessage }] },
    {
      code: '<a aria-label="click here">something</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    { code: '<a> a link </a>;', errors: [{ message: defaultErrorMessage }] },
    {
      code: '<a>a<i></i> link</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><i></i>a link</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><span>click</span> here</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><span> click </span> here</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><span aria-hidden>more text</span>learn more</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><span aria-hidden="true">more text</span>learn more</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><img alt="click here"/></a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a alt="tutorial on using eslint-plugin-jsx-a11y">click here</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><span alt="tutorial on using eslint-plugin-jsx-a11y">click here</span></a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><CustomElement>click</CustomElement> here</a>;',
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<Link>here</Link>',
      settings: linkComponentSettings,
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a><Image alt="click here" /></a>',
      settings: imageComponentSettings,
      errors: [{ message: defaultErrorMessage }],
    },
    {
      code: '<a>a disallowed word</a>',
      options: [{ words: ['a disallowed word'] }],
      errors: [{ message: expectedMessage(['a disallowed word']) }],
    },
  ],
});
