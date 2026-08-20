import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code, filename: 'file.mjs' });
const invalid = (
  code: string,
  method: string,
  suggestionOutputs: string[],
) => ({
  code,
  filename: 'file.mjs',
  errors: suggestionOutputs.map((output) => ({
    message: `Promise in \`Promise.${method}()\` should not be awaited.`,
    suggestions: [
      {
        messageId: 'no-await-in-promise-methods/suggestion',
        desc: 'Remove `await`.',
        output,
      },
    ],
  })),
});

ruleTester.run('no-await-in-promise-methods', null as never, {
  valid: [
    valid('Promise.all([promise1, promise2, promise3, promise4])'),
    valid('Promise.allSettled([promise1, promise2, promise3, promise4])'),
    valid('Promise.any([promise1, promise2, promise3, promise4])'),
    valid('Promise.race([promise1, promise2, promise3, promise4])'),
    valid('Promise.all(...[await promise])'),
    valid('Promise.all([await promise], extraArguments)'),
    valid('Promise.all()'),
    valid('Promise.all(notArrayExpression)'),
    valid('Promise.all([,])'),
    valid('Promise[all]([await promise])'),
    valid('Promise.all?.([await promise])'),
    valid('Promise?.all([await promise])'),
    valid('Promise.notListedMethod([await promise])'),
    valid('NotPromise.all([await promise])'),
    valid('Promise.all([(await promise, 0)])'),
    valid('new Promise.all([await promise])'),

    // We are not checking these cases.
    valid('globalThis.Promise.all([await promise])'),
    valid('Promise["all"]([await promise])'),
  ],
  invalid: [
    invalid('Promise.all([await promise])', 'all', ['Promise.all([promise])']),
    invalid('Promise.allSettled([await promise])', 'allSettled', [
      'Promise.allSettled([promise])',
    ]),
    invalid('Promise.any([await promise])', 'any', ['Promise.any([promise])']),
    invalid('Promise.race([await promise])', 'race', [
      'Promise.race([promise])',
    ]),
    invalid('Promise.all([, await promise])', 'all', [
      'Promise.all([, promise])',
    ]),
    invalid('Promise.all([await promise,])', 'all', [
      'Promise.all([promise,])',
    ]),
    invalid('Promise.all([await promise],)', 'all', [
      'Promise.all([promise],)',
    ]),
    invalid('Promise.all([await (0, promise)],)', 'all', [
      'Promise.all([(0, promise)],)',
    ]),
    invalid('Promise.all([await (( promise ))])', 'all', [
      'Promise.all([(( promise ))])',
    ]),
    invalid('Promise.all([await await promise])', 'all', [
      'Promise.all([await promise])',
    ]),
    invalid('Promise.all([...foo, await promise1, await promise2])', 'all', [
      'Promise.all([...foo, promise1, await promise2])',
      'Promise.all([...foo, await promise1, promise2])',
    ]),
    invalid('Promise.all([await promise1, await promise2])', 'all', [
      'Promise.all([promise1, await promise2])',
      'Promise.all([await promise1, promise2])',
    ]),
    invalid('Promise.any([await a, await b, await c])', 'any', [
      'Promise.any([a, await b, await c])',
      'Promise.any([await a, b, await c])',
      'Promise.any([await a, await b, c])',
    ]),
    invalid('Promise.all([await /* comment*/ promise])', 'all', [
      'Promise.all([/* comment*/ promise])',
    ]),
  ],
});
