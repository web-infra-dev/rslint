import type ts from 'typescript';
import type { TypeServiceCompilerData } from './type-service-compiler-data.js';
import type {
  DiagnosticFact,
  IndexInfoFact,
  NodeRef,
  QueryResults,
  ScalarRequest,
  SignatureFact,
  SymbolFact,
  TypeFact,
} from './type-service-protocol.js';

export interface ConditionalRootView {
  readonly node: object;
  readonly checkType: TypeView;
  readonly extendsType: TypeView;
  readonly isDistributive: boolean;
  readonly inferTypeParameters?: readonly TypeView[];
  readonly outerTypeParameters?: readonly TypeView[];
  readonly aliasSymbol?: SymbolView;
  readonly aliasTypeArguments?: readonly TypeView[];
}

export interface TypeView {
  readonly checker: CheckerView;
  readonly id: number;
  readonly flags: number;
  readonly symbol?: SymbolView;
  readonly aliasSymbol?: SymbolView;
  readonly intrinsicName?: string;
  readonly escapedName?: string;
  readonly objectFlags?: number;
  readonly value?: TypeFact['value'];
  readonly types?: readonly TypeView[];
  readonly texts?: readonly string[];
  readonly type?: TypeView;
  readonly objectType?: TypeView;
  readonly indexType?: TypeView;
  readonly checkType?: TypeView;
  readonly extendsType?: TypeView;
  readonly root?: ConditionalRootView;
  readonly target?: TypeView;
  readonly typeArguments?: readonly TypeView[];
  readonly aliasTypeArguments?: readonly TypeView[];
  readonly typeParameters?: readonly TypeView[];
  readonly outerTypeParameters?: readonly TypeView[];
  readonly localTypeParameters?: readonly TypeView[];
  readonly typeParameter?: TypeView;
  readonly constraintType?: TypeView;
  readonly templateType?: TypeView;
  readonly modifiersType?: TypeView;
  readonly thisType?: TypeView;
  readonly declaredIndexInfos?: readonly IndexInfoView[];
  readonly combinedFlags?: number;
  readonly readonly?: boolean;
  readonly elementFlags?: readonly number[];
  readonly minLength?: number;
  readonly fixedLength?: number;
  readonly hasRestElement?: boolean;
  readonly labeledElementDeclarations?: readonly (object | undefined)[];
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
  readonly members: ReadonlyMap<string, SymbolView> | undefined;
  readonly links:
    | { readonly checkFlags: number; readonly type: TypeView | undefined }
    | undefined;
  getName(): string;
  getEscapedName(): string;
  getFlags(): number;
  getDeclarations(): readonly object[] | undefined;
  getDocumentationComment(
    checker?: CheckerView,
  ): readonly ts.SymbolDisplayPart[];
  getJsDocTags(checker?: CheckerView): readonly ts.JSDocTagInfo[];
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
  getDocumentationComment(
    checker?: CheckerView,
  ): readonly ts.SymbolDisplayPart[];
  getJsDocTags(): readonly ts.JSDocTagInfo[];
}

