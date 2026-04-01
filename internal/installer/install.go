package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Magnetkopf/aim/internal/metadata"
)

// ExecuteInstallation finalizing the XDG setup
func ExecuteInstallation(meta *metadata.AppMetadata, action string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	appBaseDir := filepath.Join(homeDir, ".local", "share", "aim", "apps", meta.AppName)
	versionDir := filepath.Join(appBaseDir, meta.Hash)
	currentSymlink := filepath.Join(appBaseDir, "current")

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

	// Process Desktop file
	desktopPath := filepath.Join(versionDir, "squashfs-root", filepath.Base(meta.Desktop))
	desktopBytes, err := os.ReadFile(desktopPath)
	if err != nil {
		return fmt.Errorf("failed to read desktop file at %s: %w", desktopPath, err)
	}

	appRunPath := filepath.Join(currentSymlink, "squashfs-root", "AppRun")
	iconPath := filepath.Join(currentSymlink, "squashfs-root", filepath.Base(meta.IconPath))

	// Rewrite Desktop content
	lines := strings.Split(string(desktopBytes), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Exec=") {
			// Replace Exec binary with appRunPath, keep arguments
			// E.g., Exec=app %U -> Exec="/home/.../app" %U
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) > 1 {
				lines[i] = fmt.Sprintf("Exec=\"%s\" %s", appRunPath, parts[1])
			} else {
				lines[i] = fmt.Sprintf("Exec=\"%s\"", appRunPath)
			}
		} else if strings.HasPrefix(trimmed, "Icon=") {
			lines[i] = fmt.Sprintf("Icon=%s", iconPath)
		} else if strings.HasPrefix(trimmed, "TryExec=") {
			lines[i] = fmt.Sprintf("TryExec=\"%s\"", appRunPath)
		}
	}

	updatedDesktop := strings.Join(lines, "\n")

	// Add prefix
	targetDesktopName := fmt.Sprintf("aim-%s.desktop", strings.ReplaceAll(meta.AppName, " ", ""))
	targetDesktopPath := filepath.Join(homeDir, ".local", "share", "applications", targetDesktopName)

	if err := os.WriteFile(targetDesktopPath, []byte(updatedDesktop), 0644); err != nil {
		return fmt.Errorf("failed to write system desktop file: %w", err)
	}

	fmt.Printf("Successfully installed system desktop file at: %s\n", targetDesktopPath)

	// Update desktop DB
	exec.Command("update-desktop-database", filepath.Join(homeDir, ".local", "share", "applications")).Run()

	return nil
}
