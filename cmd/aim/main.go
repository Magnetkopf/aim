package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/Magnetkopf/aim/internal/installer"
	"github.com/Magnetkopf/aim/internal/intercept"
	"github.com/Magnetkopf/aim/internal/metadata"
	"github.com/Magnetkopf/aim/web"
)

func main() {
	var register bool
	flag.BoolVar(&register, "register", false, "Register aim as the default handler for .AppImage files")

	flag.Parse()

	if register {
		if err := intercept.Register(); err != nil {
			fmt.Fprintf(os.Stderr, "Error registering aim: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// For now, in Phase 1, we just verify that we received a file path.
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("aim")
		fmt.Println("Usage: aim [appimage_file]")
		fmt.Println("       aim --register")
		os.Exit(1)
	}

	filePath := args[0]

	// Ensure the file exists
	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not access file '%s': %v\n", filePath, err)
		os.Exit(1)
	}

	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: '%s' is a directory, expected an AppImage file.\n", filePath)
		os.Exit(1)
	}

	fmt.Println("Extracting...")

	meta, err := metadata.Extract(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting metadata: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting Web UI Installer...")

	// Prepare the embedded UI filesystem
	distFS, err := fs.Sub(web.UI, "dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load embedded UI: %v\n", err)
		os.Exit(1)
	}

	// Open the installer UI and wait for user's action
	action, err := installer.RunUI(meta, http.FS(distFS))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running installer interface: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User chose to: %s\n", action)
	if action == "cancel" {
		fmt.Println("Installation cancelled.")
		os.RemoveAll(meta.TmpDir)
		os.Exit(0)
	}

	if action == "install" || action == "reinstall" {
		if action == "reinstall" {
			fmt.Println("Reinstalling existing version...")
		}
		if err := installer.ExecuteInstallation(meta, action); err != nil {
			fmt.Fprintf(os.Stderr, "Installation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✔ Done")
	}
}
