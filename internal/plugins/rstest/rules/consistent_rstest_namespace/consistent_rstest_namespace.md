# consistent-rstest-namespace

## Rule Details

Rstest exposes its utility namespace under two names, `rs` and `rstest`, and they do the same thing. Mixing both in one project makes the same mocking or timer call look like two different APIs, so this rule picks one spelling and reports the other.

A named import of the reported spelling from `@rstest/core` is reported, and so is every call or tagged-template invocation made through it — `rstest.mock('./service')` or `` rstest.fn`name` `` under the default preference, `rs.mock('./service')` under `{ "fn": "rstest" }`. Parentheses, non-null assertions and TypeScript type assertions around the namespace do not hide the invocation. The namespace is recognised whether it is imported or read as a global. It is not recognised when the name means something else: an aliased import such as `import { rstest as testUtils } from '@rstest/core'`, an import of the same name from another module, a namespace import, a `require` of the whole module, a local variable, or a parameter. Property names are left alone, so `config.rstest` is never rewritten.

`import.meta.rstest` is outside the rule. It is not a choice between the two spellings, and rewriting it would depend on a binding the file may not have.

## Incorrect

```ts
import { expect, rstest, test } from '@rstest/core';

rstest.mock('./payment-gateway');

test('charges the card once', () => {
  const charge = rstest.fn();
  expect(charge).not.toHaveBeenCalled();
});
```

## Correct

```ts
import { expect, rs, test } from '@rstest/core';

rs.mock('./payment-gateway');

test('charges the card once', () => {
  const charge = rs.fn();
  expect(charge).not.toHaveBeenCalled();
});
```

## Options

```json
{
  "rstest/consistent-rstest-namespace": [
    "error",
    {
      "fn": "rstest"
    }
  ]
}
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `fn` | `string` | `"rs"` | Namespace spelling to require, either `"rs"` or `"rstest"`. |

## Autofix

A call is rewritten by replacing the namespace object alone, so the method name, optional chaining, arguments and comments stay as written. An import of the reported spelling is renamed to the preferred one, or, when the file already imports the preferred spelling, dropped from its specifier list together with the separating comma and the comments written for it. A comment on a specifier that stays is kept where it was.

A report carries no fix when the preferred spelling is not free to write. That is the case when the file declares that name as something other than the Rstest namespace — a variable, a parameter, or an import of another name — since the rewrite would either declare the name twice or make the call reach the nearer binding instead of the namespace. Scopes are not distinguished here: a declaration anywhere in the file withholds the fix.

A report also carries no fix when rewriting it would leave the file naming a binding that no longer exists. That is the case when the namespace is bound by a destructured `require` rather than an import, when the file uses the binding somewhere the rewrite does not reach — a re-export such as `export { rstest }` or `export { rstest as helper }`, a type query such as `typeof rstest`, or a reference that is not an invocation such as `const { fn } = rstest` — and when dropping the specifier would empty its declaration, since that would turn the statement into a side-effect import.
