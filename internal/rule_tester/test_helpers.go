package rule_tester

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestHelpers provides utility functions for rule testing
type TestHelpers struct {
	RootDir string
}

// NewTestHelpers creates a new TestHelpers instance
func NewTestHelpers(rootDir string) *TestHelpers {
	return &TestHelpers{RootDir: rootDir}
}

// CommonFixtures provides common test code patterns
type CommonFixtures struct{}

// NewCommonFixtures creates a new CommonFixtures instance
func NewCommonFixtures() *CommonFixtures {
	return &CommonFixtures{}
}

// Class generates a class declaration
func (f *CommonFixtures) Class(name string, body string) string {
	return fmt.Sprintf("class %s {\n%s\n}", name, body)
}

// Interface generates an interface declaration
func (f *CommonFixtures) Interface(name string, body string) string {
	return fmt.Sprintf("interface %s {\n%s\n}", name, body)
}

// Function generates a function declaration
func (f *CommonFixtures) Function(name string, params string, returnType string, body string) string {
	if returnType != "" {
		return fmt.Sprintf("function %s(%s): %s {\n%s\n}", name, params, returnType, body)
	}
	return fmt.Sprintf("function %s(%s) {\n%s\n}", name, params, body)
}

// ArrowFunction generates an arrow function
func (f *CommonFixtures) ArrowFunction(params string, returnType string, body string) string {
	if returnType != "" {
		return fmt.Sprintf("(%s): %s => {\n%s\n}", params, returnType, body)
	}
	return fmt.Sprintf("(%s) => {\n%s\n}", params, body)
}

// Method generates a method declaration
func (f *CommonFixtures) Method(name string, params string, returnType string, body string) string {
	if returnType != "" {
		return fmt.Sprintf("  %s(%s): %s {\n  %s\n  }", name, params, returnType, strings.ReplaceAll(body, "\n", "\n  "))
	}
	return fmt.Sprintf("  %s(%s) {\n  %s\n  }", name, params, strings.ReplaceAll(body, "\n", "\n  "))
}

// Property generates a property declaration
func (f *CommonFixtures) Property(name string, type_ string, value string) string {
	if value != "" {
		return fmt.Sprintf("  %s: %s = %s;", name, type_, value)
	}
	return fmt.Sprintf("  %s: %s;", name, type_)
}

// Type generates a type alias
func (f *CommonFixtures) Type(name string, definition string) string {
	return fmt.Sprintf("type %s = %s;", name, definition)
}

// Enum generates an enum declaration
func (f *CommonFixtures) Enum(name string, members string) string {
	return fmt.Sprintf("enum %s {\n%s\n}", name, members)
}

// Namespace generates a namespace declaration
func (f *CommonFixtures) Namespace(name string, body string) string {
	return fmt.Sprintf("namespace %s {\n%s\n}", name, body)
}

// Module generates a module declaration
func (f *CommonFixtures) Module(name string, body string) string {
	return fmt.Sprintf("module %s {\n%s\n}", name, body)
}

// Import generates an import statement
func (f *CommonFixtures) Import(specifiers string, from string) string {
	return fmt.Sprintf("import %s from '%s';", specifiers, from)
}

// Export generates an export statement
func (f *CommonFixtures) Export(what string) string {
	return "export " + what
}

// Const generates a const declaration
func (f *CommonFixtures) Const(name string, type_ string, value string) string {
	if type_ != "" {
		return fmt.Sprintf("const %s: %s = %s;", name, type_, value)
	}
	return fmt.Sprintf("const %s = %s;", name, value)
}

// Let generates a let declaration
func (f *CommonFixtures) Let(name string, type_ string, value string) string {
	if type_ != "" {
		return fmt.Sprintf("let %s: %s = %s;", name, type_, value)
	}
	return fmt.Sprintf("let %s = %s;", name, value)
}

// Var generates a var declaration
func (f *CommonFixtures) Var(name string, type_ string, value string) string {
	if type_ != "" {
		return fmt.Sprintf("var %s: %s = %s;", name, type_, value)
	}
	return fmt.Sprintf("var %s = %s;", name, value)
}

// ProgramHelper provides utilities for creating TypeScript programs for testing
type ProgramHelper struct {
	RootDir string
}

// NewProgramHelper creates a new ProgramHelper
func NewProgramHelper(rootDir string) *ProgramHelper {
	return &ProgramHelper{RootDir: rootDir}
}

