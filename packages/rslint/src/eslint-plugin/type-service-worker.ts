import {
  MessageChannel,
  receiveMessageOnPort,
  type MessagePort,
} from 'node:worker_threads';
import {
  FrameCancelled,
  checkFrameCancellation,
  type DecodedFrame,
} from '../languages/js/frame.js';
import {
  TypeServiceClient,
  syntaxOnlyParserServices,
  type ParserServiceOptions,
  type ProgramView,
  type TypeServiceFile,
} from '../languages/js/type-service.js';
import type { TypeServiceQuery } from '../languages/js/type-service-protocol.js';
import { encodeBinaryTypeQuery } from '../languages/js/type-service-request-codec.js';
import { TYPESCRIPT_COMPILER_DATA } from '../languages/js/type-service-compiler-data.generated.js';
import {
  getNativeBinding,
  type NativeBinding,
  type WorkerRpcDescriptor,
} from '../native/binding.js';
import {
  BRIDGE_CANCELLED_TASK_IDX,
  BRIDGE_CONTROL_BYTES,
  BRIDGE_STATE_CLOSED,
  BRIDGE_STATE_ERROR,
  BRIDGE_STATE_IDX,
  BRIDGE_STATE_OPEN,
  BRIDGE_WAKE_IDX,
  type OpenTypeServiceBridge,
  type TypeServiceBridgeReplyEnvelope,
  type TypeServiceBridgeRequestEnvelope,
  type TypeServiceWorkerRequest,
  type TypeServiceWorkerResponse,
  type PluginFileTypeServices,
} from './plugin/type-service-protocol.js';

// Status codes written by crates/rslint-native/src/rpc_transport.rs into statusOut[0].
const RPC_STATUS_OK = 0;
const RPC_STATUS_CLOSED = 1;
const RPC_STATUS_CANCELLED = 2;
const RPC_STATUS_ERROR = 3;
const RPC_STATUS_UNAVAILABLE = 5;

type ParserServices = ReturnType<TypeServiceFile['parserServices']>;
type FileServices = Extract<ParserServices, { program: ProgramView }>;
interface WorkerSession {
  generation: number;
  program: ProgramView;
  client?: TypeServiceClient;
  fileIndex?: Map<string, number>;
}
interface WorkerBridge {
  port: MessagePort;
  control: Int32Array;
  nextReqId: number;
}
interface ActiveFile {
  input: PluginFileTypeServices;
  taskId: number;
  cancelFlag?: Int32Array;
}

function isBridgeReplyEnvelope(
  value: unknown,
): value is TypeServiceBridgeReplyEnvelope {
  return (
    typeof value === 'object' &&
    value !== null &&
    'reqId' in value &&
    typeof value.reqId === 'number'
  );
}

/** Program identities last across files. One persistent synchronous channel is
 * opened lazily per worker on first type-service use and reused across files. */
export class WorkerTypeServices {
  private readonly sessions = new Map<string, WorkerSession>();
  private bridge?: WorkerBridge;
  private active?: ActiveFile;
  private rpcReqBuf?: Uint8Array;
  private readonly rpcStatusOut = new Int32Array(1);
  private readonly zeroCancelFlag = new Int32Array(1);
  private rpcCallSync?: NativeBinding['rpcCallSync'];
  private rpcDisabled = false;
  constructor(
    private readonly parent: MessagePort,
    private rpc?: WorkerRpcDescriptor,
  ) {}

  attachRpc(rpc: WorkerRpcDescriptor): void {
    this.rpc = rpc;
    this.rpcReqBuf = undefined;
  }

  private ensureBridge(): WorkerBridge {
    if (this.bridge) return this.bridge;
    const { port1, port2 } = new MessageChannel();
    const control = new SharedArrayBuffer(BRIDGE_CONTROL_BYTES);
    const message: OpenTypeServiceBridge = {
      kind: 'type-bridge',
      port: port2,
      control,
    };
    try {
      this.parent.postMessage(message, [port2]);
    } catch (error) {
      port1.close();
      port2.close();
      throw error;
    }
    this.bridge = {
      port: port1,
      control: new Int32Array(control),
      nextReqId: 0,
    };
    return this.bridge;
  }

  private closeBridge(): void {
    const bridge = this.bridge;
    if (!bridge) return;
    this.bridge = undefined;
    bridge.port.close();
  }

