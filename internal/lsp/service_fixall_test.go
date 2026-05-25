package lsp

import (
	"context"
	"sort"
	"testing"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// fixGroup represents an atomic set of text edits from a single diagnostic.
// Used by mergeNonOverlappingFixGroups tests.
type fixGroup struct {
	edits    []*lsproto.TextEdit
	minStart int
	maxEnd   int
}

// mergeNonOverlappingFixGroups sorts fix groups by start position and returns
// the merged text edits from non-overlapping groups using a greedy algorithm.
func mergeNonOverlappingFixGroups(groups []fixGroup) []*lsproto.TextEdit {
	if len(groups) == 0 {
		return nil
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].minStart < groups[j].minStart
	})
	var allEdits []*lsproto.TextEdit
	lastEnd := 0
	for _, g := range groups {
		if g.minStart >= lastEnd {
			allEdits = append(allEdits, g.edits...)
			lastEnd = g.maxEnd
		}
	}
	return allEdits
}

// ======== isFixAllRequest tests ========

func TestIsFixAllRequest_Nil(t *testing.T) {
	if isFixAllRequest(nil) {
		t.Error("nil context should not be fixAll")
	}
}

func TestIsFixAllRequest_NoOnly(t *testing.T) {
	ctx := &lsproto.CodeActionContext{}
	if isFixAllRequest(ctx) {
		t.Error("context without Only should not be fixAll")
	}
}

func TestIsFixAllRequest_QuickFixOnly(t *testing.T) {
	only := []lsproto.CodeActionKind{lsproto.CodeActionKindQuickFix}
	ctx := &lsproto.CodeActionContext{Only: &only}
	if isFixAllRequest(ctx) {
		t.Error("quickfix-only context should not be fixAll")
	}
}

func TestIsFixAllRequest_SourceFixAll(t *testing.T) {
	only := []lsproto.CodeActionKind{lsproto.CodeActionKindSourceFixAll}
	ctx := &lsproto.CodeActionContext{Only: &only}
	if !isFixAllRequest(ctx) {
		t.Error("source.fixAll should be recognized as fixAll")
	}
}

func TestIsFixAllRequest_SourceFixAllRslint(t *testing.T) {
	only := []lsproto.CodeActionKind{"source.fixAll.rslint"}
	ctx := &lsproto.CodeActionContext{Only: &only}
	if !isFixAllRequest(ctx) {
		t.Error("source.fixAll.rslint should be recognized as fixAll")
	}
}

func TestIsFixAllRequest_MixedKinds(t *testing.T) {
	only := []lsproto.CodeActionKind{lsproto.CodeActionKindQuickFix, "source.fixAll.rslint"}
	ctx := &lsproto.CodeActionContext{Only: &only}
	if !isFixAllRequest(ctx) {
		t.Error("mixed kinds containing source.fixAll.rslint should be fixAll")
	}
}

func TestIsFixAllRequest_OtherSourceFixAll(t *testing.T) {
	only := []lsproto.CodeActionKind{"source.fixAll.eslint"}
	ctx := &lsproto.CodeActionContext{Only: &only}
	if isFixAllRequest(ctx) {
		t.Error("source.fixAll.eslint should not be recognized as rslint fixAll")
	}
}

func TestIsFixAllRequest_EmptyOnly(t *testing.T) {
	only := []lsproto.CodeActionKind{}
	ctx := &lsproto.CodeActionContext{Only: &only}
	if isFixAllRequest(ctx) {
		t.Error("empty Only array should not be fixAll")
	}
}

func TestIsFixAllRequest_SourceOrganizeImports(t *testing.T) {
	only := []lsproto.CodeActionKind{lsproto.CodeActionKindSourceOrganizeImports}
	ctx := &lsproto.CodeActionContext{Only: &only}
	if isFixAllRequest(ctx) {
		t.Error("source.organizeImports should not be fixAll")
	}
}

// ======== handleFixAllCodeAction tests ========

func TestHandleFixAllCodeAction_SessionNil(t *testing.T) {
	s := newTestServer()
	resp, err := s.handleFixAllCodeAction(context.Background(), "file:///project/test.ts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CommandOrCodeActionArray == nil || len(*resp.CommandOrCodeActionArray) != 0 {
		t.Error("expected empty code actions when session is nil")
	}
}

func TestHandleFixAllCodeAction_NonTSFile(t *testing.T) {
	s := newTestServer()
	resp, err := s.handleFixAllCodeAction(context.Background(), "file:///project/styles.css")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CommandOrCodeActionArray == nil || len(*resp.CommandOrCodeActionArray) != 0 {
		t.Error("expected empty code actions for non-TS file")
	}
}

