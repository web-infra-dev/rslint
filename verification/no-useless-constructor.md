## Rule: no-useless-constructor

### Test File: no-useless-constructor.test.ts

### Validation Summary
- ✅ **CORRECT**: Core logic structure, accessibility checks, parameter property detection, constructor body analysis, spread argument handling, suggestions with fix ranges
- ⚠️ **POTENTIAL ISSUES**: Constructor overload detection logic, decorator detection method, super(...arguments) special case handling
- ❌ **INCORRECT**: Missing base rule integration, different AST node type expectations, incomplete decorator checking

### Discrepancies Found

#### 1. Base Rule Integration Missing
**TypeScript Implementation:**
```typescript
const baseRule = getESLintCoreRule('no-useless-constructor');
// ...
const rules = baseRule.create(context);
return {
  MethodDefinition(node): void {
    if (
      node.value.type === AST_NODE_TYPES.FunctionExpression &&
      checkAccessibility(node) &&
      checkParams(node)
    ) {
      rules.MethodDefinition(node);
    }
  },
};
```

**Go Implementation:**
```go
// Direct implementation without base rule integration
var NoUselessConstructorRule = rule.Rule{
    Name: "no-useless-constructor",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        return rule.RuleListeners{
            ast.KindConstructor: func(node *ast.Node) {
                // Direct logic implementation
            },
        }
    },
}
```

**Issue:** The TypeScript version leverages the base ESLint rule and only applies TypeScript-specific checks as filters, while the Go version implements all logic directly.

**Impact:** The Go version may miss some edge cases that the base ESLint rule handles, or may have different behavior for JavaScript-specific patterns.

**Test Coverage:** All test cases could potentially be affected, but particularly edge cases around JavaScript constructor patterns.

#### 2. AST Node Type Mismatch
**TypeScript Implementation:**
```typescript
return {
  MethodDefinition(node): void {
    if (node.value.type === AST_NODE_TYPES.FunctionExpression && ...) {
      rules.MethodDefinition(node);
    }
  },
};
```

**Go Implementation:**
```go
return rule.RuleListeners{
    ast.KindConstructor: func(node *ast.Node) {
        // Directly processes Constructor nodes
    },
}
```

**Issue:** TypeScript version listens for `MethodDefinition` nodes and filters for constructors, while Go version directly listens for `Constructor` nodes.

**Impact:** Different entry points may lead to different node context or missing certain constructor patterns.

**Test Coverage:** Should be revealed in test cases that expect `AST_NODE_TYPES.MethodDefinition` type in error reports.

#### 3. Decorator Detection Implementation
**TypeScript Implementation:**
```typescript
function checkParams(node: TSESTree.MethodDefinition): boolean {
  return !node.value.params.some(
    param =>
      param.type === AST_NODE_TYPES.TSParameterProperty ||
      param.decorators.length,
  );
}
```

**Go Implementation:**
```go
// Check for decorators
if ast.GetCombinedModifierFlags(param)&ast.ModifierFlagsDecorator != 0 {
    return false
}
```

**Issue:** Different approaches to decorator detection - TypeScript checks `param.decorators.length` while Go uses `GetCombinedModifierFlags`.

**Impact:** May miss or incorrectly identify decorated parameters.

**Test Coverage:** Test cases with decorators like `constructor(@Foo foo: string)` should reveal this.

#### 4. Constructor Overload Handling
**TypeScript Implementation:**
```typescript
// Implicit handling through base rule
```

**Go Implementation:**
```go
body := ctor.Body
if body == nil {
    // Constructor without body (overload signature)
    return false
}
```

**Issue:** Go version explicitly handles constructor overloads by checking for nil body, while TypeScript relies on base rule.

**Impact:** May have different behavior for constructor overloads and declare statements.

**Test Coverage:** Test cases with `constructor();` (overload signatures) and `declare class` should reveal differences.

#### 5. Super Arguments Special Case
**TypeScript Implementation:**
```typescript
// Implicit handling in base rule for super(...arguments)
```

**Go Implementation:**
```go
// Check special case: super(...arguments)
if len(superArgs.Nodes) == 1 && lastArg.Kind == ast.KindIdentifier &&
    lastArg.AsIdentifier().Text == "arguments" {
    return true
}
```

**Issue:** Go version has explicit special case handling for `super(...arguments)` while TypeScript relies on base rule.

**Impact:** May have different behavior for the `arguments` object usage pattern.

**Test Coverage:** Test case `constructor(a, b, ...c) { super(...arguments); }` should reveal this.

#### 6. Public Constructor in Extended Classes
**TypeScript Implementation:**
```typescript
case 'public':
  if (node.parent.parent.superClass) {
    return false;
  }
  break;
```

**Go Implementation:**
```go
case ast.KindPublicKeyword:
    // public constructors in classes with superClass are not useless
    if classNode != nil && ast.IsClassDeclaration(classNode) {
        classDecl := classNode.AsClassDeclaration()
        if classDecl.HeritageClauses != nil {
            for _, clause := range classDecl.HeritageClauses.Nodes {
                if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
                    return false
                }
            }
        }
    }
```

**Issue:** Both check for extends clauses but use different AST navigation patterns.

**Impact:** Should be equivalent but different implementation approaches.

**Test Coverage:** Test cases with `public constructor() {}` in extended classes should work correctly.

### Recommendations
- **Critical**: Investigate whether Go version needs base rule integration or if direct implementation covers all necessary cases
- **Important**: Verify decorator detection works correctly with `GetCombinedModifierFlags` approach
- **Important**: Ensure AST node type expectations align with test framework expectations
- **Medium**: Validate constructor overload handling matches TypeScript behavior
- **Low**: Confirm super(...arguments) special case handling is necessary and correct

### Test Cases Requiring Special Attention
- Constructor overloads: `constructor();`
- Decorated parameters: `constructor(@Foo foo: string)`
- Public constructors in extended classes: `class A extends B { public constructor() {} }`
- Super with arguments object: `super(...arguments)`
- Abstract class constructors

---