export interface IndexInfoView {
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

export interface DiagnosticView extends Omit<
  ts.Diagnostic,
  'file' | 'relatedInformation'
> {
  readonly file: object | undefined;
  readonly relatedInformation?: readonly DiagnosticView[];
}

export interface CheckerView {
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
  getVoidType(): TypeView;
  getStringLiteralType(value: string): TypeView;
  getNumberLiteralType(value: number): TypeView;
  getBigIntLiteralType(value: Readonly<ts.PseudoBigInt>): TypeView;
  getNullableType(type: object, flags: number): TypeView;
  getAmbientModules(): readonly SymbolView[];
  getAugmentedPropertiesOfType(t: object): readonly SymbolView[];
  getRootSymbols(s: object): readonly SymbolView[];
  getMergedSymbol(s: object): SymbolView | undefined;
  isArgumentsSymbol(s: object): boolean;
  isUndefinedSymbol(s: object): boolean;
  tryGetMemberInModuleExports(
    name: string,
    moduleSymbol: object,
  ): SymbolView | undefined;
  isValidPropertyAccess(n: object, name: string): boolean;
  getExportSpecifierLocalTargetSymbol(n: object): SymbolView | undefined;
  getPropertySymbolOfDestructuringAssignment(n: object): SymbolView | undefined;
  getJsxIntrinsicTagNamesAt(n: object): readonly SymbolView[];
  getConstantValue(n: object): string | number | undefined;
  isOptionalParameter(n: object): boolean;
  isImplementationOfOverload(n: object): boolean | undefined;
  getSymbolsOfParameterPropertyDeclaration(
    parameter: object,
    parameterName: string,
  ): readonly SymbolView[];
  runWithCancellationToken<T>(
    token: unknown,
    cb: (checker: CheckerView) => T,
  ): T;
  getTypeAtLocation(n: object | undefined): TypeView | undefined;
  getTypeFromTypeNode(n: object | undefined): TypeView | undefined;
  getContextualType(n: object | undefined): TypeView | undefined;
  getContextualTypeForArgumentAtIndex(
    n: object,
    index: number,
  ): TypeView | undefined;
  getFullyQualifiedName(symbol: object): string;
  symbolToString(
    s: object,
    enclosing?: object,
    meaning?: number,
    flags?: number,
  ): string;
  typeToTypeNode(
    type: object,
    enclosing?: object,
    flags?: number,
  ): object | undefined;
  getSuggestionDiagnostics(source: object): readonly DiagnosticView[];
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
  typePredicateToString(
    predicate: Readonly<TypePredicateView>,
    enclosing?: object,
    flags?: number,
  ): string;
  signatureToString(
    s: object,
    enclosing?: object,
    flags?: number,
    kind?: number,
  ): string;
  getSymbolAtLocation(n: object | undefined): SymbolView | undefined;
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
  getTypeOfSymbolAtLocation(
    s: object,
    n: object | undefined,
  ): TypeView | undefined;
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
    n: object | undefined,
    candidatesOutArray?: SignatureView[],
    argumentCount?: number,
  ): SignatureView | undefined;
  getReturnTypeOfSignature(s: object): TypeView | undefined;
  getTypePredicateOfSignature(s: object): TypePredicateView | undefined;
  getSignaturesOfType(t: object, kind: number): readonly SignatureView[];
}

export interface ProgramView {
  getConfigFileParsingDiagnostics(): readonly DiagnosticView[];
  getOptionsDiagnostics(): readonly DiagnosticView[];
  getGlobalDiagnostics(): readonly DiagnosticView[];
  getSyntacticDiagnostics(source?: object): readonly DiagnosticView[];
  getSemanticDiagnostics(source?: object): readonly DiagnosticView[];
  getDeclarationDiagnostics(source?: object): readonly DiagnosticView[];
  getRootFileNames(): readonly string[];
  emit(): never;
  getTypeChecker(): CheckerView;
  getCompilerOptions(): Readonly<Record<string, unknown>>;
  getCurrentDirectory(): string;
  getSourceFileByPath(path: string): object | undefined;
  getSourceFile(name: string): object | undefined;
  getSourceFiles(): readonly object[];
  readonly sourceFileToPackageName: {
    get(path: string): string | undefined;
  };
  isSourceFileFromExternalLibrary(source: object): boolean;
  isSourceFileDefaultLibrary(source: object): boolean;
}

export interface ViewHost {
  readonly typescript: TypeServiceCompilerData;
  readonly checker: CheckerView;
  ensure(): void;
  withReplyRecords<T>(run: () => T): T;
  query<Q extends ScalarRequest>(request: Q): QueryResults[Q['op']];
  typeFact(id: number): TypeFact;
  symbolFact(id: number): SymbolFact;
  signatureFact(id: number): SignatureFact;
  type(id: number): TypeView | undefined;
  symbol(id: number, knownName?: string): SymbolView | undefined;
  signature(id: number): SignatureView | undefined;
  typesFor(ids: readonly number[]): readonly TypeView[];
  symbolsFor(ids: readonly number[]): readonly SymbolView[];
  indexInfo(fact: IndexInfoFact): IndexInfoView;
  nodeRef(ref: NodeRef | undefined): object | undefined;
  sourceView(file: string): object | undefined;
  diagnosticSourceView(file: string, filePath?: string): object | undefined;
  conditionalRootFor(
    node: object,
    create: () => ConditionalRootView,
  ): ConditionalRootView;
}

export function memo<T>(ensure: () => void, read: () => T): () => T {
  let cached: { value: T } | undefined;
  return () => {
    ensure();
    return (cached ??= { value: read() }).value;
  };
}

