import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const lines = (...parts: string[]) => parts.join('\n');
const js = (code: string) => ({ code, filename: 'file.js' });
const invalid = (code: string, message: string) => ({
  ...js(code),
  errors: [{ message }],
});

const enforceMessage = (name: string) =>
  `Use \`new ${name}()\` instead of \`${name}()\`.`;
const disallowMessage = (name: string) =>
  `Use \`${name}()\` instead of \`new ${name}()\`.`;
const disallowCallOrNewMessage = (name: string) =>
  `\`${name}\` is not a function or constructor.`;
const dateMessage = 'Use `String(new Date())` instead of `Date()`.';

const enforceNewObjects = [
  'Object',
  'Array',
  'ArrayBuffer',
  'DataView',
  'Date',
  'Function',
  'Map',
  'WeakMap',
  'Set',
  'WeakSet',
  'Promise',
  'RegExp',
  'SharedArrayBuffer',
  'Proxy',
  'WeakRef',
  'FinalizationRegistry',
  'DisposableStack',
  'AsyncDisposableStack',
  'Error',
  'AggregateError',
  'EvalError',
  'RangeError',
  'ReferenceError',
  'SuppressedError',
  'SyntaxError',
  'TypeError',
  'URIError',
  'Intl.Collator',
  'Intl.DateTimeFormat',
  'Intl.DisplayNames',
  'Intl.DurationFormat',
  'Intl.ListFormat',
  'Intl.Locale',
  'Intl.NumberFormat',
  'Intl.PluralRules',
  'Intl.RelativeTimeFormat',
  'Intl.Segmenter',
  'Temporal.Duration',
  'Temporal.Instant',
  'Temporal.PlainDate',
  'Temporal.PlainDateTime',
  'Temporal.PlainMonthDay',
  'Temporal.PlainTime',
  'Temporal.PlainYearMonth',
  'Temporal.ZonedDateTime',
  'WebAssembly.Module',
  'WebAssembly.Instance',
  'WebAssembly.Memory',
  'WebAssembly.Table',
  'WebAssembly.Global',
  'WebAssembly.Tag',
  'WebAssembly.Exception',
  'WebAssembly.CompileError',
  'WebAssembly.LinkError',
  'WebAssembly.RuntimeError',
  'Float16Array',
  'Float32Array',
  'Float64Array',
  'Int8Array',
  'Int16Array',
  'Int32Array',
  'BigInt64Array',
  'BigUint64Array',
  'Uint8Array',
  'Uint16Array',
  'Uint32Array',
  'Uint8ClampedArray',
];

const disallowNewObjects = ['BigInt', 'Boolean', 'Number', 'String', 'Symbol'];

const disallowCallOrNewObjects = [
  'Temporal.Now',
  'WebAssembly',
  'WebAssembly.JSTag',
];

const createShadowedCallTest = (object: string) => {
  const [objectName, propertyName] = object.split('.');

  if (!propertyName) {
    return lines(
      `const ${object} = function() {};`,
      `const foo = ${object}();`,
    );
  }

  return lines(
    `const ${objectName} = {${propertyName}() {}};`,
    `const foo = ${object}();`,
  );
};

const createShadowedNewTest = (object: string) => {
  const [objectName, propertyName] = object.split('.');

  if (!propertyName) {
    return lines(
      `const ${object} = function() {};`,
      `const foo = new ${object}();`,
    );
  }

  return lines(
    `const ${objectName} = {${propertyName}: class {}};`,
    `const foo = new ${object}();`,
  );
};

