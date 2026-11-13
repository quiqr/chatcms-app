# Design

## Context

Tailwind 4 changed its preflight so that `button` gets `cursor: default`.
Our `Button` renders a `<button>`, or, with `asChild`, a link. The only
`cursor-pointer` in the frontend is on the photo upload label, and `MenuItem`
sets `cursor-default` explicitly.

## Goals / Non-Goals

**Goals:**
- The hand over every enabled control, built either way.
- A regression is caught by the gate, not by somebody noticing.

**Non-Goals:**
- Hover colours; they are a separate matter per component.

## Decisions

**One base rule, not a class per component.** A rule in `@layer base` keyed
on elements and ARIA roles covers what Radix renders (`role="menuitem"`,
`role="radio"` on a button, `role="tab"`) without touching each component,
and a component written later gets it for free. It sits in the base layer,
so a utility class on a component can still override it where there is a
reason.

**Disabled is excluded in the selector.** `:not(:disabled)` for native
controls, `[data-disabled]` for Radix items, `aria-disabled` for roles.

**The check forbids the arrow instead of requiring the hand.** Requiring a
class on every clickable thing cannot be checked from source text reliably;
forbidding `cursor-default` and `cursor-auto` can, and those are the only way
a component can undo the base rule.

**The acceptance test measures the computed cursor.** jsdom does not load the
stylesheet, so only a real browser can prove the rule applies.

## Risks / Trade-offs

- [A control built from a `div` with an `onClick` and no role misses the
  rule] → Such a control is also unreachable by keyboard, which the design
  system already forbids; giving it a role fixes both.
