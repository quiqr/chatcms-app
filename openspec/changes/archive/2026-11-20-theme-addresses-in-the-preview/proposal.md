# Proposal

Bean: [chatcms-6kfu](../../../.beans/chatcms-6kfu--preview-of-uilite-looks-broken.md) (preview of UILite looks broken).

## Why

UILite writes its styles and pictures as `{{.Site.BaseURL}}css/styles.css`.
Hugo makes that absolute with the base URL ChatCMS gives it, for example
`http://127.0.0.1:8080/preview/…`. Opened under another name, such as
`localhost:5173` in development or another host name in production, the
browser asks another site, without the session, and gets nothing.

## What Changes

- The preview proxy replaces that exact absolute prefix with its path in
  HTML, CSS, JavaScript, JSON and XML responses, also in its `\/`-escaped
  form.

## Capabilities

### New Capabilities

### Modified Capabilities

- `preview/serving`: a theme's absolute addresses work under any name.
