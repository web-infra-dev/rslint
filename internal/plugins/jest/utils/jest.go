package utils

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type JestFnType string

type JestImportMode string

const (
	JEST_GLOBAL_MODE JestImportMode = "global"
	JEST_IMPORT_MODE JestImportMode = "import"
)

const (
	JestFnTypeExpect   JestFnType = "expect"
	JestFnTypeDescribe JestFnType = "describe"
	JestFnTypeHook     JestFnType = "hook"
	JestFnTypeJest     JestFnType = "jest"
	JestFnTypeTest     JestFnType = "test"
	JestFnTypeUnknown  JestFnType = "unknown"
)

var JEST_METHOD_NAMES = map[string]bool{
	"afterAll":   true,
	"afterEach":  true,
	"beforeAll":  true,
	"beforeEach": true,
	"describe":   true,
	"expect":     true,
	"fdescribe":  true,
	"fit":        true,
	"it":         true,
	"jest":       true,
	"test":       true,
	"xdescribe":  true,
	"xit":        true,
	"xtest":      true,
}

var EQUALITY_METHOD_NAMES = map[string]bool{
	"toBe":          true,
	"toEqual":       true,
	"toStrictEqual": true,
}

var EXPECT_MODIFIER_NAMES = map[string]bool{
	"not":      true,
	"rejects":  true,
	"resolves": true,
}

var VALID_JEST_FN_CALL_CHAINS = map[string]bool{
	"afterAll":                  true,
	"afterEach":                 true,
	"beforeAll":                 true,
	"beforeEach":                true,
	"describe":                  true,
	"describe.each":             true,
	"describe.only":             true,
	"describe.only.each":        true,
	"describe.skip":             true,
	"describe.skip.each":        true,
	"fdescribe":                 true,
	"fdescribe.each":            true,
	"fit":                       true,
	"fit.each":                  true,
	"fit.failing":               true,
	"fit.fails":                 true,
	"it":                        true,
	"it.concurrent":             true,
	"it.concurrent.each":        true,
	"it.concurrent.only.each":   true,
	"it.concurrent.skip.each":   true,
	"it.each":                   true,
	"it.failing":                true,
	"it.fails":                  true,
	"it.only":                   true,
	"it.only.each":              true,
	"it.only.failing":           true,
	"it.only.fails":             true,
	"it.skip":                   true,
	"it.skip.each":              true,
	"it.skip.failing":           true,
	"it.skip.fails":             true,
	"it.todo":                   true,
	"test":                      true,
	"test.concurrent":           true,
	"test.concurrent.each":      true,
	"test.concurrent.only.each": true,
	"test.concurrent.skip.each": true,
	"test.each":                 true,
	"test.failing":              true,
	"test.fails":                true,
	"test.only":                 true,
	"test.only.each":            true,
	"test.only.failing":         true,
	"test.only.fails":           true,
	"test.skip":                 true,
	"test.skip.each":            true,
	"test.skip.failing":         true,
	"test.skip.fails":           true,
	"test.todo":                 true,
	"xdescribe":                 true,
	"xdescribe.each":            true,
	"xit":                       true,
	"xit.each":                  true,
	"xit.failing":               true,
	"xit.fails":                 true,
	"xtest":                     true,
	"xtest.each":                true,
	"xtest.failing":             true,
	"xtest.fails":               true,
}

type ParsedJestFnMemberEntry struct {
	Name string
	Node *ast.Node
}

func getPropertyName(node *ast.Node) string {
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		return node.AsPrivateIdentifier().Text
	}
	return ""
}

func GetJestKind(name string) JestFnType {
	switch name {
	case "describe", "fdescribe", "xdescribe":
		return JestFnTypeDescribe
	case "fit", "it", "test", "xit", "xtest":
		return JestFnTypeTest
	case "beforeEach", "afterEach", "beforeAll", "afterAll":
		return JestFnTypeHook
	case "jest":
		return JestFnTypeJest
	case "expect":
		return JestFnTypeExpect
	default:
		return JestFnTypeUnknown
	}
}

func GetJestFnMemberEntries(node *ast.Node) []ParsedJestFnMemberEntry {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return []ParsedJestFnMemberEntry{{
			Name: node.AsIdentifier().Text,
			Node: node,
		}}
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		left := GetJestFnMemberEntries(property.Expression)
		nameNode := property.Name()
		if name := getPropertyName(nameNode); name != "" {
			return append(left, ParsedJestFnMemberEntry{
				Name: name,
				Node: nameNode,
			})
		}
		return left
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		left := GetJestFnMemberEntries(element.Expression)
		nameNode := ast.SkipParentheses(element.ArgumentExpression)
		if name := getElementAccessName(nameNode); name != "" {
			return append(left, ParsedJestFnMemberEntry{
				Name: name,
				Node: nameNode,
			})
		}
		return left
	case ast.KindCallExpression:
		return GetJestFnMemberEntries(node.AsCallExpression().Expression)
	case ast.KindTaggedTemplateExpression:
		return GetJestFnMemberEntries(node.AsTaggedTemplateExpression().Tag)
	default:
		return nil
	}
}

