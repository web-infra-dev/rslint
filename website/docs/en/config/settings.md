# settings

- **Type:** `Record<string, any>`

Provides shared settings to all rules in a matching config entry. Native rules and compatible third-party ESLint rules can read these values from their rule context.

```ts
{
  files: ['**/*.tsx'],
  settings: {
    react: {
      version: 'detect',
    },
    'jsx-a11y': {
      polymorphicPropName: 'as',
    },
  },
}
```

When multiple matching entries provide `settings`, ordinary nested objects are merged recursively. Later arrays and scalar values replace earlier values.

```ts
export default defineConfig([
  {
    settings: {
      react: { version: 'detect', runtime: 'automatic' },
    },
  },
  {
    files: ['legacy/**'],
    settings: {
      // Keeps version and replaces runtime for matching files.
      react: { runtime: 'classic' },
    },
  },
]);
```
