package assistant

import (
	"strings"
	"testing"

	"github.com/quiqr/chatcms-app/internal/anchor"
	"github.com/quiqr/chatcms-app/internal/httpapi"
)

func TestASelectionInAFieldNamesTheFieldAndThePage(t *testing.T) {
	about := httpapi.ChatSelection{Text: "Fresh bread", Before: "Korenaar ", After: " daily.", Path: "/post/hello/"}
	said := AboutASelection("  Make it rye bread ", about, anchor.Where{Kind: anchor.InAField, Page: "posts/hello", Field: "mainContent"})
	for _, want := range []string{
		"Make it rye bread\n\n",
		`on the preview page /post/hello/:` + "\n" + `"Fresh bread"`,
		`Just before it: "Korenaar "`,
		`Just after it: " daily."`,
		`the field "mainContent" of the page with key "posts/hello"`,
		"change nothing else",
	} {
		if !strings.Contains(said, want) {
			t.Errorf("the prompt lacks %q:\n%s", want, said)
		}
	}
}

func TestASelectionOnAKnownPageWithoutAFieldNamesThePage(t *testing.T) {
	said := AboutASelection("Shorter", httpapi.ChatSelection{Text: "x y", Path: "/"}, anchor.Where{Kind: anchor.InAField, Page: "home"})
	if !strings.Contains(said, `the page with key "home"`) {
		t.Errorf("prompt = %s", said)
	}
}

func TestSharedTextIsSaidToShowOnMorePages(t *testing.T) {
	said := AboutASelection("Shorter", httpapi.ChatSelection{Text: "Open daily", Path: "/"}, anchor.Where{Kind: anchor.SharedText, Page: "settings", PageTitle: "Site settings", Field: "footer"})
	if !strings.Contains(said, `shared text: the field "footer" of "Site settings"`) || !strings.Contains(said, "every page") {
		t.Errorf("prompt = %s", said)
	}
}

func TestTheDesignIsSaidToBeOutOfReach(t *testing.T) {
	said := AboutASelection("Shorter", httpapi.ChatSelection{Text: "Made with Hugo", Path: "/"}, anchor.Where{Kind: anchor.TheDesign})
	if !strings.Contains(said, "comes from the site's design") {
		t.Errorf("prompt = %s", said)
	}
}

func TestASelectionOnNoPageSaysSo(t *testing.T) {
	said := AboutASelection("Shorter", httpapi.ChatSelection{Text: "Menu \"Home\"", Path: "/"}, anchor.Where{})
	if !strings.Contains(said, "No page was found") || strings.Contains(said, "Just before") || !strings.Contains(said, `"Menu \"Home\""`) {
		t.Errorf("prompt = %s", said)
	}
}
