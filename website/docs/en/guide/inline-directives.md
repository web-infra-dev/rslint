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
