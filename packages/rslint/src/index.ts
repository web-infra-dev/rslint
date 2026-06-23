import { NodeRslintService } from './node.js';
import {
  LintOptions,
  LintTextOptions,
  LintResponse,
  RSLintService,
  ApplyFixesRequest,
  ApplyFixesResponse,
  GetAstInfoRequest,
  GetAstInfoResponse,
} from './service.js';

export { defineConfig, globalIgnores } from './define-config.js';
export type { RslintConfigEntry, ESLintPlugin } from './define-config.js';
export {
  ts,
  js,
  reactPlugin,
  reactHooksPlugin,
  importPlugin,
  promisePlugin,
  jestPlugin,
  unicornPlugin,
  jsxA11yPlugin,
} from './configs/index.js';

// Export the main RSLintService class for direct usage
export { RSLintService } from './service.js';

// Export specific implementations for advanced usage
export { NodeRslintService } from './node.js';

// For backward compatibility and convenience
export async function lint(options: LintOptions): Promise<LintResponse> {
  const service = new RSLintService(
    new NodeRslintService({
      workingDirectory: options.workingDirectory,
    }),
  );
  const result = await service.lint(options);
  await service.close();
  return result;
}

// Convenience function for in-memory linting (no disk write), mirroring ESLint's lintText
export async function lintText(
  code: string,
  options: LintTextOptions = {},
): Promise<LintResponse> {
  const service = new RSLintService(
    new NodeRslintService({
      workingDirectory: options.workingDirectory,
    }),
  );
  let result: LintResponse;
  try {
    result = await service.lintText(code, options);
  } finally {
    await service.close();
  }
  return result;
}

// Convenience function for applying fixes
export async function applyFixes(
  options: ApplyFixesRequest,
): Promise<ApplyFixesResponse> {
  const service = new RSLintService(new NodeRslintService());
  const result = await service.applyFixes(options);
  await service.close();
  return result;
}

// Convenience function for getting AST info
export async function getAstInfo(
  options: GetAstInfoRequest,
): Promise<GetAstInfoResponse> {
  const service = new RSLintService(new NodeRslintService());
  const result = await service.getAstInfo(options);
  await service.close();
  return result;
}

// Export all types
export {
  type Diagnostic,
  type LintOptions,
  type LintTextOptions,
  type LintResponse,
  type ApplyFixesRequest,
  type ApplyFixesResponse,
  type LanguageOptions,
  type ParserOptions,
  type RSlintOptions,
  type RslintServiceInterface,
  // AST Info types
  type GetAstInfoRequest,
  type GetAstInfoResponse,
  type NodeInfo,
  type TypeInfo,
  type SymbolInfo,
  type SignatureInfo,
  type FlowInfo,
  type ParameterInfo,
  type TypeParamInfo,
  type IndexInfo,
  type TypePredicateInfo,
} from './types.js';
