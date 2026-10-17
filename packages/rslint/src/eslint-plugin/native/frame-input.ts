/** Worker-local scratch for synchronous frame decoding. No mapped view escapes. */
import {
  checkFrameCancellation,
  decodeFrame,
} from '../../languages/js/frame.js';
import { getNativeBinding, type SharedBytes } from '../../native/binding.js';

const MAX_RETAINED_INPUT = 4 * 1024 * 1024;
let scratch: Buffer | undefined;
let borrowed = false;

export function decodePluginFrame(
  input: Uint8Array | SharedBytes,
  cancelFlag?: Int32Array,
): ReturnType<typeof decodeFrame> {
  checkFrameCancellation(cancelFlag);
  if (input instanceof Uint8Array) return decodeFrame(input, cancelFlag);
  const native = getNativeBinding();
  // Reentrant decoding must not overwrite the outer decoder's byte view.
  // Large inputs keep independent storage instead of raising the cache bound.
  if (borrowed || input.length > MAX_RETAINED_INPUT) {
    return decodeFrame(native.readBytes(input), cancelFlag);
  }
  borrowed = true;
  try {
    if (!scratch || scratch.length < input.length) {
      scratch = Buffer.allocUnsafeSlow(
        Math.min(
          MAX_RETAINED_INPUT,
          Math.max(input.length, (scratch?.length ?? 0) * 2),
        ),
      );
    }
    native.readBytesInto(input, scratch);
    return decodeFrame(scratch.subarray(0, input.length), cancelFlag);
  } finally {
    borrowed = false;
  }
}
