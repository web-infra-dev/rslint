package jsx_key

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxKeyRule(t *testing.T) {
	checkFragmentShorthand := map[string]interface{}{"checkFragmentShorthand": true}
	checkKeyMustBeforeSpread := map[string]interface{}{"checkKeyMustBeforeSpread": true}
	warnOnDuplicates := map[string]interface{}{"warnOnDuplicates": true}
	pragmaSettings := map[string]interface{}{
		"react": map[string]interface{}{"pragma": "Act", "fragment": "Frag"},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxKeyRule, []rule_tester.ValidTestCase{
		// ---- Upstream valid cases ----
		{
			Code: `
[1, 2, 3].map((item) => {
  return item === 'bar' ? <div key={item}>{item}</div> : <span key={item}>{item}</span>;
})
`,
			Tsx: true,
		},
		{Code: `fn()`, Tsx: true},
		{Code: `[1, 2, 3].map(function () {})`, Tsx: true},
		{Code: `<App />;`, Tsx: true},
		{Code: `[<App key={0} />, <App key={1} />];`, Tsx: true},
		{Code: `[1, 2, 3].map(function(x) { return <App key={x} /> });`, Tsx: true},
		{Code: `[1, 2, 3].map(x => <App key={x} />);`, Tsx: true},
		{Code: `[1, 2 ,3].map(x => x && <App x={x} key={x} />);`, Tsx: true},
		{Code: `[1, 2 ,3].map(x => x ? <App x={x} key="1" /> : <OtherApp x={x} key="2" />);`, Tsx: true},
		{Code: `[1, 2, 3].map(x => { return <App key={x} /> });`, Tsx: true},
		{Code: `Array.from([1, 2, 3], function(x) { return <App key={x} /> });`, Tsx: true},
		{Code: `Array.from([1, 2, 3], (x => <App key={x} />));`, Tsx: true},
		{Code: `Array.from([1, 2, 3], (x => {return <App key={x} />}));`, Tsx: true},
		{Code: `Array.from([1, 2, 3], someFn);`, Tsx: true},
		{Code: `Array.from([1, 2, 3]);`, Tsx: true},
		{Code: `[1, 2, 3].foo(x => <App />);`, Tsx: true},
		{Code: `var App = () => <div />;`, Tsx: true},
		{Code: `[1, 2, 3].map(function(x) { return; });`, Tsx: true},
		{Code: `foo(() => <div />);`, Tsx: true},
		{Code: `foo(() => <></>);`, Tsx: true},
		{Code: `<></>;`, Tsx: true},
		{Code: `<App {...{}} />;`, Tsx: true},
		{
			Code:    `<App key="keyBeforeSpread" {...{}} />;`,
			Tsx:     true,
			Options: checkKeyMustBeforeSpread,
		},
		{
			Code:    `<div key="keyBeforeSpread" {...{}} />;`,
			Tsx:     true,
			Options: checkKeyMustBeforeSpread,
		},
		{
			Code: `
const spans = [
  <span key="notunique"/>,
  <span key="notunique"/>,
];
`,
			Tsx: true,
		},
		{
			Code: `
function Component(props) {
  return hasPayment ? (
    <div className="stuff">
      <BookingDetailSomething {...props} />
      {props.modal && props.calculatedPrice && (
        <SomeOtherThing items={props.something} discount={props.discount} />
      )}
    </div>
  ) : null;
}
`,
			Tsx: true,
		},
		{
			Code: `
// testrule.jsx
const trackLink = () => {};
const getAnalyticsUiElement = () => {};

const onTextButtonClick = (e, item) => trackLink([, getAnalyticsUiElement(item), item.name], e);
`,
			Tsx: true,
		},
		{
			Code: "\nfunction Component({ allRatings }) {\n  return (\n    <RatingDetailsStyles>\n      {Object.entries(allRatings)?.map(([key, value], index) => {\n        const rate = value?.split(/(?=[%, /])/);\n\n        if (!rate) return null;\n\n        return (\n          <li key={`${entertainment.tmdbId}${index}`}>\n            <img src={`/assets/rating/${key}.png`} />\n            <span className=\"rating-details--rate\">{rate?.[0]}</span>\n            <span className=\"rating-details--rate-suffix\">{rate?.[1]}</span>\n          </li>\n        );\n      })}\n    </RatingDetailsStyles>\n  );\n}\n",
			Tsx:  true,
		},
		{
			Code: "\nconst baz = foo?.bar?.()?.[1] ?? 'qux';\n\nqux()?.map()\n\nconst directiveRanges = comments?.map(tryParseTSDirective)\n",
			Tsx:  true,
		},
		// ---- React.Children.toArray / destructured Children.toArray ----
		{Code: `React.Children.toArray([1, 2, 3].map(x => <App />));`, Tsx: true},
		{
			Code: `
import { Children } from "react";
Children.toArray([1, 2, 3].map(x => <App />));
`,
			Tsx: true,
		},
		{
			// Pragma setting — Act.Children.toArray also counts.
			Code: `
import Act from 'react';
import { Children as ReactChildren } from 'react';

const { Children } = Act;
const { toArray } = Children;

Act.Children.toArray([1, 2, 3].map(x => <App />));
Act.Children.toArray(Array.from([1, 2, 3], x => <App />));
Children.toArray([1, 2, 3].map(x => <App />));
Children.toArray(Array.from([1, 2, 3], x => <App />));
`,
			Tsx:      true,
			Settings: pragmaSettings,
		},
		{Code: `[1, 2, 3].map(x => { return x && <App key={x} />; });`, Tsx: true},
		{Code: `[1, 2, 3].map(x => { return x && y && <App key={x} />; });`, Tsx: true},
		{Code: `[1, 2, 3].map(x => { return x && foo(); });`, Tsx: true},

		// ---- Additional edge cases ----
		// Nested JSX: keys are not required on direct JSX children of a
		// JSX parent.
		{Code: `<ul><li/><li/></ul>;`, Tsx: true},
		// Array.from with no callback (non-function 2nd arg) — no report.
		{Code: `Array.from([1, 2, 3], "not a function");`, Tsx: true},
		// `xs?.map(...)` — optional chaining on the call. Same behavior as
		// `xs.map(...)`.
		{Code: `xs?.map(x => <App key={x} />);`, Tsx: true},
		// `xs.map?.(...)` — optional call. Same behavior.
		{Code: `xs.map?.(x => <App key={x} />);`, Tsx: true},
		// Parenthesized arrow callback.
		{Code: `[1,2,3].map((x => <App key={x} />));`, Tsx: true},
		// Spread element alongside keyed JSX in an array — spread is not a
		// JsxElement, only the keyed element is inspected.
		{Code: `const rest = []; [<App key="k" />, ...rest];`, Tsx: true},
		// Omitted array element (hole) is not a JsxElement.
		{Code: `[, <App key="k" />];`, Tsx: true},
		// warnOnDuplicates: distinct literal texts do not trigger.
		{
			Code: `
const spans = [
  <span key="a"/>,
  <span key="b"/>,
];
`,
			Tsx:     true,
			Options: warnOnDuplicates,
		},
		// checkKeyMustBeforeSpread false (default): key after spread in an
		// array is NOT flagged.
		{Code: `[<App {...obj} key="k" />, <App key="k2" />];`, Tsx: true},
		// checkFragmentShorthand false (default): fragment in iterator is
		// not flagged.
		{Code: `[1, 2, 3].map(x => <>{x}</>);`, Tsx: true},
		// checkFragmentShorthand false: fragment in array literal is not
		// flagged.
		{Code: `[<></>];`, Tsx: true},
		// Typed FC returning JSX — a real-world shape from upstream's test
		// suite. Should not report.
		{
			Code: `
import React, { FC, useRef, useState } from 'react';

type Props = {
  videoUrl: string;
  videoTitle: string;
};
const ResourceVideo: FC<Props> = ({
  videoUrl,
  videoTitle,
}: Props): JSX.Element => {
  return (
    <div className="resource-video">
      <VimeoVideoPlayInModal videoUrl={videoUrl} />
      <h3>{videoTitle}</h3>
    </div>
  );
};

export default ResourceVideo;
`,
			Tsx: true,
		},
		// mobx `observable.map(...)` is NOT Array.prototype.map — the callee
		// is `observable.map`, property name "map". Upstream actually flags
		// this if there's a JSX-returning arrow argument; here the single
		// argument is a type argument generic so no callback is present.
		{
			Code: `
import { observable } from "mobx";

export interface ClusterFrameInfo {
  frameId: number;
  processId: number;
}

export const clusterFrameMap = observable.map<string, ClusterFrameInfo>();
`,
			Tsx: true,
		},
		// Nested React.Children.toArray — both the inner Children.toArray
		// and the deepest map are skipped.
		{
			Code: `
React.Children.toArray(
  React.Children.toArray([1, 2, 3].map(x => <App />))
);
`,
			Tsx: true,
		},
		// Children.toArray balanced with an adjacent unguarded map: only
		// the unguarded one would normally flag, but it's keyed — no report.
		{
			Code: `
React.Children.toArray([1, 2, 3].map(x => <A />));
[1, 2, 3].map(x => <B key={x} />);
`,
			Tsx: true,
		},
		// Fragment appearing as a JSX child (not an array element) — even
		// with checkFragmentShorthand on, upstream only reports array-level
		// fragments. Locks that behavior.
		{
			Code:    `<div><></><></></div>;`,
			Tsx:     true,
			Options: checkFragmentShorthand,
		},
		// Map callback with async arrow — still an ArrowFunction, still
		// subject to the iterator check. Keyed → no report.
		{Code: `[1, 2, 3].map(async x => <App key={x} />);`, Tsx: true},
		// Deeply nested parens around the arrow callback.
		{Code: `[1, 2, 3].map((((x => <App key={x} />))));`, Tsx: true},
		// Bracket access `xs["map"]` is NOT matched by the rule (upstream
		// uses `callee.property.name="map"`, i.e. dot access only).
		{Code: `xs["map"](x => <App />);`, Tsx: true},
		// Chained calls: `xs.filter(...).map(x => <A key={x} />)` — keyed.
		{Code: `xs.filter(x => x > 0).map(x => <App key={x} />);`, Tsx: true},
		// Array.from with a non-function second argument (identifier/any).
		{Code: `Array.from([1, 2, 3], 123);`, Tsx: true},
		// Generic call — `.map<T>(...)` still a PropertyAccessExpression.
		{Code: `xs.map<number>(x => <App key={x} />);`, Tsx: true},
	}, []rule_tester.InvalidTestCase{
		// ---- Upstream invalid cases ----
		{
			Code: `
[1, 2, 3].map((item) => {
  return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
})`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
[1, 2, 3].map(function(item) {
  return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
})`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
Array.from([1, 2, 3], (item) => {
  return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
})`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
import { Fragment } from 'react';

const ITEMS = ['bar', 'foo'];

export default function BugIssue() {
  return (
    <Fragment>
      {ITEMS.map((item) => {
        return item === 'bar' ? <div>{item}</div> : <span>{item}</span>;
      })}
    </Fragment>
  );
}
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[<App />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArrayKey", Line: 1, Column: 2},
			},
		},
		{
			Code: `[<App {...key} />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArrayKey", Line: 1, Column: 2},
			},
		},
		{
			Code: `[<App key={0}/>, <App />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArrayKey", Line: 1, Column: 18},
			},
		},
		{
			Code: `[1, 2 ,3].map(function(x) { return <App /> });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2 ,3].map(x => <App />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2 ,3].map(x => x && <App x={x} />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2 ,3].map(x => x ? <App x={x} key="1" /> : <OtherApp x={x} />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2 ,3].map(x => x ? <App x={x} /> : <OtherApp x={x} key="2" />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2 ,3].map(x => { return <App /> });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `Array.from([1, 2 ,3], function(x) { return <App /> });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `Array.from([1, 2 ,3], (x => { return <App /> }));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `Array.from([1, 2 ,3], (x => <App />));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2, 3]?.map(x => <App />)`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code:     `[1, 2, 3].map(x => <>{x}</>);`,
			Tsx:      true,
			Options:  checkFragmentShorthand,
			Settings: pragmaSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingIterKeyUsePrag",
					Message:   `Missing "key" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use Act.Frag instead`,
				},
			},
		},
		{
			Code:     `[<></>];`,
			Tsx:      true,
			Options:  checkFragmentShorthand,
			Settings: pragmaSettings,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArrayKeyUsePrag",
					Message:   `Missing "key" prop for element in array. Shorthand fragment syntax does not support providing keys. Use Act.Frag instead`,
				},
			},
		},
		{
			Code:    `[<App {...obj} key="keyAfterSpread" />];`,
			Tsx:     true,
			Options: checkKeyMustBeforeSpread,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "keyBeforeSpread"},
			},
		},
		{
			Code:    `[<div {...obj} key="keyAfterSpread" />];`,
			Tsx:     true,
			Options: checkKeyMustBeforeSpread,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "keyBeforeSpread"},
			},
		},
		{
			Code: `
const spans = [
  <span key="notunique"/>,
  <span key="notunique"/>,
];
`,
			Tsx:     true,
			Options: warnOnDuplicates,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonUniqueKeys", Line: 3},
				{MessageId: "nonUniqueKeys", Line: 4},
			},
		},
		{
			Code: `
const div = (
  <div>
    <span key="notunique"/>
    <span key="notunique"/>
  </div>
);
`,
			Tsx:     true,
			Options: warnOnDuplicates,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonUniqueKeys", Line: 4},
				{MessageId: "nonUniqueKeys", Line: 5},
			},
		},
		{
			Code: `
const Test = () => {
  const list = [1, 2, 3, 4, 5];

  return (
    <div>
      {list.map(item => {
        if (item < 2) {
          return <div>{item}</div>;
        }

        return <div />;
      })}
    </div>
  );
};
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
const TestO = () => {
  const list = [1, 2, 3, 4, 5];

  return (
    <div>
      {list.map(item => {
        if (item < 2) {
          return <div>{item}</div>;
        } else if (item < 5) {
          return <div></div>
        }  else {
          return <div></div>
        }

        return <div />;
      })}
    </div>
  );
};
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
const TestCase = () => {
  const list = [1, 2, 3, 4, 5];

  return (
    <div>
      {list.map(item => {
        if (item < 2) return <div>{item}</div>;
        else if (item < 5) return <div />;
        else return <div />;
      })}
    </div>
  );
};
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `
const TestCase = () => {
  const list = [1, 2, 3, 4, 5];

  return (
    <div>
      {list.map(x => <div {...spread} key={x} />)}
    </div>
  );
};
`,
			Tsx:     true,
			Options: checkKeyMustBeforeSpread,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "keyBeforeSpread"},
			},
		},
		{
			Code: `[1, 2, 3].map(x => { return x && <App />; });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `[1, 2, 3].map(x => { return x || y || <App />; });`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},

		// ---- Additional edge cases ----
		// Line / column coverage on a simple array-missing-key case.
		{
			Code: "\n[\n  <App />,\n  <App />,\n];\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArrayKey", Line: 3, Column: 3},
				{MessageId: "missingArrayKey", Line: 4, Column: 3},
			},
		},
		// Full message-text assertion on a plain missingIterKey.
		{
			Code: `[1].map(x => <App />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingIterKey",
					Message:   `Missing "key" prop for element in iterator`,
					Line:      1,
					Column:    14,
				},
			},
		},
		// Full message-text assertion on missingArrayKey.
		{
			Code: `[<App />];`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArrayKey",
					Message:   `Missing "key" prop for element in array`,
					Line:      1,
					Column:    2,
				},
			},
		},
		// Default pragma — missingIterKeyUsePrag uses "React.Fragment".
		{
			Code:    `[1, 2, 3].map(x => <>{x}</>);`,
			Tsx:     true,
			Options: checkFragmentShorthand,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingIterKeyUsePrag",
					Message:   `Missing "key" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead`,
				},
			},
		},
		// Default pragma — missingArrayKeyUsePrag uses "React.Fragment".
		{
			Code:    `[<></>];`,
			Tsx:     true,
			Options: checkFragmentShorthand,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArrayKeyUsePrag",
					Message:   `Missing "key" prop for element in array. Shorthand fragment syntax does not support providing keys. Use React.Fragment instead`,
				},
			},
		},
		// Optional chaining on both the member access and the call.
		{
			Code: `[1, 2, 3].map?.(x => <App />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		{
			Code: `xs?.map(x => <App />);`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		// Parenthesized arrow callback inside Array.from — ESTree flattens
		// the parens; tsgo preserves them. Regression case.
		{
			Code: `Array.from([1, 2, 3], ((x => <App />)));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		// Parenthesized JSX body of arrow map callback.
		{
			Code: `[1,2,3].map(x => (<App />));`,
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey"},
			},
		},
		// warnOnDuplicates with checkKeyMustBeforeSpread both on —
		// independent diagnostics on the same element.
		{
			Code: `
const spans = [
  <span {...o} key="a"/>,
  <span key="a"/>,
];
`,
			Tsx: true,
			Options: map[string]interface{}{
				"warnOnDuplicates":         true,
				"checkKeyMustBeforeSpread": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "keyBeforeSpread"},
				{MessageId: "nonUniqueKeys", Line: 3},
				{MessageId: "nonUniqueKeys", Line: 4},
			},
		},
		// warnOnDuplicates groups keys by raw source text — `{'a'}` and
		// `"a"` are different texts, so they're NOT duplicates (matches
		// upstream's `getText` behavior).
		{
			Code: `
const spans = [
  <span key={'a'}/>,
  <span key="a"/>,
  <span key={'a'}/>,
];
`,
			Tsx:     true,
			Options: warnOnDuplicates,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "nonUniqueKeys", Line: 3},
				{MessageId: "nonUniqueKeys", Line: 5},
			},
		},
		// ---- Additional container / positioning coverage ----
		// Map callback missing key: assert precise Line/Column/EndLine/
		// EndColumn for the JSX element in a multi-line case.
		{
			Code: "\n[1, 2, 3].map(x =>\n  <App />\n);\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingIterKey",
					Line:      3,
					Column:    3,
					EndLine:   3,
					EndColumn: 10,
				},
			},
		},
		// Array-missing-key: precise end line/column on a multi-line element.
		{
			Code: "\n[\n  <App\n    x={1}\n  />,\n];\n",
			Tsx:  true,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "missingArrayKey",
					Line:      3,
					Column:    3,
					EndLine:   5,
					EndColumn: 5,
				},
			},
		},
		// keyBeforeSpread in array context reports on the array, not the
		// offending element.
		{
			Code:    "\n[\n  <App {...x} key=\"k\" />,\n];\n",
			Tsx:     true,
			Options: checkKeyMustBeforeSpread,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "keyBeforeSpread",
					Line:      2,
					Column:    1,
				},
			},
		},
		// Three-way independence: keys across sibling arrays don't collide.
		// The outer array has JSX children (the inner arrays) that are
		// themselves arrays, not JSX — no missing-array-key on the outer.
		// Each inner array evaluates its own keys independently.
		{
			Code: `
const nested = [
  [<A key="x" />, <A />],
  [<A key="x" />],
];
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingArrayKey", Line: 3, Column: 19},
			},
		},
		// ---- Upstream nested-`Children.toArray` boolean-flag quirk ----
		//
		// eslint-plugin-react uses a plain boolean `isWithinChildrenToArray`
		// that gets toggled on/off per Children.toArray enter/exit. When TWO
		// Children.toArray calls both lexically enclose some code, the inner
		// call's exit clobbers the flag even though the outer is still
		// enclosing. Subsequent map/from calls inside the outer are then
		// treated as "not inside toArray" and reported.
		//
		// These cases reproduce the quirk and lock our output 1:1 with
		// upstream ESLint — verified by running eslint-plugin-react@7
		// against these exact inputs. A depth-counter implementation would
		// silently diverge (zero reports); do not "fix" the boolean.
		{
			Code: `
React.Children.toArray([
  React.Children.toArray(a),
  xs.map(x => <A/>),
]);
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey", Line: 4, Column: 15},
			},
		},
		{
			Code: `
React.Children.toArray(
  bar(
    React.Children.toArray(a),
    xs.map(y => <A/>)
  )
);
`,
			Tsx: true,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingIterKey", Line: 5, Column: 17},
			},
		},
		// Multiple JSX children of a JSX parent, duplicate keys across
		// spread/non-spread siblings. checkKeyMustBeforeSpread catches the
		// spread violator; warnOnDuplicates clusters the identical keys.
		{
			Code: `
const div = (
  <div>
    <span {...o} key="a" />
    <span key="a" />
  </div>
);
`,
			Tsx: true,
			Options: map[string]interface{}{
				"warnOnDuplicates":         true,
				"checkKeyMustBeforeSpread": true,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "keyBeforeSpread"},
				{MessageId: "nonUniqueKeys", Line: 4},
				{MessageId: "nonUniqueKeys", Line: 5},
			},
		},
	})
}
