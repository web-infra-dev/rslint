import { RuleTester } from '../rule-tester';

const components = [{ components: ['Heading', 'Title'] }];

const componentsSettings = {
  'jsx-a11y': {
    components: {
      CustomInput: 'input',
      Title: 'h1',
      Heading: 'h2',
    },
  },
};

const errorMessage =
  'Headings must have content and the content must be accessible by a screen reader.';

new RuleTester().run('heading-has-content', null as never, {
  valid: [
    // ---- DEFAULT ELEMENT TESTS ----
    { code: '<div />;' },
    { code: '<h1>Foo</h1>' },
    { code: '<h2>Foo</h2>' },
    { code: '<h3>Foo</h3>' },
    { code: '<h4>Foo</h4>' },
    { code: '<h5>Foo</h5>' },
    { code: '<h6>Foo</h6>' },
    { code: '<h6>123</h6>' },
    { code: '<h1><Bar /></h1>' },
    { code: '<h1>{foo}</h1>' },
    { code: '<h1>{foo.bar}</h1>' },
    { code: '<h1 dangerouslySetInnerHTML={{ __html: "foo" }} />' },
    { code: '<h1 children={children} />' },

    // ---- CUSTOM ELEMENT TESTS FOR COMPONENTS OPTION ----
    { code: '<Heading>Foo</Heading>', options: components },
    { code: '<Title>Foo</Title>', options: components },
    { code: '<Heading><Bar /></Heading>', options: components },
    { code: '<Heading>{foo}</Heading>', options: components },
    { code: '<Heading>{foo.bar}</Heading>', options: components },
    {
      code: '<Heading dangerouslySetInnerHTML={{ __html: "foo" }} />',
      options: components,
    },
    { code: '<Heading children={children} />', options: components },

    // ---- aria-hidden on the heading itself → exempt ----
    { code: '<h1 aria-hidden />' },

    // ---- CUSTOM ELEMENT TESTS FOR COMPONENTS SETTINGS ----
    { code: '<Heading>Foo</Heading>', settings: componentsSettings },
    // CustomInput stays as a custom component (no remap, no settings here),
    // so it's a non-hidden child → h1 is accessible.
    { code: '<h1><CustomInput type="hidden" /></h1>' },
  ],
  invalid: [
    // ---- DEFAULT ELEMENT TESTS ----
    { code: '<h1 />', errors: [{ message: errorMessage }] },
    {
      code: '<h1><Bar aria-hidden /></h1>',
      errors: [{ message: errorMessage }],
    },
    { code: '<h1>{undefined}</h1>', errors: [{ message: errorMessage }] },
    {
      code: '<h1><input type="hidden" /></h1>',
      errors: [{ message: errorMessage }],
    },

    // ---- CUSTOM ELEMENT TESTS FOR COMPONENTS OPTION ----
    {
      code: '<Heading />',
      errors: [{ message: errorMessage }],
      options: components,
    },
    {
      code: '<Heading><Bar aria-hidden /></Heading>',
      errors: [{ message: errorMessage }],
      options: components,
    },
    {
      code: '<Heading>{undefined}</Heading>',
      errors: [{ message: errorMessage }],
      options: components,
    },

    // ---- CUSTOM ELEMENT TESTS FOR COMPONENTS SETTINGS ----
    {
      code: '<Heading />',
      errors: [{ message: errorMessage }],
      settings: componentsSettings,
    },
    {
      code: '<h1><CustomInput type="hidden" /></h1>',
      errors: [{ message: errorMessage }],
      settings: componentsSettings,
    },
  ],
});
