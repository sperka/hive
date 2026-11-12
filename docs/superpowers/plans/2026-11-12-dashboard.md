# Dashboard implementation and verification

**Goal:** Ship the approved dashboard in v0.3.0.

**Architecture:** Keep Bubble Tea and Lip Gloss. Separate CPU counter sampling
from rendering and isolate optional Kitty graphics behind a text fallback.

**Spec:** `docs/superpowers/specs/2026-11-12-dashboard-design.md`

## Completed implementation

- [x] Keep the outer dashboard at 100 columns, including borders, with compact
  layouts for smaller terminals.
- [x] Place Sessions and Details side by side, with CPU activity directly below.
- [x] Use the original 3-4-3 text hexagons and a bold, non-italic HIVE/version
  label at the top right.
- [x] Right-align the session total and single-line state breakdown on the next
  two rows without shifting the panels.
- [x] Sample real CPU counters asynchronously and retain at most 60 readings.
  Distinguish first, failed, and valid zero-usage samples.
- [x] Show physical/logical core counts, usage history, and subdued
  average/peak/scale/sample information.
- [x] Align the CPU title, statistics, and grid at a shared left edge.
  Distribute unused space between four or eight columns.
- [x] Number cores down each column, preserving percentage associations,
  uneven counts, and hidden-core totals in short panels.
- [x] Use green/yellow/red percentages at 50%/80% thresholds and `--` when unknown.
- [x] Keep text graphics as the default, with an explicit Kitty image override.
  Preserve original spacing and only clean up Hive's own image.
- [x] Preserve existing session actions, themes, CLI commands, and switcher
  behavior; keep selection valid after session refresh.

## Verification completed during implementation

- [x] Check frame dimensions, long Unicode fields, empty selection, and compact
  viewport bounds.
- [x] Check sampling during overlays, counter reset/error handling, real history,
  and physical/logical counts.
- [x] Check column order and usage associations for full and uneven grids,
  height limits, unavailable values, and 100.0% readings.
- [x] Check optional graphics selection, PNG geometry, passthrough, theme
  changes, and image-specific cleanup.
- [x] Inspect terminal previews at wide, narrow, and short sizes.
- [x] Run the Go test suite and vet after the final layout changes.

## Release

Consolidate the local dashboard work and release preparation into one commit
above v0.2.0. Follow `RELEASING.md` for race tests, installer fixtures, all four
platform builds, draft artifact verification, publication, and a real upgrade
from a temporary v0.2.0 installation. Leave the user's installed binary unchanged.
