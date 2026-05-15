import { RuleTester } from '../rule-tester';

const componentsSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
    components: {
      Foo: 'html',
    },
  },
};

const errorMessage = 'lang attribute must have a valid value.';

new RuleTester().run('lang', null as never, {
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
      errors: [{ message: errorMessage }],
    },
    {
      code: '<html lang="zz-LL" />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<html lang={undefined} />',
      errors: [{ message: errorMessage }],
    },
    {
      code: '<Foo lang={undefined} />',
      settings: componentsSettings,
      errors: [{ message: errorMessage }],
    },
    {
      code: '<Box as="html" lang="foo" />',
      settings: componentsSettings,
      errors: [{ message: errorMessage }],
    },
  ],
});