  private checkBridgeCancel(active: ActiveFile, control: Int32Array): void {
    const state = Atomics.load(control, BRIDGE_STATE_IDX);
    if (state === BRIDGE_STATE_ERROR) {
      // Tear down a failed bridge even if cancellation also fired so the next
      // task on this worker lazily opens a fresh bridge instead of reusing a dead one.
      this.closeBridge();
    }
    checkFrameCancellation(active.cancelFlag);
    if (Atomics.load(control, BRIDGE_CANCELLED_TASK_IDX) === active.taskId) {
      throw new FrameCancelled();
    }
    if (state === BRIDGE_STATE_CLOSED) {
      this.closeBridge();
      throw new FrameCancelled();
    }
    if (state === BRIDGE_STATE_ERROR) {
      throw new Error('type service bridge failed');
    }
  }

  private call(request: TypeServiceWorkerRequest): TypeServiceWorkerResponse {
    const active = this.active;
    if (
      !active ||
      active.input.session !== request.session ||
      active.input.binding.generation !== request.generation
    )
      throw new Error('type services have no active file lease');
    checkFrameCancellation(active.cancelFlag);
    if (active.input.syntaxOnly) request = { ...request, syntaxOnly: true };
    const bridge = this.ensureBridge();
    const { port, control } = bridge;
    this.checkBridgeCancel(active, control);
    const reqId = ++bridge.nextReqId;
    const envelope: TypeServiceBridgeRequestEnvelope = {
      taskId: active.taskId,
      reqId,
      request,
    };
    port.postMessage(envelope);
    for (;;) {
      // Load wakeEpoch BEFORE checking cancel/state/port so any concurrent
      // main-thread reply or cancel bumps the epoch and prevents a lost wake-up.
      const epoch = Atomics.load(control, BRIDGE_WAKE_IDX);
      this.checkBridgeCancel(active, control);
      let raw = receiveMessageOnPort(port);
      while (raw !== undefined) {
        const message: unknown = raw.message;
        if (isBridgeReplyEnvelope(message) && message.reqId === reqId) {
          this.checkBridgeCancel(active, control);
          if (message.error) {
            if (message.error.name === 'AbortError') throw new FrameCancelled();
            if (message.error.name === 'TypeServiceClosedError')
              this.retire([request.session]);
            const error = new Error(message.error.message);
            error.name = message.error.name;
            throw error;
          }
          if (!message.result) throw new Error('invalid type service reply');
          return message.result;
        }
        raw = receiveMessageOnPort(port);
      }
      Atomics.wait(control, BRIDGE_WAKE_IDX, epoch, 50);
    }
  }

  private session(id: string, generation: number): WorkerSession {
    const current = this.sessions.get(id);
    if (current) {
      if (current.generation !== generation)
        throw new Error('foreign type service generation');
      return current;
    }
    const load = () => {
      if (this.sessions.get(id) !== session)
        throw new Error('type service session is closed');
      this.load(id, session);
      return session.client!.program;
    };
    const session: WorkerSession = {
      generation,
      program: {
        getConfigFileParsingDiagnostics: (...args) =>
          load().getConfigFileParsingDiagnostics(...args),
        getOptionsDiagnostics: (...args) =>
          load().getOptionsDiagnostics(...args),
        getGlobalDiagnostics: (...args) => load().getGlobalDiagnostics(...args),
        getSyntacticDiagnostics: (...args) =>
          load().getSyntacticDiagnostics(...args),
        getSemanticDiagnostics: (...args) =>
          load().getSemanticDiagnostics(...args),
        getDeclarationDiagnostics: (...args) =>
          load().getDeclarationDiagnostics(...args),
        getRootFileNames: (...args) => load().getRootFileNames(...args),
        emit: (...args) => load().emit(...args),
        getTypeChecker: (...args) => load().getTypeChecker(...args),
        getCompilerOptions: (...args) => load().getCompilerOptions(...args),
        getCurrentDirectory: (...args) => load().getCurrentDirectory(...args),
        getSourceFileByPath: (...args) => load().getSourceFileByPath(...args),
        getSourceFile: (...args) => load().getSourceFile(...args),
        getSourceFiles: (...args) => load().getSourceFiles(...args),
        isSourceFileFromExternalLibrary: (...args) =>
          load().isSourceFileFromExternalLibrary(...args),
        isSourceFileDefaultLibrary: (...args) =>
          load().isSourceFileDefaultLibrary(...args),
        sourceFileToPackageName: {
          get: (path) => load().sourceFileToPackageName.get(path),
        },
      },
    };
    this.sessions.set(id, session);
    return session;
  }

  private getRpcCallSync(): NativeBinding['rpcCallSync'] | undefined {
    if (this.rpcDisabled || !this.rpc) return undefined;
    if (this.rpcCallSync) return this.rpcCallSync;
    try {
      const fn = getNativeBinding().rpcCallSync;
      if (typeof fn !== 'function') {
        this.rpcDisabled = true;
        return undefined;
      }
      this.rpcCallSync = fn;
      return fn;
    } catch {
      this.rpcDisabled = true;
      return undefined;
    }
  }

