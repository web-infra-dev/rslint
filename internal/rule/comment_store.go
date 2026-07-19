package rule

import (
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// CommentStore owns the canonical comment list for one linted source file.
//
// Collection is lazy because most rule/file pairs never inspect comments.
// All returns the same read-only, source-ordered slice after the first call.
// A CommentStore belongs to one lintFile invocation and is therefore used
// serially by that file's rule initializers and listeners.
type CommentStore struct {
	sourceFile *ast.SourceFile
	comments   []*ast.CommentRange
	scanned    bool
}

func NewCommentStore(sourceFile *ast.SourceFile) *CommentStore {
	return &CommentStore{sourceFile: sourceFile}
}

// All returns every real line or block comment in the source file, ordered by
// source position and deduplicated. Treat the returned slice as read-only.
func (s *CommentStore) All() []*ast.CommentRange {
	if s == nil || s.scanned {
		if s == nil {
			return nil
		}
		return s.comments
	}
	s.scanned = true

	if s.sourceFile == nil {
		return nil
	}
	text := s.sourceFile.Text()
	// Every comment recognized by the TypeScript scanner starts with one of
	// these byte pairs. False positives in strings, templates, regexes, or JSX
	// only take the established slow path; an absence proves there is no
	// comment and avoids materializing the file's token tree.
	if !mayContainComment(text) {
		return nil
	}

	utils.ForEachComment(s.sourceFile.AsNode(), func(comment *ast.CommentRange) {
		s.comments = append(s.comments, comment)
	}, s.sourceFile)

	// ForEachComment can surface the same physical comment twice (once as a
	// token's trailing range, once as the next token's leading range) and does
	// not guarantee source order.
	sort.Slice(s.comments, func(i, j int) bool {
		return s.comments[i].Pos() < s.comments[j].Pos()
	})
	s.comments = slices.CompactFunc(s.comments, func(a, b *ast.CommentRange) bool {
		return a.Pos() == b.Pos() && a.End() == b.End()
	})
	return s.comments
}

func mayContainComment(text string) bool {
	for searchStart := 0; searchStart < len(text); {
		offset := strings.IndexByte(text[searchStart:], '/')
		if offset < 0 {
			return false
		}
		slash := searchStart + offset
		if slash+1 < len(text) && (text[slash+1] == '/' || text[slash+1] == '*') {
			return true
		}
		searchStart = slash + 1
	}
	return false
}
