# Development Environment

This project uses Nix flakes for reproducible development environments.

## Running Commands

Use `nix develop -c` to run commands inside the flake environment:

```bash
nix develop -c go build ./...
nix develop -c go test ./...
nix develop -c go mod tidy
```

## Installing

Build and install to `~/dotfiles/bin/`:

```bash
nix develop -c go build -o ~/dotfiles/bin/diffstory ./cmd/diffstory/
```

## Reference Implementation

Use [lazygit](https://github.com/jesseduffield/lazygit) as inspiration for UI patterns and codebase structure. The lazygit source code is available at:

- Local: `~/code/lazygit`
- Online: https://github.com/jesseduffield/lazygit
