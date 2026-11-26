import { expect, test } from 'rstack/test';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import v8 from 'node:v8';
import vm from 'node:vm';
import { Linter } from 'eslint';
import tsParser from '@typescript-eslint/parser';
import typedPlugin from '@typescript-eslint/eslint-plugin';
import ts from 'typescript';
import {
  TypeServiceClient,
  typeQueryKey,
  type TypeServiceDescriptor,
} from '../../../../rslint/src/languages/js/type-service.js';
import { TypePacketReader } from '../../../../rslint/src/languages/js/type-service-codec.js';
import { TARGET_DEFAULT_LIB } from '../../../../rslint/src/languages/js/scope-factory.js';
import { lintFile } from '../../../../rslint/src/eslint-plugin/linter/ecma-language-plugin.js';
import { decodeFrame } from '../../../../rslint/src/languages/js/frame.js';
import { enums, useTypeServiceHarness } from './harness.js';

const harness = useTypeServiceHarness();

test('decorator metadata flags: typed Program is authoritative, syntax-only honors overrides and preserves custom factories', async () => {
  const dir = mkdtempSync(resolve(harness.temp, 'decorator-precedence-'));
  const serviceFile = resolve(dir, 'service.ts');
  const consumerFile = resolve(dir, 'consumer.ts');
  const withDecoratorsConfig = resolve(dir, 'tsconfig.decorators.json');
  const plainConfig = resolve(dir, 'tsconfig.plain.json');
  const code = `import { Service } from './service';
function Injectable(): ClassDecorator {
  return () => {};
}
@Injectable()
export class Consumer {
  constructor(private readonly service: Service) {}
}
`;
  writeFileSync(serviceFile, 'export class Service {}\n');
  writeFileSync(consumerFile, code);
  writeFileSync(
    withDecoratorsConfig,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        experimentalDecorators: true,
        emitDecoratorMetadata: true,
      },
      include: ['service.ts', 'consumer.ts'],
    }),
  );
  writeFileSync(
    plainConfig,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
      },
      include: ['service.ts', 'consumer.ts'],
    }),
  );

  // Confirm upstream ESLint preserves the import even when conflicting flags
  // are supplied alongside a project tsconfig that enables decorator metadata.
  const eslint = new Linter({ cwd: dir });
  const upstreamEnabled = eslint.verify(
    code,
    [
      {
        files: ['**/*.ts'],
        languageOptions: {
          parser: tsParser,
          parserOptions: {
            project: withDecoratorsConfig,
            tsconfigRootDir: dir,
            emitDecoratorMetadata: false,
            experimentalDecorators: false,
          },
        },
        plugins: { '@typescript-eslint': typedPlugin as any },
        rules: { '@typescript-eslint/consistent-type-imports': 'error' },
      },
    ],
    { filename: consumerFile },
  );
  expect(upstreamEnabled).toEqual([]);

  const runMode = async (
    config: string,
    syntaxOnly: boolean,
    opts?: {
      requestParserOptions?: {
        experimentalDecorators?: boolean;
        emitDecoratorMetadata?: boolean;
      };
      fileParserOptions?: {
        experimentalDecorators?: boolean;
        emitDecoratorMetadata?: boolean;
      };
      customGetterFactory?: boolean;
      probeContext?: boolean;
    },
  ) => {
    const host = harness.createHost();
    const descriptor = host.request({
      method: 'init',
      config,
      syntaxOnly,
    }) as TypeServiceDescriptor;
    const client = new TypeServiceClient(
      enums,
      descriptor,
      (query) =>
        Buffer.from(host.request({ method: 'query', query }), 'base64'),
      host.source,
      { maxFactsBytes: 1 << 20 },
    );
    try {
      const encoded = host.request({ method: 'project', file: consumerFile });
      const frame = Buffer.from(encoded.frame, 'base64');
      const projection = {
        ...encoded,
        tree: Buffer.from(encoded.tree, 'base64'),
      };
      const handle = client.file(projection.binding);
      const rules = new Map([
        [
          '@typescript-eslint/consistent-type-imports',
          typedPlugin.rules['consistent-type-imports'],
        ],
      ]);

      let seenService: Record<string, unknown> | undefined;
      let seenLanguageOptions: unknown;
      let seenSourceCode: any;
      let originalInstance: Record<string, unknown> | undefined;
      const probeRule = {
        meta: { type: 'problem', docs: {}, schema: [], messages: {} },
        create(context: any) {
          seenService = context.sourceCode.parserServices;
          seenLanguageOptions = context.languageOptions;
          seenSourceCode = context.sourceCode;
          return {};
        },
      };
      const allRules =
        opts?.customGetterFactory || opts?.probeContext
          ? new Map([...rules, ['probe/check', probeRule as any]])
          : rules;
      const ruleConfigs: Record<string, { options: unknown[] }> = {
        '@typescript-eslint/consistent-type-imports': { options: [] },
        ...(opts?.customGetterFactory || opts?.probeContext
          ? { 'probe/check': { options: [] } }
          : {}),
      };

      const factory =
        !opts?.customGetterFactory && !opts?.fileParserOptions
          ? handle.parserServicesFactory(projection)
          : (decoded: any, parserOptions: any) => {
              const base = handle.parserServices(
                decoded,
                projection,
                projection.binding.sourceFileName,
                opts?.fileParserOptions ?? parserOptions,
              );
              if (!opts?.customGetterFactory) return base;
              // Custom class factory applies the file's parserOptions at
              // construction time; lintFile passes the original instance directly
              // to rules so private methods, `#field in this`, setters, read-only
              // fields, non-configurable defineProperty, and domain errors behave
              // identically to native class instances without double execution.
              class CustomSyntaxServices {
                readonly #tag = 'kept';
                readonly #maps = {
                  esTreeNodeToTSNodeMap: base.esTreeNodeToTSNodeMap,
                  tsNodeToESTreeNodeMap: base.tsNodeToESTreeNodeMap,
                };
                #label = 'initial';
                readonly program = null;
                readonly emitDecoratorMetadata: boolean;
                readonly experimentalDecorators: boolean;
                readonly isolatedDeclarations: boolean;
                readonly locked = 'source';
                calls = 0;
                constructor() {
                  this.emitDecoratorMetadata =
                    parserOptions?.emitDecoratorMetadata === true;
                  this.experimentalDecorators =
                    parserOptions?.experimentalDecorators === true;
                  this.isolatedDeclarations =
                    parserOptions?.isolatedDeclarations === true;
                  Object.defineProperty(this, 'locked', {
                    value: 'source',
                    writable: false,
                    enumerable: true,
                    configurable: false,
                  });
                }
                #helper() {
                  return this.#tag;
                }
                get esTreeNodeToTSNodeMap() {
                  return this.#maps.esTreeNodeToTSNodeMap;
                }
                get tsNodeToESTreeNodeMap() {
                  return this.#maps.tsNodeToESTreeNodeMap;
                }
                get label() {
                  return this.#label;
                }
                set label(value: string) {
                  this.#label = value.trim();
                }
                readMixed() {
                  this.calls += 1;
                  return [this.#helper(), this.emitDecoratorMetadata] as const;
                }
                hasPrivate() {
                  return #tag in this;
                }
                throwDomainError(): never {
                  this.calls += 1;
                  throw new TypeError('private field metadata is unavailable');
                }
              }
              const instance = new CustomSyntaxServices();
              originalInstance = instance as unknown as Record<string, unknown>;
              return instance;
            };

      const result = lintFile(
        {
          filePath: consumerFile,
          frame,
          languageOptions: opts?.requestParserOptions
            ? { parserOptions: opts.requestParserOptions }
            : undefined,
          rules: ruleConfigs,
          collectFixes: true,
          suggestionsMode: 'eager',
        },
        { rules: allRules } as any,
        factory,
      );
      expect(result.parseError).toBeUndefined();
      expect(result.ruleErrors).toBeUndefined();
      return {
        result,
        seenService,
        seenLanguageOptions,
        seenSourceCode,
        originalInstance,
      };
    } finally {
      client.close();
      await host.close();
    }
  };

  // Typed project: tsconfig is authoritative; conflicting per-file flags (either
  // passed via handle.parserServices or via lintFile request) MUST NOT override it.
  // Also verify lintFile passes req.languageOptions through untouched and
  // snapshots compiler options eagerly so deferred getScope() survives client.close().
  const typedDefault = await runMode(withDecoratorsConfig, false, {
    probeContext: true,
  });
  expect(typedDefault.result.diagnostics).toEqual([]);
  expect(typedDefault.seenLanguageOptions).toEqual({ parserOptions: {} });
  expect(
    typedDefault.seenSourceCode.getScope(typedDefault.seenSourceCode.ast),
  ).toBeDefined();

  const typedFileOverride = await runMode(withDecoratorsConfig, false, {
    fileParserOptions: {
      experimentalDecorators: false,
      emitDecoratorMetadata: false,
    },
  });
  expect(typedFileOverride.result.diagnostics).toEqual([]);

  const typedRequestOverride = await runMode(withDecoratorsConfig, false, {
    requestParserOptions: {
      experimentalDecorators: false,
      emitDecoratorMetadata: false,
    },
  });
  expect(typedRequestOverride.result.diagnostics).toEqual([]);

  // Syntax-only: tsconfig propagates decorator flags.
  const syntaxFromTsconfig = await runMode(withDecoratorsConfig, true);
  expect(syntaxFromTsconfig.result.diagnostics).toEqual([]);

  // Syntax-only: per-file parserOptions override works in both directions.
  const syntaxEnabledByOverride = await runMode(plainConfig, true, {
    requestParserOptions: {
      experimentalDecorators: true,
      emitDecoratorMetadata: true,
    },
  });
  expect(syntaxEnabledByOverride.result.diagnostics).toEqual([]);

  const syntaxDisabledByDefault = await runMode(plainConfig, true);
  expect(syntaxDisabledByDefault.result.diagnostics).toHaveLength(1);

  // Syntax-only custom class factory receives parserOptions at construction
  // time and rules receive the original class instance directly (no wrapper).
  const customWithOverride = await runMode(plainConfig, true, {
    requestParserOptions: {
      experimentalDecorators: true,
      emitDecoratorMetadata: true,
    },
    customGetterFactory: true,
  });
  expect(customWithOverride.result.diagnostics).toEqual([]);
  const customSvc = customWithOverride.seenService as any;
  expect(customSvc).toBe(customWithOverride.originalInstance);
  expect(customSvc.emitDecoratorMetadata).toBe(true);
  expect(customSvc.esTreeNodeToTSNodeMap).toBeDefined();
  expect(customSvc.readMixed()).toEqual(['kept', true]);
  expect(customSvc.calls).toBe(1);
  expect(customSvc.hasPrivate()).toBe(true);
  expect(() => customSvc.throwDomainError()).toThrow(
    'private field metadata is unavailable',
  );
  expect(customSvc.calls).toBe(2);
  customSvc.label = '  active  ';
  expect(customSvc.label).toBe('active');
  expect(Reflect.set(customSvc, 'locked', 'overwritten')).toBe(false);
  expect(customSvc.locked).toBe('source');
  Object.defineProperty(customSvc, 'memoSlot', {
    value: 7,
    writable: false,
    enumerable: false,
    configurable: false,
  });
  expect(Object.getOwnPropertyDescriptor(customSvc, 'memoSlot')).toEqual({
    value: 7,
    writable: false,
    enumerable: false,
    configurable: false,
  });
  expect(customSvc.readMixed).toBe(customSvc.readMixed);
});

