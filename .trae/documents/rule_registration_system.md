# RSLint Rule Registration System Documentation

## 1. Overview

The RSLint rule registration system is a centralized mechanism for managing, loading, and executing TypeScript ESLint rules within the Go-based linter. It provides a type-safe, extensible architecture for rule discovery, configuration, and execution.

## 2. Core Components

### 2.1 Rule Interface

```go
// internal/rule/rule.go
type Rule struct {
    Name string
    Run  func(ctx RuleContext, options any) RuleListeners
}

func CreateRule(r Rule) Rule {
    return Rule{
        Name: "@typescript-eslint/" + r.Name,
        Run:  r.Run,
    }
}
```

**Key Features:**

- **Name**: Unique identifier following TypeScript ESLint naming convention
- **Run**: Function that accepts context and options, returns event listeners
- **CreateRule**: Helper function that automatically prefixes rule names

### 2.2 Rule Registry

```go
// internal/config/rule_registry.go
type RuleRegistry struct {
    rules map[string]rule.Rule
}

func NewRuleRegistry() *RuleRegistry {
    return &RuleRegistry{
        rules: make(map[string]rule.Rule),
    }
}

func (r *RuleRegistry) Register(ruleName string, ruleImpl rule.Rule) {
    r.rules[ruleName] = ruleImpl
}

func (r *RuleRegistry) GetRule(name string) (rule.Rule, bool) {
    rule, exists := r.rules[name]
    return rule, exists
}
```

**Registry Operations:**

- **Register**: Adds a rule to the global registry
- **GetRule**: Retrieves a specific rule by name
- **GetAllRules**: Returns all registered rules
- **GetEnabledRules**: Filters rules based on configuration

### 2.3 Global Rule Registration

```go
// internal/config/config.go
var GlobalRuleRegistry = NewRuleRegistry()

func RegisterAllTypeScriptEslintPluginRules() {
    GlobalRuleRegistry.Register("@typescript-eslint/adjacent-overload-signatures", adjacent_overload_signatures.AdjacentOverloadSignaturesRule)
    GlobalRuleRegistry.Register("@typescript-eslint/array-type", array_type.ArrayTypeRule)
    GlobalRuleRegistry.Register("@typescript-eslint/await-thenable", await_thenable.AwaitThenableRule)
    GlobalRuleRegistry.Register("@typescript-eslint/class-literal-property-style", class_literal_property_style.ClassLiteralPropertyStyleRule)
    GlobalRuleRegistry.Register("@typescript-eslint/dot-notation", dot_notation.DotNotationRule)
    GlobalRuleRegistry.Register("@typescript-eslint/explicit-member-accessibility", explicit_member_accessibility.ExplicitMemberAccessibilityRule)
    GlobalRuleRegistry.Register("@typescript-eslint/max-params", max_params.MaxParamsRule)
    GlobalRuleRegistry.Register("@typescript-eslint/member-ordering", member_ordering.MemberOrderingRule)
    // ... additional rules
}
```

## 3. Rule Configuration System

### 3.1 Rule Configuration Structure

```go
type RuleConfig struct {
    Level   string                 `json:"level,omitempty"`   // "error", "warn", "off"
    Options map[string]interface{} `json:"options,omitempty"` // Rule-specific options
}

func (rc *RuleConfig) IsEnabled() bool {
    if rc == nil {
        return false
    }
    return rc.Level != "off" && rc.Level != ""
}
```

### 3.2 Configuration Loading

```go
type EnabledRuleWithConfig struct {
    Rule   rule.Rule
    Config *RuleConfig
}

func (r *RuleRegistry) GetEnabledRulesWithConfig(config RslintConfig, filePath string) []EnabledRuleWithConfig {
    enabledRuleConfigs := config.GetRulesForFile(filePath)
    var enabledRules []EnabledRuleWithConfig

    for ruleName, ruleConfig := range enabledRuleConfigs {
        if ruleConfig.IsEnabled() {
            if ruleImpl, exists := r.rules[ruleName]; exists {
                enabledRules = append(enabledRules, EnabledRuleWithConfig{
                    Rule:   ruleImpl,
                    Config: ruleConfig,
                })
            }
        }
    }
    return enabledRules
}
```

## 4. Rule Execution Context

### 4.1 Rule Context Structure

```go
type RuleContext struct {
    SourceFile                 *ast.SourceFile
    Program                    *compiler.Program
    TypeChecker                *checker.Checker
    DisableManager             *DisableManager
    ReportRange                func(textRange core.TextRange, msg RuleMessage)
    ReportRangeWithSuggestions func(textRange core.TextRange, msg RuleMessage, suggestions ...RuleSuggestion)
    ReportNode                 func(node *ast.Node, msg RuleMessage)
    ReportNodeWithFixes        func(node *ast.Node, msg RuleMessage, fixes ...RuleFix)
    ReportNodeWithSuggestions  func(node *ast.Node, msg RuleMessage, suggestions ...RuleSuggestion)
}
```

### 4.2 Rule Listeners

```go
type RuleListeners map[ast.Kind]func(node *ast.Node)

// Special listener types for pattern matching
func ListenerOnAllowPattern(kind ast.Kind) ast.Kind {
    return ast.Kind(int(kind) + 1000)
}

func ListenerOnExit(kind ast.Kind) ast.Kind {
    return ast.Kind(int(kind) + 2000)
}
```

## 5. Rule Implementation Pattern

### 5.1 Standard Rule Structure

