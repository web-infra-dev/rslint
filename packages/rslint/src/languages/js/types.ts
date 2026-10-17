import type { SourceLocation } from './positions.js';

/** Ordinary ESTree node shape shared by decoding, traversal and plugin consumers. */
export interface ESTreeNode {
  type: string;
  range: [number, number];
  loc: SourceLocation;
  parent?: ESTreeNode;
  [key: string]: unknown;
}

/** JavaScript global access values accepted by ESLint-compatible language options. */
export type GlobalAccess =
  | boolean
  | null
  | 'true'
  | 'false'
  | 'readonly'
  | 'readable'
  | 'writable'
  | 'writeable'
  | 'off';

export type GlobalsConfig = Record<string, GlobalAccess>;
