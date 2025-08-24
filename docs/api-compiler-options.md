# Compiler Options Override Support

The API mode now supports passing `compilerOptions` to override settings from the TypeScript configuration file.

## Usage

When making a lint request via the API, you can now include a `compilerOptions` field that will override any corresponding options in your `tsconfig.json`:

```json
{
  "kind": "lint",
  "id": 1,
  "data": {
    "files": ["src/**/*.ts"],
    "config": "./rslint.json",
    "compilerOptions": {
      "strict": true,
      "target": 7,
      "noImplicitAny": true,
      "skipLibCheck": true
    },
    "ruleOptions": {
      "no-unused-vars": "error"
    }
  }
}
```

## Notes

- The `compilerOptions` field accepts a map of option name to value
- These options will override any corresponding options from your `tsconfig.json`
- Numeric options (like `target`) should be passed as numbers corresponding to the TypeScript enum values
- Boolean options can be passed as `true`/`false`
- String options should be passed as strings
- Array options should be passed as arrays

## TypeScript Target Values

Common target values:

- `1` = ES3
- `2` = ES5
- `3` = ES2015 (ES6)
- `4` = ES2016
- `5` = ES2017
- `6` = ES2018
- `7` = ES2019
- `8` = ES2020
- `9` = ES2021
- `10` = ES2022
- `99` = ESNext

## Module Values

Common module values:

- `0` = None
- `1` = CommonJS
- `2` = AMD
- `3` = UMD
- `4` = System
- `5` = ES2015 (ES6)
- `6` = ES2020
- `7` = ES2022
- `99` = ESNext
- `100` = Node16
- `199` = NodeNext

## Example

```json
{
  "compilerOptions": {
    "target": 8,
    "module": 1,
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "lib": ["ES2020", "DOM"]
  }
}
```

This will override the corresponding settings in your `tsconfig.json` for the duration of the linting session.
