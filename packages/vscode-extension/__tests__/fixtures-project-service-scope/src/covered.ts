// File IN tsconfig.include (`src`). Type-aware rules should fire.
export const covered = ((command: string, args: string[], options: unknown) => {
  return { stdout: '', stderr: '', exitCode: 0 };
}) as unknown;
