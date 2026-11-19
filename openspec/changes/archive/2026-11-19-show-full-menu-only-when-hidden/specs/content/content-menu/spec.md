## MODIFIED Requirements

### Requirement: A location picker reaches any page without the column

On screens wider than a phone, the collection view and the editor SHALL show a
breadcrumb `Content / <group> / <current>`, whose last part opens a picker,
whether the column is folded or not. Choosing an entry in the picker SHALL
open it and close the picker. While the column is hidden, folded to its strip
or shown as a drawer on a narrower screen, the picker SHALL offer "Show the
full content menu", which unfolds the column. While the column is open, the
picker SHALL NOT offer it.

#### Scenario: Switching while folded

- **WHEN** the column is folded and somebody opens the picker and chooses
  News
- **THEN** News opens and the picker closes, and the column stays folded

#### Scenario: The column is already open

- **WHEN** the column is open and somebody opens the picker
- **THEN** the picker does not offer "Show the full content menu"
