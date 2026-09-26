// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester, type ValidTestCase } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';

const valid: ValidTestCase[] = [
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

valid.push(
  {
    code: "nativeDiffButtons.parentElement.after(\n\ttooltipped(\n\t\t{},\n\t\t<a className={cx('btn', isHidingWhitespace() && 'color-fg-subtle')} />,\n\t),\n);",
    filename: 'src/virtual.tsx',
  },
  {
    code: "nativeDiffButtons.parentElement.after(\n\ttooltipped(\n\t\t{},\n\t\t<>{cx('btn', isHidingWhitespace() && 'color-fg-subtle')}</>,\n\t),\n);",
    filename: 'src/virtual.tsx',
  },
  {
    code: 'foo(bar(baz(qux())));',
    filename,
    options: [{ max: 4 }],
  },
);

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

invalid.push(
  {
    code: '<Component value={foo(bar(baz(qux())))} />;',
    filename: 'src/virtual.tsx',
    errors: [
      {
        messageId: 'max-nested-calls',
        message: 'Call is nested too deeply. Maximum allowed is 3.',
      },
    ],
  },
  {
    code: 'mergeReports(await pMap(\n\tawait mergeWithFileConfigs(uniq(paths), inputOptions, configFiles),\n\tasync ({files, options, prettierOptions}) => runEslint(files, buildConfig(options, prettierOptions), {isQuiet: options.quiet}),\n));',
    filename,
    errors: [
      {
        messageId: 'max-nested-calls',
        message: 'Call is nested too deeply. Maximum allowed is 3.',
      },
    ],
  },
);

ruleTester.run('max-nested-calls', null as never, { valid, invalid });
