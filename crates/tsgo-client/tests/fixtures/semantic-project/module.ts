const object = { value: 42 };
(globalThis as any).__semanticIssue = object;
export const originalValue = 42;
export interface SymbolLinks {}