// CreateTestProgram creates a TypeScript program from source code
func (h *ProgramHelper) CreateTestProgram(code string, fileName string, tsconfigPath string) (*compiler.Program, *ast.SourceFile, error) {
	fs := utils.NewOverlayVFSForFile(tspath.ResolvePath(h.RootDir, fileName), code)
	host := utils.CreateCompilerHost(h.RootDir, fs)

	program, err := utils.CreateProgram(true, fs, h.RootDir, tsconfigPath, host)
	if err != nil {
		return nil, nil, err
	}

	sourceFile := program.GetSourceFile(fileName)
	return program, sourceFile, nil
}

// DiagnosticAssertion provides utilities for asserting diagnostic properties
type DiagnosticAssertion struct{}

// NewDiagnosticAssertion creates a new DiagnosticAssertion
func NewDiagnosticAssertion() *DiagnosticAssertion {
	return &DiagnosticAssertion{}
}

// FormatDiagnosticError formats a diagnostic error for testing output
func (a *DiagnosticAssertion) FormatDiagnosticError(messageId string, line int, column int, endLine int, endColumn int) InvalidTestCaseError {
	return InvalidTestCaseError{
		MessageId: messageId,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}

// FormatDiagnosticErrorWithSuggestions formats a diagnostic error with suggestions
func (a *DiagnosticAssertion) FormatDiagnosticErrorWithSuggestions(messageId string, line int, column int, suggestions []InvalidTestCaseSuggestion) InvalidTestCaseError {
	return InvalidTestCaseError{
		MessageId:   messageId,
		Line:        line,
		Column:      column,
		Suggestions: suggestions,
	}
}

// BatchTestBuilder helps build large test suites programmatically
type BatchTestBuilder struct {
	Valid   []ValidTestCase
	Invalid []InvalidTestCase
}

// NewBatchTestBuilder creates a new BatchTestBuilder
func NewBatchTestBuilder() *BatchTestBuilder {
	return &BatchTestBuilder{
		Valid:   make([]ValidTestCase, 0),
		Invalid: make([]InvalidTestCase, 0),
	}
}

// AddValid adds a valid test case
func (b *BatchTestBuilder) AddValid(code string) *BatchTestBuilder {
	b.Valid = append(b.Valid, ValidTestCase{Code: code})
	return b
}

// AddValidWithOptions adds a valid test case with options
func (b *BatchTestBuilder) AddValidWithOptions(code string, options any) *BatchTestBuilder {
	b.Valid = append(b.Valid, ValidTestCase{Code: code, Options: options})
	return b
}

// AddValidWithFileName adds a valid test case with a specific filename
func (b *BatchTestBuilder) AddValidWithFileName(code string, fileName string) *BatchTestBuilder {
	b.Valid = append(b.Valid, ValidTestCase{Code: code, FileName: fileName})
	return b
}

// AddInvalid adds an invalid test case
func (b *BatchTestBuilder) AddInvalid(code string, messageId string, line int, column int, output string) *BatchTestBuilder {
	var outputs []string
	if output != "" {
		outputs = []string{output}
	}
	b.Invalid = append(b.Invalid, InvalidTestCase{
		Code: code,
		Errors: []InvalidTestCaseError{
			{MessageId: messageId, Line: line, Column: column},
		},
		Output: outputs,
	})
	return b
}

// AddInvalidWithErrors adds an invalid test case with multiple errors
func (b *BatchTestBuilder) AddInvalidWithErrors(code string, errors []InvalidTestCaseError, output string) *BatchTestBuilder {
	var outputs []string
	if output != "" {
		outputs = []string{output}
	}
	b.Invalid = append(b.Invalid, InvalidTestCase{
		Code:   code,
		Errors: errors,
		Output: outputs,
	})
	return b
}

// AddInvalidWithOptions adds an invalid test case with options
func (b *BatchTestBuilder) AddInvalidWithOptions(code string, messageId string, line int, column int, output string, options any) *BatchTestBuilder {
	var outputs []string
	if output != "" {
		outputs = []string{output}
	}
	b.Invalid = append(b.Invalid, InvalidTestCase{
		Code: code,
		Errors: []InvalidTestCaseError{
			{MessageId: messageId, Line: line, Column: column},
		},
		Output:  outputs,
		Options: options,
	})
	return b
}

// Build returns the built test cases
func (b *BatchTestBuilder) Build() ([]ValidTestCase, []InvalidTestCase) {
	return b.Valid, b.Invalid
}

// GetSuite returns a TestSuite
func (b *BatchTestBuilder) GetSuite() *TestSuite {
	return &TestSuite{
		Valid:   b.Valid,
		Invalid: b.Invalid,
	}
}
