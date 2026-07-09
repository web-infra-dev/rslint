package void_use_memo

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestVoidUseMemoUpstream migrates the react-hooks/void-use-memo cases from
// upstream compiler/packages/babel-plugin-react-compiler/src/Validation/ValidateUseMemo.ts
// and related React Compiler fixtures 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// void_use_memo_extras_test.go.
func TestVoidUseMemoUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// ---- Upstream fixture: useMemo-arrow-implicit-return.js. ----
		{Code: `function Component() {
  const value = useMemo(() => computeValue(), []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-empty-return.js; explicit return is non-void for this rule. ----
		{Code: `function Component() {
  const value = useMemo(() => {
    return;
  }, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-explicit-null-return.js. ----
		{Code: `function Component() {
  const value = useMemo(() => {
    return null;
  }, []);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-multiple-returns.js. ----
		{Code: `function Component({items}) {
  const value = useMemo(() => {
    for (let item of items) {
      if (item.match) return item;
    }
    return null;
  }, [items]);
  return <div>{value}</div>;
}
`, Tsx: true},
		// ---- Upstream fixture: useMemo-switch-return.js. ----
		{Code: `function Component(props) {
  const x = useMemo(() => {
    let y;
    switch (props.switch) {
      case 'foo': {
        return 'foo';
      }
      case 'bar': {
        y = 'bar';
        break;
      }
      default: {
        y = props.y;
      }
    }
    return y;
  });
  return x;
}
`, Tsx: true},
		// ---- Upstream fixture: repro.js; skipped because the source disables validateNoVoidUseMemo via React Compiler test metadata. ----
		{Code: `function Component(props) {
  const item = props.item;
  const thumbnails = [];
  const baseVideos = getBaseVideos(item);
  useMemo(() => {
    baseVideos.forEach(video => {
      const baseVideo = video.hasBaseVideo;
      if (Boolean(baseVideo)) {
        thumbnails.push({extraVideo: true});
      }
    });
  });
  return <FlatList baseVideos={baseVideos} items={thumbnails} />;
}
`, Tsx: true, Skip: true},
	}

	invalid := []rule_tester.InvalidTestCase{
		// ---- Upstream fixture: invalid-useMemo-no-return-value.js. ----
		{
			Code: `function Component() {
  const value = useMemo(() => {
    console.log('computing');
  }, []);
  const value2 = React.useMemo(() => {
    console.log('computing');
  }, []);
  return (
    <div>
      {value}
      {value2}
    </div>
  );
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 2, 25, 4, 4),
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 5, 32, 7, 4),
			},
		},
		// ---- Upstream fixture: invalid-unused-use-memo.js. ----
		{
			Code: `function Component() {
  useMemo(() => {
    return [];
  }, []);
  return <div />;
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("unusedResult", unusedResultReason, unusedResultDescription, 2, 3, 2, 10),
			},
		},
		// ---- Upstream fixture: invalid-useMemo-return-empty.js. ----
		{
			Code: `function component(a) {
  let x = useMemo(() => {
    mutate(a);
  }, []);
  return x;
}
`,
			Errors: []rule_tester.InvalidTestCaseError{
				voidUseMemoError("missingReturn", missingReturnReason, missingReturnDescription, 2, 19, 4, 4),
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &VoidUseMemoRule, valid, invalid)
}

func voidUseMemoError(id, reason, description string, line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: id,
		Message:   buildVoidUseMemoMessage(id, reason, description).Description,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}
}
