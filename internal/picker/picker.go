package picker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type picker struct {
	message string
}

func Run() error {
	p := &picker{}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:         "Install an AppImage - aim",
			Width:         420,
			Height:        280,
			MinWidth:      420,
			MinHeight:     280,
			DisableResize: false,

			Content: ui.View(p.view),
		})
	})
	return mygo.App.Run()
}

func (p *picker) view(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Padding(24).Gap(16).Children(func() {
		ui.Text(c, "Install").FontSize(22).Bold()

		zone := ui.Column(c).Grow(1).Center().Gap(8).Radius(12).
			Background(t.Surface).Border(2, t.Border)
		if zone.FileDragOver() {
			zone.Border(2, t.Accent).Background(t.SurfaceHover)
		}
		zone.Children(func() {
			ui.Text(c, "Drop AppImage here").FontSize(14)
		})
		if files := zone.DroppedFiles(); files != nil {
			path, err := getPath(files)
			if err != nil {
				p.message = err.Error()
			} else {
				if err := launchInstaller(path); err != nil {
					p.message = fmt.Sprintf("Could not open installer: %v", err)
				} else {
					p.message = ""
				}
			}
		}

		if p.message != "" {
			ui.Text(c, p.message).TextColor(t.TextMuted)
		}
	})
}

func getPath(files []string) (string, error) {
	if len(files) != 1 {
		return "", fmt.Errorf("Drop one AppImage file at a time.")
	}
	path := files[0]
	if !strings.EqualFold(filepath.Ext(path), ".AppImage") {
		return "", fmt.Errorf("Choose a .AppImage file.")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("Could not access file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("Choose a regular AppImage file.")
	}
	return filepath.Abs(path)
}

func launchInstaller(path string) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	return exec.Command(executable, path).Start()
}
