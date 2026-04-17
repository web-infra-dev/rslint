// File NOT in tsconfig.include (test/ outside src/).
export const uncovered = ((
  command: string,
  args: string[],
  options: unknown,
) => {
  return { stdout: '', stderr: '', exitCode: 0 };
}) as unknown;
