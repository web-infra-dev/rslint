# RSLint TypeScript-ESLint Rule Implementation Guide

## Rule Naming Convention

RSLint follows a specific convention for TypeScript-ESLint rules to maintain consistency across the codebase.

### Rule Implementation Names

When implementing a TypeScript-ESLint rule, the rule's `Name` field should contain **only the base rule name without the `@typescript-eslint/` prefix**.

```go
var MyRule = rule.Rule{
    Name: "my-rule-name",  // ✅ Correct
    // NOT: "@typescript-eslint/my-rule-name" ❌
}
```

### Rule Registration

Rules are registered in `internal/config/config.go` in the `RegisterAllTypeSriptEslintPluginRules` function. When registering, use the **full name including the `@typescript-eslint/` prefix**:

```go
GlobalRuleRegistry.Register("@typescript-eslint/my-rule-name", my_rule.MyRule)
```

### Why This Pattern?

1. The rule implementation uses the short name for simplicity
2. The registry key uses the full name to match ESLint convention
3. The `getAllTypeScriptEslintPluginRules` function adds the prefix automatically when needed
4. This allows the rules to work correctly with both the plugin system and direct rule references

### Example

For the `adjacent-overload-signatures` rule:

```go
// In internal/rules/adjacent_overload_signatures/adjacent_overload_signatures.go
var AdjacentOverloadSignaturesRule = rule.Rule{
    Name: "adjacent-overload-signatures",  // Short name
    // ...
}

// In internal/config/config.go
GlobalRuleRegistry.Register("@typescript-eslint/adjacent-overload-signatures",
    adjacent_overload_signatures.AdjacentOverloadSignaturesRule)
```

## Adding New TypeScript-ESLint Rules

1. Create the rule implementation in `internal/rules/<rule_name>/<rule_name>.go`
2. Use the short name (without prefix) in the rule's `Name` field
3. Add the import to `internal/config/config.go`
4. Register the rule with the full prefixed name in `RegisterAllTypeSriptEslintPluginRules`
5. Add the struct field to `TypedRules` if needed for configuration

## Testing

- Rule tests go in `packages/rslint-test-tools/tests/typescript-eslint/rules/`
- Use the rule tester with just the short name (the prefix is handled automatically)
- Follow the existing test file patterns for consistency
