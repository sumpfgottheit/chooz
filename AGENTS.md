# AGENTS.md — `chooz`

Authoritative build spec for an AI coding agent (Codex). Implement the
tool exactly as described here. When something is unspecified, prefer the
simplest solution that satisfies the **Acceptance criteria** at the bottom, and
keep the architecture open for the items under **Out of scope (v1)**.

---

## 1. What `chooz` is

`chooz` is a small, single, statically linked Go binary: an interactive
terminal menu driven by a YAML file. Think **`gum choose` + `fzf --preview`**:
a selectable list on one side, the (multi-line) description of the highlighted
item in a preview pane, and on selection the chosen item's `name` is printed to
**stdout** so it composes in shell scripts.

It also supports **non-interactive** selection (justfile-style): pass the item
name on the command line and `chooz` validates and echoes it without ever
drawing a TUI. This makes it usable in CI / non-TTY contexts.

Pronounced "choose". The binary is `chooz`.

---

## 2. Tech stack (fixed)

- **Language:** Go (use the latest stable available; target the version pinned
  in `go.mod`, `1.24`+).
- **TUI:** [`charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea)
  runtime, [`charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss)
  for styling/adaptive colors, [`charmbracelet/bubbles`](https://github.com/charmbracelet/bubbles)
  `list` + `viewport` components.
- **YAML:** `gopkg.in/yaml.v3`.
- **CLI:** `github.com/spf13/cobra` (single root command with positional args;
  gives `--help`, `--version`, and shell completion for free).
- **TTY detection:** `golang.org/x/term` (`term.IsTerminal`).
- All dependencies are pure Go — **no CGO**. `CGO_ENABLED=0` must always build.

---

## 3. YAML input schema (authoritative)

```yaml
# All top-level keys are OPTIONAL except `items`.
title: "Select a target environment"   # optional menu header
description: "Pick where to deploy."    # optional menu-level help (multi-line ok)

items:                                  # REQUIRED, must be non-empty
  - name: prod                          # REQUIRED, see rules below
    title: "Production"                 # optional display label; fallback = name
    description: |                      # optional, multi-line allowed
      Live cluster. Be careful.
      Multiple lines are rendered and wrapped in the preview pane.
  - name: stage
    title: "Staging"
    description: "Pre-production."
  - name: dev                           # minimal item: only name

theme:                                  # optional, see §6
  cursor: "212"
```

### 3.1 Validation rules

- `items` must be present and non-empty → otherwise error (exit 1).
- Each item's **`name` is mandatory** and must match `^[A-Za-z0-9]+$`
  (ASCII letters and digits only — no spaces, hyphens, underscores, unicode).
- `name` values must be **unique** across all items → duplicate is an error.
- `title` and `description` are optional. Display label = `title` if set, else `name`.
- **Unknown keys (top-level or per-item) are ignored, not errors.** This is a
  hard requirement: future versions add fields, and old binaries must keep
  working on newer YAML. (Implement decoding so unknown fields are silently
  dropped; do NOT use `yaml.v3`'s `KnownFields(true)`.)

---

## 4. CLI interface

```
chooz [flags] <file.yaml> [name]
```

### Positional args
- `<file.yaml>` — **required**. Path to the menu YAML. The literal `-` reads
  YAML from **stdin** (nice-to-have; implement if cheap).
- `[name]` — **optional**. If given, this is a **non-interactive selection**:
  validate that an item with this `name` exists, print it to stdout, exit 0.
  If it does not exist, error to stderr and exit 1. No TUI is drawn.

### Flags
- `-d, --default <name>` — **required feature**. Dual role:
  1. **Interactive (TTY):** the item with this `name` is pre-highlighted on
     startup.
  2. **Non-interactive (no TTY) and no positional `[name]` given:** this value
     is used as the selection (the CI fallback). It is validated like a normal
     selection. *(Design decision; if it should only ever pre-highlight and
     never auto-select, that is a one-line change — but implement the dual role
     as specified.)*
  An unknown `--default` value is an error (exit 1).
- `-l, --list` — print every item's `name`, one per line, to stdout, and exit 0.
  No TUI. (Order = YAML order.)
- `-n, --non-interactive` — never draw a TUI. Requires a resolvable selection
  (positional `[name]` or `--default`); otherwise error (exit 1).
- `--height <n>` — optional max number of visible list rows (gum-style).
- `--no-color` — disable all styling. Also honor the `NO_COLOR` env var and
  `CLICOLOR_FORCE`.
- `-h, --help`, `--version`.

### TTY auto-detection
- If **stdout is not a terminal** (piped/redirected) **and** no positional
  `[name]` and no `--default` were given → do **not** draw a TUI; error to
  stderr and exit 1 with a clear message ("not a terminal; pass a name or
  --default").
- If not a TTY but `--default` (or positional `[name]`) resolves → print that
  selection and exit 0. This is what makes `chooz` CI-friendly.
- The TUI may render to `/dev/tty` when available so it never pollutes stdout
  (see I/O contract).

---

## 5. I/O contract (critical — get this right)

This is the single most important behavioral requirement. Mirror `fzf`/`gum`:

- The **TUI is rendered to stderr** (Bubble Tea: `tea.WithOutput(os.Stderr)`),
  or to `/dev/tty` directly. It must **never** write to stdout.
- On selection, the chosen item's **`name`** (not the title) is written to
  **stdout**, followed by a single `\n`, and nothing else.
- Therefore `RESULT=$(chooz menu.yaml)` must capture only the `name`.

### Exit codes
- `0` — an item was selected/resolved and its name printed.
- `130` — user cancelled the interactive menu (Esc, Ctrl-C, or `q`). Nothing on stdout.
- `1` — runtime error (file missing, invalid YAML, validation failure, name not
  found, no-TTY with no selection). Message on stderr.
- `2` — usage error (bad flags / missing required arg). Cobra default.

---

## 6. Theming & colors

- **Default palette must be legible on both dark and light terminals.** Use
  Lipgloss `lipgloss.AdaptiveColor{Light: ..., Dark: ...}` for every styled
  element, and rely on Lipgloss/termenv background detection. When detection
  fails (dumb term, piped), fall back to a conservative palette that reads on
  both backgrounds (avoid pure white/pure black; prefer mid-tone accents).
- Honor `NO_COLOR` and `--no-color` → render with no ANSI styling at all.
- Honor `CLICOLOR_FORCE` → keep colors even when not a TTY.

### Overridable via the `theme:` YAML block and env vars
Implement this minimal, well-named set. Values accept either an ANSI-256 index
(`"212"`) or a hex string (`"#aabbcc"`):

| Element             | YAML key      | Env var                 |
|---------------------|---------------|-------------------------|
| Cursor / selected   | `cursor`      | `CHOOZ_THEME_CURSOR`      |
| Selected item text  | `selected`    | `CHOOZ_THEME_SELECTED`    |
| Normal item text    | `item`        | `CHOOZ_THEME_ITEM`        |
| Menu header (title) | `header`      | `CHOOZ_THEME_HEADER`      |
| Description text    | `description` | `CHOOZ_THEME_DESCRIPTION` |
| Border / separator  | `border`      | `CHOOZ_THEME_BORDER`      |
| Help / key hints    | `help`        | `CHOOZ_THEME_HELP`        |

Precedence: env var > YAML `theme:` > adaptive default. Unknown theme keys are
ignored (same forward-compat rule as §3.1).

---

## 7. TUI behavior

- **Header:** if top-level `title`/`description` are set, render them at the top.
- **List:** one row per item showing the display label (`title` else `name`),
  with a cursor/highlight on the active row. `--height` caps visible rows;
  scroll within the list when there are more items.
- **Preview pane:** shows the highlighted item's `description`, wrapped and
  multi-line, in a `bubbles/viewport`. Layout:
  - terminal width ≥ ~80 cols → preview **to the right** of the list;
  - narrower → preview **below** the list.
  Long descriptions scroll inside the viewport.
- **Default selection:** `--default <name>` starts with that row highlighted;
  otherwise the first item.
- **Keys:** `↑`/`k` and `↓`/`j` move; `Enter` selects; `Esc`/`Ctrl-C`/`q`
  cancel (exit 130); optional `g`/`G` jump to top/bottom; `PgUp`/`PgDn` scroll
  the preview. Show a one-line help/key hint at the bottom.
- Keep the Bubble Tea `Model` thin. Put all selection-resolution logic (resolve
  by name, default fallback, not-found) in a separate `selector` package so it
  is unit-testable **without** a TTY.

---

## 8. Suggested repo layout

```
chooz/
  go.mod
  main.go                 # thin: cobra wiring + dispatch
  internal/
    config/               # YAML load + validation (+ unknown-key tolerance)
      config.go
      config_test.go
    selector/             # name resolution, default fallback, list (no TTY)
      selector.go
      selector_test.go
    theme/                # adaptive defaults + YAML/env overrides
      theme.go
      theme_test.go
    tui/                  # Bubble Tea model, view, preview layout
      model.go
  testdata/
    example.yaml
    minimal.yaml          # items with only `name`
    invalid_name.yaml
    duplicate_name.yaml
  Makefile
  .goreleaser.yaml        # optional, cross-compiled static releases
  README.md
```

---

## 9. Build & distribution

- Static binary:
  ```bash
  CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(git describe --tags --always)" -o chooz .
  ```
- Cross-compile via `GOOS`/`GOARCH` (linux/amd64, linux/arm64, darwin/*,
  windows/amd64). Optional GoReleaser config to produce these.
- **Airgapped builds:** provide `go mod vendor` and ensure
  `go build -mod=vendor` works offline. Document this in the README. (All deps
  are pure Go, so no system libraries are required.)
- Provide a `Makefile` with at least: `build`, `static`, `test`, `lint`,
  `vendor`, `clean`.

---

## 10. Testing

- **config:** valid parse; missing/empty `items`; `name` regex pass/fail;
  duplicate `name`; unknown keys are ignored (not an error); multi-line
  `description` preserved.
- **selector:** resolve by positional name (hit/miss); `--default` fallback
  when non-TTY; `--list` output order; not-found errors return correct exit.
- **theme:** precedence env > YAML > default; hex and ANSI-256 parsing;
  `NO_COLOR`/`--no-color` produce unstyled output.
- **tui (optional):** drive the Bubble Tea `Model.Update` with synthetic key
  messages (and/or `charmbracelet/x/exp/teatest`) to assert navigation and that
  `Enter` yields the highlighted item's `name`. Keep this thin — the real logic
  lives in `selector`.
- Use the files under `testdata/`.

---

## 11. Coding conventions

- `gofmt`/`goimports` clean; pass `golangci-lint` with a sane default config.
- Idiomatic, modern Go for the `go.mod` version (no obsolete patterns).
- Wrap errors with context (`fmt.Errorf("...: %w", err)`); print user-facing
  errors to **stderr**; map them to the exit codes in §5.
- No global mutable state; pass config/theme explicitly.
- `main.go` stays thin; behavior lives in `internal/*`.

---

## 12. Out of scope (v1) — but keep the door open

- **Fuzzy filtering** (typing to filter, like `gum filter`/`fzf`). Do not
  implement now, but structure the `tui.Model` so a filter input can be added
  without a rewrite.
- Additional per-item fields beyond `name`/`title`/`description` (e.g. `value`,
  `cmd`, `tags`, `disabled`). The unknown-key tolerance (§3.1) already makes
  adding these non-breaking.
- Multi-select.

---

## 13. Acceptance criteria (definition of done)

1. `chooz testdata/example.yaml` shows the menu; arrow/`j`/`k` navigate; the
   preview pane shows the highlighted item's multi-line description; `Enter`
   prints that item's `name` to **stdout**.
2. `X=$(chooz testdata/example.yaml)` captures **only** the name (TUI went to
   stderr/tty).
3. `chooz testdata/example.yaml prod` prints `prod` and exits 0 with no TUI;
   `chooz testdata/example.yaml nope` exits 1 with an error on stderr.
4. `chooz testdata/example.yaml -l` prints all names, one per line, in YAML order.
5. `chooz --default prod testdata/example.yaml` pre-highlights `prod`
   interactively; piped (no TTY) it prints `prod` and exits 0.
6. Piped with no name and no `--default` → exit 1 with a clear stderr message.
7. `Esc`/`Ctrl-C`/`q` → exit 130, nothing on stdout.
8. Invalid YAML, bad `name` (non-alphanumeric), or duplicate `name` → exit 1
   with a clear message.
9. Adding an unknown key to a YAML item does **not** break parsing.
10. `CGO_ENABLED=0 go build` produces a working static binary; colors are
    legible on both dark and light terminals; `NO_COLOR` disables styling.
