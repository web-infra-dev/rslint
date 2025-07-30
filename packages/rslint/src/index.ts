import { LintOptions, LintResponse, RSLintService } from './service.ts';

// Export the RSLintService class for direct usage
export { RSLintService } from './service.ts';

// For backward compatibility and convenience
export async function lint(options: LintOptions): Promise<LintResponse> {
  const service = new RSLintService({
    workingDirectory: options.workingDirectory,
  });
  const result = await service.lint(options);
  await service.close();
  return result;
}

export { type Diagnostic, type LintOptions, type LintResponse } from './service.ts';
