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
    // A list-style (single-schema, non-tuple) `items` schema still gets its
    // default filled in for each already-present element independently.
    name: 'list_style_items_default_applied_per_element',
    schema: {
      type: 'array',
      items: [
        {
          type: 'array',
          items: {
            type: 'object',
            properties: { foo: { type: 'string', default: 'd' } },
          },
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [[{}, { foo: 'x' }]],
  },
  {
    // list-style `items` never grows the array to satisfy minItems — it says
    // nothing about how many elements there are, unlike a tuple position
    // with its own literal default.
    name: 'list_style_items_does_not_grow_array_to_satisfy_minItems',
    schema: {
      type: 'array',
      items: [
        {
          type: 'array',
          minItems: 2,
          items: { type: 'string', default: 'd' },
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [['a']],
  },
  {
    // The real-world case this was modeled after: a rule's own top-level
    // options schema is itself list-style (a single object schema, not a
    // tuple), the way @graphql-eslint/eslint-plugin's relay-arguments does.
    // An empty `{}` option gets its declared default filled in, satisfying
    // minProperties — matching ajv, this used to be silently skipped.
    name: 'top_level_list_style_items_default_applied',
    schema: {
      type: 'array',
      maxItems: 1,
      items: {
        type: 'object',
        minProperties: 1,
        properties: { includeBoth: { type: 'boolean', default: true } },
      },
    },
    input: [{}],
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
  // The following $ref cases pin down a subtle ajv behavior: whether a
  // property/item's default propagates through a bare `{"$ref": ...}`
  // schema depends entirely on whether that property/item is *already
  // present*, never on whether it's absent — because ajv compiles $ref as a
  // call into the referenced schema's own compiled validator, and
  // object/array mutation-by-reference is what makes a nested default land
  // in the parent's data. A `$ref` target's own top-level default can never
  // fill an *absent* parent slot, because that decision is made by the
  // parent's own codegen looking at the literal (undereferenced) schema
  // text, which never contains `default` when it's just `{"$ref": ...}`.
  {
    name: 'ref_property_absent_ref_targets_own_default_not_pulled_through',
    schema: {
      type: 'array',
      definitions: {
        inner: {
          type: 'object',
          default: {},
          properties: { a: { type: 'number', default: 1 } },
          additionalProperties: false,
        },
      },
      items: [
        {
          type: 'object',
          properties: { foo: { $ref: '#/definitions/inner' } },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
  {
    name: 'ref_tuple_item_absent_ref_targets_own_default_not_pulled_through',
    schema: {
      type: 'array',
      definitions: {
        inner: {
          type: 'object',
          default: {},
          properties: { a: { type: 'number', default: 1 } },
          additionalProperties: false,
        },
      },
      items: [{ $ref: '#/definitions/inner' }],
      minItems: 0,
      maxItems: 1,
    },
    input: [],
  },
  {
    name: 'ref_tuple_item_present_as_empty_object_fills_nested_default_through_ref',
    schema: {
      type: 'array',
      definitions: {
        inner: {
          type: 'object',
          properties: { a: { type: 'number', default: 1 } },
          additionalProperties: false,
        },
      },
      items: [{ $ref: '#/definitions/inner' }],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
  {
    name: 'ref_property_present_as_empty_object_fills_nested_default_through_ref',
    schema: {
      type: 'array',
      definitions: {
        inner: {
          type: 'object',
          properties: { a: { type: 'number', default: 1 } },
          additionalProperties: false,
        },
      },
      items: [
        {
          type: 'object',
          properties: { foo: { $ref: '#/definitions/inner' } },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ foo: {} }],
  },
  {
    name: 'ref_multi_hop_chain_fills_nested_default_through_ref',
    schema: {
      type: 'array',
      definitions: {
        a: { $ref: '#/definitions/b' },
        b: {
          type: 'object',
          properties: { x: { type: 'string', default: 'hi' } },
          additionalProperties: false,
        },
      },
      items: [
        {
          type: 'object',
          properties: { foo: { $ref: '#/definitions/a' } },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ foo: {} }],
  },
  {
    name: 'ref_property_present_ref_target_is_array_with_tuple_default',
    schema: {
      type: 'array',
      definitions: {
        innerArr: {
          type: 'array',
          items: [{ type: 'string' }, { type: 'number', default: 42 }],
          minItems: 0,
          maxItems: 2,
        },
      },
      items: [
        {
          type: 'object',
          properties: { foo: { $ref: '#/definitions/innerArr' } },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ foo: [] }],
  },
  {
    // A default nested inside oneOf stays excluded even when reached via a
    // property that's a bare $ref to that oneOf (ban-ts-comment's own
    // ts-check/ts-expect-error/... pattern): the oneOf itself has no
    // Properties/Items for applyDefaults to walk into, ref-resolved or not.
    name: 'ref_to_oneOf_default_still_not_applied',
    schema: {
      type: 'array',
      definitions: {
        directiveConfig: {
          oneOf: [
            { type: 'boolean', default: true },
            { type: 'string', enum: ['allow-with-description'] },
          ],
        },
      },
      items: [
        {
          type: 'object',
          properties: { 'ts-check': { $ref: '#/definitions/directiveConfig' } },
          additionalProperties: false,
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
  {
    // allOf is unambiguous (every branch must hold), unlike anyOf/oneOf: ajv
    // fills in a default declared on a single allOf branch's property.
    name: 'allOf_single_branch_default_applied',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          allOf: [
            {
              properties: {
                mode: { type: 'string', default: 'allof-default' },
              },
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
    // Each allOf branch contributes its own properties' defaults; here two
    // branches declare defaults for two different keys and both get filled.
    name: 'allOf_multiple_branches_each_contribute_defaults',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          allOf: [
            { properties: { a: { type: 'string', default: 'a-default' } } },
            { properties: { b: { type: 'string', default: 'b-default' } } },
          ],
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
  {
    // Regression guard: anyOf/oneOf branch defaults must stay excluded even
    // now that allOf branches are folded in (they are genuinely ambiguous —
    // unlike allOf, not every branch is guaranteed to hold).
    name: 'anyOf_branch_default_still_not_applied',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          anyOf: [
            {
              properties: {
                mode: { type: 'string', default: 'anyof-default' },
              },
            },
            { properties: { other: { type: 'string' } } },
          ],
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{}],
  },
  {
    // additionalProperties as a schema (rather than a boolean) never
    // manufactures a brand-new key — ajv's useDefaults only ever creates
    // keys/items declared in a schema's own `properties`/tuple `items`.
    name: 'additionalProperties_schema_default_does_not_create_new_key',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: { known: { type: 'string' } },
          additionalProperties: { type: 'string', default: 'ap-default' },
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ known: 'x' }],
  },
  {
    // But a key that's already present and matched only by
    // additionalProperties (not by a named `properties` entry) still gets
    // its own nested defaults filled in, same as any other present value.
    name: 'additionalProperties_schema_fills_nested_default_for_existing_key',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          properties: { known: { type: 'string' } },
          additionalProperties: {
            type: 'object',
            properties: { inner: { type: 'string', default: 'inner-default' } },
          },
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ known: 'x', extra: {} }],
  },
  {
    // Same as above but via patternProperties instead of additionalProperties.
    name: 'patternProperties_fills_nested_default_for_existing_key',
    schema: {
      type: 'array',
      items: [
        {
          type: 'object',
          patternProperties: {
            '^opt_': {
              type: 'object',
              properties: { inner: { type: 'string', default: 'pp-default' } },
            },
          },
        },
      ],
      minItems: 0,
      maxItems: 1,
    },
    input: [{ opt_x: {} }],
  },
  {
    // A literal `default` written directly beside a bare `$ref` (a "$ref
    // sibling") is honored by ajv using that literal value — not the ref
    // target's own top-level default (confirmed separately to never be
    // pulled through; see ref_property_absent_ref_targets_own_default_not_pulled_through).
    name: 'ref_sibling_default_applied_using_literal_value_not_ref_target',
    schema: {
      type: 'array',
      definitions: {
        foo: { type: 'string', default: 'ref-target-default' },
      },
      items: [
        {
          type: 'object',
          properties: {
            foo: { $ref: '#/definitions/foo', default: 'sibling-default' },
          },
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
