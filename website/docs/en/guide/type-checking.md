# Type Checking

Rslint runs TypeScript semantic checks alongside or instead of lint rules. Explicit projects are checked program-wide. Automatic service discovery selects additional projects from lint targets. Each checked Program retains its complete roots and dependencies.

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

When both `parserOptions.project` and `projectService` are omitted, rslint uses `tsconfig.json` in the governing config directory when present. A declaration of `project: []` suppresses this fallback if no paths were declared; it does not remove earlier explicit declarations. Final matching false/null values disable explicit/default binding for individual lint targets, while program-wide explicit checking retains the owner's declarations.

An entry's [`basePath`](/config/base-path) anchors explicit project literals or globs unless the target has an explicit `tsconfigRootDir`. Each declaration keeps its own authored base when that root is omitted or reset. Neither option moves the implicit governing-directory fallback. Ordinary project strings and arrays retain their owner-wide declaration order; this differs from typescript-eslint's last matching project value.

## Automatic project discovery

With an explicit [`projectService: true`](/config/language-options#languageoptionsparseroptionsprojectservice), rslint discovers the projects that own the selected files. Nested tsconfigs and project references are followed; a tsconfig beside the lint config does not override the source's local project.

With service enabled and no explicit project declarations, `rslint --type-check-only packages/app/src/file.ts` finds that file's project and checks the whole project, including sibling files. It does not first build an implicit root project for that service target. No arguments select the current directory's lint scope for discovery. Type-check-only never executes lint rules.

Unowned service targets use source-only gap linting. TypeScript presets do not enable service; if another matching entry enables it, set `projectService: false` to use ordinary explicit projects. An effective project string or array, including [], conflicts with enabled service. Effective conflicts, service/root options and false/null clears are evaluated from the same matching config used for lint rules.

A matching `projectService: false` overrides earlier service settings and disables the target's implicit default binding. It preserves explicit owner declarations, including declarations from entries that do not match the target. JavaScript also accepts null for this service reset; the public type is boolean. Unmatched service/root/clear settings do not affect another target. Program-wide checking of an owner with no targets follows the separate scope rules below.

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

Rslint collects the governing owner's explicit project strings and arrays in declaration order, even from entries whose files/ignores do not match a lint target. `--type-check` and `--type-check-only` build every project in that list for each applicable root context. A later project array or [] does not remove an earlier declaration. Target binding prefers direct roots across the list, then import membership.

The construction scope depends on the operation:

| Operation                                      | Project scope                                                             |
| ---------------------------------------------- | ------------------------------------------------------------------------- |
| Plain CLI lint of the whole cwd                | Eagerly build explicit projects for active ordinary owners                |
| Focused file/subdirectory CLI lint or API lint | Select needed projects using root and import membership                   |
| `--type-check` or `--type-check-only`          | Check complete explicit declaration lists, plus service-selected Programs |

When service/root/clear options require target discovery, all actual targets contribute their effective `tsconfigRootDir` contexts to program-wide explicit checking, including service and clear targets. Each context checks the whole declaration list. Without an explicit root, the declarations keep their authored bases. A per-target clear changes lint binding; it does not remove these type-check projects.

When there are no explicit paths, only ordinary targets that allow the default can request the implicit owner tsconfig. If an owner has no selected targets at all, program-wide checking retains its original declaration/default lookup without guessing effective scoped options. An empty service-only scope therefore builds no Programs in plain lint, but type-check-only can still check the owner's default tsconfig. These scope rules belong to Rslint; ESLint has no corresponding type-check flags.

Shared explicit projects are constructed once per invocation. File-symlink declarations remain distinct because TypeScript resolves relative paths from the declared location. Explicit and service modes can require separate Programs for the same tsconfig because reference source and declaration-output behavior differs. Ownership probes may construct complete candidates to inspect import membership even when the target ultimately uses gap linting.

**Every checked Program includes its tsconfig root files and dependencies loaded through imports and references.** These filters affect lint targets and service/root context discovery, but do not trim a checked Program or the owner's ordinary explicit declaration list:

- rslint config's `files` patterns
- rslint config's `ignores` patterns (root-level or per-entry)
- `.gitignore`
- CLI file / directory arguments — `rslint --type-check-only foo.ts` still type-checks every file in the program(s), not just `foo.ts`

If a file is included by tsconfig but matched by rslint `ignores`, lint rules do not run on it, but **type errors for it are still reported**. The tsconfig's `exclude` filters `include` discovery; imports and references can still bring an excluded file into the Program. `// @ts-nocheck` disables semantic checking of that file.

### Gap files

Selected files without a project under their applicable binding settings (root-level scripts, ad-hoc config files, etc.) are called _gap files_. This includes JavaScript, TypeScript, and the other supported script extensions. The lint loader parses and binds them without providing a TypeChecker, so rules that do not require type information still run while type-aware rules are skipped. The source-only fallback itself does not participate in program-wide type checking. If the same file also belongs to a checked explicit or service Program, `--type-check` can still report TypeScript diagnostics for it through that Program. This fallback does not create a tsconfig for automatic project discovery.

With `projectService: true`, discovery can find an owning project that an explicit project list missed, such as a nested tsconfig. When discovery finds no owning project, the file follows the same gap fallback. Other files keep their selected projects. Actual config or Program failures still report errors.

To enable type information for a gap file, include it in a project selected by its effective parser settings. Upstream typescript-eslint instead rejects unowned service files by default and supports `allowDefaultProject` to provide type information for allowed files outside configured projects. Rslint does not yet support that option; its source-only gap fallback is not equivalent.

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

Skips every lint rule and runs only the type-check phase. Pure explicit-project configuration retains its path without lint-target discovery. New service/root/clear options can require target discovery to determine their applicable contexts; explicit checking still uses the complete declaration lists described above.

```bash
rslint --type-check-only .
```

`--type-check-only` implies `--type-check`; passing both is redundant.

### vs. `--type-check`

| Flag                | Lint rules | Type diagnostics | Suppresses lint-phase warnings <sup>\*</sup> |
| ------------------- | :--------: | :--------------: | :------------------------------------------: |
| `--type-check`      |     ✓      |        ✓         |                      no                      |
| `--type-check-only` |     ✗      |        ✓         |                     yes                      |

<sup>\*</sup> The lint phase emits per-file stderr warnings like `<file> was not found, skipping` and `<file> is ignored because of a matching ignore pattern`. These are suppressed in `--type-check-only`. Ignored files can still be checked through the program-wide explicit declarations or a service-selected Program (see [What gets type-checked](#what-gets-type-checked)).

## Flag matrix

| Flag             | `--type-check`                                                                      | `--type-check-only`                                                |
| ---------------- | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| `--fix`          | Applies lint fixes. Type errors have no auto-fix.                                   | **Rejected** (exit code 2).                                        |
| `--rule`         | Overrides lint rules normally.                                                      | **Rejected** (exit code 2).                                        |
| `--quiet`        | Suppresses warnings; type errors always shown.                                      | No-op — the lint phase produces nothing.                           |
| `--format`       | Type errors rendered in the chosen format.                                          | Same.                                                              |
| `--max-warnings` | Counts lint warnings only.                                                          | Always zero warnings (lint phase skipped).                         |
| File/dir args    | Scope lint targets and new parser contexts; explicit checking remains program-wide. | Scope new parser contexts; explicit checking remains program-wide. |
