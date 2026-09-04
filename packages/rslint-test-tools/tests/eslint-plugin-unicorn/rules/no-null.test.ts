import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const js = (code: string, options: unknown[] = []) => ({
  code,
  filename: 'file.js',
  options,
});
const error = {
  messageId: 'error',
  message: 'Use `undefined` instead of `null`.',
};
const invalid = (code: string, options: unknown[] = [], errors = [error]) => ({
  code,
  filename: 'file.js',
  options,
  errors,
});

const ignoreArguments = [{ checkArguments: false }];
const checkStrictEquality = [{ checkStrictEquality: true }];

ruleTester.run('no-null', null as never, {
  valid: [
    js('let foo'),
    js('Object.create(null)'),
    js('Object.create(null, {foo: {value:1}})'),
    js('let insertedNode = parentNode.insertBefore(newNode, null)'),
    js('let insertedNode = parentNode?.insertBefore(newNode, null)'),
    js('const foo = "null";'),
    js('Object.create()'),
    js('Object.create(bar)'),
    js('Object.create("null")'),
    js('useRef(null)'),
    js('React.useRef(null)'),
    js('if (foo === null) {}'),
    js('if (null === foo) {}'),
    js('if (foo !== null) {}'),
    js('if (null !== foo) {}'),
    js('if (foo === null) {}', [{ checkStrictEquality: false }]),
    js('if (null === foo) {}', [{ checkStrictEquality: false }]),
    js('if (foo !== null) {}', [{ checkStrictEquality: false }]),
    js('if (null !== foo) {}', [{ checkStrictEquality: false }]),
    js('foo(null)', ignoreArguments),
    js('foo(bar, null)', ignoreArguments),
    js('drawingManager.setMap(null)', ignoreArguments),
    js('markers[index].setMap(null)', ignoreArguments),
    js('object?.method?.(null)', ignoreArguments),
    js('foo?.(null)', ignoreArguments),
    js('new HttpResponse(null)', ignoreArguments),
    js('new HttpResponse(body, null)', ignoreArguments),
    {
      ...js('foo = Object.create(null)'),
      languageOptions: { ecmaVersion: 2019 },
    },
  ],
  invalid: [
    invalid('const foo = null'),
    invalid('foo(null)'),
    invalid('if (foo == null) {}'),
    invalid('if (foo != null) {}'),
    invalid('if (null == foo) {}'),
    invalid('if (null != foo) {}'),
    invalid('function foo() {\n\treturn null;\n}'),
    invalid('let foo = null;'),
    invalid('var foo = null;'),
    invalid('var foo = 1, bar = null, baz = 2;'),
    invalid('const foo = null;'),
    invalid('const foo = null;', ignoreArguments),
    invalid('function foo() {\n\treturn null;\n}', ignoreArguments),
    invalid('if (foo === null) {}', [
      { checkArguments: false, checkStrictEquality: true },
    ]),
    invalid('foo([null])', ignoreArguments),
    invalid('foo(bar ?? null)', ignoreArguments),
    invalid('foo(...[null])', ignoreArguments),
    invalid('new HttpResponse([null])', ignoreArguments),
    invalid('if (foo === null) {}', checkStrictEquality),
    invalid('if (null === foo) {}', checkStrictEquality),
    invalid('if (foo !== null) {}', checkStrictEquality),
    invalid('if (null !== foo) {}', checkStrictEquality),
    invalid('new Object.create(null)'),
    invalid('new foo.insertBefore(bar, null)'),
    invalid('create(null)'),
    invalid('insertBefore(bar, null)'),
    invalid('Object["create"](null)'),
    invalid('foo["insertBefore"](bar, null)'),
    invalid('Object[create](null)'),
    invalid('foo[insertBefore](bar, null)'),
    invalid('Object[null](null)', [], [error, error]),
    invalid('Object.notCreate(null)'),
    invalid('foo.notInsertBefore(foo, null)'),
    invalid('NotObject.create(null)'),
    invalid('lib.Object.create(null)'),
    invalid('Object.create(...[null])'),
    invalid('Object.create(null, bar, extraArgument)'),
    invalid('foo.insertBefore(null)'),
    invalid('foo.insertBefore(foo, null, bar)'),
    invalid('foo.insertBefore(...[foo], null)'),
    invalid('foo.insertBefore(null, bar)'),
    invalid('Object.create(bar, null)'),
  ],
});
