package comma_dangle

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	optionNever           = "never"
	optionAlways          = "always"
	optionAlwaysMultiline = "always-multiline"
	optionOnlyMultiline   = "only-multiline"
	optionIgnore          = "ignore"
)

// normalizedOptions is the per-slot config the predicates consult. Upstream
// fans out a single string option to every slot; the object form lets each
// slot be set independently. Every slot uses the same enum (never / always /
// always-multiline / only-multiline / ignore).
type normalizedOptions struct {
	arrays           string
	objects          string
	imports          string
	exports          string
	functions        string
	importAttributes string
	dynamicImports   string
	enums            string
	generics         string
	tuples           string
}

func defaultOptions() normalizedOptions {
	return normalizedOptions{
		arrays:           optionNever,
		objects:          optionNever,
		imports:          optionNever,
		exports:          optionNever,
		functions:        optionNever,
		importAttributes: optionNever,
		dynamicImports:   optionNever,
		enums:            optionNever,
		generics:         optionNever,
		tuples:           optionNever,
	}
}

func validOptionValue(s string) bool {
	switch s {
	case optionNever, optionAlways, optionAlwaysMultiline, optionOnlyMultiline:
		return true
	}
	return false
}

func validValueWithIgnore(s string) bool {
	return s == optionIgnore || validOptionValue(s)
}

// parseOptions mirrors upstream's normalizeOptions. Accepted shapes:
//
//	['comma-dangle']                                    → defaults (never everywhere)
//	['comma-dangle', 'always']                          → broadcast 'always'
//	['comma-dangle', { arrays: 'always', ... }]         → per-slot override
//
// rslint's config loader collapses a single trailing option element into the
// option directly, so accept both `[]interface{}` (Go test / multi-element
// config) and a bare string / map (CLI single-option config).
func parseOptions(raw any) normalizedOptions {
	opts := defaultOptions()

	var primary any
	switch v := raw.(type) {
	case []interface{}:
		if len(v) > 0 {
			primary = v[0]
		}
	case string, map[string]interface{}:
		primary = v
	}

	switch v := primary.(type) {
	case string:
		if validOptionValue(v) {
			opts.arrays = v
			opts.objects = v
			opts.imports = v
			opts.exports = v
			opts.functions = v
			opts.importAttributes = v
			opts.dynamicImports = v
			opts.enums = v
			opts.generics = v
			opts.tuples = v
		}
	case map[string]interface{}:
		set := func(key string, dst *string) {
			if s, ok := v[key].(string); ok && validValueWithIgnore(s) {
				*dst = s
			}
		}
		set("arrays", &opts.arrays)
		set("objects", &opts.objects)
		set("imports", &opts.imports)
		set("exports", &opts.exports)
		set("functions", &opts.functions)
		set("importAttributes", &opts.importAttributes)
		set("dynamicImports", &opts.dynamicImports)
		set("enums", &opts.enums)
		set("generics", &opts.generics)
		set("tuples", &opts.tuples)
	}
	return opts
}

// verifyInfo holds everything a predicate needs to inspect a single trailing
// position. `list` is one of object props / array elements / function params /
// etc.; `closePos` is the byte position of the bracket that terminates it
// (`}`, `]`, `)`, or `>`).
type verifyInfo struct {
	list     *ast.NodeList
	closePos int
	// lastItem is the last non-hole child of list. tsgo represents ESTree
	// `null` array holes as KindOmittedExpression; we filter them out here
	// for parity with upstream's `last(nodes)` (which the `??` short-circuits).
	lastItem *ast.Node
	// isRest is true when lastItem is a rest-binding form (ParameterDeclaration
	// or BindingElement with `...`). In that case `always`-family options
	// degrade to `never` — JS forbids a trailing comma after a rest binding.
	isRest bool
}

