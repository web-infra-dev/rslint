import { afterAll, beforeAll, describe, expect, test } from 'rstack/test';
import { mkdtempSync, writeFileSync } from 'node:fs';
import { relative, resolve } from 'node:path';
import tsParser from '@typescript-eslint/parser';
import ts from 'typescript';
import {
  TypeServiceClient,
  type TypeServiceDescriptor,
} from '../../../../rslint/src/languages/js/type-service.js';
import { lintFile } from '../../../../rslint/src/eslint-plugin/linter/ecma-language-plugin.js';
import { decodeFrame } from '../../../../rslint/src/languages/js/frame.js';
import type { ESTreeNode } from '../../../../rslint/src/languages/js/types.js';
import { VISITOR_KEYS } from '../../../../rslint/src/languages/js/visitor-keys.js';
import { cleanupTempDir } from '../../helpers/cleanup-temp-dir.js';
import { enums, Host, root, useTypeServiceHarness } from './harness.js';

const harness = useTypeServiceHarness();

function readMapping(host: Host, directory: string) {
  const file = resolve(directory, 'mapping.tsx'),
    config = resolve(directory, 'tsconfig.json');
  const longTypeName = '例'.repeat(110);
  const code = `\uFEFF\uFEFF"use strict";
import Default, { createStore as named } from "effector";
export function f<T>(plain: string, value = 1, ...rest: number[]) {
  return ((plain));
}
class C {
  #value = 1;
  m(x: string) { return x; }
  read() { return this.#value; }
}
declare const a: any;
a?.b.c;
f<string>("x");
const __name = "\\uD800" as const;
declare const unique: unique symbol;
const uniqueKeys = {[unique]: true};
const truth = true as const;
const surrogateKeys = {"\\uD800":1 as const,"\\uD801":"two" as const,"\\uFFFD":true as const};
const spreadKeys = {...surrogateKeys};
const multiline = { nested: { label: "中文" as const }, count: 1 };
function inferredScalar() { return '中文' as const; }
function inferredObject() { return { label: '中文' as const }; }
interface ${longTypeName} { value: number }
declare function multilineCall(
  { '\\é': escaped }: { '\\é': string },
  value: { label: "中文"; escaped: "\\é"; slash: "\\\\é" },
): { nested: { label: "中文" }; count: number };
/** Ordinary documentation. */
function docsTarget() {}
async function tokenContext() { await Promise.resolve(1); }
interface DocsBase {
  /** Base documentation linking to {@linkcode docsTarget | example}.
   * @param value Value to use.
   * @deprecated Replace this with {@link docsTarget}.
   */
  method(value: string): void;
}
class DocsChild implements DocsBase {
  /** @inheritDoc Additional detail. */
  method(value: string) {}
}
const docAssertion = /** @type {{ label: string }} */ ({ label: "temporary" });
const view = <Box />;
declare const Box: () => null;
declare namespace N { const value: number; interface Item<T> {} }
declare global { interface GlobalMarker {} }
type Query = typeof N.value;
namespace A.B {}
class Base<T> {}
class Derived extends Base<string> implements N.Item<number> {}
interface Child extends N.Item<boolean> {}
function meta() { return new.target; }
import.meta;
const first = 1, second = 2, third = 3;
first, second, third;
undefined;
await Promise.resolve(1);
`;
  writeFileSync(file, code);
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: false,
        strictNullChecks: true,
        target: 'ES2022',
        module: 'ESNext',
        moduleResolution: 'Bundler',
        jsx: 'react-native',
        moduleDetection: 'legacy',
        newLine: 'crlf',
        experimentalDecorators: true,
        types: [],
      },
      include: ['*.tsx'],
    }),
  );
  const descriptor = host.request({
    method: 'init',
    config,
  }) as TypeServiceDescriptor;
  const encoded = host.request({ method: 'project', file });
  const projection = {
    ...encoded,
    tree: Buffer.from(encoded.tree, 'base64'),
  };
  const sourceFileName: string = projection.binding.sourceFileName;
  const bytes = Buffer.from(encoded.frame, 'base64');
  const frame = decodeFrame(bytes);
  const client = new TypeServiceClient(
    enums,
    descriptor,
    (query) => Buffer.from(host.request({ method: 'query', query }), 'base64'),
    host.source,
    { maxFactsBytes: 1 << 16 },
  );
  const program = client.program as unknown as ts.Program;
  const beforeSource = program.getSourceFile(file);
  if (!beforeSource) throw new Error(`missing project source: ${file}`);
  const services = client
    .file(projection.binding)
    .parserServices(frame, projection, sourceFileName) as any;
  const upstream = tsParser.parseForESLint(code, {
    filePath: file,
    project: config,
    tsconfigRootDir: directory,
    range: true,
    loc: true,
  });
  const body = frame.ast.body as any[];
  const variable = (name: string) =>
    body
      .flatMap((node) => node.declarations ?? [])
      .find((declaration) => declaration.id.name === name);
  const expression = (type: string) =>
    body.find(
      (node) =>
        node.type === 'ExpressionStatement' && node.expression.type === type,
    ).expression;
  const fn = body.find(
    (node) => node.type === 'ExportNamedDeclaration',
  ).declaration;
  const nativeFn = services.esTreeNodeToTSNodeMap.get(fn);
  const referenceChecker = upstream.services.program!.getTypeChecker();
  const referenceBody = upstream.ast.body as any[];
  const referenceVariable = (name: string) =>
    referenceBody
      .flatMap((node) => node.declarations ?? [])
      .find((entry) => entry.id.name === name);
  const referenceNodes = upstream.services.esTreeNodeToTSNodeMap!;
  const referenceFn = referenceNodes.get(
    referenceBody.find((node) => node.type === 'ExportNamedDeclaration')
      .declaration,
  ) as ts.FunctionDeclaration;
  return {
    client,
    file,
    bytes,
    directory,
    longTypeName,
    beforeSource,
    frame,
    services,
    upstream,
    body,
    variable,
    expression,
    fn,
    nativeFn,
    referenceChecker,
    referenceVariable,
    referenceNodes,
    referenceFn,
    checker: services.program.getTypeChecker(),
  };
}