func TestHandleFixAllCodeAction_ClearsPendingLintURI(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const x = 1;"
	s.pendingLintURIs[uri] = struct{}{}

	_, _ = s.handleFixAllCodeAction(context.Background(), uri)

	if _, ok := s.pendingLintURIs[uri]; ok {
		t.Error("fixAll should clear pendingLintURIs for the target URI")
	}
}

func TestHandleFixAllCodeAction_DoesNotUpdateDiagnosticsCache(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const x = 1;"

	oldDiags := []rule.RuleDiagnostic{{RuleName: "old-rule"}}
	s.diagnostics[uri] = oldDiags

	_, _ = s.handleFixAllCodeAction(context.Background(), uri)

	cachedDiags := s.diagnostics[uri]
	if len(cachedDiags) != 1 || cachedDiags[0].RuleName != "old-rule" {
		t.Errorf("fixAll should not modify s.diagnostics, got %v", cachedDiags)
	}
}

func TestHandleFixAllCodeAction_OtherPendingURIsPreserved(t *testing.T) {
	s := newTestServer()
	uriA := lsproto.DocumentUri("file:///project/a.ts")
	uriB := lsproto.DocumentUri("file:///project/b.ts")
	s.documents[uriA] = "const a = 1;"
	s.pendingLintURIs[uriA] = struct{}{}
	s.pendingLintURIs[uriB] = struct{}{}

	_, _ = s.handleFixAllCodeAction(context.Background(), uriA)

	if _, ok := s.pendingLintURIs[uriA]; ok {
		t.Error("fixAll target URI should be cleared from pendingLintURIs")
	}
	if _, ok := s.pendingLintURIs[uriB]; !ok {
		t.Error("other URIs in pendingLintURIs should be preserved")
	}
}

// ======== handleDidSave pendingLintURIs optimization tests ========

func TestHandleDidSave_ClearsPendingLintURI(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const x = 1;"
	s.pendingLintURIs[uri] = struct{}{}

	savedText := "const x = 1;"
	_ = s.handleDidSave(ctx, &lsproto.DidSaveTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Text:         &savedText,
	})

	if _, ok := s.pendingLintURIs[uri]; ok {
		t.Error("didSave should clear pendingLintURIs for the saved URI")
	}
}

func TestHandleDidSave_OtherPendingURIsPreserved(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uriA := lsproto.DocumentUri("file:///project/a.ts")
	uriB := lsproto.DocumentUri("file:///project/b.ts")
	s.documents[uriA] = "const a = 1;"
	s.documents[uriB] = "const b = 1;"
	s.pendingLintURIs[uriA] = struct{}{}
	s.pendingLintURIs[uriB] = struct{}{}

	savedText := "const a = 1;"
	_ = s.handleDidSave(ctx, &lsproto.DidSaveTextDocumentParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uriA},
		Text:         &savedText,
	})

	if _, ok := s.pendingLintURIs[uriA]; ok {
		t.Error("saved URI should be removed from pendingLintURIs")
	}
	if _, ok := s.pendingLintURIs[uriB]; !ok {
		t.Error("other URIs should remain in pendingLintURIs")
	}
}

// ======== mergeNonOverlappingFixGroups tests ========

func makeEdit(startLine, startChar, endLine, endChar uint32, text string) *lsproto.TextEdit {
	return &lsproto.TextEdit{
		Range: lsproto.Range{
			Start: lsproto.Position{Line: startLine, Character: startChar},
			End:   lsproto.Position{Line: endLine, Character: endChar},
		},
		NewText: text,
	}
}

func TestMergeFixGroups_Empty(t *testing.T) {
	result := mergeNonOverlappingFixGroups(nil)
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
	result = mergeNonOverlappingFixGroups([]fixGroup{})
	if result != nil {
		t.Errorf("expected nil for empty slice, got %v", result)
	}
}

func TestMergeFixGroups_SingleGroup(t *testing.T) {
	groups := []fixGroup{
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 3, "let")}, minStart: 0, maxEnd: 3},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 1 || result[0].NewText != "let" {
		t.Errorf("expected 1 edit 'let', got %v", result)
	}
}

func TestMergeFixGroups_NoOverlap(t *testing.T) {
	groups := []fixGroup{
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 3, "let")}, minStart: 0, maxEnd: 3},
		{edits: []*lsproto.TextEdit{makeEdit(1, 0, 1, 3, "const")}, minStart: 10, maxEnd: 13},
		{edits: []*lsproto.TextEdit{makeEdit(2, 0, 2, 3, "var")}, minStart: 20, maxEnd: 23},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 3 {
		t.Fatalf("expected 3 edits, got %d", len(result))
	}
}

func TestMergeFixGroups_FullOverlap(t *testing.T) {
	groups := []fixGroup{
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 10, "A")}, minStart: 0, maxEnd: 10},
		{edits: []*lsproto.TextEdit{makeEdit(0, 5, 0, 15, "B")}, minStart: 5, maxEnd: 15},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 1 || result[0].NewText != "A" {
		t.Errorf("expected 1 edit 'A', got %v", result)
	}
}

