import type ts from 'typescript';
import { SyntaxKind as nativeSyntaxKind } from '@rslint/api';
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
  isCacheableTypeQuery,
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
  type IndexInfoFact,
  type DiagnosticFact,
  type NamedSymbolFact,
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

export interface ConditionalRootView {
  readonly instantiations?: ReadonlyMap<string, TypeView>;
  node: object;
  checkType: TypeView;
  extendsType: TypeView;
  isDistributive: boolean;
  inferTypeParameters?: readonly TypeView[];
  outerTypeParameters?: readonly TypeView[];
  aliasSymbol?: SymbolView;
  aliasTypeArguments?: readonly TypeView[];
}
interface DeclaredMemberViews {
  declaredProperties: readonly SymbolView[];
  declaredCallSignatures: readonly SignatureView[];
  declaredConstructSignatures: readonly SignatureView[];
  declaredIndexInfos: readonly IndexInfoView[];
}
export interface TypeView extends Partial<Readonly<DeclaredMemberViews>> {
  readonly checker: CheckerView;
  readonly simplifiedForReading: TypeView | undefined;
  readonly simplifiedForWriting: TypeView | undefined;
  readonly elementType: TypeView | undefined;
  readonly finalArrayType: TypeView | undefined;
  readonly root: ConditionalRootView | undefined;
  readonly resolvedTrueType: TypeView | undefined;
  readonly resolvedFalseType: TypeView | undefined;
  readonly pattern: object | undefined;
  readonly node: object | undefined;
  readonly freshType: TypeView | undefined;
  readonly regularType: TypeView | undefined;
  readonly baseType: TypeView | undefined;
  readonly constraint: TypeView | undefined;
  readonly outerTypeParameters: readonly TypeView[] | undefined;
  readonly localTypeParameters: readonly TypeView[] | undefined;
  readonly elementFlags: readonly number[] | undefined;
  readonly minLength: number | undefined;
  readonly fixedLength: number | undefined;
  readonly hasRestElement: boolean | undefined;
  readonly labeledElementDeclarations:
    readonly (object | undefined)[] | undefined;
  readonly texts: readonly string[] | undefined;
  readonly id: number;
  readonly flags: number;
  readonly symbol: SymbolView | undefined;
  readonly aliasSymbol: SymbolView | undefined;
  readonly intrinsicName: string | undefined;
  readonly escapedName: string | undefined;
  readonly objectFlags: number | undefined;
  readonly value: TypeFact['value'];
  readonly types: readonly TypeView[] | undefined;
  readonly type: TypeView | undefined;
  readonly objectType: TypeView | undefined;
  readonly indexType: TypeView | undefined;
  readonly checkType: TypeView | undefined;
  readonly extendsType: TypeView | undefined;
  readonly target: TypeView | undefined;
  readonly typeArguments: readonly TypeView[] | undefined;
  readonly aliasTypeArguments: readonly TypeView[] | undefined;
  readonly typeParameters: readonly TypeView[] | undefined;
  readonly typeParameter: TypeView | undefined;
  readonly constraintType: TypeView | undefined;
  readonly templateType: TypeView | undefined;
  readonly modifiersType: TypeView | undefined;
  readonly thisType: TypeView | undefined;
  readonly combinedFlags: number | undefined;
  readonly readonly: boolean | undefined;
  getFlags(): number;
  getSymbol(): SymbolView | undefined;
  isUnion(): boolean;
  isIntersection(): boolean;
  isUnionOrIntersection(): boolean;
  isLiteral(): boolean;
  isStringLiteral(): boolean;
  isNumberLiteral(): boolean;
  isTypeParameter(): boolean;
  isClassOrInterface(): boolean;
  isClass(): boolean;
  isIndexType(): boolean;
  getCallSignatures(): readonly SignatureView[];
  getConstructSignatures(): readonly SignatureView[];
  getProperties(): readonly SymbolView[];
  getApparentProperties(): readonly SymbolView[];
  getProperty(name: string): SymbolView | undefined;
  getNonNullableType(): TypeView | undefined;
  getConstraint(): TypeView | undefined;
  getDefault(): TypeView | undefined;
  getBaseTypes(): readonly TypeView[] | undefined;
  getStringIndexType(): TypeView | undefined;
  getNumberIndexType(): TypeView | undefined;
}
export interface SymbolView {
  readonly id: number;
  readonly name: string;
  readonly escapedName: string;
  readonly flags: number;
  readonly declarations: readonly object[] | undefined;
  readonly valueDeclaration: object | undefined;
  readonly exports: ReadonlyMap<string, SymbolView> | undefined;
  readonly globalExports: ReadonlyMap<string, SymbolView> | undefined;
  readonly members: ReadonlyMap<string, SymbolView> | undefined;
  readonly links:
    | { readonly checkFlags: number; readonly type: TypeView | undefined }
    | undefined;
  getName(): string;
  getEscapedName(): string;
  getFlags(): number;
  getDeclarations(): readonly object[] | undefined;
  getDocumentationComment(checker?: CheckerView): ts.SymbolDisplayPart[];
  getJsDocTags(checker?: CheckerView): ts.JSDocTagInfo[];
}
export interface SignatureView {
  readonly id: number;
  readonly parameters: readonly SymbolView[];
  readonly declaration: object | undefined;
  readonly typeParameters: readonly TypeView[] | undefined;
  readonly thisParameter: SymbolView | undefined;
  readonly minArgumentCount: number;
  getParameters(): readonly SymbolView[];
  getTypeParameters(): readonly TypeView[] | undefined;
  getTypeParameterAtPosition(position: number): TypeView | undefined;
  getDeclaration(): object | undefined;
  getReturnType(): TypeView | undefined;
  getDocumentationComment(checker?: CheckerView): ts.SymbolDisplayPart[];
  getJsDocTags(): ts.JSDocTagInfo[];
}
export interface IndexInfoView {
  readonly components?: readonly object[];
  readonly keyType: TypeView;
  readonly type: TypeView;
  readonly isReadonly: boolean;
  readonly declaration: object | undefined;
}
export interface TypePredicateView {
  readonly kind: number;
  readonly parameterName: string | undefined;
  readonly parameterIndex: number | undefined;
  readonly type: TypeView | undefined;
}
export interface CheckerView {
  getStringLiteralType(value: string): TypeView;
  getNumberLiteralType(value: number): TypeView;
  getBigIntLiteralType(value: ts.PseudoBigInt): TypeView;
  getNullableType(t: object, flags: number): TypeView;
  getPrivateIdentifierPropertyOfType(
    t: object,
    name: string,
    location: object,
  ): SymbolView | undefined;
  tryGetMemberInModuleExports(
    name: string,
    moduleSymbol: object,
  ): SymbolView | undefined;
  isValidPropertyAccess(n: object, name: string): boolean;
  getSymbolsOfParameterPropertyDeclaration(
    n: object,
    name: string,
  ): readonly SymbolView[];
  getSymbolOfExpando(
    n: object,
    allowDeclaration: boolean,
  ): SymbolView | undefined;
  getIndexInfosOfIndexSymbol(
    s: object,
    siblings?: readonly object[],
  ): readonly IndexInfoView[];
  signatureToSignatureDeclaration(
    s: object,
    kind: number,
    enclosing?: object,
    flags?: number,
  ): object | undefined;
  symbolToParameterDeclaration(
    s: object,
    enclosing?: object,
    flags?: number,
  ): object | undefined;
  symbolToTypeParameterDeclarations(
    s: object,
    enclosing?: object,
    flags?: number,
  ): readonly object[] | undefined;
  typeParameterToDeclaration(
    t: object,
    enclosing?: object,
    flags?: number,
  ): object | undefined;
  symbolToEntityName(
    s: object,
    meaning: number,
    enclosing?: object,
    flags?: number,
  ): object | undefined;
  symbolToExpression(
    s: object,
    meaning: number,
    enclosing?: object,
    flags?: number,
  ): object | undefined;
  symbolToString(
    s: object,
    enclosing?: object,
    meaning?: number,
    flags?: number,
  ): string;
  typePredicateToString(
    predicate: TypePredicateView,
    enclosing?: object,
    flags?: number,
  ): string;
  indexInfoToIndexSignatureDeclaration(
    info: IndexInfoView,
    enclosing?: object,
    flags?: number,
  ): object | undefined;
  getAnyType(): TypeView;
  getUnknownType(): TypeView;
  getStringType(): TypeView;
  getNumberType(): TypeView;
  getBooleanType(): TypeView;
  getBigIntType(): TypeView;
  getESSymbolType(): TypeView;
  getNonPrimitiveType(): TypeView;
  getNeverType(): TypeView;
  getNullType(): TypeView;
  getUndefinedType(): TypeView;
  getTrueType(): TypeView;
  getFalseType(): TypeView;
  getAmbientModules(): readonly SymbolView[];
  getAugmentedPropertiesOfType(t: object): readonly SymbolView[];
  getRootSymbols(s: object): readonly SymbolView[];
  getMergedSymbol(s: object): SymbolView | undefined;
  isArgumentsSymbol(s: object): boolean;
  isUndefinedSymbol(s: object): boolean;
  getExportSpecifierLocalTargetSymbol(n: object): SymbolView | undefined;
  getPropertySymbolOfDestructuringAssignment(n: object): SymbolView | undefined;
  getJsxIntrinsicTagNamesAt(n: object): readonly SymbolView[];
  getTypeOfAssignmentPattern(n: object): TypeView | undefined;
  isOptionalParameter(n: object): boolean;
  getConstantValue(n: object): string | number | undefined;
  isImplementationOfOverload(n: object): boolean | undefined;
  getTypeArgumentsForResolvedSignature(
    s: object,
  ): readonly TypeView[] | undefined;
  getTypeAtLocation(n: object): TypeView | undefined;
  getTypeFromTypeNode(n: object): TypeView | undefined;
  getContextualType(n: object): TypeView | undefined;
  getContextualTypeForArgumentAtIndex(
    n: object,
    index: number,
  ): TypeView | undefined;
  getFullyQualifiedName(symbol: object): string;
  typeToTypeNode(
    type: object,
    enclosing?: object,
    flags?: number,
  ): object | undefined;
  getSuggestionDiagnostics(
    source: object,
    cancellation?: ts.CancellationToken,
  ): DiagnosticView[];
  runWithCancellationToken<T>(
    token: ts.CancellationToken,
    cb: (checker: CheckerView) => T,
  ): T;
  getVoidType(): TypeView | undefined;
  getBaseConstraintOfType(t: object): TypeView | undefined;
  getApparentType(t: object): TypeView | undefined;
  getBaseTypeOfLiteralType(t: object): TypeView | undefined;
  getNonNullableType(t: object): TypeView | undefined;
  getAwaitedType(t: object): TypeView | undefined;
  getWidenedType(t: object): TypeView | undefined;
  getBaseTypes(t: object): readonly TypeView[];
  getDefaultFromTypeParameter(t: object): TypeView | undefined;
  isTypeAssignableTo(source: object, target: object): boolean;
  typeToString(t: object, enclosing?: object, flags?: number): string;
  signatureToString(
    s: object,
    enclosing?: object,
    flags?: number,
    kind?: number,
  ): string;
  getSymbolAtLocation(n: object): SymbolView | undefined;
  getSymbolsInScope(n: object, meaning: number): readonly SymbolView[];
  resolveName(
    name: string,
    location: object | undefined,
    meaning: number,
    excludeGlobals: boolean,
  ): SymbolView | undefined;
  getAliasedSymbol(s: object): SymbolView | undefined;
  getImmediateAliasedSymbol(s: object): SymbolView | undefined;
  isUnknownSymbol(s: object): boolean;
  getExportsOfModule(s: object): readonly SymbolView[];
  getExportSymbolOfSymbol(s: object): SymbolView | undefined;
  getShorthandAssignmentValueSymbol(
    n: object | undefined,
  ): SymbolView | undefined;
  getTypeOfSymbol(s: object): TypeView | undefined;
  getTypeOfSymbolAtLocation(s: object, n: object): TypeView | undefined;
  getDeclaredTypeOfSymbol(s: object): TypeView | undefined;
  getPropertiesOfType(t: object): readonly SymbolView[];
  getPropertyOfType(t: object, name: string): SymbolView | undefined;
  getTypeOfPropertyOfType(t: object, name: string): TypeView | undefined;
  getIndexTypeOfType(t: object, kind: number): TypeView | undefined;
  getIndexInfoOfType(t: object, kind: number): IndexInfoView | undefined;
  getIndexInfosOfType(t: object): readonly IndexInfoView[];
  getTypeArguments(t: object): readonly TypeView[];
  isArrayType(t: object): boolean;
  isArrayLikeType(t: object): boolean;
  isTupleType(t: object): boolean;
  getSignatureFromDeclaration(n: object): SignatureView | undefined;
  getResolvedSignature(
    n: object,
    candidates?: SignatureView[],
    argumentCount?: number,
  ): SignatureView | undefined;
  getReturnTypeOfSignature(s: object): TypeView | undefined;
  getTypePredicateOfSignature(s: object): TypePredicateView | undefined;
  getSignaturesOfType(t: object, kind: number): readonly SignatureView[];
}
export interface ProgramView {
  getConfigFileParsingDiagnostics(): readonly DiagnosticView[];
  getOptionsDiagnostics(
    cancellation?: ts.CancellationToken,
  ): readonly DiagnosticView[];
  getGlobalDiagnostics(
    cancellation?: ts.CancellationToken,
  ): readonly DiagnosticView[];
  getSyntacticDiagnostics(
    source?: object,
    cancellation?: ts.CancellationToken,
  ): readonly DiagnosticView[];
  getSemanticDiagnostics(
    source?: object,
    cancellation?: ts.CancellationToken,
  ): readonly DiagnosticView[];
  getDeclarationDiagnostics(
    source?: object,
    cancellation?: ts.CancellationToken,
  ): readonly DiagnosticView[];
  getNodeCount(): number;
  getIdentifierCount(): number;
  getSymbolCount(): number;
  getTypeCount(): number;
  getInstantiationCount(): number;
  getRelationCacheSizes(): {
    assignable: number;
    identity: number;
    subtype: number;
    strictSubtype: number;
  };
  getRootFileNames(): readonly string[];
  getModeForUsageLocation(source: object, usage: object): number | undefined;
  getModeForResolutionAtIndex(
    source: object,
    index: number,
  ): number | undefined;
  getProjectReferences(): readonly ts.ProjectReference[] | undefined;
  getResolvedProjectReferences():
    readonly (ResolvedReferenceView | undefined)[] | undefined;
  emit(): never;
  getTypeChecker(): CheckerView;
  getCompilerOptions(): Readonly<Record<string, unknown>>;
  getCurrentDirectory(): string;
  getSourceFileByPath(path: string): object | undefined;
  getSourceFile(name: string): object | undefined;
  getSourceFiles(): readonly object[];
  sourceFileToPackageName: { get(path: string): string | undefined };
  isSourceFileFromExternalLibrary(source: object): boolean;
  isSourceFileDefaultLibrary(source: object): boolean;
}
interface ResolvedReferenceView {
  commandLine: Omit<ts.ParsedCommandLine, 'errors' | 'options'> & {
    options: Readonly<Record<string, unknown>>;
    errors: DiagnosticView[];
  };
  sourceFile: object;
  references?: readonly (ResolvedReferenceView | undefined)[];
}
type ReferenceConfig = {
  name: string;
  path: string;
  options: Record<string, unknown>;
  fileNames: string[];
  projectReferences?: ts.ProjectReference[];
  wildcardDirectories: ts.MapLike<ts.WatchDirectoryFlags>;
  extendedSourceFiles?: string[];
  references?: number[];
  raw?: unknown;
  compileOnSave?: boolean;
  typeAcquisition?: ts.TypeAcquisition;
  watchOptions?: ts.WatchOptions;
  diagnostics?: string;
};
type DiagnosticView = Omit<ts.Diagnostic, 'file' | 'relatedInformation'> & {
  file: object | undefined;
  relatedInformation?: DiagnosticView[];
};

