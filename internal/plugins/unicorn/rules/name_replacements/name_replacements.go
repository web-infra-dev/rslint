package name_replacements

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	scopeAnalysis "github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

//go:embed name_replacements.schema.json
var schemaJSON []byte

const (
	messageIDReplace    = "replace"
	messageIDSuggestion = "suggestion"
	messageIDRename     = "rename"
	moreDescriptive     = "A more descriptive name will do too."
)

type importCheckMode uint8

const (
	importCheckNever importCheckMode = iota
	importCheckInternal
	importCheckAlways
)

type options struct {
	checkProperties                 bool
	checkVariables                  bool
	checkDefaultAndNamespaceImports importCheckMode
	checkShorthandImports           importCheckMode
	checkShorthandProperties        bool
	checkFilenames                  bool
	replacements                    map[string]map[string]bool
	allowList                       map[string]bool
	ignore                          []*esregexp.RegExp
}

type replacementResult struct {
	total   int
	samples []string
}

type nameReplacements struct {
	ctx                    rule.RuleContext
	opts                   options
	manager                *scope.Manager
	generatedNamesByScope  map[*scope.Scope]map[string]struct{}
	reportedPropertyRanges map[core.TextRange]struct{}
	referencesByVariable   map[*scope.Variable][]*scope.Reference
	classVariablesByID     map[*ast.Node][]*scope.Variable
	unresolvedScopesByName map[string][]*scope.Scope
	localeComparer         *ecmascript.LocaleComparer
}

type preparedOptionsCacheKey struct{ encoded string }

// NameReplacementsRule enforces descriptive replacements for variable,
// property, and filename names.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/docs/rules/name-replacements.md
var NameReplacementsRule = rule.Rule{
	Name:   "unicorn/name-replacements",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		var opts options
		if encoded, err := json.Marshal(rawOptions); err == nil {
			opts = rule.CachedByProgram(ctx, preparedOptionsCacheKey{encoded: string(encoded)}, func() options {
				return parseOptions(rawOptions)
			})
		} else {
			opts = parseOptions(rawOptions)
		}
		state := &nameReplacements{
			ctx:                    ctx,
			opts:                   opts,
			generatedNamesByScope:  make(map[*scope.Scope]map[string]struct{}),
			reportedPropertyRanges: make(map[core.TextRange]struct{}),
		}

		state.checkFilename()
		if opts.checkVariables {
			state.manager = scopeAnalysis.Get(ctx, scope.Options{CollectReferences: true})
		}

		listeners := rule.RuleListeners{}
		if opts.checkProperties {
			listeners[ast.KindIdentifier] = state.checkProperty
		}

		if state.manager != nil {
			statements := ctx.SourceFile.Statements.Nodes
			if len(statements) == 0 {
				state.checkVariables()
			} else {
				last := statements[len(statements)-1]
				listeners[rule.ListenerOnExit(last.Kind)] = func(node *ast.Node) {
					if node == last {
						state.checkVariables()
					}
				}
			}
		}

		return listeners
	},
}

func parseImportMode(value any, defaultMode importCheckMode) importCheckMode {
	switch value := value.(type) {
	case bool:
		if value {
			return importCheckAlways
		}
		return importCheckNever
	case string:
		if value == "internal" {
			return importCheckInternal
		}
	}
	return defaultMode
}

