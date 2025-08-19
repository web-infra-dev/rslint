# Dot Notation Rule

This rule enforces the use of dot notation whenever possible, matching the behavior of TypeScript-ESLint's `@typescript-eslint/dot-notation` rule.

## Implementation Notes

- Correctly handles template literal index signatures (e.g., `[key: \`prefix\_${string}\`]`)
- Respects private/protected class member access based on configuration
- Provides auto-fix suggestions for converting between dot and bracket notation
- Handles TypeScript's `noPropertyAccessFromIndexSignature` compiler option

## Known Issues

- Line number reporting may differ slightly from TypeScript-ESLint due to AST traversal differences
- Some TypeScript test snapshots may need updates due to infrastructure differences between Go and TypeScript implementations

## Configuration

The rule supports the same options as TypeScript-ESLint:

- `allowKeywords` - Allow keywords like `while` to be accessed with dot notation
- `allowPattern` - Regex pattern for properties that can use bracket notation
- `allowPrivateClassPropertyAccess` - Allow private class properties to be accessed with bracket notation
- `allowProtectedClassPropertyAccess` - Allow protected class properties to be accessed with bracket notation
- `allowIndexSignaturePropertyAccess` - Allow properties matching index signatures to be accessed with bracket notation
