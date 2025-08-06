# Rslint Guide

This guide provides comprehensive documentation for using Rslint, a high-performance JavaScript and TypeScript linter written in Go.

## Installation

```bash
# Install via npm
npm install -g @rslint/core

# Install via pnpm
pnpm add -g @rslint/core

# Install via yarn
yarn global add @rslint/core
```

## Basic Usage

### Command Line Interface

```bash
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

Rslint supports multiple configuration formats:

**rslint.json**

```json
[
  {
    // ignore files and folder for linting
    "ignores": ["./files-not-want-lint.ts", "./tests/**/fixtures/**.ts"],
    "languageOptions": {
      "parserOptions": {
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
