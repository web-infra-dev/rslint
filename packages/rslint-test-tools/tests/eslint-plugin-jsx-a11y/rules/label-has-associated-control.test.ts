import { RuleTester } from '../rule-tester';

// Mirrors upstream's `errorMessages` constant — keep these strings
// byte-identical to the Go-side `msg*` constants and to upstream's
// `__tests__/src/rules/label-has-associated-control-test.js`.
const errorMessages = {
  accessibleLabel: { message: 'A form label must have accessible text.' },
  htmlFor: { message: 'A form label must have a valid htmlFor attribute.' },
  nesting: {
    message: 'A form label must have an associated control as a descendant.',
  },
  either: {
    message:
      'A form label must either have a valid htmlFor attribute or a control as a descendant.',
  },
  both: {
    message:
      'A form label must have a valid htmlFor attribute and a control as a descendant.',
  },
} as const;

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

const withAssertInvalid = (
  assert: 'htmlFor' | 'nesting' | 'both' | 'either',
  expectedError: (typeof errorMessages)[keyof typeof errorMessages],
  cases: readonly RawCase[],
) =>
  cases.map((c) => {
    const baseOpts = (c.options?.[0] as Record<string, any>) ?? {};
    return {
      code: c.code,
      options: [{ ...baseOpts, assert }],
      ...(c.settings ? { settings: c.settings } : {}),
      errors: [expectedError],
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

const htmlForInvalid: RawCase[] = [
  {
    code: '<label htmlFor="js_id"><span><span><span>A label</span></span></span></label>',
    options: [{ depth: 4 }],
  },
  { code: '<label htmlFor="js_id" aria-label="A label" />' },
  { code: '<label htmlFor="js_id" aria-labelledby="A label" />' },
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
  // Custom label attributes.
  {
    code: '<label htmlFor="js_id" label="A label" />',
    options: [{ labelAttributes: ['label'] }],
  },
];

const nestingInvalid: RawCase[] = [
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
  // Custom controlComponents.
  {
    code: '<label>A label<OtherCustomInput /></label>',
    options: [{ controlComponents: ['CustomInput'] }],
  },
  {
    code: '<label><span>A label<CustomInput /></span></label>',
    options: [{ controlComponents: ['CustomInput'] }],
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
  {
    code: '<label><span>A label<CustomInput /></span></label>',
    settings: componentsSettings,
  },
  {
    code: '<CustomLabel><span>A label<CustomInput /></span></CustomLabel>',
    settings: componentsSettings,
  },
];

// neverValid — labels that fail in EVERY mode. The first few entries always
// emit `accessibleLabel` (no text); remaining entries emit the assert-
// specific error.
type NeverValidCase = RawCase & {
  useAccessibleLabel?: boolean;
};

const neverValid: NeverValidCase[] = [
  { code: '<label htmlFor="js_id" />', useAccessibleLabel: true },
  {
    code: '<label htmlFor="js_id"><input /></label>',
    useAccessibleLabel: true,
  },
  {
    code: '<label htmlFor="js_id"><textarea /></label>',
    useAccessibleLabel: true,
  },
  { code: '<label></label>', useAccessibleLabel: true },
  { code: '<label>A label</label>' },
  { code: '<div><label /><input /></div>', useAccessibleLabel: true },
  { code: '<div><label>A label</label><input /></div>' },
  // Custom label component.
  {
    code: '<CustomLabel aria-label="A label" />',
    options: [{ labelComponents: ['CustomLabel'] }],
  },
  {
    code: '<MUILabel aria-label="A label" />',
    options: [{ labelComponents: ['???Label'] }],
  },
  {
    code: '<CustomLabel label="A label" />',
    options: [{ labelAttributes: ['label'], labelComponents: ['CustomLabel'] }],
  },
  {
    code: '<CustomLabel aria-label="A label" />',
    settings: componentsSettings,
  },
  // Custom label attributes.
  {
    code: '<label label="A label" />',
    options: [{ labelAttributes: ['label'] }],
  },
  // Custom controlComponents.
  {
    code: '<label><span><CustomInput /></span></label>',
    options: [{ controlComponents: ['CustomInput'] }],
    useAccessibleLabel: true,
  },
  {
    code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
    options: [
      {
        controlComponents: ['CustomInput'],
        labelComponents: ['CustomLabel'],
      },
    ],
    useAccessibleLabel: true,
  },
  {
    code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
    options: [
      {
        controlComponents: ['CustomInput'],
        labelComponents: ['CustomLabel'],
        labelAttributes: ['label'],
      },
    ],
    useAccessibleLabel: true,
  },
  {
    code: '<label><span><CustomInput /></span></label>',
    settings: componentsSettings,
    useAccessibleLabel: true,
  },
  {
    code: '<CustomLabel><span><CustomInput /></span></CustomLabel>',
    settings: componentsSettings,
    useAccessibleLabel: true,
  },
];

const neverValidWithAssert = (
  assert: 'htmlFor' | 'nesting' | 'both' | 'either',
  assertError: (typeof errorMessages)[keyof typeof errorMessages],
) =>
  neverValid.map((c) => {
    const baseOpts = (c.options?.[0] as Record<string, any>) ?? {};
    const err = c.useAccessibleLabel
      ? errorMessages.accessibleLabel
      : assertError;
    return {
      code: c.code,
      options: [{ ...baseOpts, assert }],
      ...(c.settings ? { settings: c.settings } : {}),
      errors: [err],
    };
  });

// Four ruleTester.run blocks — once per assert mode, mirroring upstream.

new RuleTester().run('label-has-associated-control', null as never, {
  valid: [
    ...withAssert('htmlFor', alwaysValid),
    ...withAssert('htmlFor', htmlForValid),
  ],
  invalid: [
    ...neverValidWithAssert('htmlFor', errorMessages.htmlFor),
    ...withAssertInvalid('htmlFor', errorMessages.htmlFor, nestingInvalid),
  ],
});

new RuleTester().run('label-has-associated-control', null as never, {
  valid: [
    ...withAssert('nesting', alwaysValid),
    ...withAssert('nesting', nestingValid),
  ],
  invalid: [
    ...neverValidWithAssert('nesting', errorMessages.nesting),
    ...withAssertInvalid('nesting', errorMessages.nesting, htmlForInvalid),
  ],
});

new RuleTester().run('label-has-associated-control', null as never, {
  valid: [
    ...withAssert('either', alwaysValid),
    ...withAssert('either', htmlForValid),
    ...withAssert('either', nestingValid),
  ],
  invalid: [...neverValidWithAssert('either', errorMessages.either)],
});

new RuleTester().run('label-has-associated-control', null as never, {
  valid: [...withAssert('both', alwaysValid), ...withAssert('both', bothValid)],
  invalid: [
    ...neverValidWithAssert('both', errorMessages.both),
    ...withAssertInvalid('both', errorMessages.both, htmlForInvalid),
    ...withAssertInvalid('both', errorMessages.both, nestingInvalid),
  ],
});
