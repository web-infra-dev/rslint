/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
import {
  RemoteSourceFile,
  RemoteNode,
  SyntaxKind,
  NodeFlags,
  Wtf8Decoder,
} from '@rslint/api';
import type ts from 'typescript';
import type { DecodedFrame } from './frame.js';
import type { ESTreeNode } from './types.js';
import { decodeTypePacket } from './type-service-codec.js';
import type { SharedTypeCacheReader } from './shared-type-cache.js';

type Bag = Record<string, unknown>;
export interface TypeServiceDescriptor {
  generation: number;
  checker: number;
  currentDirectory: string;
  compilerOptions: Bag;
  sources: Array<{
    fileName: string;
    path: string;
    packages?: string[];
    external: boolean;
    defaultLibrary: boolean;
  }>;
}
export interface TypeServiceProjection {
  sourceFileName: string;
  tree: Uint8Array;
  origins: number[];
  reverse: Record<string, number>;
  generation: number;
}
export interface TypeServiceQuery {
  generation: number;
  checker: number;
  op: string;
  id?: number;
  file?: string;
  index?: number;
  name?: string;
  kind?: number;
  excludeGlobals?: boolean;
}
export function typeQueryKey(q: TypeServiceQuery): string {
  return JSON.stringify([
    q.generation,
    q.checker,
    q.op,
    q.id ?? 0,
    q.file ?? '',
    q.index ?? 0,
    q.name ?? '',
    q.kind ?? 0,
    q.excludeGlobals ?? false,
  ]);
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
      const value = Reflect.get(target, key, receiver);
      return typeof value === 'function'
        ? (...args: unknown[]) => {
            ensure();
            return Reflect.apply(value, receiver, args);
          }
        : value;
    },
  });
}
function translateFlags(flags: number, from: object, to: object): number {
  let result = 0;
  for (const [name, value] of Object.entries(from)) {
    if (
      typeof value !== 'number' ||
      value <= 0 ||
      (value & (value - 1)) !== 0 ||
      !(flags & value)
    )
      continue;
    const mapped = (to as Record<string, unknown>)[name];
    if (typeof mapped !== 'number') unsupported(`enum.${name}`);
    result |= mapped;
  }
  return result;
}

/** One client owns stable wrapper identity and bounded local facts. The host
 * injects a synchronous transport and owns Session/cache lifetime. Closing a
 * client invalidates retained wrappers; closing shared storage invalidates all
 * attached clients. No transport, native binding or scheduler is imported here.
 */
