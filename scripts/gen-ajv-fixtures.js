// scripts/gen-ajv-fixtures.js
//
// Generates internal/rule/testdata/ajv_fixtures.json: a corpus of
// {schema, input} pairs run through ajv@6 configured exactly the way ESLint
// configures it (see node_modules/ajv's consumer, eslint/lib/shared/ajv.js),
// so internal/rule's schema.go can be tested for output parity with the
// real-world validator ESLint itself uses for rule options.
//
// Two kinds of cases feed into the same corpus:
//   - synthetic defaulting cases (SYNTHETIC_CASES below) that pin down
//     specific ajv useDefaults edge cases: nested/tuple/$ref gaps, allOf vs.
//     anyOf/oneOf exclusion, additionalProperties/patternProperties, etc.
//   - real-world cases (REALWORLD_CASES below), sourced from real ESLint /
//     typescript-eslint / eslint-plugin-unicorn / eslint-plugin-react /
//     eslint-plugin-vue rule option schemas, exercising general validation
//     correctness against schema shapes actually shipped by real rules:
//     definitions/$defs + $ref chains, oneOf/anyOf/allOf exclusion, tuple
//     vs. list-style items, additionalItems, additionalProperties-as-schema,
//     minProperties, required, and uniqueItems.
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

// Mirrors ESLint's lib/shared/config-validator.js getRuleOptionsSchema: a
// rule's raw meta.schema, when it's a plain array of item schemas, is
// wrapped into a tuple-style `{type: "array", items: [...]}`. When it's
// already a full schema object, it's used as-is.
function wrap(rawSchema) {
  if (Array.isArray(rawSchema)) {
    if (rawSchema.length === 0) {
      return { type: 'array', minItems: 0, maxItems: 0 };
    }
    return {
      type: 'array',
      items: rawSchema,
      minItems: 0,
      maxItems: rawSchema.length,
    };
  }
  return rawSchema;
}

