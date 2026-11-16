package onboarding

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

type Template struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Author         string   `json:"author"`
	AuthorURL      string   `json:"authorUrl"`
	License        string   `json:"license"`
	LicenseURL     string   `json:"licenseUrl"`
	Homepage       string   `json:"homepage"`
	DemoURL        string   `json:"demoUrl"`
	HugoTheme      string   `json:"hugoTheme"`
	HugoVersion    string   `json:"hugoVersion"`
	Extended       bool     `json:"extended"`
	Model          []string `json:"model"`
	Forms          int      `json:"forms"`
	SourceURL      string   `json:"sourceUrl"`
	ScreenshotType string   `json:"screenshotType"`
}

type feedEntry struct {
	Name           string          `json:"QuiqrEtalageName"`
	NormalizedName string          `json:"NormalizedName"`
	Description    string          `json:"QuiqrEtalageDescription"`
	Author         string          `json:"QuiqrEtalageAuthor"`
	AuthorURL      string          `json:"QuiqrEtalageAuthorHomepage"`
	License        string          `json:"QuiqrEtalageLicense"`
	LicenseURL     string          `json:"QuiqrEtalageLicenseURL"`
	Homepage       string          `json:"QuiqrEtalageHomepage"`
	DemoURL        string          `json:"QuiqrEtalageDemoUrl"`
	HugoTheme      string          `json:"HugoTheme"`
	HugoVersion    string          `json:"HugoVersion"`
	Model          string          `json:"QuiqrModel"`
	Forms          json.RawMessage `json:"QuiqrFormsEndPoints"`
	ScreenshotType string          `json:"ScreenshotImageType"`
	SourceLink     string          `json:"SourceLink"`
}

var (
	templateID      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
	screenshotTypes = map[string]bool{"png": true, "jpg": true, "jpeg": true, "webp": true, "gif": true}
	hugoVersion     = regexp.MustCompile(`^[0-9]+(\.[0-9]+){1,2}$`)
)

func ParseFeed(data []byte, feedURL string) ([]Template, error) {
	var entries []json.RawMessage
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("the template feed is not a list: %w", err)
	}

	loopbackFeed := isLoopbackHTTP(feedURL)
	seen := map[string]bool{}
	templates := []Template{}

	for _, raw := range entries {
		var entry feedEntry
		if json.Unmarshal(raw, &entry) != nil {
			continue
		}
		template, ok := entry.template(loopbackFeed)
		if !ok || seen[template.ID] {
			continue
		}
		seen[template.ID] = true
		templates = append(templates, template)
	}

	sort.SliceStable(templates, func(a, b int) bool {
		return strings.ToLower(templates[a].Name) < strings.ToLower(templates[b].Name)
	})
	return templates, nil
}

func (e feedEntry) template(loopbackFeed bool) (Template, bool) {
	name := strings.TrimSpace(e.Name)
	id := strings.TrimSpace(e.NormalizedName)
	source := strings.TrimSpace(e.SourceLink)

	if name == "" || !templateID.MatchString(id) || !usableSource(source, loopbackFeed) {
		return Template{}, false
	}

	version, extended := splitHugoVersion(e.HugoVersion)
	shot := strings.ToLower(strings.TrimSpace(e.ScreenshotType))
	if !screenshotTypes[shot] {
		shot = ""
	}

	return Template{
		ID:             id,
		Name:           name,
		Description:    strings.TrimSpace(e.Description),
		Author:         strings.TrimSpace(e.Author),
		AuthorURL:      webLink(e.AuthorURL),
		License:        strings.TrimSpace(e.License),
		LicenseURL:     webLink(e.LicenseURL),
		Homepage:       webLink(e.Homepage),
		DemoURL:        webLink(e.DemoURL),
		HugoTheme:      strings.TrimSpace(e.HugoTheme),
		HugoVersion:    version,
		Extended:       extended,
		Model:          strings.Fields(e.Model),
		Forms:          formsCount(e.Forms),
		SourceURL:      source,
		ScreenshotType: shot,
	}, true
}

func splitHugoVersion(raw string) (string, bool) {
	version := strings.TrimSpace(raw)
	extended := false
	if rest, found := strings.CutPrefix(version, "extended_"); found {
		version, extended = rest, true
	}
	version = strings.TrimPrefix(version, "v")
	if !hugoVersion.MatchString(version) {
		return "", extended
	}
	return version, extended
}

func formsCount(raw json.RawMessage) int {
	var count int
	if json.Unmarshal(raw, &count) == nil && count > 0 {
		return count
	}
	return 0
}

func webLink(raw string) string {
	link := strings.TrimSpace(raw)
	parsed, err := url.Parse(link)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return ""
	}
	return link
}

func usableSource(source string, loopbackFeed bool) bool {
	parsed, err := url.Parse(source)
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return false
	}
	switch parsed.Scheme {
	case "https":
		return true
	case "http":
		return loopbackFeed && isLoopbackHost(parsed.Hostname())
	}
	return false
}

func isLoopbackHTTP(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "http" {
		return false
	}
	return isLoopbackHost(parsed.Hostname())
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func ScreenshotURL(feedURL string, template Template) (string, bool) {
	if template.ScreenshotType == "" {
		return "", false
	}
	base, err := url.Parse(feedURL)
	if err != nil {
		return "", false
	}
	relative := &url.URL{Path: "templates/" + template.ID + "/screenshot." + template.ScreenshotType}
	return base.ResolveReference(relative).String(), true
}
