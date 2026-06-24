declare function maybeString(): string | undefined;
declare function eitherNum(): string | number;

// Union that INCLUDES undefined → reported.
const withUndefined = maybeString();

// Union with NO undefined member → not reported (exercises the member check,
// not just isUnion()).
const withoutUndefined = eitherNum();

// Not a union at all → not reported (exercises isUnion() === false).
const plain = 'hello';

// Reference the bindings so `noUnusedLocals`-style setups don't elide them and
// so tsgo definitely assigns each a type.
export { withUndefined, withoutUndefined, plain };
