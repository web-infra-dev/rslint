/**
 * Visitor-keys table — sourced from oxc-parser's `visitorKeys` export
 * (replaces the former `@typescript-eslint/visitor-keys` /
 * `eslint-visitor-keys` direct deps). The child ORDER inside each array
 * is load-bearing: it sets ESLint-compatible traversal / listener-firing
 * order, so a drift (a reordered, dropped, or added key — or a whole
 * node type appearing/disappearing) must fail CI rather than silently
 * change rule behavior.
 *
 * The drift guard below pins a COMPLETE hardcoded snapshot of the table
 * — every node type → its exact ordered key array, captured from
 * oxc-parser 0.132 — and asserts the live export deep-equals it. The
 * snapshot is a literal embedded here ON PURPOSE: it is NOT re-derived
 * from `oxcVisitorKeys` at runtime (that would be a tautology). An oxc
 * upgrade that reorders/drops/adds any key, or changes the set of node
 * types, fails this test and forces a deliberate review + snapshot bump.
 */
import { describe, test, expect } from '@rstest/core';

import { VISITOR_KEYS, getVisitorKeys } from '../../src/ast/visitor-keys.js';

// Full snapshot of oxc-parser 0.132's `visitorKeys`. Hardcoded literal —
// do not regenerate from the live export. To intentionally adopt an oxc
// change, regenerate this block and review the diff:
//   node -e "import('oxc-parser').then(({visitorKeys:v})=>{for(const
//     k of Object.keys(v).sort())console.log(JSON.stringify(k)+': '+
//     JSON.stringify(v[k])+',')})" --input-type=module
const VISITOR_KEYS_SNAPSHOT: Record<string, readonly string[]> = {
  AccessorProperty: ['decorators', 'key', 'typeAnnotation', 'value'],
  ArrayExpression: ['elements'],
  ArrayPattern: ['decorators', 'elements', 'typeAnnotation'],
  ArrowFunctionExpression: ['typeParameters', 'params', 'returnType', 'body'],
  AssignmentExpression: ['left', 'right'],
  AssignmentPattern: ['decorators', 'left', 'right', 'typeAnnotation'],
  AwaitExpression: ['argument'],
  BinaryExpression: ['left', 'right'],
  BlockStatement: ['body'],
  BreakStatement: ['label'],
  CallExpression: ['callee', 'typeArguments', 'arguments'],
  CatchClause: ['param', 'body'],
  ChainExpression: ['expression'],
  ClassBody: ['body'],
  ClassDeclaration: [
    'decorators',
    'id',
    'typeParameters',
    'superClass',
    'superTypeArguments',
    'implements',
    'body',
  ],
  ClassExpression: [
    'decorators',
    'id',
    'typeParameters',
    'superClass',
    'superTypeArguments',
    'implements',
    'body',
  ],
  ConditionalExpression: ['test', 'consequent', 'alternate'],
  ContinueStatement: ['label'],
  DebuggerStatement: [],
  Decorator: ['expression'],
  DoWhileStatement: ['body', 'test'],
  EmptyStatement: [],
  ExportAllDeclaration: ['exported', 'source', 'attributes'],
  ExportDefaultDeclaration: ['declaration'],
  ExportNamedDeclaration: ['declaration', 'specifiers', 'source', 'attributes'],
  ExportSpecifier: ['local', 'exported'],
  ExpressionStatement: ['expression'],
  ForInStatement: ['left', 'right', 'body'],
  ForOfStatement: ['left', 'right', 'body'],
  ForStatement: ['init', 'test', 'update', 'body'],
  FunctionDeclaration: ['id', 'typeParameters', 'params', 'returnType', 'body'],
  FunctionExpression: ['id', 'typeParameters', 'params', 'returnType', 'body'],
  Identifier: ['decorators', 'typeAnnotation'],
  IfStatement: ['test', 'consequent', 'alternate'],
  ImportAttribute: ['key', 'value'],
  ImportDeclaration: ['specifiers', 'source', 'attributes'],
  ImportDefaultSpecifier: ['local'],
  ImportExpression: ['source', 'options'],
  ImportNamespaceSpecifier: ['local'],
  ImportSpecifier: ['imported', 'local'],
  JSXAttribute: ['name', 'value'],
  JSXClosingElement: ['name'],
  JSXClosingFragment: [],
  JSXElement: ['openingElement', 'children', 'closingElement'],
  JSXEmptyExpression: [],
  JSXExpressionContainer: ['expression'],
  JSXFragment: ['openingFragment', 'children', 'closingFragment'],
  JSXIdentifier: [],
  JSXMemberExpression: ['object', 'property'],
  JSXNamespacedName: ['namespace', 'name'],
  JSXOpeningElement: ['name', 'typeArguments', 'attributes'],
  JSXOpeningFragment: [],
  JSXSpreadAttribute: ['argument'],
  JSXSpreadChild: ['expression'],
  JSXText: [],
  LabeledStatement: ['label', 'body'],
  Literal: [],
  LogicalExpression: ['left', 'right'],
  MemberExpression: ['object', 'property'],
  MetaProperty: ['meta', 'property'],
  MethodDefinition: ['decorators', 'key', 'value'],
  NewExpression: ['callee', 'typeArguments', 'arguments'],
  ObjectExpression: ['properties'],
  ObjectPattern: ['decorators', 'properties', 'typeAnnotation'],
  ParenthesizedExpression: ['expression'],
  PrivateIdentifier: [],
  Program: ['body'],
  Property: ['key', 'value'],
  PropertyDefinition: ['decorators', 'key', 'typeAnnotation', 'value'],
  RestElement: ['decorators', 'argument', 'typeAnnotation'],
  ReturnStatement: ['argument'],
  SequenceExpression: ['expressions'],
  SpreadElement: ['argument'],
  StaticBlock: ['body'],
  Super: [],
  SwitchCase: ['test', 'consequent'],
  SwitchStatement: ['discriminant', 'cases'],
  TSAbstractAccessorProperty: ['decorators', 'key', 'typeAnnotation'],
  TSAbstractMethodDefinition: ['key', 'value'],
  TSAbstractPropertyDefinition: ['decorators', 'key', 'typeAnnotation'],
  TSAnyKeyword: [],
  TSArrayType: ['elementType'],
  TSAsExpression: ['expression', 'typeAnnotation'],
  TSBigIntKeyword: [],
  TSBooleanKeyword: [],
  TSCallSignatureDeclaration: ['typeParameters', 'params', 'returnType'],
  TSClassImplements: ['expression', 'typeArguments'],
  TSConditionalType: ['checkType', 'extendsType', 'trueType', 'falseType'],
  TSConstructSignatureDeclaration: ['typeParameters', 'params', 'returnType'],
  TSConstructorType: ['typeParameters', 'params', 'returnType'],
  TSDeclareFunction: ['id', 'typeParameters', 'params', 'returnType', 'body'],
  TSEmptyBodyFunctionExpression: [
    'id',
    'typeParameters',
    'params',
    'returnType',
  ],
  TSEnumBody: ['members'],
  TSEnumDeclaration: ['id', 'body'],
  TSEnumMember: ['id', 'initializer'],
  TSExportAssignment: ['expression'],
  TSExternalModuleReference: ['expression'],
  TSFunctionType: ['typeParameters', 'params', 'returnType'],
  TSImportEqualsDeclaration: ['id', 'moduleReference'],
  TSImportType: ['source', 'options', 'qualifier', 'typeArguments'],
  TSIndexSignature: ['parameters', 'typeAnnotation'],
  TSIndexedAccessType: ['objectType', 'indexType'],
  TSInferType: ['typeParameter'],
  TSInstantiationExpression: ['expression', 'typeArguments'],
  TSInterfaceBody: ['body'],
  TSInterfaceDeclaration: ['id', 'typeParameters', 'extends', 'body'],
  TSInterfaceHeritage: ['expression', 'typeArguments'],
  TSIntersectionType: ['types'],
  TSIntrinsicKeyword: [],
  TSJSDocNonNullableType: ['typeAnnotation'],
  TSJSDocNullableType: ['typeAnnotation'],
  TSJSDocUnknownType: [],
  TSLiteralType: ['literal'],
  TSMappedType: ['key', 'constraint', 'nameType', 'typeAnnotation'],
  TSMethodSignature: ['key', 'typeParameters', 'params', 'returnType'],
  TSModuleBlock: ['body'],
  TSModuleDeclaration: ['id', 'body'],
  TSNamedTupleMember: ['label', 'elementType'],
  TSNamespaceExportDeclaration: ['id'],
  TSNeverKeyword: [],
  TSNonNullExpression: ['expression'],
  TSNullKeyword: [],
  TSNumberKeyword: [],
  TSObjectKeyword: [],
  TSOptionalType: ['typeAnnotation'],
  TSParameterProperty: ['decorators', 'parameter'],
  TSParenthesizedType: ['typeAnnotation'],
  TSPropertySignature: ['key', 'typeAnnotation'],
  TSQualifiedName: ['left', 'right'],
  TSRestType: ['typeAnnotation'],
  TSSatisfiesExpression: ['expression', 'typeAnnotation'],
  TSStringKeyword: [],
  TSSymbolKeyword: [],
  TSTemplateLiteralType: ['quasis', 'types'],
  TSThisType: [],
  TSTupleType: ['elementTypes'],
  TSTypeAliasDeclaration: ['id', 'typeParameters', 'typeAnnotation'],
  TSTypeAnnotation: ['typeAnnotation'],
  TSTypeAssertion: ['typeAnnotation', 'expression'],
  TSTypeLiteral: ['members'],
  TSTypeOperator: ['typeAnnotation'],
  TSTypeParameter: ['name', 'constraint', 'default'],
  TSTypeParameterDeclaration: ['params'],
  TSTypeParameterInstantiation: ['params'],
  TSTypePredicate: ['parameterName', 'typeAnnotation'],
  TSTypeQuery: ['exprName', 'typeArguments'],
  TSTypeReference: ['typeName', 'typeArguments'],
  TSUndefinedKeyword: [],
  TSUnionType: ['types'],
  TSUnknownKeyword: [],
  TSVoidKeyword: [],
  TaggedTemplateExpression: ['tag', 'typeArguments', 'quasi'],
  TemplateElement: [],
  TemplateLiteral: ['quasis', 'expressions'],
  ThisExpression: [],
  ThrowStatement: ['argument'],
  TryStatement: ['block', 'handler', 'finalizer'],
  UnaryExpression: ['argument'],
  UpdateExpression: ['argument'],
  V8IntrinsicExpression: ['name', 'arguments'],
  VariableDeclaration: ['declarations'],
  VariableDeclarator: ['id', 'init'],
  WhileStatement: ['test', 'body'],
  WithStatement: ['object', 'body'],
  YieldExpression: ['argument'],
};

