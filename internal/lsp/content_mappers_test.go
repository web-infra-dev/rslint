package lsp

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/lsp/lsproto"
)

func TestInitializeReadsRunExternalCode(t *testing.T) {
	t.Parallel()
	enabled, disabled := true, false
	tests := []struct {
		name    string
		options *lsproto.InitializationOptionsOrNull
		want    bool
	}{
		{name: "absent"},
		{name: "null", options: &lsproto.InitializationOptionsOrNull{}},
		{name: "unset", options: &lsproto.InitializationOptionsOrNull{InitializationOptions: &lsproto.InitializationOptions{}}},
		{name: "false", options: &lsproto.InitializationOptionsOrNull{InitializationOptions: &lsproto.InitializationOptions{RunExternalCode: &disabled}}},
		{name: "true", options: &lsproto.InitializationOptionsOrNull{InitializationOptions: &lsproto.InitializationOptions{RunExternalCode: &enabled}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newTestServer()
			if _, err := s.handleInitialize(context.Background(), &lsproto.InitializeParams{InitializationOptions: tt.options}); err != nil {
				t.Fatal(err)
			}
			if s.contentMappers == nil {
				t.Fatal("initialize must create the content mapper Runtime")
			}
			if got := s.contentMappers.Enabled(); got != tt.want {
				t.Fatalf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
