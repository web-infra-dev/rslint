# RSLint API Architecture Documentation

## 1. Overview

RSLint provides a comprehensive API architecture that supports multiple communication modes including IPC (Inter-Process Communication), CLI, and LSP (Language Server Protocol). The API is designed to facilitate seamless integration between JavaScript/TypeScript frontends and the Go-based linting engine.

## 2. Core API Components

### 2.1 IPC Communication Layer

#### Protocol Structure

The IPC system implements a binary message protocol similar to esbuild:

- **Message Format**: 4-byte length header (uint32 little endian) + JSON message content
- **Bidirectional**: Supports both request-response and streaming communication
- **Thread-safe**: Uses mutex for concurrent access protection

#### Message Types

```go
type MessageKind string

const (
    KindLint      MessageKind = "lint"      // JS → Go lint request
    KindResponse  MessageKind = "response"  // Go → JS lint results
    KindError     MessageKind = "error"     // Error notifications
    KindHandshake MessageKind = "handshake" // Connection verification
    KindExit      MessageKind = "exit"      // Termination request
)
```

#### Core Message Structure

```go
type Message struct {
    Kind MessageKind `json:"kind"`
    ID   int         `json:"id"`
    Data interface{} `json:"data,omitempty"`
}
```

### 2.2 Request/Response Models

#### Lint Request

```go
type LintRequest struct {
    Files            []string               `json:"files,omitempty"`
    Config           string                 `json:"config,omitempty"`
    Format           string                 `json:"format,omitempty"`
    WorkingDirectory string                 `json:"workingDirectory,omitempty"`
    RuleOptions      map[string]interface{} `json:"ruleOptions,omitempty"`
    FileContents     map[string]string      `json:"fileContents,omitempty"`
}
```

#### Lint Response

```go
type LintResponse struct {
    Diagnostics []Diagnostic `json:"diagnostics"`
    ErrorCount  int          `json:"errorCount"`
    FileCount   int          `json:"fileCount"`
    RuleCount   int          `json:"ruleCount"`
}
```

#### Diagnostic Structure

```go
type Diagnostic struct {
    RuleName  string `json:"ruleName"`
    Message   string `json:"message"`
    FilePath  string `json:"filePath"`
    Range     Range  `json:"range"`
    Severity  string `json:"severity,omitempty"`
    MessageId string `json:"messageId"`
}

type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}

type Position struct {
    Line   int `json:"line"`
    Column int `json:"column"`
}
```

### 2.3 Service Management

#### IPC Service

```go
type Service struct {
    reader  *bufio.Reader
    writer  io.Writer
    handler Handler
    mutex   sync.Mutex
}
```

**Key Features:**

- **Concurrent Safety**: Mutex-protected message writing
- **Streaming Support**: Buffered reader for efficient message processing
- **Error Handling**: Comprehensive error propagation and recovery
- **Protocol Versioning**: Handshake-based version negotiation

#### Handler Interface

```go
type Handler interface {
    HandleLint(req LintRequest) (*LintResponse, error)
}
```

## 3. API Implementation Layers

### 3.1 IPC Handler (`cmd/rslint/api.go`)

#### Core Responsibilities

1. **Filesystem Setup**: Creates overlay VFS for file content injection
2. **Configuration Loading**: Loads rslint.json/rslint.jsonc with fallback
3. **Rule Registration**: Registers all TypeScript ESLint plugin rules
4. **TypeScript Integration**: Sets up compiler host and programs
5. **Linting Execution**: Orchestrates the linting process

#### Implementation Flow

```go
func (h *IPCHandler) HandleLint(req api.LintRequest) (*api.LintResponse, error) {
    // 1. Working directory setup
    if req.WorkingDirectory != "" {
        os.Chdir(req.WorkingDirectory)
    }

    // 2. Filesystem creation with overlay support
    fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
    if len(req.FileContents) > 0 {
        fs = utils.NewOverlayVFS(fs, req.FileContents)
    }

    // 3. Rule registration and configuration
    rslintconfig.RegisterAllTypeScriptEslintPluginRules()
    _, tsConfigs, configDirectory := rslintconfig.LoadConfigurationWithFallback(...)

    // 4. Rule filtering based on request options
    rulesWithOptions := filterRulesByOptions(origin_rules, req.RuleOptions)

    // 5. TypeScript compiler setup
    host := utils.CreateCompilerHost(configDirectory, fs)
    programs := createProgramsFromConfigs(tsConfigs, host)

    // 6. Linting execution
    return linter.RunLinter(...)
}
```

### 3.2 Message Processing

#### Read/Write Operations

```go
func (s *Service) readMessage() (*Message, error) {
    // Read 4-byte length header
    var length uint32
    binary.Read(s.reader, binary.LittleEndian, &length)

    // Read message content
    data := make([]byte, length)
    io.ReadFull(s.reader, data)

    // Unmarshal JSON
    var msg Message
    json.Unmarshal(data, &msg)
    return &msg, nil
}

func (s *Service) writeMessage(msg *Message) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    // Marshal to JSON
    data, _ := json.Marshal(msg)

    // Write length + content
    binary.Write(s.writer, binary.LittleEndian, uint32(len(data)))
    s.writer.Write(data)
    return nil
}
```

#### Message Routing

```go
func (s *Service) Start() error {
    for {
        msg, err := s.readMessage()
        if err != nil {
            return err
        }

        switch msg.Kind {
        case KindHandshake:
            s.handleHandshake(msg)
        case KindLint:
            s.handleLint(msg)
        case KindExit:
            return nil
        default:
            s.sendError(msg.ID, "unknown message kind")
        }
    }
}
```

