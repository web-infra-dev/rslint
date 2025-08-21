import {
  LintOptions,
  LintResponse,
  RSLintService,
  ApplyFixesRequest,
  ApplyFixesResponse,
} from './service.ts';

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

// Convenience function for applying fixes
export async function applyFixes(
  options: ApplyFixesRequest,
): Promise<ApplyFixesResponse> {
  const service = new RSLintService({});
  const result = await service.applyFixes(options);
  await service.close();
  return result;
}

export {
  type Diagnostic,
  type LintOptions,
  type LintResponse,
  type ApplyFixesRequest,
  type ApplyFixesResponse,
  type LanguageOptions,
  type ParserOptions,
} from './service.ts';
