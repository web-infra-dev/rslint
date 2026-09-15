# no-async-mock-factory

## Rule Details

A mock factory must hand back the module object synchronously.

Rstest installs the second argument of `rs.mock`, `rs.doMock`, `rs.mockRequire` and `rs.doMockRequire` as the mocked module's own factory and calls it while the module graph is being built, where the module itself is what has to come back. Returning a promise always fails and throws `[Rstest] An async mock factory is not supported.` — and the type system does not reject it, so the error surfaces only once the mocked module is really loaded, which can be far from the mock itself.

What matters is the value the factory actually returns, not how it is written: an `async` function, a plain function returning `Promise.resolve(…)`, `new Promise(…)`, a dynamic `import(…)` and `rs.importActual(…)` all amount to the same thing. A generator and a hand-written thenable hand back something that is not a promise, so the runtime lets them through and so does this rule.

Only a factory that can *only* return a promise is reported. One with a path that runs off the end of its body returns `undefined` there, which the runtime accepts, so it is left alone.

A factory the syntax cannot settle — one that comes from another module, or that a helper produces — is judged from its type when a TypeScript project is available, and goes unreported without type information. A type that mixes a promise-returning signature with a synchronous one is never reported, because such a call may well work.

## Incorrect

```ts
import { rs } from '@rstest/core';

rs.mock('./api', async () => {
  const actual = await rs.importActual<typeof import('./api')>('./api');
  return { ...actual, get: rs.fn() };
});

rs.mock('./config', () => Promise.resolve({ endpoint: 'http://localhost' }));
```

## Correct

```ts
import * as actual from './api' with { rstest: 'importActual' };
import { rs } from '@rstest/core';

rs.mock('./api', () => ({ ...actual, get: rs.fn() }));

rs.mock('./config', () => ({ endpoint: 'http://localhost' }));
```

## Suggestions

The rule provides no autofix. The common shape awaits the real module inside the factory, and rewriting that means adding a top-level `import … with { rstest: 'importActual' }` and reshaping the body — a change across statements that should not be applied unattended.

Two narrow suggestions are offered where the rewrite is exactly equivalent:

- **Remove `async`**, when the body neither suspends nor already hands back a promise of its own.
- **Return the value directly instead of wrapping it in `Promise.resolve(…)`**, when the factory's only returned expression is a single-argument `Promise.resolve(…)` around a value that is plainly not a promise.

Both share two preconditions. The code being rewritten must contain no comment, because each suggestion replaces a whole span and a comment sitting inside that span would be deleted with it. And the factory must carry no return type annotation of its own, because both rewrites take the promise out of the value while leaving the annotation as written, which would leave `(): Promise<T> => …` annotating a plain `T`. When either fails the rule reports without a suggestion and leaves the change to you.
