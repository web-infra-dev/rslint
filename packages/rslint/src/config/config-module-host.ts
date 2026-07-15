import { createHash } from 'node:crypto';
import { readFile } from 'node:fs/promises';
import path from 'node:path';

import {
  collectPluginMeta,
  loadConfigFile,
  loadConfigFileFresh,
  normalizeConfig,
  type PluginConfigDescriptor,
} from './config-file-loader.js';
import {
  CONFIG_DISCOVERY_PROTOCOL_VERSION,
  type ActivateConfigsRequest,
  type ActivateConfigsResponse,
  type ConfigModuleCandidate,
  type ConfigModuleEslintPluginEntry,
  type ConfigModuleLoadMode,
  type ConfigModuleLoadResult,
  type FailedConfigModuleResult,
  type LoadConfigsRequest,
  type LoadConfigsResponse,
} from './config-discovery-protocol.js';

type ConfigModuleLoader = (configPath: string) => Promise<unknown>;
type ConfigSourceReader = (configPath: string) => Promise<Uint8Array>;

export interface ConfigModuleHostOptions {
  /** Test/embedding seam; production defaults to loadConfigFile. */
  loadCached?: ConfigModuleLoader;
  /** Test/embedding seam; production defaults to loadConfigFileFresh. */
  loadFresh?: ConfigModuleLoader;
  /** Test/embedding seam; production defaults to fs.promises.readFile. */
  readSource?: ConfigSourceReader;
}

export interface EffectiveConfigModule {
  id: string;
  configPath: string;
  configDirectory: string;
  entries: Record<string, unknown>[];
  sourceFingerprint: string;
}

export type ConfigModulePluginDescriptor = PluginConfigDescriptor;

/**
 * Node-local activation data derived from exactly the effective IDs selected
 * by Go. It is exposed only to the prepare callback and is never serialized
 * as the activation response.
 */
export interface ConfigModuleActivationPlan {
  transactionId: string;
  configs: EffectiveConfigModule[];
  eslintPluginEntries: ConfigModuleEslintPluginEntry[];
  pluginConfigs: ConfigModulePluginDescriptor[];
}

interface StoredCandidate {
  candidate: ConfigModuleCandidate;
  result: StoredLoadResult;
}

interface StoredLoadedResult {
  id: string;
  status: 'loaded';
  /** JSON is the immutable session copy and exactly matches the wire shape. */
  entriesJSON: string;
  sourceFingerprint: string;
}

type StoredLoadResult = StoredLoadedResult | FailedConfigModuleResult;

interface ConfigModuleSession {
  loadMode: ConfigModuleLoadMode;
  singleThreaded: boolean;
  candidatesById: Map<string, StoredCandidate>;
  idByConfigPath: Map<string, string>;
  operation: Promise<void>;
}

class ConfigSourceChangedError extends Error {
  readonly code = 'CONFIG_CHANGED_DURING_LOAD';

  constructor(configPath: string, phase: 'load' | 'plugin-prepare' = 'load') {
    super(
      phase === 'load'
        ? `config changed while it was being loaded: ${configPath}`
        : `config changed while its plugin host was being prepared: ${configPath}`,
    );
    this.name = 'ConfigSourceChangedError';
  }
}

function assertNotAborted(signal?: AbortSignal): void {
  if (!signal?.aborted) return;
  if (signal.reason instanceof Error) throw signal.reason;
  throw new Error('config module operation was aborted');
}

function protocolError(message: string): Error {
  return new Error(`config module protocol: ${message}`);
}

