/**
 * @fileoverview Tests for keyword-spacing rule.
 * @author Toru Nagashima
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/keyword-spacing/keyword-spacing._js_.test.ts
 *   packages/eslint-plugin/rules/keyword-spacing/keyword-spacing._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('keyword-spacing', null as never, { valid, invalid })`.
 *  - The upstream helper builders are evaluated to their final literal forms:
 *      `BOTH` -> `{ before: true, after: true }`
 *      `NEITHER` -> `{ before: false, after: false }`
 *      `override(kw, v)` -> `{ before: v.before===false, after: v.after===false, overrides: { [kw]: v } }`
 *      `expectedBefore(kw)` / `expectedAfter(kw)` / `expectedBeforeAndAfter(kw)` and the
 *      `unexpected*` variants -> their `{ messageId, data: { value: kw } }` arrays.
 *    The `keyword-spacing` message templates carry a `{{value}}` placeholder, so each
 *    error keeps its `data: { value }` for the RuleTester to interpolate.
 *  - `parserOptions` (ecmaVersion / sourceType / ecmaFeatures.jsx) dropped — rslint
 *    resolves via tsconfig; the RuleTester routes a fixture to `.tsx` when real JSX is
 *    present. `parser: tsParser` markers dropped (every fixture is parsed by ts-go).
 *  - `OVERRIDES_WITH` (a `parserOptions` spread for sloppy-mode `with`) dropped along
 *    with the other parser options.
 *  - `name` / `rule` / `lang` / `linterOptions` dropped.
 *
 * KNOWN GAPS: none. Every upstream fixture parses cleanly under rslint's ts-go parser
 * and the rule's output matches upstream for all cases, so the entire upstream suite is
 * in the green set above. This was checked, not assumed: the upstream tests rely on
 * sloppy-mode / older-ES syntax in several places (`with` statements, `using` /
 * `await using` declarations, decorators, the `<Thing>x` type-assertion form), and the
 * RuleTester aborts the whole batch — failing loud and naming the offending fixture — if
 * ANY fixture is unparseable under ts-go's strict/module semantics. The suite passes, so
 * no fixture is a ts-go syntax error and no case had to be isolated.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('keyword-spacing', null as never, {
  valid: [
    // ---- from keyword-spacing._js_.test.ts ----
    { code: 'import { a } from "foo"' },
    { code: 'import { a as b } from "foo"' },
    { code: 'import { "a" as b } from "foo"' },
    { code: 'import{ a }from"foo"', options: [ { before: false, after: false } ] },
    { code: 'import{ a as b }from"foo"', options: [ { before: false, after: false } ] },
    { code: 'import{ "a"as b }from"foo"', options: [ { before: false, after: false } ] },
    { code: 'import{ "a" as b }from"foo"', options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ] },
    { code: 'import { a as b } from "foo"', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'import { "a"as b } from "foo"', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'let a; export { a };' },
    { code: 'export { "a" } from "foo";' },
    { code: 'let a; export { a as b };' },
    { code: 'let a; export { a as "b" };' },
    { code: 'export { "a" as b } from "foo";' },
    { code: 'export { "a" as "b" } from "foo";' },
    { code: 'let a; export{ a };', options: [ { before: false, after: false } ] },
    { code: 'export{ "a" }from"foo";', options: [ { before: false, after: false } ] },
    { code: 'let a; export{ a as b };', options: [ { before: false, after: false } ] },
    { code: 'let a; export{ a as"b" };', options: [ { before: false, after: false } ] },
    { code: 'export{ "a"as b }from"foo";', options: [ { before: false, after: false } ] },
    { code: 'export{ "a"as"b" }from"foo";', options: [ { before: false, after: false } ] },
    { code: 'let a; export{ a as "b" };', options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ] },
    { code: 'export{ "a" as b }from"foo";', options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ] },
    { code: 'export{ "a" as "b" }from"foo";', options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ] },
    { code: 'let a; export { a as b };', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'let a; export { a as"b" };', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'export { "a"as b } from "foo";', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'export { "a"as"b" } from "foo";', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'import * as a from "foo"' },
    { code: 'import*as a from"foo"', options: [ { before: false, after: false } ] },
    { code: 'import* as a from"foo"', options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ] },
    { code: 'import *as a from "foo"', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'export * as a from "foo"' },
    { code: 'export * as "a" from "foo"' },
    { code: 'export*as a from"foo"', options: [ { before: false, after: false } ] },
    { code: 'export*as"a"from"foo"', options: [ { before: false, after: false } ] },
    { code: 'export* as a from"foo"', options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ] },
    { code: 'export* as "a"from"foo"', options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ] },
    { code: 'export *as a from "foo"', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'export *as"a" from "foo"', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: '{} async function foo() {}' },
    { code: '{}async function foo() {}', options: [ { before: false, after: false } ] },
    { code: '{} async function foo() {}', options: [ { before: false, after: false, overrides: { async: { before: true, after: true } } } ] },
    { code: '{}async function foo() {}', options: [ { before: true, after: true, overrides: { async: { before: false, after: false } } } ] },
    { code: '{} async () => {}' },
    { code: '{}async () => {}', options: [ { before: false, after: false } ] },
    { code: '{} async () => {}', options: [ { before: false, after: false, overrides: { async: { before: true, after: true } } } ] },
    { code: '{}async () => {}', options: [ { before: true, after: true, overrides: { async: { before: false, after: false } } } ] },
    { code: '({async [b]() {}})' },
    { code: '({async[b]() {}})', options: [ { before: false, after: false } ] },
    { code: '({async [b]() {}})', options: [ { before: false, after: false, overrides: { async: { before: true, after: true } } } ] },
    { code: '({async[b]() {}})', options: [ { before: true, after: true, overrides: { async: { before: false, after: false } } } ] },
    { code: 'class A {a(){} async [b]() {}}' },
    { code: 'class A {a(){}async[b]() {}}', options: [ { before: false, after: false } ] },
    { code: 'class A {a(){} async [b]() {}}', options: [ { before: false, after: false, overrides: { async: { before: true, after: true } } } ] },
    { code: 'class A {a(){}async[b]() {}}', options: [ { before: true, after: true, overrides: { async: { before: false, after: false } } } ] },
    { code: '[async function foo() {}]' },
    { code: '[ async function foo() {}]', options: [ { before: false, after: false } ] },
    { code: '() =>async function foo() {}' },
    { code: '() => async function foo() {}', options: [ { before: false, after: false } ] },
    { code: '{async function foo() {} }' },
    { code: '{ async function foo() {} }', options: [ { before: false, after: false } ] },
    { code: '(0,async function foo() {})' },
    { code: '(0, async function foo() {})', options: [ { before: false, after: false } ] },
    { code: 'a[async function foo() {}]' },
    { code: '({[async function foo() {}]: 0})' },
    { code: 'a[ async function foo() {}]', options: [ { before: false, after: false } ] },
    { code: '({[ async function foo() {}]: 0})', options: [ { before: false, after: false } ] },
    { code: '({ async* foo() {} })' },
    { code: '({ async *foo() {} })', options: [ { before: false, after: false } ] },
    { code: '({a:async function foo() {} })' },
    { code: '({a: async function foo() {} })', options: [ { before: false, after: false } ] },
    { code: ';async function foo() {};' },
    { code: '; async function foo() {} ;', options: [ { before: false, after: false } ] },
    { code: 'async() => {}' },
    { code: 'async () => {}', options: [ { before: false, after: false } ] },
    { code: '(async function foo() {})' },
    { code: '( async function foo() {})', options: [ { before: false, after: false } ] },
    { code: 'a =async function foo() {}' },
    { code: 'a = async function foo() {}', options: [ { before: false, after: false } ] },
    { code: '!async function foo() {}' },
    { code: '! async function foo() {}', options: [ { before: false, after: false } ] },
    { code: '`${async function foo() {}}`' },
    { code: '`${ async function foo() {}}`', options: [ { before: false, after: false } ] },
    { code: '<Foo onClick={async function foo() {}} />' },
    { code: '<Foo onClick={ async function foo() {}} />', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { {} await +1 }' },
    { code: 'async function wrap() { {}await +1 }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { {} await +1 }', options: [ { before: false, after: false, overrides: { await: { before: true, after: true } } } ] },
    { code: 'async function wrap() { {}await +1 }', options: [ { before: true, after: true, overrides: { await: { before: false, after: false } } } ] },
    { code: 'async function wrap() { [await a] }' },
    { code: 'async function wrap() { [ await a] }', options: [ { before: false, after: false } ] },
    { code: 'async () =>await a' },
    { code: 'async () => await a', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { {await a } }' },
    { code: 'async function wrap() { { await a } }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { (0,await a) }' },
    { code: 'async function wrap() { (0, await a) }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { a[await a] }' },
    { code: 'async function wrap() { ({[await a]: 0}) }' },
    { code: 'async function wrap() { a[ await a] }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { ({[ await a]: 0}) }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { ({a:await a }) }' },
    { code: 'async function wrap() { ({a: await a }) }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { ;await a; }' },
    { code: 'async function wrap() { ; await a ; }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { (await a) }' },
    { code: 'async function wrap() { ( await a) }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { a =await a }' },
    { code: 'async function wrap() { a = await a }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { a+await a }' },
    { code: 'async function wrap() { a + await a }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { a<await a }' },
    { code: 'async function wrap() { a < await a }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { a>await a }' },
    { code: 'async function wrap() { a > await a }', options: [ { before: false, after: false } ] },
    { code: "async function wrap() { !await'a' }" },
    { code: "async function wrap() { ! await 'a' }", options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { `${await a}` }' },
    { code: 'async function wrap() { `${ await a}` }', options: [ { before: false, after: false } ] },
    { code: 'async function wrap() { <Foo onClick={await a} /> }' },
    { code: 'async function wrap() { <Foo onClick={ await a} /> }', options: [ { before: false, after: false } ] },
    'A: for (;;) { {} break A; }',
    { code: 'A: for(;;) { {}break A; }', options: [ { before: false, after: false } ] },
    { code: 'A: for(;;) { {} break A; }', options: [ { before: false, after: false, overrides: { break: { before: true, after: true } } } ] },
    { code: 'A: for (;;) { {}break A; }', options: [ { before: true, after: true, overrides: { break: { before: false, after: false } } } ] },
    'for (;;) {break}',
    { code: 'for(;;) { break }', options: [ { before: false, after: false } ] },
    'for (;;) { ;break; }',
    { code: 'for(;;) { ; break ; }', options: [ { before: false, after: false } ] },
    'switch (a) { case 0: {} case +1: }',
    'switch (a) { case 0: {} case (1): }',
    { code: 'switch(a) { case 0: {}case+1: }', options: [ { before: false, after: false } ] },
    { code: 'switch(a) { case 0: {}case(1): }', options: [ { before: false, after: false } ] },
    { code: 'switch(a) { case 0: {} case +1: }', options: [ { before: false, after: false, overrides: { case: { before: true, after: true } } } ] },
    { code: 'switch (a) { case 0: {}case+1: }', options: [ { before: true, after: true, overrides: { case: { before: false, after: false } } } ] },
    'switch (a) {case 0: }',
    { code: 'switch(a) { case 0: }', options: [ { before: false, after: false } ] },
    'switch (a) { case 0: ;case 1: }',
    { code: 'switch(a) { case 0: ; case 1: }', options: [ { before: false, after: false } ] },
    'try {} catch {}',
    { code: 'try{}catch{}', options: [ { before: false, after: false } ] },
    { code: 'try{} catch {}', options: [ { before: false, after: false, overrides: { catch: { before: true, after: true } } } ] },
    { code: 'try {}catch{}', options: [ { before: true, after: true, overrides: { catch: { before: false, after: false } } } ] },
    'try {}\ncatch {}',
    { code: 'try{}\ncatch{}', options: [ { before: false, after: false } ] },
    'try {} catch (e) {}',
    { code: 'try{}catch(e) {}', options: [ { before: false, after: false } ] },
    { code: 'try{} catch (e) {}', options: [ { before: false, after: false, overrides: { catch: { before: true, after: true } } } ] },
    { code: 'try {}catch(e) {}', options: [ { before: true, after: true, overrides: { catch: { before: false, after: false } } } ] },
    'try {}\ncatch (e) {}',
    { code: 'try{}\ncatch(e) {}', options: [ { before: false, after: false } ] },
    { code: '{} class Bar {}' },
    { code: '(class {})' },
    { code: '{}class Bar {}', options: [ { before: false, after: false } ] },
    { code: '(class{})', options: [ { before: false, after: false } ] },
    { code: '{} class Bar {}', options: [ { before: false, after: false, overrides: { class: { before: true, after: true } } } ] },
    { code: '{}class Bar {}', options: [ { before: true, after: true, overrides: { class: { before: false, after: false } } } ] },
    { code: '[class {}]' },
    { code: '[ class{}]', options: [ { before: false, after: false } ] },
    { code: '() =>class {}' },
    { code: '() => class{}', options: [ { before: false, after: false } ] },
    { code: '{class Bar {} }' },
    { code: '{ class Bar {} }', options: [ { before: false, after: false } ] },
    { code: '(0,class {})' },
    { code: '(0, class{})', options: [ { before: false, after: false } ] },
    { code: 'a[class {}]' },
    { code: '({[class {}]: 0})' },
    { code: 'a[ class{}]', options: [ { before: false, after: false } ] },
    { code: '({[ class{}]: 0})', options: [ { before: false, after: false } ] },
    { code: '({a:class {} })' },
    { code: '({a: class{} })', options: [ { before: false, after: false } ] },
    { code: ';class Bar {};' },
    { code: '; class Bar {} ;', options: [ { before: false, after: false } ] },
    { code: '( class{})', options: [ { before: false, after: false } ] },
    { code: 'a =class {}' },
    { code: 'a = class{}', options: [ { before: false, after: false } ] },
    { code: 'a+class {}' },
    { code: 'a + class{}', options: [ { before: false, after: false } ] },
    { code: 'a<class {}' },
    { code: 'a < class{}', options: [ { before: false, after: false } ] },
    { code: 'a>class {}' },
    { code: 'a > class{}', options: [ { before: false, after: false } ] },
    { code: '!class {}' },
    { code: '! class{}', options: [ { before: false, after: false } ] },
    { code: '`${class {}}`' },
    { code: '`${ class{}}`', options: [ { before: false, after: false } ] },
    { code: '<Foo onClick={class {}} />' },
    { code: '<Foo onClick={ class{}} />', options: [ { before: false, after: false } ] },
    { code: 'class C {\n#x;\nfoo() {\nfor (this.#x of bar){}}}', options: [ { before: false } ] },
    { code: 'class C {\n#x;\nfoo() {\nfor (this.#x in bar){}}}', options: [ { before: false } ] },
    { code: '{} const [a] = b' },
    { code: '{} const {a} = b' },
    { code: '{}const[a] = b', options: [ { before: false, after: false } ] },
    { code: '{}const{a} = b', options: [ { before: false, after: false } ] },
    { code: '{} const [a] = b', options: [ { before: false, after: false, overrides: { const: { before: true, after: true } } } ] },
    { code: '{} const {a} = b', options: [ { before: false, after: false, overrides: { const: { before: true, after: true } } } ] },
    { code: '{}const[a] = b', options: [ { before: true, after: true, overrides: { const: { before: false, after: false } } } ] },
    { code: '{}const{a} = b', options: [ { before: true, after: true, overrides: { const: { before: false, after: false } } } ] },
    { code: '{const a = b}' },
    { code: '{ const a = b}', options: [ { before: false, after: false } ] },
    { code: ';const a = b;' },
    { code: '; const a = b ;', options: [ { before: false, after: false } ] },
    'A: for (;;) { {} continue A; }',
    { code: 'A: for(;;) { {}continue A; }', options: [ { before: false, after: false } ] },
    { code: 'A: for(;;) { {} continue A; }', options: [ { before: false, after: false, overrides: { continue: { before: true, after: true } } } ] },
    { code: 'A: for (;;) { {}continue A; }', options: [ { before: true, after: true, overrides: { continue: { before: false, after: false } } } ] },
    'for (;;) {continue}',
    { code: 'for(;;) { continue }', options: [ { before: false, after: false } ] },
    'for (;;) { ;continue; }',
    { code: 'for(;;) { ; continue ; }', options: [ { before: false, after: false } ] },
    '{} debugger',
    { code: '{}debugger', options: [ { before: false, after: false } ] },
    { code: '{} debugger', options: [ { before: false, after: false, overrides: { debugger: { before: true, after: true } } } ] },
    { code: '{}debugger', options: [ { before: true, after: true, overrides: { debugger: { before: false, after: false } } } ] },
    '{debugger}',
    { code: '{ debugger }', options: [ { before: false, after: false } ] },
    ';debugger;',
    { code: '; debugger ;', options: [ { before: false, after: false } ] },
    'switch (a) { case 0: {} default: }',
    { code: 'switch(a) { case 0: {}default: }', options: [ { before: false, after: false } ] },
    { code: 'switch(a) { case 0: {} default: }', options: [ { before: false, after: false, overrides: { default: { before: true, after: true } } } ] },
    { code: 'switch (a) { case 0: {}default: }', options: [ { before: true, after: true, overrides: { default: { before: false, after: false } } } ] },
    'switch (a) {default:}',
    { code: 'switch(a) { default: }', options: [ { before: false, after: false } ] },
    'switch (a) { case 0: ;default: }',
    { code: 'switch(a) { case 0: ; default: }', options: [ { before: false, after: false } ] },
    '{} delete foo.a',
    { code: '{}delete foo.a', options: [ { before: false, after: false } ] },
    { code: '{} delete foo.a', options: [ { before: false, after: false, overrides: { delete: { before: true, after: true } } } ] },
    { code: '{}delete foo.a', options: [ { before: true, after: true, overrides: { delete: { before: false, after: false } } } ] },
    '[delete foo.a]',
    { code: '[ delete foo.a]', options: [ { before: false, after: false } ] },
    { code: '(() =>delete foo.a)' },
    { code: '(() => delete foo.a)', options: [ { before: false, after: false } ] },
    '{delete foo.a }',
    { code: '{ delete foo.a }', options: [ { before: false, after: false } ] },
    '(0,delete foo.a)',
    { code: '(0, delete foo.a)', options: [ { before: false, after: false } ] },
    'a[delete foo.a]',
    { code: '({[delete foo.a]: 0})' },
    { code: 'a[ delete foo.a]', options: [ { before: false, after: false } ] },
    { code: '({[ delete foo.a]: 0})', options: [ { before: false, after: false } ] },
    '({a:delete foo.a })',
    { code: '({a: delete foo.a })', options: [ { before: false, after: false } ] },
    ';delete foo.a',
    { code: '; delete foo.a', options: [ { before: false, after: false } ] },
    '(delete foo.a)',
    { code: '( delete foo.a)', options: [ { before: false, after: false } ] },
    'a =delete foo.a',
    { code: 'a = delete foo.a', options: [ { before: false, after: false } ] },
    'a+delete foo.a',
    { code: 'a + delete foo.a', options: [ { before: false, after: false } ] },
    'a<delete foo.a',
    { code: 'a < delete foo.a', options: [ { before: false, after: false } ] },
    'a>delete foo.a',
    { code: 'a > delete foo.a', options: [ { before: false, after: false } ] },
    '!delete(foo.a)',
    { code: '! delete (foo.a)', options: [ { before: false, after: false } ] },
    { code: '`${delete foo.a}`' },
    { code: '`${ delete foo.a}`', options: [ { before: false, after: false } ] },
    { code: '<Foo onClick={delete foo.a} />' },
    { code: '<Foo onClick={ delete foo.a} />', options: [ { before: false, after: false } ] },
    '{} do {} while (true)',
    { code: '{}do{}while(true)', options: [ { before: false, after: false } ] },
    { code: '{} do {}while(true)', options: [ { before: false, after: false, overrides: { do: { before: true, after: true } } } ] },
    { code: '{}do{} while (true)', options: [ { before: true, after: true, overrides: { do: { before: false, after: false } } } ] },
    '{}\ndo\n{} while (true)',
    { code: '{}\ndo\n{}while(true)', options: [ { before: false, after: false } ] },
    '{do {} while (true)}',
    { code: '{ do{}while(true) }', options: [ { before: false, after: false } ] },
    ';do; while (true)',
    { code: '; do ;while(true)', options: [ { before: false, after: false } ] },
    'if (a) {} else {}',
    'if (a) {} else if (b) {}',
    'if (a) {} else (0)',
    'if (a) {} else []',
    'if (a) {} else +1',
    'if (a) {} else "a"',
    { code: 'if(a){}else{}', options: [ { before: false, after: false } ] },
    { code: 'if(a){}else if(b) {}', options: [ { before: false, after: false } ] },
    { code: 'if(a) {}else(0)', options: [ { before: false, after: false } ] },
    { code: 'if(a) {}else[]', options: [ { before: false, after: false } ] },
    { code: 'if(a) {}else+1', options: [ { before: false, after: false } ] },
    { code: 'if(a) {}else"a"', options: [ { before: false, after: false } ] },
    { code: 'if(a) {} else {}', options: [ { before: false, after: false, overrides: { else: { before: true, after: true } } } ] },
    { code: 'if (a) {}else{}', options: [ { before: true, after: true, overrides: { else: { before: false, after: false } } } ] },
    'if (a) {}\nelse\n{}',
    { code: 'if(a) {}\nelse\n{}', options: [ { before: false, after: false } ] },
    { code: 'if(a){ }else{ }', options: [ { before: false, after: true, overrides: { else: { after: false }, if: { after: false } } } ] },
    { code: 'if(a){ }else{ }', options: [ { before: true, after: false, overrides: { else: { before: false }, if: { before: false } } } ] },
    'if (a);else;',
    { code: 'if(a); else ;', options: [ { before: false, after: false } ] },
    { code: 'var a = 0; {} export {a}' },
    { code: '{} export default a' },
    { code: '{} export * from "a"' },
    { code: 'var a = 0; {}export{a}', options: [ { before: false, after: false } ] },
    { code: 'var a = 0; {} export {a}', options: [ { before: false, after: false, overrides: { export: { before: true, after: true } } } ] },
    { code: 'var a = 0; {}export{a}', options: [ { before: true, after: true, overrides: { export: { before: false, after: false } } } ] },
    { code: 'var a = 0;\n;export {a}' },
    { code: 'var a = 0;\n; export{a}', options: [ { before: false, after: false } ] },
    { code: 'class Bar extends [] {}' },
    { code: 'class Bar extends[] {}', options: [ { before: false, after: false } ] },
    { code: 'class Bar extends [] {}', options: [ { before: false, after: false, overrides: { extends: { before: true, after: true } } } ] },
    { code: 'class Bar extends[] {}', options: [ { before: true, after: true, overrides: { extends: { before: false, after: false } } } ] },
    'try {} finally {}',
    { code: 'try{}finally{}', options: [ { before: false, after: false } ] },
    { code: 'try{} finally {}', options: [ { before: false, after: false, overrides: { finally: { before: true, after: true } } } ] },
    { code: 'try {}finally{}', options: [ { before: true, after: true, overrides: { finally: { before: false, after: false } } } ] },
    'try {}\nfinally\n{}',
    { code: 'try{}\nfinally\n{}', options: [ { before: false, after: false } ] },
    '{} for (;;) {}',
    '{} for (var foo in obj) {}',
    { code: '{} for (var foo of list) {}' },
    { code: '{}for(;;) {}', options: [ { before: false, after: false } ] },
    { code: '{}for(var foo in obj) {}', options: [ { before: false, after: false } ] },
    { code: '{}for(var foo of list) {}', options: [ { before: false, after: false } ] },
    { code: '{} for (;;) {}', options: [ { before: false, after: false, overrides: { for: { before: true, after: true } } } ] },
    { code: '{} for (var foo in obj) {}', options: [ { before: false, after: false, overrides: { for: { before: true, after: true } } } ] },
    { code: '{} for (var foo of list) {}', options: [ { before: false, after: false, overrides: { for: { before: true, after: true } } } ] },
    { code: '{}for(;;) {}', options: [ { before: true, after: true, overrides: { for: { before: false, after: false } } } ] },
    { code: '{}for(var foo in obj) {}', options: [ { before: true, after: true, overrides: { for: { before: false, after: false } } } ] },
    { code: '{}for(var foo of list) {}', options: [ { before: true, after: true, overrides: { for: { before: false, after: false } } } ] },
    '{for (;;) {} }',
    '{for (var foo in obj) {} }',
    { code: '{for (var foo of list) {} }' },
    { code: '{ for(;;) {} }', options: [ { before: false, after: false } ] },
    { code: '{ for(var foo in obj) {} }', options: [ { before: false, after: false } ] },
    { code: '{ for(var foo of list) {} }', options: [ { before: false, after: false } ] },
    ';for (;;) {}',
    ';for (var foo in obj) {}',
    { code: ';for (var foo of list) {}' },
    { code: '; for(;;) {}', options: [ { before: false, after: false } ] },
    { code: '; for(var foo in obj) {}', options: [ { before: false, after: false } ] },
    { code: '; for(var foo of list) {}', options: [ { before: false, after: false } ] },
    { code: 'import {foo} from "foo"' },
    { code: 'export {foo} from "foo"' },
    { code: 'export * from "foo"' },
    { code: 'export * as "x" from "foo"' },
    { code: 'import{foo}from"foo"', options: [ { before: false, after: false } ] },
    { code: 'export{foo}from"foo"', options: [ { before: false, after: false } ] },
    { code: 'export*from"foo"', options: [ { before: false, after: false } ] },
    { code: 'export*as x from"foo"', options: [ { before: false, after: false } ] },
    { code: 'export*as"x"from"foo"', options: [ { before: false, after: false } ] },
    { code: 'import{foo} from "foo"', options: [ { before: false, after: false, overrides: { from: { before: true, after: true } } } ] },
    { code: 'export{foo} from "foo"', options: [ { before: false, after: false, overrides: { from: { before: true, after: true } } } ] },
    { code: 'export* from "foo"', options: [ { before: false, after: false, overrides: { from: { before: true, after: true } } } ] },
    { code: 'export*as"x" from "foo"', options: [ { before: false, after: false, overrides: { from: { before: true, after: true } } } ] },
    { code: 'import {foo}from"foo"', options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ] },
    { code: 'export {foo}from"foo"', options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ] },
    { code: 'export *from"foo"', options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ] },
    { code: 'export * as x from"foo"', options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ] },
    { code: 'export * as "x"from"foo"', options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ] },
    '{} function foo() {}',
    { code: '{}function foo() {}', options: [ { before: false, after: false } ] },
    { code: '{} function foo() {}', options: [ { before: false, after: false, overrides: { function: { before: true, after: true } } } ] },
    { code: '{}function foo() {}', options: [ { before: true, after: true, overrides: { function: { before: false, after: false } } } ] },
    '[function() {}]',
    { code: '[ function() {}]', options: [ { before: false, after: false } ] },
    { code: '(() =>function() {})' },
    { code: '(() => function() {})', options: [ { before: false, after: false } ] },
    '{function foo() {} }',
    { code: '{ function foo() {} }', options: [ { before: false, after: false } ] },
    '(0,function() {})',
    { code: '(0, function() {})', options: [ { before: false, after: false } ] },
    'a[function() {}]',
    { code: '({[function() {}]: 0})' },
    { code: 'a[ function() {}]', options: [ { before: false, after: false } ] },
    { code: '({[ function(){}]: 0})', options: [ { before: false, after: false } ] },
    { code: 'function* foo() {}' },
    { code: 'function *foo() {}', options: [ { before: false, after: false } ] },
    '({a:function() {} })',
    { code: '({a: function() {} })', options: [ { before: false, after: false } ] },
    ';function foo() {};',
    { code: '; function foo() {} ;', options: [ { before: false, after: false } ] },
    '(function() {})',
    { code: '( function () {})', options: [ { before: false, after: false } ] },
    'a =function() {}',
    { code: 'a = function() {}', options: [ { before: false, after: false } ] },
    'a+function() {}',
    { code: 'a + function() {}', options: [ { before: false, after: false } ] },
    'a<function() {}',
    { code: 'a < function() {}', options: [ { before: false, after: false } ] },
    'a>function() {}',
    { code: 'a > function() {}', options: [ { before: false, after: false } ] },
    '!function() {}',
    { code: '! function() {}', options: [ { before: false, after: false } ] },
    { code: '`${function() {}}`' },
    { code: '`${ function() {}}`', options: [ { before: false, after: false } ] },
    { code: '<Foo onClick={function() {}} />' },
    { code: '<Foo onClick={ function() {}} />', options: [ { before: false, after: false } ] },
    { code: '({ get [b]() {} })' },
    { code: 'class A { a() {} get [b]() {} }' },
    { code: 'class A { a() {} static get [b]() {} }' },
    { code: '({ get[b]() {} })', options: [ { before: false, after: false } ] },
    { code: 'class A { a() {}get[b]() {} }', options: [ { before: false, after: false } ] },
    { code: 'class A { a() {}static get[b]() {} }', options: [ { before: false, after: false } ] },
    { code: '({ get [b]() {} })', options: [ { before: false, after: false, overrides: { get: { before: true, after: true } } } ] },
    { code: 'class A { a() {} get [b]() {} }', options: [ { before: false, after: false, overrides: { get: { before: true, after: true } } } ] },
    { code: '({ get[b]() {} })', options: [ { before: true, after: true, overrides: { get: { before: false, after: false } } } ] },
    { code: 'class A { a() {}get[b]() {} }', options: [ { before: true, after: true, overrides: { get: { before: false, after: false } } } ] },
    { code: 'class A { a; get #b() {} }' },
    { code: 'class A { a;get#b() {} }', options: [ { before: false, after: false } ] },
    { code: '({ a,get [b]() {} })' },
    { code: '({ a, get[b]() {} })', options: [ { before: false, after: false } ] },
    { code: 'class A { ;get #b() {} }' },
    { code: 'class A { ; get#b() {} }', options: [ { before: false, after: false } ] },
    '{} if (a) {}',
    'if (a) {} else if (a) {}',
    { code: '{}if(a) {}', options: [ { before: false, after: false } ] },
    { code: 'if(a) {}else if(a) {}', options: [ { before: false, after: false } ] },
    { code: '{} if (a) {}', options: [ { before: false, after: false, overrides: { if: { before: true, after: true } } } ] },
    { code: 'if (a) {}else if (a) {}', options: [ { before: false, after: false, overrides: { if: { before: true, after: true } } } ] },
    { code: '{}if(a) {}', options: [ { before: true, after: true, overrides: { if: { before: false, after: false } } } ] },
    { code: 'if(a) {} else if(a) {}', options: [ { before: true, after: true, overrides: { if: { before: false, after: false } } } ] },
    '{if (a) {} }',
    { code: '{ if(a) {} }', options: [ { before: false, after: false } ] },
    ';if (a) {}',
    { code: '; if(a) {}', options: [ { before: false, after: false } ] },
    { code: '{} import {a} from "foo"' },
    { code: '{} import a from "foo"' },
    { code: '{} import * as a from "a"' },
    { code: '{}import{a}from"foo"', options: [ { before: false, after: false } ] },
    { code: '{}import*as a from"foo"', options: [ { before: false, after: false } ] },
    { code: '{} import {a}from"foo"', options: [ { before: false, after: false, overrides: { import: { before: true, after: true } } } ] },
    { code: '{} import *as a from"foo"', options: [ { before: false, after: false, overrides: { import: { before: true, after: true } } } ] },
    { code: '{}import{a} from "foo"', options: [ { before: true, after: true, overrides: { import: { before: false, after: false } } } ] },
    { code: '{}import* as a from "foo"', options: [ { before: true, after: true, overrides: { import: { before: false, after: false } } } ] },
    { code: ';import {a} from "foo"' },
    { code: '; import{a}from"foo"', options: [ { before: false, after: false } ] },
    { code: 'for ([foo] in {foo: 0}) {}' },
    { code: 'for([foo]in{foo: 0}) {}', options: [ { before: false, after: false } ] },
    { code: 'for([foo] in {foo: 0}) {}', options: [ { before: false, after: false, overrides: { in: { before: true, after: true } } } ] },
    { code: 'for ([foo]in{foo: 0}) {}', options: [ { before: true, after: true, overrides: { in: { before: false, after: false } } } ] },
    { code: 'for ([foo] in ({foo: 0})) {}' },
    'if ("foo"in{foo: 0}) {}',
    { code: 'if("foo" in {foo: 0}) {}', options: [ { before: false, after: false } ] },
    'if ("foo"instanceof{foo: 0}) {}',
    { code: 'if("foo" instanceof {foo: 0}) {}', options: [ { before: false, after: false } ] },
    { code: '{} let [a] = b' },
    { code: '{}let[a] = b', options: [ { before: false, after: false } ] },
    { code: '{} let [a] = b', options: [ { before: false, after: false, overrides: { let: { before: true, after: true } } } ] },
    { code: '{}let[a] = b', options: [ { before: true, after: true, overrides: { let: { before: false, after: false } } } ] },
    { code: '{let [a] = b }' },
    { code: '{ let[a] = b }', options: [ { before: false, after: false } ] },
    { code: ';let [a] = b' },
    { code: '; let[a] = b', options: [ { before: false, after: false } ] },
    '{} new foo()',
    { code: '{}new foo()', options: [ { before: false, after: false } ] },
    { code: '{} new foo()', options: [ { before: false, after: false, overrides: { new: { before: true, after: true } } } ] },
    { code: '{}new foo()', options: [ { before: true, after: true, overrides: { new: { before: false, after: false } } } ] },
    '[new foo()]',
    { code: '[ new foo()]', options: [ { before: false, after: false } ] },
    { code: '(() =>new foo())' },
    { code: '(() => new foo())', options: [ { before: false, after: false } ] },
    '{new foo() }',
    { code: '{ new foo() }', options: [ { before: false, after: false } ] },
    '(0,new foo())',
    { code: '(0, new foo())', options: [ { before: false, after: false } ] },
    'a[new foo()]',
    { code: '({[new foo()]: 0})' },
    { code: 'a[ new foo()]', options: [ { before: false, after: false } ] },
    { code: '({[ new foo()]: 0})', options: [ { before: false, after: false } ] },
    '({a:new foo() })',
    { code: '({a: new foo() })', options: [ { before: false, after: false } ] },
    ';new foo()',
    { code: '; new foo()', options: [ { before: false, after: false } ] },
    '(new foo())',
    { code: '( new foo())', options: [ { before: false, after: false } ] },
    'a =new foo()',
    { code: 'a = new foo()', options: [ { before: false, after: false } ] },
    'a+new foo()',
    { code: 'a + new foo()', options: [ { before: false, after: false } ] },
    'a<new foo()',
    { code: 'a < new foo()', options: [ { before: false, after: false } ] },
    'a>new foo()',
    { code: 'a > new foo()', options: [ { before: false, after: false } ] },
    '!new(foo)()',
    { code: '! new (foo)()', options: [ { before: false, after: false } ] },
    { code: '`${new foo()}`' },
    { code: '`${ new foo()}`', options: [ { before: false, after: false } ] },
    { code: '<Foo onClick={new foo()} />' },
    { code: '<Foo onClick={ new foo()} />', options: [ { before: false, after: false } ] },
    { code: 'for ([foo] of {foo: 0}) {}' },
    { code: 'for([foo]of{foo: 0}) {}', options: [ { before: false, after: false } ] },
    { code: 'for([foo] of {foo: 0}) {}', options: [ { before: false, after: false, overrides: { of: { before: true, after: true } } } ] },
    { code: 'for ([foo]of{foo: 0}) {}', options: [ { before: true, after: true, overrides: { of: { before: false, after: false } } } ] },
    { code: 'for ([foo] of ({foo: 0})) {}' },
    'function foo() { {} return +a }',
    { code: 'function foo() { return <p/>; }' },
    { code: 'function foo() { {}return+a }', options: [ { before: false, after: false } ] },
    { code: 'function foo() { return<p/>; }', options: [ { after: false } ] },
    { code: 'function foo() { {} return +a }', options: [ { before: false, after: false, overrides: { return: { before: true, after: true } } } ] },
    { code: 'function foo() { {}return+a }', options: [ { before: true, after: true, overrides: { return: { before: false, after: false } } } ] },
    'function foo() {\nreturn\n}',
    { code: 'function foo() {\nreturn\n}', options: [ { before: false, after: false } ] },
    'function foo() {return}',
    { code: 'function foo() { return }', options: [ { before: false, after: false } ] },
    'function foo() { ;return; }',
    { code: 'function foo() { ; return ; }', options: [ { before: false, after: false } ] },
    { code: '({ set [b](value) {} })' },
    { code: 'class A { a() {} set [b](value) {} }' },
    { code: 'class A { a() {} static set [b](value) {} }' },
    { code: '({ set[b](value) {} })', options: [ { before: false, after: false } ] },
    { code: 'class A { a() {}set[b](value) {} }', options: [ { before: false, after: false } ] },
    { code: '({ set [b](value) {} })', options: [ { before: false, after: false, overrides: { set: { before: true, after: true } } } ] },
    { code: 'class A { a() {} set [b](value) {} }', options: [ { before: false, after: false, overrides: { set: { before: true, after: true } } } ] },
    { code: '({ set[b](value) {} })', options: [ { before: true, after: true, overrides: { set: { before: false, after: false } } } ] },
    { code: 'class A { a() {}set[b](value) {} }', options: [ { before: true, after: true, overrides: { set: { before: false, after: false } } } ] },
    { code: 'class A { a; set #b(value) {} }' },
    { code: 'class A { a;set#b(value) {} }', options: [ { before: false, after: false } ] },
    { code: '({ a,set [b](value) {} })' },
    { code: '({ a, set[b](value) {} })', options: [ { before: false, after: false } ] },
    { code: 'class A { ;set #b(value) {} }' },
    { code: 'class A { ; set#b(value) {} }', options: [ { before: false, after: false } ] },
    { code: 'class A { a() {} static [b]() {} }' },
    { code: 'class A { a() {}static[b]() {} }', options: [ { before: false, after: false } ] },
    { code: 'class A { a() {} static [b]() {} }', options: [ { before: false, after: false, overrides: { static: { before: true, after: true } } } ] },
    { code: 'class A { a() {}static[b]() {} }', options: [ { before: true, after: true, overrides: { static: { before: false, after: false } } } ] },
    { code: 'class A { a; static [b]; }' },
    { code: 'class A { a;static[b]; }', options: [ { before: false, after: false } ] },
    { code: 'class A { a; static #b; }' },
    { code: 'class A { a;static#b; }', options: [ { before: false, after: false } ] },
    { code: 'class A { a() {} static {} }' },
    { code: 'class A { a() {}static{} }', options: [ { before: false, after: false } ] },
    { code: 'class A { a() {} static {} }', options: [ { before: false, after: false, overrides: { static: { before: true, after: true } } } ] },
    { code: 'class A { a() {}static{} }', options: [ { before: true, after: true, overrides: { static: { before: false, after: false } } } ] },
    { code: 'class A { a() {}\nstatic\n{} }', options: [ { before: false, after: false } ] },
    { code: 'class A { static* [a]() {} }' },
    { code: 'class A { static *[a]() {} }', options: [ { before: false, after: false } ] },
    { code: 'class A { ;static a() {} }' },
    { code: 'class A { ; static a() {} }', options: [ { before: false, after: false } ] },
    { code: 'class A { ;static a; }' },
    { code: 'class A { ; static a ; }', options: [ { before: false, after: false } ] },
    { code: 'class A { ;static {} }' },
    { code: 'class A { ; static{} }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { a() { {} super[b](); } }' },
    { code: 'class A extends B { a() { {}super[b](); } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { a() { {} super[b](); } }', options: [ { before: false, after: false, overrides: { super: { before: true, after: true } } } ] },
    { code: 'class A extends B { a() { {}super[b](); } }', options: [ { before: true, after: true, overrides: { super: { before: false, after: false } } } ] },
    { code: 'class A extends B { constructor() { [super()]; } }' },
    { code: 'class A extends B { constructor() { [ super() ]; } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { () =>super(); } }' },
    { code: 'class A extends B { constructor() { () => super(); } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() {super()} }' },
    { code: 'class A extends B { constructor() { super() } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { (0,super()) } }' },
    { code: 'class A extends B { constructor() { (0, super()) } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { ({[super()]: 0}) } }' },
    { code: 'class A extends B { constructor() { ({[ super() ]: 0}) } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { ({a:super() }) } }' },
    { code: 'class A extends B { constructor() { ({a: super() }) } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { super(); } }' },
    { code: 'class A extends B { constructor() { super (); } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { ;super(); } }' },
    { code: 'class A extends B { constructor() { ; super() ; } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { (super()) } }' },
    { code: 'class A extends B { constructor() { ( super() ) } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { b =super() } }' },
    { code: 'class A extends B { constructor() { b = super() } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { b+super() } }' },
    { code: 'class A extends B { constructor() { b + super() } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { b<super() } }' },
    { code: 'class A extends B { constructor() { b < super() } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { b>super() } }' },
    { code: 'class A extends B { constructor() { b > super() } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { !super() } }' },
    { code: 'class A extends B { constructor() { ! super() } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { `${super()}` } }' },
    { code: 'class A extends B { constructor() { `${ super() }` } }', options: [ { before: false, after: false } ] },
    { code: 'class A extends B { constructor() { <Foo onClick={super()} /> } }' },
    { code: 'class A extends B { constructor() { <Foo onClick={ super() } /> } }', options: [ { before: false, after: false } ] },
    '{} switch (a) {}',
    { code: '{}switch(a) {}', options: [ { before: false, after: false } ] },
    { code: '{} switch (a) {}', options: [ { before: false, after: false, overrides: { switch: { before: true, after: true } } } ] },
    { code: '{}switch(a) {}', options: [ { before: true, after: true, overrides: { switch: { before: false, after: false } } } ] },
    '{switch (a) {} }',
    { code: '{ switch(a) {} }', options: [ { before: false, after: false } ] },
    ';switch (a) {}',
    { code: '; switch(a) {}', options: [ { before: false, after: false } ] },
    '{} this[a]',
    { code: '{}this[a]', options: [ { before: false, after: false } ] },
    { code: '{} this[a]', options: [ { before: false, after: false, overrides: { this: { before: true, after: true } } } ] },
    { code: '{}this[a]', options: [ { before: true, after: true, overrides: { this: { before: false, after: false } } } ] },
    { code: '<Thing> this.blah' },
    { code: '<Thing>this.blah', options: [ { before: true, after: false, overrides: { this: { before: false } } } ] },
    '[this]',
    { code: '[ this ]', options: [ { before: false, after: false } ] },
    { code: '(() =>this)' },
    { code: '(() => this)', options: [ { before: false, after: false } ] },
    '{this}',
    { code: '{ this }', options: [ { before: false, after: false } ] },
    '(0,this)',
    { code: '(0, this)', options: [ { before: false, after: false } ] },
    'a[this]',
    { code: '({[this]: 0})' },
    { code: 'a[ this ]', options: [ { before: false, after: false } ] },
    { code: '({[ this ]: 0})', options: [ { before: false, after: false } ] },
    '({a:this })',
    { code: '({a: this })', options: [ { before: false, after: false } ] },
    ';this',
    { code: '; this', options: [ { before: false, after: false } ] },
    '(this)',
    { code: '( this )', options: [ { before: false, after: false } ] },
    'a =this',
    { code: 'a = this', options: [ { before: false, after: false } ] },
    'a+this',
    { code: 'a + this', options: [ { before: false, after: false } ] },
    'a<this',
    { code: 'a < this', options: [ { before: false, after: false } ] },
    'a>this',
    { code: 'a > this', options: [ { before: false, after: false } ] },
    'this+a',
    { code: 'this + a', options: [ { before: false, after: false } ] },
    'this<a',
    { code: 'this < a', options: [ { before: false, after: false } ] },
    'this>a',
    { code: 'this > a', options: [ { before: false, after: false } ] },
    '!this',
    { code: '! this', options: [ { before: false, after: false } ] },
    { code: '`${this}`' },
    { code: '`${ this }`', options: [ { before: false, after: false } ] },
    { code: '<Foo onClick={this} />' },
    { code: '<Foo onClick={ this } />', options: [ { before: false, after: false } ] },
    'function foo() { {} throw +a }',
    { code: 'function foo() { {}throw+a }', options: [ { before: false, after: false } ] },
    { code: 'function foo() { {} throw +a }', options: [ { before: false, after: false, overrides: { throw: { before: true, after: true } } } ] },
    { code: 'function foo() { {}throw+a }', options: [ { before: true, after: true, overrides: { throw: { before: false, after: false } } } ] },
    'function foo() {\nthrow a\n}',
    { code: 'function foo() {\nthrow a\n}', options: [ { before: false, after: false } ] },
    'function foo() {throw a }',
    { code: 'function foo() { throw a }', options: [ { before: false, after: false } ] },
    'function foo() { ;throw a }',
    { code: 'function foo() { ; throw a }', options: [ { before: false, after: false } ] },
    '{} try {} finally {}',
    { code: '{}try{}finally{}', options: [ { before: false, after: false } ] },
    { code: '{} try {}finally{}', options: [ { before: false, after: false, overrides: { try: { before: true, after: true } } } ] },
    { code: '{}try{} finally {}', options: [ { before: true, after: true, overrides: { try: { before: false, after: false } } } ] },
    '{try {} finally {}}',
    { code: '{ try{}finally{}}', options: [ { before: false, after: false } ] },
    ';try {} finally {}',
    { code: '; try{}finally{}', options: [ { before: false, after: false } ] },
    '{} typeof foo',
    { code: '{}typeof foo', options: [ { before: false, after: false } ] },
    { code: '{} typeof foo', options: [ { before: false, after: false, overrides: { typeof: { before: true, after: true } } } ] },
    { code: '{}typeof foo', options: [ { before: true, after: true, overrides: { typeof: { before: false, after: false } } } ] },
    '[typeof foo]',
    { code: '[ typeof foo]', options: [ { before: false, after: false } ] },
    { code: '(() =>typeof foo)' },
    { code: '(() => typeof foo)', options: [ { before: false, after: false } ] },
    '{typeof foo }',
    { code: '{ typeof foo }', options: [ { before: false, after: false } ] },
    '(0,typeof foo)',
    { code: '(0, typeof foo)', options: [ { before: false, after: false } ] },
    'a[typeof foo]',
    { code: '({[typeof foo]: 0})' },
    { code: 'a[ typeof foo]', options: [ { before: false, after: false } ] },
    { code: '({[ typeof foo]: 0})', options: [ { before: false, after: false } ] },
    '({a:typeof foo })',
    { code: '({a: typeof foo })', options: [ { before: false, after: false } ] },
    ';typeof foo',
    { code: '; typeof foo', options: [ { before: false, after: false } ] },
    '(typeof foo)',
    { code: '( typeof foo)', options: [ { before: false, after: false } ] },
    'a =typeof foo',
    { code: 'a = typeof foo', options: [ { before: false, after: false } ] },
    'a+typeof foo',
    { code: 'a + typeof foo', options: [ { before: false, after: false } ] },
    'a<typeof foo',
    { code: 'a < typeof foo', options: [ { before: false, after: false } ] },
    'a>typeof foo',
    { code: 'a > typeof foo', options: [ { before: false, after: false } ] },
    '!typeof+foo',
    { code: '! typeof +foo', options: [ { before: false, after: false } ] },
    { code: '`${typeof foo}`' },
    { code: '`${ typeof foo}`', options: [ { before: false, after: false } ] },
    { code: '<Foo onClick={typeof foo} />' },
    { code: '<Foo onClick={ typeof foo} />', options: [ { before: false, after: false } ] },
    { code: '{} var [a] = b' },
    { code: '{}var[a] = b', options: [ { before: false, after: false } ] },
    { code: '{} var [a] = b', options: [ { before: false, after: false, overrides: { var: { before: true, after: true } } } ] },
    { code: '{}var[a] = b', options: [ { before: true, after: true, overrides: { var: { before: false, after: false } } } ] },
    'for (var foo in [1, 2, 3]) {}',
    '{var a = b }',
    { code: '{ var a = b }', options: [ { before: false, after: false } ] },
    ';var a = b',
    { code: '; var a = b', options: [ { before: false, after: false } ] },
    '{} void foo',
    { code: '{}void foo', options: [ { before: false, after: false } ] },
    { code: '{} void foo', options: [ { before: false, after: false, overrides: { void: { before: true, after: true } } } ] },
    { code: '{}void foo', options: [ { before: true, after: true, overrides: { void: { before: false, after: false } } } ] },
    '[void foo]',
    { code: '[ void foo]', options: [ { before: false, after: false } ] },
    { code: '(() =>void foo)' },
    { code: '(() => void foo)', options: [ { before: false, after: false } ] },
    '{void foo }',
    { code: '{ void foo }', options: [ { before: false, after: false } ] },
    '(0,void foo)',
    { code: '(0, void foo)', options: [ { before: false, after: false } ] },
    'a[void foo]',
    { code: '({[void foo]: 0})' },
    { code: 'a[ void foo]', options: [ { before: false, after: false } ] },
    { code: '({[ void foo]: 0})', options: [ { before: false, after: false } ] },
    '({a:void foo })',
    { code: '({a: void foo })', options: [ { before: false, after: false } ] },
    ';void foo',
    { code: '; void foo', options: [ { before: false, after: false } ] },
    '(void foo)',
    { code: '( void foo)', options: [ { before: false, after: false } ] },
    'a =void foo',
    { code: 'a = void foo', options: [ { before: false, after: false } ] },
    'a+void foo',
    { code: 'a + void foo', options: [ { before: false, after: false } ] },
    'a<void foo',
    { code: 'a < void foo', options: [ { before: false, after: false } ] },
    'a>void foo',
    { code: 'a > void foo', options: [ { before: false, after: false } ] },
    '!void+foo',
    { code: '! void +foo', options: [ { before: false, after: false } ] },
    { code: '`${void foo}`' },
    { code: '`${ void foo}`', options: [ { before: false, after: false } ] },
    { code: '<Foo onClick={void foo} />' },
    { code: '<Foo onClick={ void foo} />', options: [ { before: false, after: false } ] },
    '{} while (a) {}',
    'do {} while (a)',
    { code: '{}while(a) {}', options: [ { before: false, after: false } ] },
    { code: 'do{}while(a)', options: [ { before: false, after: false } ] },
    { code: '{} while (a) {}', options: [ { before: false, after: false, overrides: { while: { before: true, after: true } } } ] },
    { code: 'do{} while (a)', options: [ { before: false, after: false, overrides: { while: { before: true, after: true } } } ] },
    { code: '{}while(a) {}', options: [ { before: true, after: true, overrides: { while: { before: false, after: false } } } ] },
    { code: 'do {}while(a)', options: [ { before: true, after: true, overrides: { while: { before: false, after: false } } } ] },
    'do {}\nwhile (a)',
    { code: 'do{}\nwhile(a)', options: [ { before: false, after: false } ] },
    '{while (a) {}}',
    { code: '{ while(a) {}}', options: [ { before: false, after: false } ] },
    ';while (a);',
    'do;while (a);',
    { code: '; while(a) ;', options: [ { before: false, after: false } ] },
    { code: 'do ; while(a) ;', options: [ { before: false, after: false } ] },
    { code: '{} with (obj) {}' },
    { code: '{}with(obj) {}', options: [ { before: false, after: false } ] },
    { code: '{} with (obj) {}', options: [ { before: false, after: false, overrides: { with: { before: true, after: true } } } ] },
    { code: '{}with(obj) {}', options: [ { before: true, after: true, overrides: { with: { before: false, after: false } } } ] },
    { code: '{with (obj) {}}' },
    { code: '{ with(obj) {}}', options: [ { before: false, after: false } ] },
    { code: ';with (obj) {}' },
    { code: '; with(obj) {}', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { {} yield foo }' },
    { code: 'function* foo() { {}yield foo }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { {} yield foo }', options: [ { before: false, after: false, overrides: { yield: { before: true, after: true } } } ] },
    { code: 'function* foo() { {}yield foo }', options: [ { before: true, after: true, overrides: { yield: { before: false, after: false } } } ] },
    { code: 'function* foo() { [yield] }' },
    { code: 'function* foo() { [ yield ] }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() {yield}' },
    { code: 'function* foo() { yield }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { (0,yield foo) }' },
    { code: 'function* foo() { (0, yield foo) }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { a[yield] }' },
    { code: 'function* foo() { ({[yield]: 0}) }' },
    { code: 'function* foo() { a[ yield ] }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { ({[ yield ]: 0}) }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { ({a:yield foo }) }' },
    { code: 'function* foo() { ({a: yield foo }) }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { ;yield; }' },
    { code: 'function* foo() { ; yield ; }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { (yield) }' },
    { code: 'function* foo() { ( yield ) }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { a =yield foo }' },
    { code: 'function* foo() { a = yield foo }', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { yield+foo }' },
    { code: 'function* foo() { yield +foo }', options: [ { before: false, after: false } ] },
    { code: '`${yield}`' },
    { code: '`${ yield}`', options: [ { before: false, after: false } ] },
    { code: 'function* foo() { <Foo onClick={yield} /> }' },
    { code: 'function* foo() { <Foo onClick={ yield } /> }', options: [ { before: false, after: false } ] },
    { code: '@dec class Foo {}' },
    { code: 'class Foo { @dec get bar() {} @dec set baz() {} @dec async baw() {} }' },
    { code: 'class Foo { @dec static qux() {} @dec static get bar() {} @dec static set baz() {} @dec static async baw() {} }' },
    { code: 'symbol => 4;' },

    // ---- from keyword-spacing._ts_.test.ts ----
    { code: 'const foo = {} as {};' },
    { code: 'const foo = {}as{};', options: [ { before: false, after: false } ] },
    { code: 'const foo = {} as {};', options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ] },
    { code: 'const foo = {}as{};', options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ] },
    { code: 'const foo = {} as {};', options: [ { overrides: { as: {} } } ] },
    'const foo = {} satisfies {}',
    { code: 'const foo = {}satisfies{};', options: [ { before: false, after: false } ] },
    { code: 'const foo = {} satisfies {};', options: [ { before: false, after: false, overrides: { satisfies: { before: true, after: true } } } ] },
    { code: 'const foo = {}satisfies{};', options: [ { before: true, after: true, overrides: { satisfies: { before: false, after: false } } } ] },
    { code: 'const foo = {} satisfies {};', options: [ { before: false, after: false, overrides: { satisfies: { before: true, after: true } } } ] },
    { code: 'const a = 1 as any', options: [ { before: false, after: false } ] },
    { code: 'const a = true as any', options: [ { before: false, after: false } ] },
    { code: 'const a = b as any', options: [ { before: false, after: false } ] },
    { code: 'const a = 1 satisfies any', options: [ { before: false, after: false } ] },
    { code: 'const a = true satisfies any', options: [ { before: false, after: false } ] },
    { code: 'const a = b satisfies any', options: [ { before: false, after: false } ] },
    { code: 'import type { foo } from "foo";', options: [ { before: true, after: true } ] },
    { code: "import type * as Foo from 'foo'", options: [ { before: true, after: true } ] },
    { code: 'import type { SavedQueries } from "./SavedQueries.js";', options: [ { before: true, after: false, overrides: { else: { after: true }, return: { after: true }, try: { after: true }, catch: { after: false }, case: { after: true }, const: { after: true }, throw: { after: true }, let: { after: true }, do: { after: true }, of: { after: true }, as: { after: true }, finally: { after: true }, from: { after: true }, import: { after: true }, export: { after: true }, default: { after: true }, type: { after: true } } } ] },
    { code: 'import type { SavedQueries } from "./SavedQueries.js";', options: [ { before: true, after: true, overrides: { import: { after: false } } } ] },
    { code: "import type{SavedQueries} from './SavedQueries.js';", options: [ { before: true, after: false, overrides: { from: { after: true } } } ] },
    { code: "import type{SavedQueries} from'./SavedQueries.js';", options: [ { before: true, after: false } ] },
    { code: "import type http from 'node:http';", options: [ { before: true, after: false, overrides: { from: { after: true } } } ] },
    { code: "import type http from'node:http';", options: [ { before: true, after: false } ] },
    { code: 'import type {} from "foo";', options: [ { before: true, after: true } ] },
    { code: 'import type { foo1, foo2 } from "foo";', options: [ { before: true, after: true } ] },
    { code: 'import type { foo1 as _foo1, foo2 as _foo2 } from "foo";', options: [ { before: true, after: true } ] },
    { code: 'import type { foo as bar } from "foo";', options: [ { before: true, after: true } ] },
    { code: "import pkgJson from 'package.json' with { type: 'json' }", options: [ { before: true, after: true } ] },
    { code: "export{ name }from'package.json'with{ type: 'json' }", options: [ { before: false, after: false } ] },
    { code: "export * from 'package.json'with{ type: 'json' }", options: [ { before: true, after: true, overrides: { with: { before: false, after: false } } } ] },
    { code: 'class A { delete() {} }', options: [ { before: true, after: true } ] },
    { code: 'class C { @readonly accessor foo = 1 }', options: [ { before: false, after: false } ] },
    { code: 'export type { foo } from "foo";', options: [ { before: true, after: true } ] },
    { code: "export type * as Foo from 'foo'", options: [ { before: true, after: true } ] },
    { code: 'export type { SavedQueries } from "./SavedQueries.js";', options: [ { before: true, after: false, overrides: { else: { after: true }, return: { after: true }, try: { after: true }, catch: { after: false }, case: { after: true }, const: { after: true }, throw: { after: true }, let: { after: true }, do: { after: true }, of: { after: true }, as: { after: true }, finally: { after: true }, from: { after: true }, import: { after: true }, export: { after: true }, default: { after: true }, type: { after: true } } } ] },
    { code: 'export type { SavedQueries } from "./SavedQueries.js";', options: [ { before: true, after: true, overrides: { export: { after: false } } } ] },
    { code: "export type{SavedQueries} from './SavedQueries.js';", options: [ { before: true, after: false, overrides: { from: { after: true } } } ] },
    { code: "export type{SavedQueries} from'./SavedQueries.js';", options: [ { before: true, after: false } ] },
    { code: 'export type {} from "foo";', options: [ { before: true, after: true } ] },
    { code: 'export type { foo1, foo2 } from "foo";', options: [ { before: true, after: true } ] },
    { code: 'export type { foo1 as _foo1, foo2 as _foo2 } from "foo";', options: [ { before: true, after: true } ] },
    { code: 'export type { foo as bar } from "foo";', options: [ { before: true, after: true } ] },
    { code: 'import type{}from"foo";', options: [ { before: false, after: false } ] },
    { code: 'export type{}from"foo";', options: [ { before: false, after: false } ] },
  ],

  invalid: [
    // ---- from keyword-spacing._js_.test.ts ----
    {
      code: 'import { "a"as b } from "foo"',
      output: 'import { "a" as b } from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'import{ "a" as b }from"foo"',
      output: 'import{ "a"as b }from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'import{ "a"as b }from"foo"',
      output: 'import{ "a" as b }from"foo"',
      options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'import { "a" as b } from "foo"',
      output: 'import { "a"as b } from "foo"',
      options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'let a; export { a as"b" };',
      output: 'let a; export { a as "b" };',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'export { "a"as b } from "foo";',
      output: 'export { "a" as b } from "foo";',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export { "a"as"b" } from "foo";',
      output: 'export { "a" as "b" } from "foo";',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
        { messageId: 'expectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'let a; export{ a as "b" };',
      output: 'let a; export{ a as"b" };',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'export{ "a" as b }from"foo";',
      output: 'export{ "a"as b }from"foo";',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export{ "a" as "b" }from"foo";',
      output: 'export{ "a"as"b" }from"foo";',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
        { messageId: 'unexpectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'let a; export{ a as"b" };',
      output: 'let a; export{ a as "b" };',
      options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'export{ "a"as b }from"foo";',
      output: 'export{ "a" as b }from"foo";',
      options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export{ "a"as"b" }from"foo";',
      output: 'export{ "a" as "b" }from"foo";',
      options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
        { messageId: 'expectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'let a; export { a as "b" };',
      output: 'let a; export { a as"b" };',
      options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'export { "a" as b } from "foo";',
      output: 'export { "a"as b } from "foo";',
      options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export { "a" as "b" } from "foo";',
      output: 'export { "a"as"b" } from "foo";',
      options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
        { messageId: 'unexpectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'import *as a from "foo"',
      output: 'import * as a from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' }, line: 1, column: 9, endLine: 1, endColumn: 11 },
      ],
    },
    {
      code: 'import* as a from"foo"',
      output: 'import*as a from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' }, line: 1, column: 8, endLine: 1, endColumn: 9 },
      ],
    },
    {
      code: 'import*   as a from"foo"',
      output: 'import*as a from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' }, line: 1, column: 8, endLine: 1, endColumn: 11 },
      ],
    },
    {
      code: 'import*as a from"foo"',
      output: 'import* as a from"foo"',
      options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'import * as a from "foo"',
      output: 'import *as a from "foo"',
      options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export *as a from "foo"',
      output: 'export * as a from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export *as"a" from "foo"',
      output: 'export * as "a" from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
        { messageId: 'expectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'export* as a from"foo"',
      output: 'export*as a from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export* as "a"from"foo"',
      output: 'export*as"a"from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
        { messageId: 'unexpectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'export*as a from"foo"',
      output: 'export* as a from"foo"',
      options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export*as"a"from"foo"',
      output: 'export* as "a"from"foo"',
      options: [ { before: false, after: false, overrides: { as: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
        { messageId: 'expectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'export * as a from "foo"',
      output: 'export *as a from "foo"',
      options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'export * as "a" from "foo"',
      output: 'export *as"a" from "foo"',
      options: [ { before: true, after: true, overrides: { as: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
        { messageId: 'unexpectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: '{}async function foo() {}',
      output: '{} async function foo() {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'async' } },
      ],
    },
    {
      code: '{} async function foo() {}',
      output: '{}async function foo() {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'async' } },
      ],
    },
    {
      code: '{}async function foo() {}',
      output: '{} async function foo() {}',
      options: [ { before: false, after: false, overrides: { async: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'async' } },
      ],
    },
    {
      code: '{} async function foo() {}',
      output: '{}async function foo() {}',
      options: [ { before: true, after: true, overrides: { async: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'async' } },
      ],
    },
    {
      code: '{}async () => {}',
      output: '{} async () => {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'async' } },
      ],
    },
    {
      code: '{} async () => {}',
      output: '{}async () => {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'async' } },
      ],
    },
    {
      code: '{}async () => {}',
      output: '{} async () => {}',
      options: [ { before: false, after: false, overrides: { async: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'async' } },
      ],
    },
    {
      code: '{} async () => {}',
      output: '{}async () => {}',
      options: [ { before: true, after: true, overrides: { async: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'async' } },
      ],
    },
    {
      code: '({async[b]() {}})',
      output: '({async [b]() {}})',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'async' } },
      ],
    },
    {
      code: '({async [b]() {}})',
      output: '({async[b]() {}})',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'async' } },
      ],
    },
    {
      code: '({async[b]() {}})',
      output: '({async [b]() {}})',
      options: [ { before: false, after: false, overrides: { async: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'async' } },
      ],
    },
    {
      code: '({async [b]() {}})',
      output: '({async[b]() {}})',
      options: [ { before: true, after: true, overrides: { async: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'async' } },
      ],
    },
    {
      code: 'class A {a(){}async[b]() {}}',
      output: 'class A {a(){} async [b]() {}}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'async' } },
        { messageId: 'expectedAfter', data: { value: 'async' } },
      ],
    },
    {
      code: 'class A {a(){} async [b]() {}}',
      output: 'class A {a(){}async[b]() {}}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'async' } },
        { messageId: 'unexpectedAfter', data: { value: 'async' } },
      ],
    },
    {
      code: 'class A {a(){}async[b]() {}}',
      output: 'class A {a(){} async [b]() {}}',
      options: [ { before: false, after: false, overrides: { async: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'async' } },
        { messageId: 'expectedAfter', data: { value: 'async' } },
      ],
    },
    {
      code: 'class A {a(){} async [b]() {}}',
      output: 'class A {a(){}async[b]() {}}',
      options: [ { before: true, after: true, overrides: { async: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'async' } },
        { messageId: 'unexpectedAfter', data: { value: 'async' } },
      ],
    },
    {
      code: 'async function wrap() { {}await a }',
      output: 'async function wrap() { {} await a }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: 'async function wrap() { {} await a }',
      output: 'async function wrap() { {}await a }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: 'async function wrap() { {}await a }',
      output: 'async function wrap() { {} await a }',
      options: [ { before: false, after: false, overrides: { await: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: 'async function wrap() { {} await a }',
      output: 'async function wrap() { {}await a }',
      options: [ { before: true, after: true, overrides: { await: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: 'async function wrap() { for await(x of xs); }',
      output: 'async function wrap() { for await (x of xs); }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'await' } },
      ],
    },
    {
      code: 'async function wrap() { for await (x of xs); }',
      output: 'async function wrap() { for await(x of xs); }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'await' } },
      ],
    },
    {
      code: 'async function wrap() { for await(x of xs); }',
      output: 'async function wrap() { for await (x of xs); }',
      options: [ { before: false, after: false, overrides: { await: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'await' } },
      ],
    },
    {
      code: 'async function wrap() { for await (x of xs); }',
      output: 'async function wrap() { for await(x of xs); }',
      options: [ { before: true, after: true, overrides: { await: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'await' } },
      ],
    },
    {
      code: 'A: for (;;) { {}break A; }',
      output: 'A: for (;;) { {} break A; }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'break' } },
      ],
    },
    {
      code: 'A: for(;;) { {} break A; }',
      output: 'A: for(;;) { {}break A; }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'break' } },
      ],
    },
    {
      code: 'A: for(;;) { {}break A; }',
      output: 'A: for(;;) { {} break A; }',
      options: [ { before: false, after: false, overrides: { break: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'break' } },
      ],
    },
    {
      code: 'A: for (;;) { {} break A; }',
      output: 'A: for (;;) { {}break A; }',
      options: [ { before: true, after: true, overrides: { break: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'break' } },
      ],
    },
    {
      code: 'switch (a) { case 0: {}case+1: }',
      output: 'switch (a) { case 0: {} case +1: }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'case' } },
        { messageId: 'expectedAfter', data: { value: 'case' } },
      ],
    },
    {
      code: 'switch (a) { case 0: {}case(1): }',
      output: 'switch (a) { case 0: {} case (1): }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'case' } },
        { messageId: 'expectedAfter', data: { value: 'case' } },
      ],
    },
    {
      code: 'switch(a) { case 0: {} case +1: }',
      output: 'switch(a) { case 0: {}case+1: }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'case' } },
        { messageId: 'unexpectedAfter', data: { value: 'case' } },
      ],
    },
    {
      code: 'switch(a) { case 0: {} case (1): }',
      output: 'switch(a) { case 0: {}case(1): }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'case' } },
        { messageId: 'unexpectedAfter', data: { value: 'case' } },
      ],
    },
    {
      code: 'switch(a) { case 0: {}case+1: }',
      output: 'switch(a) { case 0: {} case +1: }',
      options: [ { before: false, after: false, overrides: { case: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'case' } },
        { messageId: 'expectedAfter', data: { value: 'case' } },
      ],
    },
    {
      code: 'switch (a) { case 0: {} case +1: }',
      output: 'switch (a) { case 0: {}case+1: }',
      options: [ { before: true, after: true, overrides: { case: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'case' } },
        { messageId: 'unexpectedAfter', data: { value: 'case' } },
      ],
    },
    {
      code: 'try {}catch{}',
      output: 'try {} catch {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'catch' } },
        { messageId: 'expectedAfter', data: { value: 'catch' } },
      ],
    },
    {
      code: 'try{} catch {}',
      output: 'try{}catch{}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'catch' } },
        { messageId: 'unexpectedAfter', data: { value: 'catch' } },
      ],
    },
    {
      code: 'try{}catch{}',
      output: 'try{} catch {}',
      options: [ { before: false, after: false, overrides: { catch: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'catch' } },
        { messageId: 'expectedAfter', data: { value: 'catch' } },
      ],
    },
    {
      code: 'try {} catch {}',
      output: 'try {}catch{}',
      options: [ { before: true, after: true, overrides: { catch: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'catch' } },
        { messageId: 'unexpectedAfter', data: { value: 'catch' } },
      ],
    },
    {
      code: 'try {}catch(e) {}',
      output: 'try {} catch(e) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'catch' } },
      ],
    },
    {
      code: 'try{} catch (e) {}',
      output: 'try{}catch (e) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'catch' } },
      ],
    },
    {
      code: 'try{}catch(e) {}',
      output: 'try{} catch(e) {}',
      options: [ { before: false, after: false, overrides: { catch: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'catch' } },
      ],
    },
    {
      code: 'try {} catch (e) {}',
      output: 'try {}catch (e) {}',
      options: [ { before: true, after: true, overrides: { catch: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'catch' } },
      ],
    },
    {
      code: '{}class Bar {}',
      output: '{} class Bar {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'class' } },
      ],
    },
    {
      code: '(class{})',
      output: '(class {})',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'class' } },
      ],
    },
    {
      code: '{} class Bar {}',
      output: '{}class Bar {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'class' } },
      ],
    },
    {
      code: '(class {})',
      output: '(class{})',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'class' } },
      ],
    },
    {
      code: '{}class Bar {}',
      output: '{} class Bar {}',
      options: [ { before: false, after: false, overrides: { class: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'class' } },
      ],
    },
    {
      code: '{} class Bar {}',
      output: '{}class Bar {}',
      options: [ { before: true, after: true, overrides: { class: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'class' } },
      ],
    },
    {
      code: '{}const[a] = b',
      output: '{} const [a] = b',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'const' } },
        { messageId: 'expectedAfter', data: { value: 'const' } },
      ],
    },
    {
      code: '{}const{a} = b',
      output: '{} const {a} = b',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'const' } },
        { messageId: 'expectedAfter', data: { value: 'const' } },
      ],
    },
    {
      code: '{} const [a] = b',
      output: '{}const[a] = b',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'const' } },
        { messageId: 'unexpectedAfter', data: { value: 'const' } },
      ],
    },
    {
      code: '{} const {a} = b',
      output: '{}const{a} = b',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'const' } },
        { messageId: 'unexpectedAfter', data: { value: 'const' } },
      ],
    },
    {
      code: '{}const[a] = b',
      output: '{} const [a] = b',
      options: [ { before: false, after: false, overrides: { const: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'const' } },
        { messageId: 'expectedAfter', data: { value: 'const' } },
      ],
    },
    {
      code: '{}const{a} = b',
      output: '{} const {a} = b',
      options: [ { before: false, after: false, overrides: { const: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'const' } },
        { messageId: 'expectedAfter', data: { value: 'const' } },
      ],
    },
    {
      code: '{} const [a] = b',
      output: '{}const[a] = b',
      options: [ { before: true, after: true, overrides: { const: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'const' } },
        { messageId: 'unexpectedAfter', data: { value: 'const' } },
      ],
    },
    {
      code: '{} const {a} = b',
      output: '{}const{a} = b',
      options: [ { before: true, after: true, overrides: { const: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'const' } },
        { messageId: 'unexpectedAfter', data: { value: 'const' } },
      ],
    },
    {
      code: '{}using a = b',
      output: '{} using a = b',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'using' } },
      ],
    },
    {
      code: '{} using a = b',
      output: '{}using a = b',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'using' } },
      ],
    },
    {
      code: '{}using a = b',
      output: '{} using a = b',
      options: [ { before: false, after: false, overrides: { using: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'using' } },
      ],
    },
    {
      code: '{} using a = b',
      output: '{}using a = b',
      options: [ { before: true, after: true, overrides: { using: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'using' } },
      ],
    },
    {
      code: '{}await using a = b',
      output: '{} await using a = b',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: '{} await using a = b',
      output: '{}await using a = b',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: '{}await using a = b',
      output: '{} await using a = b',
      options: [ { before: false, after: false, overrides: { await: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: '{} await using a = b',
      output: '{}await using a = b',
      options: [ { before: true, after: true, overrides: { await: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: 'A: for (;;) { {}continue A; }',
      output: 'A: for (;;) { {} continue A; }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'continue' } },
      ],
    },
    {
      code: 'A: for(;;) { {} continue A; }',
      output: 'A: for(;;) { {}continue A; }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'continue' } },
      ],
    },
    {
      code: 'A: for(;;) { {}continue A; }',
      output: 'A: for(;;) { {} continue A; }',
      options: [ { before: false, after: false, overrides: { continue: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'continue' } },
      ],
    },
    {
      code: 'A: for (;;) { {} continue A; }',
      output: 'A: for (;;) { {}continue A; }',
      options: [ { before: true, after: true, overrides: { continue: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'continue' } },
      ],
    },
    {
      code: '{}debugger',
      output: '{} debugger',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'debugger' } },
      ],
    },
    {
      code: '{} debugger',
      output: '{}debugger',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'debugger' } },
      ],
    },
    {
      code: '{}debugger',
      output: '{} debugger',
      options: [ { before: false, after: false, overrides: { debugger: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'debugger' } },
      ],
    },
    {
      code: '{} debugger',
      output: '{}debugger',
      options: [ { before: true, after: true, overrides: { debugger: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'debugger' } },
      ],
    },
    {
      code: 'switch (a) { case 0: {}default: }',
      output: 'switch (a) { case 0: {} default: }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'default' } },
      ],
    },
    {
      code: 'switch(a) { case 0: {} default: }',
      output: 'switch(a) { case 0: {}default: }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'default' } },
      ],
    },
    {
      code: 'switch(a) { case 0: {}default: }',
      output: 'switch(a) { case 0: {} default: }',
      options: [ { before: false, after: false, overrides: { default: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'default' } },
      ],
    },
    {
      code: 'switch (a) { case 0: {} default: }',
      output: 'switch (a) { case 0: {}default: }',
      options: [ { before: true, after: true, overrides: { default: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'default' } },
      ],
    },
    {
      code: '{}delete foo.a',
      output: '{} delete foo.a',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'delete' } },
      ],
    },
    {
      code: '{} delete foo.a',
      output: '{}delete foo.a',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'delete' } },
      ],
    },
    {
      code: '{}delete foo.a',
      output: '{} delete foo.a',
      options: [ { before: false, after: false, overrides: { delete: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'delete' } },
      ],
    },
    {
      code: '{} delete foo.a',
      output: '{}delete foo.a',
      options: [ { before: true, after: true, overrides: { delete: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'delete' } },
      ],
    },
    {
      code: '{}do{} while (true)',
      output: '{} do {} while (true)',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'do' } },
        { messageId: 'expectedAfter', data: { value: 'do' } },
      ],
    },
    {
      code: '{} do {}while(true)',
      output: '{}do{}while(true)',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'do' } },
        { messageId: 'unexpectedAfter', data: { value: 'do' } },
      ],
    },
    {
      code: '{}do{}while(true)',
      output: '{} do {}while(true)',
      options: [ { before: false, after: false, overrides: { do: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'do' } },
        { messageId: 'expectedAfter', data: { value: 'do' } },
      ],
    },
    {
      code: '{} do {} while (true)',
      output: '{}do{} while (true)',
      options: [ { before: true, after: true, overrides: { do: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'do' } },
        { messageId: 'unexpectedAfter', data: { value: 'do' } },
      ],
    },
    {
      code: 'if (a) {}else{}',
      output: 'if (a) {} else {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'else' } },
        { messageId: 'expectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if (a) {}else if (b) {}',
      output: 'if (a) {} else if (b) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'else' } },
      ],
    },
    {
      code: 'if (a) {}else(0)',
      output: 'if (a) {} else (0)',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'else' } },
        { messageId: 'expectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if (a) {}else[]',
      output: 'if (a) {} else []',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'else' } },
        { messageId: 'expectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if (a) {}else+1',
      output: 'if (a) {} else +1',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'else' } },
        { messageId: 'expectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if (a) {}else"a"',
      output: 'if (a) {} else "a"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'else' } },
        { messageId: 'expectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a){} else {}',
      output: 'if(a){}else{}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'else' } },
        { messageId: 'unexpectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a){} else if(b) {}',
      output: 'if(a){}else if(b) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a) {} else (0)',
      output: 'if(a) {}else(0)',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'else' } },
        { messageId: 'unexpectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a) {} else []',
      output: 'if(a) {}else[]',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'else' } },
        { messageId: 'unexpectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a) {} else +1',
      output: 'if(a) {}else+1',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'else' } },
        { messageId: 'unexpectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a) {} else "a"',
      output: 'if(a) {}else"a"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'else' } },
        { messageId: 'unexpectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a) {}else{}',
      output: 'if(a) {} else {}',
      options: [ { before: false, after: false, overrides: { else: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'else' } },
        { messageId: 'expectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if (a) {} else {}',
      output: 'if (a) {}else{}',
      options: [ { before: true, after: true, overrides: { else: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'else' } },
        { messageId: 'unexpectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if (a) {}else {}',
      output: 'if (a) {} else {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'else' } },
      ],
    },
    {
      code: 'if (a) {} else{}',
      output: 'if (a) {} else {}',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a) {} else{}',
      output: 'if(a) {}else{}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'else' } },
      ],
    },
    {
      code: 'if(a) {}else {}',
      output: 'if(a) {}else{}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'else' } },
      ],
    },
    {
      code: 'var a = 0; {}export{a}',
      output: 'var a = 0; {} export {a}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'export' } },
        { messageId: 'expectedAfter', data: { value: 'export' } },
      ],
    },
    {
      code: 'var a = 0; {}export default a',
      output: 'var a = 0; {} export default a',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'export' } },
      ],
    },
    {
      code: 'var a = 0; export default{a}',
      output: 'var a = 0; export default {a}',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'default' } },
      ],
    },
    {
      code: '{}export* from "a"',
      output: '{} export * from "a"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'export' } },
        { messageId: 'expectedAfter', data: { value: 'export' } },
      ],
    },
    {
      code: 'var a = 0; {} export {a}',
      output: 'var a = 0; {}export{a}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'export' } },
        { messageId: 'unexpectedAfter', data: { value: 'export' } },
      ],
    },
    {
      code: 'var a = 0; {}export{a}',
      output: 'var a = 0; {} export {a}',
      options: [ { before: false, after: false, overrides: { export: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'export' } },
        { messageId: 'expectedAfter', data: { value: 'export' } },
      ],
    },
    {
      code: 'var a = 0; {} export {a}',
      output: 'var a = 0; {}export{a}',
      options: [ { before: true, after: true, overrides: { export: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'export' } },
        { messageId: 'unexpectedAfter', data: { value: 'export' } },
      ],
    },
    {
      code: 'class Bar extends[] {}',
      output: 'class Bar extends [] {}',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'extends' } },
      ],
    },
    {
      code: '(class extends[] {})',
      output: '(class extends [] {})',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'extends' } },
      ],
    },
    {
      code: 'class Bar extends [] {}',
      output: 'class Bar extends[] {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'extends' } },
      ],
    },
    {
      code: '(class extends [] {})',
      output: '(class extends[] {})',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'extends' } },
      ],
    },
    {
      code: 'class Bar extends[] {}',
      output: 'class Bar extends [] {}',
      options: [ { before: false, after: false, overrides: { extends: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'extends' } },
      ],
    },
    {
      code: 'class Bar extends [] {}',
      output: 'class Bar extends[] {}',
      options: [ { before: true, after: true, overrides: { extends: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'extends' } },
      ],
    },
    {
      code: 'class Bar extends`}` {}',
      output: 'class Bar extends `}` {}',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'extends' } },
      ],
    },
    {
      code: 'try {}finally{}',
      output: 'try {} finally {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'finally' } },
        { messageId: 'expectedAfter', data: { value: 'finally' } },
      ],
    },
    {
      code: 'try{} finally {}',
      output: 'try{}finally{}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'finally' } },
        { messageId: 'unexpectedAfter', data: { value: 'finally' } },
      ],
    },
    {
      code: 'try{}finally{}',
      output: 'try{} finally {}',
      options: [ { before: false, after: false, overrides: { finally: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'finally' } },
        { messageId: 'expectedAfter', data: { value: 'finally' } },
      ],
    },
    {
      code: 'try {} finally {}',
      output: 'try {}finally{}',
      options: [ { before: true, after: true, overrides: { finally: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'finally' } },
        { messageId: 'unexpectedAfter', data: { value: 'finally' } },
      ],
    },
    {
      code: '{}for(;;) {}',
      output: '{} for (;;) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'for' } },
        { messageId: 'expectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{}for(var foo in obj) {}',
      output: '{} for (var foo in obj) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'for' } },
        { messageId: 'expectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{}for(var foo of list) {}',
      output: '{} for (var foo of list) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'for' } },
        { messageId: 'expectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{} for (;;) {}',
      output: '{}for(;;) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'for' } },
        { messageId: 'unexpectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{} for (var foo in obj) {}',
      output: '{}for(var foo in obj) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'for' } },
        { messageId: 'unexpectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{} for (var foo of list) {}',
      output: '{}for(var foo of list) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'for' } },
        { messageId: 'unexpectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{}for(;;) {}',
      output: '{} for (;;) {}',
      options: [ { before: false, after: false, overrides: { for: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'for' } },
        { messageId: 'expectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{}for(var foo in obj) {}',
      output: '{} for (var foo in obj) {}',
      options: [ { before: false, after: false, overrides: { for: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'for' } },
        { messageId: 'expectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{}for(var foo of list) {}',
      output: '{} for (var foo of list) {}',
      options: [ { before: false, after: false, overrides: { for: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'for' } },
        { messageId: 'expectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{} for (;;) {}',
      output: '{}for(;;) {}',
      options: [ { before: true, after: true, overrides: { for: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'for' } },
        { messageId: 'unexpectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{} for (var foo in obj) {}',
      output: '{}for(var foo in obj) {}',
      options: [ { before: true, after: true, overrides: { for: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'for' } },
        { messageId: 'unexpectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: '{} for (var foo of list) {}',
      output: '{}for(var foo of list) {}',
      options: [ { before: true, after: true, overrides: { for: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'for' } },
        { messageId: 'unexpectedAfter', data: { value: 'for' } },
      ],
    },
    {
      code: 'import {foo}from"foo"',
      output: 'import {foo} from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'from' } },
        { messageId: 'expectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export {foo}from"foo"',
      output: 'export {foo} from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'from' } },
        { messageId: 'expectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export *from"foo"',
      output: 'export * from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'from' } },
        { messageId: 'expectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export * as "a"from"foo"',
      output: 'export * as "a" from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'from' } },
        { messageId: 'expectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'import{foo} from "foo"',
      output: 'import{foo}from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export{foo} from "foo"',
      output: 'export{foo}from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export* from "foo"',
      output: 'export*from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export*as x from "foo"',
      output: 'export*as x from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export*as"x" from "foo"',
      output: 'export*as"x"from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'import{foo}from"foo"',
      output: 'import{foo} from "foo"',
      options: [ { before: false, after: false, overrides: { from: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'from' } },
        { messageId: 'expectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export{foo}from"foo"',
      output: 'export{foo} from "foo"',
      options: [ { before: false, after: false, overrides: { from: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'from' } },
        { messageId: 'expectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export*from"foo"',
      output: 'export* from "foo"',
      options: [ { before: false, after: false, overrides: { from: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'from' } },
        { messageId: 'expectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export*as"x"from"foo"',
      output: 'export*as"x" from "foo"',
      options: [ { before: false, after: false, overrides: { from: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'from' } },
        { messageId: 'expectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'import {foo} from "foo"',
      output: 'import {foo}from"foo"',
      options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export {foo} from "foo"',
      output: 'export {foo}from"foo"',
      options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export * from "foo"',
      output: 'export *from"foo"',
      options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export * as x from "foo"',
      output: 'export * as x from"foo"',
      options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export * as "x" from "foo"',
      output: 'export * as "x"from"foo"',
      options: [ { before: true, after: true, overrides: { from: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: '{}function foo() {}',
      output: '{} function foo() {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'function' } },
      ],
    },
    {
      code: '{} function foo() {}',
      output: '{}function foo() {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'function' } },
      ],
    },
    {
      code: '{}function foo() {}',
      output: '{} function foo() {}',
      options: [ { before: false, after: false, overrides: { function: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'function' } },
      ],
    },
    {
      code: '{} function foo() {}',
      output: '{}function foo() {}',
      options: [ { before: true, after: true, overrides: { function: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'function' } },
      ],
    },
    {
      code: '({ get[b]() {} })',
      output: '({ get [b]() {} })',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: 'class A { a() {}get[b]() {} }',
      output: 'class A { a() {} get [b]() {} }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'get' } },
        { messageId: 'expectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: 'class A { a() {} static get[b]() {} }',
      output: 'class A { a() {} static get [b]() {} }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: '({ get [b]() {} })',
      output: '({ get[b]() {} })',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: 'class A { a() {} get [b]() {} }',
      output: 'class A { a() {}get[b]() {} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'get' } },
        { messageId: 'unexpectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: 'class A { a() {}static get [b]() {} }',
      output: 'class A { a() {}static get[b]() {} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: '({ get[b]() {} })',
      output: '({ get [b]() {} })',
      options: [ { before: false, after: false, overrides: { get: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: 'class A { a() {}get[b]() {} }',
      output: 'class A { a() {} get [b]() {} }',
      options: [ { before: false, after: false, overrides: { get: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'get' } },
        { messageId: 'expectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: '({ get [b]() {} })',
      output: '({ get[b]() {} })',
      options: [ { before: true, after: true, overrides: { get: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: 'class A { a() {} get [b]() {} }',
      output: 'class A { a() {}get[b]() {} }',
      options: [ { before: true, after: true, overrides: { get: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'get' } },
        { messageId: 'unexpectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: 'class A { a;get#b() {} }',
      output: 'class A { a;get #b() {} }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: 'class A { a; get #b() {} }',
      output: 'class A { a; get#b() {} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'get' } },
      ],
    },
    {
      code: '{}if(a) {}',
      output: '{} if (a) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'if' } },
        { messageId: 'expectedAfter', data: { value: 'if' } },
      ],
    },
    {
      code: 'if (a) {} else if(b) {}',
      output: 'if (a) {} else if (b) {}',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'if' } },
      ],
    },
    {
      code: '{} if (a) {}',
      output: '{}if(a) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'if' } },
        { messageId: 'unexpectedAfter', data: { value: 'if' } },
      ],
    },
    {
      code: 'if(a) {}else if (b) {}',
      output: 'if(a) {}else if(b) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'if' } },
      ],
    },
    {
      code: '{}if(a) {}',
      output: '{} if (a) {}',
      options: [ { before: false, after: false, overrides: { if: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'if' } },
        { messageId: 'expectedAfter', data: { value: 'if' } },
      ],
    },
    {
      code: 'if (a) {}else if(b) {}',
      output: 'if (a) {}else if (b) {}',
      options: [ { before: false, after: false, overrides: { if: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'if' } },
      ],
    },
    {
      code: '{} if (a) {}',
      output: '{}if(a) {}',
      options: [ { before: true, after: true, overrides: { if: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'if' } },
        { messageId: 'unexpectedAfter', data: { value: 'if' } },
      ],
    },
    {
      code: 'if(a) {} else if (b) {}',
      output: 'if(a) {} else if(b) {}',
      options: [ { before: true, after: true, overrides: { if: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'if' } },
      ],
    },
    {
      code: 'import* as a from "foo"',
      output: 'import * as a from "foo"',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'import' }, line: 1, column: 1, endLine: 1, endColumn: 7 },
      ],
    },
    {
      code: 'import *as a from"foo"',
      output: 'import*as a from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'import' }, line: 1, column: 7, endLine: 1, endColumn: 8 },
      ],
    },
    {
      code: 'import   *as a from"foo"',
      output: 'import*as a from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'import' }, line: 1, column: 7, endLine: 1, endColumn: 10 },
      ],
    },
    {
      code: '{}import{a} from "foo"',
      output: '{} import {a} from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'import' } },
        { messageId: 'expectedAfter', data: { value: 'import' } },
      ],
    },
    {
      code: '{}import a from "foo"',
      output: '{} import a from "foo"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'import' } },
      ],
    },
    {
      code: '{}import* as a from "a"',
      output: '{} import * as a from "a"',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'import' } },
        { messageId: 'expectedAfter', data: { value: 'import' } },
      ],
    },
    {
      code: '{} import {a}from"foo"',
      output: '{}import{a}from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'import' } },
        { messageId: 'unexpectedAfter', data: { value: 'import' } },
      ],
    },
    {
      code: '{} import *as a from"foo"',
      output: '{}import*as a from"foo"',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'import' } },
        { messageId: 'unexpectedAfter', data: { value: 'import' } },
      ],
    },
    {
      code: '{}import{a}from"foo"',
      output: '{} import {a}from"foo"',
      options: [ { before: false, after: false, overrides: { import: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'import' } },
        { messageId: 'expectedAfter', data: { value: 'import' } },
      ],
    },
    {
      code: '{}import*as a from"foo"',
      output: '{} import *as a from"foo"',
      options: [ { before: false, after: false, overrides: { import: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'import' } },
        { messageId: 'expectedAfter', data: { value: 'import' } },
      ],
    },
    {
      code: '{} import {a} from "foo"',
      output: '{}import{a} from "foo"',
      options: [ { before: true, after: true, overrides: { import: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'import' } },
        { messageId: 'unexpectedAfter', data: { value: 'import' } },
      ],
    },
    {
      code: '{} import * as a from "foo"',
      output: '{}import* as a from "foo"',
      options: [ { before: true, after: true, overrides: { import: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'import' } },
        { messageId: 'unexpectedAfter', data: { value: 'import' } },
      ],
    },
    {
      code: 'for ([foo]in{foo: 0}) {}',
      output: 'for ([foo] in {foo: 0}) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'in' } },
        { messageId: 'expectedAfter', data: { value: 'in' } },
      ],
    },
    {
      code: 'for([foo] in {foo: 0}) {}',
      output: 'for([foo]in{foo: 0}) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'in' } },
        { messageId: 'unexpectedAfter', data: { value: 'in' } },
      ],
    },
    {
      code: 'for([foo]in{foo: 0}) {}',
      output: 'for([foo] in {foo: 0}) {}',
      options: [ { before: false, after: false, overrides: { in: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'in' } },
        { messageId: 'expectedAfter', data: { value: 'in' } },
      ],
    },
    {
      code: 'for ([foo] in {foo: 0}) {}',
      output: 'for ([foo]in{foo: 0}) {}',
      options: [ { before: true, after: true, overrides: { in: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'in' } },
        { messageId: 'unexpectedAfter', data: { value: 'in' } },
      ],
    },
    {
      code: '{}let[a] = b',
      output: '{} let [a] = b',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'let' } },
        { messageId: 'expectedAfter', data: { value: 'let' } },
      ],
    },
    {
      code: '{} let [a] = b',
      output: '{}let[a] = b',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'let' } },
        { messageId: 'unexpectedAfter', data: { value: 'let' } },
      ],
    },
    {
      code: '{}let[a] = b',
      output: '{} let [a] = b',
      options: [ { before: false, after: false, overrides: { let: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'let' } },
        { messageId: 'expectedAfter', data: { value: 'let' } },
      ],
    },
    {
      code: '{} let [a] = b',
      output: '{}let[a] = b',
      options: [ { before: true, after: true, overrides: { let: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'let' } },
        { messageId: 'unexpectedAfter', data: { value: 'let' } },
      ],
    },
    {
      code: '{}new foo()',
      output: '{} new foo()',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'new' } },
      ],
    },
    {
      code: '{} new foo()',
      output: '{}new foo()',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'new' } },
      ],
    },
    {
      code: '{}new foo()',
      output: '{} new foo()',
      options: [ { before: false, after: false, overrides: { new: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'new' } },
      ],
    },
    {
      code: '{} new foo()',
      output: '{}new foo()',
      options: [ { before: true, after: true, overrides: { new: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'new' } },
      ],
    },
    {
      code: 'for ([foo]of{foo: 0}) {}',
      output: 'for ([foo] of {foo: 0}) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'of' } },
        { messageId: 'expectedAfter', data: { value: 'of' } },
      ],
    },
    {
      code: 'for([foo] of {foo: 0}) {}',
      output: 'for([foo]of{foo: 0}) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'of' } },
        { messageId: 'unexpectedAfter', data: { value: 'of' } },
      ],
    },
    {
      code: 'for([foo]of{foo: 0}) {}',
      output: 'for([foo] of {foo: 0}) {}',
      options: [ { before: false, after: false, overrides: { of: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'of' } },
        { messageId: 'expectedAfter', data: { value: 'of' } },
      ],
    },
    {
      code: 'for ([foo] of {foo: 0}) {}',
      output: 'for ([foo]of{foo: 0}) {}',
      options: [ { before: true, after: true, overrides: { of: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'of' } },
        { messageId: 'unexpectedAfter', data: { value: 'of' } },
      ],
    },
    {
      code: 'function foo() { {}return+a }',
      output: 'function foo() { {} return +a }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'return' } },
        { messageId: 'expectedAfter', data: { value: 'return' } },
      ],
    },
    {
      code: 'function foo() { return<p/>; }',
      output: 'function foo() { return <p/>; }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'return' } },
      ],
    },
    {
      code: 'function foo() { {} return +a }',
      output: 'function foo() { {}return+a }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'return' } },
        { messageId: 'unexpectedAfter', data: { value: 'return' } },
      ],
    },
    {
      code: 'function foo() { return <p/>; }',
      output: 'function foo() { return<p/>; }',
      options: [ { after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'return' } },
      ],
    },
    {
      code: 'function foo() { {}return+a }',
      output: 'function foo() { {} return +a }',
      options: [ { before: false, after: false, overrides: { return: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'return' } },
        { messageId: 'expectedAfter', data: { value: 'return' } },
      ],
    },
    {
      code: 'function foo() { {} return +a }',
      output: 'function foo() { {}return+a }',
      options: [ { before: true, after: true, overrides: { return: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'return' } },
        { messageId: 'unexpectedAfter', data: { value: 'return' } },
      ],
    },
    {
      code: '({ set[b](value) {} })',
      output: '({ set [b](value) {} })',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: 'class A { a() {}set[b](value) {} }',
      output: 'class A { a() {} set [b](value) {} }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'set' } },
        { messageId: 'expectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: 'class A { a() {} static set[b](value) {} }',
      output: 'class A { a() {} static set [b](value) {} }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: '({ set [b](value) {} })',
      output: '({ set[b](value) {} })',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: 'class A { a() {} set [b](value) {} }',
      output: 'class A { a() {}set[b](value) {} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'set' } },
        { messageId: 'unexpectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: '({ set[b](value) {} })',
      output: '({ set [b](value) {} })',
      options: [ { before: false, after: false, overrides: { set: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: 'class A { a() {}set[b](value) {} }',
      output: 'class A { a() {} set [b](value) {} }',
      options: [ { before: false, after: false, overrides: { set: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'set' } },
        { messageId: 'expectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: '({ set [b](value) {} })',
      output: '({ set[b](value) {} })',
      options: [ { before: true, after: true, overrides: { set: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: 'class A { a() {} set [b](value) {} }',
      output: 'class A { a() {}set[b](value) {} }',
      options: [ { before: true, after: true, overrides: { set: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'set' } },
        { messageId: 'unexpectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: 'class A { a;set#b(x) {} }',
      output: 'class A { a;set #b(x) {} }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: 'class A { a; set #b(x) {} }',
      output: 'class A { a; set#b(x) {} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'set' } },
      ],
    },
    {
      code: 'class A { a() {}static[b]() {} }',
      output: 'class A { a() {} static [b]() {} }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'static' } },
        { messageId: 'expectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {}static get [b]() {} }',
      output: 'class A { a() {} static get [b]() {} }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {} static [b]() {} }',
      output: 'class A { a() {}static[b]() {} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'static' } },
        { messageId: 'unexpectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {} static get[b]() {} }',
      output: 'class A { a() {}static get[b]() {} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {}static[b]() {} }',
      output: 'class A { a() {} static [b]() {} }',
      options: [ { before: false, after: false, overrides: { static: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'static' } },
        { messageId: 'expectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {} static [b]() {} }',
      output: 'class A { a() {}static[b]() {} }',
      options: [ { before: true, after: true, overrides: { static: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'static' } },
        { messageId: 'unexpectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a;static[b]; }',
      output: 'class A { a;static [b]; }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a; static [b]; }',
      output: 'class A { a; static[b]; }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a;static#b; }',
      output: 'class A { a;static #b; }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a; static #b; }',
      output: 'class A { a; static#b; }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {}static{} }',
      output: 'class A { a() {} static {} }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'static' } },
        { messageId: 'expectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {}static{} }',
      output: 'class A { a() {} static {} }',
      options: [ { before: false, after: false, overrides: { static: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'static' } },
        { messageId: 'expectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A {  a() {}static {} }',
      output: 'class A {  a() {} static {} }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A {  a() {} static{} }',
      output: 'class A {  a() {} static {} }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {} static {} }',
      output: 'class A { a() {}static{} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'static' } },
        { messageId: 'unexpectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {} static {} }',
      output: 'class A { a() {}static{} }',
      options: [ { before: true, after: true, overrides: { static: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'static' } },
        { messageId: 'unexpectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {} static{} }',
      output: 'class A { a() {}static{} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() {}static {} }',
      output: 'class A { a() {}static{} }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'static' } },
      ],
    },
    {
      code: 'class A { a() { {}super[b]; } }',
      output: 'class A { a() { {} super[b]; } }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'super' } },
      ],
    },
    {
      code: 'class A { a() { {} super[b]; } }',
      output: 'class A { a() { {}super[b]; } }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'super' } },
      ],
    },
    {
      code: 'class A { a() { {}super[b]; } }',
      output: 'class A { a() { {} super[b]; } }',
      options: [ { before: false, after: false, overrides: { super: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'super' } },
      ],
    },
    {
      code: 'class A { a() { {} super[b]; } }',
      output: 'class A { a() { {}super[b]; } }',
      options: [ { before: true, after: true, overrides: { super: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'super' } },
      ],
    },
    {
      code: '{}switch(a) {}',
      output: '{} switch (a) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'switch' } },
        { messageId: 'expectedAfter', data: { value: 'switch' } },
      ],
    },
    {
      code: '{} switch (a) {}',
      output: '{}switch(a) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'switch' } },
        { messageId: 'unexpectedAfter', data: { value: 'switch' } },
      ],
    },
    {
      code: '{}switch(a) {}',
      output: '{} switch (a) {}',
      options: [ { before: false, after: false, overrides: { switch: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'switch' } },
        { messageId: 'expectedAfter', data: { value: 'switch' } },
      ],
    },
    {
      code: '{} switch (a) {}',
      output: '{}switch(a) {}',
      options: [ { before: true, after: true, overrides: { switch: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'switch' } },
        { messageId: 'unexpectedAfter', data: { value: 'switch' } },
      ],
    },
    {
      code: '{}this[a]',
      output: '{} this[a]',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'this' } },
      ],
    },
    {
      code: '{} this[a]',
      output: '{}this[a]',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'this' } },
      ],
    },
    {
      code: '{}this[a]',
      output: '{} this[a]',
      options: [ { before: false, after: false, overrides: { this: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'this' } },
      ],
    },
    {
      code: '{} this[a]',
      output: '{}this[a]',
      options: [ { before: true, after: true, overrides: { this: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'this' } },
      ],
    },
    {
      code: '<Thing> this.blah',
      output: '<Thing>this.blah',
      options: [ { before: true, after: false, overrides: { this: { before: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'this' } },
      ],
    },
    {
      code: '<Thing>this.blah',
      output: '<Thing> this.blah',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'this' } },
      ],
    },
    {
      code: 'function foo() { {}throw+a }',
      output: 'function foo() { {} throw +a }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'throw' } },
        { messageId: 'expectedAfter', data: { value: 'throw' } },
      ],
    },
    {
      code: 'function foo() { {} throw +a }',
      output: 'function foo() { {}throw+a }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'throw' } },
        { messageId: 'unexpectedAfter', data: { value: 'throw' } },
      ],
    },
    {
      code: 'function foo() { {}throw+a }',
      output: 'function foo() { {} throw +a }',
      options: [ { before: false, after: false, overrides: { throw: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'throw' } },
        { messageId: 'expectedAfter', data: { value: 'throw' } },
      ],
    },
    {
      code: 'function foo() { {} throw +a }',
      output: 'function foo() { {}throw+a }',
      options: [ { before: true, after: true, overrides: { throw: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'throw' } },
        { messageId: 'unexpectedAfter', data: { value: 'throw' } },
      ],
    },
    {
      code: '{}try{} finally {}',
      output: '{} try {} finally {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'try' } },
        { messageId: 'expectedAfter', data: { value: 'try' } },
      ],
    },
    {
      code: '{} try {}finally{}',
      output: '{}try{}finally{}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'try' } },
        { messageId: 'unexpectedAfter', data: { value: 'try' } },
      ],
    },
    {
      code: '{}try{}finally{}',
      output: '{} try {}finally{}',
      options: [ { before: false, after: false, overrides: { try: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'try' } },
        { messageId: 'expectedAfter', data: { value: 'try' } },
      ],
    },
    {
      code: '{} try {} finally {}',
      output: '{}try{} finally {}',
      options: [ { before: true, after: true, overrides: { try: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'try' } },
        { messageId: 'unexpectedAfter', data: { value: 'try' } },
      ],
    },
    {
      code: '{}typeof foo',
      output: '{} typeof foo',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'typeof' } },
      ],
    },
    {
      code: '{} typeof foo',
      output: '{}typeof foo',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'typeof' } },
      ],
    },
    {
      code: '{}typeof foo',
      output: '{} typeof foo',
      options: [ { before: false, after: false, overrides: { typeof: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'typeof' } },
      ],
    },
    {
      code: '{} typeof foo',
      output: '{}typeof foo',
      options: [ { before: true, after: true, overrides: { typeof: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'typeof' } },
      ],
    },
    {
      code: '{}var[a] = b',
      output: '{} var [a] = b',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'var' } },
        { messageId: 'expectedAfter', data: { value: 'var' } },
      ],
    },
    {
      code: '{} var [a] = b',
      output: '{}var[a] = b',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'var' } },
        { messageId: 'unexpectedAfter', data: { value: 'var' } },
      ],
    },
    {
      code: '{}var[a] = b',
      output: '{} var [a] = b',
      options: [ { before: false, after: false, overrides: { var: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'var' } },
        { messageId: 'expectedAfter', data: { value: 'var' } },
      ],
    },
    {
      code: '{} var [a] = b',
      output: '{}var[a] = b',
      options: [ { before: true, after: true, overrides: { var: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'var' } },
        { messageId: 'unexpectedAfter', data: { value: 'var' } },
      ],
    },
    {
      code: '{}void foo',
      output: '{} void foo',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'void' } },
      ],
    },
    {
      code: '{} void foo',
      output: '{}void foo',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'void' } },
      ],
    },
    {
      code: '{}void foo',
      output: '{} void foo',
      options: [ { before: false, after: false, overrides: { void: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'void' } },
      ],
    },
    {
      code: '{} void foo',
      output: '{}void foo',
      options: [ { before: true, after: true, overrides: { void: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'void' } },
      ],
    },
    {
      code: '{}while(a) {}',
      output: '{} while (a) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'while' } },
        { messageId: 'expectedAfter', data: { value: 'while' } },
      ],
    },
    {
      code: 'do {}while(a)',
      output: 'do {} while (a)',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'while' } },
        { messageId: 'expectedAfter', data: { value: 'while' } },
      ],
    },
    {
      code: '{} while (a) {}',
      output: '{}while(a) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'while' } },
        { messageId: 'unexpectedAfter', data: { value: 'while' } },
      ],
    },
    {
      code: 'do{} while (a)',
      output: 'do{}while(a)',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'while' } },
        { messageId: 'unexpectedAfter', data: { value: 'while' } },
      ],
    },
    {
      code: '{}while(a) {}',
      output: '{} while (a) {}',
      options: [ { before: false, after: false, overrides: { while: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'while' } },
        { messageId: 'expectedAfter', data: { value: 'while' } },
      ],
    },
    {
      code: 'do{}while(a)',
      output: 'do{} while (a)',
      options: [ { before: false, after: false, overrides: { while: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'while' } },
        { messageId: 'expectedAfter', data: { value: 'while' } },
      ],
    },
    {
      code: '{} while (a) {}',
      output: '{}while(a) {}',
      options: [ { before: true, after: true, overrides: { while: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'while' } },
        { messageId: 'unexpectedAfter', data: { value: 'while' } },
      ],
    },
    {
      code: 'do {} while (a)',
      output: 'do {}while(a)',
      options: [ { before: true, after: true, overrides: { while: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'while' } },
        { messageId: 'unexpectedAfter', data: { value: 'while' } },
      ],
    },
    {
      code: '{}with(obj) {}',
      output: '{} with (obj) {}',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'with' } },
        { messageId: 'expectedAfter', data: { value: 'with' } },
      ],
    },
    {
      code: '{} with (obj) {}',
      output: '{}with(obj) {}',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'with' } },
        { messageId: 'unexpectedAfter', data: { value: 'with' } },
      ],
    },
    {
      code: '{}with(obj) {}',
      output: '{} with (obj) {}',
      options: [ { before: false, after: false, overrides: { with: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'with' } },
        { messageId: 'expectedAfter', data: { value: 'with' } },
      ],
    },
    {
      code: '{} with (obj) {}',
      output: '{}with(obj) {}',
      options: [ { before: true, after: true, overrides: { with: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'with' } },
        { messageId: 'unexpectedAfter', data: { value: 'with' } },
      ],
    },
    {
      code: 'function* foo() { {}yield foo }',
      output: 'function* foo() { {} yield foo }',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'yield' } },
      ],
    },
    {
      code: 'function* foo() { {} yield foo }',
      output: 'function* foo() { {}yield foo }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'yield' } },
      ],
    },
    {
      code: 'function* foo() { {}yield foo }',
      output: 'function* foo() { {} yield foo }',
      options: [ { before: false, after: false, overrides: { yield: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'yield' } },
      ],
    },
    {
      code: 'function* foo() { {} yield foo }',
      output: 'function* foo() { {}yield foo }',
      options: [ { before: true, after: true, overrides: { yield: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'yield' } },
      ],
    },
    {
      code: 'class Foo { @desc({set a(value) {}, get a() {}, async c() {}}) async[foo]() {} }',
      output: 'class Foo { @desc({set a(value) {}, get a() {}, async c() {}}) async [foo]() {} }',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'async' } },
      ],
    },

    // ---- from keyword-spacing._ts_.test.ts ----
    {
      code: 'const foo = {}as {};',
      output: 'const foo = {} as {};',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'const foo = {} as{};',
      output: 'const foo = {}as{};',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'as' } },
      ],
    },
    {
      code: 'const foo = {} as{};',
      output: 'const foo = {} as {};',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'const foo = {}as {};',
      output: 'const foo = {}as{};',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'const foo = {} as{};',
      output: 'const foo = {} as {};',
      options: [ { overrides: { as: {} } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'as' } },
      ],
    },
    {
      code: 'const foo = {}satisfies {};',
      output: 'const foo = {} satisfies {};',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'satisfies' } },
      ],
    },
    {
      code: 'const foo = {} satisfies{};',
      output: 'const foo = {}satisfies{};',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'satisfies' } },
      ],
    },
    {
      code: 'const foo = {} satisfies{};',
      output: 'const foo = {} satisfies {};',
      errors: [
        { messageId: 'expectedAfter', data: { value: 'satisfies' } },
      ],
    },
    {
      code: 'const foo = {}satisfies {};',
      output: 'const foo = {}satisfies{};',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'satisfies' } },
      ],
    },
    {
      code: 'const foo = {} satisfies{};',
      output: 'const foo = {} satisfies {};',
      options: [ { overrides: { satisfies: {} } } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'satisfies' } },
      ],
    },
    {
      code: 'class C { @readonly() accessor foo = 1 }',
      output: 'class C { @readonly()accessor foo = 1 }',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'accessor' } },
      ],
    },
    {
      code: 'import type{ foo } from "foo";',
      output: 'import type { foo } from "foo";',
      options: [ { after: true, before: true } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: 'import type { foo } from"foo";',
      output: 'import type{ foo } from"foo";',
      options: [ { after: false, before: true } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: 'import type* as foo from "foo";',
      output: 'import type * as foo from "foo";',
      options: [ { after: true, before: true } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: 'import type * as foo from"foo";',
      output: 'import type* as foo from"foo";',
      options: [ { after: false, before: true } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: "import type {SavedQueries} from './SavedQueries.js';",
      output: "import type{SavedQueries} from './SavedQueries.js';",
      options: [ { before: true, after: false, overrides: { from: { after: true } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: "import type {SavedQueries} from './SavedQueries.js';",
      output: "import type{SavedQueries} from'./SavedQueries.js';",
      options: [ { before: true, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export type{ foo } from "foo";',
      output: 'export type { foo } from "foo";',
      options: [ { after: true, before: true } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: 'export type { foo } from"foo";',
      output: 'export type{ foo } from"foo";',
      options: [ { after: false, before: true } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: 'export type* as foo from "foo";',
      output: 'export type * as foo from "foo";',
      options: [ { after: true, before: true } ],
      errors: [
        { messageId: 'expectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: 'export type * as foo from"foo";',
      output: 'export type* as foo from"foo";',
      options: [ { after: false, before: true } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: "export type {SavedQueries} from './SavedQueries.js';",
      output: "export type{SavedQueries} from './SavedQueries.js';",
      options: [ { before: true, after: false, overrides: { from: { after: true } } } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
      ],
    },
    {
      code: "export type {SavedQueries} from './SavedQueries.js';",
      output: "export type{SavedQueries} from'./SavedQueries.js';",
      options: [ { before: true, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'import type {} from "foo";',
      output: 'import type{}from"foo";',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: 'export type {} from "foo";',
      output: 'export type{}from"foo";',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedAfter', data: { value: 'type' } },
        { messageId: 'unexpectedBefore', data: { value: 'from' } },
        { messageId: 'unexpectedAfter', data: { value: 'from' } },
      ],
    },
    {
      code: "import pkgJson from'package.json' with { type: 'json' }",
      output: "import pkgJson from'package.json'with{ type: 'json' }",
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'with' } },
        { messageId: 'unexpectedAfter', data: { value: 'with' } },
      ],
    },
    {
      code: "export { name } from 'package.json'with{ type: 'json' }",
      output: "export { name } from 'package.json' with { type: 'json' }",
      options: [ { before: true, after: true } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'with' } },
        { messageId: 'expectedAfter', data: { value: 'with' } },
      ],
    },
    {
      code: "export*from'package.json'with{ type: 'json' }",
      output: "export*from'package.json' with { type: 'json' }",
      options: [ { before: false, after: false, overrides: { with: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'with' } },
        { messageId: 'expectedAfter', data: { value: 'with' } },
      ],
    },
    {
      code: '{}using a = b',
      output: '{} using a = b',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'using' } },
      ],
    },
    {
      code: '{} using a = b',
      output: '{}using a = b',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'using' } },
      ],
    },
    {
      code: '{}using a = b',
      output: '{} using a = b',
      options: [ { before: false, after: false, overrides: { using: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'using' } },
      ],
    },
    {
      code: '{} using a = b',
      output: '{}using a = b',
      options: [ { before: true, after: true, overrides: { using: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'using' } },
      ],
    },
    {
      code: '{}await using a = b',
      output: '{} await using a = b',
      errors: [
        { messageId: 'expectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: '{} await using a = b',
      output: '{}await using a = b',
      options: [ { before: false, after: false } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: '{}await using a = b',
      output: '{} await using a = b',
      options: [ { before: false, after: false, overrides: { await: { before: true, after: true } } } ],
      errors: [
        { messageId: 'expectedBefore', data: { value: 'await' } },
      ],
    },
    {
      code: '{} await using a = b',
      output: '{}await using a = b',
      options: [ { before: true, after: true, overrides: { await: { before: false, after: false } } } ],
      errors: [
        { messageId: 'unexpectedBefore', data: { value: 'await' } },
      ],
    },
  ],
});