func cloneBoolMap(source map[string]bool) map[string]bool {
	result := make(map[string]bool, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func parseOptions(raw []any) options {
	result := options{
		checkVariables:                  true,
		checkDefaultAndNamespaceImports: importCheckInternal,
		checkShorthandImports:           importCheckInternal,
		checkFilenames:                  true,
	}
	values := map[string]any{}
	if len(raw) > 0 {
		values, _ = raw[0].(map[string]any)
	}

	if value, ok := values["checkProperties"].(bool); ok {
		result.checkProperties = value
	}
	if value, ok := values["checkVariables"].(bool); ok {
		result.checkVariables = value
	}
	result.checkDefaultAndNamespaceImports = parseImportMode(
		values["checkDefaultAndNamespaceImports"], importCheckInternal,
	)
	result.checkShorthandImports = parseImportMode(values["checkShorthandImports"], importCheckInternal)
	if value, ok := values["checkShorthandProperties"].(bool); ok {
		result.checkShorthandProperties = value
	}
	if value, ok := values["checkFilenames"].(bool); ok {
		result.checkFilenames = value
	}

	extendDefaults := true
	if value, ok := values["extendDefaultReplacements"].(bool); ok {
		extendDefaults = value
	}
	if extendDefaults {
		result.replacements = newDefaultReplacements()
	} else {
		result.replacements = make(map[string]map[string]bool)
	}
	if configured, ok := values["replacements"].(map[string]any); ok {
		for discouraged, rawReplacement := range configured {
			if disabled, ok := rawReplacement.(bool); ok && !disabled {
				result.replacements[discouraged] = map[string]bool{}
				continue
			}
			merged := map[string]bool{}
			if existing := result.replacements[discouraged]; existing != nil {
				merged = cloneBoolMap(existing)
			}
			if replacements, ok := rawReplacement.(map[string]any); ok {
				for replacement, rawEnabled := range replacements {
					if enabled, ok := rawEnabled.(bool); ok {
						merged[replacement] = enabled
					}
				}
			}
			result.replacements[discouraged] = merged
		}
	}

	extendAllowList := true
	if value, ok := values["extendDefaultAllowList"].(bool); ok {
		extendAllowList = value
	}
	if extendAllowList {
		result.allowList = cloneBoolMap(defaultAllowList)
	} else {
		result.allowList = make(map[string]bool)
	}
	if configured, ok := values["allowList"].(map[string]any); ok {
		for name, rawAllowed := range configured {
			if allowed, ok := rawAllowed.(bool); ok {
				result.allowList[name] = allowed
			}
		}
	}

	patterns := append([]string(nil), defaultIgnore...)
	if configured, ok := values["ignore"].([]any); ok {
		for _, value := range configured {
			if pattern, ok := value.(string); ok {
				patterns = append(patterns, pattern)
			}
		}
	}
	for _, pattern := range patterns {
		if compiled, err := esregexp.Compile(pattern, "u"); err == nil {
			result.ignore = append(result.ignore, compiled)
		}
	}
	return result
}

func jsUpperFirst(value string) string {
	if value == "" {
		return value
	}
	first, size := utf8.DecodeRuneInString(value)
	if first > 0xffff {
		return value
	}
	return ecmascript.StringToUpperCase(string(first)) + value[size:]
}

func jsLowerFirst(value string) string {
	if value == "" {
		return value
	}
	first, size := utf8.DecodeRuneInString(value)
	if first > 0xffff {
		return value
	}
	return ecmascript.StringToLowerCase(string(first)) + value[size:]
}

func isUpperCase(value string) bool {
	return value == ecmascript.StringToUpperCase(value)
}

func isUpperFirst(value string) bool {
	if value == "" {
		return false
	}
	firstByte := value[0]
	if firstByte < utf8.RuneSelf {
		return firstByte < 'a' || firstByte > 'z'
	}
	first, size := ecmascript.DecodeStringRune(value)
	if first > 0xffff {
		return true
	}
	character := value[:size]
	return character == ecmascript.StringToUpperCase(character)
}

func splitNameWords(value string) []string {
	if value == "" {
		return nil
	}
	starts := []int{0}
	var previous rune
	first := true
	for index, current := range value {
		if !first && (!unicode.IsLower(current) || !unicode.IsLetter(previous)) {
			starts = append(starts, index)
		}
		previous = current
		first = false
	}
	words := make([]string, 0, len(starts))
	for index, start := range starts {
		end := len(value)
		if index+1 < len(starts) {
			end = starts[index+1]
		}
		if start != end {
			words = append(words, value[start:end])
		}
	}
	return words
}

func (o options) ignored(name string) bool {
	for _, pattern := range o.ignore {
		if pattern.TestOrTimeout(name) {
			return true
		}
	}
	return false
}

func (o options) wordReplacements(word string, compare func(string, string) int) []string {
	if isUpperCase(word) || o.allowList[word] {
		return nil
	}

	var replacements map[string]bool
	var found bool
	for _, candidate := range []string{jsLowerFirst(word), word, jsUpperFirst(word)} {
		if replacements, found = o.replacements[candidate]; found {
			break
		}
	}
	if !found {
		return nil
	}

	transform := jsLowerFirst
	if isUpperFirst(word) {
		transform = jsUpperFirst
	}
	result := make([]string, 0, len(replacements))
	for replacement, enabled := range replacements {
		if enabled {
			result = append(result, transform(replacement))
		}
	}
	slices.SortStableFunc(result, compare)
	return result
}

const maximumRelevantReplacementCount = 103

func (o options) nameReplacements(name string, limit int, compare func(string, string) int) replacementResult {
	if isUpperCase(name) || o.allowList[name] || o.ignored(name) {
		return replacementResult{}
	}
	if exact := o.wordReplacements(name, compare); len(exact) > 0 {
		return replacementResult{total: len(exact), samples: exact[:min(limit, len(exact))]}
	}
	allLowerASCII := name != ""
	for index := range len(name) {
		if name[index] < 'a' || name[index] > 'z' {
			allLowerASCII = false
			break
		}
	}
	if allLowerASCII {
		return replacementResult{}
	}

	words := splitNameWords(name)
	combinations := make([][]string, 0, len(words))
	hasReplacements := false
	total := 1
	for _, word := range words {
		replacements := o.wordReplacements(word, compare)
		if len(replacements) == 0 {
			replacements = []string{word}
		} else {
			hasReplacements = true
		}
		combinations = append(combinations, replacements)
		if total > maximumRelevantReplacementCount/len(replacements) {
			total = maximumRelevantReplacementCount
		} else {
			total *= len(replacements)
		}
	}
	if !hasReplacements {
		return replacementResult{}
	}

	sampleCount := min(total, limit)
	samples := make([]string, 0, sampleCount)
	for sampleIndex := range sampleCount {
		remaining := sampleIndex
		parts := make([]string, len(combinations))
		for combinationIndex := len(combinations) - 1; combinationIndex >= 0; combinationIndex-- {
			items := combinations[combinationIndex]
			itemIndex := remaining % len(items)
			remaining = (remaining - itemIndex) / len(items)
			parts[combinationIndex] = items[itemIndex]
		}
		for index := len(parts) - 1; index > 0; index-- {
			if isASCIIAlpha(parts[index]) && strings.HasSuffix(parts[index-1], parts[index]) {
				parts = slices.Delete(parts, index, index+1)
			}
		}
		samples = append(samples, strings.Join(parts, ""))
	}
	return replacementResult{total: total, samples: samples}
}

func (r *nameReplacements) compareReplacements(left, right string) int {
	if r.localeComparer == nil {
		r.localeComparer = ecmascript.NewLocaleComparer("")
	}
	if result := r.localeComparer.Compare(left, right); result != 0 {
		return result
	}
	return ecmascript.CompareStrings(left, right)
}

func (r *nameReplacements) nameReplacements(name string, limit int) replacementResult {
	return r.opts.nameReplacements(name, limit, r.compareReplacements)
}

func isASCIIAlpha(value string) bool {
	if value == "" {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') {
			return false
		}
	}
	return true
}

func replacementMessage(discouraged string, replacements replacementResult, nameType string) rule.RuleMessage {
	if replacements.total == 1 {
		replacement := replacements.samples[0]
		return rule.RuleMessage{
			Id:          messageIDReplace,
			Description: fmt.Sprintf("The %s `%s` should be named `%s`. %s", nameType, discouraged, replacement, moreDescriptive),
			Data: map[string]string{
				"nameTypeText":    nameType,
				"discouragedName": discouraged,
				"replacement":     replacement,
			},
		}
	}
	quoted := make([]string, len(replacements.samples))
	for index, sample := range replacements.samples {
		quoted[index] = "`" + sample + "`"
	}
	replacementsText := strings.Join(quoted, ", ")
	if omitted := replacements.total - len(replacements.samples); omitted > 0 {
		omittedText := strconv.Itoa(omitted)
		if omitted > 99 {
			omittedText = "99+"
		}
		replacementsText += fmt.Sprintf(", ... (%s more omitted)", omittedText)
	}
	return rule.RuleMessage{
		Id:          messageIDSuggestion,
		Description: fmt.Sprintf("Please rename the %s `%s`. Suggested names are: %s. %s", nameType, discouraged, replacementsText, moreDescriptive),
		Data: map[string]string{
			"nameTypeText":     nameType,
			"discouragedName":  discouraged,
			"replacementsText": replacementsText,
		},
	}
}

func (r *nameReplacements) checkFilename() {
	if !r.opts.checkFilenames || r.ctx.SourceFile == nil {
		return
	}
	filename := path.Base(r.ctx.SourceFile.FileName())
	if filename == "" || filename == "." || filename == "<input>" || filename == "<text>" {
		return
	}
	extension := path.Ext(filename)
	name := strings.TrimSuffix(filename, extension)
	replacements := r.nameReplacements(name, 3)
	if replacements.total == 0 {
		return
	}
	for index := range replacements.samples {
		replacements.samples[index] += extension
	}
	r.ctx.ReportRange(core.NewTextRange(0, len(r.ctx.SourceFile.Text())), replacementMessage(filename, replacements, "filename"))
}

func (r *nameReplacements) checkVariables() {
	type variableTask struct {
		variable     *scope.Variable
		declarations []*scope.Variable
	}
	processed := make(map[*scope.Variable]struct{})
	tasks := make([]variableTask, 0)
	r.buildBindingIndexes()
	for _, currentScope := range r.manager.Scopes {
		for _, variable := range currentScope.Vars {
			if _, seen := processed[variable]; seen || variable.Anonymous || variable.ID == nil {
				continue
			}
			declarations := currentScope.Declarations(variable.Name)
			for _, declaration := range declarations {
				processed[declaration] = struct{}{}
			}
			// The synthetic inner class binding uses the same identifier as the
			// outer binding. It is folded into the outer variable below.
			if variable.Kind == scope.DefClassInnerName && slices.ContainsFunc(r.classVariablesByID[variable.ID], func(candidate *scope.Variable) bool {
				return candidate.Kind == scope.DefClassName
			}) {
				continue
			}
			tasks = append(tasks, variableTask{variable: variable, declarations: declarations})
		}
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].variable.ID.Pos() < tasks[j].variable.ID.Pos()
	})
	for _, task := range tasks {
		r.checkVariable(task.variable, task.declarations)
	}
}

