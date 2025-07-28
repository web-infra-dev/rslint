## Rule: ban-tslint-comment

### Test File: ban-tslint-comment.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core regex pattern matching (`^\s*tslint:(enable|disable)(?:-(line|next-line))?(:|\s|$)`)
  - Basic comment detection for both line and block comments
  - Message formatting with `toText` function
  - Fix suggestions for removing comments
  - Handling of whitespace trimming in comment text

- ⚠️ **POTENTIAL ISSUES**: 
  - Complex manual string parsing vs AST-based comment processing
  - Fix range calculation logic differs significantly
  - Different approaches to handling comment boundaries

- ❌ **INCORRECT**: 
  - **CRITICAL**: Missing AST-based comment processing - TypeScript uses `context.sourceCode.getAllComments()` while Go uses manual string searching
  - **CRITICAL**: Incorrect fix range calculation for removing comments with trailing characters
  - **CRITICAL**: Manual comment parsing may miss edge cases that AST parsing would catch

### Discrepancies Found

#### 1. Comment Processing Approach
**TypeScript Implementation:**
```typescript
const comments = context.sourceCode.getAllComments();
comments.forEach(c => {
  if (ENABLE_DISABLE_REGEX.test(c.value)) {
    // Process comment
  }
});
```

**Go Implementation:**
```go
// Process all tslint comments immediately
pos := 0
for {
    tslintPos := strings.Index(sourceText[pos:], "tslint:")
    if tslintPos == -1 {
        break
    }
    // Manual string parsing to find comment boundaries
}
```

**Issue:** The TypeScript version uses the AST's comment collection which provides structured comment objects with proper location information, while the Go version manually searches for "tslint:" strings and tries to reconstruct comment boundaries.

**Impact:** The manual approach in Go may:
- Miss comments in unusual positions
- Incorrectly identify "tslint:" strings that aren't actually in comments
- Have issues with nested or complex comment structures

**Test Coverage:** All test cases rely on proper comment detection, so this affects the entire rule.

#### 2. Fix Range Calculation
**TypeScript Implementation:**
```typescript
fix(fixer) {
  const rangeStart = context.sourceCode.getIndexFromLoc({
    column: c.loc.start.column > 0 ? c.loc.start.column - 1 : 0,
    line: c.loc.start.line,
  });
  const rangeEnd = context.sourceCode.getIndexFromLoc({
    column: c.loc.end.column,
    line: c.loc.end.line,
  });
  return fixer.removeRange([rangeStart, rangeEnd + 1]);
}
```

**Go Implementation:**
```go
// Calculate the proper range for removal
removeStart := commentStart
removeEnd := commentEnd

// Complex logic for handling line comments with preceding whitespace
if !isBlockComment {
    // ... extensive whitespace and newline handling logic
}

fixRange := utils.TrimNodeTextRange(sourceFile, sourceFile.AsNode()).WithPos(removeStart).WithEnd(removeEnd)
```

**Issue:** The TypeScript version uses a simple approach: remove from one character before the comment start to one character after the comment end. The Go version has much more complex logic for handling whitespace and newlines, which may not match the expected behavior.

**Impact:** The fix outputs may not match the expected test results, particularly for the case `'someCode(); // tslint:disable-line'` which should output `'someCode();'`.

**Test Coverage:** Test cases with `output` properties will fail if fix ranges are incorrect.

#### 3. Comment Text Extraction
**TypeScript Implementation:**
```typescript
// Comment object from AST already has parsed value
if (ENABLE_DISABLE_REGEX.test(c.value)) {
  // c.value is the comment content without comment markers
}
```

**Go Implementation:**
```go
// Manual extraction of comment value
var commentValue string
if isBlockComment {
    // Remove /* and */ from block comments
    if commentEnd > commentStart + 4 {
        commentValue = sourceText[commentStart+2 : commentEnd-2]
    }
} else {
    // Remove // from line comments
    if commentEnd > commentStart + 2 {
        commentValue = sourceText[commentStart+2 : commentEnd]
    }
}
```

**Issue:** Manual string extraction is error-prone and may not handle edge cases like malformed comments or comments at file boundaries.

**Impact:** Regex matching may fail on edge cases, causing the rule to miss violations or report false positives.

**Test Coverage:** Basic test cases should work, but edge cases may fail.

#### 4. Position Reporting
**TypeScript Implementation:**
```typescript
context.report({
  node: c,  // Reports on the comment node directly
  messageId: 'commentDetected',
  data: { text: toText(c.value, c.type) },
  // ...
});
```

**Go Implementation:**
```go
reportRange := utils.TrimNodeTextRange(sourceFile, sourceFile.AsNode()).WithPos(commentStart).WithEnd(commentEnd)
ctx.ReportRangeWithSuggestions(reportRange, buildCommentDetectedMessage(commentText), ...)
```

**Issue:** The TypeScript version reports on the comment AST node which provides accurate location information. The Go version manually calculates ranges which may have off-by-one errors or incorrect column positions.

**Impact:** Error positions in test assertions may not match (column numbers, line numbers).

**Test Coverage:** Tests with specific `column` and `line` assertions will likely fail.

#### 5. Rule Registration and Framework Integration
**TypeScript Implementation:**
```typescript
export default createRule({
  // Rule configuration and implementation
  create: context => {
    return {
      Program(): void {
        // Rule logic runs once per program
      },
    };
  },
});
```

**Go Implementation:**
```go
var BanTslintCommentRule = rule.Rule{
    Name: "ban-tslint-comment",
    Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
        // Rule logic runs immediately
        // Returns empty listeners
        return rule.RuleListeners{}
    },
}
```

**Issue:** The TypeScript version integrates with the AST traversal framework and runs during the Program node visit. The Go version runs all logic immediately and returns empty listeners, which bypasses the AST traversal system entirely.

**Impact:** This architectural difference means the Go version doesn't benefit from the AST framework's comment parsing and position tracking.

**Test Coverage:** This affects the fundamental behavior and may cause all tests to behave differently than expected.

### Recommendations
- **CRITICAL**: Rewrite the Go implementation to use AST-based comment processing instead of manual string parsing
- **CRITICAL**: Fix the range calculation logic to match TypeScript's simple approach: `[start-1, end+1]`
- **HIGH**: Integrate properly with the AST traversal framework by implementing a `Program` listener
- **HIGH**: Use the RSLint comment parsing utilities if available, or access comments through the TypeScript AST
- **MEDIUM**: Simplify the fix range logic to match the TypeScript behavior exactly
- **MEDIUM**: Add more comprehensive test cases for edge cases like comments at file boundaries, nested comments, and malformed comments
- **LOW**: Ensure message formatting exactly matches the TypeScript version

### Test Cases That Will Likely Fail
1. `'someCode(); // tslint:disable-line'` - Fix output should be `'someCode();'` but Go logic is more complex
2. Position assertions (column/line numbers) may be incorrect due to manual range calculation
3. Any edge cases involving comment boundaries or unusual whitespace

---