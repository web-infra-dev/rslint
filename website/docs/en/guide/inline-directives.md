---
description: 'Disable and re-enable Rslint rules with file and line comments using rslint or eslint directive prefixes.'
---

# Inline Directives

Rslint supports inline comments to disable or enable rules in source code. Both `rslint-` and `eslint-` prefixed directives are supported and fully equivalent.

## Disable for the rest of the file

```ts
/* rslint-disable */
// Disables all rules from this point forward

/* rslint-disable @typescript-eslint/no-explicit-any */
// Disables a specific rule from this point forward
```

## Re-enable rules

```ts
/* rslint-enable */
// Re-enables all previously disabled rules

/* rslint-enable @typescript-eslint/no-explicit-any */
// Re-enables a specific rule
```

## Disable for the current line

```ts
const x: any = 1; // rslint-disable-line
// Disables all rules for this line only

const y: any = 2; // rslint-disable-line @typescript-eslint/no-explicit-any
// Disables a specific rule for this line
```

## Disable for the next line

```ts
/* rslint-disable-next-line */
const x: any = 1;

/* rslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-call */
// Disables multiple specific rules for the next line
const y: any = fn();
```

## Notes

- `eslint-disable` / `eslint-enable` and their variants are also supported for ESLint compatibility. The two prefixes can be mixed freely (e.g. `rslint-disable` paired with `eslint-enable`).
- Both single-line (`//`) and multi-line (`/* */`) comment styles are supported.
- Omitting rule names disables/enables all rules.
- Multiple rule names can be separated by commas.
- An inline description can be added after `--` (e.g., `rslint-disable-next-line no-console -- temporary workaround`).

## Report unused directives

Reporting is off by default. Enable it per file through your configuration:

```ts
export default [
  {
    files: ['**/*.ts'],
    linterOptions: { reportUnusedDisableDirectives: 'warn' },
  },
];
```

The setting accepts `'off'`, `'warn'`, and `'error'`, or the corresponding
numbers `0`, `1`, and `2`. For ESLint configuration compatibility, `true` means
`'warn'` and `false` means `'off'`. Later matching entries override earlier
explicit settings. The JS API uses the same configuration.

Use `--report-unused-disable-directives` to report errors, or
`--report-unused-disable-directives-severity warn` to choose a severity. Explicit
CLI settings override configuration; the severity flag takes precedence over
the boolean flag regardless of argument order. Pass
`--report-unused-disable-directives=false` to turn reporting off.

Native rules and JavaScript plugin rules share usage tracking. A comment with
multiple rule names can be partially unused. `--fix` removes unused entries,
or replaces an entirely unused comment with a space to preserve token boundaries.
Unused enable comments are reported when they do not undo a disable directive.
The editor offers the same fixes through Quick Fix and Fix All.

Repeated or overlapping directives follow ESLint 10.14.0 attribution: the last
applicable block directive gets credit, and a line directive gets no credit
when a block directive already suppressed the diagnostic. Rules skipped by
execution policy exempt their entries and wildcard directives from unused
reporting. Failed or incomplete file execution produces no unused reports.

Both prefixes have identical behavior. Rslint also retains its support for
block disable/enable commands in `//` comments; ESLint requires `/* */` for
those commands. Malformed keywords such as `eslint-disabled` are not directives.
