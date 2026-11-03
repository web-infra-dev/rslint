// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
import path from 'node:path';
import { runInNewContext } from 'node:vm';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester';

const defaults = {
  filename: 'src/case.js',
  languageOptions: { sourceType: 'module' as const },
};
const error = {
  messageId: 'no-useless-fallback-in-spread',
  message: 'The empty object is useless.',
};

new RuleTester().run('no-useless-fallback-in-spread', {} as never, {
  valid: [
    // Snapshot cases
    'const array = [...(foo || [])]',
    'const array = [...(foo || {})]',
    'const array = [...(foo && {})]',
    'const object = {...(foo && {})}',
    'const object = {...({} || foo)}',
    'const object = {...({} && foo)}',
    'const object = {...({} ?? foo)}',
    'const object = {...foo}',
    'const object = {...(foo ?? ({} || {}))}',
    'const {...foo} = object',
    'function foo({...bar}){}',
    'const object = {...(foo || {}).toString()}',
    'const object = {...fn(foo || {})}',
    'const object = call({}, ...(foo || {}))',
    'const object = {...(foo || {not: "empty"})}',
    'const object = {...(foo || {...{}})}',
    'const object = {...(foo ? {bar: true} : {})}',
    'const object = {...(!foo ? {} : {bar: true})}',
    'const object = {...(foo ? {a: 1} : {b: 2})}',
    'const object = {...(foo ? {} : {})}',
    'const object = {...(foo ? foo : {})}',
    'const object = {...(!foo ? {} : foo)}',
    'const object = {...(foo.bar ? foo.bar : {})}',
    'const array = [...(foo ? {a: 1} : {})]',
    'const object = {...(foo ? {a: 1} : {}).bar}',
    // Documentation examples
    'const object = {...foo};',
    'const object = {...(foo && {})};',
    'const array = [...(foo || [])];',
  ].map((code) => ({ ...defaults, code })),
  invalid: [
    // Comment-preserving fixes
    {
      ...defaults,
      code: 'const object = {...(getObject(/* keep */) || {})};',
      output: 'const object = {...getObject(/* keep */)};',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...((/* keep */ foo) ?? {})};',
      output: 'const object = {...(/* keep */ foo)};',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...({/* keep */ value: 1} || {})};',
      output: 'const object = {...{/* keep */ value: 1}};',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(getObject(// keep\n) ?? {})};',
      output: 'const object = {...getObject(// keep\n)};',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(getObject(// keep\r\n) ?? {})};',
      output: 'const object = {...getObject(// keep\r\n)};',
      errors: [error],
    },
    // TypeScript and comments blocking fixes
    {
      ...defaults,
      code: 'const object = {...((foo as object /* keep */) ?? {})};',
      filename: 'src/case.ts',
      output: 'const object = {...(foo as object /* keep */)};',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(foo || {/* keep */})};',
      output: null,
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...((foo) /* keep */ ?? {})};',
      output: null,
      errors: [error],
    },
    // Snapshot cases
    {
      ...defaults,
      code: 'const object = {...(foo || {})}',
      output: 'const object = {...foo}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(foo ?? {})}',
      output: 'const object = {...foo}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(foo ?? (( {} )))}',
      output: 'const object = {...foo}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...((( foo )) ?? (( {} )))}',
      output: 'const object = {...(( foo ))}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(( (( foo )) ?? (( {} )) ))}',
      output: 'const object = {... (( foo )) }',
      errors: [error],
    },
    {
      ...defaults,
      code: 'async ()=> ({...((await foo) || {})})',
      output: 'async ()=> ({...(await foo)})',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(0 || {})}',
      output: 'const object = {...0}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...((-0) || {})}',
      output: 'const object = {...(-0)}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(.0 || {})}',
      output: 'const object = {....0}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(0n || {})}',
      output: 'const object = {...0n}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(false || {})}',
      output: 'const object = {...false}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(null || {})}',
      output: 'const object = {...null}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(undefined || {})}',
      output: 'const object = {...undefined}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...((a && b) || {})}',
      output: 'const object = {...(a && b)}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(NaN || {})}',
      output: 'const object = {...NaN}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...("" || {})}',
      output: 'const object = {...""}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...([] || {})}',
      output: 'const object = {...[]}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...({} || {})}',
      output: 'const object = {...{}}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(foo || {}),}',
      output: 'const object = {...foo,}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(foo /* keep */ || {})}',
      output: null,
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(foo || /* keep */ {})}',
      output: null,
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...((foo ?? {}) || {})}',
      output: 'const object = {...foo}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...((foo && {}) || {})}',
      output: 'const object = {...(foo && {})}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(foo && {} || {})}',
      output: 'const object = {...foo && {}}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...({...(foo || {})})}',
      output: 'const object = {...({...foo})}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...({...((0, foo) || {})})}',
      output: 'const object = {...({...(0, foo)})}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'function foo(a = {...(bar || {})}){}',
      output: 'function foo(a = {...bar}){}',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(document.all || {})}',
      output: 'const object = {...document.all}',
      errors: [error],
    },
    // Runtime-equivalence regression
    {
      ...defaults,
      code: 'JSON.stringify({...((value /* keep */) || {})});',
      output: 'JSON.stringify({...(value /* keep */)});',
      errors: [error],
    },
    {
      ...defaults,
      code: 'JSON.stringify({...((value /* keep */) ?? {})});',
      output: 'JSON.stringify({...(value /* keep */)});',
      errors: [error],
    },
    // Documentation examples
    {
      ...defaults,
      code: 'const object = {...(foo || {})};',
      output: 'const object = {...foo};',
      errors: [error],
    },
    {
      ...defaults,
      code: 'const object = {...(foo ?? {})};',
      output: 'const object = {...foo};',
      errors: [error],
    },
  ],
});

// Upstream also verifies that preserving a comment inside the left operand
// retains the spread result for both fallback operators and representative values.
test('comment-preserving fallback fixes retain spread results', async () => {
  const configDirectory = path.resolve(import.meta.dirname, '..');
  const filename = path.join(configDirectory, 'src/runtime.js');
  for (const operator of ['||', '??']) {
    const code = `JSON.stringify({...((value /* keep */) ${operator} {})});`;
    const result = await lint({
      configDirectory,
      workingDirectory: process.cwd(),
      config: [
        {
          plugins: ['unicorn'],
          languageOptions: {
            sourceType: 'module',
            parserOptions: { projectService: false, project: false },
          },
          rules: { 'unicorn/no-useless-fallback-in-spread': 'error' },
        },
      ],
      fileContents: { [filename]: code },
      fix: true,
    });
    expect(result.fileCount).toBe(1);
    expect(result.ruleCount).toBe(1);
    expect(result.diagnostics).toEqual([]);
    const outputs = Object.values(result.output ?? {});
    expect(outputs).toHaveLength(1);
    const output = outputs[0]!;
    expect(output).toContain('/* keep */');
    for (const value of [
      undefined,
      null,
      false,
      0,
      '',
      'text',
      { value: 1 },
      [1, 2],
    ]) {
      expect(runInNewContext(output, { value })).toBe(
        runInNewContext(code, { value }),
      );
    }
  }
});
