# no-unescaped-entities

## Rule Details

Disallow unescaped HTML entities from appearing in markup. Reports occurrences of the characters `>`, `"`, `'`, and `}` in JSX text content and suggests escape sequences as replacements.

Examples of **incorrect** code for this rule:

```jsx
<div> > </div>
<div>Don't</div>
<div>{"foo"}}</div>
```

Examples of **correct** code for this rule:

```jsx
<div> &gt; </div>
<div>Don&apos;t</div>
<div>{"foo"}</div>
<div>{">"}</div>
```

## Options

- `forbid` (default: `[">", "\"", "'", "}"]` with standard alternatives): List of forbidden characters. Each item can be either a string (the character itself) or an object `{ char: string, alternatives: string[] }` specifying replacement suggestions.

```jsonc
{
  "react/no-unescaped-entities": ["error", { "forbid": [">", "}"] }]
}
```

```jsonc
{
  "react/no-unescaped-entities": [
    "error",
    {
      "forbid": [
        { "char": ">", "alternatives": ["&gt;"] },
        { "char": "}", "alternatives": ["&#125;"] }
      ]
    }
  ]
}
```

## Differences from ESLint

TypeScript's JSX parser rejects unescaped `>` and `}` in JSX text with a syntax error before this rule can run, so for TypeScript sources those defaults are effectively enforced by the parser itself. This rule still catches `'`, `"`, and any custom characters configured via `forbid`.

## Original Documentation

- [react/no-unescaped-entities](https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/no-unescaped-entities.md)
