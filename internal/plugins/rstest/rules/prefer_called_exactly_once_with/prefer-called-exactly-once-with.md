# prefer-called-exactly-once-with

## Rule Details

This rule prefers one precise mock assertion over separate call-count and argument assertions on the same target.

`toHaveBeenCalledWith` passes when *any* recorded call matches, so pairing it with `toHaveBeenCalledOnce` states "called exactly once, with these arguments" across two statements. `toHaveBeenCalledExactlyOnceWith` states it in one, and it cannot silently weaken if the call-count assertion is later removed.

Examples of **incorrect** code for this rule:

```typescript
expect(handler).toHaveBeenCalledOnce();
expect(handler).toHaveBeenCalledWith('ready');
```

Examples of **correct** code for this rule:

```typescript
expect(handler).toHaveBeenCalledExactlyOnceWith('ready');
```

## Chai-style spellings

Rstest also exposes the Chai-style spellings `calledOnce`, `calledWith`, and `calledOnceWith`, which assert on the same spy as their `toHaveBeen…` counterparts. The rule reports them too, including chains that mix both spellings, and the merged assertion keeps the spelling of the argument-side call:

```typescript
expect(handler).to.have.been.calledOnce;
expect(handler).to.have.been.calledWith('ready');
// → expect(handler).to.have.been.calledOnceWith('ready');
```

A Chai chain can also state both halves by itself, and is reported on its own:

```typescript
expect(handler).to.have.been.calledOnce.and.calledWith('ready');
// → expect(handler).to.have.been.calledOnceWith('ready');
```

## Reported without a fix

A chain that asserts more than once is reported, but the rewrite is left to the author, because applying it would mean deleting or splicing assertions the rule does not model — and getting that wrong removes a live assertion silently:

```typescript
expect(handler).to.have.been.calledOnce.and.to.be.ok;
expect(handler).to.have.been.calledWith('ready');
// → expect(handler).to.have.been.calledOnceWith('ready').and.to.be.ok;
```

## Limits

The rule does not merge two assertions when:

- either carries `not`, because "not called once" and "not called with these arguments" is `¬ once ∧ ¬ with`, while the combined matcher negated is `¬(once ∧ with)`;
- they carry different modifiers, because each then claims something about a different value; matching `resolves` or `rejects` on both halves does merge, awaited or not;
- one is awaited and the other is not, because dropping the awaited statement would leave a floating promise whose failure escapes as an unhandled rejection;
- `mockClear`, `mockReset`, or `mockRestore` resets the target between them, at any nesting depth, because each assertion then describes a different call history;
- more than two of these assertions share one target, because which pair to merge is ambiguous.

Where a fix is offered, it preserves the surviving call's arguments, type arguments, comments, and formatting. The folded assertion is removed with its line when it owns that line, and otherwise on its own so that neighbouring statements and comments are never disturbed.
