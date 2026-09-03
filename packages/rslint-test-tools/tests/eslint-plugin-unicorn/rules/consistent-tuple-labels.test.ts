import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const messageId = 'consistent-tuple-labels';
const message =
  'This tuple element should have a label, just like the other elements.';

const valid = (code: string) => ({ code, filename: 'file.ts' });
const invalid = (code: string, errorCount = 1) => ({
  code,
  filename: 'file.ts',
  errors: Array.from({ length: errorCount }, () => ({ messageId, message })),
});

ruleTester.run('consistent-tuple-labels', null as never, {
  valid: [
    valid('type Foo = [a: string, b: number];'),
    valid('type Foo = [a: string, b: number, c: boolean];'),
    valid('type Foo = [string, number];'),
    valid('type Foo = [string, number, boolean];'),
    valid('type Foo = [];'),
    valid('type Foo = [string];'),
    valid('type Foo = [a: string];'),
    valid('type Foo = [a?: string, b?: number];'),
    valid('type Foo = [a: string, b?: number];'),
    valid('type Foo = [string?, number?];'),
    valid('type Foo = [a: string, ...b: number[]];'),
    valid('type Foo = [string, ...number[]];'),
    valid('type Foo = [[a: number, b: number]];'),
    valid('type Foo = [a: [x: number], b: [y: number]];'),
    valid('type Foo = string[];'),
    valid('type Foo = readonly string[];'),
    valid('type Foo = {a: string; b: number};'),
  ],
  invalid: [
    invalid('type Foo = [a: string, number];'),
    invalid('type Foo = [string, b: number];'),
    invalid('type Foo = [a: string, b: number, c: boolean, d];'),
    invalid('type Foo = [a: string, number, c: boolean];'),
    invalid('type Foo = [a: string, number, boolean];', 2),
    invalid('type Foo = [a?: string, number?];'),
    invalid('type Foo = [string?, b: number];'),
    invalid('type Foo = [string, ...rest: number[]];'),
    invalid('type Foo = [a: string, ...number[]];'),
    invalid('type Foo = [a?: string, number];'),
    invalid('type Foo = readonly [a: string, number];'),
    invalid('type Foo = [[a: number, number]];'),
    invalid('type Foo = [a: string, /* unlabeled */ number];'),
  ],
});
