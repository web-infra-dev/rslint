// `export = namespace` — under esModuleInterop, the namespace becomes the
// synthesized default. Under no-interop tsconfig, no default is available.
namespace Foo {
  export const a = 1;
  export const b = 2;
}
export = Foo;
