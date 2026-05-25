import { RuleTester } from '../rule-tester';

const componentsSettings = {
  'jsx-a11y': {
    components: {
      FooComponent: 'iframe',
    },
  },
};

const errorMessage = '<iframe> elements must have a unique title property.';

new RuleTester().run('iframe-has-title', null as never, {
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
    { code: '<iframe />', errors: [{ message: errorMessage }] },
    { code: '<iframe {...props} />', errors: [{ message: errorMessage }] },
    {
      code: '<iframe title={undefined} />',
      errors: [{ message: errorMessage }],
    },
    { code: '<iframe title="" />', errors: [{ message: errorMessage }] },
    { code: '<iframe title={false} />', errors: [{ message: errorMessage }] },
    { code: '<iframe title={true} />', errors: [{ message: errorMessage }] },
    { code: "<iframe title={''} />", errors: [{ message: errorMessage }] },
    { code: '<iframe title={``} />', errors: [{ message: errorMessage }] },
    { code: '<iframe title={""} />', errors: [{ message: errorMessage }] },
    { code: '<iframe title={42} />', errors: [{ message: errorMessage }] },
    {
      code: '<FooComponent />',
      settings: componentsSettings,
      errors: [{ message: errorMessage }],
    },
  ],
});
