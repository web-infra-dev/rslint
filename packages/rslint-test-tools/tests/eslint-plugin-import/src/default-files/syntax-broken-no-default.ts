// Intentional syntax errors — no default export. Rule must not crash; should
// either skip silently (module symbol unresolvable) or report cleanly.
let = 1;
export const named = 1;
