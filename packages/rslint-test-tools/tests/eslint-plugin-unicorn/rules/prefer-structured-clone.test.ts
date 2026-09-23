// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';

const valid = (code: string) => ({ code, filename });

const jsonClone = (code: string, output: string) => ({
  code,
  filename,
  errors: [
    {
      messageId: 'prefer-structured-clone/error',
      message:
        'Prefer `structuredClone(…)` over `JSON.parse(JSON.stringify(…))` to create a deep clone.',
      suggestions: [
        {
          messageId: 'prefer-structured-clone/suggestion',
          desc: 'Switch to `structuredClone(…)`.',
          output,
        },
      ],
    },
  ],
});

const functionClone = (
  code: string,
  path: string,
  output: string,
  functions?: string[],
) => ({
  code,
  filename,
  ...(functions ? { options: [{ functions }] } : {}),
  errors: [
    {
      messageId: 'prefer-structured-clone/error',
      message:
        'Prefer `structuredClone(…)` over `' +
        path +
        '(…)` to create a deep clone.',
      suggestions: [
        {
          messageId: 'prefer-structured-clone/suggestion',
          desc: 'Switch to `structuredClone(…)`.',
          output,
        },
      ],
    },
  ],
});

ruleTester.run('prefer-structured-clone', null as never, {
  valid: [
    valid('structuredClone(foo)'),
    valid('JSON.parse(new JSON.stringify(foo))'),
    valid('new JSON.parse(JSON.stringify(foo))'),
    valid('JSON.parse(JSON.stringify())'),
    valid('JSON.parse(JSON.stringify(...foo))'),
    valid('JSON.parse(JSON.stringify(foo, extraArgument))'),
    valid('JSON.parse(...JSON.stringify(foo))'),
    valid('JSON.parse(JSON.stringify(foo), extraArgument)'),
    valid('JSON.parse(JSON.stringify?.(foo))'),
    valid('JSON.parse(JSON?.stringify(foo))'),
    valid('JSON.parse?.(JSON.stringify(foo))'),
    valid('JSON?.parse(JSON.stringify(foo))'),
    valid('JSON.parse(JSON.not_stringify(foo))'),
    valid('JSON.parse(not_JSON.stringify(foo))'),
    valid('JSON.not_parse(JSON.stringify(foo))'),
    valid('not_JSON.parse(JSON.stringify(foo))'),
    valid('JSON.stringify(JSON.parse(foo))'),
    valid('JSON.parse(JSON.stringify(foo, undefined, 2))'),
    valid('new _.cloneDeep(foo)'),
    valid('notMatchedFunction(foo)'),
    valid('_.cloneDeep()'),
    valid('_.cloneDeep(...foo)'),
    valid('_.cloneDeep(foo, extraArgument)'),
    valid('_.cloneDeep?.(foo)'),
    valid('_?.cloneDeep(foo)'),
  ],
  invalid: [
    jsonClone('JSON.parse(JSON.stringify(foo))', 'structuredClone(foo)'),
    jsonClone('JSON.parse(JSON.stringify(foo),)', 'structuredClone(foo,)'),
    jsonClone('JSON.parse(JSON.stringify(foo,))', 'structuredClone(foo)'),
    jsonClone('JSON.parse(JSON.stringify(foo,),)', 'structuredClone(foo,)'),
    jsonClone(
      'JSON.parse( ((JSON.stringify)) (foo))',
      'structuredClone(  foo)',
    ),
    jsonClone(
      '(( JSON.parse)) (JSON.stringify(foo))',
      '(( structuredClone)) (foo)',
    ),
    jsonClone(
      'JSON.parse(JSON.stringify( ((foo)) ))',
      'structuredClone( ((foo)) )',
    ),
    jsonClone(
      'function foo() {\n\treturn JSON\n\t\t.parse(\n\t\t\tJSON.\n\t\t\t\tstringify(\n\t\t\t\t\tbar,\n\t\t\t\t),\n\t\t);\n}',
      'function foo() {\n\treturn structuredClone(\n\t\t\t\n\t\t\t\t\tbar\n\t\t\t\t,\n\t\t);\n}',
    ),
    functionClone('_.cloneDeep(foo)', '_.cloneDeep', 'structuredClone(foo)'),
    functionClone(
      'lodash.cloneDeep(foo)',
      'lodash.cloneDeep',
      'structuredClone(foo)',
    ),
    functionClone(
      'lodash.cloneDeep(foo,)',
      'lodash.cloneDeep',
      'structuredClone(foo,)',
    ),
    functionClone(
      'myCustomDeepCloneFunction(foo,)',
      'myCustomDeepCloneFunction',
      'structuredClone(foo,)',
      ['myCustomDeepCloneFunction'],
    ),
    functionClone(
      'my.cloneDeep(foo,)',
      'my.cloneDeep',
      'structuredClone(foo,)',
      ['my.cloneDeep'],
    ),
    functionClone(
      'class A {\n\tconstructor() {\n\t\tthis.a = new.target.cloneDeep(foo);\n\t\tthis.b = import.meta.cloneDeep(foo);\n\t}\n}',
      'new.target.cloneDeep',
      'class A {\n\tconstructor() {\n\t\tthis.a = structuredClone(foo);\n\t\tthis.b = import.meta.cloneDeep(foo);\n\t}\n}',
      ['new.target.cloneDeep'],
    ),
    functionClone(
      'class A {\n\tconstructor() {\n\t\tthis.a = super.cloneDeep(foo);\n\t}\n}',
      'super.cloneDeep',
      'class A {\n\tconstructor() {\n\t\tthis.a = structuredClone(foo);\n\t}\n}',
      ['super.cloneDeep'],
    ),
  ],
});
