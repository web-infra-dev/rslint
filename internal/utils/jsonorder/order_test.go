package jsonorder

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestPropertyKeys(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  []string
	}{
		{`{"z":1,"a":2}`, []string{"z", "a"}},
		{`{"z":1,"a":2,"z":3}`, []string{"z", "a"}},
		{`{"z":1,"10":0,"2":0,"4294967295":0,"01":0,"0":0,"a":2}`, []string{"0", "2", "10", "z", "4294967295", "01", "a"}},
		{`{"@app/special":1,"@app":2,"~key":3,"":4}`, []string{"@app/special", "@app", "~key", ""}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			order, err := Parse([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			var object map[string]any
			if err := json.Unmarshal([]byte(tc.input), &object); err != nil {
				t.Fatal(err)
			}
			if got := order.PropertyKeys(object); !slices.Equal(got, tc.want) {
				t.Fatalf("keys = %v, want %v", got, tc.want)
			}
		})
	}
	var order *Order
	if got := order.PropertyKeys(map[string]any{"z": 1, "a": 2}); !slices.Equal(got, []string{"a", "z"}) {
		t.Fatalf("Go map keys = %v", got)
	}
	if _, err := Parse([]byte(`{"broken"`)); err == nil {
		t.Fatal("invalid JSON was accepted")
	}
}

func TestMarshalCurrentValues(t *testing.T) {
	input := `{"options":[{"aliases":{"z":"old","a":"a"},"removed":true}]}`
	order, err := Parse([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		t.Fatal(err)
	}
	options := value["options"].([]any)[0].(map[string]any)
	aliases := options["aliases"].(map[string]any)
	aliases["z"] = "new"
	aliases["b"] = "added"
	delete(options, "removed")
	options["defaulted"] = true
	encoded, err := Marshal(value, order)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"options":[{"aliases":{"z":"new","a":"a","b":"added"},"defaulted":true}]}`
	if string(encoded) != want {
		t.Fatalf("JSON = %s, want %s", encoded, want)
	}
}
