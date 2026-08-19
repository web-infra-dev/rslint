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

A Chai chain can also state both halves by itself. It is reported, but not fixed, for the same reason as the chains below — the rewrite folds one matcher into another inside a single chain, and the rule does not model what else the chain asserts:

```typescript
expect(handler).to.have.been.calledOnce.and.calledWith('ready');
// merge by hand into: expect(handler).to.have.been.calledOnceWith('ready');
```

## Reported without a fix

A chain that asserts more than once is reported, but the rewrite is left to the author, because applying it would mean deleting or splicing assertions the rule does not model — and getting that wrong removes a live assertion silently:

```typescript
expect(handler).to.have.been.calledOnce.and.to.be.ok;
expect(handler).to.have.been.calledWith('ready');
// merge by hand into: expect(handler).to.have.been.calledOnceWith('ready').and.to.be.ok;
```

The same holds when the target is not stable under a second evaluation. The fix keeps one `expect(...)` call and deletes the other, so an argument such as a function call would go from evaluated twice to evaluated once — and may not even denote the same mock both times:

```typescript
expect(getMock()).toHaveBeenCalledOnce();
expect(getMock()).toHaveBeenCalledWith('ready');
```

Identifiers, property accesses, and element accesses with a literal key are stable, so those pairs are still fixed.

## Limits

The rule does not merge two assertions when:

- either carries `not` anywhere in its chain, because "not called once" and "not called with these arguments" is `¬ once ∧ ¬ with`, while the combined matcher negated is `¬(once ∧ with)`; Chai allows a modifier between matchers, so this includes `calledOnce.and.not.calledWith('ready')`;
- they carry different modifiers, because each then claims something about a different value; matching `resolves` or `rejects` on both halves does merge, awaited or not;
- one is awaited and the other is not, because dropping the awaited statement would leave a floating promise whose failure escapes as an unhandled rejection;
- anything other than an inert statement sits between them. The merge claims both assertions describe one call history, so nothing in between may call the target or rebind it. Neither question is decidable from the syntax, so the rule answers a narrower one it can: a statement may sit between the two halves only if it runs nothing of the author's. That covers an assertion whose `expect(...)` and matcher calls are its only calls, where no matcher executes what it asserts on and the arguments hold no other call, no `await`, no assignment and no spread, since `[...gen]` runs the iterator protocol the way a call runs a function; a call into TypeScript's default library, such as `console.log('checkpoint')` or `JSON.parse(text)`, provided it is not `eval`, whose source arrives as a string, and nothing callable is handed to it; and a declaration whose initializers meet the same bar, where a hoisted `var` must also declare names neither assertion reads and every declared name is a plain identifier, since a destructuring pattern runs that same protocol. Assertions on other mocks therefore keep their place between the two halves, while a reset, a reassignment, a further call to the target, and any call the rule cannot resolve leave the pair alone. Resolving the library call needs type information, so without a type checker such a call blocks the merge as well;
- either is an `expect.poll(...)` or `expect.element(...)` assertion, because each half then retries on its own until it passes and the two can settle against different call histories, which is the one thing the merge claims they do not do;
- more than two of these assertions share one target, because which pair to merge is ambiguous.

Two cases are not covered. When both halves are awaited, the second `await` is itself a suspension between the call-count check and the argument check, and the merge closes it: a mock called once before the pair and once more while the second half is suspended satisfies both halves separately — one recorded call at the first, one matching call among two at the second — and fails the merged assertion. The barrier rejects an intervening assertion that awaits for exactly this reason, but the pair's own `await` is part of a half rather than a statement between them. Refusing awaited pairs would drop the shape the rule was extended to support, so this one is left open.

The other: a matcher registered through `expect.extend` can invoke its subject the way `toThrow` does, and the rule has no way to know it. Recognising only the matchers Rstest ships would need a matcher table this repo does not have, and refusing an intervening assertion whose subject is callable would reject `expect(otherMock).toHaveBeenCalledWith(...)` — the assertion the barrier exists to let through.

Where a fix is offered, it preserves the surviving call's arguments, type arguments, comments, and formatting. The folded assertion is removed with its line when it owns that line, and otherwise on its own so that neighbouring statements and comments are never disturbed.
