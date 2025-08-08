type Foo = {
  bar: boolean;
  [key: `key_${string}`]: number;
};
declare const foo: Foo;
console.log('Testing key_baz access:');
foo['key_baz']; // This should be allowed (not flagged)
