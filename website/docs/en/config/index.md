# Configuration

Rslint uses a configuration file to define linting rules and behavior. This page describes all the configuration options available.

## Configuration File

Rslint looks for configuration files in the following order:

- `rslint.json`
- `rslint.jsonc` (JSON with comments)

You can also specify a custom configuration file using the `--config` option:

```bash
rslint --config custom-config.json
```

### Creating a Configuration File

To create a default configuration file, run:

```bash
rslint --init
```

This creates a `rslint.jsonc` file with sensible defaults.

## Configuration Format

The configuration file contains an array of configuration entries. Each entry defines rules and options for matching files:

```jsonc
[
  {
    "ignores": ["./dist/**", "./node_modules/**"],
    "languageOptions": {
      "parserOptions": {
        "project": ["./tsconfig.json"],
      },
    },
    "rules": {
      "@typescript-eslint/no-unused-vars": "error",
      "@typescript-eslint/array-type": ["warn", { "default": "array-simple" }],
    },
    "plugins": ["@typescript-eslint"],
  },
]
```

## Configuration Options

### ignores

- **Type:** `string[]`
- **Default:** `[]`

An array of glob patterns for files and directories to ignore during linting.

```jsonc
{
  "ignores": [
    "./dist/**",
    "./build/**",
    "./node_modules/**",
    "**/*.d.ts",
    "./tests/**/fixtures/**",
  ],
}
```

Patterns support:

- **Glob patterns**: `*.js`, `**/*.ts`
- **Directory patterns**: `dist/**`, `node_modules/**`
- **Negation**: `!important.ts` (when used with other patterns)

### languageOptions

- **Type:** `object`
- **Default:** `{}`

Language-specific configuration options.

#### languageOptions.parserOptions

- **Type:** `object`
- **Default:** `{}`

Parser configuration for TypeScript analysis.

##### languageOptions.parserOptions.project

- **Type:** `string[]`
- **Default:** `["./tsconfig.json"]`

Array of TypeScript project configuration files. Rslint will lint all files included in these TypeScript projects.

```jsonc
{
  "languageOptions": {
    "parserOptions": {
      "project": [
        "./tsconfig.json",
        "./packages/*/tsconfig.json",
        "./apps/*/tsconfig.json",
      ],
    },
  },
}
```

This is especially useful for monorepos where you have multiple TypeScript projects.

### rules

- **Type:** `object`
- **Default:** `{}`

Configuration for linting rules. Rules can be configured in several formats:

#### String Format

```jsonc
{
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/prefer-const": "warn",
    "@typescript-eslint/no-explicit-any": "off",
  },
}
```

Valid severity levels:

- `"error"` - Rule violations cause linting to fail
- `"warn"` - Rule violations produce warnings
- `"off"` - Rule is disabled

#### Array Format

For rules that accept configuration options:

```jsonc
{
  "rules": {
    "@typescript-eslint/array-type": ["error", { "default": "array-simple" }],
    "@typescript-eslint/no-unused-vars": [
      "warn",
      {
        "vars": "all",
        "args": "after-used",
        "ignoreRestSiblings": false,
      },
    ],
  },
}
```

#### Object Format

Alternative object-based configuration:

```jsonc
{
  "rules": {
    "@typescript-eslint/no-unused-vars": {
      "level": "error",
      "options": {
        "vars": "all",
        "args": "after-used",
      },
    },
  },
}
```

### plugins

- **Type:** `string[]`
- **Default:** `[]`

Array of plugins to enable. When a plugin is enabled, all its implemented rules are automatically available with default configurations.

```jsonc
{
  "plugins": ["@typescript-eslint", "eslint-plugin-import"],
}
```

#### Available Plugins

##### @typescript-eslint

Enables TypeScript-specific linting rules. This is the most commonly used plugin for TypeScript projects.

```jsonc
{
  "plugins": ["@typescript-eslint"],
  "rules": {
    // TypeScript-specific rules are now available
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/prefer-const": "warn",
  },
}
```

##### eslint-plugin-import

Enables import/export related rules for better module management.

```jsonc
{
  "plugins": ["eslint-plugin-import"],
  "rules": {
    // Import rules are now available
    "import/no-unresolved": "error",
    "import/order": "warn",
  },
}
```

## Complete Example

Here's a comprehensive configuration example for a typical TypeScript project:

```jsonc
[
  {
    // Ignore build outputs and dependencies
    "ignores": [
      "./dist/**",
      "./build/**",
      "./node_modules/**",
      "./coverage/**",
      "**/*.d.ts",
      "./tests/**/fixtures/**",
    ],

    // TypeScript project configuration
    "languageOptions": {
      "parserOptions": {
        "project": ["./tsconfig.json", "./packages/*/tsconfig.json"],
      },
    },

    // Enable TypeScript plugin
    "plugins": ["@typescript-eslint"],

    // Rule configuration
    "rules": {
      // Variable and import rules
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          "vars": "all",
          "args": "after-used",
          "argsIgnorePattern": "^_",
          "varsIgnorePattern": "^_",
        },
      ],

      // Type safety rules
      "@typescript-eslint/no-unsafe-argument": "error",
      "@typescript-eslint/no-unsafe-assignment": "error",
      "@typescript-eslint/no-unsafe-call": "error",
      "@typescript-eslint/no-unsafe-member-access": "error",
      "@typescript-eslint/no-unsafe-return": "error",
      "@typescript-eslint/await-thenable": "error",

      // Code style rules
      "@typescript-eslint/array-type": ["warn", { "default": "array-simple" }],
      "@typescript-eslint/prefer-const": "error",
      "@typescript-eslint/no-unnecessary-type-assertion": "warn",

      // Async/Promise rules
      "@typescript-eslint/no-floating-promises": [
        "error",
        { "ignoreVoid": true },
      ],
      "@typescript-eslint/require-await": "warn",
      "@typescript-eslint/return-await": ["error", "always"],

      // Best practices
      "@typescript-eslint/no-empty-function": [
        "error",
        { "allow": ["constructors"] },
      ],
      "@typescript-eslint/no-empty-interface": "error",
      "@typescript-eslint/prefer-as-const": "error",
    },
  },
]
```
