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
  NodeFact,
  NodeRef,
  SourceDescriptor,
  SyntaxSource,
  TypeServiceDescriptor,
  TypeServiceProjection,
  SequenceAlias,
  ChainAlias,
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

function getProp(target: object, key: PropertyKey): unknown {
  return Reflect.get(target, key);
}

function isObject(value: unknown): value is object {
  return typeof value === 'object' && value !== null;
}

function isCompilerNode(value: unknown): value is Node {
  return isObject(value) && 'kind' in value && 'getSourceFile' in value;
}

function isSyntheticRef(ref: SyntaxRef): ref is SyntheticRef {
  return 'owner' in ref;
}

function unsupported(name: string): never {
  throw new Error(`unsupported TypeScript operation: ${name}`);
}

const tokenNavigation = new Set<PropertyKey>([
  'getChildren',
  'getChildAt',
  'getChildCount',
  'getFirstToken',
  'getLastToken',
]);

// Native bits that either have no JS equivalent or depend on the node kind.
const specialNodeFlags =
  NodeFlags.Reparsed |
  NodeFlags.ReparserTransformedLiteral |
  NodeFlags.HasJSDoc |
  NodeFlags.OptionalChain |
  NodeFlags.PossiblyContainsDeprecatedTag;

const commonNodeFlags: readonly (readonly [string, number])[] = Object.entries(
  NodeFlags,
).filter((entry): entry is [string, number] => {
  const value = entry[1];
  return (
    typeof value === 'number' &&
    value > NodeFlags.None &&
    (value & (value - 1)) === 0 &&
    !(value & specialNodeFlags)
  );
});

function translateNodeFlags(
  flags: number,
  table: Readonly<Record<string, unknown>>,
  moduleKeyword: SyntaxKind = SyntaxKind.Unknown,
): number {
  const readFlag = (name: string): number => {
    const val = table[name];
    if (typeof val !== 'number') unsupported(`enum.${name}`);
    return val;
  };
  let result = 0;
  if (flags & NodeFlags.OptionalChain) {
    result |= moduleKeyword
      ? readFlag('NestedNamespace')
      : readFlag('OptionalChain');
  }
  if (moduleKeyword === SyntaxKind.NamespaceKeyword) {
    result |= readFlag('Namespace');
  }
  if (moduleKeyword === SyntaxKind.GlobalKeyword) {
    result |= readFlag('GlobalAugmentation');
  }
  if (flags & NodeFlags.PossiblyContainsDeprecatedTag) {
    result |= readFlag('Deprecated');
  }
  for (const [name, value] of commonNodeFlags) {
    if (flags & value) {
      result |= readFlag(name);
    }
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
  'checkJsDirective',
  'impliedNodeFormat',
  'referencedFiles',
  'typeReferenceDirectives',
  'libReferenceDirectives',
  'imports',
  'moduleAugmentations',
  'ambientModuleNames',
  'externalModuleIndicator',
] as const;

function publicSyntaxKind(node: Readonly<Node>): SyntaxKind {
  if (
    node.kind === SyntaxKind.QualifiedName ||
    node.kind === SyntaxKind.TypeReference
  ) {
    let parent: Node | undefined = node;
    while (parent?.kind === SyntaxKind.QualifiedName) parent = parent.parent;
    if (
      parent?.kind === SyntaxKind.TypeReference &&
      parent.parent?.kind === SyntaxKind.HeritageClause
    ) {
      return node.kind === SyntaxKind.TypeReference
        ? SyntaxKind.ExpressionWithTypeArguments
        : SyntaxKind.PropertyAccessExpression;
    }
  }
  return node.kind;
}

// The compiler's child table excludes scalar fields exposed by the JS AST.
const scalarFields: Readonly<Partial<Record<SyntaxKind, readonly string[]>>> = {
  [SyntaxKind.MethodDeclaration]: ['symbol'],
  [SyntaxKind.MethodSignature]: ['symbol'],
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
};

type JsDocKinds = Readonly<Record<string, boolean>>;

// Only differences from the reader's child table belong here. A string names
// the native slot, and false omits a slot reserved by the Go parser.
const fieldAliases: Readonly<
  Partial<Record<SyntaxKind, Readonly<Record<string, string | false>>>>
> = {
  [SyntaxKind.SyntaxList]: { _children: 'children', children: false },
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

function syntaxFields(raw: Readonly<Node>, jsDocKinds: JsDocKinds): string[] {
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
      ) {
        fields.push('exclamationToken');
      }
      continue;
    }
    fields.push(key);
  }
  if (aliases) {
    for (const key of Object.keys(aliases)) {
      if (aliases[key] !== false) fields.push(key);
    }
  }
  if (
    kind >= SyntaxKind.FirstLiteralToken &&
    kind <= SyntaxKind.LastTemplateToken
  ) {
    fields.push('text');
  }
  if (
    kind >= SyntaxKind.FirstTemplateToken &&
    kind <= SyntaxKind.LastTemplateToken
  ) {
    fields.push('rawText', 'templateFlags');
  }
  for (const key of ['operator', 'token', 'keywordToken']) {
    if (getProp(raw, key) !== undefined) fields.push(key);
  }
  const kindName = SyntaxKind[kind];
  if ((kindName && jsDocKinds[kindName]) || raw.flags & NodeFlags.HasJSDoc) {
    fields.push('jsDoc');
  }
  return fields;
}

type ReadSyntax = (
  key: string | symbol,
  proxy: object,
  args?: readonly unknown[],
) => unknown;

const readSyntax = Symbol('readSyntax');
type SyntaxMethod = (
  this: Readonly<{ [readSyntax]?: ReadSyntax }>,
  ...args: readonly unknown[]
) => unknown;
const syntaxMethods = new Map<PropertyKey, SyntaxMethod>();