test('ts.getPreEmitDiagnostics and config diagnostics preserve tsconfig.json location', async () => {
  // Include U+0130 (İ) in the fixture path so canonical-path comparisons
  // cannot rely on JavaScript String.prototype.toLowerCase().
  const dir = mkdtempSync(resolve(harness.temp, 'pre-emit-config-\u0130-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'Project.Config.json');
  writeFileSync(file, `export const bad: number = 'oops';\n`);
  writeFileSync(
    config,
    JSON.stringify(
      {
        compilerOptions: {
          strict: true,
          target: 'ES2022',
          module: 'ESNext',
          skipLibCheck: true,
          inlineSources: true,
        },
        include: ['input.ts'],
      },
      null,
      2,
    ) + '\n',
  );

  const cmd = ts.getParsedCommandLineOfConfigFile(
    config,
    {},
    {
      ...ts.sys,
      onUnRecoverableConfigFileDiagnostic: (d) => {
        throw new Error(ts.flattenDiagnosticMessageText(d.messageText, '\n'));
      },
    },
  )!;
  const reference = ts.createProgram({
    rootNames: cmd.fileNames,
    options: cmd.options,
    configFileParsingDiagnostics: cmd.errors,
  });

  const host = harness.createHost();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const owner = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 20 },
  );

  try {
    const actualProgram = owner.program as unknown as ts.Program;
    // Config file must remain excluded from Program source-file APIs.
    expect(actualProgram.getSourceFile(config)).toBeUndefined();
    expect(
      actualProgram
        .getSourceFiles()
        .some((s) => s.fileName === config.replaceAll('\\', '/')),
    ).toBe(false);

    const normalizeSlashes = (p: string | undefined) =>
      p?.replaceAll('\\', '/');
    const actualTs5051 = actualProgram
      .getOptionsDiagnostics()
      .find((d) => d.code === 5051)!;
    const refTs5051 = reference
      .getOptionsDiagnostics()
      .find((d) => d.code === 5051)!;

    // Cold reads of diagnostic.file.text, flags, and path MUST work before
    // ts.formatDiagnostic loads the file.
    expect(normalizeSlashes(actualTs5051.file?.fileName)).toBe(
      normalizeSlashes(refTs5051.file?.fileName),
    );
    expect(normalizeSlashes((actualTs5051.file as any)?.path)).toBe(
      normalizeSlashes((refTs5051.file as any)?.path),
    );
    expect(
      actualTs5051.file?.text.slice(
        actualTs5051.start!,
        actualTs5051.start! + actualTs5051.length!,
      ),
    ).toBe('"inlineSources"');
    expect(actualTs5051.file?.flags).toBe(refTs5051.file?.flags);

    const formatHost: ts.FormatDiagnosticsHost = {
      getCurrentDirectory: () => dir.replaceAll('\\', '/'),
      getCanonicalFileName: (name) => name.replaceAll('\\', '/'),
      getNewLine: () => '\n',
    };
    const summarize = (diags: readonly ts.Diagnostic[]) =>
      diags
        .map((d) => [
          d.code,
          d.start,
          d.length,
          normalizeSlashes(d.file?.fileName),
        ])
        .sort((a, b) => Number(a[0]) - Number(b[0]));

    expect(summarize(actualProgram.getOptionsDiagnostics())).toEqual(
      summarize(reference.getOptionsDiagnostics()),
    );
    expect(summarize(ts.getPreEmitDiagnostics(actualProgram))).toEqual(
      summarize(ts.getPreEmitDiagnostics(reference)),
    );
    expect(ts.formatDiagnostic(actualTs5051, formatHost)).toBe(
      ts.formatDiagnostic(refTs5051, formatHost),
    );
    expect(ts.formatDiagnostic(actualTs5051, formatHost)).toContain(
      'Project.Config.json(',
    );

    // Late-discovered inherited config diagnostics carry the compiler's
    // canonical path even when the host SyntaxSource omits `path`.
    const baseConfig = resolve(dir, 'Base.Config.json');
    const lowerBaseConfig = resolve(dir, 'base.config.json');
    const baseConfigContent =
      JSON.stringify(
        {
          compilerOptions: {
            strict: true,
            unknownInheritedOption: true,
          },
        },
        null,
        2,
      ) + '\n';
    writeFileSync(baseConfig, baseConfigContent);
    if (ts.sys.useCaseSensitiveFileNames) {
      // On case-sensitive hosts, create two distinct files so both spellings
      // produce distinct canonical paths.
      writeFileSync(lowerBaseConfig, baseConfigContent);
    }
    const childConfig = resolve(dir, 'Child.Config.json');
    writeFileSync(
      childConfig,
      JSON.stringify(
        {
          extends: './Base.Config.json',
          compilerOptions: {
            target: 'ES2022',
            module: 'ESNext',
            skipLibCheck: true,
          },
          include: ['input.ts'],
        },
        null,
        2,
      ) + '\n',
    );
    const childConfigLower = resolve(dir, 'Child.Lower.json');
    writeFileSync(
      childConfigLower,
      JSON.stringify(
        {
          extends: './base.config.json',
          compilerOptions: {
            target: 'ES2022',
            module: 'ESNext',
            skipLibCheck: true,
          },
          include: ['input.ts'],
        },
        null,
        2,
      ) + '\n',
    );
    const readInheritedDiag = async (cfg: string) => {
      const inheritedHost = harness.createHost();
      const inheritedDesc = inheritedHost.request({
        method: 'init',
        config: cfg,
      }) as TypeServiceDescriptor;
      const legacySource = (
        scope: { generation: number },
        reqFile: string,
        metadataOnly?: boolean,
      ) => {
        const reply = inheritedHost.source(scope, reqFile, metadataOnly);
        return reply
          ? { fileName: reply.fileName, tree: reply.tree }
          : undefined;
      };
      const inheritedClient = new TypeServiceClient(
        enums,
        inheritedDesc,
        (query) =>
          Buffer.from(
            inheritedHost.request({ method: 'query', query }),
            'base64',
          ),
        legacySource,
        { maxFactsBytes: 1 << 20 },
      );
      try {
        const inheritedActual =
          inheritedClient.program as unknown as ts.Program;
        const diag = inheritedActual
          .getConfigFileParsingDiagnostics()
          .find((d) => d.code === 5023)!;
        return {
          fileName: normalizeSlashes(diag.file?.fileName),
          path: normalizeSlashes((diag.file as any)?.path),
        };
      } finally {
        inheritedClient.close();
        await inheritedHost.close();
      }
    };
    const upperInherited = await readInheritedDiag(childConfig);
    const lowerInherited = await readInheritedDiag(childConfigLower);
    const canonicalDir = normalizeSlashes(
      (refTs5051.file as any)?.path,
    )!.replace(/\/project\.config\.json$/i, '');
    const normalizedBaseConfig = baseConfig.replaceAll('\\', '/');
    const normalizedLowerBaseConfig = lowerBaseConfig.replaceAll('\\', '/');
    expect(upperInherited.fileName).toBe(normalizedBaseConfig);
    expect(lowerInherited.fileName).toBe(normalizedLowerBaseConfig);
    if (ts.sys.useCaseSensitiveFileNames) {
      expect(upperInherited.path).toBe(`${canonicalDir}/Base.Config.json`);
      expect(lowerInherited.path).toBe(`${canonicalDir}/base.config.json`);
      expect(lowerInherited.path).not.toBe(upperInherited.path);
    } else {
      expect(upperInherited.path).toBe(`${canonicalDir}/base.config.json`);
      expect(lowerInherited.path).toBe(upperInherited.path);
    }
  } finally {
    owner.close();
    await host.close();
  }
});

