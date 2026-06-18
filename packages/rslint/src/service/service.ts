import type {
  RslintServiceInterface as RslintServiceBackend,
  LintOptions,
  LintResponse,
  GetAstInfoRequest,
  GetAstInfoResponse,
} from '../types.js';

/**
 * Environment-agnostic RslintService facade: drives the handshake +
 * lint / getAstInfo / close protocol over a backend (NodeRslintService or
 * BrowserRslintService) supplied by the caller.
 */
export class RSLintService {
  private readonly service: RslintServiceBackend;

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
      configDirectory,
      workingDirectory,
      fileContents,
      includeEncodedSourceFiles,
      fix,
    } = options;

    // Send handshake
    await this.service.sendMessage('handshake', { version: '1.0.0' });

    // Send lint request
    return this.service.sendMessage('lint', {
      files,
      config,
      configDirectory,
      workingDirectory,
      fileContents,
      includeEncodedSourceFiles,
      fix,
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
    // Ask the peer to exit, then tear down regardless. The peer may exit before
    // its ack frame is read; the exit handler resolves the pending on that
    // expected path, but swallow any rejection too (defensive — e.g. the
    // process was already dead) so a floating promise can't become an
    // unhandledRejection.
    try {
      await this.service.sendMessage('exit', {});
    } catch {
      // peer already gone — expected during close
    }
    this.service.terminate();
  }
}

// Re-export types for convenience
export type {
  Diagnostic,
  LintOptions,
  LintResponse,
  RSlintOptions,
  RslintServiceInterface,
  PendingMessage,
  IpcMessage,
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
} from '../types.js';
