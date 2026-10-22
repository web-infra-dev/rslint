# rslint compiler extension patch

The direct Rust lint engine owns the AST and checker in process. It does not
serialize an AST or call these JSON checker extensions for each rule. The
patches retain the compiler API work from the earlier migration experiment;
not every extension is exercised by the current direct engine.

Base: official [oxc-project/tsrs](https://github.com/oxc-project/tsrs) commit
`280b426b89cd4ecd6153cf48228826bd60a99c61`, whose TypeScript reference is
`b85298b6a81f772d080b0455de0ca9d744cd6fd6`.

`0001-rslint-checker-api.patch` adds a small API extension inside the existing
checker request path. It reuses the snapshot lease, exclusive API checker,
node handles, and symbol/type/signature registries. No parallel checker or
independent type cache is created. Three narrow checker exports make existing
private queries callable; their implementations delegate to the original
methods without replacing their algorithms.

`0002-byte-safe-checker-strings.patch` preserves JavaScript lone-surrogate
strings across the checker bridge. Literal and template text metadata includes
raw WTF-8 bytes; property-name requests accept `nameBytes`, and string-returning
queries accept `byteStrings`. These additions preserve the original JSON API
for existing callers. Binder and transient symbol names and tables also carry
exact bytes, preventing distinct surrogate property names from collapsing.

`0003-optional-symbol-location.patch` exposes the checker's existing optional
location for `getTypeOfSymbolAtLocation`. A missing location keeps the declared
type instead of applying flow narrowing, while preserving exported-symbol
resolution and the compiler's missing-type handling. It does not replace this
operation with `getTypeOfSymbol`.

Prepare from the official source, then build with the generated Cargo config:

```sh
python3 scripts/tsrs/prepare-compiler.py
cargo +1.99.1 build --locked --release -p rslint-native
cargo +1.99.1 test --locked -p rslint-core
```

An optional `--mirror /path/to/tsrs` uses a local repository only as a source
of committed Git objects. It still checks out the exact base commit, applies
the included patch, and records hashes. Dirty files from that repository are
never copied. No commit, reset, or push occurs in the supplied mirror.

The generated Cargo configuration lives under ignored `.cache`.
Compiler sources live in a separate per-checkout directory under
`$XDG_CACHE_HOME/rslint/tsrs-compiler`, defaulting to
`~/.cache/rslint/tsrs-compiler`. This keeps their Cargo workspace outside the
rslint workspace, so inherited compiler package metadata resolves correctly.
`--source-cache /absolute/cache` can select another external directory.
Each distinct patch fingerprint gets a separate checkout. Repeated preparation
verifies HEAD and the complete source diff against its receipt; it fails on
unexpected changes instead of overwriting them. Cargo overrides all compiler
workspace packages to this same checkout because mixing patched path
dependencies with Git copies would create incompatible Rust type identities.
The generated configuration contains local paths and must not be committed.
Use the compiler's pinned Rust 1.99.1 toolchain, including for metadata and tests.

## Additional API methods

Every method below is prefixed with `rslint/`. Except `checkerInfo`, parameters
include the existing `snapshot` and `project` handles. `type`, `keyType`,
`source`, and `target` are upstream type IDs; `symbol` is a full upstream
symbol reference; `location` and `errorNode` are upstream node handles.
Compiler response shapes and nullable results are preserved.

| Method                                     | Additional parameters                           | Result                                                          |
| ------------------------------------------ | ----------------------------------------------- | --------------------------------------------------------------- |
| `checkerInfo`                              | None                                            | `{version: 1, programCreationOptions: 2, noEmitDiagnostics: 1}` |
| `getErrorType`                             | None                                            | Type                                                            |
| `getIndexTypeOfType`                       | `type`, `keyType`                               | Type or null                                                    |
| `getContextualType`                        | `location`, `flags`                             | Type or null                                                    |
| `getTypeOfSymbolAtLocation`                | `symbol`, optional `location`                   | Type or null                                                    |
| `typeToString`                             | `type`, optional `location`, `flags`            | String, preserving explicit zero flags                          |
| `getPromisedTypeOfPromise`                 | `type`                                          | Type or null                                                    |
| `getMinArgumentCount`                      | `signature`                                     | Number                                                          |
| `getSignatureMinArgumentCount`             | `signature`                                     | Raw signature minimum count, without checker resolution         |
| `getMinTypeArgumentCount`                  | `types`                                         | Number                                                          |
| `getConditionalRootNode`                   | `type`                                          | Node handle or null                                             |
| `getIterationTypeOfIterable`               | `type`, `use`, `typeKind`, optional `errorNode` | Type or null                                                    |
| `isTypeStrictSubtypeOf`                    | `source`, `target`                              | Boolean                                                         |
| `getGlobalSymbol`                          | `name`, `meaning`                               | Symbol or null                                                  |
| `getMergedSymbol`, `skipAlias`             | `symbol`                                        | Symbol                                                          |
| `resolveAlias`                             | Optional `symbol`                               | `{symbol, ok}`                                                  |
| `getAccessedPropertyName`                  | `location`                                      | `{name, ok}`                                                    |
| `getPropertyNameForKnownSymbolName`        | `name`                                          | String                                                          |
| `isNonDeferredTypeReference`               | `type`                                          | Boolean                                                         |
| `getPropertyNameFromType`                  | `type`                                          | String                                                          |
| `getSymbolFlags`                           | `symbol`                                        | Number                                                          |
| `getResolvedSignature`                     | `location`, `collectCandidates`, `checkMode`    | `{signature, candidates}`                                       |
| `getRootSymbols`                           | `symbol`                                        | Symbol array                                                    |
| `getTypeOnlyAliasDeclaration`              | `symbol`                                        | Node handle or null                                             |
| `getSymbolsOfParameterPropertyDeclaration` | `location`, `name`                              | `{parameter, property}`                                         |
| `isDeprecatedDeclaration`                  | `location`                                      | Boolean                                                         |
| `isUnknownSymbol`                          | `symbol`                                        | Boolean                                                         |
| `getSuggestionDiagnostics`                 | `file`                                          | Diagnostic array                                                |
| `symbolGraph`                              | `symbol`                                        | `{snapshot, project, root, symbols}`                            |

Iteration type kinds use the compiler's `Yield = 0`, `Return = 1`, and
`Next = 2`. All flags retain their upstream numeric representation.

`symbolGraph` exports a snapshot-owned symbol graph and also supports
transient symbols. It exports the real table
keys and preserves absent versus present empty tables. It does not derive
member/export keys from symbol display names. Every returned symbol is
registered through the original API registry, preserving identity across
binder and checker queries.

Non-null `typeToString.verbosity` and `getGlobalSymbol.diagnostic` arguments
currently return explicit unsupported errors. Existing rslint callers use
null for those arguments; they are not silently discarded. This patch is
migration infrastructure, not a claim that full lint compatibility has passed.

## Program construction and diagnostic support

`createPrograms[].options` and the equivalent reconfiguration options accept:

- `currentDirectory`: an absolute directory used by the compiler host and
  relative root filename resolution.
- `useSourceOfProjectReference`: the actual program construction flag.
- `singleThreaded`: the actual program construction override, distinct from
  the compiler option of the same name.
- `configFileName`: an optional config document identifier. The project
  builder reads it through the snapshot's real config registry, retaining the
  parsed config source, raw options, and diagnostic source positions. It
  clones that command line with the explicitly supplied roots, compiler
  options, and references. Missing or unreadable configs fail the request.

When `configFileName` is present, its native parse diagnostics are already
retained. `configFileParsingDiagnostics` adds extra diagnostics; callers must
not repeat that config's parse diagnostics in this array. Without
`configFileName`, the original synthetic program behavior is retained.
Omitted construction flags keep upstream defaults. Reconfiguration rebuilds
when an explicit construction option changes.

The patch also exposes a narrow project getter for the actual diagnostics
pool's accumulated globals. It initializes the diagnostic checker without a
semantic pass, then releases the checker lease before reading the pool.
The service uses this to preserve rslint's no-emit diagnostic ordering.
An additional standard-collector overload accepts that global callback while
preserving every original diagnostic phase and short-circuit condition.
The existing compiler collector delegates to the overload with its original
program-global callback, so its behavior is unchanged.

Three narrow session helpers let the service's binder name resolver register
symbols, serialize node handles, and resolve nodes through the existing API registries and cached node
tables. They do not initialize a type checker; callers retain the snapshot
and the allocation owner while any returned compiler pointer is used.

Cached scanner tokens are absent from the structural binary AST. The API node
handle adapter supports opaque `t.` handles for actual source-file token-cache
entries, retaining parent identity. It checks UTF-8 source boundaries, the
parent range, and actual scanner kind/end before resolving a token. Both
query arguments and returned node references use this same handle adapter.

On the compiler's existing `NO_THREADS` targets, project checker pools reclaim
expired idle diagnostics/query checkers cooperatively before the next checker
acquisition. They do not start the unsupported background cleanup timer. The
original held-checker, discarded-program, escaped-region, and result-lease
safeguards still control disposal; API checker identity is never idle-cleaned.
Program/session disposal releases the remaining regions. Native targets retain
the original timer behavior.
