package params

import "testing"

type testParams struct {
	Name  string `json:"name"`
	Count int    `json:"count,omitempty"`
	Age   int    `json:"age"`
}

func TestEncode(t *testing.T) {
	p := testParams{Name: "alice", Age: 25}
	m, err := Encode(p, map[string]any{"extra": "val"})
	if err != nil {
		t.Fatal(err)
	}
	if m["name"] != "alice" {
		t.Errorf("name = %v, want alice", m["name"])
	}
	if _, ok := m["count"]; ok {
		t.Error("count should be omitted for zero value")
	}
	if m["age"].(float64) != 25 {
		t.Errorf("age = %v, want 25", m["age"])
	}
	if m["extra"] != "val" {
		t.Errorf("extra = %v, want val", m["extra"])
	}
}

func TestToOrderedString(t *testing.T) {
	got := ToOrderedString(map[string]any{"b": "2", "a": "1", "c": 3})
	want := "a=1&b=2&c=3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToOrderedStringEmpty(t *testing.T) {
	got := ToOrderedString(map[string]any{})
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
