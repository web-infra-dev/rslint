/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
import {
  childProperties,
  RemoteSourceFile,
  RemoteNode,
  SyntaxKind,
  NodeFlags,
  TokenFlags,
  Wtf8Decoder,
  type Node,
} from '@rslint/api';
import type ts from 'typescript';
import type { TypeServiceCompilerData } from './type-service-compiler-data.js';
import { WeakObjectCache } from './type-service-cache.js';
import type {
  FileBinding,
  GeneratedSyntax,
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
const tokenNavigation = new Set<PropertyKey>([
  'getChildren',
  'getChildAt',
  'getChildCount',
  'getFirstToken',
  'getLastToken',
]);
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
  'fileName',
  'path',
  'isDeclarationFile',
  'text',
  'languageVersion',
  'languageVariant',
  'scriptKind',
  'hasNoDefaultLib',
  'impliedNodeFormat',
  'referencedFiles',
  'typeReferenceDirectives',
  'libReferenceDirectives',
  'imports',
  'moduleAugmentations',
  'ambientModuleNames',
  'externalModuleIndicator',
];
function publicSyntaxKind(node: Node): SyntaxKind {
  if (
    node.kind === SyntaxKind.QualifiedName ||
    node.kind === SyntaxKind.TypeReference
  ) {
    let parent: Node | undefined = node;
    while (parent?.kind === SyntaxKind.QualifiedName) parent = parent.parent;
    if (
      parent?.kind === SyntaxKind.TypeReference &&
      parent.parent?.kind === SyntaxKind.HeritageClause
    )
      return node.kind === SyntaxKind.TypeReference
        ? SyntaxKind.ExpressionWithTypeArguments
        : SyntaxKind.PropertyAccessExpression;
  }
  return node.kind;
}

// The compiler's child table excludes scalar fields exposed by the JS AST.
const scalarFields: Partial<Record<SyntaxKind, readonly string[]>> = {
  [SyntaxKind.Identifier]: ['escapedText'],
  [SyntaxKind.StringLiteral]: ['hasExtendedUnicodeEscape'],
  [SyntaxKind.PrivateIdentifier]: ['escapedText'],
  [SyntaxKind.Block]: ['multiLine'],
  [SyntaxKind.ArrayLiteralExpression]: ['multiLine'],
  [SyntaxKind.ObjectLiteralExpression]: ['multiLine'],
  [SyntaxKind.ImportAttributes]: ['multiLine'],
  [SyntaxKind.ImportClause]: ['isTypeOnly', 'phaseModifier'],
  [SyntaxKind.ImportSpecifier]: ['isTypeOnly'],
  [SyntaxKind.ExportSpecifier]: ['isTypeOnly'],
  [SyntaxKind.ExportDeclaration]: ['isTypeOnly'],
  [SyntaxKind.ImportEqualsDeclaration]: ['isTypeOnly'],
  [SyntaxKind.ExportAssignment]: ['isExportEquals'],
  [SyntaxKind.ImportType]: ['isTypeOf'],
  [SyntaxKind.JsxText]: ['text', 'containsOnlyTriviaWhiteSpaces'],
  [SyntaxKind.JSDocText]: ['text'],
  [SyntaxKind.JSDocLink]: ['text'],
  [SyntaxKind.JSDocLinkCode]: ['text'],
  [SyntaxKind.JSDocLinkPlain]: ['text'],
  [SyntaxKind.JSDocTypeLiteral]: ['isArrayType'],
  [SyntaxKind.JSDocParameterTag]: ['isNameFirst', 'isBracketed'],
  [SyntaxKind.JSDocPropertyTag]: ['isNameFirst', 'isBracketed'],
  [SyntaxKind.SourceFile]: sourceProperties,
  [SyntaxKind.SyntaxList]: ['children'],
};

type JsDocKinds = TypeServiceCompilerData['jsDocKinds'];

// Only differences from the reader's child table belong here. A string names
// the native slot, and false omits a slot reserved by the Go parser.
const fieldAliases: Partial<
  Record<SyntaxKind, Readonly<Record<string, string | false>>>
