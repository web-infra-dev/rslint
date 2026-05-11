// `foo` must be flagged by plugin X (px/no-foo). `bar` must NOT
// be flagged here — plugin Y lives only in packages/y, and the
// per-config LoadedPlugins map keeps the two disjoint.
const foo = 1;
const bar = 2;

export { foo, bar };