function isConfigEntry(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function normalizeCandidate(
  candidate: ConfigModuleCandidate,
): ConfigModuleCandidate {
  if (!candidate || typeof candidate !== 'object') {
    throw protocolError('candidate must be an object');
  }
  if (typeof candidate.id !== 'string' || candidate.id.length === 0) {
    throw protocolError('candidate id must be a non-empty string');
  }
  if (
    typeof candidate.configPath !== 'string' ||
    !path.isAbsolute(candidate.configPath)
  ) {
    throw protocolError(
      `candidate ${JSON.stringify(candidate.id)} configPath must be absolute`,
    );
  }
  if (
    typeof candidate.configDirectory !== 'string' ||
    !path.isAbsolute(candidate.configDirectory)
  ) {
    throw protocolError(
      `candidate ${JSON.stringify(candidate.id)} configDirectory must be absolute`,
    );
  }
  return {
    id: candidate.id,
    // configPath is a Node-local filesystem operand, so normalize it to the
    // host platform before fingerprinting/importing. configDirectory is
    // different: Go owns it as an opaque routing identity and later echoes it
    // byte-for-byte in pluginLint.configKey. Normalizing that token here would
    // turn `C:/...` into `C:\\...` on Windows and break worker map lookup.
    configPath: path.normalize(candidate.configPath),
    configDirectory: candidate.configDirectory,
  };
}

function cloneEntries(entriesJSON: string): Record<string, unknown>[] {
  const entries: unknown = JSON.parse(entriesJSON);
  if (!Array.isArray(entries) || !entries.every(isConfigEntry)) {
    throw protocolError('stored config entries are malformed');
  }
  return entries;
}

function clonePluginEntries(
  entries: readonly ConfigModuleEslintPluginEntry[],
): ConfigModuleEslintPluginEntry[] {
  return entries.map(({ prefix, ruleNames }) => ({
    prefix,
    ruleNames: [...ruleNames],
  }));
}

function cloneStoredResult(result: StoredLoadResult): ConfigModuleLoadResult {
  if (result.status === 'loaded') {
    return {
      id: result.id,
      status: 'loaded',
      entries: cloneEntries(result.entriesJSON),
    };
  }
  return {
    id: result.id,
    status: 'failed',
    error: { ...result.error },
  };
}

function errorPayload(
  error: unknown,
  fallbackCode?: string,
): FailedConfigModuleResult['error'] {
  const message = error instanceof Error ? error.message : String(error);
  let code = fallbackCode;
  if (error !== null && typeof error === 'object' && 'code' in error) {
    const candidateCode = error.code;
    if (typeof candidateCode === 'string' && candidateCode.length > 0) {
      code = candidateCode;
    }
  }
  return { message, ...(code ? { code } : {}) };
}

/**
 * Executes config modules on behalf of Go's discovery coordinator.
 *
 * A host may serve multiple concurrent discovery transactions. Operations in
 * one transaction are serialized (frontier batches can safely build on earlier
 * batches), while separate transactions remain independent. Call
 * deleteSession after commit/abort so normalized entries do not remain live.
 */
export class ConfigModuleHost {
  readonly #loadCached: ConfigModuleLoader;
  readonly #loadFresh: ConfigModuleLoader;
  readonly #readSource: ConfigSourceReader;
  readonly #sessions = new Map<string, ConfigModuleSession>();

  constructor(options: ConfigModuleHostOptions = {}) {
    this.#loadCached = options.loadCached ?? loadConfigFile;
    this.#loadFresh = options.loadFresh ?? loadConfigFileFresh;
    this.#readSource = options.readSource ?? readFile;
  }

  /** Load and normalize one discovery frontier without fail-fast evaluation. */
  async loadConfigs(
    request: LoadConfigsRequest,
    signal?: AbortSignal,
  ): Promise<LoadConfigsResponse> {
    this.#validateRequest(request);
    // Reject malformed paths before allocating transaction state. Protocol
    // errors must not leave an unreachable empty session behind.
    const candidates = request.candidates.map(normalizeCandidate);
    const session = this.#getOrCreateSession(
      request.transactionId,
      request.loadMode,
      request.singleThreaded === true,
    );
    return this.#enqueue(session, async () => {
      assertNotAborted(signal);
      this.#validateCandidateBatch(session, candidates);

      const newCandidates = candidates.filter(
        ({ id }) => !session.candidatesById.has(id),
      );
      let loaded: StoredLoadResult[];
      if (request.singleThreaded) {
        loaded = [];
        for (const candidate of newCandidates) {
          loaded.push(await this.#loadCandidate(candidate, session.loadMode));
        }
      } else {
        loaded = await Promise.all(
          newCandidates.map(async (candidate) => {
            const result = await this.#loadCandidate(
              candidate,
              session.loadMode,
            );
            return result;
          }),
        );
      }
      // Module evaluation itself cannot always be interrupted. Do not publish
      // any of its results into session state after a superseding transaction.
      assertNotAborted(signal);

      for (let index = 0; index < newCandidates.length; index++) {
        const candidate = newCandidates[index];
        const result = loaded[index];
        session.candidatesById.set(candidate.id, { candidate, result });
        session.idByConfigPath.set(candidate.configPath, candidate.id);
      }

      return {
        transactionId: request.transactionId,
        results: candidates.map(({ id }) => {
          const stored = session.candidatesById.get(id);
          if (!stored) {
            throw protocolError(
              `candidate ${JSON.stringify(id)} was not loaded`,
            );
          }
          return cloneStoredResult(stored.result);
        }),
      };
    });
  }

  /**
   * Build plugin-host inputs for exactly the effective IDs selected by Go.
   * Sources are fingerprinted here before plugin-host preparation; the
   * protocol-facing activateConfigs performs the matching post-prepare check.
   */
  async #summarizeEffectiveConfigs(
    transactionId: string,
    effectiveConfigIds: readonly string[],
    signal?: AbortSignal,
  ): Promise<ConfigModuleActivationPlan> {
    const session = this.#sessions.get(transactionId);
    if (!session) {
      throw protocolError(
        `unknown transaction ${JSON.stringify(transactionId)}`,
      );
    }
    return this.#enqueue(session, async () => {
      assertNotAborted(signal);
      const stored = this.#effectiveCandidates(session, effectiveConfigIds);
      await this.#verifyFingerprints(stored, 'load', signal);

      const configs = stored.map(({ candidate, result }) => ({
        id: candidate.id,
        configPath: candidate.configPath,
        configDirectory: candidate.configDirectory,
        entries: cloneEntries(result.entriesJSON),
        sourceFingerprint: result.sourceFingerprint,
      }));
      const { eslintPluginEntries, pluginConfigs } = collectPluginMeta(configs);
      return {
        transactionId,
        configs,
        eslintPluginEntries: clonePluginEntries(eslintPluginEntries),
        pluginConfigs: pluginConfigs.map((descriptor) => ({ ...descriptor })),
      };
    });
  }

  /**
   * Protocol-facing activation transaction.
   *
   * Plugin workers re-import plugin-bearing config modules. The optional
   * prepare callback therefore runs between two source-fingerprint checks.
   * A caller cannot publish the normalized activation until every effective
   * source still matches the bytes from which that activation was derived.
   */
  async activateConfigs(
    request: ActivateConfigsRequest,
    signal?: AbortSignal,
    prepare?: (plan: ConfigModuleActivationPlan) => Promise<void>,
  ): Promise<ActivateConfigsResponse> {
    this.#validateActivationRequest(request);
    const plan = await this.#summarizeEffectiveConfigs(
      request.transactionId,
      request.effectiveConfigIds,
      signal,
    );
    const response: ActivateConfigsResponse = {
      transactionId: plan.transactionId,
      eslintPluginEntries: clonePluginEntries(plan.eslintPluginEntries),
    };
    assertNotAborted(signal);
    await prepare?.(plan);
    assertNotAborted(signal);
    await this.#verifyEffectiveConfigs(request, signal);
    return response;
  }

  /**
   * Recheck the original source fingerprints for exactly Go's effective IDs.
   * activateConfigs keeps this check private so adapters cannot accidentally
   * publish a plugin host without validating its source bytes again.
   */
  async #verifyEffectiveConfigs(
    request: ActivateConfigsRequest,
    signal?: AbortSignal,
  ): Promise<void> {
    this.#validateActivationRequest(request);
    const session = this.#sessions.get(request.transactionId);
    if (!session) {
      throw protocolError(
        `unknown transaction ${JSON.stringify(request.transactionId)}`,
      );
    }
    await this.#enqueue(session, async () => {
      assertNotAborted(signal);
      const stored = this.#effectiveCandidates(
        session,
        request.effectiveConfigIds,
      );
      await this.#verifyFingerprints(stored, 'plugin-prepare', signal);
    });
  }

  /** Drop all normalized module state for a committed or aborted transaction. */
  deleteSession(transactionId: string): boolean {
    return this.#sessions.delete(transactionId);
  }

  #validateRequest(request: LoadConfigsRequest): void {
    if (!request || typeof request !== 'object') {
      throw protocolError('load request must be an object');
    }
    if (request.protocolVersion !== CONFIG_DISCOVERY_PROTOCOL_VERSION) {
      throw protocolError(
        `unsupported version ${String(request.protocolVersion)}; expected ${CONFIG_DISCOVERY_PROTOCOL_VERSION}`,
      );
    }
    if (
      typeof request.transactionId !== 'string' ||
      request.transactionId.length === 0
    ) {
      throw protocolError('transactionId must be a non-empty string');
    }
    if (request.loadMode !== 'cached' && request.loadMode !== 'fresh') {
      throw protocolError(`unsupported loadMode ${String(request.loadMode)}`);
    }
    if (
      request.singleThreaded !== undefined &&
      typeof request.singleThreaded !== 'boolean'
    ) {
      throw protocolError('singleThreaded must be a boolean when provided');
    }
    if (!Array.isArray(request.candidates)) {
      throw protocolError('candidates must be an array');
    }
  }

  #validateActivationRequest(request: ActivateConfigsRequest): void {
    if (!request || typeof request !== 'object') {
      throw protocolError('activation request must be an object');
    }
    if (request.protocolVersion !== CONFIG_DISCOVERY_PROTOCOL_VERSION) {
      throw protocolError(
        `unsupported version ${String(request.protocolVersion)}; expected ${CONFIG_DISCOVERY_PROTOCOL_VERSION}`,
      );
    }
    if (
      typeof request.transactionId !== 'string' ||
      request.transactionId.length === 0
    ) {
      throw protocolError('transactionId must be a non-empty string');
    }
    if (!Array.isArray(request.effectiveConfigIds)) {
      throw protocolError('effectiveConfigIds must be an array');
    }
  }

  #getOrCreateSession(
    transactionId: string,
    loadMode: ConfigModuleLoadMode,
    singleThreaded: boolean,
  ): ConfigModuleSession {
    const existing = this.#sessions.get(transactionId);
    if (existing) {
      if (existing.loadMode !== loadMode) {
        throw protocolError(
          `transaction ${JSON.stringify(transactionId)} cannot change loadMode from ${existing.loadMode} to ${loadMode}`,
        );
      }
      if (existing.singleThreaded !== singleThreaded) {
        throw protocolError(
          `transaction ${JSON.stringify(transactionId)} cannot change singleThreaded mode`,
        );
      }
      return existing;
    }
    const session: ConfigModuleSession = {
      loadMode,
      singleThreaded,
      candidatesById: new Map(),
      idByConfigPath: new Map(),
      operation: Promise.resolve(),
    };
    this.#sessions.set(transactionId, session);
    return session;
  }

  #validateCandidateBatch(
    session: ConfigModuleSession,
    candidates: readonly ConfigModuleCandidate[],
  ): void {
    const ids = new Set<string>();
    const paths = new Map<string, string>();
    for (const candidate of candidates) {
      if (ids.has(candidate.id)) {
        throw protocolError(
          `candidate batch contains duplicate id ${JSON.stringify(candidate.id)}`,
        );
      }
      ids.add(candidate.id);

      const batchOwner = paths.get(candidate.configPath);
      if (batchOwner && batchOwner !== candidate.id) {
        throw protocolError(
          `candidate IDs ${JSON.stringify(batchOwner)} and ${JSON.stringify(candidate.id)} use the same configPath`,
        );
      }
      paths.set(candidate.configPath, candidate.id);

      const existing = session.candidatesById.get(candidate.id);
      if (
        existing &&
        (existing.candidate.configPath !== candidate.configPath ||
          existing.candidate.configDirectory !== candidate.configDirectory)
      ) {
        throw protocolError(
          `candidate id ${JSON.stringify(candidate.id)} changed path or directory within one transaction`,
        );
      }
      const pathOwner = session.idByConfigPath.get(candidate.configPath);
      if (pathOwner && pathOwner !== candidate.id) {
        throw protocolError(
          `candidate IDs ${JSON.stringify(pathOwner)} and ${JSON.stringify(candidate.id)} use the same configPath`,
        );
      }
    }
  }

  #effectiveCandidates(
    session: ConfigModuleSession,
    effectiveConfigIds: readonly string[],
  ): Array<{ candidate: ConfigModuleCandidate; result: StoredLoadedResult }> {
    const seen = new Set<string>();
    return effectiveConfigIds.map((id) => {
      if (seen.has(id)) {
        throw protocolError(
          `effective config list contains duplicate id ${JSON.stringify(id)}`,
        );
      }
      seen.add(id);
      const candidate = session.candidatesById.get(id);
      if (!candidate) {
        throw protocolError(
          `unknown effective config id ${JSON.stringify(id)}`,
        );
      }
      if (candidate.result.status !== 'loaded') {
        throw protocolError(
          `effective config ${JSON.stringify(id)} did not load successfully`,
        );
      }
      return {
        candidate: candidate.candidate,
        result: candidate.result,
      };
    });
  }

  async #verifyFingerprints(
    stored: ReadonlyArray<{
      candidate: ConfigModuleCandidate;
      result: StoredLoadedResult;
    }>,
    phase: 'load' | 'plugin-prepare',
    signal?: AbortSignal,
  ): Promise<void> {
    const currentFingerprints = await Promise.all(
      stored.map(async ({ candidate }) => {
        const fingerprint = await this.#fingerprint(candidate.configPath);
        return fingerprint;
      }),
    );
    assertNotAborted(signal);
    for (let index = 0; index < stored.length; index++) {
      if (
        currentFingerprints[index] !== stored[index].result.sourceFingerprint
      ) {
        throw new ConfigSourceChangedError(
          stored[index].candidate.configPath,
          phase,
        );
      }
    }
  }

  async #loadCandidate(
    candidate: ConfigModuleCandidate,
    loadMode: ConfigModuleLoadMode,
  ): Promise<StoredLoadResult> {
    // Preserve the stage at which a candidate failed. Besides making catalog
    // failures actionable, this lets CLI/API/LSP distinguish module-loading
    // failures from a module that evaluated but exported an invalid config.
    let failureCode = 'load';
    try {
      const sourceFingerprint = await this.#fingerprint(candidate.configPath);
      const rawConfig = await (loadMode === 'fresh'
        ? this.#loadFresh(candidate.configPath)
        : this.#loadCached(candidate.configPath));
      failureCode = 'invalid';
      const normalized = normalizeConfig(rawConfig);
      // This is the cross-process serialization boundary. Keeping the canonical
      // JSON also prevents a caller mutating a response from corrupting session
      // state used later by #summarizeEffectiveConfigs.
      const entriesJSON = JSON.stringify(normalized);
      const afterFingerprint = await this.#fingerprint(candidate.configPath);
      if (sourceFingerprint !== afterFingerprint) {
        throw new ConfigSourceChangedError(candidate.configPath);
      }
      return {
        id: candidate.id,
        status: 'loaded',
        entriesJSON,
        sourceFingerprint: afterFingerprint,
      };
    } catch (error) {
      return {
        id: candidate.id,
        status: 'failed',
        error: errorPayload(error, failureCode),
      };
    }
  }

  async #fingerprint(configPath: string): Promise<string> {
    const contents = await this.#readSource(configPath);
    return `${contents.byteLength}:${createHash('sha256').update(contents).digest('hex')}`;
  }

  async #enqueue<T>(
    session: ConfigModuleSession,
    operation: () => Promise<T>,
  ): Promise<T> {
    const result = session.operation.then(operation, operation);
    session.operation = result.then(
      () => undefined,
      () => undefined,
    );
    const value = await result;
    return value;
  }
}
