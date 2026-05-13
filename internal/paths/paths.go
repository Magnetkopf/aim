package paths

import (
	"os"
	"path/filepath"
)

// IsRoot returns true if the current process is running as root.
func IsRoot() bool {
	return os.Getuid() == 0
}

// BaseDir returns the base directory for aim data.
// User: ~/.local/share/aim
// Root: /usr/share/aim
func BaseDir() string {
	if IsRoot() {
		return "/usr/share/aim"
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "aim")
}

// AppsDir returns the directory where apps are installed.
func AppsDir() string {
	return filepath.Join(BaseDir(), "apps")
}

// AppDir returns the directory for a specific app.
func AppDir(appName string) string {
	return filepath.Join(AppsDir(), appName)
}

// VersionDir returns the directory for a specific app version.
func VersionDir(appName, hash string) string {
	return filepath.Join(AppDir(appName), hash)
}

// CurrentSymlink returns the path to the current symlink for an app.
func CurrentSymlink(appName string) string {
	return filepath.Join(AppDir(appName), "current")
}

// TmpDir returns the temporary extraction directory.
func TmpDir() string {
	return filepath.Join(BaseDir(), "tmp")
}

// AppTmpDir returns the temporary extraction directory for a hash.
func AppTmpDir(hash string) string {
	return filepath.Join(TmpDir(), hash)
}

// VersionsFile returns the path to versions.json for an app.
func VersionsFile(appName string) string {
	return filepath.Join(AppDir(appName), "versions.json")
}

// ApplicationsDir returns the system applications directory.
// User: ~/.local/share/applications
// Root: /usr/share/applications
func ApplicationsDir() string {
	if IsRoot() {
		return "/usr/share/applications"
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "applications")
}

// DesktopFilePath returns the path to the aim desktop file.
func DesktopFilePath() string {
	return filepath.Join(ApplicationsDir(), "aim.desktop")
}

// AppDesktopFilePath returns the path to an app's desktop file.
func AppDesktopFilePath(appName string) string {
	return filepath.Join(ApplicationsDir(), "aim-"+appName+".desktop")
}

// MimeAppsListPath returns the path to mimeapps.list.
func MimeAppsListPath() string {
	if IsRoot() {
		return "/usr/share/applications/mimeapps.list"
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "mimeapps.list")
}
