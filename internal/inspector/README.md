# Inspector Module

The `inspector` module builds code inspection information for the Playground, providing detailed AST, Type, Symbol, Signature, and Flow information.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           Frontend (React)                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐      ┌─────────────────────────────────────────────┐  │
│  │   Editor     │      │              AstInfoPanel                   │  │
│  │              │      │  ┌───────┬───────┬────────┬───────┐        │  │
│  │  cursor ─────┼──────┼─▶│ Node  │ Type  │ Symbol │ Flow  │ Tabs   │  │
│  │  position    │      │  └───┬───┴───┬───┴────┬───┴───┬───┘        │  │
│  │              │      │      ▼       ▼        ▼       ▼            │  │
│  │  ◀───────────┼──────┼─ InfoView  InfoView  InfoView  InfoView    │  │
│  │  highlight   │      │      │       │        │       │            │  │
│  └──────────────┘      └──────┼───────┼────────┼───────┼────────────┘  │
│                               │       │        │       │               │
│                               └───────┴────────┴───────┘               │
│                                       │                                │
│  ┌────────────────────────────────────┴────────────────────────────┐   │
│  │                    Lazy Loading Components                      │   │
│  │                                                                 │   │
│  │  LazyNodeView / LazyTypeView / LazySymbolView / LazySignatureView   │
│  │  ┌───────────────────────────────────────────────────────────┐  │   │
│  │  │ • Show shallow info initially (kind, pos, preview)        │  │   │
│  │  │ • On expand: fetch full info via context                  │  │   │
│  │  │ • Pass fileName for external file nodes                   │  │   │
│  │  │ • Highlight on hover (skip for external files)            │  │   │
│  │  └───────────────────────────────────────────────────────────┘  │   │
│  │                                                                 │   │
│  │  AstInfoContext: fetchAstInfo(pos, end, kind, fileName?)        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                   │
└────────────────────────────────────┼───────────────────────────────────┘
                                     │ IPC (JSON)
                                     ▼
┌────────────────────────────────────────────────────────────────────────┐
│                            Backend (Go)                                │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  cmd/rslint/api.go: HandleGetAstInfo                                   │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ 1. Get/create cached Program                                     │  │
│  │ 2. Find node: FindNodeAtPosition(sourceFile, pos, end, kind)     │  │
│  │ 3. If fileName set → load external file, find node there         │  │
│  │ 4. Build response via inspector.Builder                          │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                    │                                   │
│                                    ▼                                   │
│  internal/inspector/Builder                                            │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                                                                  │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────────┐  │  │
│  │  │ builder_    │ │ builder_    │ │ builder_    │ │ builder_   │  │  │
│  │  │ node.go     │ │ type.go     │ │ symbol.go   │ │ flow.go    │  │  │
│  │  │             │ │             │ │             │ │            │  │  │
│  │  │ BuildNode-  │ │ BuildType-  │ │ BuildSymbol-│ │ BuildFlow- │  │  │
│  │  │ Info()      │ │ Info()      │ │ Info()      │ │ Info()     │  │  │
│  │  │ BuildShallow│ │ BuildShallow│ │ BuildShallow│ │ BuildFlow- │  │  │
│  │  │ NodeInfo()  │ │ TypeInfo()  │ │ SymbolInfo()│ │ Graph()    │  │  │
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └────────────┘  │  │
│  │                                                                  │  │
│  │  helpers.go: FindNodeAtPosition, GetTypeAtNode, etc.             │  │
│  │  flags.go: GetNodeFlagNames, GetTypeFlagNames, etc.              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                    │                                   │
│                                    ▼                                   │
│  TypeScript-Go (checker.Checker, ast.Node, checker.Type, etc.)         │
└────────────────────────────────────────────────────────────────────────┘
```

## Lazy Loading Data Flow

```
Initial Request                          Lazy Expand Request
─────────────────                        ───────────────────

User clicks code                         User expands "Parent" node
       │                                        │
       ▼                                        ▼
