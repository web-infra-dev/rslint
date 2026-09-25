// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';

const valid = [
  'foo();',
  'foo(bar(baz()));',
  'foo(bar(), baz(), qux());',
  'query().filter().map().toArray();',
  'foo()[bar()]().baz();',
  'foo(() => bar(baz(qux())));',
  'foo(bar(class {field = baz(qux());}));',
  'foo(bar(class {static {baz(qux());}}));',
  'new Foo(new Bar(new Baz()));',
  'await foo(await bar(await baz()));',
  'foo?.(bar?.(baz?.()));',
  'foo(...bar(baz()));',
  'foo(condition ? bar(baz()) : qux());',
].map((code) => ({ code, filename }));

valid.push({
  code: 'foo(bar(baz(qux())));',
  filename,
  options: [{ max: 4 }],
});

const invalid = [
  { code: 'foo(bar(baz(qux())));', max: 3 },
  { code: 'foo(bar(baz()));', max: 2, options: [{ max: 2 }] },
  { code: 'new Foo(new Bar(new Baz(new Qux())));', max: 3 },
  { code: 'await foo(await bar(await baz(await qux())));', max: 3 },
  { code: 'foo?.(bar?.(baz?.(qux?.())));', max: 3 },
  { code: 'foo(...bar(baz(qux())));', max: 3 },
  { code: 'foo(condition ? bar(baz(qux())) : zed());', max: 3 },
  { code: 'foo(class {field = bar(baz(qux(zed())));});', max: 3 },
].map(({ code, max, options }) => ({
  code,
  filename,
  ...(options ? { options } : {}),
  errors: [
    {
      messageId: 'max-nested-calls',
      message: `Call is nested too deeply. Maximum allowed is ${max}.`,
    },
  ],
}));

ruleTester.run('max-nested-calls', null as never, { valid, invalid });