func TestMergeFixGroups_PartialOverlap(t *testing.T) {
	groups := []fixGroup{
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 10, "A")}, minStart: 0, maxEnd: 10},
		{edits: []*lsproto.TextEdit{makeEdit(0, 8, 0, 20, "B")}, minStart: 8, maxEnd: 20},
		{edits: []*lsproto.TextEdit{makeEdit(1, 0, 1, 5, "C")}, minStart: 25, maxEnd: 30},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 2 || result[0].NewText != "A" || result[1].NewText != "C" {
		t.Errorf("expected A,C, got %v", result)
	}
}

func TestMergeFixGroups_AdjacentNotOverlapping(t *testing.T) {
	groups := []fixGroup{
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 10, "A")}, minStart: 0, maxEnd: 10},
		{edits: []*lsproto.TextEdit{makeEdit(0, 10, 0, 20, "B")}, minStart: 10, maxEnd: 20},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 2 {
		t.Fatalf("expected 2 edits (adjacent OK), got %d", len(result))
	}
}

func TestMergeFixGroups_UnsortedInput(t *testing.T) {
	groups := []fixGroup{
		{edits: []*lsproto.TextEdit{makeEdit(2, 0, 2, 5, "C")}, minStart: 30, maxEnd: 35},
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 5, "A")}, minStart: 0, maxEnd: 5},
		{edits: []*lsproto.TextEdit{makeEdit(1, 0, 1, 5, "B")}, minStart: 15, maxEnd: 20},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 3 || result[0].NewText != "A" || result[1].NewText != "B" || result[2].NewText != "C" {
		t.Errorf("expected A,B,C, got %v", result)
	}
}

func TestMergeFixGroups_AtomicGroup_MultipleEdits(t *testing.T) {
	groups := []fixGroup{
		{
			edits:    []*lsproto.TextEdit{makeEdit(0, 0, 0, 5, "A1"), makeEdit(1, 0, 1, 5, "A2")},
			minStart: 0, maxEnd: 25,
		},
		{edits: []*lsproto.TextEdit{makeEdit(0, 3, 0, 8, "B")}, minStart: 3, maxEnd: 8},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 2 || result[0].NewText != "A1" || result[1].NewText != "A2" {
		t.Errorf("expected A1,A2, got %v", result)
	}
}

func TestMergeFixGroups_AllOverlapping(t *testing.T) {
	groups := []fixGroup{
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 20, "A")}, minStart: 0, maxEnd: 20},
		{edits: []*lsproto.TextEdit{makeEdit(0, 5, 0, 10, "B")}, minStart: 5, maxEnd: 10},
		{edits: []*lsproto.TextEdit{makeEdit(0, 15, 0, 25, "C")}, minStart: 15, maxEnd: 25},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 1 || result[0].NewText != "A" {
		t.Errorf("expected 'A' only, got %v", result)
	}
}

func TestMergeFixGroups_SameStartPosition(t *testing.T) {
	groups := []fixGroup{
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 5, "A")}, minStart: 0, maxEnd: 5},
		{edits: []*lsproto.TextEdit{makeEdit(0, 0, 0, 10, "B")}, minStart: 0, maxEnd: 10},
	}
	result := mergeNonOverlappingFixGroups(groups)
	if len(result) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(result))
	}
}

// ======== handleCodeAction routing tests ========

