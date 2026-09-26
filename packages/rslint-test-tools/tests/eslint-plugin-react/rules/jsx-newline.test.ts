// Ported from eslint-plugin-react v7.37.5 (all 31 tests and 10 documentation examples).
// https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/jsx-newline.js
import { RuleTester } from '../rule-tester';

new RuleTester().run('jsx-newline', {} as never, {
  valid: [
    // Upstream valid 1
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
    // Upstream valid 2
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
      options: [
        {
          prevent: true,
        },
      ],
    },
    // Upstream valid 3
    {
      code: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `,
      features: ['fragment'],
    },
    // Upstream valid 4
    {
      code: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `,
      options: [
        {
          prevent: true,
        },
      ],
      features: ['fragment'],
    },
    // Upstream valid 5
    {
      code: `
        {/* fake-eslint-disable-next-line react/forbid-component-props */}
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          <Icon f7='gear' />
        </Button>
      `,
    },
    // Upstream valid 6
    {
      code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* fake-eslint-disable-next-line should also work inside a component */}
          <Icon f7='gear' />
        </Button>
      `,
    },
    // Upstream valid 7
    {
      code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* should work inside a component */}
          {/* and it should work when using multiple comments */}
          <Icon f7='gear' />
        </Button>
      `,
    },
    // Upstream valid 8
    {
      code: `
        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>
          {/* this is a multiline
              block comment */}
          <Icon f7='gear' />
        </Button>
      `,
    },
    // Upstream valid 9
    {
      code: `
        <>
          {/* does this */}
          <Icon f7='gear' />

          {/* also work with multiple components and inside a fragment? */}
          <OneLineComponent />
        </>
      `,
      features: ['fragment'],
    },
    // Upstream valid 10
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
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
      features: ['fragment'],
    },
    // Upstream valid 11
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
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
    },
    // Upstream valid 12
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
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
    },
    // Upstream documentation example 4
    {
      code: `<div>
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
</div>`,
      options: [
        {
          prevent: false,
        },
      ],
    },
    // Upstream documentation example 6
    {
      code: `<div>
  <Button>{data.label}</Button>
  <List />
</div>`,
      options: [
        {
          prevent: true,
        },
      ],
    },
    // Upstream documentation example 7
    {
      code: `<div>
  <Button>{data.label}</Button>
  {showSomething === true && <Something />}
</div>`,
      options: [
        {
          prevent: true,
        },
      ],
    },
    // Upstream documentation example 8
    {
      code: `<div>
  {showSomething === true && <Something />}
  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`,
      options: [
        {
          prevent: true,
        },
      ],
    },
  ],
  invalid: [
    // Upstream invalid 1
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
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 19,
        },
      ],
    },
    // Upstream invalid 2
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
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 52,
        },
      ],
    },
    // Upstream invalid 3
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
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 4,
          column: 11,
          endLine: 4,
          endColumn: 40,
        },
      ],
    },
    // Upstream invalid 4
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
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 4,
          column: 11,
          endLine: 8,
          endColumn: 13,
        },
      ],
    },
    // Upstream invalid 5
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
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 67,
        },
      ],
    },
    // Upstream invalid 6
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
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
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
      errors: [
        {
          messageId: 'allowMultilines',
          message: 'Multiline JSX elements should start in a new line',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 41,
        },
      ],
    },
    // Upstream invalid 7
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
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
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
      errors: [
        {
          messageId: 'allowMultilines',
          message: 'Multiline JSX elements should start in a new line',
          line: 7,
          column: 11,
          endLine: 7,
          endColumn: 28,
        },
      ],
    },
    // Upstream invalid 8
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
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 5,
          column: 13,
          endLine: 5,
          endColumn: 30,
        },
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 7,
          column: 11,
          endLine: 10,
          endColumn: 17,
        },
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 9,
          column: 13,
          endLine: 9,
          endColumn: 26,
        },
      ],
    },
    // Upstream invalid 9
    {
      code: `
        <div>
          <Button>{data.label}</Button>

          <List />
        </div>
      `,
      options: [
        {
          prevent: true,
        },
      ],
      output: `
        <div>
          <Button>{data.label}</Button>
          <List />
        </div>
      `,
      errors: [
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 19,
        },
      ],
    },
    // Upstream invalid 10
    {
      code: `
        <div>
          <Button>{data.label}</Button>

          {showSomething === true && <Something />}
        </div>
      `,
      options: [
        {
          prevent: true,
        },
      ],
      output: `
        <div>
          <Button>{data.label}</Button>
          {showSomething === true && <Something />}
        </div>
      `,
      errors: [
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 52,
        },
      ],
    },
    // Upstream invalid 11
    {
      code: `
        <div>
          {showSomething === true && <Something />}

          <Button>{data.label}</Button>
        </div>
      `,
      options: [
        {
          prevent: true,
        },
      ],
      output: `
        <div>
          {showSomething === true && <Something />}
          <Button>{data.label}</Button>
        </div>
      `,
      errors: [
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 40,
        },
      ],
    },
    // Upstream invalid 12
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
      options: [
        {
          prevent: true,
        },
      ],
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
      errors: [
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 5,
          column: 11,
          endLine: 9,
          endColumn: 13,
        },
      ],
    },
    // Upstream invalid 13
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
      options: [
        {
          prevent: true,
        },
      ],
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
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 6,
          column: 13,
          endLine: 6,
          endColumn: 30,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 9,
          column: 11,
          endLine: 13,
          endColumn: 17,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 12,
          column: 13,
          endLine: 12,
          endColumn: 26,
        },
      ],
    },
    // Upstream invalid 14
    {
      code: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `,
      features: ['fragment'],
      output: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `,
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 5,
          column: 11,
          endLine: 5,
          endColumn: 45,
        },
      ],
    },
    // Upstream invalid 15
    {
      code: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `,
      options: [
        {
          prevent: true,
        },
      ],
      features: ['fragment'],
      output: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `,
      errors: [
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 6,
          column: 11,
          endLine: 6,
          endColumn: 45,
        },
      ],
    },
    // Upstream invalid 16
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
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
      features: ['fragment'],
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
        {
          messageId: 'allowMultilines',
          message: 'Multiline JSX elements should start in a new line',
          line: 5,
          column: 11,
          endLine: 8,
          endColumn: 13,
        },
        {
          messageId: 'allowMultilines',
          message: 'Multiline JSX elements should start in a new line',
          line: 9,
          column: 11,
          endLine: 9,
          endColumn: 31,
        },
      ],
    },
    // Upstream invalid 17
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
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
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
      errors: [
        {
          messageId: 'allowMultilines',
          message: 'Multiline JSX elements should start in a new line',
          line: 4,
          column: 11,
          endLine: 8,
          endColumn: 13,
        },
      ],
    },
    // Upstream invalid 18
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
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
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
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 6,
          column: 13,
          endLine: 6,
          endColumn: 30,
        },
        {
          messageId: 'allowMultilines',
          message: 'Multiline JSX elements should start in a new line',
          line: 8,
          column: 11,
          endLine: 12,
          endColumn: 17,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 11,
          column: 13,
          endLine: 11,
          endColumn: 26,
        },
      ],
    },
    // Upstream invalid 19
    {
      code: '\n        const frag: DocumentFragment = (\n          <Fragment>\n            <sni-sequence-editor-tool\n              name="forward"\n              direction="forward"\n              type="control"\n              onClick={ () => this.onClickNavigate(\'forward\') }\n            />\n            <sni-sequence-editor-tool\n              name="rotate"\n              direction="left"\n              type="control"\n              onClick={ () => this.onClickNavigate(\'left\') }\n            />\n    \n            <sni-sequence-editor-tool\n              name="rotate"\n              direction="right"\n              type="control"\n              onClick={ (): void => this.onClickNavigate(\'right\') }\n            />\n    \n            <div className="sni-sequence-editor-control-panel__delete" data-name="delete" onClick={ this.onDeleteCommand } />\n    \n            {\n              ...Array.from(this.children)\n            }\n          </Fragment>\n        )\n      ',
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
      features: ['types'],
      output:
        '\n        const frag: DocumentFragment = (\n          <Fragment>\n            <sni-sequence-editor-tool\n              name="forward"\n              direction="forward"\n              type="control"\n              onClick={ () => this.onClickNavigate(\'forward\') }\n            />\n\n            <sni-sequence-editor-tool\n              name="rotate"\n              direction="left"\n              type="control"\n              onClick={ () => this.onClickNavigate(\'left\') }\n            />\n    \n            <sni-sequence-editor-tool\n              name="rotate"\n              direction="right"\n              type="control"\n              onClick={ (): void => this.onClickNavigate(\'right\') }\n            />\n    \n            <div className="sni-sequence-editor-control-panel__delete" data-name="delete" onClick={ this.onDeleteCommand } />\n    \n            {\n              ...Array.from(this.children)\n            }\n          </Fragment>\n        )\n      ',
      errors: [
        {
          messageId: 'allowMultilines',
          message: 'Multiline JSX elements should start in a new line',
          line: 10,
          column: 13,
          endLine: 15,
          endColumn: 15,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 26,
          column: 13,
          endLine: 28,
          endColumn: 14,
        },
      ],
    },
    // Upstream documentation example 1
    {
      code: `<div>
  <Button>{data.label}</Button>
  <List />
</div>`,
      options: [
        {
          prevent: false,
        },
      ],
      output: `<div>
  <Button>{data.label}</Button>

  <List />
</div>`,
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
    // Upstream documentation example 2
    {
      code: `<div>
  <Button>{data.label}</Button>
  {showSomething === true && <Something />}
</div>`,
      options: [
        {
          prevent: false,
        },
      ],
      output: `<div>
  <Button>{data.label}</Button>

  {showSomething === true && <Something />}
</div>`,
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 44,
        },
      ],
    },
    // Upstream documentation example 3
    {
      code: `<div>
  {showSomething === true && <Something />}
  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`,
      options: [
        {
          prevent: false,
        },
      ],
      output: `<div>
  {showSomething === true && <Something />}

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`,
      errors: [
        {
          messageId: 'require',
          message: 'JSX element should start in a new line',
          line: 3,
          column: 3,
          endLine: 7,
          endColumn: 5,
        },
      ],
    },
    // Upstream documentation example 5
    {
      code: `<div>
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
</div>`,
      options: [
        {
          prevent: true,
        },
      ],
      output: `<div>
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
</div>`,
      errors: [
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 11,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 6,
          column: 3,
          endLine: 11,
          endColumn: 12,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 10,
          column: 5,
          endLine: 10,
          endColumn: 18,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 13,
          column: 3,
          endLine: 13,
          endColumn: 44,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 15,
          column: 3,
          endLine: 15,
          endColumn: 28,
        },
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 17,
          column: 3,
          endLine: 21,
          endColumn: 5,
        },
      ],
    },
    // Upstream documentation example 9
    {
      code: `<div>
  {showSomething === true && <Something />}

  <Button>Button 3</Button>
  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`,
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
      output: `<div>
  {showSomething === true && <Something />}
  <Button>Button 3</Button>

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`,
      errors: [
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 28,
        },
        {
          messageId: 'allowMultilines',
          message: 'Multiline JSX elements should start in a new line',
          line: 5,
          column: 3,
          endLine: 9,
          endColumn: 5,
        },
      ],
    },
    // Upstream documentation example 10 (labeled correct; v7.37.5 reports prevent).
    {
      code: `<div>
  {showSomething === true && <Something />}

  <Button>Button 3</Button>

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`,
      options: [
        {
          prevent: true,
          allowMultilines: true,
        },
      ],
      output: `<div>
  {showSomething === true && <Something />}
  <Button>Button 3</Button>

  {showSomethingElse === true ? (
    <SomethingElse />
  ) : (
    <ErrorMessage />
  )}
</div>`,
      errors: [
        {
          messageId: 'prevent',
          message: 'JSX element should not start in a new line',
          line: 4,
          column: 3,
          endLine: 4,
          endColumn: 28,
        },
      ],
    },
  ],
});