test('getResolvedSignature forwards candidates and argumentCount without poisoning cached signatures', async () => {
  const dir = mkdtempSync(resolve(harness.temp, 'overload-resolution-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'tsconfig.json');
  writeFileSync(
    file,
    `export function pick(a: string): 1;
export function pick(a: number, b: boolean): 2;
export function pick(a: string | number, b?: boolean): 1 | 2 {
  return b ? 2 : 1;
}
export function choose<T extends string>(x: T): 1;
export function choose<T extends string>(x: T, y: T): 2;
export function choose<T extends string>(x: T, y?: T): 1 | 2 {
  return y === undefined ? 1 : 2;
}
export const coldInvalid = pick(true);
export const genericOne = choose(true);
export const genericTwo = choose(true);
export const genericWithCandidates = choose(true);
export const genericWarm = choose(true);
export const incomplete = pick(true as unknown as never);
export const call1 = pick('ok');
export const call2 = pick(1, true);
`,
  );
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        skipLibCheck: true,
      },
      include: ['input.ts'],
    }),
  );

  const cmd = ts.getParsedCommandLineOfConfigFile(
    config,
    {},
    {
      ...ts.sys,
      onUnRecoverableConfigFileDiagnostic: (d) => {
        throw new Error(ts.flattenDiagnosticMessageText(d.messageText, '\n'));
      },
    },
  )!;
  const reference = ts.createProgram({
    rootNames: cmd.fileNames,
    options: cmd.options,
    configFileParsingDiagnostics: cmd.errors,
  });

  const host = harness.createHost();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const owner = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 20 },
  );

  try {
    const exercise = (p: ts.Program) => {
      const c = p.getTypeChecker();
      const src = p.getSourceFile(file)!;
      const callFor = (varName: string) => {
        const stmt = src.statements.find(
          (s): s is ts.VariableStatement =>
            ts.isVariableStatement(s) &&
            ts.isIdentifier(s.declarationList.declarations[0].name) &&
            s.declarationList.declarations[0].name.text === varName,
        )!;
        return stmt.declarationList.declarations[0]
          .initializer as ts.CallExpression;
      };
      const retOf = (sig: ts.Signature | undefined) =>
        sig ? c.typeToString(c.getReturnTypeOfSignature(sig)) : undefined;

      // Cold invalid non-generic overloaded call queried with (undefined, 2):
      // without a candidates array, TypeScript combines overloads (`never`).
      const coldInvalid = callFor('coldInvalid');
      const coldFirst = retOf(
        c.getResolvedSignature(coldInvalid, undefined, 2),
      );
      const coldNext = retOf(c.getResolvedSignature(coldInvalid));

      // Cold invalid generic overloaded calls: argumentCount selects the
      // matching overload even when candidatesOutArray is omitted.
      const genericOne = callFor('genericOne');
      const genericOneSeq = [
        retOf(c.getResolvedSignature(genericOne, undefined, 1)),
        retOf(c.getResolvedSignature(genericOne)),
      ];
      const genericTwo = callFor('genericTwo');
      const genericTwoSeq = [
        retOf(c.getResolvedSignature(genericTwo, undefined, 2)),
        retOf(c.getResolvedSignature(genericTwo)),
      ];
      const genericCandCall = callFor('genericWithCandidates');
      const genericCandidates: ts.Signature[] = [];
      const genericCandSeq = [
        retOf(c.getResolvedSignature(genericCandCall, genericCandidates, 2)),
        retOf(c.getResolvedSignature(genericCandCall)),
      ];
      const genericWarm = callFor('genericWarm');
      const genericWarmSeq = [
        retOf(c.getResolvedSignature(genericWarm)),
        retOf(c.getResolvedSignature(genericWarm, undefined, 2)),
        retOf(c.getResolvedSignature(genericWarm)),
      ];

      // Warm non-generic call: querying with (undefined, 2) after caching MUST NOT
      // overwrite the cached overload; passing a candidates array DOES resolve
      // with argumentCount = 2 and populates both candidates.
      const incomplete = callFor('incomplete');
      const first = retOf(c.getResolvedSignature(incomplete));
      const countWithoutCandidates = retOf(
        c.getResolvedSignature(incomplete, undefined, 2),
      );
      const after = retOf(c.getResolvedSignature(incomplete));
      const candidates: ts.Signature[] = [];
      const countWithCandidates = retOf(
        c.getResolvedSignature(incomplete, candidates, 2),
      );
      const afterCandidates = retOf(c.getResolvedSignature(incomplete));

      const call1Candidates: ts.Signature[] = [];
      const resolved1 = c.getResolvedSignature(
        callFor('call1'),
        call1Candidates,
      )!;
      const call2Candidates: ts.Signature[] = [];
      const resolved2 = c.getResolvedSignature(
        callFor('call2'),
        call2Candidates,
      )!;

      return {
        coldFirst,
        coldNext,
        genericOneSeq,
        genericTwoSeq,
        genericCandSeq,
        genericCandidates: genericCandidates.map(retOf),
        genericWarmSeq,
        first,
        countWithoutCandidates,
        after,
        countWithCandidates,
        candidatesInfo: candidates.map(retOf),
        afterCandidates,
        resolved1: c.signatureToString(resolved1),
        call1Candidates: call1Candidates.map((s) => c.signatureToString(s)),
        resolved2: c.signatureToString(resolved2),
        call2Candidates: call2Candidates.map((s) => c.signatureToString(s)),
      };
    };

    const actual = exercise(owner.program as unknown as ts.Program);
    const expected = exercise(reference);
    expect(actual).toEqual(expected);
    expect([actual.coldFirst, actual.coldNext]).toEqual(['never', 'never']);
    expect(actual.genericOneSeq).toEqual(['1', '1']);
    expect(actual.genericTwoSeq).toEqual(['2', '2']);
    expect(actual.genericWarmSeq).toEqual(['1', '1', '1']);
    expect([actual.first, actual.countWithoutCandidates, actual.after]).toEqual(
      ['1', '1', '1'],
    );
    expect(actual.candidatesInfo).toEqual(['1', '2']);
  } finally {
    owner.close();
    await host.close();
  }
});

