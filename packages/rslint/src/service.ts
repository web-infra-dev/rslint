import type {
  RslintServiceInterface as RslintServiceBackend,
  LintOptions,
  LintResponse,
  ApplyFixesRequest,
  ApplyFixesResponse,
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