const valid = [
  js('const foo = new Object()'),
  js('const foo = new Array()'),
  js('const foo = Array?.()'),
  js('const foo = Map?.()'),
  js('const foo = Date?.()'),
  js('const foo = globalThis?.Date()'),
  js('const foo = Intl.DateTimeFormat?.()'),
  js('const foo = Intl?.DateTimeFormat()'),
  js('const foo = Temporal.PlainDate?.(2024, 1, 1)'),
  js('const foo = WebAssembly.Module?.(buffer)'),
  js('const foo = WebAssembly?.Module(buffer)'),
  js('const foo = BigInt()'),
  js('const foo = Boolean()'),
  js('const foo = Number()'),
  js('const foo = String()'),
  js('const foo = Symbol()'),
  js('const foo = new Intl.DateTimeFormat()'),
  js('const foo = new globalThis.Intl.DateTimeFormat()'),
  js("const foo = new Intl.DisplayNames('en', {type: 'language'})"),
  js("const foo = new Intl.Locale('en')"),
  js('const foo = new Intl.Segmenter()'),
  js('const foo = new Temporal.PlainDate(2024, 1, 1)'),
  js('const foo = new globalThis.Temporal.PlainDate(2024, 1, 1)'),
  js("const foo = new Temporal.ZonedDateTime(0n, 'UTC')"),
  js('const foo = Temporal.Now.instant()'),
  js('const foo = new WebAssembly.Module(buffer)'),
  js('const foo = new globalThis.WebAssembly.Module(buffer)'),
  js('const foo = new WebAssembly.Memory({initial: 1})'),
  js('const foo = new WebAssembly.CompileError()'),
  js('const foo = WebAssembly.instantiate(buffer)'),
  js('const foo = WebAssembly.JSTag'),
  js(lines("import { Map } from 'immutable';", 'const m = Map();')),
  js(lines("const {Map} = require('immutable');", 'const foo = Map();')),
  js(
    lines("const {String} = require('guitar');", 'const lowE = new String();'),
  ),
  js(lines("import {String} from 'guitar';", 'const lowE = new String();')),
  js('new Foo();Bar();'),
  js('Foo();new Bar();'),
  js('const isObject = v => Object(v) === v;'),
  js('const isObject = v => globalThis.Object(v) === v;'),
  js('(x) !== Object(x)'),
];

for (const name of [
  'ArrayBuffer',
  'BigInt64Array',
  'BigUint64Array',
  'DataView',
  'Error',
  'Float16Array',
  'Float32Array',
  'Float64Array',
  'Function',
  'Int8Array',
  'Int16Array',
  'Int32Array',
  'Map',
  'WeakMap',
  'Set',
  'WeakSet',
  'Promise',
  'RegExp',
  'Uint8Array',
  'Uint16Array',
  'Uint32Array',
  'Uint8ClampedArray',
  'AggregateError',
  'TypeError',
  'SuppressedError',
  'DisposableStack',
  'AsyncDisposableStack',
]) {
  valid.push(js(`const foo = new ${name}()`));
}

for (const object of [...enforceNewObjects, ...disallowCallOrNewObjects]) {
  valid.push(js(createShadowedCallTest(object)));
}

for (const object of [...disallowNewObjects, ...disallowCallOrNewObjects]) {
  valid.push(js(createShadowedNewTest(object)));
}

