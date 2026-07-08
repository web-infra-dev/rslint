// scripts/gen-ajv-realworld-fixtures.js
//
// Generates internal/rule/testdata/ajv_realworld_fixtures.json: a corpus of
// {schema, input} pairs, sourced from real ESLint / typescript-eslint /
// eslint-plugin-unicorn / eslint-plugin-react / eslint-plugin-vue rule
// option schemas (definitions/$defs + $ref chains, oneOf/anyOf/allOf,
// tuple vs. list-style items, additionalProperties-as-schema, etc.), run
// through ajv@6 configured exactly the way ESLint configures it. This
// exercises internal/rule's schema.go against real-world schema shapes,
// not just the hand-built synthetic cases in ajv_defaults_fixtures.json.
const fs = require('fs');
const path = require('path');
const Ajv = require('ajv');
const metaSchema = require('ajv/lib/refs/json-schema-draft-04.json');

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

// ---- Real rule schemas, copied verbatim from real-world rule meta.schema
// definitions (extracted from eslint-core, @typescript-eslint/eslint-plugin,
// eslint-plugin-unicorn, eslint-plugin-react, eslint-plugin-vue). ----

const schemas = {
  'eslint-core/comma-dangle': {
    definitions: {
      value: { enum: ['always-multiline', 'always', 'never', 'only-multiline'] },
      valueWithIgnore: {
        enum: ['always-multiline', 'always', 'ignore', 'never', 'only-multiline'],
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
            items: { enum: ['*', 'block-like', 'function', 'if', 'return', 'var'] },
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
        arrayOption: { type: 'string', enum: ['array', 'generic', 'array-simple'] },
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
            'readonly', 'private', 'protected', 'public',
            'private readonly', 'protected readonly', 'public readonly',
          ],
        },
      },
      additionalProperties: false,
      properties: {
        allow: { type: 'array', items: { $ref: '#/items/0/$defs/modifier' } },
        prefer: { type: 'string', enum: ['class-property', 'parameter-property'] },
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
const cases = [
  // comma-dangle: oneOf[$ref-to-enum, object-of-$ref-to-enum], additionalItems:false
  { schema: 'eslint-core/comma-dangle', name: 'comma_dangle_bare_enum_value', input: ['always'] },
  { schema: 'eslint-core/comma-dangle', name: 'comma_dangle_object_form', input: [{ arrays: 'always', functions: 'ignore' }] },
  { schema: 'eslint-core/comma-dangle', name: 'comma_dangle_invalid_enum_value', input: ['sometimes'] },
  { schema: 'eslint-core/comma-dangle', name: 'comma_dangle_functions_rejects_ignore_is_valid_but_arrays_bad_value_invalid', input: [{ arrays: 'sometimes' }] },
  { schema: 'eslint-core/comma-dangle', name: 'comma_dangle_too_many_items_rejected_by_additionalItems', input: ['always', 'never'] },

  // func-names: 2-item tuple, both $ref to same enum, additionalItems:false
  { schema: 'eslint-core/func-names', name: 'func_names_both_items_valid', input: ['always', { generators: 'never' }] },
  { schema: 'eslint-core/func-names', name: 'func_names_first_item_invalid_enum', input: ['sometimes'] },
  { schema: 'eslint-core/func-names', name: 'func_names_extra_third_item_rejected', input: ['always', {}, 'oops'] },

  // array-element-newline: oneOf reused via $ref, minProperties:1
  { schema: 'eslint-core/array-element-newline', name: 'array_element_newline_bare_string', input: ['always'] },
  { schema: 'eslint-core/array-element-newline', name: 'array_element_newline_per_node_object', input: [{ ArrayExpression: 'always', ArrayPattern: { multiline: true } }] },
  // Looks like it should hit minProperties:1 on the second oneOf branch, but
  // {} also matches basicConfig's own object branch (all its properties are
  // optional) via the first oneOf branch, so the outer oneOf is still
  // satisfied (exactly one match) and this is valid — minProperties:1 here
  // is unreachable dead weight given the sibling $ref branch.
  { schema: 'eslint-core/array-element-newline', name: 'array_element_newline_empty_object_matches_basicConfig_not_minProperties', input: [{}] },
  // An object shaped only for the second branch (a per-node-type key) but
  // with zero of {ArrayExpression, ArrayPattern} present: this shows there's
  // no way to actually reach the minProperties:1 failure, since additionalProperties:false
  // rejects any unrelated key before minProperties is ever checked.
  { schema: 'eslint-core/array-element-newline', name: 'array_element_newline_unrelated_key_rejected_by_additionalProperties', input: [{ NotAKey: 'always' }] },

  // padding-line-between-statements: list-style items (not a tuple), required, anyOf
  { schema: 'eslint-core/padding-line-between-statements', name: 'padding_line_list_style_multiple_entries', input: [{ blankLine: 'always', prev: '*', next: 'return' }, { blankLine: 'never', prev: ['if', 'block-like'], next: '*' }] },
  { schema: 'eslint-core/padding-line-between-statements', name: 'padding_line_missing_required_field', input: [{ blankLine: 'always', prev: '*' }] },
  { schema: 'eslint-core/padding-line-between-statements', name: 'padding_line_statementType_array_must_be_nonempty_and_unique', input: [{ blankLine: 'always', prev: ['if', 'if'], next: '*' }] },

  // array-type: $defs + $ref (draft-2019-style keyword name, still resolvable as a plain JSON Pointer under a draft-4 compile)
  { schema: '@typescript-eslint/array-type', name: 'array_type_valid_both_props', input: [{ default: 'generic', readonly: 'array-simple' }] },
  { schema: '@typescript-eslint/array-type', name: 'array_type_invalid_enum_value', input: [{ default: 'tuple' }] },
  { schema: '@typescript-eslint/array-type', name: 'array_type_unknown_property_rejected', input: [{ defaults: 'array' }] },

  // ban-ts-comment: real-world confirmation that a "default" nested inside
  // oneOf (directiveConfigSchema's boolean branch) is NOT filled in, even
  // though the property itself ($ref to that oneOf) is entirely absent.
  { schema: '@typescript-eslint/ban-ts-comment', name: 'ban_ts_comment_empty_object_ts_check_not_defaulted', input: [{}] },
  { schema: '@typescript-eslint/ban-ts-comment', name: 'ban_ts_comment_ts_check_true_valid', input: [{ 'ts-check': true }] },
  { schema: '@typescript-eslint/ban-ts-comment', name: 'ban_ts_comment_ts_check_description_object_valid', input: [{ 'ts-expect-error': { descriptionFormat: '^TODO' } }] },
  { schema: '@typescript-eslint/ban-ts-comment', name: 'ban_ts_comment_ts_check_invalid_string', input: [{ 'ts-check': 'nope' }] },

  // no-restricted-types: additionalProperties as a $ref'd oneOf schema (dynamic keys)
  { schema: '@typescript-eslint/no-restricted-types', name: 'no_restricted_types_dynamic_keys_mixed_ban_configs', input: [{ types: { Foo: 'use Bar instead', Object: true, Bad: { message: 'no', suggest: ['Good'] } } }] },
  { schema: '@typescript-eslint/no-restricted-types', name: 'no_restricted_types_invalid_ban_config_value', input: [{ types: { Foo: 123 } }] },

  // parameter-properties: array items each $ref into an enum defined in $defs
  { schema: '@typescript-eslint/parameter-properties', name: 'parameter_properties_valid_allow_list', input: [{ allow: ['readonly', 'private readonly'], prefer: 'class-property' }] },
  { schema: '@typescript-eslint/parameter-properties', name: 'parameter_properties_invalid_allow_entry', input: [{ allow: ['static'] }] },

  // unicorn import-style: nested additionalProperties chain (object -> $ref -> anyOf -> $ref -> object)
  { schema: 'eslint-plugin-unicorn/import-style', name: 'import_style_nested_dynamic_keys_valid', input: [{ styles: { lodash: { named: true, default: false }, chalk: false } }] },
  { schema: 'eslint-plugin-unicorn/import-style', name: 'import_style_nested_dynamic_keys_invalid_leaf_type', input: [{ styles: { lodash: { named: 'yes' } } }] },

  // react jsx-curly-spacing: allOf inside anyOf, plus $ref to an anyOf itself (basicConfigOrBoolean)
  { schema: 'eslint-plugin-react/jsx-curly-spacing', name: 'jsx_curly_spacing_bare_enum_form', input: ['always', { allowMultiline: false }] },
  { schema: 'eslint-plugin-react/jsx-curly-spacing', name: 'jsx_curly_spacing_object_form_with_nested_attributes_boolean', input: [{ when: 'never', attributes: false, children: { when: 'always' } }] },
  { schema: 'eslint-plugin-react/jsx-curly-spacing', name: 'jsx_curly_spacing_invalid_when_value', input: [{ when: 'sometimes' }] },

  // vue html-self-closing: nested object of $ref values, maxItems:1
  { schema: 'eslint-plugin-vue/html-self-closing', name: 'html_self_closing_nested_valid', input: [{ html: { normal: 'always', void: 'never' }, svg: 'always' }] },
  { schema: 'eslint-plugin-vue/html-self-closing', name: 'html_self_closing_invalid_nested_enum', input: [{ html: { normal: 'sometimes' } }] },
  { schema: 'eslint-plugin-vue/html-self-closing', name: 'html_self_closing_too_many_top_level_items', input: [{}, {}] },
];

function run() {
  const results = cases.map(({ schema: schemaName, name, input }) => {
    const rawSchema = schemas[schemaName];
    if (!rawSchema) throw new Error(`unknown schema ${schemaName}`);
    const schema = wrap(JSON.parse(JSON.stringify(rawSchema)));
    const ajv = newEslintAjv();
    const validate = ajv.compile(schema);
    const data = JSON.parse(JSON.stringify(input));
    const valid = validate(data);
    return {
      name,
      source: schemaName,
      schema,
      input,
      output: data,
      valid,
      errors: valid ? null : validate.errors,
    };
  });

  const outPath = path.join(
    __dirname,
    '../internal/rule/testdata/ajv_realworld_fixtures.json',
  );
  fs.writeFileSync(outPath, JSON.stringify(results, null, 2) + '\n');
  console.log(`Wrote ${results.length} cases to ${outPath}`);
}

run();
