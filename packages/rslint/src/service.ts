import type {
  RslintServiceInterface as RslintServiceBackend,
  LintOptions,
  LintResponse,
  ApplyFixesRequest,
  ApplyFixesResponse,
} from './types.js';
 import { RemoteTypeChecker } from './remote-typechecker.js';

/**
 * Extended lint response that includes the type checker instance
 */
export interface LintResult extends LintResponse {
  /** Type checker instance if includeTypeChecker was true */
  typeChecker?: RemoteTypeChecker;
}

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
  async lint(options: LintOptions = {}): Promise<LintResult> {
    const {
      files,
      config,
      workingDirectory,
      ruleOptions,
      fileContents,
      languageOptions,
      includeEncodedSourceFiles,
      includeTypeChecker,
    } = options;

    // Send handshake
    await this.service.sendMessage('handshake', { version: '1.0.0' });

    // Send lint request
    const response: LintResponse = await this.service.sendMessage('lint', {
      files,
      config,
      workingDirectory,
      ruleOptions,
      fileContents,
      languageOptions,
      includeEncodedSourceFiles,
      includeTypeChecker,
      format: 'jsonline',
    });

    // Create the result with optional type checker
    const result: LintResult = { ...response };

    // If type checker is available, create the RemoteTypeChecker instance
    if (response.hasTypeChecker) {
      result.typeChecker = new RemoteTypeChecker(this.service);
    }

    return result;
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
   * Close the service
   */
  async close(): Promise<void> {
    return new Promise(resolve => {
      this.service.sendMessage('exit', {}).finally(() => {
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
} from './types.js';

// Re-export RemoteTypeChecker and related types
export { RemoteTypeChecker } from './remote-typechecker.js';
export type {
  NodeLocation,
  TypeDetails,
  SymbolDetails,
  SignatureDetails,
  FlowNodeDetails,
  NodeTypeResponse,
  NodeSymbolResponse,
  NodeSignatureResponse,
  NodeFlowNodeResponse,
  NodeInfoResponse,
} from './checker-types.js';
