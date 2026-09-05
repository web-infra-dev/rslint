package checker

import (
	"reflect"
	"testing"

	upstream "github.com/microsoft/TypeScript/tsc/internal/checker"
)

// The generated accessors cast the compiler's Checker with unsafe.Pointer.
// Private generic link stores must preserve every subsequent field's offset.
func TestCheckerMirrorLayout(t *testing.T) {
	original := reflect.TypeFor[upstream.Checker]()
	mirror := reflect.TypeFor[extra_Checker]()
	if original.Size() != mirror.Size() || original.Align() != mirror.Align() || original.NumField() != mirror.NumField() {
		t.Fatalf("Checker mirror layout differs: original size/align/fields = %d/%d/%d, mirror = %d/%d/%d",
			original.Size(), original.Align(), original.NumField(), mirror.Size(), mirror.Align(), mirror.NumField())
	}
	for i := range original.NumField() {
		want, got := original.Field(i), mirror.Field(i)
		if want.Name != got.Name || want.Offset != got.Offset || want.Type.Size() != got.Type.Size() || want.Type.Align() != got.Type.Align() {
			t.Errorf("field %d: original %s offset/size/align = %d/%d/%d, mirror %s = %d/%d/%d",
				i, want.Name, want.Offset, want.Type.Size(), want.Type.Align(), got.Name, got.Offset, got.Type.Size(), got.Type.Align())
		}
	}
}
