package jsx_closing_tag_location

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxClosingTagLocationRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxClosingTagLocationRule, []rule_tester.ValidTestCase{
		{
			Code: `var x = <div>foo</div>`,
			Tsx:  true,
		},
		{
			Code: `var x = (
<div>
  foo
</div>
)`,
			Tsx: true,
		},
		{
			Code: `var x = (
  <div>
    <span>foo</span>
  </div>
)`,
			Tsx: true,
		},
		{
			Code: `var x = <></>`,
			Tsx:  true,
		},
		{
			// "line-aligned" option: closing tag aligned with line start indent
			Code: `var x = (
  <div>
    foo
  </div>
)`,
			Tsx:     true,
			Options: "line-aligned",
		},
		{
			// Single-line element with content
			Code: `var x = <div>foo</div>`,
			Tsx:  true,
		},
		{
			// Fragment on single line
			Code: `var x = <>foo</>`,
			Tsx:  true,
		},
		{
			// Multiline fragment with proper alignment
			Code: `var x = (
<>
  foo
</>
)`,
			Tsx: true,
		},
		{
			// Nested elements with proper alignment
			Code: `var x = (
  <div>
    <span>
      foo
    </span>
  </div>
)`,
			Tsx: true,
		},
	}, []rule_tester.InvalidTestCase{
		{
			// "line-aligned" option: closing tag misaligned with line indent
			Code: `var x = (
  <div>
    foo
    </div>
)`,
			Tsx:     true,
			Options: "line-aligned",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alignWithOpening",
					Line:      4,
					Column:    5,
				},
			},
		},
		{
			Code: `var x = (
<div>
  foo
  </div>
)`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "matchIndent",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			Code: `var x = (
  <div>
    foo
      </div>
)`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "matchIndent",
					Line:      4,
					Column:    7,
				},
			},
		},
		{
			// Fragment closing tag misaligned
			Code: `var x = (
<>
  foo
  </>
)`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "matchIndent",
					Line:      4,
					Column:    3,
				},
			},
		},
		{
			// "line-aligned": closing tag over-indented
			Code: `var x = (
  <div>
    foo
      </div>
)`,
			Tsx:     true,
			Options: "line-aligned",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "alignWithOpening",
					Line:      4,
					Column:    7,
				},
			},
		},
		{
			// Closing tag on same line as content (must be on own line)
			Code: `var x = (
<App>
  foo</App>
)`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onOwnLine",
					Line:      3,
					Column:    6,
				},
			},
		},
		{
			// Fragment closing tag on same line as content
			Code: `var x = (
<>
  foo</>
)`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "onOwnLine",
					Line:      3,
					Column:    6,
				},
			},
		},
	})
}
