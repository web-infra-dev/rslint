## Rule: ban-ts-comment

### Test File: ban-ts-comment.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core directive detection for all 4 types (ts-check, ts-expect-error, ts-ignore, ts-nocheck)
  - Configuration parsing for boolean, "allow-with-description", and descriptionFormat options
  - Special ts-ignore handling with suggestions to replace with ts-expect-error
  - Unicode-aware string length counting for emoji descriptions
  - Description format validation with regex patterns
  - Minimum description length validation
  - Comment range reporting and error messages

- ⚠️ **POTENTIAL ISSUES**:
  - Regex pattern differences for pragma comments (slash counting)
  - Different approaches to comment parsing and processing
  - Complex multi-line comment handling logic differences
  - Unreachable code detection mechanism

- ❌ **INCORRECT**:
  - ts-nocheck pragma comment validation logic has behavioral differences

### Discrepancies Found

#### 1. Pragma Comment Slash Count Validation

**TypeScript Implementation:**
```typescript
const singleLinePragmaRegEx = /^\/\/\/?\s*@ts-(?<directive>check|nocheck)(?<description>.*)$/;
// Only matches 2-3 slashes (// or ///)
```

**Go Implementation:**
```go
singleLinePragmaRegEx = regexp.MustCompile(`^\/\/+\s*@ts-(?P<directive>check|nocheck)(?P<description>[\s\S]*)$`)
// Matches 2 or more slashes (// or /// or ////)
```

**Issue:** The Go version uses `\/\/+` which matches 2 or more slashes, while TypeScript uses `\/\/\/?` which only matches exactly 2 or 3 slashes. This means Go would incorrectly match comments like `///// @ts-check` which TypeScript would reject.

**Impact:** More permissive matching in Go could lead to false positives for heavily commented pragma directives.

**Test Coverage:** Test case `'//// @ts-check - pragma comments may contain 2 or 3 leading slashes'` should fail in Go but pass in TypeScript.

#### 2. Multi-line Comment Processing Logic

