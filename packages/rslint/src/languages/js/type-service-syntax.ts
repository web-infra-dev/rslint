/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
import {
  RemoteSourceFile,
  RemoteNode,
  SyntaxKind,
  NodeFlags,
  Wtf8Decoder,
  type Node,
} from '@rslint/api';
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
interface SyntheticRef {
  readonly owner: NodeRef;
  readonly child: number;
  readonly listChild?: number;
}
type SyntaxRef = NodeRef | SyntheticRef;
function unsupported(name: string): never {
  throw new Error(`unsupported TypeScript operation: ${name}`);
}
// These native bits either have no JS equivalent or depend on the node kind.
const specialNodeFlags =
  NodeFlags.Reparsed |
  NodeFlags.ReparserTransformedLiteral |
  NodeFlags.HasJSDoc |
  NodeFlags.OptionalChain |
  NodeFlags.PossiblyContainsDeprecatedTag;
const commonNodeFlags = Object.entries(NodeFlags).filter(
  (entry): entry is [string, number] => {
    const value = entry[1];
    return (
      typeof value === 'number' &&
      value > NodeFlags.None &&
      (value & (value - 1)) === 0 &&
      !(value & specialNodeFlags)
    );
  },
);
function translateNodeFlags(
  flags: number,
  table: object,
  moduleKeyword: SyntaxKind = SyntaxKind.Unknown,
): number {
  const to = table as Record<string, number>;
  let result = 0;
  if (flags & NodeFlags.OptionalChain)
    result |= moduleKeyword ? to.NestedNamespace : to.OptionalChain;
  if (moduleKeyword === SyntaxKind.NamespaceKeyword) result |= to.Namespace;
  if (moduleKeyword === SyntaxKind.GlobalKeyword)
    result |= to.GlobalAugmentation;
  if (flags & NodeFlags.PossiblyContainsDeprecatedTag) result |= to.Deprecated;
  for (const [name, value] of commonNodeFlags) {
    if (!(flags & value)) continue;
    const translated = to[name];
    if (typeof translated !== 'number') unsupported(`enum.${name}`);
    result |= translated;
  }
  return result;
}

const sourceProperties = [
  'languageVersion',
  'hasNoDefaultLib',
  'impliedNodeFormat',
  'referencedFiles',
  'typeReferenceDirectives',
  'libReferenceDirectives',
];
function sourceProperty(key: PropertyKey): boolean {
  return typeof key === 'string' && sourceProperties.includes(key);
}

function wrapList(
  value: unknown[],
  wrap: (item: unknown) => unknown,
): unknown[] {
  const list = value.map(wrap);
  for (const key of ['pos', 'end', 'hasTrailingComma'])
    if (key in value)
      Object.defineProperty(list, key, {
        value: Reflect.get(value, key),
        enumerable: true,
        writable: true,
        configurable: true,
      });
  return list;
}

/** A worker/generation owns stable native identities and a bounded decoded
 * source LRU. Proxies retain identities, never a dependency source buffer. */
