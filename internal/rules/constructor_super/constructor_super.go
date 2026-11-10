package constructor_super

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"sort"
)

func buildMissingAll() rule.RuleMessage {
	return rule.RuleMessage{Id: "missingAll", Description: "Expected to call 'super()'."}
}
func buildMissingSome() rule.RuleMessage {
	return rule.RuleMessage{Id: "missingSome", Description: "Lacked a call of 'super()' in some code paths."}
}
func buildDuplicate() rule.RuleMessage {
	return rule.RuleMessage{Id: "duplicate", Description: "Unexpected duplicate 'super()'."}
}
func buildBadSuper() rule.RuleMessage {
	return rule.RuleMessage{Id: "badSuper", Description: "Unexpected 'super()' because 'super' is not a constructor."}
}

func superClass(node *ast.Node) *ast.Node {
	clauses := utils.GetHeritageClauses(node)
	if clauses != nil {
		for _, c := range clauses.Nodes {
			h := c.AsHeritageClause()
			if h.Token == ast.KindExtendsKeyword && h.Types != nil && len(h.Types.Nodes) > 0 {
				return h.Types.Nodes[0].Expression()
			}
		}
	}
	return nil
}

// Match ESLint's syntactic possibility check, without resolving values or types.
func isPossibleConstructor(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	switch node.Kind {
	case ast.KindClassExpression, ast.KindFunctionExpression, ast.KindThisKeyword,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression,
		ast.KindNewExpression, ast.KindYieldExpression, ast.KindTaggedTemplateExpression, ast.KindMetaProperty:
		return true
	case ast.KindIdentifier:
		return node.Text() != "undefined"
	case ast.KindBinaryExpression:
		b := node.AsBinaryExpression()
		switch b.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindAmpersandAmpersandToken, ast.KindCommaToken:
			return isPossibleConstructor(b.Right)
		case ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return isPossibleConstructor(b.Left) || isPossibleConstructor(b.Right)
		}
	case ast.KindConditionalExpression:
		c := node.AsConditionalExpression()
		return isPossibleConstructor(c.WhenTrue) || isPossibleConstructor(c.WhenFalse)
	}
	return false
}

// A state is a set of paths: one bit for paths without super/return-value,
// one for paths with it. Joining branches never enumerates individual paths.
type paths uint8

const (
	uncalled paths = 1
	called   paths = 2
)

type completion struct {
	kind  ast.Kind
	label string
}
type flow map[completion]paths

var normal = completion{}

func (f flow) add(g flow) {
	for k, s := range g {
		f[k] |= s
	}
}
func (f flow) take(k completion) paths { s := f[k]; delete(f, k); return s }

type statementKey struct {
	node  *ast.Node
	state paths
	label string
}
type statementResult struct {
	output     flow
	exceptions paths
}

type analyzer struct {
	statementsCache map[statementKey]statementResult
	possible        bool
	reports         map[*ast.Node]rule.RuleMessage
	// The active catch collects states at potentially throwing expressions.
	exceptions *paths
}

func (a *analyzer) mayThrow(s paths) {
	if a.exceptions != nil {
		*a.exceptions |= s
	}
}

// Keep active and short-circuited paths separate until the optional chain
// ends. Joining at each link would incorrectly execute later call arguments
// and computed keys on the skipped paths.
func (a *analyzer) optionalChain(n *ast.Node, s paths) (paths, paths) {
	receiver := n.Expression()
	var active, skipped paths
	if receiver != nil && receiver.Flags&ast.NodeFlagsOptionalChain != 0 {
		active, skipped = a.optionalChain(receiver, s)
	} else {
		active = a.expression(receiver, s)
	}
	var token *ast.Node
	switch n.Kind {
	case ast.KindCallExpression:
		token = n.AsCallExpression().QuestionDotToken
	case ast.KindPropertyAccessExpression:
		token = n.AsPropertyAccessExpression().QuestionDotToken
	case ast.KindElementAccessExpression:
		token = n.AsElementAccessExpression().QuestionDotToken
	}
	if token != nil {
		skipped |= active
	}
	switch n.Kind {
	case ast.KindCallExpression:
		if args := n.AsCallExpression().Arguments; args != nil {
			for _, arg := range args.Nodes {
				active = a.expression(arg, active)
			}
		}
	case ast.KindElementAccessExpression:
		active = a.expression(n.AsElementAccessExpression().ArgumentExpression, active)
	}
	a.mayThrow(active)
	return active, skipped
}