function diagnosticMessage(fact: DiagnosticFact): ts.DiagnosticMessageChain {
  return {
    code: fact.code,
    category: fact.category,
    messageText: fact.text,
    next: fact.chain.length ? fact.chain.map(diagnosticMessage) : undefined,
  };
}
function unsupported(surface: string): never {
  throw new Error(`unsupported TypeScript operation: ${surface}`);
}
// Field order in the mappedFields reply.
const mappedFieldIndex = {
  typeParameter: 0,
  constraintType: 1,
  templateType: 2,
  modifiersType: 3,
} as const;
const declaredMemberKeys = new Set<PropertyKey>([
  'declaredProperties',
  'declaredCallSignatures',
  'declaredConstructSignatures',
  'declaredIndexInfos',
]);

const serviceView = Symbol();
function canonicalView<T extends object>(view: T & { [serviceView]?: T }): T {
  const original = view[serviceView];
  // A transparent proxy forwards its target's prototype. A derived object
  // inherits the marker but has its own identity and cannot name a Go type.
  return original &&
    (original === view ||
      Object.getPrototypeOf(view) === Object.getPrototypeOf(original))
    ? original
    : view;
}

const inspectCustom = Symbol.for('nodejs.util.inspect.custom');
const inspectionReceivers = new WeakSet<object>();
const rejectTypeEdit = () => false;

const typeInspection = {
  configurable: true,
  writable: true,
  value(this: object, _depth: number, options: { getters?: boolean | string }) {
    if (!options?.getters || inspectionReceivers.has(this)) return this;
    // With getters enabled, Node calls each accessor on the proxy's target.
    // Give it the rule's actual receiver. Returning this on the second pass
    // keeps cycle detection and formatting inside the original inspection.
    const receiver = new Proxy(this, {});
    inspectionReceivers.add(receiver);
    return receiver;
  },
} satisfies PropertyDescriptor;

