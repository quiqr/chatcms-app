## 1. Implementation

- [x] 1.1 `relative` in `internal/preview/decorate.go` rewrites the base URL
  and, when apart, the preview origin plus the prefix, before the overlay is
  injected. Bean chatcms-6kfu.

## 2. Verification

- [x] 2.1 Go test for pages, JSON, a picture and other hosts; e2e 55 signs in
  at `localhost` with the base URL on `127.0.0.1`: passes 4 of 4 with the
  fix, fails without it on the absolute address.
