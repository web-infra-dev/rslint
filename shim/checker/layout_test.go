package checker

import (
	"reflect"
	"testing"

	upstream "github.com/microsoft/TypeScript/tsc/internal/checker"
)

// Generated accessors cast these compiler types with unsafe.Pointer. Check
// every mirrored structure, including private stores embedded by value.
func TestCheckerMirrorLayout(t *testing.T) {
	for _, pair := range []struct{ original, mirror reflect.Type }{
		{reflect.TypeFor[upstream.Checker](), reflect.TypeFor[extra_Checker]()},
		{reflect.TypeFor[upstream.Type](), reflect.TypeFor[extra_Type]()},
		{reflect.TypeFor[upstream.ConditionalType](), reflect.TypeFor[extra_ConditionalType]()},
		{reflect.TypeFor[upstream.ConditionalRoot](), reflect.TypeFor[extra_ConditionalRoot]()},
		{reflect.TypeFor[upstream.TupleType](), reflect.TypeFor[extra_TupleType]()},
		{reflect.TypeFor[upstream.LiteralType](), reflect.TypeFor[extra_LiteralType]()},
		{reflect.TypeFor[upstream.Signature](), reflect.TypeFor[extra_Signature]()},
		{reflect.TypeFor[upstream.InterfaceType](), reflect.TypeFor[extra_InterfaceType]()},
	} {
		t.Run(pair.original.Name(), func(t *testing.T) {
			original, mirror := pair.original, pair.mirror
			if original.Size() != mirror.Size() || original.Align() != mirror.Align() || original.NumField() != mirror.NumField() {
				t.Fatalf("mirror layout differs: original size/align/fields = %d/%d/%d, mirror = %d/%d/%d",
					original.Size(), original.Align(), original.NumField(), mirror.Size(), mirror.Align(), mirror.NumField())
			}
			for i := range original.NumField() {
				want, got := original.Field(i), mirror.Field(i)
				if want.Name != got.Name || want.Offset != got.Offset || want.Type.Size() != got.Type.Size() || want.Type.Align() != got.Type.Align() {
					t.Errorf("field %d: original %s offset/size/align = %d/%d/%d, mirror %s = %d/%d/%d",
						i, want.Name, want.Offset, want.Type.Size(), want.Type.Align(), got.Name, got.Offset, got.Type.Size(), got.Type.Align())
				}
			}
		})
	}
}
