/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
import {
  RemoteSourceFile,
  RemoteNode,
  SyntaxKind,
  NodeFlags,
  Wtf8Decoder,
} from '@rslint/api';
import type ts from 'typescript';
import type {
  FileBinding,
  TypeServiceDescriptor,
  TypeServiceProjection,
  SyntaxSource,
  NodeRef,
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

/** Owns native syntax views for exactly one lint invocation. Origin maps hold
 * numbers; native nodes/proxies and reverse-only aliases appear on lookup. */
export class FileSyntaxViews {
  private rawNodes = new WeakMap<object, RemoteNode>();
  private wrappedNodes = new WeakMap<RemoteNode, object>();
  private heritageNames = new WeakMap<RemoteNode, boolean>();
  private readonly sources = new Map<string, RemoteSourceFile>();
  private forward = new WeakMap<ESTreeNode, number>();
  private reverse = new Map<number, number>();
  private aliases = new Map<
    number,
    NonNullable<TypeServiceProjection['aliases']>[number]
  >();
  private readonly materializedAliases = new Map<number, ESTreeNode>();
  private original: RemoteSourceFile | undefined;
  private nodes: DecodedFrame['nodes'] | undefined;
  private lineStarts: number[] = [];
  private readonly sequenceMembers = new Map<number, readonly ESTreeNode[]>();
  private attached = false;
  readonly stats = { nodes: 0, aliases: 0, sourceBytes: 0 };
  constructor(
    private readonly typescript: Pick<typeof ts, 'SyntaxKind' | 'NodeFlags'>,
    private readonly descriptor: TypeServiceDescriptor,
    private readonly binding: FileBinding,
    private lookup: SyntaxSource | undefined,
    private readonly ensure: () => void,
  ) {}
  close(): void {
    this.lookup = undefined;
    this.sources.clear();
    this.original = undefined;
    this.nodes = undefined;
    this.forward = new WeakMap();
    this.reverse.clear();
    this.aliases.clear();
    this.materializedAliases.clear();
    this.rawNodes = new WeakMap();
    this.wrappedNodes = new WeakMap();
    this.heritageNames = new WeakMap();
    this.lineStarts = [];
    this.sequenceMembers.clear();
    this.stats.sourceBytes = 0;
  }
  location(node: object): NodeRef {
    this.ensure();
    const raw = this.rawNodes.get(node);
    if (!raw) throw new Error('foreign compiler node handle');
    return {
      file: raw.getSourceFile().fileName,
      index: (raw as unknown as { index: number }).index,
    };
  }
  metadata(node: object) {
    const { file } = this.location(node),
      metadata = this.descriptor.sources.find((s) => s.fileName === file);
    if (!metadata) throw new Error('foreign source');
    return metadata;
  }
  sourceView(file: string): object | undefined {
    const source = this.source(file);
    return source ? (this.wrap(source) as object) : undefined;
  }
  source(file: string): RemoteSourceFile | undefined {
    this.ensure();
    let source = this.sources.get(file);
    if (!source) {
      const reply = this.lookup!(this.binding, file);
      if (!reply) return undefined;
      source = this.sources.get(reply.fileName);
      if (!source) {
        const tree = Uint8Array.from(reply.tree);
        source = new RemoteSourceFile(
          tree,
          new Wtf8Decoder('utf-8', { ignoreBOM: true }),
        );
        this.sources.set(reply.fileName, source);
        this.stats.sourceBytes += tree.byteLength;
      }
      this.sources.set(file, source);
    }
    return source;
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
    const tree = Uint8Array.from(projection.tree);
    const source = new RemoteSourceFile(
      tree,
      new Wtf8Decoder('utf-8', { ignoreBOM: true }),
    );
    if (
      source.fileName !== projection.sourceFileName ||
      source.fileName !== sourceFileName ||
      source.fileName !== this.binding.sourceFileName ||
      projection.binding.sourceFileName !== this.binding.sourceFileName ||
      source.text !== frame.text
    )
      throw new Error('type service projection belongs to a different source');
    const existing = this.sources.get(source.fileName);
    this.attached = true;
    this.original = existing ?? source;
    if (!existing) {
      this.sources.set(source.fileName, source);
      this.stats.sourceBytes += tree.byteLength;
    }
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
        ? (this.wrap(this.original!.getOrCreateNodeAtIndex(index)) as object)
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
      const raw = this.rawNodes.get(node);
      if (!raw || !Object.is(raw.getSourceFile(), this.original))
        return undefined;
      const index = (raw as unknown as { index: number }).index;
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
      this.stats.aliases++;
      return alias;
    },
    has: (node: object): boolean => {
      this.ensure();
      const raw = this.rawNodes.get(node);
      if (!raw || !Object.is(raw.getSourceFile(), this.original)) return false;
      const index = (raw as unknown as { index: number }).index;
      return this.reverse.has(index) || this.aliases.has(index);
    },
  };

  private isHeritageName(node: RemoteNode): boolean {
    const path: RemoteNode[] = [];
    let owner = node,
      value = false;
    while (owner.kind === SyntaxKind.QualifiedName) {
      const old = this.heritageNames.get(owner);
      if (old !== undefined) {
        value = old;
        break;
      }
      path.push(owner);
      owner = owner.parent as RemoteNode;
    }
    if (owner.kind !== SyntaxKind.QualifiedName)
      value =
        owner.kind === SyntaxKind.TypeReference &&
        owner.parent?.kind === SyntaxKind.HeritageClause;
    for (const current of path) this.heritageNames.set(current, value);
    return value;
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
    const old = this.wrappedNodes.get(value);
    if (old) return old;
    // tsgo represents type-only heritage as TypeReference/QualifiedName.
    // Expose TS's expression view while keeping the original compiler handle.
    const heritageReference =
      value.kind === SyntaxKind.TypeReference &&
      value.parent?.kind === SyntaxKind.HeritageClause;
    const heritageName =
      value.kind === SyntaxKind.QualifiedName && this.isHeritageName(value);
    this.stats.nodes++;
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
        if (key === 'getChildAt')
          return (index: number, sourceFile?: object) => {
            this.ensure();
            if (
              sourceFile !== undefined &&
              !Object.is(this.rawNodes.get(sourceFile), raw.getSourceFile())
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
  nodeRef(value: NodeRef | undefined): object | undefined {
    if (!value) return undefined;
    const source = this.source(value.file);
    if (!source) throw new Error('foreign source');
    return this.wrap(source.getOrCreateNodeAtIndex(value.index)) as object;
  }
}
