import {
  type ActivateConfigsRequest,
  type ActivateConfigsResponse,
  type ConfigModuleActivationPlan,
  type ConfigModuleEslintPluginEntry,
  type ConfigModulePluginDescriptor,
  type EvaluateConfigPredicatesRequest,
  type EvaluateConfigPredicatesResponse,
  type LoadConfigsRequest,
  type LoadConfigsResponse,
} from '@rslint/core/config-loader';

interface ConfigActivationWireResponse {
  transactionId: string;
  /** Empty when no matching worker generation could be staged. */
  eslintPluginEntries: ConfigModuleEslintPluginEntry[];
  /** False lets Go preserve its last-good catalog instead of committing. */
  pluginHostReady: boolean;
}

export interface ConfigTransactionControlRequest {
  transactionId: string;
}

interface ConfigCommitWireResponse {
  transactionId: string;
  committed: true;
}

interface ConfigAbortWireResponse {
  transactionId: string;
  aborted: true;
}

/** Structural seams keep the JSON-RPC transaction adapter independently testable. */
interface ConfigModuleHostAdapter {
  loadConfigs(
    request: LoadConfigsRequest,
    signal?: AbortSignal,
  ): Promise<LoadConfigsResponse>;
  activateConfigs(
    request: ActivateConfigsRequest,
    signal?: AbortSignal,
    prepare?: (plan: ConfigModuleActivationPlan) => Promise<void>,
  ): Promise<ActivateConfigsResponse>;
  evaluateConfigPredicates(
    request: EvaluateConfigPredicatesRequest,
    signal?: AbortSignal,
  ): Promise<EvaluateConfigPredicatesResponse>;
  deleteSession(transactionId: string): boolean;
}

interface PluginLintPoolAdapter {
  prepare(
    descriptors: ConfigModulePluginDescriptor[],
    fingerprint: string,
    generation: string,
  ): Promise<boolean>;
  commit(generation: string): Promise<boolean>;
  abort(generation: string): Promise<void>;
}

function throwIfAborted(signal?: AbortSignal): void {
  if (!signal?.aborted) return;
  if (signal.reason instanceof Error) throw signal.reason;
  throw new Error('config transaction was cancelled');
}

function assertTransactionControlRequest(
  request: ConfigTransactionControlRequest,
): void {
  if (!request || typeof request !== 'object') {
    throw new Error('config transaction request must be an object');
  }
  if (
    typeof request.transactionId !== 'string' ||
    request.transactionId.length === 0
  ) {
    throw new Error('config transactionId must be a non-empty string');
  }
}

/**
 * LSP transport adapter for the shared config module host.
 *
 * Go owns discovery, ignore semantics, last-good selection and catalog commit.
 * This adapter only evaluates Go's candidates, stages the matching plugin host,
 * and mirrors Go's final commit/abort for the same transaction ID.
 */
export class LspConfigTransactionAdapter {
  private readonly stagedTransactions = new Set<string>();
  private activeTransaction: string | undefined;
  private predecessorTransaction: string | undefined;
  private disposed = false;

  constructor(
    private readonly host: ConfigModuleHostAdapter,
    private readonly pluginLintPool: PluginLintPoolAdapter,
    private readonly fingerprint: (plan: ConfigModuleActivationPlan) => string,
  ) {}

  async loadConfigs(
    request: LoadConfigsRequest,
    signal?: AbortSignal,
  ): Promise<LoadConfigsResponse> {
    this.assertActive();
    assertTransactionControlRequest(request);
    throwIfAborted(signal);
    const transactionId = request.transactionId;
    this.stagedTransactions.add(transactionId);
    try {
      // Editor reloads must not reuse the config entry module. Go still sends
      // the shared envelope, but the LSP transport makes that entry-freshness
      // invariant explicit for every frontier. Static transitive imports retain
      // Node's normal module-cache semantics; full graph isolation requires a
      // separate evaluator realm rather than query-busting only the entry URL.
      const response = await this.host.loadConfigs(
        { ...request, loadMode: 'fresh' },
        signal,
      );
      this.assertActive();
      throwIfAborted(signal);
      return response;
    } catch (error) {
      this.discardStaged(transactionId);
      throw error;
    }
  }

