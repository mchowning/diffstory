# diffguide

A terminal UI viewer for code reviews. Receives structured review data via HTTP or MCP (Model Context Protocol) and displays it in an interactive TUI with syntax highlighting.

Designed for use with Claude Code - when Claude reviews your code, the results appear in a dedicated viewer with syntax-highlighted diffs, organized by topic.

## Installation

```bash
# Build from source
go build -o diffguide ./cmd/diffguide/

# Or with nix
nix develop -c go build -o diffguide ./cmd/diffguide/
```

## Using with Claude Code

### Setup

1. Build and install diffguide somewhere in your PATH:
   ```bash
   go build -o ~/bin/diffguide ./cmd/diffguide/
   ```

2. Register the MCP server with Claude Code:
   ```bash
   claude mcp add --transport stdio diffguide --scope user -- ~/bin/diffguide mcp
   ```

3. Restart Claude Code to pick up the new MCP server.

### Workflow

**Terminal 1** - Run the diffguide viewer in your project directory:
```bash
cd /path/to/your/project
diffguide
```

The viewer starts in "waiting" mode, ready to display reviews.

**Terminal 2** - Work with Claude Code as usual:
```bash
cd /path/to/your/project
claude
```

When you want Claude to review code, ask it directly:

> "Review the changes I made to the authentication module"

> "Use diffguide to review this PR"

> "Submit a code review of the recent commits"

Claude will use the `submit_review` tool to send a structured review to diffguide. The review appears instantly in the viewer with:

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

Run `diffguide` in any directory to start the viewer. It watches for reviews submitted to that directory.

```bash
cd /path/to/project
diffguide
```

**Keybindings:**

| Key | Action |
|-----|--------|
| `j` / `↓` | Next section |
| `k` / `↑` | Previous section |
| `J` | Scroll content down |
| `K` | Scroll content up |
| `?` | Toggle help |
| `q` / `Ctrl+C` | Quit |

### HTTP Server

Start an HTTP server to receive reviews:

```bash
diffguide server              # Default port 8765
diffguide server -port 9000   # Custom port
diffguide server -v           # Verbose logging
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
diffguide mcp      # Runs on stdio
diffguide mcp -v   # Verbose logging to stderr
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

1. **Storage**: Reviews are stored in `~/.diffguide/reviews/` as JSON files, hashed by working directory
2. **File Watching**: The TUI watches for file changes and updates automatically
3. **Syntax Highlighting**: Diffs are displayed with syntax-aware colorization
4. **Multi-source**: Both HTTP and MCP interfaces write to the same storage, so the TUI shows reviews from either source

## Development

```bash
# Run tests
nix develop -c go test ./...

# Build
nix develop -c go build ./cmd/diffguide/

# Run the viewer
nix develop -c go run ./cmd/diffguide/

# Run the HTTP server
nix develop -c go run ./cmd/diffguide/ server -v

# Run the MCP server
nix develop -c go run ./cmd/diffguide/ mcp -v
```

## Architecture

```
cmd/diffguide/
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
