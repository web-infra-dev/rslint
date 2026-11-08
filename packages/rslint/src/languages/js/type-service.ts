import type ts from 'typescript';
import type { DecodedFrame } from './frame.js';
import type { ESTreeNode } from './types.js';
import type { SharedTypeCacheReader } from './shared-type-cache.js';
import { WeakObjectCache } from './type-service-cache.js';
import { typeReplyPackets, TypePacketReader } from './type-service-codec.js';
import {
  FileSyntaxViews,
  GenerationSyntaxViews,
} from './type-service-syntax.js';
import {
  typeQueryKey,
  nameToUnits,
  type FileBinding,
  type TypeServiceDescriptor,
  type TypeServiceProjection,
  type QueryTransport,
  type SyntaxSource,
  type QueryResults,
  type TypeServiceQuery,
  type TypeFact,
  type SymbolFact,
  type SignatureFact,
  type ServiceRequest,
  type ScalarRequest,
  type BatchTransport,
} from './type-service-protocol.js';
export { typeQueryKey } from './type-service-protocol.js';
export type {
  TypeServiceDescriptor,
  TypeServiceProjection,
  TypeServiceQuery,
} from './type-service-protocol.js';

export interface TypeView {
  readonly id: number;
  readonly flags: number;
  readonly symbol: SymbolView | undefined;
  readonly aliasSymbol: SymbolView | undefined;
  readonly intrinsicName: string | undefined;
  readonly escapedName: string | undefined;
  readonly objectFlags: number | undefined;
  readonly value: TypeFact['value'];
  readonly types: readonly TypeView[] | undefined;
  getSymbol(): SymbolView | undefined;
  isUnion(): boolean;
  isIntersection(): boolean;
  isUnionOrIntersection(): boolean;
  isLiteral(): boolean;
  isStringLiteral(): boolean;
  isNumberLiteral(): boolean;
  isTypeParameter(): boolean;
  getCallSignatures(): readonly SignatureView[];
  getConstructSignatures(): readonly SignatureView[];
  getProperties(): readonly SymbolView[];
  getProperty(name: string): SymbolView | undefined;
  getNonNullableType(): TypeView | undefined;
  getConstraint(): TypeView | undefined;
  getNumberIndexType(): TypeView | undefined;
}
export interface SymbolView {
  readonly id: number;
  readonly name: string;
  readonly escapedName: string;
  readonly flags: number;
  readonly declarations: readonly object[] | undefined;
  readonly valueDeclaration: object | undefined;
  getName(): string;
  getFlags(): number;
  getDeclarations(): readonly object[] | undefined;
}
export interface SignatureView {
  readonly id: number;
  readonly parameters: readonly SymbolView[];
  readonly declaration: object | undefined;
  getParameters(): readonly SymbolView[];
  getDeclaration(): object | undefined;
  getReturnType(): TypeView | undefined;
}
export interface CheckerView {
  getTypeAtLocation(n: object): TypeView | undefined;
  getTypeFromTypeNode(n: object): TypeView | undefined;
  getContextualType(n: object): TypeView | undefined;
  getVoidType(): TypeView | undefined;
  getBaseConstraintOfType(t: object): TypeView | undefined;
  getApparentType(t: object): TypeView | undefined;
  getBaseTypeOfLiteralType(t: object): TypeView | undefined;
  getNonNullableType(t: object): TypeView | undefined;
  typeToString(t: object, enclosing?: object, flags?: number): string;
  getSymbolAtLocation(n: object): SymbolView | undefined;
  resolveName(
    name: string,
    location: object | undefined,
    meaning: number,
    excludeGlobals: boolean,
  ): SymbolView | undefined;
  getAliasedSymbol(s: object): SymbolView | undefined;
  getTypeOfSymbolAtLocation(s: object, n: object): TypeView | undefined;
  getDeclaredTypeOfSymbol(s: object): TypeView | undefined;
  getPropertiesOfType(t: object): readonly SymbolView[];
  getPropertyOfType(t: object, name: string): SymbolView | undefined;
  getTypeArguments(t: object): readonly TypeView[];
  isArrayType(t: object): boolean;
  isArrayLikeType(t: object): boolean;
  isTupleType(t: object): boolean;
  getSignatureFromDeclaration(n: object): SignatureView | undefined;
  getResolvedSignature(
    n: object,
    candidates?: unknown,
    argumentCount?: unknown,
  ): SignatureView | undefined;
  getReturnTypeOfSignature(s: object): TypeView | undefined;
  getSignaturesOfType(t: object, kind: number): readonly SignatureView[];
}
export interface ProgramView {
  getTypeChecker(): CheckerView;
  getCompilerOptions(): Readonly<Record<string, unknown>>;
  getCurrentDirectory(): string;
  getSourceFile(name: string): object | undefined;
  getSourceFiles(): readonly object[];
  sourceFileToPackageName: { get(path: string): string | undefined };
  isSourceFileFromExternalLibrary(source: object): boolean;
  isSourceFileDefaultLibrary(source: object): boolean;
}
function unsupported(surface: string): never {
  throw new Error(`unsupported TypeScript operation: ${surface}`);
}
function guarded<T extends object>(
  object: T,
  ensure: () => void,
  surface: string,
): T {
  return new Proxy(object, {
    get(target, key, receiver) {
      ensure();
      if (
        typeof key === 'string' &&
        !(key in target) &&
        key !== 'then' &&
        key !== 'toJSON'
      )
        unsupported(`${surface}.${key}`);
      const value: unknown = Reflect.get(target, key, receiver);
      return typeof value === 'function'
        ? (...args: unknown[]) => {
            ensure();
            return Reflect.apply(value, receiver, args);
          }
        : value;
    },
  });
}
// The wrapper caches decoded values only. This closure never captures a packet
// reader, shared byte view, request lease, or syntax transport reply.
function memo<T>(ensure: () => void, read: () => T): () => T {
  let cached: { value: T } | undefined;
  return () => {
    ensure();
    return (cached ??= { value: read() }).value;
  };
}
type NodeOperation =
  | 'type'
  | 'fromTypeNode'
  | 'context'
  | 'symbol'
  | 'signature'
  | 'resolvedSignature';
