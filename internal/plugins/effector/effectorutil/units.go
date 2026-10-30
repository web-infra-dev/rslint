// Package effectorutil holds type predicates shared by effector rules. It
// mirrors `isType` from eslint-plugin-effector's src/shared/is.ts.
package effectorutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func packageSpecifier(names ...string) []utils.TypeOrValueSpecifier {
	return []utils.TypeOrValueSpecifier{{
		From:    utils.TypeOrValueSpecifierFromPackage,
		Package: "effector",
		Name:    names,
	}}
}

var unitSpecifier = packageSpecifier("Store", "StoreWritable", "Event", "EventCallable", "Effect", "Domain")

// IsUnit reports whether t is an effector Store, Event, Effect or Domain,
// matching upstream `isType.unit`: the type name must be one of the unit
// interfaces and its declaration must come from the `effector` package.
func IsUnit(t *checker.Type, program *program.Program) bool {
	return utils.TypeMatchesSomeSpecifier(t, unitSpecifier, nil, program)
}
