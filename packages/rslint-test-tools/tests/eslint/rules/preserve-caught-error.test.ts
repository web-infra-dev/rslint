import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('preserve-caught-error', {
  valid: [
    {
      code: 'try {\n        throw new Error("Original error");\n    } catch (error) {\n        throw new Error("Failed to perform error prone operations", { cause: error });\n    }',
    },
    {
      code: 'try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error("Something failed", { \'cause\': error });\n\t}',
    },
    {
      code: 'try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error("Something failed", { "cause": error });\n\t}',
    },
    {
      code: 'try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error("Something failed", { [\'cause\']: error });\n\t}',
    },
    {
      code: 'try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error("Something failed", { ["cause"]: error });\n\t}',
    },
    {
      code: 'try {\n\t\tdoSomething();\n\t} catch (error) {\n\t\tthrow new Error("Something failed", { [`cause`]: error });\n\t}',
    },
    {
      code: 'try {\n        doSomething();\n    } catch (e) {\n        console.error(e);\n    }',
    },
    {
      code: 'try {\n        doSomething();\n    } catch (err) {\n        throw new Error("Failed", { cause: err, extra: 42 });\n    }',
    },
    {
      code: 'try {\n        doSomething();\n    } catch (error) {\n        switch (error.code) {\n            case "A":\n                throw new Error("Type A", { cause: error });\n            case "B":\n                throw new Error("Type B", { cause: error });\n            default:\n                throw new Error("Other", { cause: error });\n        }\n    }',
    },
    {
      code: 'try {\n\t\t// ...\n\t} catch (err) {\n\t\tconst opts = { cause: err }\n\t\tthrow new Error("msg", { ...opts });\n\t}\n\t',
    },
    {
      code: 'try {\n\t} catch (error) {\n\t\tfoo = {\n\t\t\tbar() {\n\t\t\t\tthrow new Error();\n\t\t\t}\n\t\t};\n\t}',
    },
    {
      code: 'try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tconst args = [];\n\t\t\t\tthrow new Error(...args);\n\t\t}',
    },
    {
      code: 'import { Error } from "./my-custom-error.js";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow Error("Failed to perform error prone operations");\n\t\t\t}',
    },
    {
      code: 'try {\n\t\tdoSomething();\n\t} catch {\n\t\tthrow new Error("Something went wrong");\n\t}',
      options: { requireCatchParameter: false },
    },
    {
      code: 'try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tthrow new Error("Something failed", { cause: anotherError, cause: error });\n\t\t}',
    },
    {
      code: 'try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tthrow new Error("Something failed", { "cause": anotherError, "cause": error });\n\t\t}',
    },
    {
      code: 'try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tthrow new Error("Something failed", { cause: anotherError, "cause": error });\n\t\t}',
    },
    {
      code: '\n\t\t\tconst errors = { AppError: class AppError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new errors.AppError("Something failed", { cause: error });\n\t\t\t}',
      options: { errorClassNames: ['AppError'] },
    },
    {
      code: '\n\t\t\tconst lib = { APIError: class APIError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new lib.APIError("Something failed", { cause: error });\n\t\t\t}',
      options: { errorClassNames: ['APIError'] },
    },
    {
      code: '\n\t\t\tclass CustomApplicationError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new CustomApplicationError("Cause not provided", { cause: err });\n\t\t\t}',
      options: { errorClassNames: ['CustomApplicationError'] },
    },
    {
      code: '\n\t\t\tclass APIError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new APIError("API failed", { cause: error });\n\t\t\t}',
      options: { errorClassNames: ['APIError'] },
    },
    {
      code: '\n\t\t\tclass CustomError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new CustomError("No cause");\n\t\t\t}',
      options: { errorClassNames: [] },
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError("failed", someData, { cause: err });\n\t\t\t}',
      options: { errorClassNames: [{ name: 'MyError', argumentPosition: 3 }] },
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new SimpleError({ cause: err });\n\t\t\t}',
      options: {
        errorClassNames: [{ name: 'SimpleError', argumentPosition: 1 }],
      },
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AppError("Message", {}, { cause: err });\n\t\t\t}',
      options: {
        errorClassNames: [
          { name: 'AppError', argumentPosition: 2 },
          { name: 'AppError', argumentPosition: 3 },
        ],
      },
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new context.Error("failed", someData, { cause: err });\n\t\t\t}',
      options: { errorClassNames: [{ name: 'Error', argumentPosition: 3 }] },
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new Error("failed", { cause: err });\n\t\t\t}',
      options: { errorClassNames: [{ name: 'Error', argumentPosition: 3 }] },
    },
    {
      code: 'import { AggregateError } from "some-module";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AggregateError({ cause: err });\n\t\t\t}',
      options: {
        errorClassNames: [{ name: 'AggregateError', argumentPosition: 1 }],
      },
    },
  ],
  invalid: [
    {
      code: 'try {\n            doSomething();\n        } catch (err) {\n            throw new Error("Something failed");\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (err) {\n            const unrelated = new Error("other");\n            throw new Error("Something failed", { cause: unrelated });\n        }',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (err) {\n            const unrelated = new Error("other");\n            throw new Error("Something failed", { "cause": unrelated });\n        }',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (err) {\n            const e = err;\n            throw new Error("Failed", { cause: e });\n        }',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (error) {\n            throw new Error("Failed", { cause: error.message });\n        }',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (error) {\n            if (shouldThrow) {\n                while (true) {\n                    if (Math.random() > 0.5) {\n                        throw new Error("Failed without cause");\n                    }\n                }\n            }\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (error) {\n            switch (error.code) {\n                case "A":\n                    throw new Error("Type A");\n                case "B":\n                    throw new Error("Type B", { cause: error });\n                default:\n                    throw new Error("Other", { cause: error });\n            }\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (error) {\n            throw new Error(`The certificate key "${chalk.yellow(keyFile)}" is invalid.\n${err.message}`);\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (error) {\n            const errorMessage = "Operation failed";\n            throw new Error(errorMessage);\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (error) {\n            const errorMessage = "Operation failed";\n            throw new Error(errorMessage, { existingOption: true, complexOption: { moreOptions: {} } });\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (err) {\n            if (err.code === "A") {\n                throw new Error("Type A");\n            }\n            throw new TypeError("Fallback error");\n        }',
      errors: [{ messageId: 'missingCause' }, { messageId: 'missingCause' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (err) {\n            throw Error("Something failed");\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n        } catch (err) {\n            my_label:\n            throw new Error("Failed without cause");\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n        } catch (err) {\n            {\n                throw new Error("Something went wrong");\n            }\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n        } catch (err) {\n            {\n                throw new Error();\n            }\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n        } catch (err) {\n            {\n                throw new AggregateError([], "Lorem ipsum");\n            }\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n        } catch (err) {\n            {\n                throw new AggregateError();\n            }\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n        } catch (err) {\n            {\n                throw new AggregateError([]);\n            }\n        }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n\t\t\tdoSomething();\n\t\t} catch {\n\t\t\tthrow new Error("Something went wrong");\n\t\t}',
      options: { requireCatchParameter: true },
      errors: [{ messageId: 'missingCatchErrorParam' }],
    },
    {
      code: 'try {\n            doSomething();\n        } catch (err) {\n            throw new Error("Something failed", { cause });\n        }',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {\n\t\t\t\tdoSomething();\n\t\t\t} catch ({ message }) {\n\t\t\t\tthrow new Error(message);\n\t\t\t}',
      errors: [{ messageId: 'partiallyLostError' }],
    },
    {
      code: 'try {\n\t\t\t\tdoSomethingElse();\n\t\t\t} catch ({ ...error }) {\n\t\t\t\tthrow new Error(error.message);\n\t\t\t}',
      errors: [{ messageId: 'partiallyLostError' }],
    },
    {
      code: 'try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tif (whatever) {\n\t\t\t\t\tconst error = anotherError;\n\t\t\t\t\tthrow new Error("Something went wrong", { cause: error });\n\t\t\t\t}\n\t\t\t}',
      errors: [{ messageId: 'caughtErrorShadowed' }],
    },
    {
      code: 'try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new Error(\n\t\t\t\t\t"Something went wrong" // some comments\n\t\t\t\t);\n\t\t\t}',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new Error("Something failed", {});\n\t\t\t}',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n\t\t\tdoSomething();\n\t\t} catch (error) {\n\t\t\tconst cause = "desc";\n\t\t\tthrow new Error("Something failed", { [cause]: "Some error" });\n\t\t}',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {\n\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\tthrow new Error("Something failed", { cause() { /* do something */ }  });\n\t\t\t}',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {} catch (error) {\n\t\t\t\tthrow new Error("Something failed", { cause: error, cause: anotherError });\n\t\t\t}',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {} catch (error) {\n\t\t\t\tthrow new Error("Something failed", { cause: error, "cause": anotherError });\n\t\t\t}',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {} catch (error) {\n\t\t\t\tthrow new Error("Something failed", { get cause() { } });\n\t\t\t}',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {} catch (error) {\n\t\t\t\tthrow new Error("Something failed", { set cause(value) { } });\n\t\t\t}',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {} catch (error) {\n\t\t\t\tthrow new Error("Something failed", {\n\t\t\t\t\tget cause() { return error; },\n\t\t\t\t\tset cause(value) { error = value; },\n\t\t\t\t});\n\t\t\t}',
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try { doSomething(); } catch (err) { throw new Error(("Something failed")); }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try { doSomething(); } catch (err) { throw new Error(("Something failed"),); }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try { doSomething(); } catch (err) { throw new AggregateError((errors)); }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try { doSomething(); } catch (err) { throw new AggregateError(errors, ("message")); }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try { doSomething(); } catch (err) { throw new CustomError((foo)); }',
      options: { errorClassNames: ['CustomError'] },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: '\n\t\t\tclass CustomApplicationError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new CustomApplicationError("Cause not provided");\n\t\t\t}',
      options: { errorClassNames: ['CustomApplicationError'] },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: '\n\t\t\tclass APIError extends Error {}\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new APIError("API failed", { cause: wrong });\n\t\t\t}',
      options: { errorClassNames: ['APIError'] },
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: '\n\t\t\tconst errors = { AppError: class AppError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new errors.AppError("Something failed");\n\t\t\t}',
      options: { errorClassNames: ['AppError'] },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: '\n\t\t\tconst lib = { APIError: class APIError extends Error {} };\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (error) {\n\t\t\t\tthrow new lib.APIError("Something failed", { cause: wrong });\n\t\t\t}',
      options: { errorClassNames: ['APIError'] },
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError("failed", someData);\n\t\t\t}',
      options: { errorClassNames: [{ name: 'MyError', argumentPosition: 3 }] },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError("failed", someData, { cause: wrong });\n\t\t\t}',
      options: { errorClassNames: [{ name: 'MyError', argumentPosition: 3 }] },
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new SimpleError();\n\t\t\t}',
      options: {
        errorClassNames: [{ name: 'SimpleError', argumentPosition: 1 }],
      },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new MyError();\n\t\t\t}',
      options: { errorClassNames: ['MyError'] },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new context.Error("failed", someData);\n\t\t\t}',
      options: { errorClassNames: [{ name: 'Error', argumentPosition: 3 }] },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: '\n\t\t\ttry {\n\t\t\t    doSomething();\n\t\t\t} catch (err) {\n\t\t\t    throw new Error("failed");\n\t\t\t}',
      options: { errorClassNames: [{ name: 'Error', argumentPosition: 3 }] },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'import { AggregateError } from "some-module";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AggregateError();\n\t\t\t}',
      options: {
        errorClassNames: [{ name: 'AggregateError', argumentPosition: 1 }],
      },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'import { AggregateError } from "some-module";\n\t\t\ttry {\n\t\t\t\tdoSomething();\n\t\t\t} catch (err) {\n\t\t\t\tthrow new AggregateError({ cause: wrong });\n\t\t\t}',
      options: {
        errorClassNames: [{ name: 'AggregateError', argumentPosition: 1 }],
      },
      errors: [{ messageId: 'incorrectCause' }],
    },
    {
      code: 'try {} catch (err) { throw new Error; }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {} catch (err) { throw new AggregateError; }',
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {} catch (err) { throw new CustomError; }',
      options: {
        errorClassNames: [{ name: 'CustomError', argumentPosition: 1 }],
      },
      errors: [{ messageId: 'missingCause' }],
    },
    {
      code: 'try {} catch (err) { throw new (Error); }',
      errors: [{ messageId: 'missingCause' }],
    },
  ],
});
