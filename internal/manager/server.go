package manager

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Magnetkopf/aim/internal/metadata"
)

// AppInfo represents installed app information
type AppInfo struct {
	Name         string `json:"name"`
	IconPath     string `json:"iconPath,omitempty"`
	CurrentHash  string `json:"currentHash,omitempty"`
	VersionCount int    `json:"versionCount"`
}

// VersionInfo represents a single version entry
type VersionInfo struct {
	Hash        string `json:"hash"`
	Version     string `json:"version"`
	InstallTime string `json:"install_time"`
}

// AppDetail represents detailed app information
type AppDetail struct {
	Name         string        `json:"name"`
	IconPath     string        `json:"iconPath,omitempty"`
	CurrentHash  string        `json:"currentHash,omitempty"`
	Versions     []VersionInfo `json:"versions"`
}

// RunManager starts the manager web UI server
func RunManager(staticFS http.FileSystem) error {
	// Create a listener on a random ephemeral port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("could not start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	mux := http.NewServeMux()

	// API: List all installed apps
	mux.HandleFunc("/api/app", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		apps, err := getInstalledApps()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apps)
	})

	// API: Get app details
	mux.HandleFunc("/api/app/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract app name from path: /api/app/{name}
		appName := r.URL.Path[len("/api/app/"):]
		if appName == "" {
			http.Error(w, "App name required", http.StatusBadRequest)
			return
		}

		detail, err := getAppDetail(appName)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "App not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(detail)
	})

	// API: Get app icon
	mux.HandleFunc("/api/icon/", func(w http.ResponseWriter, r *http.Request) {
		appName := r.URL.Path[len("/api/icon/"):]
		if appName == "" {
			http.NotFound(w, r)
			return
		}

		iconPath, err := getAppIconPath(appName)
		if err != nil || iconPath == "" {
			http.NotFound(w, r)
			return
		}

		ext := filepath.Ext(iconPath)
		if ext == ".svg" {
			w.Header().Set("Content-Type", "image/svg+xml")
		} else {
			w.Header().Set("Content-Type", "image/png")
		}
		http.ServeFile(w, r, iconPath)
	})

	// Static files
	if staticFS != nil {
		fsHandler := http.FileServer(staticFS)
		mux.Handle("/", func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				fsHandler.ServeHTTP(w, r)
			})
		}())
	}

	server := &http.Server{Handler: mux}

	// Run the server in background
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Manager server error: %v\n", err)
		}
	}()

	fmt.Printf("App Manager running at %s\n", url)
	openBrowser(url)

	// Wait forever (manager runs until user kills it)
	select {}
}

func getInstalledApps() ([]AppInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	appsDir := filepath.Join(homeDir, ".local", "share", "aim", "apps")

	// Check if directory exists
	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		return []AppInfo{}, nil
	}

	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, err
	}

	var apps []AppInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		appName := entry.Name()
		appDir := filepath.Join(appsDir, appName)

		// Get current symlink target
		currentLink := filepath.Join(appDir, "current")
		currentHash := ""
		if target, err := os.Readlink(currentLink); err == nil {
			currentHash = filepath.Base(target)
		}

		// Get icon path
		iconPath := ""
		if currentHash != "" {
			iconPath, _ = findIconInDir(filepath.Join(appDir, currentHash, "squashfs-root"))
		}

		// Count versions
		versionsFile := filepath.Join(appDir, "versions.json")
		versionCount := 0
		if data, err := os.ReadFile(versionsFile); err == nil {
			var vf metadata.AppVersionsFile
			if err := json.Unmarshal(data, &vf); err == nil {
				versionCount = len(vf.Versions)
			}
		}

		apps = append(apps, AppInfo{
			Name:         appName,
			IconPath:     iconPath,
			CurrentHash:  currentHash,
			VersionCount: versionCount,
		})
	}

	return apps, nil
}

func getAppDetail(appName string) (*AppDetail, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	appDir := filepath.Join(homeDir, ".local", "share", "aim", "apps", appName)

	// Check if app exists
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return nil, err
	}

	// Get current symlink target
	currentLink := filepath.Join(appDir, "current")
	currentHash := ""
	if target, err := os.Readlink(currentLink); err == nil {
		currentHash = filepath.Base(target)
	}

	// Get icon path
	iconPath := ""
	if currentHash != "" {
		iconPath, _ = findIconInDir(filepath.Join(appDir, currentHash, "squashfs-root"))
	}

	// Read versions
	versionsFile := filepath.Join(appDir, "versions.json")
	var versions []VersionInfo
	if data, err := os.ReadFile(versionsFile); err == nil {
		var vf metadata.AppVersionsFile
		if err := json.Unmarshal(data, &vf); err == nil {
			for _, v := range vf.Versions {
				versions = append(versions, VersionInfo{
					Hash:        v.Hash,
					Version:     v.Version,
					InstallTime: v.InstallTime.Format("2006-01-02 15:04"),
				})
			}
		}
	}

	return &AppDetail{
		Name:        appName,
		IconPath:    iconPath,
		CurrentHash: currentHash,
		Versions:    versions,
	}, nil
}

func getAppIconPath(appName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(homeDir, ".local", "share", "aim", "apps", appName)
	currentLink := filepath.Join(appDir, "current")

	target, err := os.Readlink(currentLink)
	if err != nil {
		return "", err
	}

	squashfsDir := filepath.Join(appDir, target, "squashfs-root")
	return findIconInDir(squashfsDir)
}

func findIconInDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// First look for .desktop file to get icon name
	var desktopFile string
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 8 && entry.Name()[len(entry.Name())-8:] == ".desktop" {
			desktopFile = filepath.Join(dir, entry.Name())
			break
		}
	}

	var iconName string
	if desktopFile != "" {
		if content, err := os.ReadFile(desktopFile); err == nil {
			lines := string(content)
			for _, line := range splitLines(lines) {
				if len(line) > 5 && line[:5] == "Icon=" {
					iconName = line[5:]
					break
				}
			}
		}
	}

	// Look for icon file
	if iconName != "" {
		pngPath := filepath.Join(dir, iconName+".png")
		svgPath := filepath.Join(dir, iconName+".svg")
		if _, err := os.Stat(pngPath); err == nil {
			return pngPath, nil
		}
		if _, err := os.Stat(svgPath); err == nil {
			return svgPath, nil
		}
	}

	// Fallback: find any png/svg
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > 4 && name[len(name)-4:] == ".png" {
			return filepath.Join(dir, name), nil
		}
		if len(name) > 4 && name[len(name)-4:] == ".svg" {
			return filepath.Join(dir, name), nil
		}
	}

	return "", fmt.Errorf("no icon found")
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

func openBrowser(url string) {
	if err := exec.Command("xdg-open", url).Start(); err != nil {
		fmt.Printf("Could not open browser automatically: %v\nPlease open %s in your browser manually.\n", err, url)
	}
}
