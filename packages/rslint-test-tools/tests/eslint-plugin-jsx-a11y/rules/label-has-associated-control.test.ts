import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Mirrors upstream's `errorMessages` constant — keep these strings
// byte-identical to the Go-side `msg*` constants and to upstream's
// `__tests__/src/rules/label-has-associated-control-test.js`.

const componentsSettings = {
  'jsx-a11y': {
    components: {
      CustomInput: 'input',
      CustomLabel: 'label',
    },
  },
};

const attributesSettings = {
  'jsx-a11y': {
    attributes: {
      for: ['htmlFor', 'for'],
    },
  },
};

// withAssert wraps options for the upstream "run the same case 4×, once per
// assert mode" pattern. Per-case `options[0]` (when present) is merged with
// the assert override so per-case overrides like `{ depth: 4 }` survive.
type RawCase = {
  code: string;
  options?: any[];
  settings?: Record<string, any>;
};

const withAssert = (
  assert: 'htmlFor' | 'nesting' | 'both' | 'either',
  cases: readonly RawCase[],
) =>
  cases.map((c) => {
    const baseOpts = (c.options?.[0] as Record<string, any>) ?? {};
    return {
      code: c.code,
      options: [{ ...baseOpts, assert }],
      ...(c.settings ? { settings: c.settings } : {}),
    };
  });

// Upstream `htmlForValid` / `nestingValid` / `bothValid` / `alwaysValid` /
// `htmlForInvalid` / `nestingInvalid` / `neverValid` — preserve the source
// shape so future audits can diff line-for-line against upstream's
// `__tests__/src/rules/label-has-associated-control-test.js`.

const htmlForValid: RawCase[] = [
  {
    code: '<label htmlFor="js_id"><span><span><span>A label</span></span></span></label>',
    options: [{ depth: 4 }],
  },
  { code: '<label htmlFor="js_id" aria-label="A label" />' },
  { code: '<label htmlFor="js_id" aria-labelledby="A label" />' },
  {
    code: '<div><label htmlFor="js_id">A label</label><input id="js_id" /></div>',
  },
  {
    code: '<label for="js_id"><span><span><span>A label</span></span></span></label>',
    options: [{ depth: 4 }],
    settings: attributesSettings,
  },
  {
    code: '<label for="js_id" aria-label="A label" />',
    settings: attributesSettings,
  },
  {
    code: '<label for="js_id" aria-labelledby="A label" />',
    settings: attributesSettings,
  },
  {
    code: '<div><label for="js_id">A label</label><input id="js_id" /></div>',
    settings: attributesSettings,
  },
  // Custom label component.
  {
    code: '<CustomLabel htmlFor="js_id" aria-label="A label" />',
    options: [{ labelComponents: ['CustomLabel'] }],
  },
  {
    code: '<CustomLabel htmlFor="js_id" label="A label" />',
    options: [{ labelAttributes: ['label'], labelComponents: ['CustomLabel'] }],
  },
  {
    code: '<CustomLabel htmlFor="js_id" aria-label="A label" />',
    settings: componentsSettings,
  },
  {
    code: '<MUILabel htmlFor="js_id" aria-label="A label" />',
    options: [{ labelComponents: ['*Label'] }],
  },
  {
    code: '<LabelCustom htmlFor="js_id" label="A label" />',
    options: [{ labelAttributes: ['label'], labelComponents: ['Label*'] }],
  },
  // Custom label attributes.
  {
    code: '<label htmlFor="js_id" label="A label" />',
    options: [{ labelAttributes: ['label'] }],
  },
  // Glob support for controlComponents option.
  {
    code: '<CustomLabel htmlFor="js_id" aria-label="A label" />',
    options: [{ controlComponents: ['Custom*'] }],
  },
  {
    code: '<CustomLabel htmlFor="js_id" aria-label="A label" />',
    options: [{ controlComponents: ['*Label'] }],
  },
  // Rule does not error if presence of accessible label cannot be determined.
  {
    code: '<div><label htmlFor="js_id"><CustomText /></label><input id="js_id" /></div>',
  },
];

