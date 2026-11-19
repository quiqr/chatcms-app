package app

import "testing"

func TestACombinedAnswerIsTidied(t *testing.T) {
	for said, want := range map[string]string{
		"  Korenaar by night and day \n":    "Korenaar by night and day",
		"```\nWe bake.\n```":                "We bake.",
		"```markdown\n# Bread\n\nRye.\n```": "# Bread\n\nRye.",
	} {
		if got := tidyCombined(said); got != want {
			t.Errorf("tidyCombined(%q) = %q, want %q", said, got, want)
		}
	}
	if prompt := combinePrompt("Home", "Title", "A", "B"); prompt != "Page: Home\nField: Title\n\nFirst version:\nA\n\nSecond version:\nB" {
		t.Errorf("prompt = %q", prompt)
	}
}
