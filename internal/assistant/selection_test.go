package assistant

import (
	"strings"
	"testing"

	"github.com/quiqr/chatcms-app/internal/httpapi"
)

func TestASelectionOnAKnownPageNamesThePage(t *testing.T) {
	about := httpapi.ChatSelection{Text: "Fresh bread", Before: "Korenaar ", After: " daily.", Path: "/post/hello/"}
	said := AboutASelection("  Make it rye bread ", about, "posts/hello")
	for _, want := range []string{
		"Make it rye bread\n\n",
		`on the preview page /post/hello/:` + "\n" + `"Fresh bread"`,
		`Just before it: "Korenaar "`,
		`Just after it: " daily."`,
		`the page with key "posts/hello"`,
		"change nothing else",
	} {
		if !strings.Contains(said, want) {
			t.Errorf("the prompt lacks %q:\n%s", want, said)
		}
	}
}

func TestASelectionOnNoPageSaysSo(t *testing.T) {
	said := AboutASelection("Shorter", httpapi.ChatSelection{Text: "Menu \"Home\"", Path: "/"}, "")
	if !strings.Contains(said, "No page was found") || strings.Contains(said, "Just before") || !strings.Contains(said, `"Menu \"Home\""`) {
		t.Errorf("prompt = %s", said)
	}
}