describe('visitor-keys table (from oxc-parser)', () => {
  test('live export deep-equals the complete hardcoded snapshot', () => {
    // `toEqual` compares the full object: every node type must be
    // present (no silent additions/removals) and every key array must
    // match in EXACT order (no silent reordering). This is the drift
    // guard — an oxc upgrade that changes the table fails here.
    expect({ ...VISITOR_KEYS }).toEqual(VISITOR_KEYS_SNAPSHOT);
  });

  test('exact node-type count (locks additions/removals numerically)', () => {
    // Redundant with the deep-equal above, but pins the count as a
    // fast, unambiguous signal when the snapshot diff is large.
    expect(Object.keys(VISITOR_KEYS)).toHaveLength(
      Object.keys(VISITOR_KEYS_SNAPSHOT).length,
    );
    expect(Object.keys(VISITOR_KEYS)).toHaveLength(165);
  });

  test('getVisitorKeys returns the table entry for known types', () => {
    expect(getVisitorKeys({ type: 'JSXElement' })).toEqual(
      VISITOR_KEYS['JSXElement'],
    );
  });

  test('getVisitorKeys fallback drops non-child keys for unknown types', () => {
    const keys = getVisitorKeys({
      type: 'SomeSyntheticNodeType',
      // child-ish fields
      foo: {},
      bar: [],
      // non-child fields that must be dropped
      parent: {},
      leadingComments: [],
      trailingComments: [],
    });
    expect(keys).toContain('foo');
    expect(keys).toContain('bar');
    expect(keys).not.toContain('parent');
    expect(keys).not.toContain('leadingComments');
    expect(keys).not.toContain('trailingComments');
  });
});
