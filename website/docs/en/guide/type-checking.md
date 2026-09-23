# Type Checking

Rslint can report TypeScript errors alongside lint diagnostics or run type checking on its own. Type checking covers entire configured projects, including imported files, even when the command names a single file.

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

To enable type-aware lint rules, configure `project` or `projectService` for the files you want to lint. Without either option, rules that do not require types still run. The type-check commands check a broader set of files, described in [What gets type-checked](#what-gets-type-checked).

For relative tsconfig paths, see the [`project` configuration reference](/config/language-options#languageoptionsparseroptionsproject).

## Automatic project discovery

Set [`projectService: true`](/config/language-options#languageoptionsparseroptionsprojectservice) to find a TypeScript project for each selected file automatically. Rslint looks for the nearest tsconfig or jsconfig that includes the file through `files` or `include`, following project references as needed. TypeScript presets do not enable this option automatically.

For example, with `projectService: true`, `rslint --type-check-only packages/app/src/file.ts` finds that file's project and checks the whole project, including sibling files. With no file arguments, Rslint uses files selected for linting in the current directory to find projects. `--type-check-only` never runs lint rules.

During linting, files outside discovered projects still receive rules that do not require types; see [gap files](#gap-files).

Do not combine `projectService: true` with a `project` string or array for the same file. To use explicit project paths instead, set `projectService: false`. See the [configuration reference](/config/language-options#languageoptionsparseroptionsprojectservice) for overrides and supported options.

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

Both type-check commands check every project listed in the loaded Rslint configuration, including projects in entries whose `files` patterns do not match the selected files. Overriding or clearing `project` in a later entry changes type-aware linting, but does not remove earlier projects from type checking. Projects found through `projectService` are also checked in full.

For lint rules, each file uses its final matching [`project` or `projectService` settings](/config/language-options#languageoptionsparseroptionsproject). These options provide type information; they do not select additional files to lint. Without either option, lint skips type-aware rules even when a `tsconfig.json` exists beside the Rslint config.

Relative project paths follow `basePath` and `tsconfigRootDir`. If selected files use different `tsconfigRootDir` values, the declared project paths are checked from each of those directories.

When no `project` paths are declared, type-check commands can use `tsconfig.json` beside the Rslint config. Files with `projectService` enabled, `project: false` or `null`, or `projectService: false` do not select this additional default project. `project: []` also disables the default. If no files are selected for linting, the default can still be checked. Neither `basePath` nor `tsconfigRootDir` changes its location. **This default applies only to the type-check commands.**

**Each project is checked in full, including files loaded through imports and references.** The following options control which files are linted and used for automatic project discovery, but do not exclude files from a project's TypeScript checks:

- rslint config's `files` patterns
- rslint config's `ignores` patterns (root-level or per-entry)
- `.gitignore`
- CLI file / directory arguments — `rslint --type-check-only foo.ts` still checks the entire selected project, not just `foo.ts`

If a file is included by tsconfig but matched by Rslint `ignores`, lint rules do not run on it, but **type errors for it are still reported**. The tsconfig's `exclude` only limits which files `include` selects; imported or referenced files can still be type-checked. `// @ts-nocheck` disables semantic checking of that file.

### Gap files

Files selected for linting but not included through `files` or `include` in any applicable TypeScript project are called _gap files_. These often include scripts and configuration files. Rslint still reports syntax errors and runs rules and fixes that do not require types. It skips type-aware rules.

For example, suppose tsconfig has `files: ["main.ts"]` and `main.ts` imports `helper.ts`, with no `include` pattern or other configured project selecting `helper.ts`:

- Linting `helper.ts` runs only rules that do not require types.
- Adding `helper.ts` to the tsconfig's `files` or `include` enables type-aware lint rules for it.
- `--type-check` and `--type-check-only` can still report TypeScript errors in `helper.ts` because the checked project imports it.

With `projectService: true`, Rslint may find a nested tsconfig that includes a gap file. If no project includes it, type-aware rules remain skipped. The typescript-eslint `allowDefaultProject` option is not supported.

An invalid configuration, such as a missing project path or conflicting project options, still reports an error. In the editor, the error appears on affected documents; other files continue using their own settings. Correcting the configuration restores normal lint diagnostics and fixes.

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

In combined mode, the displayed file count includes linted files and files selected by the checked tsconfigs' `files` or `include` settings, counting each file once. In color-enabled terminals, the parenthesized execution details are dimmed.

### Exit codes

| Code | When                                                                 |
| :--: | -------------------------------------------------------------------- |
|  0   | No errors. (Warnings still allowed unless `--max-warnings` rejects.) |
|  1   | At least one error (lint or type), or a runtime failure.             |
|  2   | Flag misuse — `--type-check-only` combined with `--fix` or `--rule`. |

## Alignment with `tsc --noEmit`

For the same TypeScript project, `--type-check` and `--type-check-only` produce the same diagnostics as `tsc --noEmit` / `tsgo --noEmit` — the same error code, file, line, and column.

TypeScript diagnostics that are not associated with a file, such as `TS18003` "No inputs were found in config file" or `TS5108` removed-option warnings, are not reported. Run `tsc --noEmit` directly to see these configuration errors.

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

Runs TypeScript checks without running any lint rules. It checks the same projects as `--type-check`.

```bash
rslint --type-check-only .
```

`--type-check-only` implies `--type-check`; passing both is redundant.

### vs. `--type-check`

| Flag                | Lint rules | Type diagnostics | Suppresses lint-phase warnings <sup>\*</sup> |
| ------------------- | :--------: | :--------------: | :------------------------------------------: |
| `--type-check`      |     ✓      |        ✓         |                      no                      |
| `--type-check-only` |     ✗      |        ✓         |                     yes                      |

<sup>\*</sup> Linting emits warnings like `<file> was not found, skipping` and `<file> is ignored because of a matching ignore pattern`. These are suppressed in `--type-check-only`. Ignored files can still receive TypeScript checks if they belong to a checked project; see [What gets type-checked](#what-gets-type-checked).

## Flag matrix

| Flag             | `--type-check`                                                                      | `--type-check-only`                                        |
| ---------------- | ----------------------------------------------------------------------------------- | ---------------------------------------------------------- |
| `--fix`          | Applies lint fixes. Type errors have no auto-fix.                                   | **Rejected** (exit code 2).                                |
| `--rule`         | Overrides lint rules normally.                                                      | **Rejected** (exit code 2).                                |
| `--quiet`        | Suppresses warnings; type errors always shown.                                      | No-op — the lint phase produces nothing.                   |
| `--format`       | Type errors rendered in the chosen format.                                          | Same.                                                      |
| `--max-warnings` | Counts lint warnings only.                                                          | Always zero warnings (lint phase skipped).                 |
| File/dir args    | Select files to lint and find projects with `projectService`; check whole projects. | Find projects with `projectService`; check whole projects. |
