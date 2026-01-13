# diffstory

A terminal UI viewer for code reviews. Organizes diffs into a narrative story that walks you through not only *what* has changed, but more importantly, *why* code changed.

## Why diffstory?

With AI generating massive amounts of code, it is even more critical to be able to quickly absorb the meaning and intention behind code changes that you have not written yourself. `diffstory` makes it possible to review code more quickly and with better comprehension.

`diffstory` can be used with any coding agent, like Claude Code. The coding agent constructs a story out of the selected code changes and shows the results in a dedicated viewer with syntax-highlighted diffs, organized by topic.

![Navigation Demo](https://github.com/mchowning/diffstory/raw/assets/navigation.gif)

## Quick Start

If you have [Claude Code](https://www.anthropic.com/claude-code) installed:

1. Build and install:
   ```bash
   go build -o ~/bin/diffstory ./cmd/diffstory/
   ```

2. Make some changes to your code, then:
   ```bash
   diffstory
   ```

3. Press `G` to generate a review of your uncommitted changes.

That's it! diffstory will use Claude Code to analyze your diff and present it as a narrative story.

## Installation

### From Source (with Go)

```bash
go build -o diffstory ./cmd/diffstory/
```

### From Source (with Nix)

```bash
nix develop -c go build -o diffstory ./cmd/diffstory/
```

### Build with Version

```bash
go build -ldflags "-X main.Version=1.0.0" -o diffstory ./cmd/diffstory/
```

## Configuration

diffstory looks for a config file at:
1. `$XDG_CONFIG_HOME/diffstory/config.jsonc` (if XDG_CONFIG_HOME is set)
2. `~/.config/diffstory/config.jsonc`

Both `.json` and `.jsonc` extensions are supported. JSONC allows comments.

See `config.example.jsonc` for a documented example.

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `llmCommand` | `string[]` | `["claude", "-p"]` | Command to invoke your LLM. Prompt is appended as final arg. |
| `diffCommand` | `string[]` | `["git", "diff", "HEAD"]` | Command to generate the diff to review. |
| `defaultFilterLevel` | `string` | `"low"` | Initial importance filter: `"low"`, `"medium"`, or `"high"`. |
| `debugLoggingEnabled` | `bool` | `false` | Enable debug logging to `/tmp/diffstory.log`. |

### Using a Different LLM

Configure `llmCommand` to use any CLI tool that accepts a prompt as the final argument:

```jsonc
{
  // Using llm CLI (https://llm.datasette.io/)
  "llmCommand": ["llm", "prompt"]
}
```

## Usage

### Generating Reviews (G keybinding)

Press `G` in the viewer to generate a review of your local changes:

1. **Choose diff source**: Select what to review (uncommitted changes, staged changes, commit range, etc.)
2. **Add context** (optional): Provide guidance for the LLM
3. **Wait for generation**: The LLM analyzes your diff and creates a structured review
4. **Browse the story**: Navigate the review organized by topic

#### Requirements

- An LLM CLI tool that accepts a prompt as the final argument
- By default, diffstory uses Claude Code (`claude -p`)
- Configure a different LLM via `llmCommand` in your config file

#### Example Workflow

```bash
# Make some changes to your code
vim src/auth.go

# Start diffstory
diffstory

# Press G, select "Uncommitted changes", optionally add context
# Wait for the LLM to generate the review
# Navigate the story with j/k/h/l
```

### TUI Viewer

Run `diffstory` in any directory to start the viewer. It watches for reviews submitted to that directory.

```bash
cd /path/to/project
diffstory
```

**Keybindings:**

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate in focused panel (sections, files, or diff) |
| `J` / `K` | Scroll diff content |
| `h` / `l` | Cycle panel focus |
| `0` / `1` / `2` | Focus Diff/Section/Files panel |
| `<` / `>` | Jump to first/last item |
| `,` / `.` | Page up/down |
| `enter` | Select file (when in files panel) |
| `f` | Cycle importance filter |
| `t` | Cycle test filter |
| `G` | Generate review (LLM) |
| `?` / `Esc` | Toggle/close help |
| `q` / `Ctrl+C` | Quit |

**Filtering:**

The TUI supports two filter dimensions that work together:

- **Importance filter** (`f`): Cycles through Low (all) -> Medium -> High only
- **Test filter** (`t`): Cycles through All -> Excluding Tests -> Only Tests

Filters combine - a hunk must pass both filters to be displayed. For example, with importance "High only" and test filter "Excluding Tests", only high-importance production code hunks are shown.

The filter indicator at the bottom shows current state: `Diff filter: High only | Excluding tests`

### Lazygit Integration

I primarily use [lazygit](https://github.com/jesseduffield/lazygit) for viewing diffs day-to-day. When I'm having trouble wrapping my head around a complex set of changes, I trigger diffstory from within lazygit to get the AI-powered narrative breakdown.

My lazygit config includes a custom command to launch diffstory:

```yaml
customCommands:
  - key: "<c-d>"
    command: "diffstory"
    context: "global"
    description: "Open diffstory"
    output: terminal
```

This binds `Ctrl+D` to open diffstory from any lazygit panel. You might find a similar setup works well for you.

## Review Data Format

Reviews are JSON objects with this structure:

```json
{
  "workingDirectory": "/absolute/path/to/project",
  "title": "Review Title",
  "sections": [
    {
      "id": "unique-section-id",
      "narrative": "Explanation of changes in this section",
      "importance": "high|medium|low",
      "hunks": [
        {
          "file": "relative/path/to/file.go",
          "startLine": 10,
          "diff": "@@ -10,3 +10,5 @@\n context\n+added line\n-removed line"
        }
      ]
    }
  ]
}
```

**Field guidance:**

- **narrative**: Should be understandable on its own and connect smoothly to adjacent sections, building a coherent narrative arc
- **diff**: Complete unified diff content - include all lines, do not truncate or summarize
- **importance**: `high` (critical changes), `medium` (significant changes), or `low` (minor changes)
- **isTest** (optional): `true` for test code changes, `false` for production code

## How It Works

1. **Storage**: Reviews are stored in `~/.cache/diffstory/` (or `XDG_CACHE_HOME/diffstory/`) as JSON files, hashed by working directory
2. **File Watching**: The TUI watches for file changes and updates automatically
3. **Syntax Highlighting**: Diffs are displayed with syntax-aware colorization

## Development

```bash
# Run tests
nix develop -c go test ./...

# Build
nix develop -c go build ./cmd/diffstory/

# Run the viewer
nix develop -c go run ./cmd/diffstory/
```

## Architecture

```
cmd/diffstory/
  main.go      # CLI entry point
  version.go   # Version info (set via ldflags)

internal/
  config/      # Configuration loading
  diff/        # Diff parsing utilities
  highlight/   # Syntax highlighting
  logging/     # Debug logging
  model/       # Review data structures
  review/      # Shared business logic (validation, normalization)
  storage/     # File-based persistence
  tui/         # Terminal UI (Bubble Tea)
  watcher/     # File system watcher
```

## Inspirations & Alternatives

**[lazygit](https://github.com/jesseduffield/lazygit)** - The primary inspiration for diffstory's UI patterns, split-pane interface, and keybindings.

**[review.fast](https://review.fast/)** - If you want AI-powered code reviews directly on GitHub PRs rather than reviewing local code in the terminal, review.fast may be a better fit.

## License

MIT
