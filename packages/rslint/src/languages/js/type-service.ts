import type ts from 'typescript';
import type { DecodedFrame } from './frame.js';
import type { ESTreeNode } from './types.js';
import type { SharedTypeCacheReader } from './shared-type-cache.js';
import { TypePacketReader } from './type-service-codec.js';
import { FileSyntaxViews } from './type-service-syntax.js';
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
type ObjectKind = 'type' | 'symbol' | 'signature';
type TypeScriptEnums = Pick<
  typeof ts,
  'SyntaxKind' | 'TypeFlags' | 'NodeFlags' | 'JsxEmit' | 'ModuleDetectionKind'
>;

/** Exactly one eligible file-lint invocation owns this client. Close it in
 * finally after the rule run. Stable wrappers/requested arrays live until close;
 * maxFactsBytes bounds disposable packet readers, not plugin-retained objects or
 * independently owned syntax views. The Go Session may outlive many clients. */
export class FileTypeServiceClient {
  private closed = false;
  private readonly facts = new Map<string, TypePacketReader>();
  private readonly types = new Map<number, TypeView>();
  private readonly symbols = new Map<number, SymbolView>();
  private readonly signatures = new Map<number, SignatureView>();
  private readonly brands = new WeakMap<
    object,
    { kind: ObjectKind; id: number }
  >();
  private shared: SharedTypeCacheReader | undefined;
  private readonly maxFactsBytes: number;
  private readonly syntax: FileSyntaxViews;
  readonly stats = { queries: 0, localHits: 0, sharedHits: 0, factBytes: 0 };
  readonly program: ProgramView;
  readonly checker: CheckerView;
  constructor(
    private readonly typescript: TypeScriptEnums,
    readonly descriptor: TypeServiceDescriptor,
    readonly binding: FileBinding,
    private transport: QueryTransport | undefined,
    syntax: SyntaxSource,
    options: { maxFactsBytes: number; shared?: SharedTypeCacheReader },
  ) {
    if (
      binding.generation !== descriptor.generation ||
      binding.checker !== descriptor.checker
    )
      throw new Error('foreign file binding');
    if (
      !Number.isSafeInteger(options.maxFactsBytes) ||
      options.maxFactsBytes < 0
    )
      throw new Error('invalid type service client budget');
    this.maxFactsBytes = options.maxFactsBytes;
    this.shared = options.shared;
    if (this.shared && this.shared.descriptor.generation !== binding.generation)
      throw new Error('foreign shared type cache');
    this.syntax = new FileSyntaxViews(
      typescript,
      descriptor,
      binding,
      syntax,
      () => {
        this.ensure();
      },
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
            return descriptor.sources.find((s) => s.path === path)
              ?.packages?.[0];
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
    this.syntax.close();
    this.shared?.close();
    this.shared = undefined;
    this.transport = undefined;
  }
  private read<K>(
    request: ServiceRequest,
    read: (packet: TypePacketReader) => K,
  ): K {
    this.ensure();
    const query: TypeServiceQuery = {
      generation: this.binding.generation,
      checker: this.binding.checker,
      ...request,
    };
    const key = typeQueryKey(query);
    let reader = this.facts.get(key),
      retained = true;
    if (reader) {
      this.stats.localHits++;
      this.facts.delete(key);
      this.facts.set(key, reader);
    } else {
      let bytes = this.shared?.get(key);
      if (bytes) this.stats.sharedHits++;
      else {
        this.stats.queries++;
        bytes = Uint8Array.from(this.transport!(query, key));
      }
      reader = new TypePacketReader(bytes);
      retained = reader.retainedBytes <= this.maxFactsBytes;
      if (retained) {
        while (
          this.stats.factBytes + reader.retainedBytes >
          this.maxFactsBytes
        ) {
          const oldest = this.facts.keys().next().value!;
          const evicted = this.facts.get(oldest)!;
          this.stats.factBytes -= evicted.retainedBytes;
          evicted.close();
          this.facts.delete(oldest);
        }
        this.facts.set(key, reader);
        this.stats.factBytes += reader.retainedBytes;
      }
    }
    try {
      return read(reader);
    } finally {
      if (!retained) reader.close();
    }
  }
  private query<Q extends ScalarRequest>(request: Q): QueryResults[Q['op']] {
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
    map: Map<number, T>,
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
      const symbol = memo(ensure, () =>
          this.symbol(this.typeField(id, 'symbol')),
        ),
        alias = memo(ensure, () =>
          this.symbol(this.typeField(id, 'aliasSymbol')),
        );
      const intrinsic = memo(ensure, () => this.typeField(id, 'intrinsicName')),
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
  parserServices(
    frame: DecodedFrame,
    projection: TypeServiceProjection,
    sourceFileName: string,
  ) {
    this.syntax.attach(frame, projection, sourceFileName);
    const toTS = (node: ESTreeNode): object => {
      const native = this.syntax.esTreeNodeToTSNodeMap.get(node);
      if (!native)
        throw new Error(`ESTree node has no compiler origin: ${node.type}`);
      return native;
    };
    const options = this.descriptor.compilerOptions;
    return {
      program: this.program,
      esTreeNodeToTSNodeMap: this.syntax.esTreeNodeToTSNodeMap,
      tsNodeToESTreeNodeMap: this.syntax.tsNodeToESTreeNodeMap,
      getTypeAtLocation: (node: ESTreeNode) =>
        this.checker.getTypeAtLocation(toTS(node)),
      getSymbolAtLocation: (node: ESTreeNode) =>
        this.checker.getSymbolAtLocation(toTS(node)),
      emitDecoratorMetadata: options.emitDecoratorMetadata === true,
      experimentalDecorators: options.experimentalDecorators === true,
      isolatedDeclarations: options.isolatedDeclarations === true,
    };
  }
}
