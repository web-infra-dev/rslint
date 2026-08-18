# rules

- **Type:** `Record<string, RuleSeverity | readonly [RuleSeverity, ...options]>`
- **RuleSeverity:** `'off' | 'warn' | 'error' | 0 | 1 | 2`

Configures individual rules with a severity level and optional positional options.

| Value          | Description                               |
| -------------- | ----------------------------------------- |
| `"error"`, `2` | Reports as an error; causes non-zero exit |
| `"warn"`, `1`  | Reports as a warning                      |
| `"off"`, `0`   | Disables the rule                         |

Use a severity directly when a rule has no options to configure:

```ts
{
  rules: {
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/require-await': 'off',
  },
}
```

Use an array to pass every item after the severity to the rule as a positional option:

```ts
{
  rules: {
    '@typescript-eslint/array-type': ['warn', { default: 'array-simple' }],
    '@typescript-eslint/no-unused-vars': [
      'error',
      {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_',
      },
    ],
  },
}
```

Invalid severities and rule value shapes are rejected while the configuration is loaded.

## Merging

Later matching config entries override earlier entries. When a later entry changes only the severity, the rule keeps options from the earlier entry. Supplying any positional option in the later array replaces all earlier options.

```ts
export default defineConfig([
  {
    rules: {
      '@typescript-eslint/no-unused-vars': ['error', { args: 'all' }],
    },
  },
  {
    files: ['tests/**'],
    rules: {
      // Keeps { args: 'all' } while changing the severity.
      '@typescript-eslint/no-unused-vars': 'warn',
    },
  },
]);
```

See [Rules & Presets](/config/rules-and-presets) for preset selection, or browse the complete [Rules](/rules/) reference for rule-specific options.
