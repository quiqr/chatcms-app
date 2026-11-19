package anchor

import (
	"testing"
	"testing/fstest"

	"github.com/quiqr/chatcms-app/internal/model"
)

const theModel = `singles:
  - key: home
    title: Home page
    file: content/_index.md
    fields:
      - key: title
        type: string
      - key: intro
        type: string
      - key: opening
        type: accordion
        fields:
          - key: day
            type: string
      - key: mainContent
        type: markdown
  - key: settings
    title: Site settings
    file: data/settings.yaml
    fields:
      - key: footer
        type: string
collections:
  - key: news
    title: News
    folder: content/news
    itemtitle: article
    fields:
      - key: title
        type: string
      - key: mainContent
        type: markdown
`

func aSite(t *testing.T) (fstest.MapFS, *model.Model) {
	t.Helper()
	site := fstest.MapFS{
		"quiqr/model/base.yaml":  {Data: []byte(theModel)},
		"content/_index.md":      {Data: []byte("---\ntitle: Korenaar Bakery\nintro: We bake with **apples** &amp; pears.\nopening:\n  - day: Monday\n  - day: Saturday\n---\n\nCome in and [smell it](/smell/).\n\nFresh bread every morning.\n")},
		"data/settings.yaml":     {Data: []byte("footer: Open every day but Sunday\n")},
		"content/news/spring.md": {Data: []byte("---\ntitle: Spring prices\n---\n\nFresh bread every morning, and rye on Fridays.\n")},
	}
	m, err := model.Load(site)
	if err != nil {
		t.Fatal(err)
	}
	return site, m
}

func TestASelectionIsFoundInTheFieldThatHoldsIt(t *testing.T) {
	site, m := aSite(t)
	for name, c := range map[string]struct {
		asked Asked
		want  Where
	}{
		"the title":                    {Asked{Text: "Korenaar Bakery", Page: "home"}, Where{Kind: InAField, Page: "home", PageTitle: "Home page", Field: "title", Matches: 1}},
		"through markdown and html":    {Asked{Text: "with apples & pears", Page: "home"}, Where{Kind: InAField, Page: "home", PageTitle: "Home page", Field: "intro", Matches: 1}},
		"a link's words in the body":   {Asked{Text: "Come in and smell it.", Page: "home"}, Where{Kind: InAField, Page: "home", PageTitle: "Home page", Field: "mainContent", Matches: 1}},
		"an item of a list":            {Asked{Text: "Saturday", Page: "home"}, Where{Kind: InAField, Page: "home", PageTitle: "Home page", Field: "opening[1].day", Matches: 1}},
		"shared text":                  {Asked{Text: "every day but Sunday", Page: "home"}, Where{Kind: SharedText, Page: "settings", PageTitle: "Site settings", Field: "footer", Matches: 1}},
		"another page, without a hint": {Asked{Text: "rye on Fridays"}, Where{Kind: InAField, Page: "news/spring", PageTitle: "Spring prices", Field: "mainContent", Matches: 1}},
		"nowhere: the design":          {Asked{Text: "Made with Hugo", Page: "home"}, Where{Kind: TheDesign}},
		"too short to say":             {Asked{Text: "a", Page: "home"}, Where{Kind: TheDesign}},
	} {
		t.Run(name, func(t *testing.T) {
			if got := Find(site, m, c.asked); got != c.want {
				t.Fatalf("Find() = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestTheHintedPageComesFirst(t *testing.T) {
	site, m := aSite(t)
	got := Find(site, m, Asked{Text: "Fresh bread every morning", Page: "news/spring"})
	if got.Page != "news/spring" {
		t.Fatalf("Find() = %+v, want the hinted page", got)
	}
	got = Find(site, m, Asked{Text: "Fresh bread every morning", Page: "home"})
	if got.Page != "home" || got.Field != "mainContent" {
		t.Fatalf("Find() = %+v, want the home page's text", got)
	}
}

func TestPlainText(t *testing.T) {
	for in, want := range map[string]string{
		"## A *heading* with `code`":             "a heading with code",
		"See [the shop](/shop/) and ![x](a.png)": "see the shop and x",
		"{{< figure src=\"a.jpg\" >}} Caption":   "caption",
		"<p>Hello&nbsp;there</p>":                "hello there",
		"- one\n- two":                           "one two",
	} {
		if got := Plain(in); got != want {
			t.Errorf("Plain(%q) = %q, want %q", in, got, want)
		}
	}
}