function unsupportedFlag(name: string): never {
  throw new Error(`unsupported TypeScript operation: ${name}`);
}

function defineReadonlyGetter(
  target: object,
  key: string,
  get: () => unknown,
): void {
  Object.defineProperty(target, key, {
    get,
    enumerable: true,
    configurable: false,
  });
}

function defineReadonlyValue(
  target: object,
  key: string,
  value: unknown,
): void {
  Object.defineProperty(target, key, {
    value,
    writable: false,
    enumerable: true,
    configurable: false,
  });
}

export type TypeFactory = (id: number) => TypeView;

export function makeTypeFactory(
  checker: Readonly<CheckerView>,
  typescript: TypeServiceCompilerData,
  ensure: () => void,
): TypeFactory {
  const { TypeFlags: tf, ObjectFlags: of } = typescript;
  function cached<T>(read: (type: TypeView) => T) {
    const values = new WeakMap<TypeView, { readonly value: T }>();
    return function (this: TypeView): T {
      ensure();
      const entry = values.get(this);
      if (entry) return entry.value;
      const value = read(this);
      values.set(this, { value });
      return value;
    };
  }
  const typeArguments = cached((t) =>
    (t.objectFlags ?? 0) & of.Reference
      ? Object.freeze(checker.getTypeArguments(t))
      : undefined,
  );
  const getCallSignatures = cached((t) => checker.getSignaturesOfType(t, 0));
  const getConstructSignatures = cached((t) =>
    checker.getSignaturesOfType(t, 1),
  );
  const getProperties = cached((t) => checker.getPropertiesOfType(t));
  const getApparentProperties = cached((t) =>
    checker.getAugmentedPropertiesOfType(t),
  );
  const getNonNullableType = cached((t) => checker.getNonNullableType(t));
  const getConstraint = cached((t) => checker.getBaseConstraintOfType(t));
  const getDefault = cached((t) =>
    t.isTypeParameter() ? checker.getDefaultFromTypeParameter(t) : undefined,
  );
  const getBaseTypes = cached((t) =>
    t.isClassOrInterface() ? checker.getBaseTypes(t) : undefined,
  );
  const getStringIndexType = cached((t) => checker.getIndexTypeOfType(t, 0));
  const getNumberIndexType = cached((t) => checker.getIndexTypeOfType(t, 1));

  class TypeViewImpl implements TypeView {
    declare readonly id: number;
    declare readonly flags: number;
    declare readonly checker: CheckerView;
    declare readonly symbol?: SymbolView;
    declare readonly objectFlags?: number;

    constructor(id: number) {
      defineReadonlyValue(this, 'id', id);
    }
    get typeArguments(): readonly TypeView[] | undefined {
      return typeArguments.call(this);
    }
    getFlags() {
      return this.flags;
    }
    getSymbol() {
      return this.symbol;
    }
    isUnion() {
      return !!(this.flags & tf.Union);
    }
    isIntersection() {
      return !!(this.flags & tf.Intersection);
    }
    isUnionOrIntersection() {
      return !!(this.flags & tf.UnionOrIntersection);
    }
    isLiteral() {
      return !!(
        this.flags &
        (tf.StringLiteral | tf.NumberLiteral | tf.BigIntLiteral)
      );
    }
    isStringLiteral() {
      return !!(this.flags & tf.StringLiteral);
    }
    isNumberLiteral() {
      return !!(this.flags & tf.NumberLiteral);
    }
    isTypeParameter() {
      return !!(this.flags & tf.TypeParameter);
    }
    isClassOrInterface() {
      return !!((this.objectFlags ?? 0) & of.ClassOrInterface);
    }
    isClass() {
      return !!((this.objectFlags ?? 0) & of.Class);
    }
    isIndexType() {
      return !!(this.flags & tf.Index);
    }
    getCallSignatures() {
      return getCallSignatures.call(this);
    }
    getConstructSignatures() {
      return getConstructSignatures.call(this);
    }
    getProperties() {
      return getProperties.call(this);
    }
    getApparentProperties() {
      return getApparentProperties.call(this);
    }
    getProperty(name: string) {
      return checker.getPropertyOfType(this, name);
    }
    getNonNullableType() {
      return getNonNullableType.call(this);
    }
    getConstraint() {
      return getConstraint.call(this);
    }
    getDefault() {
      return getDefault.call(this);
    }
    getBaseTypes() {
      return getBaseTypes.call(this);
    }
    getStringIndexType() {
      return getStringIndexType.call(this);
    }
    getNumberIndexType() {
      return getNumberIndexType.call(this);
    }
  }

  return (id: number) => new TypeViewImpl(id);
}

