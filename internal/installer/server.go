package installer

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/Magnetkopf/aim/internal/metadata"
)

type ActionPayload struct {
	Action string `json:"action"`
}

// RunUI starts the server, opens the browser, and waits for action.
func RunUI(meta *metadata.AppMetadata, staticFS http.FileSystem) (string, error) {
	// Create a listener on a random ephemeral port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("could not start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://127.0.0.1:%d", port)

	actionChan := make(chan string)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/metadata", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(meta)
	})

	mux.HandleFunc("/api/icon", func(w http.ResponseWriter, r *http.Request) {
		if meta.IconPath == "" {
			http.NotFound(w, r)
			return
		}

		// Serve file
		ext := filepath.Ext(meta.IconPath)
		if ext == ".svg" {
			w.Header().Set("Content-Type", "image/svg+xml")
		} else {
			w.Header().Set("Content-Type", "image/png")
		}
		http.ServeFile(w, r, meta.IconPath)
	})

	mux.HandleFunc("/api/action", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload ActionPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))

		actionChan <- payload.Action
	})

	if staticFS != nil {
		fsHandler := http.FileServer(staticFS)
		mux.Handle("/", func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				fsHandler.ServeHTTP(w, r)
			})
		}())
	} else {
		// Mock handler for debug without built dist
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Please build the webui first. Run: pnpm build in /web"))
		})
	}

	server := &http.Server{Handler: mux}

	// Run the server in background
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("UI server error: %v\n", err)
		}
	}()

	fmt.Printf("Installer UI running at %s\n", url)
	openBrowser(url)

	// Wait for the user to make a choice
	action := <-actionChan

	// Shut down safely
	server.Shutdown(context.Background())

	return action, nil
}

// openBrowser attempts to open the given URL via the default web browser
func openBrowser(url string) {
	var err error

	err = exec.Command("xdg-open", url).Start()

	if err != nil {
		fmt.Printf("Could not open browser automatically: %v\nPlease open %s in your browser manually.\n", err, url)
	}
}
