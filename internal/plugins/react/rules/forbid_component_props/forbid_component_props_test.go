package forbid_component_props

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestForbidComponentPropsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ForbidComponentPropsRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		// DOM intrinsic element with default forbid (className, style).
		{
			Code: `
        var First = createReactClass({
          render: function() {
            return <div className="foo" />;
          }
        });
      `,
			Tsx: true,
		},
		// DOM intrinsic with explicit forbid omitting className.
		{
			Code: `
        var First = createReactClass({
          render: function() {
            return <div style={{color: "red"}} />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"style"}},
		},
		// Component with non-forbidden prop.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo bar="baz" />;
          }
        });
      `,
			Tsx: true,
		},
		// Component with className when only style is forbidden.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"style"}},
		},
		// Multiple non-matching forbid entries.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"style", "foo"}},
		},
		// `<this.Foo>` with default forbid: prop "bar" not in forbid → no report.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <this.Foo bar="baz" />;
          }
        });
      `,
			Tsx: true,
		},
		// `<this.foo>` (lowercase rightmost) — treated as DOM by upstream's
		// componentName check, so the className prop is not checked.
		{
			Code: `
        class First extends createReactClass {
          render() {
            return <this.foo className="bar" />;
          }
        }
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"style"}},
		},
		// JSX spread on `<this.Foo>` — no JSXAttribute to check (spread is
		// JsxSpreadAttribute) so default forbid doesn't apply.
		{
			Code: `
        const First = (props) => (
          <this.Foo {...props} />
        );
      `,
			Tsx: true,
		},
		// allowedFor matches: ReactModal allowed for className.
		{
			Code: `
        const item = (<ReactModal className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":   "className",
					"allowedFor": []interface{}{"ReactModal"},
				},
			}},
		},
		// allowedFor matches a member-expression tag literally.
		{
			Code: `
        const item = (<MyLayout.Content className="customFoo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":   "className",
					"allowedFor": []interface{}{"MyLayout.Content"},
				},
			}},
		},
		// allowedFor matches "this.<X>".
		{
			Code: `
        const item = (<this.ReactModal className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":   "className",
					"allowedFor": []interface{}{"this.ReactModal"},
				},
			}},
		},
		// disallowedFor doesn't match Foo, so className stays allowed.
		{
			Code: `
        const item = (<Foo className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"ReactModal"},
				},
			}},
		},
		// JSX namespaced name (`<fbt:param>`) — `name` and `number` aren't
		// in default forbid, so no report. This locks in the namespaced-tag
		// path of the listener.
		{
			Code: `
        <fbt:param name="Total number of files" number={true} />
      `,
			Tsx: true,
		},
		// Combined disallowedFor across two entries; matched tag is excluded.
		{
			Code: `
        const item = (
          <Foo className="bar">
            <ReactModal style={{color: "red"}} />
          </Foo>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"OtherModal", "ReactModal"},
				},
				map[string]interface{}{
					"propName":      "style",
					"disallowedFor": []interface{}{"Foo"},
				},
			}},
		},
		// disallowedFor on one entry plus allowedFor on the other.
		{
			Code: `
        const item = (
          <Foo className="bar">
            <ReactModal style={{color: "red"}} />
          </Foo>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"OtherModal", "ReactModal"},
				},
				map[string]interface{}{
					"propName":   "style",
					"allowedFor": []interface{}{"ReactModal"},
				},
			}},
		},
		// disallowedFor entry doesn't match `this.ReactModal` (different tag).
		{
			Code: `
        const item = (<this.ReactModal className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"ReactModal"},
				},
			}},
		},
		// propNamePattern with allowedFor on a DOM intrinsic — `<div>` is DOM
		// and skipped before the pattern check runs.
		{
			Code: `
        const MyComponent = () => (
          <div aria-label="welcome" />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propNamePattern": "**-**",
					"allowedFor":      []interface{}{"div"},
				},
			}},
		},
		// allowedForPatterns: glob matches across several Components.
		{
			Code: `
        const rootElement = (
          <Root>
            <SomeIcon className="size-lg" />
            <AnotherIcon className="size-lg" />
            <SomeSvg className="size-lg" />
            <UICard className="size-lg" />
            <UIButton className="size-lg" />
          </Root>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":           "className",
					"allowedForPatterns": []interface{}{"*Icon", "*Svg", "UI*"},
				},
			}},
		},
		// allowedFor + allowedForPatterns combined: legacy explicit name plus pattern.
		{
			Code: `
        const rootElement = (
          <Root>
            <SomeIcon className="size-lg" />
            <AnotherIcon className="size-lg" />
            <SomeSvg className="size-lg" />
            <UICard className="size-lg" />
            <UIButton className="size-lg" />
            <ButtonLegacy className="size-lg" />
          </Root>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":           "className",
					"allowedFor":         []interface{}{"ButtonLegacy"},
					"allowedForPatterns": []interface{}{"*Icon", "*Svg", "UI*"},
				},
			}},
		},
		// disallowedFor + disallowedForPatterns: tags not matching either are allowed.
		{
			Code: `
        const rootElement = (
          <Root>
            <SomeIcon className="size-lg" />
            <AnotherIcon className="size-lg" />
            <SomeSvg className="size-lg" />
            <UICard className="size-lg" />
            <UIButton className="size-lg" />
          </Root>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":              "className",
					"disallowedFor":         []interface{}{"Modal"},
					"disallowedForPatterns": []interface{}{"*Legacy", "Shared*"},
				},
			}},
		},

		// ---- Additional edge cases (Dimension 4 universal edge shapes) ----
		// `<Foo.bar>` — rightmost lowercase, treated as DOM. tag = "Foo.bar".
		{Code: `<Foo.bar className="x" />;`, Tsx: true},
		// Bare `<this />` — KindThisKeyword tag; `componentName = "this"` is
		// lowercase, so the DOM-skip path fires and the prop is not checked
		// (matches upstream where `parentName.name === "this"` and
		// `"this"[0] === "t"` is lowercase).
		{Code: `<this bar="x" />;`, Tsx: true},
		// Bare `<this className="x" />` — same DOM-skip path. Even though
		// `className` is in default forbid, the lowercase componentName
		// short-circuits the rule.
		{Code: `<this className="x" />;`, Tsx: true},
		// `<a.B>` — rightmost uppercase, NOT DOM. tag = "a.B"; allowedFor
		// matches by exact string.
		{
			Code: `const x = (<a.B className="x" />);`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":   "className",
					"allowedFor": []interface{}{"a.B"},
				},
			}},
		},
		// Empty forbid list: nothing is forbidden, even className/style.
		{
			Code:    `<Foo className="x" style={{}} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{}},
		},
		// Spread attributes are not JsxAttributes — never reported.
		{
			Code: `<Foo {...props} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "className"},
			}},
		},
		// Spread + named attribute coexisting: spread is ignored, named is checked.
		{
			Code: `<Foo {...props} bar="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "className"},
			}},
		},
		// Boolean shorthand prop: `<Foo disabled />` — initializer absent. The
		// listener still receives the JsxAttribute and can match by name.
		{Code: `<Foo bar />;`, Tsx: true},
		// JSX Fragment containing Components: each Component is checked
		// independently; tags inside fragments are not affected by parent shape.
		{
			Code: `<><Foo bar="x" /><Bar baz="y" /></>;`,
			Tsx:  true,
		},
		// Empty string entries in forbid are skipped; a real entry alongside still applies.
		{
			Code:    `<Foo other="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"", "className"}},
		},
		// `forbid` not an array (schema-violating): falls back to defaults
		// `['className', 'style']`. Here `other` isn't in defaults → valid.
		{
			Code:    `<Foo other="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": "className"},
		},
		// Nested Components: outer/inner DOM and outer/inner Components mixed.
		// `<a>` and `<span>` are DOM — their props aren't checked even if forbidden.
		{
			Code: `
        <a className="outer">
          <Foo bar="ok">
            <span className="ok" />
          </Foo>
        </a>
      `,
			Tsx: true,
		},
		// Custom message with empty string falls back to default messageId
		// path — mirrors the truthy check in upstream `customMessage || ...`.
		{
			Code: `<MyComp other="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName": "className",
					"message":  "",
				},
			}},
		},
		// Non-string forbid entries (numbers, nulls, nested arrays) are ignored.
		// Defaults are NOT re-applied because `forbid` is present (a non-nil
		// array). Nothing is forbidden.
		{
			Code:    `<Foo className="x" style={{}} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{42, nil, []interface{}{"className"}}},
		},

		// ---- TS-specific syntax & wrappers (Dimension 1) ----
		// TS generic on a component tag: `<Foo<T>>` — tagName is still
		// Identifier "Foo"; type arguments don't affect tag string.
		{
			Code: `function W<T>() { return <Foo<T> bar="x" />; }`,
			Tsx:  true,
		},
		// TS non-null on the JSX child expression doesn't affect the
		// JSXAttribute listener.
		{
			Code: `const F: any = null; <a>{F!}</a>;`,
			Tsx:  true,
		},
		// TS `as` cast inside a JSX expression initializer — value side, not
		// inspected by the rule. Tag is DOM (`a`).
		{
			Code: `<a href={("/x" as string)} />;`,
			Tsx:  true,
		},

		// ---- Whitespace & comment robustness (Dimension 4) ----
		// Tag name immediately followed by self-close: `<Foo/>` (no space).
		{Code: `<Foo bar="x"/>;`, Tsx: true},
		// Multi-line attribute on a Component: only the prop name matters.
		{
			Code: `
        <Foo
          bar="multi-line"
        />
      `,
			Tsx: true,
		},
		// Block comment between tag and attribute — should not break parsing.
		{
			Code: `<Foo /* hi */ bar="x" />;`,
			Tsx:  true,
		},

		// ---- Defaults & options edge shapes ----
		// `null` options → defaults apply but bar is not in defaults.
		{Code: `<Foo bar="x" />;`, Tsx: true, Options: nil},
		// Options array shape `[opts]` — exercises GetOptionsMap unwrap path.
		{
			Code:    `<Foo bar="x" />;`,
			Tsx:     true,
			Options: []interface{}{map[string]interface{}{"forbid": []interface{}{"className"}}},
		},
		// Options object shape `opts` — single-option CLI path.
		{
			Code:    `<Foo bar="x" />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"className"}},
		},
		// Object entry with neither propName nor propNamePattern is silently
		// skipped (upstream produces an entry keyed by undefined). Bar is
		// then unrestricted.
		{
			Code: `<Foo bar="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"allowedFor": []interface{}{"X"}},
			}},
		},
		// `allowedFor` non-array (schema-violating) is silently ignored —
		// fall back to "no allow list, no allow patterns" → defaults forbid
		// applies. Here `bar` isn't forbidden so still valid.
		{
			Code: `<Foo bar="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "className", "allowedFor": "ReactModal"},
			}},
		},

		// ---- Real-world allow patterns (designed for typical app configs) ----
		// 1) Strict whitelist: only `Theme.*` may use `style`.
		{
			Code: `<Theme.Box style={{padding: 8}} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":           "style",
					"allowedForPatterns": []interface{}{"Theme.*"},
				},
			}},
		},
		// 2) Mixed glob — both `Card`-suffixed and `Modal`-suffixed allowed.
		{
			Code: `
        <UserCard className="foo" />;
        <UserModal className="bar" />;
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":           "className",
					"allowedForPatterns": []interface{}{"*Card", "*Modal"},
				},
			}},
		},
		// 3) Combined disallow + custom message — none of the listed
		// components present; quietly valid.
		{
			Code: `<HappyButton onClick={() => {}} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":              "onClick",
					"disallowedFor":         []interface{}{"DangerousButton"},
					"disallowedForPatterns": []interface{}{"Legacy*"},
					"message":               "Use IconButton#onPress",
				},
			}},
		},
		// 4) Multiple unrelated forbid entries; matched-rule routing.
		{
			Code: `<Foo onClick={() => {}} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "className"},
				map[string]interface{}{"propName": "style"},
				map[string]interface{}{"propName": "onClick", "allowedFor": []interface{}{"Foo"}},
			}},
		},
		// 5) JSX inside `key={...}` expression is checked separately as a
		// nested element — the inner `<Inner>` renders inside an attribute
		// expression, but its own props are still checked.
		{
			Code: `<div data-x={(<Inner bar="x" />)} />;`,
			Tsx:  true,
		},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		// Default forbid catches className on a Component.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 5, Column: 25},
			},
		},
		// Default forbid catches style on a Component.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo style={{color: "red"}} />;
          }
        });
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 5, Column: 25},
			},
		},
		// Explicit string forbid list catches className.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo className="bar" />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"className", "style"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 5, Column: 25},
			},
		},
		// Explicit string forbid list catches style.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo style={{color: "red"}} />;
          }
        });
      `,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"className", "style"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 5, Column: 25},
			},
		},
		// disallowedFor matches the Component tag.
		{
			Code: `
        var First = createReactClass({
          propTypes: externalPropTypes,
          render: function() {
            return <Foo style={{color: "red"}} />;
          }
        });
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "style",
					"disallowedFor": []interface{}{"Foo"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 5, Column: 25},
			},
		},
		// allowedFor not matching → forbidden.
		{
			Code: `
        const item = (<Foo className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":   "className",
					"allowedFor": []interface{}{"ReactModal"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 2, Column: 28},
			},
		},
		// `<this.ReactModal>` with allowedFor matching only `ReactModal` — tag
		// is "this.ReactModal", not in allowList, forbidden.
		{
			Code: `
        const item = (<this.ReactModal className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":   "className",
					"allowedFor": []interface{}{"ReactModal"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 2, Column: 40},
			},
		},
		// `<this.ReactModal>` with disallowedFor including the dotted name.
		{
			Code: `
        const item = (<this.ReactModal className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"this.ReactModal"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 2, Column: 40},
			},
		},
		// disallowedFor exact match on a single-name Component.
		{
			Code: `
        const item = (<ReactModal className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"ReactModal"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 2, Column: 35},
			},
		},
		// disallowedFor on a member-expression tag.
		{
			Code: `
        const item = (<MyLayout.Content className="customFoo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"MyLayout.Content"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 2, Column: 41},
			},
		},
		// Custom message replaces the default; no messageId is emitted.
		{
			Code: `
        const item = (<Foo className="foo" />);
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName": "className",
					"message":  "Please use ourCoolClassName instead of ClassName",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Please use ourCoolClassName instead of ClassName", Line: 2, Column: 28},
			},
		},
		// Multiple custom messages, in source order across nested elements.
		{
			Code: `
        const item = () => (
          <Foo className="foo">
            <Bar option="high" />
          </Foo>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName": "className",
					"message":  "Please use ourCoolClassName instead of ClassName",
				},
				map[string]interface{}{
					"propName": "option",
					"message":  "Avoid using option",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Please use ourCoolClassName instead of ClassName", Line: 3, Column: 16},
				{Message: "Avoid using option", Line: 4, Column: 18},
			},
		},
		// Mix of default-message and custom-message entries.
		{
			Code: `
        const item = () => (
          <Foo className="foo">
            <Bar option="high" />
          </Foo>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propName": "className"},
				map[string]interface{}{
					"propName": "option",
					"message":  "Avoid using option",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 16},
				{Message: "Avoid using option", Line: 4, Column: 18},
			},
		},
		// propNamePattern matches a kebab-case prop on a Component.
		{
			Code: `
        const MyComponent = () => (
          <Foo kebab-case-prop={123} />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{"propNamePattern": "**-**"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 16},
			},
		},
		// propNamePattern + custom message.
		{
			Code: `
        const MyComponent = () => (
          <Foo kebab-case-prop={123} />
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propNamePattern": "**-**",
					"message":         "Avoid using kebab-case",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Avoid using kebab-case", Line: 3, Column: 16},
			},
		},
		// propNamePattern + allowedFor: DOM intrinsic skipped, Component flagged.
		{
			Code: `
        const MyComponent = () => (
          <div>
            <div aria-label="Hello World" />
            <Foo kebab-case-prop={123} />
          </div>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propNamePattern": "**-**",
					"allowedFor":      []interface{}{"div"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 5, Column: 18},
			},
		},
		// propNamePattern + disallowedFor: DOM intrinsics skipped (h1, div).
		{
			Code: `
        const MyComponent = () => (
          <div>
            <div aria-label="Hello World" />
            <h1 data-id="my-heading" />
            <Foo kebab-case-prop={123} />
          </div>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propNamePattern": "**-**",
					"disallowedFor":   []interface{}{"Foo"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 6, Column: 18},
			},
		},
		// allowedForPatterns + custom message.
		{
			Code: `
        const rootElement = () => (
          <Root>
            <SomeIcon className="size-lg" />
            <SomeSvg className="size-lg" />
          </Root>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":           "className",
					"message":            "className available only for icons",
					"allowedForPatterns": []interface{}{"*Icon"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "className available only for icons", Line: 5, Column: 22},
			},
		},
		// Two separate entries, each with allowedForPatterns + custom message.
		{
			Code: `
        const rootElement = () => (
          <Root>
            <UICard style={{backgroundColor: black}}/>
            <SomeIcon className="size-lg" />
            <SomeSvg className="size-lg" style={{fill: currentColor}} />
          </Root>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":           "className",
					"message":            "className available only for icons",
					"allowedForPatterns": []interface{}{"*Icon"},
				},
				map[string]interface{}{
					"propName":           "style",
					"message":            "style available only for SVGs",
					"allowedForPatterns": []interface{}{"*Svg"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "style available only for SVGs", Line: 4, Column: 21},
				{Message: "className available only for icons", Line: 6, Column: 22},
			},
		},
		// disallowedFor + disallowedForPatterns combined; multi-component flagging.
		{
			Code: `
        const rootElement = (
          <Root>
            <SomeIcon className="size-lg" />
            <AnotherIcon className="size-lg" />
            <SomeSvg className="size-lg" />
            <UICard className="size-lg" />
            <ButtonLegacy className="size-lg" />
          </Root>
        );
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":              "className",
					"disallowedFor":         []interface{}{"SomeSvg"},
					"disallowedForPatterns": []interface{}{"UI*", "*Icon"},
					"message":               "Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns", Line: 4, Column: 23},
				{Message: "Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns", Line: 5, Column: 26},
				{Message: "Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns", Line: 6, Column: 22},
				{Message: "Avoid using className for SomeSvg and components that match the `UI*` and `*Icon` patterns", Line: 7, Column: 21},
			},
		},

		// ---- Additional lock-in cases ----
		// `<Foo.Bar.Baz>` deep-nested member access: upstream produces
		// "undefined.Baz" as the tag string. Our port mirrors that quirk so
		// allow/disallow lists match upstream byte-for-byte. This locks in
		// the behavior — a future refactor that walks the chain cleanly
		// would change the resulting tag string and should explicitly
		// flip this test.
		{
			Code: `<Foo.Bar.Baz className="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "className",
					"disallowedFor": []interface{}{"undefined.Baz"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 14},
			},
		},
		// Options passed as bare object (single-option CLI shape) — exercises
		// `utils.GetOptionsMap`'s array-vs-map handling.
		{
			Code:    `<Foo style={{}} />;`,
			Tsx:     true,
			Options: map[string]interface{}{"forbid": []interface{}{"style"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// JSX namespaced name with a forbidden prop. Upstream: namespaced
		// tags bypass the `componentName[0]` lowercase check (because
		// componentName is a JSXIdentifier object, not a string). We mirror
		// that — `<fbt:param>` is treated as a Component, so `className` is
		// reported. Locks in the namespaced-tag path of the listener.
		{
			Code: `<fbt:param className="x" />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 12},
			},
		},
		// Same namespaced tag with explicit allowedFor matching `ns:name`.
		// Confirms our synthetic tag-string format `"fbt:param"` is what
		// users would naturally write in their config.
		{
			Code: `<fbt:Foo style={{a: 1}} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":   "style",
					"allowedFor": []interface{}{"fbt:notFoo"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 10},
			},
		},
		// Multi-level Component nesting: every JsxAttribute on a Component
		// is checked, regardless of how deeply nested. DOM `<div>` is skipped.
		{
			Code: `
        const tree = (
          <Outer className="o1">
            <div className="d1">
              <Inner className="i1">
                <Leaf className="l1" />
              </Inner>
            </div>
          </Outer>
        );
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 18},
				{MessageId: "propIsForbidden", Line: 5, Column: 22},
				{MessageId: "propIsForbidden", Line: 6, Column: 23},
			},
		},
		// Multiple JsxAttributes on the same Component, mixed with spread:
		// each forbidden named attr is reported independently.
		{
			Code: `<Foo {...rest} className="x" style={{}} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 16},
				{MessageId: "propIsForbidden", Line: 1, Column: 30},
			},
		},
		// Direct propName lookup wins over a propNamePattern that would also
		// match — locks in upstream `forbid.get(prop) || patternMatch` order.
		// Pattern entry has a custom message; direct entry has none, so the
		// reported message is the default.
		{
			Code: `<Foo className="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propNamePattern": "*Name",
					"message":         "Pattern wins (should NOT see this)",
				},
				map[string]interface{}{"propName": "className"},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// Same prop hits TWO pattern entries; first-matched-in-insertion-order
		// wins (upstream `Array.find`). The first entry has a custom message.
		{
			Code: `<Foo data-id="x" />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propNamePattern": "data-*",
					"message":         "First pattern",
				},
				map[string]interface{}{
					"propNamePattern": "*-id",
					"message":         "Second pattern (should NOT see this)",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "First pattern", Line: 1, Column: 6},
			},
		},
		// JSX expression initializer (no string literal): the rule only looks
		// at the attribute name, never the value, so this is reported.
		{
			Code: `<Foo style={someStyle} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// `<Foo.Bar>` 2-segment member with disallowedFor — locks in the
		// canonical `"Foo.Bar"` tag string.
		{
			Code: `<Foo.Bar style={{}} />;`,
			Tsx:  true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":      "style",
					"disallowedFor": []interface{}{"Foo.Bar"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 10},
			},
		},

		// ---- Reporting position robustness (Go-vs-ESLint ECMA columns) ----
		// JSXOpening (`<Foo>`) variant — the listener fires on the attribute
		// inside the opening element, not on the closing tag.
		{
			Code: `<Foo className="x">child text</Foo>;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// Indented multi-line attribute — column counts from line start of
		// the JsxAttribute, not the opening tag.
		{
			Code: `
        <Foo
          className="x"
        />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 11},
			},
		},
		// Tab-indented attribute — UTF-16 column counting (rule_tester uses
		// ECMA spec line/column, tabs count as 1).
		{
			Code: "<Foo\n\tclassName=\"x\"\n/>;",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 2, Column: 2},
			},
		},
		// Multi-byte (CJK) child content — column on the attribute line is
		// not affected by the multi-byte payload elsewhere in the file.
		{
			Code: `
        const t = "中文";
        <Foo className={t} />;
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 14},
			},
		},

		// ---- TS-specific syntax on the Component side ----
		// TS generic on a Component tag combined with forbidden prop.
		{
			Code: `function W<T>() { return <Foo<T> className="x" />; }`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 34},
			},
		},
		// JSX expression value with TS `as` cast — the rule still flags
		// the attribute by name.
		{
			Code: `<Foo style={({} as React.CSSProperties)} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 6},
			},
		},
		// JSX expression value with TS non-null `!` — same.
		{
			Code: `const s: any = null; <Foo style={s!} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 27},
			},
		},

		// ---- Real-world configurations ----
		// Design-system enforcement: forbid `className` everywhere except
		// utility primitives matching `Box`/`Stack`/`Inline`.
		{
			Code: `
        <Card className="bad" />;
        <Box className="ok" />;
        <Stack className="ok" />;
        <Inline className="ok" />;
        <CustomThing className="bad" />;
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propName":           "className",
					"allowedForPatterns": []interface{}{"Box", "Stack", "Inline"},
					"message":            "Use design-system primitives for className.",
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{Message: "Use design-system primitives for className.", Line: 2, Column: 15},
				{Message: "Use design-system primitives for className.", Line: 6, Column: 22},
			},
		},
		// `data-*` ban on Components but allowed on `<TestHarness>`.
		{
			Code: `
        <TestHarness data-x="ok" />;
        <Widget data-y="bad" />;
      `,
			Tsx: true,
			Options: map[string]interface{}{"forbid": []interface{}{
				map[string]interface{}{
					"propNamePattern": "data-*",
					"allowedFor":      []interface{}{"TestHarness"},
				},
			}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 17},
			},
		},
		// Many siblings, alternating Component vs DOM, single forbid rule;
		// only Component children are reported.
		{
			Code: `
        <Group>
          <Foo className="a" />
          <span className="b" />
          <Bar className="c" />
          <i className="d" />
          <Baz className="e" />
        </Group>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 3, Column: 16},
				{MessageId: "propIsForbidden", Line: 5, Column: 16},
				{MessageId: "propIsForbidden", Line: 7, Column: 16},
			},
		},
		// JSX child elements (`<Foo>` containing nested children incl. another
		// Component with the same forbidden prop).
		{
			Code: `
        <Foo className="outer">
          <Bar className="inner" />
        </Foo>
      `,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 2, Column: 14},
				{MessageId: "propIsForbidden", Line: 3, Column: 16},
			},
		},
		// Component inside a parent attribute expression: rule fires on
		// `<Inner className=...>` even though it's nested in a JSX expression.
		{
			Code: `<div data-x={(<Inner className="x" />)} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 22},
			},
		},
		// Conditional rendering with both branches producing Components.
		{
			Code: `const cond = true; const x = cond ? <Foo style={{}} /> : <Bar style={{}} />;`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 42},
				{MessageId: "propIsForbidden", Line: 1, Column: 63},
			},
		},
		// Mapping over a list — JSX inside an arrow body still triggers.
		{
			Code: `const xs: any[] = []; xs.map(x => <Item key={x} className="row" />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "propIsForbidden", Line: 1, Column: 49},
			},
		},
	})
}
