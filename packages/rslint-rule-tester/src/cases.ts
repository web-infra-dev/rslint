import assert from 'node:assert/strict';

/** Rejects expectations that only invalid cases can carry. */
export function checkValidCase(item: object): void {
  for (const key of ['errors', 'output']) {
    assert.ok(
      !Object.hasOwn(item, key),
      `Valid test case must not have '${key}' property`,
    );
  }
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
