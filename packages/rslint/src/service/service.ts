import type {
  RslintServiceInterface as RslintServiceBackend,
  LintOptions,
  LintResponse,
  GetAstInfoRequest,
  GetAstInfoResponse,
  IpcMessage,
  LintInboundHandlers,
} from '../types.js';

/**
 * Environment-agnostic RslintService facade: drives the handshake +
 * lint / getAstInfo / close protocol over a backend (NodeRslintService or
 * BrowserRslintService) supplied by the caller.
 */
export class RSLintService {
  private readonly service: RslintServiceBackend;
  private activeLintHandlers: LintInboundHandlers | null = null;
  private requestQueue: Promise<void> = Promise.resolve();
  private closeRequested = false;
  private closePromise: Promise<void> | null = null;

  constructor(service: RslintServiceBackend) {
    this.service = service;
    this.service.setInboundHandler?.((message) =>
      this.handleInboundRequest(message),
    );
  }

  /**
   * Run the linter on specified files
   */
  async lint(
    options: LintOptions = {},
    handlers: LintInboundHandlers = {},
  ): Promise<LintResponse> {
    if (this.closeRequested) {
      throw new Error('rslint service is closing');
    }

    return this.enqueue(async () => {
      if (handlers.pluginLint && !this.service.setInboundHandler) {
        throw new Error(
          'rslint backend does not support reverse pluginLint requests',
        );
      }

      this.activeLintHandlers = handlers;
      try {
        return await this.lintExclusive(options);
      } finally {
        this.activeLintHandlers = null;
      }
    });
  }

  private async lintExclusive(options: LintOptions): Promise<LintResponse> {
    const {
      files,
      config,
      eslintPlugins,
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
      eslintPlugins,
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
    if (this.closeRequested) {
      throw new Error('rslint service is closing');
    }

    return this.enqueue(() => this.getAstInfoExclusive(options));
  }

  private async getAstInfoExclusive(
    options: GetAstInfoRequest,
  ): Promise<GetAstInfoResponse> {
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
  close(): Promise<void> {
    if (this.closePromise) return this.closePromise;
    this.closeRequested = true;

    this.closePromise = this.enqueue(async () => {
      // Ask the peer to exit, then tear down regardless. The peer may exit
      // before its ack frame is read; swallow rejection on that expected path.
      try {
        await this.service.sendMessage('exit', {});
      } catch {
        // peer already gone — expected during close
      }
      this.service.terminate();
    });
    return this.closePromise;
  }

  private enqueue<T>(operation: () => Promise<T>): Promise<T> {
    const result = this.requestQueue.then(operation, operation);
    this.requestQueue = result.then(
      () => undefined,
      () => undefined,
    );
    return result;
  }

  private handleInboundRequest(
    message: IpcMessage,
  ): Promise<unknown> | unknown {
    if (message.kind !== 'pluginLint') {
      throw new Error(
        `rslint service received unexpected inbound request '${message.kind}'`,
      );
    }
    const handler = this.activeLintHandlers?.pluginLint;
    if (!handler) {
      throw new Error(
        'rslint service received pluginLint without an active plugin host',
      );
    }
    return handler(message.data);
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
  InboundRequestHandler,
  LintInboundHandlers,
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