func (a *analyzer) expression(n *ast.Node, s paths) paths {
	if n == nil || s == 0 {
		return s
	}
	if n.Flags&ast.NodeFlagsOptionalChain != 0 &&
		(n.Kind == ast.KindCallExpression || n.Kind == ast.KindPropertyAccessExpression || n.Kind == ast.KindElementAccessExpression) {
		active, skipped := a.optionalChain(n, s)
		return active | skipped
	}
	if utils.IsFunctionLikeContainer(n) {
		// Computed method names execute in the enclosing scope; bodies do not.
		if name := n.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
			return a.expression(name, s)
		}
		return s
	}
	switch n.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		s = a.expression(superClass(n), s)
		for _, m := range n.Members() {
			if name := m.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
				s = a.expression(name, s)
			}
		}
		return s
	case ast.KindConditionalExpression:
		c := n.AsConditionalExpression()
		s = a.expression(c.Condition, s)
		return a.expression(c.WhenTrue, s) | a.expression(c.WhenFalse, s)
	case ast.KindBinaryExpression:
		b := n.AsBinaryExpression()
		s = a.expression(b.Left, s)
		switch b.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return s | a.expression(b.Right, s)
		}
		return a.expression(b.Right, s)
	}
	n.ForEachChild(func(c *ast.Node) bool { s = a.expression(c, s); return false })
	switch n.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindNewExpression, ast.KindCallExpression:
		a.mayThrow(s)
	}
	if n.Kind == ast.KindCallExpression && n.Expression() != nil && n.Expression().Kind == ast.KindSuperKeyword {
		if s&called != 0 {
			// Upstream only reconsiders initially valid calls on a loop edge;
			// a badSuper call is not also reported as its own duplicate.
			if a.reports[n].Id != "badSuper" {
				a.reports[n] = buildDuplicate()
			}
		} else if !a.possible && a.reports[n].Id != "duplicate" {
			a.reports[n] = buildBadSuper()
		}
		return called
	}
	return s
}
func (a *analyzer) statements(nodes []*ast.Node, s paths) flow {
	out := flow{normal: s}
	for _, n := range nodes {
		s = out.take(normal)
		if s == 0 {
			break
		}
		out.add(a.statement(n, s, ""))
	}
	return out
}

// Memoize the three possible nonempty input states per AST node. Without this,
// nested loops could repeatedly analyze the same subtree at each fixed point.
func (a *analyzer) statement(n *ast.Node, s paths, label string) flow {
	key := statementKey{n, s, label}
	if a.statementsCache == nil {
		a.statementsCache = make(map[statementKey]statementResult)
	}
	if cached, ok := a.statementsCache[key]; ok {
		a.mayThrow(cached.exceptions)
		cloned := flow{}
		cloned.add(cached.output)
		return cloned
	}
	upper := a.exceptions
	exceptions := paths(0)
	a.exceptions = &exceptions
	output := a.statementUncached(n, s, label)
	a.exceptions = upper
	a.mayThrow(exceptions)
	cloned := flow{}
	cloned.add(output)
	a.statementsCache[key] = statementResult{cloned, exceptions}
	return output
}

