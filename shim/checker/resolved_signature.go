package checker

import (
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
)

// Checker_getResolvedSignatureWithArgumentCount mirrors TypeScript's
// getResolvedSignatureWorker(node, candidatesOutArray, argumentCount, checkMode)
// by forwarding candidatesOutArray and argumentCount independently while
// restoring apparentArgumentCount before returning.
func Checker_getResolvedSignatureWithArgumentCount(
	c *checker.Checker,
	node *ast.Node,
	candidatesOutArray *[]*checker.Signature,
	argumentCount *int,
	checkMode checker.CheckMode,
) *checker.Signature {
	if argumentCount != nil {
		mirror := (*extra_Checker)(unsafe.Pointer(c))
		previous := mirror.apparentArgumentCount
		mirror.apparentArgumentCount = argumentCount
		defer func() {
			mirror.apparentArgumentCount = previous
		}()
	}
	return Checker_getResolvedSignature(c, node, candidatesOutArray, checkMode)
}