func getElementAccessName(node *ast.Node) string {
	if node == nil {
		return ""
	}

	node = ast.SkipParentheses(node)
	if node == nil {
		return ""
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	default:
		return ""
	}
}

// DefaultJestVersion is used when the Jest version cannot be resolved from settings or package.json.
const DefaultJestVersion = "29.0.0"

// JestVersionMajor extracts the major version number from an npm version or range (e.g. "^29.0.0", "~27.1.0").
// It returns 29 when the version cannot be parsed, matching the historical default of Jest 29.
func JestVersionMajor(v string) int {
	const fallback = 29
	s := strings.TrimSpace(v)
	if s == "" {
		return fallback
	}
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "workspace:") || strings.HasPrefix(low, "file:") || strings.HasPrefix(low, "link:") {
		return fallback
	}
	if low == "latest" || low == "*" {
		return fallback
	}
	if strings.HasPrefix(low, "npm:") {
		if at := strings.LastIndexByte(s, '@'); at >= 0 && at+1 < len(s) {
			s = strings.TrimSpace(s[at+1:])
		} else {
			return fallback
		}
	}
	// Remove leading range operators
	for {
		changed := false
		if strings.HasPrefix(s, ">=") {
			s = strings.TrimSpace(s[2:])
			changed = true
		} else if strings.HasPrefix(s, "<=") {
			s = strings.TrimSpace(s[2:])
			changed = true
		} else if strings.HasPrefix(s, ">") || strings.HasPrefix(s, "<") {
			s = strings.TrimSpace(s[1:])
			changed = true
		} else if strings.HasPrefix(s, "^") || strings.HasPrefix(s, "~") {
			s = strings.TrimSpace(s[1:])
			changed = true
		}
		if !changed {
			break
		}
	}
	s = strings.TrimLeft(s, "vV")
	if s == "" {
		return fallback
	}
	major, _, _ := strings.Cut(s, ".")
	firstNonDigit := strings.IndexFunc(major, func(r rune) bool {
		return r < '0' || r > '9'
	})
	if firstNonDigit >= 0 {
		major = major[:firstNonDigit]
	}
	n, err := strconv.Atoi(major)
	if err != nil {
		return fallback
	}
	return n
}

// jestVersionFromSettings returns the Jest version from rslint settings (ESLint style settings.jest.version).
func jestVersionFromSettings(settings map[string]interface{}) (string, bool) {
	if settings == nil {
		return "", false
	}
	raw, ok := settings["jest"]
	if !ok {
		return "", false
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return "", false
	}
	var ver string
	switch v := m["version"].(type) {
	case string:
		ver = v
	case float64:
		// JSON numbers decode into interface{} as float64; treat them as a major version.
		ver = strconv.Itoa(int(v))
	case int:
		ver = strconv.Itoa(v)
	case int64:
		ver = strconv.FormatInt(v, 10)
	default:
		return "", false
	}
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return "", false
	}
	return ver, true
}

// jestVersionFromPackageJSONText reads the "jest" dependency from JSON text.
func jestVersionFromPackageJSONText(data string) string {
	var m struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		PeerDeps        map[string]string `json:"peerDependencies"`
		OptDependencies map[string]string `json:"optionalDependencies"`
	}
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return ""
	}
	// Aligned with packagejson.HasDependency: dependencies, then dev, then peer, then optional
	if m.Dependencies != nil {
		if v, ok := m.Dependencies["jest"]; ok {
			return v
		}
	}
	if m.DevDependencies != nil {
		if v, ok := m.DevDependencies["jest"]; ok {
			return v
		}
	}
	if m.PeerDeps != nil {
		if v, ok := m.PeerDeps["jest"]; ok {
			return v
		}
	}
	if m.OptDependencies != nil {
		if v, ok := m.OptDependencies["jest"]; ok {
			return v
		}
	}
	return ""
}

// readJestVersionFromPackageJson resolves the jest version from the nearest package.json (same package
// as the current source file) using the TypeScript program's host filesystem.
func readJestVersionFromPackageJson(program *compiler.Program, sourceFile *ast.SourceFile) string {
	if program == nil || sourceFile == nil {
		return ""
	}
	dir := tspath.GetDirectoryPath(sourceFile.FileName())
	pkgDir := program.GetNearestAncestorDirectoryWithPackageJson(dir)
	if pkgDir == "" {
		return ""
	}
	pkgPath := tspath.CombinePaths(pkgDir, "package.json")
	if !program.FileExists(pkgPath) {
		return ""
	}
	text, ok := program.Host().FS().ReadFile(pkgPath)
	if !ok {
		return ""
	}
	return jestVersionFromPackageJSONText(text)
}

// GetJestVersion returns the effective Jest version: explicit settings, then the nearest package.json,
// then DefaultJestVersion.
func GetJestVersion(ctx rule.RuleContext) string {
	if s, ok := jestVersionFromSettings(ctx.Settings); ok {
		return s
	}
	if v := readJestVersionFromPackageJson(ctx.Program, ctx.SourceFile); v != "" {
		return v
	}

	return DefaultJestVersion
}
