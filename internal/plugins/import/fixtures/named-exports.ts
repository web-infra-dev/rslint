export const a = 1;
const b = 2;
export const bar = 2;
export const foo = 3;
const c = 3;

export { b };
export { c as d };

export class ExportedClass {}

export * as deep from './deep-namespace-chain/top-level-member';
