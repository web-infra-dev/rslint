import type { TypeServiceCompilerData } from './type-service-compiler-data.js';
import type { DecodedFrame } from './frame.js';
import type { ESTreeNode } from './types.js';
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
  type IndexInfoFact,
  type NamedSymbolFact,
  type QueryResults,
  type QueryTransport,
  type ScalarRequest,
  type ServiceRequest,
  type SignatureFact,
  type SymbolFact,
  type SyntaxSource,
  type TypeFact,
  type TypeServiceDescriptor,
  type TypeServiceProjection,
  type TypeServiceQuery,
} from './type-service-protocol.js';
import {
  createDiagnosticView,
  createIndexInfoView,
  createSignatureView,
  createSymbolView,
  createTypeView,
  makeTypeFactory,
  memo,
  type CheckerView,
  type ConditionalRootView,
  type DiagnosticView,
  type IndexInfoView,
  type ProgramView,
  type SignatureView,
  type SymbolView,
  type TypeFactory,
  type TypeView,
  type ViewHost,
} from './type-service-views.js';

export { typeQueryKey } from './type-service-protocol.js';
export type {
  TypeServiceDescriptor,
  TypeServiceProjection,
  TypeServiceQuery,
} from './type-service-protocol.js';
export type {
  CheckerView,
  ConditionalRootView,
  DiagnosticView,
  IndexInfoView,
  ProgramView,
  SignatureView,
  SymbolView,
  TypePredicateView,
  TypeView,
} from './type-service-views.js';

function unsupported(surface: string): never {
  throw new Error(`unsupported TypeScript operation: ${surface}`);
}

type ObjectKind = 'type' | 'symbol' | 'signature';
interface ObjectBrand {
  readonly kind: ObjectKind;
  readonly id: number;
  knownName?: string;
}

function readCompilerOptions(
  typescript: TypeServiceCompilerData,
  input: Readonly<Record<string, unknown>>,
): Readonly<Record<string, unknown>> {
  const compilerOptions: Record<string, unknown> = { ...input };
  for (const [key, group] of [
    ['jsx', typescript.JsxEmit],
    ['moduleDetection', typescript.ModuleDetectionKind],
  ] as const) {
    const value = compilerOptions[key];
    if (typeof value === 'string') {
      const translated: unknown = Reflect.get(group, value);
      if (typeof translated !== 'number') {
        unsupported(`compilerOptions.${key}.${value}`);
      }
      compilerOptions[key] = translated;
    }
  }
  return compilerOptions;
}

export interface TypeServiceClientOptions {
  readonly maxFactsBytes: number;
  readonly maxSyntaxBytes?: number;
  readonly maxSyntaxFactsBytes?: number;
}

interface CancellationTokenLike {
  readonly isCancellationRequested?: () => boolean;
  readonly throwIfCancellationRequested?: () => void;
}

/**
 * One worker owns this client for one Program generation. File handles own
 * ESTree mappings; canonical semantic and syntax identities survive file close.
 */
export class TypeServiceClient {
  private closed = false;
  private readonly facts = new Map<string, TypePacketReader>();
  private replyRecords: Map<string, Uint8Array> | undefined;
  private cancellationToken: CancellationTokenLike | undefined;
  private readonly types = new WeakObjectCache<number, TypeView>();
  private readonly symbols = new WeakObjectCache<number, SymbolView>();
  private readonly signatures = new WeakObjectCache<number, SignatureView>();
  private readonly indexInfos = new WeakObjectCache<number, IndexInfoView>();
  private readonly brands = new WeakMap<object, ObjectBrand>();
  private readonly conditionalRoots = new WeakMap<
    object,
    ConditionalRootView
  >();
  private readonly maxFactsBytes: number;
  private readonly syntax: GenerationSyntaxViews;
  private readonly typeFactory: TypeFactory;
  private readonly viewHost: ViewHost;
  private readonly files = new Set<TypeServiceFile>();
  private factBytes = 0;
  readonly program: ProgramView;
  readonly checker: CheckerView;

