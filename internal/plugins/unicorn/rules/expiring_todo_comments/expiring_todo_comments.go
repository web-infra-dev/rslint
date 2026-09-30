package expiring_todo_comments

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/npmsemver"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
	"github.com/web-infra-dev/rslint/internal/utils/warningcomments"
)

//go:embed expiring_todo_comments.schema.json
var schemaJSON []byte

// ExpiringTodoCommentsRule follows eslint-plugin-unicorn v76.0.0.
var ExpiringTodoCommentsRule = rule.Rule{
	Name:   "unicorn/expiring-todo-comments",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		comments := ctx.Comments.All()
		if len(comments) == 0 {
			return rule.RuleListeners{}
		}
		opts := parseOptions(options)
		pkg := packagejson.FindNearest(ctx.Program(), ctx.SourceFile.FileName())
		// Upstream parseArgument checks package availability from cwd, while
		// comparisons read the package nearest the linted file.
		cwd := ctx.ProcessCurrentDirectory()
		if cwd == "" {
			cwd = ctx.Program().CurrentDirectory()
		}
		hasPackage := packagejson.FindNearest(ctx.Program(), tspath.ResolvePath(cwd, "__rslint__.js")) != nil
		var patterns []*esregexp.RegExp
		if !opts.allowWarningComments {
			for _, term := range opts.terms {
				// Unicorn passes no location to the core rule, which selects anywhere.
				patterns = append(patterns, warningcomments.Pattern(term, "anywhere", ""))
			}
		}
		for _, comment := range comments {
			value := utils.CommentValue(ctx.SourceFile.Text(), comment)
			trimmed := ecmascript.StringTrim(value)
			if utils.IsDirectiveComment(comment.Kind, trimmed) {
				pattern := disableDirective
				// ESLint only treats disable-line/disable-next-line as line
				// directives; rslint also accepts its own block controls here.
				if comment.Kind == ast.KindSingleLineCommentTrivia && strings.HasPrefix(trimmed, "eslint-") {
					pattern = lineDisableDirective
				}
				if pattern.Test(value) {
					continue
				}
			}
			var unused []string
			for line := range strings.SplitSeq(value, "\n") {
				ignored := false
				for _, pattern := range opts.ignore {
					if pattern.TestOrTimeout(line) {
						ignored = true
						break
					}
				}
				if ignored {
					continue
				}
				report := func(id, message string) {
					ctx.ReportRange(core.NewTextRange(comment.Pos(), comment.End()), rule.RuleMessage{Id: id, Description: message})
				}
				if processComment(line, opts, hasPackage, pkg, report) || opts.allowWarningComments {
					continue
				}
				if utils.IsDirectiveComment(comment.Kind, ecmascript.StringTrim(line)) && selfConfig.Test(line) {
					continue
				}
				unused = append(unused, line)
			}
			// Upstream checks expiration conditions before its warning fallback.
			// Several lines of a block share one diagnostic range.
			for _, line := range unused {
				for i, pattern := range patterns {
					if pattern.Test(line) {
						ctx.ReportRange(core.NewTextRange(comment.Pos(), comment.End()), rule.RuleMessage{
							Id:          "unexpectedComment",
							Description: fmt.Sprintf("Unexpected '%s': '%s'.", opts.terms[i], warningcomments.Display(line)),
						})
					}
				}
			}
		}
		return rule.RuleListeners{}
	},
}

type ruleOptions struct {
	terms                            []string
	ignore                           []*esregexp.RegExp
	date                             string
	checkDates, allowWarningComments bool
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{terms: []string{"todo", "fixme", "xxx"}, date: time.Now().UTC().Format(time.DateOnly), allowWarningComments: true}
	if len(options) == 0 {
		return opts
	}
	object, _ := options[0].(map[string]any)
	if terms, ok := object["terms"].([]any); ok {
		opts.terms = nil
		for _, term := range terms {
			if text, ok := term.(string); ok {
				opts.terms = append(opts.terms, text)
			}
		}
	}
	if date, ok := object["date"].(string); ok {
		opts.date = date
	}
	if value, ok := object["allowWarningComments"].(bool); ok {
		opts.allowWarningComments = value
	}
	opts.checkDates, _ = object["checkDates"].(bool)
	patterns, _ := object["ignore"].([]any)
	for _, value := range patterns {
		pattern := ecmascript.JSONValueToString(value)
		if compiled, err := esregexp.Compile(pattern, "u"); err == nil {
			opts.ignore = append(opts.ignore, compiled)
		}
	}
	return opts
}

