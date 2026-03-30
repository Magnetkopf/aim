package intercept

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Register configures system to open .AppImage files with aim
func Register() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	appDir := filepath.Join(homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create applications directory: %w", err)
	}

	desktopFilePath := filepath.Join(appDir, "aim.desktop")

	desktopFileContent := fmt.Sprintf(`[Desktop Entry]
Name=aim
Comment=Love ur AppImage
Exec=%s %%f
Icon=system-software-install
Type=Application
Categories=System;Utility;Core;
MimeType=application/vnd.appimage;
NoDisplay=true
`, execPath)

	if err := os.WriteFile(desktopFilePath, []byte(desktopFileContent), 0644); err != nil {
		return fmt.Errorf("failed to write aim.desktop: %w", err)
	}
	fmt.Printf("Created desktop entry at: %s\n", desktopFilePath)

	fmt.Println("Updating desktop database...")
	updateCmd := exec.Command("update-desktop-database", appDir)
	if err := updateCmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to run update-desktop-database: %v\n", err)
	}

	fmt.Println("Setting aim as default handler for AppImage...")
	xdgCmd := exec.Command("xdg-mime", "default", "aim.desktop", "application/vnd.appimage")
	if err := xdgCmd.Run(); err != nil {
		return fmt.Errorf("failed to run xdg-mime: %w", err)
	}

	fmt.Println("Done. aim is now the default handler for AppImage.")
	return nil
}
