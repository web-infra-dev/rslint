/** Canonical ESTree child order, shared by decoding and rule traversal. */
import { VISITOR_KEYS_TABLE } from './visitor-keys-table.js';

const _vkObj = Object.create(null) as Record<string, readonly string[]>;
for (const k of Object.keys(VISITOR_KEYS_TABLE)) {
  _vkObj[k] = VISITOR_KEYS_TABLE[k];
}
Object.freeze(_vkObj);
export const VISITOR_KEYS: Readonly<Record<string, readonly string[]>> = _vkObj;

const NON_CHILD_KEYS = new Set([
  'parent',
  'leadingComments',
  'trailingComments',
]);

/** ESLint also permits caller-created synthetic nodes in SourceCode queries.
 * Frame decoding validates known schemas independently of this API fallback. */
export function getVisitorKeys(node: {
  type?: string;
  [key: string]: unknown;
}): readonly string[] {
  const keys = node.type === undefined ? undefined : VISITOR_KEYS[node.type];
  return keys ?? Object.keys(node).filter((key) => !NON_CHILD_KEYS.has(key));
}
