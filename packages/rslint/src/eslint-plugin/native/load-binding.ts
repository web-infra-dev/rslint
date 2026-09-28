/** Worker parser entry; the CLI uses the same loader without loading this runtime. */
import { getNativeBinding } from '../../native/binding.js';

export type {
  CommentObj,
  ParseResult,
  SharedBytes,
} from '../../native/binding.js';

const binding = getNativeBinding();
export const parse = binding.parse;
export const parseSharedBytes = binding.parseSharedBytes;
