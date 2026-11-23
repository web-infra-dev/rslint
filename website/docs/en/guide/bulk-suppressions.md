---
description: 'Enable lint rules on an existing project by recording error counts per file and rule, checking them in CI, and pruning resolved violations.'
---

# Bulk suppressions

Bulk suppressions let you enable a lint rule on an existing project and fix the current violations over time. The Rslint CLI records an error count per file and rule, following the same workflow and JSON format as [ESLint's bulk suppressions](https://eslint.org/docs/latest/use/suppressions).

**Only lint rule errors are eligible.** A rule configured as `"warn"` still reports warnings, even if you name it with `--suppress-rule`. Warnings count toward `--max-warnings`. Rslint reports syntax errors and TypeScript compiler diagnostics independently of the baseline.

With `--type-check-only`, Rslint skips the baseline and rejects recording and pruning.

## Record existing errors

Enable the rules at `"error"` severity in your config, then record their current violations from the project directory.

```bash
rslint --suppress-rule "[no-var,no-debugger]"
```

You can also repeat `--suppress-rule` or pass comma-separated IDs. To record all error-level rules, use `rslint --suppress-all`. Choose one of `--suppress-all`, `--suppress-rule`, or `--prune-suppressions` per run. Add `--fix` to apply available fixes before counting the remaining errors.

Commit the generated `rslint-suppressions.json` with your config. Developers and CI can then run `rslint` from the same working directory to check against it.

## How the suppressions file works

This baseline records two `no-debugger` errors in `src/app.js`.

```json
{
  "src/app.js": {
    "no-debugger": { "count": 2 }
  }
}
```

If a third `no-debugger` error appears in that file, Rslint reports all three. Moving the two original errors to different lines leaves the count at two. Each file and rule has its own count.

File keys use the name Rslint linted, relative to the physical working directory, with `/` as the separator. A file on another Windows drive or UNC share uses its absolute path.

Entering the project through a directory symlink and running `rslint .` produces the same keys as entering the real directory. A symlink named in the lint target keeps its own spelling. For example, if `linked.js` points to `src/app.js`, checking `linked.js` alone updates its own entry. Each spelling can match different lint rules. If a run selects both names, Rslint checks the file once and applies counts for the selected name.

To reuse an ESLint baseline, run `rslint --suppressions-location eslint-suppressions.json` from the directory where it was recorded. Entries apply when the file path and rule ID match.

Pass `--suppressions-location` on every run that uses a custom location. An existing directory or a path ending in a directory separator selects `rslint-suppressions_<hash>` inside that directory. Create the directory first. The hash comes from the absolute working directory. To reuse a file ESLint generated in that directory, pass its full filename.

## Remove unused suppressions

If the count in `src/app.js` falls from two to one, Rslint suppresses the remaining error and exits with code `2` because one recorded suppression is unused. Update the baseline and commit the change.

```bash
rslint --prune-suppressions
```

Pruning lowers counts, removes entries that reach zero, and removes entries for deleted files. A partial run keeps entries for unchecked files unless Rslint can verify their deletion. This includes keeping entries for a symlink whose target is unavailable, for files in inaccessible directories, and for files on disconnected drives. Entries with UNC paths such as `//server/share/file.ts` are updated only when Rslint checks those files.

On a case-insensitive filesystem, a rename from `file.js` to `File.js` can leave the old spelling accessible. Remove its JSON entry if you no longer lint that spelling.

Ignored files and files with syntax errors keep their entries. If a file moves outside its TypeScript project, Rslint keeps the counts for enabled rules that require type information. Other rules in that file can still be checked and pruned. If you turn a rule off or lower its severity to `"warn"`, pruning removes its old count the next time Rslint checks the file.

A plugin rule that throws reports its failure and any findings it emitted. An incomplete rule keeps its recorded count, while completed rules can still record and prune theirs. If plugin results are missing for an entire file, `--suppress-all` and `--prune-suppressions` preserve that file's entries and update other files normally. `--suppress-rule` can still record a completed native rule in the affected file. The CLI exits with code `2` for incomplete file results. If a plugin batch request fails during the final check, the CLI leaves the baseline untouched.

If a config fails to load while checking or updating a baseline, a partial lint run exits with code `2`, even with an empty `{}` baseline. Files that lost their config keep their entries. Files linted with a fallback config report their findings, and files with a working config can still record and prune their counts.

To defer cleanup, run `rslint --pass-on-unpruned-suppressions`. Lint errors and the configured warning limit still determine the exit status.
