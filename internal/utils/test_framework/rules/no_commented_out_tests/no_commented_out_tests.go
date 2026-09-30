package no_commented_out_tests

import (
	"regexp"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// RootKind distinguishes test registrations from suite registrations.
type RootKind uint8

const (
	RootTest RootKind = iota + 1
	RootDescribe
)

// Root describes one identifier that opens a registration chain.
type Root struct {
	Kind RootKind
	// Extendable marks roots that accept `.extend(fixtures)`.
	Extendable bool
}

// MemberKind describes how a member access changes a registration chain.
type MemberKind uint8

const (
	// MemberModifier keeps the chain callable, e.g. `.only`.
	MemberModifier MemberKind = iota + 1
	// MemberConditionalFactory must be called once to return the API again,
	// e.g. `.skipIf(condition)`.
	MemberConditionalFactory
	// MemberParameterizedFactory returns a registrar when called or tagged,
	// e.g. `.each(rows)` or `` .each`table` ``.
	MemberParameterizedFactory
	// MemberExtendFactory returns another extendable API when called.
	MemberExtendFactory
)

// Member describes one member name accepted after a root.
type Member struct {
	Kind MemberKind
	// TestOnly restricts the member to RootTest roots.
	TestOnly bool
}

// Profile is the framework-specific syntax accepted for a commented-out
// registration. The engine owns comment reconstruction, parsing, registration
// checks and diagnostics; a profile only names the legal shapes.
type Profile struct {
	Roots   map[string]Root
	Members map[string]Member
	// AcceptUnknownRootMember additionally accepts one unlisted member
	// directly after a root, such as `test.futureModifier("x")`.
	AcceptUnknownRootMember bool

	rootNames     []string
	rootCandidate *regexp.Regexp
}

// Config configures a no-commented-out-tests rule for one framework.
type Config struct {
	Name    string
	Profile Profile
}

func (p *Profile) prepare() {
	p.rootNames = p.rootNames[:0]
	for name := range p.Roots {
		p.rootNames = append(p.rootNames, name)
	}
	sort.Strings(p.rootNames)
	quoted := make([]string, len(p.rootNames))
	for i, name := range p.rootNames {
		quoted[i] = regexp.QuoteMeta(name)
	}
	p.rootCandidate = regexp.MustCompile(
		`(?m)^[\t ]*[;(]*[\t ]*(` + strings.Join(quoted, "|") + `)\b`,
	)
}

func buildCommentedTestsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "commentedTests",
		Description: "Do not comment out tests",
	}
}

// NewRule creates a no-commented-out-tests rule for a test framework.
func NewRule(config Config) rule.Rule {
	profile := config.Profile
	profile.prepare()

	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			sourceText := ctx.SourceFile.Text()
			if !profile.mayContainCommentedRoot(sourceText) {
				return nil
			}
			for _, block := range buildCommentBlocks(sourceText, ctx.Comments.All()) {
				reportedComments := make(map[*ast.CommentRange]struct{})
				for _, offset := range profile.findCommentedRegistrations(block.text, ctx.SourceFile.ScriptKind) {
					comment := block.commentAt(offset)
					if comment == nil {
						continue
					}
					if _, reported := reportedComments[comment]; reported {
						continue
					}
					reportedComments[comment] = struct{}{}
					ctx.ReportRange(
						core.NewTextRange(comment.Pos(), comment.End()),
						buildCommentedTestsMessage(),
					)
				}
			}
			return nil
		},
	}
}
