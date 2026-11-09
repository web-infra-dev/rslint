import assert from 'node:assert/strict';

/** Requires a skipped case to state why. */
export function checkSkip(item: { skip?: unknown }): void {
  if (!Object.hasOwn(item, 'skip')) return;
  assert.ok(
    typeof item.skip === 'string' && item.skip.trim() !== '',
    "Test case property 'skip' must be a non-empty reason",
  );
}

/** Rejects expectations that only invalid cases can carry. */
export function checkValidCase(item: object): void {
  for (const key of ['errors', 'output']) {
    assert.ok(
      !Object.hasOwn(item, key),
      `Valid test case must not have '${key}' property`,
    );
  }
}

/**
 * Requires an invalid case to expect at least one error, and an expected
 * autofix to change the code.
 */
export function checkInvalidCase(item: {
  code: string;
  errors?: unknown;
  output?: string;
}): void {
  assert.ok(
    Array.isArray(item.errors),
    "Invalid test case must have an 'errors' array",
  );
  assert.ok(
    item.errors.length > 0,
    'Invalid test case must have at least one error',
  );
  assert.notEqual(
    item.output,
    item.code,
    "Test property 'output' matches 'code'. Omit 'output' if no autofix is expected.",
  );
}

const requiredErrorFields = {
  messageId: ['string'],
  message: ['string', 'function'],
  line: ['number'],
  column: ['number'],
  endLine: ['number'],
  endColumn: ['number'],
} as const;

/** Requires an expected error to state its message and full range. */
export function checkExpectedError(error: unknown, index: number): void {
  assert.ok(
    typeof error === 'object' && error !== null,
    `errors[${index}] must be an object`,
  );
  for (const [key, types] of Object.entries(requiredErrorFields)) {
    const type: string = typeof Reflect.get(error, key);
    assert.ok(
      (types as readonly string[]).includes(type),
      `errors[${index}].${key} must be ${types.join(' or ')}, got ${type}`,
    );
  }
}
