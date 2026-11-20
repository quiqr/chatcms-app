## 1. Implementation

- [x] 1.1 Publishing built pages records where the site's own files stood in
  `refs/chatcms/published-source/<branch>`, and the review and the count of
  what is waiting compare against it, so a site whose own files never reach
  the remote no longer says everything is waiting forever.

## 2. Verification

- [x] 2.1 `TestPublishingPagesSaysWhatIsSavedSinceIsPublished`; `go test
  ./internal/gitsync/ ./internal/publish/ ./internal/app/` pass.