> = {
  [SyntaxKind.TypeParameter]: { default: 'defaultType', defaultType: false },
  [SyntaxKind.BinaryExpression]: { type: false, modifiers: false },
  [SyntaxKind.PropertyAssignment]: { type: false },
  [SyntaxKind.ShorthandPropertyAssignment]: { type: false },
  [SyntaxKind.ForInStatement]: { awaitModifier: false },
  [SyntaxKind.ExportAssignment]: { type: false },
  [SyntaxKind.ModuleDeclaration]: { attributes: false },
  [SyntaxKind.DefaultClause]: { expression: false },
  [SyntaxKind.ImportAttributes]: { elements: 'attributes', attributes: false },
  [SyntaxKind.ImportDeclaration]: { assertClause: 'attributes' },
  [SyntaxKind.ExportDeclaration]: { assertClause: 'attributes' },
  [SyntaxKind.JSDocSeeTag]: { name: 'nameExpression', nameExpression: false },
  [SyntaxKind.JSDocAugmentsTag]: { class: 'className', className: false },
  [SyntaxKind.JSDocImplementsTag]: { class: 'className', className: false },
  [SyntaxKind.JSDocTypedefTag]: { fullName: 'name' },
  [SyntaxKind.JSDocCallbackTag]: { fullName: 'name' },
  [SyntaxKind.JSDocTypeLiteral]: {
    jsDocPropertyTags: 'jsdocPropertyTags',
    jsdocPropertyTags: false,
  },
};

function syntaxFields(raw: Node, jsDocKinds: JsDocKinds): string[] {
  const kind = publicSyntaxKind(raw);
  const aliases = fieldAliases[kind];
  const fields = [
    'pos',
    'end',
    'kind',
    'flags',
    'parent',
    'original',
    ...(scalarFields[kind] ?? []),
  ];
  for (const key of childProperties[kind] ?? []) {
    if (key === undefined || aliases?.[key] === false) continue;
    if (key === 'postfixToken') {
      fields.push('questionToken');
      if (
        kind !== SyntaxKind.MethodSignature &&
        kind !== SyntaxKind.PropertySignature
      )
        fields.push('exclamationToken');
      continue;
    }
    fields.push(key);
  }
  if (aliases)
    for (const key of Object.keys(aliases))
      if (aliases[key] !== false) fields.push(key);
  if (
    kind >= SyntaxKind.FirstLiteralToken &&
    kind <= SyntaxKind.LastTemplateToken
  )
    fields.push('text');
  if (
    kind >= SyntaxKind.FirstTemplateToken &&
    kind <= SyntaxKind.LastTemplateToken
  )
    fields.push('rawText', 'templateFlags');
  for (const key of ['operator', 'token', 'keywordToken'])
    if (Reflect.get(raw, key) !== undefined) fields.push(key);
  if (jsDocKinds[SyntaxKind[kind]] || raw.flags & NodeFlags.HasJSDoc)
    fields.push('jsDoc');
  return fields;
}

function syntaxView(
  load: () => Node,
  read: (key: string | symbol, proxy: object) => unknown,
  ensure: () => void,
  jsDocKinds: JsDocKinds,
): object {
  const proxy: object = new Proxy(
    {},
    {
      get(target, key, receiver) {
        ensure();
        // TypeScript helpers write node-local caches, such as modifier flags.
        return Reflect.has(target, key)
          ? Reflect.get(target, key, receiver)
          : read(key, proxy);
      },
      has(target, key) {
        ensure();
        if (Reflect.has(target, key)) return true;
        const node = load();
        return (
          syntaxFields(node, jsDocKinds).includes(String(key)) ||
          (key === 'text' &&
            (node.kind === SyntaxKind.Identifier ||
              node.kind === SyntaxKind.PrivateIdentifier)) ||
          typeof Reflect.get(node, key) === 'function'
        );
      },
      ownKeys(target) {
        ensure();
        return [
          ...new Set([
            ...syntaxFields(load(), jsDocKinds),
            ...Reflect.ownKeys(target),
          ]),
        ];
      },
      getOwnPropertyDescriptor(target, key) {
        ensure();
        return (
          Reflect.getOwnPropertyDescriptor(target, key) ??
          (syntaxFields(load(), jsDocKinds).includes(String(key))
            ? {
                configurable: true,
                enumerable: true,
                writable: true,
                value: read(key, proxy),
              }
            : undefined)
        );
      },
    },
  );
  return proxy;
}

function wrapList(
  value: unknown[],
  wrap: (item: unknown) => unknown,
): unknown[] {
  const list: unknown[] = [];
  for (const item of value) {
    const visible = wrap(item);
    if (visible !== undefined) list.push(visible);
  }
  for (const key of ['pos', 'end', 'hasTrailingComma'])
    if (key in value) {
      const field = Reflect.get(value, key);
      Reflect.set(
        list,
        key,
        key === 'hasTrailingComma' ? field : (field as number) | 0,
      );
    }
  return list;
}