  constructor(
    private readonly typescript: TypeServiceCompilerData,
    readonly descriptor: TypeServiceDescriptor,
    private transport: QueryTransport | undefined,
    syntax: SyntaxSource,
    options: TypeServiceClientOptions,
  ) {
    const maxSyntaxBytes = options.maxSyntaxBytes ?? 16 << 20;
    const maxSyntaxFactsBytes = options.maxSyntaxFactsBytes ?? 1 << 20;
    if (
      !Number.isSafeInteger(options.maxFactsBytes) ||
      options.maxFactsBytes < 0 ||
      !Number.isSafeInteger(maxSyntaxBytes) ||
      maxSyntaxBytes < 0
    ) {
      throw new Error('invalid type service client budget');
    }
    if (!Number.isSafeInteger(maxSyntaxFactsBytes) || maxSyntaxFactsBytes < 0) {
      throw new Error('invalid syntax fact budget');
    }
    this.maxFactsBytes = options.maxFactsBytes;
    const ensure = () => {
      this.ensure();
    };
    this.syntax = new GenerationSyntaxViews(
      typescript,
      descriptor,
      syntax,
      (node) =>
        this.read({ op: 'nodeRecord', args: { node } }, (packet) =>
          packet.nodeFacts(),
        ),
      ensure,
      maxSyntaxBytes,
      maxSyntaxFactsBytes,
      (node) =>
        this.symbol(this.query({ op: 'declarationSymbol', args: { node } })),
    );

    const errorType = () => this.query({ op: 'errorType', args: {} });
    const requiredType = (readId: () => number): TypeView =>
      this.withReplyRecords(() => {
        const t = this.type(readId());
        if (!t) throw new Error('missing expected compiler type');
        return t;
      });
    const nodeHandle = (
      op:
        | 'type'
        | 'fromTypeNode'
        | 'context'
        | 'symbol'
        | 'signature'
        | 'exportSpecifierLocalTargetSymbol'
        | 'propertySymbolOfDestructuringAssignment'
        | 'shorthandValueSymbol',
      n: object | undefined,
    ) => {
      const node = this.syntax.parseTreeLocation(n);
      if (node) return this.query({ op, args: { node } });
      return op === 'type' || op === 'fromTypeNode' ? errorType() : 0;
    };
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
    ) =>
      this.withReplyRecords(() =>
        this.type(this.query({ op, args: { id: this.id(t, 'type') } })),
      );