const mappedFieldIndex = {
  typeParameter: 0,
  constraintType: 1,
  templateType: 2,
  modifiersType: 3,
} as const;

/**
 * Constructs a plain `TypeView` object with kind-appropriate own getters.
 * Avoids JS Proxy reflection traps while preserving accurate `in` checks
 * (such as `'typeParameter' in type` and `'types' in type`) and read-only
 * compiler fields.
 */
export function createTypeView(
  host: Readonly<ViewHost>,
  createBase: TypeFactory,
  id: number,
): TypeView {
  const ensure = () => {
    host.ensure();
  };
  const tf = host.typescript.TypeFlags;
  const of = host.typescript.ObjectFlags;

  const fact = host.typeFact(id);
  const flagsValue = fact.flags.reduce((acc, name) => {
    if (name === 'None') return acc;
    const flag: unknown = Reflect.get(tf, name);
    if (typeof flag !== 'number') unsupportedFlag(`TypeFlags.${name}`);
    return acc | flag;
  }, 0);

  const objectFlagsValue = fact.objectFlags?.reduce((acc, name) => {
    const flag: unknown = Reflect.get(of, name);
    if (typeof flag !== 'number') unsupportedFlag(`ObjectFlags.${name}`);
    return acc | flag;
  }, 0);

  const symbolId = fact.symbol;
  const aliasId = fact.aliasSymbol;
  const targetId = fact.target;
  const intrinsicName = fact.intrinsicName;
  const escapedName = fact.escapedName;
  const literalRaw = fact.value;
  const literalValue =
    typeof literalRaw === 'object'
      ? Object.freeze({ ...literalRaw })
      : literalRaw;
  const tuple = fact.tuple;
  const typeOperandIds = fact.typeOperands;
  const aliasTypeArgumentIds = fact.aliasTypeArguments;
  const typeParameterIds = fact.typeParameters;

  const hasFlag = (mask: number) => (flagsValue & mask) !== 0;
  const hasObjectFlag = (mask: number) =>
    ((objectFlagsValue ?? 0) & mask) !== 0;

  const typeList = (ids: readonly number[] | undefined) =>
    ids ? Object.freeze(host.typesFor(ids)) : undefined;

  const details = memo(ensure, () =>
    host.query({ op: 'typeDetails', args: { id } }),
  );
  const symbol = memo(ensure, () => host.symbol(symbolId));
  const aliasSymbol = memo(ensure, () => host.symbol(aliasId));
  const target = memo(ensure, () => host.type(targetId));
  const typeOperands = memo(ensure, () => typeList(typeOperandIds));
  const aliasTypeArguments = memo(ensure, () => typeList(aliasTypeArgumentIds));
  const typeParameters = memo(ensure, () => typeList(typeParameterIds));
  const conditionalRoot = memo(ensure, () => {
    const d = details();
    const node = host.nodeRef(d.root);
    if (!node) return undefined;
    return host.conditionalRootFor(node, () => {
      const checkType = host.type(d.rootCheck);
      const extendsType = host.type(d.rootExtends);
      if (!checkType || !extendsType) {
        throw new Error('incomplete conditional root type');
      }
      return Object.freeze({
        node,
        checkType,
        extendsType,
        isDistributive: d.distributive,
        inferTypeParameters: typeList(d.infer),
        outerTypeParameters: typeList(d.rootOuter),
        aliasSymbol: host.symbol(d.rootAlias),
        aliasTypeArguments: typeList(d.rootArguments),
      });
    });
  });

  const view = createBase(id);
  defineReadonlyGetter(view, 'checker', () => {
    ensure();
    return host.checker;
  });
  defineReadonlyGetter(view, 'flags', () => {
    ensure();
    return flagsValue;
  });

  if (objectFlagsValue !== undefined) {
    defineReadonlyGetter(view, 'objectFlags', () => {
      ensure();
      return objectFlagsValue;
    });
  }
  if (
    hasFlag(
      tf.Object |
        tf.TypeParameter |
        tf.StringLiteral |
        tf.NumberLiteral |
        tf.BigIntLiteral |
        tf.StringMapping,
    ) ||
    symbolId !== 0
  ) {
    defineReadonlyGetter(view, 'symbol', symbol);
  }
  if (
    aliasId !== 0 ||
    hasFlag(tf.UnionOrIntersection | tf.Conditional | tf.IndexedAccess) ||
    hasObjectFlag(of.Mapped | of.Instantiated)
  ) {
    defineReadonlyGetter(view, 'aliasSymbol', aliasSymbol);
    defineReadonlyGetter(view, 'aliasTypeArguments', aliasTypeArguments);
  }
  if (intrinsicName !== undefined) {
    defineReadonlyGetter(view, 'intrinsicName', () => {
      ensure();
      return intrinsicName;
    });
  }
  if (escapedName !== undefined) {
    defineReadonlyGetter(view, 'escapedName', () => {
      ensure();
      return escapedName;
    });
  }
  if (literalValue !== undefined) {
    defineReadonlyGetter(view, 'value', () => {
      ensure();
      return literalValue;
    });
  }
  if (hasFlag(tf.UnionOrIntersection | tf.TemplateLiteral)) {
    defineReadonlyGetter(view, 'types', typeOperands);
  }
  if (hasFlag(tf.TemplateLiteral)) {
    defineReadonlyGetter(view, 'texts', () => {
      const texts = details().texts;
      return texts && Object.freeze(texts);
    });
  }
  if (hasFlag(tf.Index | tf.StringMapping)) {
    defineReadonlyGetter(view, 'type', () => typeOperands()?.[0]);
  }
  if (hasFlag(tf.IndexedAccess)) {
    defineReadonlyGetter(view, 'objectType', () => typeOperands()?.[0]);
    defineReadonlyGetter(view, 'indexType', () => typeOperands()?.[1]);
  }
  if (hasFlag(tf.Conditional)) {
    defineReadonlyGetter(view, 'checkType', () => typeOperands()?.[0]);
    defineReadonlyGetter(view, 'extendsType', () => typeOperands()?.[1]);
    defineReadonlyGetter(view, 'root', conditionalRoot);
  }
  if (targetId !== 0) {
    defineReadonlyGetter(view, 'target', target);
  }
  if (
    hasObjectFlag(of.ClassOrInterface | of.Tuple) &&
    (targetId === 0 || targetId === id)
  ) {
    const thisType = memo(ensure, () =>
      host.withReplyRecords(() =>
        host.type(host.query({ op: 'classThisType', args: { id } })),
      ),
    );
    defineReadonlyGetter(view, 'typeParameters', typeParameters);
    defineReadonlyGetter(view, 'outerTypeParameters', () =>
      typeList(details().outer),
    );
    defineReadonlyGetter(view, 'localTypeParameters', () =>
      typeList(details().local),
    );
    defineReadonlyGetter(view, 'thisType', thisType);
    defineReadonlyGetter(view, 'declaredIndexInfos', () =>
      host.withReplyRecords(() => {
        const infos = host.query({ op: 'declaredIndexInfos', args: { id } });
        return infos
          ? Object.freeze(infos.map((info) => host.indexInfo(info)))
          : undefined;
      }),
    );
  }
  if (tuple !== undefined) {
    defineReadonlyGetter(view, 'combinedFlags', () => {
      ensure();
      return tuple.combinedFlags;
    });
    defineReadonlyGetter(view, 'readonly', () => {
      ensure();
      return tuple.readonly;
    });
    defineReadonlyGetter(view, 'elementFlags', () => {
      const elements = details().elements;
      return elements && Object.freeze(elements);
    });
    defineReadonlyGetter(view, 'minLength', () => details().minLength);
    defineReadonlyGetter(view, 'fixedLength', () => details().fixedLength);
    defineReadonlyGetter(view, 'hasRestElement', () =>
      details().elements?.some((flags) => (flags & 12) !== 0),
    );
    defineReadonlyGetter(view, 'labeledElementDeclarations', () => {
      const labels = details().labels;
      return labels && Object.freeze(labels.map((ref) => host.nodeRef(ref)));
    });
  }
  if (hasObjectFlag(of.Mapped)) {
    const readMapped = (key: keyof typeof mappedFieldIndex) =>
      host.withReplyRecords(() => {
        const ids = host.query({ op: 'mappedFields', args: { id } });
        return host.type(ids[mappedFieldIndex[key]] ?? 0);
      });
    defineReadonlyGetter(view, 'typeParameter', () =>
      readMapped('typeParameter'),
    );
    defineReadonlyGetter(view, 'constraintType', () =>
      readMapped('constraintType'),
    );
    defineReadonlyGetter(view, 'templateType', () =>
      readMapped('templateType'),
    );
    defineReadonlyGetter(view, 'modifiersType', () =>
      readMapped('modifiersType'),
    );
  }
  return view;
}