// buildInfo scans forward from list.End() past trivia to locate the close
// bracket. Returns nil when there's nothing to check (list nil/empty, ends
// in a hole, or close-bracket character isn't where we'd expect — parser
// recovery). `container` is the AST node that owns the list (ArrayLiteral,
// ObjectLiteral, FunctionDeclaration, etc.), used to disambiguate
// SpreadElement / SpreadAssignment in destructuring-LHS position.
func buildInfo(text string, list *ast.NodeList, closeChar byte, container *ast.Node) *verifyInfo {
	if list == nil || len(list.Nodes) == 0 {
		return nil
	}
	last := list.Nodes[len(list.Nodes)-1]
	if last == nil || last.Kind == ast.KindOmittedExpression {
		return nil
	}
	closePos := scanner.SkipTrivia(text, list.End())
	if closePos >= len(text) || text[closePos] != closeChar {
		return nil
	}
	return &verifyInfo{
		list:     list,
		closePos: closePos,
		lastItem: last,
		isRest:   isRestElement(last, container),
	}
}

// isRestElement mirrors upstream's `lastItem.type !== 'RestElement'` guard.
// In ESTree, RestElement only appears in patterns (function params, array /
// object destructuring); spread in array literals (SpreadElement) and object
// literals (SpreadAssignment) does NOT count and may take a trailing comma.
//
// tsgo doesn't fork ArrayLiteralExpression / ObjectLiteralExpression into
// pattern forms — `[a, ...rest] = []` parses as an ArrayLiteralExpression
// with a SpreadElement, same shape as the literal `[a, ...spread]`. We
// therefore treat SpreadElement / SpreadAssignment as rest only when the
// containing literal sits in a destructuring-LHS position, detected via
// `IsArrayLiteralOrObjectLiteralDestructuringPattern` (which walks up
// through PropertyAssignment and nested patterns).
func isRestElement(node *ast.Node, container *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindParameter:
		if pd := node.AsParameterDeclaration(); pd != nil {
			return pd.DotDotDotToken != nil
		}
	case ast.KindBindingElement:
		if be := node.AsBindingElement(); be != nil {
			return be.DotDotDotToken != nil
		}
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return container != nil && ast.IsArrayLiteralOrObjectLiteralDestructuringPattern(container)
	}
	return false
}

// findTrailingCommaPos returns the position of the trailing comma in
// [lastItem.End(), closePos), or -1 when no comma is present. The only
// non-trivia content allowed in that range is a single `,`.
func findTrailingCommaPos(text string, info *verifyInfo) int {
	pos := scanner.SkipTrivia(text, info.lastItem.End())
	if pos < info.closePos && text[pos] == ',' {
		return pos
	}
	return -1
}

// isMultiline mirrors upstream's `isMultiline(info)`. The "last token" is
// the comma if it exists, otherwise the last item — matching how
// `sourceCode.getTokenAfter(lastItem)` resolves in upstream.
func isMultiline(text string, info *verifyInfo, commaPos int) bool {
	var refEnd int
	if commaPos >= 0 {
		refEnd = commaPos + 1
	} else {
		refEnd = info.lastItem.End()
	}
	return utils.ContainsLineTerminator(text, refEnd, info.closePos)
}

// predicateContext bundles the inputs the predicate functions all read.
type predicateContext struct {
	ctx   rule.RuleContext
	text  string
	isTSX bool
}

// reportUnexpected emits the "unexpected trailing comma" diagnostic with
// a removal fix. The diagnostic range is the single-char span of the comma,
// matching upstream's `loc: trailingToken.loc`.
func (pc *predicateContext) reportUnexpected(commaPos int) {
	pc.ctx.ReportRangeWithFixes(
		core.NewTextRange(commaPos, commaPos+1),
		rule.RuleMessage{
			Id:          "unexpected",
			Description: "Unexpected trailing comma.",
		},
		rule.RuleFix{
			Text:  "",
			Range: core.NewTextRange(commaPos, commaPos+1),
		},
	)
}