  /**
   * Direct binary shared-memory fast path for semantic checker queries.
   * Preserves exact active-lease, cancellation, `syntaxOnly`, and error
   * semantics, transparently falling back to `HostTypeBridge` when the query
   * is excluded (`ProgramRequest`, `suggestionDiagnostics`), exceeds slot
   * capacity, or the SHM channel is unavailable.
   */
  private queryPacket(
    session: string,
    generation: number,
    owner: WorkerSession,
    query: TypeServiceQuery,
  ): Uint8Array {
    const active = this.active;
    if (
      !active ||
      active.input.session !== session ||
      active.input.binding.generation !== generation
    ) {
      throw new Error('type services have no active file lease');
    }
    checkFrameCancellation(active.cancelFlag);

    const rpc = this.rpc;
    if (
      !active.input.syntaxOnly &&
      rpc &&
      !this.rpcDisabled &&
      owner.fileIndex
    ) {
      const callSync = this.getRpcCallSync();
      if (callSync) {
        if (this.bridge) {
          this.checkBridgeCancel(active, this.bridge.control);
        }
        this.rpcReqBuf ??= new Uint8Array(rpc.slotSize);
        const reqLen = encodeBinaryTypeQuery(
          this.rpcReqBuf,
          session,
          query,
          owner.fileIndex,
        );
        if (reqLen > 0) {
          const replyBuf = callSync(
            rpc.channelId,
            rpc.slot,
            rpc.epoch,
            this.rpcReqBuf,
            reqLen,
            active.cancelFlag ?? this.zeroCancelFlag,
            this.rpcStatusOut,
          );
          const status = this.rpcStatusOut[0];
          if (status === RPC_STATUS_CLOSED) {
            this.retire([session]);
          }
          // Check task/bridge cancellation unconditionally before surfacing
          // RPC_STATUS_CLOSED or RPC_STATUS_ERROR so retired/cancelled tasks
          // always throw FrameCancelled rather than parseError.
          if (this.bridge) {
            this.checkBridgeCancel(active, this.bridge.control);
          } else {
            checkFrameCancellation(active.cancelFlag);
          }
          if (status === RPC_STATUS_OK) {
            if (!replyBuf) {
              throw new Error('incomplete type service query reply');
            }
            return replyBuf;
          }
          if (status === RPC_STATUS_CLOSED) {
            const error = new Error('type service session is stale or closed');
            error.name = 'TypeServiceClosedError';
            throw error;
          }
          if (status === RPC_STATUS_CANCELLED) {
            throw new FrameCancelled();
          }
          if (status === RPC_STATUS_ERROR) {
            const message =
              replyBuf && replyBuf.byteLength > 0
                ? replyBuf.toString('utf8')
                : 'type service query failed';
            throw new Error(message);
          }
          if (status === RPC_STATUS_UNAVAILABLE) {
            this.rpcDisabled = true;
          }
          // RPC_STATUS_FALLBACK / RPC_STATUS_UNAVAILABLE: transparently fall back.
        }
      }
    }

    const answer = this.call({
      session,
      generation,
      queries: [query],
    });
    if (answer.packets?.length !== 1) {
      throw new Error('incomplete type service query reply');
    }
    return answer.packets[0];
  }

  private load(
    id: string,
    owner: WorkerSession,
    projection?: string,
  ): TypeServiceWorkerResponse {
    if (this.sessions.get(id) !== owner)
      throw new Error('type service session is closed');
    if (owner.client && projection === undefined) return {};
    const generation = owner.generation;
    const reply = this.call({
      session: id,
      generation,
      open: owner.client ? undefined : true,
      projection,
    });
    if (!owner.client) {
      if (!reply.descriptor) throw new Error('missing type service descriptor');
      checkFrameCancellation(this.active?.cancelFlag);
      const fileIndex = new Map<string, number>();
      const sources = reply.descriptor.sources;
      for (let i = 0; i < sources.length; i++) {
        fileIndex.set(sources[i].fileName, i);
      }
      owner.fileIndex = fileIndex;
      owner.client = new TypeServiceClient(
        TYPESCRIPT_COMPILER_DATA,
        reply.descriptor,
        (query) => this.queryPacket(id, generation, owner, query),
        (scope, file, metadataOnly) =>
          this.call({
            session: id,
            generation: scope.generation,
            source: file,
            metadataOnly,
          }).source,
        { maxFactsBytes: 8 << 20 },
      );
    }
    return reply;
  }

