package domain

import (
	"encoding/json"
	"testing"
)

func TestEvidenceUnmarshalValueStringOrNumber(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{`{"source":"s","value":"45%"}`, "45%"},
		{`{"source":"s","value":45}`, "45"},
		{`{"source":"s","value":45.5}`, "45.5"},
		{`{"source":"s","value":1400000000}`, "1400000000"},
		{`{"source":"s","value":true}`, "true"},
		{`{"source":"s"}`, ""},
		{`{"source":"s","value":null}`, ""},
	}

	for _, c := range cases {
		var ev Evidence
		if err := json.Unmarshal([]byte(c.raw), &ev); err != nil {
			t.Fatalf("Unmarshal(%s): %v", c.raw, err)
		}
		if ev.Value != c.want {
			t.Errorf("Unmarshal(%s).Value = %q, want %q", c.raw, ev.Value, c.want)
		}
	}

	// La sérialisation reste une chaîne (compat API/DB).
	out, err := json.Marshal(Evidence{Source: "s", Metric: "m", Value: "45%"})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"source":"s","metric":"m","value":"45%"}` {
		t.Errorf("marshal = %s", out)
	}
}
