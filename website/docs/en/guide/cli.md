# CLI Reference

## Usage

```bash
rslint [options] [files...]
```

## Options

| Flag                 | Description                                             |
| -------------------- | ------------------------------------------------------- |
| `--init`             | Generate a default config file in the current directory |
| `--config <path>`    | Specify which config file to use                        |
| `--fix`              | Automatically fix problems                              |
| `--format <format>`  | Output format: `default`, `jsonline`, or `github`       |
| `--quiet`            | Report errors only, suppress warnings                   |
| `--max-warnings <n>` | Exit with error if warning count exceeds this number    |
| `--no-color`         | Disable colored output                                  |
| `--force-color`      | Force colored output                                    |
| `--help`, `-h`       | Show help information                                   |

## File Arguments

When file paths are provided, rslint only lints the specified files instead of scanning all files from configuration. Configuration rules still apply to the matched files.

```bash
# Lint specific files
rslint src/index.ts src/utils.ts

# Use with --fix
rslint --fix src/index.ts src/utils.ts
```

When no file arguments are given, rslint scans all files based on the project configuration.

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