const nestingValid: RawCase[] = [
  { code: '<label>A label<input /></label>' },
  { code: '<label>A label<textarea /></label>' },
  { code: '<label><img alt="A label" /><input /></label>' },
  { code: '<label><img aria-label="A label" /><input /></label>' },
  { code: '<label><span>A label<input /></span></label>' },
  {
    code: '<label><span><span>A label<input /></span></span></label>',
    options: [{ depth: 3 }],
  },
  {
    code: '<label><span><span><span>A label<input /></span></span></span></label>',
    options: [{ depth: 4 }],
  },
  {
    code: '<label><span><span><span><span>A label</span><input /></span></span></span></label>',
    options: [{ depth: 5 }],
  },
  {
    code: '<label><span><span><span><span aria-label="A label" /><input /></span></span></span></label>',
    options: [{ depth: 5 }],
  },
  {
    code: '<label><span><span><span><input aria-label="A label" /></span></span></span></label>',
    options: [{ depth: 5 }],
  },
  // Other controls.
  { code: '<label>foo<meter /></label>' },
  { code: '<label>foo<output /></label>' },
  { code: '<label>foo<progress /></label>' },
  { code: '<label>foo<textarea /></label>' },
  // Custom controlComponents.
  {
    code: '<label>A label<CustomInput /></label>',
    options: [{ controlComponents: ['CustomInput'] }],
  },
  {
    code: '<label><span>A label<CustomInput /></span></label>',
    options: [{ controlComponents: ['CustomInput'] }],
  },
  {
    code: '<label><span>A label<CustomInput /></span></label>',
    settings: componentsSettings,
  },
  {
    code: '<CustomLabel><span>A label<CustomInput /></span></CustomLabel>',
    options: [
      {
        controlComponents: ['CustomInput'],
        labelComponents: ['CustomLabel'],
      },
    ],
  },
  {
    code: '<CustomLabel><span label="A label"><CustomInput /></span></CustomLabel>',
    options: [
      {
        controlComponents: ['CustomInput'],
        labelComponents: ['CustomLabel'],
        labelAttributes: ['label'],
      },
    ],
  },
  // Glob support for controlComponents.
  {
    code: '<label><span>A label<CustomInput /></span></label>',
    options: [{ controlComponents: ['Custom*'] }],
  },
  {
    code: '<label><span>A label<CustomInput /></span></label>',
    options: [{ controlComponents: ['*Input'] }],
  },
  {
    code: '<label><span>A label<TextInput /></span></label>',
    options: [{ controlComponents: ['????Input'] }],
  },
  // Rule does not error if presence of accessible label cannot be determined.
  { code: '<label><CustomText /><input /></label>' },
];

const bothValid: RawCase[] = [
  {
    code: '<label htmlFor="js_id"><span><span><span>A label<input /></span></span></span></label>',
    options: [{ depth: 4 }],
  },
  { code: '<label htmlFor="js_id" aria-label="A label"><input /></label>' },
  {
    code: '<label htmlFor="js_id" aria-labelledby="A label"><input /></label>',
  },
  {
    code: '<label htmlFor="js_id" aria-labelledby="A label"><textarea /></label>',
  },
  // Custom label component.
  {
    code: '<CustomLabel htmlFor="js_id" aria-label="A label"><input /></CustomLabel>',
    options: [{ labelComponents: ['CustomLabel'] }],
  },
  {
    code: '<CustomLabel htmlFor="js_id" label="A label"><input /></CustomLabel>',
    options: [{ labelAttributes: ['label'], labelComponents: ['CustomLabel'] }],
  },
  {
    code: '<CustomLabel htmlFor="js_id" label="A label"><input /></CustomLabel>',
    options: [{ labelAttributes: ['label'], labelComponents: ['*Label'] }],
  },
  {
    code: '<CustomLabel htmlFor="js_id" aria-label="A label"><input /></CustomLabel>',
    settings: componentsSettings,
  },
  {
    code: '<CustomLabel htmlFor="js_id" aria-label="A label"><CustomInput /></CustomLabel>',
    settings: componentsSettings,
  },
  // Custom label attributes.
  {
    code: '<label htmlFor="js_id" label="A label"><input /></label>',
    options: [{ labelAttributes: ['label'] }],
  },
  {
    code: '<label htmlFor="selectInput">Some text<select id="selectInput" /></label>',
  },
];

const alwaysValid: RawCase[] = [
  { code: '<div />' },
  { code: '<CustomElement />' },
  { code: '<input type="hidden" />' },
];

// neverValid — labels that fail in EVERY mode. The first few entries always
// emit `accessibleLabel` (no text); remaining entries emit the assert-
// specific error.

// Four ruleTester.run blocks — once per assert mode, mirroring upstream.

