import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';
import { diagnostic } from '../expectations';

function expectSortPropsByAlphaError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'sortPropsByAlpha',
    'Props should be sorted alphabetically',
    range,
  );
}

function expectListCallbacksLastError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'listCallbacksLast',
    'Callbacks must be listed after all other props',
    range,
  );
}

function expectListShorthandFirstError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'listShorthandFirst',
    'Shorthand props must be listed before all other props',
    range,
  );
}

function expectListReservedPropsFirstError(
  range: Parameters<typeof diagnostic>[2],
) {
  return diagnostic(
    'listReservedPropsFirst',
    'Reserved props must be listed before all other props',
    range,
  );
}

function expectListShorthandLastError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'listShorthandLast',
    'Shorthand props must be listed after all other props',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-sort-props', {
  valid: [
    { code: '<App />' },
    { code: '<App {...props} a b />' },
    { code: '<App a z onBar onFoo />', options: [{ callbacksLast: true }] },
    { code: '<App a b="b" />', options: [{ shorthandFirst: true }] },
    { code: '<App a="a" b="b" x y />', options: [{ shorthandLast: true }] },
    {
      code: '<App children={<App />} key={0} ref="r" a />',
      options: [{ reservedFirst: true }],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 0 (7.37.5)',
      code: `<App />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `<App {...this.props} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `<App a b c />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `<App {...this.props} a b c />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `<App c {...this.props} a b />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<App a="c" b="b" c="a" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 6 (7.37.5)',
      code: `<App {...this.props} a="c" b="b" c="a" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `<App c="a" {...this.props} a="c" b="b" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `<App A a />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<App aB aa/>;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<App aA aB />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `<App aB aaa />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 12 (7.37.5)',
      code: `<App a aB aa />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 13 (7.37.5)',
      code: `<App Number="2" name="John" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 14 (7.37.5)',
      code: `<App a A />;`,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 15 (7.37.5)',
      code: `<App aa aB />;`,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 16 (7.37.5)',
      code: `<App a B c />;`,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 17 (7.37.5)',
      code: `<App A b C />;`,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 18 (7.37.5)',
      code: `<App name="John" Number="2" />;`,
      options: [{ ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 19 (7.37.5)',
      code: `<App a z onBar onFoo />;`,
      options: [{ callbacksLast: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 20 (7.37.5)',
      code: `<App z onBar onFoo />;`,
      options: [{ callbacksLast: true, ignoreCase: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 21 (7.37.5)',
      code: `<App a b="b" />;`,
      options: [{ shorthandFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 22 (7.37.5)',
      code: `<App z a="a" />;`,
      options: [{ shorthandFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 23 (7.37.5)',
      code: `<App x y z a="a" b="b" />;`,
      options: [{ shorthandFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 24 (7.37.5)',
      code: `<App a="a" b="b" x y z />;`,
      options: [{ shorthandLast: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 25 (7.37.5)',
      code: `<App a="a" b="b" x y z onBar onFoo />;`,
      options: [{ callbacksLast: true, shorthandLast: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 26 (7.37.5)',
      code: `
        <App
          a={{
            aA: 1,
          }}
          b
        />
      `,
      options: [{ multiline: 'first' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 27 (7.37.5)',
      code: `
        <App
          a={{
            aA: 1,
          }}
          b={[
            1,
          ]}
          c
          d
        />
      `,
      options: [{ multiline: 'first' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 28 (7.37.5)',
      code: `
        <App
          a
          b
          c={{
            cC: 1,
          }}
          d={[
            1,
          ]}
          e="1"
        />
      `,
      options: [{ multiline: 'first', shorthandFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 29 (7.37.5)',
      code: `
        <App
          a
          b={{
            bB: 1,
          }}
        />
      `,
      options: [{ multiline: 'last' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 30 (7.37.5)',
      code: `
        <App
          a
          b
          c="1"
          d={{
            dD: 1,
          }}
          e={[
            1,
          ]}
        />
      `,
      options: [{ multiline: 'last' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 31 (7.37.5)',
      code: `
        <App
          a={1}
          b="1"
          c={{
            cC: 1,
          }}
          d={() => (
            1
          )}
          e
          f
          onClick={() => ({
            gG: 1,
          })}
        />
      `,
      options: [
        { multiline: 'last', shorthandLast: true, callbacksLast: true },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 32 (7.37.5)',
      code: `<App a b />;`,
      options: [{ noSortAlphabetically: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 33 (7.37.5)',
      code: `<App b a />;`,
      options: [{ noSortAlphabetically: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 34 (7.37.5)',
      code: `<App children={<App />} key={0} ref="r" a b c />`,
      options: [{ reservedFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 35 (7.37.5)',
      code: `<App children={<App />} key={0} ref="r" a b c dangerouslySetInnerHTML={{__html: "EPR"}} />`,
      options: [{ reservedFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 36 (7.37.5)',
      code: `<App children={<App />} key={0} a ref="r" />`,
      options: [
        { reservedFirst: ['children', 'dangerouslySetInnerHTML', 'key'] },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 37 (7.37.5)',
      code: `<App children={<App />} key={0} a dangerouslySetInnerHTML={{__html: "EPR"}} ref="r" />`,
      options: [
        { reservedFirst: ['children', 'dangerouslySetInnerHTML', 'key'] },
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 38 (7.37.5)',
      code: `<App ref="r" key={0} children={<App />} b a c />`,
      options: [{ noSortAlphabetically: true, reservedFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 39 (7.37.5)',
      code: `<div ref="r" dangerouslySetInnerHTML={{__html: "EPR"}} key={0} children={<App />} b a c />`,
      options: [{ noSortAlphabetically: true, reservedFirst: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 40 (7.37.5)',
      code: `<App key="key" c="c" b />`,
      options: [{ reservedFirst: true, shorthandLast: true }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 41 (7.37.5)',
      code: `
        <RawFileField
          onChange={handleChange}
          onFileRemove={asMedia ? null : handleRemove}
          {...props}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 42 (7.37.5)',
      code: `
        <RawFileField
          onFileRemove={asMedia ? null : handleRemove}
          onChange={handleChange}
          {...props}
        />
      `,
      options: [{ locale: 'sk-SK' }],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: '<App b a />',
      errors: [expectSortPropsByAlphaError([1, 8, 1, 9])],
      output: `<App a b />`,
    },
    {
      code: '<App a onBar onFoo z />',
      options: [{ callbacksLast: true }],
      errors: [expectListCallbacksLastError([1, 14, 1, 19])],
      output: `<App a z onBar onFoo />`,
    },
    {
      code: '<App a="a" b />',
      options: [{ shorthandFirst: true }],
      errors: [expectListShorthandFirstError([1, 12, 1, 13])],
      output: `<App b a="a" />`,
    },
    {
      code: '<App a key={1} />',
      options: [{ reservedFirst: true }],
      errors: [expectListReservedPropsFirstError([1, 8, 1, 11])],
      output: `<App key={1} a />`,
    },
    {
      code: '<App key={4} />',
      options: [{ reservedFirst: [] }],
      errors: [
        diagnostic(
          'listIsEmpty',
          'A customized reserved first list must not be empty',
          [1, 6, 1, 13],
        ),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `<App b a />;`,
      errors: [expectSortPropsByAlphaError([1, 8, 1, 9])],
      output: `<App a b />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `<App aB a />;`,
      errors: [expectSortPropsByAlphaError([1, 9, 1, 10])],
      output: `<App a aB />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `<App fistName="John" tel={5555555} name="John Smith" lastName="Smith" Number="2" />;`,
      errors: [
        expectSortPropsByAlphaError([1, 36, 1, 40]),
        expectSortPropsByAlphaError([1, 54, 1, 62]),
        expectSortPropsByAlphaError([1, 71, 1, 77]),
      ],
      output: `<App Number="2" fistName="John" lastName="Smith" name="John Smith" tel={5555555} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<App aa aB />;`,
      errors: [expectSortPropsByAlphaError([1, 9, 1, 11])],
      output: `<App aB aa />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 4 (7.37.5)',
      code: `<App aB aA />;`,
      errors: [expectSortPropsByAlphaError([1, 9, 1, 11])],
      output: `<App aA aB />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `<App aaB aA />;`,
      errors: [expectSortPropsByAlphaError([1, 10, 1, 12])],
      output: `<App aA aaB />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `<App aaB aaa aA a />;`,
      errors: [
        expectSortPropsByAlphaError([1, 14, 1, 16]),
        expectSortPropsByAlphaError([1, 17, 1, 18]),
      ],
      output: `<App a aA aaB aaa />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `<App {...this.props} b a />;`,
      errors: [expectSortPropsByAlphaError([1, 24, 1, 25])],
      output: `<App {...this.props} a b />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `<App c {...this.props} b a />;`,
      errors: [expectSortPropsByAlphaError([1, 26, 1, 27])],
      output: `<App c {...this.props} a b />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `<App fistName="John" tel={5555555} name="John Smith" lastName="Smith" Number="2" />;`,
      options: [{ ignoreCase: true }],
      errors: [
        expectSortPropsByAlphaError([1, 36, 1, 40]),
        expectSortPropsByAlphaError([1, 54, 1, 62]),
        expectSortPropsByAlphaError([1, 71, 1, 77]),
      ],
      output: `<App fistName="John" lastName="Smith" name="John Smith" Number="2" tel={5555555} />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `<App B a />;`,
      options: [{ ignoreCase: true }],
      errors: [expectSortPropsByAlphaError([1, 8, 1, 9])],
      output: `<App a B />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `<App B A c />;`,
      options: [{ ignoreCase: true }],
      errors: [expectSortPropsByAlphaError([1, 8, 1, 9])],
      output: `<App A B c />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 12 (7.37.5)',
      code: `<App c="a" a="c" b="b" />;`,
      errors: [
        expectSortPropsByAlphaError([1, 12, 1, 13]),
        expectSortPropsByAlphaError([1, 18, 1, 19]),
      ],
      output: `<App a="c" b="b" c="a" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 13 (7.37.5)',
      code: `<App {...this.props} c="a" a="c" b="b" />;`,
      errors: [
        expectSortPropsByAlphaError([1, 28, 1, 29]),
        expectSortPropsByAlphaError([1, 34, 1, 35]),
      ],
      output: `<App {...this.props} a="c" b="b" c="a" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 14 (7.37.5)',
      code: `<App d="d" b="b" {...this.props} c="a" a="c" />;`,
      errors: [
        expectSortPropsByAlphaError([1, 12, 1, 13]),
        expectSortPropsByAlphaError([1, 40, 1, 41]),
      ],
      output: `<App b="b" d="d" {...this.props} a="c" c="a" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 15 (7.37.5)',
      code: `
        <App
          a={true}
          z
          r
          _onClick={function(){}}
          onHandle={function(){}}
          {...this.props}
          b={false}
          {...otherProps}
        >
          {test}
        </App>
      `,
      errors: [
        expectSortPropsByAlphaError([5, 11, 5, 12]),
        expectSortPropsByAlphaError([6, 11, 6, 19]),
        expectSortPropsByAlphaError([7, 11, 7, 19]),
      ],
      output: `
        <App
          _onClick={function(){}}
          a={true}
          onHandle={function(){}}
          r
          z
          {...this.props}
          b={false}
          {...otherProps}
        >
          {test}
        </App>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 16 (7.37.5)',
      code: `<App b={2} c={3} d={4} e={5} f={6} g={7} h={8} i={9} j={10} k={11} a={1} />`,
      errors: [expectSortPropsByAlphaError([1, 68, 1, 69])],
      output: `<App a={1} b={2} c={3} d={4} e={5} f={6} g={7} h={8} i={9} j={10} k={11} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      code: `
        <List
          className={className}
          onStageAnswer={onStageAnswer}
          onCommitAnswer={onCommitAnswer}
          isFocused={isFocused}
          direction={direction}
          allowMultipleSelection={allowMultipleSelection}
          measureLongestChildNode={measureLongestChildNode}
          layoutItemsSize={layoutItemsSize}
          handleAppScroll={handleAppScroll}
          isActive={isActive}
          resetSelection={resetSelection}
          onKeyboardChoiceHovered={onKeyboardChoiceHovered}
          keyboardShortcutType
        />
      `,
      errors: [
        expectSortPropsByAlphaError([5, 11, 5, 25]),
        expectSortPropsByAlphaError([6, 11, 6, 20]),
        expectSortPropsByAlphaError([7, 11, 7, 20]),
        expectSortPropsByAlphaError([8, 11, 8, 33]),
        expectSortPropsByAlphaError([9, 11, 9, 34]),
        expectSortPropsByAlphaError([10, 11, 10, 26]),
        expectSortPropsByAlphaError([11, 11, 11, 26]),
        expectSortPropsByAlphaError([12, 11, 12, 19]),
        expectSortPropsByAlphaError([14, 11, 14, 34]),
        expectSortPropsByAlphaError([15, 11, 15, 31]),
      ],
      output: `
        <List
          allowMultipleSelection={allowMultipleSelection}
          className={className}
          direction={direction}
          handleAppScroll={handleAppScroll}
          isActive={isActive}
          isFocused={isFocused}
          keyboardShortcutType
          layoutItemsSize={layoutItemsSize}
          measureLongestChildNode={measureLongestChildNode}
          onCommitAnswer={onCommitAnswer}
          onKeyboardChoiceHovered={onKeyboardChoiceHovered}
          onStageAnswer={onStageAnswer}
          resetSelection={resetSelection}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 18 (7.37.5)',
      code: `
        <CreateNewJob
          closed={false}
          flagOptions={flagOptions}
          jobHeight={300}
          jobWidth={200}
          campaign='Some Campaign name'
          campaignStart={moment('2018-07-28 00:00:00')}
          campaignFinish={moment('2018-09-01 00:00:00')}
          jobNumber={'Job Number can be a String'}
          jobTemplateOptions={jobTemplateOptions}
          numberOfPages={30}
          onChange={onChange}
          onClose={onClose}
          spreadSheetTemplateOptions={spreadSheetTemplateOptions}
          stateMachineOptions={stateMachineOptions}
          workflowTemplateOptions={workflowTemplateOptions}
          workflowTemplateSteps={workflowTemplateSteps}
          description='Some description for this job'

          jobTemplate='1'
          stateMachine='1'
          flag='1'
          spreadSheetTemplate='1'
          workflowTemplate='1'
          validation={validation}
          onSubmit={onSubmit}
        />
      `,
      errors: [
        expectSortPropsByAlphaError([7, 11, 7, 19]),
        expectSortPropsByAlphaError([8, 11, 8, 24]),
        expectSortPropsByAlphaError([9, 11, 9, 25]),
        expectSortPropsByAlphaError([10, 11, 10, 20]),
        expectSortPropsByAlphaError([11, 11, 11, 29]),
        expectSortPropsByAlphaError([19, 11, 19, 22]),
        expectSortPropsByAlphaError([21, 11, 21, 22]),
        expectSortPropsByAlphaError([22, 11, 22, 23]),
        expectSortPropsByAlphaError([23, 11, 23, 15]),
        expectSortPropsByAlphaError([24, 11, 24, 30]),
        expectSortPropsByAlphaError([25, 11, 25, 27]),
        expectSortPropsByAlphaError([26, 11, 26, 21]),
        expectSortPropsByAlphaError([27, 11, 27, 19]),
      ],
      output: `
        <CreateNewJob
          campaign='Some Campaign name'
          campaignFinish={moment('2018-09-01 00:00:00')}
          campaignStart={moment('2018-07-28 00:00:00')}
          closed={false}
          description='Some description for this job'
          flag='1'
          flagOptions={flagOptions}
          jobHeight={300}
          jobNumber={'Job Number can be a String'}
          jobTemplate='1'
          jobTemplateOptions={jobTemplateOptions}
          jobWidth={200}
          numberOfPages={30}
          onChange={onChange}
          onClose={onClose}
          onSubmit={onSubmit}
          spreadSheetTemplate='1'

          spreadSheetTemplateOptions={spreadSheetTemplateOptions}
          stateMachine='1'
          stateMachineOptions={stateMachineOptions}
          validation={validation}
          workflowTemplate='1'
          workflowTemplateOptions={workflowTemplateOptions}
          workflowTemplateSteps={workflowTemplateSteps}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 19 (7.37.5)',
      code: `<App key="key" b c="c" />`,
      options: [{ reservedFirst: true, shorthandLast: true }],
      errors: [expectListShorthandLastError([1, 16, 1, 17])],
      output: `<App key="key" c="c" b />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 20 (7.37.5)',
      code: `<App ref="ref" key="key" isShorthand veryLastAttribute="yes" />`,
      options: [{ reservedFirst: true, shorthandLast: true }],
      errors: [
        expectSortPropsByAlphaError([1, 16, 1, 19]),
        expectListShorthandLastError([1, 26, 1, 37]),
      ],
      output: `<App key="key" ref="ref" veryLastAttribute="yes" isShorthand />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 21 (7.37.5)',
      code: `<App a z onFoo onBar />;`,
      options: [{ callbacksLast: true }],
      errors: [expectSortPropsByAlphaError([1, 16, 1, 21])],
      output: `<App a z onBar onFoo />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 22 (7.37.5)',
      code: `<App a onBar onFoo z />;`,
      options: [{ callbacksLast: true }],
      errors: [expectListCallbacksLastError([1, 14, 1, 19])],
      output: `<App a z onBar onFoo />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 23 (7.37.5)',
      code: `<App a="a" b />;`,
      options: [{ shorthandFirst: true }],
      errors: [expectListShorthandFirstError([1, 12, 1, 13])],
      output: `<App b a="a" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 24 (7.37.5)',
      code: `<App z x a="a" />;`,
      options: [{ shorthandFirst: true }],
      errors: [expectSortPropsByAlphaError([1, 8, 1, 9])],
      output: `<App x z a="a" />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 25 (7.37.5)',
      code: `<App b a="a" />;`,
      options: [{ shorthandLast: true }],
      errors: [expectListShorthandLastError([1, 6, 1, 7])],
      output: `<App a="a" b />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 26 (7.37.5)',
      code: `<App a="a" onBar onFoo z x />;`,
      options: [{ shorthandLast: true }],
      errors: [expectSortPropsByAlphaError([1, 26, 1, 27])],
      output: `<App a="a" onBar onFoo x z />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 27 (7.37.5)',
      code: `<App b a />;`,
      options: [{ noSortAlphabetically: false }],
      errors: [expectSortPropsByAlphaError([1, 8, 1, 9])],
      output: `<App a b />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 29 (7.37.5)',
      code: `<div a dangerouslySetInnerHTML={{__html: "EPR"}} />`,
      options: [{ reservedFirst: true }],
      errors: [expectListReservedPropsFirstError([1, 8, 1, 31])],
      output: `<div dangerouslySetInnerHTML={{__html: "EPR"}} a />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 30 (7.37.5)',
      code: `<App ref="r" key={2} b />`,
      options: [{ reservedFirst: true }],
      errors: [expectSortPropsByAlphaError([1, 14, 1, 17])],
      output: `<App key={2} ref="r" b />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 31 (7.37.5)',
      code: `<App key={2} b a />`,
      options: [{ reservedFirst: true }],
      errors: [expectSortPropsByAlphaError([1, 16, 1, 17])],
      output: `<App key={2} a b />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 32 (7.37.5)',
      code: `<App b a />`,
      options: [{ reservedFirst: true }],
      errors: [expectSortPropsByAlphaError([1, 8, 1, 9])],
      output: `<App a b />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 33 (7.37.5)',
      code: `<App dangerouslySetInnerHTML={{__html: "EPR"}} e key={2} b />`,
      options: [{ reservedFirst: true }],
      errors: [
        expectListReservedPropsFirstError([1, 50, 1, 53]),
        expectSortPropsByAlphaError([1, 58, 1, 59]),
      ],
      output: `<App key={2} b dangerouslySetInnerHTML={{__html: "EPR"}} e />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 34 (7.37.5)',
      code: `<App key={3} children={<App />} />`,
      options: [
        { reservedFirst: ['children', 'dangerouslySetInnerHTML', 'key'] },
      ],
      errors: [expectSortPropsByAlphaError([1, 14, 1, 22])],
      output: `<App children={<App />} key={3} />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 35 (7.37.5)',
      code: `<App z ref="r" />`,
      options: [{ noSortAlphabetically: true, reservedFirst: true }],
      errors: [expectListReservedPropsFirstError([1, 8, 1, 11])],
      output: `<App ref="r" z />`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 37 (7.37.5)',
      code: `<App key={5} />`,
      options: [{ reservedFirst: ['notReserved'] }],
      errors: [
        diagnostic(
          'noUnreservedProps',
          'A customized reserved first list must only contain a subset of React reserved props. Remove: notReserved',
          [1, 6, 1, 13],
        ),
      ],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 38 (7.37.5)',
      code: `<App onBar z />;`,
      options: [{ callbacksLast: true, reservedFirst: true }],
      errors: [expectListCallbacksLastError([1, 6, 1, 11])],
      output: `<App z onBar />;`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 39 (7.37.5)',
      code: `
        <App
          a
          b={{
            bB: 1,
          }}
        />
      `,
      options: [{ multiline: 'first' }],
      errors: [
        diagnostic(
          'listMultilineFirst',
          'Multiline props must be listed before all other props',
          [4, 11, 4, 12],
        ),
      ],
      output: `
        <App
          b={{
            bB: 1,
          }}
          a
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 40 (7.37.5)',
      code: `
        <App
          a={1}
          b={{
            bB: 1,
          }}
          c
        />
      `,
      options: [{ multiline: 'first', shorthandFirst: true }],
      errors: [
        diagnostic(
          'listMultilineFirst',
          'Multiline props must be listed before all other props',
          [4, 11, 4, 12],
        ),
        expectListShorthandFirstError([7, 11, 7, 12]),
      ],
      output: `
        <App
          c
          b={{
            bB: 1,
          }}
          a={1}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 41 (7.37.5)',
      code: `
        <App
          a={{
            aA: 1,
          }}
          b
        />
      `,
      options: [{ multiline: 'last' }],
      errors: [
        diagnostic(
          'listMultilineLast',
          'Multiline props must be listed after all other props',
          [3, 11, 3, 12],
        ),
      ],
      output: `
        <App
          b
          a={{
            aA: 1,
          }}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 42 (7.37.5)',
      code: `
        <App
          a={{
            aA: 1,
          }}
          b
          inline={1}
          onClick={() => ({
            c: 1
          })}
          d="dD"
          e={() => ({
            eE: 1
          })}
          f
        />
      `,
      options: [
        { multiline: 'last', shorthandLast: true, callbacksLast: true },
      ],
      errors: [
        expectListShorthandLastError([6, 11, 6, 12]),
        expectListCallbacksLastError([8, 11, 8, 18]),
      ],
      output: `
        <App
          d="dD"
          inline={1}
          a={{
            aA: 1,
          }}
          e={() => ({
            eE: 1
          })}
          b
          f
          onClick={() => ({
            c: 1
          })}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 43 (7.37.5)',
      code: `
        <Typography
          float
          className={classNames(classes.inputWidth, {
            [classes.noBorder]: isActive === "values",
          })}
          disabled={isDisabled}
          initialValue={computePercentage(number, count)}
          InputProps={{
            ...customInputProps,
          }}
          key={index}
          isRequired
          {...sharedTypographyProps}
          ref={textRef}
          min="0"
          name="fieldName"
          placeholder={getTranslation("field")}
          onValidate={validate}
          inputProps={{
            className: inputClassName,
          }}
          outlined
          {...rest}
        />
      `,
      options: [
        {
          multiline: 'last',
          shorthandFirst: true,
          callbacksLast: true,
          reservedFirst: true,
          ignoreCase: true,
        },
      ],
      errors: [
        diagnostic(
          'listMultilineLast',
          'Multiline props must be listed after all other props',
          [4, 11, 4, 20],
        ),
        expectListReservedPropsFirstError([12, 11, 12, 14]),
        expectListShorthandFirstError([13, 11, 13, 21]),
        expectListCallbacksLastError([19, 11, 19, 21]),
      ],
      output: `
        <Typography
          key={index}
          float
          isRequired
          disabled={isDisabled}
          initialValue={computePercentage(number, count)}
          className={classNames(classes.inputWidth, {
            [classes.noBorder]: isActive === "values",
          })}
          InputProps={{
            ...customInputProps,
          }}
          {...sharedTypographyProps}
          ref={textRef}
          outlined
          min="0"
          name="fieldName"
          placeholder={getTranslation("field")}
          inputProps={{
            className: inputClassName,
          }}
          onValidate={validate}
          {...rest}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 44 (7.37.5)',
      code: `
        <foo
          m={0}
          n={0} // this is n
          o={0}
          c={0} // this is c
          // fofof
          f={0} // this is f
          a={0}
          b={0}
          d={0}
        />
      `,
      errors: [
        expectSortPropsByAlphaError([6, 11, 6, 12]),
        expectSortPropsByAlphaError([8, 11, 8, 12]),
        expectSortPropsByAlphaError([9, 11, 9, 12]),
        expectSortPropsByAlphaError([10, 11, 10, 12]),
        expectSortPropsByAlphaError([11, 11, 11, 12]),
      ],
      output: `
        <foo
          a={0}
          b={0}
          d={0}
          m={0}
          n={0} // this is n
          o={0}
          c={0} // this is c
          // fofof
          f={0} // this is f
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 45 (7.37.5)',
      code: `
        <foo
          m={0}
          n={0} // this is n
          o={0}
          c={0} // this is c
          f={0} // this is f
          e={0}
          a={0}
          b={0}
          d={0}
        />
      `,
      errors: [
        expectSortPropsByAlphaError([6, 11, 6, 12]),
        expectSortPropsByAlphaError([7, 11, 7, 12]),
        expectSortPropsByAlphaError([8, 11, 8, 12]),
        expectSortPropsByAlphaError([9, 11, 9, 12]),
        expectSortPropsByAlphaError([10, 11, 10, 12]),
        expectSortPropsByAlphaError([11, 11, 11, 12]),
      ],
      output: `
        <foo
          a={0}
          b={0}
          c={0} // this is c
          d={0}
          e={0}
          f={0} // this is f
          m={0}
          n={0} // this is n
          o={0}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 46 (7.37.5)',
      code: `
        <foo
          a1={0}
          g={0}
          d={0} // comment for d
          // comment for d and aa
          aa={0}
          c={0} // comment for c
          // comment for c and e
          e={1}
          ab={1} // comment for ab
          f={0}
        />
      `,
      errors: [
        expectSortPropsByAlphaError([5, 11, 5, 12]),
        expectSortPropsByAlphaError([7, 11, 7, 13]),
        expectSortPropsByAlphaError([8, 11, 8, 12]),
        expectSortPropsByAlphaError([10, 11, 10, 12]),
        expectSortPropsByAlphaError([11, 11, 11, 13]),
        expectSortPropsByAlphaError([12, 11, 12, 12]),
      ],
      output: `
        <foo
          a1={0}
          ab={1} // comment for ab
          f={0}
          g={0}
          c={0} // comment for c
          // comment for c and e
          e={1}
          d={0} // comment for d
          // comment for d and aa
          aa={0}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 47 (7.37.5)',
      code: `
        <foo
          a1={0}
          ab={1}
          // comment for ab and f
          f={0}
          g={0}
          c={0} // comment for c
          // comment for c and e
          e={1}
          d={0}
          aa={1} // comment for aa
        />
      `,
      errors: [
        expectSortPropsByAlphaError([8, 11, 8, 12]),
        expectSortPropsByAlphaError([10, 11, 10, 12]),
        expectSortPropsByAlphaError([11, 11, 11, 12]),
        expectSortPropsByAlphaError([12, 11, 12, 13]),
      ],
      output: `
        <foo
          a1={0}
          aa={1} // comment for aa
          d={0}
          g={0}
          ab={1}
          // comment for ab and f
          f={0}
          c={0} // comment for c
          // comment for c and e
          e={1}
        />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 48 (7.37.5)',
      code: `
        <foo a={0} b={1} /* comment for b and ab */ ab={1} aa={0} />
      `,
      errors: [
        expectSortPropsByAlphaError([2, 53, 2, 55]),
        expectSortPropsByAlphaError([2, 60, 2, 62]),
      ],
      output: `
        <foo a={0} aa={0} b={1} /* comment for b and ab */ ab={1} />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 49 (7.37.5)',
      code: `
        <ReactJson src={rowResult} name="data" collapsed={4} collapseStringsAfterLength={60} onEdit={onEdit} /* onDelete={onEdit} */ />
      `,
      errors: [
        expectSortPropsByAlphaError([2, 36, 2, 40]),
        expectSortPropsByAlphaError([2, 48, 2, 57]),
        expectSortPropsByAlphaError([2, 62, 2, 88]),
        expectSortPropsByAlphaError([2, 94, 2, 100]),
      ],
      output: `
        <ReactJson collapseStringsAfterLength={60} collapsed={4} name="data" src={rowResult} onEdit={onEdit} /* onDelete={onEdit} */ />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 50 (7.37.5)',
      code: `
        <Page
          // Pass all the props to the Page component.
          {...props}
          // Use the platform specific props from the doc.ts file.
          {...TemplatePageProps[platform]}
          // Use the getSubTitle helper function to get the page header subtitle from the active platform.
          subTitle={getSubTitle(platform)}
          // You can define custom sections using the \`otherSections\` prop.
          // Here it is using a method that takes the platform as an argument to return the correct array of section props.
          otherSections={_otherSections(platform) as IPageSectionProps[]}

          // You can hide the side rail by setting \`showSideRail\` to false.
          // showSideRail={false}

          // You can pass a custom className to the page wrapper if needed.
          // className="customPageClassName"
        />
      `,
      errors: [expectSortPropsByAlphaError([11, 11, 11, 24])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
