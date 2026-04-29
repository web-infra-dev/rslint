# Type Checking

Rslint can perform TypeScript semantic type checking alongside lint diagnostics. This allows you to replace `tsc --noEmit` with a single `rslint --type-check` command.

## Quick Start

```bash
rslint --type-check .
```

When `--type-check` is enabled, TypeScript semantic errors (e.g. type mismatches, missing properties, unresolved modules) are reported together with lint diagnostics in a single unified output.

## Replacing `tsc --noEmit`

Traditionally, TypeScript projects run type checking and linting as two separate steps:

```bash
tsc --noEmit        # type checking
rslint .            # linting
```

With `--type-check`, both can be done in one pass:

```bash
rslint --type-check .
```

This provides several advantages:

- **Single command** — no need to orchestrate two tools in scripts or CI pipelines.
- **Shared TypeScript program** — type checking reuses the same TypeScript program built for type-aware lint rules, so it adds minimal overhead compared to running a separate `tsc` process.
- **Unified output** — type errors and lint errors appear in the same output stream with a consistent format, making it easier to triage issues.

### CI Migration

#### GitHub Actions

Before:

```yaml
steps:
  - run: npx tsc --noEmit
  - run: npx rslint .
```

After:

```yaml
steps:
  - run: npx rslint --type-check .
```

To get inline annotations on pull request diffs, use `--format github`:

```yaml
steps:
  - run: npx rslint --type-check --format github .
```

#### Other CI Environments

```bash
# Lint + type check in one command
npx rslint --type-check .

# With error threshold for warnings
npx rslint --type-check --max-warnings 10 .

# Errors only (cleaner CI logs)
npx rslint --type-check --quiet .
```

## Output

Type errors are reported with their original TypeScript error codes as the rule name, e.g. `TypeScript(TS2322)`. All type errors have severity `error` and cause a non-zero exit code.

Type errors are included in all output formats (`default`, `jsonline`, `github`).

### Default format

```
  TypeScript(TS2322)  — [error] Type 'string' is not assignable to type 'number'.
  ╭─┴──────────( src/utils.ts:3:7 )─────
  │ 2 │  const name = 'hello';
  │ 3 │  const count: number = name;
  │ 4 │
  ╰────────────────────────────────
```

For errors with detailed explanations, TypeScript's message chain is displayed with indentation:

```
  TypeScript(TS2322)  — [error] Type 'B' is not assignable to type 'A'.
    The types of 'x.y.z' are incompatible between these types.
      Type 'number' is not assignable to type 'string'.
  ╭─┴──────────( src/types.ts:3:7 )─────
```

### Summary line

When `--type-check` is enabled, the summary always splits lint errors and type errors:

```
Found 3 lint errors, 2 type errors and 1 warning (linted 42 files in 120ms using 8 threads)
```

## Alignment with `tsc --noEmit`

`--type-check` is designed to produce the same diagnostics as `tsc --noEmit` (and `tsgo --noEmit`) for any given TypeScript program — same error code, same file, same line and column.

The one intentional difference: TypeScript diagnostics that have no source-file anchor (e.g. tsconfig validation messages such as `TS18003` "No inputs were found in config file" or `TS5108` "Option … has been removed") are not reported by `--type-check`, because rslint's output is rendered per file. Run `tsc --noEmit` directly when you need to surface those configuration-level errors.

## Files Without tsconfig Coverage

Files that match your config's `files` patterns but are not included in any tsconfig (e.g. root-level scripts, config files) are still linted with syntax-level rules. However, `--type-check` will **not** report semantic type errors for these files, since reliable type information requires tsconfig coverage.

To enable full type checking for these files, add them to your tsconfig's `include` or create a separate tsconfig that covers them.

## Interaction with Other Flags

| Flag             | Behavior with `--type-check`                                                     |
| ---------------- | -------------------------------------------------------------------------------- |
| `--fix`          | Applies lint auto-fixes as usual. Type errors have no auto-fix.                  |
| `--quiet`        | Suppresses warnings. Type errors (severity `error`) are always shown.            |
| `--format`       | Type errors are rendered in the chosen format (`default`, `jsonline`, `github`). |
| `--max-warnings` | Only counts lint warnings, not type errors (which are always `error`).           |