test('core checker helpers and runWithCancellationToken match TypeScript', async () => {
  const dir = mkdtempSync(resolve(harness.temp, 'checker-helpers-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'tsconfig.json');
  writeFileSync(
    file,
    `export function pick(a: string): string;
export function pick(a: number, b: boolean): number;
export function pick(a: string | number, b?: boolean): string | number {
  return b ? a : a;
}
export class Holder {
  constructor(public readonly value: number, opt?: string) {
    void opt;
  }
}
export function isNum(x: unknown): x is number {
  return typeof x === 'number';
}
export function assertPresent(x: unknown): asserts x {
  if (!x) throw new Error('missing');
}
`,
  );
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        skipLibCheck: true,
      },
      include: ['input.ts'],
    }),
  );

  const cmd = ts.getParsedCommandLineOfConfigFile(
    config,
    {},
    {
      ...ts.sys,
      onUnRecoverableConfigFileDiagnostic: (d) => {
        throw new Error(ts.flattenDiagnosticMessageText(d.messageText, '\n'));
      },
    },
  )!;
  const reference = ts.createProgram({
    rootNames: cmd.fileNames,
    options: cmd.options,
    configFileParsingDiagnostics: cmd.errors,
  });

  const host = harness.createHost();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const owner = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 20 },
  );

  try {
    const exercise = (p: ts.Program) => {
      const c = p.getTypeChecker();
      const src = p.getSourceFile(file)!;
      const pickDecls = src.statements.filter(
        (s): s is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(s) && s.name?.text === 'pick',
      );
      const holder = src.statements.find(
        (s): s is ts.ClassDeclaration =>
          ts.isClassDeclaration(s) && s.name?.text === 'Holder',
      )!;
      const ctor = holder.members.find(ts.isConstructorDeclaration)!;
      const isNum = src.statements.find(
        (s): s is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(s) && s.name?.text === 'isNum',
      )!;
      const assertPresent = src.statements.find(
        (s): s is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(s) && s.name?.text === 'assertPresent',
      )!;
      const pred = c.getTypePredicateOfSignature(
        c.getSignatureFromDeclaration(isNum)!,
      )!;
      const assertsPred = c.getTypePredicateOfSignature(
        c.getSignatureFromDeclaration(assertPresent)!,
      )!;

      // Cancellation: pre-cancelled token throws during checker operations.
      let tokenPolls = 0;
      const cancelledToken: ts.CancellationToken = {
        isCancellationRequested: () => true,
        throwIfCancellationRequested() {
          tokenPolls++;
          throw new Error('cancelled by caller');
        },
      };
      const stringType = c.getStringType();
      let cancellationOutcome = '';
      try {
        cancellationOutcome = c.runWithCancellationToken(
          cancelledToken,
          (inner) => inner.typeToString(stringType),
        );
      } catch (err) {
        cancellationOutcome = String(err);
      }
      // After runWithCancellationToken exits, subsequent queries must not poll it.
      const pollsBeforeOutside = tokenPolls;
      const afterTokenFormat = c.typeToString(stringType);
      const pollsAfterOutside = tokenPolls;

      return {
        overloads: pickDecls.map((d) => c.isImplementationOfOverload(d)),
        optionalParams: ctor.parameters.map((param) =>
          c.isOptionalParameter(param),
        ),
        paramPropSymbols: c
          .getSymbolsOfParameterPropertyDeclaration(
            ctor.parameters[0] as ts.ParameterDeclaration,
            'value',
          )
          .map((s) => [s.name, s.flags]),
        literals: [
          c.typeToString(c.getStringLiteralType('hello')),
          c.typeToString(c.getNumberLiteralType(42)),
          c.typeToString(c.getNumberLiteralType(Number.NaN)),
          c.typeToString(c.getNumberLiteralType(Number.POSITIVE_INFINITY)),
          c.typeToString(
            c.getBigIntLiteralType({ negative: true, base10Value: '99' }),
          ),
          c.typeToString(
            c.getBigIntLiteralType({ negative: true, base10Value: '0' }),
          ),
          c.typeToString(c.getTrueType()),
          c.typeToString(c.getFalseType()),
          c.typeToString(
            c.getNullableType(c.getStringType(), ts.TypeFlags.Undefined),
          ),
        ],
        predicateString: c.typePredicateToString(pred, isNum),
        assertsPredicateString: c.typePredicateToString(
          assertsPred,
          assertPresent,
        ),
        withActiveToken: c.runWithCancellationToken(
          {
            isCancellationRequested: () => false,
            throwIfCancellationRequested() {},
          },
          (inner) => inner.typeToString(inner.getTrueType()),
        ),
        cancellationOutcome,
        polledAtLeastOnce: tokenPolls >= 1,
        restoredAfterExit: pollsAfterOutside === pollsBeforeOutside,
        afterTokenFormat,
      };
    };

    expect(exercise(owner.program as unknown as ts.Program)).toEqual(
      exercise(reference),
    );
  } finally {
    owner.close();
    await host.close();
  }
});

