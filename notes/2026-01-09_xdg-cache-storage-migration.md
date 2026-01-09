---
date: 2026-01-09 15:17:04 EST
git_commit: 7355b52ab7f23664b905fe257bd15adfe57e113f
branch: main
repository: diffstory
topic: "XDG Cache Storage Migration"
tags: [implementation, storage, xdg]
last_updated: 2026-01-09
---

# XDG Cache Storage Migration

## Summary

Migrated review storage from `~/.diffstory/reviews/` to `~/.cache/diffstory/` with `XDG_CACHE_HOME` environment variable support, aligning with the XDG Base Directory Specification.

## Overview

Reviews in diffstory are regenerable cache data—they can always be recreated by re-running the LLM analysis. The previous storage location (`~/.diffstory/reviews/`) didn't follow XDG conventions, while the config system already used `XDG_CONFIG_HOME`. This change moves reviews to the appropriate XDG cache location, making the storage behavior consistent with the config system and allowing users to clear their cache without affecting configuration.

The implementation checks for `XDG_CACHE_HOME` first (for users who set custom cache directories), falling back to `~/.cache/diffstory/` when unset. Existing reviews in the old location (`~/.diffstory/reviews/`) are not migrated—they can be regenerated on demand.

## Technical Details

### Storage Path Resolution

The `NewStore()` function in `internal/storage/store.go` was updated to follow XDG conventions:

```go
func NewStore() (*Store, error) {
	var baseDir string
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		baseDir = filepath.Join(xdg, "diffstory")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		baseDir = filepath.Join(home, ".cache", "diffstory")
	}
	return NewStoreWithDir(baseDir)
}
```

The change prioritizes `XDG_CACHE_HOME` when set, otherwise defaults to `~/.cache/diffstory/`. This matches the pattern used in `internal/config/config.go` for configuration files.

### Test Coverage

Two tests were added to `internal/storage/store_test.go` to verify the XDG behavior:

1. `TestNewStore_UsesXDGCacheHomeWhenSet` - Sets `XDG_CACHE_HOME` to a temp directory and verifies the store uses it
2. `TestNewStore_FallsBackToDotCacheWhenXDGUnset` - Unsets `XDG_CACHE_HOME` and verifies the store falls back to `~/.cache/diffstory`

Both tests properly save and restore the original environment variable value using `t.Cleanup()`.

### Documentation Updates

- `internal/watcher/watcher.go:27` - Updated comment to reference new storage location
- `README.md:181` - Updated "How It Works" section to document `~/.cache/diffstory/` with `XDG_CACHE_HOME` note

## Git References

**Branch**: `main`

**Commit Range**: Single commit on main

**Commits Documented**:

**7355b52ab7f23664b905fe257bd15adfe57e113f** (2026-01-09)
Move review storage to XDG cache directory

Reviews are regenerable cache data, so move storage from ~/.diffstory/reviews/
to ~/.cache/diffstory/ (with XDG_CACHE_HOME support). This aligns with the
XDG Base Directory Specification, matching the existing XDG_CONFIG_HOME
pattern for config files.

- NewStore() now checks XDG_CACHE_HOME env var with ~/.cache/diffstory fallback
- Add tests for XDG_CACHE_HOME behavior
- Update documentation
