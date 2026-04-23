package json

import (
	"testing"
)

func TestRawMessage(t *testing.T) {
	type T struct {
		Name string     `json:"name"`
		Raw  RawMessage `json:"raw"`
	}

	// Unmarshal
	input := `{"name":"test","raw":{"key":42}}`
	var v T
	if err := UnmarshalString(input, &v); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if v.Name != "test" {
		t.Fatalf("Name = %q, want %q", v.Name, "test")
	}
	if string(v.Raw) != `{"key":42}` {
		t.Fatalf("Raw = %q, want %q", v.Raw, `{"key":42}`)
	}

	// Marshal
	v2 := T{Name: "foo", Raw: RawMessage(`[1,2,3]`)}
	bs, err := Marshal(&v2)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	want := `{"name":"foo","raw":[1,2,3]}`
	if string(bs) != want {
		t.Fatalf("Marshal = %q, want %q", bs, want)
	}

	// nil RawMessage marshals to null
	v3 := T{Name: "bar"}
	bs, err = Marshal(&v3)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	want = `{"name":"bar","raw":null}`
	if string(bs) != want {
		t.Fatalf("Marshal = %q, want %q", bs, want)
	}
}