describe('TypeScript node maps and type queries', () => {
  let directory: string, host: Host, mapping: ReturnType<typeof readMapping>;
  beforeAll(() => {
    directory = mkdtempSync(
      resolve(root, 'packages/rslint-test-tools/.type-services-'),
    );
    host = harness.createHost();
    mapping = readMapping(host, directory);
  });
  afterAll(async () => {
    try {
      mapping?.client.close();
      await host?.close();
    } finally {
      await cleanupTempDir(directory);
    }
  });

  test('node maps, tokens and aliases match TypeScript', () => {
    const {
      beforeSource,
      frame,
      services,
      upstream,
      expression,
      fn,
      nativeFn,
    } = mapping;
    expect(services.esTreeNodeToTSNodeMap.has(frame.ast)).toBe(true);
    const collect = (ast: any, maps: any) => {
      const entries: any[] = [];
      const walk = (node: any, path: string) => {
        const native = maps.esTreeNodeToTSNodeMap.get(node);
        if (node.type === 'MetaProperty')
          expect(native.keywordToken).toBe(
            node.meta.name === 'new'
              ? ts.SyntaxKind.NewKeyword
              : ts.SyntaxKind.ImportKeyword,
          );
        const reverse = native && maps.tsNodeToESTreeNodeMap.get(native);
        entries.push([
          path,
          node.type,
          native ? ts.SyntaxKind[native.kind] : undefined,
          native?.getStart(),
          native?.end,
          reverse?.type,
          reverse?.range,
        ]);
        for (const key of VISITOR_KEYS[node.type] ?? []) {
          const value = node[key];
          if (Array.isArray(value))
            value.forEach(
              (child, index) =>
                child && walk(child, `${path}.${key}[${index}]`),
            );
          else if (value) walk(value, `${path}.${key}`);
        }
      };
      walk(ast, 'root');
      return entries;
    };
    expect(collect(frame.ast, services)).toEqual(
      collect(upstream.ast, upstream.services),
    );
    const nativeAwait = services.esTreeNodeToTSNodeMap.get(
      expression('AwaitExpression'),
    );
    expect(nativeAwait.getChildAt(1)).toBe(nativeAwait.expression);
    expect(nativeAwait.getChildAt(1).kind).toBe(ts.SyntaxKind.CallExpression);
    const awaitToken = nativeAwait.getChildAt(0);
    expect(awaitToken.kind).toBe(ts.SyntaxKind.AwaitKeyword);
    expect(awaitToken.parent).toBe(nativeAwait);
    expect(awaitToken.getText()).toBe('await');
    expect(nativeAwait.getFirstToken()).toBe(awaitToken);
    const call = nativeAwait.expression;
    const argumentsList = call.getChildAt(2);
    expect(argumentsList.kind).toBe(ts.SyntaxKind.SyntaxList);
    expect(argumentsList.parent).toBe(call);
    expect(argumentsList.getChildAt(0)).toBe(call.arguments[0]);
    expect(call.getChildCount()).toBe(4);
    expect(call.getChildren()).toBe(call.getChildren());
    expect(call.getLastToken()).toBe(call.getChildAt(3));
    expect(call.getLastToken().getSourceFile()).toBe(beforeSource);
    expect(services.tsNodeToESTreeNodeMap.has(argumentsList)).toBe(false);
    for (let i = 0; i < fn.params.length; i++)
      expect(services.tsNodeToESTreeNodeMap.get(nativeFn.parameters[i])).toBe(
        fn.params[i],
      );
    const returnExpression = nativeFn.body.statements[0].expression;
    expect(services.tsNodeToESTreeNodeMap.get(returnExpression).type).toBe(
      'Identifier',
    );
    const outerChain = expression('ChainExpression');
    const innerTS = services.esTreeNodeToTSNodeMap.get(
      outerChain.expression.object,
    );
    expect(services.tsNodeToESTreeNodeMap.get(innerTS).type).toBe(
      'ChainExpression',
    );
    const canonical = expression('SequenceExpression');
    const sequence = services.esTreeNodeToTSNodeMap.get(canonical);
    const members = canonical.expressions;
    canonical.expressions = [members[2]];
    expect(services.tsNodeToESTreeNodeMap.has(sequence.left)).toBe(true);
    const detached = services.tsNodeToESTreeNodeMap.get(sequence.left);
    expect(services.tsNodeToESTreeNodeMap.get(sequence.left)).toBe(detached);
    expect(services.esTreeNodeToTSNodeMap.get(detached)).toBe(sequence.left);
    const upstreamSequence = (upstream.ast.body as any[]).find(
      (node) =>
        node.type === 'ExpressionStatement' &&
        node.expression.type === 'SequenceExpression',
    ).expression;
    const upstreamNative =
      upstream.services.esTreeNodeToTSNodeMap!.get(upstreamSequence);
    const upstreamAlias = upstream.services.tsNodeToESTreeNodeMap!.get(
      (upstreamNative as ts.BinaryExpression).left,
    );
    expect([detached.range, detached.loc]).toEqual([
      upstreamAlias.range,
      upstreamAlias.loc,
    ]);
    expect(detached.type).toBe('SequenceExpression');
    expect(detached.expressions.map((node: any) => node.name)).toEqual([
      'first',
      'second',
    ]);
    detached.expressions = [members[0]];
    expect(detached.expressions).toEqual([members[0]]);
    canonical.expressions = members;
  });

  test('parser services expose compiler options and unmapped type annotations', () => {
    const { directory, file, services, upstream, fn, nativeFn } = mapping;
    const options = services.program.getCompilerOptions();
    const referenceOptions = upstream.services.program!.getCompilerOptions();
    for (const key of [
      'strict',
      'strictNullChecks',
      'jsx',
      'moduleDetection',
      'newLine',
      'target',
      'module',
      'moduleResolution',
      'experimentalDecorators',
    ])
      expect(options[key], key).toBe(referenceOptions[key]);
    expect(services.experimentalDecorators).toBe(true);
    expect(
      services.program.getSourceFile(resolve(directory, 'missing.ts')),
    ).toBeUndefined();
    expect(services.program.getSourceFile(file)).toBe(nativeFn.getSourceFile());
    const annotation = fn.params[0].typeAnnotation;
    expect(services.esTreeNodeToTSNodeMap.has(annotation)).toBe(false);
    expect(services.getSymbolAtLocation(annotation)).toBeUndefined();
    expect(services.getTypeAtLocation(annotation).flags).toBe(ts.TypeFlags.Any);
  });

  test('lintFile returns a file error when parser services fail to initialize', () => {
    const { file, bytes } = mapping;
    const result = lintFile(
      {
        filePath: file,
        frame: bytes,
        rules: {},
        collectFixes: false,
        suggestionsMode: 'off',
      },
      { plugins: [], rules: new Map() },
      () => {
        throw new Error('stale file binding');
      },
    );
    expect(result.parseError).toBe('parser services: stale file binding');
  });

  test('literal types and symbol lookup match TypeScript', () => {
    const {
      body,
      expression,
      variable,
      services,
      checker,
      referenceChecker,
      referenceNodes,
      referenceVariable,
    } = mapping;
    const declaration = variable('__name');
    const stringType = services.getTypeAtLocation(declaration.id);
    expect(stringType.value).toBe('\uD800');
    expect(stringType.flags & ts.TypeFlags.StringLiteral).toBeTruthy();

    const symbol = services.getSymbolAtLocation(declaration.id);
    expect([symbol.name, symbol.escapedName]).toEqual(['__name', '___name']);
    expect(services.esTreeNodeToTSNodeMap.get(declaration.id).escapedText).toBe(
      '___name',
    );
    const uniqueType = services.getTypeAtLocation(variable('unique').id);
    const uniqueProperty = services
      .getTypeAtLocation(variable('uniqueKeys').id)
      .getProperties()[0];
    expect(uniqueType.flags & ts.TypeFlags.UniqueESSymbol).toBeTruthy();
    expect(uniqueType.escapedName).toBe(uniqueProperty.escapedName);
    expect(uniqueType.escapedName).toMatch(/^__@unique@\d+$/);
    const booleanType = services.getTypeAtLocation(variable('truth').id);
    expect(booleanType.intrinsicName).toBe('true');
    expect(booleanType.isLiteral()).toBe(false);
    expect(booleanType.value).toBeUndefined();
    const keyDeclaration = variable('surrogateKeys');
    const keyType = services.getTypeAtLocation(keyDeclaration.id);
    const keyLocation = services.esTreeNodeToTSNodeMap.get(keyDeclaration.id);
    expect(
      ['\uD800', '\uD801', '\uFFFD'].map((name) =>
        checker.typeToString(
          checker.getTypeOfSymbolAtLocation(
            checker.getPropertyOfType(keyType, name),
            keyLocation,
          ),
        ),
      ),
    ).toEqual(['1', '"two"', 'true']);

    const referenceSpreadType = referenceChecker.getTypeAtLocation(
      referenceNodes.get(referenceVariable('spreadKeys').init),
    ) as ts.ObjectType;
    expect(
      referenceSpreadType.objectFlags & ts.ObjectFlags.ContainsSpread,
    ).toBe(ts.ObjectFlags.ContainsSpread);
    expect(
      services.getTypeAtLocation(variable('spreadKeys').init).objectFlags &
        ts.ObjectFlags.ContainsSpread,
    ).toBe(referenceSpreadType.objectFlags & ts.ObjectFlags.ContainsSpread);
    const truthLocation = services.esTreeNodeToTSNodeMap.get(
      variable('truth').id,
    );
    const referenceTruthLocation = referenceNodes.get(
      referenceVariable('truth').id,
    );
    for (const meaning of [
      ts.SymbolFlags.Value,
      ts.SymbolFlags.Type,
      ts.SymbolFlags.All,
    ]) {
      const actual = checker.resolveName(
        'truth',
        truthLocation,
        meaning,
        false,
      );
      const expected = referenceChecker.resolveName(
        'truth',
        referenceTruthLocation,
        meaning,
        false,
      );
      expect([actual?.name, actual?.flags]).toEqual([
        expected?.name,
        expected?.flags,
      ]);
    }
    const promise = checker.resolveName(
      'Promise',
      undefined,
      ts.SymbolFlags.All,
      false,
    );
    const referencePromise = referenceChecker.resolveName(
      'Promise',
      undefined,
      ts.SymbolFlags.All,
      false,
    );
    expect(referencePromise).toBeDefined();
    expect([promise?.name, promise?.flags]).toEqual([
      referencePromise?.name,
      referencePromise?.flags,
    ]);

    expect(
      services.getSymbolAtLocation(expression('Identifier')).declarations,
    ).toBeUndefined();
    const privateKey = body
      .find((node) => node.type === 'ClassDeclaration' && node.id.name === 'C')
      .body.body.find((node: any) => node.key.type === 'PrivateIdentifier').key;
    expect(services.esTreeNodeToTSNodeMap.get(privateKey).escapedText).toBe(
      '#value',
    );
    expect(services.getSymbolAtLocation(privateKey).name).toBe('#value');
  });

  test('generated type syntax and compiler node flags match TypeScript', () => {
    const {
      beforeSource,
      variable,
      services,
      checker,
      referenceChecker,
      referenceVariable,
      referenceNodes,
      referenceFn,
    } = mapping;
    const id = variable('surrogateKeys').id;
    const keyType = services.getTypeAtLocation(id);
    const keyLocation = services.esTreeNodeToTSNodeMap.get(id);
    // The node builder returns detached syntax. Its public kinds, child fields
    // and visitation must agree even when the compiler enums differ.
    const generatedShape = (n: ts.Node): unknown => {
      const children: unknown[] = [];
      n.forEachChild((child) => {
        children.push(generatedShape(child));
      });
      return [
        n.kind,
        // tsgo's node builder can reuse a source literal and its range.
        // For these leaves compare the literal's value as consumed by rules.
        ts.isStringLiteral(n) || ts.isNumericLiteral(n)
          ? n.text
          : [n.pos, n.end],
        ts.isIdentifier(n) ? n.escapedText : undefined,
        children,
      ];
    };
    const generated = checker.typeToTypeNode(
      keyType,
      keyLocation,
      ts.NodeBuilderFlags.InTypeAlias,
    )!;
    const referenceLocation = referenceNodes.get(
      referenceVariable('surrogateKeys').id,
    )!;
    expect(generatedShape(generated)).toEqual(
      generatedShape(
        referenceChecker.typeToTypeNode(
          referenceChecker.getTypeAtLocation(referenceLocation),
          referenceLocation,
          ts.NodeBuilderFlags.InTypeAlias,
        )!,
      ),
    );
    expect(generated.getSourceFile()).toBeUndefined();
    expect('referencedFiles' in generated).toBe(false);
    const contextualToken = (source: ts.SourceFile) => {
      const fn = source.statements.find(
        (node): node is ts.FunctionDeclaration =>
          ts.isFunctionDeclaration(node) && node.name?.text === 'tokenContext',
      )!;
      return (
        fn.body!.statements[0] as ts.ExpressionStatement
      ).expression.getFirstToken()!;
    };
    const contextFlags = contextualToken(referenceFn.getSourceFile()).flags;
    expect(contextFlags & ts.NodeFlags.ContextFlags).toBeTruthy();
    expect(contextualToken(beforeSource).flags).toBe(contextFlags);
    const flagsAndModifiers = (source: ts.SourceFile) => {
      const samples: number[] = [];
      const visit = (n: ts.Node) => {
        if (
          ts.isModuleDeclaration(n) ||
          ts.isMethodSignature(n) ||
          (ts.isPropertyAccessExpression(n) && n.questionDotToken) ||
          (ts.isFunctionDeclaration(n) && n.name?.text === 'docsTarget')
        )
          samples.push(n.flags);
        if (ts.isModuleDeclaration(n))
          samples.push(ts.getModifiers(n)?.length ?? 0);
        n.forEachChild(visit);
      };
      visit(source);
      return samples;
    };
    expect(flagsAndModifiers(beforeSource)).toEqual(
      flagsAndModifiers(referenceFn.getSourceFile()),
    );
  });

  test('type and signature formatting matches TypeScript', () => {
    const {
      beforeSource,
      variable,
      services,
      checker,
      nativeFn,
      referenceChecker,
      referenceNodes,
      referenceVariable,
      referenceFn,
      longTypeName,
    } = mapping;
    const uniqueType = services.getTypeAtLocation(variable('unique').id);
    const formatted = (
      typeChecker: ts.TypeChecker,
      declaration: ts.FunctionDeclaration,
      unique: ts.Type,
    ) => {
      const namespace = declaration
        .getSourceFile()
        .statements.filter(ts.isModuleDeclaration)
        .find((node) => node.name.text === 'N')!;
      const item = (namespace.body as ts.ModuleBlock).statements[1];
      const itemType = typeChecker.getTypeAtLocation(item);
      const rest = typeChecker.getTypeAtLocation(declaration.parameters[2]);
      const signature = typeChecker.getSignatureFromDeclaration(declaration)!;
      const none = ts.TypeFormatFlags.None;
      return [
        typeChecker.typeToString(unique),
        typeChecker.typeToString(unique, undefined, none),
        typeChecker.typeToString(rest, declaration, none),
        typeChecker.typeToString(
          rest,
          declaration,
          ts.TypeFormatFlags.WriteArrayAsGenericType,
        ),
        typeChecker.typeToString(itemType, item, none),
        typeChecker.typeToString(itemType, declaration, none),
        ...[ts.SignatureKind.Call, ts.SignatureKind.Construct].map((kind) =>
          typeChecker.signatureToString(
            signature,
            declaration,
            ts.TypeFormatFlags.WriteArrowStyleSignature,
            kind,
          ),
        ),
      ];
    };
    const expectedFormats = formatted(
      referenceChecker,
      referenceFn,
      referenceChecker.getTypeAtLocation(
        referenceNodes.get(referenceVariable('unique').id),
      ),
    );
    expect(formatted(checker, nativeFn, uniqueType)).toEqual(expectedFormats);
    // Multiline, quote-style, and non-ASCII truncation formatting use tsgo's
    // native printer; verify semantic options and UTF-8 boundary safety rather
    // than TypeScript 6's ASCII-escape and single-line newline quirks.
    const multilineDecl = beforeSource.statements.find(
      (node): node is ts.FunctionDeclaration =>
        ts.isFunctionDeclaration(node) && node.name?.text === 'multilineCall',
    )!;
    const multilineSig = checker.getSignatureFromDeclaration(multilineDecl)!;
    const multilineObj = checker.getTypeAtLocation(
      services.esTreeNodeToTSNodeMap.get(variable('multiline').id),
    );
    const longType = beforeSource.statements.find(
      (node): node is ts.InterfaceDeclaration =>
        ts.isInterfaceDeclaration(node) && node.name.text === longTypeName,
    )!;
    const longFormatted = checker.typeToString(
      checker.getTypeAtLocation(longType.name),
    );
    expect(longFormatted).not.toContain('\uFFFD');
    expect(longFormatted.startsWith('例'.repeat(20))).toBe(true);

    const multilineFlags =
      ts.TypeFormatFlags.MultilineObjectLiterals |
      ts.TypeFormatFlags.NoTruncation;
    const multilineObjText = checker.typeToString(
      multilineObj,
      multilineDecl,
      multilineFlags,
    );
    expect(multilineObjText).toContain('\n');
    expect(multilineObjText).toContain('nested:');
    expect(multilineObjText).toContain('label: "中文"');
    expect(multilineObjText).toContain('count: number');

    const arrowCall = checker.signatureToString(
      multilineSig,
      multilineDecl,
      multilineFlags | ts.TypeFormatFlags.WriteArrowStyleSignature,
      ts.SignatureKind.Call,
    );
    const arrowConstruct = checker.signatureToString(
      multilineSig,
      multilineDecl,
      multilineFlags | ts.TypeFormatFlags.WriteArrowStyleSignature,
      ts.SignatureKind.Construct,
    );
    expect(arrowCall).toContain('=>');
    expect(arrowConstruct).toBe(`new ${arrowCall}`);

    const singleQuoteSig = checker.signatureToString(
      multilineSig,
      multilineDecl,
      multilineFlags | ts.TypeFormatFlags.UseSingleQuotesForStringLiteralType,
      ts.SignatureKind.Call,
    );
    expect(singleQuoteSig).toContain("label: '中文'");
    expect(singleQuoteSig).toContain("escaped: 'é'");
  });

  test('symbol and signature documentation matches TypeScript', () => {
    const {
      beforeSource,
      variable,
      services,
      checker,
      referenceChecker,
      referenceNodes,
      referenceVariable,
      referenceFn,
    } = mapping;
    const readDocumentation = (
      typeChecker: ts.TypeChecker,
      source: ts.SourceFile,
    ) => {
      const owner = source.statements.find(
        (node): node is ts.ClassDeclaration =>
          ts.isClassDeclaration(node) && node.name?.text === 'DocsChild',
      )!;
      const method = owner.members[0] as ts.MethodDeclaration;
      const symbol = typeChecker.getSymbolAtLocation(method.name)!;
      const signature = typeChecker.getSignatureFromDeclaration(method)!;
      const documentation = symbol.getDocumentationComment(typeChecker);
      const tags = symbol.getJsDocTags(typeChecker);
      expect(symbol.getDocumentationComment(typeChecker)).toBe(documentation);
      expect(symbol.getJsDocTags(typeChecker)).toBe(tags);
      expect(documentation.map((part) => part.kind)).toContain('linkName');
      return [
        documentation,
        tags,
        signature.getDocumentationComment(typeChecker),
        signature.getJsDocTags(),
      ];
    };
    expect(readDocumentation(checker, beforeSource)).toEqual(
      readDocumentation(referenceChecker, referenceFn.getSourceFile()),
    );
    const symbolTags = (typeChecker: ts.TypeChecker, node: ts.Node) =>
      typeChecker.getSymbolAtLocation(node)!.getJsDocTags(typeChecker);
    expect(
      symbolTags(
        checker,
        services.esTreeNodeToTSNodeMap.get(variable('docAssertion').id),
      ),
    ).toEqual(
      symbolTags(
        referenceChecker,
        referenceNodes.get(referenceVariable('docAssertion').id),
      ),
    );
    // Regression: calls without a checker must not poison subsequent calls with a checker,
    // and cached arrays must be frozen against cross-rule mutation.
    const docsChild = beforeSource.statements.find(
      (node): node is ts.ClassDeclaration =>
        ts.isClassDeclaration(node) && node.name?.text === 'DocsChild',
    )!;
    const childMethod = docsChild.members[0] as ts.MethodDeclaration;
    const childSym = checker.getSymbolAtLocation(childMethod.name)!;
    const noCheckerDocs = childSym.getDocumentationComment(undefined);
    const withCheckerDocs = childSym.getDocumentationComment(checker);
    const noCheckerTags = childSym.getJsDocTags(undefined);
    const withCheckerTags = childSym.getJsDocTags(checker);
    expect(Object.isFrozen(noCheckerDocs)).toBe(true);
    expect(Object.isFrozen(withCheckerDocs)).toBe(true);
    expect(Object.isFrozen(noCheckerTags)).toBe(true);
    expect(Object.isFrozen(withCheckerTags)).toBe(true);
    expect(noCheckerDocs.map((p: any) => p.kind)).not.toContain('linkName');
    expect(withCheckerDocs.map((p: any) => p.kind)).toContain('linkName');
    expect(noCheckerTags.map((t: any) => t.name)).not.toContain('deprecated');
    expect(withCheckerTags.map((t: any) => t.name)).toContain('deprecated');
  });
});

