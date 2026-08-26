// cspell:ignore FAQI Qpage
package filename_case

import (
	"slices"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestLatestUpstreamAcronymCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		fn    func(string) string
		want  string
	}{
		{name: "camel inner HTML", input: "innerHTML", fn: toCamelCaseWithAcronyms, want: "innerHTML"},
		{name: "camel DOM range", input: "getDOMRangeRect", fn: toCamelCaseWithAcronyms, want: "getDOMRangeRect"},
		{name: "camel URL suffix", input: "apiURL", fn: toCamelCaseWithAcronyms, want: "apiURL"},
		{name: "camel numbered acronym", input: "getHTML5Parser", fn: toCamelCaseWithAcronyms, want: "getHTML5Parser"},
		{name: "camel leading acronym", input: "HTMLParser", fn: toCamelCaseWithAcronyms, want: "htmlParser"},
		{name: "camel mixed acronym", input: "XMLHttpRequest", fn: toCamelCaseWithAcronyms, want: "xmlHttpRequest"},
		{name: "pascal FAQ", input: "FAQPage", fn: toPascalCaseWithLeadingAcronym, want: "FAQPage"},
		{name: "pascal DIY", input: "DIYWidget", fn: toPascalCaseWithLeadingAcronym, want: "DIYWidget"},
		{name: "pascal numbered separator", input: "URL2Path", fn: toPascalCaseWithLeadingAcronym, want: "URL2Path"},
		{name: "pascal numbered suffix", input: "FAQI18n", fn: toPascalCaseWithLeadingAcronym, want: "FAQI18n"},
		{name: "pascal acronym not leading", input: "PageFAQ", fn: toPascalCaseWithLeadingAcronym, want: "PageFaq"},
		{name: "pascal separated acronym", input: "FAQ-Page", fn: toPascalCaseWithLeadingAcronym, want: "FaqPage"},
		{name: "pascal lowercase suffix", input: "FAQpage", fn: toPascalCaseWithLeadingAcronym, want: "FaQpage"},
		{name: "pascal invalid suffix", input: "FAQPageFOO", fn: toPascalCaseWithLeadingAcronym, want: "FaqPageFoo"},
		{name: "pascal lowercase after number", input: "URL2path", fn: toPascalCaseWithLeadingAcronym, want: "Url2path"},
		{name: "pascal two-letter prefix", input: "UIPath", fn: toPascalCaseWithLeadingAcronym, want: "UiPath"},
		{name: "pascal two-letter numbered prefix", input: "UI2Path", fn: toPascalCaseWithLeadingAcronym, want: "Ui2Path"},
		{name: "pascal terminal acronym", input: "FOO2", fn: toPascalCaseWithLeadingAcronym, want: "Foo2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.fn(test.input); got != test.want {
				t.Fatalf("transform(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestGetPathSegments(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		cwd      string
		want     []string
	}{
		{name: "absolute inside cwd", fileName: "/repo/src/FooBar/file.js", cwd: "/repo", want: []string{"src", "FooBar", "file.js"}},
		{name: "relative inside cwd", fileName: "src/FooBar/file.js", cwd: "/repo", want: []string{"src", "FooBar", "file.js"}},
		{name: "outside cwd", fileName: "/other/Src/fooBar.js", cwd: "/repo", want: []string{"fooBar.js"}},
		{name: "cwd prefix collision", fileName: "/repository/Src/fooBar.js", cwd: "/repo", want: []string{"fooBar.js"}},
		{name: "root cwd", fileName: "/src/FooBar/file.js", cwd: "/", want: []string{"src", "FooBar", "file.js"}},
		{name: "dot segments use general path resolution", fileName: "/repo/src/../FooBar/file.js", cwd: "/repo", want: []string{"FooBar", "file.js"}},
		{name: "empty cwd", fileName: "/repo/src/FooBar/file.js", want: []string{"file.js"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := getPathSegments(test.fileName, test.cwd)
			if len(got) != len(test.want) {
				t.Fatalf("getPathSegments(%q, %q) = %v, want %v", test.fileName, test.cwd, got, test.want)
			}
			for index := range got {
				if got[index] != test.want[index] {
					t.Fatalf("getPathSegments(%q, %q) = %v, want %v", test.fileName, test.cwd, got, test.want)
				}
			}
		})
	}
}

func TestCanonicalPOSIXPathSegments(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		cwd      string
		want     []string
		wantFast bool
	}{
		{name: "inside cwd", fileName: "/repo/src/file.js", cwd: "/repo", want: []string{"src", "file.js"}, wantFast: true},
		{name: "root cwd", fileName: "/src/file.js", cwd: "/", want: []string{"src", "file.js"}, wantFast: true},
		{name: "outside cwd", fileName: "/other/src/file.js", cwd: "/repo", want: []string{"file.js"}, wantFast: true},
		{name: "cwd prefix collision", fileName: "/repository/file.js", cwd: "/repo", want: []string{"file.js"}, wantFast: true},
		{name: "same path", fileName: "/repo", cwd: "/repo", want: []string{"repo"}, wantFast: true},
		{name: "relative filename", fileName: "src/file.js", cwd: "/repo"},
		{name: "relative cwd", fileName: "/repo/src/file.js", cwd: "repo"},
		{name: "dot segment", fileName: "/repo/src/../file.js", cwd: "/repo"},
		{name: "duplicate separator", fileName: "/repo//file.js", cwd: "/repo"},
		{name: "untitled root segment", fileName: "/^/file.js", cwd: "/"},
		{name: "DOS volume segment", fileName: "/repo/C:/file.js", cwd: "/repo"},
		{name: "trailing cwd separator", fileName: "/repo/file.js", cwd: "/repo/"},
		{name: "backslash", fileName: `/repo/src\file.js`, cwd: "/repo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := getCanonicalPOSIXPathSegments(test.fileName, test.cwd)
			if ok != test.wantFast {
				t.Fatalf("fast path = %v, want %v", ok, test.wantFast)
			}
			if len(got) != len(test.want) {
				t.Fatalf("segments = %v, want %v", got, test.want)
			}
			for index := range got {
				if got[index] != test.want[index] {
					t.Fatalf("segments = %v, want %v", got, test.want)
				}
			}
			if ok {
				slow := getPathSegmentsWithTspath(test.fileName, test.cwd)
				if !slices.Equal(got, slow) {
					t.Fatalf("fast segments = %v, tspath segments = %v", got, slow)
				}
			}
		})
	}
}

