package runtime

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

const quarto = `
packages:
  quarto: "1.6.40"
  typst: "0.12.0"
preview:
  type: pdf
  open: /_book/book.pdf
  command: [quarto, preview, "{source}", --port, "{port}", --host, "{host}", --no-browser]
build:
  output: _book
  command: [quarto, render, "{source}"]
`

func TestADeclaredRuntimeIsRead(t *testing.T) {
	definition, err := Parse([]byte(quarto))
	if err != nil {
		t.Fatal(err)
	}
	if definition.Packages["quarto"].Version != "1.6.40" || definition.Preview.Type != PDF || definition.Preview.Open != "/_book/book.pdf" || definition.Build.Output != "_book" {
		t.Errorf("Parse() = %+v", definition)
	}
}

func TestDefaultsFillTheGaps(t *testing.T) {
	definition, err := Parse([]byte("packages: {hugo: \"0.145.0\"}\npreview:\n  command: [hugo, server]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if definition.Preview.Type != Web || definition.Preview.Open != "/" || definition.Build != nil {
		t.Errorf("Parse() = %+v", definition)
	}
}

func TestAWrongDefinitionSaysWhatIsWrong(t *testing.T) {
	for text, want := range map[string]string{
		"packages: [": "cannot be read",
		"packages: {hugo: \"1\"}\npreview: {command: [hugo]}\ncolour: red\n":                            "cannot be read",
		"preview: {command: [hugo]}\n":                                                                  "no packages",
		"packages: {\"../x\": \"1\"}\npreview: {command: [hugo]}\n":                                     "not a package name",
		"packages: {hugo: \"1; rm\"}\npreview: {command: [hugo]}\n":                                     "not a version",
		"packages: {hugo: \"1\"}\npreview: {type: video, command: [hugo]}\n":                            "not one of web, pdf or other",
		"packages: {hugo: \"1\"}\npreview: {open: /../etc, command: [hugo]}\n":                          "stay inside",
		"packages: {hugo: \"1\"}\npreview: {open: x, command: [hugo]}\n":                                "must start with /",
		"packages: {hugo: \"1\"}\npreview: {}\n":                                                        "preview command is empty",
		"packages: {hugo: \"1\"}\npreview: {command: [/bin/sh]}\n":                                      "program's name",
		"packages: {hugo: \"1\"}\npreview: {command: [hugo, \"{secret}\"]}\n":                           "{secret}",
		"packages: {hugo: \"1\"}\npreview: {command: [hugo]}\nbuild: {output: /etc, command: [hugo]}\n": "inside the site",
		"packages: {hugo: \"1\"}\npreview: {command: [hugo]}\nbuild: {output: ../x, command: [hugo]}\n": "inside the site",
		"packages: {hugo: \"1\"}\npreview: {command: [hugo]}\nbuild: {output: public}\n":                "build command is empty",
	} {
		_, err := Parse([]byte(text))
		if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), want) {
			t.Errorf("Parse(%q) error = %v, want one saying %q", text, err, want)
		}
	}
}

func TestASiteWithoutADefinitionGetsTodaysHugo(t *testing.T) {
	definition, declared, err := Read(fstest.MapFS{}, "0.139.0")
	if err != nil || declared || definition.Packages["hugo"].Version != "0.139.0" || definition.Preview.Command[0] != "hugo" {
		t.Errorf("Read() = %+v, %v, %v", definition, declared, err)
	}
	site := fstest.MapFS{File: {Data: []byte(quarto)}}
	definition, declared, err = Read(site, "0.139.0")
	if err != nil || !declared || definition.Packages["quarto"].Version == "" {
		t.Errorf("Read() of a declared site = %+v, %v, %v", definition, declared, err)
	}
	if _, _, err := Read(fstest.MapFS{File: {Data: []byte("nonsense: [")}}, ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("Read() of a broken file error = %v", err)
	}
}

type pins map[string]string

func (p pins) Path(ctx context.Context, name, version, revision string) (string, error) {
	key := name + "@" + version
	if revision != "" {
		key += "@" + revision
	}
	path, found := p[key]
	if !found {
		return "", errors.New("no such version: " + name + " " + version)
	}
	return path, nil
}

