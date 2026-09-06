package require_to_throw_message

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func TestShouldReport(t *testing.T) {
	matcherNode := &ast.Node{}
	matcherEntry := &testFramework.MemberEntry{Node: matcherNode}
	argument := &ast.Node{}
	tests := []struct {
		name string
		call *ExpectCall
		want bool
	}{
		{name: "toThrow", call: &ExpectCall{Matcher: "toThrow", MatcherEntry: matcherEntry}, want: true},
		{name: "toThrowError", call: &ExpectCall{Matcher: "toThrowError", MatcherEntry: matcherEntry}, want: true},
		{name: "argument", call: &ExpectCall{Matcher: "toThrow", MatcherEntry: matcherEntry, MatcherArgs: []*ast.Node{argument}}},
		{name: "not", call: &ExpectCall{Matcher: "toThrow", MatcherEntry: matcherEntry, Modifiers: []string{"not"}}},
		{name: "rejects", call: &ExpectCall{Matcher: "toThrow", MatcherEntry: matcherEntry, Modifiers: []string{"rejects"}}, want: true},
		{name: "other matcher", call: &ExpectCall{Matcher: "throw", MatcherEntry: matcherEntry}},
		{name: "missing matcher entry", call: &ExpectCall{Matcher: "toThrow"}},
		{name: "missing matcher node", call: &ExpectCall{Matcher: "toThrow", MatcherEntry: &testFramework.MemberEntry{}}},
		{name: "unparsed call"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldReport(test.call); got != test.want {
				t.Fatalf("shouldReport() = %t, want %t", got, test.want)
			}
		})
	}
}
