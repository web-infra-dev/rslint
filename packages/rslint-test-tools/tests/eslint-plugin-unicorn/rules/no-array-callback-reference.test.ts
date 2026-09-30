// Ported from eslint-plugin-unicorn v76.0.0 tests and documentation; see LICENSE.
// The suite wrapper checks messages and IDs; Go tests assert ranges and edits.
import { RuleTester } from '../rule-tester';

function valid(code: string, filename = 'file.js', options?: unknown[]) {
  return { code, filename, options };
}

function error(method: string, name?: string) {
  return name
    ? {
        messageId: 'error-with-name',
        message: `Do not pass function \`${name}\` directly to \`.${method}(…)\`.`,
      }
    : {
        messageId: 'error-without-name',
        message: `Do not pass function directly to \`.${method}(…)\`.`,
      };
}

function invalid(
  code: string,
  errors: ReturnType<typeof error>[],
  filename = 'file.js',
  options?: unknown[],
) {
  return { code, errors, filename, options };
}

// JavaScript and options (includes all not-function-types fixtures).
new RuleTester().run('no-array-callback-reference', {} as never, {
  valid: [
    valid('foo.every(element => fn(element))'),
    valid('foo.filter(element => fn(element))'),
    valid('foo.find(element => fn(element))'),
    valid('foo.findIndex(element => fn(element))'),
    valid('foo.findLast(element => fn(element))'),
    valid('foo.findLastIndex(element => fn(element))'),
    valid('foo.flatMap(element => fn(element))'),
    valid('foo.forEach(element => fn(element))'),
    valid('foo.map(element => fn(element))'),
    valid('foo.some(element => fn(element))'),
    valid('foo.reduce((accumulator, element) => fn(element))'),
    valid('foo.reduceRight((accumulator, element) => fn(element))'),
    valid('foo?.every(element => fn(element))'),
    valid('foo?.filter(element => fn(element))'),
    valid('foo?.find(element => fn(element))'),
    valid('foo?.findIndex(element => fn(element))'),
    valid('foo?.findLast(element => fn(element))'),
    valid('foo?.findLastIndex(element => fn(element))'),
    valid('foo?.flatMap(element => fn(element))'),
    valid('foo?.forEach(element => fn(element))'),
    valid('foo?.map(element => fn(element))'),
    valid('foo?.some(element => fn(element))'),
    valid('foo?.reduce((accumulator, element) => fn(element))'),
    valid('foo?.reduceRight((accumulator, element) => fn(element))'),
    valid('this.every(fn)'),
    valid('this.filter(fn)'),
    valid('this.find(fn)'),
    valid('this.findIndex(fn)'),
    valid('this.findLast(fn)'),
    valid('this.findLastIndex(fn)'),
    valid('this.flatMap(fn)'),
    valid('this.forEach(fn)'),
    valid('this.map(fn)'),
    valid('this.some(fn)'),
    valid('this.reduce(fn)'),
    valid('this.reduceRight(fn)'),
    valid('foo.find(Boolean)'),
    valid('foo.some(Boolean)'),
    valid('foo.map(String)'),
    valid('foo.map(Number)'),
    valid('foo.map(BigInt)'),
    valid('foo.map(Boolean)'),
    valid('foo.map(Symbol)'),
    valid('new foo.map(fn);'),
    valid('map(fn);'),
    valid("foo['map'](fn);"),
    valid('foo[map](fn);'),
    valid('foo.notListedMethod(fn);'),
    valid('foo.map();'),
    valid('foo.map(fn, extraArgument1, extraArgument2);'),
    valid('foo.map(...argumentsArray)'),
    valid('Promise.map(fn)'),
    valid('Promise.forEach(fn)'),
    valid('lodash.map(fn)'),
    valid('underscore.map(fn)'),
    valid('_.map(fn)'),
    valid('Async.map(list, fn)'),
    valid('async.map(list, fn)'),
    valid('React.Children.forEach(children, fn)'),
    valid('Children.forEach(children, fn)'),
    valid('Vue.filter(name, fn)'),
    valid('$(this).find(tooltip)'),
    valid('$.map(realArray, function(value, index) {});'),
    valid('$(this).filter(tooltip)'),
    valid('jQuery(this).find(tooltip)'),
    valid('jQuery.map(realArray, function(value, index) {});'),
    valid('jQuery(this).filter(tooltip)'),
    valid('Angular.forEach(list, fn)', 'file.js', [{ ignore: ['Angular'] }]),
    valid('P.map(list, fn)', 'file.js', [{ ignore: ['P'] }]),
    valid('myLib.utils.map(list, fn)', 'file.js', [
      { ignore: ['myLib.utils'] },
    ]),
    valid('myLib(args).map(fn)', 'file.js', [{ ignore: ['myLib'] }]),
    valid('Promise.map(list, fn)', 'file.js', [{ ignore: ['Angular'] }]),
    valid('lodash.map(list, fn)', 'file.js', [{ ignore: ['Angular'] }]),
    valid('foo.map([])'),
    valid('foo.map([element])'),
    valid('foo.map([...elements])'),
    valid('foo.map(1 + fn)'),
    valid('foo.map("length" in fn)'),
    valid('foo.map(fn instanceof Function)'),
    valid('foo.map(class ClassCantUseAsFunction {})'),
    valid('foo.map(0)'),
    valid('foo.map(1)'),
    valid('foo.map(0.1)'),
    valid('foo.map("")'),
    valid('foo.map("string")'),
    valid('foo.map(/regex/)'),
    valid('foo.map(null)'),
    valid('foo.map(0n)'),
    valid('foo.map(1n)'),
    valid('foo.map(true)'),
    valid('foo.map(false)'),
    valid('foo.map({})'),
    valid('foo.map(`templateLiteral`)'),
    valid('foo.map(undefined)'),
    valid('foo.map(- fn)'),
    valid('foo.map(+ fn)'),
    valid('foo.map(~ fn)'),
    valid('foo.map(typeof fn)'),
    valid('foo.map(void fn)'),
    valid('foo.map(delete foo.fn)'),
    valid('foo.map(++ fn)'),
    valid('foo.map(-- fn)'),
    valid('foo.map(a = fn)'),
    valid('foo.map(fn())'),
    valid('foo.map(new ClassReturnsFunction())'),
    valid('foo.map(new Function())'),
    valid('foo.map(fn``)'),
    valid('foo.map(this)'),
    valid('const query = {}; model.find(query)'),
    valid('const taskName = "task"; service.find(taskName)'),
    valid('const values = []; collection.map(values)'),
    valid('const NotCallable = class {}; collection.map(NotCallable)'),
    valid('const index = 1 + 1; collection.findIndex(index)'),
    valid('const query = {}; const criteria = query; model.find(criteria)'),
    valid('foo.map(() => {})'),
    valid('foo.map(function() {})'),
    valid('foo.map(function bar() {})'),
    valid('(async () => await foo.every(bar))()'),
    valid('(async () => await foo.filter(bar))()'),
    valid('(async () => await foo.find(bar))()'),
    valid('(async () => await foo.findIndex(bar))()'),
    valid('(async () => await foo.findLast(bar))()'),
    valid('(async () => await foo.findLastIndex(bar))()'),
    valid('(async () => await foo.flatMap(bar))()'),
    valid('(async () => await foo.forEach(bar))()'),
    valid('(async () => await foo.map(bar))()'),
    valid('(async () => await foo.some(bar))()'),
    valid('foo.map(function (a) {}.bind(bar))'),
    valid(
      'async function foo() {\n\tconst clientId = 20\n\tconst client = await oidc.Client.find(clientId)\n}',
    ),
    valid(
      'const results = collection\n\t.find({\n\t\t$and: [cursorQuery, params.query]\n\t}, {\n\t\tprojection: params.projection\n\t})\n\t.sort($sort)\n\t.limit(params.limit + 1)\n\t.toArray()',
    ),
    valid(
      "const EventsStore = types.model('EventsStore', {\n\tevents: types.optional(types.map(Event), {}),\n})",
    ),
    valid('const collection = new Set(); collection.forEach(callback);'),
    valid('const collection = new Map(); collection.forEach(callback);'),
    valid(
      'class Foo {} const collection = new Foo(); collection.map(callback);',
    ),
  ],
  invalid: [
    invalid('foo.every(fn)', [error('every', 'fn')]),
    invalid('foo.filter(fn)', [error('filter', 'fn')]),
    invalid('foo.find(fn)', [error('find', 'fn')]),
    invalid('foo.findIndex(fn)', [error('findIndex', 'fn')]),
    invalid('foo.findLast(fn)', [error('findLast', 'fn')]),
    invalid('foo.findLastIndex(fn)', [error('findLastIndex', 'fn')]),
    invalid('foo.flatMap(fn)', [error('flatMap', 'fn')]),
    invalid('foo.map(fn)', [error('map', 'fn')]),
    invalid('foo.some(fn)', [error('some', 'fn')]),
    invalid('foo?.every(fn)', [error('every', 'fn')]),
    invalid('foo?.filter(fn)', [error('filter', 'fn')]),
    invalid('foo?.find(fn)', [error('find', 'fn')]),
    invalid('foo?.findIndex(fn)', [error('findIndex', 'fn')]),
    invalid('foo?.findLast(fn)', [error('findLast', 'fn')]),
    invalid('foo?.findLastIndex(fn)', [error('findLastIndex', 'fn')]),
    invalid('foo?.flatMap(fn)', [error('flatMap', 'fn')]),
    invalid('foo?.map(fn)', [error('map', 'fn')]),
    invalid('foo?.some(fn)', [error('some', 'fn')]),
    invalid('foo.forEach(fn)', [error('forEach', 'fn')]),
    invalid('const callback = value => value; array.map(callback);', [
      error('map', 'callback'),
    ]),
    invalid('let query = {}; model.find(query);', [error('find', 'query')]),
    invalid('array.map(fn)', [error('map', 'fn')]),
    invalid('foo.map(element)', [error('map', 'element')]),
    invalid('items.map(fn)', [error('map', 'fn')]),
    invalid('items.map(item)', [error('map', 'item')]),
    invalid('items.map(items)', [error('map', 'items')]),
    invalid('items.map(index)', [error('map', 'index')]),
    invalid('items.map(item.fn)', [error('map')]),
    invalid('classes.map(fn)', [error('map', 'fn')]),
    invalid('indices.map(fn)', [error('map', 'fn')]),
    invalid('items.forEach(fn)', [error('forEach', 'fn')]),
    invalid('items.reduce(fn)', [error('reduce', 'fn')]),
    invalid('items.reduceRight(fn)', [error('reduceRight', 'fn')]),
    invalid('items.reduce(accumulator)', [error('reduce', 'accumulator')]),
    invalid('foo.reduce(fn)', [error('reduce', 'fn')]),
    invalid('foo.reduceRight(fn)', [error('reduceRight', 'fn')]),
    invalid('foo.every(fn, thisArgument)', [error('every', 'fn')]),
    invalid('foo.filter(fn, thisArgument)', [error('filter', 'fn')]),
    invalid('foo.find(fn, thisArgument)', [error('find', 'fn')]),
    invalid('foo.findIndex(fn, thisArgument)', [error('findIndex', 'fn')]),
    invalid('foo.findLast(fn, thisArgument)', [error('findLast', 'fn')]),
    invalid('foo.findLastIndex(fn, thisArgument)', [
      error('findLastIndex', 'fn'),
    ]),
    invalid('foo.flatMap(fn, thisArgument)', [error('flatMap', 'fn')]),
    invalid('foo.map(fn, thisArgument)', [error('map', 'fn')]),
    invalid('foo.some(fn, thisArgument)', [error('some', 'fn')]),
    invalid('foo.forEach(fn, thisArgument)', [error('forEach', 'fn')]),
    invalid('foo.reduce(fn, initialValue)', [error('reduce', 'fn')]),
    invalid('foo.reduceRight(fn, initialValue)', [error('reduceRight', 'fn')]),
    invalid('foo.reduce(Boolean, initialValue)', [error('reduce', 'Boolean')]),
    invalid('foo.reduceRight(Boolean, initialValue)', [
      error('reduceRight', 'Boolean'),
    ]),
    invalid('foo.forEach(Boolean)', [error('forEach', 'Boolean')]),
    invalid('foo.every(lib.fn)', [error('every')]),
    invalid('foo.filter(lib.fn)', [error('filter')]),
    invalid('foo.find(lib.fn)', [error('find')]),
    invalid('foo.findIndex(lib.fn)', [error('findIndex')]),
    invalid('foo.findLast(lib.fn)', [error('findLast')]),
    invalid('foo.findLastIndex(lib.fn)', [error('findLastIndex')]),
    invalid('foo.flatMap(lib.fn)', [error('flatMap')]),
    invalid('foo.map(lib.fn)', [error('map')]),
    invalid('foo.some(lib.fn)', [error('some')]),
    invalid('foo.reduce(lib.fn)', [error('reduce')]),
    invalid('foo.reduceRight(lib.fn)', [error('reduceRight')]),
    invalid('foo.map(a || b)', [error('map')]),
    invalid('array.map(condition ? toFile : toBuffer);', [
      error('map', 'toFile'),
      error('map', 'toBuffer'),
    ]),
    invalid(
      'function * foo() { array.map(condition ? (yield toFile) : toBuffer); }',
      [error('map'), error('map', 'toBuffer')],
    ),
    invalid(
      'async function foo() { array.map((await condition) ? toFile : toBuffer); }',
      [error('map', 'toFile'), error('map', 'toBuffer')],
    ),
    invalid('bar.map(fn)', [error('map', 'fn')]),
    invalid('bar.reduce(fn)', [error('reduce', 'fn')]),
    invalid('foo.map(lib.fn)', [error('map')]),
    invalid('foo.reduce(lib.fn)', [error('reduce')]),
    invalid(
      'const fn = async () => {\n\tawait Promise.all(foo.map(toPromise));\n}',
      [error('map', 'toPromise')],
    ),
    invalid(
      'async function fn() {\n\tfor await (const foo of bar.map(toPromise)) {}\n}',
      [error('map', 'toPromise')],
    ),
    invalid(
      'async function fn() {\n\tawait foo.reduce(foo, Promise.resolve())\n}',
      [error('reduce', 'foo')],
    ),
    invalid('const fn = (x, y) => x + y;\n[1, 2, 3].map(fn);', [
      error('map', 'fn'),
    ]),
    invalid('Other.forEach(fn)', [error('forEach', 'fn')], 'file.js', [
      { ignore: ['Angular'] },
    ]),
  ],
});

