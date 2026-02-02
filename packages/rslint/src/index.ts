import { NodeRslintService } from './node.js';
import {
  LintOptions,
  LintResponse,
  RSLintService,
  ApplyFixesRequest,
  ApplyFixesResponse,
  GetAstInfoRequest,
  GetAstInfoResponse,
} from './service.js';

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