const SYNTHETIC_CASES = [
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

// ---- Real rule schemas, copied verbatim from real-world rule meta.schema
// definitions (extracted from eslint-core, @typescript-eslint/eslint-plugin,
// eslint-plugin-unicorn, eslint-plugin-react, eslint-plugin-vue). ----

const REALWORLD_SCHEMAS = {
  'eslint-core/comma-dangle': {
    definitions: {
      value: {
        enum: ['always-multiline', 'always', 'never', 'only-multiline'],
      },
      valueWithIgnore: {
        enum: [
          'always-multiline',
          'always',
          'ignore',
          'never',
          'only-multiline',
        ],
      },
    },
    type: 'array',
    items: [
      {
        oneOf: [
          { $ref: '#/definitions/value' },
          {
            type: 'object',
            properties: {
              arrays: { $ref: '#/definitions/valueWithIgnore' },
              objects: { $ref: '#/definitions/valueWithIgnore' },
              imports: { $ref: '#/definitions/valueWithIgnore' },
              exports: { $ref: '#/definitions/valueWithIgnore' },
              functions: { $ref: '#/definitions/valueWithIgnore' },
            },
            additionalProperties: false,
          },
        ],
      },
    ],
    additionalItems: false,
  },
  'eslint-core/func-names': {
    definitions: {
      value: { enum: ['always', 'as-needed', 'never'] },
    },
    type: 'array',
    items: [
      { $ref: '#/definitions/value' },
      {
        type: 'object',
        properties: { generators: { $ref: '#/definitions/value' } },
        additionalProperties: false,
      },
    ],
    additionalItems: false,
  },
  'eslint-core/array-element-newline': {
    definitions: {
      basicConfig: {
        oneOf: [
          { enum: ['always', 'never', 'consistent'] },
          {
            type: 'object',
            properties: {
              multiline: { type: 'boolean' },
              minItems: { type: ['integer', 'null'], minimum: 0 },
            },
            additionalProperties: false,
          },
        ],
      },
    },
    type: 'array',
    items: [
      {
        oneOf: [
          { $ref: '#/definitions/basicConfig' },
          {
            type: 'object',
            properties: {
              ArrayExpression: { $ref: '#/definitions/basicConfig' },
              ArrayPattern: { $ref: '#/definitions/basicConfig' },
            },
            additionalProperties: false,
            minProperties: 1,
          },
        ],
      },
    ],
  },
  'eslint-core/padding-line-between-statements': {
    definitions: {
      paddingType: { enum: ['any', 'never', 'always'] },
      statementType: {
        anyOf: [
          { enum: ['*', 'block-like', 'function', 'if', 'return', 'var'] },
          {
            type: 'array',
            items: {
              enum: ['*', 'block-like', 'function', 'if', 'return', 'var'],
            },
            minItems: 1,
            uniqueItems: true,
          },
        ],
      },
    },
    type: 'array',
    items: {
      type: 'object',
      properties: {
        blankLine: { $ref: '#/definitions/paddingType' },
        prev: { $ref: '#/definitions/statementType' },
        next: { $ref: '#/definitions/statementType' },
      },
      additionalProperties: false,
      required: ['blankLine', 'prev', 'next'],
    },
  },
  '@typescript-eslint/array-type': wrap([
    {
      type: 'object',
      $defs: {
        arrayOption: {
          type: 'string',
          enum: ['array', 'generic', 'array-simple'],
        },
      },
      additionalProperties: false,
      properties: {
        default: { $ref: '#/items/0/$defs/arrayOption' },
        readonly: { $ref: '#/items/0/$defs/arrayOption' },
      },
    },
  ]),
  '@typescript-eslint/ban-ts-comment': wrap([
    {
      type: 'object',
      $defs: {
        directiveConfigSchema: {
          oneOf: [
            { type: 'boolean', default: true },
            { type: 'string', enum: ['allow-with-description'] },
            {
              type: 'object',
              additionalProperties: false,
              properties: { descriptionFormat: { type: 'string' } },
            },
          ],
        },
      },
      additionalProperties: false,
      properties: {
        minimumDescriptionLength: { type: 'number' },
        'ts-check': { $ref: '#/items/0/$defs/directiveConfigSchema' },
        'ts-expect-error': { $ref: '#/items/0/$defs/directiveConfigSchema' },
        'ts-ignore': { $ref: '#/items/0/$defs/directiveConfigSchema' },
        'ts-nocheck': { $ref: '#/items/0/$defs/directiveConfigSchema' },
      },
    },
  ]),
  '@typescript-eslint/no-restricted-types': wrap([
    {
      type: 'object',
      $defs: {
        banConfig: {
          oneOf: [
            { type: 'boolean', enum: [true] },
            { type: 'string' },
            {
              type: 'object',
              additionalProperties: false,
              properties: {
                fixWith: { type: 'string' },
                message: { type: 'string' },
                suggest: { type: 'array', items: { type: 'string' } },
              },
            },
          ],
        },
      },
      additionalProperties: false,
      properties: {
        types: {
          type: 'object',
          additionalProperties: { $ref: '#/items/0/$defs/banConfig' },
        },
      },
    },
  ]),
  '@typescript-eslint/parameter-properties': wrap([
    {
      type: 'object',
      $defs: {
        modifier: {
          type: 'string',
          enum: [
            'readonly',
            'private',
            'protected',
            'public',
            'private readonly',
            'protected readonly',
            'public readonly',
          ],
        },
      },
      additionalProperties: false,
      properties: {
        allow: { type: 'array', items: { $ref: '#/items/0/$defs/modifier' } },
        prefer: {
          type: 'string',
          enum: ['class-property', 'parameter-property'],
        },
      },
    },
  ]),
  'eslint-plugin-unicorn/import-style': {
    type: 'array',
    additionalItems: false,
    items: [
      {
        type: 'object',
        additionalProperties: false,
        properties: {
          checkImport: { type: 'boolean' },
          checkDynamicImport: { type: 'boolean' },
          checkExportFrom: { type: 'boolean' },
          checkRequire: { type: 'boolean' },
          extendDefaultStyles: { type: 'boolean' },
          styles: { $ref: '#/definitions/moduleStyles' },
        },
      },
    ],
    definitions: {
      moduleStyles: {
        type: 'object',
        additionalProperties: { $ref: '#/definitions/styles' },
      },
      styles: {
        anyOf: [{ enum: [false] }, { $ref: '#/definitions/booleanObject' }],
      },
      booleanObject: {
        type: 'object',
        additionalProperties: { type: 'boolean' },
      },
    },
  },
  'eslint-plugin-react/jsx-curly-spacing': {
    definitions: {
      basicConfig: {
        type: 'object',
        properties: {
          when: { enum: ['always', 'never'] },
          allowMultiline: { type: 'boolean' },
          spacing: {
            type: 'object',
            properties: { objectLiterals: { enum: ['always', 'never'] } },
          },
        },
      },
      basicConfigOrBoolean: {
        anyOf: [{ $ref: '#/definitions/basicConfig' }, { type: 'boolean' }],
      },
    },
    type: 'array',
    items: [
      {
        anyOf: [
          {
            allOf: [
              { $ref: '#/definitions/basicConfig' },
              {
                type: 'object',
                properties: {
                  attributes: { $ref: '#/definitions/basicConfigOrBoolean' },
                  children: { $ref: '#/definitions/basicConfigOrBoolean' },
                },
              },
            ],
          },
          { enum: ['always', 'never'] },
        ],
      },
      {
        type: 'object',
        properties: {
          allowMultiline: { type: 'boolean' },
          spacing: {
            type: 'object',
            properties: { objectLiterals: { enum: ['always', 'never'] } },
          },
        },
        additionalProperties: false,
      },
    ],
  },
  'eslint-plugin-vue/html-self-closing': {
    definitions: {
      optionValue: { enum: ['always', 'never', 'any'] },
    },
    type: 'array',
    items: [
      {
        type: 'object',
        properties: {
          html: {
            type: 'object',
            properties: {
              normal: { $ref: '#/definitions/optionValue' },
              void: { $ref: '#/definitions/optionValue' },
              component: { $ref: '#/definitions/optionValue' },
            },
            additionalProperties: false,
          },
          svg: { $ref: '#/definitions/optionValue' },
          math: { $ref: '#/definitions/optionValue' },
        },
        additionalProperties: false,
      },
    ],
    maxItems: 1,
  },
};

// ---- {schema-name, input} cases exercising each schema's distinguishing
// structure: oneOf/anyOf/allOf exclusivity, $ref via definitions vs $defs,
// tuple vs. list-style items, additionalItems, additionalProperties-as-
// schema, minProperties, required, uniqueItems, and (for ban-ts-comment)
// confirming with a real rule that a "default" nested inside oneOf is not
// applied. ----
const REALWORLD_CASES = [
  // comma-dangle: oneOf[$ref-to-enum, object-of-$ref-to-enum], additionalItems:false
  {
    schema: 'eslint-core/comma-dangle',
    name: 'comma_dangle_bare_enum_value',
    input: ['always'],
  },
  {
    schema: 'eslint-core/comma-dangle',
    name: 'comma_dangle_object_form',
    input: [{ arrays: 'always', functions: 'ignore' }],
  },
  {
    schema: 'eslint-core/comma-dangle',
    name: 'comma_dangle_invalid_enum_value',
    input: ['sometimes'],
  },
  {
    schema: 'eslint-core/comma-dangle',
    name: 'comma_dangle_functions_rejects_ignore_is_valid_but_arrays_bad_value_invalid',
    input: [{ arrays: 'sometimes' }],
  },
  {
    schema: 'eslint-core/comma-dangle',
    name: 'comma_dangle_too_many_items_rejected_by_additionalItems',
    input: ['always', 'never'],
  },

  // func-names: 2-item tuple, both $ref to same enum, additionalItems:false
  {
    schema: 'eslint-core/func-names',
    name: 'func_names_both_items_valid',
    input: ['always', { generators: 'never' }],
  },
  {
    schema: 'eslint-core/func-names',
    name: 'func_names_first_item_invalid_enum',
    input: ['sometimes'],
  },
  {
    schema: 'eslint-core/func-names',
    name: 'func_names_extra_third_item_rejected',
    input: ['always', {}, 'oops'],
  },

  // array-element-newline: oneOf reused via $ref, minProperties:1
  {
    schema: 'eslint-core/array-element-newline',
    name: 'array_element_newline_bare_string',
    input: ['always'],
  },
  {
    schema: 'eslint-core/array-element-newline',
    name: 'array_element_newline_per_node_object',
    input: [{ ArrayExpression: 'always', ArrayPattern: { multiline: true } }],
  },
  // Looks like it should hit minProperties:1 on the second oneOf branch, but
  // {} also matches basicConfig's own object branch (all its properties are
  // optional) via the first oneOf branch, so the outer oneOf is still
  // satisfied (exactly one match) and this is valid — minProperties:1 here
  // is unreachable dead weight given the sibling $ref branch.
  {
    schema: 'eslint-core/array-element-newline',
    name: 'array_element_newline_empty_object_matches_basicConfig_not_minProperties',
    input: [{}],
  },
  // An object shaped only for the second branch (a per-node-type key) but
  // with zero of {ArrayExpression, ArrayPattern} present: this shows there's
  // no way to actually reach the minProperties:1 failure, since additionalProperties:false
  // rejects any unrelated key before minProperties is ever checked.
  {
    schema: 'eslint-core/array-element-newline',
    name: 'array_element_newline_unrelated_key_rejected_by_additionalProperties',
    input: [{ NotAKey: 'always' }],
  },

  // padding-line-between-statements: list-style items (not a tuple), required, anyOf
  {
    schema: 'eslint-core/padding-line-between-statements',
    name: 'padding_line_list_style_multiple_entries',
    input: [
      { blankLine: 'always', prev: '*', next: 'return' },
      { blankLine: 'never', prev: ['if', 'block-like'], next: '*' },
    ],
  },
  {
    schema: 'eslint-core/padding-line-between-statements',
    name: 'padding_line_missing_required_field',
    input: [{ blankLine: 'always', prev: '*' }],
  },
  {
    schema: 'eslint-core/padding-line-between-statements',
    name: 'padding_line_statementType_array_must_be_nonempty_and_unique',
    input: [{ blankLine: 'always', prev: ['if', 'if'], next: '*' }],
  },

  // array-type: $defs + $ref (draft-2019-style keyword name, still resolvable as a plain JSON Pointer under a draft-4 compile)
  {
    schema: '@typescript-eslint/array-type',
    name: 'array_type_valid_both_props',
    input: [{ default: 'generic', readonly: 'array-simple' }],
  },
  {
    schema: '@typescript-eslint/array-type',
    name: 'array_type_invalid_enum_value',
    input: [{ default: 'tuple' }],
  },
  {
    schema: '@typescript-eslint/array-type',
    name: 'array_type_unknown_property_rejected',
    input: [{ defaults: 'array' }],
  },

  // ban-ts-comment: real-world confirmation that a "default" nested inside
  // oneOf (directiveConfigSchema's boolean branch) is NOT filled in, even
  // though the property itself ($ref to that oneOf) is entirely absent.
  {
    schema: '@typescript-eslint/ban-ts-comment',
    name: 'ban_ts_comment_empty_object_ts_check_not_defaulted',
    input: [{}],
  },
  {
    schema: '@typescript-eslint/ban-ts-comment',
    name: 'ban_ts_comment_ts_check_true_valid',
    input: [{ 'ts-check': true }],
  },
  {
    schema: '@typescript-eslint/ban-ts-comment',
    name: 'ban_ts_comment_ts_check_description_object_valid',
    input: [{ 'ts-expect-error': { descriptionFormat: '^TODO' } }],
  },
  {
    schema: '@typescript-eslint/ban-ts-comment',
    name: 'ban_ts_comment_ts_check_invalid_string',
    input: [{ 'ts-check': 'nope' }],
  },

  // no-restricted-types: additionalProperties as a $ref'd oneOf schema (dynamic keys)
  {
    schema: '@typescript-eslint/no-restricted-types',
    name: 'no_restricted_types_dynamic_keys_mixed_ban_configs',
    input: [
      {
        types: {
          Foo: 'use Bar instead',
          Object: true,
          Bad: { message: 'no', suggest: ['Good'] },
        },
      },
    ],
  },
  {
    schema: '@typescript-eslint/no-restricted-types',
    name: 'no_restricted_types_invalid_ban_config_value',
    input: [{ types: { Foo: 123 } }],
  },

  // parameter-properties: array items each $ref into an enum defined in $defs
  {
    schema: '@typescript-eslint/parameter-properties',
    name: 'parameter_properties_valid_allow_list',
    input: [
      { allow: ['readonly', 'private readonly'], prefer: 'class-property' },
    ],
  },
  {
    schema: '@typescript-eslint/parameter-properties',
    name: 'parameter_properties_invalid_allow_entry',
    input: [{ allow: ['static'] }],
  },

  // unicorn import-style: nested additionalProperties chain (object -> $ref -> anyOf -> $ref -> object)
  {
    schema: 'eslint-plugin-unicorn/import-style',
    name: 'import_style_nested_dynamic_keys_valid',
    input: [
      { styles: { lodash: { named: true, default: false }, chalk: false } },
    ],
  },
  {
    schema: 'eslint-plugin-unicorn/import-style',
    name: 'import_style_nested_dynamic_keys_invalid_leaf_type',
    input: [{ styles: { lodash: { named: 'yes' } } }],
  },

  // react jsx-curly-spacing: allOf inside anyOf, plus $ref to an anyOf itself (basicConfigOrBoolean)
  {
    schema: 'eslint-plugin-react/jsx-curly-spacing',
    name: 'jsx_curly_spacing_bare_enum_form',
    input: ['always', { allowMultiline: false }],
  },
  {
    schema: 'eslint-plugin-react/jsx-curly-spacing',
    name: 'jsx_curly_spacing_object_form_with_nested_attributes_boolean',
    input: [{ when: 'never', attributes: false, children: { when: 'always' } }],
  },
  {
    schema: 'eslint-plugin-react/jsx-curly-spacing',
    name: 'jsx_curly_spacing_invalid_when_value',
    input: [{ when: 'sometimes' }],
  },

  // vue html-self-closing: nested object of $ref values, maxItems:1
  {
    schema: 'eslint-plugin-vue/html-self-closing',
    name: 'html_self_closing_nested_valid',
    input: [{ html: { normal: 'always', void: 'never' }, svg: 'always' }],
  },
  {
    schema: 'eslint-plugin-vue/html-self-closing',
    name: 'html_self_closing_invalid_nested_enum',
    input: [{ html: { normal: 'sometimes' } }],
  },
  {
    schema: 'eslint-plugin-vue/html-self-closing',
    name: 'html_self_closing_too_many_top_level_items',
    input: [{}, {}],
  },
];

function runCase(schema, name, input, source) {
  const ajv = newEslintAjv();
  const validate = ajv.compile(schema);
  // ajv mutates the data argument in place; keep the original input
  // untouched in the fixture for the Go test to feed in fresh.
  const data = JSON.parse(JSON.stringify(input));
  const valid = validate(data);
  const result = {
    name,
    schema,
    input,
    output: data,
    valid,
    errors: valid ? null : validate.errors,
  };
  if (source) result.source = source;
  return result;
}

function run() {
  const synthetic = SYNTHETIC_CASES.map(({ name, schema, input }) =>
    runCase(schema, name, input),
  );
  const realworld = REALWORLD_CASES.map(
    ({ schema: schemaName, name, input }) => {
      const rawSchema = REALWORLD_SCHEMAS[schemaName];
      if (!rawSchema) throw new Error(`unknown schema ${schemaName}`);
      const schema = wrap(JSON.parse(JSON.stringify(rawSchema)));
      return runCase(schema, name, input, schemaName);
    },
  );
  const results = [...synthetic, ...realworld];

  const outPath = path.join(
    __dirname,
    '../internal/rule/testdata/ajv_fixtures.json',
  );
  fs.writeFileSync(outPath, JSON.stringify(results, null, 2) + '\n');
  console.log(`Wrote ${results.length} cases to ${outPath}`);
}

run();
