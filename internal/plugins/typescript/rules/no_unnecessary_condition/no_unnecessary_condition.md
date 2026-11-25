# no-unnecessary-condition

## Rule Details

Disallow conditionals where the type is always truthy or always falsy.

Any expression being used as a condition must be able to evaluate as truthy or falsy in order to be considered necessary. Conversely, any expression that always evaluates to truthy or always evaluates to falsy is considered unnecessary and will be flagged by this rule.

Examples of **incorrect** code for this rule:

```typescript
function head<T>(items: T[]) {
  // items is always truthy (arrays are objects)
  if (items) {
    return items[0].toUpperCase();
  }
}

function foo(arg: 'bar' | 'baz') {
  // arg is always truthy (non-empty string literals)
  if (arg) {
  }
}
```

Examples of **correct** code for this rule:

```typescript
function head<T>(items: T[]) {
  // items.length can be zero
  if (items.length) {
    return items[0].toUpperCase();
  }
}

function foo(arg: string) {
  // string can be empty
  if (arg) {
  }
}
```

## Differences from upstream

With `noUncheckedIndexedAccess` enabled, rslint preserves optional chaining on a
possibly missing indexed value, including a key variable with a literal type.
typescript-eslint 8.74.0 can suggest removing the last `?.` in these examples,
which introduces a TypeScript error.

```typescript
declare const items: { id: string }[] | undefined;
declare const records: Record<number, { id: string }> | undefined;
declare const key: 0;

items?.[(key)]?.id;
records?.[(key)]?.id;
```

The same protection applies when an index signature explicitly includes
`undefined` or `null`, or a property keyed by a unique symbol is optional. Index
matching follows TypeScript's rules for numeric keys, numeric strings, symbols,
and template literal keys.

rslint also reports constant conditions in array predicate callbacks and
truthiness assertions, unreachable logical fallbacks, and computed reads narrowed
to `null` or `undefined`. typescript-eslint 8.74.0 can miss these cases:

```typescript
declare const flags: boolean[];
declare const label: string | undefined;

flags.filter(flag => flag || true);
const display = label || 'Untitled' || 'Fallback';

function read(value: { x?: number }) {
  if (value.x === undefined) return value['x'] ?? 42;
  return value.x;
}
```

Async and generator callbacks return objects, regardless of the value in their
`return` statement. For example, `flags.filter(async () => false)` keeps every
element. rslint reports the callback as always truthy instead of reporting its
body as always falsy.

Logical assignments check the value being read. Ordinary fields retain
control-flow narrowing. Accessors use the getter's declared type: a setter can
transform its argument, and each getter call can return a different value.
The same protection applies to properties of an object returned by a getter.
Values copied by object spread, object rest, or `Object.assign` retain ordinary
field narrowing, as do undecorated private auto-accessors. A later optional
spread may leave a getter already on the target in place; it remains an accessor.
`Object.assign` also preserves target getters, including those on class instances
and parameter objects; writing through a setter does not replace the getter.
An unknown target typed as `object` may also contain a getter, so its structural
type alone does not establish an ordinary data copy. This also applies when the
source uses ordinary fields or index signatures: its write types do not establish
the target getters' read types. A target with `__proto__` can inherit a getter and
setter that handle the assignment. A target's declared getter contract is preserved
before inspecting its constructor, which can return a different object.
This also applies through nested `Object.assign` calls and index signatures.
Reading an unknown target
into a stable local, such as `const item = copy.item`, does not prove that
`item?.id` is redundant. Destructured reads retain the same uncertainty; subsequent
guards on the captured value can still make a fallback unnecessary. Captured
getter values also preserve the getter's read type rather than a preceding setter
argument, including explicitly typed locals and destructuring defaults. Ordinary
fields on a captured object retain their own control-flow narrowing.
Fresh arrays and functions retain ordinary data-copy narrowing when no descriptor
changes or unaccounted calls are visible through their aliases. Deleting properties,
changing the prototype, or copying a possible own `__proto__` key prevents that
assumption. Objects with a `null` prototype also retain ordinary copy narrowing.
Initializer types never narrow a getter's public read type, and an annotation on
the copy cannot erase the target getter's actual read type. This also covers values
supplied by later spreads or descriptor changes. Calls resolve the applicable
overload of each possible function through the checker. A function with one fixed
signature can ignore extra arguments accepted by another member of a callable union.
For a union of keys, the read type includes all selected properties. Setter types
and write intersections must not determine whether the condition is necessary.
For example, this `??=` remains necessary even after the setter receives `null`:

```typescript
interface Box {
  get value(): number | undefined;
  set value(reset: null);
}

function reset(box: Box) {
  box['value'] = null;
  box['value'] ??= null;
}
```

When `allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing` is enabled, optional
property and parameter markers still protect nullish fallbacks. For example, with
`strictNullChecks` disabled, `options: { title?: string }` can be `{}`, so
`options.title ?? 'Untitled'` remains necessary.
After an assignment of a statically known value, such as
`options.title = 'set'`, the fallback is reported as unnecessary. Unrelated
simple variable declarations initialized by primitive literals, identifiers, empty
objects or arrays, or function expressions preserve this proof. Destructuring,
object spread, and other expressions may invoke user code,
so they stop this proof. Other optional
property and parameter fallbacks are retained conservatively: without strict null checking,
the compiler also erases nullish values from assignment sources. For example,
assigning an optional parameter with `options.title = title` does not prove the
property is present.

A non-null assertion inside an optional chain does not cancel an earlier short
circuit. For example, the final `?.` remains necessary here because `root` can be
`undefined`:

```typescript
interface Leaf {
  get item(): { id: number };
}

function read(root: { middle: Leaf } | undefined) {
  return (root?.middle!.item)?.id;
}
```

## Original Documentation

- [typescript-eslint: no-unnecessary-condition](https://typescript-eslint.io/rules/no-unnecessary-condition)
- [Source code](https://github.com/typescript-eslint/typescript-eslint/blob/v8.74.0/packages/eslint-plugin/src/rules/no-unnecessary-condition.ts)
