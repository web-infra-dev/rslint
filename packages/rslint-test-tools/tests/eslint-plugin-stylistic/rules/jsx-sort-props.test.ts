/**
 * @fileoverview Tests for jsx-sort-props
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-sort-props/jsx-sort-props.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, ... })` ->
 *    `ruleTester.run('jsx-sort-props', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaFeatures.jsx`) dropped — rslint resolves via tsconfig
 *    and the RuleTester routes every JSX fixture (these carry `/>` / `</Tag`) to
 *    a `.tsx` file, where ts-go parses JSX correctly.
 *  - The local `expected*Error` helpers (e.g. `expectedError`,
 *    `expectedCallbackError`, `expectedReservedFirstError`, …) are inlined to
 *    their final `{ messageId: '<id>' }` objects. The plugin's `meta.messages`
 *    for every id is a STATIC string (no `{{placeholder}}`), so the RuleTester
 *    asserts the rendered message directly and no `data` is needed.
 *  - `options` arrays are inlined from the `*Args` consts to their literal value.
 *
 * The upstream file wraps every case in the `valids()` / `invalids()` helpers
 * from `shared/test-utils/parsers-jsx.ts`. Those helpers multiplex each case
 * across several PARSERS (ESLint-default, @babel/eslint-parser,
 * @typescript-eslint/parser) and append a `// features: [...], parser: ...`
 * comment to `code`/`output`. With the resolved toolchain (ESLint 10.x) the
 * babel variant is skipped (`skipBabel = gte(ESLint.version, '10.0.0')` ===
 * true). That parser-multiplexing is a pure upstream-harness artifact with no
 * rslint analog, so each case is ported ONCE as its literal source (the
 * `features` field and the appended parser-comment are dropped); the code
 * itself is verbatim, including the leading newline + indentation of every plain
 * backtick template (load-bearing: the `line` pins are computed against that
 * exact indented source).
 *
 * The final invalid case carried `verifyFixChanges: false` upstream (the only
 * effect of that flag, per eslint-vitest-rule-tester, is to drop the "fix must
 * change the code" assertion). It pins `errors` but no `output`, so it is ported
 * `output`-less: the RuleTester asserts its diagnostics and does not check the
 * fix — faithful to upstream.
 *
 * Several invalid cases pin `output` plus a numeric `errors` count: rslint fixes
 * to a fixpoint (multi-pass) whereas ESLint RuleTester `output` is a single pass,
 * but jsx-sort-props reorders all attributes in one pass, so the fixpoint equals
 * the single-pass output and every `output` matches verbatim.
 *
 * There are NO `$` unindent template tags, NO `readFileSync` external-fixture
 * cases, and NO `suggestions`. The `._css_` / `._json_` / `._markdown_` test
 * files don't exist for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so nothing is moved to KNOWN GAPS.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-sort-props', null as never, {
  valid: [
    { code: '<App />;' },
    { code: '<App {...this.props} />;' },
    { code: '<App a b c />;' },
    { code: '<App {...this.props} a b c />;' },
    { code: '<App c {...this.props} a b />;' },
    { code: '<App a="c" b="b" c="a" />;' },
    { code: '<App {...this.props} a="c" b="b" c="a" />;' },
    { code: '<App c="a" {...this.props} a="c" b="b" />;' },
    { code: '<App A a />;' },
    { code: '<App aB aa/>;' },
    { code: '<App aA aB />;' },
    { code: '<App aB aaa />;' },
    { code: '<App a aB aa />;' },
    { code: '<App Number="2" name="John" />;' },
    // Ignoring case
    { code: '<App a A />;', options: [{ ignoreCase: true }] },
    { code: '<App aa aB />;', options: [{ ignoreCase: true }] },
    { code: '<App a B c />;', options: [{ ignoreCase: true }] },
    { code: '<App A b C />;', options: [{ ignoreCase: true }] },
    { code: '<App name="John" Number="2" />;', options: [{ ignoreCase: true }] },
    // Sorting callbacks below all other props
    { code: '<App a z onBar onFoo />;', options: [{ callbacksLast: true }] },
    { code: '<App z onBar onFoo />;', options: [{ callbacksLast: true, ignoreCase: true }] },
    // Sorting shorthand props before others
    { code: '<App a b="b" />;', options: [{ shorthandFirst: true }] },
    { code: '<App z a="a" />;', options: [{ shorthandFirst: true }] },
    { code: '<App x y z a="a" b="b" />;', options: [{ shorthandFirst: true }] },
    { code: '<App a="a" b="b" x y z />;', options: [{ shorthandLast: true }] },
    {
      code: '<App a="a" b="b" x y z onBar onFoo />;',
      options: [{ callbacksLast: true, shorthandLast: true }],
    },
    // Sorting multiline props before others
    {
      code: `
        <App
          a={{
            aA: 1,
          }}
          b
        />
      `,
      options: [{ multiline: 'first' }],
    },
    {
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
    },
    {
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
    },
    // Sorting multiline props after others
    {
      code: `
        <App
          a
          b={{
            bB: 1,
          }}
        />
      `,
      options: [{ multiline: 'last' }],
    },
    {
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
    },
    {
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
      options: [{ multiline: 'last', shorthandLast: true, callbacksLast: true }],
    },
    // noSortAlphabetically
    { code: '<App a b />;', options: [{ noSortAlphabetically: true }] },
    { code: '<App b a />;', options: [{ noSortAlphabetically: true }] },
    // reservedFirst
    {
      code: '<App children={<App />} key={0} ref="r" a b c />',
      options: [{ reservedFirst: true }],
    },
    {
      code: '<App children={<App />} key={0} ref="r" a b c dangerouslySetInnerHTML={{__html: "EPR"}} />',
      options: [{ reservedFirst: true }],
    },
    {
      code: '<App children={<App />} key={0} a ref="r" />',
      options: [{ reservedFirst: ['children', 'dangerouslySetInnerHTML', 'key'] }],
    },
    {
      code: '<App children={<App />} key={0} a dangerouslySetInnerHTML={{__html: "EPR"}} ref="r" />',
      options: [{ reservedFirst: ['children', 'dangerouslySetInnerHTML', 'key'] }],
    },
    {
      code: '<App ref="r" key={0} children={<App />} b a c />',
      options: [{ noSortAlphabetically: true, reservedFirst: true }],
    },
    {
      code: '<div ref="r" dangerouslySetInnerHTML={{__html: "EPR"}} key={0} children={<App />} b a c />',
      options: [{ noSortAlphabetically: true, reservedFirst: true }],
    },
    {
      code: '<App key="key" c="c" b />',
      options: [{ reservedFirst: true, shorthandLast: true }],
    },
    {
      code: `
        <RawFileField
          onChange={handleChange}
          onFileRemove={asMedia ? null : handleRemove}
          {...props}
        />
      `,
    },
    {
      code: `
        <RawFileField
          onFileRemove={asMedia ? null : handleRemove}
          onChange={handleChange}
          {...props}
        />
      `,
      options: [{ locale: 'sk-SK' }],
    },
  ],
  invalid: [
    {
      code: '<App b a />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App a b />;',
    },
    {
      code: '<App aB a />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App a aB />;',
    },
    {
      code: '<App fistName="John" tel={5555555} name="John Smith" lastName="Smith" Number="2" />;',
      errors: [{ messageId: 'sortPropsByAlpha' }, { messageId: 'sortPropsByAlpha' }, { messageId: 'sortPropsByAlpha' }],
      output: '<App Number="2" fistName="John" lastName="Smith" name="John Smith" tel={5555555} />;',
    },
    {
      code: '<App aa aB />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App aB aa />;',
    },
    {
      code: '<App aB aA />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App aA aB />;',
    },
    {
      code: '<App aaB aA />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App aA aaB />;',
    },
    {
      code: '<App aaB aaa aA a />;',
      errors: [{ messageId: 'sortPropsByAlpha' }, { messageId: 'sortPropsByAlpha' }],
      output: '<App a aA aaB aaa />;',
    },
    {
      code: '<App {...this.props} b a />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App {...this.props} a b />;',
    },
    {
      code: '<App c {...this.props} b a />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App c {...this.props} a b />;',
    },
    {
      code: '<App fistName="John" tel={5555555} name="John Smith" lastName="Smith" Number="2" />;',
      options: [{ ignoreCase: true }],
      errors: [{ messageId: 'sortPropsByAlpha' }, { messageId: 'sortPropsByAlpha' }, { messageId: 'sortPropsByAlpha' }],
      output: '<App fistName="John" lastName="Smith" name="John Smith" Number="2" tel={5555555} />;',
    },
    {
      code: '<App B a />;',
      options: [{ ignoreCase: true }],
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App a B />;',
    },
    {
      code: '<App B A c />;',
      options: [{ ignoreCase: true }],
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App A B c />;',
    },
    {
      code: '<App c="a" a="c" b="b" />;',
      output: '<App a="c" b="b" c="a" />;',
      errors: 2,
    },
    {
      code: '<App {...this.props} c="a" a="c" b="b" />;',
      output: '<App {...this.props} a="c" b="b" c="a" />;',
      errors: 2,
    },
    {
      code: '<App d="d" b="b" {...this.props} c="a" a="c" />;',
      output: '<App b="b" d="d" {...this.props} a="c" c="a" />;',
      errors: 2,
    },
    {
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
      errors: 3,
    },
    {
      code: '<App b={2} c={3} d={4} e={5} f={6} g={7} h={8} i={9} j={10} k={11} a={1} />',
      output: '<App a={1} b={2} c={3} d={4} e={5} f={6} g={7} h={8} i={9} j={10} k={11} />',
      errors: 1,
    },
    {
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
      errors: 10,
    },
    {
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
      errors: 13,
    },
    {
      code: '<App key="key" b c="c" />',
      errors: [{ messageId: 'listShorthandLast' }],
      options: [{ reservedFirst: true, shorthandLast: true }],
      output: '<App key="key" c="c" b />',
    },
    {
      code: '<App ref="ref" key="key" isShorthand veryLastAttribute="yes" />',
      errors: [{ messageId: 'sortPropsByAlpha' }, { messageId: 'listShorthandLast' }],
      options: [{ reservedFirst: true, shorthandLast: true }],
      output: '<App key="key" ref="ref" veryLastAttribute="yes" isShorthand />',
    },
    {
      code: '<App a v-for={i in 4} v-if={true} b />',
      errors: [{ messageId: 'listReservedPropsFirst' }, { messageId: 'listReservedPropsFirst' }],
      options: [{ reservedFirst: ['v-if', 'v-for'] }],
      output: '<App v-if={true} v-for={i in 4} a b />',
    },
    {
      code: '<App v-slot={{ foo }} v-slots={{}} onClick={() => {}} />',
      errors: [{ messageId: 'listReservedPropsLast' }, { messageId: 'listReservedPropsLast' }],
      options: [{ reservedLast: ['v-slot', 'v-slots'] }],
      output: '<App onClick={() => {}} v-slots={{}} v-slot={{ foo }} />',
    },
    {
      code: '<App a v-model:b={foo} b v-model={foo} c v-slot:b={{ foo }} v-slot:a={{ foo }} onClick={() => {}} />',
      errors: [{ messageId: 'listReservedPropsFirst' }, { messageId: 'listReservedPropsFirst' }, { messageId: 'sortPropsByAlpha' }, { messageId: 'listReservedPropsLast' }],
      options: [{ reservedFirst: ['v-model'], reservedLast: ['v-slot'] }],
      output: '<App v-model={foo} v-model:b={foo} a b c onClick={() => {}} v-slot:a={{ foo }} v-slot:b={{ foo }} />',
    },
    {
      code: '<App a v-model:b={foo} b v-model={foo} c v-slot:b={{ foo }} v-slot:a={{ foo }} onClick={() => {}} />',
      errors: [{ messageId: 'listReservedPropsFirst' }, { messageId: 'listReservedPropsFirst' }, { messageId: 'listReservedPropsLast' }],
      options: [{ noSortAlphabetically: true, reservedFirst: ['v-model'], reservedLast: ['v-slot'] }],
      output: '<App v-model:b={foo} v-model={foo} a b c onClick={() => {}} v-slot:b={{ foo }} v-slot:a={{ foo }} />',
    },
    {
      code: '<App a z onFoo onBar />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      options: [{ callbacksLast: true }],
      output: '<App a z onBar onFoo />;',
    },
    {
      code: '<App a onBar onFoo z />;',
      errors: [{ messageId: 'listCallbacksLast' }],
      options: [{ callbacksLast: true }],
      output: '<App a z onBar onFoo />;',
    },
    {
      code: '<App a="a" b />;',
      errors: [{ messageId: 'listShorthandFirst' }],
      options: [{ shorthandFirst: true }],
      output: '<App b a="a" />;',
    },
    {
      code: '<App z x a="a" />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      options: [{ shorthandFirst: true }],
      output: '<App x z a="a" />;',
    },
    {
      code: '<App b a="a" />;',
      errors: [{ messageId: 'listShorthandLast' }],
      options: [{ shorthandLast: true }],
      output: '<App a="a" b />;',
    },
    {
      code: '<App a="a" onBar onFoo z x />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      options: [{ shorthandLast: true }],
      output: '<App a="a" onBar onFoo x z />;',
    },
    {
      code: '<App b a />;',
      errors: [{ messageId: 'sortPropsByAlpha' }],
      options: [{ noSortAlphabetically: false }],
      output: '<App a b />;',
    },
    // reservedFirst
    {
      code: '<App a key={1} />',
      options: [{ reservedFirst: true }],
      errors: [{ messageId: 'listReservedPropsFirst' }],
      output: '<App key={1} a />',
    },
    {
      code: '<div a dangerouslySetInnerHTML={{__html: "EPR"}} />',
      options: [{ reservedFirst: true }],
      errors: [{ messageId: 'listReservedPropsFirst' }],
      output: '<div dangerouslySetInnerHTML={{__html: "EPR"}} a />',
    },
    {
      code: '<App ref="r" key={2} b />',
      options: [{ reservedFirst: true }],
      errors: [{ messageId: 'sortPropsByAlpha' }],
      output: '<App key={2} ref="r" b />',
    },
    {
      code: '<App key={2} b a />',
      options: [{ reservedFirst: true }],
      output: '<App key={2} a b />',
      errors: [{ messageId: 'sortPropsByAlpha' }],
    },
    {
      code: '<App b a />',
      options: [{ reservedFirst: true }],
      output: '<App a b />',
      errors: [{ messageId: 'sortPropsByAlpha' }],
    },
    {
      code: '<App dangerouslySetInnerHTML={{__html: "EPR"}} e key={2} b />',
      options: [{ reservedFirst: true }],
      output: '<App key={2} b dangerouslySetInnerHTML={{__html: "EPR"}} e />',
      errors: [{ messageId: 'listReservedPropsFirst' }, { messageId: 'sortPropsByAlpha' }],
    },
    {
      code: '<App key={3} children={<App />} />',
      options: [{ reservedFirst: ['children', 'dangerouslySetInnerHTML', 'key'] }],
      errors: [{ messageId: 'listReservedPropsFirst' }],
      output: '<App children={<App />} key={3} />',
    },
    {
      code: '<App z ref="r" />',
      options: [{ noSortAlphabetically: true, reservedFirst: true }],
      errors: [{ messageId: 'listReservedPropsFirst' }],
      output: '<App ref="r" z />',
    },
    {
      code: '<App key={4} />',
      options: [{ reservedFirst: [] }],
      errors: [{ messageId: 'listIsEmpty' }],
    },
    {
      code: '<App onBar z />;',
      output: '<App z onBar />;',
      options: [{ callbacksLast: true, reservedFirst: true }],
      errors: [{ messageId: 'listCallbacksLast' }],
    // multiline first
    },
    {
      code: `
        <App
          a
          b={{
            bB: 1,
          }}
        />
      `,
      options: [{ multiline: 'first' }],
      errors: [{ messageId: 'listMultilineFirst' }],
      output: `
        <App
          b={{
            bB: 1,
          }}
          a
        />
      `,
    },
    {
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
      errors: [{ messageId: 'listMultilineFirst' }, { messageId: 'listShorthandFirst' }],
      output: `
        <App
          c
          b={{
            bB: 1,
          }}
          a={1}
        />
      `,
    },
    // multiline last
    {
      code: `
        <App
          a={{
            aA: 1,
          }}
          b
        />
      `,
      options: [{ multiline: 'last' }],
      errors: [{ messageId: 'listMultilineLast' }],
      output: `
        <App
          b
          a={{
            aA: 1,
          }}
        />
      `,
    },
    {
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
      options: [{ multiline: 'last', shorthandLast: true, callbacksLast: true }],
      errors: [
        {
          messageId: 'listShorthandLast',
          line: 6,
        },
        {
          messageId: 'listCallbacksLast',
          line: 8,
        },
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
    },
    {
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
      errors: [
        {
          messageId: 'listMultilineLast',
          line: 4,
        },
        {
          messageId: 'listReservedPropsFirst',
          line: 12,
        },
        {
          messageId: 'listShorthandFirst',
          line: 13,
        },
        {
          messageId: 'listCallbacksLast',
          line: 19,
        },
      ],
    },
    {
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
      errors: [
        {
          messageId: 'sortPropsByAlpha',
          line: 6,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 8,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 9,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 10,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 11,
        },
      ],
    },
    {
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
      errors: [
        {
          messageId: 'sortPropsByAlpha',
          line: 6,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 7,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 8,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 9,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 10,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 11,
        },
      ],
    },
    {
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
      errors: [
        {
          messageId: 'sortPropsByAlpha',
          line: 5,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 7,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 8,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 10,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 11,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 12,
        },
      ],
    },
    {
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
      errors: [
        {
          messageId: 'sortPropsByAlpha',
          line: 8,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 10,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 11,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 12,
        },
      ],
    },
    {
      code: `
        <foo a={0} b={1} /* comment for b and ab */ ab={1} aa={0} />
      `,
      output: `
        <foo a={0} aa={0} b={1} /* comment for b and ab */ ab={1} />
      `,
      errors: [
        {
          messageId: 'sortPropsByAlpha',
          line: 2,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 2,
        },
      ],
    },
    {
      code: `
        <ReactJson src={rowResult} name="data" collapsed={4} collapseStringsAfterLength={60} onEdit={onEdit} /* onDelete={onEdit} */ />
      `,
      output: `
        <ReactJson collapseStringsAfterLength={60} collapsed={4} name="data" src={rowResult} onEdit={onEdit} /* onDelete={onEdit} */ />
      `,
      errors: [
        {
          messageId: 'sortPropsByAlpha',
          line: 2,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 2,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 2,
        },
        {
          messageId: 'sortPropsByAlpha',
          line: 2,
        },
      ],
    },
    {
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
      errors: [
        {
          messageId: 'sortPropsByAlpha',
          line: 11,
        },
      ],
    },
  ],
});