test('public checker, Program, mapped types and declared index infos match TypeScript', async () => {
  const dir = mkdtempSync(resolve(harness.temp, 'public-'));
  const file = resolve(dir, 'input.ts'),
    config = resolve(dir, 'tsconfig.json');
  const code =
    [
      "export enum Values { N = 3, S = 'three', C = id(3) }",
      'export type Pair = readonly [first: string, second?: number];',
      'export function id<T extends { valueOf?<U,>(): number }>(input: T): T { return input; }',
      'export const result = id(42);',
      "export const incorrect: number = 'bad';",
      'export type Dictionary = { [name: string]: number };',
      "export function isString(input: unknown): input is string { return typeof input === 'string'; }",
      'export type Select<T> = T extends string ? true : false;',
      'export type Picked<T> = { [K in keyof T]: T[K] };',
      'export type UsedPicked = Picked<{ p: string }>;',
      'export interface Parent { inherited: string; }',
      'export interface Child extends Parent { own: number; [index: number]: boolean; }',
      'export class Box { read() { return 1; } }',
      'export { id as alias };',
      'export const record = { n: 1 }; export const {n} = record;',
      "export * from './external';",
    ].join('\n') + '\n';
  writeFileSync(file, code);
  writeFileSync(resolve(dir, 'external.ts'), 'export type External = string;');
  writeFileSync(
    config,
    JSON.stringify({
      compilerOptions: {
        strict: true,
        target: 'esnext',
        module: 'esnext',
        lib: ['es2022'],
        types: [],
      },
      files: ['input.ts'],
    }),
  );
  const host = harness.createHost();
  let closeOwner: (() => void) | undefined;
  try {
    const descriptor = host.request({ method: 'init', config });
    const owner = new TypeServiceClient(
      enums,
      descriptor,
      (query) =>
        Buffer.from(host.request({ method: 'query', query }), 'base64'),
      host.source,
      { maxFactsBytes: 1 << 18 },
    );
    closeOwner = () => owner.close();
    const parsed = ts.getParsedCommandLineOfConfigFile(
      config,
      {},
      { ...ts.sys, onUnRecoverableConfigFileDiagnostic() {} },
    )!;
    const reference = ts.createProgram(parsed.fileNames, parsed.options);
    const statement = (source: ts.SourceFile, name: string) => {
      const found = source.statements.find((node) => {
        const declaration = ts.isVariableStatement(node)
          ? node.declarationList.declarations[0]
          : (node as ts.DeclarationStatement);
        const candidate = ts.getNameOfDeclaration(declaration);
        return (
          candidate && ts.isIdentifier(candidate) && candidate.text === name
        );
      });
      expect(found, name).toBeDefined();
      return found!;
    };
    const exercise = (p: ts.Program) => {
      const c = p.getTypeChecker(),
        src = p.getSourceFile(file)!;
      const enm = statement(src, 'Values');
      const pair = statement(src, 'Pair');
      const fn = statement(src, 'id') as ts.FunctionDeclaration;
      const dict = statement(src, 'Dictionary');
      const select = statement(src, 'Select');
      const picked = statement(src, 'Picked') as ts.TypeAliasDeclaration;
      const usedPicked = statement(
        src,
        'UsedPicked',
      ) as ts.TypeAliasDeclaration;
      const childIface = statement(src, 'Child') as ts.InterfaceDeclaration;
      const param = fn.parameters[0];
      const symbol = c.getSymbolAtLocation(param.name)!;
      const fileSymbol = c.getSymbolAtLocation(src)!;
      const t = c.getTypeFromTypeNode(
        (pair as ts.TypeAliasDeclaration).type,
      ) as ts.TupleTypeReference;
      const tuple = t.target;
      const cond = c.getTypeFromTypeNode(
        (select as ts.TypeAliasDeclaration).type,
      ) as ts.ConditionalType;
      const mapped = c.getTypeFromTypeNode(picked.type) as any;
      const usedMapped = c.getTypeFromTypeNode(usedPicked.type);
      const childType = c.getDeclaredTypeOfSymbol(
        c.getSymbolAtLocation(childIface.name)!,
      ) as ts.InterfaceTypeWithDeclaredMembers;
      const exportSpec = (
        (
          src.statements.find(
            (node) =>
              ts.isExportDeclaration(node) &&
              node.exportClause &&
              ts.isNamedExports(node.exportClause) &&
              node.exportClause.elements.some(
                (element) => element.name.text === 'alias',
              ),
          ) as ts.ExportDeclaration
        ).exportClause as ts.NamedExports
      ).elements[0];
      const format = (value: ts.Type) => c.typeToString(value);
      const name = (node: ts.Node | undefined) =>
        node && ts.SyntaxKind[node.kind];

      const beforeTemplate = mapped.templateType;
      (c as any).getTypeOfPropertyOfType(usedMapped, 'p');
      const afterTemplate = mapped.templateType;

      const beforeDeclaredIndex = childType.declaredIndexInfos;
      const allIndexInfos = c.getIndexInfosOfType(childType);
      const afterDeclaredIndex = childType.declaredIndexInfos;

      return {
        primitive: [
          c.getAnyType(),
          c.getUnknownType(),
          c.getStringType(),
          c.getNumberType(),
          c.getBigIntType(),
          c.getBooleanType(),
          c.getESSymbolType(),
          c.getNonPrimitiveType(),
          c.getNeverType(),
          c.getNullType(),
          c.getUndefinedType(),
          c.getVoidType(),
        ].map(format),
        interned: c.getStringType() === c.getStringType(),
        constants: (enm as ts.EnumDeclaration).members.map((member) =>
          c.getConstantValue(member),
        ),
        rootSymbols: c.getRootSymbols(symbol).map((s) => s.name),
        merged: c.getMergedSymbol(symbol) === symbol,
        symbolPredicates: [
          c.isArgumentsSymbol(symbol),
          c.isUndefinedSymbol(symbol),
        ],
        localExport: c.getExportSpecifierLocalTargetSymbol(exportSpec)?.name,
        moduleMember: c.tryGetMemberInModuleExports('alias', fileSymbol)?.name,
        rawExports: [...(fileSymbol.exports?.keys() ?? [])].sort(),
        exports: c
          .getExportsOfModule(fileSymbol)
          .map((s) => s.name)
          .sort(),
        symbolString: c.symbolToString(symbol, fn),
        shape: [
          'types' in c.getStringType(),
          'target' in c.getStringType(),
          'thisType' in t,
          tuple.thisType && format(tuple.thisType),
          'types' in c.getBooleanType(),
          'root' in cond,
          'root' in c.getStringType(),
          'elementFlags' in tuple,
          'elementFlags' in t,
          'typeParameter' in mapped,
          'typeParameter' in c.getStringType(),
        ],
        mapped: [
          mapped.typeParameter?.symbol?.name,
          format(mapped.constraintType),
          beforeTemplate === undefined,
          afterTemplate?.indexType === mapped.typeParameter,
        ],
        declaredIndex: [
          beforeDeclaredIndex === undefined,
          afterDeclaredIndex?.map((info) => [
            allIndexInfos.includes(info),
            format(info.keyType),
            format(info.type),
          ]),
        ],
        tuple: [
          tuple.minLength,
          tuple.fixedLength,
          tuple.hasRestElement,
          tuple.elementFlags,
          tuple.labeledElementDeclarations?.map(name),
        ],
        conditional: [
          name(cond.root.node),
          cond.root.isDistributive,
          format(cond.root.checkType),
          format(cond.root.extendsType),
        ],
        roots: p.getRootFileNames(),
        sourceByPath:
          p.getSourceFileByPath(
            (src as ts.SourceFile & { path: ts.Path }).path,
          ) === src,
        diagnostics: {
          semantic: p
            .getSemanticDiagnostics(src)
            .map((d) => [d.code, d.start, d.length, d.file === src]),
          syntactic: p.getSyntacticDiagnostics(src).length,
        },
        dictIndex: format(
          c.getIndexInfoOfType(
            c.getTypeFromTypeNode((dict as ts.TypeAliasDeclaration).type),
            ts.IndexKind.String,
          )!.type,
        ),
      };
    };
    const actual = owner.program as unknown as ts.Program;
    expect(exercise(actual)).toEqual(exercise(reference));

    const checker = owner.program.getTypeChecker();
    const src = owner.program.getSourceFile(file) as ts.SourceFile;
    const stringType = checker.getStringType();
    expect(Reflect.set(stringType, 'flags', 0)).toBe(false);
    expect(() => owner.program.emit()).toThrow('Program.emit');
    expect(() =>
      src.update('', { span: { start: 0, length: 0 }, newLength: 0 }),
    ).toThrow('SourceFile.update');
    const dictionary = checker.getTypeFromTypeNode(
      (statement(src, 'Dictionary') as ts.TypeAliasDeclaration).type,
    )!;
    const index = checker.getIndexInfoOfType(dictionary, ts.IndexKind.String)!;
    expect(Reflect.set(index, 'type', stringType)).toBe(false);
  } finally {
    closeOwner?.();
    await host.close();
  }
});