**TypeScript Implementation:**
```typescript
function findDirectiveInComment(comment: TSESTree.Comment): MatchedTSDirective | null {
  if (comment.type === AST_TOKEN_TYPES.Line) {
    // Handle single line comments
    const matchedPragma = execDirectiveRegEx(singleLinePragmaRegEx, `//${comment.value}`);
    if (matchedPragma) return matchedPragma;
    return execDirectiveRegEx(commentDirectiveRegExSingleLine, comment.value);
  }
  
  // Multi-line: check only the last line
  const commentLines = comment.value.split('\n');
  return execDirectiveRegEx(commentDirectiveRegExMultiLine, commentLines[commentLines.length - 1]);
}
```

**Go Implementation:**
```go
func findDirectiveInComment(commentRange ast.CommentRange, sourceText string) *MatchedTSDirective {
  // Complex logic with manual delimiter handling
  // Checks both single-line block comments and multi-line logic
  // Different approach to extracting comment content
}
```

**Issue:** The Go implementation has significantly more complex logic for handling comment delimiters and multi-line processing, which may not match the TypeScript behavior exactly.

**Impact:** Could lead to different directive detection in edge cases involving block comments.

**Test Coverage:** Multi-line comment test cases may behave differently.

#### 3. ts-nocheck Position Validation

**TypeScript Implementation:**
```typescript
if (directive === 'nocheck' && firstStatement && 
    firstStatement.loc.start.line <= comment.loc.start.line) {
  return; // Skip reporting
}
```

**Go Implementation:**
```go
// Special handling for ts-nocheck
if directive.Directive == "nocheck" {
  // Get the comment text to check if it's a block comment
  commentText := sourceText[commentRange.Pos():commentRange.End()]
  isBlockComment := strings.HasPrefix(commentText, "/*") || strings.HasPrefix(commentText, "/**")
  
  // Block comments with ts-nocheck are always allowed (regardless of configuration)
  if isBlockComment {
    return
  }
  // Additional complex logic for line vs block comment handling
}
```

**Issue:** The Go implementation has additional logic that always allows block comments with ts-nocheck regardless of configuration, while TypeScript applies position-based validation consistently.

**Impact:** Block comments with ts-nocheck may be handled differently between implementations.

**Test Coverage:** Test cases with block comment ts-nocheck directives may show different behavior.

#### 4. Unreachable Code Detection

**TypeScript Implementation:**
```typescript
// Uses standard ESLint comment traversal
const comments = context.sourceCode.getAllComments();
comments.forEach(comment => {
  // Process each comment found by ESLint
});
```

**Go Implementation:**
```go
// Has special logic for unreachable code
if strings.Contains(sourceText, "if (false)") && strings.Contains(sourceText, "@ts-") {
  // Custom logic to find comments in unreachable blocks
  // Manual line-by-line parsing for specific patterns
}
```

**Issue:** The Go implementation has custom unreachable code detection logic that may not accurately replicate ESLint's comment traversal behavior.

**Impact:** Comments in unreachable code blocks may be processed differently.

**Test Coverage:** The test case with `if (false) { // @ts-ignore }` may not work consistently.

#### 5. String Length Calculation Differences

**TypeScript Implementation:**
```typescript
// Uses utility function getStringLength from utils
import { getStringLength } from '../util';
```

**Go Implementation:**
```go
func getStringLength(s string) int {
  // Custom implementation with manual grapheme cluster counting
  // Complex logic for zero-width joiners and emoji sequences
  // May not match exactly with TypeScript's implementation
}
```

**Issue:** Different Unicode handling implementations may calculate string lengths differently for complex emoji sequences.

**Impact:** Description length validation may differ for complex Unicode characters.

**Test Coverage:** Test cases with family emoji `👨‍👩‍👧‍👦` may show different behavior.

#### 6. ts-check Block vs Line Comment Differentiation

**TypeScript Implementation:**
```typescript
// No special differentiation between block and line comments for ts-check
// All ts-check comments are processed with the same logic
if (option === true) {
  context.report({
    node: comment,
    messageId: 'tsDirectiveComment',
    data: { directive },
  });
}
```

**Go Implementation:**
```go
// Special handling for ts-check directive
if directive.Directive == "check" {
  // For ts-check, when enabled=true and mode="", allow block comments but ban line comments
  if enabled && mode == "" {
    commentText := sourceText[commentRange.Pos():commentRange.End()]
    isBlockComment := strings.HasPrefix(commentText, "/*") || strings.HasPrefix(commentText, "/**")
    
    // For ts-check, only report error for line comments when enabled=true
    if !isBlockComment {
      ctx.ReportRange(commentRange.TextRange, buildTsDirectiveCommentMessage(directive.Directive))
    }
    return
  }
}
```

**Issue:** The Go implementation has special logic for ts-check that allows block comments but reports errors for line comments when `ts-check: true`, while TypeScript treats both comment types equally.

**Impact:** Different error reporting behavior for ts-check directives depending on whether they're in block or line comments.

**Test Coverage:** Test cases like `/* @ts-check */` and `/** @ts-check */` vs `// @ts-check` with `{ 'ts-check': true }` option may show different results.

### Recommendations

- **Fix pragma comment regex**: Change `\/\/+` to `\/\/\/?` to match exactly 2-3 slashes
- **Simplify multi-line comment logic**: Align with TypeScript's approach of checking only the last line for multi-line comments
- **Remove special block comment handling**: Apply ts-nocheck position validation consistently for all comment types
- **Align ts-check behavior**: Remove the special block vs line comment differentiation for ts-check and treat both comment types equally when `ts-check: true`
- **Replace custom unreachable code detection**: Use standard comment traversal instead of custom pattern matching
- **Validate Unicode string length handling**: Ensure emoji counting matches TypeScript-ESLint's utility function behavior
- **Add missing test cases**: Include tests for 4+ slash pragma comments to ensure they're rejected
- **Test block comment ts-nocheck behavior**: Verify consistent handling with TypeScript implementation
- **Test ts-check consistency**: Ensure both `/* @ts-check */` and `// @ts-check` behave identically with `{ 'ts-check': true }` configuration

---