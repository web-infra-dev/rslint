// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import {
  RuleTester,
  type ValidTestCase,
  type InvalidTestCase,
} from '../rule-tester';

new RuleTester().run('prefer-optional-catch-binding', {} as never, {
  valid: [
    ...[
      'try {} catch {}',
      'try {} catch {\n\terror\n}',
      'try {} catch(used) {\n\tconsole.error(used);\n}',
      'try {} catch(usedInADeeperScope) {\n\tfunction foo() {\n\t\tfunction bar() {\n\t\t\tconsole.error(usedInADeeperScope);\n\t\t}\n\t}\n}',
      'try {} catch ({message}) {alert(message)}',
      'try {} catch ({cause: {message}}) {alert(message)}',
      'try {} catch({nonExistsProperty = thisWillExecute()}) {}',
      '// ✅\ntry {\n\t// do something\n} catch {\n\t// ignore error\n}\n',
      '// ✅\ntry {\n\tawait fetch(url);\n} catch {\n\t// ignore fetch errors\n}\n',
      '// ✅\ntry {\n\tdoSomething();\n} catch (error) {\n\t// error is actually used\n\tconsole.log(error.message);\n}\n',
    ].map((code): ValidTestCase => ({
      code,
      ...{
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
  ],
  invalid: [
    ...['try {} catch (_) {}'].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Remove unused catch binding `_`.',
            messageId: 'with-name',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['try {} catch (foo) {\n\tfunction bar(foo) {}\n}'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: 'Remove unused catch binding `foo`.',
              messageId: 'with-name',
            },
          ],
          filename: 'src/virtual.js',
          languageOptions: {
            globals: {
              global: 'readonly',
              self: 'readonly',
              window: 'readonly',
            },
            sourceType: 'module',
          },
        },
      }),
    ),
    ...[
      'try {} catch (outer) {\n\ttry {} catch (inner) {\n\t}\n}\ntry {\n\ttry {} catch (inTry) {\n\t}\n} catch (another) {\n\ttry {} catch (inCatch) {\n\t}\n} finally {\n\ttry {} catch (inFinally) {\n\t}\n}',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Remove unused catch binding `outer`.',
            messageId: 'with-name',
          },
          {
            message: 'Remove unused catch binding `inner`.',
            messageId: 'with-name',
          },
          {
            message: 'Remove unused catch binding `inTry`.',
            messageId: 'with-name',
          },
          {
            message: 'Remove unused catch binding `another`.',
            messageId: 'with-name',
          },
          {
            message: 'Remove unused catch binding `inCatch`.',
            messageId: 'with-name',
          },
          {
            message: 'Remove unused catch binding `inFinally`.',
            messageId: 'with-name',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['try {} catch (theRealErrorName) {}'].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Remove unused catch binding `theRealErrorName`.',
            messageId: 'with-name',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      '/* comment */\ntry {\n\t/* comment */\n\t// comment\n} catch (\n\t/* comment */\n\t// comment\n\tunused\n\t/* comment */\n\t// comment\n) {\n\t/* comment */\n\t// comment\n}\n/* comment */',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Remove unused catch binding `unused`.',
            messageId: 'with-name',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      'try    {    } catch    (e)  \n  \t  {    }',
      'try {} catch(e) {}',
      'try {} catch (e){}',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Remove unused catch binding `e`.',
            messageId: 'with-name',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      'try {} catch ({}) {}',
      'try {} catch ({/* inner comment */ message}) {}',
      'try {} catch ({message}) {}',
      'try {} catch ({message: notUsedMessage}) {}',
      'try {} catch ({cause: {message}}) {}',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Remove unused catch binding.',
            messageId: 'without-name',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      '// ❌\ntry {\n\t// do something\n} catch (notUsedError) {\n\t// ignore error\n}\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Remove unused catch binding `notUsedError`.',
            messageId: 'with-name',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...[
      '// ❌\ntry {\n\tawait fetch(url);\n} catch (error) {\n\t// error is not used, just continue\n}\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Remove unused catch binding `error`.',
            messageId: 'with-name',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
  ],
});
