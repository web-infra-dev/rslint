package rule

import (
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

func TestFilterNonTypeAwareRules(t *testing.T) {
	tests := []struct {
		name  string
		rules []ConfiguredRule
		want  []string
	}{
		{
			name: "mixed",
			rules: []ConfiguredRule{
				{Name: "syntax-rule"},
				{Name: "type-rule", RequiresTypeInfo: true},
				{Name: "another-syntax"},
			},
			want: []string{"syntax-rule", "another-syntax"},
		},
		{
			name: "all type aware",
			rules: []ConfiguredRule{
				{Name: "type-rule-1", RequiresTypeInfo: true},
				{Name: "type-rule-2", RequiresTypeInfo: true},
			},
		},
		{
			name:  "all syntax",
			rules: []ConfiguredRule{{Name: "rule-a"}, {Name: "rule-b"}},
			want:  []string{"rule-a", "rule-b"},
		},
		{name: "empty"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filtered := FilterNonTypeAwareRules(test.rules)
			if len(filtered) != len(test.want) {
				t.Fatalf("got %d rules, want %d", len(filtered), len(test.want))
			}
			for index, name := range test.want {
				if filtered[index].Name != name {
					t.Fatalf("rule %d = %q, want %q", index, filtered[index].Name, name)
				}
			}
		})
	}
}

func TestCreateRulePreservesRequiresTypeInfo(t *testing.T) {
	configured := CreateRule(Rule{
		Name:             "test-rule",
		RequiresTypeInfo: true,
		Run:              func(RuleContext, []any) RuleListeners { return nil },
	})

	if configured.Name != "@typescript-eslint/test-rule" {
		t.Fatalf("unexpected name: %s", configured.Name)
	}
	if !configured.RequiresTypeInfo {
		t.Fatal("RequiresTypeInfo should survive CreateRule")
	}
}

func TestPreparedRuleIdentityAndFileContext(t *testing.T) {
	preparations := 0
	var events []string
	r := WithPreparation(Rule{Name: "prepared", Schema: EmptyArraySchema}, func(options []any) FileRunner {
		preparations++
		value := options[0].(string)
		return func(ctx RuleContext) RuleListeners {
			file := ctx.Settings["file"].(string)
			return RuleListeners{ast.KindSourceFile: func(*ast.Node) {
				events = append(events, value+":"+file)
			}}
		}
	})
	first := r.Configure([]any{"first"})
	second := r.Configure([]any{"second"})
	copyOfFirst := FilterNonTypeAwareRules([]ConfiguredRule{first})[0]
	if preparations != 0 {
		t.Fatal("configuration eagerly prepared a rule")
	}
	var executor Executor
	for i, configured := range []ConfiguredRule{first, second, copyOfFirst} {
		ctx := RuleContext{Settings: map[string]any{"file": strconv.Itoa(i)}}
		executor.Run(configured, ctx)[ast.KindSourceFile](nil)
	}
	if preparations != 2 {
		t.Fatalf("prepared %d instances, want one per configuration", preparations)
	}
	if want := []string{"first:0", "second:1", "first:2"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("file/configuration state leaked: got %v, want %v", events, want)
	}
	executor.Release()
	if len(executor.runners) != 0 {
		t.Fatal("released executor retained configuration references")
	}
	// Identical names and options in a fresh configuration still get their own
	// identity. A refreshed resolver must never borrow the previous generation.
	fresh := r.Configure(first.Options)
	if fresh.prepared == first.prepared || copyOfFirst.prepared != first.prepared {
		t.Fatal("configuration identity did not survive copies or isolate refresh")
	}
	executor.Run(fresh, RuleContext{Settings: map[string]any{"file": "fresh"}})[ast.KindSourceFile](nil)
	if preparations != 3 {
		t.Fatalf("fresh configuration reused an old instance: %d preparations", preparations)
	}
	executor.Release()
}

func TestPreparedRuleConcurrentExecutors(t *testing.T) {
	const workers = 12
	var preparations atomic.Int32
	r := WithPreparation(Rule{Name: "prepared"}, func([]any) FileRunner {
		preparations.Add(1)
		// Deliberately non-atomic: this state is exclusively task-owned even
		// while callbacks run. The race detector verifies that contract too.
		calls := 0
		return func(ctx RuleContext) RuleListeners {
			expect := ctx.Settings["call"].(int)
			return RuleListeners{ast.KindSourceFile: func(*ast.Node) {
				calls++
				if calls != expect {
					t.Errorf("runner shared by active tasks: got call %d, want %d", calls, expect)
				}
			}}
		}
	}).Configure(nil)
	ready := make(chan struct{}, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			var executor Executor
			listener := executor.Run(r, RuleContext{Settings: map[string]any{"call": 1}})[ast.KindSourceFile]
			ready <- struct{}{}
			<-start
			listener(nil)
			for call := 2; call <= 20; call++ {
				executor.Run(r, RuleContext{Settings: map[string]any{"call": call}})[ast.KindSourceFile](nil)
			}
			executor.Release()
		})
	}
	for range workers {
		<-ready
	}
	if got := preparations.Load(); got != workers {
		t.Errorf("got %d preparations for %d simultaneous leases", got, workers)
	}
	close(start)
	wg.Wait()
}

func TestPreparedRuleDirectAndLegacyExecution(t *testing.T) {
	calls := 0
	legacy := Rule{Name: "legacy", Run: func(_ RuleContext, options []any) RuleListeners {
		calls += options[0].(int)
		return nil
	}}
	var executor Executor
	configured := legacy.Configure([]any{2})
	executor.Run(configured, RuleContext{})
	executor.Run(configured, RuleContext{})
	if calls != 4 || executor.runners != nil || configured.prepared != nil {
		t.Fatal("ordinary rule changed behavior or allocated prepared state")
	}
	preparations := 0
	prepared := CreateRule(WithPreparation(Rule{Name: "prepared", RequiresTypeInfo: true}, func([]any) FileRunner {
		preparations++
		return func(RuleContext) RuleListeners { calls++; return nil }
	}))
	configured = prepared.Configure(nil)
	if configured.prepared == nil || !configured.RequiresTypeInfo || configured.Name != "@typescript-eslint/prepared" {
		t.Fatal("TypeScript wrapper lost preparation or metadata")
	}
	prepared.Run(RuleContext{}, nil)
	configured.Run(RuleContext{})
	if preparations != 2 || calls != 6 {
		t.Fatalf("direct execution must create independent instances: preparations=%d calls=%d", preparations, calls)
	}
}

func BenchmarkConfiguredRulePreparation(b *testing.B) {
	r := WithPreparation(Rule{Name: "pattern"}, func(options []any) FileRunner {
		pattern, err := esregexp.Compile(options[0].(string), "u")
		if err != nil {
			panic(err)
		}
		return func(RuleContext) RuleListeners {
			return RuleListeners{ast.KindIdentifier: func(*ast.Node) {
				for range 25 {
					pattern.Test("validIdentifier")
				}
			}}
		}
	})
	for _, reuse := range []bool{false, true} {
		name := "PerFile"
		if reuse {
			name = "Prepared"
		}
		b.Run(name, func(b *testing.B) {
			configured := r.Configure([]any{"^(?!forbidden)[A-Za-z_$][A-Za-z0-9_$]*$"})
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				var executor Executor
				for pb.Next() {
					var listeners RuleListeners
					if reuse {
						listeners = executor.Run(configured, RuleContext{})
					} else {
						listeners = configured.Run(RuleContext{})
					}
					listeners[ast.KindIdentifier](nil)
				}
				executor.Release()
			})
		})
	}
}
