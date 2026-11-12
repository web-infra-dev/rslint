import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

const componentsSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
    components: {
      Foo: 'html',
    },
  },
};

new RuleTester(config).run('lang', {
  valid: [
    { code: '<div />;' },
    { code: '<div foo="bar" />;' },
    { code: '<div lang="foo" />;' },
    { code: '<html lang="en" />' },
    { code: '<html lang="en-US" />' },
    { code: '<html lang="zh-Hans" />' },
    { code: '<html lang="zh-Hant-HK" />' },
    { code: '<html lang="zh-yue-Hant" />' },
    { code: '<html lang="ja-Latn" />' },
    { code: '<html lang={foo} />' },
    { code: '<HTML lang="foo" />' },
    { code: '<Foo lang={undefined} />' },
    { code: '<Foo lang="en" />', settings: componentsSettings },
    { code: '<Box as="html" lang="en"  />', settings: componentsSettings },
  ],
  invalid: [
    {
      code: '<html lang="foo" />',
      errors: [
        {
          messageId: 'invalidLangValue',
          message: 'lang attribute must have a valid value.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 17,
        },
      ],
    },
    {
      code: '<html lang="zz-LL" />',
      errors: [
        {
          messageId: 'invalidLangValue',
          message: 'lang attribute must have a valid value.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<html lang={undefined} />',
      errors: [
        {
          messageId: 'invalidLangValue',
          message: 'lang attribute must have a valid value.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<Foo lang={undefined} />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'invalidLangValue',
          message: 'lang attribute must have a valid value.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<Box as="html" lang="foo" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'invalidLangValue',
          message: 'lang attribute must have a valid value.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
  ],
});