func (r *nameReplacements) buildBindingIndexes() {
	if r.referencesByVariable != nil {
		return
	}
	r.referencesByVariable = make(map[*scope.Variable][]*scope.Reference)
	r.classVariablesByID = make(map[*ast.Node][]*scope.Variable)
	r.unresolvedScopesByName = make(map[string][]*scope.Scope)
	for _, currentScope := range r.manager.Scopes {
		for _, variable := range currentScope.Vars {
			if variable.ID != nil && (variable.Kind == scope.DefClassName || variable.Kind == scope.DefClassInnerName) {
				r.classVariablesByID[variable.ID] = append(r.classVariablesByID[variable.ID], variable)
			}
		}
	}
	for _, reference := range r.manager.References {
		if len(reference.Declarations) == 0 {
			name := reference.Identifier.Text()
			r.unresolvedScopesByName[name] = append(r.unresolvedScopesByName[name], reference.From)
			continue
		}
		key := reference.Declarations[0]
		r.referencesByVariable[key] = append(r.referencesByVariable[key], reference)
	}
}

func (r *nameReplacements) classVariables(variable *scope.Variable, declarations []*scope.Variable) []*scope.Variable {
	result := append([]*scope.Variable(nil), declarations...)
	if variable.Kind == scope.DefClassName {
		for _, candidate := range r.classVariablesByID[variable.ID] {
			if !slices.Contains(result, candidate) {
				result = append(result, candidate)
			}
		}
	}
	return result
}

