import { RuleTester } from '../rule-tester';

const componentsSettings = {
  'jsx-a11y': {
    polymorphicPropName: 'as',
    components: {
      Input: 'input',
    },
  },
};

const arrayOpts = [
  {
    img: ['Thumbnail', 'Image'],
    object: ['Object'],
    area: ['Area'],
    'input[type="image"]': ['InputImage'],
  },
];

new RuleTester().run('alt-text', null as never, {
  valid: [
    // ---- DEFAULT ELEMENT 'img' TESTS ----
    { code: '<img alt="foo" />;' },
    { code: '<img alt={"foo"} />;' },
    { code: '<img alt={alt} />;' },
    { code: '<img ALT="foo" />;' },
    { code: '<img ALT={`This is the ${alt} text`} />;' },
    { code: '<img ALt="foo" />;' },
    { code: '<img alt="foo" salt={undefined} />;' },
    { code: '<img {...this.props} alt="foo" />' },
    { code: '<a />' },
    { code: '<div />' },
    { code: '<img alt={function(e) {} } />' },
    { code: '<div alt={function(e) {} } />' },
    { code: '<img alt={() => void 0} />' },
    { code: '<IMG />' },
    { code: '<UX.Layout>test</UX.Layout>' },
    { code: '<img alt={alt || "Alt text" } />' },
    { code: '<img alt={photo.caption} />;' },
    { code: '<img alt={bar()} />;' },
    { code: '<img alt={foo.bar || ""} />' },
    { code: '<img alt={bar() || ""} />' },
    { code: '<img alt={foo.bar() || ""} />' },
    { code: '<img alt="" />' },
    { code: '<img alt={`${undefined}`} />' },
    { code: '<img alt=" " />' },
    { code: '<img alt="" role="presentation" />' },
    { code: '<img alt="" role="none" />' },
    { code: '<img alt="" role={`presentation`} />' },
    { code: '<img alt="" role={"presentation"} />' },
    { code: '<img alt="this is lit..." role="presentation" />' },
    { code: '<img alt={error ? "not working": "working"} />' },
    { code: '<img alt={undefined ? "working": "not working"} />' },
    { code: '<img alt={plugin.name + " Logo"} />' },
    { code: '<img aria-label="foo" />' },
    { code: '<img aria-labelledby="id1" />' },

    // ---- DEFAULT <object> TESTS ----
    { code: '<object aria-label="foo" />' },
    { code: '<object aria-labelledby="id1" />' },
    { code: '<object>Foo</object>' },
    { code: '<object><p>This is descriptive!</p></object>' },
    { code: '<Object />' },
    { code: '<object title="An object" />' },

    // ---- DEFAULT <area> TESTS ----
    { code: '<area aria-label="foo" />' },
    { code: '<area aria-labelledby="id1" />' },
    { code: '<area alt="" />' },
    { code: '<area alt="This is descriptive!" />' },
    { code: '<area alt={altText} />' },
    { code: '<Area />' },

    // ---- DEFAULT <input type="image"> TESTS ----
    { code: '<input />' },
    { code: '<input type="foo" />' },
    { code: '<input type="image" aria-label="foo" />' },
    { code: '<input type="image" aria-labelledby="id1" />' },
    { code: '<input type="image" alt="" />' },
    { code: '<input type="image" alt="This is descriptive!" />' },
    { code: '<input type="image" alt={altText} />' },
    { code: '<InputImage />' },
    { code: '<Input type="image" alt="" />', settings: componentsSettings },
    {
      code: '<SomeComponent as="input" type="image" alt="" />',
      settings: componentsSettings,
    },

    // ---- CUSTOM ELEMENT TESTS FOR ARRAY OPTION TESTS ----
    { code: '<Thumbnail alt="foo" />;', options: arrayOpts },
    { code: '<Thumbnail alt={"foo"} />;', options: arrayOpts },
    { code: '<Thumbnail alt={alt} />;', options: arrayOpts },
    { code: '<Thumbnail ALT="foo" />;', options: arrayOpts },
    {
      code: '<Thumbnail ALT={`This is the ${alt} text`} />;',
      options: arrayOpts,
    },
    { code: '<Thumbnail ALt="foo" />;', options: arrayOpts },
    { code: '<Thumbnail alt="foo" salt={undefined} />;', options: arrayOpts },
    { code: '<Thumbnail {...this.props} alt="foo" />', options: arrayOpts },
    { code: '<thumbnail />', options: arrayOpts },
    { code: '<Thumbnail alt={function(e) {} } />', options: arrayOpts },
    { code: '<div alt={function(e) {} } />', options: arrayOpts },
    { code: '<Thumbnail alt={() => void 0} />', options: arrayOpts },
    { code: '<THUMBNAIL />', options: arrayOpts },
    { code: '<Thumbnail alt={alt || "foo" } />', options: arrayOpts },
    { code: '<Image alt="foo" />;', options: arrayOpts },
    { code: '<Image alt={"foo"} />;', options: arrayOpts },
    { code: '<Image alt={alt} />;', options: arrayOpts },
    { code: '<Image ALT="foo" />;', options: arrayOpts },
    { code: '<Image ALT={`This is the ${alt} text`} />;', options: arrayOpts },
    { code: '<Image ALt="foo" />;', options: arrayOpts },
    { code: '<Image alt="foo" salt={undefined} />;', options: arrayOpts },
    { code: '<Image {...this.props} alt="foo" />', options: arrayOpts },
    { code: '<image />', options: arrayOpts },
    { code: '<Image alt={function(e) {} } />', options: arrayOpts },
    { code: '<div alt={function(e) {} } />', options: arrayOpts },
    { code: '<Image alt={() => void 0} />', options: arrayOpts },
    { code: '<IMAGE />', options: arrayOpts },
    { code: '<Image alt={alt || "foo" } />', options: arrayOpts },
    { code: '<Object aria-label="foo" />', options: arrayOpts },
    { code: '<Object aria-labelledby="id1" />', options: arrayOpts },
    { code: '<Object>Foo</Object>', options: arrayOpts },
    {
      code: '<Object><p>This is descriptive!</p></Object>',
      options: arrayOpts,
    },
    { code: '<Object title="An object" />', options: arrayOpts },
    { code: '<Area aria-label="foo" />', options: arrayOpts },
    { code: '<Area aria-labelledby="id1" />', options: arrayOpts },
    { code: '<Area alt="" />', options: arrayOpts },
    { code: '<Area alt="This is descriptive!" />', options: arrayOpts },
    { code: '<Area alt={altText} />', options: arrayOpts },
    { code: '<InputImage aria-label="foo" />', options: arrayOpts },
    { code: '<InputImage aria-labelledby="id1" />', options: arrayOpts },
    { code: '<InputImage alt="" />', options: arrayOpts },
    { code: '<InputImage alt="This is descriptive!" />', options: arrayOpts },
    { code: '<InputImage alt={altText} />', options: arrayOpts },
  ],
  invalid: [
    // ---- DEFAULT ELEMENT 'img' TESTS ----
    {
      code: '<img />;',
      errors: [
        {
          message:
            'img elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<img alt />;',
      errors: [
        {
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<img alt={undefined} />;',
      errors: [
        {
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<img src="xyz" />',
      errors: [
        {
          message:
            'img elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<img role />',
      errors: [
        {
          message:
            'img elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<img {...this.props} />',
      errors: [
        {
          message:
            'img elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<img alt={false || false} />',
      errors: [
        {
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<img alt={undefined} role="presentation" />;',
      errors: [
        {
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<img alt role="presentation" />;',
      errors: [
        {
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<img role="presentation" />;',
      errors: [
        {
          message:
            'Prefer alt="" over a presentational role. First rule of aria is to not use aria if it can be achieved via native HTML.',
        },
      ],
    },
    {
      code: '<img role="none" />;',
      errors: [
        {
          message:
            'Prefer alt="" over a presentational role. First rule of aria is to not use aria if it can be achieved via native HTML.',
        },
      ],
    },
    {
      code: '<img aria-label={undefined} />',
      errors: [
        {
          message:
            'The aria-label attribute must have a value. The alt attribute is preferred over aria-label for images.',
        },
      ],
    },
    {
      code: '<img aria-labelledby={undefined} />',
      errors: [
        {
          message:
            'The aria-labelledby attribute must have a value. The alt attribute is preferred over aria-labelledby for images.',
        },
      ],
    },
    {
      code: '<img aria-label="" />',
      errors: [
        {
          message:
            'The aria-label attribute must have a value. The alt attribute is preferred over aria-label for images.',
        },
      ],
    },
    {
      code: '<img aria-labelledby="" />',
      errors: [
        {
          message:
            'The aria-labelledby attribute must have a value. The alt attribute is preferred over aria-labelledby for images.',
        },
      ],
    },
    {
      code: '<SomeComponent as="img" aria-label="" />',
      settings: componentsSettings,
      errors: [
        {
          message:
            'The aria-label attribute must have a value. The alt attribute is preferred over aria-label for images.',
        },
      ],
    },

    // ---- DEFAULT ELEMENT 'object' TESTS ----
    {
      code: '<object />',
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<object><div aria-hidden /></object>',
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<object title={undefined} />',
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<object aria-label="" />',
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<object aria-labelledby="" />',
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<object aria-label={undefined} />',
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<object aria-labelledby={undefined} />',
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },

    // ---- DEFAULT ELEMENT 'area' TESTS ----
    {
      code: '<area />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<area alt />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<area alt={undefined} />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<area src="xyz" />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<area {...this.props} />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<area aria-label="" />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<area aria-label={undefined} />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<area aria-labelledby="" />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<area aria-labelledby={undefined} />',
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },

    // ---- DEFAULT ELEMENT 'input type="image"' TESTS ----
    {
      code: '<input type="image" />',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<input type="image" alt />',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<input type="image" alt={undefined} />',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<input type="image">Foo</input>',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<input type="image" {...this.props} />',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<input type="image" aria-label="" />',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<input type="image" aria-label={undefined} />',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<input type="image" aria-labelledby="" />',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<input type="image" aria-labelledby={undefined} />',
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },

    // ---- CUSTOM ELEMENT TESTS FOR ARRAY OPTION TESTS ----
    {
      code: '<Thumbnail />;',
      options: arrayOpts,
      errors: [
        {
          message:
            'Thumbnail elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<Thumbnail alt />;',
      options: arrayOpts,
      errors: [
        {
          message:
            'Invalid alt value for Thumbnail. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<Thumbnail alt={undefined} />;',
      options: arrayOpts,
      errors: [
        {
          message:
            'Invalid alt value for Thumbnail. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<Thumbnail src="xyz" />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Thumbnail elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<Thumbnail {...this.props} />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Thumbnail elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<Image />;',
      options: arrayOpts,
      errors: [
        {
          message:
            'Image elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<Image alt />;',
      options: arrayOpts,
      errors: [
        {
          message:
            'Invalid alt value for Image. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<Image alt={undefined} />;',
      options: arrayOpts,
      errors: [
        {
          message:
            'Invalid alt value for Image. Use alt="" for presentational images.',
        },
      ],
    },
    {
      code: '<Image src="xyz" />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Image elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<Image {...this.props} />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Image elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
        },
      ],
    },
    {
      code: '<Object />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<Object><div aria-hidden /></Object>',
      options: arrayOpts,
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<Object title={undefined} />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
        },
      ],
    },
    {
      code: '<Area />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<Area alt />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<Area alt={undefined} />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<Area src="xyz" />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<Area {...this.props} />',
      options: arrayOpts,
      errors: [
        {
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<InputImage />',
      options: arrayOpts,
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<InputImage alt />',
      options: arrayOpts,
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<InputImage alt={undefined} />',
      options: arrayOpts,
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<InputImage>Foo</InputImage>',
      options: arrayOpts,
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<InputImage {...this.props} />',
      options: arrayOpts,
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
    {
      code: '<Input type="image" />',
      settings: componentsSettings,
      errors: [
        {
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
        },
      ],
    },
  ],
});
