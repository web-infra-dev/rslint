// Fixtures producing real node2type (tokenStart,end) span collisions, to verify
// Build resolves each by keeping the DEEPER node (the typescript-eslint anchor)
// instead of emitting two entries the worker would pick between arbitrarily.
// Every collision below is a strict parent/child sharing one span.

import * as mod from './mod';

export const m = new Map<string, number>();
export const arr: Array<{ x: number }> = [];

// namespace import: the ImportClause (→`any`) shares its span with the inner
// NamespaceImport (→`typeof import(...)`). The deeper NamespaceImport must win.
export const useMod = mod.val;

export function run(): void {
  // for-of array destructuring (NO initializer): the VariableDeclaration shares
  // its span with the inner ArrayBindingPattern. VariableDeclaration→`any`,
  // ArrayBindingPattern→the tuple [string, number]. The deeper one must win.
  for (const [k, v] of m) {
    void k;
    void v;
  }
  // for-of object destructuring: VariableDeclaration vs ObjectBindingPattern.
  for (const { x } of arr) {
    void x;
  }
}

class Base {
  declare b: number;
}

// `extends Base` with NO type args: tsgo wraps the Identifier `Base` in an
// ExpressionWithTypeArguments at the SAME span (both end at `Base`). With type
// args (`extends Base<number>`) the EWTA span includes `<number>` and would NOT
// collide — the bare form is what collides. EWTA→the instance type Base,
// Identifier→`typeof Base`; typescript-eslint's superClass anchors on the inner
// Identifier (convert.js maps superClass to types[0].expression), so the deeper
// Identifier (`typeof Base`) must win.
export class Derived extends Base {
  declare y: number;
}
