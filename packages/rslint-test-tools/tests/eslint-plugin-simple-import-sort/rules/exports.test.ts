/**
 * @fileoverview Alignment tests for eslint-plugin-simple-import-sort 'exports' rule.
 *
 * Ported verbatim from eslint-plugin-simple-import-sort v13.0.0:
 *   test/exports.test.js
 *
 * Upstream structure: three suites driven by ESLint's RuleTester —
 *   - `baseTests`        run 3x there (JS / Flow / TS parsers); the cases are
 *                       identical, so they are ported here ONCE.
 *   - `typescriptTests` TypeScript-specific cases — merged into run().
 *   - `flowTests`        Flow-specific cases — those that parse under ts-go are
 *                       merged into run(); the rest are isolated (see KNOWN GAPS).
 *
 * Transformations applied per the porting spec:
 *  - The upstream `input\`…\`` template tag is evaluated to its real multi-line
 *    string (strip 10-space + `|` prefix; drop the leading/trailing newline).
 *  - Each invalid case's `output: (actual) => expect(actual).toMatchInlineSnapshot(\`…\`)`
 *    is evaluated to the real expected fixed source (pipes stripped; the helper's
 *    visible-char hacks decoded: `→` → TAB, `<CR>` → CR).
 *  - Strings are emitted as escaped literals for byte-exact fidelity (trailing
 *    spaces / tabs / CR / unicode are load-bearing).
 *  - `parserOptions` dropped (rslint resolves via tsconfig); `options` kept verbatim.
 *  - The single message id is `sort` ("Run autofix to sort these exports!"),
 *    no data interpolation.
 *
 * Every upstream invalid case pins BOTH `output` (the fixed source) AND `errors`
 * (a count, or — for one positional case — `{ messageId, line, column, endLine,
 * endColumn }`). None are output-only.
 *
 * Cases that surface a real rslint<->upstream gap are NOT deleted or altered: they
 * are listed in the KNOWN GAPS block at the bottom, each annotated with the upstream
 * expectation vs. what rslint does.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('exports', null as never, {
  valid: [
    // Simple cases.
    "export {a} from \"a\"",
    "export {a,b} from \"a\"",
    "export {} from \"a\"",
    "export {    } from \"a\"",
    "export * as a from \"a\"",
    "export var one = 1;",
    "export let two = 2;",
    "export const three = 3;",
    "export function f() {}",
    "export class C {}",
    "export { a, b as c }; var a, b;",
    "export default whatever;",
    // Sorted alphabetically.
    "export {x1} from \"a\";\nexport {x2} from \"b\"",
    // Opt-out.
    "// eslint-disable-next-line\nexport {x2} from \"b\"\nexport {x1} from \"a\";",
    // Whitespace before comment at last specifier should stay.
    "export {\n  a, // a\n  b // b\n} from \"specifiers-comment-space\"\nexport {\n  c, // c\n  d, // d\n} from \"specifiers-comment-space-2\"",
    // Accidental trailing spaces doesn’t produce a sorting error.
    "export {a} from \"a\"    \nexport {b} from \"b\";    \nexport {c} from \"c\";  /* comment */  ",
    // Commenting out an export doesn’t produce a sorting error.
    "export {a} from \"a\"\n// export {b} from \"b\";\nexport {c} from \"c\";",
    // Simple cases.
    "export type T = string;",
    "export type { T, U as V }; type T = 1; type U = 1;",
    // type specifiers.
    "export { a, type b, c, type d } from \"a\"",
    "export { a, type b, c, type d }",
    // Sorted alphabetically.
    "export type {x1} from \"a\";\nexport type {x2} from \"b\"",
    // Simple cases.
    "export type T = string;",
    "export type { T, U as V }; type T = 1; type U = 1;",
    // Sorted alphabetically.
    "export type {x1} from \"a\";\nexport type {x2} from \"b\"",
  ],

  invalid: [
    // Sorting alphabetically.
    {
      code: "export {x2} from \"b\"\nexport {x1} from \"a\";",
      output: "export {x1} from \"a\";\nexport {x2} from \"b\"",
      errors: 1,
    },
    // Using comments for grouping.
    {
      code: "export * from \"g\"\nexport * from \"f\";\n// Group 2\nexport * from \"e\"\nexport * from \"d\"\n/* Group 3 */\n\nexport * from \"c\"\n\n\nexport * from \"b\"\n\n\n /* Group 4\n */\n\n   export * from \"a\"",
      output: "export * from \"f\";\nexport * from \"g\"\n// Group 2\nexport * from \"d\"\nexport * from \"e\"\n/* Group 3 */\n\nexport * from \"b\"\nexport * from \"c\"\n\n\n /* Group 4\n */\n\n   export * from \"a\"",
      errors: 3,
    },
    // Sorting specifiers.
    // In `a as c`, the “c” is used since that’s the “stable” name, while the
    // internal `a` name can change at any time without affecting the module
    // interface. In other words, this is “backwards” compared to
    // `import {a as c} from "x"`.
    {
      code: "export { d, a as c, a as b2, b, a } from \"specifiers\"",
      output: "export { a,b, a as b2, a as c, d } from \"specifiers\"",
      errors: 1,
    },
    {
      code: "export { d, a as c, a as b2, b, a, }; var d, a, b;",
      output: "export { a,b, a as b2, a as c, d,  }; var d, a, b;",
      errors: 1,
    },
    // Comments on the same line as something else don’t count for grouping.
    {
      code: "export * from \"g\"\n/* f1 */export * from \"f\"; // f2\nexport * from \"e\" /* d\n */\nexport * from \"d\"\nexport * from \"c\" /*\n b */ export * from \"b\"\n /* a\n */ export * from \"a\"",
      output: " /* a\n */ export * from \"a\"\n/*\n b */ export * from \"b\"\nexport * from \"c\" \n/* d\n */\nexport * from \"d\"\nexport * from \"e\" \n/* f1 */export * from \"f\"; // f2\nexport * from \"g\"",
      errors: 1,
    },
    // Sorting with lots of comments.
    {
      code: "/*1*//*2*/export/*3*/*/*4*/as/*as*/foo/*foo1*//*foo2*/from/*6*/\"specifiers-lots-of-comments\"/*7*//*8*/\nexport { // start\n  /* c1 */ c /* c2 */, // c3\n  // b1\n\n  b as /* b2 */ renamed\n  , /* b3 */ /* a1\n  */ a /* not-a\n  */ // comment at end\n} from \"specifiers-lots-of-comments-multiline\";\nexport {\n  e,\n  d, /* d */ /* not-d\n  */ // comment at end after trailing comma\n};\nvar e, d;",
      output: "/*1*//*2*/export/*3*/*/*4*/as/*as*/foo/*foo1*//*foo2*/from/*6*/\"specifiers-lots-of-comments\"/*7*//*8*/\nexport { // start\n/* a1\n  */ a, \n  /* c1 */ c /* c2 */, // c3\n  // b1\n  b as /* b2 */ renamed\n  /* b3 */ /* not-a\n  */ // comment at end\n} from \"specifiers-lots-of-comments-multiline\";\nexport {\n  d, /* d */   e,\n/* not-d\n  */ // comment at end after trailing comma\n};\nvar e, d;",
      errors: 2,
    },
    // Collapse blank lines inside export statements.
    {
      code: "export\n\n// export\n\n/* default */\n\n\n\n// default\n\n {\n\n  // c\n\n  c /*c*/,\n\n  /* b\n   */\n\n  b // b\n  ,\n\n  // a1\n\n  // a2\n\n  a\n\n  // a3\n\n  as\n\n  // a4\n\n  d\n\n  // a5\n\n  , // a6\n\n  // last\n\n}\n\n// from1\n\nfrom\n\n// from2\n\n\"c\"\n\n// final\n\n;",
      output: "export\n// export\n/* default */\n// default\n {\n  /* b\n   */\n  b // b\n  ,\n  // c\n  c /*c*/,\n  // a1\n  // a2\n  a\n  // a3\n  as\n  // a4\n  d\n  // a5\n  , // a6\n  // last\n}\n// from1\nfrom\n// from2\n\"c\"\n// final\n;",
      errors: 1,
    },
    // Collapse blank lines inside empty specifier list.
    {
      code: "export {\n\n    } from \"specifiers-empty\"",
      output: "export {\n    } from \"specifiers-empty\"",
      errors: 1,
    },
    // Do not collapse empty lines inside export code.
    {
      code: "export const options = {\n\n    a: 1,\n\n    b: 2\n    }, a = 1\nexport {options as options2, a as a2}",
      output: "export const options = {\n\n    a: 1,\n\n    b: 2\n    }, a = 1\nexport {a as a2,options as options2}",
      errors: 1,
    },
    // Preserve indentation (for `<script>` tags).
    {
      code: "  export {e} from \"e\"\n  export {\n    b4, b3,\n    b2\n  } from \"b\";\n  /* a */ export {a} from \"a\"; export {c} from \"c\"\n  \n    export {d} from \"d\"",
      output: "  /* a */ export {a} from \"a\"; \n  export {\n    b2,\nb3,\n    b4  } from \"b\";\nexport {c} from \"c\"\n    export {d} from \"d\"\n  export {e} from \"e\"",
      errors: 1,
    },
    // Handling last semicolon.
    {
      code: "export {x2} from \"b\"\nexport {x1} from \"a\"\n\n;[].forEach()",
      output: "export {x1} from \"a\"\nexport {x2} from \"b\"\n\n;[].forEach()",
      errors: 1,
    },
    // Handling `as default` (issue #58).
    {
      code: "export { something, something as default } from './something'",
      output: "export { something as default,something } from './something'",
      errors: 1,
    },
    // Tricky `default` cases.
    {
      code: "export {default as default, default as def, default as fault} from \"b\"",
      output: "export {default as def, default as default, default as fault} from \"b\"",
      errors: 1,
    },
    // Test messageId, lines and columns.
    {
      code: "// before\n/* also\nbefore */ export * from \"b\";\nexport * from \"a\"; /*a*/ /* comment\nafter */ // after",
      output: "// before\n/* also\nbefore */ export * from \"a\"; /*a*/ \nexport * from \"b\";/* comment\nafter */ // after",
      errors: [
        { messageId: "sort", line: 3, column: 11, endLine: 4, endColumn: 26 },
      ],
    },
    // https://github.com/5monkeys/djedi-cms/blob/133a24a9ddcc0f133aaac6bd2f13db4d6dfe2dce/djedi-react/src/index.js
    {
      code: "export { default as djedi } from \"./djedi\";\nexport { default as Node, NodeContext } from \"./Node\";\nexport { default as ForceNodes } from \"./ForceNodes\";\nexport { default as md } from \"dedent-js\";",
      output: "export { default as djedi } from \"./djedi\";\nexport { default as ForceNodes } from \"./ForceNodes\";\nexport { default as Node, NodeContext } from \"./Node\";\nexport { default as md } from \"dedent-js\";",
      errors: 1,
    },
    // https://gitlab.com/appsemble/appsemble/-/blob/247705f90c606741149fec53c6738cce28a386a7/packages/node-utils/src/index.ts
    {
      code: "export * from './logger';\nexport * from './AppsembleError';\nexport * from './basicAuth';\nexport * from './commandDirOptions';\nexport * from './getWorkspaces';\nexport * from './handleError';\nexport * from './interceptors';\nexport * from './loggerMiddleware';\nexport * from './readFileOrString';\nexport * from './fs';",
      output: "export * from './AppsembleError';\nexport * from './basicAuth';\nexport * from './commandDirOptions';\nexport * from './fs';\nexport * from './getWorkspaces';\nexport * from './handleError';\nexport * from './interceptors';\nexport * from './logger';\nexport * from './loggerMiddleware';\nexport * from './readFileOrString';",
      errors: 1,
    },
    // https://github.com/facebook/react/blob/4c7036e807fa18a3e21a5182983c7c0f05c5936e/packages/react-dom/src/client/ReactDOM.js#L193-L217
    {
      code: "var\n  createPortal,\n  batchedUpdates,\n  flushSync,\n  Internals,\n  ReactVersion,\n  findDOMNode,\n  hydrate,\n  render,\n  unmountComponentAtNode,\n  createRoot,\n  createBlockingRoot,\n  flushControlled,\n  scheduleHydration,\n  renderSubtreeIntoContainer,\n  unstable_createPortal,\n  createEventHandle\n;\nexport {\n  createPortal,\n  batchedUpdates as unstable_batchedUpdates,\n  flushSync,\n  Internals as __SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED,\n  ReactVersion as version,\n  // Disabled behind disableLegacyReactDOMAPIs\n  findDOMNode,\n  hydrate,\n  render,\n  unmountComponentAtNode,\n  // exposeConcurrentModeAPIs\n  createRoot,\n  createBlockingRoot,\n  flushControlled as unstable_flushControlled,\n  scheduleHydration as unstable_scheduleHydration,\n  // Disabled behind disableUnstableRenderSubtreeIntoContainer\n  renderSubtreeIntoContainer as unstable_renderSubtreeIntoContainer,\n  // Disabled behind disableUnstableCreatePortal\n  // Temporary alias since we already shipped React 16 RC with it.\n  // Todo: remove in React 18.\n  unstable_createPortal,\n  // enableCreateEventHandleAPI\n  createEventHandle as unstable_createEventHandle,\n};",
      output: "var\n  createPortal,\n  batchedUpdates,\n  flushSync,\n  Internals,\n  ReactVersion,\n  findDOMNode,\n  hydrate,\n  render,\n  unmountComponentAtNode,\n  createRoot,\n  createBlockingRoot,\n  flushControlled,\n  scheduleHydration,\n  renderSubtreeIntoContainer,\n  unstable_createPortal,\n  createEventHandle\n;\nexport {\n  Internals as __SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED,\n  createBlockingRoot,\n  createPortal,\n  // exposeConcurrentModeAPIs\n  createRoot,\n  // Disabled behind disableLegacyReactDOMAPIs\n  findDOMNode,\n  flushSync,\n  hydrate,\n  render,\n  unmountComponentAtNode,\n  batchedUpdates as unstable_batchedUpdates,\n  // enableCreateEventHandleAPI\n  createEventHandle as unstable_createEventHandle,\n  // Disabled behind disableUnstableCreatePortal\n  // Temporary alias since we already shipped React 16 RC with it.\n  // Todo: remove in React 18.\n  unstable_createPortal,\n  flushControlled as unstable_flushControlled,\n  // Disabled behind disableUnstableRenderSubtreeIntoContainer\n  renderSubtreeIntoContainer as unstable_renderSubtreeIntoContainer,\n  scheduleHydration as unstable_scheduleHydration,\n  ReactVersion as version,\n};",
      errors: 1,
    },
    // https://github.com/apollographql/apollo-client/blob/39942881567ff9825a0f17bbf114ec441590f8bb/src/core/index.ts#L1-L98
    {
      code: "/* Core */\n\nexport {\n  ApolloClient,\n  ApolloClientOptions,\n  DefaultOptions\n} from '../ApolloClient';\nexport {\n  ObservableQuery,\n  FetchMoreOptions,\n  UpdateQueryOptions,\n  ApolloCurrentQueryResult,\n} from '../core/ObservableQuery';\nexport {\n  QueryBaseOptions,\n  QueryOptions,\n  WatchQueryOptions,\n  MutationOptions,\n  SubscriptionOptions,\n  FetchPolicy,\n  WatchQueryFetchPolicy,\n  ErrorPolicy,\n  FetchMoreQueryOptions,\n  SubscribeToMoreOptions,\n  MutationUpdaterFn,\n} from '../core/watchQueryOptions';\nexport { NetworkStatus } from '../core/networkStatus';\nexport * from '../core/types';\nexport {\n  Resolver,\n  FragmentMatcher as LocalStateFragmentMatcher,\n} from '../core/LocalState';\nexport { isApolloError, ApolloError } from '../errors/ApolloError';\n\n/* Cache */\n\nexport * from '../cache';\n\n/* Link */\n\nexport { empty } from '../link/core/empty';\nexport { from } from '../link/core/from';\nexport { split } from '../link/core/split';\nexport { concat } from '../link/core/concat';\nexport { execute } from '../link/core/execute';\nexport { ApolloLink } from '../link/core/ApolloLink';\nexport * from '../link/core/types';\nexport {\n  parseAndCheckHttpResponse,\n  ServerParseError\n} from '../link/http/parseAndCheckHttpResponse';\nexport {\n  serializeFetchParameter,\n  ClientParseError\n} from '../link/http/serializeFetchParameter';\nexport {\n  HttpOptions,\n  fallbackHttpConfig,\n  selectHttpOptionsAndBody,\n  UriFunction\n} from '../link/http/selectHttpOptionsAndBody';\nexport { checkFetcher } from '../link/http/checkFetcher';\nexport { createSignalIfSupported } from '../link/http/createSignalIfSupported';\nexport { selectURI } from '../link/http/selectURI';\nexport { createHttpLink } from '../link/http/createHttpLink';\nexport { HttpLink } from '../link/http/HttpLink';\nexport { fromError } from '../link/utils/fromError';\nexport { toPromise } from '../link/utils/toPromise';\nexport { fromPromise } from '../link/utils/fromPromise';\nexport { ServerError, throwServerError } from '../link/utils/throwServerError';\nexport {\n  Observable,\n  Observer,\n  ObservableSubscription\n} from '../utilities/observables/Observable';\n\n/* Supporting */\n\n// Note that importing `gql` by itself, then destructuring\n// additional properties separately before exporting, is intentional...\nimport gql from 'graphql-tag';\nexport const {\n  resetCaches,\n  disableFragmentWarnings,\n  enableExperimentalFragmentVariables,\n  disableExperimentalFragmentVariables\n} = gql;\nexport { gql };",
      output: "/* Core */\n\nexport {\n  ApolloClient,\n  ApolloClientOptions,\n  DefaultOptions\n} from '../ApolloClient';\nexport {\n  FragmentMatcher as LocalStateFragmentMatcher,\n  Resolver,\n} from '../core/LocalState';\nexport { NetworkStatus } from '../core/networkStatus';\nexport {\n  ApolloCurrentQueryResult,\n  FetchMoreOptions,\n  ObservableQuery,\n  UpdateQueryOptions,\n} from '../core/ObservableQuery';\nexport * from '../core/types';\nexport {\n  ErrorPolicy,\n  FetchMoreQueryOptions,\n  FetchPolicy,\n  MutationOptions,\n  MutationUpdaterFn,\n  QueryBaseOptions,\n  QueryOptions,\n  SubscribeToMoreOptions,\n  SubscriptionOptions,\n  WatchQueryFetchPolicy,\n  WatchQueryOptions,\n} from '../core/watchQueryOptions';\nexport { ApolloError,isApolloError } from '../errors/ApolloError';\n\n/* Cache */\n\nexport * from '../cache';\n\n/* Link */\n\nexport { ApolloLink } from '../link/core/ApolloLink';\nexport { concat } from '../link/core/concat';\nexport { empty } from '../link/core/empty';\nexport { execute } from '../link/core/execute';\nexport { from } from '../link/core/from';\nexport { split } from '../link/core/split';\nexport * from '../link/core/types';\nexport { checkFetcher } from '../link/http/checkFetcher';\nexport { createHttpLink } from '../link/http/createHttpLink';\nexport { createSignalIfSupported } from '../link/http/createSignalIfSupported';\nexport { HttpLink } from '../link/http/HttpLink';\nexport {\n  parseAndCheckHttpResponse,\n  ServerParseError\n} from '../link/http/parseAndCheckHttpResponse';\nexport {\n  fallbackHttpConfig,\n  HttpOptions,\n  selectHttpOptionsAndBody,\n  UriFunction\n} from '../link/http/selectHttpOptionsAndBody';\nexport { selectURI } from '../link/http/selectURI';\nexport {\n  ClientParseError,\n  serializeFetchParameter} from '../link/http/serializeFetchParameter';\nexport { fromError } from '../link/utils/fromError';\nexport { fromPromise } from '../link/utils/fromPromise';\nexport { ServerError, throwServerError } from '../link/utils/throwServerError';\nexport { toPromise } from '../link/utils/toPromise';\nexport {\n  Observable,\n  ObservableSubscription,\n  Observer} from '../utilities/observables/Observable';\n\n/* Supporting */\n\n// Note that importing `gql` by itself, then destructuring\n// additional properties separately before exporting, is intentional...\nimport gql from 'graphql-tag';\nexport const {\n  resetCaches,\n  disableFragmentWarnings,\n  enableExperimentalFragmentVariables,\n  disableExperimentalFragmentVariables\n} = gql;\nexport { gql };",
      errors: 2,
    },
    // Type exports.
    {
      code: "export type {Z} from \"Z\";\nexport type Y = 5;\nexport type {X} from \"X\";\nexport type {B} from \"./B\";\nexport type {C} from \"/B\";\nexport type {E} from \"@/B\";\nexport {a, type type as type, z} from \"../type\";",
      output: "export type {Z} from \"Z\";\nexport type Y = 5;\nexport {a, type type as type, z} from \"../type\";\nexport type {B} from \"./B\";\nexport type {C} from \"/B\";\nexport type {E} from \"@/B\";\nexport type {X} from \"X\";",
      errors: 1,
    },
    // Type import with same name as regular import comes first.
    {
      code: "export {MyClass, type MyClass} from \"../type\";",
      output: "export {type MyClass,MyClass} from \"../type\";",
      errors: 1,
    },
    // Exports inside module declarations.
    {
      code: "export type {X} from \"X\";\nexport type {B} from \"./B\";\n\ndeclare module 'my-module' {\n  export type { PlatformPath, ParsedPath } from 'path';\n  export { type CopyOptions } from 'fs'; interface Something {}\n  export {a, type type as type, z} from \"../type\";\n  // comment\n    export * as d from \"d\"\nexport {c} from \"c\"; /*\n  */\texport {} from \"b\"; // b\n}",
      output: "export type {B} from \"./B\";\nexport type {X} from \"X\";\n\ndeclare module 'my-module' {\n  export { type CopyOptions } from 'fs'; \n  export type { ParsedPath,PlatformPath } from 'path';interface Something {}\n  export {a, type type as type, z} from \"../type\";\n  // comment\n/*\n  */\texport {} from \"b\"; // b\nexport {c} from \"c\"; \n    export * as d from \"d\"\n}",
      errors: 3,
    },
    // Type exports.
    {
      code: "export type {Z} from \"Z\";\nexport type Y = 5;\nexport type {X} from \"X\";\nexport type {B} from \"./B\";\nexport type {C} from \"/B\";\nexport type {E} from \"@/B\";",
      output: "export type {Z} from \"Z\";\nexport type Y = 5;\nexport type {B} from \"./B\";\nexport type {C} from \"/B\";\nexport type {E} from \"@/B\";\nexport type {X} from \"X\";",
      errors: 1,
    },
    // https://github.com/graphql/graphql-js/blob/f7061fdcf461a2e4b3c78077afaebefc2226c8e3/src/utilities/index.js#L1-L115
    {
      code: "// @flow strict\n\n// Produce the GraphQL query recommended for a full schema introspection.\n// Accepts optional IntrospectionOptions.\nexport { getIntrospectionQuery } from './getIntrospectionQuery';\n\nexport type {\n  IntrospectionOptions,\n  IntrospectionQuery,\n  IntrospectionSchema,\n  IntrospectionType,\n  IntrospectionInputType,\n  IntrospectionOutputType,\n  IntrospectionScalarType,\n  IntrospectionObjectType,\n  IntrospectionInterfaceType,\n  IntrospectionUnionType,\n  IntrospectionEnumType,\n  IntrospectionInputObjectType,\n  IntrospectionTypeRef,\n  IntrospectionInputTypeRef,\n  IntrospectionOutputTypeRef,\n  IntrospectionNamedTypeRef,\n  IntrospectionListTypeRef,\n  IntrospectionNonNullTypeRef,\n  IntrospectionField,\n  IntrospectionInputValue,\n  IntrospectionEnumValue,\n  IntrospectionDirective,\n} from './getIntrospectionQuery';\n\n// Gets the target Operation from a Document.\nexport { getOperationAST } from './getOperationAST';\n\n// Gets the Type for the target Operation AST.\nexport { getOperationRootType } from './getOperationRootType';\n\n// Convert a GraphQLSchema to an IntrospectionQuery.\nexport { introspectionFromSchema } from './introspectionFromSchema';\n\n// Build a GraphQLSchema from an introspection result.\nexport { buildClientSchema } from './buildClientSchema';\n\n// Build a GraphQLSchema from GraphQL Schema language.\nexport { buildASTSchema, buildSchema } from './buildASTSchema';\nexport type { BuildSchemaOptions } from './buildASTSchema';\n\n// Extends an existing GraphQLSchema from a parsed GraphQL Schema language AST.\nexport {\n  extendSchema,\n  // @deprecated: Get the description from a schema AST node and supports legacy\n  // syntax for specifying descriptions - will be removed in v16.\n  getDescription,\n} from './extendSchema';\n\n// Sort a GraphQLSchema.\nexport { lexicographicSortSchema } from './lexicographicSortSchema';\n\n// Print a GraphQLSchema to GraphQL Schema language.\nexport {\n  printSchema,\n  printType,\n  printIntrospectionSchema,\n} from './printSchema';\n\n// Create a GraphQLType from a GraphQL language AST.\nexport { typeFromAST } from './typeFromAST';\n\n// Create a JavaScript value from a GraphQL language AST with a type.\nexport { valueFromAST } from './valueFromAST';\n\n// Create a JavaScript value from a GraphQL language AST without a type.\nexport { valueFromASTUntyped } from './valueFromASTUntyped';\n\n// Create a GraphQL language AST from a JavaScript value.\nexport { astFromValue } from './astFromValue';\n\n// A helper to use within recursive-descent visitors which need to be aware of\n// the GraphQL type system.\nexport { TypeInfo, visitWithTypeInfo } from './TypeInfo';\n\n// Coerces a JavaScript value to a GraphQL type, or produces errors.\nexport { coerceInputValue } from './coerceInputValue';\n\n// Concatenates multiple AST together.\nexport { concatAST } from './concatAST';\n\n// Separates an AST into an AST per Operation.\nexport { separateOperations } from './separateOperations';\n\n// Strips characters that are not significant to the validity or execution\n// of a GraphQL document.\nexport { stripIgnoredCharacters } from './stripIgnoredCharacters';\n\n// Comparators for types\nexport {\n  isEqualType,\n  isTypeSubTypeOf,\n  doTypesOverlap,\n} from './typeComparators';\n\n// Asserts that a string is a valid GraphQL name\nexport { assertValidName, isValidNameError } from './assertValidName';\n\n// Compares two GraphQLSchemas and detects breaking changes.\nexport {\n  BreakingChangeType,\n  DangerousChangeType,\n  findBreakingChanges,\n  findDangerousChanges,\n} from './findBreakingChanges';\nexport type { BreakingChange, DangerousChange } from './findBreakingChanges';\n\n// Report all deprecated usage within a GraphQL document.\nexport { findDeprecatedUsages } from './findDeprecatedUsages';",
      output: "// @flow strict\n\n// Produce the GraphQL query recommended for a full schema introspection.\n// Accepts optional IntrospectionOptions.\nexport type {\n  IntrospectionDirective,\n  IntrospectionEnumType,\n  IntrospectionEnumValue,\n  IntrospectionField,\n  IntrospectionInputObjectType,\n  IntrospectionInputType,\n  IntrospectionInputTypeRef,\n  IntrospectionInputValue,\n  IntrospectionInterfaceType,\n  IntrospectionListTypeRef,\n  IntrospectionNamedTypeRef,\n  IntrospectionNonNullTypeRef,\n  IntrospectionObjectType,\n  IntrospectionOptions,\n  IntrospectionOutputType,\n  IntrospectionOutputTypeRef,\n  IntrospectionQuery,\n  IntrospectionScalarType,\n  IntrospectionSchema,\n  IntrospectionType,\n  IntrospectionTypeRef,\n  IntrospectionUnionType,\n} from './getIntrospectionQuery';\nexport { getIntrospectionQuery } from './getIntrospectionQuery';\n\n// Gets the target Operation from a Document.\nexport { getOperationAST } from './getOperationAST';\n\n// Gets the Type for the target Operation AST.\nexport { getOperationRootType } from './getOperationRootType';\n\n// Convert a GraphQLSchema to an IntrospectionQuery.\nexport { introspectionFromSchema } from './introspectionFromSchema';\n\n// Build a GraphQLSchema from an introspection result.\nexport { buildClientSchema } from './buildClientSchema';\n\n// Build a GraphQLSchema from GraphQL Schema language.\nexport type { BuildSchemaOptions } from './buildASTSchema';\nexport { buildASTSchema, buildSchema } from './buildASTSchema';\n\n// Extends an existing GraphQLSchema from a parsed GraphQL Schema language AST.\nexport {\n  extendSchema,\n  // @deprecated: Get the description from a schema AST node and supports legacy\n  // syntax for specifying descriptions - will be removed in v16.\n  getDescription,\n} from './extendSchema';\n\n// Sort a GraphQLSchema.\nexport { lexicographicSortSchema } from './lexicographicSortSchema';\n\n// Print a GraphQLSchema to GraphQL Schema language.\nexport {\n  printIntrospectionSchema,\n  printSchema,\n  printType,\n} from './printSchema';\n\n// Create a GraphQLType from a GraphQL language AST.\nexport { typeFromAST } from './typeFromAST';\n\n// Create a JavaScript value from a GraphQL language AST with a type.\nexport { valueFromAST } from './valueFromAST';\n\n// Create a JavaScript value from a GraphQL language AST without a type.\nexport { valueFromASTUntyped } from './valueFromASTUntyped';\n\n// Create a GraphQL language AST from a JavaScript value.\nexport { astFromValue } from './astFromValue';\n\n// A helper to use within recursive-descent visitors which need to be aware of\n// the GraphQL type system.\nexport { TypeInfo, visitWithTypeInfo } from './TypeInfo';\n\n// Coerces a JavaScript value to a GraphQL type, or produces errors.\nexport { coerceInputValue } from './coerceInputValue';\n\n// Concatenates multiple AST together.\nexport { concatAST } from './concatAST';\n\n// Separates an AST into an AST per Operation.\nexport { separateOperations } from './separateOperations';\n\n// Strips characters that are not significant to the validity or execution\n// of a GraphQL document.\nexport { stripIgnoredCharacters } from './stripIgnoredCharacters';\n\n// Comparators for types\nexport {\n  doTypesOverlap,\n  isEqualType,\n  isTypeSubTypeOf,\n} from './typeComparators';\n\n// Asserts that a string is a valid GraphQL name\nexport { assertValidName, isValidNameError } from './assertValidName';\n\n// Compares two GraphQLSchemas and detects breaking changes.\nexport type { BreakingChange, DangerousChange } from './findBreakingChanges';\nexport {\n  BreakingChangeType,\n  DangerousChangeType,\n  findBreakingChanges,\n  findDangerousChanges,\n} from './findBreakingChanges';\n\n// Report all deprecated usage within a GraphQL document.\nexport { findDeprecatedUsages } from './findDeprecatedUsages';",
      errors: 5,
    },
  ],
});
