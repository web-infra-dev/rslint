/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
import {
  RemoteSourceFile,
  RemoteNode,
  SyntaxKind,
  NodeFlags,
  Wtf8Decoder,
} from '../../../compiler-syntax/index.js';
import type ts from 'typescript';
import { WeakObjectCache } from './type-service-cache.js';
import type {
  FileBinding,
  TypeServiceDescriptor,
  TypeServiceProjection,
  SyntaxSource,
  NodeRef,
  NodeFact,
} from './type-service-protocol.js';
import type { DecodedFrame } from './frame.js';
import type { ESTreeNode } from './types.js';
import { offsetToLineColumn } from './positions.js';
type Bag = Record<string, unknown>;
function unsupported(name: string): never {
  throw new Error(`unsupported TypeScript operation: ${name}`);
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
    const translated = (to as Record<string, unknown>)[name];
    if (typeof translated !== 'number') unsupported(`enum.${name}`);
    result |= translated;
  }
  return result;
}

/** A worker/generation owns stable native identities and a bounded decoded
 * source LRU. Proxies retain identities, never a dependency source buffer. */
export class GenerationSyntaxViews {
  private readonly identities = new WeakMap<object, NodeRef>();
  private readonly wrapped = new WeakObjectCache<string, object>();
  private readonly sources = new Map<string, RemoteSourceFile>();
  private readonly summaries = new Map<
    string,
    { fact: NodeFact; bytes: number }
  >();
  private readonly sourceOrdinals = new Map<string, number>();
  private readonly metadataByName = new Map<
    string,
    TypeServiceDescriptor['sources'][number]
  >();
  readonly stats = {
    nodes: 0,
    aliases: 0,
    sourceBytes: 0,
    sourceRequests: 0,
    summaryBytes: 0,
  };
  constructor(
    private readonly typescript: Pick<typeof ts, 'SyntaxKind' | 'NodeFlags'>,
    private readonly descriptor: TypeServiceDescriptor,
    private lookup: SyntaxSource | undefined,
    private readonly readNodes: (node: NodeRef) => readonly NodeFact[],
    private readonly ensure: () => void,
    private readonly maxSourceBytes: number,
    private readonly maxSummaryBytes: number,
  ) {
    for (const [index, source] of descriptor.sources.entries()) {
      this.sourceOrdinals.set(source.fileName, index + 1);
      this.metadataByName.set(source.fileName, source);
      this.metadataByName.set(source.path, source);
    }
  }
  private key(ref: NodeRef): string {
    return `${this.sourceOrdinals.get(ref.file)}:${ref.index}`;
  }
  locationOrUndefined(node: object): NodeRef | undefined {
    return this.identities.get(node);
  }
  location(node: object): NodeRef {
    this.ensure();
    const ref = this.identities.get(node);
    if (!ref) throw new Error('foreign compiler node handle');
    return ref;
  }
  metadata(node: object) {
    const metadata = this.metadataByName.get(this.location(node).file);
    if (!metadata) throw new Error('foreign source');
    return metadata;
  }
  packageName(path: string): string | undefined {
    this.ensure();
    return this.metadataByName.get(path)?.packages?.[0];
  }
  sourceView(file: string): object | undefined {
    this.ensure();
    let metadata = this.metadataByName.get(file);
    if (!metadata) {
      const reply = this.lookup!(this.descriptor, file, true);
      metadata = reply && this.metadataByName.get(reply.fileName);
      if (metadata) this.metadataByName.set(file, metadata);
    }
    return metadata
      ? this.nodeRef({ file: metadata.fileName, index: 1 })
      : undefined;
  }
  private retain(source: RemoteSourceFile, bytes: number): void {
    const previous = this.sources.get(source.fileName);
    if (previous) {
      this.sources.delete(source.fileName);
      this.sources.set(source.fileName, previous);
      return;
    }
    // A source larger than the budget can service this access without retention.
    if (bytes > this.maxSourceBytes) return;
    while (this.stats.sourceBytes + bytes > this.maxSourceBytes) {
      const key = this.sources.keys().next().value!;
      this.stats.sourceBytes -= this.sourceSizes.get(key)!;
      this.sourceSizes.delete(key);
      this.sources.delete(key);
    }
    this.sources.set(source.fileName, source);
    this.sourceSizes.set(source.fileName, bytes);
    this.stats.sourceBytes += bytes;
  }
  private readonly sourceSizes = new Map<string, number>();
  attach(
    tree: Uint8Array,
    expected?: { file: string; text: string },
  ): RemoteSourceFile {
    const source = new RemoteSourceFile(
      Uint8Array.from(tree),
      new Wtf8Decoder('utf-8', { ignoreBOM: true }),
    );
    if (
      expected &&
      (source.fileName !== expected.file || source.text !== expected.text)
    )
      throw new Error('type service projection belongs to a different source');
    const previous = this.sources.get(source.fileName);
    if (previous) return previous;
    this.retain(source, tree.byteLength);
    return source;
  }
  source(file: string): RemoteSourceFile | undefined {
    this.ensure();
    const metadata = this.metadataByName.get(file);
    if (!metadata) return undefined;
    const old = this.sources.get(metadata.fileName);
    if (old) {
      this.sources.delete(metadata.fileName);
      this.sources.set(metadata.fileName, old);
      return old;
    }
    this.stats.sourceRequests++;
    const reply = this.lookup!(this.descriptor, metadata.fileName);
    if (!reply) return undefined;
    if (reply.fileName !== metadata.fileName)
      throw new Error('syntax source identity changed');
    if (!reply.tree) throw new Error('missing source syntax packet');
    return this.attach(reply.tree);
  }
  shape(ref: NodeRef): string {
    const source = this.sources.get(ref.file);
    if (!source) return '';
    const raw = source.getOrCreateNodeAtIndex(ref.index),
      parent = raw.parent;
    const role =
      ['expression', 'name', 'type', 'initializer'].find(
        (key) => parent && Reflect.get(parent, key, parent) === raw,
      ) ?? 'other';
    return `${raw.kind}:${parent?.kind}:${role}`;
  }
  matching<Operation extends string>(
    file: string,
    shapes: ReadonlyMap<string, ReadonlySet<Operation>>,
    limit: number,
  ): Array<{ op: Operation; node: NodeRef }> {
    const source = this.sources.get(file),
      requests: Array<{ op: Operation; node: NodeRef }> = [];
    if (!source) return requests;
    const visit = (node: RemoteNode): void => {
      if (requests.length > limit) return;
      const ref = { file, index: (node as unknown as { index: number }).index };
      for (const op of shapes.get(this.shape(ref)) ?? [])
        requests.push({ op, node: ref });
      node.forEachChild((child) => {
        if (child instanceof RemoteNode) visit(child);
      });
    };
    visit(source);
    return requests;
  }
  raw(node: object): RemoteNode {
    const ref = this.location(node),
      source = this.source(ref.file);
    if (!source) throw new Error('foreign source');
    const raw = source.getOrCreateNodeAtIndex(ref.index);
    if (!(raw instanceof RemoteNode))
      throw new Error('invalid compiler node handle');
    return raw;
  }
  private summary(ref: NodeRef): NodeFact {
    const key = this.key(ref),
      cached = this.summaries.get(key);
    if (cached) {
      this.summaries.delete(key);
      this.summaries.set(key, cached);
      return cached.fact;
    }
    let result: NodeFact | undefined;
    for (const fact of this.readNodes(ref)) {
      const key = this.key(fact.node);
      if (key === this.key(ref)) result = fact;
      if (this.summaries.has(key)) continue;
      const bytes =
        64 +
        2 *
          (fact.kind.length +
            (fact.text?.length ?? 0) +
            fact.node.file.length +
            (fact.parent?.file.length ?? 0) +
            (fact.name?.file.length ?? 0));
      if (bytes > this.maxSummaryBytes) continue;
      while (this.stats.summaryBytes + bytes > this.maxSummaryBytes) {
        const oldest = this.summaries.keys().next().value!;
        this.stats.summaryBytes -= this.summaries.get(oldest)!.bytes;
        this.summaries.delete(oldest);
      }
      this.summaries.set(key, { fact, bytes });
      this.stats.summaryBytes += bytes;
    }
    if (!result) throw new Error('missing syntax fact');
    return result;
  }
  private nativeKind(ref: NodeRef): string {
    if (ref.index === 1) return 'SourceFile';
    const source = this.sources.get(ref.file);
    if (!source) return this.summary(ref).kind;
    const kind = source.getOrCreateNodeAtIndex(ref.index).kind;
    if (kind === SyntaxKind.QualifiedName) return 'QualifiedName';
    if (kind === SyntaxKind.TypeReference) return 'TypeReference';
    if (kind === SyntaxKind.HeritageClause) return 'HeritageClause';
    return SyntaxKind[kind];
  }
  private isHeritage(ref: NodeRef): boolean {
    let current: NodeRef | undefined = ref;
    while (current && this.nativeKind(current) === 'QualifiedName')
      current = this.summary(current).parent;
    return Boolean(
      current &&
      this.nativeKind(current) === 'TypeReference' &&
      this.summary(current).parent &&
      this.nativeKind(this.summary(current).parent!) === 'HeritageClause',
    );
  }
  private kind(ref: NodeRef, name: string): number {
    if (name === 'TypeReference' && this.isHeritage(ref))
      name = 'ExpressionWithTypeArguments';
    else if (name === 'QualifiedName' && this.isHeritage(ref))
      name = 'PropertyAccessExpression';
    const result =
      this.typescript.SyntaxKind[name as keyof typeof ts.SyntaxKind];
    if (typeof result !== 'number') unsupported(`SyntaxKind.${name}`);
    return result;
  }
  nodeRef(ref: NodeRef | undefined): object | undefined {
    this.ensure();
    if (!ref) return undefined;
    const metadata = this.metadataByName.get(ref.file);
    if (!metadata) throw new Error('foreign source');
    ref = { file: metadata.fileName, index: ref.index };
    const key = this.key(ref),
      old = this.wrapped.get(key);
    if (old) return old;
    const identity = ref;
    const proxy: object = new Proxy(
      {},
      {
        get: (_target, key) => this.get(identity, key),
        has: (_target, key) => {
          this.ensure();
          return key in this.raw(proxy);
        },
        ownKeys: () => {
          this.ensure();
          return Reflect.ownKeys(this.raw(proxy));
        },
        getOwnPropertyDescriptor: (_target, key) => {
          this.ensure();
          const raw = this.raw(proxy);
          return Object.hasOwn(raw, key)
            ? {
                configurable: true,
                enumerable: true,
                get: () => this.get(identity, key),
              }
            : undefined;
        },
      },
    );
    this.identities.set(proxy, identity);
    this.wrapped.set(key, proxy);
    this.stats.nodes++;
    return proxy;
  }
  wrap(value: unknown): unknown {
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
    return this.nodeRef({
      file: value.getSourceFile().fileName,
      index: (value as unknown as { index: number }).index,
    });
  }
  private get(ref: NodeRef, key: string | symbol): unknown {
    this.ensure();
    if (key === 'getSourceFile')
      return () => {
        this.ensure();
        return this.sourceView(ref.file);
      };
    if (ref.index === 1) {
      const source = this.metadataByName.get(ref.file)!;
      if (key === 'fileName' || key === 'path' || key === 'isDeclarationFile')
        return source[key];
      if (key === 'kind') return this.typescript.SyntaxKind.SourceFile;
      if (key === 'parent') return undefined;
    }
    if (
      !this.sources.has(ref.file) &&
      ['kind', 'flags', 'parent', 'name', 'text', 'escapedText'].includes(
        String(key),
      )
    ) {
      const fact = this.summary(ref);
      if (key === 'kind') return this.kind(ref, fact.kind);
      if (key === 'flags')
        return translateFlags(fact.flags, NodeFlags, this.typescript.NodeFlags);
      if (key === 'parent') return this.nodeRef(fact.parent);
      if (
        key === 'name' &&
        !(fact.kind === 'QualifiedName' && this.isHeritage(ref))
      )
        return this.nodeRef(fact.name);
      if (key === 'text' && fact.text !== undefined) return fact.text;
      if (
        key === 'escapedText' &&
        (fact.kind === 'Identifier' || fact.kind === 'PrivateIdentifier')
      )
        return fact.text?.startsWith('__') ? `_${fact.text}` : fact.text;
    }
    const source = this.source(ref.file);
    if (!source) throw new Error('foreign source');
    const raw = source.getOrCreateNodeAtIndex(ref.index);
    const heritageReference =
      raw.kind === SyntaxKind.TypeReference &&
      raw.parent?.kind === SyntaxKind.HeritageClause;
    const heritageName =
      raw.kind === SyntaxKind.QualifiedName && this.isHeritage(ref);
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
      return translateFlags(raw.flags, NodeFlags, this.typescript.NodeFlags);
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
    if (key === 'getChildAt')
      return (index: number, sourceFile?: object) => {
        this.ensure();
        if (
          sourceFile !== undefined &&
          this.location(sourceFile).file !== raw.getSourceFile().fileName
        )
          unsupported('node.getChildAt with a foreign source');
        const child = raw.getChildAt(index);
        // The decoder also creates token/SyntaxList objects with no Go
        // compiler handle. Keep those explicit until their syntax views
        // can translate kinds and preserve identity without inventing one.
        if (child !== undefined && !(child instanceof RemoteNode))
          unsupported('node.getChildAt synthetic token or SyntaxList');
        return this.wrap(child);
      };
    if (
      typeof key === 'string' &&
      [
        'getChildren',
        'getFirstToken',
        'getLastToken',
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
                ? this.locationOrUndefined(a)
                  ? this.raw(a)
                  : a
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
  }
  close(): void {
    this.lookup = undefined;
    this.sources.clear();
    this.sourceSizes.clear();
    this.summaries.clear();
    this.wrapped.clear();
    this.stats.sourceBytes = 0;
    this.stats.summaryBytes = 0;
  }
}

/** Per-file ESTree mappings share the generation's canonical native objects. */
export class FileSyntaxViews {
  private closed = false;
  private forward = new WeakMap<ESTreeNode, number>();
  private reverse = new Map<number, number>();
  private aliases = new Map<
    number,
    NonNullable<TypeServiceProjection['aliases']>[number]
  >();
  private readonly materializedAliases = new Map<number, ESTreeNode>();
  private nodes: DecodedFrame['nodes'] | undefined;
  private lineStarts: number[] = [];
  private readonly sequenceMembers = new Map<number, readonly ESTreeNode[]>();
  private attached = false;
  constructor(
    private readonly registry: GenerationSyntaxViews,
    private readonly binding: FileBinding,
    private readonly ensureGeneration: () => void,
  ) {}
  private ensure(): void {
    this.ensureGeneration();
    if (this.closed) throw new Error('file type services are closed');
  }
  close(): void {
    this.closed = true;
    this.forward = new WeakMap();
    this.reverse.clear();
    this.aliases.clear();
    this.materializedAliases.clear();
    this.sequenceMembers.clear();
    this.nodes = undefined;
    this.lineStarts = [];
  }
  attach(
    frame: DecodedFrame,
    projection: TypeServiceProjection,
    sourceFileName: string,
  ): void {
    this.ensure();
    if (this.attached) throw new Error('file type services already attached');
    if (
      projection.binding.generation !== this.binding.generation ||
      projection.binding.checker !== this.binding.checker ||
      projection.origins.length !== frame.nodes.length
    )
      throw new Error('stale or incompatible projection origins');
    if (
      projection.sourceFileName !== sourceFileName ||
      sourceFileName !== this.binding.sourceFileName ||
      projection.binding.sourceFileName !== this.binding.sourceFileName
    )
      throw new Error('type service projection belongs to a different source');
    this.registry.attach(projection.tree, {
      file: sourceFileName,
      text: frame.text,
    });
    this.attached = true;
    this.nodes = frame.nodes;
    this.lineStarts = frame.lineStarts;
    for (let id = 1; id < frame.nodes.length; id++) {
      const index = projection.origins[id];
      if (index) this.forward.set(frame.nodes[id]!, index);
    }
    this.reverse = new Map(
      Object.entries(projection.reverse).map(([index, id]) => [
        Number(index),
        id,
      ]),
    );
    this.aliases = new Map(
      (projection.aliases ?? []).map((alias) => [alias.index, alias]),
    );
    for (const alias of this.aliases.values())
      if (!this.sequenceMembers.has(alias.sequence))
        this.sequenceMembers.set(
          alias.sequence,
          (frame.nodes[alias.sequence]!.expressions as ESTreeNode[]).slice(),
        );
  }
  readonly esTreeNodeToTSNodeMap = {
    get: (node: ESTreeNode): object | undefined => {
      this.ensure();
      const index = this.forward.get(node);
      return index
        ? this.registry.nodeRef({ file: this.binding.sourceFileName, index })
        : undefined;
    },
    has: (node: ESTreeNode): boolean => {
      this.ensure();
      return this.forward.has(node);
    },
  };
  readonly tsNodeToESTreeNodeMap = {
    get: (node: object): ESTreeNode | undefined => {
      this.ensure();
      const ref = this.registry.locationOrUndefined(node);
      if (!ref || ref.file !== this.binding.sourceFileName) return undefined;
      const index = ref.index;
      const id = this.reverse.get(index);
      if (id) return this.nodes![id];
      const old = this.materializedAliases.get(index);
      if (old) return old;
      const info = this.aliases.get(index);
      if (!info) return undefined;
      const alias: ESTreeNode = {
        type: 'SequenceExpression',
        range: [info.start, info.end],
        loc: {
          start: offsetToLineColumn(info.start, this.lineStarts),
          end: offsetToLineColumn(info.end, this.lineStarts),
        },
        expressions: this.sequenceMembers
          .get(info.sequence)!
          .slice(0, info.count),
      };
      this.materializedAliases.set(index, alias);
      this.forward.set(alias, index);
      this.registry.stats.aliases++;
      return alias;
    },
    has: (node: object): boolean => {
      this.ensure();
      const ref = this.registry.locationOrUndefined(node);
      if (!ref || ref.file !== this.binding.sourceFileName) return false;
      const index = ref.index;
      return this.reverse.has(index) || this.aliases.has(index);
    },
  };
}
