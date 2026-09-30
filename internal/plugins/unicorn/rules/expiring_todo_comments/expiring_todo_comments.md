# expiring-todo-comments

## Rule details

Reports TODO comments when their expiration conditions are met.

This rule checks `TODO`, `FIXME`, and `XXX` comments, ignoring case. Each condition
that is met produces a separate diagnostic. It does not offer fixes or suggestions.

### Conditions

| Condition | Reports when |
| --- | --- |
| `[2026-09-30]` | The reference UTC date is later than the deadline, with `checkDates: true`. |
| `[>=2.0.0]` | The nearest package.json version satisfies the comparison. |
| `[+react]` / `[-jquery]` | A dependency is installed / absent in dependencies and devDependencies. |
| `[react@>=19]` | The declared dependency version satisfies the comparison. |
| `[peer:eslint@>=9]` | The minimum supported peer dependency version satisfies the comparison. |
| `[engine:node@>=22]` | The declared Node.js engine version satisfies the comparison. |

Version comparisons support `>` and `>=`, including partial versions and
prereleases. For example, `>1` means version 2 or later, while `>1.0.0` includes
1.0.1. Peer dependency ranges such as `^8 || ^9` use their lowest supported
version, 8.0.0 in this example. Dependency and engine comparisons use the first
coercible version in their declaration. Dependency version checks report
`catalog:` references as unsupported; presence checks still work.

Separate conditions with commas. Multiple dates or multiple package-version
conditions on the same line are reported regardless of `checkDates`.

```js
// TODO [2026-09-30]: Remove this temporary fallback.
// FIXME [>=2.0.0, peer:eslint@>=9]: Remove the deprecated API.
// TODO [-jquery, +react]: Migrate this widget.
/*
 * TODO [engine:node@>=22]: Use the newer Node.js API.
 * TODO [react@>=19]: Update the component.
 */
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| `terms` | `["todo", "fixme", "xxx"]` | Comment terms to check; replaces the defaults. |
| `ignore` | `[]` | Regular expression strings that exclude matching comment lines. |
| `checkDates` | `false` | Enable expiration-date checks. |
| `checkDatesOnPullRequests` | — | **Unsupported.** Accepted only for configuration compatibility; its value is ignored. Use `checkDates` to control date checks. |
| `allowWarningComments` | `true` | Allow warning comments without recognized conditions. |
| `date` | Today in UTC | Reference date in `YYYY-MM-DD` format. |

```js
export default [{
  plugins: ['unicorn'],
  rules: {
    'unicorn/expiring-todo-comments': ['error', {
      checkDates: true,
      allowWarningComments: false,
      ignore: ['ISSUE-\\d+'],
    }],
  },
}];
```

With `allowWarningComments: false`, an ordinary `// TODO: Add tests` reports too.
Warning terms can appear anywhere in the comment. Block-comment lines are checked
individually, and their diagnostics highlight the enclosing comment.

Package conditions require a readable package.json in the working directory or
one of its parents. Comparisons use the package.json nearest the linted file.

## Differences from upstream

- `checkDatesOnPullRequests` is unsupported. Its value is accepted for
  configuration compatibility but ignored. For example, with
  `checkDates: true` and `checkDatesOnPullRequests: false`, expired dates still
  report. Set `checkDates: false` to disable these diagnostics.
- Write `ignore` entries as pattern strings. JavaScript `RegExp` objects do not
  work as ignore patterns. For example, replace `/issue-\d+/i` with
  `'[iI][sS][sS][uU][eE]-\\d+'` to ignore both `ISSUE-123` and `issue-123`.
- Invalid or unsupported ignore patterns are ignored. For example,
  `'\\p{Script=Han}+'` does not suppress an expired comment containing `中`,
  while upstream suppresses it. Use a supported expression such as `'[\\u4E00-\\u9FFF]+'`
  for that character range.

## Original documentation

- [Upstream documentation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/expiring-todo-comments.md)
- [Upstream implementation](https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/rules/expiring-todo-comments.js)