func (p pins) Environment(ctx context.Context, kind, revision string, packages []string) (string, error) {
	path, found := p[kind+"@"+revision+"@"+strings.Join(packages, ",")]
	if !found {
		return "", errors.New("no such environment: " + kind)
	}
	return path, nil
}

func inStore(paths ...string) func(string) bool {
	return func(path string) bool {
		for _, known := range paths {
			if known == path {
				return true
			}
		}
		return false
	}
}

func TestPackagesAndProgramsAreFound(t *testing.T) {
	definition, _ := Parse([]byte(quarto))
	resolver := Resolver{
		Pins:   pins{"quarto@1.6.40": "/nix/store/q-quarto", "typst@0.12.0": "/nix/store/t-typst"},
		Exists: inStore("/nix/store/q-quarto/bin/quarto"),
	}
	resolved, err := resolver.Resolve(t.Context(), definition, true)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Preview.Program != "/nix/store/q-quarto/bin/quarto" || resolved.Build.Program != resolved.Preview.Program || resolved.Output != "_book" || !resolved.Declared {
		t.Errorf("Resolve() = %+v", resolved)
	}
	if strings.Join(resolved.Path, ":") != "/nix/store/q-quarto/bin:/nix/store/t-typst/bin" {
		t.Errorf("Path = %v", resolved.Path)
	}
}

func TestAProgramNoPackageProvidesIsRefused(t *testing.T) {
	definition, _ := Parse([]byte("packages: {typst: \"0.12.0\"}\npreview: {command: [bash, -c, x]}\n"))
	resolver := Resolver{Pins: pins{"typst@0.12.0": "/nix/store/t-typst"}, Exists: inStore()}
	if _, err := resolver.Resolve(t.Context(), definition, true); !errors.Is(err, ErrNoProgram) || !strings.Contains(err.Error(), "bash") {
		t.Errorf("Resolve() error = %v, want the program named", err)
	}
}

func TestWithoutNixOnlyTheInstallationsHugoIsThere(t *testing.T) {
	resolver := Resolver{InstalledHugo: "/usr/local/bin/hugo", Exists: inStore("/usr/local/bin/hugo")}
	resolved, err := resolver.Resolve(t.Context(), Default("0.139.0"), false)
	if err != nil || resolved.Preview.Program != "/usr/local/bin/hugo" {
		t.Errorf("Resolve() = %+v, %v", resolved, err)
	}
	quartoDefinition, _ := Parse([]byte(quarto))
	if _, err := resolver.Resolve(t.Context(), quartoDefinition, true); !errors.Is(err, ErrNoPackages) {
		t.Errorf("Resolve() of quarto without Nix error = %v", err)
	}
	if _, err := (Resolver{}).Resolve(t.Context(), Default(""), false); !errors.Is(err, ErrNoHugo) {
		t.Errorf("Resolve() with no Hugo at all error = %v", err)
	}
}

func TestAVersionThatCannotBeHadIsPassedOn(t *testing.T) {
	definition, _ := Parse([]byte(quarto))
	if _, err := (Resolver{Pins: pins{}}).Resolve(t.Context(), definition, true); err == nil || !strings.Contains(err.Error(), "no such version") {
		t.Errorf("Resolve() error = %v", err)
	}
}

func TestPlaceholdersAreFilledWordByWord(t *testing.T) {
	filled := Fill([]string{"server", "--port", "{port}", "--baseURL", "{url}", "--liveReloadPort", "{reload}", "hugo; rm -rf /", "x{port}"}, map[string]string{
		"{port}": "1313",
		"{url}":  "http://h/v/key/",
	})
	want := "server --port 1313 --baseURL http://h/v/key/ hugo; rm -rf / x{port}"
	if strings.Join(filled, " ") != want {
		t.Errorf("Fill() = %q, want %q", strings.Join(filled, " "), want)
	}
}

const rev = "02c087f3fe673c9e317f206a627cb6b910336dbe"

