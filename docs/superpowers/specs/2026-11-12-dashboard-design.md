# Dashboard redesign

Implement the requested 100-column dashboard for terminal use, particularly
Ghostty. The outer border is included in the 100 columns. Preserve existing
keyboard actions, themes, session lifecycle, CLI commands, and switcher behavior.

## Layout

```text
+--------------------------------------------------------------------------------------------------+
| [hexagons]  HIVE              [session count and status counts]                                    |
|                                                                                                  |
| +--------------------------------------------+  +--------------------------------------------+   |
| | SESSIONS                                   |  | DETAILS                                    |   |
| | selectable session list                    |  | selected session metadata                  |   |
| +--------------------------------------------+  +--------------------------------------------+   |
|                                                                                                  |
| +--------------------------------------------------------------------------------------------+   |
| | CPU: current usage                         physical cores / logical CPUs                    |   |
| | real usage history, average and peak                                                       |   |
| | per-logical-CPU utilization                                                                |   |
| +--------------------------------------------------------------------------------------------+   |
+--------------------------------------------------------------------------------------------------+
```

The wordmark is compact text, replacing the six-row banner. Session and detail
panels have matching heights; details remain present with an empty state when
there is no selection. Width calculations include borders and padding. Short
and narrow terminals use a compact arrangement without overflowing the viewport.
Long or Unicode names and paths must not push panels outside the frame.

## CPU

Replace fabricated startup history with a collecting-data state. Sample real CPU
counters once per second in a Bubble Tea command, outside Update/View. Show
aggregate utilization, real history with average/peak, and per-logical-CPU usage.
Show physical cores and logical CPUs separately; report an unavailable count
honestly. A failed sample is not an idle sample. Bound the history to 60 samples.
Compact terminals may omit detailed charts while retaining current usage and
core counts.

## Graphics

Use the original font-rendered hexagons by default in every terminal. For the
explicit `HIVE_GRAPHICS=kitty` override, use Kitty Unicode placeholders for a
small transparent hexagon mark so it moves and disappears with the text layout.
Send a generated PNG directly, in quiet mode,
using the existing Charm ANSI library. No downloaded assets or runtime dependencies.
Use a per-process image ID, reuse it for theme changes, and delete only that image
on exit. Through tmux,
require existing passthrough support and wrap transmission in tmux passthrough.
Do not change the user's tmux configuration. Provide HIVE_GRAPHICS=off and
HIVE_GRAPHICS=kitty overrides. Keep CPU charts as terminal text.

## Approaches

- Chosen: Bubble Tea/Lip Gloss layout with a small optional Kitty mark. Retains
  terminal navigation and supports both Ghostty and ordinary terminals.
- Text only: simpler, but does not use the graphics capability the user raised.
- Rasterize the dashboard: would complicate selection, text alignment, themes,
  and tmux behavior without helping the requested information layout.

The requested arrangement is settled. Compact-width behavior, physical versus
logical CPU labels, and limiting graphics to the mark are implementation choices.
The completed dashboard ships in v0.3.0, with installation and rollback covered
in `RELEASING.md`.

## Readability refinement

The follow-up keeps the 100-column frame and side-by-side panels. Render HIVE
as bold text without italics, followed by one space and the full version.
Align this group to the top right beside the original 3-4-3 honeycomb, with its
original staggered spacing and a blank row above and below the mark.
Keep HIVE/version, the right-aligned session total, and the right-aligned state
breakdown on their existing screen rows, beside the three hexagon rows.
Use compact icon/count labels in narrow headers.
Keep the session/detail panels and CPU panel at their existing vertical
positions with no gap between them. The optional Kitty mark matches.

Render CPU average/peak/scale/sample text without bold and with faint styling.
Replace per-core bars with
`#1 23.4%` entries: eight columns when they fit, otherwise four. Keep one
decimal place and align percentages. Color percentages green below 50%,
yellow from 50%, and red from 80%. Unknown values remain `--`.
Align the core grid with the CPU title, graph, and statistics at one shared
left edge. Distribute spare width between the columns, keeping their positions
fixed as readings change and reserving enough room for 100.0%.
Number cores down each column before continuing in the next. Balance column
heights when the count is uneven and retain the hidden-core count in short panels.
Keep the existing history sampling and viewport bounds.
