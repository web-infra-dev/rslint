package linter

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// sampleTypeScriptCode is a realistic TypeScript file with various constructs
// that exercise different lint rules.
const sampleTypeScriptCode = `
interface User {
  name: string;
  age: number;
  email?: string;
}

type Status = 'active' | 'inactive' | 'pending';

function processUser(user: User): string {
  var result = '';
  var temp = user.name;

  if (user.age == 18) {
    console.log('Adult user:', user.name);
    result = temp + ' is an adult';
  }

  for (var i = 0; i < 10; i++) {
    result += String(i);
  }

  const items: string[] = [];
  for (var item of ['a', 'b', 'c']) {
    items.push(item);
  }

  if (user.email != null) {
    console.warn('Email exists');
  }

  switch (user.age) {
    case 18:
      result = 'young';
      break;
    case 30:
      result = 'mid';
      break;
    default:
      result = 'other';
  }

  return result;
}

class UserService {
  private users: User[] = [];

  addUser(user: User): void {
    var existing = this.users.find(u => u.name === user.name);
    if (existing == undefined) {
      this.users.push(user);
    }
  }

  getUsers(): User[] {
    return this.users;
  }

  processAll(): string[] {
    var results: string[] = [];
    for (var u of this.users) {
      results.push(processUser(u));
    }
    return results;
  }
}

export { UserService, processUser };
export type { User, Status };
`

// benchNoVarRule returns a configured rule that reports all var declarations.
func benchNoVarRule() ConfiguredRule {
	return ConfiguredRule{
		Name:     "no-var",
		Severity: rule.SeverityWarning,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindVariableDeclarationList: func(node *ast.Node) {
					if node.Flags&ast.NodeFlagsBlockScoped != 0 {
						return
					}
					ctx.ReportNode(node.Parent, rule.RuleMessage{
						Id:          "unexpectedVar",
						Description: "Unexpected var, use let or const instead.",
					})
				},
			}
		},
	}
}

// benchEqeqeqRule returns a configured rule that reports loose equality comparisons.
func benchEqeqeqRule() ConfiguredRule {
	return ConfiguredRule{
		Name:     "eqeqeq",
		Severity: rule.SeverityWarning,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindBinaryExpression: func(node *ast.Node) {
					bin := node.AsBinaryExpression()
					if bin == nil {
						return
					}
					op := bin.OperatorToken.Kind
					if op == ast.KindEqualsEqualsToken || op == ast.KindExclamationEqualsToken {
						ctx.ReportNode(node, rule.RuleMessage{
							Id:          "unexpected",
							Description: "Expected '===' and instead saw '=='.",
						})
					}
				},
			}
		},
	}
}

// benchNoConsoleRule returns a configured rule that reports console usage.
func benchNoConsoleRule() ConfiguredRule {
	return ConfiguredRule{
		Name:     "no-console",
		Severity: rule.SeverityWarning,
		Run: func(ctx rule.RuleContext) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindPropertyAccessExpression: func(node *ast.Node) {
					propAccess := node.AsPropertyAccessExpression()
					if propAccess == nil {
						return
					}
					if propAccess.Expression.Kind != ast.KindIdentifier {
						return
					}
					if propAccess.Expression.AsIdentifier().Text != "console" {
						return
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Unexpected console statement.",
					})
				},
			}
		},
	}
}

