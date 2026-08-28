package reactutil

import "testing"

// TestJsxAnnotationInComment pins the port of upstream pragmaUtil's
// `/@jsx\s+([^\s]+)/` to the cases where a compiled RE2 pattern would answer
// differently, or where the marker is only a prefix of a longer annotation.
func TestJsxAnnotationInComment(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "plain", value: " @jsx Preact.h ", want: "Preact.h"},
		{name: "tab separated", value: " @jsx\tPreact ", want: "Preact"},
		{name: "newline separated", value: " @jsx\nPreact ", want: "Preact"},
		{name: "javascript-only whitespace", value: " @jsx Preact ", want: "Preact"},
		{name: "no whitespace after marker", value: " @jsxPreact ", want: ""},
		{name: "longer marker is not a match", value: " @jsxFrag Preact.Fragment ", want: ""},
		{name: "later annotation after a non-match", value: " @jsxFrag F @jsx Preact ", want: "Preact"},
		{name: "nothing after the marker", value: " @jsx ", want: ""},
		{name: "absent", value: " nothing to see ", want: ""},
		{name: "non-identifier is still an annotation", value: " @jsx Foo-Bar ", want: "Foo-Bar"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, found := jsxAnnotationInComment(testCase.value)
			if testCase.want == "" {
				if found {
					t.Fatalf("jsxAnnotationInComment(%q) = %q, true; want not found", testCase.value, got)
				}
				return
			}
			if !found || got != testCase.want {
				t.Fatalf("jsxAnnotationInComment(%q) = %q, %v; want %q, true", testCase.value, got, found, testCase.want)
			}
		})
	}
}

// TestIsJavaScriptIdentifier pins the port of upstream's
// `/^[_$a-zA-Z][_$a-zA-Z0-9]*$/`, which is ASCII-only and does not exclude
// reserved words.
func TestIsJavaScriptIdentifier(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name string
		want bool
	}{
		{name: "React", want: true},
		{name: "_private", want: true},
		{name: "$dollar", want: true},
		{name: "h2", want: true},
		{name: "class", want: true},
		{name: "", want: false},
		{name: "2h", want: false},
		{name: "Foo-Bar", want: false},
		{name: "Preact.h", want: false},
		{name: "\u00e9lan", want: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := isJavaScriptIdentifier(testCase.name); got != testCase.want {
				t.Fatalf("isJavaScriptIdentifier(%q) = %v, want %v", testCase.name, got, testCase.want)
			}
		})
	}
}
