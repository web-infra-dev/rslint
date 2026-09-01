import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code, filename: 'file.mjs' });
const invalid = (
  code: string,
  suggestionOutputs: string[],
  filename: string = 'file.mjs',
) => ({
  code,
  filename,
  errors: suggestionOutputs.map((output) => ({
    message:
      'Prefer `.then(…).catch(…)` over passing a rejection handler to `.then()`.',
    suggestions: [
      {
        messageId: 'prefer-then-catch/suggestion',
        desc: 'Move the rejection handler to `.catch()`.',
        output,
      },
    ],
  })),
});

const invalidNoFix = (code: string, filename: string = 'file.mjs') => ({
  code,
  filename,
  errors: [
    {
      message:
        'Prefer `.then(…).catch(…)` over passing a rejection handler to `.then()`.',
    },
  ],
});

ruleTester.run('prefer-then-catch', null as never, {
  valid: [
    // ---- argument count ----
    valid('promise.then();'),
    valid('promise.then(onFulfilled);'),
    valid('promise.then(onFulfilled, onRejected, extraArgument);'),

    // ---- nullish handler ----
    valid('promise.then(undefined, onRejected);'),
    valid('promise.then(null, onRejected);'),
    valid('promise.then(void 0, onRejected);'),
    valid('promise.then(onFulfilled, undefined);'),
    valid('promise.then(onFulfilled, null);'),
    valid('promise.then(onFulfilled, void 0);'),

    // ---- spread arguments ----
    valid('promise.then(...handlers, onRejected);'),
    valid('promise.then(onFulfilled, ...handlers);'),

    // ---- member access shape ----
    valid('promise["then"](onFulfilled, onRejected);'),
    valid('promise?.then(onFulfilled, onRejected);'),
    valid('promise.then?.(onFulfilled, onRejected);'),
  ],
  invalid: [
    // ---- simple form ----
    invalid('promise.then(onFulfilled, onRejected);', [
      'promise.then(onFulfilled).catch(onRejected);',
    ]),
    invalid('promise.then(onFulfilled, function onRejected() {});', [
      'promise.then(onFulfilled).catch(function onRejected() {});',
    ]),

    // ---- shadowed `undefined` parameter is not global ----
    invalid(
      'function handlePromise(undefined) { promise.then(onFulfilled, undefined); }',
      [
        'function handlePromise(undefined) { promise.then(onFulfilled).catch(undefined); }',
      ],
    ),

    // ---- non-method-call receiver shape ----
    invalid('Promise.resolve(value).then(onFulfilled, onRejected);', [
      'Promise.resolve(value).then(onFulfilled).catch(onRejected);',
    ]),

    // ---- expression-statement wrappers ----
    invalid('void promise.then(onFulfilled, onRejected);', [
      'void promise.then(onFulfilled).catch(onRejected);',
    ]),

    // ---- trailing comma ----
    invalid('promise.then(onFulfilled, onRejected,);', [
      'promise.then(onFulfilled).catch(onRejected);',
    ]),

    // ---- parentheses around rejection handler ----
    invalid('promise.then(onFulfilled, (onRejected));', [
      'promise.then(onFulfilled).catch(onRejected);',
    ]),

    // ---- rejection handlers not safe to move ----
    invalidNoFix('promise.then(onFulfilled, createRejectionHandler());'),
    invalidNoFix('promise.then(onFulfilled, handlers.onRejected);'),

    // ---- comment arguments block the suggestion ----
    invalidNoFix(
      'promise.then(onFulfilled, /* Do not move this comment. */ onRejected);',
    ),
    invalidNoFix(
      'promise.then(onFulfilled, onRejected /* Do not move this comment. */);',
    ),
  ],
});