export function guarded<T extends object>(object: T, ensure: () => void): T {
  const proxy: T = new Proxy(object, {
    get(target, key, receiver) {
      ensure();
      if (key === serviceView) return proxy;
      const value: unknown = Reflect.get(target, key, receiver);
      return typeof value === 'function'
        ? (...args: unknown[]) => {
            ensure();
            return Reflect.apply(value, receiver, args);
          }
        : value;
    },
  });
  return proxy;
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
// Compiler records can carry plugin metadata, but their native data fields
// describe the same object returned by other checker queries.
function readonlyDataFields<T extends object>(record: T): T {
  for (const key of Object.keys(record))
    Object.defineProperty(record, key, {
      configurable: false,
      writable: false,
    });
  return record;
}
type ObjectKind = 'type' | 'symbol' | 'signature';
interface ObjectBrand {
  readonly kind: ObjectKind;
  readonly id: number;
  knownName?: string;
}
type TypeScriptEnums = Pick<
  typeof ts,
  | 'SyntaxKind'
  | 'TypeFlags'
  | 'ObjectFlags'
  | 'NodeFlags'
  | 'JsxEmit'
  | 'ModuleDetectionKind'
  | 'ScriptTarget'
> & { SymbolFlags: Pick<typeof ts.SymbolFlags, 'TypeLiteral'> };

type TypePrototype = ReturnType<typeof makeTypePrototype>;

function makeTypePrototype(
  checker: CheckerView,
  typescript: TypeScriptEnums,
  ensure: () => void,
) {
  const { TypeFlags: tf, ObjectFlags: of } = typescript;
  function cached<T>(read: (type: TypeView) => T) {
    const values = new WeakMap<TypeView, T>();
    return function (this: TypeView) {
      ensure();
      const type = canonicalView(this);
      if (!values.has(type)) values.set(type, read(type));
      return values.get(type)!;
    };
  }
  const typeArguments = cached((t) =>
    (t.objectFlags ?? 0) & of.Reference
      ? Object.freeze(checker.getTypeArguments(t))
      : undefined,
  );
  const prototype = {
    get typeArguments() {
      return typeArguments.call(this);
    },
    getFlags() {
      return this.flags;
    },
    getSymbol() {
      return this.symbol;
    },
    isUnion() {
      return !!(this.flags & tf.Union);
    },
    isIntersection() {
      return !!(this.flags & tf.Intersection);
    },
    isUnionOrIntersection() {
      return !!(this.flags & tf.UnionOrIntersection);
    },
    isLiteral() {
      return !!(
        this.flags &
        (tf.StringLiteral | tf.NumberLiteral | tf.BigIntLiteral)
      );
    },
    isStringLiteral() {
      return !!(this.flags & tf.StringLiteral);
    },
    isNumberLiteral() {
      return !!(this.flags & tf.NumberLiteral);
    },
    isTypeParameter() {
      return !!(this.flags & tf.TypeParameter);
    },
    isClassOrInterface() {
      return !!((this.objectFlags ?? 0) & of.ClassOrInterface);
    },
    isClass() {
      return !!((this.objectFlags ?? 0) & of.Class);
    },
    isIndexType() {
      return !!(this.flags & tf.Index);
    },
    getCallSignatures: cached((t) => checker.getSignaturesOfType(t, 0)),
    getConstructSignatures: cached((t) => checker.getSignaturesOfType(t, 1)),
    getProperties: cached((t) => checker.getPropertiesOfType(t)),
    getApparentProperties: cached((t) =>
      checker.getAugmentedPropertiesOfType(t),
    ),
    getProperty(name) {
      return checker.getPropertyOfType(this, name);
    },
    getNonNullableType: cached((t) => checker.getNonNullableType(t)),
    getConstraint: cached((t) => checker.getBaseConstraintOfType(t)),
    getDefault: cached((t) =>
      t.isTypeParameter() ? checker.getDefaultFromTypeParameter(t) : undefined,
    ),
    getBaseTypes: cached((t) =>
      t.isClassOrInterface() ? checker.getBaseTypes(t) : undefined,
    ),
    getStringIndexType: cached((t) => checker.getIndexTypeOfType(t, 0)),
    getNumberIndexType: cached((t) => checker.getIndexTypeOfType(t, 1)),
  } satisfies Partial<TypeView> & ThisType<TypeView>;
  for (const key of Object.keys(prototype)) {
    Object.defineProperty(prototype, key, { enumerable: false });
  }
  return prototype;
}

function inheritTypePrototype<T extends object>(
  fields: T,
  prototype: TypePrototype,
): asserts fields is T & TypePrototype {
  Object.setPrototypeOf(fields, prototype);
}

function readCompilerOptions(
  typescript: TypeScriptEnums,
  input: Record<string, unknown>,
) {
  const compilerOptions = { ...input };
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
  return compilerOptions;
}
// A retained checker cache reflects later queries in the same generation.
class LiveTypeMap implements ReadonlyMap<string, TypeView> {
  constructor(private readonly snapshot: () => Map<string, TypeView>) {}
  get size() {
    return this.snapshot().size;
  }
  get(key: string) {
    return this.snapshot().get(key);
  }
  has(key: string) {
    return this.snapshot().has(key);
  }
  keys() {
    return this.snapshot().keys();
  }
  values() {
    return this.snapshot().values();
  }
  entries() {
    return this.snapshot().entries();
  }
  [Symbol.iterator]() {
    return this.entries();
  }
  forEach(
    cb: (
      value: TypeView,
      key: string,
      map: ReadonlyMap<string, TypeView>,
    ) => void,
    thisArg?: unknown,
  ) {
    for (const [key, value] of this.entries())
      cb.call(thisArg, value, key, this);
  }
}
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
  private readonly indexInfos = new WeakObjectCache<number, IndexInfoView>();
  private readonly brands = new WeakMap<object, ObjectBrand>();
  private readonly conditionalRoots = new WeakMap<
    object,
    ConditionalRootView
  >();
  private shared: SharedTypeCacheReader | undefined;
  private readonly maxFactsBytes: number;
  private readonly syntax: GenerationSyntaxViews;
  private readonly typePrototype: TypePrototype;
  private readonly files = new Set<TypeServiceFile>();
  private batchTransport: BatchTransport | undefined;
  private factBytes = 0;
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
    this.syntax = new GenerationSyntaxViews(
      typescript,
      descriptor,
      syntax,
      (node) =>
        this.read({ op: 'nodeRecord', args: { node } }, (packet) =>
          packet.nodeFacts(),
        ),
      () => {
        this.ensure();
      },
      options.maxSyntaxBytes ?? 16 << 20,
      options.maxSyntaxFactsBytes ?? 1 << 20,
    );
    const nodeType = (op: 'type' | 'fromTypeNode' | 'context', n: object) =>
      this.type(
        this.query({ op: op, args: { node: this.syntax.location(n) } }),
      );
    const typeType = (
      op:
        | 'constraint'
        | 'apparent'
        | 'baseLiteral'
        | 'nonNullable'
        | 'awaited'
        | 'widened'
        | 'defaultType'
        | 'stringIndexType'
        | 'numberIndexType',
      t: object,
    ) => this.type(this.query({ op: op, args: { id: this.id(t, 'type') } }));
    this.checker = guarded<CheckerView>(
      {
        getAnyType: () => this.type(this.query({ op: 'anyType', args: {} }))!,
        getUnknownType: () =>
          this.type(this.query({ op: 'unknownType', args: {} }))!,
        getStringType: () =>
          this.type(this.query({ op: 'stringType', args: {} }))!,
        getNumberType: () =>
          this.type(this.query({ op: 'numberType', args: {} }))!,
        getBooleanType: () =>
          this.type(this.query({ op: 'booleanType', args: {} }))!,
        getBigIntType: () =>
          this.type(this.query({ op: 'bigIntType', args: {} }))!,
        getESSymbolType: () =>
          this.type(this.query({ op: 'eSSymbolType', args: {} }))!,
        getNonPrimitiveType: () =>
          this.type(this.query({ op: 'nonPrimitiveType', args: {} }))!,
        getNeverType: () =>
          this.type(this.query({ op: 'neverType', args: {} }))!,
        getNullType: () => this.type(this.query({ op: 'nullType', args: {} }))!,
        getUndefinedType: () =>
          this.type(this.query({ op: 'undefinedType', args: {} }))!,
        getTrueType: () => this.type(this.query({ op: 'trueType', args: {} }))!,
        getFalseType: () =>
          this.type(this.query({ op: 'falseType', args: {} }))!,
        getAmbientModules: () =>
          this.namedSymbolsFor(this.query({ op: 'ambientModules', args: {} })),
        getAugmentedPropertiesOfType: (t) =>
          this.namedSymbolsFor(
            this.query({
              op: 'augmentedPropertiesOfType',
              args: { id: this.id(t, 'type') },
            }),
          ),
        getRootSymbols: (s) =>
          this.namedSymbolsFor(
            this.query({
              op: 'rootSymbols',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
        getMergedSymbol: (s) =>
          this.symbol(
            this.query({
              op: 'mergedSymbol',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
        isArgumentsSymbol: (s) =>
          this.query({
            op: 'isArgumentsSymbol',
            args: { id: this.id(s, 'symbol') },
          }),
        isUndefinedSymbol: (s) =>
          this.query({
            op: 'isUndefinedSymbol',
            args: { id: this.id(s, 'symbol') },
          }),
        getExportSpecifierLocalTargetSymbol: (n) =>
          this.symbol(
            this.query({
              op: 'exportSpecifierLocalTargetSymbol',
              args: { node: this.syntax.location(n) },
            }),
          ),
        getPropertySymbolOfDestructuringAssignment: (n) =>
          this.symbol(
            this.query({
              op: 'propertySymbolOfDestructuringAssignment',
              args: { node: this.syntax.location(n) },
            }),
          ),
        getJsxIntrinsicTagNamesAt: (n) =>
          this.namedSymbolsFor(
            this.query({
              op: 'jsxIntrinsicTagNamesAt',
              args: { node: this.syntax.location(n) },
            }),
          ),
        getTypeOfAssignmentPattern: (n) =>
          this.type(
            this.query({
              op: 'typeOfAssignmentPattern',
              args: { node: this.syntax.location(n) },
            }),
          ),
        isOptionalParameter: (n) =>
          this.query({
            op: 'isOptionalParameter',
            args: { node: this.syntax.location(n) },
          }),
        getConstantValue: (n) =>
          this.query({
            op: 'constantValue',
            args: { node: this.syntax.location(n) },
          }),
        isImplementationOfOverload: (n) =>
          this.query({
            op: 'isImplementationOfOverload',
            args: { node: this.syntax.location(n) },
          }),
        getTypeArgumentsForResolvedSignature: (s) => {
          const ids = this.query({
            op: 'typeArgumentsForResolvedSignature',
            args: { id: this.id(s, 'signature') },
          });
          return ids && this.typesFor(ids);
        },
        getStringLiteralType: (value) =>
          this.type(
            this.query({
              op: 'stringLiteralType',
              args: { name: nameToUnits(value) },
            }),
          )!,
        getNumberLiteralType: (value) =>
          this.type(
            this.query({
              op: 'numberLiteralType',
              args: { value: Object.is(value, -0) ? '-0' : String(value) },
            }),
          )!,
        getBigIntLiteralType: (value) =>
          this.type(
            this.query({
              op: 'bigintLiteralType',
              args: { negative: value.negative, digits: value.base10Value },
            }),
          )!,
        getNullableType: (t, flags) =>
          this.type(
            this.query({
              op: 'nullableType',
              args: {
                id: this.id(t, 'type'),
                nullable: !!(flags & typescript.TypeFlags.Null),
                undefinable: !!(flags & typescript.TypeFlags.Undefined),
              },
            }),
          )!,
        getPrivateIdentifierPropertyOfType: (t, name, location) =>
          this.symbol(
            this.query({
              op: 'privateIdentifierProperty',
              args: {
                id: this.id(t, 'type'),
                name: nameToUnits(name),
                node: this.syntax.location(location),
              },
            }),
          ),
        tryGetMemberInModuleExports: (name, moduleSymbol) =>
          this.symbol(
            this.query({
              op: 'moduleMember',
              args: {
                id: this.id(moduleSymbol, 'symbol'),
                name: nameToUnits(name),
              },
            }),
          ),
        isValidPropertyAccess: (n, name) =>
          this.query({
            op: 'validPropertyAccess',
            args: { node: this.syntax.location(n), name: nameToUnits(name) },
          }),
        getSymbolsOfParameterPropertyDeclaration: (n, name) =>
          this.symbolsFor(
            this.query({
              op: 'parameterPropertySymbols',
              args: { node: this.syntax.location(n), name: nameToUnits(name) },
            }),
          ),
        getSymbolOfExpando: (n, allowDeclaration) =>
          this.symbol(
            this.query({
              op: 'expandoSymbol',
              args: { node: this.syntax.location(n), allowDeclaration },
            }),
          ),
        getIndexInfosOfIndexSymbol: (s, siblings) =>
          this.query({
            op: 'indexSymbolInfos',
            args: {
              id: this.id(s, 'symbol'),
              siblings: siblings?.map((sibling) => this.id(sibling, 'symbol')),
            },
          }).map((info) => this.indexInfo(info)),
        signatureToSignatureDeclaration: (s, kind, enclosing, flags) => {
          const name = typescript.SyntaxKind[kind];
          const nativeKind: unknown = Reflect.get(nativeSyntaxKind, name);
          if (typeof nativeKind !== 'number') unsupported(`SyntaxKind.${name}`);
          const data = this.query({
            op: 'signatureDeclaration',
            args: {
              id: this.id(s, 'signature'),
              kind: nativeKind,
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          });
          return data && this.syntax.generatedTypeNode(data);
        },
        symbolToParameterDeclaration: (s, enclosing, flags) => {
          const data = this.query({
            op: 'parameterDeclaration',
            args: {
              id: this.id(s, 'symbol'),
              kind: 0,
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          });
          return data && this.syntax.generatedTypeNode(data);
        },
        symbolToTypeParameterDeclarations: (s, enclosing, flags) => {
          const data = this.query({
            op: 'typeParameterDeclarations',
            args: {
              id: this.id(s, 'symbol'),
              kind: 0,
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          });
          if (!data) return undefined;
          const parameters: unknown = Reflect.get(
            this.syntax.generatedTypeNode(data),
            'typeParameters',
          );
          if (parameters === undefined || Array.isArray(parameters))
            return parameters;
          throw new Error('invalid generated type parameter list');
        },
        typeParameterToDeclaration: (t, enclosing, flags) => {
          const data = this.query({
            op: 'typeParameterDeclaration',
            args: {
              id: this.id(t, 'type'),
              kind: 0,
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          });
          return data && this.syntax.generatedTypeNode(data);
        },
        symbolToEntityName: (s, meaning, enclosing, flags) => {
          const data = this.query({
            op: 'entityName',
            args: {
              id: this.id(s, 'symbol'),
              kind: meaning >>> 0,
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          });
          return data && this.syntax.generatedTypeNode(data);
        },
        symbolToExpression: (s, meaning, enclosing, flags) => {
          const data = this.query({
            op: 'symbolExpression',
            args: {
              id: this.id(s, 'symbol'),
              kind: meaning >>> 0,
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          });
          return data && this.syntax.generatedTypeNode(data);
        },
        symbolToString: (s, enclosing, meaning, flags) =>
          this.query({
            op: 'symbolString',
            args: {
              id: this.id(s, 'symbol'),
              kind: (meaning ?? 0) >>> 0,
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          }),
        typePredicateToString: (predicate, enclosing, flags) =>
          this.query({
            op: 'predicateString',
            args: {
              kind: predicate.kind,
              name:
                predicate.parameterName === undefined
                  ? undefined
                  : nameToUnits(predicate.parameterName),
              position: predicate.parameterIndex,
              type: predicate.type ? this.id(predicate.type, 'type') : 0,
              node: enclosing && this.syntax.location(enclosing),
              flags: flags ?? 16384,
            },
          }),
        indexInfoToIndexSignatureDeclaration: (info, enclosing, flags) => {
          const data = this.query({
            op: 'indexDeclaration',
            args: {
              key: this.id(info.keyType, 'type'),
              value: this.id(info.type, 'type'),
              readonly: info.isReadonly,
              declaration:
                info.declaration && this.syntax.location(info.declaration),
              node: enclosing && this.syntax.location(enclosing),
              flags,
            },
          });
          return data && this.syntax.generatedTypeNode(data);
        },
        getTypeAtLocation: (n) => nodeType('type', n),
        getTypeFromTypeNode: (n) => nodeType('fromTypeNode', n),
        getContextualType: (n) => nodeType('context', n),
        getContextualTypeForArgumentAtIndex: (n, index) =>
          this.type(
            this.query({
              op: 'contextualArgument',
              args: { node: this.syntax.location(n), index },
            }),
          ),
        getFullyQualifiedName: (symbol) =>
          this.query({
            op: 'fullyQualifiedName',
            args: { id: this.id(symbol, 'symbol') },
          }),
        typeToTypeNode: (t, enclosing, flags) => {
          const data = this.query({
            op: 'typeNode',
            args: {
              id: this.id(t, 'type'),
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          });
          return data && this.syntax.generatedTypeNode(data);
        },
        getSuggestionDiagnostics: (source, cancellation) => {
          cancellation?.throwIfCancellationRequested();
          const diagnostics = this.query({
            op: 'suggestionDiagnostics',
            args: { node: this.syntax.location(source) },
          });
          cancellation?.throwIfCancellationRequested();
          return diagnostics.map((d) => this.diagnostic(d));
        },
        runWithCancellationToken: (token, cb) => {
          token.throwIfCancellationRequested();
          const previous = this.cancellation;
          this.cancellation = token;
          try {
            const result = cb(this.checker);
            token.throwIfCancellationRequested();
            return result;
          } finally {
            this.cancellation = previous;
          }
        },
        getVoidType: () => this.type(this.query({ op: 'void', args: {} })),
        getBaseConstraintOfType: (t) => typeType('constraint', t),
        getApparentType: (t) => typeType('apparent', t),
        getBaseTypeOfLiteralType: (t) => typeType('baseLiteral', t),
        getNonNullableType: (t) => typeType('nonNullable', t),
        getAwaitedType: (t) => typeType('awaited', t),
        getWidenedType: (t) => typeType('widened', t),
        getDefaultFromTypeParameter: (t) => typeType('defaultType', t),
        getBaseTypes: (t) =>
          this.typesFor(
            this.query({ op: 'baseTypes', args: { id: this.id(t, 'type') } }),
          ),
        isTypeAssignableTo: (source, target) =>
          this.query({
            op: 'isTypeAssignableTo',
            args: {
              source: this.id(source, 'type'),
              target: this.id(target, 'type'),
            },
          }),
        typeToString: (t, enclosing, flags) =>
          this.query({
            op: 'typeString',
            args: {
              id: this.id(t, 'type'),
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
            },
          }),
        signatureToString: (s, enclosing, flags, kind = 0) =>
          this.query({
            op: 'signatureString',
            args: {
              id: this.id(s, 'signature'),
              node: enclosing && this.syntax.location(enclosing),
              flags: flags === undefined ? undefined : flags >>> 0,
              kind,
            },
          }),
        getSymbolAtLocation: (n) =>
          this.symbol(
            this.query({
              op: 'symbol',
              args: { node: this.syntax.location(n) },
            }),
          ),
        getSymbolsInScope: (n, meaning) =>
          this.namedSymbolsFor(
            this.query({
              op: 'symbolsInScope',
              args: { node: this.syntax.location(n), meaning: meaning | 0 },
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
        getImmediateAliasedSymbol: (s) =>
          this.symbol(
            this.query({
              op: 'immediateAliasedSymbol',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
        isUnknownSymbol: (s) =>
          this.query({
            op: 'isUnknownSymbol',
            args: { id: this.id(s, 'symbol') },
          }),
        getExportsOfModule: (s) =>
          this.namedSymbolsFor(
            this.query({
              op: 'exportsOfModule',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
        getExportSymbolOfSymbol: (s) =>
          this.symbol(
            this.query({
              op: 'exportSymbol',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
        getShorthandAssignmentValueSymbol: (n) =>
          n &&
          this.symbol(
            this.query({
              op: 'shorthandValueSymbol',
              args: { node: this.syntax.location(n) },
            }),
          ),
        getTypeOfSymbol: (s) =>
          this.type(
            this.query({
              op: 'typeOfSymbol',
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
        getTypeOfPropertyOfType: (t, name) =>
          this.type(
            this.query({
              op: 'propertyType',
              args: { id: this.id(t, 'type'), name: nameToUnits(name) },
            }),
          ),
        getIndexTypeOfType: (t, kind) =>
          typeType(kind === 0 ? 'stringIndexType' : 'numberIndexType', t),
        getIndexInfoOfType: (t, kind) => {
          const [info] = this.query({
            op: 'indexInfo',
            args: { id: this.id(t, 'type'), kind },
          });
          return info && this.indexInfo(info);
        },
        getIndexInfosOfType: (t) =>
          this.query({
            op: 'indexInfos',
            args: { id: this.id(t, 'type') },
          }).map((info) => this.indexInfo(info)),
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
          const node = this.syntax.location(n);
          if (candidates !== undefined || argumentCount !== undefined) {
            if (candidates?.length)
              throw new Error('signature candidate output array must be empty');
            const [resolved, ...ids] = this.query({
              op: 'resolvedSignatureWithOptions',
              args: {
                node,
                withCandidates: candidates !== undefined,
                argumentCount,
              },
            });
            if (candidates && ids.length)
              candidates.push(...this.signaturesFor(ids));
            return this.signature(resolved);
          }
          return this.signature(
            this.query({ op: 'resolvedSignature', args: { node } }),
          );
        },
        getReturnTypeOfSignature: (s) =>
          this.type(
            this.query({
              op: 'returnType',
              args: { id: this.id(s, 'signature') },
            }),
          ),
        getTypePredicateOfSignature: (s) => {
          const fact = this.query({
            op: 'typePredicate',
            args: { id: this.id(s, 'signature') },
          });
          return fact && { ...fact, type: this.type(fact.type) };
        },
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
    );
    const compilerOptions = readCompilerOptions(
      typescript,
      descriptor.compilerOptions,
    );
    const ensure = this.ensure.bind(this);
    this.typePrototype = makeTypePrototype(this.checker, typescript, ensure);
    this.program = guarded<ProgramView>(
      {
        getTypeChecker: () => this.checker,
        getCompilerOptions: () => ({ ...compilerOptions }),
        getCurrentDirectory: () => descriptor.currentDirectory,
        emit: () => {
          throw new Error('unsupported TypeScript operation: Program.emit');
        },
        getSourceFileByPath: (path) => this.syntax.sourceByPath(path),
        getConfigFileParsingDiagnostics: () =>
          this.diagnostics({ op: 'configFileParsingDiagnostics', args: {} }),
        getOptionsDiagnostics: (cancellation) =>
          this.diagnostics(
            { op: 'optionsDiagnostics', args: {} },
            cancellation,
          ),
        getGlobalDiagnostics: (cancellation) =>
          this.diagnostics({ op: 'globalDiagnostics', args: {} }, cancellation),
        getSyntacticDiagnostics: (source, cancellation) =>
          this.diagnostics(
            {
              op: 'syntacticDiagnostics',
              args: { node: source && this.syntax.location(source) },
            },
            cancellation,
          ),
        getSemanticDiagnostics: (source, cancellation) =>
          this.diagnostics(
            {
              op: 'semanticDiagnostics',
              args: { node: source && this.syntax.location(source) },
            },
            cancellation,
          ),
        getDeclarationDiagnostics: (source, cancellation) =>
          this.diagnostics(
            {
              op: 'declarationDiagnostics',
              args: { node: source && this.syntax.location(source) },
            },
            cancellation,
          ),
        getNodeCount: () => this.query({ op: 'nodeCount', args: {} }),
        getIdentifierCount: () =>
          this.query({ op: 'identifierCount', args: {} }),
        getSymbolCount: () => this.query({ op: 'symbolCount', args: {} }),
        getTypeCount: () => this.query({ op: 'typeCount', args: {} }),
        getInstantiationCount: () =>
          this.query({ op: 'instantiationCount', args: {} }),
        getRelationCacheSizes: () => {
          const [assignable, identity, subtype, strictSubtype] = this.query({
            op: 'relationCacheSizes',
            args: {},
          });
          return { assignable, identity, subtype, strictSubtype };
        },
        getRootFileNames: () => this.query({ op: 'rootFileNames', args: {} }),
        getModeForUsageLocation: (source, usage) =>
          this.query({
            op: 'modeForUsageLocation',
            args: {
              file: this.syntax.location(source).file,
              node: this.syntax.location(usage),
            },
          }) || undefined,
        getModeForResolutionAtIndex: (source, index) =>
          this.query({
            op: 'modeForResolutionAtIndex',
            args: { node: this.syntax.location(source), index },
          }) || undefined,
        getProjectReferences: () => {
          const refs: ts.ProjectReference[] | null = JSON.parse(
            this.query({ op: 'projectReferences', args: {} }),
          );
          return refs ?? undefined;
        },
        getResolvedProjectReferences: () => this.resolvedReferences(),
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
    );
  }
  private resolvedReferences():
    readonly (ResolvedReferenceView | undefined)[] | undefined {
    const graph: { roots: number[]; configs: ReferenceConfig[] } | null =
      JSON.parse(this.query({ op: 'resolvedProjectReferences', args: {} }));
    if (!graph) return undefined;
    const configs: ResolvedReferenceView[] = graph.configs.map((config) => {
      const {
        name,
        path,
        options,
        diagnostics,
        references: _references,
        extendedSourceFiles,
        ...commandLine
      } = config;
      const errors = diagnostics
        ? new TypePacketReader(Buffer.from(diagnostics, 'base64'))
            .scalar('configFileParsingDiagnostics')
            .map((d) => this.diagnostic(d))
        : [];
      const sourceFile = this.syntax.configSource(name, path);
      if (
        extendedSourceFiles &&
        !Object.hasOwn(sourceFile, 'extendedSourceFiles')
      )
        Object.assign(sourceFile, { extendedSourceFiles });
      return {
        sourceFile,
        commandLine: {
          ...commandLine,
          fileNames: commandLine.fileNames ?? [],
          options: readCompilerOptions(this.typescript, options),
          errors,
        },
      };
    });
    const lookup = (id: number) => configs[id - 1];
    graph.configs.forEach((config, i) => {
      if (config.references)
        configs[i].references = config.references.map(lookup);
    });
    return graph.roots.map(lookup);
  }
  private diagnostics(
    request: ServiceRequest<
      | 'configFileParsingDiagnostics'
      | 'optionsDiagnostics'
      | 'globalDiagnostics'
      | 'syntacticDiagnostics'
      | 'semanticDiagnostics'
      | 'declarationDiagnostics'
    >,
    token?: ts.CancellationToken,
  ): DiagnosticView[] {
    token?.throwIfCancellationRequested();
    const result = this.query(request).map((d) => this.diagnostic(d));
    token?.throwIfCancellationRequested();
    return result;
  }
  private diagnostic(d: DiagnosticFact): DiagnosticView {
    return {
      file: d.fileName ? this.syntax.diagnosticSource(d.fileName) : undefined,
      start: d.fileName ? d.start : undefined,
      length: d.fileName ? d.length : undefined,
      code: d.code,
      category: d.category,
      source: d.source,
      messageText: d.chain.length ? diagnosticMessage(d) : d.text,
      reportsDeprecated: d.reportsDeprecated,
      reportsUnnecessary: d.reportsUnnecessary,
      relatedInformation: d.related.length
        ? d.related.map((item) => this.diagnostic(item))
        : undefined,
    };
  }
  private cancellation?: ts.CancellationToken;
  private ensure(): void {
    this.cancellation?.throwIfCancellationRequested();
    if (this.closed) throw new Error('type service client is closed');
    this.shared?.assertOpen();
  }
  close(): void {
    this.closed = true;
    for (const reader of this.facts.values()) reader.close();
    this.facts.clear();
    this.factBytes = 0;
    this.types.clear();
    this.symbols.clear();
    this.signatures.clear();
    this.indexInfos.clear();
    for (const file of this.files) file.close();
    this.files.clear();
    this.syntax.close();
    this.shared?.close();
    this.shared = undefined;
    this.transport = undefined;
    this.batchTransport = undefined;
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
    const key = typeQueryKey(query);
    if (
      !isCacheableTypeQuery(query) &&
      !(
        (query.op === 'type' || query.op === 'constantValue') &&
        this.syntax.isStableNodeQuery(query.op, query.args.node)
      )
    ) {
      const reader = this.ingest(
        query,
        this.transport!(query, key, false),
        false,
      );
      try {
        return read(reader);
      } finally {
        reader.close();
      }
    }
    let reader = this.facts.get(key);
    if (reader) {
      this.facts.delete(key);
      this.facts.set(key, reader);
    } else {
      const shared = this.shared?.get(key);
      reader = this.ingest(query, shared ?? this.transport!(query, key, true));
    }
    try {
      return read(reader);
    } finally {
      if (this.facts.get(key) !== reader) reader.close();
    }
  }
  private store(key: string, bytes: Uint8Array): TypePacketReader {
    const existing = this.facts.get(key);
    if (existing) return existing;
    const reader = new TypePacketReader(Uint8Array.from(bytes));
    if (reader.retainedBytes <= this.maxFactsBytes) {
      while (this.factBytes + reader.retainedBytes > this.maxFactsBytes) {
        const oldest = this.facts.keys().next().value!,
          evicted = this.facts.get(oldest)!;
        this.factBytes -= evicted.retainedBytes;
        evicted.close();
        this.facts.delete(oldest);
      }
      this.facts.set(key, reader);
      this.factBytes += reader.retainedBytes;
    }
    return reader;
  }
  private ingest(
    query: TypeServiceQuery,
    bytes: Uint8Array,
    retain = true,
  ): TypePacketReader {
    const reply = typeReplyPackets(bytes);
    for (const record of reply.records) {
      const key = typeQueryKey({
        ...query,
        op: record.op,
        args: { id: record.id },
      });
      const reader = this.store(key, record.packet);
      if (this.facts.get(key) !== reader) reader.close();
    }
    return retain
      ? this.store(typeQueryKey(query), reply.packet)
      : new TypePacketReader(reply.packet);
  }
  /** Merge known misses only. Cache hits, including negative answers, never
   * cross the bridge. The scalar API remains the authoritative fallback. */
  prefetch(requests: readonly ServiceRequest[]): void {
    this.ensure();
    if (!this.batchTransport) return;
    const missing = new Map<string, TypeServiceQuery>();
    for (const request of requests) {
      const query = {
        generation: this.descriptor.generation,
        checker: this.descriptor.checker,
        ...request,
      };
      if (!isCacheableTypeQuery(query)) continue;
      const key = typeQueryKey(query);
      if (this.facts.has(key) || missing.has(key)) continue;
      const shared = this.shared?.get(key);
      if (shared) {
        const reader = this.ingest(query, shared);
        if (this.facts.get(key) !== reader) reader.close();
      } else missing.set(key, query);
    }
    const all = [...missing.entries()];
    for (let offset = 0; offset < all.length; offset += 256) {
      const batch = all.slice(offset, offset + 256),
        queries = batch.map(([, query]) => query),
        keys = batch.map(([key]) => key);
      const packets = this.batchTransport(queries, keys);
      if (packets.length !== batch.length)
        throw new Error('incomplete type service batch');
      packets.forEach((packet, index) => {
        const reader = this.ingest(queries[index], packet);
        if (this.facts.get(keys[index]) !== reader) reader.close();
      });
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
    const brand =
      this.brands.get(object) ?? this.brands.get(canonicalView(object));
    if (brand?.kind !== kind) throw new Error(`foreign ${kind} handle`);
    return brand.id;
  }
  private intern<T extends object>(
    map: WeakObjectCache<number, T>,
    kind: ObjectKind,
    id: number,
    create: (client: this, brand: ObjectBrand) => T,
  ): T | undefined {
    this.ensure();
    if (!id) return undefined;
    const old = map.get(id);
    if (old) return old;
    const brand = { kind, id };
    const fields = create(this, brand);
    const object =
      kind === 'type'
        ? fields
        : guarded(fields, () => {
            this.ensure();
          });
    this.brands.set(object, brand);
    map.set(id, object);
    return object;
  }
  private typesFor(ids: readonly number[]): TypeView[] {
    return ids.map((id) => this.type(id)!);
  }
  private symbolsFor(ids: readonly number[]): SymbolView[] {
    return ids.map((id) => this.symbol(id)!);
  }
  private namedSymbolsFor(facts: readonly NamedSymbolFact[]): SymbolView[] {
    return facts.map(({ id, name }) => this.symbol(id, name)!);
  }
  private signaturesFor(ids: readonly number[]): SignatureView[] {
    return ids.map((id) => this.signature(id)!);
  }
  private indexInfo(fact: IndexInfoFact): IndexInfoView {
    const old = fact.id && this.indexInfos.get(fact.id);
    if (old) return old;
    const info = readonlyDataFields({
      components:
        fact.components &&
        Object.freeze(fact.components.map((ref) => this.syntax.nodeRef(ref)!)),
      keyType: this.type(fact.keyType)!,
      type: this.type(fact.type)!,
      isReadonly: fact.isReadonly,
      declaration: this.syntax.nodeRef(fact.declaration),
    });
    if (fact.id) this.indexInfos.set(fact.id, info);
    return info;
  }
  private typeFieldNames?: ReadonlySet<PropertyKey>;
  private canInspectType(): boolean {
    this.cancellation?.throwIfCancellationRequested();
    return !this.closed && (this.shared?.isOpen ?? true);
  }
  private wrapType(
    fields: TypeView,
    hasField: (
      key: PropertyKey,
      includePending?: boolean,
      mapped?: readonly number[],
    ) => boolean,
    mappedFields: () => readonly number[] | undefined,
    hasDeclaredMembers: () => boolean,
  ): TypeView {
    Object.defineProperties(fields, {
      id: { writable: false },
      checker: { writable: false },
    });
    const native = (this.typeFieldNames ??= new Set(Reflect.ownKeys(fields)));
    // A getter returning this can make Node unwrap more than one proxy. The
    // ordinary target holds descriptors for reflection, never the rule's state.
    const target: TypeView = Object.create(this.typePrototype);
    Object.defineProperty(fields, inspectCustom, typeInspection);
    let pluginFields: Set<PropertyKey> | undefined;
    const managed = (key: PropertyKey) =>
      native.has(key) && !pluginFields?.has(key);
    const nativeDescriptor = (key: PropertyKey) => {
      const previous = Reflect.getOwnPropertyDescriptor(target, key);
      if (previous && !previous.configurable) return previous;
      const descriptor = Reflect.getOwnPropertyDescriptor(fields, key)!;
      descriptor.configurable = false;
      Object.defineProperty(target, key, descriptor);
      return descriptor;
    };
    const own = (key: PropertyKey, live = true) => {
      const descriptor = Reflect.getOwnPropertyDescriptor(fields, key);
      if (!descriptor || !managed(key)) return descriptor;
      const fixed = Reflect.getOwnPropertyDescriptor(target, key);
      if (fixed && !fixed.configurable) return fixed;
      return key === 'id' || (live && hasField(key))
        ? nativeDescriptor(key)
        : undefined;
    };
    const keys = (live = true) => {
      const all = Reflect.ownKeys(fields);
      if (!Object.isExtensible(fields)) return all;
      if (!live) return all.filter((key) => own(key, false));
      const declared = hasDeclaredMembers(),
        mapped = mappedFields();
      return all.filter((key) => {
        if (!managed(key)) return true;
        if (
          !(declaredMemberKeys.has(key)
            ? declared
            : hasField(key, false, mapped))
        )
          return false;
        nativeDescriptor(key);
        return true;
      });
    };
    const definePluginField = (
      key: PropertyKey,
      attributes: PropertyDescriptor,
    ) => {
      if (!Reflect.defineProperty(fields, key, attributes)) return false;
      const descriptor = Reflect.getOwnPropertyDescriptor(fields, key)!;
      if (!descriptor.configurable || !Object.isExtensible(target))
        Object.defineProperty(target, key, descriptor);
      return true;
    };
    const backing: TypeView = new Proxy(target, {
      set: rejectTypeEdit,
      defineProperty: rejectTypeEdit,
      deleteProperty: rejectTypeEdit,
      preventExtensions: rejectTypeEdit,
      setPrototypeOf: rejectTypeEdit,
      get: (_target, key, receiver: object) => {
        const self =
          receiver === backing &&
          (key === Symbol.toStringTag || key === inspectCustom)
            ? view
            : receiver;
        if (key === serviceView) return view;
        if (key === Symbol.toStringTag && !this.canInspectType())
          return Reflect.getOwnPropertyDescriptor(fields, key)?.value;
        if (receiver !== backing && managed(key) && key !== 'id') this.ensure();
        const value: unknown = Reflect.get(fields, key, self);
        if (value === undefined && managed(key)) {
          const proto = Reflect.getPrototypeOf(fields);
          if (proto && Reflect.has(proto, key) && !hasField(key))
            return Reflect.get(proto, key, self);
        }
        return value;
      },
      has: (_target, key) => {
        if (!managed(key) || !Object.hasOwn(fields, key))
          return Reflect.has(fields, key);
        const proto = Reflect.getPrototypeOf(fields);
        return (
          !!own(key, this.canInspectType()) ||
          (proto !== null && Reflect.has(proto, key))
        );
      },
      ownKeys: () => keys(this.canInspectType()),
      getOwnPropertyDescriptor: (_object, key) =>
        own(key, this.canInspectType()),
    });
    const view = new Proxy(backing, {
      setPrototypeOf: (_target, proto) => {
        this.ensure();
        return (
          Reflect.setPrototypeOf(target, proto) &&
          Reflect.setPrototypeOf(fields, proto)
        );
      },
      set: (_target, key, value, receiver) => {
        this.ensure();
        if (managed(key) && !hasField(key, true))
          return Reflect.set(
            Reflect.getPrototypeOf(fields) ?? Object.create(null),
            key,
            value,
            receiver,
          );
        return Reflect.set(fields, key, value, receiver);
      },
      defineProperty: (_target, key, attributes) => {
        this.ensure();
        if (key === 'typeArguments') return false;
        if (!managed(key)) return definePluginField(key, attributes);
        if (hasField(key, true))
          return !!own(key) && Reflect.defineProperty(target, key, attributes);
        // An absent placeholder must not supply the descriptor defaults for
        // a new plugin property.
        if (!Object.isExtensible(fields)) return false;
        Reflect.deleteProperty(fields, key);
        (pluginFields ??= new Set()).add(key);
        return definePluginField(key, attributes);
      },
      deleteProperty: (_target, key) => {
        this.ensure();
        return managed(key)
          ? !own(key)
          : Reflect.deleteProperty(target, key) &&
              Reflect.deleteProperty(fields, key);
      },
      preventExtensions: () => {
        this.ensure();
        if (Object.isExtensible(fields)) {
          const keep = new Set(keys());
          for (const key of Reflect.ownKeys(target))
            if (!keep.has(key) && !Reflect.deleteProperty(target, key))
              return false;
          for (const key of Reflect.ownKeys(fields)) {
            if (!keep.has(key)) Reflect.deleteProperty(fields, key);
            else if (!managed(key))
              Object.defineProperty(
                target,
                key,
                Reflect.getOwnPropertyDescriptor(fields, key)!,
              );
          }
          Reflect.preventExtensions(target);
        }
        return Reflect.preventExtensions(fields);
      },
    });
    return view;
  }
  private type(id: number): TypeView | undefined {
    const create = (client: this): TypeView => {
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
      const details = memo(ensure, () =>
        this.query({ op: 'typeDetails', args: { id } }),
      );
      const typeList = (ids: readonly number[] | undefined) =>
        ids && Object.freeze(this.typesFor(ids));
      const conditionalRoot = memo(ensure, () => {
        const d = details();
        const node = this.syntax.nodeRef(d.root);
        if (!node) return undefined;
        const old = this.conditionalRoots.get(node);
        if (old) return old;
        const root: ConditionalRootView = {
          node,
          checkType: this.type(d.rootCheck)!,
          extendsType: this.type(d.rootExtends)!,
          isDistributive: d.distributive,
          inferTypeParameters: typeList(d.infer),
          outerTypeParameters: typeList(d.rootOuter),
          aliasSymbol: this.symbol(d.rootAlias),
          aliasTypeArguments: typeList(d.rootArguments),
          instantiations: d.rootOuter?.length
            ? new LiveTypeMap(
                () =>
                  new Map(
                    this.query({
                      op: 'conditionalInstantiations',
                      args: { id },
                    }).map((entry) => [entry.name, this.type(entry.id)!]),
                  ),
              )
            : undefined,
        };
        readonlyDataFields(root);
        this.conditionalRoots.set(node, root);
        return root;
      });
      const objectFlags = memo(ensure, () =>
        this.typeField(id, 'objectFlags')?.reduce((value, name) => {
          const flag = Reflect.get(this.typescript.ObjectFlags, name);
          if (typeof flag !== 'number') unsupported(`ObjectFlags.${name}`);
          return value | flag;
        }, 0),
      );
      const symbolId = memo(ensure, () => this.typeField(id, 'symbol'));
      const aliasId = memo(ensure, () => this.typeField(id, 'aliasSymbol'));
      const symbol = memo(ensure, () => this.symbol(symbolId()));
      const alias = memo(ensure, () => this.symbol(aliasId()));
      const intrinsic = memo(ensure, () => this.typeField(id, 'intrinsicName')),
        escapedName = memo(ensure, () => this.typeField(id, 'escapedName')),
        value = memo(ensure, () => {
          const literal = this.typeField(id, 'value');
          return typeof literal === 'object'
            ? readonlyDataFields(literal)
            : literal;
        });
      const typeOperands = memo(ensure, () =>
        typeList(this.typeField(id, 'typeOperands')),
      );
      const targetId = memo(ensure, () => this.typeField(id, 'target'));
      const target = memo(ensure, () => this.type(targetId())),
        aliasTypeArguments = memo(ensure, () =>
          typeList(this.typeField(id, 'aliasTypeArguments')),
        ),
        typeParameters = memo(ensure, () =>
          typeList(this.typeField(id, 'typeParameters')),
        ),
        tuple = memo(ensure, () => this.typeField(id, 'tuple'));
      const flag = (mask: number) => Boolean(flags() & mask),
        objectFlag = (mask: number) => Boolean((objectFlags() ?? 0) & mask),
        tf = this.typescript.TypeFlags,
        of = this.typescript.ObjectFlags;
      let mappedIds: readonly number[] | undefined;
      const mappedFields = (index?: number) => {
        if (!objectFlag(of.Mapped)) return undefined;
        if (
          !mappedIds ||
          !(index === undefined ? mappedIds.every(Boolean) : mappedIds[index])
        )
          mappedIds = this.query({ op: 'mappedFields', args: { id } });
        return mappedIds;
      };
      const mapped = (key: keyof typeof mappedFieldIndex) => {
        const index = mappedFieldIndex[key];
        return this.type(mappedFields(index)?.[index] ?? 0);
      };
      let fresh = 0,
        regular = 0;
      const freshId = () =>
        flag(tf.Literal | tf.Enum)
          ? (fresh ||= this.query({ op: 'freshLiteral', args: { id } }))
          : 0;
      const regularId = () => {
        if (flag(tf.Literal | tf.Enum)) return details().regular;
        if (flag(tf.Union) || objectFlag(of.FreshLiteral))
          return (regular ||= this.query({ op: 'regularType', args: { id } }));
        return 0;
      };
      let memberFact: QueryResults['declaredMembers'];
      const declared = () =>
        objectFlag(of.ClassOrInterface | of.Tuple)
          ? (memberFact ??= this.query({ op: 'declaredMembers', args: { id } }))
          : undefined;
      let resolvedMembers: DeclaredMemberViews | undefined;
      const declaredMembers = () => {
        const fact = declared();
        if (fact && !resolvedMembers) {
          resolvedMembers = {
            declaredProperties: Object.freeze(this.symbolsFor(fact.properties)),
            declaredCallSignatures: Object.freeze(
              this.signaturesFor(fact.calls),
            ),
            declaredConstructSignatures: Object.freeze(
              this.signaturesFor(fact.constructs),
            ),
            declaredIndexInfos: Object.freeze(
              fact.indices.map((info) => this.indexInfo(info)),
            ),
          };
        }
        return resolvedMembers;
      };
      // Presence reads flags, handles and facts. Only a field getter constructs
      // member views. A writable plugin property may use an inapplicable name,
      // but cannot shadow a compiler slot that another client can populate.
      const hasField = (
        key: PropertyKey,
        includePending = false,
        mappedIds?: readonly number[],
      ): boolean => {
        switch (key) {
          case 'id':
          case 'checker':
          case 'flags':
            return true;
          case 'types':
            return flag(tf.UnionOrIntersection | tf.TemplateLiteral);
          case 'type':
            return flag(tf.Index | tf.StringMapping);
          case 'objectType':
          case 'indexType':
            return flag(tf.IndexedAccess);
          case 'root':
          case 'checkType':
          case 'extendsType':
            return flag(tf.Conditional);
          case 'target':
            return !!targetId();
          case 'typeParameters':
          case 'localTypeParameters':
          case 'outerTypeParameters':
          case 'thisType':
            return (
              objectFlag(of.ClassOrInterface | of.Tuple) && targetId() === id
            );
          case 'symbol':
            return (
              flag(
                tf.Object |
                  tf.TypeParameter |
                  tf.StringLiteral |
                  tf.NumberLiteral |
                  tf.BigIntLiteral |
                  tf.StringMapping,
              ) || !!symbolId()
            );
          case 'aliasSymbol':
          case 'aliasTypeArguments':
            return (
              !!aliasId() ||
              flag(
                tf.UnionOrIntersection | tf.Conditional | tf.IndexedAccess,
              ) ||
              objectFlag(of.Mapped | of.Instantiated) ||
              (objectFlag(of.Anonymous) &&
                !!symbolId() &&
                (this.symbolField(symbolId(), 'flags') as ts.SymbolFlags) ===
                  this.typescript.SymbolFlags.TypeLiteral) ||
              (objectFlag(of.Reference) && details().node !== undefined)
            );
          case 'elementFlags':
          case 'combinedFlags':
          case 'readonly':
          case 'minLength':
          case 'fixedLength':
          case 'hasRestElement':
          case 'labeledElementDeclarations':
            return tuple() !== undefined;
          case 'typeParameter':
          case 'constraintType':
          case 'templateType':
          case 'modifiersType': {
            const index = mappedFieldIndex[key];
            return (
              objectFlag(of.Mapped) &&
              (includePending || !!(mappedIds ?? mappedFields(index))?.[index])
            );
          }
          case 'objectFlags':
            return objectFlags() !== undefined;
          case 'freshType':
            return (
              flag(tf.Literal | tf.Enum) && (includePending || !!freshId())
            );
          case 'regularType':
            return (
              (includePending &&
                (flag(tf.Literal | tf.Enum | tf.Union) ||
                  objectFlag(of.FreshLiteral))) ||
              !!regularId()
            );
          case 'simplifiedForReading':
          case 'simplifiedForWriting':
            return (
              flag(tf.IndexedAccess) &&
              (includePending || !!this.query({ op: key, args: { id } }))
            );
          case 'resolvedTrueType':
          case 'resolvedFalseType':
            return (
              flag(tf.Conditional) &&
              (includePending || !!this.query({ op: key, args: { id } }))
            );
          case 'elementType':
          case 'finalArrayType':
            return (
              objectFlag(of.EvolvingArray) &&
              (includePending ||
                !!this.query({
                  op: key === 'elementType' ? 'evolvingElement' : 'finalArray',
                  args: { id },
                }))
            );
          case 'constraint':
            return flag(tf.TypeParameter)
              ? includePending ||
                  !!this.query({ op: 'typeParameterConstraint', args: { id } })
              : flag(tf.Substitution) && !!details().constraint;
          case 'intrinsicName':
            return intrinsic() !== undefined;
          case 'escapedName':
            return escapedName() !== undefined;
          case 'value':
            return value() !== undefined;
          case 'pattern':
          case 'node':
          case 'texts':
            return details()[key] !== undefined;
          case 'baseType':
            return !!details().base;
          default:
            return (
              declaredMemberKeys.has(key) &&
              objectFlag(of.ClassOrInterface | of.Tuple) &&
              (includePending || !!declared())
            );
        }
      };
      const fields = {
        id,
        checker: client.checker,
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
        get declaredProperties() {
          return declaredMembers()?.declaredProperties;
        },
        get declaredCallSignatures() {
          return declaredMembers()?.declaredCallSignatures;
        },
        get declaredConstructSignatures() {
          return declaredMembers()?.declaredConstructSignatures;
        },
        get declaredIndexInfos() {
          return declaredMembers()?.declaredIndexInfos;
        },
        get objectFlags() {
          return objectFlags();
        },
        get value() {
          return value();
        },
        get types() {
          return flag(tf.UnionOrIntersection | tf.TemplateLiteral)
            ? typeOperands()
            : undefined;
        },
        get type() {
          return flag(tf.Index | tf.StringMapping)
            ? typeOperands()?.[0]
            : undefined;
        },
        get objectType() {
          return flag(tf.IndexedAccess) ? typeOperands()?.[0] : undefined;
        },
        get indexType() {
          return flag(tf.IndexedAccess) ? typeOperands()?.[1] : undefined;
        },
        get checkType() {
          return flag(tf.Conditional) ? typeOperands()?.[0] : undefined;
        },
        get extendsType() {
          return flag(tf.Conditional) ? typeOperands()?.[1] : undefined;
        },
        get pattern() {
          const d = details();
          return client.syntax.nodeRef(d.pattern);
        },
        get node() {
          const d = details();
          return client.syntax.nodeRef(d.node);
        },
        get freshType() {
          return client.type(freshId());
        },
        get elementType() {
          return client.type(
            client.query({ op: 'evolvingElement', args: { id } }),
          );
        },
        get finalArrayType() {
          return client.type(client.query({ op: 'finalArray', args: { id } }));
        },
        get simplifiedForReading() {
          return client.type(
            client.query({ op: 'simplifiedForReading', args: { id } }),
          );
        },
        get simplifiedForWriting() {
          return client.type(
            client.query({ op: 'simplifiedForWriting', args: { id } }),
          );
        },
        get regularType() {
          return client.type(regularId());
        },
        get baseType() {
          const d = details();
          return client.type(d.base);
        },
        get constraint() {
          if (flag(tf.TypeParameter))
            return client.type(
              client.query({ op: 'typeParameterConstraint', args: { id } }),
            );
          return client.type(details().constraint);
        },
        get outerTypeParameters() {
          const d = details();
          return typeList(d.outer);
        },
        get localTypeParameters() {
          const d = details();
          return typeList(d.local);
        },
        get elementFlags() {
          const d = details();
          return d.elements && Object.freeze(d.elements);
        },
        get minLength() {
          const d = details();
          return d.minLength;
        },
        get fixedLength() {
          const d = details();
          return d.fixedLength;
        },
        get hasRestElement() {
          const d = details();
          return d.elements?.some((flags) => (flags & 12) !== 0);
        },
        get labeledElementDeclarations() {
          const d = details();
          return (
            d.labels &&
            Object.freeze(d.labels.map((ref) => client.syntax.nodeRef(ref)))
          );
        },
        get texts() {
          const d = details();
          return d.texts && Object.freeze(d.texts);
        },
        get root() {
          return conditionalRoot();
        },
        get resolvedTrueType() {
          return client.type(
            client.query({ op: 'resolvedTrueType', args: { id } }),
          );
        },
        get resolvedFalseType() {
          return client.type(
            client.query({ op: 'resolvedFalseType', args: { id } }),
          );
        },
        get target() {
          return target();
        },
        get aliasTypeArguments() {
          return aliasTypeArguments();
        },
        get typeParameters() {
          return typeParameters();
        },
        get typeParameter() {
          return mapped('typeParameter');
        },
        get constraintType() {
          return mapped('constraintType');
        },
        get templateType() {
          return mapped('templateType');
        },
        get modifiersType() {
          return mapped('modifiersType');
        },
        get thisType() {
          return objectFlag(of.ClassOrInterface | of.Tuple) &&
            target()?.id === id
            ? client.type(client.query({ op: 'classThisType', args: { id } }))
            : undefined;
        },
        get combinedFlags() {
          return tuple()?.combinedFlags;
        },
        get readonly() {
          return tuple()?.readonly;
        },
      };
      inheritTypePrototype(fields, client.typePrototype);
      return this.wrapType(fields, hasField, mappedFields, () => !!declared());
    };
    return this.intern(this.types, 'type', id, create);
  }
  private symbol(id: number, knownName?: string): SymbolView | undefined {
    const view = this.intern(this.symbols, 'symbol', id, (client, brand) => {
      const ensure = () => {
        this.ensure();
      };
      const readTable = (
        flag: 'hasExports' | 'hasGlobalExports',
        op: 'symbolExports' | 'symbolGlobalExports',
      ) =>
        memo(ensure, () =>
          this.symbolField(id, flag)
            ? new Map(
                this.query({ op, args: { id } }).map((member) => [
                  member.name,
                  this.symbol(member.id)!,
                ]),
              )
            : undefined,
        );
      const exportTable = readTable('hasExports', 'symbolExports');
      const globalExportTable = readTable(
        'hasGlobalExports',
        'symbolGlobalExports',
      );
      const name = memo(
          ensure,
          () => brand.knownName ?? this.symbolField(id, 'name'),
        ),
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
      const withChecker = (checker: CheckerView | undefined): boolean => {
        if (checker !== undefined && checker !== this.checker)
          throw new Error('foreign type service checker');
        return checker !== undefined;
      };
      let documentation: ts.SymbolDisplayPart[] | undefined;
      let tags: ts.JSDocTagInfo[] | undefined;
      let links: SymbolView['links'];
      let members: Map<string, SymbolView> | undefined;
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
        get exports() {
          return exportTable();
        },
        get globalExports() {
          return globalExportTable();
        },
        get members() {
          if (members || !client.symbolField(id, 'hasMembers')) return members;
          return (members = new Map(
            client
              .query({ op: 'symbolMembers', args: { id } })
              .map((member) => [member.name, client.symbol(member.id)!]),
          ));
        },
        get links() {
          if (links) return links;
          const flags = client.symbolField(id, 'checkFlags');
          if (flags === undefined) return undefined;
          return (links = {
            get checkFlags() {
              ensure();
              return flags;
            },
            get type() {
              return client.type(
                client.query({ op: 'symbolLinkType', args: { id } }),
              );
            },
          });
        },
        getName: name,
        getEscapedName: escaped,
        getFlags: flags,
        getDeclarations: declarations,
        getDocumentationComment: (checker) => {
          const useChecker = withChecker(checker);
          return (documentation ??= this.query({
            op: 'symbolDocumentation',
            args: { id, withChecker: useChecker },
          }));
        },
        getJsDocTags: (checker) => {
          const useChecker = withChecker(checker);
          return (tags ??= this.query({
            op: 'symbolJSDocTags',
            args: { id, withChecker: useChecker },
          }));
        },
      };
    });
    if (view && knownName !== undefined)
      this.brands.get(view)!.knownName = knownName;
    return view;
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
        typeParameters = memo(ensure, () => {
          const ids = this.signatureField(id, 'typeParameters');
          return ids ? this.typesFor(ids) : undefined;
        }),
        thisParameter = memo(ensure, () =>
          this.symbol(this.signatureField(id, 'thisParameter')),
        ),
        minArgumentCount = memo(ensure, () =>
          this.signatureField(id, 'minArgumentCount'),
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
        get typeParameters() {
          return typeParameters();
        },
        get thisParameter() {
          return thisParameter();
        },
        get minArgumentCount() {
          return minArgumentCount();
        },
        getParameters: parameters,
        getTypeParameters: typeParameters,
        getTypeParameterAtPosition: (position) =>
          this.type(
            this.query({ op: 'parameterType', args: { id, position } }),
          ),
        getDeclaration: declaration,
        getReturnType: returned,
        getDocumentationComment: memo(ensure, () =>
          this.query({ op: 'signatureDocumentation', args: { id } }),
        ),
        getJsDocTags: memo(ensure, () =>
          this.query({ op: 'signatureJSDocTags', args: { id } }),
        ),
      };
    });
  }
  file(binding: FileBinding): TypeServiceFile {
    this.ensure();
    if (
      binding.generation !== this.descriptor.generation ||
      binding.checker !== this.descriptor.checker
    )
      throw new Error('foreign file binding');
    const file = new TypeServiceFile(
      this,
      binding,
      new FileSyntaxViews(this.syntax, binding, () => {
        this.ensure();
      }),
    );
    this.files.add(file);
    return file;
  }
  releaseFile(file: TypeServiceFile): void {
    this.files.delete(file);
  }
}

type NodeMaps = Pick<
  FileSyntaxViews,
  'esTreeNodeToTSNodeMap' | 'tsNodeToESTreeNodeMap'
>;

export function syntaxOnlyParserServices(maps: NodeMaps) {
  return {
    ...maps,
    program: null,
    emitDecoratorMetadata: undefined,
    experimentalDecorators: undefined,
    isolatedDeclarations: undefined,
  };
}

export class TypeServiceFile {
  private closed = false;
  constructor(
    private readonly owner: TypeServiceClient,
    readonly binding: FileBinding,
    private readonly syntax: FileSyntaxViews,
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
    if (this.owner.descriptor.syntaxOnly)
      return syntaxOnlyParserServices({
        esTreeNodeToTSNodeMap: this.syntax.esTreeNodeToTSNodeMap,
        tsNodeToESTreeNodeMap: this.syntax.tsNodeToESTreeNodeMap,
      });
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
      getContextualType: (node: ESTreeNode) =>
        this.owner.checker.getContextualType(toTS(node)),
      getResolvedSignature: (node: ESTreeNode) =>
        this.owner.checker.getResolvedSignature(toTS(node)),
      getTypeAtLocation: (node: ESTreeNode) =>
        this.owner.checker.getTypeAtLocation(toTS(node)),
      getSymbolAtLocation: (node: ESTreeNode) =>
        this.owner.checker.getSymbolAtLocation(toTS(node)),
      getTypeFromTypeNode: (node: ESTreeNode) =>
        this.owner.checker.getTypeFromTypeNode(toTS(node)),
      getTypeOfSymbolAtLocation: (symbol: SymbolView, node: ESTreeNode) =>
        this.owner.checker.getTypeOfSymbolAtLocation(symbol, toTS(node)),
      emitDecoratorMetadata: options.emitDecoratorMetadata === true,
      experimentalDecorators: options.experimentalDecorators === true,
      isolatedDeclarations: options.isolatedDeclarations === true,
    };
  }
}
