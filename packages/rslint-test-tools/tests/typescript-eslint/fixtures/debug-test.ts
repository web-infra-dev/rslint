class Foo {
  // This should be out of order - method before field
  doSomething() {}
  name: string = '';
}