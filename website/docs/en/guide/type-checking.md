# Type Checking

Rslint runs TypeScript semantic type-check alongside or instead of lint rules — a drop-in replacement for `tsc --noEmit` in CI.

- `--type-check` — lint rules **and** type-check, in one pass.
- `--type-check-only` — type-check only; lint phase is skipped entirely.

## Quick start

Point rslint at your tsconfig(s) via `languageOptions.parserOptions.project`:

```js
// rslint.config.mjs
export default [
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: { project: ['./tsconfig.json'] },
    },
  },
];
```

Then:

```bash
rslint --type-check .         # lint + type-check
rslint --type-check-only .    # type-check only
```

When `parserOptions.project` is omitted, rslint uses `tsconfig.json` in the governing config directory when present. If neither configured projects nor that fallback tsconfig exist, no real TypeScript Program is built and type-check produces no diagnostics for that config.

## What gets type-checked

`parserOptions.project` accepts one or more tsconfig paths:

```js
// Single tsconfig
parserOptions: { project: ['./tsconfig.json'] }

// Multiple tsconfigs (monorepo, separate test/build configs, …)
parserOptions: {
  project: ['./tsconfig.json', './packages/*/tsconfig.json'],
}
```

Each canonical tsconfig produces one TypeScript Program, even when multiple rslint configs reference it. Rslint retains every config association and project declaration order for lint-rule binding. Type-check runs over every real Program independently.

**The type-check scope is each tsconfig's `include` / `files` minus `exclude` — nothing in your rslint config or on the CLI changes it.** Specifically, the following are **lint-phase concepts** that do **not** affect type-check scope:

- rslint config's `files` patterns
- rslint config's `ignores` patterns (root-level or per-entry)
- `.gitignore`
- CLI file / directory arguments — `rslint --type-check-only foo.ts` still type-checks every file in the program(s), not just `foo.ts`

If a file is included by tsconfig but matched by rslint `ignores`, lint rules do not run on it, but **type errors for it are still reported**. To exclude it from type-check as well, add it to the tsconfig's `exclude` or prepend `// @ts-nocheck` to the file.

### Gap files

Selected files that are **not** present in any tsconfig Program declared by their governing config (root-level scripts, ad-hoc config files, etc.) are called _gap files_. They receive an AST-only fallback Program, so syntax-only rules still run but type-aware rules do not. The program-wide type-check phase also skips the fallback. To enable type information, add the file to one of the governing config's tsconfigs or declare a dedicated project there.

## Output

Type errors carry `TypeScript(TS<code>)` as the rule name and severity `error`:

```
  TypeScript(TS2322)  — [error] Type 'string' is not assignable to type 'number'.
  ╭─┴──────────( src/utils.ts:3:7 )─────
  │ 2 │  const name = 'hello';
  │ 3 │  const count: number = name;
  │ 4 │
  ╰────────────────────────────────
```

Chained errors indent the TypeScript message chain:

```
  TypeScript(TS2322)  — [error] Type 'B' is not assignable to type 'A'.
    The types of 'x.y.z' are incompatible between these types.
      Type 'number' is not assignable to type 'string'.
```

Type errors appear in every output format (`default`, `jsonline`, `github`).

### Summary line

Three templates depending on mode:

```
# Plain lint
Found 3 errors and 1 warning (linted 42 files with 5 rules in 120ms using 8 threads)

# --type-check
Found 3 lint errors, 2 type errors and 1 warning (linted 42 files with 5 rules in 120ms using 8 threads)

# --type-check-only
Found 2 type errors (type-checked 42 files in 80ms using 8 threads)
```

### Exit codes

| Code | When                                                                 |
| :--: | -------------------------------------------------------------------- |
|  0   | No errors. (Warnings still allowed unless `--max-warnings` rejects.) |
|  1   | At least one error (lint or type), or a runtime failure.             |
|  2   | Flag misuse — `--type-check-only` combined with `--fix` or `--rule`. |

## Alignment with `tsc --noEmit`

For any given program, `--type-check` (and `--type-check-only`) produces the same diagnostics as `tsc --noEmit` / `tsgo --noEmit` — same error code, same file, same line and column.

One intentional difference: TypeScript diagnostics without a source-file anchor (e.g. `TS18003` "No inputs were found in config file", `TS5108` removed-option warnings) are not reported, because rslint output is per file. Run `tsc --noEmit` directly to surface these configuration-level errors.

## Replacing `tsc --noEmit` in CI

```yaml
# Before — two steps
steps:
  - run: npx tsc --noEmit
  - run: npx rslint .

# After — one combined step
steps:
  - run: npx rslint --type-check .
```

For inline annotations on PR diffs:

```yaml
- run: npx rslint --type-check --format github .
```

If your CI keeps lint and type-check as separate jobs, use `--type-check-only` in the type-check job:

```yaml
jobs:
  type-check:
    steps:
      - run: npx rslint --type-check-only .
  lint:
    steps:
      - run: npx rslint .
```

## `--type-check-only`

Skips every lint rule and runs only the type-check phase. Use this when CI splits "type-check" and "lint" into separate steps and you want the type-check step to pay zero lint-side cost.

```bash
rslint --type-check-only .
```

`--type-check-only` implies `--type-check`; passing both is redundant.

### vs. `--type-check`

| Flag                | Lint rules | Type diagnostics | Suppresses lint-phase warnings <sup>\*</sup> |
| ------------------- | :--------: | :--------------: | :------------------------------------------: |
| `--type-check`      |     ✓      |        ✓         |                      no                      |
| `--type-check-only` |     ✗      |        ✓         |                     yes                      |

<sup>\*</sup> The lint phase emits per-file stderr warnings like `<file> was not found, skipping` and `<file> is ignored because of a matching ignore pattern`. In `--type-check-only` the lint phase doesn't run, so these are suppressed — they would otherwise mislead users into thinking the file wasn't type-checked, when in fact Phase 2 is independent of CLI scope and rslint ignores (see [What gets type-checked](#what-gets-type-checked)).

## Flag matrix

| Flag             | `--type-check`                                       | `--type-check-only`                          |
| ---------------- | ---------------------------------------------------- | -------------------------------------------- |
| `--fix`          | Applies lint fixes. Type errors have no auto-fix.    | **Rejected** (exit code 2).                  |
| `--rule`         | Overrides lint rules normally.                       | **Rejected** (exit code 2).                  |
| `--quiet`        | Suppresses warnings; type errors always shown.       | No-op — the lint phase produces nothing.     |
| `--format`       | Type errors rendered in the chosen format.           | Same.                                        |
| `--max-warnings` | Counts lint warnings only.                           | Always zero warnings (lint phase skipped).   |
| File/dir args    | Restricts lint scope. Type-check stays program-wide. | Lint skipped. Type-check still program-wide. |
