import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'Missing the separator argument.';
const valid = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string) => ({
  code,
  filename: 'file.js',
  errors: [{ message }],
});

ruleTester.run('require-array-join-separator', null as never, {
  valid: [
    valid('foo.join(",")'),
    valid('join()'),
    valid('foo.join(...[])'),
    valid('foo.join?.()'),
    valid('foo?.join?.()'),
    valid('foo[join]()'),
    valid('foo["join"]()'),
    valid('[].join.call(foo, ",")'),
    valid('[].join.call()'),
    valid('[].join.call(...[foo])'),
    valid('[].join?.call(foo)'),
    valid('[]?.join.call(foo)'),
    valid('[].join[call](foo)'),
    valid('[][join].call(foo)'),
    valid('[,].join.call(foo)'),
    valid('[].join.notCall(foo)'),
    valid('[].notJoin.call(foo)'),
    valid('Array.prototype.join.call(foo, "")'),
    valid('Array.prototype.join.call()'),
    valid('Array.prototype.join.call(...[foo])'),
    valid('Array.prototype.join?.call(foo)'),
    valid('Array.prototype?.join.call(foo)'),
    valid('Array?.prototype.join.call(foo)'),
    valid('Array.prototype.join[call](foo, "")'),
    valid('Array.prototype[join].call(foo)'),
    valid('Array[prototype].join.call(foo)'),
    valid('Array.prototype.join.notCall(foo)'),
    valid('Array.prototype.notJoin.call(foo)'),
    valid('Array.notPrototype.join.call(foo)'),
    valid('NotArray.prototype.join.call(foo)'),
    valid('path.join(__dirname, "./foo.js")'),
  ],
  invalid: [
    invalid('foo.join()'),
    invalid('[].join.call(foo)'),
    invalid('[].join.call(foo,)'),
    invalid('[].join.call(foo , );'),
    invalid('Array.prototype.join.call(foo)'),
    invalid('Array.prototype.join.call(foo, )'),
    invalid('foo?.join()'),
  ],
});
