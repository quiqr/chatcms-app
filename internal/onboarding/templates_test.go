package onboarding

import (
	"os"
	"path/filepath"
	"testing"
)

const communityFeed = "https://quiqr.github.io/quiqr-community-templates/templates.json"

func recordedFeed(t *testing.T) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "test", "fixtures", "templates", "templates.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func byID(templates []Template) map[string]Template {
	found := map[string]Template{}
	for _, template := range templates {
		found[template.ID] = template
	}
	return found
}

func TestTheRecordedFeedKeepsEveryValidEntry(t *testing.T) {
	templates, err := ParseFeed(recordedFeed(t), communityFeed)
	if err != nil {
		t.Fatal(err)
	}

	if len(templates) != 9 {
		t.Fatalf("kept %d templates, want the 9 valid ones", len(templates))
	}
	if _, kept := byID(templates)["No-Source"]; kept {
		t.Error("the entry without a source repository was kept")
	}
}

func TestTemplatesAreSortedByName(t *testing.T) {
	templates, err := ParseFeed([]byte(`[
		{"QuiqrEtalageName": "Summer", "NormalizedName": "Summer", "SourceLink": "https://example.org/summer"},
		{"QuiqrEtalageName": "paper", "NormalizedName": "Paper", "SourceLink": "https://example.org/paper"},
		{"QuiqrEtalageName": "Bexer", "NormalizedName": "Bexer", "SourceLink": "https://example.org/bexer"}
	]`), communityFeed)
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, template := range templates {
		names = append(names, template.Name)
	}
	if got, want := len(names), 3; got != want {
		t.Fatalf("names = %v", names)
	}
	if names[0] != "Bexer" || names[1] != "paper" || names[2] != "Summer" {
		t.Errorf("order = %v, want Bexer, paper, Summer", names)
	}
}

func TestDetailsAreMappedFromTheFeed(t *testing.T) {
	templates, err := ParseFeed(recordedFeed(t), communityFeed)
	if err != nil {
		t.Fatal(err)
	}
	summer := byID(templates)["Summer"]

	if summer.Name != "Summer" || summer.SourceURL != "https://github.com/mipmip/summer-qremix" {
		t.Errorf("summer = %+v", summer)
	}
	if summer.HugoVersion != "0.101.0" || !summer.Extended {
		t.Errorf("hugo = %q extended %v, want 0.101.0 extended", summer.HugoVersion, summer.Extended)
	}
	if len(summer.Model) != 4 || summer.Model[0] != "base" {
		t.Errorf("model = %v, want the four parts", summer.Model)
	}
	if summer.DemoURL == "" || summer.ScreenshotType != "jpg" {
		t.Errorf("demo %q screenshot %q", summer.DemoURL, summer.ScreenshotType)
	}

	kitchen := byID(templates)["KitchenSink"]
	if kitchen.HugoVersion != "0.111.3" || kitchen.Extended {
		t.Errorf("kitchen sink hugo = %q extended %v", kitchen.HugoVersion, kitchen.Extended)
	}
	if kitchen.DemoURL != "" {
		t.Errorf("an empty demo became %q", kitchen.DemoURL)
	}
}

func TestEntriesThatCannotBeUsedAreDropped(t *testing.T) {
	templates, err := ParseFeed([]byte(`[
		{"QuiqrEtalageName": "", "NormalizedName": "nameless", "SourceLink": "https://example.org/a"},
		{"QuiqrEtalageName": "Slash", "NormalizedName": "../etc", "SourceLink": "https://example.org/b"},
		{"QuiqrEtalageName": "File", "NormalizedName": "file", "SourceLink": "file:///etc"},
		{"QuiqrEtalageName": "Ssh", "NormalizedName": "ssh", "SourceLink": "git@github.com:a/b.git"},
		{"QuiqrEtalageName": "Ext", "NormalizedName": "ext", "SourceLink": "ext::sh -c touch% /tmp/x"},
		{"QuiqrEtalageName": "Plain", "NormalizedName": "plain", "SourceLink": "http://example.org/c"},
		{"QuiqrEtalageName": "Loop", "NormalizedName": "loop", "SourceLink": "http://127.0.0.1:9/d"},
		{"QuiqrEtalageName": "Creds", "NormalizedName": "creds", "SourceLink": "https://u:p@example.org/e"},
		"not an object",
		{"QuiqrEtalageName": "Good", "NormalizedName": "good", "SourceLink": "https://example.org/f"},
		{"QuiqrEtalageName": "Twice", "NormalizedName": "good", "SourceLink": "https://example.org/g"}
	]`), communityFeed)
	if err != nil {
		t.Fatal(err)
	}

	if len(templates) != 1 || templates[0].Name != "Good" {
		t.Errorf("kept %+v, want only Good", templates)
	}
}

func TestALoopbackFeedMayNameLoopbackSources(t *testing.T) {
	feed := []byte(`[{"QuiqrEtalageName": "Loop", "NormalizedName": "loop", "SourceLink": "http://127.0.0.1:9/d"}]`)

	templates, err := ParseFeed(feed, "http://127.0.0.1:8000/templates.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 1 {
		t.Errorf("a loopback feed dropped its loopback source: %+v", templates)
	}
}

func TestAFeedThatIsNotAList(t *testing.T) {
	if _, err := ParseFeed([]byte(`{"templates": []}`), communityFeed); err == nil {
		t.Error("an object was accepted as a feed")
	}
}

func TestLinksThatAreNotWebAddressesAreDropped(t *testing.T) {
	templates, err := ParseFeed([]byte(`[{"QuiqrEtalageName": "X", "NormalizedName": "x",
		"SourceLink": "https://example.org/x", "QuiqrEtalageDemoUrl": "javascript:alert(1)",
		"QuiqrEtalageLicenseURL": "https://example.org/licence", "QuiqrFormsEndPoints": 3,
		"HugoVersion": "latest", "ScreenshotImageType": "svg"}]`), communityFeed)
	if err != nil || len(templates) != 1 {
		t.Fatalf("templates = %+v, err = %v", templates, err)
	}
	x := templates[0]
	if x.DemoURL != "" || x.LicenseURL == "" || x.Forms != 3 || x.HugoVersion != "" || x.ScreenshotType != "" {
		t.Errorf("x = %+v", x)
	}
}

func TestTheScreenshotLivesBesideTheFeed(t *testing.T) {
	got, ok := ScreenshotURL(communityFeed, Template{ID: "Summer", ScreenshotType: "jpg"})
	want := "https://quiqr.github.io/quiqr-community-templates/templates/Summer/screenshot.jpg"
	if !ok || got != want {
		t.Errorf("ScreenshotURL = %q, %v, want %q", got, ok, want)
	}

	if _, ok := ScreenshotURL(communityFeed, Template{ID: "Summer"}); ok {
		t.Error("a template without a screenshot type got an address")
	}
}

func TestTheInstalledHugoSaysItsVersion(t *testing.T) {
	cases := map[string]Hugo{
		"hugo v0.139.0-6c1e7b7f+extended linux/amd64 BuildDate=unknown VendorInfo=nixpkgs":  {Version: "0.139.0", Extended: true},
		"hugo v0.111.3-5d4eb5154e1fed125ca8e9b5a0315c4180dab192 linux/amd64 BuildDate=2023": {Version: "0.111.3"},
		"Hugo Static Site Generator v0.54.0/extended linux/amd64":                           {Version: "0.54.0"},
		"command not found": {},
	}
	for output, want := range cases {
		if got := ParseHugoVersion(output); got != want {
			t.Errorf("ParseHugoVersion(%q) = %+v, want %+v", output, got, want)
		}
	}
}
