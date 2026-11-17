// Minimal Jest matcher types needed by the upstream unbound-method case.
// A method declaration is intentional: a function-valued property would not
// exercise the same unbound-method check.
interface JestMatchers {
  toHaveBeenCalledTimes(count: number): void;
}

declare function expect(value: unknown): JestMatchers;
