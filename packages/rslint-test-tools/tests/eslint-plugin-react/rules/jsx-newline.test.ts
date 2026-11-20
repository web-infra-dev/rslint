// Ported from eslint-plugin-react v7.37.5 (all 31 tests and 10 documentation examples).
// https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/jsx-newline.js
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

function expectRequireError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('require', 'JSX element should start in a new line', range);
}

function expectAllowMultilinesError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'allowMultilines',
    'Multiline JSX elements should start in a new line',
    range,
  );
}

function expectPreventError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'prevent',
    'JSX element should not start in a new line',
    range,
  );
}

new RuleTester(fixtureConfig).run('jsx-newline', {
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
      errors: [expectRequireError([4, 11, 4, 19])],
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
      errors: [expectRequireError([4, 11, 4, 52])],
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
      errors: [expectRequireError([4, 11, 4, 40])],
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
      errors: [expectRequireError([4, 11, 8, 13])],
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
      errors: [expectRequireError([7, 11, 7, 67])],
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
      errors: [expectAllowMultilinesError([7, 11, 7, 41])],
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
      errors: [expectAllowMultilinesError([7, 11, 7, 28])],
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
        expectRequireError([5, 13, 5, 30]),
        expectRequireError([7, 11, 10, 17]),
        expectRequireError([9, 13, 9, 26]),
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
      errors: [expectPreventError([5, 11, 5, 19])],
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
      errors: [expectPreventError([5, 11, 5, 52])],
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
      errors: [expectPreventError([5, 11, 5, 40])],
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
      errors: [expectPreventError([5, 11, 9, 13])],
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
        expectPreventError([6, 13, 6, 30]),
        expectPreventError([9, 11, 13, 17]),
        expectPreventError([12, 13, 12, 26]),
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

      output: `
        <>
          <Button>{data.label}</Button>
          Test

          <span>Should be in new line</span>
        </>
      `,
      errors: [expectRequireError([5, 11, 5, 45])],
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

      output: `
        <>
          <Button>{data.label}</Button>
          Test
          <span>Should be in new line</span>
        </>
      `,
      errors: [expectPreventError([6, 11, 6, 45])],
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
        expectAllowMultilinesError([5, 11, 8, 13]),
        expectAllowMultilinesError([9, 11, 9, 31]),
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
      errors: [expectAllowMultilinesError([4, 11, 8, 13])],
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
        expectPreventError([6, 13, 6, 30]),
        expectAllowMultilinesError([8, 11, 12, 17]),
        expectPreventError([11, 13, 11, 26]),
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

      output:
        '\n        const frag: DocumentFragment = (\n          <Fragment>\n            <sni-sequence-editor-tool\n              name="forward"\n              direction="forward"\n              type="control"\n              onClick={ () => this.onClickNavigate(\'forward\') }\n            />\n\n            <sni-sequence-editor-tool\n              name="rotate"\n              direction="left"\n              type="control"\n              onClick={ () => this.onClickNavigate(\'left\') }\n            />\n    \n            <sni-sequence-editor-tool\n              name="rotate"\n              direction="right"\n              type="control"\n              onClick={ (): void => this.onClickNavigate(\'right\') }\n            />\n    \n            <div className="sni-sequence-editor-control-panel__delete" data-name="delete" onClick={ this.onDeleteCommand } />\n    \n            {\n              ...Array.from(this.children)\n            }\n          </Fragment>\n        )\n      ',
      errors: [
        expectAllowMultilinesError([10, 13, 15, 15]),
        expectPreventError([26, 13, 28, 14]),
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
      errors: [expectRequireError([3, 3, 3, 11])],
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
      errors: [expectRequireError([3, 3, 3, 44])],
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
      errors: [expectRequireError([3, 3, 7, 5])],
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
        expectPreventError([4, 3, 4, 11]),
        expectPreventError([6, 3, 11, 12]),
        expectPreventError([10, 5, 10, 18]),
        expectPreventError([13, 3, 13, 44]),
        expectPreventError([15, 3, 15, 28]),
        expectPreventError([17, 3, 21, 5]),
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
        expectPreventError([4, 3, 4, 28]),
        expectAllowMultilinesError([5, 3, 9, 5]),
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
      errors: [expectPreventError([4, 3, 4, 28])],
    },
  ],
});
