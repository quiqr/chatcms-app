# Tasks

## 1. The server

- [x] 1.1 `Chat.Stop` aborts the person's OpenCode session; `POST …/chat/stop`;
  proven by Go tests.

## 2. The chat

- [x] 2.1 `ChatPanel` keeps an id and a state per request (working, done,
  question), reports them, undoes a request in one step, stops one; questions
  count on the badge; proven by Vitest.

## 3. The overlay

- [x] 3.1 Pills drawn from `pills`, placed at their anchor, with Stop, Undo,
  Show in chat and Answer in chat; the success tint; done pills stay 8 s and
  until the next touch; proven by Vitest.
- [x] 3.2 `Preview` posts the pills on change and on every ready; passes the
  actions up; proven by Vitest.
- [x] 3.3 Strings in `en.ts` and `nl.ts`.

## 4. Proof and ship

- [x] 4.1 e2e 48: Working then Changed with Undo at the selection, Undo puts
  the title back, two requests give two pills and a count; screenshots next to
  references 04 to 07.
- [x] 4.2 `CHANGELOG.md` under [Unreleased].
- [x] 4.3 `scripts/gate.sh ship` green; close chatcms-6s8v.