func containsVariable(variables []*scope.Variable, candidate *scope.Variable) bool {
	return slices.Contains(variables, candidate)
}

func (r *nameReplacements) referencesFor(variables []*scope.Variable) []*scope.Reference {
	keys := make(map[*scope.Variable]struct{}, len(variables))
	result := make([]*scope.Reference, 0)
	for _, variable := range variables {
		key := variable
		if declarations := variable.Scope.Declarations(variable.Name); len(declarations) > 0 {
			key = declarations[0]
		}
		if _, duplicate := keys[key]; duplicate {
			continue
		}
		keys[key] = struct{}{}
		result = append(result, r.referencesByVariable[key]...)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Identifier.Pos() < result[j].Identifier.Pos()
	})
	result = slices.Compact(result)
	return result
}

func (r *nameReplacements) checkVariable(variable *scope.Variable, declarations []*scope.Variable) {
	if variable.Kind == scope.DefEnumMember || !r.shouldCheckImport(variable) ||
		(!r.opts.checkShorthandProperties && isShorthandBinding(variable)) {
		return
	}
	replacements := r.nameReplacements(variable.Name, 3)
	if replacements.total == 0 {
		return
	}

	variables := r.classVariables(variable, declarations)
	references := r.referencesFor(variables)
	scopes := make([]*scope.Scope, 0, len(references)+len(variables))
	scopeSet := make(map[*scope.Scope]struct{}, cap(scopes))
	appendScope := func(current *scope.Scope) {
		if current == nil {
			return
		}
		if _, duplicate := scopeSet[current]; duplicate {
			return
		}
		scopeSet[current] = struct{}{}
		scopes = append(scopes, current)
	}
	for _, reference := range references {
		appendScope(reference.From)
	}
	for _, declaration := range variables {
		appendScope(declaration.Scope)
	}
	filtered := replacements.samples[:0]
	for _, candidate := range replacements.samples {
		if available := r.availableName(candidate, scopes, variables); available != "" {
			filtered = append(filtered, available)
		}
	}
	replacements.samples = filtered
	if len(replacements.samples) == 0 {
		return
	}

	message := replacementMessage(variable.ID.Text(), replacements, "variable")
	reportRange := utils.GetESTreeBindingIdentifierRange(r.ctx.SourceFile, variable.ID)
	canRename := r.canRename(variable, variables, references)
	canFixParameter := r.canFixParameter(variable)
	if replacements.total == 1 && canRename && canFixParameter {
		replacement := replacements.samples[0]
		for _, currentScope := range scopes {
			generated := r.generatedNamesByScope[currentScope]
			if generated == nil {
				generated = make(map[string]struct{})
				r.generatedNamesByScope[currentScope] = generated
			}
			generated[replacement] = struct{}{}
		}
		r.ctx.ReportRangeWithDeferredFixes(reportRange, message, func() []rule.RuleFix {
			return r.renameFixes(variable, variables, references, replacement)
		})
		return
	}
	if replacements.total > 1 && canRename && canFixParameter {
		r.ctx.ReportRangeWithDeferredSuggestions(reportRange, message, func() []rule.RuleSuggestion {
			return r.renameSuggestions(variable, variables, references, replacements.samples)
		})
		return
	}
	r.ctx.ReportRange(reportRange, message)
}