// reportMissing emits the "missing trailing comma" diagnostic with an
// insertion fix. Range is `[lastItem.End(), lastItem.End()+1)` — a 1-char
// span starting right after the last token. When that span crosses a
// newline (last item ends a line, `}` opens the next), the line/col
// renderer maps the end position to (line+1, col 1), matching upstream's
// `getNextLocation` shape.
func (pc *predicateContext) reportMissing(info *verifyInfo) {
	insertPos := info.lastItem.End()
	endPos := insertPos + 1
	// CRLF special-case: when the byte right after the last token is `\r`
	// and the next is `\n`, extend the diagnostic span past the whole
	// `\r\n` so the line/col renderer correctly resolves the end to the
	// start of the next line. Without this, end maps to "line N col X+1"
	// (still inside the CRLF sequence) instead of "line N+1 col 1",
	// silently diverging from ESLint's `getNextLocation` shape.
	if endPos < len(pc.text) && pc.text[insertPos] == '\r' && pc.text[endPos] == '\n' {
		endPos++
	}
	if endPos > len(pc.text) {
		endPos = len(pc.text)
	}
	pc.ctx.ReportRangeWithFixes(
		core.NewTextRange(insertPos, endPos),
		rule.RuleMessage{
			Id:          "missing",
			Description: "Missing trailing comma.",
		},
		rule.RuleFix{
			Text:  ",",
			Range: core.NewTextRange(insertPos, insertPos),
		},
	)
}

// forbidTrailingComma mirrors upstream's `forbidTrailingComma`. `tsxCarveOut`
// is the upstream TSX `<T,>` exemption: a single-element TypeParameterDecl in
// a `.tsx` file is the JSX-disambiguation form and must not be reported, no
// matter which option-arm we landed in.
func (pc *predicateContext) forbidTrailingComma(info *verifyInfo, tsxCarveOut bool) {
	if tsxCarveOut {
		return
	}
	commaPos := findTrailingCommaPos(pc.text, info)
	if commaPos >= 0 {
		pc.reportUnexpected(commaPos)
	}
}

// forceTrailingComma mirrors upstream's `forceTrailingComma`. For a rest
// binding the call degrades to forbidTrailingComma (with carve-out), since
// JS forbids a trailing comma after a rest.
func (pc *predicateContext) forceTrailingComma(info *verifyInfo, tsxCarveOut bool) {
	if info.isRest {
		pc.forbidTrailingComma(info, tsxCarveOut)
		return
	}
	commaPos := findTrailingCommaPos(pc.text, info)
	if commaPos >= 0 {
		return
	}
	pc.reportMissing(info)
}

func (pc *predicateContext) forceTrailingCommaIfMultiline(info *verifyInfo, tsxCarveOut bool) {
	commaPos := findTrailingCommaPos(pc.text, info)
	if isMultiline(pc.text, info, commaPos) {
		pc.forceTrailingComma(info, tsxCarveOut)
	} else {
		pc.forbidTrailingComma(info, tsxCarveOut)
	}
}

func (pc *predicateContext) allowTrailingCommaIfMultiline(info *verifyInfo, tsxCarveOut bool) {
	commaPos := findTrailingCommaPos(pc.text, info)
	if !isMultiline(pc.text, info, commaPos) {
		pc.forbidTrailingComma(info, tsxCarveOut)
	}
}

// dispatch routes a verifyInfo to the predicate selected by `option`.
// `tsxCarveOut` is the `<T,>` exemption (see forbidTrailingComma).
func (pc *predicateContext) dispatch(info *verifyInfo, option string, tsxCarveOut bool) {
	if info == nil || option == optionIgnore {
		return
	}
	switch option {
	case optionAlways:
		pc.forceTrailingComma(info, tsxCarveOut)
	case optionAlwaysMultiline:
		pc.forceTrailingCommaIfMultiline(info, tsxCarveOut)
	case optionOnlyMultiline:
		pc.allowTrailingCommaIfMultiline(info, tsxCarveOut)
	default: // "never" (also covers unknown strings — same as upstream)
		pc.forbidTrailingComma(info, tsxCarveOut)
	}
}

