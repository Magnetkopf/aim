# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

The Vue frontend must be built before the Go binary because the dist is embedded via `//go:embed`:

```bash
# Build frontend
cd web && pnpm install && pnpm run build && cd ..

# Build Go binary
go build -o aim ./cmd/aim
```

## Architecture

aim is an AppImage manager that provides an Android APK-style installation experience. It registers as a MIME handler for `.AppImage` files.

### Installation Flow

1. **Intercept** (`internal/intercept/`) - One-time `--register` flag creates `aim.desktop` and registers via `xdg-mime` as the default handler for `application/vnd.appimage`

2. **Metadata Extraction** (`internal/metadata/`) - When an AppImage is opened:
   - Computes SHA256 hash for versioning
   - Runs `--appimage-extract` to unpack to temp dir
   - Parses `.desktop` file for app name and icon
   - Checks if hash already exists (`AlreadyInstalled` field)

3. **Web UI** (`internal/installer/server.go`) - Spawns ephemeral HTTP server on random port, serves Vue SPA with metadata API endpoints, waits for user action

4. **Installation** (`internal/installer/install.go`) - Moves extracted files to `~/.local/share/aim/apps/<AppName>/<hash>/`, creates `current` symlink, rewrites `.desktop` with correct `Exec` and `Icon` paths

### Key Paths

- Apps installed to: `~/.local/share/aim/apps/<AppName>/<hash>/squashfs-root/`
- Current version symlink: `~/.local/share/aim/apps/<AppName>/current`
- System desktop entry: `~/.local/share/applications/aim-<AppName>.desktop`

### Embedding

The Vue build output (`web/dist/`) is embedded into the Go binary via `web/embed.go` using `//go:embed all:dist`. The frontend must be rebuilt before any Go build to include UI changes.

### Frontend Stack

Vue 3 + TypeScript + TailwindCSS 4 + shadcn-vue. Single `App.vue` component handles all UI states.
