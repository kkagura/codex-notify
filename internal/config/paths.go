package config

import (
	"os"
	"path/filepath"
)

const appDirName = "codex-notify"

type Paths struct {
	BaseDir   string
	LogDir    string
	LogFile   string
	StateDir  string
	StateFile string
}

func ResolvePaths() Paths {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		if dir, err := os.UserCacheDir(); err == nil && dir != "" {
			base = dir
		} else {
			base = os.TempDir()
		}
	}

	root := filepath.Join(base, appDirName)
	logDir := filepath.Join(root, "logs")
	stateDir := filepath.Join(root, "state")

	return Paths{
		BaseDir:   root,
		LogDir:    logDir,
		LogFile:   filepath.Join(logDir, "app.log"),
		StateDir:  stateDir,
		StateFile: filepath.Join(stateDir, "dedupe.json"),
	}
}
