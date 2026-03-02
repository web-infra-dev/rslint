import type {
  RslintServiceInterface as RslintServiceBackend,
  LintOptions,
  LintResponse,
  ApplyFixesRequest,
  ApplyFixesResponse,
  GetAstInfoRequest,
  GetAstInfoResponse,
} from './types.js';

/**
 * Main RslintService class that automatically uses the appropriate implementation
 */
export class RSLintService {
  private service: RslintServiceBackend;

  constructor(service: RslintServiceBackend) {
    this.service = service;
  }

  /**
   * Run the linter on specified files
   */
  async lint(options: LintOptions = {}): Promise<LintResponse> {
    const {
      files,
      config,
      workingDirectory,
      ruleOptions,
      fileContents,
      languageOptions,
      includeEncodedSourceFiles,
    } = options;

    // Send handshake
    await this.service.sendMessage('handshake', { version: '1.0.0' });

    // Send lint request
    return this.service.sendMessage('lint', {
      files,
      config,
      workingDirectory,
      ruleOptions,
      fileContents,
      languageOptions,
      includeEncodedSourceFiles,
      format: 'jsonline',
    });
  }

  /**
   * Apply fixes to a file based on diagnostics
   */
  async applyFixes(options: ApplyFixesRequest): Promise<ApplyFixesResponse> {
    const { fileContent, diagnostics } = options;

    // Send handshake
    await this.service.sendMessage('handshake', { version: '1.0.0' });

    // Send apply fixes request
    return this.service.sendMessage('applyFixes', {
      fileContent,
      diagnostics,
    });
  }

  /**
   * Get detailed AST information at a specific position
   * Returns Node, Type, Symbol, Signature, and Flow information
   */
  async getAstInfo(options: GetAstInfoRequest): Promise<GetAstInfoResponse> {
    const {
      fileContent,
      position,
      end,
      kind,
      depth = 2,
      fileName,
      compilerOptions,
    } = options;

    // Send handshake
    await this.service.sendMessage('handshake', { version: '1.0.0' });

    // Send getAstInfo request
    return this.service.sendMessage('getAstInfo', {
      fileContent,
      position,
      end,
      kind,
      depth,
      fileName,
      compilerOptions,
    });
  }

  /**
   * Close the service
   */
  async close(): Promise<void> {
    return new Promise(resolve => {
      void this.service.sendMessage('exit', {}).finally(() => {
        this.service.terminate();
        resolve();
      });
    });
  }
}

// Re-export types for convenience
export type {
  Diagnostic,
  LintOptions,
  LintResponse,
  ApplyFixesRequest,
  ApplyFixesResponse,
  LanguageOptions,
  ParserOptions,
  RSlintOptions,
  RslintServiceInterface,
  // AST Info types
  GetAstInfoRequest,
  GetAstInfoResponse,
  NodeInfo,
  TypeInfo,
  SymbolInfo,
  SignatureInfo,
  FlowInfo,
  ParameterInfo,
  TypeParamInfo,
  IndexInfo,
  TypePredicateInfo,
} from './types.js';
