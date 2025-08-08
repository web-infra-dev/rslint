type Foo = {
  bar: boolean;
  [key: `key_${string}`]: number;
};
declare const foo: Foo;
foo['key_baz'];
