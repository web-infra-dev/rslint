package utils

import "testing"

func TestNormalizeGlobalAccess(t *testing.T) {
	tests := []struct {
		value any
		want  GlobalAccess
		ok    bool
	}{
		{true, GlobalAccessWritable, true},
		{"true", GlobalAccessWritable, true},
		{"writable", GlobalAccessWritable, true},
		{"writeable", GlobalAccessWritable, true},
		{false, GlobalAccessReadonly, true},
		{"false", GlobalAccessReadonly, true},
		{"readonly", GlobalAccessReadonly, true},
		{"readable", GlobalAccessReadonly, true},
		{nil, GlobalAccessReadonly, true},
		{"off", GlobalAccessOff, true},
		{"", GlobalAccessUnset, false},
		{"Off", GlobalAccessUnset, false},
		{"nonsense", GlobalAccessUnset, false},
		{0, GlobalAccessUnset, false},
		{[]string{"readonly"}, GlobalAccessUnset, false},
	}
	for _, test := range tests {
		got, ok := NormalizeGlobalAccess(test.value)
		if got != test.want || ok != test.ok {
			t.Errorf("NormalizeGlobalAccess(%#v) = (%v, %t), want (%v, %t)", test.value, got, ok, test.want, test.ok)
		}
	}
}

func TestNormalizeInlineGlobalAccess(t *testing.T) {
	tests := []struct {
		setting    string
		hasSetting bool
		want       GlobalAccess
		ok         bool
	}{
		{"", false, GlobalAccessReadonly, true},
		{"", true, GlobalAccessUnset, false},
		{"writable", true, GlobalAccessWritable, true},
		{"readonly", true, GlobalAccessReadonly, true},
		{"off", true, GlobalAccessOff, true},
		{"bar", true, GlobalAccessUnset, false},
	}
	for _, test := range tests {
		got, ok := NormalizeInlineGlobalAccess(test.setting, test.hasSetting)
		if got != test.want || ok != test.ok {
			t.Errorf("NormalizeInlineGlobalAccess(%q, %t) = (%v, %t), want (%v, %t)", test.setting, test.hasSetting, got, ok, test.want, test.ok)
		}
	}
}

func TestGlobalAccessPredicates(t *testing.T) {
	tests := []struct {
		access     GlobalAccess
		isDeclared bool
		isWritable bool
		text       string
	}{
		{GlobalAccessUnset, false, false, "unset"},
		{GlobalAccessOff, false, false, "off"},
		{GlobalAccessReadonly, true, false, "readonly"},
		{GlobalAccessWritable, true, true, "writable"},
	}
	for _, test := range tests {
		if got := test.access.IsDeclared(); got != test.isDeclared {
			t.Errorf("%v.IsDeclared() = %t, want %t", test.access, got, test.isDeclared)
		}
		if got := test.access.IsWritable(); got != test.isWritable {
			t.Errorf("%v.IsWritable() = %t, want %t", test.access, got, test.isWritable)
		}
		if got := test.access.String(); got != test.text {
			t.Errorf("String() = %q, want %q", got, test.text)
		}
	}
}
