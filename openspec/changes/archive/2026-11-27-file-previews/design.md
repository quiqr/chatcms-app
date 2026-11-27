# Design

## The kinds

One table in `internal/content` maps an extension to a preview kind and a
content type: `png jpg jpeg gif webp avif bmp ico` as image, `mp4 webm ogv
mov` as video, `mp3 wav oga ogg m4a flac` as audio, `pdf` as pdf. `svg` is
left out: it is text and opens in the editor, and served as itself it could
run script.

## Reading stays the same, with one field more

`ReadFile` sets `Preview` on a file that is binary or too large when its
extension is in the table. The screen picks the preview by that field, so
the browser never guesses from a name.

## Serving the bytes

`OpenFile` returns the file as an `io.ReadSeeker` with its modification time,
so `http.ServeContent` answers range requests. A file that cannot seek is
read into memory. The site is released when the handler closes the file.

Headers: the kind's content type, `nosniff`, `Content-Disposition: inline`
with the name, `Cache-Control: private, max-age=0`, and
`Content-Security-Policy: sandbox` for every kind but PDF.
