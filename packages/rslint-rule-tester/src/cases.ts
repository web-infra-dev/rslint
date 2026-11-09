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
