# prefer-to-contain

## Rule Details

This rule requires `toContain()` when an equality assertion checks the boolean result of a single-argument `includes()` call. The dedicated matcher states the intended containment check directly and produces a more useful failure message.

The rule recognizes `toBe()`, `toEqual()` and `toStrictEqual()`, including negated forms, and preserves the assertion's meaning when deciding whether the replacement needs `.not`. It recognizes Rstest globals, imports and aliases, namespace and `require` bindings, `rstack/test`, `import.meta.rstest`, `expect.soft()`, Playwright's ordinary value assertions, and the `expect` supplied by a [test context](https://rstest.rs/api/runtime-api/test-api/test#testcontext).

Polling assertions, promise modifiers, browser-only `expect.element()` assertions, Chai-style assertions, dynamic accessors and explicit `NaN` items are left unchanged. A statically known string with a regular-expression item is also excluded. These cases differ at runtime: Rstest's array `toContain()` does not match `Array.prototype.includes()` for `NaN`, while native string `includes()` rejects regular expressions.

## Incorrect

```ts
expect(flavors.includes('lime')).toBe(true);
expect(roles.includes('admin')).not.toEqual(false);
expect(tags.includes('deprecated')).toStrictEqual(false);
```

## Correct

```ts
expect(flavors).toContain('lime');
expect(roles).toContain('admin');
expect(tags).not.toContain('deprecated');
```

## Autofix

The autofix moves the `includes()` receiver and item into `expect(receiver).toContain(item)`, removes equality-matcher type arguments, then adds or removes `.not` to preserve the boolean assertion. Quoted and template-literal matcher accessors retain their delimiters. If the moved item may have side effects, `expect()` has a custom message argument, or an edited expression contains a comment, the rule reports the assertion without a fix.
