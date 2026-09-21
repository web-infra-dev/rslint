package test_framework_test

import (
	"testing"

	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func TestIsSubjectMutatingChaiMatcher(t *testing.T) {
	for _, name := range []string{
		"property", "ownProperty", "haveOwnProperty",
		"ownPropertyDescriptor", "haveOwnPropertyDescriptor",
		"toContain", "toThrow", "toThrowError", "throw", "throws", "Throw",
	} {
		if !testFramework.IsSubjectMutatingChaiMatcher(name) {
			t.Errorf("expected %q to mutate the assertion subject", name)
		}
	}
	for _, name := range []string{
		"toHaveLength", "toHaveBeenCalledTimes", "toBe", "toEqual",
		"toContainEqual", "not", "resolves", "", "PROPERTY",
	} {
		if testFramework.IsSubjectMutatingChaiMatcher(name) {
			t.Errorf("expected %q to keep the assertion subject", name)
		}
	}
}
