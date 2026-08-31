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

When `parserOptions.project` is omitted, rslint uses `tsconfig.json` in the governing config directory when present. An explicit `project: []` disables that fallback for the governing config. If neither configured projects nor that fallback tsconfig exist, no real TypeScript Program is built and type-check produces no diagnostics for that config.

An entry's [`basePath`](/config/base-path) becomes the anchor for its explicit project literals or globs. It does not move the implicit governing-directory `tsconfig.json` fallback or change Rslint's existing owner-wide project collection. Program-wide type checking still builds every explicit project declaration in the effective config catalog; `files`, `ignores`, `.gitignore`, and CLI target scope do not reduce that Program's TypeScript diagnostics.

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

Each normalized declared tsconfig path in the effective loaded config catalog produces one TypeScript Program, even when multiple rslint configs reference that path. Parent global ignores can prevent a nested config from entering that catalog during directory discovery. File-symlink declarations remain distinct because TypeScript resolves relative paths from the declared location. Rslint retains every config association and project declaration order for lint-rule binding. Type-check runs over every real Program independently.

**After the effective config catalog is established, each Program's type-check scope is its tsconfig `include` / `files` minus `exclude`.** The following lint-phase concepts do not filter that Program scope:

- rslint config's `files` patterns
- rslint config's `ignores` patterns (root-level or per-entry)
- `.gitignore`
- CLI file / directory arguments — `rslint --type-check-only foo.ts` still type-checks every file in the program(s), not just `foo.ts`

If a file is included by tsconfig but matched by rslint `ignores`, lint rules do not run on it, but **type errors for it are still reported**. To exclude it from type-check as well, add it to the tsconfig's `exclude` or prepend `// @ts-nocheck` to the file.

### Gap files

Selected files that are **not** present in any tsconfig Program declared by their governing config (root-level scripts, ad-hoc config files, etc.) are called _gap files_. CLI lint parses and binds them as standalone source files without creating a synthetic Program or TypeChecker, so rules that do not require type information still run while type-aware rules do not. Gap files never enter the program-wide type-check phase. To enable type information, add the file to one of the governing config's tsconfigs or declare a dedicated project there.

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

Type errors appear in every output format (`default`, `jsonline`, `github`, `gitlab`).

### Lifecycle status

The default format uses mode-specific start and completed status lines:

```
# Plain lint
start   Linting...
error   Lint failed with 3 errors and 1 warning in 120ms (42 files, 5 rules, 8 threads)

# --type-check
start   Linting and type checking...
error   Lint and type check failed with 3 lint errors, 2 TypeScript errors, and 1 warning in 120ms (47 files, 5 rules, 8 threads)

# --type-check-only
start   Type checking...
error   Type check failed with 2 TypeScript errors in 80ms (42 files, 8 threads)
```

In combined mode, the displayed file count is the canonical, deduplicated union of lint targets and root files from every compiler-capable tsconfig Program. It is not the larger of two counts: partially overlapping sets still contribute every distinct file. In color-enabled terminals, the complete parenthesized execution details are rendered dim.

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
