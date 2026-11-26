package unicornutil

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// IsDisableOrEnableDirective reports an eslint or rslint disable/enable
// directive comment. Configuration and global comments are not included.
// value is the comment text between its delimiters.
func IsDisableOrEnableDirective(comment *ast.CommentRange, value string) bool {
	label := ecmascript.StringTrim(value)
	if end := strings.IndexFunc(label, ecmascript.IsWhiteSpaceOrLineTerminator); end >= 0 {
		label = label[:end]
	}
	switch label {
	case "eslint-disable", "eslint-enable", "rslint-disable", "rslint-enable":
		return comment.Kind == ast.KindMultiLineCommentTrivia
	case "eslint-disable-next-line", "rslint-disable-next-line":
		return true
	case "eslint-disable-line", "rslint-disable-line":
		return !strings.ContainsAny(value, ecmascript.LineTerminators)
	default:
		return false
	}
}

// closestJSDocTags mirrors Unicorn's getDocumentationTags. It follows tsgo's
// JSDoc owners from host, and the first owner with a leading comment other
// than a disable/enable directive decides: a JSDoc block supplies its tags,
// while an ordinary comment hides older documentation. Statement
// documentation describes only a single binding. source is the text that the
// returned tags' positions index.
func closestJSDocTags(ctx rule.RuleContext, host *ast.Node) (tags []*ast.Node, source string) {
	text := ctx.SourceFile.Text()
	for owner := host; owner != nil; owner = ast.GetNextJSDocCommentLocation(owner) {
		if ast.IsVariableDeclarationList(owner) && len(owner.AsVariableDeclarationList().Declarations.Nodes) > 1 {
			return nil, ""
		}
		start := owner.Pos()
		comments := utils.CommentsInSpan(ctx.Comments.All(), start, scanner.SkipTrivia(text, start))
		for i := len(comments) - 1; i >= 0; i-- {
			comment := comments[i]
			value := utils.CommentValue(text, comment)
			if IsDisableOrEnableDirective(comment, value) {
				continue
			}
			if comment.Kind != ast.KindMultiLineCommentTrivia || !strings.HasPrefix(value, "*") {
				return nil, ""
			}
			docs := owner.JSDoc(ctx.SourceFile)
			if len(docs) == 0 || docs[len(docs)-1].Pos() != comment.Pos() {
				// tsgo can attach a same-line JSDoc to the previous statement.
				source = text[comment.Pos():comment.End()]
				docs = parseCallableJSDoc(source, owner)
			} else {
				source = text
			}
			if len(docs) == 0 || docs[len(docs)-1].AsJSDoc().Tags == nil {
				return nil, ""
			}
			return docs[len(docs)-1].AsJSDoc().Tags.Nodes, source
		}
	}
	return nil, ""
}

// uniqueJSDocTag returns the only tag that matches, as Unicorn ignores an
// ambiguous annotation.
func uniqueJSDocTag(tags []*ast.Node, matches func(*ast.Node) bool) *ast.Node {
	var match *ast.Node
	for _, tag := range tags {
		if matches(tag) {
			if match != nil {
				return nil
			}
			match = tag
		}
	}
	return match
}