var (
	argumentsPattern      = esregexp.MustCompile(`\[([^}]+)]`, "")
	datePattern           = esregexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`, "")
	inclusionPattern      = esregexp.MustCompile(`^[+-]\s*(\S+)`, "")
	comparisonPattern     = esregexp.MustCompile(`^(\S+)@(>|>=)(\d+(?:\.\d+){0,2}(?:-[\d\-a-z]+(?:\.[\d\-a-z]+)*)?(?:\+[\d\-a-z]+(?:\.[\d\-a-z]+)*)?)`, "i")
	packageVersionPattern = esregexp.MustCompile(`^(>|>=)(\d+(?:\.\d+){0,2}(?:-[\d\-a-z]+(?:\.[\d\-a-z]+)*)?(?:\+[\d\-a-z]+(?:\.[\d\-a-z]+)*)?)\s*$`, "")
	disableDirective      = esregexp.MustCompile(`^\s*(?:eslint|rslint)-(?:enable|disable(?:-next-line|-line)?)(?:\s|$)`, "u")
	lineDisableDirective  = esregexp.MustCompile(`^\s*eslint-disable-(?:next-line|line)(?:\s|$)`, "u")
	selfConfig            = esregexp.MustCompile(`\bno-warning-comments\b`, "u")
)

type argument struct{ kind, name, condition, version string }

func captures(pattern *esregexp.RegExp, text string) []string {
	match, _ := pattern.Unwrap().FindStringMatch(text)
	if match == nil {
		return nil
	}
	groups := match.Groups()
	result := make([]string, len(groups))
	for i, group := range groups {
		result[i] = group.String()
	}
	return result
}

func parseArgument(text string, hasPackage bool) argument {
	if datePattern.Test(text) {
		return argument{kind: "dates", version: text}
	}
	if hasPackage {
		// Upstream's non-Unicode \S{2,} counts UTF-16 units, so one emoji
		// qualifies. Keep that length check outside the code-point matcher.
		if match := captures(inclusionPattern, text); match != nil && ecmascript.StringCodeUnitCount(match[1]) >= 2 {
			return argument{kind: "dependencies", name: ecmascript.StringTrim(text[1:]), condition: text[:1]}
		}
		if match := captures(comparisonPattern, text); match != nil && ecmascript.StringCodeUnitCount(match[1]) >= 2 {
			arg := argument{kind: "dependencies", name: match[1], condition: match[2], version: match[3]}
			switch {
			case arg.name == "engine:node":
				arg.kind = "engines"
			case strings.HasPrefix(arg.name, "engine:"):
				return argument{kind: "unknowns", name: text}
			case strings.HasPrefix(arg.name, "peer:"):
				arg.kind = "peerDependencies"
				arg.name = strings.TrimPrefix(arg.name, "peer:")
			}
			return arg
		}
		if match := captures(packageVersionPattern, text); match != nil {
			return argument{kind: "packageVersions", condition: match[1], version: match[2]}
		}
	}
	return argument{kind: "unknowns", name: text}
}

func processComment(text string, opts ruleOptions, hasPackage bool, pkg *packagejson.Package, report func(string, string)) bool {
	if !strings.Contains(text, "[") {
		return false
	}
	lower := ecmascript.StringToLowerCase(text)
	hasTerm := false
	for _, term := range opts.terms {
		if strings.Contains(lower, ecmascript.StringToLowerCase(term)) {
			hasTerm = true
			break
		}
	}
	if !hasTerm {
		return false
	}
	match := captures(argumentsPattern, text)
	if match == nil {
		return false
	}
	groups := make(map[string][]argument)
	for raw := range strings.SplitSeq(match[1], ",") {
		arg := parseArgument(ecmascript.StringTrim(raw), hasPackage)
		groups[arg.kind] = append(groups[arg.kind], arg)
	}
	message := ecmascript.StringTrim(text[strings.IndexByte(text, ']')+1:])
	message = ecmascript.StringTrim(strings.TrimPrefix(message, ":"))
	emit := func(id, prefix string) { report("unicorn/"+id, prefix+". "+message) }
	field := func(names ...string) any {
		if pkg == nil {
			return nil
		}
		return pkg.Field(names...)
	}
	used := len(groups) > 1 || len(groups["unknowns"]) == 0
	if dates := groups["dates"]; len(dates) > 1 {
		values := make([]string, len(dates))
		for i, date := range dates {
			values[i] = date.version
		}
		emit("avoidMultipleDates", "Avoid using multiple expiration dates: "+strings.Join(values, ", "))
	} else if len(dates) == 1 && opts.checkDates && reachedDate(dates[0].version, opts.date) {
		emit("expiredTodo", "Past due date: "+dates[0].version)
	}
	if versions := groups["packageVersions"]; len(versions) > 1 {
		values := make([]string, len(versions))
		for i, v := range versions {
			values[i] = v.condition + v.version
		}
		emit("avoidMultiplePackageVersions", "Avoid using multiple package versions: "+strings.Join(values, ", "))
	} else if len(versions) == 1 && versionMatches(field("version"), versions[0]) {
		emit("reachedPackageVersion", "Past due package version: "+versions[0].condition+versions[0].version)
	}
	for _, dependency := range groups["dependencies"] {
		raw := field("dependencies", dependency.name)
		if dev, ok := field("devDependencies").(map[string]any); ok {
			if value, exists := dev[dependency.name]; exists {
				raw = value
			}
		}
		switch dependency.condition {
		case "+":
			if truthy(raw) {
				emit("havePackage", "Due since "+dependency.name+" was installed")
			}
		case "-":
			if !truthy(raw) {
				emit("dontHavePackage", "Due since "+dependency.name+" was removed")
			}
		default:
			if isCatalog(raw) {
				emit("unsupportedCatalogProtocol", "Cannot check dependency version because "+dependency.name+" uses the unsupported `catalog:` protocol")
			} else if versionMatches(raw, dependency) {
				emit("versionMatches", "Due since package version matched: "+dependency.name+" "+dependency.condition+" "+dependency.version)
			}
		}
	}
	for _, peer := range groups["peerDependencies"] {
		raw, _ := field("peerDependencies", peer.name).(string)
		if isCatalog(raw) {
			emit("unsupportedCatalogProtocol", "Cannot check peer dependency version because "+peer.name+" uses the unsupported `catalog:` protocol")
			continue
		}
		if versions, ok := npmsemver.Parse(raw); ok && raw != "" {
			if floor, ok := versions.MinVersion(); ok && npmsemver.SatisfiesComparison(floor, peer.condition+peer.version) {
				emit("peerVersionMatches", "Due since peer dependency version matched: "+peer.name+" "+peer.condition+" "+peer.version)
			}
		}
	}
	for _, engine := range groups["engines"] {
		if versionMatches(field("engines", "node"), engine) {
			emit("engineMatches", "Due since Node.js version matched: node"+engine.condition+engine.version)
		}
	}
	for _, unknown := range groups["unknowns"] {
		original := unknown.name
		if index := strings.IndexByte(original, '>'); index >= 0 && !strings.Contains(original, "@") {
			fixed := original[:index] + "@" + original[index:]
			if parseArgument(fixed, hasPackage).kind != "unknowns" {
				used = true
				emit("missingAtSymbol", fmt.Sprintf("Missing '@' on TODO argument. On '%s' use '%s'", original, fixed))
				continue
			}
		}
		fixed := strings.ReplaceAll(original, " ", "")
		if parseArgument(fixed, hasPackage).kind != "unknowns" {
			used = true
			emit("removeWhitespaces", fmt.Sprintf("Avoid using whitespace on TODO argument. On '%s' use '%s'", original, fixed))
		}
	}
	return used
}

func isCatalog(value any) bool { text, _ := value.(string); return strings.HasPrefix(text, "catalog:") }

func truthy(value any) bool {
	switch value := value.(type) {
	case nil:
		return false
	case bool:
		return value
	case string:
		return value != ""
	case float64:
		return value != 0
	default:
		return true
	}
}

func versionMatches(raw any, arg argument) bool {
	if !truthy(raw) {
		return false
	}
	text := ecmascript.JSONValueToString(raw)
	for _, prefix := range []string{">=", "<=", ">", "<", "~", "^"} {
		if strings.HasPrefix(text, prefix) {
			text = strings.TrimPrefix(text, prefix)
			break
		}
	}
	text, _, _ = strings.Cut(text, " ")
	version, ok := npmsemver.CoerceVersion(text)
	return ok && npmsemver.SatisfiesComparison(version, arg.condition+arg.version)
}

func reachedDate(past, now string) bool {
	parse := func(text string) (time.Time, bool) {
		if !datePattern.Test(text) {
			return time.Time{}, false
		}
		year, _ := strconv.Atoi(text[:4])
		month, _ := strconv.Atoi(text[5:7])
		day, _ := strconv.Atoi(text[8:])
		if month < 1 || month > 12 || day < 1 || day > 31 {
			return time.Time{}, false
		}
		// Date.parse normalizes e.g. February 30, but rejects day 32.
		return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), true
	}
	before, validBefore := parse(past)
	after, validAfter := parse(now)
	return validBefore && validAfter && before.Before(after)
}
