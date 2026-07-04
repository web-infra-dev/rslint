import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const objectMessage = 'Do not add `then` to an object.';
const exportMessage = 'Do not export `then`.';
const classMessage = 'Do not add `then` to a class.';

const valid = (code: string) => ({ code, filename: 'file.js' });
const invalidObject = (code: string) => ({
  code,
  filename: 'file.js',
  errors: [{ message: objectMessage }],
});
const invalidExport = (code: string) => ({
  code,
  filename: 'file.js',
  errors: [{ message: exportMessage }],
});
const invalidClass = (code: string) => ({
  code,
  filename: 'file.js',
  errors: [{ message: classMessage }],
});

ruleTester.run('no-thenable', null as never, {
  valid: [
    // `object`
    valid('const then = {}'),
    valid('const notThen = then'),
    valid('const then = then.then'),
    valid('const foo = {notThen: 1}'),
    valid('const foo = {notThen() {}}'),
    valid('const foo = {[then]: 1}'),
    valid('const NOT_THEN = "no-then";const foo = {[NOT_THEN]: 1}'),
    valid('function foo({then}) {}'),
    valid('({[Symbol.prototype]: 1})'),

    // `class`
    valid('class then {}'),
    valid('class Foo {notThen}'),
    valid('class Foo {notThen() {}}'),
    valid('class Foo {[then]}'),
    valid('class Foo {#then}'),
    valid('class Foo {#then() {}}'),
    valid('class Foo {[then]() {}}'),
    valid('class Foo {get notThen() {}}'),
    valid('class Foo {get #then() {}}'),
    valid('class Foo {get [then]() {}}'),
    valid('class Foo {static notThen}'),
    valid('class Foo {static notThen() {}}'),
    valid('class Foo {static #then}'),
    valid('class Foo {static #then() {}}'),
    valid('class Foo {static [then]}'),
    valid('class Foo {static [then]() {}}'),
    valid('class Foo {static get notThen() {}}'),
    valid('class Foo {static get #then() {}}'),
    valid('class Foo {static get [then]() {}}'),
    valid('class Foo {notThen = then}'),
    valid('class Foo {[Symbol.property]}'),
    valid('class Foo {static [Symbol.property]}'),
    valid('class Foo {get [Symbol.property]() {}}'),
    valid('class Foo {[Symbol.property]() {}}'),
    valid('class Foo {static get [Symbol.property]() {}}'),

    // Assign
    valid('foo[then] = 1'),
    valid('foo.notThen = 1'),
    valid('then.notThen = then.then'),
    valid('const NOT_THEN = "no-then";foo[NOT_THEN] = 1'),
    valid('foo.then ++'),
    valid('++ foo.then'),
    valid('delete foo.then'),
    valid('typeof foo.then'),
    valid('foo.then != 1'),
    valid('foo[Symbol.property] = 1'),

    // `Object.fromEntries`
    valid('Object.fromEntries([then, 1])'),
    valid('Object.fromEntries([,,])'),
    valid('Object.fromEntries([[,,],[]])'),
    valid('const NOT_THEN = "not-then";Object.fromEntries([[NOT_THEN, 1]])'),
    valid('Object.fromEntries([[["then", 1]]])'),
    valid('NotObject.fromEntries([["then", 1]])'),
    valid('Object.notFromEntries([["then", 1]])'),
    valid('Object.fromEntries?.([["then", 1]])'),
    valid('Object?.fromEntries([["then", 1]])'),
    valid('Object.fromEntries([[..."then", 1]])'),
    valid('Object.fromEntries([["then", 1]], extraArgument)'),
    valid('Object.fromEntries(...[["then", 1]])'),
    valid('Object.fromEntries([[Symbol.property, 1]])'),

    // `{Object,Reflect}.defineProperty`
    valid('Object.defineProperty(foo, then, 1)'),
    valid('Object.defineProperty(foo, "not-then", 1)'),
    valid('const then = "no-then";Object.defineProperty(foo, then, 1)'),
    valid('Reflect.defineProperty(foo, then, 1)'),
    valid('Reflect.defineProperty(foo, "not-then", 1)'),
    valid('const then = "no-then";Reflect.defineProperty(foo, then, 1)'),
    valid('Object.defineProperty(foo, "then", )'),
    valid('Object.defineProperty(...foo, "then", 1)'),
    valid('Object.defineProperty(foo, ...["then", 1])'),
    valid('Object.defineProperty(foo, Symbol.property, 1)'),
    valid('Reflect.defineProperty(foo, Symbol.property, 1)'),

    // `export`
    valid('export {default} from "then"'),
    valid('const then = 1; export {then as notThen}'),
    valid('export default then'),
    valid('export function notThen(){}'),
    valid('export class notThen {}'),
    valid('export default function then (){}'),
    valid('export default class then {}'),
    valid('export default function (){}'),
    valid('export default class {}'),

    // `export variables`
    valid('export const notThen = 1'),
    valid('export const {then: notThen} = 1'),
    valid('export const {then: notThen = then} = 1'),
  ],
  invalid: [
    // `object`
    invalidObject('const foo = {then: 1}'),
    invalidObject('const foo = {["then"]: 1}'),
    invalidObject('const foo = {[`then`]: 1}'),
    invalidObject('const THEN = "then";const foo = {[THEN]: 1}'),
    invalidObject('const foo = {then() {}}'),
    invalidObject('const foo = {["then"]() {}}'),
    invalidObject('const foo = {[`then`]() {}}'),
    invalidObject('const THEN = "then";const foo = {[THEN]() {}}'),
    invalidObject('const foo = {get then() {}}'),
    invalidObject('const foo = {set then(v) {}}'),
    invalidObject('const foo = {get ["then"]() {}}'),
    invalidObject('const foo = {get [`then`]() {}}'),
    invalidObject('const THEN = "then";const foo = {get [THEN]() {}}'),

    // `class`
    invalidClass('class Foo {then}'),
    invalidClass('const Foo = class {then}'),
    invalidClass('class Foo {["then"]}'),
    invalidClass('class Foo {[`then`]}'),
    invalidClass('const THEN = "then";class Foo {[THEN]}'),
    invalidClass('class Foo {then() {}}'),
    invalidClass('class Foo {["then"]() {}}'),
    invalidClass('class Foo {[`then`]() {}}'),
    invalidClass('const THEN = "then";class Foo {[THEN]() {}}'),
    invalidClass('class Foo {static then}'),
    invalidClass('class Foo {static ["then"]}'),
    invalidClass('class Foo {static [`then`]}'),
    invalidClass('const THEN = "then";class Foo {static [THEN]}'),
    invalidClass('class Foo {static then() {}}'),
    invalidClass('class Foo {static ["then"]() {}}'),
    invalidClass('class Foo {static [`then`]() {}}'),
    invalidClass('const THEN = "then";class Foo {static [THEN]() {}}'),
    invalidClass('class Foo {get then() {}}'),
    invalidClass('class Foo {get ["then"]() {}}'),
    invalidClass('class Foo {get [`then`]() {}}'),
    invalidClass('const THEN = "then";class Foo {get [THEN]() {}}'),
    invalidClass('class Foo {set then(v) {}}'),
    invalidClass('class Foo {set ["then"](v) {}}'),
    invalidClass('class Foo {set [`then`](v) {}}'),
    invalidClass('const THEN = "then";class Foo {set [THEN](v) {}}'),
    invalidClass('class Foo {static get then() {}}'),
    invalidClass('class Foo {static set then(v) {}}'),
    invalidClass('class Foo {static get ["then"]() {}}'),
    invalidClass('class Foo {static get [`then`]() {}}'),
    invalidClass('const THEN = "then";class Foo {static get [THEN]() {}}'),

    // Assign
    invalidObject('foo.then = 1'),
    invalidObject('foo["then"] = 1'),
    invalidObject('foo[`then`] = 1'),
    invalidObject('const THEN = "then";foo[THEN] = 1'),
    invalidObject('foo.then += 1'),
    invalidObject('foo.then ||= 1'),
    invalidObject('foo.then ??= 1'),

    // `{Object,Reflect}.defineProperty`
    invalidObject('Object.defineProperty(foo, "then", 1)'),
    invalidObject('Object.defineProperty(foo, `then`, 1)'),
    invalidObject('const THEN = "then";Object.defineProperty(foo, THEN, 1)'),
    invalidObject('Reflect.defineProperty(foo, "then", 1)'),
    invalidObject('Reflect.defineProperty(foo, `then`, 1)'),
    invalidObject('const THEN = "then";Reflect.defineProperty(foo, THEN, 1)'),

    // `Object.fromEntries`
    invalidObject('Object.fromEntries([["then", 1]])'),
    invalidObject('Object.fromEntries([["then"]])'),
    invalidObject('Object.fromEntries([[`then`, 1]])'),
    invalidObject('const THEN = "then";Object.fromEntries([[THEN, 1]])'),
    invalidObject('Object.fromEntries([foo, ["then", 1]])'),

    // `export`
    invalidExport('const then = 1; export {then}'),
    invalidExport('const notThen = 1; export {notThen as then}'),
    invalidExport('export {then} from "foo"'),
    invalidExport('export function then() {}'),
    invalidExport('export async function then() {}'),
    invalidExport('export function * then() {}'),
    invalidExport('export async function * then() {}'),
    invalidExport('export class then {}'),

    // `export variables`
    invalidExport('export const then = 1'),
    invalidExport('export let then = 1'),
    invalidExport('export var then = 1'),
    invalidExport('export const [then] = 1'),
    invalidExport('export let [then] = 1'),
    invalidExport('export var [then] = 1'),
    invalidExport('export const [, then] = 1'),
    invalidExport('export let [, then] = 1'),
    invalidExport('export var [, then] = 1'),
    invalidExport('export const [, ...then] = 1'),
    invalidExport('export let [, ...then] = 1'),
    invalidExport('export var [, ...then] = 1'),
    invalidExport('export const {then} = 1'),
    invalidExport('export let {then} = 1'),
    invalidExport('export var {then} = 1'),
    invalidExport('export const {foo, ...then} = 1'),
    invalidExport('export let {foo, ...then} = 1'),
    invalidExport('export var {foo, ...then} = 1'),
    invalidExport('export const {foo: {bar: [{baz: then}]}} = 1'),
    invalidExport('export const notThen = 1, then = 1'),
  ],
});