func TestAPackageMayNameItsRevision(t *testing.T) {
	definition, err := Parse([]byte("packages:\n  quarto: {version: \"1.10.18\", nixpkgs: " + rev + "}\n  typst: \"0.15.1\"\npreview: {command: [quarto, preview, \"{source}\"]}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if definition.Packages["quarto"] != (Package{Version: "1.10.18", Nixpkgs: rev}) || definition.Packages["typst"] != (Package{Version: "0.15.1"}) {
		t.Errorf("Packages = %+v", definition.Packages)
	}
	resolver := Resolver{
		Pins:   pins{"quarto@1.10.18@" + rev: "/nix/store/q-quarto-1.10.18", "typst@0.15.1": "/nix/store/t-typst-0.15.1"},
		Exists: inStore("/nix/store/q-quarto-1.10.18/bin/quarto"),
	}
	resolved, err := resolver.Resolve(t.Context(), definition, true)
	if err != nil || resolved.Preview.Program != "/nix/store/q-quarto-1.10.18/bin/quarto" || !resolved.Rooted {
		t.Errorf("Resolve() = %+v, %v", resolved, err)
	}
}

func TestARevisionMustBeAWholeCommit(t *testing.T) {
	for _, bad := range []string{
		"packages: {quarto: {version: \"1.10.18\", nixpkgs: main}}\npreview: {command: [quarto]}\n",
		"packages: {quarto: {version: \"1.10.18\", channel: x}}\npreview: {command: [quarto]}\n",
		"packages: {quarto: \"1.10.18\"}\npython: {packages: [pandas]}\npreview: {command: [quarto]}\n",
		"packages: {quarto: \"1.10.18\"}\npython: {nixpkgs: " + rev + ", packages: [\"pandas; rm\"]}\npreview: {command: [quarto]}\n",
	} {
		if _, err := Parse([]byte(bad)); !errors.Is(err, ErrInvalid) {
			t.Errorf("Parse(%q) error = %v", bad, err)
		}
	}
}

func TestAProvidedPackageOfTheRightVersionComesFirst(t *testing.T) {
	definition, _ := Parse([]byte("packages: {quarto: \"1.10.18\"}\npreview: {command: [quarto, preview]}\n"))
	resolver := Resolver{
		Provided: map[string]string{"quarto": "/nix/store/p-quarto-1.10.18"},
		Exists:   inStore("/nix/store/p-quarto-1.10.18/bin/quarto"),
	}
	resolved, err := resolver.Resolve(t.Context(), definition, true)
	if err != nil || resolved.Preview.Program != "/nix/store/p-quarto-1.10.18/bin/quarto" {
		t.Errorf("Resolve() = %+v, %v", resolved, err)
	}

	resolver.Provided["quarto"] = "/nix/store/p-quarto-1.9.37"
	if _, err := resolver.Resolve(t.Context(), definition, true); !errors.Is(err, ErrNoPackages) {
		t.Errorf("another version was used: %v", err)
	}
}

func TestCellPackagesBecomeQuartosPythonAndR(t *testing.T) {
	definition, err := Parse([]byte("packages: {quarto: \"1.10.18\"}\npython: {nixpkgs: " + rev + ", packages: [pandas]}\nr: {nixpkgs: " + rev + ", packages: [ggplot2]}\npreview: {command: [quarto, preview]}\n"))
	if err != nil {
		t.Fatal(err)
	}
	resolver := Resolver{
		Pins: pins{
			"quarto@1.10.18":            "/nix/store/q-quarto-1.10.18",
			"python@" + rev + "@pandas": "/nix/store/py-env",
			"r@" + rev + "@ggplot2":     "/nix/store/r-env",
		},
		Exists: inStore("/nix/store/q-quarto-1.10.18/bin/quarto"),
	}
	resolved, err := resolver.Resolve(t.Context(), definition, true)
	if err != nil || resolved.Env["QUARTO_PYTHON"] != "/nix/store/py-env/bin/python3" || resolved.Env["QUARTO_R"] != "/nix/store/r-env/bin/R" {
		t.Errorf("Resolve() = %+v, %v", resolved, err)
	}
	if _, err := (Resolver{Provided: map[string]string{"quarto": "/nix/store/p-quarto-1.10.18"}}).Resolve(t.Context(), definition, true); !errors.Is(err, ErrNoPackages) {
		t.Errorf("cells without Nix error = %v", err)
	}
}

func TestAPreviewToldItsAddressIsNotRooted(t *testing.T) {
	resolved, err := (Resolver{InstalledHugo: "/bin/hugo", Exists: inStore("/bin/hugo")}).Resolve(t.Context(), Default(""), false)
	if err != nil || resolved.Rooted {
		t.Errorf("Hugo Rooted = %v, %v", resolved.Rooted, err)
	}
}