test('zero-byte fact cache uses single-pass reads without retaining pending records', async () => {
  const dir = mkdtempSync(resolve(harness.temp, 'zero-cache-budget-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'tsconfig.json');
  const propLines = Array.from(
    { length: 1000 },
    (_, i) => `  p${i}: number;`,
  ).join('\n');
  writeFileSync(
    file,
    `export interface Huge {\n${propLines}\n}\nexport declare const huge: Huge;\nexport function isNum(x: unknown): x is number {\n  return typeof x === 'number';\n}\n`,
  );
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        skipLibCheck: true,
      },
      include: ['input.ts'],
    }),
  );

  const host = harness.createHost();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;

  let transportCalls = 0;
  let transportRecordCalls = 0;
  let serverTrips = 0;
  let recordServerTrips = 0;
  let lastReplyBytes = 0;
  const sharedTransportCache = new Map<string, Uint8Array>();
  const sharedTransport = (
    query: Parameters<typeof typeQueryKey>[0],
    key: string,
    cacheable: boolean,
  ): Uint8Array => {
    transportCalls++;
    const isRecordOp =
      query.op === 'typeRecord' ||
      query.op === 'symbolRecord' ||
      query.op === 'signatureRecord';
    if (isRecordOp) transportRecordCalls++;
    if (cacheable) {
      const cached = sharedTransportCache.get(key);
      if (cached) {
        lastReplyBytes = cached.byteLength;
        return cached;
      }
    }
    serverTrips++;
    if (isRecordOp) recordServerTrips++;
    const reply = Buffer.from(
      host.request({ method: 'query', query }),
      'base64',
    );
    if (cacheable) sharedTransportCache.set(key, reply);
    lastReplyBytes = reply.byteLength;
    return reply;
  };
  const zeroCacheClient = new TypeServiceClient(
    enums,
    descriptor,
    sharedTransport,
    host.source,
    { maxFactsBytes: 0 },
  );

  try {
    const zc = zeroCacheClient.program.getTypeChecker();

    // Compact vs full reply shapes MUST produce distinct transport cache keys
    // for operations that bundle supplemental records, while operations whose
    // replies are byte-identical across modes MUST share the same key.
    expect(
      typeQueryKey({
        generation: descriptor.generation,
        op: 'properties',
        args: { id: 1 },
        skipRecords: true,
      }),
    ).not.toBe(
      typeQueryKey({
        generation: descriptor.generation,
        op: 'properties',
        args: { id: 1 },
      }),
    );
    expect(
      typeQueryKey({
        generation: descriptor.generation,
        op: 'typeString',
        args: { id: 1, node: undefined, flags: undefined },
        skipRecords: true,
      }),
    ).toBe(
      typeQueryKey({
        generation: descriptor.generation,
        op: 'typeString',
        args: { id: 1, node: undefined, flags: undefined },
      }),
    );
    expect(
      typeQueryKey({
        generation: descriptor.generation,
        op: 'isArray',
        args: { id: 1 },
        skipRecords: true,
      }),
    ).toBe(
      typeQueryKey({
        generation: descriptor.generation,
        op: 'isArray',
        args: { id: 1 },
      }),
    );

    // Cold getStringType() consumes the bundled typeRecord during eager view
    // construction: 1 transport call total, 0 record calls, 0 retained bytes.
    transportCalls = 0;
    transportRecordCalls = 0;
    const zString = zc.getStringType();
    expect(zString.flags).toBe(ts.TypeFlags.String);
    expect(zString.intrinsicName).toBe('string');
    expect(transportCalls).toBe(1);
    expect(transportRecordCalls).toBe(0);

    // Reading all shallow fields of a lazy SymbolView / SignatureView with a
    // zero-byte cache decodes the record once rather than per field (1 record
    // call instead of 5-6).
    const zSrc = zeroCacheClient.program.getSourceFile(file) as ts.SourceFile;
    const zIsNum = zSrc.statements.find(
      (s): s is ts.FunctionDeclaration =>
        ts.isFunctionDeclaration(s) && s.name?.text === 'isNum',
    )!;
    transportCalls = 0;
    transportRecordCalls = 0;
    const zSym = zc.getSymbolAtLocation(zIsNum.name!)!;
    void zSym.name;
    void zSym.escapedName;
    void zSym.flags;
    void zSym.declarations;
    void zSym.valueDeclaration;
    void zSym.links;
    expect(transportRecordCalls).toBeLessThanOrEqual(1);

    transportCalls = 0;
    transportRecordCalls = 0;
    const zSig = zc.getSignatureFromDeclaration(zIsNum)!;
    void zSig.parameters;
    void zSig.declaration;
    void zSig.typeParameters;
    void zSig.thisParameter;
    void zSig.minArgumentCount;
    expect(transportRecordCalls).toBeLessThanOrEqual(1);

    // Listing 1,000 properties and reading every property name under
    // maxFactsBytes: 0 takes 1 transport call total (0 symbolRecord round trips)
    // and skips shipping discarded supplemental symbol records over the wire.
    const hugeDecl = zSrc.statements.find(
      (s): s is ts.VariableStatement =>
        ts.isVariableStatement(s) &&
        ts.isIdentifier(s.declarationList.declarations[0].name) &&
        s.declarationList.declarations[0].name.text === 'huge',
    )!.declarationList.declarations[0];
    const hugeType = zc.getTypeAtLocation(hugeDecl)!;
    transportCalls = 0;
    transportRecordCalls = 0;
    serverTrips = 0;
    lastReplyBytes = 0;
    const props = zc.getPropertiesOfType(hugeType);
    const propNames = props.map((p) => p.name);
    expect(propNames).toHaveLength(1000);
    expect(propNames[0]).toBe('p0');
    expect(propNames[999]).toBe('p999');
    expect(transportCalls).toBe(1);
    expect(transportRecordCalls).toBe(0);
    expect(serverTrips).toBe(1);
    expect(lastReplyBytes).toBeLessThan(40_000);

    // Seed shared transport cache with zero-budget replies for operations that
    // never bundle supplemental records (`typeToString`, `isArrayType`).
    const zSampleTypes = [zc.getStringType(), zc.getNumberType(), hugeType];
    const zFormatted = zSampleTypes.map((t) => zc.typeToString(t));
    const zArrays = zSampleTypes.map((t) => zc.isArrayType(t));

    // A second client with an 8 MiB fact budget sharing the same transport
    // cache MUST NOT receive the zero-budget client's compact `properties`
    // reply (1 bulk server trip, 0 per-symbol trips), AND MUST reuse cached
    // `typeToString` / `isArrayType` replies with 0 server trips.
    const cachedClient = new TypeServiceClient(
      enums,
      descriptor,
      sharedTransport,
      host.source,
      { maxFactsBytes: 8 << 20 },
    );
    try {
      const cc = cachedClient.program.getTypeChecker();
      const cSrc = cachedClient.program.getSourceFile(file) as ts.SourceFile;
      const cHugeDecl = cSrc.statements.find(
        (s): s is ts.VariableStatement =>
          ts.isVariableStatement(s) &&
          ts.isIdentifier(s.declarationList.declarations[0].name) &&
          s.declarationList.declarations[0].name.text === 'huge',
      )!.declarationList.declarations[0];
      const cHugeType = cc.getTypeAtLocation(cHugeDecl)!;
      serverTrips = 0;
      recordServerTrips = 0;
      const cachedProps = cc.getPropertiesOfType(cHugeType);
      const flags = cachedProps.map((p) => p.flags);
      expect(flags).toHaveLength(1000);
      expect(flags[0]).toBe(ts.SymbolFlags.Property);
      expect(flags[999]).toBe(ts.SymbolFlags.Property);
      expect(serverTrips).toBe(1);
      expect(recordServerTrips).toBe(0);

      const cSampleTypes = [cc.getStringType(), cc.getNumberType(), cHugeType];
      serverTrips = 0;
      recordServerTrips = 0;
      expect(cSampleTypes.map((t) => cc.typeToString(t))).toEqual(zFormatted);
      expect(cSampleTypes.map((t) => cc.isArrayType(t))).toEqual(zArrays);
      expect(serverTrips).toBe(0);
      expect(recordServerTrips).toBe(0);
    } finally {
      cachedClient.close();
    }
  } finally {
    zeroCacheClient.close();
    await host.close();
  }
});

