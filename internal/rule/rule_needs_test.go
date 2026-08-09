package rule

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
)

func TestCreateRulePreservesNeeds(t *testing.T) {
	needs := RuleNeeds{Refs: RefNeedReferences | RefNeedImportBindings}
	created := CreateRule(Rule{Name: "example", Needs: needs})
	if created.Needs != needs {
		t.Fatalf("CreateRule Needs = %v, want %v", created.Needs, needs)
	}
}

func TestRuleNeedsFitExistingReporterLayout(t *testing.T) {
	type legacyReporter struct {
		_ string
		DiagnosticSeverity
		DiagnosticConsumer
	}
	if got, want := unsafe.Sizeof(ruleContextReporter{}), unsafe.Sizeof(legacyReporter{}); got != want {
		t.Fatalf("ruleContextReporter size = %d, want legacy size %d", got, want)
	}
}

func TestRefStoreZeroConsumerLayoutAndControlPlane(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 8 && unsafe.Sizeof(RefStore{}) > 160 {
		t.Fatalf("RefStore size = %d, exceeds the existing 160-byte allocation class", unsafe.Sizeof(RefStore{}))
	}
	storeType := reflect.TypeOf((*RefStore)(nil))
	for _, name := range []string{"Start", "Observe", "Complete", "StartCollection", "ObserveIdentifier", "CompleteCollection"} {
		if _, ok := storeType.MethodByName(name); ok {
			t.Fatalf("RefStore exposes traversal control method %s to RuleContext consumers", name)
		}
	}
	contextType := reflect.TypeOf(RuleContext{})
	collectorType := reflect.TypeOf(RefCollector{})
	observerType := reflect.TypeOf(RefObserver{})
	for index := range contextType.NumField() {
		fieldType := contextType.Field(index).Type
		if fieldType == collectorType || fieldType == observerType {
			t.Fatalf("RuleContext exposes RefStore traversal control through field %s", contextType.Field(index).Name)
		}
	}
}

func TestFileFinalizeKindDoesNotOverlapListenerRanges(t *testing.T) {
	lastSyntheticExit := ListenerOnExit(ListenerOnNotAllowPattern(ast.KindCount - 1))
	if ListenerOnFileFinalize() <= lastSyntheticExit {
		t.Fatalf("file-finalize kind %d overlaps synthetic listener range ending at %d", ListenerOnFileFinalize(), lastSyntheticExit)
	}
}

func TestRuleContextRefRequestsValidateRuleCeiling(t *testing.T) {
	_, refs := newBoundRefStore(t, "/request-context.ts", core.ScriptKindTS, "const value = 1;")
	ctx := RuleContext{Refs: refs}.WithRuleNeeds(RuleNeeds{
		Refs: RefNeedReferences | RefNeedImportBindings,
	})
	if !ctx.RequestRefs(RefNeedReferences) || refs.requested != RefNeedReferences {
		t.Fatalf("declared request was not recorded: %v", refs.requested)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("undeclared RefStore request did not panic")
		}
	}()
	ctx.RequestRefs(RefNeedBindingDeclarations)
}

func TestRuleContextRefRequestWithoutStoreReturnsFalse(t *testing.T) {
	ctx := RuleContext{}.WithRuleNeeds(RuleNeeds{Refs: RefNeedReferences})
	if ctx.RequestRefs(RefNeedReferences) {
		t.Fatal("request without a RefStore returned true")
	}
}
