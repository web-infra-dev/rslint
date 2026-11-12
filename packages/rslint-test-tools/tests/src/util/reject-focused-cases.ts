// Every rule-tester runs all cases of a rule inside one `test('valid')` and one
// `test('invalid')`, so the test runner never sees a focused case. Honoring
// `only: true` there would silently skip every other case of the rule while
// the run still passes, so the rule-testers reject focused cases instead,
// matching the Go rule tester.
export function rejectFocusedCases(cases: {
  valid: readonly (string | { only?: boolean })[];
  invalid: readonly { only?: boolean }[];
}): void {
  if (
    cases.valid.some((x) => typeof x === 'object' && x.only) ||
    cases.invalid.some((x) => x.only)
  ) {
    throw new Error(
      "focused rule test cases (only) are not allowed; select tests with the test runner's filters instead",
    );
  }
}