test('TARGET_DEFAULT_LIB aligns with TypeScript ScriptTarget', () => {
  expect(TARGET_DEFAULT_LIB[ts.ScriptTarget.ES2015]).toBe('es6');
  expect(TARGET_DEFAULT_LIB[ts.ScriptTarget.ES2016]).toBe('es2016.full');
  expect(TARGET_DEFAULT_LIB[ts.ScriptTarget.ES2022]).toBe('es2022.full');
  expect(TARGET_DEFAULT_LIB[ts.ScriptTarget.ES2025]).toBe('es2025.full');
  expect(TARGET_DEFAULT_LIB[ts.ScriptTarget.ESNext]).toBe('esnext.full');
});

test('withFile supports sync/async callbacks and GC preserves live node maps', async () => {
  v8.setFlagsFromString('--expose_gc');
  const runGC = vm.runInNewContext('gc') as () => void;
  const forceGC = async () => {
    for (let i = 0; i < 4; i++) {
      runGC();
      await new Promise((done) => setTimeout(done, 10));
    }
  };

  const dir = mkdtempSync(resolve(harness.temp, 'file-lifecycle-'));
  const file = resolve(dir, 'input.ts');
  const config = resolve(dir, 'tsconfig.json');
  writeFileSync(file, 'export const ok = 1;\n');
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        skipLibCheck: true,
      },
      include: ['input.ts'],
    }),
  );
  const host = harness.createHost();
  const syntaxHost = harness.createHost();
  const typedDescriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const syntaxDescriptor = syntaxHost.request({
    method: 'init',
    config,
    syntaxOnly: true,
  }) as TypeServiceDescriptor;
  const client = new TypeServiceClient(
    enums,
    typedDescriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 20 },
  );
  const syntaxClient = new TypeServiceClient(
    enums,
    syntaxDescriptor,
    (query) =>
      Buffer.from(syntaxHost.request({ method: 'query', query }), 'base64'),
    syntaxHost.source,
    { maxFactsBytes: 1 << 20 },
  );
  try {
    const encoded = host.request({
      method: 'project',
      file,
    });
    const projection = {
      ...encoded,
      tree: Buffer.from(encoded.tree, 'base64'),
    };
    const decoded = decodeFrame(Buffer.from(encoded.frame, 'base64'));
    const stmt = (decoded.ast.body as readonly any[])[0];

    const syntaxEncoded = syntaxHost.request({
      method: 'project',
      file,
    });
    const syntaxProjection = {
      ...syntaxEncoded,
      tree: Buffer.from(syntaxEncoded.tree, 'base64'),
    };
    const syntaxDecoded = decodeFrame(
      Buffer.from(syntaxEncoded.frame, 'base64'),
    );
    const syntaxStmt = (syntaxDecoded.ast.body as readonly any[])[0];

    // Sync withFile closes both the handle and its node maps upon return.
    let syncHandle: ReturnType<TypeServiceClient['file']> | undefined;
    let syncServices: ReturnType<TypeServiceClient['file']> extends {
      parserServices: (...args: any[]) => infer R;
    }
      ? R
      : never;
    const syncKind = client.withFile(projection.binding, (handle) => {
      syncHandle = handle;
      syncServices = handle.parserServices(decoded, projection);
      return (syncServices.esTreeNodeToTSNodeMap.get(stmt) as any)?.kind;
    });
    expect(syncKind).toBe(ts.SyntaxKind.VariableStatement);
    expect(() => syncHandle!.parserServices(decoded, projection)).toThrow(
      'file type services are closed',
    );
    expect(() => syncServices!.esTreeNodeToTSNodeMap.get(stmt)).toThrow(
      'file type services are closed',
    );

    // Async withFile keeps the handle open across awaits and closes on resolve.
    let asyncServices: typeof syncServices | undefined;
    const asyncKind = await client.withFile(
      projection.binding,
      async (handle) => {
        asyncServices = handle.parserServices(decoded, projection);
        await Promise.resolve();
        return (asyncServices.esTreeNodeToTSNodeMap.get(stmt) as any)?.kind;
      },
    );
    expect(asyncKind).toBe(ts.SyntaxKind.VariableStatement);
    expect(() => asyncServices!.esTreeNodeToTSNodeMap.get(stmt)).toThrow(
      'file type services are closed',
    );

    // Async withFile also closes when the returned promise rejects.
    let rejectedServices: typeof syncServices | undefined;
    await expect(
      client.withFile(projection.binding, async (handle) => {
        rejectedServices = handle.parserServices(decoded, projection);
        await Promise.resolve();
        expect(
          (rejectedServices.esTreeNodeToTSNodeMap.get(stmt) as any)?.kind,
        ).toBe(ts.SyntaxKind.VariableStatement);
        throw new Error('async failure');
      }),
    ).rejects.toThrow('async failure');
    expect(() => rejectedServices!.esTreeNodeToTSNodeMap.get(stmt)).toThrow(
      'file type services are closed',
    );

    // Custom thenables with one-shot `then` getters must only have `then` read
    // once, matching native `await` / `Promise.resolve`.
    let thenReads = 0;
    let thenableServices: typeof syncServices | undefined;
    const thenableValue = await client.withFile(
      projection.binding,
      (handle) => {
        thenableServices = handle.parserServices(decoded, projection);
        return {
          get then() {
            thenReads++;
            if (thenReads > 1) {
              throw new Error('then was read twice');
            }
            return (resolveThenable: (value: number) => void) => {
              expect(
                (thenableServices!.esTreeNodeToTSNodeMap.get(stmt) as any)
                  ?.kind,
              ).toBe(ts.SyntaxKind.VariableStatement);
              resolveThenable(42);
            };
          },
        };
      },
    );
    expect(thenableValue).toBe(42);
    expect(thenReads).toBe(1);
    expect(() => thenableServices!.esTreeNodeToTSNodeMap.get(stmt)).toThrow(
      'file type services are closed',
    );

    // Union return types containing a Promise subclass normalize the pending
    // branch to a plain Promise at the type level (no stale subclass methods).
    class CancellableJob<V> extends Promise<V> {
      cancel() {
        return 'cancelled';
      }
    }
    const runMixed = (asyncBranch: boolean) =>
      client.withFile(projection.binding, (handle) => {
        const s = handle.parserServices(decoded, projection);
        if (asyncBranch) {
          return new CancellableJob<number>((resolveJob) => {
            resolveJob((s.esTreeNodeToTSNodeMap.get(stmt) as any)?.kind);
          });
        }
        return (s.esTreeNodeToTSNodeMap.get(stmt) as any)?.kind as number;
      });
    const mixedSync = runMixed(false);
    if (typeof mixedSync !== 'number') {
      // @ts-expect-error normalized union branch is a plain Promise, not CancellableJob
      mixedSync.cancel();
    }
    expect(mixedSync).toBe(ts.SyntaxKind.VariableStatement);
    const mixedAsync = runMixed(true);
    expect(mixedAsync).toBeInstanceOf(Promise);
    expect(await mixedAsync).toBe(ts.SyntaxKind.VariableStatement);

    // Explicit Symbol.dispose closes handles and maps immediately.
    const disposed = client.file(projection.binding);
    const disposedServices = disposed.parserServices(decoded, projection);
    expect(
      (disposedServices.esTreeNodeToTSNodeMap.get(stmt) as any)?.kind,
    ).toBe(ts.SyntaxKind.VariableStatement);
    disposed[Symbol.dispose]();
    expect(() => disposed.parserServices(decoded, projection)).toThrow(
      'file type services are closed',
    );
    expect(() => disposedServices.esTreeNodeToTSNodeMap.get(stmt)).toThrow(
      'file type services are closed',
    );

    // Collecting an unreferenced TypeServiceFile handle must NOT close live
    // syntax-only parserServices or detached node maps that callers still hold.
    const detachedSyntaxServices = (() => {
      const handle = syntaxClient.file(syntaxProjection.binding);
      return {
        weakHandle: new WeakRef(handle),
        services: handle.parserServices(syntaxDecoded, syntaxProjection),
      };
    })();
    const detachedSyntaxMap = (() => {
      const handle = syntaxClient.file(syntaxProjection.binding);
      return {
        weakHandle: new WeakRef(handle),
        map: handle.parserServices(syntaxDecoded, syntaxProjection)
          .esTreeNodeToTSNodeMap,
      };
    })();
    const detachedTypedMap = (() => {
      const handle = client.file(projection.binding);
      return {
        weakHandle: new WeakRef(handle),
        map: handle.parserServices(decoded, projection).esTreeNodeToTSNodeMap,
      };
    })();

    expect(
      (
        detachedSyntaxServices.services.esTreeNodeToTSNodeMap.get(
          syntaxStmt,
        ) as any
      )?.kind,
    ).toBe(ts.SyntaxKind.VariableStatement);
    expect((detachedSyntaxMap.map.get(syntaxStmt) as any)?.kind).toBe(
      ts.SyntaxKind.VariableStatement,
    );
    expect((detachedTypedMap.map.get(stmt) as any)?.kind).toBe(
      ts.SyntaxKind.VariableStatement,
    );

    await forceGC();

    expect(detachedSyntaxServices.weakHandle.deref()).toBeUndefined();
    expect(detachedSyntaxMap.weakHandle.deref()).toBeUndefined();
    expect(detachedTypedMap.weakHandle.deref()).toBeUndefined();

    expect(
      (
        detachedSyntaxServices.services.esTreeNodeToTSNodeMap.get(
          syntaxStmt,
        ) as any
      )?.kind,
    ).toBe(ts.SyntaxKind.VariableStatement);
    expect((detachedSyntaxMap.map.get(syntaxStmt) as any)?.kind).toBe(
      ts.SyntaxKind.VariableStatement,
    );
    expect((detachedTypedMap.map.get(stmt) as any)?.kind).toBe(
      ts.SyntaxKind.VariableStatement,
    );

    // Closing the owning client still invalidates all live detached maps.
    client[Symbol.dispose]();
    syntaxClient[Symbol.dispose]();
    expect(() => detachedTypedMap.map.get(stmt)).toThrow(
      'type service client is closed',
    );
    expect(() => detachedSyntaxMap.map.get(syntaxStmt)).toThrow(
      'type service client is closed',
    );
  } finally {
    client[Symbol.dispose]();
    syntaxClient[Symbol.dispose]();
    await host.close();
    await syntaxHost.close();
  }
});

