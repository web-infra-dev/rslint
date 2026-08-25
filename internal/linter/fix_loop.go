package linter

// maxAutofixPasses matches ESLint's write-pass limit. The limit stays private
// so integrations cannot reproduce the loop and drift on its boundary.
const maxAutofixPasses = 10

// AutofixLoopOptions controls integration-specific work after the last
// writable pass. CLI and API consumers need a final diagnostic generation;
// content-only consumers such as LSP fix-all do not.
type AutofixLoopOptions struct {
	VerifyAfterLimit bool
}

// AutofixPass describes one lint pass. Index is zero-based. AllowApply is false
// only for the optional verification pass after all writable slots were used.
type AutofixPass struct {
	Index      int
	AllowApply bool
}

// AutofixPassResult reports whether a writable pass changed any source.
type AutofixPassResult struct {
	Applied bool
}

// RunAutofixLoop runs at most ten writable lint-fix passes. A pass that applies
// nothing is already the final verified generation. If every writable pass
// changes source and VerifyAfterLimit is set, one additional read-only pass is
// run so diagnostics describe the bytes produced by the tenth write.
//
// The callback owns integration-specific linting, Program reconstruction,
// plugin dispatch, source storage, and fix application. This coordinator owns
// only pass lifecycle and error propagation.
func RunAutofixLoop(
	options AutofixLoopOptions,
	run func(AutofixPass) (AutofixPassResult, error),
) error {
	for index := range maxAutofixPasses {
		passResult, err := run(AutofixPass{Index: index, AllowApply: true})
		if err != nil {
			return err
		}
		if !passResult.Applied {
			return nil
		}
	}

	if options.VerifyAfterLimit {
		_, err := run(AutofixPass{Index: maxAutofixPasses, AllowApply: false})
		if err != nil {
			return err
		}
	}
	return nil
}
