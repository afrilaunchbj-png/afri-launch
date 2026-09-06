package jobs

import "testing"

// TestParseResearchResultNumericEvidence vérifie qu'un `value` numérique
// dans l'evidence renvoyée par le LLM ne fait pas échouer le décodage
// (régression : "cannot unmarshal number into field of type string").
func TestParseResearchResultNumericEvidence(t *testing.T) {
	content := `{"opportunities":[{"country":"Bénin","title":"Formation en ligne","summary":"s.","difficulty":"low","signal":"verified","score":72,"scores":{"demand":70,"pain":65,"competition":55,"purchasing_power":50,"digital_fit":80,"evidence_strength":60},"evidence":[{"source":"x","title":"r","url":"https://example.test","metric":"pénétration mobile","value":45},{"source":"y","metric":"valeur","value":"1,2 M"}]}]}`
	ops, err := ParseResearchResult(content)
	if err != nil {
		t.Fatalf("ParseResearchResult: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops = %d", len(ops))
	}
	if got := ops[0].Evidence[0].Value; got != "45" {
		t.Errorf("evidence[0].value = %q, want \"45\"", got)
	}
	if got := ops[0].Evidence[1].Value; got != "1,2 M" {
		t.Errorf("evidence[1].value = %q, want \"1,2 M\"", got)
	}
}

func TestParseResearchResultStripFences(t *testing.T) {
	content := "```json\n{\"opportunities\":[{\"country\":\"CI\",\"title\":\"T\",\"summary\":\"S\",\"evidence\":[]}]}\n```"
	ops, err := ParseResearchResult(content)
	if err != nil {
		t.Fatal(err)
	}
	if ops[0].Country != "CI" {
		t.Errorf("country = %q", ops[0].Country)
	}
}