// Ternaries and upstream snapshot cases.
new RuleTester().run('no-array-callback-reference', {} as never, {
  valid: [
    valid('foo.map(_ ? () => {} : _ ? () => {} : () => {})'),
    valid('foo.reduce(_ ? () => {} : _ ? () => {} : () => {})'),
    valid('foo.every(_ ? Boolean : _ ? Boolean : Boolean)'),
    valid('foo.map(_ ? String : _ ? Number : Boolean)'),
  ],
  invalid: [
    invalid(
      'foo.map(\n\t_\n\t\t? String // This one should be ignored\n\t\t: callback\n);',
      [error('map', 'callback')],
    ),
    invalid(
      'foo.forEach(\n\t_\n\t\t? callbackA\n\t\t: _\n\t\t\t\t? callbackB\n\t\t\t\t: callbackC\n);',
      [
        error('forEach', 'callbackA'),
        error('forEach', 'callbackB'),
        error('forEach', 'callbackC'),
      ],
    ),
    invalid(
      "async function * foo () {\n\tfoo.map((0, bar));\n\tfoo.map(yield bar);\n\tfoo.map(yield* bar);\n\tfoo.map(() => bar);\n\tfoo.map(bar &&= baz);\n\tfoo.map(bar || baz);\n\tfoo.map(bar + bar);\n\tfoo.map(+ bar);\n\tfoo.map(++ bar);\n\tfoo.map(new Function(''));\n}",
      [error('map'), error('map'), error('map'), error('map')],
    ),
  ],
});