function jsDocComment(node: Node) {
  const parts = Reflect.get(node, 'comment') as readonly Node[] | undefined;
  if (!parts?.length) return undefined;
  return parts.length === 1 && parts[0].kind === SyntaxKind.JSDocText
    ? (Reflect.get(parts[0], 'text') as string)
    : parts;
}

function forEachSyntaxChild(
  node: Node,
  wrap: (item: unknown) => unknown,
  cb: (node: unknown) => unknown,
  cbNodes?: (nodes: unknown) => unknown,
): unknown {
  const visit = (child: Node | readonly Node[], callback = cb) => {
    const visible = wrap(child);
    return visible === undefined ? undefined : callback(visible);
  };
  const visitList = (list: readonly Node[]): unknown => {
    if (cbNodes) return visit(list, cbNodes);
    for (const child of list) {
      const result = visit(child);
      if (result) return result;
    }
    return undefined;
  };
  const kind = node.kind;
  if (kind === SyntaxKind.JSDocSignature) return node.forEachChild(visit);
  if (
    kind !== SyntaxKind.JSDoc &&
    (kind < SyntaxKind.FirstJSDocTagNode || kind > SyntaxKind.LastJSDocTagNode)
  )
    return node.forEachChild(visit, cbNodes && visitList);

  let fields = childProperties[kind];
  if (
    (kind === SyntaxKind.JSDocParameterTag ||
      kind === SyntaxKind.JSDocPropertyTag) &&
    !Reflect.get(node, 'isNameFirst')
  )
    fields = ['tagName', 'typeExpression', 'name', 'comment'];
  if (
    kind === SyntaxKind.JSDocCallbackTag ||
    (kind === SyntaxKind.JSDocTypedefTag &&
      Reflect.get(node, 'typeExpression')?.kind !==
        SyntaxKind.JSDocTypeExpression)
  )
    fields = ['tagName', 'name', 'typeExpression', 'comment'];
  for (const field of fields ?? []) {
    if (!field) continue;
    const value: unknown =
      field === 'comment' ? jsDocComment(node) : Reflect.get(node, field);
    if (!value || typeof value === 'string') continue;
    const result = Array.isArray(value)
      ? visitList(value)
      : visit(value as Node);
    if (result) return result;
  }
  return undefined;
}

function isJSDocCast(node: Node | undefined): boolean {
  return !!(
    node &&
    (node.kind === SyntaxKind.AsExpression ||
      node.kind === SyntaxKind.SatisfiesExpression) &&
    node.flags & NodeFlags.JavaScriptFile &&
    (Reflect.get(node, 'type') as Node).flags & NodeFlags.Reparsed
  );
}

// These literals use TypeScript's NodeObject methods, like non-token nodes.
function needsRealPosition(kind: SyntaxKind): boolean {
  return (
    kind > SyntaxKind.LastToken ||
    kind === SyntaxKind.StringLiteral ||
    kind === SyntaxKind.NumericLiteral ||
    kind === SyntaxKind.NoSubstitutionTemplateLiteral
  );
}

/** A worker/generation owns stable native identities and a bounded decoded
 * source LRU. Proxies retain identities, never a dependency source buffer. */
