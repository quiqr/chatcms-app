package preview

import "testing"

func TestTheAddressIsAPathOnTheAppsOwnAddress(t *testing.T) {
	s := &Service{baseURL: "http://127.0.0.1:8080"}
	for given, want := range map[string]string{
		"http://127.0.0.1:8080/preview/technative/korenaar/":        "/preview/technative/korenaar/",
		"http://127.0.0.1:8080/preview/technative/korenaar/p/site/": "/preview/technative/korenaar/p/site/",
		"": "",
	} {
		if got := s.addressOf(given); got != want {
			t.Errorf("addressOf(%q) = %q, want %q", given, got, want)
		}
	}
	s.ServeFrom("http://localhost:8080")
	if got := s.addressOf("http://localhost:8080/v/k1/"); got != "http://localhost:8080/v/k1/" {
		t.Errorf("on a preview origin addressOf() = %q", got)
	}
}