function syntaxMethod(key: string | symbol): SyntaxMethod {
  let method = syntaxMethods.get(key);
  if (!method) {
    method = function (...args) {
      const read = this[readSyntax];
      if (read) return read(key, this, args);
      if (key === 'update') return unsupported('SourceFile.update');
      const proto: unknown = Object.getPrototypeOf(this);
      const native =
        (isObject(proto) ? getProp(proto, key) : undefined) ??
        getProp(this, key);
      const fn =
        typeof native === 'function' && native !== method
          ? native
          : getProp(RemoteNode.prototype, key);
      if (typeof fn !== 'function') {
        throw new Error(`missing syntax method: ${String(key)}`);
      }
      return Reflect.apply(fn, this, [...args]);
    };
    syntaxMethods.set(key, method);
  }
  return method;
}

function positionMethod(key: string | symbol): unknown {
  switch (key) {
    case 'getText':
    case 'getFullText':
    case 'getFullStart':
    case 'getEnd':
    case 'getWidth':
    case 'getFullWidth':
    case 'getLeadingTriviaWidth':
      return getProp(RemoteNode.prototype, key);
    default:
      return undefined;
  }
}

function syntaxView(
  load: () => Node,
  read: ReadSyntax,
  ensure: () => void,
  jsDocKinds: JsDocKinds,
  target: object = {},
): object {
  const proxy: object = new Proxy(target, {
    get(t, key, receiver) {
      ensure();
      if (key === readSyntax) return read;
      // TypeScript helpers write node-local caches such as modifier flags.
      if (Reflect.has(t, key)) {
        const cached: unknown = Reflect.get(t, key, receiver);
        return cached;
      }
      if (key === 'assertHasRealPosition') {
        return needsRealPosition(load()) ? assertHasRealPosition : undefined;
      }
      return read(key, proxy);
    },
    set(t, key, value, receiver) {
      ensure();
      if (receiver === proxy && !Reflect.has(t, key)) {
        return Reflect.set(t, key, value);
      }
      return Reflect.set(t, key, value, receiver);
    },
    has(t, key) {
      ensure();
      if (Reflect.has(t, key)) return true;
      const node = load();
      if (key === 'assertHasRealPosition') return needsRealPosition(node);
      return (
        syntaxFields(node, jsDocKinds).includes(String(key)) ||
        (key === 'text' &&
          (node.kind === SyntaxKind.Identifier ||
            node.kind === SyntaxKind.PrivateIdentifier)) ||
        typeof getProp(node, key) === 'function'
      );
    },
    ownKeys(t) {
      ensure();
      return [
        ...new Set([
          ...syntaxFields(load(), jsDocKinds),
          ...Reflect.ownKeys(t),
        ]),
      ];
    },
    getOwnPropertyDescriptor(t, key) {
      ensure();
      return (
        Reflect.getOwnPropertyDescriptor(t, key) ??
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
  });
  return proxy;
}

function wrapList(
  value: readonly unknown[],
  wrap: (item: unknown) => unknown,
): unknown[] & Partial<ts.TextRange> {
  const list: unknown[] = [];
  for (const item of value) {
    const visible = wrap(item);
    if (visible !== undefined) list.push(visible);
  }
  for (const key of ['pos', 'end', 'hasTrailingComma']) {
    if (key in value) {
      const field = getProp(value, key);
      Reflect.set(
        list,
        key,
        key === 'hasTrailingComma'
          ? field
          : typeof field === 'number'
            ? field | 0
            : 0,
      );
    }
  }
  return list;
}

function shareGeneratedRange(
  value: object,
  range: Readonly<Partial<ts.ReadonlyTextRange>>,
): void {
  Object.defineProperties(value, {
    pos: { get: () => range.pos },
    end: { get: () => range.end },
  });
}

function jsDocComment(node: Readonly<Node>): unknown {
  const parts = getProp(node, 'comment');
  if (!Array.isArray(parts) || parts.length === 0) return undefined;
  const first: unknown = parts[0];
  return parts.length === 1 &&
    isCompilerNode(first) &&
    first.kind === SyntaxKind.JSDocText
    ? getProp(first, 'text')
    : parts;
}

function forEachSyntaxChild(
  node: Readonly<Node>,
  wrap: (item: unknown) => unknown,
  args: readonly unknown[],
): unknown {
  const cb = args[0];
  const cbNodes = args[1];
  if (typeof cb !== 'function') return undefined;
  const invoke = (fn: typeof cb, arg: unknown): unknown =>
    Reflect.apply(fn, undefined, [arg]);
  const listCallback = typeof cbNodes === 'function' ? cbNodes : undefined;

  const visit = (child: unknown, fn = cb): unknown => {
    const visible = wrap(child);
    return visible === undefined ? undefined : invoke(fn, visible);
  };
  const visitList = (list: readonly Node[]): unknown => {
    if (listCallback) return visit(list, listCallback);
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
  ) {
    return node.forEachChild(visit, listCallback && visitList);
  }

  let fields: readonly (string | undefined)[] | undefined =
    childProperties[kind];
  if (
    (kind === SyntaxKind.JSDocParameterTag ||
      kind === SyntaxKind.JSDocPropertyTag) &&
    !getProp(node, 'isNameFirst')
  ) {
    fields = ['tagName', 'typeExpression', 'name', 'comment'];
  }
  const typeExpr = getProp(node, 'typeExpression');
  if (
    kind === SyntaxKind.JSDocCallbackTag ||
    (kind === SyntaxKind.JSDocTypedefTag &&
      (!isCompilerNode(typeExpr) ||
        typeExpr.kind !== SyntaxKind.JSDocTypeExpression))
  ) {
    fields = ['tagName', 'name', 'typeExpression', 'comment'];
  }
  for (const field of fields ?? []) {
    if (!field) continue;
    const value =
      field === 'comment' ? jsDocComment(node) : getProp(node, field);
    if (!value || typeof value === 'string') continue;
    const result = Array.isArray(value)
      ? visitList(value as readonly Node[])
      : visit(value);
    if (result) return result;
  }
  return undefined;
}

function isJSDocCast(node: Readonly<Node> | undefined): boolean {
  if (
    !node ||
    (node.kind !== SyntaxKind.AsExpression &&
      node.kind !== SyntaxKind.SatisfiesExpression) ||
    !(node.flags & NodeFlags.JavaScriptFile)
  ) {
    return false;
  }
  const typeNode = getProp(node, 'type');
  return isCompilerNode(typeNode) && !!(typeNode.flags & NodeFlags.Reparsed);
}

function needsRealPosition(node: Readonly<Pick<Node, 'kind'>>): boolean {
  const { kind } = node;
  return (
    kind > SyntaxKind.LastToken ||
    ((kind === SyntaxKind.StringLiteral ||
      kind === SyntaxKind.NumericLiteral ||
      kind === SyntaxKind.JsxText ||
      kind === SyntaxKind.NoSubstitutionTemplateLiteral) &&
      node instanceof RemoteNode)
  );
}

function assertHasRealPosition(
  this: Readonly<ts.TextRange>,
  message?: string,
): void {
  if (!(this.pos >= 0 && this.end >= 0)) {
    throw new Error(
      message ?? 'Node must have a real position for this operation',
    );
  }
}

function readSubtree(tree: Readonly<Uint8Array>): RemoteNode {
  const reader = new RemoteSourceFile(
    tree,
    new Wtf8Decoder('utf-8', { ignoreBOM: true }),
  );
  const root = new RemoteNode(
    reader.view,
    1,
    reader,
    reader,
    reader._offsetNodes,
  );
  reader.nodes[1] = root;
  return root;
}

function remoteIndex(node: object): number {
  const idx = getProp(node, 'index');
  if (typeof idx !== 'number') {
    throw new Error('missing encoded compiler node index');
  }
  return idx;
}

/**
 * A generation owns canonical TS syntax proxies and a bounded decoded source
 * LRU. Proxies retain lightweight NodeRef identities, never raw source buffers.
 */
export class GenerationSyntaxViews {
  private readonly identities = new WeakMap<object, SyntaxRef>();
  private readonly wrapped = new WeakObjectCache<string, object>();
  private readonly generated = new WeakSet<object>();
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
  private oversizedSource?: {
    key: string;
    source: WeakRef<RemoteSourceFile>;
  };
  private readonly summaries = new Map<
    string,
    { fact: NodeFact; bytes: number }
  >();
  private readonly sourceOrdinals = new Map<string, number>();
  private readonly metadataByName = new Map<string, SourceDescriptor>();
  private readonly diagnosticMetadata = new Map<string, SourceDescriptor>();
  private sourceBytes = 0;
  private summaryBytes = 0;

  constructor(
    private readonly typescript: Pick<
      TypeServiceCompilerData,
      | 'SyntaxKind'
      | 'NodeFlags'
      | 'ScriptTarget'
      | 'builderEmitFlags'
      | 'jsDocKinds'
    >,
    private readonly descriptor: TypeServiceDescriptor,
    private lookup: SyntaxSource | undefined,
    private readonly readNodes: (node: NodeRef) => readonly NodeFact[],
    private readonly ensure: () => void,
    private readonly maxSourceBytes: number,
    private readonly maxSummaryBytes: number,
    private readonly readDeclarationSymbol: (
      node: NodeRef,
    ) => object | undefined,
  ) {
    for (const [index, source] of descriptor.sources.entries()) {
      this.sourceOrdinals.set(source.fileName, index + 1);
      this.metadataByName.set(source.fileName, source);
      this.metadataByName.set(source.path, source);
    }
    for (const source of descriptor.diagnosticSources ?? []) {
      this.registerDiagnosticSource(source);
    }
  }

  private registerDiagnosticSource(source: SourceDescriptor): void {
    if (!this.sourceOrdinals.has(source.fileName)) {
      this.sourceOrdinals.set(source.fileName, this.sourceOrdinals.size + 1);
    }
    this.diagnosticMetadata.set(source.fileName, source);
    if (source.path) this.diagnosticMetadata.set(source.path, source);
  }

  private ordinal(ref: NodeRef): number | undefined {
    return this.sourceOrdinals.get(ref.file);
  }

  private key(ref: NodeRef): string {
    return `${this.ordinal(ref) ?? 0}:${ref.index}`;
  }

  locationOrUndefined(node: object): NodeRef | undefined {
    const ref = this.identities.get(node);
    return ref && !isSyntheticRef(ref) ? ref : undefined;
  }

  isStableNodeQuery(op: 'type' | 'constantValue', ref: NodeRef): boolean {
    if (ref.token) return false;
    const source = this.cachedSource(ref.file);
    const node =
      source && ref.index < source.nodes.length
        ? source.getOrCreateNodeAtIndex(ref.index)
        : undefined;
    if (op === 'constantValue') return node?.kind === SyntaxKind.EnumMember;
    return (
      node?.kind === SyntaxKind.Identifier &&
      node.parent?.kind === SyntaxKind.VariableDeclaration &&
      getProp(node.parent, 'name') === node
    );
  }

  sourceLocation(node: object): NodeRef | undefined {
    this.ensure();
    const ref = this.identities.get(node);
    if (!ref) {
      const flags = getProp(node, 'flags');
      const isSynthesized =
        typeof flags === 'number' &&
        !!(flags & this.typescript.NodeFlags.Synthesized);
      if (this.generated.has(node) || isSynthesized) return undefined;

      // Support scanned token nodes borrowed via getChildren.
      const tokenKind = getProp(node, 'kind');
      const jsKind = this.typescript.SyntaxKind;
      if (
        typeof tokenKind === 'number' &&
        (tokenKind <= Number(jsKind.LastToken) ||
          tokenKind === Number(jsKind.SyntaxList))
      ) {
        const parentObj = getProp(node, 'parent');
        const parent = isObject(parentObj)
          ? this.locationOrUndefined(parentObj)
          : undefined;
        if (parent) {
          const tokenPos = getProp(node, 'pos');
          const tokenEnd = getProp(node, 'end');
          if (typeof tokenPos === 'number' && typeof tokenEnd === 'number') {
            const children = this.rawNode(parent).getChildren();
            const matches = (child: Readonly<Node>) =>
              child.pos === tokenPos &&
              child.end === tokenEnd &&
              Number(this.translateKind(SyntaxKind[child.kind])) === tokenKind;
            let child = children.find(matches);
            if (!child) {
              const list = children.find(
                (n) =>
                  n.kind === SyntaxKind.SyntaxList &&
                  n.pos <= tokenPos &&
                  tokenEnd <= n.end,
              );
              child = list?.getChildren().find(matches);
            }
            if (child instanceof RemoteNode) return this.ref(child);
            if (child && child.parent instanceof RemoteNode) {
              const { kind, pos, end } = child;
              return { ...this.ref(child.parent), token: { kind, pos, end } };
            }
          }
        } else if (isObject(parentObj)) {
          const parentFlags = getProp(parentObj, 'flags');
          if (
            (typeof parentFlags === 'number' &&
              !!(parentFlags & this.typescript.NodeFlags.Synthesized)) ||
            this.generated.has(parentObj)
          ) {
            return undefined;
          }
        }
      }
      throw new Error('foreign compiler node handle');
    }
    if (!isSyntheticRef(ref)) return ref;
    const { kind, pos, end } = this.rawSyntax(ref);
    return { ...ref.owner, token: { kind, pos, end } };
  }

  parseTreeLocation(node: object | undefined): NodeRef | undefined {
    if (!node) return undefined;
    const ref = this.sourceLocation(node);
    if (ref) return ref;
    let original = getProp(node, 'original');
    while (isObject(original)) {
      const source = this.sourceLocation(original);
      if (source) return source;
      original = getProp(original, 'original');
    }
    return undefined;
  }

  location(node: object): NodeRef {
    const ref = this.parseTreeLocation(node);
    if (!ref) throw new Error('generated compiler node has no source location');
    return ref;
  }

  private lookupMetadata(file: string): SourceDescriptor | undefined {
    return this.metadataByName.get(file) ?? this.diagnosticMetadata.get(file);
  }

  metadata(node: object): SourceDescriptor {
    const ref = this.location(node);
    const metadata = this.lookupMetadata(ref.file);
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
    return source?.path === path
      ? this.nodeRef({ file: source.fileName, index: 1 })
      : undefined;
  }

  sourceView(file: string): object | undefined {
    this.ensure();
    let metadata = this.metadataByName.get(file);
    if (!metadata && this.lookup) {
      const reply = this.lookup(this.descriptor, file, true);
      metadata = reply && this.metadataByName.get(reply.fileName);
      if (metadata) this.metadataByName.set(file, metadata);
    }
    return metadata
      ? this.nodeRef({ file: metadata.fileName, index: 1 })
      : undefined;
  }

  diagnosticSourceView(file: string): object | undefined {
    this.ensure();
    const programSource = this.sourceView(file);
    if (programSource) return programSource;
    let metadata = this.diagnosticMetadata.get(file);
    if (!metadata && this.lookup) {
      const reply = this.lookup(this.descriptor, file, true);
      if (reply) {
        metadata = this.diagnosticMetadata.get(reply.fileName) ?? {
          fileName: reply.fileName,
          path: reply.path ?? reply.fileName,
          external: false,
          defaultLibrary: false,
          isDeclarationFile: false,
        };
        this.registerDiagnosticSource(metadata);
        this.diagnosticMetadata.set(file, metadata);
      }
    }
    return metadata
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
    const oversized = this.oversizedSource;
    return oversized?.key === file ? oversized.source.deref() : undefined;
  }

  private retain(source: RemoteSourceFile, bytes: number): void {
    const key = source.fileName;
    if (bytes > this.maxSourceBytes) {
      this.oversizedSource = { key, source: new WeakRef(source) };
      return;
    }
    while (this.sourceBytes + bytes > this.maxSourceBytes) {
      const oldest = this.sources.entries().next().value;
      if (!oldest) break;
      this.sources.delete(oldest[0]);
      this.sourceBytes -= oldest[1].bytes;
    }
    this.sources.set(key, { source, bytes });
    this.sourceBytes += bytes;
  }

  attach(
    tree: Readonly<Uint8Array>,
    expected?: Readonly<{ file: string; text: string }>,
  ): RemoteSourceFile {
    const source = new RemoteSourceFile(
      Uint8Array.from(tree),
      new Wtf8Decoder('utf-8', { ignoreBOM: true }),
    );
    if (
      expected &&
      (source.fileName !== expected.file || source.text !== expected.text)
    ) {
      throw new Error('type service projection belongs to a different source');
    }
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
    const metadata = this.lookupMetadata(file);
    if (!metadata) return undefined;
    const old = this.cachedSource(metadata.fileName);
    if (old) return old;
    if (!this.lookup) return undefined;
    const reply = this.lookup(this.descriptor, metadata.fileName);
    if (!reply) return undefined;
    if (reply.fileName !== metadata.fileName) {
      throw new Error('syntax source identity changed');
    }
    if (!reply.tree) throw new Error('missing source syntax packet');
    return this.attach(reply.tree);
  }

  private rawNode(ref: NodeRef): RemoteNode {
    const source = this.source(ref.file);
    if (!source) throw new Error('foreign source');
    const raw = source.getOrCreateNodeAtIndex(ref.index);
    if (!(raw instanceof RemoteNode)) {
      throw new Error('invalid compiler node handle');
    }
    return raw;
  }

  private rawSyntax(ref: SyntaxRef): Node {
    if (!isSyntheticRef(ref)) return this.rawNode(ref);
    const child = this.rawNode(ref.owner).getChildAt(ref.child);
    return ref.listChild === undefined
      ? child
      : child.getChildAt(ref.listChild);
  }

  private positionedNode(ref: SyntaxRef): Node {
    const node = this.rawSyntax(ref);
    if (node.pos < 0 && needsRealPosition(node)) {
      assertHasRealPosition.call(node);
    }
    return node;
  }

  private summary(ref: NodeRef): NodeFact {
    const requestedKey = this.key(ref);
    const cached = this.summaries.get(requestedKey);
    if (cached) {
      this.summaries.delete(requestedKey);
      this.summaries.set(requestedKey, cached);
      return cached.fact;
    }
    let result: NodeFact | undefined;
    for (const fact of this.readNodes(ref)) {
      const key = this.key(fact.node);
      if (key === requestedKey) result = fact;
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
        const oldest = this.summaries.entries().next().value;
        if (!oldest) break;
        this.summaryBytes -= oldest[1].bytes;
        this.summaries.delete(oldest[0]);
      }
      this.summaries.set(key, { fact, bytes });
      this.summaryBytes += bytes;
    }
    if (!result) throw new Error('missing syntax fact');
    return result;
  }

  private isHeritage(initial: NodeFact): boolean {
    let fact = initial;
    while (fact.kind === 'QualifiedName' && fact.parent) {
      fact = this.summary(fact.parent);
    }
    return Boolean(
      fact.kind === 'TypeReference' &&
      fact.parent &&
      this.summary(fact.parent).kind === 'HeritageClause',
    );
  }

  private translateKind(name: string): ts.SyntaxKind {
    const normalized =
      name === 'EndOfFile'
        ? 'EndOfFileToken'
        : name === 'JSDocUnknownTag'
          ? 'JSDocTag'
          : name;
    const result = getProp(this.typescript.SyntaxKind, normalized);
    if (typeof result !== 'number') unsupported(`SyntaxKind.${normalized}`);
    return result;
  }

  nodeRef(ref: NodeRef | undefined): object | undefined {
    this.ensure();
    if (!ref) return undefined;
    const metadata = this.lookupMetadata(ref.file);
    if (!metadata) throw new Error('foreign source');
    const identity: NodeRef = { file: metadata.fileName, index: ref.index };
    return this.wrapRef(identity, this.key(identity));
  }

  private wrapRef(identity: SyntaxRef, cacheKey: string): object {
    const old = this.wrapped.get(cacheKey);
    if (old) return old;
    const proxy = syntaxView(
      () => this.rawSyntax(identity),
      (key, view, args) => this.get(identity, key, view, args),
      this.ensure,
      this.typescript.jsDocKinds,
    );
    this.identities.set(proxy, identity);
    this.wrapped.set(cacheKey, proxy);
    return proxy;
  }

  private ref(value: RemoteNode, index = remoteIndex(value)): NodeRef {
    return { file: value.getSourceFile().fileName, index };
  }

  private wrapNode(initial: Node): object {
    let value = initial;
    while (isJSDocCast(value)) {
      const expr = getProp(value, 'expression');
      if (!isCompilerNode(expr)) break;
      value = expr;
    }
    if (value instanceof RemoteNode) {
      const wrapped = this.nodeRef(this.ref(value));
      if (!wrapped) throw new Error('foreign source');
      return wrapped;
    }
    const parent = value.parent;
    if (!(parent instanceof RemoteNode)) {
      throw new Error('missing syntax owner');
    }
    let ref = this.syntheticRefs.get(value);
    if (!ref) {
      const owner = this.ref(parent);
      for (const [child, item] of parent.getChildren().entries()) {
        if (item instanceof RemoteNode) continue;
        this.syntheticRefs.set(item, { owner, child });
        if (item.kind !== SyntaxKind.SyntaxList) continue;
        for (const [listChild, token] of item.getChildren().entries()) {
          if (!(token instanceof RemoteNode)) {
            this.syntheticRefs.set(token, { owner, child, listChild });
          }
        }
      }
      ref = this.syntheticRefs.get(value);
      if (!ref) throw new Error('missing synthetic syntax node');
    }
    const key =
      value.kind === SyntaxKind.SyntaxList
        ? `list:${this.key(ref.owner)}:${ref.child}`
        : `token:${this.ordinal(ref.owner) ?? 0}:${value.pos}:${value.end}`;
    return this.wrapRef(ref, key);
  }

  firstToken(ref: NodeRef): object | undefined {
    const token = this.rawNode(ref).getFirstToken();
    return token && this.wrapNode(token);
  }

  wrap(value: unknown): unknown {
    if (Array.isArray(value)) {
      const parent = getProp(value, 'parent');
      if (parent instanceof RemoteNode) {
        const key = this.key(this.ref(parent, remoteIndex(value)));
        const found = this.wrapped.get(key);
        if (found) return found;
        const list = wrapList(value, (item) => this.wrap(item));
        if (
          value.length > 0 &&
          list.length === 0 &&
          (getProp(parent, 'typeParameters') === value ||
            getProp(parent, 'typeArguments') === value ||
            getProp(parent, 'heritageClauses') === value ||
            getProp(parent, 'modifiers') === value)
        ) {
          return undefined;
        }
        this.wrapped.set(key, list);
        return list;
      }
      return wrapList(value, (item) => this.wrap(item));
    }
    if (value instanceof RemoteNode) {
      return value.flags & NodeFlags.Reparsed
        ? undefined
        : this.wrapNode(value);
    }
    if (isCompilerNode(value)) return this.wrapNode(value);
    return value;
  }

  private get(
    ref: SyntaxRef,
    key: string | symbol,
    proxy: object,
    args?: readonly unknown[],
  ): unknown {
    const position = positionMethod(key);
    if (typeof position === 'function') {
      if (!args) return syntaxMethod(key);
      if (!this.identities.has(proxy)) {
        return Reflect.apply(position, proxy, [...args]);
      }
      if (
        key === 'getFullStart' ||
        key === 'getEnd' ||
        key === 'getFullWidth'
      ) {
        const value: unknown = Reflect.apply(position, proxy, [...args]);
        const pos = getProp(proxy, 'pos');
        if (
          typeof value === 'number' &&
          (value < 0 ||
            (key === 'getFullWidth' &&
              value === 0 &&
              typeof pos === 'number' &&
              pos < 0))
        ) {
          this.positionedNode(ref);
        }
        return value;
      }
      const callArgs = [...args];
      const firstArg = callArgs[0];
      const source = isObject(firstArg)
        ? this.identities.get(firstArg)
        : undefined;
      if (source) callArgs[0] = this.rawSyntax(source);
      return Reflect.apply(position, this.positionedNode(ref), callArgs);
    }

    const real = isSyntheticRef(ref) ? undefined : ref;
    const owner = isSyntheticRef(ref) ? ref.owner : ref;
    const { file } = owner;

    if (key === 'symbol') {
      return real &&
        !this.descriptor.syntaxOnly &&
        scalarFields[publicSyntaxKind(this.rawNode(real))]?.includes(key)
        ? this.readDeclarationSymbol(real)
        : undefined;
    }
    if (key === 'getSourceFile') {
      if (!args) return syntaxMethod(key);
      return this.diagnosticSourceView(file);
    }
    if (real?.index === 1) {
      const source = this.lookupMetadata(file);
      if (!source) throw new Error('foreign source');
      if (key === 'fileName' || key === 'path' || key === 'isDeclarationFile') {
        return source[key];
      }
      if (key === 'impliedNodeFormat' || key === 'checkJsDirective') {
        return source[key];
      }
      if (
        key === 'referencedFiles' ||
        key === 'typeReferenceDirectives' ||
        key === 'libReferenceDirectives'
      ) {
        const sf = this.source(file);
        if (!sf) throw new Error('foreign source');
        const references = sf[key].map(
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
      if (key === 'hasNoDefaultLib') return false;
      if (key === 'languageVersion') {
        const targets = this.typescript.ScriptTarget;
        if (file.toLowerCase().endsWith('.json')) return targets.ES2015;
        const value = this.descriptor.compilerOptions.target;
        const target = typeof value === 'number' ? value : undefined;
        if (this.descriptor.syntaxOnly) return target ?? targets.Latest;
        return target === undefined || target === Number(targets.ES3)
          ? targets.LatestStandard
          : target;
      }
      if (key === 'update') {
        return args ? unsupported('SourceFile.update') : syntaxMethod(key);
      }
      if (key === 'kind') return this.typescript.SyntaxKind.SourceFile;
      if (key === 'parent') return undefined;
    }

    if (
      real &&
      !this.cachedSource(file) &&
      !(real.index === 1 && key === 'text') &&
      (key === 'kind' ||
        key === 'flags' ||
        key === 'parent' ||
        key === 'name' ||
        key === 'text' ||
        key === 'escapedText')
    ) {
      const fact = this.summary(real);
      if (key === 'kind') {
        if (fact.kind === 'TypeReference' && this.isHeritage(fact)) {
          return this.typescript.SyntaxKind.ExpressionWithTypeArguments;
        }
        if (fact.kind === 'QualifiedName' && this.isHeritage(fact)) {
          return this.typescript.SyntaxKind.PropertyAccessExpression;
        }
        return this.translateKind(fact.kind);
      }
      if (key === 'flags') {
        return translateNodeFlags(
          fact.flags,
          this.typescript.NodeFlags,
          fact.moduleKeyword,
        );
      }
      if (key === 'parent') return this.nodeRef(fact.parent);
      if (
        key === 'name' &&
        !(fact.kind === 'QualifiedName' && this.isHeritage(fact))
      ) {
        return this.nodeRef(fact.name);
      }
      if (key === 'text' && fact.text !== undefined) return fact.text;
      if (
        key === 'escapedText' &&
        (fact.kind === 'Identifier' || fact.kind === 'PrivateIdentifier')
      ) {
        return fact.text?.startsWith('__') ? `_${fact.text}` : fact.text;
      }
    }

    if (key === 'forEachChild') {
      if (!args) return syntaxMethod(key);
      return forEachSyntaxChild(
        this.rawSyntax(ref),
        (item) => this.wrap(item),
        args,
      );
    }

    if (tokenNavigation.has(key)) {
      if (!args) return syntaxMethod(key);
      const sourceArg = args[key === 'getChildAt' ? 1 : 0];
      if (isObject(sourceArg)) this.checkSource(file, sourceArg);
      if (key === 'getChildren') return this.childNodes(ref, proxy);
      if (key === 'getChildAt') {
        const index = args[0];
        if (typeof index !== 'number') return undefined;
        const children = this.children.get(proxy);
        if (children) return children[index];
        const child = this.positionedNode(ref).getChildAt(index);
        return child && this.wrapNode(child);
      }
      if (key === 'getChildCount') {
        return (
          this.children.get(proxy)?.length ??
          this.positionedNode(ref).getChildCount()
        );
      }
      const node = this.positionedNode(ref);
      const token =
        key === 'getFirstToken' ? node.getFirstToken() : node.getLastToken();
      return token && this.wrapNode(token);
    }

    if (isSyntheticRef(ref)) {
      if (key === 'parent') return this.nodeRef(ref.owner);
      if (key === 'flags') {
        const ownerNode = this.nodeRef(ref.owner);
        const flags = ownerNode ? getProp(ownerNode, 'flags') : 0;
        return typeof flags === 'number'
          ? flags & this.typescript.NodeFlags.ContextFlags
          : 0;
      }
    }

    const raw =
      key === 'getStart' && args
        ? this.positionedNode(ref)
        : this.rawSyntax(ref);

    if (
      key === 'flags' &&
      real &&
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

    if (key === '_children' && raw.kind === SyntaxKind.SyntaxList) {
      return this.childNodes(ref, proxy);
    }

    const member = this.syntaxMember(raw, key);
    if (typeof member === 'function') {
      if (!args) return syntaxMethod(key);
      return this.wrap(
        Reflect.apply(
          member,
          raw,
          args.map((arg) => {
            const identity = isObject(arg)
              ? this.identities.get(arg)
              : undefined;
            return identity ? this.rawSyntax(identity) : arg;
          }),
        ),
      );
    }

    const value = this.wrap(member);
    if (Array.isArray(value)) {
      Object.defineProperty(proxy, key, {
        value,
        writable: true,
        configurable: true,
        enumerable: true,
      });
    }
    return value;
  }

  private syntaxMember(raw: Readonly<Node>, key: string | symbol): unknown {
    if (key === 'parent') {
      let parent = raw.parent;
      while (parent && isJSDocCast(parent)) parent = parent.parent;
      return parent;
    }
    if (key === 'postfixToken') return undefined;
    const kind = publicSyntaxKind(raw);
    if (key === 'jsDocPropertyTags' && kind === SyntaxKind.JSDocTypeLiteral) {
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
      let name = getProp(raw, 'name');
      while (
        isCompilerNode(name) &&
        name.kind === SyntaxKind.ModuleDeclaration
      ) {
        name = getProp(name, 'body') ?? getProp(name, 'name');
      }
      return name;
    }
    if (
      key === 'comment' &&
      kind >= SyntaxKind.FirstJSDocNode &&
      kind <= SyntaxKind.LastJSDocNode
    ) {
      return jsDocComment(raw);
    }
    if (
      key === 'hasExtendedUnicodeEscape' &&
      kind === SyntaxKind.StringLiteral
    ) {
      const tokenFlags = getProp(raw, 'tokenFlags');
      return raw.pos < 0
        ? undefined
        : typeof tokenFlags === 'number' &&
            !!(tokenFlags & TokenFlags.ExtendedUnicodeEscape);
    }
    const alias =
      typeof key === 'string' ? fieldAliases[kind]?.[key] : undefined;
    if (alias === false) return undefined;
    if (typeof alias === 'string') return getProp(raw, alias);
    if (kind !== raw.kind) {
      if (key === 'expression') {
        return getProp(
          raw,
          kind === SyntaxKind.PropertyAccessExpression ? 'left' : 'typeName',
        );
      }
      if (key === 'name' && kind === SyntaxKind.PropertyAccessExpression) {
        return getProp(raw, 'right');
      }
      if (key === 'typeName' || key === 'left' || key === 'right') {
        return undefined;
      }
    }
    if (key === 'kind') return this.translateKind(SyntaxKind[kind]);
    if (
      (key === 'questionToken' || key === 'exclamationToken') &&
      childProperties[raw.kind]?.includes('postfixToken')
    ) {
      const token = getProp(raw, 'postfixToken');
      return isCompilerNode(token) &&
        token.kind ===
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
      const tokenKind = getProp(raw, key);
      return typeof tokenKind === 'number'
        ? this.translateKind(SyntaxKind[tokenKind])
        : undefined;
    }
    if (key === 'flags') {
      const kw =
        raw.kind === SyntaxKind.ModuleDeclaration
          ? getProp(raw, 'keyword')
          : undefined;
      return translateNodeFlags(
        raw.flags,
        this.typescript.NodeFlags,
        typeof kw === 'number' ? kw : 0,
      );
    }
    if (
      key === 'escapedText' &&
      (raw.kind === SyntaxKind.Identifier ||
        raw.kind === SyntaxKind.PrivateIdentifier)
    ) {
      const text = getProp(raw, 'text');
      if (typeof text !== 'string') return undefined;
      return text.startsWith('__') ? `_${text}` : text;
    }
    return getProp(raw, key);
  }

  /**
   * Decodes a detached syntax tree returned by `checker.typeToTypeNode`.
   * Generated nodes keep their own identity while exposing `original` when
   * tsgo's node builder preserved a parse-tree origin.
   */
  generatedTypeNode({ tree, origins, emitFlags }: GeneratedSyntax): object {
    const root = readSubtree(tree);
    const views = new WeakMap<object, object>();
    const getSourceFile = function (this: Node): unknown {
      const view = views.get(this);
      const parent = view ? getProp(view, 'parent') : undefined;
      if (!isObject(parent)) return undefined;
      const fn = getProp(parent, 'getSourceFile');
      return typeof fn === 'function'
        ? Reflect.apply(fn, parent, [])
        : undefined;
    };
    const scanSource = (arg: unknown): unknown => {
      if (!isObject(arg)) return arg;
      let source = views.get(arg);
      if (!source) {
        const created: object = {};
        Object.setPrototypeOf(created, arg);
        Object.defineProperties(created, {
          tokenCache: { writable: true },
          childrenCache: { writable: true },
        });
        source = created;
        views.set(arg, source);
      }
      return source;
    };
    const wrap = (value: unknown): unknown => {
      if (!isObject(value)) return value;
      const old = views.get(value);
      if (old) return old;
      if (Array.isArray(value)) {
        const list = wrapList(value, wrap);
        if ('pos' in list && 'end' in list) shareGeneratedRange(value, list);
        views.set(value, list);
        return list;
      }
      if (!isCompilerNode(value)) return value;
      const node = value;
      const encoded = node instanceof RemoteNode;
      const index = encoded ? remoteIndex(node) : 0;
      const flags = emitFlags?.[index];
      const fields = {
        pos: node.pos,
        end: node.end,
        emitNode: flags
          ? { flags: this.typescript.builderEmitFlags[flags] }
          : undefined,
      };
      shareGeneratedRange(node, fields);
      Object.defineProperty(node, 'getSourceFile', { value: getSourceFile });
      const original = origins?.[index];
      const proxy = syntaxView(
        () => node,
        (key, view, args) => {
          if (key === 'original') return this.nodeRef(original);
          if (
            key === 'hasExtendedUnicodeEscape' &&
            encoded &&
            !original &&
            node.kind === SyntaxKind.StringLiteral
          ) {
            return undefined;
          }
          if (key === 'parent') return encoded ? undefined : wrap(node.parent);
          if (key === 'flags') {
            const nodeFlags = this.syntaxMember(node, key);
            if (typeof nodeFlags !== 'number') return nodeFlags;
            return encoded
              ? nodeFlags | this.typescript.NodeFlags.Synthesized
              : nodeFlags;
          }
          if (key === 'getSourceFile') {
            if (!args) return syntaxMethod(key);
            const parent = getProp(view, 'parent');
            if (!isObject(parent)) return undefined;
            const fn = getProp(parent, 'getSourceFile');
            return typeof fn === 'function'
              ? Reflect.apply(fn, parent, [])
              : undefined;
          }
          if (key === 'forEachChild') {
            return args
              ? forEachSyntaxChild(node, wrap, args)
              : syntaxMethod(key);
          }
          const member = this.syntaxMember(node, key);
          if (typeof member === 'function') {
            if (!args) return syntaxMethod(key);
            if (needsRealPosition(node)) {
              assertHasRealPosition.call(node);
            } else {
              const position = positionMethod(key);
              if (typeof position === 'function') {
                return Reflect.apply(position, view, [...args]);
              }
            }
            const callArgs = [...args];
            if (tokenNavigation.has(key)) {
              const idx = key === 'getChildAt' ? 1 : 0;
              callArgs[idx] = scanSource(callArgs[idx] ?? node.getSourceFile());
            }
            return wrap(Reflect.apply(member, node, callArgs));
          }
          return wrap(member);
        },
        this.ensure,
        this.typescript.jsDocKinds,
        fields,
      );
      this.generated.add(proxy);
      views.set(value, proxy);
      return proxy;
    };
    const wrapped = wrap(root);
    if (!isObject(wrapped)) throw new Error('invalid generated syntax node');
    return wrapped;
  }

  private childNodes(ref: SyntaxRef, proxy: object): object[] {
    let children = this.children.get(proxy);
    if (!children) {
      children = this.positionedNode(ref)
        .getChildren()
        .map((child) => this.wrapNode(child));
      this.children.set(proxy, children);
    }
    return children;
  }

  private checkSource(file: string, sourceFile: object): void {
    const ref = this.locationOrUndefined(sourceFile);
    if (ref?.file !== file || ref.index !== 1) {
      throw new Error('foreign syntax source');
    }
  }

  close(): void {
    this.lookup = undefined;
    this.children = new WeakMap();
    this.activeSources.clear();
    this.oversizedSource = undefined;
    this.sources.clear();
    this.diagnosticMetadata.clear();
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
    number | { readonly firstTokenOf: number }
  >();
  private reverse?: Readonly<Record<string, number>>;
  private readonly aliases = new Map<number, SequenceAlias | ChainAlias>();
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
    this.reverse = undefined;
    this.aliases.clear();
    this.materializedAliases.clear();
    this.sequenceMembers.clear();
    if (this.attached) this.registry.releaseFile(this.binding.sourceFileName);
    this.attached = false;
    this.nodes = undefined;
    this.lineStarts = [];
  }

  attach(
    frame: Readonly<DecodedFrame>,
    projection: TypeServiceProjection,
    sourceFileName: string,
  ): void {
    this.ensure();
    if (this.attached) throw new Error('file type services already attached');
    if (
      projection.binding.generation !== this.binding.generation ||
      projection.origins.length !== frame.nodes.length
    ) {
      throw new Error('stale or incompatible projection origins');
    }
    if (
      projection.sourceFileName !== this.binding.sourceFileName ||
      projection.binding.sourceFileName !== this.binding.sourceFileName ||
      (sourceFileName !== this.binding.sourceFileName &&
        this.registry.sourceView(sourceFileName) !==
          this.registry.sourceView(this.binding.sourceFileName))
    ) {
      throw new Error('type service projection belongs to a different source');
    }
    this.registry.attach(projection.tree, {
      file: this.binding.sourceFileName,
      text: frame.text,
    });
    this.attached = true;
    this.nodes = frame.nodes;
    this.lineStarts = frame.lineStarts;
    for (let id = 1; id < frame.nodes.length; id++) {
      const index = projection.origins[id];
      const node = frame.nodes[id];
      if (index && node) this.forward.set(node, index);
    }
    for (const [id, firstTokenOf] of Object.entries(
      projection.firstTokens ?? {},
    )) {
      const node = frame.nodes[Number(id)];
      if (node) this.forward.set(node, { firstTokenOf });
    }
    this.reverse = { ...projection.reverse };
    for (const alias of projection.aliases ?? []) {
      this.aliases.set(alias.index, alias);
      if (this.sequenceMembers.has(alias.sequence)) continue;
      const memberIds = projection.sequenceMembers?.[alias.sequence];
      if (memberIds) {
        const resolved: ESTreeNode[] = [];
        for (const id of memberIds) {
          const item = frame.nodes[id];
          if (item) resolved.push(item);
        }
        this.sequenceMembers.set(alias.sequence, resolved);
      } else {
        const seqNode = frame.nodes[alias.sequence];
        const expressions = seqNode
          ? getProp(seqNode, 'expressions')
          : undefined;
        if (Array.isArray(expressions)) {
          this.sequenceMembers.set(
            alias.sequence,
            (expressions as readonly ESTreeNode[]).slice(),
          );
        }
      }
    }
    for (const alias of projection.chainAliases ?? []) {
      this.aliases.set(alias.index, alias);
    }
  }

  readonly esTreeNodeToTSNodeMap = {
    get: (node: Readonly<ESTreeNode>): object | undefined => {
      this.ensure();
      const origin = this.forward.get(node);
      if (origin === undefined) return undefined;
      const file = this.binding.sourceFileName;
      return typeof origin === 'number'
        ? this.registry.nodeRef({ file, index: origin })
        : this.registry.firstToken({ file, index: origin.firstTokenOf });
    },
    has: (node: Readonly<ESTreeNode>): boolean => {
      this.ensure();
      return this.forward.has(node);
    },
  };

  readonly tsNodeToESTreeNodeMap = {
    get: (node: object): ESTreeNode | undefined => {
      this.ensure();
      const ref = this.registry.locationOrUndefined(node);
      if (ref?.file !== this.binding.sourceFileName) return undefined;
      const index = ref.index;
      const id = this.reverse?.[index];
      if (id) return this.nodes?.[id];
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
            expression: this.nodes?.[info.expression],
          }
        : {
            type: 'SequenceExpression',
            ...position,
            expressions: (this.sequenceMembers.get(info.sequence) ?? []).slice(
              0,
              info.count,
            ),
          };
      const origin = chain ? info.origin : index;
      this.materializedAliases.set(key, alias);
      if (origin) this.forward.set(alias, origin);
      return alias;
    },
    has: (node: object): boolean => {
      this.ensure();
      const ref = this.registry.locationOrUndefined(node);
      if (ref?.file !== this.binding.sourceFileName) return false;
      const index = ref.index;
      return this.reverse?.[index] !== undefined || this.aliases.has(index);
    },
  };
}
