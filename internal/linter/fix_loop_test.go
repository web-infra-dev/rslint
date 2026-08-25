package linter

import (
	"errors"
	"testing"
)

func TestRunAutofixLoopStopsAtStableGeneration(t *testing.T) {
	var passes []AutofixPass
	err := RunAutofixLoop(AutofixLoopOptions{VerifyAfterLimit: true}, func(pass AutofixPass) (AutofixPassResult, error) {
		passes = append(passes, pass)
		return AutofixPassResult{Applied: pass.Index < 2}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(passes) != 3 {
		t.Fatalf("passes = %d, want two writes and one stable lint", len(passes))
	}
	for index, pass := range passes {
		if pass.Index != index || !pass.AllowApply {
			t.Fatalf("pass %d = %+v, want writable index %d", index, pass, index)
		}
	}
}

func TestRunAutofixLoopVerifiesAfterLimit(t *testing.T) {
	var passes []AutofixPass
	err := RunAutofixLoop(AutofixLoopOptions{VerifyAfterLimit: true}, func(pass AutofixPass) (AutofixPassResult, error) {
		passes = append(passes, pass)
		return AutofixPassResult{Applied: pass.AllowApply}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(passes) != 11 {
		t.Fatalf("passes = %d, want ten writes and one verification", len(passes))
	}
	verification := passes[len(passes)-1]
	if verification.Index != 10 || verification.AllowApply {
		t.Fatalf("verification pass = %+v, want read-only index 10", verification)
	}
}

func TestRunAutofixLoopCanSkipLimitVerification(t *testing.T) {
	passes := 0
	err := RunAutofixLoop(AutofixLoopOptions{}, func(pass AutofixPass) (AutofixPassResult, error) {
		passes++
		if !pass.AllowApply {
			t.Fatal("content-only loop must not run a read-only verification pass")
		}
		return AutofixPassResult{Applied: true}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if passes != 10 {
		t.Fatalf("passes = %d, want ten writable passes", passes)
	}
}

func TestRunAutofixLoopPropagatesPassError(t *testing.T) {
	want := errors.New("lint failed")
	err := RunAutofixLoop(AutofixLoopOptions{}, func(pass AutofixPass) (AutofixPassResult, error) {
		if pass.Index == 1 {
			return AutofixPassResult{}, want
		}
		return AutofixPassResult{Applied: true}, nil
	})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
