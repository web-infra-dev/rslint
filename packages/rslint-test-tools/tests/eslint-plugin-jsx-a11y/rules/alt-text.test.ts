import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

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

new RuleTester(config).run('alt-text', {
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
          messageId: 'missingProp',
          message:
            'img elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<img alt />;',
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
        },
      ],
    },
    {
      code: '<img alt={undefined} />;',
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<img src="xyz" />',
      errors: [
        {
          messageId: 'missingProp',
          message:
            'img elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<img role />',
      errors: [
        {
          messageId: 'missingProp',
          message:
            'img elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<img {...this.props} />',
      errors: [
        {
          messageId: 'missingProp',
          message:
            'img elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<img alt={false || false} />',
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<img alt={undefined} role="presentation" />;',
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 44,
        },
      ],
    },
    {
      code: '<img alt role="presentation" />;',
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for img. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<img role="presentation" />;',
      errors: [
        {
          messageId: 'preferAlt',
          message:
            'Prefer alt="" over a presentational role. First rule of aria is to not use aria if it can be achieved via native HTML.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<img role="none" />;',
      errors: [
        {
          messageId: 'preferAlt',
          message:
            'Prefer alt="" over a presentational role. First rule of aria is to not use aria if it can be achieved via native HTML.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<img aria-label={undefined} />',
      errors: [
        {
          messageId: 'ariaLabelValue',
          message:
            'The aria-label attribute must have a value. The alt attribute is preferred over aria-label for images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<img aria-labelledby={undefined} />',
      errors: [
        {
          messageId: 'ariaLabelledByValue',
          message:
            'The aria-labelledby attribute must have a value. The alt attribute is preferred over aria-labelledby for images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: '<img aria-label="" />',
      errors: [
        {
          messageId: 'ariaLabelValue',
          message:
            'The aria-label attribute must have a value. The alt attribute is preferred over aria-label for images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      code: '<img aria-labelledby="" />',
      errors: [
        {
          messageId: 'ariaLabelledByValue',
          message:
            'The aria-labelledby attribute must have a value. The alt attribute is preferred over aria-labelledby for images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<SomeComponent as="img" aria-label="" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'ariaLabelValue',
          message:
            'The aria-label attribute must have a value. The alt attribute is preferred over aria-label for images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },

    // ---- DEFAULT ELEMENT 'object' TESTS ----
    {
      code: '<object />',
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<object><div aria-hidden /></object>',
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<object title={undefined} />',
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<object aria-label="" />',
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<object aria-labelledby="" />',
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<object aria-label={undefined} />',
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<object aria-labelledby={undefined} />',
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },

    // ---- DEFAULT ELEMENT 'area' TESTS ----
    {
      code: '<area />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<area alt />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<area alt={undefined} />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<area src="xyz" />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<area {...this.props} />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<area aria-label="" />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<area aria-label={undefined} />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<area aria-labelledby="" />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    {
      code: '<area aria-labelledby={undefined} />',
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },

    // ---- DEFAULT ELEMENT 'input type="image"' TESTS ----
    {
      code: '<input type="image" />',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: '<input type="image" alt />',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      code: '<input type="image" alt={undefined} />',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<input type="image">Foo</input>',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 21,
        },
      ],
    },
    {
      code: '<input type="image" {...this.props} />',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: '<input type="image" aria-label="" />',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<input type="image" aria-label={undefined} />',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 46,
        },
      ],
    },
    {
      code: '<input type="image" aria-labelledby="" />',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<input type="image" aria-labelledby={undefined} />',
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },

    // ---- CUSTOM ELEMENT TESTS FOR ARRAY OPTION TESTS ----
    {
      code: '<Thumbnail />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'missingProp',
          message:
            'Thumbnail elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<Thumbnail alt />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for Thumbnail. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<Thumbnail alt={undefined} />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for Thumbnail. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<Thumbnail src="xyz" />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'missingProp',
          message:
            'Thumbnail elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<Thumbnail {...this.props} />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'missingProp',
          message:
            'Thumbnail elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 30,
        },
      ],
    },
    {
      code: '<Image />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'missingProp',
          message:
            'Image elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 10,
        },
      ],
    },
    {
      code: '<Image alt />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for Image. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<Image alt={undefined} />;',
      options: arrayOpts,
      errors: [
        {
          messageId: 'altValue',
          message:
            'Invalid alt value for Image. Use alt="" for presentational images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<Image src="xyz" />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'missingProp',
          message:
            'Image elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<Image {...this.props} />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'missingProp',
          message:
            'Image elements must have an alt prop, either with meaningful text, or an empty string for decorative images.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<Object />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 11,
        },
      ],
    },
    {
      code: '<Object><div aria-hidden /></Object>',
      options: arrayOpts,
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<Object title={undefined} />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'object',
          message:
            'Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: '<Area />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 9,
        },
      ],
    },
    {
      code: '<Area alt />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<Area alt={undefined} />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<Area src="xyz" />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<Area {...this.props} />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'area',
          message:
            'Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: '<InputImage />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<InputImage alt />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
    },
    {
      code: '<InputImage alt={undefined} />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<InputImage>Foo</InputImage>',
      options: arrayOpts,
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<InputImage {...this.props} />',
      options: arrayOpts,
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      code: '<Input type="image" />',
      settings: componentsSettings,
      errors: [
        {
          messageId: 'inputImage',
          message:
            '<input> elements with type="image" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
  ],
});
