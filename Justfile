import? 'Justfile.local'

# Build and install diffstory to specified path
install path:
    nix develop -c go build -o {{path}} ./cmd/diffstory/