export class GenerationSyntaxViews {
  private readonly identities = new WeakMap<object, SyntaxRef>();
  private readonly wrapped = new WeakObjectCache<string, object>();
  private readonly generatedNodes = new WeakSet<object>();
  private readonly syntheticRefs = new WeakMap<Node, SyntheticRef>();
  private children = new WeakMap<object, object[]>();
  private readonly activeSources = new Map<
    string,
    { source: RemoteSourceFile; users: number }
  >();
  private readonly sources = new Map<
    string,
    { source: RemoteSourceFile; bytes: number }
  >();
  private readonly summaries = new Map<
    string,
    { fact: NodeFact; bytes: number }
  >();
  private readonly sourceOrdinals = new Map<string, number>();
  private readonly metadataByName = new Map<
    string,
    TypeServiceDescriptor['sources'][number]
  >();
  private sourceBytes = 0;
  private summaryBytes = 0;
  constructor(
    private readonly typescript: Pick<
      typeof ts,
      'SyntaxKind' | 'NodeFlags' | 'ScriptTarget'
    >,
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
    const ref = this.identities.get(node);
    return ref && !('owner' in ref) ? ref : undefined;
  }
  isStableNodeQuery(op: 'type' | 'constantValue', ref: NodeRef): boolean {
    if (ref.token) return false;
    const node = this.cachedSource(ref.file)?.getOrCreateNodeAtIndex(ref.index);
    // Enum declarations have fixed values. Other constant-value queries can
    // depend on a property access. Identifier uses can likewise have flow types.
    if (op === 'constantValue') return node?.kind === SyntaxKind.EnumMember;
    return (
      node?.kind === SyntaxKind.Identifier &&
      node.parent?.kind === SyntaxKind.VariableDeclaration &&
      Reflect.get(node.parent, 'name') === node
    );
  }
  parseTreeLocation(node: object): NodeRef | undefined {
    this.ensure();
    const ref = this.identities.get(node);
    if (!ref) {
      if (this.generatedNodes.has(node)) return undefined;
      throw new Error('foreign compiler node handle');
    }
    if (!('owner' in ref)) return ref;
    const { kind, pos, end } = this.rawSyntax(ref);
    return { ...ref.owner, token: { kind, pos, end } };
  }
  location(node: object): NodeRef {
    const ref = this.parseTreeLocation(node);
    if (!ref) throw new Error('generated compiler node has no source location');
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
  sourceByPath(path: string): object | undefined {
    this.ensure();
    const source = this.metadataByName.get(path);
    return source?.path === path && !this.configs.has(source.fileName)
      ? this.nodeRef({ file: source.fileName, index: 1 })
      : undefined;
  }
  private readonly configs = new Set<string>();
  diagnosticSource(file: string): object | undefined {
    const known = this.sourceView(file);
    if (known) return known;
    const found = this.lookup!(this.descriptor, file, true);
    return found && this.configSource(found.fileName);
  }
  configSource(file: string, path = file): object {
    if (!this.metadataByName.has(file)) {
      this.configs.add(file);
      const source = {
        fileName: file,
        path,
        external: false,
        defaultLibrary: false,
        isDeclarationFile: false,
      };
      this.sourceOrdinals.set(file, this.sourceOrdinals.size + 1);
      this.metadataByName.set(file, source);
      this.metadataByName.set(path, source);
    }
    return this.nodeRef({ file, index: 1 })!;
  }
  sourceView(file: string): object | undefined {
    this.ensure();
    let metadata = this.metadataByName.get(file);
    if (!metadata) {
      const reply = this.lookup!(this.descriptor, file, true);
      metadata = reply && this.metadataByName.get(reply.fileName);
      if (metadata) this.metadataByName.set(file, metadata);
    }
    return metadata && !this.configs.has(metadata.fileName)
      ? this.nodeRef({ file: metadata.fileName, index: 1 })
      : undefined;
  }
  private cachedSource(file: string): RemoteSourceFile | undefined {
    const active = this.activeSources.get(file);
    if (active) return active.source;
    const entry = this.sources.get(file);
    if (entry) {
      this.sources.delete(file);
      this.sources.set(file, entry);
    }
    return entry?.source;
  }
  private retain(source: RemoteSourceFile, bytes: number): void {
    if (bytes > this.maxSourceBytes) return;
    while (this.sourceBytes + bytes > this.maxSourceBytes) {
      const [key, entry] = this.sources.entries().next().value!;
      this.sources.delete(key);
      this.sourceBytes -= entry.bytes;
    }
    this.sources.set(source.fileName, { source, bytes });
    this.sourceBytes += bytes;
  }
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
    const previous = this.cachedSource(source.fileName);
    if (!previous) this.retain(source, tree.byteLength);
    const shared = previous ?? source;
    if (expected) {
      const active = this.activeSources.get(expected.file);
      if (active) active.users++;
      else this.activeSources.set(expected.file, { source: shared, users: 1 });
    }
    return shared;
  }
  releaseFile(file: string): void {
    const active = this.activeSources.get(file);
    if (active && --active.users === 0) this.activeSources.delete(file);
  }
  source(file: string): RemoteSourceFile | undefined {
    this.ensure();
    const metadata = this.metadataByName.get(file);
    if (!metadata) return undefined;
    const old = this.cachedSource(metadata.fileName);
    if (old) return old;
    const reply = this.lookup!(this.descriptor, metadata.fileName);
    if (!reply) return undefined;
    if (reply.fileName !== metadata.fileName)
      throw new Error('syntax source identity changed');
    if (!reply.tree) throw new Error('missing source syntax packet');
    return this.attach(reply.tree);
  }
  private rawNode(ref: NodeRef): RemoteNode {
    const source = this.source(ref.file);
    if (!source) throw new Error('foreign source');
    const raw = source.getOrCreateNodeAtIndex(ref.index);
    if (!(raw instanceof RemoteNode))
      throw new Error('invalid compiler node handle');
    return raw;
  }
  private rawSyntax(ref: SyntaxRef): Node {
    if (!('owner' in ref)) return this.rawNode(ref);
    const child = this.rawNode(ref.owner).getChildAt(ref.child);
    return ref.listChild === undefined
      ? child
      : child.getChildAt(ref.listChild);
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
      while (this.summaryBytes + bytes > this.maxSummaryBytes) {
        const oldest = this.summaries.keys().next().value!;
        this.summaryBytes -= this.summaries.get(oldest)!.bytes;
        this.summaries.delete(oldest);
      }
      this.summaries.set(key, { fact, bytes });
      this.summaryBytes += bytes;
    }
    if (!result) throw new Error('missing syntax fact');
    return result;
  }
  private nativeKind(ref: NodeRef): string {
    if (ref.index === 1) return 'SourceFile';
    const source = this.cachedSource(ref.file);
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
  private kind(ref: SyntaxRef, name: string): number {
    if (!('owner' in ref) && name === 'TypeReference' && this.isHeritage(ref))
      name = 'ExpressionWithTypeArguments';
    else if (
      !('owner' in ref) &&
      name === 'QualifiedName' &&
      this.isHeritage(ref)
    )
      name = 'PropertyAccessExpression';
    return this.translateKind(name);
  }
  private translateKind(name: string): number {
    if (name === 'EndOfFile') name = 'EndOfFileToken';
    if (name === 'JSDocUnknownTag') name = 'JSDocTag';
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
    const identity = { file: metadata.fileName, index: ref.index };
    return this.wrapRef(identity, this.key(identity));
  }
  private wrapRef(identity: SyntaxRef, cacheKey: string): object {
    const old = this.wrapped.get(cacheKey);
    if (old) return old;
    const source = !('owner' in identity) && identity.index === 1;
    const proxy: object = new Proxy(
      {},
      {
        get: (target, key, receiver) => {
          if (!Object.hasOwn(target, key))
            return this.get(identity, key, proxy);
          this.ensure();
          // TypeScript helpers write node-local caches, such as modifier flags.
          return Reflect.get(target, key, receiver);
        },
        has: (target, key) => {
          this.ensure();
          return (
            Object.hasOwn(target, key) ||
            (source && sourceProperty(key)) ||
            key in this.rawSyntax(identity)
          );
        },
        ownKeys: (target) => {
          this.ensure();
          return [
            ...new Set([
              ...Reflect.ownKeys(this.rawSyntax(identity)),
              ...(source ? sourceProperties : []),
              ...Reflect.ownKeys(target),
            ]),
          ];
        },
        getOwnPropertyDescriptor: (target, key) => {
          this.ensure();
          const own = Reflect.getOwnPropertyDescriptor(target, key);
          if (own) return own;
          if (source && sourceProperty(key))
            return {
              value: this.get(identity, key, proxy),
              writable: true,
              enumerable: true,
              configurable: true,
            };
          const raw = this.rawSyntax(identity);
          return Object.hasOwn(raw, key)
            ? {
                configurable: true,
                enumerable: true,
                get: () => this.get(identity, key, proxy),
              }
            : undefined;
        },
      },
    );
    this.identities.set(proxy, identity);
    this.wrapped.set(cacheKey, proxy);
    return proxy;
  }
  private ref(value: RemoteNode): NodeRef {
    return {
      file: value.getSourceFile().fileName,
      index: (value as unknown as { index: number }).index,
    };
  }
  private wrapNode(value: Node): object {
    if (value instanceof RemoteNode) return this.nodeRef(this.ref(value))!;
    const parent = value.parent;
    if (!(parent instanceof RemoteNode))
      throw new Error('missing syntax owner');
    let ref = this.syntheticRefs.get(value);
    if (!ref) {
      const owner = this.ref(parent);
      // Index the reader's cached child objects once per decoded parent.
      for (const [child, item] of parent.getChildren().entries()) {
        if (item instanceof RemoteNode) continue;
        this.syntheticRefs.set(item, { owner, child });
        if (item.kind !== SyntaxKind.SyntaxList) continue;
        for (const [listChild, token] of item.getChildren().entries()) {
          if (!(token instanceof RemoteNode))
            this.syntheticRefs.set(token, { owner, child, listChild });
        }
      }
      ref = this.syntheticRefs.get(value);
      if (!ref) throw new Error('missing synthetic syntax node');
    }
    const key =
      value.kind === SyntaxKind.SyntaxList
        ? `list:${this.key(ref.owner)}:${ref.child}`
        : `token:${this.sourceOrdinals.get(ref.owner.file)}:${value.pos}:${value.end}`;
    return this.wrapRef(ref, key);
  }
  firstToken(ref: NodeRef): object | undefined {
    const token = this.rawNode(ref).getFirstToken();
    return token && this.wrapNode(token);
  }
  wrap(value: unknown): unknown {
    if (Array.isArray(value)) {
      // The pinned reader gives each NodeArray its own encoded node index.
      if ('parent' in value && value.parent instanceof RemoteNode) {
        const key = this.key({
          file: value.parent.getSourceFile().fileName,
          index: (value as unknown as { index: number }).index,
        });
        const found = this.wrapped.get(key);
        if (found) return found;
        const list = wrapList(value, (item) => this.wrapNode(item as Node));
        this.wrapped.set(key, list);
        return list;
      }
      return wrapList(value, (item) => this.wrap(item));
    }
    if (
      value instanceof RemoteNode ||
      (typeof value === 'object' &&
        value !== null &&
        'kind' in value &&
        'getSourceFile' in value)
    )
      return this.wrapNode(value as Node);
    return value;
  }
  private get(ref: SyntaxRef, key: string | symbol, proxy: object): unknown {
    this.ensure();
    const real = 'owner' in ref ? undefined : ref,
      file = 'owner' in ref ? ref.owner.file : ref.file;
    if (key === 'getSourceFile')
      return () => {
        this.ensure();
        return this.configs.has(file)
          ? this.configSource(file)
          : this.sourceView(file);
      };
    if (real?.index === 1) {
      const source = this.metadataByName.get(file)!;
      if (key === 'fileName' || key === 'path' || key === 'isDeclarationFile')
        return source[key];
      if (key === 'impliedNodeFormat') return source.impliedNodeFormat;
      if (
        key === 'referencedFiles' ||
        key === 'typeReferenceDirectives' ||
        key === 'libReferenceDirectives'
      ) {
        const references = this.source(file)![key].map(
          ({ pos, end, fileName, resolutionMode, preserve }) => ({
            pos,
            end,
            fileName,
            ...(resolutionMode ? { resolutionMode } : {}),
            ...(preserve ? { preserve } : {}),
          }),
        );
        Object.defineProperty(proxy, key, {
          value: references,
          writable: true,
          configurable: true,
          enumerable: true,
        });
        return references;
      }
      // TypeScript 6 keeps this legacy property false, including for no-default-lib directives.
      if (key === 'hasNoDefaultLib') return false;
      if (key === 'languageVersion') {
        const targets = this.typescript.ScriptTarget;
        if (this.configs.has(file) || file.toLowerCase().endsWith('.json'))
          return targets.ES2015;
        if (this.descriptor.syntaxOnly) return targets.Latest;
        const value = this.descriptor.compilerOptions.target;
        const target: ts.ScriptTarget | undefined =
          typeof value === 'number' ? value : undefined;
        return target === undefined || target === targets.ES3
          ? targets.LatestStandard
          : target;
      }
      if (key === 'update') return () => unsupported('SourceFile.update');
      if (key === 'getLineEndOfPosition')
        return (position: number) => {
          this.ensure();
          const sourceFile = this.source(file)!;
          const { line } = sourceFile.getLineAndCharacterOfPosition(position);
          const starts = sourceFile.getLineStarts();
          const end =
            (line + 1 >= starts.length && sourceFile.getEnd()) ||
            starts[line + 1] - 1;
          return sourceFile.text[end] === '\n' &&
            sourceFile.text[end - 1] === '\r'
            ? end - 1
            : end;
        };
      if (key === 'kind') return this.typescript.SyntaxKind.SourceFile;
      if (key === 'parent') return undefined;
    }
    if (
      real &&
      !this.sources.has(file) &&
      !this.configs.has(file) &&
      ['kind', 'flags', 'parent', 'name', 'text', 'escapedText'].includes(
        String(key),
      )
    ) {
      const fact = this.summary(real);
      if (key === 'kind') return this.kind(real, fact.kind);
      if (key === 'flags')
        return translateNodeFlags(
          fact.flags,
          this.typescript.NodeFlags,
          fact.moduleKeyword,
        );
      if (key === 'parent') return this.nodeRef(fact.parent);
      if (
        key === 'name' &&
        !(fact.kind === 'QualifiedName' && this.isHeritage(real))
      )
        return this.nodeRef(fact.name);
      if (key === 'text' && fact.text !== undefined) return fact.text;
      if (
        key === 'escapedText' &&
        (fact.kind === 'Identifier' || fact.kind === 'PrivateIdentifier')
      )
        return fact.text?.startsWith('__') ? `_${fact.text}` : fact.text;
    }
    if (key === 'forEachChild')
      return (
        cb: (node: unknown) => unknown,
        cbNodes?: (nodes: unknown) => unknown,
      ) => {
        this.ensure();
        return this.rawSyntax(ref).forEachChild(
          (n) => cb(this.wrapNode(n)),
          cbNodes ? (nodes) => cbNodes(this.wrap(nodes)) : undefined,
        );
      };
    if (key === 'getChildren')
      return (sourceFile?: object) => {
        this.checkSource(file, sourceFile);
        return this.childNodes(ref, proxy);
      };
    if (key === 'getChildAt')
      return (index: number, sourceFile?: object) => {
        this.checkSource(file, sourceFile);
        const children = this.children.get(proxy);
        if (children) return children[index];
        const child = this.rawSyntax(ref).getChildAt(index);
        return child && this.wrapNode(child);
      };
    if (key === 'getChildCount')
      return (sourceFile?: object) => {
        this.checkSource(file, sourceFile);
        return (
          this.children.get(proxy)?.length ??
          this.rawSyntax(ref).getChildCount()
        );
      };
    if (key === 'getFirstToken' || key === 'getLastToken')
      return (sourceFile?: object) => {
        this.checkSource(file, sourceFile);
        const token = this.rawSyntax(ref)[key]();
        return token && this.wrapNode(token);
      };
    if ('owner' in ref) {
      if (key === 'parent') return this.nodeRef(ref.owner);
      if (key === 'flags') {
        const flags = Reflect.get(this.nodeRef(ref.owner)!, 'flags') as number;
        return flags & this.typescript.NodeFlags.ContextFlags;
      }
    }
    const raw = this.rawSyntax(ref);
    if (
      key === 'flags' &&
      real &&
      !this.configs.has(file) &&
      (raw.kind === SyntaxKind.ModuleDeclaration ||
        raw.flags & NodeFlags.PossiblyContainsDeprecatedTag)
    ) {
      const fact = this.summary(real);
      return translateNodeFlags(
        fact.flags,
        this.typescript.NodeFlags,
        fact.moduleKeyword,
      );
    }
    if (key === 'children' && raw.kind === SyntaxKind.SyntaxList)
      return this.childNodes(ref, proxy);
    const heritageReference =
      raw.kind === SyntaxKind.TypeReference &&
      raw.parent?.kind === SyntaxKind.HeritageClause;
    const heritageName =
      real && raw.kind === SyntaxKind.QualifiedName && this.isHeritage(real);
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
    const member = this.syntaxMember(raw, key);
    if (typeof member === 'function')
      return (...args: unknown[]) => {
        this.ensure();
        const current = this.rawSyntax(ref);
        return this.wrap(
          Reflect.apply(
            member,
            current,
            args.map((arg) => {
              const identity =
                typeof arg === 'object' && arg !== null
                  ? this.identities.get(arg)
                  : undefined;
              return identity ? this.rawSyntax(identity) : arg;
            }),
          ),
        );
      };
    const value = this.wrap(member);
    if (Array.isArray(value))
      Object.defineProperty(proxy, key, {
        value,
        writable: true,
        configurable: true,
        enumerable: true,
      });
    return value;
  }
  private syntaxMember(raw: Node, key: string | symbol): unknown {
    if (
      key === 'kind' ||
      key === 'operator' ||
      key === 'token' ||
      key === 'keywordToken' ||
      key === 'keyword' ||
      key === 'phaseModifier'
    ) {
      const kind = Reflect.get(raw, key) as number | undefined;
      return kind === undefined
        ? undefined
        : this.translateKind(SyntaxKind[kind]);
    }
    if (key === 'flags')
      return translateNodeFlags(
        raw.flags,
        this.typescript.NodeFlags,
        raw.kind === SyntaxKind.ModuleDeclaration
          ? (Reflect.get(raw, 'keyword') as number)
          : 0,
      );
    if (
      key === 'escapedText' &&
      (raw.kind === SyntaxKind.Identifier ||
        raw.kind === SyntaxKind.PrivateIdentifier)
    ) {
      const text = Reflect.get(raw, 'text') as string;
      return text.startsWith('__') ? `_${text}` : text;
    }
    return Reflect.get(
      raw,
      key === 'default' && raw.kind === SyntaxKind.TypeParameter
        ? 'defaultType'
        : key,
    );
  }

  // A checker-generated tree owns its packet. Its nodes have no Program source
  // indices and must never enter the session's source-node identity map.
  generatedTypeNode(bytes: Uint8Array): object {
    const root = new RemoteSourceFile(
      bytes,
      new Wtf8Decoder('utf-8', { ignoreBOM: true }),
    );
    const views = new WeakMap<object, object>();
    const wrap = (value: unknown): unknown => {
      if (typeof value !== 'object' || value === null) return value;
      const old = views.get(value);
      if (old) return old;
      if (Array.isArray(value)) {
        const list = wrapList(value, wrap);
        views.set(value, list);
        return list;
      }
      if (!(value instanceof RemoteNode)) return value;
      const proxy = new Proxy(
        {},
        {
          get: (target, key) => {
            this.ensure();
            if (Object.hasOwn(target, key)) return Reflect.get(target, key);
            if (key === 'parent') return undefined;
            if (key === 'getSourceFile')
              return () => {
                this.ensure();
                return undefined;
              };
            if (key === 'forEachChild')
              return (
                cb: (n: unknown) => unknown,
                cbNodes?: (nodes: unknown) => unknown,
              ) => {
                this.ensure();
                return value.forEachChild(
                  (n) => cb(wrap(n)),
                  cbNodes ? (nodes) => cbNodes(wrap(nodes)) : undefined,
                );
              };
            const member = this.syntaxMember(value, key);
            if (typeof member === 'function')
              return (...args: unknown[]) => {
                this.ensure();
                return wrap(Reflect.apply(member, value, args));
              };
            return wrap(member);
          },
          has: (target, key) => {
            this.ensure();
            return Object.hasOwn(target, key) || key in value;
          },
        },
      );
      this.generatedNodes.add(proxy);
      views.set(value, proxy);
      return proxy;
    };
    return wrap(root) as object;
  }
  private childNodes(ref: SyntaxRef, proxy: object): object[] {
    let children = this.children.get(proxy);
    if (!children) {
      children = this.rawSyntax(ref)
        .getChildren()
        .map((child) => this.wrapNode(child));
      this.children.set(proxy, children);
    }
    return children;
  }
  private checkSource(file: string, sourceFile: object | undefined): void {
    this.ensure();
    if (sourceFile !== undefined) {
      const ref = this.locationOrUndefined(sourceFile);
      if (ref?.file !== file || ref.index !== 1)
        throw new Error('foreign syntax source');
    }
  }
  close(): void {
    this.lookup = undefined;
    this.children = new WeakMap();
    this.activeSources.clear();
    this.sources.clear();
    this.summaries.clear();
    this.wrapped.clear();
    this.sourceBytes = 0;
    this.summaryBytes = 0;
  }
}

