package kenv

import (
	"fmt"
	"os"
	"path/filepath"
)

type Info struct {
	KphpRoot string
}

func NewInfo() *Info {
	return &Info{}
}

func (info *Info) FindRoot() error {
	envRoot := os.Getenv("KPHP_ROOT")
	if envRoot != "" {
		if !fileExists(envRoot) {
			return fmt.Errorf("KPHP_ROOT points to a non-existing directory")
		}
		info.KphpRoot = envRoot
		return nil
	}

	home, err := os.UserHomeDir()
	homeKphp := filepath.Join(home, "kphp")
	if err == nil && fileExists(homeKphp) {
		info.KphpRoot = homeKphp
		return nil
	}

	return fmt.Errorf("KPHP_ROOT is not set")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