new RuleTester(config).run('label-has-associated-control', {
  valid: [
    ...withAssert('htmlFor', alwaysValid),
    ...withAssert('htmlFor', htmlForValid),
  ],
  invalid: [
    {
      code: '<label htmlFor="js_id" />',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><input /></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><textarea /></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label>A label</label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<div><label /><input /></div>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<div><label>A label</label><input /></div>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<CustomLabel aria-label="A label" />',
      options: [
        {
          labelComponents: ['CustomLabel'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<MUILabel aria-label="A label" />',
      options: [
        {
          labelComponents: ['???Label'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<CustomLabel label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          labelComponents: ['CustomLabel'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<CustomLabel aria-label="A label" />',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<label label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<label><span><CustomInput /></span></label>',
      options: [
        {
          controlComponents: ['CustomInput'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          labelAttributes: ['label'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label><span><CustomInput /></span></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label>A label<input /></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label>A label<textarea /></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><img alt="A label" /><input /></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><img aria-label="A label" /><input /></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span>A label<input /></span></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span>A label<input /></span></span></label>',
      options: [
        {
          depth: 3,
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span><span>A label<input /></span></span></span></label>',
      options: [
        {
          depth: 4,
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span><span><span>A label</span><input /></span></span></span></label>',
      options: [
        {
          depth: 5,
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span><span><span aria-label="A label" /><input /></span></span></span></label>',
      options: [
        {
          depth: 5,
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span><span><input aria-label="A label" /></span></span></span></label>',
      options: [
        {
          depth: 5,
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label>A label<OtherCustomInput /></label>',
      options: [
        {
          controlComponents: ['CustomInput'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span>A label<CustomInput /></span></label>',
      options: [
        {
          controlComponents: ['CustomInput'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span>A label<CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<CustomLabel><span label="A label"><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          labelAttributes: ['label'],
          assert: 'htmlFor',
        },
      ],
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label><span>A label<CustomInput /></span></label>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span>A label<CustomInput /></span></CustomLabel>',
      options: [
        {
          assert: 'htmlFor',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'htmlFor',
          message: 'A form label must have a valid htmlFor attribute.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
  ],
});

new RuleTester(config).run('label-has-associated-control', {
  valid: [
    ...withAssert('nesting', alwaysValid),
    ...withAssert('nesting', nestingValid),
  ],
  invalid: [
    {
      code: '<label htmlFor="js_id" />',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><input /></label>',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><textarea /></label>',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label></label>',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label>A label</label>',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<div><label /><input /></div>',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<div><label>A label</label><input /></div>',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<CustomLabel aria-label="A label" />',
      options: [
        {
          labelComponents: ['CustomLabel'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<MUILabel aria-label="A label" />',
      options: [
        {
          labelComponents: ['???Label'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<CustomLabel label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          labelComponents: ['CustomLabel'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<CustomLabel aria-label="A label" />',
      options: [
        {
          assert: 'nesting',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<label label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<label><span><CustomInput /></span></label>',
      options: [
        {
          controlComponents: ['CustomInput'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          labelAttributes: ['label'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label><span><CustomInput /></span></label>',
      options: [
        {
          assert: 'nesting',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          assert: 'nesting',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><span><span><span>A label</span></span></span></label>',
      options: [
        {
          depth: 4,
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id" aria-label="A label" />',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id" aria-labelledby="A label" />',
      options: [
        {
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<CustomLabel htmlFor="js_id" aria-label="A label" />',
      options: [
        {
          labelComponents: ['CustomLabel'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<CustomLabel htmlFor="js_id" label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          labelComponents: ['CustomLabel'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<CustomLabel htmlFor="js_id" aria-label="A label" />',
      options: [
        {
          assert: 'nesting',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id" label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          assert: 'nesting',
        },
      ],
      errors: [
        {
          messageId: 'nesting',
          message:
            'A form label must have an associated control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
  ],
});

new RuleTester(config).run('label-has-associated-control', {
  valid: [
    ...withAssert('either', alwaysValid),
    ...withAssert('either', htmlForValid),
    ...withAssert('either', nestingValid),
  ],
  invalid: [
    {
      code: '<label htmlFor="js_id" />',
      options: [
        {
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><input /></label>',
      options: [
        {
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><textarea /></label>',
      options: [
        {
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label></label>',
      options: [
        {
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label>A label</label>',
      options: [
        {
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'either',
          message:
            'A form label must either have a valid htmlFor attribute or a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<div><label /><input /></div>',
      options: [
        {
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<div><label>A label</label><input /></div>',
      options: [
        {
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'either',
          message:
            'A form label must either have a valid htmlFor attribute or a control as a descendant.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<CustomLabel aria-label="A label" />',
      options: [
        {
          labelComponents: ['CustomLabel'],
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'either',
          message:
            'A form label must either have a valid htmlFor attribute or a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<MUILabel aria-label="A label" />',
      options: [
        {
          labelComponents: ['???Label'],
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'either',
          message:
            'A form label must either have a valid htmlFor attribute or a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<CustomLabel label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          labelComponents: ['CustomLabel'],
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'either',
          message:
            'A form label must either have a valid htmlFor attribute or a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<CustomLabel aria-label="A label" />',
      options: [
        {
          assert: 'either',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'either',
          message:
            'A form label must either have a valid htmlFor attribute or a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<label label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'either',
          message:
            'A form label must either have a valid htmlFor attribute or a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<label><span><CustomInput /></span></label>',
      options: [
        {
          controlComponents: ['CustomInput'],
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          labelAttributes: ['label'],
          assert: 'either',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label><span><CustomInput /></span></label>',
      options: [
        {
          assert: 'either',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          assert: 'either',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
  ],
});

new RuleTester(config).run('label-has-associated-control', {
  valid: [...withAssert('both', alwaysValid), ...withAssert('both', bothValid)],
  invalid: [
    {
      code: '<label htmlFor="js_id" />',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><input /></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><textarea /></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label>A label</label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<div><label /><input /></div>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 15,
        },
      ],
    },
    {
      code: '<div><label>A label</label><input /></div>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 13,
        },
      ],
    },
    {
      code: '<CustomLabel aria-label="A label" />',
      options: [
        {
          labelComponents: ['CustomLabel'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<MUILabel aria-label="A label" />',
      options: [
        {
          labelComponents: ['???Label'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 34,
        },
      ],
    },
    {
      code: '<CustomLabel label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          labelComponents: ['CustomLabel'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: '<CustomLabel aria-label="A label" />',
      options: [
        {
          assert: 'both',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: '<label label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: '<label><span><CustomInput /></span></label>',
      options: [
        {
          controlComponents: ['CustomInput'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          labelAttributes: ['label'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label><span><CustomInput /></span></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
      options: [
        {
          assert: 'both',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'accessibleLabel',
          message: 'A form label must have accessible text.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id"><span><span><span>A label</span></span></span></label>',
      options: [
        {
          depth: 4,
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id" aria-label="A label" />',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 47,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id" aria-labelledby="A label" />',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 52,
        },
      ],
    },
    {
      code: '<CustomLabel htmlFor="js_id" aria-label="A label" />',
      options: [
        {
          labelComponents: ['CustomLabel'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<CustomLabel htmlFor="js_id" label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          labelComponents: ['CustomLabel'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
    {
      code: '<CustomLabel htmlFor="js_id" aria-label="A label" />',
      options: [
        {
          assert: 'both',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: '<label htmlFor="js_id" label="A label" />',
      options: [
        {
          labelAttributes: ['label'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 42,
        },
      ],
    },
    {
      code: '<label>A label<input /></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label>A label<textarea /></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><img alt="A label" /><input /></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><img aria-label="A label" /><input /></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span>A label<input /></span></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span>A label<input /></span></span></label>',
      options: [
        {
          depth: 3,
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span><span>A label<input /></span></span></span></label>',
      options: [
        {
          depth: 4,
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span><span><span>A label</span><input /></span></span></span></label>',
      options: [
        {
          depth: 5,
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span><span><span aria-label="A label" /><input /></span></span></span></label>',
      options: [
        {
          depth: 5,
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span><span><span><input aria-label="A label" /></span></span></span></label>',
      options: [
        {
          depth: 5,
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label>A label<OtherCustomInput /></label>',
      options: [
        {
          controlComponents: ['CustomInput'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<label><span>A label<CustomInput /></span></label>',
      options: [
        {
          controlComponents: ['CustomInput'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span>A label<CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<CustomLabel><span label="A label"><CustomInput /></span></CustomLabel>',
      options: [
        {
          controlComponents: ['CustomInput'],
          labelComponents: ['CustomLabel'],
          labelAttributes: ['label'],
          assert: 'both',
        },
      ],
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
    {
      code: '<label><span>A label<CustomInput /></span></label>',
      options: [
        {
          assert: 'both',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<CustomLabel><span>A label<CustomInput /></span></CustomLabel>',
      options: [
        {
          assert: 'both',
        },
      ],
      settings: {
        'jsx-a11y': {
          components: {
            CustomInput: 'input',
            CustomLabel: 'label',
          },
        },
      },
      errors: [
        {
          messageId: 'both',
          message:
            'A form label must have a valid htmlFor attribute and a control as a descendant.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 14,
        },
      ],
    },
  ],
});
