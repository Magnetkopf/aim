package intercept

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Unregister removes the system configuration for aim
func Unregister() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	appDir := filepath.Join(homeDir, ".local", "share", "applications")
	desktopFilePath := filepath.Join(appDir, "aim.desktop")

	fmt.Println("Removing aim as default handler for AppImage...")
	mimeAppsPath := filepath.Join(homeDir, ".config", "mimeapps.list")
	if err := removeFromMimeAppsList(mimeAppsPath, "application/vnd.appimage", "aim.desktop"); err != nil {
		fmt.Printf("Warning: Failed to update mimeapps.list: %v\n", err)
	}

	if _, err := os.Stat(desktopFilePath); err == nil {
		if err := os.Remove(desktopFilePath); err != nil {
			return fmt.Errorf("failed to remove aim.desktop: %w", err)
		}
		fmt.Printf("Removed desktop entry at: %s\n", desktopFilePath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check aim.desktop: %w", err)
	}

	fmt.Println("Updating desktop database...")
	updateCmd := exec.Command("update-desktop-database", appDir)
	if err := updateCmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to run update-desktop-database: %v\n", err)
	}

	fmt.Println("Done. aim is no longer registered as a handler.")
	return nil
}

func removeFromMimeAppsList(mimeAppsPath, mimeType, desktopFile string) error {
	content, err := os.ReadFile(mimeAppsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var result []string
	inSection := false
	skipLine := false

	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inSection = strings.HasPrefix(line, "[Default Applications") ||
				strings.HasPrefix(line, "[Default Actions")
		}

		if inSection && strings.Contains(line, mimeType) && strings.Contains(line, desktopFile) {
			skipLine = true
		} else if inSection && skipLine && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			skipLine = false
		}

		if !skipLine || !inSection {
			result = append(result, line)
		}
		skipLine = false
	}

	return os.WriteFile(mimeAppsPath, []byte(strings.Join(result, "\n")), 0644)
}
