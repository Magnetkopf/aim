package installer

import (
	"fmt"
	"os"

	"github.com/Magnetkopf/aim/internal/desktop"
	"github.com/Magnetkopf/aim/internal/metadata"
	"github.com/Magnetkopf/aim/internal/paths"
)

// ExecuteInstallation finalizing the XDG setup
func ExecuteInstallation(meta *metadata.AppMetadata, action string) error {
	appBaseDir := paths.AppDir(meta.AppName)
	versionDir := paths.VersionDir(meta.AppName, meta.Hash)
	currentSymlink := paths.CurrentSymlink(meta.AppName)

	// Move TmpDir to VersionDir
	if err := os.MkdirAll(appBaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create app base dir: %w", err)
	}

	// Only reinstall can overwrite
	if _, err := os.Stat(versionDir); err == nil {
		if action != "reinstall" {
			return fmt.Errorf("version directory %s already exists (use reinstall to overwrite)", versionDir)
		}
		os.RemoveAll(versionDir)
	}

	if err := os.Rename(meta.TmpDir, versionDir); err != nil {
		return fmt.Errorf("failed to move extracted files to installation directory: %w", err)
	}

	// Update symlink
	os.Remove(currentSymlink)
	if err := os.Symlink(meta.Hash, currentSymlink); err != nil {
		return fmt.Errorf("failed to create current symlink: %w", err)
	}

	// Update versions.json
	if err := metadata.AddVersion(meta.AppName, meta.Hash, meta.Version); err != nil {
		fmt.Printf("Warning: failed to update versions.json: %v\n", err)
	}

	if err := desktop.UpdateDesktopEntry(meta.AppName); err != nil {
		return fmt.Errorf("failed to update desktop entry: %w", err)
	}

	fmt.Printf("Successfully installed system desktop file for %s\n", meta.AppName)

	return nil
}
