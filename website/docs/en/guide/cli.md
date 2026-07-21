# CLI Reference

## Usage

```bash
rslint [options] [files/directories...]
```

## Options

| Flag                                                     | Description                                                                                                           |
| -------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `--init`                                                 | Generate a default config file, or migrate an existing JSON config to JS/TS                                           |
| `-c, --config <path>`                                    | Specify which config file to use                                                                                      |
| `--fix`                                                  | Automatically fix problems                                                                                            |
| `--type-check`                                           | Enable TypeScript semantic type checking ([details](/guide/type-checking))                                            |
| `--type-check-only`                                      | Run TypeScript semantic type checking without lint rules ([details](/guide/type-checking))                            |
| `--format <format>`                                      | Output format: `default`, `jsonline`, `github`, or `gitlab` ([details](/guide/output-formats))                        |
| `--quiet`                                                | Report errors only, suppress warnings                                                                                 |
| `--max-warnings <n>`                                     | Exit with error if warning count exceeds this number                                                                  |
| `--rule <rule>`                                          | Override a rule's severity or options (repeatable, see [details](#rule-overrides))                                    |
| `--report-unused-disable-directives`                     | Report unused `eslint-disable` directives as warnings ([details](#report-unused-disable-directives))                  |
| `--report-unused-disable-directives-severity <severity>` | Report unused `eslint-disable` directives at `off`, `warn`, or `error` ([details](#report-unused-disable-directives)) |
| `--no-color`                                             | Disable colored output ([details](/guide/environment-variables))                                                      |
| `--force-color`                                          | Force colored output ([details](/guide/environment-variables))                                                        |
| `--help`, `-h`                                           | Show help information                                                                                                 |

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

## Report Unused Disable Directives

Use `--report-unused-disable-directives` (or the finer-grained `--report-unused-disable-directives-severity`) to flag `eslint-disable` / `eslint-disable-line` / `eslint-disable-next-line` comments (and their `rslint-` equivalents) that didn't actually suppress any diagnostic.

```bash
# Report unused directives as warnings
rslint --report-unused-disable-directives

# Report unused directives at a specific severity
rslint --report-unused-disable-directives-severity error
rslint --report-unused-disable-directives-severity warn
rslint --report-unused-disable-directives-severity off
```

This can also be configured in `linterOptions.reportUnusedDisableDirectives` — see [Configuration Options](/config/#linteroptions).

**Behavior:**

- `--report-unused-disable-directives` reports unused directives as `warn`. For a specific severity, use `--report-unused-disable-directives-severity` instead — the two flags are mutually exclusive.
- A CLI flag (either form) takes precedence over the config file's `linterOptions.reportUnusedDisableDirectives`.
- `eslint-enable` directives are never reported as unused — only directives that disable a rule are checked.
- Only Rslint's native (Go) rules are tracked. Directives that reference third-party ESLint plugin rules (dispatched to the Node plugin worker, see [ESLint plugin compatibility](/guide/eslint-plugins)) are not checked and are never reported as unused.
- Reporting at `error` severity causes a non-zero exit code, same as any other error-level diagnostic.

## Exit Codes

| Code | Meaning                                           |
| ---- | ------------------------------------------------- |
| `0`  | No errors (warnings may be present)               |
| `1`  | Errors found, or warnings exceed `--max-warnings` |
| `2`  | Invalid command-line usage or flag combinations   |
