# Development Environment

This project uses Nix flakes for reproducible development environments.

## Running Commands

Use `nix develop -c` to run commands inside the flake environment:

```bash
nix develop -c go build ./...
nix develop -c go test ./...
nix develop -c go mod tidy
```

## Reference Implementations

Use [lazygit](https://github.com/jesseduffield/lazygit) as inspiration for UI patterns and codebase structure.

Use [review.fast](https://review.fast/) as inspiration for functionality and feature ideas.

## Keeping README Up-to-Date

**IMPORTANT**: Updating the README is a required part of any change that affects documented behavior. Before completing a task, verify whether the README needs updates.

### When to Update

| If you change... | Update this README section |
|------------------|---------------------------|
| `internal/tui/update.go` keybindings | Keybindings table |
| `internal/model/review.go` structs | Review Data Format (JSON schema + field guidance) |
| `internal/model/review.go` field names | Review Data Format (JSON example + field guidance) |
| New packages in `internal/` | Architecture diagram |
| `cmd/diffstory/` CLI args | Usage examples |
| `internal/storage/` paths | How It Works section |
| `internal/config/` options | Configuration Options table |

### Review Data Format Section

The "Review Data Format" section must match `internal/model/review.go` exactly:

- JSON example should reflect all struct fields and their nesting (Review → Chapter → Section → Hunk)
- Field guidance should explain the purpose of each field
- When adding/removing/renaming fields, update both the JSON example and the field guidance

# Additional Information

Read `AGENTS.local.md` immediately for additional instructions that are relevant to all workflows