    this.checker = {
      getAnyType: () =>
        requiredType(() => this.query({ op: 'anyType', args: {} })),
      getUnknownType: () =>
        requiredType(() => this.query({ op: 'unknownType', args: {} })),
      getStringType: () =>
        requiredType(() => this.query({ op: 'stringType', args: {} })),
      getNumberType: () =>
        requiredType(() => this.query({ op: 'numberType', args: {} })),
      getBooleanType: () =>
        requiredType(() => this.query({ op: 'booleanType', args: {} })),
      getBigIntType: () =>
        requiredType(() => this.query({ op: 'bigIntType', args: {} })),
      getESSymbolType: () =>
        requiredType(() => this.query({ op: 'eSSymbolType', args: {} })),
      getNonPrimitiveType: () =>
        requiredType(() => this.query({ op: 'nonPrimitiveType', args: {} })),
      getNeverType: () =>
        requiredType(() => this.query({ op: 'neverType', args: {} })),
      getNullType: () =>
        requiredType(() => this.query({ op: 'nullType', args: {} })),
      getUndefinedType: () =>
        requiredType(() => this.query({ op: 'undefinedType', args: {} })),
      getTrueType: () =>
        requiredType(() => this.query({ op: 'trueType', args: {} })),
      getFalseType: () =>
        requiredType(() => this.query({ op: 'falseType', args: {} })),
      getVoidType: () =>
        requiredType(() => this.query({ op: 'void', args: {} })),
      getStringLiteralType: (value) =>
        requiredType(() =>
          this.query({
            op: 'stringLiteralType',
            args: { value: nameToUnits(value) },
          }),
        ),
      getNumberLiteralType: (value) =>
        requiredType(() =>
          this.query({
            op: 'numberLiteralType',
            args: Number.isNaN(value)
              ? { value: 0, nonFinite: 'NaN' }
              : value === Infinity
                ? { value: 0, nonFinite: 'Infinity' }
                : value === -Infinity
                  ? { value: 0, nonFinite: '-Infinity' }
                  : { value: Object.is(value, -0) ? 0 : value },
          }),
        ),
      getBigIntLiteralType: (value) => {
        const base10Value = value.base10Value || '0';
        return requiredType(() =>
          this.query({
            op: 'bigIntLiteralType',
            args: {
              negative: value.negative && base10Value !== '0',
              base10Value,
            },
          }),
        );
      },
      getNullableType: (type, flags) =>
        requiredType(() =>
          this.query({
            op: 'nullableType',
            args: { id: this.id(type, 'type'), flags: flags >>> 0 },
          }),
        ),
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
      isValidPropertyAccess: (n, name) => {
        const node = this.syntax.parseTreeLocation(n);
        return node
          ? this.query({
              op: 'validPropertyAccess',
              args: { node, name: nameToUnits(name) },
            })
          : false;
      },
      getExportSpecifierLocalTargetSymbol: (n) =>
        this.symbol(nodeHandle('exportSpecifierLocalTargetSymbol', n)),
      getPropertySymbolOfDestructuringAssignment: (n) =>
        this.symbol(nodeHandle('propertySymbolOfDestructuringAssignment', n)),
      getJsxIntrinsicTagNamesAt: (n) => {
        const node = this.syntax.sourceLocation(n);
        if (!node) {
          unsupported(
            'getJsxIntrinsicTagNamesAt on a node without a source file',
          );
        }
        return this.namedSymbolsFor(
          this.query({ op: 'jsxIntrinsicTagNamesAt', args: { node } }),
        );
      },
      getConstantValue: (n) => {
        const node = this.syntax.parseTreeLocation(n);
        return node
          ? this.query({ op: 'constantValue', args: { node } })
          : undefined;
      },
      isOptionalParameter: (n) => {
        const node = this.syntax.parseTreeLocation(n);
        return node
          ? this.query({ op: 'isOptionalParameter', args: { node } })
          : false;
      },
      isImplementationOfOverload: (n) => {
        const node = this.syntax.parseTreeLocation(n);
        return node
          ? this.query({ op: 'isImplementationOfOverload', args: { node } })
          : undefined;
      },
      getSymbolsOfParameterPropertyDeclaration: (parameter, parameterName) => {
        const node = this.syntax.parseTreeLocation(parameter);
        return node
          ? this.symbolsFor(
              this.query({
                op: 'parameterPropertySymbols',
                args: { node, name: nameToUnits(parameterName) },
              }),
            )
          : [];
      },
      runWithCancellationToken: (token, cb) => {
        ensure();
        const previous = this.cancellationToken;
        this.cancellationToken =
          typeof token === 'object' && token !== null
            ? (token as CancellationTokenLike)
            : undefined;
        try {
          return cb(this.checker);
        } finally {
          this.cancellationToken = previous;
        }
      },
      getTypeAtLocation: (n) =>
        this.withReplyRecords(() => this.type(nodeHandle('type', n))),
      getTypeFromTypeNode: (n) =>
        this.withReplyRecords(() => this.type(nodeHandle('fromTypeNode', n))),
      getContextualType: (n) =>
        this.withReplyRecords(() => this.type(nodeHandle('context', n))),
      getContextualTypeForArgumentAtIndex: (n, index) => {
        const node = this.syntax.parseTreeLocation(n);
        return node
          ? this.withReplyRecords(() =>
              this.type(
                this.query({ op: 'contextualArgument', args: { node, index } }),
              ),
            )
          : undefined;
      },
      getFullyQualifiedName: (symbol) =>
        this.query({
          op: 'fullyQualifiedName',
          args: { id: this.id(symbol, 'symbol') },
        }),
      symbolToString: (s, enclosing, meaning, flags) =>
        this.query({
          op: 'symbolString',
          args: {
            id: this.id(s, 'symbol'),
            kind: (meaning ?? 0) >>> 0,
            node: enclosing && this.syntax.parseTreeLocation(enclosing),
            flags: flags === undefined ? undefined : flags >>> 0,
          },
        }),
      typeToTypeNode: (t, enclosing, flags) => {
        const data = this.query({
          op: 'typeNode',
          args: {
            id: this.id(t, 'type'),
            node: enclosing && this.syntax.parseTreeLocation(enclosing),
            flags: flags === undefined ? undefined : flags >>> 0,
          },
        });
        return data && this.syntax.generatedTypeNode(data);
      },
      getSuggestionDiagnostics: (source) =>
        this.query({
          op: 'suggestionDiagnostics',
          args: { node: this.syntax.location(source) },
        }).map((d) => createDiagnosticView(this.viewHost, d)),
      getBaseConstraintOfType: (t) => typeType('constraint', t),
      getApparentType: (t) => typeType('apparent', t),
      getBaseTypeOfLiteralType: (t) => typeType('baseLiteral', t),
      getNonNullableType: (t) => typeType('nonNullable', t),
      getAwaitedType: (t) => typeType('awaited', t),
      getWidenedType: (t) => typeType('widened', t),
      getDefaultFromTypeParameter: (t) => typeType('defaultType', t),
      getBaseTypes: (t) =>
        this.withReplyRecords(() =>
          this.typesFor(
            this.query({ op: 'baseTypes', args: { id: this.id(t, 'type') } }),
          ),
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
            node: enclosing && this.syntax.parseTreeLocation(enclosing),
            flags: flags === undefined ? undefined : flags >>> 0,
          },
        }),
      typePredicateToString: (predicate, enclosing, flags) =>
        this.query({
          op: 'typePredicateString',
          args: {
            kind: predicate.kind >>> 0,
            parameterName:
              predicate.parameterName === undefined
                ? undefined
                : nameToUnits(predicate.parameterName),
            parameterIndex: predicate.parameterIndex,
            type: predicate.type ? this.id(predicate.type, 'type') : 0,
            node: enclosing && this.syntax.parseTreeLocation(enclosing),
            flags: flags === undefined ? undefined : flags >>> 0,
          },
        }),
      signatureToString: (s, enclosing, flags, kind = 0) =>
        this.query({
          op: 'signatureString',
          args: {
            id: this.id(s, 'signature'),
            node: enclosing && this.syntax.parseTreeLocation(enclosing),
            flags: flags === undefined ? undefined : flags >>> 0,
            kind,
          },
        }),
      getSymbolAtLocation: (n) => this.symbol(nodeHandle('symbol', n)),
      getSymbolsInScope: (n, meaning) => {
        const node = this.syntax.parseTreeLocation(n);
        return node
          ? this.namedSymbolsFor(
              this.query({
                op: 'symbolsInScope',
                args: { node, meaning: meaning | 0 },
              }),
            )
          : [];
      },
      resolveName: (name, location, kind, excludeGlobals) =>
        this.symbol(
          this.query({
            op: 'resolveName',
            args: {
              name: nameToUnits(name),
              node: location && this.syntax.sourceLocation(location),
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
        this.symbol(nodeHandle('shorthandValueSymbol', n)),
      getTypeOfSymbol: (s) =>
        this.withReplyRecords(() =>
          this.type(
            this.query({
              op: 'typeOfSymbol',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
        ),
      getTypeOfSymbolAtLocation: (s, n) => {
        const id = this.id(s, 'symbol');
        const node = this.syntax.parseTreeLocation(n);
        return this.withReplyRecords(() =>
          this.type(
            node
              ? this.query({ op: 'symbolType', args: { id, node } })
              : errorType(),
          ),
        );
      },
      getDeclaredTypeOfSymbol: (s) =>
        this.withReplyRecords(() =>
          this.type(
            this.query({
              op: 'declaredType',
              args: { id: this.id(s, 'symbol') },
            }),
          ),
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
        this.withReplyRecords(() =>
          this.type(
            this.query({
              op: 'propertyType',
              args: { id: this.id(t, 'type'), name: nameToUnits(name) },
            }),
          ),
        ),
      getIndexTypeOfType: (t, kind) =>
        typeType(kind === 0 ? 'stringIndexType' : 'numberIndexType', t),
      getIndexInfoOfType: (t, kind) =>
        this.withReplyRecords(() => {
          const [info] = this.query({
            op: 'indexInfo',
            args: { id: this.id(t, 'type'), kind },
          });
          return info && this.indexInfo(info);
        }),
      getIndexInfosOfType: (t) =>
        this.withReplyRecords(() =>
          this.query({
            op: 'indexInfos',
            args: { id: this.id(t, 'type') },
          }).map((info) => this.indexInfo(info)),
        ),
      getTypeArguments: (t) =>
        this.withReplyRecords(() =>
          this.typesFor(
            this.query({
              op: 'typeArguments',
              args: { id: this.id(t, 'type') },
            }),
          ),
        ),
      isArrayType: (t) =>
        this.query({ op: 'isArray', args: { id: this.id(t, 'type') } }),
      isArrayLikeType: (t) =>
        this.query({ op: 'isArrayLike', args: { id: this.id(t, 'type') } }),
      isTupleType: (t) =>
        this.query({ op: 'isTuple', args: { id: this.id(t, 'type') } }),
      getSignatureFromDeclaration: (n) =>
        this.signature(nodeHandle('signature', n)),
      getResolvedSignature: (n, candidatesOutArray, argumentCount) => {
        const node = this.syntax.parseTreeLocation(n);
        if (!node) return undefined;
        const count =
          argumentCount !== undefined && Number.isFinite(argumentCount)
            ? Math.trunc(argumentCount)
            : undefined;
        const ids = this.query({
          op: 'resolvedSignature',
          args: {
            node,
            withCandidates: candidatesOutArray !== undefined,
            argumentCount: count,
          },
        });
        if (candidatesOutArray !== undefined) {
          for (let i = 1; i < ids.length; i++) {
            const candidate = this.signature(ids[i]);
            if (candidate) candidatesOutArray.push(candidate);
          }
        }
        return this.signature(ids[0] ?? 0);
      },
      getReturnTypeOfSignature: (s) =>
        this.withReplyRecords(() =>
          this.type(
            this.query({
              op: 'returnType',
              args: { id: this.id(s, 'signature') },
            }),
          ),
        ),
      getTypePredicateOfSignature: (s) =>
        this.withReplyRecords(() => {
          const fact = this.query({
            op: 'typePredicate',
            args: { id: this.id(s, 'signature') },
          });
          return fact && { ...fact, type: this.type(fact.type) };
        }),
      getSignaturesOfType: (t, kind) =>
        this.signaturesFor(
          this.query({
            op: 'signatures',
            args: { id: this.id(t, 'type'), kind },
          }),
        ),
    };

    this.typeFactory = makeTypeFactory(this.checker, typescript, ensure);

    this.viewHost = {
      typescript: this.typescript,
      checker: this.checker,
      ensure,
      withReplyRecords: (run) => this.withReplyRecords(run),
      query: (req) => this.query(req),
      typeFact: (id) => this.typeFact(id),
      symbolFact: (id) => this.symbolFact(id),
      signatureFact: (id) => this.signatureFact(id),
      type: (id) => this.type(id),
      symbol: (id, knownName) => this.symbol(id, knownName),
      signature: (id) => this.signature(id),
      typesFor: (ids) => this.typesFor(ids),
      symbolsFor: (ids) => this.symbolsFor(ids),
      indexInfo: (fact) => this.indexInfo(fact),
      nodeRef: (ref) => this.syntax.nodeRef(ref),
      sourceView: (file) => this.syntax.sourceView(file),
      diagnosticSourceView: (file) => this.syntax.diagnosticSourceView(file),
      conditionalRootFor: (node, create) => {
        const existing = this.conditionalRoots.get(node);
        if (existing) return existing;
        const created = create();
        this.conditionalRoots.set(node, created);
        return created;
      },
    };

    const compilerOptions = memo(ensure, () =>
      readCompilerOptions(typescript, descriptor.compilerOptions),
    );
    const rootFileNames = memo(ensure, () =>
      this.query({ op: 'rootFileNames', args: {} }),
    );
    const sourceFiles = memo(ensure, () =>
      descriptor.sources.map((s) => {
        const source = this.syntax.sourceView(s.fileName);
        if (!source) throw new Error('Program source unavailable');
        return source;
      }),
    );

    this.program = {
      getTypeChecker: () => {
        ensure();
        return this.checker;
      },
      getCompilerOptions: compilerOptions,
      getCurrentDirectory: () => {
        ensure();
        return descriptor.currentDirectory;
      },
      emit: () => {
        ensure();
        return unsupported('Program.emit');
      },
      getSourceFileByPath: (path) => this.syntax.sourceByPath(path),
      getConfigFileParsingDiagnostics: () =>
        this.diagnostics({ op: 'configFileParsingDiagnostics', args: {} }),
      getOptionsDiagnostics: () =>
        this.diagnostics({ op: 'optionsDiagnostics', args: {} }),
      getGlobalDiagnostics: () =>
        this.diagnostics({ op: 'globalDiagnostics', args: {} }),
      getSyntacticDiagnostics: (source) =>
        this.diagnostics({
          op: 'syntacticDiagnostics',
          args: { node: source && this.syntax.location(source) },
        }),
      getSemanticDiagnostics: (source) =>
        this.diagnostics({
          op: 'semanticDiagnostics',
          args: { node: source && this.syntax.location(source) },
        }),
      getDeclarationDiagnostics: (source) =>
        this.diagnostics({
          op: 'declarationDiagnostics',
          args: { node: source && this.syntax.location(source) },
        }),
      getRootFileNames: rootFileNames,
      getSourceFile: (name) => this.syntax.sourceView(name),
      getSourceFiles: sourceFiles,
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
    };
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
  ): readonly DiagnosticView[] {
    return this.query(request).map((d) =>
      createDiagnosticView(this.viewHost, d),
    );
  }

  private ensure(): void {
    if (this.closed) throw new Error('type service client is closed');
  }

  close(): void {
    this.closed = true;
    for (const reader of this.facts.values()) reader.close();
    this.facts.clear();
    this.replyRecords = undefined;
    this.cancellationToken = undefined;
    this.factBytes = 0;
    this.types.clear();
    this.symbols.clear();
    this.signatures.clear();
    this.indexInfos.clear();
    for (const file of this.files) file.close();
    this.files.clear();
    this.syntax.close();
    this.transport = undefined;
  }

  private withReplyRecords<T>(run: () => T): T {
    const outer = this.replyRecords;
    this.replyRecords = new Map();
    try {
      return run();
    } finally {
      this.replyRecords = outer;
    }
  }

  private checkCancellation(): void {
    const token = this.cancellationToken;
    if (!token) return;
    if (typeof token.throwIfCancellationRequested === 'function') {
      token.throwIfCancellationRequested();
    } else if (
      typeof token.isCancellationRequested === 'function' &&
      token.isCancellationRequested()
    ) {
      throw new Error('Operation was canceled.');
    }
  }

  private read<K>(
    request: ServiceRequest,
    read: (packet: TypePacketReader) => K,
  ): K {
    this.ensure();
    this.checkCancellation();
    const transport = this.transport;
    if (!transport) throw new Error('type service transport unavailable');
    const query: TypeServiceQuery = {
      generation: this.descriptor.generation,
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
      const reader = this.ingest(query, transport(query, key, false));
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
      const transientBytes = this.replyRecords?.get(key);
      if (transientBytes) {
        this.replyRecords?.delete(key);
        const transient = new TypePacketReader(transientBytes);
        try {
          return read(transient);
        } finally {
          transient.close();
        }
      }
      reader = this.ingest(query, transport(query, key, true), key);
    }
    try {
      return read(reader);
    } finally {
      if (this.facts.get(key) !== reader) reader.close();
    }
  }

  private store(key: string, bytes: Readonly<Uint8Array>): TypePacketReader {
    const existing = this.facts.get(key);
    if (existing) return existing;
    const size = key.length * 2 + bytes.byteLength;
    if (size > this.maxFactsBytes) return new TypePacketReader(bytes);
    const reader = new TypePacketReader(Uint8Array.from(bytes));
    while (this.factBytes + size > this.maxFactsBytes) {
      const oldest = this.facts.entries().next().value;
      if (!oldest) break;
      this.factBytes -= oldest[0].length * 2 + oldest[1].retainedBytes;
      oldest[1].close();
      this.facts.delete(oldest[0]);
    }
    this.facts.set(key, reader);
    this.factBytes += size;
    return reader;
  }

  private ingest(
    query: TypeServiceQuery,
    bytes: Readonly<Uint8Array>,
    cacheKey?: string,
  ): TypePacketReader {
    const reply = typeReplyPackets(bytes);
    for (const record of reply.records) {
      const key = typeQueryKey({
        generation: query.generation,
        op: record.op,
        args: { id: record.id },
      });
      const reader = this.store(key, record.packet);
      if (this.facts.get(key) !== reader) {
        reader.close();
        // Keep uncached type records only for the enclosing eager view build
        // so zero-budget clients avoid re-querying without retaining memory.
        if (this.replyRecords && record.op === 'typeRecord') {
          this.replyRecords.set(key, Uint8Array.from(record.packet));
        }
      }
    }
    return cacheKey === undefined
      ? new TypePacketReader(reply.packet)
      : this.store(cacheKey, reply.packet);
  }

  private query<Q extends ScalarRequest>(request: Q): QueryResults[Q['op']] {
    return this.read(request, (packet) => packet.scalar<Q['op']>(request.op));
  }

  private typeFact(id: number): TypeFact {
    return this.read({ op: 'typeRecord', args: { id } }, (packet) =>
      packet.typeFact(),
    );
  }

  private symbolFact(id: number): SymbolFact {
    return this.read({ op: 'symbolRecord', args: { id } }, (packet) =>
      packet.symbolFact(),
    );
  }

  private signatureFact(id: number): SignatureFact {
    return this.read({ op: 'signatureRecord', args: { id } }, (packet) =>
      packet.signatureFact(),
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
    create: (brand: ObjectBrand) => T,
  ): T | undefined {
    this.ensure();
    if (!id) return undefined;
    const old = map.get(id);
    if (old) return old;
    const brand: ObjectBrand = { kind, id };
    const object = create(brand);
    this.brands.set(object, brand);
    map.set(id, object);
    return object;
  }

  private typesFor(ids: readonly number[]): TypeView[] {
    const out: TypeView[] = [];
    for (const id of ids) {
      const t = this.type(id);
      if (t) out.push(t);
    }
    return out;
  }

  private symbolsFor(ids: readonly number[]): SymbolView[] {
    const out: SymbolView[] = [];
    for (const id of ids) {
      const s = this.symbol(id);
      if (s) out.push(s);
    }
    return out;
  }

  private namedSymbolsFor(facts: readonly NamedSymbolFact[]): SymbolView[] {
    const out: SymbolView[] = [];
    for (const { id, name } of facts) {
      const s = this.symbol(id, name);
      if (s) out.push(s);
    }
    return out;
  }

  private signaturesFor(ids: readonly number[]): SignatureView[] {
    const out: SignatureView[] = [];
    for (const id of ids) {
      const s = this.signature(id);
      if (s) out.push(s);
    }
    return out;
  }

  private indexInfo(fact: IndexInfoFact): IndexInfoView {
    const old = this.indexInfos.get(fact.id);
    if (old) return old;
    const info = createIndexInfoView(this.viewHost, fact);
    this.indexInfos.set(fact.id, info);
    return info;
  }

  private type(id: number): TypeView | undefined {
    return this.intern(this.types, 'type', id, () =>
      createTypeView(this.viewHost, this.typeFactory, id),
    );
  }

  private symbol(id: number, knownName?: string): SymbolView | undefined {
    const view = this.intern(this.symbols, 'symbol', id, (brand) =>
      createSymbolView(this.viewHost, id, () => brand.knownName),
    );
    if (view && knownName !== undefined) {
      const brand = this.brands.get(view);
      if (brand) brand.knownName = knownName;
    }
    return view;
  }

  private signature(id: number): SignatureView | undefined {
    return this.intern(this.signatures, 'signature', id, () =>
      createSignatureView(this.viewHost, id),
    );
  }

  file(binding: FileBinding): TypeServiceFile {
    this.ensure();
    if (binding.generation !== this.descriptor.generation) {
      throw new Error('foreign file binding');
    }
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

export interface ParserServiceOptions {
  readonly emitDecoratorMetadata?: boolean;
  readonly experimentalDecorators?: boolean;
  readonly isolatedDeclarations?: boolean;
}

export function syntaxOnlyParserServices(
  { esTreeNodeToTSNodeMap, tsNodeToESTreeNodeMap }: Readonly<NodeMaps>,
  options?: Readonly<ParserServiceOptions>,
) {
  return {
    esTreeNodeToTSNodeMap,
    tsNodeToESTreeNodeMap,
    program: null,
    emitDecoratorMetadata: options?.emitDecoratorMetadata === true,
    experimentalDecorators: options?.experimentalDecorators === true,
    isolatedDeclarations: options?.isolatedDeclarations === true,
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
    frame: Readonly<DecodedFrame>,
    projection: TypeServiceProjection,
    sourceFileName = this.binding.sourceFileName,
    parserOptions?: Readonly<ParserServiceOptions>,
  ) {
    if (this.closed) throw new Error('file type services are closed');
    this.syntax.attach(frame, projection, sourceFileName);
    const options = this.owner.descriptor.compilerOptions;
    if (this.owner.descriptor.syntaxOnly) {
      return syntaxOnlyParserServices(this.syntax, {
        emitDecoratorMetadata:
          parserOptions?.emitDecoratorMetadata ??
          options.emitDecoratorMetadata === true,
        experimentalDecorators:
          parserOptions?.experimentalDecorators ??
          options.experimentalDecorators === true,
        isolatedDeclarations:
          parserOptions?.isolatedDeclarations ??
          options.isolatedDeclarations === true,
      });
    }
    // A typed Program is authoritative for project compiler flags; ignore
    // per-file parserOptions overrides to match @typescript-eslint/parser.
    const toTS = this.syntax.esTreeNodeToTSNodeMap.get;
    return {
      program: this.owner.program,
      esTreeNodeToTSNodeMap: this.syntax.esTreeNodeToTSNodeMap,
      tsNodeToESTreeNodeMap: this.syntax.tsNodeToESTreeNodeMap,
      getContextualType: (node: Readonly<ESTreeNode>) =>
        this.owner.checker.getContextualType(toTS(node)),
      getResolvedSignature: (
        node: Readonly<ESTreeNode>,
        candidatesOutArray?: SignatureView[],
        argumentCount?: number,
      ) =>
        this.owner.checker.getResolvedSignature(
          toTS(node),
          candidatesOutArray,
          argumentCount,
        ),
      getTypeAtLocation: (node: Readonly<ESTreeNode>) =>
        this.owner.checker.getTypeAtLocation(toTS(node)),
      getSymbolAtLocation: (node: Readonly<ESTreeNode>) =>
        this.owner.checker.getSymbolAtLocation(toTS(node)),
      getTypeFromTypeNode: (node: Readonly<ESTreeNode>) =>
        this.owner.checker.getTypeFromTypeNode(toTS(node)),
      getTypeOfSymbolAtLocation: (
        symbol: Readonly<SymbolView>,
        node: Readonly<ESTreeNode>,
      ) => this.owner.checker.getTypeOfSymbolAtLocation(symbol, toTS(node)),
      emitDecoratorMetadata: options.emitDecoratorMetadata === true,
      experimentalDecorators: options.experimentalDecorators === true,
      isolatedDeclarations: options.isolatedDeclarations === true,
    };
  }
}
