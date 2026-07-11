import type {
  RslintServiceInterface as RslintServiceBackend,
  LintOptions,
  LintResponse,
  GetAstInfoRequest,
  GetAstInfoResponse,
  IpcMessage,
  LintInboundHandlers,
} from '../types.js';
import {
  API_PROTOCOL_VERSION,
  API_REVERSE_PLUGIN_LINT_CAPABILITY,
} from './protocol.js';

const EXIT_REQUEST_TIMEOUT_MS = 1_000;

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
        return await this.lintExclusive(options, Boolean(handlers.pluginLint));
      } finally {
        this.activeLintHandlers = null;
      }
    });
  }

  private async lintExclusive(
    options: LintOptions,
    requiresReversePluginLint: boolean,
  ): Promise<LintResponse> {
    const {
      files,
      canonicalFiles,
      config,
      eslintPlugins,
      configDirectory,
      pluginConfigDirectory,
      workingDirectory,
      fileContents,
      includeEncodedSourceFiles,
      fix,
    } = options;

    await this.handshake(requiresReversePluginLint);

    // Send lint request
    return this.service.sendMessage('lint', {
      files,
      canonicalFiles,
      config,
      eslintPlugins,
      configDirectory,
      pluginConfigDirectory,
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

    return this.enqueue(async () => {
      const response = await this.getAstInfoExclusive(options);
      return response;
    });
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

    await this.handshake(false);

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
    if (this.closePromise) {
      await this.closePromise;
      return;
    }
    this.closeRequested = true;

    let timedOut = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    const timeout = new Promise<void>((resolve) => {
      timer = setTimeout(() => {
        timedOut = true;
        resolve();
      }, EXIT_REQUEST_TIMEOUT_MS);
    });
    const gracefulExit = this.enqueue(async () => {
      // Never interleave exit with an active request. If waiting for the queue
      // consumed the shutdown budget, terminate without sending a late exit.
      if (timedOut) return;
      try {
        await this.service.sendMessage('exit', {});
      } catch {
        // peer already gone — expected during close
      }
    });

    this.closePromise = (async () => {
      try {
        await Promise.race([gracefulExit, timeout]);
      } finally {
        timedOut = true;
        if (timer) clearTimeout(timer);
        this.service.terminate();
      }
    })();
    await this.closePromise;
  }

  private async handshake(requiresReversePluginLint: boolean): Promise<void> {
    const requestedCapabilities = requiresReversePluginLint
      ? [API_REVERSE_PLUGIN_LINT_CAPABILITY]
      : [];
    const response: unknown = await this.service.sendMessage('handshake', {
      version: API_PROTOCOL_VERSION,
      capabilities: requestedCapabilities,
    });
    if (response === null || typeof response !== 'object') {
      throw new Error('rslint backend returned an invalid handshake response');
    }
    const handshake = response as {
      version?: unknown;
      ok?: unknown;
      capabilities?: unknown;
    };
    if (handshake.ok !== true || handshake.version !== API_PROTOCOL_VERSION) {
      throw new Error(
        `rslint API protocol mismatch: expected ${API_PROTOCOL_VERSION}, received ${String(handshake.version)}`,
      );
    }
    if (requiresReversePluginLint) {
      const capabilities = Array.isArray(handshake.capabilities)
        ? handshake.capabilities
        : [];
      if (!capabilities.includes(API_REVERSE_PLUGIN_LINT_CAPABILITY)) {
        throw new Error(
          'rslint backend does not support reverse pluginLint requests',
        );
      }
    }
  }

  private async enqueue<T>(operation: () => Promise<T>): Promise<T> {
    const result = this.requestQueue.then(operation, operation);
    this.requestQueue = result.then(
      () => undefined,
      () => undefined,
    );
    const value = await result;
    return value;
  }

  private handleInboundRequest(message: IpcMessage): unknown {
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
