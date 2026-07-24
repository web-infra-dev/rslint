# CLI Reference

## Usage

```bash
rslint [options] [files/directories...]
```

## Options

| Flag                  | Description                                                                                    |
| --------------------- | ---------------------------------------------------------------------------------------------- |
| `--init`              | Generate a default config file, or migrate an existing JSON config to JS/TS                    |
| `-c, --config <path>` | Specify which config file to use                                                               |
| `--fix`               | Automatically fix problems                                                                     |
| `--type-check`        | Enable TypeScript semantic type checking ([details](/guide/type-checking))                     |
| `--type-check-only`   | Run TypeScript semantic type checking without lint rules ([details](/guide/type-checking))     |
| `--format <format>`   | Output format: `default`, `jsonline`, `github`, or `gitlab` ([details](/guide/output-formats)) |
| `--quiet`             | Report errors only, suppress warnings                                                          |
| `--timing`            | Print a per-rule timing table after the run (see [details](#rule-timing))                      |
| `--timing-top <n>`    | Print only the top N rules in the timing table (implies `--timing`)                            |
| `--max-warnings <n>`  | Exit with error if warning count exceeds this number                                           |
| `--rule <rule>`       | Override a rule's severity or options (repeatable, see [details](#rule-overrides))             |
| `--no-color`          | Disable colored output ([details](/guide/environment-variables))                               |
| `--force-color`       | Force colored output ([details](/guide/environment-variables))                                 |
| `--help`, `-h`        | Show help information                                                                          |

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

For each target file or directory, rslint searches for `rslint.config.{js,mjs,ts,mts}` starting from that location and walking upward to the filesystem root. The nearest successfully loaded config is used. If a discovered config cannot be loaded, rslint warns and tries its next ancestor; the command fails when none of the discovered candidates can be loaded.

`.cjs` and `.cts` config files are not discovered automatically, but remain supported when passed explicitly with `--config` or `-c`.

In monorepo setups, rslint automatically discovers nested configs and applies the nearest one to each file:

```bash
# Lint from monorepo root (discovers all sub-package configs)
rslint

# Lint a specific package
rslint packages/foo/

# Lint files from different packages (each uses its nearest config)
rslint packages/foo/src/a.ts packages/bar/src/b.ts
```

Use `--config` or `-c` to override automatic config discovery:

```bash
rslint --config custom.config.ts src/
rslint -c custom.config.ts src/
```

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

## Rule Timing

Use `--timing` to print a per-rule timing table after the run, sorted by total time. Use `--timing-top` to keep only the top N rules (it implies `--timing`):

```bash
rslint --timing src/
rslint --timing-top 10 src/
```

```
Rule                                    | Source | Time (ms) | Files | Relative
----------------------------------------|--------|-----------|-------|---------
@typescript-eslint/no-misused-promises  | native |    1203.5 |   842 |    31.2%
@typescript-eslint/no-floating-promises | native |     801.2 |   842 |    20.8%
jsdoc/no-types                          | js     |     311.4 |   842 |     8.1%
no-control-regex                        | native |     102.9 |   842 |     2.7%
```

- **Source** — `native` for built-in Go rules, `js` for rules run through the ESLint plugin compatibility layer.
- **Time (ms)** — total time spent in the rule across all files: building its listeners plus running them during AST traversal, including diagnostic and fix construction.
- **Files** — number of files the rule executed on.
- **Relative** — the rule's share of the summed rule time.

The table is written to stderr, so machine-readable output formats such as `jsonline` stay parseable. Files are linted by parallel workers, so summed rule time can exceed the run's wall-clock time. With `--fix`, times accumulate across all re-lint passes. Rules executed through the ESLint plugin compatibility layer are included: their time is measured inside the Node.js worker (rule `create` plus listener invocations), excluding parse and IPC overhead.

## Exit Codes

| Code | Meaning                                           |
| ---- | ------------------------------------------------- |
| `0`  | No errors (warnings may be present)               |
| `1`  | Errors found, or warnings exceed `--max-warnings` |
| `2`  | Invalid command-line usage or flag combinations   |