  async activateConfigs(
    request: ActivateConfigsRequest,
    signal?: AbortSignal,
  ): Promise<ConfigActivationWireResponse> {
    this.assertActive();
    assertTransactionControlRequest(request);
    throwIfAborted(signal);
    const transactionId = request.transactionId;
    try {
      let pluginHostReady = false;
      const activation = await this.host.activateConfigs(
        request,
        signal,
        async (candidate) => {
          this.assertActive();
          throwIfAborted(signal);
          pluginHostReady = await this.pluginLintPool.prepare(
            candidate.pluginConfigs,
            this.fingerprint(candidate),
            transactionId,
          );
          this.assertActive();
          throwIfAborted(signal);
        },
      );
      this.assertActive();
      throwIfAborted(signal);
      return {
        transactionId: activation.transactionId,
        // Never ask Go to register/dispatch placeholder rules without the
        // matching worker generation. On first startup Go may still commit the
        // ordinary native config as a degraded no-host generation; with a
        // last-good generation it instead aborts this transaction.
        eslintPluginEntries: pluginHostReady
          ? activation.eslintPluginEntries
          : [],
        pluginHostReady,
      };
    } catch (error) {
      await this.pluginLintPool.abort(transactionId).catch(() => undefined);
      this.discardStaged(transactionId);
      throw error;
    }
  }

  async evaluateConfigPredicates(
    request: EvaluateConfigPredicatesRequest,
    signal?: AbortSignal,
  ): Promise<EvaluateConfigPredicatesResponse> {
    this.assertActive();
    assertTransactionControlRequest(request);
    throwIfAborted(signal);
    const response = await this.host.evaluateConfigPredicates(request, signal);
    this.assertActive();
    throwIfAborted(signal);
    return response;
  }

  async commitConfigs(
    request: ConfigTransactionControlRequest,
  ): Promise<ConfigCommitWireResponse> {
    this.assertActive();
    assertTransactionControlRequest(request);
    const transactionId = request.transactionId;
    if (
      transactionId !== this.activeTransaction &&
      !this.stagedTransactions.has(transactionId)
    ) {
      throw new Error(
        `cannot commit unknown config transaction ${JSON.stringify(transactionId)}`,
      );
    }
    if (!(await this.pluginLintPool.commit(transactionId))) {
      throw new Error(
        `failed to commit plugin-host generation ${JSON.stringify(transactionId)}`,
      );
    }
    if (transactionId !== this.activeTransaction) {
      if (this.predecessorTransaction) {
        this.host.deleteSession(this.predecessorTransaction);
      }
      this.predecessorTransaction = this.activeTransaction;
      this.activeTransaction = transactionId;
      this.stagedTransactions.delete(transactionId);
    }
    return {
      transactionId,
      committed: true,
    };
  }

  async abortConfigs(
    request: ConfigTransactionControlRequest,
  ): Promise<ConfigAbortWireResponse> {
    assertTransactionControlRequest(request);
    const transactionId = request.transactionId;
    try {
      await this.pluginLintPool.abort(transactionId);
    } finally {
      if (transactionId === this.activeTransaction) {
        this.host.deleteSession(transactionId);
        this.activeTransaction = this.predecessorTransaction;
        this.predecessorTransaction = undefined;
      } else if (this.stagedTransactions.has(transactionId)) {
        this.discardStaged(transactionId);
      }
    }
    return {
      transactionId,
      aborted: true,
    };
  }

  /**
   * Drop transactions orphaned by a native-server restart while keeping the
   * adapter reusable for the replacement process. A transaction that reached
   * PluginLintPool.commit but whose response was lost is compensated by abort;
   * an older fully committed host is not in this set and remains available
   * until the replacement server commits its first catalog.
   */
  async resetForServerRestart(): Promise<void> {
    this.assertActive();
    const orphaned = [...this.stagedTransactions];
    this.stagedTransactions.clear();
    for (const transactionId of orphaned) {
      this.host.deleteSession(transactionId);
    }
    await Promise.allSettled(
      orphaned.map(async (transactionId) =>
        this.pluginLintPool.abort(transactionId),
      ),
    );
  }

  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    const transactions = new Set([
      ...this.stagedTransactions,
      this.activeTransaction,
      this.predecessorTransaction,
    ]);
    for (const transactionId of transactions) {
      if (!transactionId) continue;
      this.host.deleteSession(transactionId);
    }
    this.stagedTransactions.clear();
    this.activeTransaction = undefined;
    this.predecessorTransaction = undefined;
  }

  private discardStaged(transactionId: string): void {
    this.host.deleteSession(transactionId);
    this.stagedTransactions.delete(transactionId);
  }

  private assertActive(): void {
    if (this.disposed) {
      throw new Error('config transaction adapter is disposed');
    }
  }
}