export function createSymbolView(
  host: Readonly<ViewHost>,
  id: number,
  knownName: () => string | undefined,
): SymbolView {
  const ensure = () => {
    host.ensure();
  };
  const fact = memo(ensure, () => host.symbolFact(id));
  const name = memo(ensure, () => knownName() ?? fact().name);
  const escapedName = memo(ensure, () => fact().escapedName);
  const flags = memo(ensure, () => fact().flags);
  const declarations = memo(ensure, () => {
    const refs = fact().declarations;
    if (!refs) return undefined;
    const nodes: object[] = [];
    for (const ref of refs) {
      const node = host.nodeRef(ref);
      if (node) nodes.push(node);
    }
    return Object.freeze(nodes);
  });
  const valueDeclaration = memo(ensure, () =>
    host.nodeRef(fact().valueDeclaration),
  );
  const readTable = (
    op: 'symbolMembers' | 'symbolExports',
  ): ReadonlyMap<string, SymbolView> | undefined => {
    const entries = host.query({ op, args: { id } });
    if (!entries) return undefined;
    const map = new Map<string, SymbolView>();
    for (const entry of entries) {
      const sym = host.symbol(entry.id);
      if (sym) map.set(entry.name, sym);
    }
    return map;
  };
  const checkWithChecker = (checker: CheckerView | undefined): boolean => {
    ensure();
    if (checker !== undefined && checker !== host.checker) {
      throw new Error('foreign type service checker');
    }
    return checker !== undefined;
  };

  // Cache separately by `withChecker` so calls with/without checker never poison
  // each other, and freeze returned arrays so rules cannot mutate cached results.
  let docWithChecker: readonly ts.SymbolDisplayPart[] | undefined;
  let docWithoutChecker: readonly ts.SymbolDisplayPart[] | undefined;
  let tagsWithChecker: readonly ts.JSDocTagInfo[] | undefined;
  let tagsWithoutChecker: readonly ts.JSDocTagInfo[] | undefined;
  let links: SymbolView['links'];

  return {
    id,
    get name() {
      return name();
    },
    get escapedName() {
      return escapedName();
    },
    get flags() {
      return flags();
    },
    get declarations() {
      return declarations();
    },
    get valueDeclaration() {
      return valueDeclaration();
    },
    get exports() {
      return readTable('symbolExports');
    },
    get members() {
      return readTable('symbolMembers');
    },
    get links() {
      ensure();
      if (links) return links;
      const checkFlags = fact().checkFlags;
      if (checkFlags === undefined) return undefined;
      return (links = {
        get checkFlags() {
          ensure();
          return checkFlags;
        },
        get type() {
          return host.withReplyRecords(() =>
            host.type(host.query({ op: 'typeOfSymbol', args: { id } })),
          );
        },
      });
    },
    getName: name,
    getEscapedName: escapedName,
    getFlags: flags,
    getDeclarations: declarations,
    getDocumentationComment: (checker) => {
      const useChecker = checkWithChecker(checker);
      if (useChecker) {
        return (docWithChecker ??= Object.freeze(
          host.query({
            op: 'symbolDocumentation',
            args: { id, withChecker: true },
          }),
        ));
      }
      return (docWithoutChecker ??= Object.freeze(
        host.query({
          op: 'symbolDocumentation',
          args: { id, withChecker: false },
        }),
      ));
    },
    getJsDocTags: (checker) => {
      const useChecker = checkWithChecker(checker);
      if (useChecker) {
        return (tagsWithChecker ??= Object.freeze(
          host.query({
            op: 'symbolJSDocTags',
            args: { id, withChecker: true },
          }),
        ));
      }
      return (tagsWithoutChecker ??= Object.freeze(
        host.query({
          op: 'symbolJSDocTags',
          args: { id, withChecker: false },
        }),
      ));
    },
  };
}

