# CLI Reference

## Usage

```bash
rslint [options] [files/directories...]
```

## Options

| Flag                 | Description                                                                        |
| -------------------- | ---------------------------------------------------------------------------------- |
| `--init`             | Generate a default config file, or migrate an existing JSON config to JS/TS        |
| `--config <path>`    | Specify which config file to use                                                   |
| `--fix`              | Automatically fix problems                                                         |
| `--type-check`       | Enable TypeScript semantic type checking ([details](/guide/type-checking))         |
| `--format <format>`  | Output format: `default`, `jsonline`, or `github`                                  |
| `--quiet`            | Report errors only, suppress warnings                                              |
| `--max-warnings <n>` | Exit with error if warning count exceeds this number                               |
| `--rule <rule>`      | Override a rule's severity or options (repeatable, see [details](#rule-overrides)) |
| `--no-color`         | Disable colored output                                                             |
| `--force-color`      | Force colored output                                                               |
| `--help`, `-h`       | Show help information                                                              |

## File and Directory Arguments

You can pass file paths, directory paths, or a mix of both. Rslint discovers the config file by walking upward from the target location.

```bash
# Lint specific files
rslint src/index.ts src/utils.ts

# Lint a directory (only files under that directory are linted)
rslint src/

# Mix files and directories
rslint src/ lib/utils.ts

# Use with --fix
rslint --fix src/index.ts
```

When no arguments are given, rslint scopes linting to the current working directory.

### Config Discovery

For each target file or directory, rslint searches for `rslint.config.{js,mjs,ts,mts}` starting from that location and walking upward to the filesystem root. The first config found is used.

In monorepo setups, rslint automatically discovers nested configs and applies the nearest one to each file:

```bash
# Lint from monorepo root (discovers all sub-package configs)
rslint

# Lint a specific package
rslint packages/foo/

# Lint files from different packages (each uses its nearest config)
rslint packages/foo/src/a.ts packages/bar/src/b.ts
```

Use `--config` to override automatic config discovery:

```bash
rslint --config custom.config.ts src/
```

## Output Formats

### default

Human-readable terminal output with colored code snippets and diagnostic highlighting.

```bash
rslint .
```

```
src/index.ts:5:7
  error  @typescript-eslint/no-unused-vars  'foo' is declared but its value is never read.

Found 1 error and 0 warnings (linted 12 files in 42ms using 8 threads)
```

With `--type-check`, type errors are also included (see [Type Checking](/guide/type-checking) for details):

```bash
rslint --type-check .
```

```
src/index.ts:5:7
  error  @typescript-eslint/no-unused-vars  'foo' is declared but its value is never read.

src/utils.ts:3:7
  error  TypeScript(TS2322)  Type 'string' is not assignable to type 'number'.

Found 1 lint error, 1 type error and 0 warnings (linted 12 files in 85ms using 8 threads)
```

### jsonline

One diagnostic per line as compact JSON. Suitable for programmatic consumption.

```bash
rslint --format jsonline .
```

### github

GitHub Actions workflow command format. Creates annotations directly on pull request diffs.

```bash
rslint --format github .
```

## Type Checking

Use `--type-check` to enable TypeScript semantic type checking alongside lint diagnostics. This can fully replace `tsc --noEmit` in your workflow.

```bash
rslint --type-check .
```

For details on output format, CI migration, and how it works, see the [Type Checking](/guide/type-checking) page.

## Rule Overrides

Use `--rule` to override a rule's severity or options from the command line, without modifying your config file. This is useful for quick debugging, CI one-offs, or temporarily enabling/disabling rules.

```bash
# Override severity
rslint --rule 'no-console: off'
rslint --rule 'no-debugger: error'
rslint --rule 'no-debugger: warn'

# Override severity with options (JSON array format)
rslint --rule 'no-console: ["error", {"allow": ["warn", "error"]}]'

# Plugin rules
rslint --rule '@typescript-eslint/no-explicit-any: off'

# Multiple overrides
rslint --rule 'no-console: off' --rule 'no-debugger: error'
```

`--rule` can appear anywhere in the argument list — before or after file paths and other flags:

```bash
rslint --rule 'no-console: off' src/
rslint src/ --rule 'no-console: off'
rslint src/ --rule 'no-console: off' --format github
```

**Behavior:**

- CLI rules have the **highest precedence** and override all config file entries, including per-file overrides.
- When the same rule is specified multiple times, the **last one wins**.
- Rules that don't exist in the registry are silently ignored.

## Exit Codes

| Code | Meaning                                           |
| ---- | ------------------------------------------------- |
| `0`  | No errors (warnings may be present)               |
| `1`  | Errors found, or warnings exceed `--max-warnings` |

## CI Integration

### GitHub Actions

Use `--format github` to get inline annotations on pull request diffs:

```yaml
name: Lint
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - run: npm ci
      - run: npx rslint --format github .
```

Rslint auto-detects the `GITHUB_ACTIONS` environment variable and enables colored output in CI.

### Other CI Environments

```bash
# Fail on any error
npx rslint .

# Lint with type checking
npx rslint --type-check .

# Fail on errors or if warnings exceed threshold
npx rslint --max-warnings 10 .

# Report errors only (cleaner CI logs)
npx rslint --quiet .
```

## Environment Variables

| Variable      | Description            |
| ------------- | ---------------------- |
| `NO_COLOR`    | Disable colored output |
| `FORCE_COLOR` | Force colored output   |
