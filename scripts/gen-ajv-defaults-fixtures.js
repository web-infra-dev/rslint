// scripts/gen-ajv-defaults-fixtures.js
//
// Generates internal/rule/testdata/ajv_defaults_fixtures.json: a corpus of
// {schema, input} pairs run through ajv@6 configured exactly the way ESLint
// configures it (see node_modules/ajv's consumer, eslint/lib/shared/ajv.js),
// so internal/rule's schema.go can be tested for output parity with the
// real-world validator ESLint itself uses for rule options.
const fs = require('fs');
const path = require('path');
const Ajv = require('ajv');
const metaSchema = require('ajv/lib/refs/json-schema-draft-04.json');

// Mirrors eslint/lib/shared/ajv.js exactly.
function newEslintAjv() {
  const ajv = new Ajv({
    meta: false,
    useDefaults: true,
    validateSchema: false,
    missingRefs: 'ignore',
    verbose: true,
    schemaId: 'auto',
  });
  ajv.addMetaSchema(metaSchema);
  ajv._opts.defaultMeta = metaSchema.id;
  return ajv;
}

const cases = [
  {
    name: 'object_property_defaults_fill_missing_keys',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: {
            allow: { type: 'array', items: { type: 'string' }, default: [] },
            strict: { type: 'boolean', default: true },
          },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ strict: false }],
  },
  {
    name: 'object_property_defaults_fill_all_when_empty',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: {
            allow: { type: 'array', items: { type: 'string' }, default: [] },
            strict: { type: 'boolean', default: true },
          },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
  {
    name: 'no_options_provided_first_tuple_item_filled_from_default',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: { allow: { type: 'array', default: [] } },
          additionalProperties: false,
          default: {},
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [],
  },
  {
    name: 'multiple_trailing_tuple_items_filled_from_defaults',
    schema: {
      type: 'array',
      items: [
        { type: 'string', default: 'a' },
        { type: 'number', default: 1 },
        { type: 'boolean', default: true },
      ],
      minItems: 0,
      maxItems: 3,
    },
    input: ['given'],
  },
  {
    name: 'later_tuple_item_default_fills_creating_gap',
    schema: {
      type: 'array',
      items: [{ type: 'string' }, { type: 'number', default: 42 }],
      minItems: 0,
      maxItems: 2,
    },
    input: [],
  },
  {
    name: 'nested_object_in_object_defaults',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: {
            outer: {
              type: 'object',
              properties: {
                inner: { type: 'string', default: 'fallback' },
              },
              additionalProperties: false,
              default: {},
            },
          },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
  {
    name: 'nested_object_in_tuple_item_defaults',
    schema: {
      type: 'array',
      items: [
        { type: 'string' },
        {
          type: 'object',
          properties: { level: { type: 'string', default: 'warn' } },
          additionalProperties: false,
        },
      ],
      minItems: 1,
      maxItems: 2,
    },
    input: ['first', {}],
  },
  {
    name: 'existing_property_is_not_overwritten',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: { allow: { type: 'array', default: ['x'] } },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ allow: ['y'] }],
  },
  {
    name: 'explicit_null_is_not_overwritten_by_default',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: { allow: { type: ['array', 'null'], default: ['x'] } },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ allow: null }],
  },
  {
    name: 'oneOf_default_not_applied_inside_composite_rule',
    schema: {
      type: 'array',
      items: [
        {
          oneOf: [
            {
              type: 'object',
              properties: { mode: { type: 'string', default: 'a' } },
              additionalProperties: false,
            },
            {
              type: 'object',
              properties: { mode: { const: 'b' } },
              additionalProperties: false,
            },
          ],
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
  {
    name: 'list_style_items_default_not_applied_per_element',
    schema: {
      type: 'array',
      items: [
        {
          type: 'array',
          items: { type: 'string', default: 'x' },
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [['a']],
  },
  {
    name: 'no_defaults_needed_output_unchanged',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: {
            allow: { type: 'array', items: { type: 'string' }, default: [] },
          },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ allow: ['warn', 'error'] }],
  },
  {
    name: 'empty_options_array_with_no_items_schema_stays_empty',
    schema: { type: 'array', maxItems: 0 },
    input: [],
  },
  {
    name: 'defaults_still_applied_when_result_is_invalid',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: {
            allow: { type: 'array', default: [] },
          },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ unknown: true }],
  },
  {
    name: 'numeric_default_type',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: { max: { type: 'number', default: 3.5 } },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
];

function run() {
  const results = cases.map(({ name, schema, input }) => {
    const ajv = newEslintAjv();
    const validate = ajv.compile(schema);
    // ajv mutates the data argument in place; keep the original input
    // untouched in the fixture for the Go test to feed in fresh.
    const data = JSON.parse(JSON.stringify(input));
    const valid = validate(data);
    return {
      name,
      schema,
      input,
      output: data,
      valid,
      errors: valid ? null : validate.errors,
    };
  });

  const outPath = path.join(
    __dirname,
    '../internal/rule/testdata/ajv_defaults_fixtures.json',
  );
  fs.writeFileSync(outPath, JSON.stringify(results, null, 2) + '\n');
  console.log(`Wrote ${results.length} cases to ${outPath}`);
}

run();