export function createSignatureView(
  host: Readonly<ViewHost>,
  id: number,
): SignatureView {
  const ensure = () => {
    host.ensure();
  };
  const fact = memo(ensure, () => host.signatureFact(id));
  const parameters = memo(ensure, () =>
    Object.freeze(host.symbolsFor(fact().parameters)),
  );
  const declaration = memo(ensure, () => host.nodeRef(fact().declaration));
  const typeParameters = memo(ensure, () => {
    const ids = fact().typeParameters;
    return ids ? Object.freeze(host.typesFor(ids)) : undefined;
  });
  const thisParameter = memo(ensure, () => host.symbol(fact().thisParameter));
  const minArgumentCount = memo(ensure, () => fact().minArgumentCount);
  const returnType = memo(ensure, () =>
    host.withReplyRecords(() =>
      host.type(host.query({ op: 'returnType', args: { id } })),
    ),
  );
  const documentation = memo(ensure, () =>
    Object.freeze(host.query({ op: 'signatureDocumentation', args: { id } })),
  );
  const jsDocTags = memo(ensure, () =>
    Object.freeze(host.query({ op: 'signatureJSDocTags', args: { id } })),
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
      host.withReplyRecords(() =>
        host.type(host.query({ op: 'parameterType', args: { id, position } })),
      ),
    getDeclaration: declaration,
    getReturnType: returnType,
    getDocumentationComment: documentation,
    getJsDocTags: jsDocTags,
  };
}

