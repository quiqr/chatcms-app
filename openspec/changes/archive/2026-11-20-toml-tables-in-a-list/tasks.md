## 1. Implementation

- [x] 1.1 `Normalise` turns a TOML array of tables (`[]map[string]any`) into
  a list, so `[[build]]` and `[[collections]]` are no longer refused or
  ignored; the `formats-mixed` golden now holds its TOML collection.

## 2. Verification

- [x] 2.1 `TestTOMLTablesInAListAreAList` and `go test ./...` pass; the UILite
  template from the gallery loads (12 singles, 2 collections, 1 build).