interface QueryProfile {
  shapes: Map<string, Set<NodeOperation>>;
  maxQueries: number;
}
type ObjectKind = 'type' | 'symbol' | 'signature';
type TypeScriptEnums = Pick<
  typeof ts,
  | 'SyntaxKind'
  | 'TypeFlags'
  | 'ObjectFlags'
  | 'NodeFlags'
  | 'JsxEmit'
  | 'ModuleDetectionKind'
>;

export interface TypeServiceClientOptions {
  maxFactsBytes: number;
  maxSyntaxBytes?: number;
  maxSyntaxFactsBytes?: number;
  shared?: SharedTypeCacheReader;
  batchTransport?: BatchTransport;
}
/** One worker owns this client for one Program generation. File handles own
 * ESTree mappings; immutable semantic and native identities survive file close. */
export class TypeServiceClient {
  private closed = false;
  private readonly facts = new Map<string, TypePacketReader>();
  private readonly types = new WeakObjectCache<number, TypeView>();
  private readonly symbols = new WeakObjectCache<number, SymbolView>();
  private readonly signatures = new WeakObjectCache<number, SignatureView>();
  private readonly brands = new WeakMap<
    object,
    { kind: ObjectKind; id: number }
  >();
  private shared: SharedTypeCacheReader | undefined;
  private readonly maxFactsBytes: number;
  private readonly syntax: GenerationSyntaxViews;
  private readonly fileOrdinals = new Map<string, number>();
  private readonly files = new Set<TypeServiceFile>();
  private batchTransport: BatchTransport | undefined;
  private readonly profiles = new Map<string, QueryProfile>();
  private activeFile:
    | {
        file: TypeServiceFile;
        name: string;
        profile: QueryProfile;
        queries: number;
      }
    | undefined;
  readonly stats = { queries: 0, localHits: 0, sharedHits: 0, factBytes: 0 };
  readonly program: ProgramView;
  readonly checker: CheckerView;
  constructor(
    private readonly typescript: TypeScriptEnums,
    readonly descriptor: TypeServiceDescriptor,
    private transport: QueryTransport | undefined,
    syntax: SyntaxSource,
    options: TypeServiceClientOptions,
  ) {
    if (
      !Number.isSafeInteger(options.maxFactsBytes) ||
      options.maxFactsBytes < 0 ||
      !Number.isSafeInteger(options.maxSyntaxBytes ?? 16 << 20) ||
      (options.maxSyntaxBytes ?? 16 << 20) < 0
    )
      throw new Error('invalid type service client budget');
    if (
      !Number.isSafeInteger(options.maxSyntaxFactsBytes ?? 1 << 20) ||
      (options.maxSyntaxFactsBytes ?? 1 << 20) < 0
    )
      throw new Error('invalid syntax fact budget');
    this.maxFactsBytes = options.maxFactsBytes;
    this.shared = options.shared;
    this.batchTransport = options.batchTransport;
    if (
      this.shared &&
      this.shared.descriptor.generation !== descriptor.generation
    )
      throw new Error('foreign shared type cache');
    descriptor.sources.forEach((source, index) => {
      this.fileOrdinals.set(source.fileName, index + 1);
      this.fileOrdinals.set(source.path, index + 1);
    });
    this.syntax = new GenerationSyntaxViews(
      typescript,
      descriptor,
      syntax,
      (node) =>
        this.read({ op: 'nodeRecord', args: { node } }, (packet) =>
          packet.nodeFacts(),
        ),
      () => this.ensure(),
      options.maxSyntaxBytes ?? 16 << 20,
      options.maxSyntaxFactsBytes ?? 1 << 20,
    );
    const nodeType = (op: 'type' | 'fromTypeNode' | 'context', n: object) =>
      this.type(
        this.query({ op: op, args: { node: this.syntax.location(n) } }),
      );
    const typeType = (
      op: 'constraint' | 'apparent' | 'baseLiteral' | 'nonNullable',
      t: object,
    ) => this.type(this.query({ op: op, args: { id: this.id(t, 'type') } }));
    this.checker = guarded<CheckerView>(
      {
        getTypeAtLocation: (n) => nodeType('type', n),
        getTypeFromTypeNode: (n) => nodeType('fromTypeNode', n),
        getContextualType: (n) => nodeType('context', n),
        getVoidType: () => this.type(this.query({ op: 'void', args: {} })),
        getBaseConstraintOfType: (t) => typeType('constraint', t),
        getApparentType: (t) => typeType('apparent', t),
        getBaseTypeOfLiteralType: (t) => typeType('baseLiteral', t),
        getNonNullableType: (t) => typeType('nonNullable', t),
        typeToString: (t, enclosing, flags) => {
          if (enclosing !== undefined || flags !== undefined)
            unsupported('checker.typeToString formatting arguments');
          return this.query({
            op: 'typeString',
            args: { id: this.id(t, 'type') },
          });
        },
        getSymbolAtLocation: (n) =>
          this.symbol(
            this.query({
              op: 'symbol',
              args: { node: this.syntax.location(n) },
            }),
          ),
        resolveName: (name, location, kind, excludeGlobals) =>
          this.symbol(
            this.query({
              op: 'resolveName',
              args: {
                name: nameToUnits(name),
                node: location ? this.syntax.location(location) : undefined,
                kind,
                excludeGlobals,
              },
            }),
          ),
        getAliasedSymbol: (s) =>
          this.symbol(
            this.query({
              op: 'aliasedSymbol',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
        getTypeOfSymbolAtLocation: (s, n) =>
          this.type(
            this.query({
              op: 'symbolType',
              args: {
                id: this.id(s, 'symbol'),
                node: this.syntax.location(n),
              },
            }),
          ),
        getDeclaredTypeOfSymbol: (s) =>
          this.type(
            this.query({
              op: 'declaredType',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
        getPropertiesOfType: (t) =>
          this.symbolsFor(
            this.query({ op: 'properties', args: { id: this.id(t, 'type') } }),
          ),
        getPropertyOfType: (t, name) =>
          this.symbol(
            this.query({
              op: 'property',
              args: { id: this.id(t, 'type'), name: nameToUnits(name) },
            }),
          ),
        getTypeArguments: (t) =>
          this.typesFor(
            this.query({
              op: 'typeArguments',
              args: { id: this.id(t, 'type') },
            }),
          ),
        isArrayType: (t) =>
          this.query({ op: 'isArray', args: { id: this.id(t, 'type') } }),
        isArrayLikeType: (t) =>
          this.query({ op: 'isArrayLike', args: { id: this.id(t, 'type') } }),
        isTupleType: (t) =>
          this.query({ op: 'isTuple', args: { id: this.id(t, 'type') } }),
        getSignatureFromDeclaration: (n) =>
          this.signature(
            this.query({
              op: 'signature',
              args: { node: this.syntax.location(n) },
            }),
          ),
        getResolvedSignature: (n, candidates, argumentCount) => {
          if (candidates !== undefined || argumentCount !== undefined)
            unsupported('checker.getResolvedSignature optional arguments');
          return this.signature(
            this.query({
              op: 'resolvedSignature',
              args: { node: this.syntax.location(n) },
            }),
          );
        },
        getReturnTypeOfSignature: (s) =>
          this.type(
            this.query({
              op: 'returnType',
              args: { id: this.id(s, 'signature') },
            }),
          ),
        getSignaturesOfType: (t, kind) =>
          this.signaturesFor(
            this.query({
              op: 'signatures',
              args: { id: this.id(t, 'type'), kind },
            }),
          ),
      },
      () => {
        this.ensure();
      },
      'checker',
    );
    const compilerOptions = { ...descriptor.compilerOptions };
    for (const [key, group] of [
      ['jsx', typescript.JsxEmit],
      ['moduleDetection', typescript.ModuleDetectionKind],
    ] as const) {
      const value = compilerOptions[key];
      if (typeof value === 'string') {
        const translated: unknown = Reflect.get(group, value);
        if (typeof translated !== 'number')
          unsupported(`compilerOptions.${key}.${value}`);
        compilerOptions[key] = translated;
      }
    }
    this.program = guarded<ProgramView>(
      {
        getTypeChecker: () => this.checker,
        getCompilerOptions: () => ({ ...compilerOptions }),
        getCurrentDirectory: () => descriptor.currentDirectory,
        getSourceFile: (name) => this.syntax.sourceView(name),
        getSourceFiles: () =>
          descriptor.sources.map((s) => {
            const source = this.syntax.sourceView(s.fileName);
            if (!source) throw new Error('Program source unavailable');
            return source;
          }),
        sourceFileToPackageName: {
          get: (path) => {
            this.ensure();
            return this.syntax.packageName(path);
          },
        },
        isSourceFileFromExternalLibrary: (source) =>
          this.syntax.metadata(source).external,
        isSourceFileDefaultLibrary: (source) =>
          this.syntax.metadata(source).defaultLibrary,
      },
      () => {
        this.ensure();
      },
      'program',
    );
  }
  get syntaxStats() {
    return this.syntax.stats;
  }
  private ensure(): void {
    if (this.closed) throw new Error('type service client is closed');
    this.shared?.assertOpen();
  }
  close(): void {
    this.closed = true;
    for (const reader of this.facts.values()) reader.close();
    this.facts.clear();
    this.stats.factBytes = 0;
    this.types.clear();
    this.symbols.clear();
    this.signatures.clear();
    for (const file of this.files) file.close();
    this.files.clear();
    this.syntax.close();
    this.shared?.close();
    this.shared = undefined;
    this.transport = undefined;
    this.batchTransport = undefined;
    this.profiles.clear();
    this.activeFile = undefined;
  }
  private read<K>(
    request: ServiceRequest,
    read: (packet: TypePacketReader) => K,
  ): K {
    this.ensure();
    const query: TypeServiceQuery = {
      generation: this.descriptor.generation,
      checker: this.descriptor.checker,
      ...request,
    };
    const key = this.key(query);
    let reader = this.facts.get(key);
    if (reader) {
      this.stats.localHits++;
      this.facts.delete(key);
      this.facts.set(key, reader);
    } else {
      const shared = this.shared?.get(key);
      if (shared) this.stats.sharedHits++;
      else this.stats.queries++;
      reader = this.ingest(query, shared ?? this.transport!(query, key));
    }
    try {
      return read(reader);
    } finally {
      if (this.facts.get(key) !== reader) reader.close();
    }
  }
  key(query: TypeServiceQuery): string {
    return typeQueryKey(query, this.fileOrdinals);
  }
  private store(key: string, bytes: Uint8Array): TypePacketReader {
    const existing = this.facts.get(key);
    if (existing) return existing;
    const reader = new TypePacketReader(Uint8Array.from(bytes));
    if (reader.retainedBytes <= this.maxFactsBytes) {
      while (this.stats.factBytes + reader.retainedBytes > this.maxFactsBytes) {
        const oldest = this.facts.keys().next().value!,
          evicted = this.facts.get(oldest)!;
        this.stats.factBytes -= evicted.retainedBytes;
        evicted.close();
        this.facts.delete(oldest);
      }
      this.facts.set(key, reader);
      this.stats.factBytes += reader.retainedBytes;
    }
    return reader;
  }
  private ingest(query: TypeServiceQuery, bytes: Uint8Array): TypePacketReader {
    const reply = typeReplyPackets(bytes);
    for (const record of reply.records) {
      const key = this.key({
        ...query,
        op: record.op,
        args: { id: record.id },
      });
      const reader = this.store(key, record.packet);
      if (this.facts.get(key) !== reader) reader.close();
    }
    return this.store(this.key(query), reply.packet);
  }
  /** Merge known misses only. Cache hits, including negative answers, never
   * cross the bridge. The scalar API remains the authoritative fallback. */
  prefetch(requests: readonly ServiceRequest[]): void {
    this.prepare(requests, false);
  }
  private prepare(
    requests: readonly ServiceRequest[],
    optional: boolean,
  ): void {
    this.ensure();
    if (!this.batchTransport) return;
    const missing = new Map<string, TypeServiceQuery>();
    for (const request of requests) {
      const query = {
        generation: this.descriptor.generation,
        checker: this.descriptor.checker,
        ...request,
      };
      const key = this.key(query);
      if (this.facts.has(key) || missing.has(key)) continue;
      const shared = this.shared?.get(key);
      if (shared) {
        this.stats.sharedHits++;
        const reader = this.ingest(query, shared);
        if (this.facts.get(key) !== reader) reader.close();
      } else missing.set(key, query);
    }
    const all = [...missing.entries()];
    for (let offset = 0; offset < all.length; offset += 256) {
      const batch = all.slice(offset, offset + 256),
        queries = batch.map(([, query]) => query),
        keys = batch.map(([key]) => key);
      let packets: readonly Uint8Array[];
      try {
        packets = this.batchTransport(queries, keys);
      } catch (error) {
        this.ensure();
        if (
          !optional ||
          (error instanceof Error && error.name === 'AbortError')
        )
          throw error;
        return;
      }
      if (packets.length !== batch.length)
        throw new Error('incomplete type service batch');
      this.stats.queries += batch.length;
      packets.forEach((packet, index) => {
        const reader = this.ingest(queries[index]!, packet);
        if (this.facts.get(keys[index]!) !== reader) reader.close();
      });
    }
  }
  private query<Q extends ScalarRequest>(request: Q): QueryResults[Q['op']] {
    const active = this.activeFile;
    if (
      active &&
      'node' in request.args &&
      request.args.node?.file === active.name &&
      [
        'type',
        'fromTypeNode',
        'context',
        'symbol',
        'signature',
        'resolvedSignature',
      ].includes(request.op)
    ) {
      active.queries++;
      const shape = this.syntax.shape(request.args.node);
      if (shape) {
        let operations = active.profile.shapes.get(shape);
        if (!operations && active.profile.shapes.size < 128)
          active.profile.shapes.set(shape, (operations = new Set()));
        operations?.add(request.op as NodeOperation);
      }
    }
    return this.read(request, (packet) => packet.scalar<Q['op']>(request.op));
  }
  private typeField<K extends keyof TypeFact>(
    id: number,
    field: K,
  ): TypeFact[K] {
    return this.read({ op: 'typeRecord', args: { id } }, (packet) =>
      packet.typeField(field),
    );
  }
  private symbolField<K extends keyof SymbolFact>(
    id: number,
    field: K,
  ): SymbolFact[K] {
    return this.read({ op: 'symbolRecord', args: { id } }, (packet) =>
      packet.symbolField(field),
    );
  }
  private signatureField<K extends keyof SignatureFact>(
    id: number,
    field: K,
  ): SignatureFact[K] {
    return this.read({ op: 'signatureRecord', args: { id } }, (packet) =>
      packet.signatureField(field),
    );
  }
  private id(object: object, kind: ObjectKind): number {
    this.ensure();
    const brand = this.brands.get(object);
    if (brand?.kind !== kind) throw new Error(`foreign ${kind} handle`);
    return brand.id;
  }
  private intern<T extends object>(
    map: WeakObjectCache<number, T>,
    kind: ObjectKind,
    id: number,
    create: () => T,
  ): T | undefined {
    this.ensure();
    if (!id) return undefined;
    const old = map.get(id);
    if (old) return old;
    const object = guarded(
      create(),
      () => {
        this.ensure();
      },
      kind,
    );
    this.brands.set(object, { kind, id });
    map.set(id, object);
    return object;
  }
  private typesFor(ids: readonly number[]): TypeView[] {
    return ids.map((id) => this.type(id)!);
  }
  private symbolsFor(ids: readonly number[]): SymbolView[] {
    return ids.map((id) => this.symbol(id)!);
  }
  private signaturesFor(ids: readonly number[]): SignatureView[] {
    return ids.map((id) => this.signature(id)!);
  }
  private type(id: number): TypeView | undefined {
    return this.intern(this.types, 'type', id, () => {
      const ensure = () => {
        this.ensure();
      };
      const flags = memo(ensure, () =>
        this.typeField(id, 'flags').reduce((value, name) => {
          const flag = Reflect.get(this.typescript.TypeFlags, name);
          if (typeof flag !== 'number') unsupported(`TypeFlags.${name}`);
          return value | flag;
        }, 0),
      );
      const objectFlags = memo(ensure, () =>
        this.typeField(id, 'objectFlags')?.reduce((value, name) => {
          const flag = Reflect.get(this.typescript.ObjectFlags, name);
          if (typeof flag !== 'number') unsupported(`ObjectFlags.${name}`);
          return value | flag;
        }, 0),
      );
      const symbol = memo(ensure, () =>
          this.symbol(this.typeField(id, 'symbol')),
        ),
        alias = memo(ensure, () =>
          this.symbol(this.typeField(id, 'aliasSymbol')),
        );
      const intrinsic = memo(ensure, () => this.typeField(id, 'intrinsicName')),
        escapedName = memo(ensure, () => this.typeField(id, 'escapedName')),
        value = memo(ensure, () => this.typeField(id, 'value'));
      const types = memo(ensure, () => {
        const ids = this.typeField(id, 'types');
        return ids ? this.typesFor(ids) : undefined;
      });
      const calls = memo(ensure, () =>
          this.signaturesFor(
            this.query({ op: 'signatures', args: { id, kind: 0 } }),
          ),
        ),
        constructs = memo(ensure, () =>
          this.signaturesFor(
            this.query({ op: 'signatures', args: { id, kind: 1 } }),
          ),
        );
      const properties = memo(ensure, () =>
          this.symbolsFor(this.query({ op: 'properties', args: { id } })),
        ),
        nonNullable = memo(ensure, () =>
          this.type(this.query({ op: 'nonNullable', args: { id } })),
        ),
        constraint = memo(ensure, () =>
          this.type(this.query({ op: 'constraint', args: { id } })),
        ),
        numberIndexType = memo(ensure, () =>
          this.type(this.query({ op: 'numberIndexType', args: { id } })),
        );
      const flag = (mask: number) => Boolean(flags() & mask),
        tf = this.typescript.TypeFlags;
      return {
        id,
        get flags() {
          return flags();
        },
        get symbol() {
          return symbol();
        },
        get aliasSymbol() {
          return alias();
        },
        get intrinsicName() {
          return intrinsic();
        },
        get escapedName() {
          return escapedName();
        },
        get objectFlags() {
          return objectFlags();
        },
        get value() {
          return value();
        },
        get types() {
          return types();
        },
        getSymbol: symbol,
        isUnion: () => flag(tf.Union),
        isIntersection: () => flag(tf.Intersection),
        isUnionOrIntersection: () => flag(tf.UnionOrIntersection),
        isLiteral: () =>
          flag(tf.StringLiteral | tf.NumberLiteral | tf.BigIntLiteral),
        isStringLiteral: () => flag(tf.StringLiteral),
        isNumberLiteral: () => flag(tf.NumberLiteral),
        isTypeParameter: () => flag(tf.TypeParameter),
        getCallSignatures: calls,
        getConstructSignatures: constructs,
        getProperties: properties,
        getProperty: (name) =>
          this.symbol(
            this.query({
              op: 'property',
              args: { id, name: nameToUnits(name) },
            }),
          ),
        getNonNullableType: nonNullable,
        getConstraint: constraint,
        getNumberIndexType: numberIndexType,
      };
    });
  }
  private symbol(id: number): SymbolView | undefined {
    return this.intern(this.symbols, 'symbol', id, () => {
      const ensure = () => {
        this.ensure();
      };
      const name = memo(ensure, () => this.symbolField(id, 'name')),
        escaped = memo(ensure, () => this.symbolField(id, 'escapedName')),
        flags = memo(ensure, () => this.symbolField(id, 'flags'));
      const declarations = memo(ensure, () =>
        this.symbolField(id, 'declarations')?.map((node) =>
          this.syntax.nodeRef(node)!,
        ),
      );
      const value = memo(ensure, () =>
        this.syntax.nodeRef(this.symbolField(id, 'valueDeclaration')),
      );
      return {
        id,
        get name() {
          return name();
        },
        get escapedName() {
          return escaped();
        },
        get flags() {
          return flags();
        },
        get declarations() {
          return declarations();
        },
        get valueDeclaration() {
          return value();
        },
        getName: name,
        getFlags: flags,
        getDeclarations: declarations,
      };
    });
  }
  private signature(id: number): SignatureView | undefined {
    return this.intern(this.signatures, 'signature', id, () => {
      const ensure = () => {
          this.ensure();
        },
        parameters = memo(ensure, () =>
          this.symbolsFor(this.signatureField(id, 'parameters')),
        ),
        declaration = memo(ensure, () =>
          this.syntax.nodeRef(this.signatureField(id, 'declaration')),
        ),
        returned = memo(ensure, () =>
          this.type(this.query({ op: 'returnType', args: { id } })),
        );
      return {
        id,
        get parameters() {
          return parameters();
        },
        get declaration() {
          return declaration();
        },
        getParameters: parameters,
        getDeclaration: declaration,
        getReturnType: returned,
      };
    });
  }
  file(binding: FileBinding, profileKey?: string): TypeServiceFile {
    this.ensure();
    if (
      binding.generation !== this.descriptor.generation ||
      binding.checker !== this.descriptor.checker
    )
      throw new Error('foreign file binding');
    const file = new TypeServiceFile(
      this,
      binding,
      new FileSyntaxViews(this.syntax, binding, () => this.ensure()),
      profileKey,
    );
    this.files.add(file);
    return file;
  }
  prepareFile(
    file: TypeServiceFile,
    name: string,
    profileKey: string | undefined,
  ): void {
    if (!profileKey || !this.batchTransport) return;
    let profile = this.profiles.get(profileKey);
    if (!profile) {
      if (this.profiles.size >= 64)
        this.profiles.delete(this.profiles.keys().next().value!);
      this.profiles.set(
        profileKey,
        (profile = { shapes: new Map(), maxQueries: 0 }),
      );
    }
    this.activeFile = { file, name, profile, queries: 0 };
    if (profile.maxQueries < 16) return;
    const cap = Math.max(256, Math.min(4096, profile.maxQueries * 4));
    const requests = this.syntax.matching(name, profile.shapes, cap);
    if (requests.length > cap) return;
    this.prepare(
      requests.map(({ op, node }) => ({
        op: op as NodeOperation,
        args: { node },
      })),
      true,
    );
  }
  releaseFile(file: TypeServiceFile): void {
    this.files.delete(file);
    if (this.activeFile?.file === file) {
      this.activeFile.profile.maxQueries = Math.max(
        this.activeFile.profile.maxQueries,
        this.activeFile.queries,
      );
      this.activeFile = undefined;
    }
  }
}

export class TypeServiceFile {
  private closed = false;
  constructor(
    private readonly owner: TypeServiceClient,
    readonly binding: FileBinding,
    private readonly syntax: FileSyntaxViews,
    private readonly profileKey?: string,
  ) {}
  close(): void {
    if (this.closed) return;
    this.closed = true;
    this.syntax.close();
    this.owner.releaseFile(this);
  }
  parserServices(
    frame: DecodedFrame,
    projection: TypeServiceProjection,
    sourceFileName: string,
  ) {
    if (this.closed) throw new Error('file type services are closed');
    this.syntax.attach(frame, projection, sourceFileName);
    this.owner.prepareFile(this, sourceFileName, this.profileKey);
    const toTS = (node: ESTreeNode): object => {
      const native = this.syntax.esTreeNodeToTSNodeMap.get(node);
      if (!native)
        throw new Error(`ESTree node has no compiler origin: ${node.type}`);
      return native;
    };
    const options = this.owner.descriptor.compilerOptions;
    return {
      program: this.owner.program,
      esTreeNodeToTSNodeMap: this.syntax.esTreeNodeToTSNodeMap,
      tsNodeToESTreeNodeMap: this.syntax.tsNodeToESTreeNodeMap,
      getTypeAtLocation: (node: ESTreeNode) =>
        this.owner.checker.getTypeAtLocation(toTS(node)),
      getSymbolAtLocation: (node: ESTreeNode) =>
        this.owner.checker.getSymbolAtLocation(toTS(node)),
      emitDecoratorMetadata: options.emitDecoratorMetadata === true,
      experimentalDecorators: options.experimentalDecorators === true,
      isolatedDeclarations: options.isolatedDeclarations === true,
    };
  }
}

/** Compatibility convenience for a standalone invocation. Worker pools should
 * retain TypeServiceClient and create/close file handles for each lint task. */
export class FileTypeServiceClient extends TypeServiceClient {
  private readonly handle: TypeServiceFile;
  constructor(
    typescript: TypeScriptEnums,
    descriptor: TypeServiceDescriptor,
    binding: FileBinding,
    transport: QueryTransport,
    syntax: SyntaxSource,
    options: TypeServiceClientOptions,
  ) {
    super(typescript, descriptor, transport, syntax, options);
    this.handle = this.file(binding);
  }
  parserServices(
    frame: DecodedFrame,
    projection: TypeServiceProjection,
    sourceFileName: string,
  ) {
    return this.handle.parserServices(frame, projection, sourceFileName);
  }
}
