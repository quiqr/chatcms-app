---
# chatcms-6kfu
title: preview of UILite looks broken
status: completed
type: bug
priority: normal
created_at: 2026-11-20T14:32:59Z
updated_at: 2026-11-20T21:30:42Z
openspec-link: openspec/changes/archive/2026-11-20-theme-addresses-in-the-preview
---

looks like this

/home/pim/Afbeeldingen/Schermafbeeldingen/2026-11-20-153226_hyprshot.png


should look like this:

/home/pim/Afbeeldingen/Schermafbeeldingen/2026-11-20-153352_hyprshot.png

## Summary of Changes

- Cause: UILite writes its styles and pictures as `{{.Site.BaseURL}}...`, which
  Hugo makes absolute with the base URL ChatCMS gives it
  (`http://127.0.0.1:8080/preview/...`). Opened as `localhost:5173`, the browser
  asked another host without the session and got nothing; CDN files loaded.
- The preview proxy now turns that exact absolute prefix into its path in
  HTML, CSS, JavaScript, JSON and XML, also `\/`-escaped; pictures and other
  hosts are untouched.
- e2e 55 signs in at `localhost` with the base URL on `127.0.0.1` and checks
  the style applies: passes with the fix, fails without it.