func TestHandleCodeAction_RoutesToFixAll_WhenSourceFixAll(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const x = 1;"

	only := []lsproto.CodeActionKind{"source.fixAll.rslint"}
	resp, err := s.handleCodeAction(ctx, &lsproto.CodeActionParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Range:        lsproto.Range{},
		Context:      &lsproto.CodeActionContext{Only: &only},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CommandOrCodeActionArray == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestHandleCodeAction_RoutesToQuickFix_WhenNoOnly(t *testing.T) {
	s := newTestServer()
	ctx := context.Background()
	uri := lsproto.DocumentUri("file:///project/test.ts")
	s.documents[uri] = "const x = 1;"

	resp, err := s.handleCodeAction(ctx, &lsproto.CodeActionParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Range:        lsproto.Range{},
		Context:      &lsproto.CodeActionContext{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CommandOrCodeActionArray == nil {
		t.Fatal("expected non-nil response")
	}
}

// ======== computeEndPosition tests ========

func TestComputeEndPosition_Empty(t *testing.T) {
	line, char := computeEndPosition("")
	if line != 0 || char != 0 {
		t.Errorf("empty: got (%d, %d), want (0, 0)", line, char)
	}
}

func TestComputeEndPosition_SingleLine(t *testing.T) {
	line, char := computeEndPosition("hello")
	if line != 0 || char != 5 {
		t.Errorf("single line: got (%d, %d), want (0, 5)", line, char)
	}
}

func TestComputeEndPosition_MultipleLines(t *testing.T) {
	line, char := computeEndPosition("line1\nline2\nline3")
	if line != 2 || char != 5 {
		t.Errorf("3 lines: got (%d, %d), want (2, 5)", line, char)
	}
}

func TestComputeEndPosition_TrailingNewline(t *testing.T) {
	line, char := computeEndPosition("line1\nline2\n")
	if line != 2 || char != 0 {
		t.Errorf("trailing newline: got (%d, %d), want (2, 0)", line, char)
	}
}

func TestComputeEndPosition_OnlyNewlines(t *testing.T) {
	line, char := computeEndPosition("\n\n\n")
	if line != 3 || char != 0 {
		t.Errorf("only newlines: got (%d, %d), want (3, 0)", line, char)
	}
}

func TestComputeEndPosition_WindowsLineEndings(t *testing.T) {
	line, char := computeEndPosition("line1\r\nline2\r\nline3")
	if line != 2 || char != 5 {
		t.Errorf("CRLF: got (%d, %d), want (2, 5)", line, char)
	}
}

func TestComputeEndPosition_WindowsTrailingCRLF(t *testing.T) {
	line, char := computeEndPosition("line1\r\nline2\r\n")
	if line != 2 || char != 0 {
		t.Errorf("CRLF trailing: got (%d, %d), want (2, 0)", line, char)
	}
}

func TestComputeEndPosition_SingleChar(t *testing.T) {
	line, char := computeEndPosition("a")
	if line != 0 || char != 1 {
		t.Errorf("single char: got (%d, %d), want (0, 1)", line, char)
	}
}

func TestComputeEndPosition_Emoji(t *testing.T) {
	// 😂 is U+1F602 — a surrogate pair in UTF-16, counts as 2 code units
	line, char := computeEndPosition("ab😂cd")
	// a(1) b(1) 😂(2) c(1) d(1) = 6
	if line != 0 || char != 6 {
		t.Errorf("emoji: got (%d, %d), want (0, 6)", line, char)
	}
}

func TestComputeEndPosition_EmojiOnSecondLine(t *testing.T) {
	line, char := computeEndPosition("abc\néf😂")
	// line 1: e(1) f(1) 😂(2) = 4
	if line != 1 || char != 4 {
		t.Errorf("emoji on second line: got (%d, %d), want (1, 4)", line, char)
	}
}

func TestComputeEndPosition_EmojiWithTrailingNewline(t *testing.T) {
	line, char := computeEndPosition("abc\néf😂\n")
	if line != 2 || char != 0 {
		t.Errorf("emoji trailing newline: got (%d, %d), want (2, 0)", line, char)
	}
}

func TestComputeEndPosition_CJKCharacters(t *testing.T) {
	// CJK characters are in the BMP (U+4E00-U+9FFF), count as 1 UTF-16 code unit each
	line, char := computeEndPosition("你好世界")
	if line != 0 || char != 4 {
		t.Errorf("CJK: got (%d, %d), want (0, 4)", line, char)
	}
}

func TestComputeEndPosition_MixedASCIIAndMultiByte(t *testing.T) {
	// "const x = '😂';\n" — line 0: c,o,n,s,t, ,x, ,=, ,'(1),😂(2),'(1),;(1) = 15
	line, char := computeEndPosition("const x = '😂';\n")
	if line != 1 || char != 0 {
		t.Errorf("mixed with trailing newline: got (%d, %d), want (1, 0)", line, char)
	}
}

func TestComputeEndPosition_MultipleEmoji(t *testing.T) {
	// Each emoji outside BMP counts as 2 UTF-16 code units
	line, char := computeEndPosition("😂🎉🔥")
	// 😂(2) + 🎉(2) + 🔥(2) = 6
	if line != 0 || char != 6 {
		t.Errorf("multiple emoji: got (%d, %d), want (0, 6)", line, char)
	}
}

// ======== handleFixAllCodeAction edge cases ========

func TestHandleFixAllCodeAction_DocumentNotInMap(t *testing.T) {
	s := newTestServer()
	uri := lsproto.DocumentUri("file:///project/unknown.ts")
	// uri is NOT in s.documents — s.documents[uri] returns ""

	resp, err := s.handleFixAllCodeAction(context.Background(), uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Session is nil, so it returns empty without panicking
	if resp.CommandOrCodeActionArray == nil || len(*resp.CommandOrCodeActionArray) != 0 {
		t.Error("expected empty code actions for document not in map")
	}
}

// ======== maxFixPasses constant test ========

func TestMaxFixPasses_IsReasonable(t *testing.T) {
	if maxFixPasses < 1 || maxFixPasses > 100 {
		t.Errorf("maxFixPasses = %d, should be between 1 and 100", maxFixPasses)
	}
}
