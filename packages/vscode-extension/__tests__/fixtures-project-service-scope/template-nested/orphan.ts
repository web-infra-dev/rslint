// This file lives under a config directory that has NO tsconfig.json.
// The nearest rslint config says `projectService: true` + no explicit
// `project`, so rslint can't resolve a tsconfig for it.
//
// The LSP must not run `@typescript-eslint/no-unused-vars` (a type-aware rule)
// here, but non-type-aware native rules and their fixes must remain active.
export const orphan = ((command: string, args: string[], options: unknown) => {
  var output = command;
  return { stdout: output, stderr: '', exitCode: 0 };
}) as unknown;
