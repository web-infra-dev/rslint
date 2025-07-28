interface Foo<T> {
  [key: string]: Foo<T> | string;
}
