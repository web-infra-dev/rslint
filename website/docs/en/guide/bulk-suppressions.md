---
description: 'Enable lint rules on an existing project by recording error counts per file and rule, checking them in CI, and pruning resolved violations.'
---

# Bulk suppressions

Bulk suppressions let you enable stricter lint rules on an existing project without fixing every legacy violation up front. Rslint records the current error count per file and rule in a JSON file (default `rslint-suppressions.json`), using the same workflow and format as [ESLint's bulk suppressions](https://eslint.org/docs/latest/use/suppressions).

Existing violations up to the recorded count are suppressed, while new violations fail linting as usual.

## Record existing errors

1. Enable the rules you want to adopt at `"error"` severity in your Rslint config.
2. Run Rslint from your project root to record current violations:

```bash
# Record specific rules
rslint --suppress-rule no-var --suppress-rule no-debugger
rslint --suppress-rule "no-var,no-debugger"
rslint --suppress-rule "[no-var,no-debugger]"

# Or record all current error-level rules
rslint --suppress-all

# Apply autofixes first, then record the remaining errors
rslint --fix --suppress-all
```

3. Commit the generated `rslint-suppressions.json` alongside your config.

After that, running `rslint` from the same project directory automatically reads `rslint-suppressions.json` and suppresses the recorded errors.

## How the suppressions file works

Each entry stores the allowed error count for a file and rule:

```json
{
  "src/app.js": {
    "no-debugger": {
      "count": 2
    }
  }
}
```

- **New errors are caught**: if `src/app.js` grows from 2 to 3 `no-debugger` errors, Rslint reports all 3 errors and exits with code `1`.
- **Line moves don't break suppressions**: moving the 2 existing `no-debugger` calls to other lines keeps the count at 2.
- **Paths are relative to the working directory**: run `rslint` from the same project directory locally and in CI so file paths match.

## Remove unused suppressions

When you fix legacy violations, the recorded count becomes higher than the actual error count. Rslint still suppresses the remaining errors, but exits with code `2` to remind you to update the baseline.

Run `--prune-suppressions` to lower counts and remove entries that are no longer needed:

```bash
rslint --prune-suppressions
```

Pruning lowers counts for fixed violations, deletes rules that reach zero, and removes entries for files that no longer exist on disk. Run it from a full checkout so files outside the current check are not treated as deleted.

If you want to defer cleanup temporarily (for example, while iterating locally), pass `--pass-on-unpruned-suppressions`:

```bash
rslint --pass-on-unpruned-suppressions
```

## Custom location and ESLint migration

Use `--suppressions-location` to store the suppressions file somewhere else, or to reuse an existing ESLint suppressions file directly:

```bash
# Reuse an ESLint suppressions file
rslint --suppressions-location eslint-suppressions.json

# Store the suppressions file in a custom file or directory
rslint --suppress-all --suppressions-location config/suppressions.json
rslint --suppress-all --suppressions-location .cache/suppressions/
```

Pass `--suppressions-location` on every subsequent run that should use that file. When you pass an existing directory (or a path ending with `/` or `\`), Rslint stores a project-specific `rslint-suppressions_<hash>` file inside it.

## Options

| Flag                              | Description                                                                      |
| --------------------------------- | -------------------------------------------------------------------------------- |
| `--suppress-all`                  | Record all current error-level lint violations                                   |
| `--suppress-rule <rules>`         | Record error-level violations for specific rules (repeatable, comma/list)        |
| `--suppressions-location <path>`  | Use a custom suppressions file or directory (default `rslint-suppressions.json`) |
| `--prune-suppressions`            | Lower counts and remove unused entries from the suppressions file                |
| `--pass-on-unpruned-suppressions` | Do not fail with exit code `2` when unused suppressions remain                   |

## Notes

- **Only lint errors are suppressed**: rules configured as `"warn"` are still reported and still count toward `--max-warnings`. Syntax errors and TypeScript compiler diagnostics (`--type-check`) are always reported.
- **One mode per run**: choose at most one of `--suppress-all`, `--suppress-rule`, or `--prune-suppressions` in a single invocation.
- **`--type-check-only`**: skips the suppressions file and cannot be combined with `--suppress-all`, `--suppress-rule`, or `--prune-suppressions`.
- **Safe on incomplete runs**: ignored files, files with syntax errors, and rules or files that could not finish checking (such as during a config or plugin error) keep their existing suppression counts instead of being cleared unexpectedly.
- **Avoid concurrent writes to the same file**: do not run multiple recording or pruning processes against the same suppressions file at the same time. Run Rslint once from the project root, or give each parallel job its own `--suppressions-location`.
