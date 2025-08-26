import type {
  RslintServiceInterface,
  LintOptions,
  LintResponse,
  ApplyFixesRequest,
  ApplyFixesResponse,
  RSlintOptions,
} from './types.js';

// Import implementations
import { NodeRslintService } from './node.js';
import { BrowserRslintService } from './browser.js';

/**
 * Environment detection
 */
function isNode(): boolean {
  return (
    typeof process !== 'undefined' &&
    process.versions != null &&
    process.versions.node != null
  );
}

function isBrowser(): boolean {
  return (
    typeof globalThis !== 'undefined' &&
    typeof (globalThis as unknown as { Worker?: unknown }).Worker !==
      'undefined' &&
    typeof (globalThis as unknown as { window?: unknown }).window !==
      'undefined'
  );
}

/**
 * Factory function to create the appropriate RslintService implementation
 */
export function createRslintService(
  options: RSlintOptions = {},
): RslintServiceInterface {
  if (isNode()) {
    return new NodeRslintService(options);
  } else if (isBrowser()) {
    return new BrowserRslintService(options);
  } else {
    throw new Error(
      'Unsupported environment. RslintService requires Node.js or a browser with Web Worker support.',
    );
  }
}

/**
 * Main RslintService class that automatically uses the appropriate implementation
 */
export class RSLintService {
  private service: RslintServiceInterface;

  constructor(options: RSlintOptions = {}) {
    this.service = createRslintService(options);
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
