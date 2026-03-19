# Inline Directives

Rslint supports ESLint-compatible inline comments to disable rules in source code.

## Disable for the rest of the file

```ts
/* eslint-disable */
// Disables all rules from this point forward

/* eslint-disable @typescript-eslint/no-explicit-any */
// Disables a specific rule from this point forward
```

## Re-enable rules

```ts
/* eslint-enable */
// Re-enables all previously disabled rules

/* eslint-enable @typescript-eslint/no-explicit-any */
// Re-enables a specific rule
```

## Disable for the current line

```ts
const x: any = 1; // eslint-disable-line
// Disables all rules for this line only

const y: any = 2; // eslint-disable-line @typescript-eslint/no-explicit-any
// Disables a specific rule for this line
```

## Disable for the next line

```ts
/* eslint-disable-next-line */
const x: any = 1;

/* eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-call */
// Disables multiple specific rules for the next line
const y: any = fn();
```

## Notes

- Both single-line (`//`) and multi-line (`/* */`) comment styles are supported.
- Omitting rule names disables/enables all rules.
- Multiple rule names can be separated by commas.
