import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const js = (code: string) => ({ code, filename: 'file.js' });
const invalidIdentifier = (code: string, parameter = 'foo') => ({
  code,
  filename: 'file.js',
  errors: [
    {
      messageId: 'identifier',
      message: `Do not use an object literal as default for parameter \`${parameter}\`.`,
      data: { parameter },
    },
  ],
});
const invalidNonIdentifier = (code: string) => ({
  code,
  filename: 'file.js',
  errors: [
    {
      messageId: 'non-identifier',
      message: 'Do not use an object literal as default.',
    },
  ],
});

ruleTester.run('no-object-as-default-parameter', null as never, {
  valid: [
    js('const abc = {};'),
    js('const abc = {foo: 123};'),
    js('function abc(foo) {}'),
    js('function abc(foo = null) {}'),
    js('function abc(foo = undefined) {}'),
    js('function abc(foo = 123) {}'),
    js('function abc(foo = true) {}'),
    js('function abc(foo = "bar") {}'),
    js('function abc(foo = 123, bar = "foo") {}'),
    js('function abc(foo = {}) {}'),
    js('function abc({foo = 123} = {}) {}'),
    js('(function abc() {})(foo = {a: 123})'),
    js('const abc = foo => {};'),
    js('const abc = (foo = null) => {};'),
    js('const abc = (foo = undefined) => {};'),
    js('const abc = (foo = 123) => {};'),
    js('const abc = (foo = true) => {};'),
    js('const abc = (foo = "bar") => {};'),
    js('const abc = (foo = 123, bar = "foo") => {};'),
    js('const abc = (foo = {}) => {};'),
    js('const abc = ({a = true, b = "foo"}) => {};'),
    js('const abc = function(foo = 123) {}'),
    js('const {abc = {foo: 123}} = bar;'),
    js('const {abc = {null: "baz"}} = bar;'),
    js('const {abc = {foo: undefined}} = undefined;'),
    js('const abc = ([{foo = false, bar = 123}]) => {};'),
    js('const abc = ({foo = {a: 123}}) => {};'),
    js('const abc = ([foo = {a: 123}]) => {};'),
    js('const abc = ({foo: bar = {a: 123}}) => {};'),
    js('const abc = () => (foo = {a: 123});'),
    js('class A {\n\t[foo = {a: 123}]() {}\n}'),
    js('class A extends (foo = {a: 123}) {\n\ta() {}\n}'),
  ],
  invalid: [
    invalidIdentifier('function abc(foo = {a: 123}) {}'),
    invalidIdentifier('async function * abc(foo = {a: 123}) {}'),
    invalidIdentifier('function abc(foo = {a: false}) {}'),
    invalidIdentifier('function abc(foo = {a: "bar"}) {}'),
    invalidIdentifier('function abc(foo = {a: "bar", b: {c: true}}) {}'),
    invalidIdentifier('const abc = (foo = {a: false}) => {};'),
    invalidIdentifier('const abc = (foo = {a: 123, b: false}) => {};'),
    invalidIdentifier(
      'const abc = (foo = {a: false, b: 1, c: "test", d: null}) => {};',
    ),
    invalidIdentifier('const abc = function(foo = {a: 123}) {}'),
    invalidIdentifier('class A {\n\tabc(foo = {a: 123}) {}\n}'),
    invalidIdentifier('class A {\n\tconstructor(foo = {a: 123}) {}\n}'),
    invalidIdentifier('class A {\n\tset abc(foo = {a: 123}) {}\n}'),
    invalidIdentifier('class A {\n\tstatic abc(foo = {a: 123}) {}\n}'),
    invalidIdentifier('class A {\n\t* abc(foo = {a: 123}) {}\n}'),
    invalidIdentifier('class A {\n\tstatic async * abc(foo = {a: 123}) {}\n}'),
    invalidIdentifier('class A {\n\t[foo = {a: 123}](foo = {a: 123}) {}\n}'),
    invalidIdentifier('const A = class {\n\tabc(foo = {a: 123}) {}\n}'),
    invalidIdentifier('object = {\n\tabc(foo = {a: 123}) {}\n};'),
    invalidNonIdentifier('const A = class {\n\tabc({a} = {a: 123}) {}\n}'),

    // Upstream snapshot cases.
    invalidIdentifier('/**/function abc(foo = {a: 123}) {}'),
    invalidIdentifier('const abc = (foo = {a: false}) => {};'),
    invalidNonIdentifier('function abc({a} = {a: 123}) {}'),
    invalidNonIdentifier('function abc([a] = {a: 123}) {}'),
  ],
});
