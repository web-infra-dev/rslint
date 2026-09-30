// The part of @types/jest the upstream fixture program relies on.
declare const expect: jest.Expect;
declare namespace jest {
  interface Expect {
    <T = any>(actual: T): JestMatchers<T>;
  }
  type JestMatchers<T> = Matchers<void, T>;
  interface Matchers<R, T = {}> {
    toHaveBeenCalledTimes(expected: number): R;
  }
}