  file(input: PluginFileTypeServices, taskId: number, cancelFlag?: Int32Array) {
    const active: ActiveFile = { input, taskId, cancelFlag };
    this.active = active;
    const { session: id, binding } = input;
    let owner: WorkerSession | undefined;
    let frame: DecodedFrame | undefined;
    let fileParserOptions: Readonly<ParserServiceOptions> | undefined;
    let file: TypeServiceFile | undefined;
    let services: ParserServices | undefined;
    const ensureFile = () => {
      if (this.active !== active)
        throw new Error('file type services are closed');
      checkFrameCancellation(cancelFlag);
    };
    const session = () => {
      if (!owner) {
        ensureFile();
        owner = this.session(id, binding.generation);
      }
      return owner;
    };
    const syntax = () => {
      ensureFile();
      if (services) return services;
      const current = session();
      const sourceFileName = binding.sourceFileName;
      const { projection } = this.load(id, current, sourceFileName);
      if (!projection || !frame)
        throw new Error('missing type service projection');
      file = current.client!.file(binding);
      try {
        services = file.parserServices(
          frame,
          { ...projection, binding, sourceFileName },
          sourceFileName,
          fileParserOptions,
        );
      } catch (error) {
        file.close();
        file = undefined;
        throw error;
      }
      return services;
    };
    const typed = (): FileServices => {
      const result = syntax();
      if (!result.program)
        throw new Error('semantic queries require a typed project');
      return result;
    };
    return {
      services: (
        parsed: DecodedFrame,
        parserOptions?: Readonly<ParserServiceOptions>,
      ): ParserServices => {
        frame = parsed;
        fileParserOptions = parserOptions;
        const maps: Pick<
          FileServices,
          'esTreeNodeToTSNodeMap' | 'tsNodeToESTreeNodeMap'
        > = {
          esTreeNodeToTSNodeMap: {
            get: (node) => syntax().esTreeNodeToTSNodeMap.get(node),
            has: (node) => syntax().esTreeNodeToTSNodeMap.has(node),
          },
          tsNodeToESTreeNodeMap: {
            get: (node) => syntax().tsNodeToESTreeNodeMap.get(node),
            has: (node) => syntax().tsNodeToESTreeNodeMap.has(node),
          },
        };
        if (input.syntaxOnly) {
          return syntaxOnlyParserServices(maps, {
            emitDecoratorMetadata:
              parserOptions?.emitDecoratorMetadata ??
              input.emitDecoratorMetadata === true,
            experimentalDecorators:
              parserOptions?.experimentalDecorators ??
              input.experimentalDecorators === true,
            isolatedDeclarations:
              parserOptions?.isolatedDeclarations ??
              input.isolatedDeclarations === true,
          });
        }
        return {
          ...maps,
          get program() {
            return session().program;
          },
          getContextualType: (...args) => typed().getContextualType(...args),
          getResolvedSignature: (...args) =>
            typed().getResolvedSignature(...args),
          getTypeAtLocation: (...args) => typed().getTypeAtLocation(...args),
          getSymbolAtLocation: (...args) =>
            typed().getSymbolAtLocation(...args),
          getTypeFromTypeNode: (...args) =>
            typed().getTypeFromTypeNode(...args),
          getTypeOfSymbolAtLocation: (...args) =>
            typed().getTypeOfSymbolAtLocation(...args),
          emitDecoratorMetadata: input.emitDecoratorMetadata === true,
          experimentalDecorators: input.experimentalDecorators === true,
          isolatedDeclarations: input.isolatedDeclarations === true,
        };
      },
      close: () => {
        frame = undefined;
        fileParserOptions = undefined;
        file?.close();
        if (input.syntaxOnly && owner) this.retire([id]);
        file = undefined;
        services = undefined;
        // Tear down a dead bridge or drain any unread reply left over from a
        // cancelled call so buffers are released immediately.
        if (this.bridge) {
          if (
            Atomics.load(this.bridge.control, BRIDGE_STATE_IDX) !==
            BRIDGE_STATE_OPEN
          ) {
            this.closeBridge();
          } else {
            while (receiveMessageOnPort(this.bridge.port) !== undefined) {
              /* discard stale reply */
            }
          }
        }
        if (this.active === active) this.active = undefined;
      },
    };
  }
  retire(sessions: readonly string[]): void {
    for (const id of sessions) {
      const owner = this.sessions.get(id);
      this.sessions.delete(id);
      owner?.client?.close();
      if (owner) owner.client = undefined;
    }
  }
  close(): void {
    this.retire([...this.sessions.keys()]);
    this.closeBridge();
  }
}
