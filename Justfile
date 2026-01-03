# Build and install diffguide to ~/dotfiles/bin/
install:
    go build -o ~/dotfiles/bin/diffstory ./cmd/diffstory/

# Start the MCP server
mcp:
    go run ./cmd/diffstory mcp
