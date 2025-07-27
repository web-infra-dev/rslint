## Rule: ban-ts-comment

### Test File: ban-ts-comment.test.ts

### Validation Summary
- ✅ **CORRECT**: 
  - Core directive detection logic (ts-expect-error, ts-ignore, ts-nocheck, ts-check)
  - Configuration option parsing and handling
  - Error message generation and suggestion system
  - Description length validation with Unicode support
  - Regular expression pattern matching for directive extraction
  - Multi-line comment processing (last line detection)
  - Special handling for ts-ignore to suggest ts-expect-error replacement

- ⚠️ **POTENTIAL ISSUES**:
  - Regex pattern differences for pragma comments (slash count handling)
  - Comment text extraction methodology differences
  - ts-nocheck positioning logic complexity
  - Unicode string length calculation variations
  - Comment scanning in unreachable code blocks

- ❌ **INCORRECT**:
  - Pragma comment slash count validation
  - ts-check block comment vs line comment handling logic
  - Comment position boundary checking

### Discrepancies Found

#### 1. Pragma Comment Slash Count Validation

**TypeScript Implementation:**
```typescript
const singleLinePragmaRegEx = /^\/\/\/?\s*@ts-(?<directive>check|nocheck)(?<description>.*)$/;
// Later used with:
const matchedPragma = execDirectiveRegEx(singleLinePragmaRegEx, `//${comment.value}`);
```

**Go Implementation:**
```go
singleLinePragmaRegEx = regexp.MustCompile(`^\/\/+\s*@ts-(?P<directive>check|nocheck)(?P<description>[\s\S]*)$`)
// With additional check:
originalSlashCount := len(commentText) - len(strings.TrimLeft(commentText, "/"))
if originalSlashCount <= 3 {
    if matchedPragma := execDirectiveRegEx(singleLinePragmaRegExForMultiLine, commentValue); matchedPragma != nil {
        return matchedPragma
    }
}
```

**Issue:** TypeScript allows exactly 2-3 slashes (`\/\/\/?`) while Go allows 1+ slashes (`\/\/+`) but then manually checks for ≤3 slashes. This creates inconsistent behavior.

**Impact:** Test case `//// @ts-nocheck` (4 slashes) should be valid according to TypeScript-ESLint tests, but the implementations handle this differently.

**Test Coverage:** The test case `//// @ts-nocheck - pragma comments may contain 2 or 3 leading slashes` appears to contradict the regex pattern.

#### 2. ts-check Block Comment Handling

**TypeScript Implementation:**
```typescript
// No special logic distinguishing block vs line comments for ts-check
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
if directive.Directive == "check" {
    if enabled && mode == "" {
        commentText := sourceText[commentRange.Pos():commentRange.End()]
        isBlockComment := strings.HasPrefix(commentText, "/*") || strings.HasPrefix(commentText, "/**")
        
        if !isBlockComment {
            ctx.ReportRange(commentRange.TextRange, buildTsDirectiveCommentMessage(directive.Directive))
        }
        return
    }
}
```

**Issue:** Go implementation adds special logic to allow block comments with ts-check when enabled=true, but only reports line comments. TypeScript implementation treats all comments equally.

**Impact:** Test cases expecting `/* @ts-check */` to be valid with `ts-check: true` option would behave differently.

**Test Coverage:** Multiple test cases show block comments should be valid: `/* @ts-check */`, `/** @ts-check */`

#### 3. Comment Text Extraction and Boundary Handling

**TypeScript Implementation:**
```typescript
const commentLines = comment.value.split('\n');
return execDirectiveRegEx(commentDirectiveRegExMultiLine, commentLines[commentLines.length - 1]);
```

**Go Implementation:**
```go
if startPos < 0 || startPos >= len(sourceText) || endPos <= startPos || endPos > len(sourceText) {
    return nil
}
commentText := sourceText[startPos:endPos]
```

**Issue:** TypeScript works with preprocessed comment values while Go extracts raw text including delimiters, requiring additional processing to handle `/*` and `*/`.

**Impact:** Edge cases with malformed comments or boundary conditions might behave differently.

**Test Coverage:** Tests with complex multi-line comment structures may reveal parsing differences.

#### 4. ts-nocheck Position Logic Complexity

**TypeScript Implementation:**
```typescript
if (directive === 'nocheck' && firstStatement && firstStatement.loc.start.line <= comment.loc.start.line) {
    return;
}
```

**Go Implementation:**
```go
if directive.Directive == "nocheck" {
    commentText := sourceText[commentRange.Pos():commentRange.End()]
    isBlockComment := strings.HasPrefix(commentText, "/*") || strings.HasPrefix(commentText, "/**")
    
    if isBlockComment {
        return  // Block comments with ts-nocheck are always allowed
    }
    
    if firstStatement == nil {
        return  // No statements in file, allow ts-nocheck
    }
    if commentRange.Pos() < firstStatement.Pos() {
        return
    }
}
```

**Issue:** Go adds complexity by treating block comments with ts-nocheck as always valid, while TypeScript only checks line position regardless of comment type.

**Impact:** Block comments with ts-nocheck might be incorrectly allowed in positions where they shouldn't be.

**Test Coverage:** Test case with ts-nocheck after first statement should fail, but block comment behavior differs.

#### 5. Unicode String Length Calculation

**TypeScript Implementation:**
```typescript
// Uses getStringLength utility from '@typescript-eslint/utils'
if (getStringLength(description.trim()) < nullThrows(minimumDescriptionLength))
```

**Go Implementation:**
```go
func getStringLength(s string) int {
    count := 0
    runes := []rune(s)
    
    for i := 0; i < len(runes); i++ {
        r := runes[i]
        
        // Count ASCII letters, numbers, and meaningful punctuation
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
            r == ' ' || r == '.' || /* ... many specific chars ... */ {
            count++
        } else if r >= 0x1F000 {
            // Complex emoji handling logic
            count++
            // Skip over zero-width joiners and variation selectors
        }
    }
    return count
}
```

**Issue:** The Go implementation has custom Unicode counting logic that may differ from TypeScript-ESLint's `getStringLength` utility, particularly for complex emoji sequences.

**Impact:** Test cases with emoji descriptions like `👨‍👩‍👧‍👦` might count differently, affecting minimum length validation.

**Test Coverage:** Tests with emoji in descriptions verify proper Unicode length calculation.

### Recommendations

- **Fix pragma comment regex**: Align Go regex pattern to match TypeScript's exact 2-3 slash requirement
- **Standardize ts-check handling**: Remove Go's special block comment logic for ts-check or verify it matches expected behavior
- **Validate Unicode counting**: Ensure Go's `getStringLength` produces identical results to TypeScript-ESLint's utility
- **Simplify ts-nocheck logic**: Align Go's position checking with TypeScript's simpler line-based approach
- **Test comment boundary cases**: Add tests for edge cases in comment parsing and delimiter handling
- **Verify unreachable code scanning**: The Go implementation has additional scanning logic for comments in `if (false)` blocks that may not be necessary

### Test Cases Requiring Attention

1. `//// @ts-nocheck` - Should this be valid with 4 slashes?
2. `/* @ts-check */` with `ts-check: true` - Should block comments be allowed?
3. Block comment `/* @ts-nocheck */` positioning - Should location matter?
4. Emoji length counting consistency across implementations
5. Complex multi-line comment directive detection on last line only

---