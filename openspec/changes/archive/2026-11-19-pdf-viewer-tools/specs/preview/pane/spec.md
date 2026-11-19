## MODIFIED Requirements

### Requirement: A PDF preview keeps its own size

A pdf preview SHALL start with its pages at their own size, one PDF point to
one CSS point, whatever the width of the pane, and SHALL scroll where the
pane is smaller. On a phone it SHALL instead start with each page scaled down
to the width of the screen less a margin, keeping its proportions, and never
past its own size. The pages SHALL stand on a grey ground, so that the edges
of the page format show. A different size SHALL only come from the viewer's
zoom, which keeps the page's proportions; the pane SHALL NOT reflow a PDF to
its width.

#### Scenario: a narrow pane

- **WHEN** an A5 PDF is previewed in a pane narrower than the page
- **THEN** the page SHALL keep its size and the pane SHALL scroll

#### Scenario: on a phone

- **WHEN** a PDF is previewed on a phone 390 px wide
- **THEN** the whole width of the page SHALL show, with the grey ground
  around it, and its proportions SHALL be those of the PDF
