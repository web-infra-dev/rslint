// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-block-statement-over-iife.js
import { RuleTester, type ValidTestCase } from '../rule-tester';

const ruleName = 'prefer-block-statement-over-iife';
const error = {
  messageId: ruleName,
  message: 'Prefer a block statement over an IIFE used only for scoping.',
};

function testCase(
  code: string,
  filename = 'src/virtual.js',
  sourceType?: 'script' | 'module',
): ValidTestCase {
  return {
    code,
    filename,
    languageOptions: sourceType ? { sourceType } : undefined,
  };
}

// The Go upstream suite also asserts the exact ranges and autofixes.
new RuleTester().run(ruleName, {} as never, {
  valid: [
    // The shared upstream runner also checks the empty file.
    testCase(''),
    testCase('const value = (() => { return getValue(); })();'),
    testCase('const value = (() => { getValue(); })();'),
    testCase('(() => value)();'),
    testCase('(async () => { await run(); })();'),
    testCase('(function * () { yield value; })();'),
    testCase('(function named() { run(); })();'),
    testCase('(function (value) { run(value); })(value);'),
    testCase('(() => { return value; })();'),
    testCase('(() => { if (condition) { return; } })();'),
    testCase('(() => { var value = 1; run(value); })();'),
    testCase('(() => { if (condition) { var value = 1; } })();'),
    testCase('(() => { eval(code); })();'),
    testCase('(() => { if (condition) { eval(code); } })();'),
    testCase('(() => { (eval)(code); })();'),
    testCase("(() => { (eval as any)('var x = 1'); })();", 'src/virtual.ts'),
    testCase("(() => { eval!('var x = 1'); })();", 'src/virtual.ts'),
    testCase('(function () { this.run(); })();'),
    testCase('(function () { arguments[0](); })();'),
    testCase('(function () { console.log(new.target); })();'),
    testCase('(function () { const run = () => this.run(); run(); })();'),
    testCase(
      '(function () { const getArguments = () => arguments; getArguments(); })();',
    ),
    testCase(
      '(function () { const getNewTarget = () => new.target; getNewTarget(); })();',
    ),
    testCase("(function () { 'use strict'; run(); })();"),
    testCase("(() => { 'use strict'; run(); })();"),
    testCase('(function () { run(); }).call(this);'),
    testCase('(() => { run(); })?.();'),
    testCase(
      '(() => {\n\tfunction foo() {}\n\tfoo();\n})();',
      'src/virtual.js',
      'script',
    ),
    testCase(
      '(() => {\n\tif (condition) {\n\t\tfunction foo() {}\n\t\tfoo();\n\t}\n})();',
      'src/virtual.js',
      'script',
    ),

    // Documentation examples. Both incorrect examples are in the suite below.
    testCase('{\n\tconst value = getValue();\n\trun(value);\n}'),
    testCase('const value = (() => {\n\treturn getValue();\n})();'),
    testCase('(async () => {\n\tawait run();\n})();'),
  ],
  invalid: [
    testCase('(() => { eval?.(code); })();'),
    testCase('(() => {\n\tconst value = getValue();\n\trun(value);\n})();'),
    testCase(
      '(function () {\n\tconst value = getValue();\n\trun(value);\n})();',
    ),
    testCase(
      'if (condition) {\n\t(() => {\n\t\tconst value = getValue();\n\t\trun(value);\n\t})();\n}',
    ),
    testCase('(() => {\n\t// Keep this comment.\n\trun();\n})();'),
    testCase('(() => {\n\tthis.run();\n})();'),
    testCase(
      '(() => {\n\tfunction nested() {\n\t\treturn value;\n\t}\n\n\trun(nested);\n})();',
    ),
    testCase(
      '(() => {\n\tfunction nested() {\n\t\tthis.run();\n\t}\n\n\trun(nested);\n})();',
    ),
    testCase(
      '(() => {\n\tconst value = getValue() satisfies number;\n\trun(value);\n})();',
      'src/virtual.ts',
    ),
    testCase(
      '(() => {\n\tfunction foo() {}\n\tfoo();\n})();',
      'src/virtual.js',
      'module',
    ),
    testCase('(/* comment */ () => {\n\trun();\n})();'),
    testCase('(() => {\n\trun();\n} /* comment */)();'),
    testCase('(() => {\n\trun();\n})() /* comment */;'),
  ].map((item) => ({ ...item, errors: [error] })),
});