// setupBenchProgram creates a compiler.Program with the given source files for benchmarking.
// This follows the same approach as createTestProgramWithFiles in linter_test.go.
func setupBenchProgram(b *testing.B, sourceFiles map[string]string) (*compiler.Program, map[string]string) {
	b.Helper()

	tmpDir := b.TempDir()

	includes := make([]string, 0, len(sourceFiles))
	normalizedPaths := make(map[string]string, len(sourceFiles))
	for name, content := range sourceFiles {
		filePath := filepath.Join(tmpDir, name)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to write %s: %v", name, err)
		}
		includes = append(includes, "./"+name)
		normalizedPaths[name] = tspath.NormalizePath(filePath)
	}

	includeJSON := `"` + joinStrings(includes, `","`) + `"`
	tsconfig := `{"compilerOptions":{"strict":true,"target":"esnext","module":"commonjs"},"include":[` + includeJSON + `]}`
	if err := os.WriteFile(filepath.Join(tmpDir, "tsconfig.json"), []byte(tsconfig), 0644); err != nil {
		b.Fatalf("Failed to write tsconfig: %v", err)
	}

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	host := utils.CreateCompilerHost(tmpDir, fs)
	program, err := utils.CreateProgram(true, fs, tmpDir, "tsconfig.json", host)
	if err != nil {
		b.Fatalf("Failed to create program: %v", err)
	}

	return program, normalizedPaths
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

// runLintBenchmark is the common linting benchmark harness.
func runLintBenchmark(b *testing.B, code string, rules []ConfiguredRule) {
	b.Helper()

	program, normalizedPaths := setupBenchProgram(b, map[string]string{"file.ts": code})
	filePath := normalizedPaths["file.ts"]
	allowedFiles := []string{filePath}

	// Verify the program was created properly
	sourceFile := program.GetSourceFile(filePath)
	if sourceFile == nil {
		b.Fatalf("Source file not found at path: %s", filePath)
	}

	b.ResetTimer()
	for b.Loop() {
		var mu sync.Mutex
		diagnostics := make([]rule.RuleDiagnostic, 0, 16)

		RunLinterInProgram(
			program,
			allowedFiles,
			nil,
			[]string{},
			func(sourceFile *ast.SourceFile) []ConfiguredRule {
				return rules
			},
			false,
			func(diagnostic rule.RuleDiagnostic) {
				mu.Lock()
				diagnostics = append(diagnostics, diagnostic)
				mu.Unlock()
			},
			nil,
			nil,
		)
	}
}

func BenchmarkLintSingleRule_NoVar(b *testing.B) {
	runLintBenchmark(b, sampleTypeScriptCode, []ConfiguredRule{benchNoVarRule()})
}

func BenchmarkLintSingleRule_Eqeqeq(b *testing.B) {
	runLintBenchmark(b, sampleTypeScriptCode, []ConfiguredRule{benchEqeqeqRule()})
}

func BenchmarkLintSingleRule_NoConsole(b *testing.B) {
	runLintBenchmark(b, sampleTypeScriptCode, []ConfiguredRule{benchNoConsoleRule()})
}

func BenchmarkLintMultipleRules(b *testing.B) {
	rules := []ConfiguredRule{
		benchNoVarRule(),
		benchEqeqeqRule(),
		benchNoConsoleRule(),
	}
	runLintBenchmark(b, sampleTypeScriptCode, rules)
}

// benchLargeTypeScriptCode generates a larger TypeScript file to benchmark scaling behavior.
func benchLargeTypeScriptCode() string {
	base := "function process%d(input: string): string {\n" +
		"  var result = input;\n" +
		"  if (result == '') {\n" +
		"    console.log('empty');\n" +
		"    result = 'default';\n" +
		"  }\n" +
		"  for (var i = 0; i < 10; i++) {\n" +
		"    result += String(i);\n" +
		"  }\n" +
		"  return result;\n" +
		"}\n\n"

	var code string
	for i := 0; i < 50; i++ {
		code += benchReplacePercD(base, i)
	}
	code += "\nexport {};\n"
	return code
}

func benchItoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	n := i
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func benchReplacePercD(s string, i int) string {
	result := ""
	for j := 0; j < len(s); j++ {
		if j+1 < len(s) && s[j] == '%' && s[j+1] == 'd' {
			result += benchItoa(i)
			j++
		} else {
			result += string(s[j])
		}
	}
	return result
}

func BenchmarkLintLargeFile(b *testing.B) {
	code := benchLargeTypeScriptCode()
	rules := []ConfiguredRule{
		benchNoVarRule(),
		benchEqeqeqRule(),
		benchNoConsoleRule(),
	}
	runLintBenchmark(b, code, rules)
}
