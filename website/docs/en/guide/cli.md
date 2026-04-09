# CLI Reference

## Usage

```bash
rslint [options] [files/directories...]
```

## Options

| Flag                 | Description                                                                          |
| -------------------- | ------------------------------------------------------------------------------------ |
| `--init`             | Generate a default config file, or migrate an existing JSON config to JS/TS          |
| `--config <path>`    | Specify which config file to use                                                     |
| `--fix`              | Automatically fix problems                                                           |
| `--type-check`       | Enable TypeScript semantic type checking ([details](/guide/type-checking))           |
| `--format <format>`  | Output format: `default`, `jsonline`, or `github` ([details](/guide/output-formats)) |
| `--quiet`            | Report errors only, suppress warnings                                                |
| `--max-warnings <n>` | Exit with error if warning count exceeds this number                                 |
| `--rule <rule>`      | Override a rule's severity or options (repeatable, see [details](#rule-overrides))   |
| `--no-color`         | Disable colored output ([details](/guide/environment-variables))                     |
| `--force-color`      | Force colored output ([details](/guide/environment-variables))                       |
| `--help`, `-h`       | Show help information                                                                |

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