getAstInfo({ pos: 42 })                  getAstInfo({
       │                                   pos: 10,      ← from shallow
       ▼                                   end: 100,     ← from shallow
┌─────────────────────┐                    kind: 80,     ← from shallow
│ Response:           │                    fileName: ""  ← from shallow
│                     │                  })
│ node: { FULL }      │                         │
│ type: { FULL }      │                         ▼
│ symbol: { FULL }    │                  ┌─────────────────────┐
│                     │                  │ Response:           │
│ node.parent:        │                  │ node: { FULL }      │
│   { SHALLOW }  ─────┼─────────────────▶│ type: { FULL }      │
│ type.properties:    │                  │ ...                 │
│   [{ SHALLOW }]     │                  └─────────────────────┘
└─────────────────────┘
```

## Shallow vs Full Info

| Aspect      | Full Info                   | Shallow Info             |
| ----------- | --------------------------- | ------------------------ |
| Built by    | `BuildNodeInfo()`           | `BuildShallowNodeInfo()` |
| Contains    | All fields + nested objects | Basic fields only        |
| Nested refs | Shallow info                | None                     |
| Use case    | Top-level display           | Lazy-loadable references |

```
Full Info Example:                    Shallow Info Example:
{                                     {
  id: 12345,                            kind: 211,
  kind: 80,                             kindName: "CallExpression",
  kindName: "Identifier",               pos: 0,
  pos: 42,                              end: 100,
  end: 49,                              fileName: "",  // optional
  text: "console",                      text: "log"    // optional
  parent: { shallow... },             }
  locals: [...],
  ...
}
```

## External File Handling

When a node references external files (e.g., `lib.d.ts`):

```
Click on "console"
       │
       ▼
┌──────────────────────────────────────────────────┐
│ symbol.valueDeclaration: {                       │
│   kind: 260,                                     │
│   pos: 21850,        ← position in lib.dom.d.ts │
│   end: 21869,                                    │
│   fileName: "bundled:///libs/lib.dom.d.ts" ◀──── │
│ }                                                │
└──────────────────────────────────────────────────┘
       │
       │ User expands valueDeclaration
       ▼
┌──────────────────────────────────────────────────┐
│ getAstInfo({                                     │
│   pos: 21850,                                    │
│   end: 21869,                                    │
│   kind: 260,                                     │
│   fileName: "bundled:///libs/lib.dom.d.ts" ◀──── │
│ })                                               │
└──────────────────────────────────────────────────┘
       │
       ▼
Backend loads lib.dom.d.ts → finds node → returns full info
```

**Note**: Highlighting is disabled for external file nodes since positions don't correspond to the current editor content.

## Symbol vs Type.Symbol

For an identifier like `console`, there are two different symbols:

|             | Symbol Panel                        | Type.Symbol                        |
| ----------- | ----------------------------------- | ---------------------------------- |
| Source      | `checker.GetSymbolAtLocation(node)` | `type.Symbol()`                    |
| Represents  | The **value** (variable `console`)  | The **type** (interface `Console`) |
| Flags       | Variable                            | Interface                          |
| Declaration | `declare var console: Console`      | `interface Console { ... }`        |

This reflects TypeScript's separation of values and types - they are intentionally different.

## Directory Structure

```
internal/inspector/
├── types.go              # Data type definitions
├── builder.go            # Builder entry point
├── builder_node.go       # AST node info builder
├── builder_type.go       # Type info builder
├── builder_symbol.go     # Symbol info builder
├── builder_signature.go  # Signature info builder
├── builder_flow.go       # Control flow info builder
├── flags.go              # Flag name resolvers
└── helpers.go            # Helper functions
```

## Core Types

### Request/Response

```go
// Request for AST info
type GetAstInfoRequest struct {
    FileContent     string         // Source code content
    Position        int            // Start position
    End             int            // End position (optional, for exact matching)
    Kind            int            // Node kind filter (optional)
    FileName        string         // Target file (empty for user file, external path for lib.d.ts etc.)
    CompilerOptions map[string]any // TypeScript compiler options
}