func isShorthandBinding(variable *scope.Variable) bool {
	if variable.ID == nil {
		return false
	}
	for node := variable.ID.Parent; node != nil; node = node.Parent {
		if node.Kind == ast.KindBindingElement {
			binding := node.AsBindingElement()
			return binding.PropertyName == nil && binding.Initializer == nil && node.Parent != nil &&
				node.Parent.Kind == ast.KindObjectBindingPattern
		}
		if node.Kind == ast.KindShorthandPropertyAssignment {
			return true
		}
		if node == variable.DefNode {
			break
		}
	}
	return false
}

func importSource(variable *scope.Variable) string {
	for node := variable.DefNode; node != nil; node = node.Parent {
		if node.Kind == ast.KindImportDeclaration {
			module := node.AsImportDeclaration().ModuleSpecifier
			if module != nil && ast.IsStringLiteralLike(module) {
				return module.Text()
			}
		}
		if node.Kind == ast.KindVariableStatement || node.Kind == ast.KindSourceFile {
			break
		}
	}
	if variable.DefNode != nil && variable.DefNode.Kind == ast.KindVariableDeclaration {
		initializer := variable.DefNode.Initializer()
		if initializer == nil {
			return ""
		}
		initializer = ast.SkipParentheses(initializer)
		if initializer != nil && initializer.Kind == ast.KindCallExpression {
			call := initializer.AsCallExpression()
			if call.Expression.Kind == ast.KindIdentifier && call.Expression.Text() == "require" &&
				len(call.Arguments.Nodes) == 1 && ast.IsStringLiteralLike(call.Arguments.Nodes[0]) {
				return call.Arguments.Nodes[0].Text()
			}
		}
	}
	return ""
}