export function createIndexInfoView(
  host: Readonly<ViewHost>,
  fact: IndexInfoFact,
): IndexInfoView {
  const keyType = host.type(fact.keyType);
  const valueType = host.type(fact.type);
  if (!keyType || !valueType) {
    throw new Error('incomplete index info type');
  }
  return Object.freeze({
    keyType,
    type: valueType,
    isReadonly: fact.isReadonly,
    declaration: host.nodeRef(fact.declaration),
  });
}

function diagnosticMessage(root: DiagnosticFact): ts.DiagnosticMessageChain {
  const rootNode: ts.DiagnosticMessageChain = {
    code: root.code,
    category: root.category,
    messageText: root.text,
    next: undefined,
  };
  const stack: {
    readonly fact: DiagnosticFact;
    readonly node: ts.DiagnosticMessageChain;
  }[] = [{ fact: root, node: rootNode }];
  while (stack.length > 0) {
    const entry = stack.pop();
    if (!entry || entry.fact.chain.length === 0) continue;
    const next: ts.DiagnosticMessageChain[] = [];
    for (const child of entry.fact.chain) {
      const childNode: ts.DiagnosticMessageChain = {
        code: child.code,
        category: child.category,
        messageText: child.text,
        next: undefined,
      };
      next.push(childNode);
      if (child.chain.length > 0) {
        stack.push({ fact: child, node: childNode });
      }
    }
    entry.node.next = next;
  }
  return rootNode;
}

export function createDiagnosticView(
  host: Readonly<ViewHost>,
  root: DiagnosticFact,
): DiagnosticView {
  type MutableDiagnosticView = {
    -readonly [K in keyof DiagnosticView]: DiagnosticView[K];
  };
  const makeView = (d: DiagnosticFact): MutableDiagnosticView => ({
    file: d.fileName
      ? host.diagnosticSourceView(d.fileName, d.filePath)
      : undefined,
    start: d.fileName ? d.start : undefined,
    length: d.fileName ? d.length : undefined,
    code: d.code,
    category: d.category,
    source: d.source,
    messageText: d.chain.length ? diagnosticMessage(d) : d.text,
    reportsDeprecated: d.reportsDeprecated,
    reportsUnnecessary: d.reportsUnnecessary,
    relatedInformation: undefined,
  });
  const rootView = makeView(root);
  const stack: {
    readonly fact: DiagnosticFact;
    readonly view: MutableDiagnosticView;
  }[] = [{ fact: root, view: rootView }];
  while (stack.length > 0) {
    const entry = stack.pop();
    if (!entry || entry.fact.related.length === 0) continue;
    const related: MutableDiagnosticView[] = [];
    for (const item of entry.fact.related) {
      const itemView = makeView(item);
      related.push(itemView);
      if (item.related.length > 0) {
        stack.push({ fact: item, view: itemView });
      }
    }
    entry.view.relatedInformation = related;
  }
  return rootView;
}
