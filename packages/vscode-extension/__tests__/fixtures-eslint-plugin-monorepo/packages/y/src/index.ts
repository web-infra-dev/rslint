// `bar` must be flagged by plugin Y (py/no-bar). `foo` must NOT
// be flagged here — plugin X lives only in packages/x.
const foo = 1;
const bar = 2;

export { foo, bar };
