import? 'Justfile.local'

# Build and install diffstory to specified path
install path:
    nix develop -c go build -o {{path}} ./cmd/diffstory/

# Start the MCP server
mcp:
    nix develop -c go run ./cmd/diffstory mcp
