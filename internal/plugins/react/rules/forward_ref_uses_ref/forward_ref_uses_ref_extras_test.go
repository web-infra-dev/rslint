package forward_ref_uses_ref

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestForwardRefUsesRefExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers,
// so future refactors can't silently regress them without breaking a named
// lock-in.
func TestForwardRefUsesRefExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForwardRefUsesRefRule, []rule_tester.ValidTestCase{
		// ---- Dimension 4: parameter-count boundary ----
		{Code: `forwardRef(() => null);`, Tsx: true},
		{Code: `forwardRef((props, ref, extra) => null);`, Tsx: true},
		{Code: `forwardRef((props, ...ref) => null);`, Tsx: true},

		// ---- Dimension 4: parenthesized callee and callback with a valid ref ----
		{Code: `(React.forwardRef)(((props, ref) => null));`, Tsx: true},

		// ---- Dimension 4: TS wrappers are upstream-opaque ----
		{Code: `forwardRef(((props) => null) as any);`, Tsx: true},
		{Code: `forwardRef(((props) => null)!);`, Tsx: true},
		{Code: `forwardRef(((props) => null) satisfies any);`, Tsx: true},
		{Code: `(React.forwardRef as any)((props) => null);`, Tsx: true},
		{Code: `forwardRef(((props, ref) => null) as any);`, Tsx: true},
		{Code: `(React.forwardRef as any)((props, ref) => null);`, Tsx: true},

		// ---- Dimension 4: computed property access is not a forwardRef property identifier ----
		{Code: `React["forwardRef"]((props) => null);`, Tsx: true},
		{Code: "React[`forwardRef`]((props) => null);", Tsx: true},

		// ---- Dimension 4: graceful degradation for empty and non-function arguments ----
		{Code: `forwardRef();`, Tsx: true},
		{Code: `forwardRef(Component);`, Tsx: true},
		{Code: `forwardReference((props) => null);`, Tsx: true},
		{Code: `React.forwardRef2((props) => null);`, Tsx: true},
		{Code: `React.forwardRef.call(null, (props) => null);`, Tsx: true},
		{Code: `React.forwardRef.apply(null, [(props) => null]);`, Tsx: true},

		// ---- Real-user: #3684 shadcn/ui typed React.forwardRef with destructured props ----
		{
			Code: `const FormItem = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    const id = React.useId();
    return <div className={className} ref={ref} {...props}>{id}</div>;
  }
);`,
			Tsx: true,
		},

		// ---- Real-user: #3524 React.memo + typed React.forwardRef nesting ----
		{
			Code: `const LazyImage = React.memo<Props>(
  React.forwardRef<Image, Props>(
    ({ foo }, ref) => {
      console.log("foo", foo);
      return <Image ref={ref} />;
    }
  )
);`,
			Tsx: true,
		},

		// ---- Real-user: #3738 shadcn/ui default argument under forwardRef ----
		{
			Code: `const BreadcrumbLink = React.forwardRef<
  HTMLAnchorElement,
  React.ComponentPropsWithoutRef<"a"> & { asChild?: boolean }
>(({ asChild = false, className, ...props }, ref) => {
  const Comp = asChild ? Slot : "a";
  return <Comp ref={ref} className={className} {...props} />;
});`,
			Tsx: true,
		},

		// ---- Real-user: #2172 export default React.forwardRef sibling wrapper shape ----
		{
			Code: `export default React.forwardRef((props, ref) => <StoreListItem {...props} forwardRef={ref} />);`,
			Tsx:  true,
		},

		// N/A: Object/class property key forms do not apply; this rule only
		// inspects call callees and function parameters.
		// N/A: Class declaration / class expression containers do not apply;
		// upstream only listens for function and arrow expressions.
		// N/A: Body-absent overload / abstract / declare members do not apply;
		// the inspected callbacks are expression nodes with parameter lists.
	}, []rule_tester.InvalidTestCase{
		// ---- Real-user: #3158 original missing-ref request ----
		{
			Code: `const Button = forwardRef(props => <button {...props} />);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 27,
				`const Button = forwardRef((props, ref) => <button {...props} />);`,
				`const Button = props => <button {...props} />;`)},
			Tsx: true,
		},

		// ---- Dimension 4: nested wrapper chain ----
		{
			Code: `React.memo(React.forwardRef((props) => null));`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 29,
				`React.memo(React.forwardRef((props, ref) => null));`,
				`React.memo((props) => null);`)},
			Tsx: true,
		},

		// ---- Dimension 4: parenthesized callee ----
		// Locks in upstream isForwardRefCall() arm 1: Identifier callee.
		{
			Code: `(forwardRef)((props) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 14,
				`(forwardRef)((props, ref) => null);`,
				`(props) => null;`)},
			Tsx: true,
		},

		// ---- Dimension 4: optional bare call ----
		{
			Code: `forwardRef?.((props) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 14,
				`forwardRef?.((props, ref) => null);`,
				`(props) => null;`)},
			Tsx: true,
		},
		{
			Code: `forwardRef?.(props => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 14,
				`forwardRef?.((props, ref) => null);`,
				`props => null;`)},
			Tsx: true,
		},

		// ---- Dimension 4: optional member expression ----
		// Locks in upstream isForwardRefCall() arm 2: MemberExpression property.
		{
			Code: `React?.forwardRef((props) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 19,
				`React?.forwardRef((props, ref) => null);`,
				`(props) => null;`)},
			Tsx: true,
		},
		{
			Code: `React.forwardRef?.(props => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 20,
				`React.forwardRef?.((props, ref) => null);`,
				`props => null;`)},
			Tsx: true,
		},
		{
			Code: `React.forwardRef?.((props) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 20,
				`React.forwardRef?.((props, ref) => null);`,
				`(props) => null;`)},
			Tsx: true,
		},

		// ---- Dimension 4: parenthesized callback parent ----
		{
			Code: `forwardRef((((props) => null)));`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 14,
				`forwardRef((((props, ref) => null)));`,
				`(props) => null;`)},
			Tsx: true,
		},

		// Locks in upstream isForwardRefCall() arm 2: object name is ignored.
		{
			Code: `Other.forwardRef((props) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 18,
				`Other.forwardRef((props, ref) => null);`,
				`(props) => null;`)},
			Tsx: true,
		},

		// ---- Dimension 4: typed generic forwardRef call ----
		{
			Code: `React.forwardRef<HTMLDivElement, Props>((props: Props) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 41,
				`React.forwardRef<HTMLDivElement, Props>((props: Props, ref) => null);`,
				`(props: Props) => null;`)},
			Tsx: true,
		},

		// ---- Dimension 4: parameter forms ----
		{
			Code: `forwardRef(({ className, ...props }) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 12,
				`forwardRef(({ className, ...props }, ref) => null);`,
				`({ className, ...props }) => null;`)},
			Tsx: true,
		},
		{
			Code: `forwardRef((props = {}) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 12,
				`forwardRef((props = {}, ref) => null);`,
				`(props = {}) => null;`)},
			Tsx: true,
		},
		{
			Code: `forwardRef((props /* keep */) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 12,
				`forwardRef((props, ref /* keep */) => null);`,
				`(props /* keep */) => null;`)},
			Tsx: true,
		},
		{
			Code: `forwardRef((...props) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 12,
				`forwardRef((...props, ref) => null);`,
				`(...props) => null;`)},
			Tsx: true,
		},
		{
			Code: `forwardRef(async props => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 12,
				`forwardRef(async (props, ref) => null);`,
				`async props => null;`)},
			Tsx: true,
		},
		{
			Code: `forwardRef(function* Component(props) { return null; });`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 12,
				`forwardRef(function* Component(props, ref) { return null; });`,
				`function* Component(props) { return null; };`)},
			Tsx: true,
		},

		// Locks in upstream listener shape: direct function argument position is
		// not checked, only the parent CallExpression callee.
		{
			Code: `forwardRef(Component, (props) => null);`,
			Errors: []rule_tester.InvalidTestCaseError{missingRefError(1, 23,
				`forwardRef(Component, (props, ref) => null);`,
				`(props) => null;`)},
			Tsx: true,
		},
	})
}
