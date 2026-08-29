package ideas

import "testing"

func TestParseRevision(t *testing.T) {
	text := "TITLE: Comptabilité express\nHOOK: Gérez vos comptes en 10 minutes par jour\nEXPLANATION: Une app mobile simple pour les micro-commerçants."
	title, hook, explanation, err := parseRevision(text)
	if err != nil {
		t.Fatalf("parseRevision: %v", err)
	}
	if title != "Comptabilité express" {
		t.Errorf("title = %q", title)
	}
	if hook != "Gérez vos comptes en 10 minutes par jour" {
		t.Errorf("hook = %q", hook)
	}
	if explanation != "Une app mobile simple pour les micro-commerçants." {
		t.Errorf("explanation = %q", explanation)
	}
}

func TestParseRevisionEmpty(t *testing.T) {
	if _, _, _, err := parseRevision("n'importe quoi sans clé"); err == nil {
		t.Fatal("expected error for unparseable text")
	}
}