```go
// Example: internal/rules/dot_notation/dot_notation.go
package dot_notation

import (
    "github.com/web-infra-dev/rslint/internal/rule"
    "github.com/microsoft/typescript-go/shim/ast"
)

var DotNotationRule = rule.CreateRule(rule.Rule{
    Name: "dot-notation",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindMemberExpression: func(node *ast.Node) {
                // Rule implementation logic
                memberExpr := node.AsMemberExpression()
                // Check conditions and report diagnostics
                if shouldUseDotNotation(memberExpr) {
                    ctx.ReportNode(node, rule.RuleMessage{
                        Id:          "preferDot",
                        Description: "Prefer dot notation over bracket notation",
                    })
                }
            },
        }
    },
})
```

### 5.2 Rule Testing Framework

```go
// Testing pattern for rules
func TestRuleNameRule(t *testing.T) {
    rule.RunTest(t, RuleNameRule, []rule.TestCase{
        {
            Name: "valid case",
            Code: "obj.property",
            Valid: true,
        },
        {
            Name: "invalid case",
            Code: "obj['property']",
            Valid: false,
            ExpectedDiagnostics: []rule.ExpectedDiagnostic{
                {
                    MessageId: "preferDot",
                    Line:      1,
                    Column:    1,
                },
            },
        },
    })
}
```

## 6. Rule Discovery and Loading

### 6.1 Automatic Rule Discovery

Rules are discovered through explicit registration in the `RegisterAllTypeScriptEslintPluginRules()` function. This approach provides:

- **Compile-time safety**: Missing rules cause compilation errors
- **Explicit dependencies**: Clear visibility of all available rules
- **Performance**: No runtime reflection or file system scanning

### 6.2 Dynamic Rule Loading

```go
// API mode rule filtering
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

## 7. Rule Disable Management

### 7.1 Disable Manager

```go
type DisableManager struct {
    sourceFile *ast.SourceFile
    // Internal state for tracking disable comments
}

func NewDisableManager(sourceFile *ast.SourceFile) *DisableManager {
    return &DisableManager{
        sourceFile: sourceFile,
    }
}

func (dm *DisableManager) IsRuleDisabled(ruleName string, position int) bool {
    // Check if rule is disabled at the given position
    // Supports eslint-disable comments
    return false // Implementation details
}
```

### 7.2 Disable Comment Support

The system supports standard ESLint disable comments:

- `// eslint-disable-next-line @typescript-eslint/rule-name`
- `/* eslint-disable @typescript-eslint/rule-name */`
- `/* eslint-enable @typescript-eslint/rule-name */`

## 8. Performance Optimizations

### 8.1 Listener Registration

```go
registeredListeners := make(map[ast.Kind][](func(node *ast.Node)), 20)

for _, r := range rules {
    for kind, listener := range r.Run(ctx) {
        listeners, ok := registeredListeners[kind]
        if !ok {
            listeners = make([](func(node *ast.Node)), 0, len(rules))
        }
        registeredListeners[kind] = append(listeners, listener)
    }
}
```

### 8.2 Efficient AST Traversal

```go
runListeners := func(kind ast.Kind, node *ast.Node) {
    if listeners, ok := registeredListeners[kind]; ok {
        for _, listener := range listeners {
            listener(node)
        }
    }
}
```

## 9. Error Handling and Diagnostics

### 9.1 Diagnostic Reporting

```go
type RuleDiagnostic struct {
    RuleName    string
    Range       core.TextRange
    Message     RuleMessage
    SourceFile  *ast.SourceFile
    Severity    DiagnosticSeverity
    FixesPtr    *[]RuleFix
    Suggestions *[]RuleSuggestion
}
```

### 9.2 Severity Levels

```go
type DiagnosticSeverity int

const (
    SeverityError DiagnosticSeverity = iota
    SeverityWarning
    SeverityInfo
    SeverityHint
)
```

## 10. Integration Points

### 10.1 CLI Integration

Rules are automatically loaded and executed through the CLI interface:

```bash
rslint [files...] --config rslint.json
```

### 10.2 API Integration

Rules can be selectively enabled through the IPC API:

```typescript
const result = await lint({
  files: ['src/**/*.ts'],
  ruleOptions: {
    '@typescript-eslint/dot-notation': 'error',
    '@typescript-eslint/max-params': ['error', { max: 3 }],
  },
});
```

### 10.3 LSP Integration

Rules integrate with the Language Server Protocol for real-time linting in editors.

## 11. Extension and Customization

### 11.1 Adding New Rules

1. Create rule implementation in `internal/rules/rule_name/`
2. Export rule variable following naming convention
3. Register rule in `RegisterAllTypeScriptEslintPluginRules()`
4. Add to hardcoded rule list in API mode
5. Write comprehensive tests

### 11.2 Rule Options

Rules can accept configuration options through the `options any` parameter:

```go
Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
    // Parse options
    opts := parseOptions(options)

    return rule.RuleListeners{
        ast.KindFunctionDeclaration: func(node *ast.Node) {
            // Use options in rule logic
            if shouldCheck(node, opts) {
                // Report diagnostic
            }
        },
    }
}
```

## 12. Best Practices

### 12.1 Rule Implementation

- Use specific AST node listeners for performance
- Implement comprehensive test coverage
- Handle edge cases gracefully
- Provide clear diagnostic messages
- Support fixable diagnostics when possible

### 12.2 Performance Considerations

- Minimize work in hot paths
- Use efficient data structures
- Avoid unnecessary AST traversals
- Cache expensive computations
- Profile rule performance regularly

### 12.3 Testing Guidelines

- Test both valid and invalid cases
- Include edge cases and boundary conditions
- Test with various TypeScript configurations
- Verify fix suggestions work correctly
- Test disable comment functionality

This rule registration system provides a robust, extensible foundation for implementing and managing TypeScript ESLint rules within the RSLint architecture.