func (a *analyzer) statementUncached(n *ast.Node, s paths, label string) flow {
	if n == nil || s == 0 {
		return flow{normal: s}
	}
	switch n.Kind {
	case ast.KindBlock:
		return a.statements(n.Statements(), s)
	case ast.KindIfStatement:
		i := n.AsIfStatement()
		s = a.expression(i.Expression, s)
		out := a.statement(i.ThenStatement, s, "")
		out.add(a.statement(i.ElseStatement, s, ""))
		return out
	case ast.KindReturnStatement:
		// Upstream marks returning a value before visiting its expression.
		if n.Expression() != nil {
			s = called
			s = a.expression(n.Expression(), s)
		}
		return flow{completion{kind: ast.KindReturnStatement}: s}
	case ast.KindThrowStatement:
		s = a.expression(n.Expression(), s)
		a.mayThrow(s)
		return flow{completion{kind: ast.KindThrowStatement}: s}
	case ast.KindBreakStatement:
		l := ""
		if n.AsBreakStatement().Label != nil {
			l = n.AsBreakStatement().Label.Text()
		}
		return flow{completion{kind: n.Kind, label: l}: s}
	case ast.KindContinueStatement:
		l := ""
		if n.AsContinueStatement().Label != nil {
			l = n.AsContinueStatement().Label.Text()
		}
		return flow{completion{kind: n.Kind, label: l}: s}
	case ast.KindLabeledStatement:
		l := n.AsLabeledStatement()
		out := a.statement(l.Statement, s, l.Label.Text())
		out[normal] |= out.take(completion{kind: ast.KindBreakStatement, label: l.Label.Text()})
		return out
	case ast.KindWhileStatement:
		w := n.AsWhileStatement()
		return a.loop(w.Statement, w.Expression, nil, s, false, label)
	case ast.KindDoStatement:
		d := n.AsDoStatement()
		return a.loop(d.Statement, d.Expression, nil, s, true, label)
	case ast.KindForStatement:
		f := n.AsForStatement()
		s = a.expression(f.Initializer, s)
		return a.loop(f.Statement, f.Condition, f.Incrementor, s, false, label)
	case ast.KindForInStatement, ast.KindForOfStatement:
		f := n.AsForInOrOfStatement()
		s = a.expression(f.Expression, s)
		s = a.expression(f.Initializer, s)
		// Enumeration can execute zero times, and can repeat.
		return a.loop(f.Statement, n, nil, s, false, label)
	case ast.KindSwitchStatement:
		return a.switchStatement(n, s)
	case ast.KindTryStatement:
		return a.tryStatement(n, s)
	default:
		return flow{normal: a.expression(n, s)}
	}
}
func (a *analyzer) loop(body, condition, update *ast.Node, s paths, do bool, label string) flow {
	out := flow{}
	head := s
	// Two bits make this monotone fixed point converge in at most three visits.
	for {
		entry := head
		if !do && condition != nil && condition.Kind != ast.KindForInStatement && condition.Kind != ast.KindForOfStatement {
			entry = a.expression(condition, entry)
		}
		if !do && condition != nil && condition.Kind == ast.KindFalseKeyword {
			out[normal] |= entry
			break
		}
		if !do && condition != nil && condition.Kind != ast.KindTrueKeyword {
			out[normal] |= entry
		}
		f := a.statement(body, entry, "")
		back := f.take(normal) | f.take(completion{kind: ast.KindContinueStatement})
		if label != "" {
			back |= f.take(completion{kind: ast.KindContinueStatement, label: label})
		}
		out[normal] |= f.take(completion{kind: ast.KindBreakStatement})
		if label != "" {
			out[normal] |= f.take(completion{kind: ast.KindBreakStatement, label: label})
		}
		out.add(f)
		back = a.expression(update, back)
		if do {
			back = a.expression(condition, back)
			if condition != nil && condition.Kind != ast.KindTrueKeyword {
				out[normal] |= back
			}
		}
		// ESLint revisits do-while bodies even for a literal false test.
		next := head | back
		if next == head {
			break
		}
		head = next
	}
	return out
}
func (a *analyzer) switchStatement(n *ast.Node, s paths) flow {
	sw := n.AsSwitchStatement()
	s = a.expression(sw.Expression, s)
	out := flow{}
	fall := paths(0)
	hasDefault := false
	for _, c := range sw.CaseBlock.AsCaseBlock().Clauses.Nodes {
		clause := c.AsCaseOrDefaultClause()
		if c.Kind == ast.KindDefaultClause {
			hasDefault = true
		} else {
			s = a.expression(clause.Expression, s)
		}
		f := a.statements(clause.Statements.Nodes, s|fall)
		fall = f.take(normal)
		out[normal] |= f.take(completion{kind: ast.KindBreakStatement})
		out.add(f)
	}
	out[normal] |= fall
	if !hasDefault {
		out[normal] |= s
	}
	return out
}
func (a *analyzer) tryStatement(n *ast.Node, s paths) flow {
	t := n.AsTryStatement()
	exceptions := paths(0)
	upper := a.exceptions
	a.exceptions = &exceptions
	out := a.statement(t.TryBlock, s, "")
	a.exceptions = upper
	if t.CatchClause != nil {
		// ESLint's catch entry also joins the try's fall-through segment.
		// This keeps empty try/catch paths and calls at the segment end in
		// the same conservative model as its code-path analyzer.
		exceptions |= out[normal] | out.take(completion{kind: ast.KindThrowStatement})
		out.add(a.statement(t.CatchClause.AsCatchClause().Block, exceptions, ""))
	} else if upper != nil {
		*upper |= exceptions
	}
	if t.FinallyBlock != nil {
		// Diagnose against all entries together before transferring each
		// completion separately. Map iteration order must not affect reports.
		entries := paths(0)
		for _, state := range out {
			entries |= state
		}
		a.statement(t.FinallyBlock, entries, "")
		result := flow{}
		for k, state := range out {
			f := a.statement(t.FinallyBlock, state, "")
			result[k] |= f.take(normal)
			result.add(f)
		}
		out = result
	}
	return out
}

var ConstructorSuperRule = rule.Rule{
	Name: "constructor-super", Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{ast.KindConstructor: func(n *ast.Node) {
			if n.Body() == nil || n.Parent == nil {
				return
			}
			base := superClass(n.Parent)

			a := analyzer{possible: isPossibleConstructor(base), reports: map[*ast.Node]rule.RuleMessage{}}
			f := a.statement(n.Body(), uncalled, "")
			returned := f[normal] | f[completion{kind: ast.KindReturnStatement}]
			if base != nil && returned&uncalled != 0 {
				if returned&called != 0 {
					ctx.ReportNode(n, buildMissingSome())
				} else {
					ctx.ReportNode(n, buildMissingAll())
				}
			}
			// A loop or finally block can visit a call repeatedly. Report each
			// call once, in source order, without rescanning nested constructors.
			calls := make([]*ast.Node, 0, len(a.reports))
			for call := range a.reports {
				calls = append(calls, call)
			}
			sort.Slice(calls, func(i, j int) bool { return calls[i].Pos() < calls[j].Pos() })
			for _, call := range calls {
				ctx.ReportNode(call, a.reports[call])
			}
		}}
	},
}