// listSpec describes one list to verify. `closeChar` is the bracket expected
// to close it, used when scanning past trivia from list.End().
type listSpec struct {
	list      *ast.NodeList
	closeChar byte
	option    string
	// container is the AST node that owns the list (ArrayLiteralExpression,
	// FunctionDeclaration, NamedImports, etc.). Passed through to
	// isRestElement so SpreadElement / SpreadAssignment can be recognized as
	// rest when the literal is in destructuring-LHS position.
	container *ast.Node
	// isTypeParam=true marks a TypeParameters list, which gets the TSX `<T,>`
	// carve-out when the list has exactly one parameter.
	isTypeParam bool
}

func (pc *predicateContext) check(spec listSpec) {
	info := buildInfo(pc.text, spec.list, spec.closeChar, spec.container)
	if info == nil {
		return
	}
	tsxCarveOut := pc.isTSX && spec.isTypeParam && len(spec.list.Nodes) == 1
	pc.dispatch(info, spec.option, tsxCarveOut)
}

// CommaDangleRule enforces consistent use of trailing commas in object / array
// literals, parameter lists, import / export specifiers, dynamic imports,
// import attributes, TS enums, TS tuple types, TS generics, and TS function
// types. Ported from @stylistic/eslint-plugin's comma-dangle.
//
// Listener fan-out vs. upstream ESTree:
//
//   - tsgo collapses ESTree's MethodDefinition.value = FunctionExpression
//     into a single MethodDeclaration / accessor / Constructor node, so
//     the `functions` option needs explicit listeners on those node kinds
//     to match upstream's "all class-method param lists" coverage.
//   - tsgo has no separate `TSDeclareFunction` kind; `declare function f()`
//     is a body-less FunctionDeclaration, naturally covered by the
//     FunctionDeclaration listener.
//   - tsgo has no `ImportExpression` kind; dynamic `import(source, opts)`
//     parses as a CallExpression whose `Expression.Kind` is `ImportKeyword`,
//     which is what we branch on inside the CallExpression listener to
//     pick the `dynamicImports` slot instead of `functions`.
//   - tsgo has no `TSTypeParameterDeclaration` / `TSTypeParameterInstantiation`
//     node kinds; TypeParameters / TypeArguments are fields on a wide set of
//     declaration / type nodes. We delegate to `node.TypeArgumentList()` /
//     `node.TypeParameterList()` (typescript-go/internal/ast/ast.go:483, 515)
//     so the carrier-kind set stays in lockstep with tsgo. Hand-rolling a
//     per-kind switch is what caused JsxOpeningElement / JsxSelfClosingElement
//     to be silently skipped in an earlier revision.
var CommaDangleRule = rule.Rule{
	Name: "@stylistic/comma-dangle",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		text := ctx.SourceFile.Text()
		isTSX := strings.HasSuffix(strings.ToLower(ctx.SourceFile.FileName()), ".tsx")
		pc := &predicateContext{ctx: ctx, text: text, isTSX: isTSX}

		// checkTypeArgs handles every TypeArguments-carrying node (`Bar<T>`,
		// `foo<T>()`, `new Foo<T>()`, `tag<T>`...``, `import('x').Y<T>`,
		// `typeof x.Y<T>`, `<Foo<T> />`, `<Foo<T>>x</Foo>`). Upstream's rule
		// hard-codes `predicate.never` for type-argument lists
		// (TSTypeParameterInstantiation), so the `generics` slot is bypassed
		// in favor of a literal 'never'. Delegating to tsgo's
		// `node.TypeArgumentList()` keeps the list of carrier kinds in one
		// place — manually fanning out per-kind via `node.AsXxx().TypeArguments`
		// is what caused JsxOpeningElement / JsxSelfClosingElement to be
		// dropped in an earlier revision.
		checkTypeArgs := func(node *ast.Node) {
			if list := node.TypeArgumentList(); list != nil {
				pc.check(listSpec{list: list, closeChar: '>', option: optionNever, container: node})
			}
		}

		// checkTypeParams handles every TypeParameters-carrying node — function
		// / method / class / interface / type-alias declarations and their
		// expression forms. Delegates to tsgo's `node.TypeParameterList()` for
		// the same reason as `checkTypeArgs`. The carve-out for TSX `<T,>` lives
		// in `check` via `isTypeParam: true` + 1-element list + .tsx filename.
		checkTypeParams := func(node *ast.Node) {
			if list := node.TypeParameterList(); list != nil {
				pc.check(listSpec{list: list, closeChar: '>', option: opts.generics, container: node, isTypeParam: true})
			}
		}

		checkFunctionLike := func(node *ast.Node) {
			// Body-less class members (abstract methods, `declare class`
			// methods, overload signatures) are parsed by @typescript-eslint
			// as `TSEmptyBodyFunctionExpression` — distinct from
			// `FunctionExpression` — so upstream's `FunctionExpression`
			// listener does not match them, and no diagnostic is emitted.
			// tsgo collapses these into the same `MethodDeclaration` /
			// `Constructor` / accessor kinds with a nil Body. Detect and
			// skip the Parameters check (but still walk TypeParameters —
			// upstream still visits TSTypeParameterDeclaration on
			// `TSAbstractMethodDefinition`, though that requires a real
			// abstract method to test).
			//
			// FunctionDeclaration is intentionally not part of this guard:
			// upstream visits both `FunctionDeclaration` (with body) and
			// `TSDeclareFunction` (body-less); collapsing both in tsgo's
			// FunctionDeclaration means we must always check.
			skipParams := false
			switch node.Kind {
			case ast.KindMethodDeclaration, ast.KindConstructor,
				ast.KindGetAccessor, ast.KindSetAccessor:
				if node.Body() == nil {
					skipParams = true
				}
			}
			fl := node.FunctionLikeData()
			if !skipParams && fl != nil && fl.Parameters != nil {
				pc.check(listSpec{list: fl.Parameters, closeChar: ')', option: opts.functions, container: node})
			}
			checkTypeParams(node)
		}

		return rule.RuleListeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if obj := node.AsObjectLiteralExpression(); obj != nil {
					pc.check(listSpec{list: obj.Properties, closeChar: '}', option: opts.objects, container: node})
				}
			},
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				if bp := node.AsBindingPattern(); bp != nil {
					pc.check(listSpec{list: bp.Elements, closeChar: '}', option: opts.objects, container: node})
				}
			},
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				if arr := node.AsArrayLiteralExpression(); arr != nil {
					pc.check(listSpec{list: arr.Elements, closeChar: ']', option: opts.arrays, container: node})
				}
			},
			ast.KindArrayBindingPattern: func(node *ast.Node) {
				if bp := node.AsBindingPattern(); bp != nil {
					pc.check(listSpec{list: bp.Elements, closeChar: ']', option: opts.arrays, container: node})
				}
			},
			ast.KindImportDeclaration: func(node *ast.Node) {
				id := node.AsImportDeclaration()
				if id == nil {
					return
				}
				// imports: only when NamedBindings is a NamedImports `{ a, b }`
				// group (not a namespace import or a bare default import).
				if id.ImportClause != nil {
					if ic := id.ImportClause.AsImportClause(); ic != nil &&
						ic.NamedBindings != nil &&
						ic.NamedBindings.Kind == ast.KindNamedImports {
						named := ic.NamedBindings.AsNamedImports()
						pc.check(listSpec{list: named.Elements, closeChar: '}', option: opts.imports, container: ic.NamedBindings})
					}
				}
				// importAttributes: `with { type: 'json' }` clause.
				if id.Attributes != nil {
					if attrs := id.Attributes.AsImportAttributes(); attrs != nil {
						pc.check(listSpec{list: attrs.Attributes, closeChar: '}', option: opts.importAttributes, container: id.Attributes})
					}
				}
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				ed := node.AsExportDeclaration()
				if ed == nil {
					return
				}
				if ed.ExportClause != nil && ed.ExportClause.Kind == ast.KindNamedExports {
					named := ed.ExportClause.AsNamedExports()
					pc.check(listSpec{list: named.Elements, closeChar: '}', option: opts.exports, container: ed.ExportClause})
				}
				if ed.Attributes != nil {
					if attrs := ed.Attributes.AsImportAttributes(); attrs != nil {
						pc.check(listSpec{list: attrs.Attributes, closeChar: '}', option: opts.importAttributes, container: ed.Attributes})
					}
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil {
					return
				}
				// Dynamic `import(...)` lives in the `dynamicImports` slot;
				// regular calls go to `functions`.
				argOption := opts.functions
				if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword {
					argOption = opts.dynamicImports
				}
				if call.Arguments != nil {
					pc.check(listSpec{list: call.Arguments, closeChar: ')', option: argOption, container: node})
				}
				checkTypeArgs(node)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				ne := node.AsNewExpression()
				if ne == nil {
					return
				}
				if ne.Arguments != nil {
					pc.check(listSpec{list: ne.Arguments, closeChar: ')', option: opts.functions, container: node})
				}
				checkTypeArgs(node)
			},
			// Pure type-argument carriers — delegating to `node.TypeArgumentList()`
			// so the set of kinds stays in lockstep with tsgo's own dispatch.
			ast.KindTaggedTemplateExpression:    checkTypeArgs,
			ast.KindTypeReference:               checkTypeArgs,
			ast.KindExpressionWithTypeArguments: checkTypeArgs,
			ast.KindImportType:                  checkTypeArgs,
			ast.KindTypeQuery:                   checkTypeArgs,
			ast.KindJsxOpeningElement:           checkTypeArgs,
			ast.KindJsxSelfClosingElement:       checkTypeArgs,
			// Function-like kinds — Parameters (`functions`) + TypeParameters
			// (`generics`). Mirrors upstream's listener set:
			//   FunctionDeclaration, FunctionExpression, ArrowFunctionExpression,
			//   TSDeclareFunction, TSFunctionType — plus the class-context
			//   FunctionExpression sites that ESTree wraps in MethodDefinition
			//   (collapsed in tsgo into MethodDeclaration / Constructor /
			//   Get|SetAccessor directly).
			//
			// NOT included: `KindConstructorType` (TS `new (...) => T`).
			// Upstream has no `TSConstructorType` listener — verified by running
			// @stylistic/eslint-plugin on a test corpus. Adding it would
			// over-report on every `new (a, b,) => T` shape that upstream
			// silently accepts.
			ast.KindFunctionDeclaration: checkFunctionLike,
			ast.KindFunctionExpression:  checkFunctionLike,
			ast.KindArrowFunction:       checkFunctionLike,
			ast.KindMethodDeclaration:   checkFunctionLike,
			ast.KindConstructor:         checkFunctionLike,
			ast.KindGetAccessor:         checkFunctionLike,
			ast.KindSetAccessor:         checkFunctionLike,
			ast.KindFunctionType:        checkFunctionLike,
			// Class- / interface- / alias-like kinds — generics only.
			ast.KindClassDeclaration:     checkTypeParams,
			ast.KindClassExpression:      checkTypeParams,
			ast.KindInterfaceDeclaration: checkTypeParams,
			ast.KindTypeAliasDeclaration: checkTypeParams,
			ast.KindEnumDeclaration: func(node *ast.Node) {
				if ed := node.AsEnumDeclaration(); ed != nil {
					pc.check(listSpec{list: ed.Members, closeChar: '}', option: opts.enums, container: node})
				}
			},
			ast.KindTupleType: func(node *ast.Node) {
				if tt := node.AsTupleTypeNode(); tt != nil {
					pc.check(listSpec{list: tt.Elements, closeChar: ']', option: opts.tuples, container: node})
				}
			},
		}
	},
}
