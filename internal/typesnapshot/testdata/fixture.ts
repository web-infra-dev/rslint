const a: undefined = undefined;
const c: string | undefined = undefined;
const d: undefined[] = [];
const e: { a: number } & { b: string } = { a: 1, b: 's' };
function g(): undefined {
  return undefined;
}
function id<T>(x: T): T {
  return x;
}
const f = id(undefined);
const x: string = 's';
