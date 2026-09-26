package schemagen

import "testing"

func TestOrderedJSONRoundTripPreservesKeyOrder(t *testing.T) {
	input := []byte(`{"z":1,"nested":{"b":true,"a":null},"items":[3,2,1]}`)
	node, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	output, err := node.MarshalIndent()
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"z\": 1,\n  \"nested\": {\n    \"b\": true,\n    \"a\": null\n  },\n  \"items\": [\n    3,\n    2,\n    1\n  ]\n}\n"
	if string(output) != want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestOrderedJSONSetAndDelete(t *testing.T) {
	node, err := Parse([]byte(`{"first":1,"second":2}`))
	if err != nil {
		t.Fatal(err)
	}
	object, ok := node.AsObject()
	if !ok {
		t.Fatal("root is not an object")
	}
	object.Set("third", newScalar(3))
	object.Delete("first")
	output, err := node.MarshalIndent()
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"second\": 2,\n  \"third\": 3\n}\n"
	if string(output) != want {
		t.Fatalf("unexpected output:\n%s", output)
	}
}
