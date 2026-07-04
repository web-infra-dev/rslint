import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const message = 'Useless case in switch statement.';
const error = { message };

const lines = (...parts: string[]) => parts.join('\n');
const js = (code: string) => ({ code, filename: 'file.js' });
const ts = (code: string) => ({ code, filename: 'file.ts' });

ruleTester.run('no-useless-switch-case', null as never, {
  valid: [
    js(
      lines(
        'switch (foo) {',
        '\tcase a:',
        '\tcase b:',
        '\t\thandleDefaultCase();',
        '\t\tbreak;',
        '}',
      ),
    ),
    js(
      lines(
        'switch (foo) {',
        '\tcase a:',
        '\t\thandleCaseA();',
        '\t\tbreak;',
        '\tdefault:',
        '\t\thandleDefaultCase();',
        '\t\tbreak;',
        '}',
      ),
    ),
    js(
      lines(
        'switch (foo) {',
        '\tcase a:',
        '\t\thandleCaseA();',
        '\tdefault:',
        '\t\thandleDefaultCase();',
        '\t\tbreak;',
        '}',
      ),
    ),
    js(
      lines(
        'switch (foo) {',
        '\tcase a:',
        '\t\tbreak;',
        '\tdefault:',
        '\t\thandleDefaultCase();',
        '\t\tbreak;',
        '}',
      ),
    ),
    js(
      lines(
        'switch (foo) {',
        '\tcase a:',
        '\t\thandleCaseA();',
        '\t\t// Fallthrough',
        '\tdefault:',
        '\t\thandleDefaultCase();',
        '\t\tbreak;',
        '}',
      ),
    ),
    js(
      lines(
        'switch (foo) {',
        '\tcase a:',
        '\tdefault:',
        '\t\thandleDefaultCase();',
        '\t\tbreak;',
        '\tcase b:',
        '\t\thandleCaseB();',
        '\t\tbreak;',
        '}',
      ),
    ),
    js(
      lines(
        'switch (1) {',
        '\t\t// This is not useless',
        '\t\tcase 1:',
        '\t\tdefault:',
        "\t\t\t\tconsole.log('1')",
        '\t\tcase 1:',
        "\t\t\t\tconsole.log('2')",
        '}',
      ),
    ),
  ],
  invalid: [
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase undefined:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase null:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...ts(
        lines(
          'switch (foo) {',
          '\tcase undefined:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...ts(
        lines(
          'switch (foo) {',
          '\tcase null:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...ts(
        lines(
          'switch (foo) {',
          '\tcase null:',
          '\tcase undefined:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error, error],
    },
    {
      ...ts(
        lines(
          'switch (foo) {',
          '\tcase a:',
          '\tcase undefined:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error, error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase undefined:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...ts(
        lines(
          'switch (foo) {',
          '\tcase a:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase a:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase a: {',
          '\t}',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase a: {',
          '\t\t;;',
          '\t\t{',
          '\t\t\t;;',
          '\t\t\t{',
          '\t\t\t\t;;',
          '\t\t\t}',
          '\t\t}',
          '\t}',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase a:',
          '\tcase (( b ))         :',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error, error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase a:',
          '\tcase b:',
          '\t\thandleCaseAB();',
          '\t\tbreak;',
          '\tcase d:',
          '\tcase d:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error, error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase a:',
          '\tcase b:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error, error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\t// eslint-disable-next-line',
          '\tcase a:',
          '\tcase b:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
    {
      ...js(
        lines(
          'switch (foo) {',
          '\tcase a:',
          '\t// eslint-disable-next-line',
          '\tcase b:',
          '\tdefault:',
          '\t\thandleDefaultCase();',
          '\t\tbreak;',
          '}',
        ),
      ),
      errors: [error],
    },
  ],
});
