/** Canonical ESTree child order, shared by decoding and rule traversal. */
import { visitorKeys as canonicalVisitorKeys } from '@typescript-eslint/visitor-keys';
import { getKeys } from 'eslint-visitor-keys';

const _vkObj = Object.create(null) as Record<string, readonly string[]>;
for (const [type, keys] of Object.entries(canonicalVisitorKeys)) {
  if (keys !== undefined) _vkObj[type] = keys;
}
Object.freeze(_vkObj);
export const VISITOR_KEYS: Readonly<Record<string, readonly string[]>> = _vkObj;

/** ESLint also permits caller-created synthetic nodes in SourceCode queries.
 * Frame decoding uses the fixed schema table independently of this fallback. */
export function getVisitorKeys(node: {
  type?: string;
  [key: string]: unknown;
}): readonly string[] {
  const keys = node.type === undefined ? undefined : VISITOR_KEYS[node.type];
  return keys ?? getKeys(node);
}
