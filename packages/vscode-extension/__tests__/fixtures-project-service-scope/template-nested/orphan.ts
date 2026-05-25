// This file lives under a config directory that has NO tsconfig.json.
// The nearest rslint config says `projectService: true` + no explicit
// `project`, so rslint can't resolve a tsconfig for it.
//
// CLI and LSP must agree on whether `@typescript-eslint/no-unused-vars`
// (a type-aware rule) runs here.
export const orphan = ((command: string, args: string[], options: unknown) => {
  return { stdout: '', stderr: '', exitCode: 0 };
}) as unknown;
