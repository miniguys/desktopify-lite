package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DefaultBrowser       string
	DefaultURLTemplate   string
	DefaultExtraFlags    string
	DefaultProxy         string
	DisableGoogleFavicon bool
	WithDebug            bool
}

type Input struct {
	URL             string
	IconURL         string
	Name            string
	Browser         string
	URLTemplate     string
	ExtraFlags      string
	StartupWMClass  string
	Proxy           string
	IconURLExplicit bool
}

type Paths struct {
	Home            string
	ApplicationsDir string
	IconsDir        string
}

func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("cannot determine home directory: %w", err)
	}

	dataHome := strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
	if dataHome == "" {
		dataHome = filepath.Join(home, ".local", "share")
	}

	return Paths{
		Home:            home,
		ApplicationsDir: filepath.Join(dataHome, "applications"),
		IconsDir:        filepath.Join(dataHome, "icons"),
	}, nil
}

func EnsureDirs(p Paths) error {
	if p.ApplicationsDir == "" || p.IconsDir == "" {
		return errors.New("internal error: empty paths")
	}
	if err := os.MkdirAll(p.ApplicationsDir, 0o755); err != nil {
		return fmt.Errorf("cannot create applications dir: %w", err)
	}
	if err := os.MkdirAll(p.IconsDir, 0o755); err != nil {
		return fmt.Errorf("cannot create icons dir: %w", err)
	}
	return nil
}

func (p Paths) DesktopFilePath(filename string) string {
	return filepath.Join(p.ApplicationsDir, filename)
}
