# valid-title

## Rule Details

Validates Rstest `describe`, `test`, and `it` titles. Clear, consistently formatted titles make test output easier to scan and project conventions easier to enforce.

The rule rejects empty titles, titles that are not strings, titles with leading or trailing whitespace, and titles that repeat the registration name as their first word, such as `test('test creates a user')`. The `disallowedWords`, `mustMatch`, and `mustNotMatch` options add project-specific title conventions on top of that.

In array-based `.each` and `.for` titles the rule also checks format specifiers, accepting `%s`, `%d`, `%i`, `%f`, `%j`, `%o`, `%O`, `%c`, `%#`, `%$`, and `%%`.

## Incorrect

```ts
test('', () => {});
test('test creates a user', () => {});
describe('  user service  ', () => {});
```

## Correct

```ts
test('creates a user', () => {});
describe('user service', () => {});
```

## Options

```json
{
  "rstest/valid-title": [
    "error",
    {
      "disallowedWords": ["should"],
      "mustMatch": "^[a-z]"
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `ignoreSpaces` | `boolean` | `false` | Allow leading and trailing whitespace. |
| `ignoreTypeOfDescribeName` | `boolean` | `false` | Allow non-string `describe` titles. |
| `ignoreTypeOfTestName` | `boolean` | `false` | Allow non-string `test` and `it` titles. |
| `disallowedWords` | `string[]` | `[]` | Words that titles must not contain. |
| `mustMatch` | matcher | none | Pattern a title must match. |
| `mustNotMatch` | matcher | none | Pattern a title must not match. |

A matcher is a pattern string, `[pattern, message]`, or an object with separate `describe`, `test`, and `it` matchers.

## Autofix

Removes literal surrounding whitespace and duplicate `test`, `it`, or `describe` prefixes from titles.
