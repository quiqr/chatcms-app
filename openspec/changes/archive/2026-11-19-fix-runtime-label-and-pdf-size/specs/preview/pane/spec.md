# Spec Delta

## ADDED Requirements

### Requirement: A PDF preview keeps its own size

A pdf preview SHALL show its pages at their own size, one PDF point to one
CSS point, whatever the width of the pane, and SHALL scroll where the pane is
smaller. It SHALL NOT offer a choice of width.

#### Scenario: a narrow pane

- **WHEN** an A5 PDF is previewed in a pane narrower than the page
- **THEN** the page SHALL keep its size and the pane SHALL scroll
