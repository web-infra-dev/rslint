// Sibling so the directory is non-empty but lacks an index entrypoint.
// `import X from "./empty-dir"` should fail to resolve (no index).
export const unrelated = 1;
