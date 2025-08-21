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

##### languageOptions.parserOptions.projectService

- **Type:** `boolean`
- **Default:** `false`

Whether to use TypeScript's project service for better performance in large projects.

```jsonc
{
  "languageOptions": {
    "parserOptions": {
      "projectService": true,
      "project": ["./tsconfig.json"],
    },
  },
}
```

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
        "projectService": false,
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

## Monorepo Configuration

For monorepos, you can define multiple configuration entries for different parts of your project:

```jsonc
[
  // Frontend packages
  {
    "files": ["./packages/frontend/**/*.ts", "./packages/frontend/**/*.tsx"],
    "ignores": ["./packages/frontend/dist/**"],
    "languageOptions": {
      "parserOptions": {
        "project": ["./packages/frontend/tsconfig.json"],
      },
    },
    "plugins": ["@typescript-eslint"],
    "rules": {
      "@typescript-eslint/no-unused-vars": "error",
      // Frontend-specific rules
    },
  },

  // Backend packages
  {
    "files": ["./packages/backend/**/*.ts"],
    "ignores": ["./packages/backend/dist/**"],
    "languageOptions": {
      "parserOptions": {
        "project": ["./packages/backend/tsconfig.json"],
      },
    },
    "plugins": ["@typescript-eslint"],
    "rules": {
      "@typescript-eslint/no-unused-vars": "error",
      // Backend-specific rules
    },
  },

  // Test files
  {
    "files": ["**/*.test.ts", "**/*.spec.ts"],
    "languageOptions": {
      "parserOptions": {
        "project": ["./tsconfig.json"],
      },
    },
    "plugins": ["@typescript-eslint"],
    "rules": {
      // More relaxed rules for tests
      "@typescript-eslint/no-unsafe-assignment": "off",
      "@typescript-eslint/no-explicit-any": "off",
    },
  },
]
```

## Migration from ESLint

If you're migrating from ESLint, you can use most of your existing TypeScript-ESLint rule configurations directly:

```jsonc
// Your existing .eslintrc.json rules section
{
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/prefer-const": "warn"
  }
}

// Can be used directly in rslint.json
[
  {
    "plugins": ["@typescript-eslint"],
    "rules": {
      "@typescript-eslint/no-unused-vars": "error",
      "@typescript-eslint/prefer-const": "warn"
    }
  }
]
```

## Performance Tips

For optimal performance:

1. **Use specific TypeScript projects**: Include only the `tsconfig.json` files you need
2. **Ignore unnecessary files**: Use `ignores` to exclude build outputs, dependencies, and generated files
3. **Consider projectService**: For very large projects, enable `projectService: true`
4. **Minimize rule scope**: Use `files` patterns to apply rules only where needed

```jsonc
{
  "languageOptions": {
    "parserOptions": {
      // ✅ Good: Specific projects
      "project": ["./src/tsconfig.json", "./tests/tsconfig.json"],

      // ❌ Avoid: Too broad
      "project": ["./**/tsconfig.json"],
    },
  },
}
```
