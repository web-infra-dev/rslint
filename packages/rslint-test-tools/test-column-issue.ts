// Test case from dot-notation.test.ts
// Line 1
a['SHOUT_CASE'];
//^
// Column 3 (0-indexed) or Column 4 (1-indexed)
// The '[' character is at byte offset 5 from start of file:
// - 'a' (1 byte)
// - '\n' (1 byte)
// - ' ' (1 byte)
// - ' ' (1 byte)
// - '[' (1 byte) = position 4 (0-indexed)
// From start of line 2 (after '\n'):
// - ' ' (1 byte)
// - ' ' (1 byte)
// - '[' (1 byte) = position 2 from line start (0-indexed), or 3 (1-indexed)