const invalidCases = [
  invalid('const object = (Object)();', enforceMessage('Object')),
  invalid('const isObject = v => Object(v) == v;', enforceMessage('Object')),
  invalid('const symbol = new (Symbol)("");', disallowMessage('Symbol')),
  invalid(
    'const symbol = new /* comment */ Symbol("");',
    disallowMessage('Symbol'),
  ),
  invalid('const symbol = new Symbol;', disallowMessage('Symbol')),
  {
    ...invalid('const s = new Symbol()!;', disallowMessage('Symbol')),
    filename: 'file.ts',
  },
  invalid('new globalThis.String()', disallowMessage('String')),
  invalid('new global.String()', disallowMessage('String')),
  invalid('new self.String()', disallowMessage('String')),
  invalid('new window.String()', disallowMessage('String')),
  invalid(
    lines('const {String} = globalThis;', 'new String();'),
    disallowMessage('String'),
  ),
  invalid(
    lines(
      'const {String: RenamedString} = globalThis;',
      'new RenamedString();',
    ),
    disallowMessage('String'),
  ),
  invalid(
    lines('const RenamedString = globalThis.String;', 'new RenamedString();'),
    disallowMessage('String'),
  ),
  invalid('globalThis.Array()', enforceMessage('Array')),
  invalid('global.Array()', enforceMessage('Array')),
  invalid('self.Array()', enforceMessage('Array')),
  invalid('window.Array()', enforceMessage('Array')),
  invalid(
    lines('const {Array: RenamedArray} = globalThis;', 'RenamedArray();'),
    enforceMessage('Array'),
  ),
  invalid('const foo = Error("Foo bar")', enforceMessage('Error')),
  invalid('const foo = (( Map ))()', enforceMessage('Map')),
  invalid(
    "const foo = Map([['foo', 'bar'], ['unicorn', 'rainbow']])",
    enforceMessage('Map'),
  ),
  invalid(
    "const foo = Intl.DisplayNames('en', {type: 'language'})",
    enforceMessage('Intl.DisplayNames'),
  ),
  invalid("const foo = Intl.Locale('en')", enforceMessage('Intl.Locale')),
  invalid(
    'const foo = Temporal.PlainDate(2024, 1, 1)',
    enforceMessage('Temporal.PlainDate'),
  ),
  invalid(
    'const foo = globalThis.Temporal.PlainDate(2024, 1, 1)',
    enforceMessage('Temporal.PlainDate'),
  ),
  invalid(
    'const foo = Temporal.Now()',
    disallowCallOrNewMessage('Temporal.Now'),
  ),
  invalid(
    'const foo = Temporal.Now?.()',
    disallowCallOrNewMessage('Temporal.Now'),
  ),
  invalid(
    'const foo = Temporal?.Now()',
    disallowCallOrNewMessage('Temporal.Now'),
  ),
  invalid(
    'const foo = new Temporal.Now()',
    disallowCallOrNewMessage('Temporal.Now'),
  ),
  invalid('const foo = WebAssembly()', disallowCallOrNewMessage('WebAssembly')),
  invalid(
    'const foo = new WebAssembly()',
    disallowCallOrNewMessage('WebAssembly'),
  ),
  invalid(
    'const foo = WebAssembly.JSTag()',
    disallowCallOrNewMessage('WebAssembly.JSTag'),
  ),
  invalid(
    'const foo = new WebAssembly.JSTag()',
    disallowCallOrNewMessage('WebAssembly.JSTag'),
  ),
  invalid('const foo = new BigInt(123)', disallowMessage('BigInt')),
  invalid('const foo = new Boolean()', disallowMessage('Boolean')),
  invalid('const foo = new Number()', disallowMessage('Number')),
  invalid("const foo = new Number('123')", disallowMessage('Number')),
  invalid('const foo = new String()', disallowMessage('String')),
  invalid('const foo = new Symbol()', disallowMessage('Symbol')),
  invalid('const foo = Date();', dateMessage),
  invalid('const foo = globalThis.Date();', dateMessage),
  invalid('const foo = Date(/*comment*/);', dateMessage),
  invalid('const foo = globalThis/*comment*/.Date();', dateMessage),
  invalid('const foo = Date(bar);', dateMessage),
];

for (const name of [
  'Object',
  'Array',
  'ArrayBuffer',
  'BigInt64Array',
  'BigUint64Array',
  'DataView',
  'AggregateError',
  'EvalError',
  'RangeError',
  'ReferenceError',
  'SuppressedError',
  'SyntaxError',
  'TypeError',
  'URIError',
  'DisposableStack',
  'AsyncDisposableStack',
  'Float16Array',
  'Float32Array',
  'Float64Array',
  'Function',
  'Int8Array',
  'Int16Array',
  'Int32Array',
  'WeakMap',
  'Set',
  'WeakSet',
  'Promise',
  'RegExp',
  'Uint8Array',
  'Uint16Array',
  'Uint32Array',
  'Uint8ClampedArray',
  'Intl.Collator',
  'Intl.DateTimeFormat',
  'Intl.DurationFormat',
  'Intl.ListFormat',
  'Intl.NumberFormat',
  'Intl.PluralRules',
  'Intl.RelativeTimeFormat',
  'Intl.Segmenter',
  'Temporal.Duration',
  'Temporal.Instant',
  'Temporal.PlainDateTime',
  'Temporal.PlainMonthDay',
  'Temporal.PlainTime',
  'Temporal.PlainYearMonth',
  'Temporal.ZonedDateTime',
  'WebAssembly.Module',
  'WebAssembly.Instance',
  'WebAssembly.Memory',
  'WebAssembly.Table',
  'WebAssembly.Global',
  'WebAssembly.Tag',
  'WebAssembly.Exception',
  'WebAssembly.CompileError',
  'WebAssembly.LinkError',
  'WebAssembly.RuntimeError',
]) {
  invalidCases.push(invalid(`const foo = ${name}()`, enforceMessage(name)));
}

ruleTester.run('new-for-builtins', null as never, {
  valid,
  invalid: invalidCases,
});
