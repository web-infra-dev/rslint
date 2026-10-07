// Ported from eslint-plugin-unicorn v77.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/no-negated-condition.js
import { RuleTester } from '../rule-tester';

const error = {
  messageId: 'no-negated-condition',
  message: 'Unexpected negated condition.',
};

new RuleTester().run('no-negated-condition', {} as never, {
  valid: [
    // Upstream snapshots.
    { code: 'if (a) {}', filename: 'src/virtual.js' },
    { code: 'if (a) {} else {}', filename: 'src/virtual.js' },
    { code: 'if (!a) {}', filename: 'src/virtual.js' },
    { code: 'if (!a) {} else if (b) {}', filename: 'src/virtual.js' },
    { code: 'if (!a) {} else if (b) {} else {}', filename: 'src/virtual.js' },
    { code: 'if (a == b) {}', filename: 'src/virtual.js' },
    { code: 'if (a == b) {} else {}', filename: 'src/virtual.js' },
    { code: 'if (a != b) {}', filename: 'src/virtual.js' },
    { code: 'if (a != b) {} else if (b) {}', filename: 'src/virtual.js' },
    {
      code: 'if (a != b) {} else if (b) {} else {}',
      filename: 'src/virtual.js',
    },
    { code: 'if (a !== b) {}', filename: 'src/virtual.js' },
    { code: 'if (a === b) {} else {}', filename: 'src/virtual.js' },
    { code: 'a ? b : c', filename: 'src/virtual.js' },
    // Upstream documentation.
    {
      code: 'if (a) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'if (a === b) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}',
      filename: 'src/virtual.js',
    },
    { code: 'a ? b : c', filename: 'src/virtual.js' },
    {
      code: 'if (a == b) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}',
      filename: 'src/virtual.js',
    },
    { code: 'if (!a) {\n\tdoSomething();\n}', filename: 'src/virtual.js' },
    {
      code: 'if (!a) {\n\tdoSomething();\n} else if (b) {\n\tdoSomethingElse();\n}',
      filename: 'src/virtual.js',
    },
    { code: 'if (a != b) {\n\tdoSomething();\n}', filename: 'src/virtual.js' },
  ],
  invalid: [
    // Upstream comments.
    {
      code: '!x ? /* one */ 1 : /* two */ 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? /* two */ 2 : /* one */ 1',
    },
    {
      code: '!x ? 1 /* one */ : 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: '!x ? 1 : 2 /* two */',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: '!x ? (1) /* one */ : (2)',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: '!x ? 1 : (2) /* two */',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: '!x ? 1 // one\n : 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: 'x != y ? 1 /* one */ : 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: 'x !== y ? 1 : 2 /* two */',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: 'if (!x) {one();} /* one */ else {two();}',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: 'if (!x) {one();} else {two();} /* two */',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: 'if (!x) one(); /* one */ else two();',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    {
      code: 'if (!x) one(); else two(); // two',
      filename: 'src/virtual.js',
      errors: [error],
      output: null,
    },
    // Upstream snapshots.
    {
      code: 'if (!a) {;} else {;}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (a) {;} else {;}',
    },
    {
      code: 'if (a != b) {;} else {;}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (a == b) {;} else {;}',
    },
    {
      code: 'if (a !== b) {;} else {;}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (a === b) {;} else {;}',
    },
    {
      code: '!a ? b : c',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a ? c : b',
    },
    {
      code: 'a != b ? c : d',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a == b ? d : c',
    },
    {
      code: 'a !== b ? c : d',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a === b ? d : c',
    },
    {
      code: '(( !a )) ? b : c',
      filename: 'src/virtual.js',
      errors: [error],
      output: '(( a )) ? c : b',
    },
    {
      code: '!(( a )) ? b : c',
      filename: 'src/virtual.js',
      errors: [error],
      output: '(( a )) ? c : b',
    },
    {
      code: 'if(!(( a ))) b(); else c();',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if( a ) {c();} else {b();}',
    },
    {
      code: 'if((( !a ))) b(); else c();',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if((( a ))) {c();} else {b();}',
    },
    {
      code: 'function a() {return!a ? b : c}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'function a() {return a ? c : b}',
    },
    {
      code: 'function a() {return!(( a )) ? b : c}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'function a() {return (( a )) ? c : b}',
    },
    {
      code: 'function a() {\n\treturn ! // comment\n\t\ta ? b : c;\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'function a() {\n\treturn (  // comment\n\t\ta ? c : b);\n}',
    },
    {
      code: 'function a() {\n\treturn (! // ReturnStatement argument is parenthesized\n\t\ta ? b : c);\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output:
        'function a() {\n\treturn ( // ReturnStatement argument is parenthesized\n\t\ta ? c : b);\n}',
    },
    {
      code: 'function a() {\n\treturn (\n\t\t! // UnaryExpression argument is parenthesized\n\t\ta) ? b : c;\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output:
        'function a() {\n\treturn (\n\t\t // UnaryExpression argument is parenthesized\n\t\ta) ? c : b;\n}',
    },
    {
      code: 'function a() {\n\tthrow ! // comment\n\t\ta ? b : c;\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'function a() {\n\tthrow (  // comment\n\t\ta ? c : b);\n}',
    },
    {
      code: '!a ? b : c ? d : e',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a ? c ? d : e : b',
    },
    {
      code: '!a ? b : (( c ? d : e ))',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a ? (( c ? d : e )) : b',
    },
    {
      code: 'a\n![] ? b : c',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a\n;[] ? c : b',
    },
    {
      code: 'a\n!+b ? c : d',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a\n;+b ? d : c',
    },
    {
      code: 'a\n!(b) ? c : d',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a\n;(b) ? d : c',
    },
    {
      code: 'a\n!b ? c : d',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a\nb ? d : c',
    },
    {
      code: 'if (!a)\n\tb()\nelse\n\tc()',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (a)\n\t{c()}\nelse\n\t{b()}',
    },
    {
      code: 'if(!a) b(); else c()',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if(a) {c()} else {b();}',
    },
    {
      code: 'function fn() {\n\tif(!a) b(); else return\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'function fn() {\n\tif(a) {return} else {b();}\n}',
    },
    {
      code: 'if(!a) {b()} else {c()}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if(a) {c()} else {b()}',
    },
    {
      code: 'if(!!a) b(); else c();',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if(a) {b();} else {c();}',
    },
    {
      code: '(!!a) ? b() : c();',
      filename: 'src/virtual.js',
      errors: [error],
      output: '(a) ? b() : c();',
    },
    {
      code: 'function fn() {\n\treturn!a !== b ? c : d\n\treturn((!((a)) != b)) ? c : d\n}',
      filename: 'src/virtual.js',
      errors: [error, error],
      output:
        'function fn() {\n\treturn!a === b ? d : c\n\treturn((!((a)) == b)) ? d : c\n}',
    },
    {
      code: 'if (!a) {\n\tb();\n} else if (!c) {\n\td();\n} else {\n\te();\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (!a) {\n\tb();\n} else if (c) {\n\te();\n} else {\n\td();\n}',
    },
    {
      code: '!x ? /* one */ 1 : 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? 2 : /* one */ 1',
    },
    {
      code: '!x ? 1 : /* two */ 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? /* two */ 2 : 1',
    },
    {
      code: '!x ? /* one */ /* first */ 1 : /* two */ /* second */ 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? /* two */ /* second */ 2 : /* one */ /* first */ 1',
    },
    {
      code: '!x ? /* one */ 1 : /* two */ 1',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? /* two */ 1 : /* one */ 1',
    },
    {
      code: '!x ? /* one */ ((1)) : /* two */ ((2))',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? /* two */ ((2)) : /* one */ ((1))',
    },
    {
      code: '!x ? (/* one */ 1) : (/* two */ 2)',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? (/* two */ 2) : (/* one */ 1)',
    },
    {
      code: '!x ? /* one */ (1 /* inside */) : /* two */ 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? /* two */ 2 : /* one */ (1 /* inside */)',
    },
    {
      code: '!x ? /* one */ first(/* inside */ 1) : /* two */ second(2)',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? /* two */ second(2) : /* one */ first(/* inside */ 1)',
    },
    {
      code: 'x != y ? /* one */ 1 : /* two */ 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x == y ? /* two */ 2 : /* one */ 1',
    },
    {
      code: 'x !== y ? /* one */ 1 : /* two */ 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x === y ? /* two */ 2 : /* one */ 1',
    },
    {
      code: '!x ? // one\n 1 : // two\n 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? // two\n 2 : // one\n 1',
    },
    {
      code: '!x ? /* one */ 1 : // two\n 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? // two\n 2 : /* one */ 1',
    },
    {
      code: '!x ? // one\r\n 1 : 2',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'x ? 2 : // one\r\n 1',
    },
    {
      code: '!x ? /* outer */ (!y ? /* one */ 1 : /* two */ 2) : /* three */ 3',
      filename: 'src/virtual.js',
      errors: [error, error],
      output: 'x ? /* three */ 3 : /* outer */ (y ? /* two */ 2 : /* one */ 1)',
    },
    {
      code: 'if (!x) /* one */ {one();} else /* two */ {two();}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (x) /* two */ {two();} else /* one */ {one();}',
    },
    {
      code: 'if (!x) /* one */ one(); else /* two */ two();',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (x) {/* two */ two();} else {/* one */ one();}',
    },
    {
      code: 'if (!x) // one\n one(); else // two\n two();',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (x) {// two\n two();} else {// one\n one();}',
    },
    {
      code: 'if (!x) { /* one */ one(); } else { /* two */ two(); }',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (x) { /* two */ two(); } else { /* one */ one(); }',
    },
    {
      code: 'if (!x) /* one */ one(); else {two();}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (x) {two();} else {/* one */ one();}',
    },
    {
      code: 'if (!x) {one();} else /* two */ two();',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (x) {/* two */ two();} else {one();}',
    },
    {
      code: '!x ? /* one */ (one as number) : /* two */ two!',
      filename: 'src/virtual.ts',
      errors: [error],
      output: 'x ? /* two */ two! : /* one */ (one as number)',
    },
    // Upstream documentation.
    {
      code: 'if (!a) {\n\tdoSomethingC();\n} else {\n\tdoSomethingB();\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'if (a) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}',
    },
    {
      code: 'if (a !== b) {\n\tdoSomethingC();\n} else {\n\tdoSomethingB();\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output:
        'if (a === b) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}',
    },
    {
      code: '!a ? c : b',
      filename: 'src/virtual.js',
      errors: [error],
      output: 'a ? b : c',
    },
    {
      code: 'if (a != b) {\n\tdoSomethingC();\n} else {\n\tdoSomethingB();\n}',
      filename: 'src/virtual.js',
      errors: [error],
      output:
        'if (a == b) {\n\tdoSomethingB();\n} else {\n\tdoSomethingC();\n}',
    },
  ],
});
