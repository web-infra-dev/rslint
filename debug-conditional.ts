interface Foo<T> {
  [s: string]: Foo extends T ? string : number;
}
