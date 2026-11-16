# Design

## Context

`scripts/dev-instance.sh` already sets `inbox_port`. `DevSignIn.tsx` is
loaded lazily behind `import.meta.env.DEV`, so a production build leaves it
out. Vite exposes only `VITE_`-prefixed variables to the app, unless
`vite.config.ts` defines them.

## Goals / Non-Goals

**Goals:**

- One component for the link, used by both screens.

**Non-Goals:**

- Checking whether Mailpit is running.
- Proxying Mailpit through Vite.

## Decisions

### `dev-fe.sh` exports `CHATCMS_INBOX_PORT`, `vite.config.ts` defines it

`vite.config.ts` sets `import.meta.env.VITE_CHATCMS_INBOX_PORT` from
`CHATCMS_INBOX_PORT`, defaulting to `8025`. All other settings go through the
`CHATCMS_` names, and this keeps that true. Using a bare `VITE_` variable in
the script would add a second naming scheme.

### The link takes its host from `window.location.hostname`

A phone that opened the frontend at the machine's address reaches Mailpit at
the same address, because `dev-be.sh` binds Mailpit to `CHATCMS_HOST`.
Passing the host from the script would break when someone opens the page
under a different name than the one the script knows.

### `DevMailLink.tsx` is loaded lazily behind `import.meta.env.DEV`

`DevSignIn` imports it directly, since that block is already lazy.
`ForgotPasswordPage` loads it with `lazy()` behind `import.meta.env.DEV`,
the same way `SignInPage` loads `DevSignIn`. The block on the forgot-password
screen uses the same dashed development styling, so it is clear it is not part
of the product.

## Risks / Trade-offs

- [Mailpit not running, so the link leads nowhere] → The browser says so. It
  is a development machine.
