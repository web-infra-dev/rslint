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

Rslint records the filename it checked, relative to the working directory's physical path, using `/` as the separator. Entering the same directory through a symlink keeps this base stable. Relative CLI arguments such as `../app.js` still follow the working directory used for file selection. The same conventions apply with `--config`. On Windows, a file on another drive or network share uses its absolute path.

For example, if `linked.js` points to `src/app.js`, the two names can match different rules. Checking them in separate runs gives each name its own count. Pruning `linked.js` leaves the entry for `src/app.js` intact.

On macOS, Rslint uses the spelling stored on disk for the working directory. If a directory symlink has several hard links and permissions prevent Rslint from identifying the directory entry, the baseline uses the spelling you supplied.

When a project scan or explicit arguments select both names under the same config, Rslint checks the physical file once. If you ask to record counts for both names, or either name already has a count, Rslint exits with code `2` and identifies both paths. Run Rslint on each path separately, or exclude one name in your config so project scans select the intended entry. Repeating the same filename as `app.js` and `./app.js` works normally. If different config files govern the two names, target selection reports a config conflict and exits with code `1` before applying suppressions.

To reuse an ESLint baseline, run `rslint --suppressions-location eslint-suppressions.json` from the directory where it was recorded. Entries apply when the file path and rule ID match.

Pass `--suppressions-location` on every run that uses a custom location. An existing directory or a path ending in a directory separator selects `rslint-suppressions_<hash>` inside that directory. Create the directory first. The hash comes from the absolute working directory.

## Remove unused suppressions

If the count in `src/app.js` falls from two to one, Rslint suppresses the remaining error and exits with code `2` because one recorded suppression is unused. Update the baseline and commit the change.

```bash
rslint --prune-suppressions
```

Pruning lowers counts, removes entries that reach zero, and removes entries for deleted files. A partial run preserves entries for existing files outside its scope. A Windows drive must be accessible to verify a deletion. Entries with UNC paths such as `//server/share/file.ts` are updated only when Rslint checks those files.

Ignored files and files with syntax errors keep their entries. If a file moves outside its TypeScript project, Rslint keeps the counts for enabled rules that require type information. Other rules in that file can still be checked and pruned. If you turn a rule off or lower its severity to `"warn"`, pruning removes its old count the next time Rslint checks the file.

If a plugin rule throws, Rslint reports the failure and any findings that rule emitted. Completed rules in the same run still use their baseline. By default, Rslint also reports any unused suppressions for those rules and exits with code `2` when it finds one. Fix plugin or configuration failures before recording or pruning.

To defer cleanup, run `rslint --pass-on-unpruned-suppressions`. Lint errors and the configured warning limit still determine the exit status.
