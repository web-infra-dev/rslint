/**
 * @fileoverview Tests for jsx-newline rule.
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/jsx-newline/jsx-newline.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, parserOptions, valid, invalid })`
 *    -> `ruleTester.run('jsx-newline', null as never, { valid, invalid })`.
 *  - Upstream wraps every case in `valids(...)` / `invalids(...)` (from
 *    `#test/parsers-jsx`), which fan each case out across the default / babel /
 *    typescript-eslint parsers and append a `// features: [...], parser: ...`
 *    line comment to `code`/`output`. That is upstream test-harness machinery to
 *    exercise multiple parsers, not part of the case data — rslint runs the one
 *    ts-go parser, so the underlying case (code / output / options / errors) is
 *    ported and the fan-out + comment-append are dropped. (A trailing line
 *    comment after the JSX root would not change this rule's diagnostics anyway:
 *    the rule only inspects JSX children, and the same comment is appended to
 *    both code and output so the fix diff is unaffected.)
 *  - `features: ['fragment']` / `features: ['types']` dropped — they only select
 *    upstream parser variants; the JSX/TS code itself is valid TSX and ts-go
 *    parses fragments and type annotations natively.
 *  - `parserOptions.ecmaFeatures.jsx` dropped — the RuleTester routes JSX to a
 *    `.tsx` fixture and ts-go enables JSX there.
 *  - All `code`/`output` are plain backtick templates (leading newline + shared
 *    indentation); that indentation is load-bearing for this whitespace rule and
 *    is preserved byte-for-byte. The one `${'    '}` interpolation in an upstream
 *    `output` (a whitespace-only line) is kept verbatim — it evaluates to four
 *    literal spaces, exactly the line the fix produces.
 *  - The `require` / `prevent` / `allowMultilines` messages are static (no
 *    `{{ }}` interpolation), so `{ messageId }` errors assert the literal text.
 *
 * No suggestions, no `skipBabel`-gated block, and no external-fixture cases exist
 * upstream. The `._css_` / `._json_` / `._markdown_` files don't exist for this
 * rule.
 *
 * KNOWN GAPS (real rslint<->upstream differences) are moved out of the live
 * `valid`/`invalid` arrays into the commented block at the bottom, each annotated
 * with upstream-expected vs. rslint-actual.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('jsx-newline', null as never, {
  valid: [
    {
      code: `
        <div>
          <Button>{data.label}</Button>

          <List />

          <Button>
            <IconPreview />
            Button 2

            <span></span>
          </Button>

          {showSomething === true && <Something />}

          <Button>Button 3</Button>

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `,
    },
    {
      code: `
        <div>
          <Button>{data.label}</Button>
          <List />
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
          {showSomething === true && <Something />}
          <Button>Button 3</Button>
          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `,
      options: [{ prevent: true }],
    },
    {
      code: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `,
    },
    {
      code: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `,
      options: [{ prevent: true }],
    },
    {
      code: `
        {/* fake-eslint-disable-next-line react/forbid-component-props */}
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          <Icon f7='gear' />
        </Button>
      `,
    },
    {
      code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* fake-eslint-disable-next-line should also work inside a component */}
          <Icon f7='gear' />
        </Button>
      `,
    },
    {
      code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* should work inside a component */}
          {/* and it should work when using multiple comments */}
          <Icon f7='gear' />
        </Button>
      `,
    },
    {
      code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* this is a multiline
              block comment */}
          <Icon f7='gear' />
        </Button>
      `,
    },
    {
      code: `
        <>
          {/* does this */}
          <Icon f7='gear' />

          {/* also work with multiple components and inside a fragment? */}
          <OneLineComponent />
        </>
      `,
    },
    {
      code: `
        <>
          <OneLineComponent />
          <AnotherOneLineComponent prop={prop} />

          <MultilineComponent
            prop1={prop1}
            prop2={prop2}
          />

          <OneLineComponent />
        </>
      `,
      options: [{ prevent: true, allowMultilines: true }],
    },
    {
      code: `
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />

          {/* but this one needs one */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `,
      options: [{ prevent: true, allowMultilines: true }],
    },
    {
      code: `
        <div>
          <Button>{data.label}</Button>
          <List />

          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>

          {showSomething === true && <Something />}
          <Button>Button 3</Button>

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}

        </div>
      `,
      options: [{ prevent: true, allowMultilines: true }],
    },
  ],
  invalid: [
    {
      code: `
        <div>
          <Button>{data.label}</Button>
          <List />
        </div>
      `,
      output: `
        <div>
          <Button>{data.label}</Button>

          <List />
        </div>
      `,
      errors: [{
        messageId: 'require',
      }],
    },
    {
      code: `
        <div>
          <Button>{data.label}</Button>
          {showSomething === true && <Something />}
        </div>
      `,
      output: `
        <div>
          <Button>{data.label}</Button>

          {showSomething === true && <Something />}
        </div>
      `,
      errors: [{ messageId: 'require' }],
    },
    {
      code: `
        <div>
          {showSomething === true && <Something />}
          <Button>{data.label}</Button>
        </div>
      `,
      output: `
        <div>
          {showSomething === true && <Something />}

          <Button>{data.label}</Button>
        </div>
      `,
      errors: [{ messageId: 'require' }],
    },
    {
      code: `
        <div>
          {showSomething === true && <Something />}
          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `,
      output: `
        <div>
          {showSomething === true && <Something />}

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `,
      errors: [{ messageId: 'require' }],
    },
    {
      code: `
        <div>
          {/* This should however still not work*/}
          <Icon f7='gear' />

          <OneLineComponent />
          {/* Comments between components still need a newLine */}
          <OneLineComponent />
        </div>
      `,
      output: `
        <div>
          {/* This should however still not work*/}
          <Icon f7='gear' />

          <OneLineComponent />

          {/* Comments between components still need a newLine */}
          <OneLineComponent />
        </div>
      `,
      errors: [{ messageId: 'require' }],
    },
    {
      code: `
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />
          {/* but this one needs one */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `,
      output: `
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />

          {/* but this one needs one */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `,
      options: [{ prevent: true, allowMultilines: true }],
      errors: [{ messageId: 'allowMultilines' }],
    },
    {
      code: `
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />
          {/* Multiline */}
          {/* Block comments */}
          {/* Stick to MultilineComponent */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `,
      output: `
        <div>
          {/* this does not have a newline */}
          <Icon f7='gear' />
          {/* neither does this */}
          <OneLineComponent />

          {/* Multiline */}
          {/* Block comments */}
          {/* Stick to MultilineComponent */}
          <Button>
            <IconPreview />
            Button 2
            <span></span>
          </Button>
        </div>
      `,
      options: [{ prevent: true, allowMultilines: true }],
      errors: [{ messageId: 'allowMultilines' }],
    },
    {
      code: `
        <div>
          <div>
            <button></button>
            <button></button>
          </div>
          <div>
            <span></span>
            <span></span>
          </div>
        </div>
      `,
      output: `
        <div>
          <div>
            <button></button>

            <button></button>
          </div>

          <div>
            <span></span>

            <span></span>
          </div>
        </div>
      `,
      errors: [
        { messageId: 'require' },
        { messageId: 'require' },
        { messageId: 'require' },
      ],
    },
    {
      output: `
        <div>
          <Button>{data.label}</Button>
          <List />
        </div>
      `,
      code: `
        <div>
          <Button>{data.label}</Button>

          <List />
        </div>
      `,
      errors: [{ messageId: 'prevent' }],
      options: [{ prevent: true }],
    },
    {
      output: `
        <div>
          <Button>{data.label}</Button>
          {showSomething === true && <Something />}
        </div>
      `,
      code: `
        <div>
          <Button>{data.label}</Button>

          {showSomething === true && <Something />}
        </div>
      `,
      errors: [{ messageId: 'prevent' }],
      options: [{ prevent: true }],
    },
    {
      output: `
        <div>
          {showSomething === true && <Something />}
          <Button>{data.label}</Button>
        </div>
      `,
      code: `
        <div>
          {showSomething === true && <Something />}

          <Button>{data.label}</Button>
        </div>
      `,
      errors: [{ messageId: 'prevent' }],
      options: [{ prevent: true }],
    },
    {
      output: `
        <div>
          {showSomething === true && <Something />}
          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `,
      code: `
        <div>
          {showSomething === true && <Something />}

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `,
      errors: [{ messageId: 'prevent' }],
      options: [{ prevent: true }],
    },
    {
      output: `
        <div>
          <div>
            <button></button>
            <button></button>
          </div>
          <div>
            <span></span>
            <span></span>
          </div>
        </div>
      `,
      code: `
        <div>
          <div>
            <button></button>

            <button></button>
          </div>

          <div>
            <span></span>

            <span></span>
          </div>
        </div>
      `,
      errors: [
        { messageId: 'prevent' },
        { messageId: 'prevent' },
        { messageId: 'prevent' },
      ],
      options: [{ prevent: true }],
    },
    {
      code: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `,
      output: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `,
      errors: [{ messageId: 'require' }],
    },
    {
      output: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `,
      code: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `,
      errors: [{ messageId: 'prevent' }],
      options: [{ prevent: true }],
    },
    {
      code: `
        <>
          <OneLineComponent />
          <AnotherOneLineComponent prop={prop} />
          <MultilineComponent
            prop1={prop1}
            prop2={prop2}
          />
          <OneLineComponent />
        </>
      `,
      output: `
        <>
          <OneLineComponent />
          <AnotherOneLineComponent prop={prop} />

          <MultilineComponent
            prop1={prop1}
            prop2={prop2}
          />

          <OneLineComponent />
        </>
      `,
      errors: [
        { messageId: 'allowMultilines' },
        { messageId: 'allowMultilines' },
      ],
      options: [{ prevent: true, allowMultilines: true }],
    },
    {
      code: `
        <div>
          {showSomething === true && <Something />}
          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `,
      output: `
        <div>
          {showSomething === true && <Something />}

          {showSomethingElse === true ? (
            <SomethingElse />
          ) : (
            <ErrorMessage />
          )}
        </div>
      `,
      errors: [{ messageId: 'allowMultilines' }],
      options: [{ prevent: true, allowMultilines: true }],
    },
    {
      output: `
        <div>
          <div>
            <button></button>
            <button></button>
          </div>

          <div>
            <span></span>
            <span></span>
          </div>
        </div>
      `,
      code: `
        <div>
          <div>
            <button></button>

            <button></button>
          </div>
          <div>
            <span></span>

            <span></span>
          </div>
        </div>
      `,
      errors: [
        { messageId: 'prevent' },
        { messageId: 'allowMultilines' },
        { messageId: 'prevent' },
      ],
      options: [{ prevent: true, allowMultilines: true }],
    },
    // NOTE: upstream embeds whitespace-only lines (four literal spaces) inside
    // both `code` and `output`; they are reproduced here as `${'    '}`
    // interpolations so the four-space lines survive editors/formatters and stay
    // byte-identical to upstream's evaluated strings.
    {
      code: `
        const frag: DocumentFragment = (
          <Fragment>
            <sni-sequence-editor-tool
              name="forward"
              direction="forward"
              type="control"
              onClick={ () => this.onClickNavigate('forward') }
            />
            <sni-sequence-editor-tool
              name="rotate"
              direction="left"
              type="control"
              onClick={ () => this.onClickNavigate('left') }
            />
${'    '}
            <sni-sequence-editor-tool
              name="rotate"
              direction="right"
              type="control"
              onClick={ (): void => this.onClickNavigate('right') }
            />
${'    '}
            <div className="sni-sequence-editor-control-panel__delete" data-name="delete" onClick={ this.onDeleteCommand } />
${'    '}
            {
              ...Array.from(this.children)
            }
          </Fragment>
        )
      `,
      output: `
        const frag: DocumentFragment = (
          <Fragment>
            <sni-sequence-editor-tool
              name="forward"
              direction="forward"
              type="control"
              onClick={ () => this.onClickNavigate('forward') }
            />

            <sni-sequence-editor-tool
              name="rotate"
              direction="left"
              type="control"
              onClick={ () => this.onClickNavigate('left') }
            />
${'    '}
            <sni-sequence-editor-tool
              name="rotate"
              direction="right"
              type="control"
              onClick={ (): void => this.onClickNavigate('right') }
            />
${'    '}
            <div className="sni-sequence-editor-control-panel__delete" data-name="delete" onClick={ this.onDeleteCommand } />
${'    '}
            {
              ...Array.from(this.children)
            }
          </Fragment>
        )
      `,
      options: [{ prevent: true, allowMultilines: true }],
      errors: [
        { messageId: 'allowMultilines', line: 10 },
        { messageId: 'prevent', line: 26 },
      ],
    },
  ],
});

/*
 * ============================ jsx-newline — KNOWN GAPS ============================
 *
 * NONE. Every upstream case (12 valid + 19 invalid) passes against rslint
 * verbatim: diagnostics (count + message + the two pinned `line`s) and every
 * autofix `output` match exactly — including the spread-child
 * `{ ...Array.from(this.children) }` case and the whitespace-only-line fix.
 */