func isInternalSource(source string) bool {
	return !strings.Contains(source, "node_modules") &&
		(strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/"))
}

func (r *nameReplacements) shouldCheckImport(variable *scope.Variable) bool {
	if variable.Kind != scope.DefImport && importSource(variable) == "" {
		return true
	}
	source := importSource(variable)
	mode := r.opts.checkDefaultAndNamespaceImports
	if variable.Kind == scope.DefImport && variable.ID != nil && variable.ID.Parent != nil &&
		variable.ID.Parent.Kind == ast.KindImportSpecifier {
		specifier := variable.ID.Parent.AsImportSpecifier()
		if specifier.PropertyName != nil && specifier.PropertyName.Text() == "default" {
			mode = r.opts.checkDefaultAndNamespaceImports
		} else if specifier.PropertyName == nil || specifier.PropertyName.Text() == specifier.Name().Text() {
			mode = r.opts.checkShorthandImports
		} else {
			return true
		}
	}
	switch mode {
	case importCheckNever:
		return false
	case importCheckInternal:
		return isInternalSource(source)
	default:
		return true
	}
}

func scopeResolvesName(current *scope.Scope, name string, own []*scope.Variable) bool {
	for candidateScope := current; candidateScope != nil; candidateScope = candidateScope.Parent {
		declarations := candidateScope.Declarations(name)
		for _, declaration := range declarations {
			if !containsVariable(own, declaration) {
				return true
			}
		}
		if len(declarations) > 0 {
			return false
		}
	}
	return false
}

func isDescendantScope(candidate, ancestor *scope.Scope) bool {
	for current := candidate; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

func (r *nameReplacements) isSafeName(name string, scopes []*scope.Scope, own []*scope.Variable) bool {
	if name == "arguments" || r.ctx.Globals.Access(name).IsDeclared() {
		return false
	}
	for _, currentScope := range scopes {
		if scopeResolvesName(currentScope, name, own) {
			return false
		}
		if _, generated := r.generatedNamesByScope[currentScope][name]; generated {
			return false
		}
		for _, referenceScope := range r.getUnresolvedScopesByName()[name] {
			if isDescendantScope(referenceScope, currentScope) {
				return false
			}
		}
	}
	return true
}

func (r *nameReplacements) getUnresolvedScopesByName() map[string][]*scope.Scope {
	r.buildBindingIndexes()
	return r.unresolvedScopesByName
}

var typescriptReservedWords = map[string]struct{}{
	"break": {}, "case": {}, "catch": {}, "class": {}, "const": {}, "continue": {},
	"debugger": {}, "default": {}, "delete": {}, "do": {}, "else": {}, "enum": {},
	"export": {}, "extends": {}, "false": {}, "finally": {}, "for": {}, "function": {},
	"if": {}, "import": {}, "in": {}, "instanceof": {}, "new": {}, "null": {},
	"return": {}, "super": {}, "switch": {}, "this": {}, "throw": {}, "true": {},
	"try": {}, "typeof": {}, "var": {}, "void": {}, "while": {}, "with": {}, "as": {},
	"implements": {}, "interface": {}, "let": {}, "package": {}, "private": {},
	"protected": {}, "public": {}, "static": {}, "yield": {}, "any": {}, "boolean": {},
	"constructor": {}, "declare": {}, "get": {}, "module": {}, "require": {}, "number": {},
	"set": {}, "string": {}, "symbol": {}, "type": {}, "from": {}, "of": {},
}

func validVariableName(name string) bool {
	_, reserved := typescriptReservedWords[name]
	return scanner.IsValidIdentifier(name) && !reserved
}

func (r *nameReplacements) availableName(candidate string, scopes []*scope.Scope, own []*scope.Variable) string {
	if !validVariableName(candidate) {
		candidate += "_"
		if !validVariableName(candidate) {
			return ""
		}
	}
	for !r.isSafeName(candidate, scopes, own) {
		candidate += "_"
	}
	return candidate
}

func isExportedDeclaration(variable *scope.Variable) bool {
	switch variable.Kind {
	case scope.DefParameter, scope.DefTypeParameter, scope.DefClassInnerName:
		return false
	}
	for node := variable.DefNode; node != nil && node.Kind != ast.KindSourceFile; node = node.Parent {
		flags := ast.GetCombinedModifierFlags(node)
		if flags&ast.ModifierFlagsExport != 0 {
			return flags&ast.ModifierFlagsDefault == 0
		}
		if node.Kind == ast.KindBlock || ast.IsFunctionLikeDeclaration(node) {
			break
		}
	}
	return false
}

func isJSXNamePosition(node *ast.Node) bool {
	for current := node; current != nil && current.Parent != nil; current = current.Parent {
		if ast.IsJsxTagName(current) {
			return true
		}
		parent := current.Parent
		if parent.Kind == ast.KindJsxAttribute {
			return parent.Name() == current
		}
		if parent.Kind != ast.KindPropertyAccessExpression && parent.Kind != ast.KindJsxNamespacedName {
			return false
		}
	}
	return false
}

func (r *nameReplacements) canRename(variable *scope.Variable, variables []*scope.Variable, references []*scope.Reference) bool {
	for _, declaration := range variables {
		if isExportedDeclaration(declaration) || isJSXNamePosition(declaration.ID) {
			return false
		}
	}
	for _, reference := range references {
		if isJSXNamePosition(reference.Identifier) {
			return false
		}
	}
	return true
}

func (r *nameReplacements) canFixParameter(variable *scope.Variable) bool {
	if variable.Kind != scope.DefParameter {
		return true
	}
	if variable.DefNode != nil && variable.DefNode.Kind == ast.KindParameter &&
		variable.DefNode.ModifierFlags()&ast.ModifierFlagsParameterPropertyModifier != 0 {
		return false
	}
	function := variable.DefNode
	for function != nil && !isFunctionLikeWithParameters(function) {
		function = function.Parent
	}
	if function == nil {
		return true
	}
	return !r.hasAttachedJSDocParameterComment(function)
}

func (r *nameReplacements) hasAttachedJSDocParameterComment(function *ast.Node) bool {
	candidate := function
	var comment *ast.CommentRange
	for candidate != nil {
		previousComment, previousIsComment := r.previousSourceItem(candidate)
		if previousIsComment {
			comment = previousComment
			break
		}
		if candidate.Parent == nil || !isCommentAttachmentParent(candidate.Parent) {
			return false
		}
		candidate = candidate.Parent
	}
	if comment == nil || comment.Kind != ast.KindMultiLineCommentTrivia {
		return false
	}
	candidateStart := utils.TrimNodeTextRange(r.ctx.SourceFile, candidate).Pos()
	lineMap := r.ctx.SourceFile.ECMALineMap()
	if scanner.ComputeLineOfPosition(lineMap, candidateStart)-scanner.ComputeLineOfPosition(lineMap, comment.End()) > 1 {
		return false
	}
	value := utils.CommentValue(r.ctx.SourceFile.Text(), comment)
	value = strings.TrimLeftFunc(value, ecmascript.IsWhiteSpaceOrLineTerminator)
	return strings.HasPrefix(value, "*") && containsJSDocParameterTag(value)
}

func (r *nameReplacements) previousSourceItem(node *ast.Node) (*ast.CommentRange, bool) {
	start := utils.TrimNodeTextRange(r.ctx.SourceFile, node).Pos()
	comments := r.ctx.Comments.All()
	commentIndex := sort.Search(len(comments), func(index int) bool {
		return comments[index].End() > start
	})
	var comment *ast.CommentRange
	if commentIndex > 0 {
		comment = comments[commentIndex-1]
	}
	token, hasToken := utils.TokenBeforePosition(r.ctx.SourceFile, start)
	if comment != nil && (!hasToken || comment.End() > token.End) {
		return comment, true
	}
	return nil, false
}

func isCommentAttachmentParent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindExportAssignment, ast.KindExpressionStatement,
		ast.KindMethodDeclaration, ast.KindPropertyAssignment,
		ast.KindPropertyDeclaration, ast.KindPropertySignature,
		ast.KindMethodSignature, ast.KindTypeAliasDeclaration,
		ast.KindVariableDeclaration, ast.KindVariableDeclarationList, ast.KindVariableStatement,
		ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	case ast.KindBinaryExpression:
		return ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind)
	default:
		return false
	}
}

