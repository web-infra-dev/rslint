// File NOT in tsconfig.include (test/ outside src/).
// The `console.log` below triggers the non-type-aware `no-console` rule,
// acting as a marker so tests can wait for rslint to finalize diagnostics
// on this file without resorting to a fixed-duration sleep.
console.log('skills.test fixture loaded');

export const uncovered = ((
  command: string,
  args: string[],
  options: unknown,
) => {
  return { stdout: '', stderr: '', exitCode: 0 };
}) as unknown;
