package jsx_key

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/react/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// All 57 canonical cases from tests/lib/rules/jsx-key.js at v7.37.5.
// Parser-matrix duplicates collapse to the shared tsgo TSX parser; both upstream
// optional-chain tag variants and both TypeScript-only inputs remain explicit.
func TestJsxKeyUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&JsxKeyRule,
		[]rule_tester.ValidTestCase{
			// eslint-plugin-react@7.37.5 valid/1
			{
				Code: "fn()",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/2
			{
				Code: "[1, 2, 3].map(function () {})",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/3
			{
				Code: "<App />;",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/4
			{
				Code: "[<App key={0} />, <App key={1} />];",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/5
			{
				Code: "[1, 2, 3].map(function(x) { return <App key={x} /> });",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/6
			{
				Code: "[1, 2, 3].map(x => <App key={x} />);",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/7
			{
				Code: "[1, 2 ,3].map(x => x && <App x={x} key={x} />);",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/8
			{
				Code: "[1, 2 ,3].map(x => x ? <App x={x} key=\"1\" /> : <OtherApp x={x} key=\"2\" />);",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/9
			{
				Code: "[1, 2, 3].map(x => { return <App key={x} /> });",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/10
			{
				Code: "Array.from([1, 2, 3], function(x) { return <App key={x} /> });",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/11
			{
				Code: "Array.from([1, 2, 3], (x => <App key={x} />));",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/12
			{
				Code: "Array.from([1, 2, 3], (x => {return <App key={x} />}));",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/13
			{
				Code: "Array.from([1, 2, 3], someFn);",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/14
			{
				Code: "Array.from([1, 2, 3]);",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/15
			{
				Code: "[1, 2, 3].foo(x => <App />);",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/16
			{
				Code: "var App = () => <div />;",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/17
			{
				Code: "[1, 2, 3].map(function(x) { return; });",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/18
			{
				Code: "foo(() => <div />);",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/19
			{
				Code: "foo(() => <></>);",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/20
			{
				Code: "<></>;",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/21
			{
				Code: "<App {...{}} />;",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/22
			{
				Code: "<App key=\"keyBeforeSpread\" {...{}} />;",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"checkKeyMustBeforeSpread": true,
				},
			},
			// eslint-plugin-react@7.37.5 valid/23
			{
				Code: "<div key=\"keyBeforeSpread\" {...{}} />;",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"checkKeyMustBeforeSpread": true,
				},
			},
			// eslint-plugin-react@7.37.5 valid/24
			{
				Code: `
        const spans = [
          <span key="notunique"/>,
          <span key="notunique"/>,
        ];
      `,
				Tsx: true,
			},
			// eslint-plugin-react@7.37.5 valid/25
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
			// eslint-plugin-react@7.37.5 valid/26
			{
				Code: `
        import React, { FC, useRef, useState } from 'react';

        import './ResourceVideo.sass';
        import VimeoVideoPlayInModal from '../vimeoVideoPlayInModal/VimeoVideoPlayInModal';

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
			// eslint-plugin-react@7.37.5 valid/27
			{
				Code: `
        // testrule.jsx
        const trackLink = () => {};
        const getAnalyticsUiElement = () => {};

        const onTextButtonClick = (e, item) => trackLink([, getAnalyticsUiElement(item), item.name], e);
      `,
				Tsx: true,
			},
			// eslint-plugin-react@7.37.5 valid/28
			{
				Code: "\n        function Component({ allRatings }) {\n          return (\n            <RatingDetailsStyles>\n              {Object.entries(allRatings)?.map(([key, value], index) => {\n                const rate = value?.split(/(?=[%, /])/);\n\n                if (!rate) return null;\n\n                return (\n                  <li key={`${entertainment.tmdbId}${index}`}>\n                    <img src={`/assets/rating/${key}.png`} />\n                    <span className=\"rating-details--rate\">{rate?.[0]}</span>\n                    <span className=\"rating-details--rate-suffix\">{rate?.[1]}</span>\n                  </li>\n                );\n              })}\n            </RatingDetailsStyles>\n          );\n        }\n      ",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/29
			{
				Code: `
        const baz = foo?.bar?.()?.[1] ?? 'qux';

        qux()?.map()

        const directiveRanges = comments?.map(tryParseTSDirective)
      `,
				Tsx: true,
			},
			// eslint-plugin-react@7.37.5 valid/30
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
			// eslint-plugin-react@7.37.5 valid/31
			{
				Code: "React.Children.toArray([1, 2 ,3].map(x => <App />));",
				Tsx:  true,
			},
			// eslint-plugin-react@7.37.5 valid/32
			{
				Code: `
        import { Children } from "react";
        Children.toArray([1, 2 ,3].map(x => <App />));
      `,
				Tsx: true,
			},
			// eslint-plugin-react@7.37.5 valid/33
			{
				Code: `
        import Act from 'react';
        import { Children as ReactChildren } from 'react';

        const { Children } = Act;
        const { toArray } = Children;

        Act.Children.toArray([1, 2 ,3].map(x => <App />));
        Act.Children.toArray(Array.from([1, 2 ,3], x => <App />));
        Children.toArray([1, 2 ,3].map(x => <App />));
        Children.toArray(Array.from([1, 2 ,3], x => <App />));
        // ReactChildren.toArray([1, 2 ,3].map(x => <App />));
        // ReactChildren.toArray(Array.from([1, 2 ,3], x => <App />));
        // toArray([1, 2 ,3].map(x => <App />));
        // toArray(Array.from([1, 2 ,3], x => <App />));
      `,
				Tsx: true,
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// eslint-plugin-react@7.37.5 invalid/1
			{
				Code: "[<App />];",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/2
			{
				Code: "[<App {...key} />];",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/3
			{
				Code: "[<App key={0}/>, <App />];",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/4
			{
				Code: "[1, 2 ,3].map(function(x) { return <App /> });",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/5
			{
				Code: "[1, 2 ,3].map(x => <App />);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/6
			{
				Code: "[1, 2 ,3].map(x => x && <App x={x} />);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/7
			{
				Code: "[1, 2 ,3].map(x => x ? <App x={x} key=\"1\" /> : <OtherApp x={x} />);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/8
			{
				Code: "[1, 2 ,3].map(x => x ? <App x={x} /> : <OtherApp x={x} key=\"2\" />);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/9
			{
				Code: "[1, 2 ,3].map(x => { return <App /> });",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/10
			{
				Code: "Array.from([1, 2 ,3], function(x) { return <App /> });",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/11
			{
				Code: "Array.from([1, 2 ,3], (x => { return <App /> }));",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/12
			{
				Code: "Array.from([1, 2 ,3], (x => <App />));",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/13
			{
				Code: "[1, 2, 3]?.map(x => <BabelEslintApp />)",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/14
			{
				Code: "[1, 2, 3]?.map(x => <TypescriptEslintApp />)",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/15
			{
				Code: "[1, 2, 3].map(x => <>{x}</>);",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"checkFragmentShorthand": true,
				},
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingIterKeyUsePrag",
						Message:   "Missing \"key\" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use Act.Frag instead",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/16
			{
				Code: "[<></>];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"checkFragmentShorthand": true,
				},
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missingArrayKeyUsePrag",
						Message:   "Missing \"key\" prop for element in array. Shorthand fragment syntax does not support providing keys. Use Act.Frag instead",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/17
			{
				Code: "[<App {...obj} key=\"keyAfterSpread\" />];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"checkKeyMustBeforeSpread": true,
				},
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "keyBeforeSpread",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/18
			{
				Code: "[<div {...obj} key=\"keyAfterSpread\" />];",
				Tsx:  true,
				Options: map[string]interface {
				}{
					"checkKeyMustBeforeSpread": true,
				},
				Settings: map[string]interface {
				}{
					"react": map[string]interface {
					}{
						"pragma":   "Act",
						"fragment": "Frag",
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "keyBeforeSpread",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/19
			{
				Code: `
        const spans = [
          <span key="notunique"/>,
          <span key="notunique"/>,
        ];
      `,
				Tsx: true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "nonUniqueKeys",
						Line:      3,
					},
					{
						MessageId: "nonUniqueKeys",
						Line:      4,
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/20
			{
				Code: `
        const div = (
          <div>
            <span key="notunique"/>
            <span key="notunique"/>
          </div>
        );
      `,
				Tsx: true,
				Options: map[string]interface {
				}{
					"warnOnDuplicates": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "nonUniqueKeys",
						Line:      4,
					},
					{
						MessageId: "nonUniqueKeys",
						Line:      5,
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/21
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
					{
						MessageId: "missingIterKey",
					},
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/22
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
					{
						MessageId: "missingIterKey",
					},
					{
						MessageId: "missingIterKey",
					},
					{
						MessageId: "missingIterKey",
					},
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/23
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
					{
						MessageId: "missingIterKey",
					},
					{
						MessageId: "missingIterKey",
					},
					{
						MessageId: "missingIterKey",
					},
				},
			},
			// eslint-plugin-react@7.37.5 invalid/24
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
				Tsx: true,
				Options: map[string]interface {
				}{
					"checkKeyMustBeforeSpread": true,
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "keyBeforeSpread",
					},
				},
			},
		})
}