func containsJSDocParameterTag(value string) bool {
	for searchStart := 0; searchStart < len(value); {
		offset := strings.Index(value[searchStart:], "@param")
		if offset < 0 {
			return false
		}
		end := searchStart + offset + len("@param")
		if end == len(value) || !isASCIIWordByte(value[end]) {
			return true
		}
		searchStart = end
	}
	return false
}

func isASCIIWordByte(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9' || value == '_'
}

func isFunctionLikeWithParameters(node *ast.Node) bool {
	if ast.IsFunctionLikeDeclaration(node) {
		return true
	}
	switch node.Kind {
	case ast.KindCallSignature, ast.KindConstructSignature, ast.KindConstructorType,
		ast.KindFunctionType, ast.KindMethodSignature:
		return true
	default:
		return false
	}
}

func (r *nameReplacements) renameFixes(variable *scope.Variable, variables []*scope.Variable, references []*scope.Reference, replacement string) []rule.RuleFix {
	nodes := make([]*ast.Node, 0, len(variables)+len(references))
	for _, declaration := range variables {
		nodes = append(nodes, declaration.ID)
	}
	for _, reference := range references {
		nodes = append(nodes, reference.Identifier)
	}
	seen := make(map[core.TextRange]struct{}, len(nodes))
	fixes := make([]rule.RuleFix, 0, len(nodes))
	for _, node := range nodes {
		textRange, text := r.replacementForIdentifier(node, replacement)
		if _, duplicate := seen[textRange]; duplicate {
			continue
		}
		seen[textRange] = struct{}{}
		fixes = append(fixes, rule.RuleFixReplaceRange(textRange, text))
	}
	sort.Slice(fixes, func(i, j int) bool { return fixes[i].Range.Pos() < fixes[j].Range.Pos() })
	return fixes
}

func (r *nameReplacements) replacementForIdentifier(node *ast.Node, replacement string) (core.TextRange, string) {
	textRange := utils.TrimNodeTextRange(r.ctx.SourceFile, node)
	parent := node.Parent
	if parent == nil {
		return textRange, replacement
	}
	bindingRange := utils.GetESTreeBindingIdentifierRange(r.ctx.SourceFile, node)
	if bindingRange != textRange && r.rangeContainsComment(bindingRange) {
		return textRange, replacement
	}
	if parent.Kind == ast.KindParameter && parent.Name() == node {
		parameter := parent.AsParameterDeclaration()
		if parameter.QuestionToken != nil && parameter.Type != nil {
			if colon, ok := utils.TokenBeforePosition(r.ctx.SourceFile, parameter.Type.Pos()); ok &&
				colon.Kind == ast.KindColonToken && colon.Start >= textRange.End() {
				return core.NewTextRange(textRange.Pos(), colon.Start), replacement + "?"
			}
		}
	}
	if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
		return textRange, node.Text() + ": " + replacement
	}
	if parent.Kind == ast.KindBindingElement {
		binding := parent.AsBindingElement()
		if binding.PropertyName == nil && parent.Parent != nil && parent.Parent.Kind == ast.KindObjectBindingPattern {
			return textRange, node.Text() + ": " + replacement
		}
	}
	if parent.Kind == ast.KindImportSpecifier {
		specifier := parent.AsImportSpecifier()
		if specifier.PropertyName == nil {
			return textRange, node.Text() + " as " + replacement
		}
	}
	if parent.Kind == ast.KindExportSpecifier {
		specifier := parent.AsExportSpecifier()
		if specifier.PropertyName == nil {
			return textRange, replacement + " as " + node.Text()
		}
	}
	return textRange, replacement
}

