# TypeScript ESLint Rule Porting Status

## Summary

- Total TypeScript ESLint rules: 131
- Rules ported to RSLint Go: 52 (39.69%)
- Rules remaining to port: 79 (60.31%)

## Rules to Port

### High Priority Rules (Common/Important)

- `no-non-null-assertion`: Disallows using `!` postfix operator for non-null assertions
- `no-non-null-asserted-optional-chain`: Prevents using non-null assertions after optional chains
- `prefer-optional-chain`: Enforces using optional chaining instead of logical chaining
- `explicit-function-return-type`: Requires explicit return types on functions and class methods
- `no-unnecessary-condition`: Prevents conditionals where the type is always truthy or always falsy
- `no-unsafe-declaration-merging`: Disallows unsafe declaration merging
- `prefer-ts-expect-error`: Enforces using `@ts-expect-error` over `@ts-ignore`

### Medium Priority Rules

- `ban-ts-comment`: Bans specific TypeScript comment directives
- `ban-tslint-comment`: Bans `// tslint:<rule-flag>` comments
- `consistent-type-imports`: Enforces consistent usage of type imports
- `consistent-type-exports`: Enforces consistent usage of type exports
- `consistent-type-assertions`: Enforces consistent usage of type assertions
- `member-ordering`: Requires a consistent member declaration order
- `method-signature-style`: Enforces using a particular method signature syntax
- `naming-convention`: Enforces naming conventions for everything
- `no-empty-object-type`: Disallows {} (empty object) as a type
- `no-inferrable-types`: Disallows explicit type declarations for variables or parameters
- `no-invalid-void-type`: Disallows void type outside of generic or return types
- `no-misused-new`: Enforces valid definition of new and constructor
- `no-shadow`: Disallow variable declarations from shadowing variables

### Remaining Rules

- `class-methods-use-this`: Enforces that class methods utilize this
- `consistent-generic-constructors`: Enforces consistent usage of type parameters in constructors
- `consistent-indexed-object-style`: Enforces record type or index signature for objects with numeric or string index
- `consistent-return`: Requires return statements to either always or never specify values
- `consistent-type-definitions`: Consistent use of either interface or type
- `default-param-last`: Enforces default parameters to be last
- `dot-notation`: Enforces dot notation whenever possible
- `explicit-member-accessibility`: Requires explicit accessibility modifiers on class properties and methods
- `explicit-module-boundary-types`: Requires explicit return and argument types on exported functions
- `init-declarations`: Requires or disallows initialization in variable declarations
- `max-params`: Enforces maximum number of parameters in function definitions
- `no-array-constructor`: Disallows Array constructor
- `no-confusing-non-null-assertion`: Disallows confusing non-null assertion in expressions
- `no-deprecated`: Bans deprecated APIs
- `no-dupe-class-members`: Disallows duplicate class members
- `no-duplicate-enum-values`: Disallows duplicate enum member values
- `no-dynamic-delete`: Disallows delete operator with computed key expressions
- `no-extra-non-null-assertion`: Disallows extra non-null assertion
- `no-extraneous-class`: Forbids empty classes
- `no-import-type-side-effects`: Enforces using `import type` for imports that are only used in types
- `no-invalid-this`: Disallows this keywords outside of classes or class-like objects
- `no-loop-func`: Disallows function declarations that contain unsafe references inside loop statements
- `no-loss-of-precision`: Disallows literal numbers that lose precision
- `no-magic-numbers`: Disallows magic numbers
- `no-non-null-asserted-nullish-coalescing`: Disallows non-null assertions in the left operand of nullish coalescing
- `no-redeclare`: Disallows variable redeclarations
- `no-restricted-imports`: Disallows specified modules from being imported
- `no-restricted-types`: Disallows specific types from being used
- `no-this-alias`: Disallows aliasing this
- `no-type-alias`: Disallows type aliases
- `no-unnecessary-parameter-property-assignment`: Disallows unnecessary parameter property assignment
- `no-unnecessary-qualifier`: Disallows unnecessary namespace qualifiers
- `no-unnecessary-type-constraint`: Disallows unnecessary constraints on generic types
- `no-unnecessary-type-conversion`: Disallows unnecessary type conversion
- `no-unnecessary-type-parameters`: Disallows unnecessary type parameters
- `no-unsafe-function-type`: Disallows function parameters to have unsafe types
- `no-unused-expressions`: Disallows unused expressions
- `no-use-before-define`: Disallows the use of variables before they are defined
- `no-useless-constructor`: Disallows unnecessary constructors
- `no-wrapper-object-types`: Disallows use of the wrapper class constructors
- `parameter-properties`: Requires or disallows parameter properties in class constructors
- `prefer-destructuring`: Requires destructuring from arrays and/or objects
- `prefer-enum-initializers`: Prefer initializing enum members
- `prefer-find`: Prefer Array.find() over Array.filter()[0]
- `prefer-for-of`: Prefer for...of loop over traditional for loop
- `prefer-function-type`: Use function types instead of interfaces with call signatures
- `prefer-includes`: Enforce includes method over indexOf method
- `prefer-literal-enum-member`: Require that all enum members be literal values
- `prefer-namespace-keyword`: Requires using namespace keyword over module keyword
- `prefer-readonly`: Requires private members to be marked as readonly
- `prefer-readonly-parameter-types`: Requires function parameters to be typed as readonly
- `prefer-regexp-exec`: Prefer RegExp.exec() over String.match()
- `prefer-string-starts-ends-with`: Enforce usage of String.startsWith and String.endsWith
- `sort-type-constituents`: Enforces that members of a type union/intersection are sorted alphabetically
- `strict-boolean-expressions`: Restricts the types allowed in boolean expressions
- `triple-slash-reference`: Disallows triple slash references
- `typedef`: Requires type annotations to exist
- `unified-signatures`: Disallows overloads that could be unified into one

## Notes

- Rule porting should focus on high priority rules first, followed by medium priority
- Each rule ported should include tests and documentation
- Consider compatibility with existing RSLint architecture
