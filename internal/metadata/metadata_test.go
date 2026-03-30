package metadata

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateHash(t *testing.T) {
	content := []byte("fake appimage content for hashing")

	// Create temp file
	tmpFile := filepath.Join(t.TempDir(), "test.appimage")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("could not write temp file: %v", err)
	}

	got, err := GenerateHash(tmpFile)
	if err != nil {
		t.Fatalf("GenerateHash returned error: %v", err)
	}

	want := fmt.Sprintf("%x", sha256.Sum256(content))
	if got != want {
		t.Errorf("GenerateHash() = %q, want %q", got, want)
	}
}

func TestHashExists(t *testing.T) {
	homeDir := t.TempDir()
	// Override os.UserHomeDir behavior by creating path relative to temp home
	appName := "TestApp"
	hash := "abc123"

	// Must construct the same path layout that HashExists uses
	versionDir := filepath.Join(homeDir, ".local", "share", "aim", "apps", appName, hash)

	if HashExists(appName, hash) {
		t.Logf("HashExists returned true unexpectedly (may exist in real home)")
	}

	// Create dir to test positive case by setting HOME env temporarily
	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		t.Fatalf("could not create version dir: %v", err)
	}

	if !HashExists(appName, hash) {
		t.Errorf("HashExists() = false, want true")
	}
}

func TestParseExtractedMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	squashfsDir := filepath.Join(tmpDir, "squashfs-root")
	if err := os.MkdirAll(squashfsDir, 0755); err != nil {
		t.Fatalf("could not create squashfs-root: %v", err)
	}

	desktopContent := `[Desktop Entry]
Name=TestApp
Exec=testapp
Icon=testapp-icon
Type=Application
`
	desktopPath := filepath.Join(squashfsDir, "testapp.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
		t.Fatalf("could not write desktop file: %v", err)
	}

	iconPath := filepath.Join(squashfsDir, "testapp-icon.png")
	if err := os.WriteFile(iconPath, []byte("fake png"), 0644); err != nil {
		t.Fatalf("could not write icon file: %v", err)
	}

	hash := "deadbeef"
	meta, err := parseExtractedMetadata(hash, tmpDir, squashfsDir)
	if err != nil {
		t.Fatalf("parseExtractedMetadata returned error: %v", err)
	}

	if meta.Hash != hash {
		t.Errorf("Hash = %q, want %q", meta.Hash, hash)
	}
	if meta.AppName != "TestApp" {
		t.Errorf("AppName = %q, want %q", meta.AppName, "TestApp")
	}
	if meta.Desktop != desktopPath {
		t.Errorf("Desktop = %q, want %q", meta.Desktop, desktopPath)
	}
	if meta.IconPath != iconPath {
		t.Errorf("IconPath = %q, want %q", meta.IconPath, iconPath)
	}
	if meta.TmpDir != tmpDir {
		t.Errorf("TmpDir = %q, want %q", meta.TmpDir, tmpDir)
	}
}

func TestParseExtractedMetadata_NoDesktopFile(t *testing.T) {
	tmpDir := t.TempDir()
	squashfsDir := filepath.Join(tmpDir, "squashfs-root")
	if err := os.MkdirAll(squashfsDir, 0755); err != nil {
		t.Fatalf("could not create squashfs-root: %v", err)
	}

	_, err := parseExtractedMetadata("hash", tmpDir, squashfsDir)
	if err == nil {
		t.Fatalf("expected error when no .desktop file found, got nil")
	}
}

func TestParseExtractedMetadata_FallbackIcon(t *testing.T) {
	tmpDir := t.TempDir()
	squashfsDir := filepath.Join(tmpDir, "squashfs-root")
	if err := os.MkdirAll(squashfsDir, 0755); err != nil {
		t.Fatalf("could not create squashfs-root: %v", err)
	}

	desktopContent := `[Desktop Entry]
Name=FallbackApp
Exec=app
Icon=nonexistent-icon
`
	desktopPath := filepath.Join(squashfsDir, "fallback.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
		t.Fatalf("could not write desktop file: %v", err)
	}

	fallbackIcon := filepath.Join(squashfsDir, "random-icon.svg")
	if err := os.WriteFile(fallbackIcon, []byte("fake svg"), 0644); err != nil {
		t.Fatalf("could not write fallback icon: %v", err)
	}

	meta, err := parseExtractedMetadata("hash", tmpDir, squashfsDir)
	if err != nil {
		t.Fatalf("parseExtractedMetadata returned error: %v", err)
	}

	if meta.IconPath != fallbackIcon {
		t.Errorf("IconPath = %q, want %q", meta.IconPath, fallbackIcon)
	}
}

func TestParseExtractedMetadata_MultipleDesktopEntries(t *testing.T) {
	tmpDir := t.TempDir()
	squashfsDir := filepath.Join(tmpDir, "squashfs-root")
	if err := os.MkdirAll(squashfsDir, 0755); err != nil {
		t.Fatalf("could not create squashfs-root: %v", err)
	}

	// Desktop file with multiple sections; parser should only read [Desktop Entry]
	desktopContent := `[Desktop Action Open]
Name=Open Extra
Icon=wrong-icon

[Desktop Entry]
Name=RealApp
Exec=realapp
Icon=real-icon
Type=Application

[Another Section]
Name=Ignore Me
Icon=ignore-icon

[Desktop Action idk]
Name=idk
`
	desktopPath := filepath.Join(squashfsDir, "realapp.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
		t.Fatalf("could not write desktop file: %v", err)
	}

	iconPath := filepath.Join(squashfsDir, "real-icon.png")
	if err := os.WriteFile(iconPath, []byte("fake png"), 0644); err != nil {
		t.Fatalf("could not write icon file: %v", err)
	}

	meta, err := parseExtractedMetadata("hash", tmpDir, squashfsDir)
	if err != nil {
		t.Fatalf("parseExtractedMetadata returned error: %v", err)
	}

	if meta.AppName != "RealApp" {
		t.Errorf("AppName = %q, want %q", meta.AppName, "RealApp")
	}
	if meta.IconPath != iconPath {
		t.Errorf("IconPath = %q, want %q", meta.IconPath, iconPath)
	}
}

func TestParseExtractedMetadata_NoName(t *testing.T) {
	tmpDir := t.TempDir()
	squashfsDir := filepath.Join(tmpDir, "squashfs-root")
	if err := os.MkdirAll(squashfsDir, 0755); err != nil {
		t.Fatalf("could not create squashfs-root: %v", err)
	}

	desktopContent := `[Desktop Entry]
Exec=noname
`
	desktopPath := filepath.Join(squashfsDir, "noname.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0644); err != nil {
		t.Fatalf("could not write desktop file: %v", err)
	}

	meta, err := parseExtractedMetadata("hash", tmpDir, squashfsDir)
	if err != nil {
		t.Fatalf("parseExtractedMetadata returned error: %v", err)
	}

	if meta.AppName != "nameless app" {
		t.Errorf("AppName = %q, want %q", meta.AppName, "nameless app")
	}
}