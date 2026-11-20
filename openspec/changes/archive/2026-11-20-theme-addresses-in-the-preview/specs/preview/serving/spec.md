## ADDED Requirements

### Requirement: A theme's own absolute addresses work under any name

The preview SHALL turn the absolute address it gave the site generator as
its base URL into a path, in the pages, styles, scripts and feeds it serves,
so a theme that writes its addresses with `{{ .Site.BaseURL }}` shows its
styles and pictures however the person reached the app. Other addresses and
binary files SHALL be left as they are.

#### Scenario: the app opened under another name

- **WHEN** the app's base URL is `http://127.0.0.1:8080` and a person opens it
  as `http://localhost:5173`
- **AND** a theme links `{{ .Site.BaseURL }}css/styles.css`
- **THEN** the preview SHALL link `/preview/<organisation>/<project>/css/styles.css`
  and the style SHALL apply
