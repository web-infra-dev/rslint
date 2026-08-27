# no-conditional-in-test

## Rule Details

Disallows conditional control flow in test bodies. Separate tests make each expected behavior and failure easier to identify.

Reported forms are `if` and `switch` statements, conditional expressions, and the logical operators `&&`, `||`, and `??`. Optional chaining is reported as well once `allowOptionalChaining` is turned off.

The rule checks test callbacks and functions declared inside them. Conditions used for suite setup or helpers declared outside a test are not reported.

## Incorrect

```ts
test('renders the selected view', () => {
  if (mode === 'compact') {
    expect(render(mode)).toContain('Compact');
  } else {
    expect(render(mode)).toContain('Full');
  }
});
```

## Correct

```ts
test('renders the compact view', () => {
  expect(render('compact')).toContain('Compact');
});

test('renders the full view', () => {
  expect(render('full')).toContain('Full');
});
```

## Options

```json
{
  "rstest/no-conditional-in-test": [
    "error",
    {
      "allowOptionalChaining": false
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `allowOptionalChaining` | `boolean` | `true` | Allow optional chaining in test bodies. |
