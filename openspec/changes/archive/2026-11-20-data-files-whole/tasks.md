# Tasks

## 1. Writing

- [x] 1.1 A JSON emitter matching `JSON.stringify(value, null, 2)`; proven
  against every `.json.golden`.
- [x] 1.2 A TOML emitter porting tomlify-j0.4 with `space: 2`; proven against
  every `.toml.golden`.

## 2. Reading and the document

- [x] 2.1 TOML is read with key order kept at every depth and arrays of
  tables as lists.
- [x] 2.2 `Document.Whole`, `ParseData` and `Bytes` for whole files;
  `forms.Load` picks by extension; proven by round-trip tests per format.

## 3. Through the product

- [x] 3.1 A content test: a TOML config single and a YAML data single read
  their values and save without front matter.
- [x] 3.2 e2e: a JSON data single is edited and saved, and the remote holds
  the reference's bytes.

## 4. Ship

- [x] 4.1 `CHANGELOG.md` under [Unreleased].
- [x] 4.2 `scripts/gate.sh ship` green; close chatcms-hf29.