export class TypeServiceClient {
  private closed = false;
  private readonly facts = new Map<string, unknown>();
  private readonly objects = new Map<string, Bag>();
  private rawNodes = new WeakMap<object, RemoteNode>();
  private wrappedNodes = new WeakMap<RemoteNode, object>();
  private readonly sources = new Map<string, RemoteSourceFile>();
  private readonly brands = new WeakMap<object, { kind: string; id: number }>();
  readonly stats = { queries: 0, localHits: 0, sharedHits: 0 };
  readonly program: Bag;
  readonly checker: Bag;
  constructor(
    private readonly typescript: Pick<
      typeof ts,
      | 'SyntaxKind'
      | 'TypeFlags'
      | 'NodeFlags'
      | 'JsxEmit'
      | 'ModuleDetectionKind'
    >,
    readonly descriptor: TypeServiceDescriptor,
    private transport:
      ((query: TypeServiceQuery, key: string) => Uint8Array) | undefined,
    private readonly options: {
      shared?: SharedTypeCacheReader;
      maxFacts: number;
    },
  ) {
    this.options = { ...options };
    if (!Number.isSafeInteger(options.maxFacts) || options.maxFacts < 0)
      throw new Error('invalid type service client budget');
    if (
      options.shared &&
      options.shared.descriptor.generation !== descriptor.generation
    )
      throw new Error('foreign shared type cache');
    const nodeType = (op: string, node: object): unknown =>
      this.type(this.query(op, this.location(node)) as number);
    const typeType = (op: string, type: object): unknown =>
      this.type(this.query(op, { id: this.id(type, 'type') }) as number);
    this.checker = guarded(
      {
        getTypeAtLocation: (n: object) => nodeType('type', n),
        getTypeFromTypeNode: (n: object) => nodeType('fromTypeNode', n),
        getContextualType: (n: object) => nodeType('context', n),
        getVoidType: () => this.type(this.query('void') as number),
        getBaseConstraintOfType: (t: object) => typeType('constraint', t),
        getApparentType: (t: object) => typeType('apparent', t),
        getBaseTypeOfLiteralType: (t: object) => typeType('baseLiteral', t),
        getNonNullableType: (t: object) => typeType('nonNullable', t),
        typeToString: (t: object, enclosing?: object, flags?: number) => {
          if (enclosing !== undefined || flags !== undefined)
            unsupported('checker.typeToString formatting arguments');
          return this.query('typeString', { id: this.id(t, 'type') });
        },
        getSymbolAtLocation: (n: object) =>
          this.symbol(this.query('symbol', this.location(n)) as number),
        resolveName: (
          name: string,
          location: object | undefined,
          meaning: number,
          excludeGlobals: boolean,
        ) =>
          this.symbol(
            this.query('resolveName', {
              name,
              ...(location ? this.location(location) : {}),
              kind: meaning,
              excludeGlobals,
            }) as number,
          ),
        getAliasedSymbol: (s: object) =>
          this.symbol(
            this.query('aliasedSymbol', { id: this.id(s, 'symbol') }) as number,
          ),
        getTypeOfSymbolAtLocation: (s: object, n: object) =>
          this.type(
            this.query('symbolType', {
              id: this.id(s, 'symbol'),
              ...this.location(n),
            }) as number,
          ),
        getDeclaredTypeOfSymbol: (s: object) =>
          this.type(
            this.query('declaredType', { id: this.id(s, 'symbol') }) as number,
          ),
        getPropertiesOfType: (t: object) =>
          this.symbols(this.query('properties', { id: this.id(t, 'type') })),
        getPropertyOfType: (t: object, name: string) =>
          this.symbol(
            this.query('property', { id: this.id(t, 'type'), name }) as number,
          ),
        getTypeArguments: (t: object) =>
          this.types(this.query('typeArguments', { id: this.id(t, 'type') })),
        isArrayType: (t: object) =>
          this.query('isArray', { id: this.id(t, 'type') }),
        isTupleType: (t: object) =>
          this.query('isTuple', { id: this.id(t, 'type') }),
        getSignatureFromDeclaration: (n: object) =>
          this.signature(this.query('signature', this.location(n)) as number),
        getResolvedSignature: (
          n: object,
          candidates?: unknown,
          argumentCount?: unknown,
        ) => {
          if (candidates !== undefined || argumentCount !== undefined)
            unsupported('checker.getResolvedSignature optional arguments');
          return this.signature(
            this.query('resolvedSignature', this.location(n)) as number,
          );
        },
        getReturnTypeOfSignature: (s: object) =>
          this.type(
            this.query('returnType', { id: this.id(s, 'signature') }) as number,
          ),
        getSignaturesOfType: (t: object, kind: number) =>
          this.signatures(
            this.query('signatures', { id: this.id(t, 'type'), kind }),
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
        const translated = group[value as keyof typeof group];
        if (typeof translated !== 'number')
          unsupported(`compilerOptions.${key}.${value}`);
        compilerOptions[key] = translated;
      }
    }
    this.program = guarded(
      {
        getTypeChecker: () => this.checker,
        getCompilerOptions: () => ({ ...compilerOptions }),
        getCurrentDirectory: () => descriptor.currentDirectory,
        getSourceFile: (name: string) => this.wrap(this.source(name)),
        getSourceFiles: () =>
          descriptor.sources.map((s) => this.wrap(this.source(s.fileName))),
        sourceFileToPackageName: {
          get: (path: string) => {
            this.ensure();
            return descriptor.sources.find((s) => s.path === path)
              ?.packages?.[0];
          },
        },
        isSourceFileFromExternalLibrary: (source: object) =>
          this.sourceMetadata(source).external,
        isSourceFileDefaultLibrary: (source: object) =>
          this.sourceMetadata(source).defaultLibrary,
      },
      () => {
        this.ensure();
      },
      'program',
    );
  }
  private ensure(): void {
    if (this.closed) throw new Error('type service client is closed');
    this.options.shared?.assertOpen();
  }
  close(): void {
    this.closed = true;
    this.facts.clear();
    this.objects.clear();
    this.sources.clear();
    this.options.shared?.close();
    this.options.shared = undefined;
    this.transport = undefined;
    this.rawNodes = new WeakMap();
    this.wrappedNodes = new WeakMap();
  }
  private query(op: string, args: Partial<TypeServiceQuery> = {}): unknown {
    this.ensure();
    const q: TypeServiceQuery = {
      generation: this.descriptor.generation,
      checker: this.descriptor.checker,
      op,
      ...args,
    };
    const key = typeQueryKey(q);
    if (this.facts.has(key)) {
      this.stats.localHits++;
      return this.facts.get(key);
    }
    let bytes = this.options.shared?.get(key);
    if (bytes) this.stats.sharedHits++;
    else {
      this.stats.queries++;
      bytes = this.transport!(q, key);
    }
    const value = decodeTypePacket(bytes);
    if (this.options.maxFacts > 0) {
      if (this.facts.size === this.options.maxFacts)
        this.facts.delete(this.facts.keys().next().value!);
      this.facts.set(key, value);
    }
    return value;
  }
  private id(object: object, kind: string): number {
    this.ensure();
    const brand = this.brands.get(object);
    if (brand?.kind !== kind) throw new Error(`foreign ${kind} handle`);
    return brand.id;
  }
  private location(node: object): { file: string; index: number } {
    this.ensure();
    const raw = this.rawNodes.get(node);
    if (!raw) throw new Error('foreign compiler node handle');
    return {
      file: raw.getSourceFile().fileName,
      index: (raw as unknown as { index: number }).index,
    };
  }
  private sourceMetadata(source: object) {
    const { file } = this.location(source);
    const metadata = this.descriptor.sources.find((s) => s.fileName === file);
    if (!metadata) throw new Error('foreign source');
    return metadata;
  }
  private source(file: string): RemoteSourceFile | undefined {
    let source = this.sources.get(file);
    if (!source) {
      const record = this.query('source', { file }) as
        { tree: Uint8Array; fileName: string } | undefined;
      if (!record) return undefined;
      const existing = this.sources.get(record.fileName);
      if (existing) {
        this.sources.set(file, existing);
        return existing;
      }
      source = new RemoteSourceFile(
        record.tree,
        new Wtf8Decoder('utf-8', { ignoreBOM: true }),
      );
      this.sources.set(file, source);
      this.sources.set(record.fileName, source);
    }
    return source;
  }
  private wrap(value: unknown): unknown {
    if (Array.isArray(value)) {
      const list = value.map((v) => this.wrap(v));
      for (const key of ['pos', 'end', 'hasTrailingComma'])
        if (key in value)
          Object.defineProperty(list, key, {
            value: (value as unknown as Bag)[key],
            enumerable: true,
          });
      return list;
    }
    if (!(value instanceof RemoteNode)) return value;
    const old = this.wrappedNodes.get(value);
    if (old) return old;
    // tsgo represents type-only heritage as TypeReference/QualifiedName.
    // Expose TS's expression view while keeping the original compiler handle.
    const heritageReference =
      value.kind === SyntaxKind.TypeReference &&
      value.parent?.kind === SyntaxKind.HeritageClause;
    let nameOwner = value;
    while (
      nameOwner.kind === SyntaxKind.QualifiedName &&
      nameOwner.parent?.kind === SyntaxKind.QualifiedName
    )
      nameOwner = nameOwner.parent as RemoteNode;
    const heritageName =
      value.kind === SyntaxKind.QualifiedName &&
      nameOwner.parent?.kind === SyntaxKind.TypeReference &&
      nameOwner.parent.parent?.kind === SyntaxKind.HeritageClause;
    const wrapped = new Proxy(value, {
      get: (raw, key) => {
        this.ensure();
        if (key === 'kind' && heritageReference)
          return this.typescript.SyntaxKind.ExpressionWithTypeArguments;
        if (key === 'kind' && heritageName)
          return this.typescript.SyntaxKind.PropertyAccessExpression;
        if (key === 'expression' && heritageReference)
          return this.wrap(Reflect.get(raw, 'typeName', raw));
        if (heritageName && (key === 'expression' || key === 'name'))
          return this.wrap(
            Reflect.get(raw, key === 'expression' ? 'left' : 'right', raw),
          );
        if (
          key === 'kind' ||
          key === 'operator' ||
          key === 'token' ||
          key === 'keywordToken' ||
          key === 'keyword' ||
          key === 'phaseModifier'
        ) {
          const native = Reflect.get(raw, key, raw) as number;
          if (native === undefined) return undefined;
          const name = SyntaxKind[native];
          const kind =
            this.typescript.SyntaxKind[name as keyof typeof ts.SyntaxKind];
          if (typeof kind !== 'number') unsupported(`SyntaxKind.${name}`);
          return kind;
        }
        if (key === 'flags')
          return translateFlags(
            raw.flags,
            NodeFlags,
            this.typescript.NodeFlags,
          );
        if (
          key === 'escapedText' &&
          (raw.kind === SyntaxKind.Identifier ||
            raw.kind === SyntaxKind.PrivateIdentifier)
        ) {
          const text = (raw as unknown as { text: string }).text;
          return text.startsWith('__') ? `_${text}` : text;
        }
        if (key === 'forEachChild')
          return (
            cb: (node: unknown) => unknown,
            cbNodes?: (nodes: unknown) => unknown,
          ) => {
            this.ensure();
            return raw.forEachChild(
              (n) => cb(this.wrap(n)),
              cbNodes ? (nodes) => cbNodes(this.wrap(nodes)) : undefined,
            );
          };
        if (
          typeof key === 'string' &&
          [
            'getChildren',
            'getFirstToken',
            'getLastToken',
            'getChildAt',
            'getChildCount',
          ].includes(key)
        )
          unsupported(`node.${key}`);
        const member: unknown = Reflect.get(raw, key, raw);
        if (typeof member === 'function')
          return (...args: unknown[]) => {
            this.ensure();
            return this.wrap(
              Reflect.apply(
                member,
                raw,
                args.map((a) =>
                  typeof a === 'object' && a !== null
                    ? (this.rawNodes.get(a) ?? a)
                    : a,
                ),
              ),
            );
          };
        if (
          member === undefined &&
          typeof key === 'string' &&
          /^(get|is|forEach)/.test(key)
        )
          unsupported(`node.${key}`);
        return this.wrap(member);
      },
    });
    this.rawNodes.set(wrapped, value);
    this.wrappedNodes.set(value, wrapped);
    return wrapped;
  }
  private nodeRef(value: unknown): unknown {
    if (!value) return undefined;
    const ref = value as { file: string; index: number };
    return this.wrap(this.source(ref.file)!.getOrCreateNodeAtIndex(ref.index));
  }
  private types(value: unknown): unknown[] {
    return (value as number[]).map((id) => this.type(id));
  }
  private symbols(value: unknown): unknown[] {
    return (value as number[]).map((id) => this.symbol(id));
  }
  private signatures(value: unknown): unknown[] {
    return (value as number[]).map((id) => this.signature(id));
  }
  private object(kind: string, id: number, create: () => Bag): Bag | undefined {
    this.ensure();
    if (!id) return undefined;
    const key = `${kind}:${id}`,
      old = this.objects.get(key);
    if (old) return old;
    const object = guarded(
      create(),
      () => {
        this.ensure();
      },
      kind,
    );
    this.brands.set(object, { kind, id });
    this.objects.set(key, object);
    return object;
  }
  private type(id: number): Bag | undefined {
    return this.object('type', id, () => {
      const record = (): Bag => this.query('typeRecord', { id }) as Bag;
      const flags = (): number =>
        (record().flags as string[]).reduce((value, name) => {
          const translated =
            this.typescript.TypeFlags[name as keyof typeof ts.TypeFlags];
          if (typeof translated !== 'number') unsupported(`TypeFlags.${name}`);
          return value | (translated as number);
        }, 0);
      const object: Bag = {
        id,
        get flags() {
          return flags();
        },
        getSymbol: () => this.symbol(record().symbol as number),
        isUnion: () => Boolean(flags() & this.typescript.TypeFlags.Union),
        isIntersection: () =>
          Boolean(flags() & this.typescript.TypeFlags.Intersection),
        isUnionOrIntersection: () =>
          Boolean(flags() & this.typescript.TypeFlags.UnionOrIntersection),
        isLiteral: () =>
          Boolean(
            flags() &
            (this.typescript.TypeFlags.StringLiteral |
              this.typescript.TypeFlags.NumberLiteral |
              this.typescript.TypeFlags.BigIntLiteral),
          ),
        isStringLiteral: () =>
          Boolean(flags() & this.typescript.TypeFlags.StringLiteral),
        isNumberLiteral: () =>
          Boolean(flags() & this.typescript.TypeFlags.NumberLiteral),
        isTypeParameter: () =>
          Boolean(flags() & this.typescript.TypeFlags.TypeParameter),
        getCallSignatures: () =>
          this.signatures(this.query('signatures', { id, kind: 0 })),
        getConstructSignatures: () =>
          this.signatures(this.query('signatures', { id, kind: 1 })),
        getProperties: () => this.symbols(this.query('properties', { id })),
        getProperty: (name: string) =>
          this.symbol(this.query('property', { id, name }) as number),
        getNonNullableType: () =>
          this.type(this.query('nonNullable', { id }) as number),
        getConstraint: () =>
          this.type(this.query('constraint', { id }) as number),
      };
      for (const key of ['intrinsicName', 'value'])
        Object.defineProperty(object, key, { get: () => record()[key] });
      for (const key of ['symbol', 'aliasSymbol'])
        Object.defineProperty(object, key, {
          get: () => this.symbol(record()[key] as number),
        });
      Object.defineProperty(object, 'types', {
        get: () => (record().types ? this.types(record().types) : undefined),
      });
      return object;
    });
  }
  private symbol(id: number): Bag | undefined {
    return this.object('symbol', id, () => {
      const record = (): Bag => this.query('symbolRecord', { id }) as Bag;
      const escapedName = (): string => record().escapedName as string;
      const name = (): string => record().name as string;
      const declarations = () =>
        (record().declarations as unknown[] | undefined)?.map((value) =>
          this.nodeRef(value),
        );
      const valueDeclaration = () => this.nodeRef(record().valueDeclaration);
      return {
        id,
        get name() {
          return name();
        },
        get escapedName() {
          return escapedName();
        },
        get flags() {
          return record().flags;
        },
        get declarations() {
          return declarations();
        },
        get valueDeclaration() {
          return valueDeclaration();
        },
        getName: name,
        getDeclarations: declarations,
        getFlags: () => record().flags,
      };
    });
  }
  private signature(id: number): Bag | undefined {
    return this.object('signature', id, () => {
      const record = (): Bag => this.query('signatureRecord', { id }) as Bag;
      const parameters = () => this.symbols(record().parameters);
      const declaration = () => this.nodeRef(record().declaration);
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
        getReturnType: () =>
          this.type(this.query('returnType', { id }) as number),
      };
    });
  }
  parserServices(
    frame: DecodedFrame,
    projection: TypeServiceProjection,
    sourceFileName: string,
  ): Bag {
    this.ensure();
    if (
      projection.generation !== this.descriptor.generation ||
      projection.origins.length !== frame.nodes.length
    )
      throw new Error('stale or incompatible projection origins');
    const source = new RemoteSourceFile(
      projection.tree,
      new Wtf8Decoder('utf-8', { ignoreBOM: true }),
    );
    const file = source.fileName;
    if (
      file !== projection.sourceFileName ||
      file !== sourceFileName ||
      source.text !== frame.text
    )
      throw new Error('type service projection belongs to a different source');
    const existing = this.sources.get(file);
    if (!existing) this.sources.set(file, source);
    const original = existing ?? source;
    const forward = new WeakMap<ESTreeNode, object>(),
      reverse = new WeakMap<object, ESTreeNode>();
    for (let id = 1; id < frame.nodes.length; id++) {
      const index = projection.origins[id];
      if (index)
        forward.set(
          frame.nodes[id]!,
          this.wrap(original.getOrCreateNodeAtIndex(index)) as object,
        );
    }
    for (const [index, id] of Object.entries(projection.reverse))
      reverse.set(
        this.wrap(original.getOrCreateNodeAtIndex(Number(index))) as object,
        frame.nodes[id]!,
      );
    const map = <K extends object, V>(values: WeakMap<K, V>) => ({
      get: (key: K): V | undefined => {
        this.ensure();
        return values.get(key);
      },
      has: (key: K): boolean => {
        this.ensure();
        return values.has(key);
      },
    });
    const toTS = (node: ESTreeNode): object => {
      this.ensure();
      const tsNode = forward.get(node);
      if (!tsNode)
        throw new Error(`ESTree node has no compiler origin: ${node.type}`);
      return tsNode;
    };
    const options = this.descriptor.compilerOptions;
    return {
      program: this.program,
      esTreeNodeToTSNodeMap: map(forward),
      tsNodeToESTreeNodeMap: map(reverse),
      getTypeAtLocation: (node: ESTreeNode) =>
        (this.checker.getTypeAtLocation as (n: object) => unknown)(toTS(node)),
      getSymbolAtLocation: (node: ESTreeNode) =>
        (this.checker.getSymbolAtLocation as (n: object) => unknown)(
          toTS(node),
        ),
      emitDecoratorMetadata: options.emitDecoratorMetadata === true,
      experimentalDecorators: options.experimentalDecorators === true,
      isolatedDeclarations: options.isolatedDeclarations === true,
    };
  }
}
