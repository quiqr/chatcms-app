---
# chatcms-u5zn
title: Picture thumbnails use the site address, which the app does not serve
status: todo
type: bug
priority: normal
created_at: 2026-11-19T19:09:09Z
updated_at: 2026-11-19T19:09:09Z
parent: chatcms-iqog
---

A picture field in a form, and a declared shortcode's block in the page text,
show a thumbnail with `<img src="/images/loaf.png">`. That is the address on
the site, not one the app serves, so the thumbnail does not load. The block
now leaves a picture out when it does not load; the form shows a broken image.

The media library already shows thumbnails through its own address; the
fields should use that.
