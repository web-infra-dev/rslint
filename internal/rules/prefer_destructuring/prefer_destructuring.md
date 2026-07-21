# prefer-destructuring

## Rule Details

Requires destructuring when extracting values from array indexes or object
properties in variable declarations and assignment expressions.

Examples of **incorrect** code for this rule:

```javascript
const foo = array[0];
const bar = object.bar;
bar = object.bar;
```

Examples of **correct** code for this rule:

```javascript
const [foo] = array;
const { bar } = object;
({ bar } = object);
```

Integer literal keys are treated as array access. Other computed keys are
treated as object access, regardless of the receiver's runtime type. Without
`enforceForRenamedProperties`, object access is reported only when an
identifier or string-literal property matches the target name.

For a nested member chain, the final property is destructured from its
immediate receiver:

```javascript
const bar = object.foo.bar;
```

This is fixed to:

```javascript
const { bar } = object.foo;
```

Direct optional-chain access, direct `super` access, and private properties
are not reported because they cannot be replaced safely with equivalent
destructuring.

## Options

This rule accepts up to two option objects.

### Array and object checks

Without options, both array and object checks are enabled for variable
declarations and assignment expressions.

The first option controls which access types are checked. The flat form applies
the same settings to declarations and assignments:

```json
{ "prefer-destructuring": ["error", { "array": false, "object": true }] }
```

Supplying this object replaces the defaults, so an omitted or `false` property
is disabled.

Checks can also be configured separately with `VariableDeclarator` and
`AssignmentExpression`. An omitted context is disabled:

```json
{
  "prefer-destructuring": [
    "error",
    {
      "VariableDeclarator": { "array": true, "object": true },
      "AssignmentExpression": { "array": false, "object": true }
    }
  ]
}
```

### Renamed properties

The optional second object accepts `enforceForRenamedProperties`, which defaults
to `false`. When it is `true`, object access is reported even when the target
name differs from the accessed property.

Examples of **incorrect** code for this rule with
`{ "enforceForRenamedProperties": true }`:

```json
{
  "prefer-destructuring": [
    "error",
    { "object": true },
    { "enforceForRenamedProperties": true }
  ]
}
```

```javascript
const foo = object.bar;
```

## Autofix

The autofix is limited to same-name, non-computed object property declarations
such as `const foo = object.foo`. Array access, assignments, renamed or
computed properties, and rewrites that would remove comments are reported
without a fix.

For TypeScript projects that need type-aware numeric-index handling or special
handling of declaration type annotations, use
`@typescript-eslint/prefer-destructuring`. The core rule follows ESLint's
autofix behavior and may remove an inline type annotation while converting a
declaration.

## Original Documentation

- [ESLint prefer-destructuring](https://eslint.org/docs/latest/rules/prefer-destructuring)