// Response
type GetAstInfoResponse struct {
    Node      *NodeInfo      // AST node info
    Type      *TypeInfo      // Type info
    Symbol    *SymbolInfo    // Symbol info
    Signature *SignatureInfo // Signature info
    Flow      *FlowInfo      // Control flow info
}
```

### NodeInfo

Contains detailed AST node information:

| Field        | Type       | Description                                  |
| ------------ | ---------- | -------------------------------------------- |
| `Id`         | uint64     | Node ID                                      |
| `Kind`       | int        | Node kind (SyntaxKind)                       |
| `KindName`   | string     | Kind name (e.g., "Identifier")               |
| `Pos`        | int        | Start position                               |
| `End`        | int        | End position                                 |
| `Flags`      | int        | Node flags                                   |
| `FlagNames`  | []string   | Flag name list                               |
| `Text`       | string     | Text content (identifiers/literals)          |
| `FileName`   | string     | External file path (only for external nodes) |
| `Parent`     | \*NodeInfo | Parent node (shallow)                        |
| `Name`       | \*NodeInfo | Name node                                    |
| `Expression` | \*NodeInfo | Expression node                              |
| ...          |            | More node-specific properties                |

### TypeInfo

Contains TypeScript type information:

| Field            | Type              | Description                   |
| ---------------- | ----------------- | ----------------------------- |
| `Id`             | uint32            | Type ID                       |
| `Flags`          | uint32            | TypeFlags                     |
| `FlagNames`      | []string          | Flag names                    |
| `ObjectFlags`    | uint32            | ObjectFlags                   |
| `TypeString`     | string            | Type string representation    |
| `Symbol`         | \*SymbolInfo      | Associated symbol             |
| `TypeArguments`  | []\*TypeInfo      | Generic type arguments        |
| `Properties`     | []\*SymbolInfo    | Type properties               |
| `CallSignatures` | []\*SignatureInfo | Call signatures               |
| ...              |                   | More type-specific properties |

### SymbolInfo

Contains symbol table information:

| Field              | Type           | Description               |
| ------------------ | -------------- | ------------------------- |
| `Id`               | uint64         | Symbol ID                 |
| `Name`             | string         | Symbol name               |
| `EscapedName`      | string         | Escaped name              |
| `Flags`            | uint32         | SymbolFlags               |
| `FlagNames`        | []string       | Flag names                |
| `Declarations`     | []\*NodeInfo   | Declaration node list     |
| `ValueDeclaration` | \*NodeInfo     | Primary value declaration |
| `Members`          | []\*SymbolInfo | Member symbols            |
| `Exports`          | []\*SymbolInfo | Exported symbols          |

### SignatureInfo

Contains function/method signature information:

| Field              | Type           | Description            |
| ------------------ | -------------- | ---------------------- |
| `Flags`            | uint32         | SignatureFlags         |
| `MinArgumentCount` | int            | Minimum argument count |
| `Parameters`       | []\*SymbolInfo | Parameter symbol list  |
| `TypeParameters`   | []\*TypeInfo   | Type parameters        |
| `ReturnType`       | \*TypeInfo     | Return type            |
| `Declaration`      | \*NodeInfo     | Declaration node       |

### FlowInfo

Contains control flow analysis information:

| Field         | Type         | Description                          |
| ------------- | ------------ | ------------------------------------ |
| `Flags`       | uint32       | FlowFlags                            |
| `FlagNames`   | []string     | Flag names                           |
| `Node`        | \*NodeInfo   | Associated AST node                  |
| `Antecedent`  | \*FlowInfo   | Predecessor node                     |
| `Antecedents` | []\*FlowInfo | Multiple predecessors (branch merge) |
| `Graph`       | \*FlowGraph  | Complete flow graph (top-level only) |

## Builder Usage

```go
import (
    "github.com/web-infra-dev/rslint/internal/inspector"
)

// Create Builder
builder := inspector.NewBuilder(typeChecker, sourceFile)

// Build node info
nodeInfo := builder.BuildNodeInfo(node)

// Build type info
typeInfo := builder.BuildTypeInfo(t)

// Build symbol info
symbolInfo := builder.BuildSymbolInfo(symbol)

// Build signature info
signatureInfo := builder.BuildSignatureInfo(sig)

// Build flow info
flowInfo := builder.BuildFlowInfo(flowNode)
```

## Helper Functions

```go
// Find node at specified position
node := inspector.FindNodeAtPosition(sourceFile, start, end, kind)

// Get type of a node
t := inspector.GetTypeAtNode(checker, node)

// Get signature of a node
sig := inspector.GetSignatureOfNode(checker, node)

// Get flow node of a node
flow := inspector.GetFlowNodeOfNode(node)
```

## Lazy Loading Support

To support frontend lazy loading, info is built in two forms:

1. **Full info**: Built via `BuildNodeInfo`, `BuildTypeInfo`, etc. Contains all fields.
2. **Shallow info**: Built via `BuildShallowNodeInfo`, `BuildShallowTypeInfo`, etc. Contains only basic fields and position info.

Shallow info includes `Pos`, `End`, `Kind`, `FileName` fields, allowing the frontend to make new requests for full info.

## External File Support

When a node comes from an external file (e.g., `lib.d.ts`):

1. The `FileName` field is set to the external file path (e.g., `bundled:///libs/lib.dom.d.ts`)
2. Frontend can request node info from that file by passing the `FileName` parameter
3. `Builder` automatically handles position calculation for external file nodes