func (r *nameReplacements) rangeContainsComment(textRange core.TextRange) bool {
	comments := r.ctx.Comments.All()
	index := sort.Search(len(comments), func(index int) bool {
		return comments[index].End() > textRange.Pos()
	})
	for ; index < len(comments) && comments[index].Pos() < textRange.End(); index++ {
		comment := comments[index]
		if comment.Pos() >= textRange.Pos() && comment.End() <= textRange.End() {
			return true
		}
	}
	return false
}

func (r *nameReplacements) renameSuggestions(variable *scope.Variable, variables []*scope.Variable, references []*scope.Reference, replacements []string) []rule.RuleSuggestion {
	result := make([]rule.RuleSuggestion, 0, len(replacements))
	for _, replacement := range replacements {
		if !validVariableName(replacement) {
			continue
		}
		result = append(result, rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          messageIDRename,
				Description: "Rename to `" + replacement + "`.",
				Data:        map[string]string{"replacement": replacement},
			},
			FixesArr: r.renameFixes(variable, variables, references, replacement),
		})
	}
	return result
}

func (r *nameReplacements) checkProperty(node *ast.Node) {
	if node == nil || node.Text() == "__proto__" || utils.IsInJSDocSyntax(node) || isJSXNamePosition(node) {
		return
	}
	property, exported := propertyIdentifier(node)
	if !property {
		return
	}
	textRange := utils.TrimNodeTextRange(r.ctx.SourceFile, node)
	if _, duplicate := r.reportedPropertyRanges[textRange]; duplicate {
		return
	}
	replacements := r.nameReplacements(node.Text(), 3)
	if replacements.total == 0 {
		return
	}
	r.reportedPropertyRanges[textRange] = struct{}{}
	message := replacementMessage(node.Text(), replacements, "property")
	if replacements.total > 1 && !exported {
		r.ctx.ReportRangeWithDeferredSuggestions(textRange, message, func() []rule.RuleSuggestion {
			result := make([]rule.RuleSuggestion, 0, len(replacements.samples))
			for _, replacement := range replacements.samples {
				if !scanner.IsValidIdentifier(replacement) {
					continue
				}
				result = append(result, rule.RuleSuggestion{
					Message: rule.RuleMessage{
						Id:          messageIDRename,
						Description: "Rename to `" + replacement + "`.",
						Data:        map[string]string{"replacement": replacement},
					},
					FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(textRange, replacement)},
				})
			}
			return result
		})
		return
	}
	r.ctx.ReportRange(textRange, message)
}

func isAssignmentTarget(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindBinaryExpression {
		expression := parent.AsBinaryExpression()
		return expression.Left == current && ast.IsAssignmentOperator(expression.OperatorToken.Kind)
	}
	if parent.Kind == ast.KindPrefixUnaryExpression {
		operator := parent.AsPrefixUnaryExpression().Operator
		return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
	}
	if parent.Kind == ast.KindPostfixUnaryExpression {
		operator := parent.AsPostfixUnaryExpression().Operator
		return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
	}
	return false
}

func propertyIdentifier(node *ast.Node) (property bool, exported bool) {
	parent := node.Parent
	if parent == nil || parent.Kind == ast.KindComputedPropertyName {
		return false, false
	}
	if parent.Kind == ast.KindPropertyAccessExpression && parent.Name() == node {
		return isAssignmentTarget(parent), false
	}
	if parent.Kind == ast.KindExportSpecifier {
		specifier := parent.AsExportSpecifier()
		return specifier.PropertyName != nil && specifier.Name() == node, true
	}
	if parent.Name() != node {
		return false, false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment:
		return parent.Parent != nil && parent.Parent.Kind == ast.KindObjectLiteralExpression, false
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindMethodSignature:
		return true, false
	}
	return false, false
}
