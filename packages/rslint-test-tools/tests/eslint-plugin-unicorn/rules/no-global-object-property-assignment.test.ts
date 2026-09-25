// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const js = (code: string) => ({ code, filename: 'src/virtual.js' });
const ts = (code: string) => ({ code, filename: 'src/virtual.ts' });
const error = {
  messageId: 'no-global-object-property-assignment',
  message: 'Do not assign properties on the global object.',
};

ruleTester.run('no-global-object-property-assignment', null as never, {
  valid: [
    js('globalThis.foo'),
    js('window.foo()'),
    js('globalThis = value'),
    js('globalThis[property] = value'),
    js('delete globalThis.foo'),
    js('Object.assign(globalThis, {foo: value})'),
    js('Reflect.set(globalThis, "foo", value)'),
    js('function test(window) { window.foo = 1; }'),
    js('const global = {}; global.foo = 1;'),
    js('const globalThis = {}; globalThis.foo = 1;'),
    js('const self = {}; self.foo++;'),
    js('const root = globalThis; root.foo = 1;'),
  ],
  invalid: [
    js('globalThis.foo = 1'),
    js('window.foo += 1'),
    js('self.foo ||= value'),
    js('global.foo++'),
    js('globalThis["foo"] = 1'),
    js('({foo: globalThis.foo} = object);'),
    js('[globalThis.foo] = array'),
    js('({...globalThis.foo} = object)'),
    js('[...globalThis.foo] = array'),
    js('for (globalThis.foo of iterable) {}'),
    js('for (globalThis.foo in object) {}'),
    ts('globalThis!.foo = 1'),
    ts('(globalThis as any).foo = 1'),
    ts('(<any>globalThis).foo = 1'),
    ts('(globalThis satisfies any).foo = 1'),
    ts('globalThis.foo! = 1'),
  ].map((testCase) => ({ ...testCase, errors: [error] })),
});
