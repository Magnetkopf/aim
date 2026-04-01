package metadata

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type AppMetadata struct {
	Hash             string
	AppName          string
	Version          string
	IconPath         string
	Desktop          string // path to the extracted .desktop file
	TmpDir           string // where squashfs-root is currently located
	AlreadyInstalled bool   // true if this exact hash version is already installed
}

// AppVersion represents a single installed version in versions.json
type AppVersion struct {
	Hash        string    `json:"hash"`
	Version     string    `json:"version"`
	InstallTime time.Time `json:"install_time"`
}

// AppVersionsFile represents the structure of versions.json
type AppVersionsFile struct {
	Versions []AppVersion `json:"versions"`
}

// GetVersionsFile returns the path to the versions.json for an app
func GetVersionsFile(appName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "share", "aim", "apps", appName, "versions.json"), nil
}

// AddVersion records a newly installed version in versions.json
func AddVersion(appName, hash, version string) error {
	versionsPath, err := GetVersionsFile(appName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(versionsPath), 0755); err != nil {
		return fmt.Errorf("failed to create versions directory: %w", err)
	}

	var versionsFile AppVersionsFile

	// Read existing file if it exists
	if data, err := os.ReadFile(versionsPath); err == nil {
		if err := json.Unmarshal(data, &versionsFile); err != nil {
			// If we can't parse it, we'll just start fresh rather than failing the install
			fmt.Printf("Warning: Could not parse existing versions.json: %v\n", err)
		}
	}

	// Check if this hash is already in the list
	found := false
	for i, v := range versionsFile.Versions {
		if v.Hash == hash {
			// Update install time for existing
			versionsFile.Versions[i].InstallTime = time.Now()
			found = true
			break
		}
	}

	if !found {
		// Add new version
		versionsFile.Versions = append(versionsFile.Versions, AppVersion{
			Hash:        hash,
			Version:     version,
			InstallTime: time.Now(),
		})
	}

	// Write back
	data, err := json.MarshalIndent(versionsFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode versions.json: %w", err)
	}

	if err := os.WriteFile(versionsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write versions.json: %w", err)
	}

	return nil
}

// HashExists checks if the given hash version is already installed for the app
func HashExists(appName, hash string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	versionDir := filepath.Join(homeDir, ".local", "share", "aim", "apps", appName, hash)
	if _, err := os.Stat(versionDir); err == nil {
		return true
	}
	return false
}

// GenerateHash computes the SHA256 hash of the given AppImage file.
// Used for versioning to determine if a newer/different version is being installed.
func GenerateHash(appImagePath string) (string, error) {
	f, err := os.Open(appImagePath)
	if err != nil {
		return "", fmt.Errorf("could not open AppImage for hashing: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("could not compute hash: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Extract unpacks the AppImage
func Extract(appImagePath string) (*AppMetadata, error) {
	hash, err := GenerateHash(appImagePath)
	if err != nil {
		return nil, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not find user home directory: %w", err)
	}

	// 2. Setup extraction tmp directory
	tmpDir := filepath.Join(homeDir, ".local", "share", "aim", "tmp", hash)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temporary extraction dir: %w", err)
	}

	squashfsDir := filepath.Join(tmpDir, "squashfs-root")

	// 3. Extract the AppImage if it hasn't been extracted yet
	if _, err := os.Stat(squashfsDir); os.IsNotExist(err) {
		// Ensure AppImage is executable
		if err := os.Chmod(appImagePath, 0755); err != nil {
			fmt.Printf("Warning: Could not chmod +x the AppImage: %v\n", err)
		}

		cmd := exec.Command(appImagePath, "--appimage-extract")
		cmd.Dir = tmpDir

		// In some environments, extracting might be verbose, but we just capture error if any
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to extract AppImage (is it executable/valid?): %w\nOutput: %s", err, string(output))
		}
	}

	// 4. Parse the extracted metadata
	return parseExtractedMetadata(hash, tmpDir, squashfsDir)
}

// parseExtractedMetadata looks for .desktop file and the icon
func parseExtractedMetadata(hash, tmpDir, squashfsDir string) (*AppMetadata, error) {
	meta := &AppMetadata{
		Hash:   hash,
		TmpDir: tmpDir,
	}

	// Read content of squashfs-root to find .desktop file
	entries, err := os.ReadDir(squashfsDir)
	if err != nil {
		return nil, fmt.Errorf("could not read extracted directory: %w", err)
	}

	var desktopFile string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".desktop") {
			desktopFile = filepath.Join(squashfsDir, entry.Name())
			break
		}
	}

	if desktopFile == "" {
		return nil, fmt.Errorf("no .desktop file found in extracted AppImage")
	}
	meta.Desktop = desktopFile

	// Parse the .desktop file for Name= and Icon=
	contentBytes, err := os.ReadFile(desktopFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read .desktop file: %w", err)
	}

	lines := strings.Split(string(contentBytes), "\n")
	var iconName string
	var fallbackName string
	inDesktopEntry := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inDesktopEntry = line == "[Desktop Entry]"
			continue
		}
		// Only parse Name and Icon when inside [Desktop Entry] section
		if inDesktopEntry {
			if strings.HasPrefix(line, "X-AppImage-Name=") && meta.AppName == "" {
				meta.AppName = strings.TrimPrefix(line, "X-AppImage-Name=")
			}
			if strings.HasPrefix(line, "X-AppImage-Version=") && meta.Version == "" {
				meta.Version = strings.TrimPrefix(line, "X-AppImage-Version=")
			}
			if strings.HasPrefix(line, "Name=") && fallbackName == "" {
				fallbackName = strings.TrimPrefix(line, "Name=")
			}
			if strings.HasPrefix(line, "Icon=") {
				iconName = strings.TrimPrefix(line, "Icon=")
			}
		}
	}

	if meta.AppName == "" && fallbackName != "" {
		meta.AppName = fallbackName
	}

	if meta.AppName == "" {
		meta.AppName = "nameless app"
	}

	// Check if this exact hash version is already installed
	meta.AlreadyInstalled = HashExists(meta.AppName, hash)

	// Find the icon file
	if iconName != "" {
		pngPath := filepath.Join(squashfsDir, iconName+".png")
		svgPath := filepath.Join(squashfsDir, iconName+".svg")

		if _, err := os.Stat(pngPath); err == nil {
			meta.IconPath = pngPath
		} else if _, err := os.Stat(svgPath); err == nil {
			meta.IconPath = svgPath
		} else {
			// fallback: some silly one just have a .png or .svg without matching the exact name
			for _, entry := range entries {
				if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".png") || strings.HasSuffix(entry.Name(), ".svg")) {
					meta.IconPath = filepath.Join(squashfsDir, entry.Name())
					break
				}
			}
		}
	}

	return meta, nil
}