test('decodes deep diagnostic message chains above 64 levels and rejects cyclic diagnostic packets', async () => {
  const depth = 70;
  let targetType = 'number';
  let sourceType = 'string';
  for (let i = 0; i < depth; i++) {
    targetType = `() => ${targetType}`;
    sourceType = `() => ${sourceType}`;
  }
  const sourceText = [
    `declare const deepSource: ${sourceType};`,
    `export const deepTarget: ${targetType} = deepSource;`,
    `export const shallowTarget: number = 'oops';`,
  ].join('\n');

  const dir = mkdtempSync(resolve(harness.temp, 'deep-diagnostics-'));
  const file = resolve(dir, 'deep.ts');
  const config = resolve(dir, 'tsconfig.json');
  writeFileSync(file, `${sourceText}\n`);
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'ES2022',
        module: 'ESNext',
        skipLibCheck: true,
      },
      include: ['deep.ts'],
    }),
  );

  const chainDepth = (message: string | ts.DiagnosticMessageChain): number => {
    let level = 1;
    let cur = typeof message === 'string' ? undefined : message;
    while (cur?.next && cur.next.length > 0) {
      level++;
      cur = cur.next[0];
    }
    return level;
  };

  const parsedConfig = ts.parseJsonConfigFileContent(
    JSON.parse(
      JSON.stringify({
        compilerOptions: {
          strict: true,
          target: 'ES2022',
          module: 'ESNext',
          skipLibCheck: true,
        },
        include: ['deep.ts'],
      }),
    ),
    ts.sys,
    dir,
  );
  const nativeProgram = ts.createProgram({
    rootNames: parsedConfig.fileNames,
    options: parsedConfig.options,
  });
  const nativeSf = nativeProgram.getSourceFile(file);
  expect(nativeSf).toBeDefined();
  const nativeDiags = nativeProgram.getSemanticDiagnostics(nativeSf);
  expect(nativeDiags).toHaveLength(2);
  expect(chainDepth(nativeDiags[0].messageText)).toBeGreaterThan(64);
  expect(chainDepth(nativeDiags[1].messageText)).toBe(1);

  const host = harness.createHost();
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const client = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 20 },
  );
  try {
    const sf = client.program.getSourceFile(file);
    expect(sf).toBeDefined();
    const diags = client.program.getSemanticDiagnostics(sf);
    expect(diags).toHaveLength(2);
    expect(diags.map((d) => d.code)).toEqual(nativeDiags.map((d) => d.code));
    expect(diags.map((d) => chainDepth(d.messageText))).toEqual(
      nativeDiags.map((d) => chainDepth(d.messageText)),
    );
    const leafMessage = (
      message: string | ts.DiagnosticMessageChain,
    ): string => {
      if (typeof message === 'string') return message;
      let cur = message;
      while (cur.next && cur.next.length > 0) cur = cur.next[0];
      return cur.messageText;
    };
    expect(leafMessage(diags[0].messageText)).toBe(
      leafMessage(nativeDiags[0].messageText),
    );
    expect(ts.flattenDiagnosticMessageText(diags[1].messageText, '\n')).toBe(
      ts.flattenDiagnosticMessageText(nativeDiags[1].messageText, '\n'),
    );
  } finally {
    client[Symbol.dispose]();
    await host.close();
  }

  // Crafted cyclic diagnostic packet (chain pointing to its own list at offset 12)
  // must still be rejected cleanly without overflowing the stack.
  const cyclic = Buffer.alloc(12 + 4 + 44 + 4);
  cyclic.writeUInt32LE(0x35535452, 0); // "RTS5"
  cyclic.writeUInt32LE(15, 4); // packetDiagnostics
  cyclic.writeUInt32LE(12, 8); // root list offset
  cyclic.writeUInt32LE(1, 12); // count = 1
  const emptyStrOffset = 12 + 4 + 44;
  cyclic.writeUInt32LE(emptyStrOffset, 16 + 20); // message text offset
  cyclic.writeUInt32LE(12, 16 + 32); // cyclic chain -> offset 12
  const cyclicReader = new TypePacketReader(cyclic);
  expect(() => cyclicReader.scalar('semanticDiagnostics')).toThrow(
    'invalid type service packet',
  );
  cyclicReader.close();

  // Crafted diamond-DAG diagnostic packet (chain and related sharing a child list)
  // must also be rejected to prevent exponential blowup.
  const childOffset = 12 + 4 + 44;
  const strOffset = childOffset + 4 + 44;
  const diamond = Buffer.alloc(strOffset + 4);
  diamond.writeUInt32LE(0x35535452, 0);
  diamond.writeUInt32LE(15, 4);
  diamond.writeUInt32LE(12, 8);
  diamond.writeUInt32LE(1, 12); // root count = 1
  diamond.writeUInt32LE(strOffset, 16 + 20);
  diamond.writeUInt32LE(childOffset, 16 + 32); // chain -> childOffset
  diamond.writeUInt32LE(childOffset, 16 + 36); // related -> same childOffset
  diamond.writeUInt32LE(1, childOffset); // child count = 1
  diamond.writeUInt32LE(strOffset, childOffset + 4 + 20);
  const diamondReader = new TypePacketReader(diamond);
  expect(() => diamondReader.scalar('semanticDiagnostics')).toThrow(
    'invalid type service packet',
  );
  diamondReader.close();
});
