# Rslint Guide

This guide provides basic documentation for using Rslint, a high-performance JavaScript and TypeScript linter written in Go.

## Installation

```bash
# Install via npm
npm install @rslint/core

# Install via pnpm
pnpm add @rslint/core

# Install via yarn
yarn add @rslint/core
```

## Basic Usage

### Command Line Interface

```bash
# create a default.json(optional)
rslint --init

# use default rslint.json
rslint
# use custom configuration file
rslint --config rslint.json

# Lint with auto-fix
rslint --fix .

# Show help
rslint --help
```

### Configuration

**rslint.json**

```json
[
  {
    // ignore files and folders for linting
    "ignores": ["./files-not-want-lint.ts", "./tests/**/fixtures/**.ts"],
    "languageOptions": {
      "parserOptions": {
        // Rslint will lint all files included in your typescript projects defined here
        // support lint multi packages in monorepo
        "project": ["./tsconfig.json", "packages/app1/tsconfig.json"]
      }
    },
    // same configuration as https://typescript-eslint.io/rules/
    "rules": {
      "@typescript-eslint/require-await": "off",
      "@typescript-eslint/no-unnecessary-type-assertion": "warn",
      "@typescript-eslint/array-type": ["warn", { "default": "array-simple" }]
    },
    "plugins": [
      "@typescript-eslint" // will enable all implemented @typescript-eslint rules by default
    ]
  }
]
```

## Editor Integration

### VS Code

Install the extension from the marketplace [rslint](https://marketplace.visualstudio.com/items?itemName=rstack.rslint)
