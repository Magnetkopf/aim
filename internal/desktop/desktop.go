package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Magnetkopf/aim/internal/paths"
)

// UpdateDesktopEntry updates the system .desktop file for the given app by rewriting
func UpdateDesktopEntry(appName string) error {
	appDir := paths.AppDir(appName)
	currentSymlink := paths.CurrentSymlink(appName)

	target, err := os.Readlink(currentSymlink)
	if err != nil {
		return fmt.Errorf("failed to read current symlink: %w", err)
	}
	hash := filepath.Base(target)

	versionDir := filepath.Join(appDir, hash)
	squashfsDir := filepath.Join(versionDir, "squashfs-root")

	entries, err := os.ReadDir(squashfsDir)
	if err != nil {
		return fmt.Errorf("failed to read squashfs-root: %w", err)
	}

	var desktopFile string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".desktop") {
			desktopFile = filepath.Join(squashfsDir, entry.Name())
			break
		}
	}
	if desktopFile == "" {
		return fmt.Errorf("no .desktop file found")
	}

	desktopBytes, err := os.ReadFile(desktopFile)
	if err != nil {
		return fmt.Errorf("failed to read desktop file: %w", err)
	}

	appRunPath := filepath.Join(currentSymlink, "squashfs-root", "AppRun")
	iconPath := ""
	lines := splitLines(string(desktopBytes))
	for _, line := range lines {
		if strings.HasPrefix(line, "Icon=") {
			iconPath = filepath.Join(currentSymlink, "squashfs-root", filepath.Base(strings.TrimPrefix(line, "Icon=")))
			break
		}
	}

	var updatedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Exec=") {
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) > 1 {
				updatedLines = append(updatedLines, fmt.Sprintf("Exec=\"%s\" %s", appRunPath, parts[1]))
			} else {
				updatedLines = append(updatedLines, fmt.Sprintf("Exec=\"%s\"", appRunPath))
			}
		} else if strings.HasPrefix(trimmed, "Icon=") {
			updatedLines = append(updatedLines, fmt.Sprintf("Icon=%s", iconPath))
		} else if strings.HasPrefix(trimmed, "TryExec=") {
			updatedLines = append(updatedLines, fmt.Sprintf("TryExec=\"%s\"", appRunPath))
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	updatedDesktop := strings.Join(updatedLines, "\n")

	targetDesktopName := fmt.Sprintf("aim-%s.desktop", strings.ReplaceAll(appName, " ", ""))
	targetDesktopPath := filepath.Join(paths.ApplicationsDir(), targetDesktopName)

	if err := os.WriteFile(targetDesktopPath, []byte(updatedDesktop), 0644); err != nil {
		return fmt.Errorf("failed to write system desktop file: %w", err)
	}

	// Update desktop DB
	exec.Command("update-desktop-database", paths.ApplicationsDir()).Run()

	return nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
