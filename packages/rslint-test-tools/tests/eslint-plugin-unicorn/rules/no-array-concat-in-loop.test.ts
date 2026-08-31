import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const message = 'Do not use `Array#concat()` to accumulate an array in a loop.';
const valid = (code: string, filename = 'file.js') => ({ code, filename });
const invalid = (code: string, filename = 'file.js') => ({
  code,
  filename,
  errors: [{ message }],
});

ruleTester.run('no-array-concat-in-loop', null as never, {
  valid: [
    valid(`let result = [];
result = result.concat(chunk);`),
    valid(`let result = [];
for (const chunk of chunks) {
  result = other.concat(chunk);
}`),
    valid(`let result = [];
for (const chunk of chunks) {
  other = result.concat(chunk);
}`),
    valid(`let result = [initial];
for (const chunk of chunks) {
  result = result.concat(chunk);
}`),
    valid(`let result;
for (const chunk of chunks) {
  result = result.concat(chunk);
}`),
    valid(`const result = [];
for (const chunk of chunks) {
  result = result.concat(chunk);
}`),
    valid(`const text = '';
for (const part of parts) {
  text.concat(part);
}`),
    valid(`let text = '';
for (const part of parts) {
  text = text.concat(part);
}`),
    valid(`let result = [];
for (const chunk of chunks) {
  result = result?.concat(chunk);
}`),
    valid(`let result = [];
for (const chunk of chunks) {
  result = result['concat'](chunk);
}`),
    valid(`let result = [];
for (const chunk of chunks) {
  result = result.concat(chunk).filter(Boolean);
}`),
    valid(`let result = [];
for (const chunk of chunks) {
  result = result.concat();
}`),
    valid(`for (const chunk of chunks) {
  let result = [];
  result = result.concat(chunk);
}`),
    valid(`let result = [];
for (const chunk of chunks) {
  function append() {
    result = result.concat(chunk);
  }
}`),
    valid(`let result = [];
const append = () => {
  result = result.concat(chunk);
};

for (const chunk of chunks) {
  append(chunk);
}`),
    valid(`this.result = [];
for (const chunk of chunks) {
  this.result = this.result.concat(chunk);
}`),
    valid(
      `const result = chunks.reduce((result, chunk) => result.concat(chunk), []);`,
    ),
    valid(
      `let result = ['initial'] as string[];
for (const chunk of chunks) {
  result = (result as string[]).concat(chunk);
}`,
      'file.ts',
    ),
  ],
  invalid: [
    invalid(`let result = [];
for (const chunk of chunks) {
  result = result.concat(chunk);
}`),
    invalid(`let result = [];
for (let index = 0; index < chunks.length; index++) {
  result = result.concat(chunks[index]);
}`),
    invalid(`let result = [];
for (const index in chunks) {
  result = result.concat(chunks[index]);
}`),
    invalid(`let result = [];
while (chunks.length > 0) {
  result = result.concat(chunks.pop());
}`),
    invalid(`let result = [];
do {
  result = result.concat(getChunk());
} while (hasMoreChunks());`),
    invalid(`let result = [];
for (let index = 0; index < chunks.length; result = result.concat(chunks[index++])) {}`),
    invalid(`let result = [];
for (const chunk of chunks) {
  result = (result.concat(chunk));
}`),
    invalid(`var result = [];
for (const chunk of chunks) {
  result = result.concat(chunk);
}`),
    invalid(`let result = [];
for (const chunk of chunks) {
  (result) = (result).concat(chunk);
}`),
    invalid(`let result = [];
for (const chunk of chunks) {
  result = result.concat(first, second);
}`),
    invalid(`let result = [];
for (const chunk of chunks) {
  result = result.concat(...chunkGroups);
}`),
    invalid(
      `let result = [] as string[];
for (const chunk of chunks) {
  result = (result as string[]).concat(chunk);
}`,
      'file.ts',
    ),
    invalid(
      `let result = [] satisfies string[];
for (const chunk of chunks) {
  result = result!.concat(chunk);
}`,
      'file.ts',
    ),
    invalid(
      `let result = <string[]>[];
for (const chunk of chunks) {
  result = (<string[]>result).concat(chunk);
}`,
      'file.ts',
    ),
    invalid(
      `for (let result = []; condition; result = result.concat(chunk)) {}`,
    ),
    invalid(`for (let result = []; condition;) {
  result = result.concat(chunk);
}`),
  ],
});
