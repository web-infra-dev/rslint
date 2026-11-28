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
import { TYPESCRIPT_COMPILER_DATA } from '../languages/js/type-service-compiler-data.generated.js';
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

type ParserServices = ReturnType<TypeServiceFile['parserServices']>;
type FileServices = Extract<ParserServices, { program: ProgramView }>;
interface WorkerSession {
  generation: number;
  program: ProgramView;
  client?: TypeServiceClient;
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

/** Program identities last across files. One persistent synchronous channel is
 * opened lazily per worker on first type-service use and reused across files. */
export class WorkerTypeServices {
  private readonly sessions = new Map<string, WorkerSession>();
  private bridge?: WorkerBridge;
  private active?: ActiveFile;
  constructor(private readonly parent: MessagePort) {}

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
        const received = raw.message as
          TypeServiceBridgeReplyEnvelope | undefined;
        if (received && received.reqId === reqId) {
          this.checkBridgeCancel(active, control);
          if (received.error) {
            if (received.error.name === 'AbortError')
              throw new FrameCancelled();
            if (received.error.name === 'TypeServiceClosedError')
              this.retire([request.session]);
            const error = new Error(received.error.message);
            error.name = received.error.name;
            throw error;
          }
          if (!received.result) throw new Error('invalid type service reply');
          return received.result;
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
      owner.client = new TypeServiceClient(
        TYPESCRIPT_COMPILER_DATA,
        reply.descriptor,
        (query) => {
          const answer = this.call({
            session: id,
            generation,
            queries: [query],
          });
          if (answer.packets?.length !== 1)
            throw new Error('incomplete type service query reply');
          return answer.packets[0];
        },
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
