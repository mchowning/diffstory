# Build and install diffguide to ~/dotfiles/bin/
install:
    go build -o ~/dotfiles/bin/diffguide ./cmd/diffguide/

# Start the MCP server
mcp:
    go run ./cmd/diffguide mcp
