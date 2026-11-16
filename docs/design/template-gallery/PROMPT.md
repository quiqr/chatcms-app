# Prompt for Claude Code

Put this folder in the repo as `docs/design/template-gallery/`, then paste:

```
We're adding "Start from a template" to project creation in Quiqr ChatCMS.
Read docs/design/template-gallery/TEMPLATE-GALLERY.md fully first. Reference
screens: docs/design/template-gallery/reference/screenshots/ (desktop 1440,
phone 390) and reference/html/ (exact inline styles). Use our tokens and
shadcn components, not the inline hex values.

Build in this order, and stop after each step to show me screenshots next
to the matching reference:
1. Backend: fetch + cache + validate the Quiqr community templates feed and
   proxy the screenshots; GET /api/templates (§1). Include a test with a
   recorded copy of templates.json.
2. Gallery page with search, loading, error and no-match states (§2), as the
   third tab of New project.
3. Step 2 "Name & repository" with validation and limit checks (§3).
4. Wire it into the existing onboarding: deploy key → copy the template into
   the empty repository with a fresh history → publishing → invite people.
5. Phone layout (§4).

Finish with the acceptance checks in §7 and list any that fail.
```
