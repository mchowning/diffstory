# diffstory

A terminal UI viewer for code reviews. Receives structured review data via HTTP or MCP (Model Context Protocol) and displays it in an interactive TUI with syntax highlighting.

Designed for use with Claude Code - when Claude reviews your code, the results appear in a dedicated viewer with syntax-highlighted diffs, organized by topic.

## Installation

```bash
# Build from source
go build -o diffstory ./cmd/diffstory/

# Or with nix
nix develop -c go build -o diffstory ./cmd/diffstory/
```

## Using with Claude Code

### Setup

1. Build and install diffstory somewhere in your PATH:
   ```bash
   go build -o ~/bin/diffstory ./cmd/diffstory/
   ```

2. Register the MCP server with Claude Code:
   ```bash
   claude mcp add --transport stdio diffstory --scope user -- ~/bin/diffstory mcp
   ```

3. Restart Claude Code to pick up the new MCP server.

### Workflow

**Terminal 1** - Run the diffstory viewer in your project directory:
```bash
cd /path/to/your/project
diffstory
```

The viewer starts in "waiting" mode, ready to display reviews.

**Terminal 2** - Work with Claude Code as usual:
```bash
cd /path/to/your/project
claude
```

When you want Claude to review code, ask it directly:

> "Review the changes I made to the authentication module"

> "Use diffstory to review this PR"

> "Submit a code review of the recent commits"

Claude will use the `submit_review` tool to send a structured review to diffstory. The review appears instantly in the viewer with:

- Sections organized by topic/concern
- Narrative explanations for each section
- Syntax-highlighted diffs showing the relevant code
- Importance levels (high/medium/low)

### What Claude Sends

When Claude submits a review, it provides:

- **Title**: Summary of what's being reviewed
- **Sections**: Grouped by topic (e.g., "Error Handling", "Performance", "Security")
- **Narrative**: Claude's explanation of each concern or suggestion
- **Hunks**: The actual code diffs with file paths and line numbers

The viewer displays this in a split-pane interface - sections on the left, details on the right.

## Usage

### TUI Viewer (default)

Run `diffstory` in any directory to start the viewer. It watches for reviews submitted to that directory.

```bash
cd /path/to/project
diffstory
```

**Keybindings:**

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate sections up/down |
| `J` / `K` | Scroll diff content |
| `h` / `l` | Cycle panel focus |
| `0` / `1` / `2` | Focus Diff/Section/Files panel |
| `f` | Cycle importance filter |
| `t` | Cycle test filter |
| `G` | Generate review (LLM) |
| `?` / `Esc` | Toggle/close help |
| `q` / `Ctrl+C` | Quit |

**Filtering:**

The TUI supports two filter dimensions that work together:

- **Importance filter** (`f`): Cycles through Low (all) → Medium → High only
- **Test filter** (`t`): Cycles through All → Excluding Tests → Only Tests

Filters combine - a hunk must pass both filters to be displayed. For example, with importance "High only" and test filter "Excluding Tests", only high-importance production code hunks are shown.

The filter indicator at the bottom shows current state: `Diff filter: High only | Excluding tests`

### HTTP Server

Start an HTTP server to receive reviews:

```bash
diffstory server              # Default port 8765
diffstory server -port 9000   # Custom port
diffstory server -v           # Verbose logging
```

Submit reviews via POST:

```bash
curl -X POST http://localhost:8765/review \
  -H "Content-Type: application/json" \
  -d '{
    "workingDirectory": "/path/to/project",
    "title": "Code Review",
    "sections": [
      {
        "id": "section-1",
        "narrative": "Improved error handling",
        "importance": "high",
        "hunks": [
          {
            "file": "main.go",
            "startLine": 42,
            "diff": "@@ -42,3 +42,5 @@\n func main() {\n+    if err != nil {\n+        return err\n+    }\n }"
          }
        ]
      }
    ]
  }'
```

### MCP Server

Run as an MCP server (used by Claude Code):

```bash
diffstory mcp      # Runs on stdio
diffstory mcp -v   # Verbose logging to stderr
```

See [Using with Claude Code](#using-with-claude-code) for setup instructions.

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

## How It Works

1. **Storage**: Reviews are stored in `~/.diffstory/reviews/` as JSON files, hashed by working directory
2. **File Watching**: The TUI watches for file changes and updates automatically
3. **Syntax Highlighting**: Diffs are displayed with syntax-aware colorization
4. **Multi-source**: Both HTTP and MCP interfaces write to the same storage, so the TUI shows reviews from either source

## Development

```bash
# Run tests
nix develop -c go test ./...

# Build
nix develop -c go build ./cmd/diffstory/

# Run the viewer
nix develop -c go run ./cmd/diffstory/

# Run the HTTP server
nix develop -c go run ./cmd/diffstory/ server -v

# Run the MCP server
nix develop -c go run ./cmd/diffstory/ mcp -v
```

## Architecture

```
cmd/diffstory/
  main.go      # CLI entry point
  server.go    # HTTP server runner
  mcp.go       # MCP server runner

internal/
  model/       # Review data structures
  storage/     # File-based persistence
  review/      # Shared business logic (validation, normalization)
  server/      # HTTP server implementation
  mcpserver/   # MCP server implementation
  watcher/     # File system watcher
  tui/         # Terminal UI (Bubble Tea)
  highlight/   # Syntax highlighting
```

## License

MIT