/** Per-file ESTree mappings share the generation's canonical native objects. */
export class FileSyntaxViews {
  private closed = false;
  private forward = new WeakMap<
    ESTreeNode,
    number | { firstTokenOf: number }
  >();
  private reverse = new Map<number, number>();
  private readonly aliases = new Map<
    number,
    | NonNullable<TypeServiceProjection['aliases']>[number]
    | NonNullable<TypeServiceProjection['chainAliases']>[number]
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
    if (this.attached) this.registry.releaseFile(this.binding.sourceFileName);
    this.attached = false;
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
      projection.sourceFileName !== this.binding.sourceFileName ||
      projection.binding.sourceFileName !== this.binding.sourceFileName ||
      (sourceFileName !== this.binding.sourceFileName &&
        this.registry.sourceView(sourceFileName) !==
          this.registry.sourceView(this.binding.sourceFileName))
    )
      throw new Error('type service projection belongs to a different source');
    this.registry.attach(projection.tree, {
      file: this.binding.sourceFileName,
      text: frame.text,
    });
    this.attached = true;
    this.nodes = frame.nodes;
    this.lineStarts = frame.lineStarts;
    for (let id = 1; id < frame.nodes.length; id++) {
      const index = projection.origins[id];
      if (index) this.forward.set(frame.nodes[id]!, index);
    }
    for (const [id, firstTokenOf] of Object.entries(
      projection.firstTokens ?? {},
    ))
      this.forward.set(frame.nodes[Number(id)]!, { firstTokenOf });
    this.reverse = new Map(
      Object.entries(projection.reverse).map(([index, id]) => [
        Number(index),
        id,
      ]),
    );
    for (const alias of projection.aliases ?? []) {
      this.aliases.set(alias.index, alias);
      if (this.sequenceMembers.has(alias.sequence)) continue;
      const members = projection.sequenceMembers?.[alias.sequence];
      this.sequenceMembers.set(
        alias.sequence,
        members
          ? members.map((id) => frame.nodes[id]!)
          : (frame.nodes[alias.sequence]!.expressions as ESTreeNode[]).slice(),
      );
    }
    for (const alias of projection.chainAliases ?? [])
      this.aliases.set(alias.index, alias);
  }
  readonly esTreeNodeToTSNodeMap = {
    get: (node: ESTreeNode): object | undefined => {
      this.ensure();
      const origin = this.forward.get(node);
      if (origin === undefined) return undefined;
      const file = this.binding.sourceFileName;
      return typeof origin === 'number'
        ? this.registry.nodeRef({ file, index: origin })
        : this.registry.firstToken({ file, index: origin.firstTokenOf });
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
      const info = this.aliases.get(index);
      if (!info) return undefined;
      const chain = 'expression' in info;
      // Reverse aliases for the same chain wrapper share its expression's
      // frame ID. Negative keys keep those IDs distinct from native indices.
      const key = chain ? -info.expression : index;
      const old = this.materializedAliases.get(key);
      if (old) return old;
      const position: Pick<ESTreeNode, 'range' | 'loc'> = {
        range: [info.start, info.end],
        loc: {
          start: offsetToLineColumn(info.start, this.lineStarts),
          end: offsetToLineColumn(info.end, this.lineStarts),
        },
      };
      const alias: ESTreeNode = chain
        ? {
            type: 'ChainExpression',
            start: info.start,
            end: info.end,
            ...position,
            expression: this.nodes![info.expression],
          }
        : {
            type: 'SequenceExpression',
            ...position,
            expressions: this.sequenceMembers
              .get(info.sequence)!
              .slice(0, info.count),
          };
      const origin = chain ? info.origin : index;
      this.materializedAliases.set(key, alias);
      if (origin) this.forward.set(alias, origin);
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
