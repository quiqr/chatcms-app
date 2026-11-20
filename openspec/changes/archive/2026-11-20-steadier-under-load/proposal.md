# Proposal

Beans: [chatcms-0xlf](../../../.beans/chatcms-0xlf--signing-out-sometimes-lands-on-the-sign-in-screen.md),
[chatcms-v58o](../../../.beans/chatcms-v58o--a-preview-of-ones-own-sometimes-stays-empty-on-a-p.md),
[chatcms-1jm8](../../../.beans/chatcms-1jm8--an-editor-undo-test-times-out-under-a-full-vitest.md).

## Why

Three tests failed now and then on a busy machine, and two of them because
the product itself raced: signing out could land on the sign-in screen
without saying so, and a preview page asked for while the site was still
being built stayed empty.

## What Changes

- Signing out stops refused requests from redirecting.
- An empty preview page is loaded again, a few times at most.
- The undo tests no longer wait real seconds between saves.

## Capabilities

### New Capabilities

### Modified Capabilities

- `identity/authentication`: signing out always says so.
- `preview/pane`: an empty page is loaded again.
