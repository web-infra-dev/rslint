// TestStrictVoidReturnUpstream migrates the full valid/invalid suite from
// typescript-eslint's tests/rules/strict-void-return.test.ts 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific lock-in
// cases (Dimension 4 universal edge shapes, upstream branch lock-ins, real-user
// regressions) live in strict_void_return_extras_test.go.
package strict_void_return

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestStrictVoidReturnUpstream(t *testing.T) {
	allowReturnAny := map[string]interface{}{"allowReturnAny": true}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &StrictVoidReturnRule, []rule_tester.ValidTestCase{
		{Code: `
        declare function foo(cb: {}): void;
        foo(() => () => []);
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        type Void = void;
        foo((): Void => {
          return;
        });
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo((): ReturnType<typeof foo> => {
          return;
        });
      `},
		{Code: `
        declare function foo(cb: any): void;
        foo(() => () => []);
      `},
		{Code: `
        declare class Foo {
          constructor(cb: unknown): void;
        }
        new Foo(() => ({}));
      `},
		{Code: `
        declare function foo(cb: () => {}): void;
        foo(() => 1 as any);
      `, Options: allowReturnAny},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => {
          throw new Error('boom');
        });
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        declare function boom(): never;
        foo(() => boom());
        foo(boom);
      `},
		{Code: `
        declare const Foo: {
          new (cb: () => any): void;
        };
        new Foo(function () {
          return 1;
        });
      `},
		{Code: `
        declare const Foo: {
          new (cb: () => unknown): void;
        };
        new Foo(function () {
          return 1;
        });
      `},
		{Code: `
        declare const foo: {
          bar(cb1: () => unknown, cb2: () => void): void;
        };
        foo.bar(
          function () {
            return 1;
          },
          function () {
            return;
          },
        );
      `},
		{Code: `
        declare const Foo: {
          new (cb: () => string | void): void;
        };
        new Foo(() => {
          if (maybe) {
            return 'a';
          } else {
            return 'b';
          }
        });
      `},
		{Code: `
        declare function foo<Cb extends (...args: any[]) => void>(cb: Cb): void;
        foo(() => {
          console.log('a');
        });
      `},
		{Code: `
        declare function foo(cb: (() => void) | (() => string)): void;
        foo(() => {
          label: while (maybe) {
            for (let i = 0; i < 10; i++) {
              switch (i) {
                case 0:
                  continue;
                case 1:
                  return 'a';
              }
            }
          }
        });
      `},
		{Code: `
        declare function foo(cb: (() => void) | null): void;
        foo(null);
      `},
		{Code: `
        interface Cb {
          (): void;
          (): string;
        }
        declare const Foo: {
          new (cb: Cb): void;
        };
        new Foo(() => {
          do {
            try {
              throw 1;
            } catch {
              return 'a';
            }
          } while (maybe);
        });
      `},
		{Code: `
        declare const foo: ((cb: () => boolean) => void) | ((cb: () => void) => void);
        foo(() => false);
      `},
		{Code: `
        declare const foo: {
          (cb: () => boolean): void;
          (cb: () => void): void;
        };
        foo(function () {
          with ({}) {
            return false;
          }
        });
      `},
		{Code: `
        declare const Foo: {
          new (cb: () => void): void;
          (cb: () => unknown): void;
        };
        Foo(() => false);
      `},
		{Code: `
        declare const Foo: {
          new (cb: () => any): void;
          (cb: () => void): void;
        };
        new Foo(() => false);
      `},
		{Code: `
        declare function foo(cb: () => boolean): void;
        declare function foo(cb: () => void): void;
        foo(() => false);
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        declare function foo(cb: () => boolean): void;
        foo(() => false);
      `},
		{Code: `
        declare function foo(cb: () => Promise<void>): void;
        declare function foo(cb: () => void): void;
        foo(async () => {});
      `},
		{Code: `
        declare function foo(fn: () => void);
        declare function foo(fn: () => Promise<void>);

        foo(async () => {});
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => 1 as any);
      `, Options: allowReturnAny},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => {});
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        const cb = () => {};
        foo(cb);
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(function () {});
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(cb);
        function cb() {}
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => undefined);
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(function () {
          return;
        });
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(function () {
          return void 0;
        });
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => {
          return;
        });
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        declare function cb(): never;
        foo(cb);
      `},
		{Code: `
        declare class Foo {
          constructor(cb: () => void): any;
        }
        declare function cb(): void;
        new Foo(cb);
      `},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(cb);
        function cb() {
          throw new Error('boom');
        }
      `},
		{Code: `
        declare function foo(arg: string, cb: () => void): void;
        declare function cb(): undefined;
        foo('arg', cb);
      `},
		{Code: `
        declare function foo(cb?: () => void): void;
        foo();
      `},
		{Code: `
        declare class Foo {
          constructor(cb?: () => void): void;
        }
        declare function cb(): void;
        new Foo(cb);
      `},
		{Code: `
        declare function foo(...cbs: Array<() => void>): void;
        foo(
          () => {},
          () => void null,
          () => undefined,
        );
      `},
		{Code: `
        declare function foo(...cbs: Array<() => void>): void;
        declare const cbs: Array<() => void>;
        foo(...cbs);
      `},
		{Code: `
        declare function foo(...cbs: [() => any, () => void, (() => void)?]): void;
        foo(
          async () => {},
          () => void null,
          () => undefined,
        );
      `},
		{Code: `
        let cb;
        cb = async () => 10;
      `},
		{Code: `
        const foo: () => void = () => {};
      `},
		{Code: `
        declare function cb(): void;
        const foo: () => void = cb;
      `},
		{Code: `
        const foo: () => void = function () {
          throw new Error('boom');
        };
      `},
		{Code: `
        const foo: { (): string; (): void } = () => {
          return 'a';
        };
      `},
		{Code: `
        const foo: (() => void) | (() => number) = () => {
          return 1;
        };
      `},
		{Code: `
        type Foo = () => void;
        const foo: Foo = cb;
        function cb() {
          return void null;
        }
      `},
		{Code: `
        interface Foo {
          (): void;
        }
        const foo: Foo = cb;
        function cb() {
          return undefined;
        }
      `},
		{Code: `
        declare function cb(): void;
        declare let foo: () => void;
        foo = cb;
      `},
		{Code: `
        declare let foo: () => void;
        foo += () => 1;
      `},
		{Code: `
        declare function defaultCb(): object;
        declare let foo: { cb?: () => void };
        // default doesn't have to be void
        const { cb = defaultCb } = foo;
      `},
		{Code: `
        let foo: (() => void) | null = null;
        foo &&= null;
      `},
		{Code: `
        declare function cb(): void;
        let foo: (() => void) | boolean = false;
        foo ||= cb;
      `},
		{Code: `
        declare function Foo(props: { cb: () => void }): unknown;
        return <Foo cb={() => {}} />;
      `, Tsx: true},
		{Code: `
        declare function Foo(props: { cb: () => void }): unknown;
        return <Foo cb="() => {}" />;
      `, Tsx: true},
		{Code: `
        declare function Foo(props: { cb: () => void }): unknown;
        return <Foo cb={} />;
      `, Tsx: true},
		{Code: `
        declare function Foo(props: { cb: () => void }): unknown;
        return <Bar children=<Foo cb={() => {}} /> />;
      `, Tsx: true},
		{Code: `
        type Cb = () => void;
        declare function Foo(props: { cb: Cb; s: string }): unknown;
        return <Foo cb={function () {}} s="asd" />;
      `, Tsx: true},
		{Code: `
        type Cb = () => void;
        declare function Foo(props: { x: number; cb?: Cb }): unknown;
        return <Foo x={123} />;
      `, Tsx: true},
		{Code: `
        type Cb = (() => void) | (() => number);
        declare function Foo(props: { cb?: Cb }): unknown;
        return (
          <Foo
            cb={function (arg) {
              return 123;
            }}
          />
        );
      `, Tsx: true},
		{Code: `
        interface Props {
          cb: ((arg: unknown) => void) | boolean;
        }
        declare function Foo(props: Props): unknown;
        return <Foo cb />;
      `, Tsx: true},
		{Code: `
        interface Props {
          cb: (() => void) | (() => Promise<void>);
        }
        declare function Foo(props: Props): any;
        const _ = <Foo cb={async () => {}} />;
      `, Tsx: true},
		{Code: `
        interface Props {
          children: (arg: unknown) => void;
        }
        declare function Foo(props: Props): unknown;
        declare function cb(): void;
        return <Foo>{cb}</Foo>;
      `, Tsx: true},
		{Code: `
        declare function foo(cbs: { arg: number; cb: () => void }): void;
        foo({ arg: 1, cb: () => undefined });
      `},
		{Code: `
        declare let foo: { arg?: string; cb: () => void };
        foo = {
          cb: () => {
            return something;
          },
        };
      `, Options: allowReturnAny},
		{Code: `
        declare let foo: { cb: () => void };
        foo = {
          cb() {
            return something;
          },
        };
      `, Options: allowReturnAny},
		{Code: `
        declare let foo: { cb: () => void };
        foo = {
          // don't check this thing
          cb = () => 1,
        };
      `},
		{Code: `
        declare let foo: { cb: (n: number) => void };
        let method = 'cb';
        foo = {
          // don't check computed methods
          [method](n) {
            return n;
          },
        };
      `},
		{Code: `
        // no contextual type for object
        let foo = {
          cb(n) {
            return n;
          },
        };
      `},
		{Code: `
        interface Foo {
          fn(): void;
        }
        // no symbol for method cb
        let foo: Foo = {
          cb(n) {
            return n;
          },
        };
      `},
		{Code: `
        declare let foo: { cb: (() => void) | number };
        foo = {
          cb: 0,
        };
      `},
		{Code: `
        declare function cb(): void;
        const foo: Record<string, () => void> = {
          cb1: cb,
          cb2: cb,
        };
      `},
		{Code: `
        declare function cb(): string;
        const foo: Record<string, () => void> = {
          ...cb,
        };
      `},
		{Code: `
        declare function cb(): string;
        const foo: Record<string, () => void> = {
          ...cb,
          ...{},
        };
      `},
		{Code: `
        declare function cb(): void;
        const foo: Array<(() => void) | false> = [false, cb, () => cb()];
      `},
		{Code: `
        declare function cb(): void;
        const foo: [string, () => void, (() => void)?] = ['asd', cb];
      `},
		{Code: `
        const foo: { cbs: Array<() => void> | null } = {
          cbs: [
            function () {
              return undefined;
            },
            () => {
              return void 0;
            },
            null,
          ],
        };
      `},
		{Code: `
        const foo: { cb: () => void } = class {
          static cb = () => {};
        };
      `},
		{Code: `
        class Foo {
          foo;
        }
      `},
		{Code: `
        class Bar {
          foo() {}
        }
        class Foo extends Bar {
          foo();
        }
      `},
		{Code: `
        interface Bar {
          foo(): void;
        }
        class Foo implements Bar {
          get foo() {
            return new Date();
          }
          set foo() {
            return new Date('wtf');
          }
        }
      `},
		{Code: `
        class Foo {
          foo: () => void = () => undefined;
        }
      `},
		{Code: `
        class Bar {}
        class Foo extends Bar {
          foo = () => 1;
        }
      `},
		{Code: `
        class Foo extends Wtf {
          foo = () => 1;
        }
      `},
		{Code: `
        class Foo extends Wtf {
          [unknown] = () => 1;
        }
      `},
		{Code: `
        class Foo {
          cb = () => {
            console.log('hello');
          };
        }
        class Bar extends Foo {
          cb = () => {
            console.log('nara');
          };
        }
      `},
		{Code: `
        class Foo {
          cb1 = () => {};
        }
        class Bar extends Foo {
          cb2() {}
        }
        class Baz extends Bar {
          cb1 = () => {
            console.log('hello');
          };
          cb2() {
            console.log('nara');
          }
        }
      `},
		{Code: `
        class Foo {
          fn() {
            return 'a';
          }
          cb() {}
        }
        void class extends Foo {
          cb() {
            if (maybe) {
              console.log('hello');
            } else {
              console.log('nara');
            }
          }
        };
      `},
		{Code: `
        abstract class Foo {
          abstract cb(): void;
        }
        class Bar extends Foo {
          cb() {
            console.log('a');
          }
        }
      `},
		{Code: `
        class Bar implements Foo {
          cb = () => 1;
        }
      `},
		{Code: `
        interface Foo {
          cb: () => void;
        }
        class Bar implements Foo {
          cb = () => {};
        }
      `},
		{Code: `
        interface Foo {
          cb: () => void;
        }
        class Bar implements Foo {
          get cb() {
            return () => {};
          }
        }
      `},
		{Code: `
        interface Foo {
          cb(): void;
        }
        class Bar implements Foo {
          cb() {
            return undefined;
          }
        }
      `},
		{Code: `
        interface Foo1 {
          cb1(): void;
        }
        interface Foo2 {
          cb2: () => void;
        }
        class Bar implements Foo1, Foo2 {
          cb1() {}
          cb2() {}
        }
      `},
		{Code: `
        interface Foo1 {
          cb1(): void;
        }
        interface Foo2 extends Foo1 {
          cb2: () => void;
        }
        class Bar implements Foo2 {
          cb1() {}
          cb2() {}
        }
      `},
		{Code: `
        declare let foo: () => () => void;
        foo = () => () => {};
      `},
		{Code: `
        declare let foo: { f(): () => void };
        foo = {
          f() {
            return () => undefined;
          },
        };
        function cb() {}
      `},
		{Code: `
        declare let foo: { f(): () => void };
        foo.f = function () {
          return () => {};
        };
      `},
		{Code: `
        declare let foo: () => (() => void) | string;
        foo = () => 'asd' + 'zxc';
      `},
		{Code: `
        declare function foo(cb: () => () => void): void;
        foo(function () {
          return () => {};
        });
      `},
		{Code: `
        declare function foo(cb: (arg: string) => () => void): void;
        declare function foo(cb: (arg: number) => () => boolean): void;
        foo((arg: number) => {
          return cb;
        });
        function cb() {
          return true;
        }
      `},
		{Code: `
        declare function f<T extends void>(arg: T, cb: () => T): void;
        declare function f<T extends string>(arg: T, cb: () => T): void;

        f('test', () => 'test');
        f(undefined, () => {});
      `},
		{Code: `
        interface HookFunction<T extends void | Hook = void> {
          (fn: () => void): T;
          (fn: () => Promise<void>): T;
        }

        class Hook {}

        declare var beforeEach: HookFunction<Hook>;

        beforeEach(() => {});

        beforeEach(async () => {});
      `},
	}, []rule_tester.InvalidTestCase{
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => null);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 19},
		}},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => (((true))));
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 22},
		}},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => {
          if (maybe) {
            return (((1) + 1));
          }
        });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 13},
		}},
		{Code: `
        declare function foo(arg: number, cb: () => void): void;
        foo(0, () => 0);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 22},
		}},
		{Code: `
        declare function foo(cb?: { (): void }): void;
        foo(() => () => {});
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 19},
		}},
		{Code: `
        declare const obj: { foo(cb: () => void) } | null;
        obj?.foo(() => JSON.parse('{}'));
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 24},
		}},
		{Code: `
        ((cb: () => void) => cb())!(() => 1);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 2, Column: 43},
		}},
		{Code: `
        declare function foo(cb: { (): void }): void;
        declare function cb(): string;
        foo(cb);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 13},
		}},
		{Code: `
        type AnyFunc = (...args: unknown[]) => unknown;
        declare function foo<F extends AnyFunc>(cb: F): void;
        foo(async () => ({}));
        foo<() => void>(async () => ({}));
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 5, Column: 34},
		}},
		{Code: `
        function foo<T extends {}>(arg: T, cb: () => T);
        function foo(arg: null, cb: () => void);
        function foo(arg: any, cb: () => any) {}

        foo(null, () => Math.random());
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 25},
		}},
		{Code: `
        declare function foo<T extends {}>(arg: T, cb: () => T): void;
        declare function foo(arg: any, cb: () => void): void;

        foo(null, async () => {});
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 5, Column: 28},
		}},
		{Code: `
        declare function foo(cb: () => void): void;
        declare function foo(cb: () => any): void;
        foo(async () => {
          return Math.random();
        });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 4, Column: 22},
		}},
		{Code: `
        declare function f<T extends void>(arg: T, cb: () => T): void;

        f(undefined, () => 'test');
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 28},
		}},
		{Code: `
        declare function foo(cb: { (): void }): void;
        foo(cb);
        async function cb() {}
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 3, Column: 13},
		}},
		{Code: `
        declare function foo<Cb extends (...args: any[]) => void>(cb: Cb): void;
        foo(() => {
          console.log('a');
          return 1;
        });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 11},
		}},
		{Code: `
        declare function foo(cb: () => void): void;
        function bar<Cb extends () => number>(cb: Cb) {
          foo(cb);
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 15},
		}},
		{Code: `
        declare function foo(cb: { (): void }): void;
        const cb = () => dunno;
        foo!(cb);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 14},
		}},
		{Code: `
        declare const foo: {
          (arg: boolean, cb: () => void): void;
        };
        foo(false, () => Promise.resolve(undefined));
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 26},
		}},
		{Code: `
        declare const foo: {
          bar(cb1: () => any, cb2: () => void): void;
        };
        foo.bar(
          () => Promise.resolve(1),
          () => Promise.resolve(1),
        );
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 7, Column: 17},
		}},
		{Code: `
        declare const Foo: {
          new (cb: () => void): void;
        };
        new Foo(async () => {});
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 5, Column: 26},
		}},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => {
          label: while (maybe) {
            for (const i of [1, 2, 3]) {
              if (maybe) return null;
              else return null;
            }
          }
          return void 0;
        });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 26},
			{MessageId: "nonVoidReturn", Line: 7, Column: 20},
		}},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(() => {
          do {
            try {
              throw 1;
            } catch (e) {
              return null;
            } finally {
              console.log('finally');
            }
          } while (maybe);
        });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 8, Column: 15},
		}},
		{Code: `
        declare function foo(cb: () => void): void;
        foo(async () => {
          try {
            await Promise.resolve();
          } catch {
            console.error('fail');
          }
        });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 3, Column: 22},
		}},
		{Code: `
        declare const Foo: {
          new (cb: () => void): void;
          (cb: () => unknown): void;
        };
        new Foo(() => false);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 23},
		}},
		{Code: `
        declare const Foo: {
          new (cb: () => any): void;
          (cb: () => void): void;
        };
        Foo(() => false);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 19},
		}},
		{Code: `
        interface Cb {
          (arg: string): void;
          (arg: number): void;
        }
        declare function foo(cb: Cb): void;
        foo(cb);
        function cb() {
          return true;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 7, Column: 13},
		}},
		{Code: `
        declare function foo(
          cb: ((arg: number) => void) | ((arg: string) => void),
        ): void;
        foo(cb);
        function cb() {
          return 1 + 1;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 5, Column: 13},
		}},
		{Code: `
        declare function foo(cb: (() => void) | null): void;
        declare function cb(): boolean;
        foo(cb);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 13},
		}},
		{Code: `
        declare function foo(...cbs: Array<() => void>): void;
        foo(
          () => {},
          () => false,
          () => 0,
          () => '',
        );
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 17},
			{MessageId: "nonVoidReturn", Line: 6, Column: 17},
			{MessageId: "nonVoidReturn", Line: 7, Column: 17},
		}},
		{Code: `
        declare function foo(...cbs: [() => void, () => void, (() => void)?]): void;
        foo(
          () => {},
          () => Math.random(),
          () => (1).toString(),
        );
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 17},
			{MessageId: "nonVoidReturn", Line: 6, Column: 17},
		}},
		{Code: `
        interface Ev {}
        interface EvMap {
          DOMContentLoaded: Ev;
        }
        type EvListOrEvListObj = EvList | EvListObj;
        interface EvList {
          (evt: Event): void;
        }
        interface EvListObj {
          handleEvent(object: Ev): void;
        }
        interface Win {
          addEventListener<K extends keyof EvMap>(
            type: K,
            listener: (ev: EvMap[K]) => any,
          ): void;
          addEventListener(type: string, listener: EvListOrEvListObj): void;
        }
        declare const win: Win;
        win.addEventListener('DOMContentLoaded', ev => ev);
        win.addEventListener('custom', ev => ev);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 21, Column: 56},
			{MessageId: "nonVoidReturn", Line: 22, Column: 46},
		}},
		{Code: `
        declare function foo(x: null, cb: () => void): void;
        declare function foo(x: unknown, cb: () => any): void;
        foo({}, async () => {});
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 4, Column: 26},
		}},
		{Code: `
        const arr = [1, 2];
        arr.forEach(async x => {
          console.log(x);
        });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 3, Column: 29},
		}},
		{Code: `
        [1, 2].forEach(async x => console.log(x));
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 2, Column: 32},
		}},
		{Code: `
        const foo: () => void = () => false;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 2, Column: 39},
		}},
		{Code: `
        const { name }: () => void = function foo() {
          return false;
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 11},
		}},
		{Code: `
        declare const foo: Record<string, () => void>;
        foo['a' + 'b'] = () => true;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 32},
		}},
		{Code: `
        const foo: () => void = async () => Promise.resolve(true);
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 2, Column: 42},
		}},
		{Code: `const cb: () => void = (): Array<number> => [];`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 1, Column: 45},
		}},
		{Code: `
        const cb: () => void = (): Array<number> => {
          return [];
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 2, Column: 36},
		}},
		{Code: `const cb: () => void = function*foo() {}`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 1, Column: 24},
		}},
		{Code: `const cb: () => void = (): Promise<number> => Promise.resolve(1);`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 1, Column: 47},
		}},
		{Code: `
        const cb: () => void = async (): Promise<number> => {
          try {
            return Promise.resolve(1);
          } catch {}
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 2, Column: 58},
		}},
		{Code: `const cb: () => void = async (): Promise<number> => Promise.resolve(1);`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 1, Column: 50},
		}},
		{Code: `
        const foo: () => void = async () => {
          try {
            return 1;
          } catch {}
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 2, Column: 42},
		}},
		{Code: `
        const foo: () => void = async (): Promise<void> => {
          try {
            await Promise.resolve();
          } finally {
          }
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 2, Column: 57},
		}},
		{Code: `
        const foo: () => void = async () => {
          try {
            await Promise.resolve();
          } catch (err) {
            console.error(err);
          }
          console.log('ok');
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 2, Column: 42},
		}},
		{Code: `const foo: () => void = (): number => {};`, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 1, Column: 29},
		}},
		{Code: `
        declare function cb(): boolean;
        const foo: () => void = cb;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 3, Column: 33},
		}},
		{Code: `
        const foo: () => void = function () {
          if (maybe) {
            return null;
          } else {
            return null;
          }
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 13},
			{MessageId: "nonVoidReturn", Line: 6, Column: 13},
		}},
		{Code: `
        const foo: () => void = function () {
          if (maybe) {
            console.log('elo');
            return { [1]: Math.random() };
          }
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 13},
		}},
		{Code: `
        const foo: { (arg: number): void; (arg: string): void } = arg => {
          console.log('foo');
          switch (typeof arg) {
            case 'number':
              return 0;
            case 'string':
              return '';
          }
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 15},
			{MessageId: "nonVoidReturn", Line: 8, Column: 15},
		}},
		{Code: `
        const foo: ((arg: number) => void) | ((arg: string) => void) = async () => {
          return 1;
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 2, Column: 81},
		}},
		{Code: `
        type Foo = () => void;
        const foo: Foo = cb;
        function cb() {
          return [1, 2, 3];
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 3, Column: 26},
		}},
		{Code: `
        interface Foo {
          (): void;
        }
        const foo: Foo = cb;
        function cb() {
          return { a: 1 };
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 5, Column: 26},
		}},
		{Code: `
        declare function cb(): unknown;
        declare let foo: () => void;
        foo = cb;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 15},
		}},
		{Code: `
        declare let foo: { arg?: string; cb?: () => void };
        foo.cb = () => {
          return 'hello';
          console.log('hello');
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 11},
		}},
		{Code: `
        declare function cb(): unknown;
        let foo: (() => void) | null = null;
        foo ??= cb;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 17},
		}},
		{Code: `
        declare function cb(): unknown;
        let foo: (() => void) | boolean = false;
        foo ||= cb;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 17},
		}},
		{Code: `
        declare function cb(): unknown;
        let foo: (() => void) | boolean = false;
        foo &&= cb;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 17},
		}},
		{Code: `
        declare function Foo(props: { cb: () => void }): unknown;
        return <Foo cb={() => 1} />;
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 31},
		}},
		{Code: `
        declare function Foo(props: { cb: () => void }): unknown;
        declare function getNull(): null;
        return (
          <Foo
            cb={() => {
              if (maybe) return Math.random();
              else return getNull();
            }}
          />
        );
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 7, Column: 26},
			{MessageId: "nonVoidReturn", Line: 8, Column: 20},
		}},
		{Code: `
        type Cb = () => void;
        declare function Foo(props: { cb: Cb; s: string }): unknown;
        return <Foo cb={async function () {}} s="!@#jp2gmd" />;
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 4, Column: 25},
		}},
		{Code: `
        type Cb = () => void;
        declare function Foo(props: { n: number; cb?: Cb }): unknown;
        return <Foo n={2137} cb={function* () {}} />;
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 34},
		}},
		{Code: `
        type Cb = ((arg: string) => void) | ((arg: number) => void);
        declare function Foo(props: { cb?: Cb }): unknown;
        return (
          <Foo
            cb={async function* (arg) {
              await arg;
              yield arg;
            }}
          />
        );
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 6, Column: 17},
		}},
		{Code: `
        interface Props {
          cb: ((arg: unknown) => void) | boolean;
        }
        declare function Foo(props: Props): unknown;
        return <Foo cb={x => x} />;
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 30},
		}},
		{Code: `
        type EventHandler<E> = { bivarianceHack(event: E): void }['bivarianceHack'];
        interface ButtonProps {
          onClick?: EventHandler<unknown> | undefined;
        }
        declare function Button(props: ButtonProps): unknown;
        function App() {
          return <Button onClick={x => x} />;
        }
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 8, Column: 40},
		}},
		{Code: `
        declare function foo(cbs: { arg: number; cb: () => void }): void;
        foo({ arg: 1, cb: () => 1 });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 33},
		}},
		{Code: `
        declare let foo: { arg?: string; cb: () => void };
        foo = {
          cb: () => {
            let x = 'hello';
            return x;
          },
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 13},
		}},
		{Code: `
        declare let foo: { cb: (n: number) => void };
        foo = {
          cb(n) {
            return n;
          },
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 13},
		}},
		{Code: `
        declare let foo: { 1234: (n: number) => void };
        foo = {
          1234(n) {
            return n;
          },
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 13},
		}},
		{Code: `
        declare let foo: { '1e+21': () => void };
        foo = {
          1_000_000_000_000_000_000_000: () => 1,
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 48},
		}},
		{Code: `
        declare let foo: { cb: (() => void) | number };
        foo = {
          cb: async () => {
            if (maybe) {
              return 'asd';
            }
          },
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 4, Column: 11},
		}},
		{Code: `
        declare function cb(): number;
        const foo: Record<string, () => void> = {
          cb1: cb,
          cb2: cb,
          ...cb,
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 16},
			{MessageId: "nonVoidFunc", Line: 5, Column: 16},
		}},
		{Code: `
        declare function cb(): number;
        const foo: Array<(() => void) | false> = [false, cb, () => cb()];
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 3, Column: 58},
			{MessageId: "nonVoidReturn", Line: 3, Column: 68},
		}},
		{Code: `
        declare function cb(): number;
        const foo: [string, () => void, (() => void)?] = ['asd', cb];
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 3, Column: 66},
		}},
		{Code: `
        const foo: { cbs: Array<() => void> | null } = {
          cbs: [
            function* () {
              yield 1;
            },
            async () => {
              await 1;
            },
            null,
          ],
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 13},
			{MessageId: "asyncFunc", Line: 7, Column: 22},
		}},
		{Code: `
        const foo: { cb: () => void } = class {
          static cb = () => ({});
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 30},
		}},
		{Code: `
        class Foo {
          foo: () => void = () => [];
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 35},
		}},
		{Code: `
        class Foo {
          static foo: () => void = Math.random;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 3, Column: 36},
		}},
		{Code: `
        class Foo {
          cb = () => {};
        }
        class Bar extends Foo {
          cb = Math.random;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 6, Column: 16},
		}},
		{Code: `
        const foo = () =>
          class {
            cb = () => {};
          };
        class Bar extends foo() {
          cb = Math.random;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 7, Column: 16},
		}},
		{Code: `
        class Foo {
          cb() {
            console.log('hello');
          }
        }
        const method = 'cb' as const;
        class Bar extends Foo {
          [method]() {
            return 'nara';
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 10, Column: 13},
		}},
		{Code: `
        class Bar {
          foo() {}
        }
        class Foo extends Bar {
          get foo() {
            return () => 1;
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 7, Column: 13},
		}},
		{Code: `
        class Foo {
          cb() {}
        }
        void class extends Foo {
          cb() {
            return Math.random();
          }
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 7, Column: 13},
		}},
		{Code: `
        class Foo {
          cb1 = () => {};
        }
        class Bar extends Foo {
          cb2() {}
        }
        class Baz extends Bar {
          cb1 = () => Math.random();
          cb2() {
            return Math.random();
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 9, Column: 23},
			{MessageId: "nonVoidReturn", Line: 11, Column: 13},
		}},
		{Code: `
        declare function f(): Promise<void>;
        interface Foo {
          cb: () => void;
        }
        class Bar {
          cb = () => {};
        }
        class Baz extends Bar implements Foo {
          cb: () => void = f;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 10, Column: 28},
		}},
		{Code: `
        class Foo {
          fn() {
            return 'a';
          }
          cb() {}
        }
        class Bar extends Foo {
          cb() {
            if (maybe) {
              return Promise.resolve('hello');
            } else {
              return Promise.resolve('nara');
            }
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 11, Column: 15},
			{MessageId: "nonVoidReturn", Line: 13, Column: 15},
		}},
		{Code: `
        abstract class Foo {
          abstract cb(): void;
        }
        class Bar extends Foo {
          async cb() {}
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 6, Column: 11},
		}},
		{Code: `
        class Foo {
          fn() {
            return 'a';
          }
          cb() {}
        }
        class Bar extends Foo {
          *cb() {}
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 9, Column: 11},
		}},
		{Code: `
        interface Foo {
          cb: () => void;
        }
        class Bar implements Foo {
          cb = Math.random;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 6, Column: 16},
		}},
		{Code: `
        const o = { cb() {} };
        type O = typeof o;
        class Bar implements O {
          cb = Math.random;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 5, Column: 16},
		}},
		{Code: `
        class Foo {
          cb() {}
        }
        class Bar extends Foo {
          async*cb() {}
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 6, Column: 11},
		}},
		{Code: `
        interface Foo {
          cb(): void;
        }
        class Bar implements Foo {
          async cb(): Promise<string> {
            return Promise.resolve('hello');
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 6, Column: 11},
		}},
		{Code: `
        interface Foo {
          cb(): void;
        }
        class Bar implements Foo {
          async cb() {
            try {
              return { a: ['asdf', 1234] };
            } catch {
              console.error('error');
            }
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 6, Column: 11},
		}},
		{Code: `
        interface Foo {
          cb(): void;
        }
        class Bar implements Foo {
          cb() {
            if (maybe) {
              return Promise.resolve(1);
            } else {
              return;
            }
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 8, Column: 15},
		}},
		{Code: `
        interface Foo1 {
          cb1(): void;
        }
        interface Foo2 {
          cb2: () => void;
        }
        class Bar implements Foo1, Foo2 {
          async cb1() {}
          async *cb2() {}
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 9, Column: 11},
			{MessageId: "nonVoidFunc", Line: 10, Column: 11},
		}},
		{Code: `
        interface Foo1 {
          cb1(): void;
        }
        interface Foo2 {
          cb2: () => void;
        }
        class Baz {
          cb3() {}
        }
        class Bar extends Baz implements Foo1, Foo2 {
          async cb1() {}
          async *cb2() {}
          cb3() {
            return Math.random();
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 12, Column: 11},
			{MessageId: "nonVoidFunc", Line: 13, Column: 11},
			{MessageId: "nonVoidReturn", Line: 15, Column: 13},
		}},
		{Code: `
        class A extends class {
          cb() {}
        } {
          cb() {
            return Math.random();
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 13},
		}},
		{Code: `
        class A extends class B {
          cb() {}
        } {
          cb() {
            return Math.random();
          }
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 13},
		}},
		{Code: `
        interface Foo1 {
          cb1(): void;
        }
        interface Foo2 extends Foo1 {
          cb2: () => void;
        }
        class Bar implements Foo2 {
          async cb1() {}
          async *cb2() {}
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 9, Column: 11},
			{MessageId: "nonVoidFunc", Line: 10, Column: 11},
		}},
		{Code: `
        declare let foo: () => () => void;
        foo = () => () => 1 + 1;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 27},
		}},
		{Code: `
        declare let foo: () => () => void;
        foo = () => () => Math.random();
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 3, Column: 27},
		}},
		{Code: `
        declare let foo: () => () => void;
        declare const cb: () => null | false;
        foo = () => cb;
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 4, Column: 21},
		}},
		{Code: `
        declare let foo: { f(): () => void };
        foo = {
          f() {
            return () => cb;
          },
        };
        function cb() {}
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 26},
		}},
		{Code: `
        declare let foo: { f(): () => void };
        foo.f = function () {
          return () => {
            return null;
          };
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 5, Column: 13},
		}},
		{Code: `
        declare let foo: () => (() => void) | string;
        foo = () => () => {
          return 'asd' + 'zxc';
        };
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 4, Column: 11},
		}},
		{Code: `
        declare function foo(cb: () => () => void): void;
        foo(function () {
          return async () => {};
        });
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "asyncFunc", Line: 4, Column: 27},
		}},
		{Code: `
        declare function foo(cb: () => () => void): void;
        foo(() => () => {
          if (n == 1) {
            console.log('asd')
            return [1].map(x => x)
          }
          if (n == 2) {
            console.log('asd')
            return -Math.random()
          }
          if (n == 3) {
            console.log('asd')
            return ` + "`x`" + `.toUpperCase()
          }
          return <i>{Math.random()}</i>
        });
      `, Tsx: true, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidReturn", Line: 6, Column: 13},
			{MessageId: "nonVoidReturn", Line: 10, Column: 13},
			{MessageId: "nonVoidReturn", Line: 14, Column: 13},
			{MessageId: "nonVoidReturn", Line: 16, Column: 11},
		}},
		{Code: `
        declare function foo(cb: (arg: string) => () => void): void;
        declare function foo(cb: (arg: number) => () => boolean): void;
        foo((arg: string) => {
          return cb;
        });
        async function* cb() {
          yield true;
        }
      `, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "nonVoidFunc", Line: 5, Column: 18},
		}},
	})
}
