package consistent_test_filename

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed consistent_test_filename.schema.json
var schemaJSON []byte

// Rstest discovers tests by filename glob, defaulting to
// `**/*.{test,spec}.?(c|m)[jt]s?(x)`. Both defaults below accept the same
// extensions that glob does, so a file Rstest actually runs is never mistaken
// for an ordinary source file.
const (
	defaultPattern        = `.*\.test\.(c|m)?[tj]sx?$`
	defaultAllTestPattern = `.*\.(test|spec)\.(c|m)?[tj]sx?$`
)

type options struct {
	pattern        *esregexp.RegExp
	patternSource  string
	allTestPattern *esregexp.RegExp
}

// parseOptions reports ok=false when either pattern fails to compile. The
// schema already rejects an option that is not a valid regular expression, so
// this only guards against the two engines disagreeing; the rule then stays
// silent rather than reporting against a convention it could not honour.
func parseOptions(rawOptions []any) (options, bool) {
	patternSource := defaultPattern
	allTestPatternSource := defaultAllTestPattern

	if len(rawOptions) > 0 {
		optsMap, _ := rawOptions[0].(map[string]any)
		if v, ok := optsMap["pattern"].(string); ok {
			patternSource = v
		}
		if v, ok := optsMap["allTestPattern"].(string); ok {
			allTestPatternSource = v
		}
	}

	pattern, err := esregexp.Compile(patternSource, "")
	if err != nil {
		return options{}, false
	}
	allTestPattern, err := esregexp.Compile(allTestPatternSource, "")
	if err != nil {
		return options{}, false
	}

	return options{
		pattern:        pattern,
		patternSource:  patternSource,
		allTestPattern: allTestPattern,
	}, true
}

func buildConsistentTestFilenameMessage(patternSource string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "consistentTestFilename",
		Description: "Use test file name pattern " + patternSource,
	}
}

var ConsistentTestFilenameRule = rule.Rule{
	Name:   "rstest/consistent-test-filename",
	Schema: rule.NewSchema(schemaJSON),
	// The rule is decided entirely by the file's path — nothing in the source
	// text can change the answer. `Run` is invoked once per file, so the work
	// happens here and no listener is registered. (The linter's visitor walks
	// the source file's children but never the source file node itself, so an
	// `ast.KindSourceFile` listener would silently never fire.)
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.SourceFile == nil {
			return nil
		}
		fileName := ctx.SourceFile.FileName()
		if fileName == "" {
			return nil
		}

		opts, ok := parseOptions(options)
		if !ok {
			return nil
		}

		if !opts.allTestPattern.TestOrTimeout(fileName) {
			return nil
		}
		if opts.pattern.TestOrTimeout(fileName) {
			return nil
		}

		// Upstream reports on the `Program` node, so the diagnostic spans the
		// whole file. `ReportNode` would trim leading trivia off that range, so
		// the source file's own range is built directly.
		ctx.ReportRange(
			core.NewTextRange(0, ctx.SourceFile.End()),
			buildConsistentTestFilenameMessage(opts.patternSource),
		)
		return nil
	},
}
