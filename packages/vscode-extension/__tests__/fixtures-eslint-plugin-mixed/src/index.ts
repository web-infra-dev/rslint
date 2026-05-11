// In-tsconfig file. `debugger` fires the native `no-debugger` rule;
// `forbidden` fires the plugin `fx/no-forbidden` rule. Both must land
// on this SAME file — the mixed native+plugin path through the LSP.
debugger;
const forbidden = 1;
export { forbidden };