export class GenerationSyntaxViews {
  private readonly identities = new WeakMap<object, SyntaxRef>();
  private readonly wrapped = new WeakObjectCache<string, object>();
  private readonly generatedOrigins = new WeakMap<object, NodeRef | null>();
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
  private oversizedSource?: WeakRef<RemoteSourceFile>;
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
      TypeServiceCompilerData,
      'SyntaxKind' | 'NodeFlags' | 'ScriptTarget' | 'jsDocKinds'
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
  sourceLocation(node: object): NodeRef | undefined {
    this.ensure();
    const ref = this.identities.get(node);
    if (!ref) {
      if (this.generatedOrigins.has(node)) return undefined;
      throw new Error('foreign compiler node handle');
    }
    if (!('owner' in ref)) return ref;
    const { kind, pos, end } = this.rawSyntax(ref);
    return { ...ref.owner, token: { kind, pos, end } };
  }
  parseTreeLocation(node: object): NodeRef | undefined {
    return (
      this.sourceLocation(node) ?? this.generatedOrigins.get(node) ?? undefined
    );
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
      return entry.source;
    }
    const oversized = this.oversizedSource?.deref();
    return oversized?.fileName === file ? oversized : undefined;
  }
  private retain(source: RemoteSourceFile, bytes: number): void {
    if (bytes > this.maxSourceBytes) {
      // Reuse consecutive reads of a large dependency without pinning its buffer.
      this.oversizedSource = new WeakRef(source);
      return;
    }
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
  private isHeritage(fact: NodeFact): boolean {
    while (fact.kind === 'QualifiedName' && fact.parent)
      fact = this.summary(fact.parent);
    return Boolean(
      fact.kind === 'TypeReference' &&
      fact.parent &&
      this.summary(fact.parent).kind === 'HeritageClause',
    );
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
    const proxy = syntaxView(
      () => this.rawSyntax(identity),
      (key, proxy) => this.get(identity, key, proxy),
      this.ensure,
      this.typescript.jsDocKinds,
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
    while (isJSDocCast(value)) value = Reflect.get(value, 'expression') as Node;
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
        const list = wrapList(value, (item) => this.wrap(item));
        if (
          value.length &&
          !list.length &&
          (Reflect.get(value.parent, 'typeParameters') === value ||
            Reflect.get(value.parent, 'typeArguments') === value ||
            Reflect.get(value.parent, 'heritageClauses') === value ||
            Reflect.get(value.parent, 'modifiers') === value)
        )
          return undefined;
        this.wrapped.set(key, list);
        return list;
      }
      return wrapList(value, (item) => this.wrap(item));
    }
    if (value instanceof RemoteNode)
      return value.flags & NodeFlags.Reparsed
        ? undefined
        : this.wrapNode(value);
    if (
      typeof value === 'object' &&
      value !== null &&
      'kind' in value &&
      'getSourceFile' in value
    )
      return this.wrapNode(value as Node);
    return value;
  }
  private get(ref: SyntaxRef, key: string | symbol, proxy: object): unknown {
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
      !this.cachedSource(file) &&
      !this.configs.has(file) &&
      ['kind', 'flags', 'parent', 'name', 'text', 'escapedText'].includes(
        String(key),
      )
    ) {
      const fact = this.summary(real);
      if (key === 'kind') {
        if (fact.kind === 'TypeReference' && this.isHeritage(fact))
          return this.typescript.SyntaxKind.ExpressionWithTypeArguments;
        if (fact.kind === 'QualifiedName' && this.isHeritage(fact))
          return this.typescript.SyntaxKind.PropertyAccessExpression;
        return this.translateKind(fact.kind);
      }
      if (key === 'flags')
        return translateNodeFlags(
          fact.flags,
          this.typescript.NodeFlags,
          fact.moduleKeyword,
        );
      if (key === 'parent') return this.nodeRef(fact.parent);
      if (
        key === 'name' &&
        !(fact.kind === 'QualifiedName' && this.isHeritage(fact))
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
        return forEachSyntaxChild(
          this.rawSyntax(ref),
          (item) => this.wrap(item),
          cb,
          cbNodes,
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
    if (key === 'parent') {
      let parent = raw.parent;
      while (parent && isJSDocCast(parent)) parent = parent.parent;
      return parent;
    }
    if (key === 'postfixToken') return undefined;
    const kind = publicSyntaxKind(raw);
    if (key === 'jsDocPropertyTags' && kind === SyntaxKind.JSDocTypeLiteral) {
      // This native field is an unwrapped slice of child nodes.
      const tags: Node[] = [];
      raw.forEachChild((tag) => {
        tags.push(tag);
      });
      return tags.length ? tags : undefined;
    }
    if (
      key === 'name' &&
      (kind === SyntaxKind.JSDocTypedefTag ||
        kind === SyntaxKind.JSDocCallbackTag)
    ) {
      let name = Reflect.get(raw, 'name') as Node | undefined;
      while (name?.kind === SyntaxKind.ModuleDeclaration)
        name = (Reflect.get(name, 'body') ?? Reflect.get(name, 'name')) as Node;
      return name;
    }
    if (
      key === 'comment' &&
      kind >= SyntaxKind.FirstJSDocNode &&
      kind <= SyntaxKind.LastJSDocNode
    ) {
      return jsDocComment(raw);
    }
    if (key === 'hasExtendedUnicodeEscape' && kind === SyntaxKind.StringLiteral)
      return !!(
        (Reflect.get(raw, 'tokenFlags') as number) &
        TokenFlags.ExtendedUnicodeEscape
      );
    const alias =
      typeof key === 'string' ? fieldAliases[kind]?.[key] : undefined;
    if (alias === false) return undefined;
    if (typeof alias === 'string') return Reflect.get(raw, alias);
    if (kind !== raw.kind) {
      if (key === 'expression')
        return Reflect.get(
          raw,
          kind === SyntaxKind.PropertyAccessExpression ? 'left' : 'typeName',
        );
      if (key === 'name' && kind === SyntaxKind.PropertyAccessExpression)
        return Reflect.get(raw, 'right');
      if (key === 'typeName' || key === 'left' || key === 'right')
        return undefined;
    }
    if (key === 'kind') return this.translateKind(SyntaxKind[kind]);
    if (
      (key === 'questionToken' || key === 'exclamationToken') &&
      childProperties[raw.kind]?.includes('postfixToken')
    ) {
      const token = Reflect.get(raw, 'postfixToken') as Node | undefined;
      return token?.kind ===
        (key === 'questionToken'
          ? SyntaxKind.QuestionToken
          : SyntaxKind.ExclamationToken)
        ? token
        : undefined;
    }
    if (
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
    return Reflect.get(raw, key);
  }

  // Generated nodes keep their own syntax identity. Checker queries can follow
  // an original parse node when the compiler supplies one.
  generatedTypeNode({ tree, origins }: GeneratedSyntax): object {
    const reader = new RemoteSourceFile(
      tree,
      new Wtf8Decoder('utf-8', { ignoreBOM: true }),
    );
    // The decoder's first slot holds a generated node, with no SourceFile fields.
    const root = new RemoteNode(
      reader.view,
      1,
      reader,
      reader,
      reader._offsetNodes,
    );
    reader.nodes[1] = root;
    const views = new WeakMap<object, object>();
    const getSourceFile = () => {
      this.ensure();
      return undefined;
    };
    // The reader uses these caches for recursive navigation. Generated tokens
    // need their own parents even when their spans match tokens in the file.
    const scanSource = (arg: unknown) => {
      if (typeof arg !== 'object' || arg === null) return arg;
      let source = views.get(arg);
      if (!source) {
        source = Object.create(arg, {
          tokenCache: { writable: true },
          childrenCache: { writable: true },
        }) as object;
        views.set(arg, source);
      }
      return source;
    };
    const wrap = (value: unknown): unknown => {
      if (typeof value !== 'object' || value === null) return value;
      const old = views.get(value);
      if (old) return old;
      if (Array.isArray(value)) {
        const list = wrapList(value, wrap);
        views.set(value, list);
        return list;
      }
      if (!('kind' in value && 'getSourceFile' in value)) return value;
      const node = value as Node;
      // The packet uses a RemoteSourceFile as its decoder, not as a syntax parent.
      Object.defineProperty(node, 'getSourceFile', { value: getSourceFile });
      const encoded = node instanceof RemoteNode;
      const original = encoded
        ? origins?.[(node as unknown as { index: number }).index]
        : undefined;
      const proxy = syntaxView(
        () => node,
        (key) => {
          if (key === 'original') return this.nodeRef(original);
          if (
            key === 'hasExtendedUnicodeEscape' &&
            encoded &&
            !original &&
            node.kind === SyntaxKind.StringLiteral
          )
            return undefined;
          if (key === 'parent') return encoded ? undefined : wrap(node.parent);
          if (key === 'flags') {
            const flags = this.syntaxMember(node, key) as number;
            return encoded
              ? flags | this.typescript.NodeFlags.Synthesized
              : flags;
          }
          if (key === 'getSourceFile') return getSourceFile;
          if (key === 'forEachChild')
            return (
              cb: (n: unknown) => unknown,
              cbNodes?: (nodes: unknown) => unknown,
            ) => {
              this.ensure();
              return forEachSyntaxChild(node, wrap, cb, cbNodes);
            };
          const member = this.syntaxMember(node, key);
          if (typeof member === 'function')
            return (...args: unknown[]) => {
              this.ensure();
              if (
                needsRealPosition(node.kind) &&
                (node.pos < 0 || node.end < 0)
              )
                throw new Error(
                  'Node must have a real position for this operation',
                );
              if (tokenNavigation.has(key)) {
                const index = key === 'getChildAt' ? 1 : 0;
                if (args[index] !== undefined)
                  args[index] = scanSource(args[index]);
              }
              return wrap(Reflect.apply(member, node, args));
            };
          return wrap(member);
        },
        this.ensure,
        this.typescript.jsDocKinds,
      );
      this.generatedOrigins.set(proxy, original ?? null);
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
    this.oversizedSource = undefined;
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