// Typed receivers.
new RuleTester().run('no-array-callback-reference', {} as never, {
  valid: [
    valid(
      'interface SearchService {\n\tfind(callback: Function): unknown;\n}\ndeclare const callback: Function;\ndeclare const service: SearchService;\nservice.find(callback);',
      'file.ts',
    ),
    valid(
      "class SearchService {\n\tfind(taskName: string): unknown {\n\t\treturn taskName;\n\t}\n}\nconst service = new SearchService();\nconst taskName = 'task';\nservice.find(taskName);",
      'file.ts',
    ),
    valid(
      'declare const callback: Function;\nclass Collection {\n\tmap(callback: Function) {}\n}\nconst collection = new Collection();\ncollection.map(callback);',
      'file.ts',
    ),
    valid(
      'interface Model {\n\tfind(query: object): unknown;\n}\ndeclare const AccountModel: Model;\nconst query = {};\nAccountModel.find(query);',
      'file.ts',
    ),
    valid(
      'interface NgMocks {\n\tfind(component: unknown): unknown;\n}\ndeclare const ngMocks: NgMocks;\ndeclare const MyComponent: unknown;\nngMocks.find(MyComponent);',
      'file.ts',
    ),
    valid(
      'declare const callback: Function;\ndeclare const collection: string[] | {map(callback: Function): unknown};\ncollection.map(callback);',
      'file.ts',
    ),
    valid(
      'declare const callback: Function;\ndeclare const set: Set<string>;\nset.forEach(callback);',
      'file.ts',
    ),
    valid(
      'declare const callback: Function;\ndeclare const map: Map<string, string>;\nmap.forEach(callback);',
      'file.ts',
    ),
    valid(
      'declare const callback: Function;\ndeclare const service: {find(callback: Function): unknown} | undefined;\nservice?.find(callback);',
      'file.ts',
    ),
    valid(
      'export {};\ntype Array<T> = {map(callback: Function): unknown};\ndeclare const callback: Function;\ndeclare const collection: Array<string>;\ncollection.map(callback);',
      'file.ts',
    ),
    valid(
      'export {};\nclass Uint8Array {\n\tmap(callback: Function) {}\n}\ndeclare const callback: Function;\nconst collection = new Uint8Array();\ncollection.map(callback);',
      'file.ts',
    ),
  ],
  invalid: [
    invalid(
      'declare const callback: Function; declare const array: string[]; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: readonly string[]; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: [string, string]; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: Array<string>; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: ReadonlyArray<string>; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: Uint8Array; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: string[] | readonly number[]; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: string[] | Uint8Array; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: string[] & {foo: string}; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: string[] | undefined; array?.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; function run<T extends string[]>(array: T) { array.map(callback); }',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; function run<T extends readonly string[]>(array: T) { array.map(callback); }',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; class Strings extends Array<string> {} const array = new Strings(); array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; interface S extends ReadonlyArray<string> {} declare const array: S; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; function run<T extends Uint8Array>(array: T) { array.map(callback); }',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: any; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: unknown; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const callback: Function; declare const array: MissingType; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
  ],
});

// TypeScript syntax.
new RuleTester().run('no-array-callback-reference', {} as never, {
  valid: [
    valid(
      "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.filter(isString);",
      'file.ts',
    ),
    valid(
      "const isString = (value: unknown): value is string => typeof value === 'string';\nfoo.filter(isString);",
      'file.ts',
    ),
    valid(
      "const isString = function (value: unknown): value is string {\n\treturn typeof value === 'string';\n};\nfoo.filter(isString);",
      'file.ts',
    ),
    valid(
      "const isString: (value: unknown) => value is string = value => typeof value === 'string';\nfoo.filter(isString);",
      'file.ts',
    ),
    valid(
      'function run(predicate: (value: unknown) => value is string) {\n\tfoo.filter(predicate);\n}',
      'file.ts',
    ),
    valid(
      'function runEvery(predicate: (value: unknown) => value is string) {\n\tfoo.every(predicate);\n}',
      'file.ts',
    ),
    valid(
      'function runFind(predicate: (value: unknown) => value is string) {\n\tfoo.find(predicate);\n}',
      'file.ts',
    ),
    valid(
      'function runFindLast(predicate: (value: unknown) => value is string) {\n\tfoo.findLast(predicate);\n}',
      'file.ts',
    ),
    valid(
      "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nconst guard: (value: unknown) => value is string = isString;\nfoo.filter(guard);",
      'file.ts',
    ),
    valid(
      "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.find(isString);",
      'file.ts',
    ),
    valid(
      "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.every(isString);",
      'file.ts',
    ),
    valid(
      "import {isString} from './guards';\nfoo.filter(isString);",
      'file.ts',
    ),
    valid("import {isString} from './guards';\nfoo.find(isString);", 'file.ts'),
    valid(
      "import {isString} from './guards';\nfoo.findLast(isString);",
      'file.ts',
    ),
    valid(
      "import {isString} from './guards';\nfoo.every(isString);",
      'file.ts',
    ),
    valid(
      "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfunction isNumber(value: unknown): value is number {\n\treturn typeof value === 'number';\n}\nfoo.filter(condition ? isString : isNumber);",
      'file.ts',
    ),
    valid('const query = {} as const;\nmodel.find(query);', 'file.ts'),
    valid(
      "const taskName = 'task' satisfies string;\nservice.find(taskName);",
      'file.ts',
    ),
    valid(
      'declare const collection: Set<string>; collection.forEach(callback);',
      'file.ts',
    ),
    valid(
      'declare const collection: Map<string, string>; collection.forEach(callback);',
      'file.ts',
    ),
    valid(
      'function run(collection: ReadonlySet<string>) { collection.forEach(callback); }',
      'file.ts',
    ),
    valid(
      'class Uint8Array {\n\tmap(callback: Function) {}\n}\nconst collection = new Uint8Array();\ncollection.map(callback);',
      'file.ts',
    ),
    valid(
      'declare const collection: string[] | Set<string>; collection.forEach(callback);',
      'file.ts',
    ),
  ],
  invalid: [
    invalid(
      'declare const array: string[]; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const array: readonly string[]; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const array: Array<string>; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const array: Uint8Array; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'const array = new Uint8Array(); array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const array: string[] | Uint8Array; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      'declare const array: string[] | undefined; array.map(callback);',
      [error('map', 'callback')],
      'file.ts',
    ),
    invalid(
      "function isString(value: unknown): boolean {\n\treturn typeof value === 'string';\n}\nfoo.filter(isString);",
      [error('filter', 'isString')],
      'file.ts',
    ),
    invalid(
      "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfoo.map(isString);",
      [error('map', 'isString')],
      'file.ts',
    ),
    invalid(
      "function isString(value: unknown): value is string {\n\treturn typeof value === 'string';\n}\nfunction isObject(value: unknown): boolean {\n\treturn typeof value === 'object';\n}\nfoo.filter(condition ? isString : isObject);",
      [error('filter', 'isObject')],
      'file.ts',
    ),
  ],
});

// Documentation examples.
new RuleTester().run('no-array-callback-reference', {} as never, {
  valid: [
    valid('const unicorn = x => x + 1;\n\nexport default unicorn;'),
    valid(
      'const unicorn = (x, y) => x + (y ? y : 1);\n\nexport default unicorn;',
    ),
    valid(
      "import unicorn from 'unicorn';\n\n[1, 2, 3].map(x => unicorn(x));\n//=> [2, 3, 4]",
    ),
    valid('const foo = array.map(element => callback(element));'),
    valid('const foo = array.map(Boolean);'),
    valid('array.forEach(element => {\n\tcallback(element);\n});'),
    valid('const foo = array.every(element => callback(element));'),
    valid('const foo = array.filter(element => callback(element));'),
    valid('const foo = array.filter(Boolean);'),
    valid('const foo = array.find(element => callback(element));'),
    valid('const index = array.findIndex(element => callback(element));'),
    valid('const foo = array.some(element => callback(element));'),
    valid(
      'const foo = array.reduce(\n\t(accumulator, element) => accumulator + callback(element),\n\t0\n);',
    ),
    valid(
      'const foo = array.reduceRight(\n\t(accumulator, element) => [\n\t\t...accumulator,\n\t\tcallback(element)\n\t],\n\t[]\n);',
    ),
    valid('const foo = array.flatMap(element => callback(element));'),
    // Upstream ignores factory calls, despite the incorrect marker in its documentation.
    valid("array.forEach(someFunction({foo: 'bar'}));"),
    // Upstream ignores factory calls, despite the incorrect marker in its documentation.
    valid(
      "const callback = someFunction({foo: 'bar'});\n\narray.forEach(element => {\n\tcallback(element);\n});",
    ),
    valid(
      'array.forEach(function (element) {\n\tcallback(element, this);\n}, thisArgument);',
    ),
    valid(
      "function readFile(filename) {\n\treturn fs.readFile(filename, 'utf8');\n}\n\nPromise.map(filenames, readFile);",
    ),
    valid(
      '/* eslint unicorn/no-array-callback-reference: ["error", {"ignore": ["Angular"]}] */\nAngular.forEach(list, fn); // Passes',
      'file.js',
      [{ ignore: ['Angular'] }],
    ),
  ],
  invalid: [
    invalid(
      "import unicorn from 'unicorn';\n\n[1, 2, 3].map(unicorn);\n//=> [2, 3, 4]",
      [error('map', 'unicorn')],
    ),
    invalid(
      "import unicorn from 'unicorn';\n\n[1, 2, 3].map(unicorn);\n//=> [2, 3, 5]",
      [error('map', 'unicorn')],
    ),
    invalid('const foo = array.map(callback);', [error('map', 'callback')]),
    invalid('array.forEach(callback);', [error('forEach', 'callback')]),
    invalid('const foo = array.every(callback);', [error('every', 'callback')]),
    invalid('const foo = array.filter(callback);', [
      error('filter', 'callback'),
    ]),
    invalid('const foo = array.find(callback);', [error('find', 'callback')]),
    invalid('const index = array.findIndex(callback);', [
      error('findIndex', 'callback'),
    ]),
    invalid('const foo = array.some(callback);', [error('some', 'callback')]),
    invalid('const foo = array.reduce(callback, 0);', [
      error('reduce', 'callback'),
    ]),
    invalid('const foo = array.reduceRight(callback, []);', [
      error('reduceRight', 'callback'),
    ]),
    invalid('const foo = array.flatMap(callback);', [
      error('flatMap', 'callback'),
    ]),
    invalid('array.forEach(callback, thisArgument);', [
      error('forEach', 'callback'),
    ]),
  ],
});