## 4. Integration Points

### 4.1 TypeScript Compiler Integration

#### Compiler Host Setup

```go
host := utils.CreateCompilerHost(configDirectory, fs)
comparePathOptions := tspath.ComparePathsOptions{
    CurrentDirectory:          host.GetCurrentDirectory(),
    UseCaseSensitiveFileNames: host.FS().UseCaseSensitiveFileNames(),
}
```

#### Program Creation

```go
programs := []*compiler.Program{}
for _, configFileName := range tsConfigs {
    program, err := utils.CreateProgram(false, fs, configDirectory, configFileName, host)
    if err != nil {
        return nil, fmt.Errorf("error creating TS program for %s: %w", configFileName, err)
    }
    programs = append(programs, program)
}
```

### 4.2 Virtual File System (VFS)

#### Overlay VFS Support

- **Base Layer**: `bundled.WrapFS(cachedvfs.From(osvfs.FS()))`
- **Overlay Layer**: `utils.NewOverlayVFS(fs, req.FileContents)`
- **Purpose**: Allows in-memory file content injection for IDE integration

### 4.3 Rule System Integration

#### Rule Registration

```go
rslintconfig.RegisterAllTypeScriptEslintPluginRules()
```

#### Rule Filtering

```go
type RuleWithOption struct {
    rule   rule.Rule
    option interface{}
}

rulesWithOptions := []RuleWithOption{}
if len(req.RuleOptions) > 0 {
    for _, r := range origin_rules {
        if option, ok := req.RuleOptions[r.Name]; ok {
            rulesWithOptions = append(rulesWithOptions, RuleWithOption{
                rule:   r,
                option: option,
            })
        }
    }
}
```

## 5. Error Handling and Diagnostics

### 5.1 Error Response Structure

```go
type ErrorResponse struct {
    Message string `json:"message"`
}
```

### 5.2 Error Propagation

- **IPC Level**: Protocol-level errors (malformed messages, connection issues)
- **Application Level**: Configuration errors, TypeScript compilation errors
- **Rule Level**: Individual rule execution errors

### 5.3 Diagnostic Formatting

- **Structured Output**: JSON-based diagnostic format
- **Position Mapping**: 1-based line/column indexing for editor compatibility
- **Severity Levels**: Error, Warning, Info classification

## 6. Performance Considerations

### 6.1 Caching Strategy

- **VFS Caching**: `cachedvfs.From(osvfs.FS())` for filesystem operation optimization
- **Program Reuse**: TypeScript programs cached across requests
- **Rule Instance Reuse**: Rules instantiated once and reused

### 6.2 Concurrency

- **Thread-safe Messaging**: Mutex-protected write operations
- **Parallel Rule Execution**: Rules can be executed concurrently per file
- **Async Response Handling**: Non-blocking message processing

## 7. Protocol Versioning

### 7.1 Handshake Protocol

```go
type HandshakeRequest struct {
    Version string `json:"version"`
}

type HandshakeResponse struct {
    Version string `json:"version"`
    OK      bool   `json:"ok"`
}
```

### 7.2 Version Compatibility

- **Current Version**: "1.0.0"
- **Backward Compatibility**: Graceful degradation for version mismatches
- **Feature Detection**: Capability-based feature negotiation

## 8. Security Considerations

### 8.1 Input Validation

- **Path Sanitization**: All file paths normalized and validated
- **Content Size Limits**: Reasonable limits on file content size
- **Configuration Validation**: Schema-based config validation

### 8.2 Sandboxing

- **VFS Isolation**: Virtual filesystem prevents unauthorized file access
- **Working Directory Control**: Controlled directory changes
- **Resource Limits**: Memory and CPU usage monitoring

## 9. Extension Points

### 9.1 Custom Handlers

```go
type Handler interface {
    HandleLint(req LintRequest) (*LintResponse, error)
}
```

### 9.2 Plugin Architecture

- **Rule Plugins**: Dynamic rule registration
- **Format Plugins**: Custom output format support
- **VFS Plugins**: Custom filesystem implementations

## 10. Usage Examples

### 10.1 Basic IPC Usage

```go
// Create service
service := ipc.NewService(os.Stdin, os.Stdout, &IPCHandler{})

// Start processing
if err := service.Start(); err != nil {
    log.Fatal(err)
}
```

### 10.2 Custom Request

```json
{
  "kind": "lint",
  "id": 1,
  "data": {
    "files": ["src/index.ts"],
    "config": "rslint.json",
    "ruleOptions": {
      "@typescript-eslint/no-unused-vars": "error"
    },
    "fileContents": {
      "src/index.ts": "const unused = 42;"
    }
  }
}
```

### 10.3 Response Format

```json
{
  "kind": "response",
  "id": 1,
  "data": {
    "diagnostics": [
      {
        "ruleName": "@typescript-eslint/no-unused-vars",
        "message": "'unused' is assigned a value but never used.",
        "filePath": "src/index.ts",
        "range": {
          "start": { "line": 1, "column": 7 },
          "end": { "line": 1, "column": 13 }
        },
        "severity": "error"
      }
    ],
    "errorCount": 1,
    "fileCount": 1,
    "ruleCount": 1
  }
}
```

This API architecture provides a robust, scalable foundation for RSLint's linting capabilities while maintaining compatibility with various integration scenarios including IDEs, CI/CD systems, and standalone usage.
