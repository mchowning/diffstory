# Development Environment

This project uses Nix flakes for reproducible development environments.

## Running Commands

Use `nix develop -c` to run commands inside the flake environment:

```bash
nix develop -c go build ./...
nix develop -c go test ./...
nix develop -c go mod tidy
```

## Reference Implementation

Use [lazygit](https://github.com/jesseduffield/lazygit) as inspiration for UI patterns and codebase structure.

## Keeping README Up-to-Date

When making changes, check if the README needs updating:

- **Keybindings**: Update the keybindings table when adding/changing/removing keyboard shortcuts in `internal/tui/update.go`
- **Architecture**: Update the architecture section when adding new packages to `internal/`
- **CLI flags**: Update usage examples when changing command-line arguments in `cmd/diffstory/`
- **Data format**: Update the review JSON schema when changing `internal/model/`
- **Storage**: Update the "How It Works" section if storage paths or behavior changes

# Additional Information

Read `AGENTS.local.md` immediately for additional instructions that are relevant to all workflows