func FuzzCanonicalPOSIXPathSegments(f *testing.F) {
	seeds := [][2]string{
		{"/repo/src/file.js", "/repo"},
		{"/repo", "/repo"},
		{"/repository/file.js", "/repo"},
		{"/src/file.js", "/"},
		{"/repo/目录/file.js", "/repo"},
		{"/$route/a-b_0/file.js", "/$route"},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1])
	}

	f.Fuzz(func(t *testing.T, fileName string, cwd string) {
		got, ok := getCanonicalPOSIXPathSegments(fileName, cwd)
		if !ok {
			return
		}
		want := getPathSegmentsWithTspath(fileName, cwd)
		if !slices.Equal(got, want) {
			t.Fatalf("fast segments = %v, tspath segments = %v for file %q cwd %q", got, want, fileName, cwd)
		}
	})
}

func TestLatestUpstreamPathBehavior(t *testing.T) {
	tests := []struct {
		name    string
		cwd     string
		file    string
		options []any
		wantID  string
		want    string
	}{
		{
			name:   "checks directory",
			cwd:    "/repo",
			file:   "/repo/src/FooBar/file.js",
			wantID: "directory-case",
			want:   "Directory name `FooBar` is not in kebab case. Rename it to `foo-bar`.",
		},
		{
			name:   "checks index parent directory",
			cwd:    "/repo",
			file:   "/repo/src/FooBar/index.js",
			wantID: "directory-case",
			want:   "Directory name `FooBar` is not in kebab case. Rename it to `foo-bar`.",
		},
		{
			name:   "checks complete dotted directory segment",
			cwd:    "/repo",
			file:   "/repo/src/foo.JS/bar.js",
			wantID: "directory-case",
			want:   "Directory name `foo.JS` is not in kebab case. Rename it to `foo.js`.",
		},
		{
			name: "directory checking disabled",
			cwd:  "/repo",
			file: "/repo/src/FooBar/foo_bar.js",
			options: []any{map[string]any{
				"case": "kebabCase", "checkDirectories": false,
			}},
			wantID: "filename-case",
			want:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
		},
		{
			name: "ignore matches directory segment",
			cwd:  "/repo",
			file: "/repo/src/meta/BadName.js",
			options: []any{map[string]any{
				"case": "kebabCase", "ignore": []any{"^meta$"},
			}},
		},
		{
			name:   "dollar directory skipped",
			cwd:    "/repo",
			file:   "/repo/src/$UserId/fooBar.js",
			wantID: "filename-case",
			want:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
		},
		{
			name:   "dollar filename still checks extension",
			cwd:    "/repo",
			file:   "/repo/src/foo/$userId.TSX",
			wantID: "filename-extension",
			want:   "File extension `.TSX` is not in lowercase. Rename it to `$userId.tsx`.",
		},
		{
			name: "dollar filename skips case",
			cwd:  "/repo",
			file: "/repo/src/foo/$foo_bar.js",
		},
		{
			name:   "outside cwd checks basename only",
			cwd:    "/repo",
			file:   "/other/Src/fooBar.js",
			wantID: "filename-case",
			want:   "Filename is not in kebab case. Rename it to `foo-bar.js`.",
		},
		{
			name:   "first invalid directory wins",
			cwd:    "/repo",
			file:   "/repo/src/FooBar/BazQux.js",
			wantID: "directory-case",
			want:   "Directory name `FooBar` is not in kebab case. Rename it to `foo-bar`.",
		},
		{
			name: "canonical case order applies to directories",
			cwd:  "/repo",
			file: "/repo/src/foo-bar/file.js",
			options: []any{map[string]any{"cases": map[string]any{
				"pascalCase": true, "camelCase": true,
			}}},
			wantID: "directory-case",
			want:   "Directory name `foo-bar` is not in camel case or pascal case. Rename it to `fooBar` or `FooBar`.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diagnostics := runFilenameCaseForPath(t, test.cwd, test.file, test.options)
			if test.want == "" {
				if len(diagnostics) != 0 {
					t.Fatalf("diagnostics = %v, want none", diagnostics)
				}
				return
			}
			if len(diagnostics) != 1 {
				t.Fatalf("diagnostics = %v, want one", diagnostics)
			}
			if got := diagnostics[0].Message.Id; got != test.wantID {
				t.Fatalf("message id = %q, want %q", got, test.wantID)
			}
			if got := diagnostics[0].Message.Description; got != test.want {
				t.Fatalf("message = %q, want %q", got, test.want)
			}
		})
	}
}

func runFilenameCaseForPath(t *testing.T, cwd string, fileName string, options []any) []rule.RuleDiagnostic {
	t.Helper()
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: fileName,
		Path:     tspath.Path(fileName),
	}, "export {};", core.ScriptKindTS)
	var diagnostics []rule.RuleDiagnostic
	ctx := (rule.RuleContext{
		SourceFile:     sourceFile,
		DisableManager: rule.NewDisableManager(sourceFile, rule.NewCommentStore(sourceFile)),
	}).WithFileCache(rule.NewFileCacheWithProcessCurrentDirectory(cwd)).WithReporter(
		filenameCaseRuleName,
		rule.SeverityWarning,
		func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	)
	FilenameCaseRule.Run(ctx, options)
	return diagnostics
}
